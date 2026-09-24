package assetlibrary

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentRejectionClassificationSeparatesPolicyFromTransientFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "real person in upstream code", err: &upstreamAssetError{statusCode: 400, code: "RealPersonDetected"}, want: model.AssetUnavailableRealPerson},
		{name: "sensitive content in upstream message", err: &upstreamAssetError{statusCode: 422, message: "素材包含敏感信息"}, want: model.AssetUnavailableSensitiveContent},
		{name: "server error remains retryable", err: &upstreamAssetError{statusCode: 503, code: "SensitiveContent"}},
		{name: "authentication error remains retryable", err: &upstreamAssetError{statusCode: 403, code: "AccessDenied", message: "permission denied"}},
		{name: "explicit policy error on 403 is terminal", err: &upstreamAssetError{statusCode: 403, code: "SensitiveContent"}, want: model.AssetUnavailableSensitiveContent},
		{name: "network error remains retryable", err: errors.New("sensitive content in transport error")},
		{name: "moderation service outage remains retryable", err: &upstreamAssetError{statusCode: 400, message: "真人识别服务异常，请稍后重试"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, upstreamContentRejectionReason(test.err))
		})
	}
	assert.Equal(t, model.AssetUnavailablePolicyRejected, processingContentRejectionReason(providerResult{Status: "rejected"}))
	assert.Empty(t, processingContentRejectionReason(providerResult{Status: "failed", Message: "temporary processing error"}))
}

func TestRejectedUpstreamAssetBecomesUnavailableAndVideoMappingReturnsActionableError(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "asset-rejection-integration-test-key")
	credential, err := encryptChannelCredential("provider-secret")
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/api/v1/assets/upstream-123", request.URL.Path)
		assert.Equal(t, http.MethodGet, request.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"id":"upstream-123","status":"rejected","error":{"code":"SensitiveContent","message":"processing failed"}}}`)
	}))
	defer server.Close()
	group := model.AssetGroup{PublicID: "group-local", OwnerUserID: 42, Name: "Campaign", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	asset := model.MediaAsset{PublicID: "asset-20260922120000-local", OwnerUserID: 42, GroupID: group.ID, StorageObjectID: 1001, Name: "portrait", AssetType: model.AssetTypeImage, ContentType: "image/png", Size: 10, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	config := model.AssetChannelConfig{ChannelID: 9, Enabled: true, Protocol: model.AssetChannelProtocolYoufangREST, AuthType: model.AssetChannelAuthBearer, BaseURL: server.URL, EncryptedCredential: credential, QPM: 10000}
	require.NoError(t, db.Create(&config).Error)
	replica := model.AssetReplica{AssetID: asset.ID, ChannelID: 9, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusSyncing, UpstreamAssetID: "upstream-123", LockedBy: "worker-1", Progress: 50}
	require.NoError(t, db.Create(&replica).Error)

	syncReplica("worker-1", &replica)
	require.NoError(t, db.First(&asset, asset.ID).Error)
	require.NoError(t, db.First(&replica, replica.ID).Error)
	assert.Equal(t, model.AssetStatusUnavailable, asset.Status)
	assert.Equal(t, model.AssetUnavailableSensitiveContent, asset.UnavailableReason)
	assert.Equal(t, model.AssetReplicaStatusRejected, replica.Status)
	assert.Zero(t, replica.NextSyncAt)
	assert.Empty(t, replica.LockedBy)
	assert.Equal(t, model.AssetUnavailableSensitiveContent, replica.LastError)

	requestBody := []byte(`{"content":[{"image_url":{"url":"asset://asset-20260922120000-local"}}]}`)
	_, count, err := ResolveRequestAssetIDs(requestBody, 42, 9)
	require.Error(t, err)
	assert.Zero(t, count)
	var requestErr *RequestError
	require.ErrorAs(t, err, &requestErr)
	assert.Equal(t, http.StatusUnprocessableEntity, requestErr.HTTPStatusCode())
	assert.Equal(t, "asset_unavailable", requestErr.ErrorCode())
	assert.True(t, strings.Contains(err.Error(), "upload revised material"))
	assert.NotContains(t, err.Error(), "processing failed")

	views := assetViews([]model.MediaAsset{asset}, map[int64]string{group.ID: group.PublicID})
	require.Len(t, views, 1)
	assert.Equal(t, model.AssetUnavailableSensitiveContent, views[0].UnavailableReason)
	activePage, err := ListAssets(AssetListInput{OwnerUserID: 42, Statuses: []string{model.AssetStatusReady}})
	require.NoError(t, err)
	assert.Zero(t, activePage.Total)
	failedPage, err := ListAssets(AssetListInput{OwnerUserID: 42, Statuses: []string{model.AssetStatusUnavailable}})
	require.NoError(t, err)
	assert.EqualValues(t, 1, failedPage.Total)
	assert.Error(t, RetrySyncJob(replica.ID))
	queuedReplica := model.AssetReplica{AssetID: asset.ID, ChannelID: 10, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusSyncing, LockedBy: "worker-2"}
	require.NoError(t, db.Create(&queuedReplica).Error)
	syncReplica("worker-2", &queuedReplica)
	require.NoError(t, db.First(&queuedReplica, queuedReplica.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusRejected, queuedReplica.Status)
	assert.Equal(t, model.AssetUnavailableSensitiveContent, queuedReplica.LastError)
	assert.Empty(t, queuedReplica.LockedBy)
	otherReplica := model.AssetReplica{AssetID: asset.ID, ChannelID: 11, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusFailed}
	require.NoError(t, db.Create(&otherReplica).Error)
	assert.Error(t, RetrySyncJob(otherReplica.ID))
}
