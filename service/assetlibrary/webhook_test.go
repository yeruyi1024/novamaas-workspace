package assetlibrary

import (
	"fmt"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	webhookService "github.com/QuantumNous/new-api/service/asyncwebhook"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type legacyAssetWebhookEndpoint struct {
	ID                   int64 `gorm:"primaryKey"`
	EncryptedSecret      string
	SecretHint           string
	CredentialKeyVersion string
}

func (legacyAssetWebhookEndpoint) TableName() string {
	return "asset_webhook_endpoints"
}

func seedAssetWebhookOwner(t *testing.T) {
	t.Helper()
	require.NoError(t, model.DB.Create(&model.User{Id: 42, Username: "asset-webhook-owner", Password: "password"}).Error)
	require.NoError(t, model.DB.Create(&model.AssetWebhookEndpoint{PublicID: model.SystemAssetWebhookEndpointID, OwnerUserID: 0, Status: model.AssetWebhookEndpointStatusEnabled, EventTypes: `["asset.active","asset.failed"]`}).Error)
}

func TestConcurrentWebhookCreationCannotExceedEndpointLimit(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	seedAssetWebhookOwner(t)
	for index := 0; index < webhookService.MaxWebhookEndpointsPerUser-1; index++ {
		endpoint := model.AssetWebhookEndpoint{
			PublicID: fmt.Sprintf("we_existing_%d", index), OwnerUserID: 42,
			Name: "existing", URL: "https://customer.example.com/assets",
			EventTypes: `[]`, Status: model.AssetWebhookEndpointStatusEnabled,
		}
		require.NoError(t, db.Create(&endpoint).Error)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for index := 0; index < 2; index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			_, err := webhookService.CreateWebhookEndpoint(42, webhookService.WebhookEndpointInput{
				Name: "concurrent", URL: "https://customer.example.com/assets",
				EventTypes: []string{WebhookEventAssetFailed},
			})
			results <- err
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	created := 0
	conflicts := 0
	for err := range results {
		if err == nil {
			created++
			continue
		}
		var requestErr *webhookService.RequestError
		require.ErrorAs(t, err, &requestErr)
		assert.Equal(t, 409, requestErr.HTTPStatusCode())
		conflicts++
	}
	assert.Equal(t, 1, created)
	assert.Equal(t, 1, conflicts)

	var count int64
	require.NoError(t, db.Model(&model.AssetWebhookEndpoint{}).
		Where("owner_user_id = ? AND status = ?", 42, model.AssetWebhookEndpointStatusEnabled).
		Count(&count).Error)
	assert.EqualValues(t, webhookService.MaxWebhookEndpointsPerUser, count)
}

func TestCreateWebhookEndpointWithoutEncryptionConfiguration(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	seedAssetWebhookOwner(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "")
	t.Setenv("CRYPTO_SECRET", "")
	t.Setenv("SESSION_SECRET", "")

	created, err := webhookService.CreateWebhookEndpoint(42, webhookService.WebhookEndpointInput{
		Name:       "production",
		URL:        "https://customer.example.com/webhooks/assets",
		EventTypes: []string{WebhookEventAssetActive, WebhookEventAssetFailed},
	})
	require.NoError(t, err)
	payload, err := common.Marshal(created)
	require.NoError(t, err)
	assert.NotContains(t, string(payload), "signing_secret")

	var stored model.AssetWebhookEndpoint
	require.NoError(t, db.Where("public_id = ?", created.ID).First(&stored).Error)
	assert.Equal(t, model.AssetWebhookEndpointStatusEnabled, stored.Status)
	for _, column := range []string{"encrypted_secret", "secret_hint", "credential_key_version"} {
		assert.False(t, db.Migrator().HasColumn(&model.AssetWebhookEndpoint{}, column), column)
	}

	listed, err := webhookService.ListWebhookEndpoints(42)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, created.ID, listed[0].ID)
	assert.Equal(t, []string{WebhookEventAssetActive, WebhookEventAssetFailed}, listed[0].EventTypes)
}

func TestWebhookEndpointKeepsLegacyCredentialColumnsWithoutUsingThem(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	seedAssetWebhookOwner(t)
	require.NoError(t, db.AutoMigrate(&legacyAssetWebhookEndpoint{}))
	for _, column := range []string{"encrypted_secret", "secret_hint", "credential_key_version"} {
		require.True(t, db.Migrator().HasColumn(&legacyAssetWebhookEndpoint{}, column), column)
	}
	require.NoError(t, db.AutoMigrate(&model.AssetWebhookEndpoint{}))

	created, err := webhookService.CreateWebhookEndpoint(42, webhookService.WebhookEndpointInput{
		Name: "legacy database", URL: "https://customer.example.com/webhooks/assets",
		EventTypes: []string{WebhookEventAssetFailed},
	})
	require.NoError(t, err)
	var stored legacyAssetWebhookEndpoint
	require.NoError(t, db.Where("public_id = ?", created.ID).First(&stored).Error)
	assert.Empty(t, stored.EncryptedSecret)
	assert.Empty(t, stored.SecretHint)
	assert.Empty(t, stored.CredentialKeyVersion)
}

func TestInitialActivationQueuesWebhookAndPromotesAsset(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	seedAssetWebhookOwner(t)

	_, err := webhookService.CreateWebhookEndpoint(42, webhookService.WebhookEndpointInput{
		Name: "production", URL: "https://customer.example.com/webhooks/assets",
		EventTypes: []string{WebhookEventAssetActive},
	})
	require.NoError(t, err)
	group := model.AssetGroup{PublicID: "group-activation", OwnerUserID: 42, Name: "Campaign", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	asset := model.MediaAsset{PublicID: "asset-processing", OwnerUserID: 42, GroupID: group.ID, StorageObjectID: 7002, Name: "portrait", AssetType: model.AssetTypeImage, ContentType: "image/png", Size: 10, Status: model.AssetStatusProcessing}
	require.NoError(t, db.Create(&asset).Error)
	replica := model.AssetReplica{AssetID: asset.ID, ChannelID: 9, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusSyncing, UpstreamAssetID: "upstream-456", LockedBy: "worker-1", Progress: 90, CreatedAt: common.GetTimestamp()}
	require.NoError(t, db.Create(&replica).Error)

	require.NoError(t, activateReplicaAndQueueWebhook("worker-1", &replica))
	require.NoError(t, db.First(&asset, asset.ID).Error)
	assert.Equal(t, model.AssetStatusReady, asset.Status)
	require.NoError(t, db.First(&replica, replica.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusActive, replica.Status)
	assert.Greater(t, replica.NextSyncAt, common.GetTimestamp())

	var delivery model.AssetWebhookDelivery
	require.NoError(t, db.First(&delivery).Error)
	assert.Equal(t, WebhookEventAssetActive, delivery.EventType)
	assert.Contains(t, delivery.Payload, `"status":"Active"`)
	assert.Contains(t, delivery.Payload, `"state_version":1`)
}

func TestConfirmedDelayedRejectionQueuesWebhookAfterLocalStateChange(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	seedAssetWebhookOwner(t)

	endpoint, err := webhookService.CreateWebhookEndpoint(42, webhookService.WebhookEndpointInput{
		Name:       "production",
		URL:        "https://customer.example.com/webhooks/assets",
		EventTypes: []string{WebhookEventAssetFailed},
	})
	require.NoError(t, err)
	group := model.AssetGroup{PublicID: "group-webhook", OwnerUserID: 42, Name: "Campaign", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	asset := model.MediaAsset{PublicID: "asset-delayed-rejection", OwnerUserID: 42, GroupID: group.ID, StorageObjectID: 7001, Name: "portrait", AssetType: model.AssetTypeImage, ContentType: "image/png", Size: 10, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	replica := model.AssetReplica{AssetID: asset.ID, ChannelID: 9, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusSyncing, UpstreamAssetID: "upstream-123", LockedBy: "worker-1", Progress: 100}
	require.NoError(t, db.Create(&replica).Error)

	require.NoError(t, rejectReplicaAndQueueWebhook("worker-1", &replica, model.AssetUnavailableSensitiveContent))
	require.NoError(t, db.First(&asset, asset.ID).Error)
	assert.Equal(t, model.AssetStatusUnavailable, asset.Status)
	assert.Equal(t, model.AssetUnavailableSensitiveContent, asset.UnavailableReason)

	var delivery model.AssetWebhookDelivery
	require.NoError(t, db.First(&delivery).Error)
	assert.Equal(t, endpoint.ID, delivery.EndpointPublicID)
	assert.Equal(t, WebhookEventAssetFailed, delivery.EventType)
	assert.Equal(t, model.AssetWebhookDeliveryStatusPending, delivery.Status)
	assert.Contains(t, delivery.Payload, `"id":"asset-delayed-rejection"`)
	assert.Contains(t, delivery.Payload, `"failure_reason":"sensitive_content"`)
	assert.Contains(t, delivery.Payload, `"state_version":2`)
	assert.NotZero(t, delivery.NextAttemptAt)
}

func TestDelayedRejectionSupersedesUndeliveredActiveWebhook(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	seedAssetWebhookOwner(t)

	_, err := webhookService.CreateWebhookEndpoint(42, webhookService.WebhookEndpointInput{
		Name: "production", URL: "https://customer.example.com/webhooks/assets",
		EventTypes: []string{WebhookEventAssetActive, WebhookEventAssetFailed},
	})
	require.NoError(t, err)
	group := model.AssetGroup{PublicID: "group-ordering", OwnerUserID: 42, Name: "Campaign", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	asset := model.MediaAsset{
		PublicID: "asset-ordering", OwnerUserID: 42, GroupID: group.ID, StorageObjectID: 7003,
		Name: "portrait", AssetType: model.AssetTypeImage, ContentType: "image/png", Size: 10,
		Status: model.AssetStatusReady,
	}
	require.NoError(t, db.Create(&asset).Error)
	require.NoError(t, queueAssetWebhookDeliveries(db, &asset, WebhookEventAssetActive))
	replica := model.AssetReplica{
		AssetID: asset.ID, ChannelID: 9, Operation: model.AssetReplicaOperationSync,
		Status: model.AssetReplicaStatusSyncing, UpstreamAssetID: "upstream-ordering",
		LockedBy: "worker-ordering", Progress: 100,
	}
	require.NoError(t, db.Create(&replica).Error)

	require.NoError(t, rejectReplicaAndQueueWebhook("worker-ordering", &replica, model.AssetUnavailableSensitiveContent))

	var activeDelivery model.AssetWebhookDelivery
	require.NoError(t, db.Where("event_type = ?", WebhookEventAssetActive).First(&activeDelivery).Error)
	assert.Equal(t, model.AssetWebhookDeliveryStatusSuperseded, activeDelivery.Status)
	assert.Equal(t, int64(0), activeDelivery.NextAttemptAt)

	var failedDelivery model.AssetWebhookDelivery
	require.NoError(t, db.Where("event_type = ?", WebhookEventAssetFailed).First(&failedDelivery).Error)
	assert.Equal(t, model.AssetWebhookDeliveryStatusPending, failedDelivery.Status)
	var activeEvent, failedEvent assetWebhookEvent
	require.NoError(t, common.UnmarshalJsonStr(activeDelivery.Payload, &activeEvent))
	require.NoError(t, common.UnmarshalJsonStr(failedDelivery.Payload, &failedEvent))
	assert.Equal(t, activeEvent.Data.ID, failedEvent.Data.ID)
	assert.Greater(t, failedEvent.Data.StateVersion, activeEvent.Data.StateVersion,
		"receivers must be able to reject late Active events independently of timestamp resolution")
}

func TestActivationCannotReviveTerminalAsset(t *testing.T) {
	for _, status := range []string{model.AssetStatusUnavailable, model.AssetStatusDeleted} {
		t.Run(status, func(t *testing.T) {
			db := setupAssetLibraryTestDB(t)
			asset := model.MediaAsset{PublicID: "asset-terminal", StorageObjectID: 8001, Status: status}
			require.NoError(t, db.Create(&asset).Error)
			replica := model.AssetReplica{AssetID: asset.ID, ChannelID: 1, Operation: model.AssetReplicaOperationSync,
				Status: model.AssetReplicaStatusSyncing, UpstreamAssetID: "late-active-copy", LockedBy: "activation-worker"}
			require.NoError(t, db.Create(&replica).Error)
			err := activateReplicaAndQueueWebhook("activation-worker", &replica)
			if status == model.AssetStatusDeleted {
				require.NoError(t, err, "a late Active result must continue deletion")
			} else {
				require.Error(t, err)
			}
			require.NoError(t, db.First(&asset, asset.ID).Error)
			assert.Equal(t, status, asset.Status)
			require.NoError(t, db.First(&replica, replica.ID).Error)
			if status == model.AssetStatusDeleted {
				assert.Equal(t, model.AssetReplicaStatusDeleting, replica.Status)
				assert.Equal(t, model.AssetReplicaOperationDelete, replica.Operation)
				assert.Equal(t, "late-active-copy", replica.UpstreamAssetID)
				assert.Empty(t, replica.LockedBy)
			} else {
				assert.Equal(t, model.AssetReplicaStatusSyncing, replica.Status)
			}
			var count int64
			require.NoError(t, db.Model(&model.AssetWebhookDelivery{}).Count(&count).Error)
			assert.Zero(t, count)
		})
	}
}

func TestActiveReplicaRemainsScheduledForDelayedReview(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	now := common.GetTimestamp()
	replica := model.AssetReplica{
		AssetID: 1, ChannelID: 9, Operation: model.AssetReplicaOperationSync,
		Status: model.AssetReplicaStatusActive, NextSyncAt: now - 1,
	}
	require.NoError(t, db.Create(&replica).Error)

	claimed, err := model.ClaimAssetReplicas(now, now+60, "review-worker", 10)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	assert.Equal(t, replica.ID, claimed[0].ID)
	assert.Equal(t, model.AssetReplicaStatusActive, claimed[0].Status)
	assert.Equal(t, model.AssetReplicaStatusActive, claimed[0].ClaimedFromStatus)

	var stored model.AssetReplica
	require.NoError(t, db.First(&stored, replica.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusActive, stored.Status)
	assert.Equal(t, "review-worker", stored.LockedBy)

	claimedAgain, err := model.ClaimAssetReplicas(now, now+60, "other-worker", 10)
	require.NoError(t, err)
	assert.Empty(t, claimedAgain)
}
