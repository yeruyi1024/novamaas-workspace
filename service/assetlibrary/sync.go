package assetlibrary

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"

	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
)

const (
	assetSyncScanInterval        = time.Second
	assetSyncLease               = 2 * time.Minute
	assetSyncCandidateLimit      = 5
	assetFirstPollDelay          = 8 * time.Second
	assetActiveReviewWindow      = 48 * time.Hour
	assetActiveReviewInterval    = 30 * time.Minute
	assetActiveReconcileInterval = 24 * time.Hour
	assetActiveRetryInterval     = 5 * time.Minute
)

// assetSyncTurns is the weighted rotation used when several kinds of work are
// waiting on one channel. Status checks get two turns so finished uploads are
// noticed while a batch is still being submitted. A kind with no due work
// passes its turn on, so an idle kind never leaves request slots unused.
var assetSyncTurns = [...]model.AssetReplicaWorkload{
	model.AssetReplicaWorkloadSubmit,
	model.AssetReplicaWorkloadPoll,
	model.AssetReplicaWorkloadPoll,
	model.AssetReplicaWorkloadMaintenance,
	model.AssetReplicaWorkloadReview,
}

var assetSyncOnce sync.Once

// Eligible API nodes compete for one shared database lease. An explicit
// ASSET_SYNC_ENABLED=false excludes a node without changing NODE_TYPE.
func StartSyncTask() {
	assetSyncOnce.Do(func() {
		if err := RecordLibraryIdentities(); err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("asset-library identity recording deferred; existing configurations are retained: %v", err))
		}
		if !common.GetEnvOrDefaultBool("ASSET_SYNC_ENABLED", true) {
			return
		}
		scheduler := newAssetSyncScheduler(fmt.Sprintf("%s-assets-%s", common.NodeName, common.GetRandomString(8)))
		gopool.Go(func() { scheduler.run(context.Background()) })
	})
}

type assetSyncQueue struct {
	channelID int
	workload  model.AssetReplicaWorkload
}

// assetSyncScheduler is the single dispatcher for asset synchronization. Each
// time a channel has a free request slot it reads that channel's due work,
// picks the next kind by weighted rotation and leases only the replica it is
// about to send. Work that cannot be sent yet stays unleased in the database.
type assetSyncScheduler struct {
	runnerID     string
	mu           sync.Mutex
	inFlight     map[assetSyncQueue]int
	turns        map[int]int
	pending      map[assetSyncQueue][]*model.AssetReplica
	scanAfter    map[assetSyncQueue]time.Time
	configs      []model.AssetChannelConfig
	configsUntil time.Time
	// dispatched counts started jobs; only the dispatch goroutine touches it.
	dispatched int
	jobs       sync.WaitGroup
	finished   chan struct{}
}

func newAssetSyncScheduler(runnerID string) *assetSyncScheduler {
	return &assetSyncScheduler{
		runnerID:  runnerID,
		inFlight:  make(map[assetSyncQueue]int),
		pending:   make(map[assetSyncQueue][]*model.AssetReplica),
		scanAfter: make(map[assetSyncQueue]time.Time),
		turns:     make(map[int]int),
		finished:  make(chan struct{}, 1),
	}
}

func (scheduler *assetSyncScheduler) run(ctx context.Context) {
	timer := time.NewTimer(0)
	defer timer.Stop()
	var leaderCtx context.Context
	var cancelLeader context.CancelFunc
	var expiry *time.Timer
	defer func() {
		if expiry != nil {
			expiry.Stop()
		}
		if cancelLeader != nil {
			cancelLeader()
		}
		scheduler.jobs.Wait()
	}()
	var checkAt, lastError time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		case <-scheduler.finished:
		}
		if started := time.Now(); !started.Before(checkAt) {
			checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			held, err := model.TryHoldAssetSyncLeadership(checkCtx, scheduler.runnerID, 30)
			cancel()
			checkAt = started.Add(10 * time.Second)
			wasLeader := leaderCtx != nil && leaderCtx.Err() == nil
			if err != nil || !held {
				if cancelLeader != nil {
					cancelLeader()
				}
				leaderCtx = nil
				if err != nil && (lastError.IsZero() || time.Since(lastError) >= time.Minute) {
					logger.LogWarn(ctx, fmt.Sprintf("asset synchronization leader check failed: %v", err))
					lastError = time.Now()
				}
			} else {
				// If the old expiry callback has started, its context cannot be
				// reused. Its cancellation must never affect the new lease.
				if expiry != nil && !expiry.Stop() {
					cancelLeader()
					leaderCtx = nil
				}
				if leaderCtx == nil || leaderCtx.Err() != nil {
					leaderCtx, cancelLeader = context.WithCancel(ctx)
				}
				// Stop before the DB lease expires, even if a claim/renewal stalls.
				// Starting from the request time also subtracts DB/network latency.
				expiry = time.AfterFunc(time.Until(started.Add(27*time.Second)), cancelLeader)
			}
			isLeader := leaderCtx != nil && leaderCtx.Err() == nil
			if wasLeader != isLeader {
				common.SysLog(fmt.Sprintf("asset synchronization leader=%t runner=%s", isLeader, scheduler.runnerID))
			}
		}
		if leaderCtx == nil || leaderCtx.Err() != nil {
			timer.Reset(max(time.Millisecond, time.Until(checkAt)))
			continue
		}
		timer.Reset(min(scheduler.dispatch(leaderCtx), max(time.Millisecond, time.Until(checkAt))))
	}
}

// dispatch starts at most one job per channel and returns how long to wait
// before the next round.
func (scheduler *assetSyncScheduler) dispatch(parent context.Context) time.Duration {
	if parent.Err() != nil {
		return assetSyncScanInterval
	}
	wait := assetSyncScanInterval
	if !time.Now().Before(scheduler.configsUntil) {
		configs, err := assetSyncConfigs()
		if err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("load asset channels for synchronization failed: %v", err))
			return 5 * time.Second
		}
		scheduler.configs = configs
		scheduler.configsUntil = time.Now().Add(assetSyncScanInterval)
	}
	for _, config := range scheduler.configs {
		if config.DeleteRequestedAt > 0 {
			finished, err := model.FinalizeAssetLibraryDeletion(config.LibraryID)
			if err != nil {
				logger.LogWarn(parent, fmt.Sprintf("finalize asset library deletion failed: library_id=%d error=%v", config.LibraryID, err))
				continue
			}
			if finished {
				scheduler.configsUntil = time.Time{}
				if err := RecordLibraryIdentities(); err != nil {
					logger.LogWarn(parent, fmt.Sprintf("refresh asset library identities after deletion failed: %v", err))
				}
				continue
			}
		}
		rateID := config.RateLimitID()
		if parent.Err() != nil {
			break
		}
		if delay := channelRateDelay(rateID); delay > 0 {
			wait = min(wait, delay)
			continue
		}
		replica, err := scheduler.claimNextConfig(config)
		if err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("claim asset replica failed: channel_id=%d error=%v", config.ChannelID, err))
			continue
		}
		if replica == nil {
			continue
		}
		if parent.Err() != nil {
			_ = model.FinishAssetReplica(replica.ID, replica.LockedBy, replica.ClaimedFromStatus, replica.Progress, replica.UpstreamAssetID, replica.NextSyncAt, replica.LastError, replica.Operation)
			break
		}
		allowedAt := reserveChannelRate(rateID, config.QPM)
		ctx := withChannelRateReservation(parent, rateID, config.QPM, allowedAt, replicaWorkload(replica))
		key := assetSyncQueue{config.ChannelID, replicaWorkload(replica)}
		if config.LibraryID > 0 {
			key.channelID = -config.LibraryID
		}
		scheduler.mu.Lock()
		scheduler.inFlight[key]++
		scheduler.mu.Unlock()
		scheduler.dispatched++
		scheduler.jobs.Add(1)
		gopool.Go(func() {
			defer scheduler.finish(key, ctx)
			syncReplica(ctx, replica.LockedBy, replica)
		})
		// More work may already be due on this channel once its next slot opens.
		wait = min(wait, max(channelRateDelay(rateID), 10*time.Millisecond))
	}
	return wait
}

func (scheduler *assetSyncScheduler) finish(key assetSyncQueue, ctx context.Context) {
	releaseUnusedChannelReservation(ctx)
	scheduler.mu.Lock()
	scheduler.inFlight[key]--
	scheduler.mu.Unlock()
	scheduler.jobs.Done()
	select {
	case scheduler.finished <- struct{}{}:
	default:
	}
}

// claimNext leases the oldest due replica of the next kind in the channel's
// rotation, skipping kinds with nothing due.
func (scheduler *assetSyncScheduler) claimNext(channelID int) (*model.AssetReplica, error) {
	return scheduler.claimNextConfig(model.AssetChannelConfig{ChannelID: channelID, Enabled: true})
}

func (scheduler *assetSyncScheduler) claimNextConfig(config model.AssetChannelConfig) (*model.AssetReplica, error) {
	channelID := config.ChannelID
	if config.LibraryID > 0 {
		channelID = -config.LibraryID
	}
	now := common.GetTimestamp()
	start := scheduler.turns[channelID]
	loaded := make(map[model.AssetReplicaWorkload]bool, len(assetSyncTurns))
	for offset := range assetSyncTurns {
		turn := (start + offset) % len(assetSyncTurns)
		workload := assetSyncTurns[turn]
		if config.LibraryID > 0 && !config.Enabled && workload != model.AssetReplicaWorkloadMaintenance {
			continue
		}
		if loaded[workload] {
			continue // already found nothing claimable for this kind
		}
		loaded[workload] = true
		key := assetSyncQueue{channelID, workload}
		scheduler.mu.Lock()
		busy := scheduler.inFlight[key] > 0
		scheduler.mu.Unlock()
		if busy || (workload == model.AssetReplicaWorkloadReview && channelReviewRateDelay(config.RateLimitID()) > 0) {
			continue
		}
		if len(scheduler.pending[key]) == 0 {
			if time.Now().Before(scheduler.scanAfter[key]) {
				continue
			}
			var candidates []*model.AssetReplica
			var err error
			if config.LibraryID > 0 {
				candidates, err = model.ListDueAssetLibraryReplicas(now, config.LibraryID, workload, assetSyncCandidateLimit)
			} else {
				candidates, err = model.ListDueAssetReplicas(now, config.ChannelID, workload, assetSyncCandidateLimit)
			}
			if err != nil {
				return nil, err
			}
			scheduler.pending[key] = candidates
			scheduler.scanAfter[key] = time.Now().Add(assetSyncScanInterval)
		}
		for len(scheduler.pending[key]) > 0 {
			candidate := scheduler.pending[key][0]
			scheduler.pending[key] = scheduler.pending[key][1:]
			attemptID := scheduler.runnerID + "-" + common.GetRandomString(16)
			claimed, err := model.ClaimAssetReplica(candidate, now, now+int64(assetSyncLease/time.Second), attemptID)
			if err != nil {
				return nil, err
			}
			if claimed {
				scheduler.turns[channelID] = turn + 1
				return candidate, nil
			}
		}
	}
	return nil, nil
}

func replicaWorkload(replica *model.AssetReplica) model.AssetReplicaWorkload {
	if replica.Operation == model.AssetReplicaOperationDelete {
		return model.AssetReplicaWorkloadMaintenance
	}
	if replica.Status == model.AssetReplicaStatusActive || replica.ClaimedFromStatus == model.AssetReplicaStatusActive || replica.Progress >= 100 {
		return model.AssetReplicaWorkloadReview
	}
	if replica.UpstreamAssetID == "" {
		return model.AssetReplicaWorkloadSubmit
	}
	return model.AssetReplicaWorkloadPoll
}

// processingPollDelay spaces status checks out as an upload takes longer, so a
// slow asset does not keep spending the channel's request budget.
func processingPollDelay(replica *model.AssetReplica, now int64) time.Duration {
	if replica.CreatedAt <= 0 {
		return 15 * time.Second
	}
	switch elapsed := now - replica.CreatedAt; {
	case elapsed < 60:
		return 3 * time.Second
	case elapsed < 120:
		return 5 * time.Second
	case elapsed < 300:
		return 10 * time.Second
	default:
		return 15 * time.Second
	}
}

func syncReplica(parent context.Context, runnerID string, replica *model.AssetReplica) {
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	defer cancel()
	ctx = withRequestLogContext(ctx, "sync", replica.ID, replica.AssetID)
	var asset model.MediaAsset
	if err := model.DB.First(&asset, replica.AssetID).Error; err != nil {
		failReplica(runnerID, replica, err)
		return
	}
	if asset.Status == model.AssetStatusUnavailable && replica.Operation != model.AssetReplicaOperationDelete {
		if err := model.FinishAssetReplica(replica.ID, runnerID, model.AssetReplicaStatusRejected, replica.Progress, replica.UpstreamAssetID, 0, asset.UnavailableReason); err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("finish unavailable asset replica failed: replica_id=%d error=%v", replica.ID, err))
		}
		return
	}
	deleting := asset.Status == model.AssetStatusDeleted || replica.Operation == model.AssetReplicaOperationDelete
	var deleteID string
	if deleting {
		from := replica.ClaimedFromStatus
		if err := model.DB.Where("id = ? AND locked_by = ?", replica.ID, runnerID).First(replica).Error; err != nil {
			failReplica(runnerID, replica, err)
			return
		}
		replica.ClaimedFromStatus = from
		deleteID = replica.UpstreamAssetID
		if deleteID == "" {
			if err := model.AcknowledgeAssetReplicaDeletion(replica.ID, runnerID, ""); err != nil {
				logger.LogWarn(ctx, fmt.Sprintf("finish asset cancellation failed: %v", err))
			}
			return
		}
	}
	if !deleting && replica.ClaimedFromStatus != model.AssetReplicaStatusActive && replica.Progress < 100 {
		if replica.CreatedAt == 0 {
			replica.CreatedAt = common.GetTimestamp()
			if err := model.DB.Model(replica).Where("locked_by = ?", runnerID).Update("created_at", replica.CreatedAt).Error; err != nil {
				failReplica(runnerID, replica, err)
				return
			}
		}
		if common.GetTimestamp() >= replica.CreatedAt+300 {
			finishReplicaLater(runnerID, replica, model.AssetReplicaStatusFailed, replica.Progress, replica.UpstreamAssetID, 0, "asset synchronization timed out after 5 minutes")
			return
		}
	}
	var config model.AssetChannelConfig
	if replica.LibraryID != nil {
		var library model.AssetLibraryConnection
		if err := model.DB.First(&library, *replica.LibraryID).Error; err != nil {
			failReplica(runnerID, replica, err)
			return
		}
		config = library.ProviderConfig(0)
	} else if err := model.DB.Where("channel_id = ?", replica.ChannelID).First(&config).Error; err != nil {
		failReplica(runnerID, replica, err)
		return
	}
	client, err := newProviderClient(&config)
	if err != nil {
		failReplica(runnerID, replica, err)
		return
	}
	if deleting {
		shared := false
		if config.DeleteRequestedAt > 0 {
			shared, err = upstreamAssetIsShared(config, deleteID)
			if err != nil {
				failReplica(runnerID, replica, err)
				return
			}
		}
		if !shared {
			err = client.deleteAsset(ctx, deleteID)
		}
		if err != nil && !isProviderNotFound(err) {
			failReplica(runnerID, replica, err)
			return
		}
		if err := model.AcknowledgeAssetReplicaDeletion(replica.ID, runnerID, deleteID); err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("acknowledge asset deletion failed: %v", err))
		}
		return
	}
	if !config.Enabled {
		finishReplicaLater(runnerID, replica, model.AssetReplicaStatusFailed, replica.Progress, replica.UpstreamAssetID, 24*time.Hour, "asset channel is disabled")
		return
	}
	if replica.UpstreamAssetID == "" {
		upstreamGroupID := ""
		if client.requiresGroup() {
			var group model.AssetGroup
			if err = model.DB.First(&group, asset.GroupID).Error; err != nil {
				failReplica(runnerID, replica, err)
				return
			}
			groupReplica, groupErr := ensureGroupReplica(ctx, client, &group, replica.ChannelID, config.LibraryID)
			if groupErr != nil {
				failReplica(runnerID, replica, groupErr)
				return
			}
			upstreamGroupID = groupReplica.UpstreamGroupID
		}
		sourceURL, err := storageService.PresignAssetObject(ctx, asset.StorageObjectID, 24*time.Hour, storageService.AssetObjectURLOriginal, "")
		if err != nil {
			failReplica(runnerID, replica, err)
			return
		}
		allowed, err := model.CanSubmitAssetReplica(replica.ID, runnerID)
		if err != nil {
			failReplica(runnerID, replica, err)
			return
		}
		if !allowed {
			if err := model.FinishAssetReplica(replica.ID, runnerID, model.AssetReplicaStatusPending, replica.Progress, replica.UpstreamAssetID, common.GetTimestamp(), ""); err != nil {
				logger.LogWarn(ctx, fmt.Sprintf("cancel stale submission failed: %v", err))
			}
			return
		}
		upstreamID, err := client.createAsset(ctx, upstreamGroupID, asset.Name, asset.AssetType, sourceURL)
		if err != nil {
			if reason := upstreamContentRejectionReason(err); reason != "" {
				rejectReplica(runnerID, replica, reason)
				return
			}
			failReplica(runnerID, replica, err)
			return
		}
		// Upstream processing takes several seconds; an earlier check would
		// almost always spend a request on a "processing" answer.
		if err := model.FinishAssetReplicaSubmission(replica.ID, runnerID, upstreamID, common.GetTimestamp()+int64(assetFirstPollDelay/time.Second)); err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("persist asset submission failed: replica_id=%d error=%v", replica.ID, err))
		}
		return
	}
	result, err := client.getAsset(ctx, replica.UpstreamAssetID)
	if err != nil {
		if reason := upstreamContentRejectionReason(err); reason != "" {
			rejectReplica(runnerID, replica, reason)
			return
		}
		failReplica(runnerID, replica, err)
		return
	}
	switch strings.ToLower(result.Status) {
	case "active", "ready", "succeeded", "success":
		if err = activateReplicaAndQueueWebhook(runnerID, replica); err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("finish active asset replica failed: replica_id=%d error=%v", replica.ID, err))
		}
	case "failed", "error", "rejected":
		if reason := processingContentRejectionReason(result); reason != "" {
			rejectReplica(runnerID, replica, reason)
			return
		}
		message := result.Message
		if message == "" {
			message = "upstream asset processing failed"
		}
		finishReplicaLater(runnerID, replica, model.AssetReplicaStatusFailed, replica.Progress, replica.UpstreamAssetID, 0, message)
	default:
		progress := 75
		if replica.ClaimedFromStatus == model.AssetReplicaStatusActive || replica.Progress >= 100 {
			progress = 100 // Review processing must not restart initial synchronization.
		}
		finishReplicaLater(runnerID, replica, model.AssetReplicaStatusProcessing, progress, replica.UpstreamAssetID, processingPollDelay(replica, common.GetTimestamp()), "")
	}
}

func rejectReplica(runnerID string, replica *model.AssetReplica, reason string) {
	if err := rejectReplicaAndQueueWebhook(runnerID, replica, reason); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("record asset content rejection failed: replica_id=%d error=%v", replica.ID, err))
		return
	}
	logger.LogWarn(context.Background(), fmt.Sprintf("asset content rejected by upstream: replica_id=%d channel_id=%d reason=%s", replica.ID, replica.ChannelID, reason))
}

func activateReplicaAndQueueWebhook(runnerID string, replica *model.AssetReplica) error {
	next := common.GetTimestamp() + int64(activeReviewDelay(replica)/time.Second)
	return model.DB.Transaction(func(tx *gorm.DB) error {
		asset, changed, err := model.ActivateAssetReplicaWithTx(tx, replica.ID, runnerID, replica.UpstreamAssetID, next)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}
		return queueAssetWebhookDeliveries(tx, asset, WebhookEventAssetActive)
	})
}

func rejectReplicaAndQueueWebhook(runnerID string, replica *model.AssetReplica, reason string) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		asset, changed, err := model.RejectAssetReplicaWithTx(tx, replica.ID, runnerID, reason, replica.UpstreamAssetID)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}
		return queueAssetWebhookDeliveries(tx, asset, WebhookEventAssetFailed)
	})
}

func activeReviewDelay(replica *model.AssetReplica) time.Duration {
	if replica != nil && replica.CreatedAt > 0 && common.GetTimestamp()-replica.CreatedAt < int64(assetActiveReviewWindow/time.Second) {
		return jitteredAssetReviewDelay(replica.ID, assetActiveReviewInterval)
	}
	return jitteredAssetReviewDelay(replica.ID, assetActiveReconcileInterval)
}

func jitteredAssetReviewDelay(id int64, interval time.Duration) time.Duration {
	h := fnv.New32a()
	_, _ = fmt.Fprint(h, id)
	return interval * time.Duration(80+h.Sum32()%41) / 100
}

func ensureGroupReplica(ctx context.Context, client assetProvider, group *model.AssetGroup, channelID int, libraryIDs ...int) (*model.AssetGroupReplica, error) {
	var replica model.AssetGroupReplica
	query := model.DB.Where("group_id = ?", group.ID)
	if len(libraryIDs) > 0 && libraryIDs[0] > 0 {
		query = query.Where("library_id = ?", libraryIDs[0])
	} else {
		query = query.Where("channel_id = ? AND library_id IS NULL", channelID)
	}
	err := query.First(&replica).Error
	if err == nil && replica.UpstreamGroupID != "" && replica.Status == model.AssetReplicaStatusActive {
		return &replica, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	upstreamID, err := client.createGroup(ctx, group.Name, group.Description)
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	replica.GroupID = group.ID
	replica.ChannelID = channelID
	if len(libraryIDs) > 0 && libraryIDs[0] > 0 {
		replica.ChannelID = 0
		replica.LibraryID = new(int)
		*replica.LibraryID = libraryIDs[0]
	}
	replica.UpstreamGroupID = upstreamID
	replica.Status = model.AssetReplicaStatusActive
	replica.Progress = 100
	replica.LastError = ""
	replica.LastSyncedAt = now
	replica.UpdatedAt = now
	if replica.ID == 0 {
		err = model.DB.Create(&replica).Error
	} else {
		err = model.DB.Model(&replica).Select("upstream_group_id", "status", "progress", "last_error", "last_synced_at", "updated_at").Updates(&replica).Error
	}
	return &replica, err
}

func failReplica(runnerID string, replica *model.AssetReplica, err error) {
	if errors.Is(err, context.Canceled) {
		_ = model.FinishAssetReplica(replica.ID, runnerID, replica.ClaimedFromStatus, replica.Progress, replica.UpstreamAssetID, common.GetTimestamp(), replica.LastError, replica.Operation)
		return
	}
	if replica.ClaimedFromStatus == model.AssetReplicaStatusActive || replica.Progress >= 100 {
		status := model.AssetReplicaStatusProcessing
		if replica.ClaimedFromStatus == model.AssetReplicaStatusActive {
			status = model.AssetReplicaStatusActive
		}
		finishReplicaLater(runnerID, replica, status, 100, replica.UpstreamAssetID, assetActiveRetryInterval, truncateError(err))
		logger.LogWarn(context.Background(), fmt.Sprintf("active asset review failed and will retry: replica_id=%d channel_id=%d error=%v", replica.ID, replica.ChannelID, err))
		return
	}
	next := common.GetTimestamp() + int64(model.AssetReplicaBackoff(replica.Attempts)/time.Second)
	if replica.CreatedAt > 0 && replica.Progress < 100 && replica.Operation != model.AssetReplicaOperationDelete {
		next = min(next, replica.CreatedAt+300)
	}
	if finishErr := model.FinishAssetReplica(replica.ID, runnerID, model.AssetReplicaStatusFailed, replica.Progress, replica.UpstreamAssetID, next, truncateError(err), replica.Operation); finishErr != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("record asset replica failure: replica_id=%d error=%v", replica.ID, finishErr))
	}
	logger.LogWarn(context.Background(), fmt.Sprintf("asset replica synchronization failed: replica_id=%d channel_id=%d error=%v", replica.ID, replica.ChannelID, err))
}

func finishReplicaLater(runnerID string, replica *model.AssetReplica, status string, progress int, upstreamID string, delay time.Duration, message string) {
	next := common.GetTimestamp() + int64(delay/time.Second)
	if delay == 0 {
		next = 0
	} else if replica.CreatedAt > 0 && replica.Progress < 100 && replica.ClaimedFromStatus != model.AssetReplicaStatusActive && replica.Operation != model.AssetReplicaOperationDelete {
		next = min(next, replica.CreatedAt+300)
	}
	if err := model.FinishAssetReplica(replica.ID, runnerID, status, progress, upstreamID, next, message); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("finish asset replica pass failed: replica_id=%d error=%v", replica.ID, err))
	}
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 2000 {
		message = message[:2000]
	}
	return message
}

func isProviderNotFound(err error) bool {
	var upstream *upstreamAssetError
	if !errors.As(err, &upstream) {
		return false
	}
	if upstream.statusCode == 404 {
		return true
	}
	if upstream.statusCode != 400 && (upstream.statusCode < 200 || upstream.statusCode >= 300) {
		return false
	}
	code := strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(upstream.code))
	return strings.HasSuffix(code, "notfound")
}
