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
	storageService "github.com/QuantumNous/new-api/service/storage"
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

func TestPendingAssetSyncStartsTimeoutClockWhenClaimed(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "scheduler-fixture")
	require.NoError(t, db.AutoMigrate(&model.StorageProfile{}, &model.StorageCredential{}))
	profile, err := storageService.SaveProfile(0, storageService.ProfileInput{
		Name: "queue-wait-oss", ProviderType: model.StorageProviderAliyunOSS, Status: model.StorageProfileStatusEnabled,
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", Region: "cn-hangzhou", Bucket: "test-assets-bucket",
		AuthType: model.StorageCredentialAuthStatic, AccessKeyID: "fixture-id", AccessKeySecret: "fixture-secret",
	})
	require.NoError(t, err)
	storageObject := model.StorageObject{
		ObjectID: "queued-obj", StorageProfileID: profile.ID, ObjectKey: "assets/queued.png",
		Purpose: model.StorageObjectPurposeAssetLibrary, Status: model.StorageObjectStatusUploaded,
	}
	require.NoError(t, db.Create(&storageObject).Error)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("Action") {
		case "CreateAssetGroup":
			_, _ = io.WriteString(w, `{"Result":{"Id":"queued-group"}}`)
		case "CreateAsset":
			_, _ = io.WriteString(w, `{"Result":{"Id":"queued-upstream-id"}}`)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(upstream.Close)
	seedSyncChannel(t, 7183, 1000000, model.AssetChannelProtocolVolcAction, upstream.URL)

	now := common.GetTimestamp()
	replicas := seedSyncReplicas(t, 7183,
		model.AssetReplica{Status: model.AssetReplicaStatusPending, CreatedAt: now - 600, NextSyncAt: now - 600},
		model.AssetReplica{Status: model.AssetReplicaStatusFailed, Attempts: 1, CreatedAt: now - 301, NextSyncAt: now},
	)
	require.NoError(t, db.Model(&model.MediaAsset{}).Where("id = ?", replicas[0].AssetID).Update("storage_object_id", storageObject.ID).Error)

	queued := replicas[0]
	claimed, err := model.ClaimAssetReplica(&queued, now, now+120, "queue-submit")
	require.NoError(t, err)
	require.True(t, claimed)
	syncReplica(context.Background(), "queue-submit", &queued)
	require.NoError(t, db.First(&queued, queued.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusProcessing, queued.Status)
	assert.Equal(t, "queued-upstream-id", queued.UpstreamAssetID)
	assert.Empty(t, queued.LastError)
	assert.GreaterOrEqual(t, queued.CreatedAt, now)

	retriedFailure := replicas[1]
	claimed, err = model.ClaimAssetReplica(&retriedFailure, now, now+120, "retry-timeout")
	require.NoError(t, err)
	require.True(t, claimed)
	syncReplica(context.Background(), "retry-timeout", &retriedFailure)
	require.NoError(t, db.First(&retriedFailure, retriedFailure.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusFailed, retriedFailure.Status)
	assert.Zero(t, retriedFailure.NextSyncAt)
	assert.Equal(t, "asset synchronization timed out after 5 minutes", retriedFailure.LastError)
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
			// A review just sent spaces the next one on its channel by ten
			// request slots. With the tiny slots these fixtures use, wait that
			// out instead of ending the pass early; real spacing ends it.
			spacing := time.Duration(0)
			channelRateState.Lock()
			for _, next := range channelRateState.reviewNext {
				spacing = max(spacing, time.Until(next))
			}
			channelRateState.Unlock()
			if spacing <= 0 || spacing > 50*time.Millisecond {
				return
			}
			time.Sleep(spacing)
		}
	}
}
