package assetlibrary

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const (
	assetSyncScanInterval        = time.Second
	assetSyncLease               = 2 * time.Minute
	assetSyncChannelConcurrency  = 2
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
}

var assetSyncOnce sync.Once

const (
	assetSyncLeaderKey   = "asset_sync:leader"
	assetSyncLeaderTTL   = 30 * time.Second
	assetSyncLeaderRenew = 10 * time.Second
)

// Renew only a lock this node still owns, so a node that stalled past the TTL
// cannot extend the lease of the node that replaced it.
var renewAssetSyncLeader = redis.NewScript(`if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("PEXPIRE", KEYS[1], ARGV[2]) else return 0 end`)

// StartSyncTask starts the asset scheduler on every eligible node; a Redis
// lock lets exactly one of them dispatch at a time. Each node paces upstream
// requests in memory, so this single dispatcher is what keeps a channel within
// its QPM. If the leader stops, another node takes over once the lock expires.
// Without Redis every eligible node dispatches, as before. ASSET_SYNC_ENABLED
// can exclude a node without changing NODE_TYPE.
func StartSyncTask() {
	assetSyncOnce.Do(func() {
		if !common.GetEnvOrDefaultBool("ASSET_SYNC_ENABLED", common.IsMasterNode) {
			return
		}
		if !common.RedisEnabled {
			common.SysLog("asset synchronization runs without leader election because Redis is disabled; run it on one node to keep upstream QPM accurate")
		}
		scheduler := newAssetSyncScheduler(fmt.Sprintf("%s-assets-%s", common.NodeName, common.GetRandomString(8)))
		gopool.Go(scheduler.run)
	})
}

// holdsAssetSyncLeadership acquires or renews the lock that makes this node the
// only asset dispatcher.
func holdsAssetSyncLeadership(ctx context.Context, runnerID string) (bool, error) {
	if !common.RedisEnabled {
		return true, nil
	}
	renewed, err := renewAssetSyncLeader.Run(ctx, common.RDB, []string{assetSyncLeaderKey}, runnerID, assetSyncLeaderTTL.Milliseconds()).Int()
	if err != nil {
		return false, err
	}
	if renewed == 1 {
		return true, nil
	}
	return common.RDB.SetNX(ctx, assetSyncLeaderKey, runnerID, assetSyncLeaderTTL).Result()
}

// assetSyncScheduler is the single dispatcher for asset synchronization. Each
// time a channel has a free request slot it reads that channel's due work,
// picks the next kind by weighted rotation and leases only the replica it is
// about to send. Work that cannot be sent yet stays unleased in the database.
type assetSyncScheduler struct {
	runnerID string
	mu       sync.Mutex
	inFlight map[int]int
	turns    map[int]int
	// dispatched counts started jobs; only the dispatch goroutine touches it.
	dispatched int
	jobs       sync.WaitGroup
	finished   chan struct{}
}

func newAssetSyncScheduler(runnerID string) *assetSyncScheduler {
	return &assetSyncScheduler{
		runnerID: runnerID,
		inFlight: make(map[int]int),
		turns:    make(map[int]int),
		finished: make(chan struct{}, 1),
	}
}

func (scheduler *assetSyncScheduler) run() {
	timer := time.NewTimer(0)
	defer timer.Stop()
	leader := false
	var nextLeaderCheck time.Time
	for {
		select {
		case <-timer.C:
		case <-scheduler.finished:
		}
		if now := time.Now(); !now.Before(nextLeaderCheck) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			held, err := holdsAssetSyncLeadership(ctx, scheduler.runnerID)
			cancel()
			nextLeaderCheck = now.Add(assetSyncLeaderRenew)
			if err != nil {
				// Leadership cannot be proven, so stop dispatching rather than risk
				// two nodes sending against one channel budget.
				logger.LogWarn(context.Background(), fmt.Sprintf("asset synchronization leader check failed: %v", err))
			} else if held != leader {
				common.SysLog(fmt.Sprintf("asset synchronization leader=%t runner=%s", held, scheduler.runnerID))
			}
			leader = held && err == nil
		}
		// Since Go 1.23 Reset discards a pending tick, so no drain is needed.
		if !leader {
			timer.Reset(time.Until(nextLeaderCheck))
			continue
		}
		timer.Reset(min(scheduler.dispatch(), time.Until(nextLeaderCheck)))
	}
}

// runSyncPass processes currently due work round by round until none is left.
func runSyncPass(runnerID string) {
	scheduler := newAssetSyncScheduler(runnerID)
	for range 20 {
		before := scheduler.dispatched
		scheduler.dispatch()
		scheduler.jobs.Wait()
		if scheduler.dispatched == before {
			return
		}
	}
}

// dispatch starts at most one job per channel and returns how long to wait
// before the next round.
func (scheduler *assetSyncScheduler) dispatch() time.Duration {
	wait := assetSyncScanInterval
	var configs []model.AssetChannelConfig
	if err := model.DB.Select("channel_id", "qpm").Find(&configs).Error; err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("load asset channels for synchronization failed: %v", err))
		return 5 * time.Second
	}
	for _, config := range configs {
		scheduler.mu.Lock()
		busy := scheduler.inFlight[config.ChannelID] >= assetSyncChannelConcurrency
		scheduler.mu.Unlock()
		if busy {
			continue // a finishing job wakes the scheduler
		}
		if delay := channelRateDelay(config.ChannelID); delay > 0 {
			wait = min(wait, delay)
			continue
		}
		replica, err := scheduler.claimNext(config.ChannelID)
		if err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("claim asset replica failed: channel_id=%d error=%v", config.ChannelID, err))
			continue
		}
		if replica == nil {
			continue
		}
		reserveChannelRate(config.ChannelID, config.QPM)
		scheduler.mu.Lock()
		scheduler.inFlight[config.ChannelID]++
		scheduler.mu.Unlock()
		scheduler.dispatched++
		scheduler.jobs.Add(1)
		gopool.Go(func() {
			defer scheduler.finish(replica.ChannelID)
			syncReplica(withChannelRateReservation(context.Background(), replica.ChannelID), replica.LockedBy, replica)
		})
		// More work may already be due on this channel once its next slot opens.
		wait = min(wait, max(channelRateDelay(config.ChannelID), 10*time.Millisecond))
	}
	return wait
}

func (scheduler *assetSyncScheduler) finish(channelID int) {
	scheduler.mu.Lock()
	scheduler.inFlight[channelID]--
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
	now := common.GetTimestamp()
	start := scheduler.turns[channelID]
	loaded := make(map[model.AssetReplicaWorkload]bool, len(assetSyncTurns))
	for offset := range assetSyncTurns {
		turn := (start + offset) % len(assetSyncTurns)
		workload := assetSyncTurns[turn]
		if loaded[workload] {
			continue // already found nothing claimable for this kind
		}
		loaded[workload] = true
		candidates, err := model.ListDueAssetReplicas(now, channelID, workload, assetSyncCandidateLimit)
		if err != nil {
			return nil, err
		}
		for _, candidate := range candidates {
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
	var config model.AssetChannelConfig
	if err := model.DB.Where("channel_id = ?", replica.ChannelID).First(&config).Error; err != nil {
		failReplica(runnerID, replica, err)
		return
	}
	client, err := newProviderClient(&config)
	if err != nil {
		failReplica(runnerID, replica, err)
		return
	}
	if asset.Status == model.AssetStatusDeleted || replica.Operation == model.AssetReplicaOperationDelete {
		if replica.UpstreamAssetID != "" {
			if err = client.deleteAsset(ctx, replica.UpstreamAssetID); err != nil && !isProviderNotFound(err) {
				failReplica(runnerID, replica, err)
				return
			}
		}
		if err = model.FinishAssetReplica(replica.ID, runnerID, model.AssetReplicaStatusDeleted, 100, replica.UpstreamAssetID, 0, ""); err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("finish asset replica deletion failed: replica_id=%d error=%v", replica.ID, err))
			return
		}
		cleanupDeletedAssetStorage(&asset)
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
			groupReplica, groupErr := ensureGroupReplica(ctx, client, &group, replica.ChannelID)
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
		finishReplicaLater(runnerID, replica, model.AssetReplicaStatusProcessing, 50, upstreamID, assetFirstPollDelay, "")
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
		failReplica(runnerID, replica, errors.New(message))
	default:
		finishReplicaLater(runnerID, replica, model.AssetReplicaStatusProcessing, 75, replica.UpstreamAssetID, processingPollDelay(replica, common.GetTimestamp()), "")
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
		return assetActiveReviewInterval
	}
	return assetActiveReconcileInterval
}

func ensureGroupReplica(ctx context.Context, client assetProvider, group *model.AssetGroup, channelID int) (*model.AssetGroupReplica, error) {
	var replica model.AssetGroupReplica
	err := model.DB.Where("group_id = ? AND channel_id = ?", group.ID, channelID).First(&replica).Error
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
	replica.UpstreamGroupID = upstreamID
	replica.Status = model.AssetReplicaStatusActive
	replica.Progress = 100
	replica.LastError = ""
	replica.LastSyncedAt = now
	replica.UpdatedAt = now
	if replica.ID == 0 {
		err = model.DB.Create(&replica).Error
	} else {
		err = model.DB.Save(&replica).Error
	}
	return &replica, err
}

func failReplica(runnerID string, replica *model.AssetReplica, err error) {
	if replica.ClaimedFromStatus == model.AssetReplicaStatusActive {
		finishReplicaLater(runnerID, replica, model.AssetReplicaStatusActive, 100, replica.UpstreamAssetID, assetActiveRetryInterval, truncateError(err))
		logger.LogWarn(context.Background(), fmt.Sprintf("active asset review failed and will retry: replica_id=%d channel_id=%d error=%v", replica.ID, replica.ChannelID, err))
		return
	}
	next := common.GetTimestamp() + int64(model.AssetReplicaBackoff(replica.Attempts)/time.Second)
	if finishErr := model.FinishAssetReplica(replica.ID, runnerID, model.AssetReplicaStatusFailed, replica.Progress, replica.UpstreamAssetID, next, truncateError(err)); finishErr != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("record asset replica failure: replica_id=%d error=%v", replica.ID, finishErr))
	}
	logger.LogWarn(context.Background(), fmt.Sprintf("asset replica synchronization failed: replica_id=%d channel_id=%d error=%v", replica.ID, replica.ChannelID, err))
}

func finishReplicaLater(runnerID string, replica *model.AssetReplica, status string, progress int, upstreamID string, delay time.Duration, message string) {
	next := common.GetTimestamp() + int64(delay/time.Second)
	if err := model.FinishAssetReplica(replica.ID, runnerID, status, progress, upstreamID, next, message); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("finish asset replica pass failed: replica_id=%d error=%v", replica.ID, err))
	}
}

func cleanupDeletedAssetStorage(asset *model.MediaAsset) {
	var remaining int64
	if err := model.DB.Model(&model.AssetReplica{}).Where("asset_id = ? AND status <> ?", asset.ID, model.AssetReplicaStatusDeleted).Count(&remaining).Error; err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("count deleting asset replicas failed: asset_id=%s error=%v", asset.PublicID, err))
		return
	}
	if remaining == 0 {
		if err := storageService.DeleteAssetObject(asset.StorageObjectID, "asset replicas deleted"); err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("schedule asset object deletion failed: asset_id=%s error=%v", asset.PublicID, err))
		}
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
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "notfound") || strings.Contains(message, "not found") || strings.Contains(message, "404")
}
