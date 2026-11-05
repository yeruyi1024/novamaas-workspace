package assetlibrary

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedSyncReplicas(t *testing.T, channelID int, replicas ...model.AssetReplica) []model.AssetReplica {
	t.Helper()
	group := model.AssetGroup{PublicID: fmt.Sprintf("group-%d", channelID), Status: model.AssetStatusReady}
	require.NoError(t, model.DB.Create(&group).Error)
	for index := range replicas {
		asset := model.MediaAsset{PublicID: fmt.Sprintf("asset-%d-%d", channelID, index), GroupID: group.ID,
			StorageObjectID: int64(index + 1), AssetType: model.AssetTypeImage, Status: model.AssetStatusProcessing}
		require.NoError(t, model.DB.Create(&asset).Error)
		replicas[index].AssetID, replicas[index].ChannelID = asset.ID, channelID
		replicas[index].Operation = model.AssetReplicaOperationSync
	}
	require.NoError(t, model.DB.Create(&replicas).Error)
	return replicas
}

func seedSyncChannel(t *testing.T, channelID, qpm int, protocol, endpoint string) {
	t.Helper()
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "scheduler-fixture")
	credential, err := encryptChannelCredential("fixture-secret")
	require.NoError(t, err)
	require.NoError(t, model.DB.Create(&model.AssetChannelConfig{ChannelID: channelID, Enabled: true, QPM: qpm,
		Protocol: protocol, AuthType: model.AssetChannelAuthBearer, BaseURL: endpoint, EncryptedCredential: credential}).Error)
	t.Cleanup(func() {
		channelRateState.Lock()
		defer channelRateState.Unlock()
		delete(channelRateState.next, channelID)
		delete(channelRateState.sentNext, channelID)
		delete(channelRateState.reviewNext, channelID)
	})
}

func TestAssetSyncQueriesWhileSubmissionIsBlocked(t *testing.T) {
	for _, pollChannel := range []int{701, 702} {
		t.Run(fmt.Sprint(pollChannel), func(t *testing.T) {
			setupAssetLibraryTestDB(t)
			submitted, queried := make(chan struct{}, 1), make(chan struct{}, 1)
			release := make(chan struct{})
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Query().Get("Action") == "CreateAssetGroup" {
					submitted <- struct{}{}
					<-release
					return
				}
				queried <- struct{}{}
				_, _ = io.WriteString(w, "{\"Result\":{\"Status\":\"Processing\"}}")
			}))
			t.Cleanup(upstream.Close)
			t.Cleanup(func() { close(release) })
			qpm := 1000000
			if pollChannel != 701 {
				qpm = 1 // One channel's minute-long quota wait must not block the other.
				seedSyncChannel(t, pollChannel, 1000000, model.AssetChannelProtocolVolcAction, upstream.URL)
			}
			seedSyncChannel(t, 701, qpm, model.AssetChannelProtocolVolcAction, upstream.URL)
			seedSyncReplicas(t, 701, model.AssetReplica{Status: model.AssetReplicaStatusPending}, model.AssetReplica{Status: model.AssetReplicaStatusPending})
			asset := model.MediaAsset{PublicID: "waiting-poll", StorageObjectID: 3, Status: model.AssetStatusProcessing}
			require.NoError(t, model.DB.Create(&asset).Error)
			require.NoError(t, model.DB.Create(&model.AssetReplica{AssetID: asset.ID, ChannelID: pollChannel,
				Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusProcessing, UpstreamAssetID: "already-uploaded"}).Error)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			done := make(chan struct{})
			go func() {
				defer close(done)
				newAssetSyncScheduler("blocked-submission").run(ctx)
			}()
			defer func() {
				cancel()
				<-done
			}()
			select {
			case <-submitted:
			case <-ctx.Done():
				t.Fatal("submission did not reach upstream")
			}
			select {
			case <-queried:
			case <-ctx.Done():
				t.Fatal("readiness query was blocked by submission or another channel's quota")
			}
			var leasedSubmissions int64
			require.NoError(t, model.DB.Model(&model.AssetReplica{}).Where("channel_id = ? AND upstream_asset_id = ? AND locked_by <> ?", 701, "", "").Count(&leasedSubmissions).Error)
			assert.EqualValues(t, 1, leasedSubmissions, "other submissions must stay unleased while the request is blocked")
		})
	}
}

func TestAssetSyncTimeoutStopsUntilExplicitRetry(t *testing.T) {
	for _, upstreamID := range []string{"", "known-id"} {
		t.Run("id="+upstreamID, func(t *testing.T) {
			db := setupAssetLibraryTestDB(t)
			replica := seedSyncReplicas(t, 7182, model.AssetReplica{Status: model.AssetReplicaStatusProcessing,
				UpstreamAssetID: upstreamID, CreatedAt: common.GetTimestamp() - 301})[0]
			now := common.GetTimestamp()
			claimed, err := model.ClaimAssetReplica(&replica, now, now+120, "timeout")
			require.NoError(t, err)
			require.True(t, claimed)
			syncReplica(context.Background(), "timeout", &replica)
			require.NoError(t, db.First(&replica, replica.ID).Error)
			assert.Equal(t, model.AssetReplicaStatusFailed, replica.Status)
			assert.Zero(t, replica.NextSyncAt)
			var asset model.MediaAsset
			require.NoError(t, db.First(&asset, replica.AssetID).Error)
			assert.Equal(t, model.AssetStatusSyncFailed, asset.Status)
			require.NoError(t, RetrySyncJob(replica.ID))
			require.NoError(t, db.First(&replica, replica.ID).Error)
			assert.Equal(t, upstreamID, replica.UpstreamAssetID)
			assert.GreaterOrEqual(t, replica.CreatedAt, now)
			require.NoError(t, db.First(&asset, replica.AssetID).Error)
			assert.Equal(t, model.AssetStatusProcessing, asset.Status)
		})
	}
}

func TestAssetReviewProcessingStaysOutsideInitialTimeout(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "{\"status\":\"Processing\"}")
	}))
	t.Cleanup(upstream.Close)
	seedSyncChannel(t, 7391, 1000000, model.AssetChannelProtocolYoufangREST, upstream.URL)
	replica := seedSyncReplicas(t, 7391, model.AssetReplica{Status: model.AssetReplicaStatusActive,
		UpstreamAssetID: "review-id", Progress: 100, CreatedAt: common.GetTimestamp() - 86400})[0]
	runSyncPass("review")
	require.NoError(t, db.First(&replica, replica.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusProcessing, replica.Status)
	candidates, err := model.ListDueAssetReplicas(replica.NextSyncAt, 7391, model.AssetReplicaWorkloadReview, 5)
	require.NoError(t, err)
	require.Len(t, candidates, 1, "a completed upload must stay in the bounded review workload")
	assert.Equal(t, replica.ID, candidates[0].ID)
}

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
