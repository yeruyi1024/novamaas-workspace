package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedLibraryAdminResponseRedactsKeysAndSavingDoesNotSynchronize(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Channel{}, &model.AssetLibraryConnection{}, &model.AssetChannelConfig{}, &model.AssetReplica{}, &model.MediaAsset{}))
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "admin-library-fixture")
	require.NoError(t, db.Create(&model.Channel{Id: 41, Name: "pricing channel", Type: 61}).Error)
	require.NoError(t, db.Create(&model.MediaAsset{PublicID: "asset-already-saved", StorageObjectID: 71, Status: model.AssetStatusReady}).Error)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Set("id", 1)
	ctx.Set("role", common.RoleRootUser)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/asset-library/admin/libraries", strings.NewReader(`{"name":"Production account","provider_name":"Provider","enabled":true,"protocol":"volc_action","auth_type":"ak_sk","base_url":"https://provider.example","access_key_id":"fixture-access-id","credential":"fixture-secret-value"}`))
	SaveAssetLibraryConnection(ctx)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var result struct {
		Success bool
		Data    assetService.LibraryConnectionView
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
	require.True(t, result.Success)
	assert.Positive(t, result.Data.ID)
	assert.Empty(t, result.Data.Channels)
	assert.True(t, result.Data.CredentialConfigured)
	assert.NotContains(t, response.Body.String(), "fixture-access-id")
	assert.NotContains(t, response.Body.String(), "fixture-secret-value")
	assert.NotContains(t, response.Body.String(), "encrypted_credential")
	assert.NotContains(t, response.Body.String(), "resource_key")
	var count int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Count(&count).Error)
	assert.Zero(t, count, "saving parameters must not submit an upload")
}

func TestUnlinkingChannelDoesNotRequireDecryptingAccountCredentials(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Channel{}, &model.AssetChannelConfig{}, &model.AssetLibraryConnection{}, &model.AssetReplica{}, &model.MediaAsset{}, &model.AssetGroupReplica{}))
	library := model.AssetLibraryConnection{Name: "backup account", EncryptedCredential: "unreadable-ciphertext"}
	require.NoError(t, db.Create(&library).Error)
	require.NoError(t, db.Create(&model.Channel{Id: 41, Name: "video", Type: 61}).Error)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: 41, LibraryID: library.ID, Enabled: true}).Error)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Set("id", 1)
	ctx.Set("role", common.RoleRootUser)
	ctx.Params = gin.Params{{Key: "channel_id", Value: "41"}}
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/asset-library/admin/channels/41", strings.NewReader(`{"library_id":0}`))
	UpdateAssetChannelConfig(ctx)
	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.NotContains(t, response.Body.String(), "unreadable-ciphertext")
	var binding model.AssetChannelConfig
	require.NoError(t, db.First(&binding).Error)
	assert.Zero(t, binding.LibraryID)
}

func TestMergeAssetLibrariesWithNothingToMergeChangesNothing(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Channel{}, &model.AssetChannelConfig{}, &model.AssetLibraryConnection{}, &model.AssetReplica{}, &model.AssetGroupReplica{}, &model.MediaAsset{}))
	library := model.AssetLibraryConnection{Name: "only account", Enabled: true}
	require.NoError(t, db.Create(&library).Error)
	require.NoError(t, db.Create(&model.Channel{Id: 41, Name: "video", Type: 61}).Error)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: 41, LibraryID: library.ID, Enabled: true}).Error)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Set("id", 1)
	ctx.Set("role", common.RoleRootUser)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/asset-library/admin/libraries/merge", nil)
	MergeAssetLibraries(ctx)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var result struct {
		Success bool
		Data    assetService.LibraryMergeReport
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
	assert.True(t, result.Success)
	assert.Zero(t, result.Data.MergedGroups)
	assert.Zero(t, result.Data.DeferredGroups)
	var accounts int64
	require.NoError(t, db.Model(&model.AssetLibraryConnection{}).Count(&accounts).Error)
	assert.EqualValues(t, 1, accounts)
}

type assetGroupListResponse struct {
	Success bool                     `json:"success"`
	Data    []assetService.GroupView `json:"data"`
}

func requestAssetGroups(t *testing.T, userID int, role int) assetGroupListResponse {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", userID)
	ctx.Set("role", role)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/asset-library/groups?scope=all", nil)

	ListAssetGroups(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response assetGroupListResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	return response
}

func TestGetMediaAssetPreviewRejectsUnknownVariantBeforeSigning(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 42)
	ctx.Params = gin.Params{{Key: "id", Value: "asset-public-id"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/asset-library/assets/asset-public-id/preview?variant=unexpected", nil)

	GetMediaAssetPreview(ctx)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "invalid asset preview variant")
}

func TestListAssetGroupsEnforcesOwnerScopeUnlessRequesterIsAdmin(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.AssetGroup{}))
	alice := model.User{Username: "alice-assets", Password: "password", Status: common.UserStatusEnabled, Group: "default", AffCode: "asset-groups-alice"}
	bob := model.User{Username: "bob-assets", Password: "password", Status: common.UserStatusEnabled, Group: "default", AffCode: "asset-groups-bob"}
	require.NoError(t, db.Create(&alice).Error)
	require.NoError(t, db.Create(&bob).Error)
	require.NoError(t, db.Create(&[]model.AssetGroup{
		{PublicID: "group-alice", OwnerUserID: alice.Id, Name: "Campaign", Status: model.AssetStatusReady},
		{PublicID: "group-bob", OwnerUserID: bob.Id, Name: "Campaign", Status: model.AssetStatusReady},
	}).Error)

	memberResponse := requestAssetGroups(t, alice.Id, common.RoleCommonUser)
	require.Len(t, memberResponse.Data, 1)
	assert.Equal(t, "group-alice", memberResponse.Data[0].ID)

	adminResponse := requestAssetGroups(t, alice.Id, common.RoleAdminUser)
	require.Len(t, adminResponse.Data, 2)
	creators := make(map[string]string, len(adminResponse.Data))
	for _, group := range adminResponse.Data {
		creators[group.ID] = group.OwnerName
	}
	assert.Equal(t, map[string]string{
		"group-alice": "alice-assets",
		"group-bob":   "bob-assets",
	}, creators)
}

func TestListAssetGroupsPagePreservesAdminScopeAndReturnsPaginationMetadata(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.AssetGroup{}))
	owner := model.User{Username: "paged-owner", Password: "password", Status: common.UserStatusEnabled, Group: "default", AffCode: "paged-owner"}
	other := model.User{Username: "paged-other", Password: "password", Status: common.UserStatusEnabled, Group: "default", AffCode: "paged-other"}
	require.NoError(t, db.Create(&owner).Error)
	require.NoError(t, db.Create(&other).Error)
	require.NoError(t, db.Create(&[]model.AssetGroup{
		{PublicID: "group-zeta", OwnerUserID: owner.Id, Name: "Zeta", Status: model.AssetStatusReady},
		{PublicID: "group-alpha", OwnerUserID: other.Id, Name: "Alpha", Status: model.AssetStatusReady},
	}).Error)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", owner.Id)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/asset-library/groups?scope=all&p=1&page_size=1", nil)
	ListAssetGroups(ctx)
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Items    []assetService.GroupView `json:"items"`
			Total    int64                    `json:"total"`
			Page     int                      `json:"page"`
			PageSize int                      `json:"page_size"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	assert.EqualValues(t, 2, response.Data.Total)
	assert.Equal(t, 1, response.Data.Page)
	assert.Equal(t, 1, response.Data.PageSize)
	require.Len(t, response.Data.Items, 1)
	assert.Equal(t, "group-alpha", response.Data.Items[0].ID)
	assert.Equal(t, "paged-other", response.Data.Items[0].OwnerName)
}
