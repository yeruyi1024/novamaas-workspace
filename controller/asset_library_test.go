package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
