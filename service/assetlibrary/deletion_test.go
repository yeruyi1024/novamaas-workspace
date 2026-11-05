package assetlibrary

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetDeletionKeepsLateIDsAndRetriesBeforeCleaningSource(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	var attempts atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.Header().Set("Content-Type", "application/json")
		code := http.StatusNotFound
		switch attempts.Add(1) {
		case 1:
			code = http.StatusInternalServerError
		case 2:
			code = http.StatusForbidden
		}
		w.WriteHeader(code)
		_, _ = io.WriteString(w, "{\"error\":{\"code\":\"AssetNotFound\",\"message\":\"request 404; asset not found in cache\"}}")
	}))
	t.Cleanup(upstream.Close)
	for _, channel := range []int{7191, 7192} {
		seedSyncChannel(t, channel, 1000000, model.AssetChannelProtocolYoufangREST, upstream.URL)
		require.NoError(t, db.Model(&model.AssetChannelConfig{}).Where("channel_id = ?", channel).Update("enabled", false).Error)
	}
	source := model.StorageObject{ObjectID: "delete-source", Purpose: model.StorageObjectPurposeAssetLibrary, Status: model.StorageObjectStatusUploaded}
	require.NoError(t, db.Create(&source).Error)
	asset := model.MediaAsset{PublicID: "delete-material", OwnerUserID: 42, StorageObjectID: source.ID, Status: model.AssetStatusProcessing}
	require.NoError(t, db.Create(&asset).Error)
	now := common.GetTimestamp()
	replicas := []model.AssetReplica{
		{AssetID: asset.ID, ChannelID: 7191, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusActive, UpstreamAssetID: "known-id"},
		{AssetID: asset.ID, ChannelID: 7192, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusSyncing, LockedBy: "creating", LeaseUntil: now + 120},
	}
	require.NoError(t, db.Create(&replicas).Error)
	require.NoError(t, DeleteAsset(asset.PublicID, 42, false))
	require.NoError(t, db.First(&source, source.ID).Error)
	assert.Equal(t, model.StorageObjectStatusUploaded, source.Status, "live creation still needs its source")
	require.NoError(t, model.FinishAssetReplicaSubmission(replicas[1].ID, "creating", "late-id", now+8))

	for step, channel := range []int{7191, 7191, 7191, 7192} {
		candidates, err := model.ListDueAssetReplicas(now, channel, model.AssetReplicaWorkloadMaintenance, 5)
		require.NoError(t, err)
		require.Len(t, candidates, 1)
		replica := candidates[0]
		claimed, err := model.ClaimAssetReplica(replica, now, now+120, "deleting")
		require.NoError(t, err)
		require.True(t, claimed)
		assert.Equal(t, map[int]string{7191: "known-id", 7192: "late-id"}[channel], replica.UpstreamAssetID)
		syncReplica(context.Background(), "deleting", replica)
		require.NoError(t, db.First(replica, replica.ID).Error)
		require.NoError(t, db.First(&source, source.ID).Error)
		if step < 2 {
			assert.Equal(t, model.AssetReplicaStatusDeleting, replica.Status, "500/403 must not be mistaken for NotFound")
			assert.NotEmpty(t, replica.UpstreamAssetID)
			assert.Greater(t, replica.NextSyncAt, now)
			now = replica.NextSyncAt
		} else {
			assert.Equal(t, model.AssetReplicaStatusDeleted, replica.Status)
		}
		expectedSource := model.StorageObjectStatusUploaded
		if step == 3 {
			expectedSource = model.StorageObjectStatusDeletePending
		}
		assert.Equal(t, expectedSource, source.Status, "source cleanup must wait for every channel")
	}
	claimed, err := model.ClaimAssetReplica(&replicas[1], now, now+120, "stale-candidate")
	require.NoError(t, err)
	assert.False(t, claimed, "a prefetched candidate must not overwrite completed deletion")
}
