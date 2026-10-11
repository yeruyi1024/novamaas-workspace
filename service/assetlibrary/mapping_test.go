package assetlibrary

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

func setupAssetLibraryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	previous := model.DB
	previousDatabaseType := common.MainDatabaseType()
	common.SetMainDatabaseType(common.DatabaseTypeSQLite)
	dsn := fmt.Sprintf("file:asset-library-%s?mode=memory&cache=shared&_busy_timeout=30000", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.User{},
		&model.Channel{},
		&model.AssetChannelConfig{},
		&model.AssetAccessKey{},
		&model.AssetWebhookEndpoint{},
		&model.AssetWebhookDelivery{},
		&model.AssetGroup{},
		&model.MediaAsset{},
		&model.AssetGroupReplica{},
		&model.AssetReplica{},
		&model.StorageObject{},
	))
	model.DB = db
	t.Cleanup(func() {
		model.DB = previous
		common.SetMainDatabaseType(previousDatabaseType)
	})
	return db
}

func TestResolveRequestAssetIDsMapsOnlyOwnedLocalMediaReferences(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	config := model.AssetChannelConfig{ChannelID: 9, Enabled: true, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://example.com", Region: "cn-beijing", Service: "ark", APIVersion: "2024-01-01", QPM: 60}
	require.NoError(t, db.Create(&config).Error)
	group := model.AssetGroup{PublicID: "group_local_owner", OwnerUserID: 42, Name: "avatars", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	asset := model.MediaAsset{PublicID: "asset-20260922120000-local", OwnerUserID: 42, GroupID: group.ID, StorageObjectID: 1001, Name: "portrait", AssetType: model.AssetTypeImage, ContentType: "image/png", Size: 10, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	replica := model.AssetReplica{AssetID: asset.ID, ChannelID: 9, Operation: model.AssetReplicaOperationSync, UpstreamAssetID: "asset-upstream-123", Status: model.AssetReplicaStatusActive, Progress: 100}
	require.NoError(t, db.Create(&replica).Error)

	body := []byte(`{"content":[{"type":"image_url","image_url":{"url":"asset://asset-20260922120000-local"}},{"type":"audio_url","audio_url":{"url":"asset://asset-20260922120000-local"}},{"type":"video_url","video_url":{"url":"asset://asset-upstream-existing"}},{"type":"text","text":"asset://asset-20260922120000-local"}]}`)
	mapped, count, err := ResolveRequestAssetIDs(body, 42, 9)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, "asset://asset-upstream-123", gjson.GetBytes(mapped, "content.0.image_url.url").String())
	assert.Equal(t, "asset://asset-upstream-123", gjson.GetBytes(mapped, "content.1.audio_url.url").String())
	assert.Equal(t, "asset://asset-upstream-existing", gjson.GetBytes(mapped, "content.2.video_url.url").String())
	assert.Equal(t, "asset://asset-20260922120000-local", gjson.GetBytes(mapped, "content.3.text").String())
}

func TestPublicIDsMatchVolcengineStyle(t *testing.T) {
	assetID, err := newPublicID("asset-")
	require.NoError(t, err)
	groupID, err := newPublicID("group-")
	require.NoError(t, err)

	assert.Regexp(t, regexp.MustCompile(`^asset-\d{14}-[a-z0-9]{5}$`), assetID)
	assert.Regexp(t, regexp.MustCompile(`^group-\d{14}-[a-z0-9]{5}$`), groupID)
	assert.NotContains(t, assetID, "local")
}

func TestListAssetsFiltersMediaTypeWithoutCrossTenantLeakage(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	assets := []model.MediaAsset{
		{PublicID: "asset-image", OwnerUserID: 42, StorageObjectID: 1, Name: "image", AssetType: model.AssetTypeImage, Status: model.AssetStatusReady},
		{PublicID: "asset-video", OwnerUserID: 42, StorageObjectID: 2, Name: "video", AssetType: model.AssetTypeVideo, Status: model.AssetStatusReady},
		{PublicID: "asset-other-owner", OwnerUserID: 43, StorageObjectID: 3, Name: "other", AssetType: model.AssetTypeImage, Status: model.AssetStatusReady},
	}
	require.NoError(t, db.Create(&assets).Error)

	page, err := ListAssets(AssetListInput{OwnerUserID: 42, AssetType: model.AssetTypeImage, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	assert.Equal(t, "asset-image", page.Items[0].ID)

	_, err = ListAssets(AssetListInput{OwnerUserID: 42, AssetType: "document", Page: 1, PageSize: 20})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "image, video, or audio")
}

func TestResolveRequestAssetIDsRejectsCrossTenantAndUnreadyAssets(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: 4, Enabled: true, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://example.com", Region: "cn-beijing", Service: "ark", APIVersion: "2024-01-01", QPM: 60}).Error)
	asset := model.MediaAsset{PublicID: "asset-20260922120001-owner", OwnerUserID: 7, GroupID: 1, StorageObjectID: 1002, Name: "private", AssetType: model.AssetTypeVideo, ContentType: "video/mp4", Size: 10, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	require.NoError(t, db.Create(&model.AssetReplica{AssetID: asset.ID, ChannelID: 4, Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusProcessing, Progress: 75, UpstreamAssetID: "asset-upstream-processing"}).Error)
	body := []byte(`{"content":[{"video_url":{"url":"asset://asset-20260922120001-owner"}}]}`)

	_, _, err := ResolveRequestAssetIDs(body, 8, 4)
	require.Error(t, err)
	var requestErr *RequestError
	require.True(t, errors.As(err, &requestErr))
	assert.Equal(t, http.StatusNotFound, requestErr.HTTPStatusCode())

	_, _, err = ResolveRequestAssetIDs(body, 7, 4)
	require.Error(t, err)
	require.True(t, errors.As(err, &requestErr))
	assert.Equal(t, http.StatusConflict, requestErr.HTTPStatusCode())
	assert.Contains(t, err.Error(), model.AssetReplicaStatusProcessing)
}
