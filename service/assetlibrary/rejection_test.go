package assetlibrary

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	coreService "github.com/QuantumNous/new-api/service"
	webhookService "github.com/QuantumNous/new-api/service/asyncwebhook"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelayedUpstreamRejectionUpdatesPollingStateAndReachesUnsignedWebhook(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	seedAssetWebhookOwner(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "delayed-rejection-end-to-end-key")
	fetchSetting := system_setting.GetFetchSetting()
	originalFetchSetting := *fetchSetting
	fetchSetting.EnableSSRFProtection = false
	coreService.InitHttpClient()
	t.Cleanup(func() { *fetchSetting = originalFetchSetting })

	type receivedWebhook struct {
		Body      []byte
		ID        string
		Timestamp string
		Signature string
	}
	received := make(chan receivedWebhook, 1)
	callback := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		received <- receivedWebhook{
			Body: body, ID: request.Header.Get("webhook-id"),
			Timestamp: request.Header.Get("webhook-timestamp"),
			Signature: request.Header.Get("webhook-signature"),
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer callback.Close()

	endpoint, err := webhookService.CreateWebhookEndpoint(42, webhookService.WebhookEndpointInput{
		Name: "customer", URL: "https://customer.example.com/assets",
		EventTypes: []string{WebhookEventAssetFailed},
	})
	require.NoError(t, err)
	require.NoError(t, db.Model(&model.AssetWebhookEndpoint{}).
		Where("public_id = ?", endpoint.ID).Update("url", callback.URL).Error)

	credential, err := encryptChannelCredential("provider-secret")
	require.NoError(t, err)
	upstreamActions := make(chan string, 1)
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		upstreamActions <- request.URL.Query().Get("Action")
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"Result":{"Status":"Failed","FailureReason":"sensitive_content"}}`)
	}))
	defer upstream.Close()

	group := model.AssetGroup{PublicID: "group-end-to-end", OwnerUserID: 42, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	asset := model.MediaAsset{
		PublicID: "asset-end-to-end", OwnerUserID: 42, GroupID: group.ID,
		StorageObjectID: 10001, AssetType: model.AssetTypeImage, Status: model.AssetStatusReady,
	}
	require.NoError(t, db.Create(&asset).Error)
	config := model.AssetChannelConfig{
		ChannelID: 901, Enabled: true, Protocol: model.AssetChannelProtocolVolcAction,
		AuthType: model.AssetChannelAuthBearer, BaseURL: upstream.URL,
		EncryptedCredential: credential, QPM: 1000000,
	}
	require.NoError(t, db.Create(&config).Error)
	require.NoError(t, db.Create(&model.AssetGroupReplica{
		GroupID: group.ID, ChannelID: config.ChannelID, UpstreamGroupID: "upstream-group-end-to-end",
		Status: model.AssetReplicaStatusActive,
	}).Error)
	require.NoError(t, db.Create(&model.AssetReplica{
		AssetID: asset.ID, ChannelID: config.ChannelID, UpstreamAssetID: "upstream-end-to-end",
		Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusActive,
	}).Error)

	runSyncPass("end-to-end-review")
	assert.Equal(t, "GetAsset", <-upstreamActions)

	require.NoError(t, db.First(&asset, asset.ID).Error)
	assert.Equal(t, model.AssetStatusUnavailable, asset.Status)
	assert.Equal(t, model.AssetUnavailableSensitiveContent, asset.UnavailableReason)
	polled, err := ListAssets(AssetListInput{OwnerUserID: 42, Statuses: []string{model.AssetStatusUnavailable}})
	require.NoError(t, err)
	require.Len(t, polled.Items, 1)
	assert.Equal(t, asset.PublicID, polled.Items[0].ID)
	assert.Equal(t, model.AssetUnavailableSensitiveContent, polled.Items[0].UnavailableReason)

	coreService.RunWebhookDeliveryPass("end-to-end-webhook")

	var notification receivedWebhook
	select {
	case notification = <-received:
	default:
		require.FailNow(t, "customer callback did not receive the failed asset event")
	}
	var event assetWebhookEvent
	require.NoError(t, common.Unmarshal(notification.Body, &event))
	assert.Equal(t, WebhookEventAssetFailed, event.Type)
	assert.Equal(t, asset.PublicID, event.Data.ID)
	assert.Equal(t, "Failed", event.Data.Status)
	assert.Equal(t, model.AssetUnavailableSensitiveContent, event.Data.FailureReason)
	assert.Equal(t, 2, event.Data.StateVersion)
	assert.Empty(t, notification.Signature)

	var delivery model.AssetWebhookDelivery
	require.NoError(t, db.Where("asset_id = ?", asset.ID).First(&delivery).Error)
	assert.Equal(t, model.AssetWebhookDeliveryStatusSucceeded, delivery.Status)
	assert.Equal(t, http.StatusNoContent, delivery.ResponseStatus)
}

func TestVolcDelayedFailureReasonUpdatesAssetsAndQueuesNotifications(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	seedAssetWebhookOwner(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "customer-rejection-regression")
	credential, err := encryptChannelCredential("provider-secret")
	require.NoError(t, err)
	_, err = webhookService.CreateWebhookEndpoint(42, webhookService.WebhookEndpointInput{Name: "customer", URL: "https://customer.example.com/assets", EventTypes: []string{WebhookEventAssetFailed}})
	require.NoError(t, err)
	group := model.AssetGroup{PublicID: "group-customer", OwnerUserID: 42, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	observed := make(chan int64, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("Action") != "GetAsset" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var leased int64
		if err := db.Model(&model.AssetReplica{}).Where("locked_by <> ''").Count(&leased).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		observed <- leased
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"Result":{"Status":"Failed","FailureReason":"sensitive_content"}}`)
	}))
	defer server.Close()
	config := model.AssetChannelConfig{ChannelID: 91, Enabled: true, Protocol: model.AssetChannelProtocolVolcAction,
		AuthType: model.AssetChannelAuthBearer, BaseURL: server.URL, EncryptedCredential: credential, QPM: 1000000}
	require.NoError(t, db.Create(&config).Error)
	require.NoError(t, db.Create(&model.AssetGroupReplica{GroupID: group.ID, ChannelID: 91, UpstreamGroupID: "upstream-group", Status: model.AssetReplicaStatusActive}).Error)
	for i, id := range []string{"asset-20260923163433-giwbv", "asset-20260923163435-bwgx2"} {
		asset := model.MediaAsset{PublicID: id, OwnerUserID: 42, GroupID: group.ID, StorageObjectID: int64(9900 + i), Status: model.AssetStatusReady}
		require.NoError(t, db.Create(&asset).Error)
		require.NoError(t, db.Create(&model.AssetReplica{AssetID: asset.ID, ChannelID: 91, UpstreamAssetID: id, Status: model.AssetReplicaStatusActive, Operation: model.AssetReplicaOperationSync}).Error)
	}
	runSyncPass("customer-review")
	require.Len(t, observed, 2)
	assert.EqualValues(t, 1, <-observed)
	assert.EqualValues(t, 1, <-observed)
	var assets []model.MediaAsset
	require.NoError(t, db.Find(&assets).Error)
	require.Len(t, assets, 2)
	for _, asset := range assets {
		assert.Equal(t, model.AssetStatusUnavailable, asset.Status)
		assert.Equal(t, model.AssetUnavailableSensitiveContent, asset.UnavailableReason)
		var delivery model.AssetWebhookDelivery
		require.NoError(t, db.Where("asset_id = ?", asset.ID).First(&delivery).Error)
		var event assetWebhookEvent
		require.NoError(t, common.UnmarshalJsonStr(delivery.Payload, &event))
		assert.Equal(t, WebhookEventAssetFailed, event.Type)
		assert.Equal(t, asset.PublicID, event.Data.ID)
		assert.Equal(t, "sensitive_content", event.Data.FailureReason)
		assert.Equal(t, 2, event.Data.StateVersion)
	}
}

func TestContentRejectionClassificationSeparatesPolicyFromTransientFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "real person in upstream code", err: &upstreamAssetError{statusCode: 400, code: "RealPersonDetected"}, want: model.AssetUnavailableRealPerson},
		{name: "snake case sensitive content", err: &upstreamAssetError{statusCode: 400, code: "sensitive_content"}, want: model.AssetUnavailableSensitiveContent},
		{name: "snake case real person", err: &upstreamAssetError{statusCode: 400, code: "real_person"}, want: model.AssetUnavailableRealPerson},
		{name: "snake case policy rejection", err: &upstreamAssetError{statusCode: 400, code: "policy_rejected"}, want: model.AssetUnavailablePolicyRejected},
		{name: "snake case service outage", err: &upstreamAssetError{statusCode: 400, code: "sensitive_content", message: "service_unavailable"}},
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

	syncReplica(context.Background(), "worker-1", &replica)
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
	syncReplica(context.Background(), "worker-2", &queuedReplica)
	require.NoError(t, db.First(&queuedReplica, queuedReplica.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusRejected, queuedReplica.Status)
	assert.Equal(t, model.AssetUnavailableSensitiveContent, queuedReplica.LastError)
	assert.Empty(t, queuedReplica.LockedBy)
	otherReplica := model.AssetReplica{AssetID: asset.ID, ChannelID: 11, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusFailed}
	require.NoError(t, db.Create(&otherReplica).Error)
	assert.Error(t, RetrySyncJob(otherReplica.ID))
}
