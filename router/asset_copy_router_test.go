package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetCopyRoutesAllowAdministratorsForCustomersOnly(t *testing.T) {
	setupRelayRouterTestDB(t)
	db := model.DB
	require.NoError(t, db.AutoMigrate(&model.UserSession{}, &model.Log{}, &model.Channel{}, &model.Ability{}, &model.AssetGroup{}, &model.MediaAsset{}, &model.AssetCopyBatch{}, &model.AssetLibraryConnection{}, &model.AssetChannelConfig{}, &model.AssetReplica{}, &model.StorageObject{}))
	savedSecret := common.SessionSecret
	savedGroups := setting.UserUsableGroups2JSONString()
	savedRatios, err := common.Marshal(ratio_setting.GetGroupRatioCopy())
	require.NoError(t, err)
	common.SessionSecret = "asset-copy-route-fixture"
	require.NoError(t, setting.UpdateUserUsableGroupsByJSONString(`{}`))
	require.NoError(t, ratio_setting.UpdateGroupRatioByJSONString(`{"default":1,"target-only":0.8}`))
	t.Cleanup(func() {
		common.SessionSecret = savedSecret
		require.NoError(t, setting.UpdateUserUsableGroupsByJSONString(savedGroups))
		require.NoError(t, ratio_setting.UpdateGroupRatioByJSONString(string(savedRatios)))
	})
	users := make(map[int]model.User)
	tokens := make(map[int]string)
	for _, role := range []int{common.RoleCommonUser, common.RoleAdminUser, common.RoleRootUser} {
		name := fmt.Sprintf("asset-copy-role-%d", role)
		user := model.User{Username: name, AffCode: name, Role: role, Status: common.UserStatusEnabled, Group: "default", AuthVersion: 1}
		require.NoError(t, db.Create(&user).Error)
		session, err := service.CreateLoginSession(user.Id, "password", "127.0.0.1", "test")
		require.NoError(t, err)
		users[role], tokens[role] = user, session.AccessToken
	}
	target := model.User{Username: "asset-copy-target", AffCode: "asset-copy-target", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Group: "target-only"}
	require.NoError(t, db.Create(&target).Error)
	source := users[common.RoleCommonUser]
	group := model.AssetGroup{PublicID: "group-route-source", OwnerUserID: source.Id, Name: "Campaign", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	object := model.StorageObject{ObjectID: "route-copy-object", ObjectKey: "assets/route", Purpose: model.StorageObjectPurposeAssetLibrary, Status: model.StorageObjectStatusUploaded}
	require.NoError(t, db.Create(&object).Error)
	asset := model.MediaAsset{PublicID: "asset-route-source", OwnerUserID: source.Id, GroupID: group.ID, StorageObjectID: object.ID, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	channel := model.Channel{Id: 91, Name: "target channel", Status: common.ChannelStatusEnabled}
	require.NoError(t, db.Create(&channel).Error)
	require.NoError(t, db.Create(&model.Ability{Group: "target-only", Model: "video", ChannelId: channel.Id, Enabled: true}).Error)
	library := model.AssetLibraryConnection{Name: "upstream", Enabled: true}
	require.NoError(t, db.Create(&library).Error)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: channel.Id, LibraryID: library.ID, Enabled: true}).Error)
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	SetApiRouter(engine)
	request := func(role int, method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if token := tokens[role]; token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, req)
		return recorder
	}
	copyPath := "/api/asset-library/admin/copies"
	batchPath := copyPath + "/copy-admin-route-test"
	input := fmt.Sprintf(`{"request_id":"copy-admin-route-test","source_user_id":%d,"target_user_id":%d,"group_ids":["group-route-source"]}`, source.Id, target.Id)
	for _, route := range []struct{ method, path string }{{"POST", copyPath}, {"GET", batchPath}} {
		assert.Equal(t, http.StatusUnauthorized, request(0, route.method, route.path, input).Code)
		assert.Equal(t, http.StatusForbidden, request(common.RoleCommonUser, route.method, route.path, input).Code)
	}
	for _, route := range []struct{ method, path string }{{"POST", copyPath}, {"GET", batchPath}} {
		response := request(common.RoleAdminUser, route.method, route.path, input)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		assert.Contains(t, response.Body.String(), `"success":true`)
	}
	var copied model.AssetGroup
	require.NoError(t, db.Where("owner_user_id = ?", target.Id).First(&copied).Error)
	assert.Equal(t, group.Name, copied.Name)
	var copiedAsset model.MediaAsset
	require.NoError(t, db.Where("owner_user_id = ?", target.Id).First(&copiedAsset).Error)
	var replica model.AssetReplica
	require.NoError(t, db.Where("asset_id = ? AND library_id = ?", copiedAsset.ID, library.ID).First(&replica).Error)
	assert.Equal(t, model.AssetReplicaStatusPending, replica.Status)
	assert.Empty(t, replica.UpstreamAssetID)
	assert.Equal(t, http.StatusConflict, request(common.RoleAdminUser, "POST", batchPath+"/delete-sources", "").Code)
	require.NoError(t, db.Model(&copiedAsset).Update("status", model.AssetStatusReady).Error)
	require.NoError(t, db.Model(&replica).Updates(map[string]any{"status": model.AssetReplicaStatusActive, "upstream_asset_id": "target-copy-upstream"}).Error)
	assert.Equal(t, http.StatusOK, request(common.RoleAdminUser, "POST", batchPath+"/delete-sources", "").Code)
	assert.Equal(t, http.StatusForbidden, request(common.RoleAdminUser, "GET", "/api/asset-library/admin/libraries", "").Code)
	assert.Equal(t, http.StatusForbidden, request(common.RoleAdminUser, "POST", fmt.Sprintf("/api/asset-library/admin/libraries/%d/rebuild", library.ID), "").Code)
	for _, protectedRole := range []int{common.RoleAdminUser, common.RoleRootUser} {
		protectedUser := users[protectedRole]
		for _, pair := range [][2]int{{protectedUser.Id, target.Id}, {source.Id, protectedUser.Id}} {
			body := fmt.Sprintf(`{"request_id":"protected-request-001","source_user_id":%d,"target_user_id":%d,"group_ids":["group-route-source"]}`, pair[0], pair[1])
			assert.Equal(t, http.StatusForbidden, request(common.RoleAdminUser, "POST", copyPath, body).Code)
		}
		batch := model.AssetCopyBatch{ID: fmt.Sprintf("protected-batch-%d", protectedRole), SourceUserID: protectedUser.Id, TargetUserID: target.Id, Manifest: `[]`}
		require.NoError(t, db.Create(&batch).Error)
		path := copyPath + "/" + batch.ID
		assert.Equal(t, http.StatusForbidden, request(common.RoleAdminUser, "GET", path, "").Code)
		assert.Equal(t, http.StatusForbidden, request(common.RoleAdminUser, "POST", path+"/delete-sources", "").Code)
		response := request(common.RoleRootUser, "GET", path, "")
		assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
		purgeProtectedPath := fmt.Sprintf("/api/asset-library/admin/users/%d/purge", protectedUser.Id)
		assert.Equal(t, http.StatusForbidden, request(common.RoleAdminUser, "POST", purgeProtectedPath, "").Code)
	}
	purgeTargetPath := fmt.Sprintf("/api/asset-library/admin/users/%d/purge", target.Id)
	assert.Equal(t, http.StatusUnauthorized, request(0, "POST", purgeTargetPath, "").Code)
	assert.Equal(t, http.StatusForbidden, request(common.RoleCommonUser, "POST", purgeTargetPath, "").Code)
	purgeResp := request(common.RoleAdminUser, "POST", purgeTargetPath, "")
	require.Equal(t, http.StatusOK, purgeResp.Code, purgeResp.Body.String())
	assert.Contains(t, purgeResp.Body.String(), `"deleted_asset_count":1`)
	assert.Contains(t, purgeResp.Body.String(), `"deleted_group_count":1`)
}

func TestRetryAssetSyncJobSkipsAuditLog(t *testing.T) {
	setupRelayRouterTestDB(t)
	db := model.DB
	require.NoError(t, db.AutoMigrate(&model.UserSession{}, &model.Log{}, &model.Channel{}, &model.AssetGroup{}, &model.MediaAsset{}, &model.AssetLibraryConnection{}, &model.AssetChannelConfig{}, &model.AssetReplica{}))
	savedSecret := common.SessionSecret
	common.SessionSecret = "asset-retry-audit-fixture"
	t.Cleanup(func() { common.SessionSecret = savedSecret })

	root := model.User{Username: "asset-retry-root", AffCode: "asset-retry-root", Role: common.RoleRootUser, Status: common.UserStatusEnabled, Group: "default", AuthVersion: 1}
	require.NoError(t, db.Create(&root).Error)
	session, err := service.CreateLoginSession(root.Id, "password", "127.0.0.1", "test")
	require.NoError(t, err)
	group := model.AssetGroup{PublicID: "group-retry-audit", OwnerUserID: root.Id, Name: "Audit", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	asset := model.MediaAsset{PublicID: "asset-retry-audit", OwnerUserID: root.Id, GroupID: group.ID, StorageObjectID: 1, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	library := model.AssetLibraryConnection{Name: "upstream", Enabled: true}
	require.NoError(t, db.Create(&library).Error)
	replica := model.AssetReplica{
		AssetID: asset.ID, LibraryID: &library.ID, Operation: model.AssetReplicaOperationSync,
		Status: model.AssetReplicaStatusFailed, Attempts: 1, LastError: "asset synchronization timed out after 5 minutes",
	}
	require.NoError(t, db.Create(&replica).Error)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	SetApiRouter(engine)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/asset-library/admin/sync-jobs/%d/retry", replica.ID), nil)
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusAccepted, recorder.Code, recorder.Body.String())

	var updated model.AssetReplica
	require.NoError(t, db.First(&updated, replica.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusPending, updated.Status)
	assert.Zero(t, updated.Attempts)
	assert.Empty(t, updated.LastError)

	var auditCount int64
	require.NoError(t, db.Model(&model.Log{}).Count(&auditCount).Error)
	assert.Zero(t, auditCount, "single sync job retry must not write admin audit logs")
}
