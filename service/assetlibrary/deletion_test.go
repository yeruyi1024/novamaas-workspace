package assetlibrary

import (
	"context"
	"errors"
	"fmt"
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

func TestAssetDeletionRetriesUpstreamFailureBeforeCleaningSource(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "asset-deletion-fixture")
	credential, err := encryptChannelCredential("fixture-secret")
	require.NoError(t, err)
	var attempts atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "upstream-delete-id")
		w.Header().Set("Content-Type", "application/json")
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, `{"error":{"code":"ServiceUnavailable","message":"request 404 failed; asset was not found in a temporary cache"}}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"code":"AssetNotFound","message":"asset not found"}}`)
	}))
	t.Cleanup(upstream.Close)
	require.NoError(t, db.Create(&model.AssetChannelConfig{
		ChannelID: 7191, Enabled: false, Protocol: model.AssetChannelProtocolYoufangREST,
		AuthType: model.AssetChannelAuthBearer, BaseURL: upstream.URL,
		EncryptedCredential: credential, QPM: 1000000,
	}).Error)
	source := model.StorageObject{ObjectID: "delete-source", Purpose: model.StorageObjectPurposeAssetLibrary, Status: model.StorageObjectStatusUploaded}
	require.NoError(t, db.Create(&source).Error)
	asset := model.MediaAsset{PublicID: "delete-material", OwnerUserID: 42, StorageObjectID: source.ID, Status: model.AssetStatusProcessing}
	require.NoError(t, db.Create(&asset).Error)
	replica := model.AssetReplica{AssetID: asset.ID, ChannelID: 7191, Status: model.AssetReplicaStatusProcessing,
		Operation: model.AssetReplicaOperationSync, UpstreamAssetID: "upstream-delete-id"}
	require.NoError(t, db.Create(&replica).Error)
	require.NoError(t, DeleteAsset(asset.PublicID, 42, false))
	now := common.GetTimestamp()
	for attempt := range 2 {
		candidates, err := model.ListDueAssetReplicas(now, 7191, model.AssetReplicaWorkloadMaintenance, 5)
		require.NoError(t, err)
		require.Len(t, candidates, 1)
		owner := fmt.Sprintf("delete-attempt-%d", attempt)
		claimed, err := model.ClaimAssetReplica(candidates[0], now, now+120, owner)
		require.NoError(t, err)
		require.True(t, claimed)
		syncReplica(context.Background(), owner, candidates[0])
		require.NoError(t, db.First(&replica, replica.ID).Error)
		require.NoError(t, db.First(&source, source.ID).Error)
		assert.Equal(t, "upstream-delete-id", replica.UpstreamAssetID)
		if attempt == 0 {
			assert.Equal(t, model.AssetReplicaStatusDeleting, replica.Status)
			assert.Equal(t, model.AssetReplicaOperationDelete, replica.Operation)
			assert.Equal(t, model.StorageObjectStatusUploaded, source.Status, "a failed remote delete must retain the source")
			assert.Greater(t, replica.NextSyncAt, now)
			now = replica.NextSyncAt
		} else {
			assert.Equal(t, model.AssetReplicaStatusDeleted, replica.Status)
			assert.Equal(t, model.StorageObjectStatusDeletePending, source.Status)
		}
	}
	assert.EqualValues(t, 2, attempts.Load(), "disabled channels must still finish upstream deletion")
}

func TestAssetDeletionRecognizesOnlyStructuredNotFound(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		want bool
	}{
		{"http404", &upstreamAssetError{statusCode: 404}, true},
		{"documented-code", fmt.Errorf("delete failed: %w", &upstreamAssetError{statusCode: 400, code: "InvalidAssetId.NotFound"}), true},
		{"server-error", &upstreamAssetError{statusCode: 500, code: "AssetNotFound", message: "request 404"}, false},
		{"authentication-error", &upstreamAssetError{statusCode: 403, code: "AssetNotFound"}, false},
		{"unstructured-error", errors.New("request 404: asset not found in cache"), false},
	} {
		t.Run(test.name, func(t *testing.T) { assert.Equal(t, test.want, isProviderNotFound(test.err)) })
	}
}
