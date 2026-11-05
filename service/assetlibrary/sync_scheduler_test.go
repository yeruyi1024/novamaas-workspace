package assetlibrary

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type dueReplicaSeed struct {
	label      string
	status     string
	upstreamID string
}

func seedDueReplicas(t *testing.T, db *gorm.DB, channelID int, seeds []dueReplicaSeed) map[int64]string {
	t.Helper()
	group := model.AssetGroup{PublicID: fmt.Sprintf("group-%d", channelID), OwnerUserID: 42, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	due := common.GetTimestamp() - 10
	labels := make(map[int64]string, len(seeds))
	for index, seed := range seeds {
		asset := model.MediaAsset{
			PublicID: fmt.Sprintf("asset-%d-%d", channelID, index), OwnerUserID: 42, GroupID: group.ID,
			StorageObjectID: int64(channelID*100 + index + 1), AssetType: model.AssetTypeImage, Status: model.AssetStatusProcessing,
		}
		require.NoError(t, db.Create(&asset).Error)
		replica := model.AssetReplica{
			AssetID: asset.ID, ChannelID: channelID, Operation: model.AssetReplicaOperationSync,
			Status: seed.status, UpstreamAssetID: seed.upstreamID, NextSyncAt: due,
		}
		require.NoError(t, db.Create(&replica).Error)
		labels[replica.ID] = seed.label
	}
	return labels
}

func TestAssetSchedulerRotatesWorkloadsAndLeasesOnlyTheDispatchedReplica(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	labels := seedDueReplicas(t, db, 7101, []dueReplicaSeed{
		{"submit-1", model.AssetReplicaStatusPending, ""},
		{"submit-2", model.AssetReplicaStatusPending, ""},
		{"poll-1", model.AssetReplicaStatusProcessing, "upstream-1"},
		{"poll-2", model.AssetReplicaStatusProcessing, "upstream-2"},
		{"poll-3", model.AssetReplicaStatusProcessing, "upstream-3"},
		{"review-1", model.AssetReplicaStatusActive, "upstream-4"},
	})
	scheduler := newAssetSyncScheduler("scheduler-test")

	first, err := scheduler.claimNext(7101)
	require.NoError(t, err)
	require.NotNil(t, first)
	var leased int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("locked_by <> ?", "").Count(&leased).Error)
	assert.EqualValues(t, 1, leased, "waiting work must stay unleased in the database")

	order := []string{labels[first.ID]}
	for {
		replica, err := scheduler.claimNext(7101)
		require.NoError(t, err)
		if replica == nil {
			break
		}
		order = append(order, labels[replica.ID])
	}
	// Polls take two turns; once a kind runs out, its turns go to the others.
	assert.Equal(t, []string{"submit-1", "poll-1", "poll-2", "review-1", "submit-2", "poll-3"}, order)
}

func TestAssetSchedulerGivesIdleTurnsToWaitingWork(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	labels := seedDueReplicas(t, db, 7102, []dueReplicaSeed{
		{"submit-1", model.AssetReplicaStatusPending, ""},
		{"submit-2", model.AssetReplicaStatusPending, ""},
		{"submit-3", model.AssetReplicaStatusPending, ""},
	})
	scheduler := newAssetSyncScheduler("scheduler-test")

	var order []string
	for range 3 {
		replica, err := scheduler.claimNext(7102)
		require.NoError(t, err)
		require.NotNil(t, replica)
		order = append(order, labels[replica.ID])
	}
	assert.Equal(t, []string{"submit-1", "submit-2", "submit-3"}, order)
}

func TestAssetSchedulerLeavesWorkUnleasedWhileChannelSlotIsTaken(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	const channelID = 7103
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: channelID, QPM: 1}).Error)
	seedDueReplicas(t, db, channelID, []dueReplicaSeed{{"submit-1", model.AssetReplicaStatusPending, ""}})
	reserveChannelRate(channelID, 1)
	t.Cleanup(func() {
		channelRateState.Lock()
		delete(channelRateState.next, channelID)
		channelRateState.Unlock()
	})

	wait := newAssetSyncScheduler("scheduler-test").dispatch(context.Background())

	assert.Greater(t, wait, time.Duration(0))
	assert.LessOrEqual(t, wait, assetSyncScanInterval)
	var replica model.AssetReplica
	require.NoError(t, db.Where("channel_id = ?", channelID).First(&replica).Error)
	assert.Empty(t, replica.LockedBy)
	assert.Equal(t, model.AssetReplicaStatusPending, replica.Status)
}

func TestDispatchedJobUsesItsReservedRequestSlotOnce(t *testing.T) {
	const channelID = 7104
	t.Cleanup(func() {
		channelRateState.Lock()
		delete(channelRateState.next, channelID)
		channelRateState.Unlock()
	})
	allowedAt := reserveChannelRate(channelID, 1)
	ctx := withChannelRateReservation(context.Background(), channelID, 1, allowedAt, model.AssetReplicaWorkloadPoll)

	require.NoError(t, waitForChannelRate(ctx, channelID, 1), "the first call uses the slot taken at dispatch")

	// A further call in the same job needs a new slot, one minute later at QPM 1.
	waitCtx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	assert.ErrorIs(t, waitForChannelRate(waitCtx, channelID, 1), context.DeadlineExceeded)
}

func TestProcessingPollDelayBacksOffAsUploadAges(t *testing.T) {
	const now int64 = 10_000
	tests := []struct {
		name      string
		createdAt int64
		want      time.Duration
	}{
		{"first minute", now - 10, 3 * time.Second},
		{"second minute", now - 90, 5 * time.Second},
		{"within five minutes", now - 200, 10 * time.Second},
		{"older", now - 600, 15 * time.Second},
		{"unknown age", 0, 15 * time.Second},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, processingPollDelay(&model.AssetReplica{CreatedAt: test.createdAt}, now))
		})
	}
}

func startAssetSyncSchedulerForTest(t *testing.T, id string) func() {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		newAssetSyncScheduler(id).run(ctx)
	}()
	stop := sync.OnceFunc(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			require.FailNow(t, "asset synchronization workers did not finish")
		}
	})
	t.Cleanup(stop)
	return stop
}

func TestAssetSyncQueriesReadinessWhileSubmissionIsBlocked(t *testing.T) {
	for _, test := range []struct {
		name        string
		pollChannel int
	}{
		{"same-channel", 701}, {"independent-channel-and-quota", 702},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := setupAssetLibraryTestDB(t)
			seedAssetWebhookOwner(t)
			t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "independent-poll-test-key")
			credential, err := encryptChannelCredential("provider-secret")
			require.NoError(t, err)
			submitted, queried := make(chan struct{}, 1), make(chan struct{}, 1)
			release := make(chan struct{})
			unblock := sync.OnceFunc(func() { close(release) })
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Query().Get("Action") {
				case "CreateAssetGroup":
					submitted <- struct{}{}
					<-release
					_, _ = io.WriteString(w, `{"Result":{"Id":"upstream-group"}}`)
				case "GetAsset":
					var payload struct {
						ID string `json:"Id"`
					}
					assert.NoError(t, common.DecodeJson(r.Body, &payload))
					assert.Equal(t, "already-ready", payload.ID)
					queried <- struct{}{}
					_, _ = io.WriteString(w, `{"Result":{"Status":"Active"}}`)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			t.Cleanup(upstream.Close)
			channelIDs := []int{701}
			if test.pollChannel != 701 {
				channelIDs = append(channelIDs, test.pollChannel)
			}
			for _, channelID := range channelIDs {
				qpm := 1000000
				if test.pollChannel != 701 && channelID == 701 {
					qpm = 1
				}
				require.NoError(t, db.Create(&model.AssetChannelConfig{
					ChannelID: channelID, Enabled: true, Protocol: model.AssetChannelProtocolVolcAction,
					AuthType: model.AssetChannelAuthBearer, BaseURL: upstream.URL,
					EncryptedCredential: credential, QPM: qpm,
				}).Error)
			}
			t.Cleanup(func() {
				channelRateState.Lock()
				defer channelRateState.Unlock()
				for _, channelID := range channelIDs {
					delete(channelRateState.next, channelID)
					delete(channelRateState.sentNext, channelID)
					delete(channelRateState.reviewNext, channelID)
				}
			})
			group := model.AssetGroup{PublicID: "group-independent-poll", OwnerUserID: 42, Status: model.AssetStatusReady}
			require.NoError(t, db.Create(&group).Error)
			assets := []model.MediaAsset{
				{PublicID: "asset-submitting", OwnerUserID: 42, GroupID: group.ID, StorageObjectID: 101, AssetType: model.AssetTypeImage, Status: model.AssetStatusProcessing},
				{PublicID: "asset-awaiting-poll", OwnerUserID: 42, GroupID: group.ID, StorageObjectID: 102, AssetType: model.AssetTypeImage, Status: model.AssetStatusProcessing},
			}
			require.NoError(t, db.Create(&assets).Error)
			require.NoError(t, db.Create(&[]model.AssetReplica{
				{AssetID: assets[0].ID, ChannelID: 701, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusPending},
				{AssetID: assets[1].ID, ChannelID: test.pollChannel, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusProcessing, UpstreamAssetID: "already-ready"},
			}).Error)
			stop := startAssetSyncSchedulerForTest(t, "unified-scheduler")
			t.Cleanup(unblock)
			select {
			case <-submitted:
			case <-time.After(5 * time.Second):
				require.FailNow(t, "submission did not reach the upstream")
			}
			select {
			case <-queried:
			case <-time.After(5 * time.Second):
				require.FailNow(t, "readiness query was blocked by submission")
			}
			unblock()
			require.Eventually(t, func() bool {
				var ready model.MediaAsset
				return db.First(&ready, assets[1].ID).Error == nil && ready.Status == model.AssetStatusReady
			}, 5*time.Second, 10*time.Millisecond)
			stop()
			require.NoError(t, db.First(&assets[1], assets[1].ID).Error)
			assert.Equal(t, model.AssetStatusReady, assets[1].Status)
		})
	}
}

func TestAssetSchedulerReleasesUnusedSlotAndKeepsTypesIndependent(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "local-scheduler-fixture")
	credential, err := encryptChannelCredential("fixture-secret")
	require.NoError(t, err)
	const channelID = 7181
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: channelID, QPM: 1, EncryptedCredential: credential}).Error)
	seedDueReplicas(t, db, channelID, []dueReplicaSeed{
		{"disabled-1", model.AssetReplicaStatusPending, ""},
		{"disabled-2", model.AssetReplicaStatusPending, ""},
	})
	scheduler := newAssetSyncScheduler("disabled-channel")
	scheduler.dispatch(context.Background())
	scheduler.jobs.Wait()
	scheduler.dispatch(context.Background())
	scheduler.jobs.Wait()
	assert.Equal(t, 2, scheduler.dispatched, "locally rejected work must not consume a one-minute upstream slot")

	seedDueReplicas(t, db, channelID+1, []dueReplicaSeed{
		{"waiting-submit", model.AssetReplicaStatusPending, ""},
		{"due-poll", model.AssetReplicaStatusProcessing, "ready-id"},
	})
	scheduler.inFlight[assetSyncQueue{channelID + 1, model.AssetReplicaWorkloadSubmit}] = 1
	candidate, err := scheduler.claimNext(channelID + 1)
	require.NoError(t, err)
	require.NotNil(t, candidate)
	assert.Equal(t, "ready-id", candidate.UpstreamAssetID, "a slow submit must leave the query slot usable")
}

func TestAssetSchedulerProcessingTimeoutStopsUntilExplicitRetry(t *testing.T) {
	for _, test := range []struct{ name, upstreamID, initialStatus, retryStatus string }{
		{"submitted", "upstream-expired", model.AssetReplicaStatusProcessing, model.AssetReplicaStatusProcessing},
		{"not-yet-submitted", "", model.AssetReplicaStatusPending, model.AssetReplicaStatusPending},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := setupAssetLibraryTestDB(t)
			labels := seedDueReplicas(t, db, 7182, []dueReplicaSeed{{"expired", test.initialStatus, test.upstreamID}})
			var id int64
			for key := range labels {
				id = key
			}
			require.NoError(t, db.Model(&model.AssetReplica{}).Where("id = ?", id).Update("created_at", common.GetTimestamp()-301).Error)
			replica, err := newAssetSyncScheduler("timeout").claimNext(7182)
			require.NoError(t, err)
			require.NotNil(t, replica)
			syncReplica(context.Background(), replica.LockedBy, replica)
			require.NoError(t, db.First(replica, id).Error)
			assert.Equal(t, model.AssetReplicaStatusFailed, replica.Status)
			assert.Zero(t, replica.NextSyncAt)
			assert.Contains(t, replica.LastError, "timed out")
			var asset model.MediaAsset
			require.NoError(t, db.First(&asset, replica.AssetID).Error)
			assert.Equal(t, model.AssetStatusSyncFailed, asset.Status)
			candidates, err := model.ListDueAssetReplicas(common.GetTimestamp()+3600, 7182, model.AssetReplicaWorkloadPoll, 5)
			require.NoError(t, err)
			assert.Empty(t, candidates)
			require.NoError(t, RetrySyncJob(id))
			require.NoError(t, db.First(replica, id).Error)
			assert.Equal(t, test.retryStatus, replica.Status)
			assert.Greater(t, replica.CreatedAt, common.GetTimestamp()-10)
			assert.Equal(t, test.upstreamID, replica.UpstreamAssetID)
			require.NoError(t, db.First(&asset, replica.AssetID).Error)
			assert.Equal(t, model.AssetStatusProcessing, asset.Status)
		})
	}
}

func TestAssetReviewProcessingDoesNotRestartInitialTimeout(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "review-processing-fixture")
	credential, err := encryptChannelCredential("fixture-secret")
	require.NoError(t, err)
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		call := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if call == 1 {
			_, _ = io.WriteString(w, `{"status":"Processing"}`)
			return
		}
		_, _ = io.WriteString(w, `{"status":"Active"}`)
	}))
	t.Cleanup(upstream.Close)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: 7391, Enabled: true,
		Protocol: model.AssetChannelProtocolYoufangREST, AuthType: model.AssetChannelAuthBearer,
		BaseURL: upstream.URL, EncryptedCredential: credential, QPM: 1000000}).Error)
	asset := model.MediaAsset{PublicID: "review-processing", StorageObjectID: 7391, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	now := common.GetTimestamp()
	replica := model.AssetReplica{AssetID: asset.ID, ChannelID: 7391, Operation: model.AssetReplicaOperationSync,
		Status: model.AssetReplicaStatusActive, UpstreamAssetID: "ready-before-review", Progress: 100, CreatedAt: now - 86400}
	require.NoError(t, db.Create(&replica).Error)
	for range 2 {
		candidates, err := model.ListDueAssetReplicas(now, 7391, model.AssetReplicaWorkloadReview, 5)
		require.NoError(t, err)
		require.Len(t, candidates, 1, "a previously active copy must stay in the bounded review workload")
		claimed, err := model.ClaimAssetReplica(candidates[0], now, now+120, "review-worker")
		require.NoError(t, err)
		require.True(t, claimed)
		syncReplica(context.Background(), "review-worker", candidates[0])
		require.NoError(t, db.First(&replica, replica.ID).Error)
		assert.Equal(t, 100, replica.Progress, "review processing must not reset the completed initial synchronization")
		assert.NotEqual(t, model.AssetReplicaStatusFailed, replica.Status)
		now = replica.NextSyncAt
	}
	assert.EqualValues(t, 2, calls.Load())
	assert.Equal(t, model.AssetReplicaStatusActive, replica.Status)
}

// runSyncPass processes currently due work round by round until none is left.
func runSyncPass(runnerID string) {
	scheduler := newAssetSyncScheduler(runnerID)
	for range 20 {
		before := scheduler.dispatched
		scheduler.dispatch(context.Background())
		scheduler.jobs.Wait()
		if scheduler.dispatched == before {
			return
		}
	}
}
