package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service/authz"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func studioRequest(handler gin.HandlerFunc, user, role int, path string, params gin.Params) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
	c.Set("id", user)
	c.Set("role", role)
	c.Params = params
	handler(c)
	return recorder
}

func TestImageStudioReviewRequiresPermission(t *testing.T) {
	db := setupManageUserTestDB(t)
	previousMaster := common.IsMasterNode
	common.IsMasterNode = true
	t.Cleanup(func() { common.IsMasterNode = previousMaster })
	require.NoError(t, authz.Init(db))
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}))
	run := &model.ImageStudioRun{UserID: 7, IdempotencyKey: "k", Prompt: "private prompt", Status: model.ImageStudioRunStatusFailed}
	require.NoError(t, db.Create(run).Error)
	for _, tc := range []struct {
		name    string
		role    int
		granted bool
		status  int
	}{
		{"ordinary user with grant", common.RoleCommonUser, true, http.StatusForbidden},
		{"ordinary admin", common.RoleAdminUser, false, http.StatusForbidden},
		{"granted admin", common.RoleAdminUser, true, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, authz.SetUserPermissions(8, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: tc.granted}}))
			list := studioRequest(ListImageStudioRuns, 8, tc.role, "/?scope=all", nil)
			assert.Equal(t, tc.status, list.Code)
			detail := studioRequest(GetImageStudioRun, 8, tc.role, "/?scope=all", gin.Params{{Key: "id", Value: run.PublicID}})
			assert.Equal(t, tc.status, detail.Code)
			assert.Equal(t, tc.status == http.StatusOK, strings.Contains(detail.Body.String(), "private prompt"))
		})
	}
	// Without scope=all a user only sees their own runs.
	own := studioRequest(GetImageStudioRun, 8, common.RoleRootUser, "/", gin.Params{{Key: "id", Value: run.PublicID}})
	assert.Equal(t, http.StatusNotFound, own.Code)
}

// Disabling the feature blocks the studio API but keeps runs and images, so
// they are listed again once it is re-enabled.
func TestDisabledImageStudioBlocksApiAndKeepsHistory(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}))
	require.NoError(t, db.Create(&model.ImageStudioRun{UserID: 8, IdempotencyKey: "kept", Status: model.ImageStudioRunStatusFailed}).Error)
	setting := operation_setting.GetImageStudioSetting()
	previousEnabled := setting.Enabled
	t.Cleanup(func() { setting.Enabled = previousEnabled })
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set("id", 8); c.Set("role", common.RoleRootUser) })
	engine.GET("/settings", GetImageStudioSettings)
	engine.GET("/runs", ImageStudioEnabled(), ListImageStudioRuns)
	engine.GET("/logs", ListImageStudioLogs)
	serve := func(path string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		return recorder
	}

	setting.Enabled = false
	assert.Equal(t, http.StatusOK, serve("/settings").Code, "the page reads the switch to block its route")
	assert.Equal(t, http.StatusServiceUnavailable, serve("/runs").Code)
	assert.Equal(t, http.StatusOK, serve("/logs").Code, "tracking remains readable when generation is disabled")
	assert.Equal(t, http.StatusServiceUnavailable, serve("/runs?scope=all").Code, "review must also be closed for root")

	setting.Enabled = true
	list := serve("/runs")
	require.Equal(t, http.StatusOK, list.Code)
	assert.Contains(t, list.Body.String(), `"total":1`)
}

func TestReviewDeleteRequiresReadAndWriteAndKeepsLibraryCopy(t *testing.T) {
	db := setupManageUserTestDB(t)
	previousMaster := common.IsMasterNode
	common.IsMasterNode = true
	t.Cleanup(func() { common.IsMasterNode = previousMaster })
	require.NoError(t, authz.Init(db))
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.StorageObject{}, &model.MediaAsset{}))
	run := &model.ImageStudioRun{UserID: 7, IdempotencyKey: "review-delete", Status: model.ImageStudioRunStatusCompleted}
	require.NoError(t, db.Create(run).Error)
	object := &model.StorageObject{ObjectID: "generated-review-image", OwnerUserID: 7, Purpose: model.StorageObjectPurposeImageStudio, Status: model.StorageObjectStatusUploaded, DeleteAfter: common.GetTimestamp() + 3600}
	require.NoError(t, db.Create(object).Error)
	asset := &model.MediaAsset{PublicID: "saved-library-copy", OwnerUserID: 7, StorageObjectID: 1234, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(asset).Error)
	require.NoError(t, db.Create(&model.ImageStudioResult{RunID: run.ID, StorageObjectID: object.ID, MediaAssetID: asset.ID}).Error)
	params := gin.Params{{Key: "id", Value: run.PublicID}}
	for _, tc := range []struct {
		name        string
		role        int
		read, write bool
	}{
		{"ordinary user", common.RoleCommonUser, true, true},
		{"read only", common.RoleAdminUser, true, false},
		{"write without read", common.RoleAdminUser, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, authz.SetUserPermissions(8, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: tc.read, authz.ActionWrite: tc.write}}))
			response := studioRequest(DeleteImageStudioRun, 8, tc.role, "/?scope=all", params)
			assert.Equal(t, http.StatusForbidden, response.Code)
		})
	}
	require.NoError(t, authz.SetUserPermissions(8, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: true, authz.ActionWrite: true}}))
	own := studioRequest(DeleteImageStudioRun, 8, common.RoleAdminUser, "/", params)
	assert.Equal(t, http.StatusNotFound, own.Code)
	response := studioRequest(DeleteImageStudioRun, 8, common.RoleAdminUser, "/?scope=all", params)
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, db.First(run, run.ID).Error)
	assert.Equal(t, model.ImageStudioRunStatusDeleted, run.Status)
	require.NoError(t, db.First(asset, asset.ID).Error)
	assert.Equal(t, model.AssetStatusReady, asset.Status)
	require.NoError(t, db.First(object, object.ID).Error)
	assert.Equal(t, model.StorageObjectStatusDeletePending, object.Status)
	var logs []model.Log
	require.NoError(t, db.Where("type = ?", model.LogTypeManage).Find(&logs).Error)
	require.Len(t, logs, 1)
	assert.Equal(t, 8, logs[0].UserId)
	assert.Contains(t, logs[0].Other, "image_studio.review_delete")
	assert.Contains(t, logs[0].Other, run.PublicID)
}

func TestImageStudioBatchPollIsBoundedAndOwnerScoped(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}))
	own := &model.ImageStudioRun{UserID: 7, IdempotencyKey: "batch-own", Status: model.ImageStudioRunStatusRunning}
	foreign := &model.ImageStudioRun{UserID: 8, IdempotencyKey: "batch-foreign", Status: model.ImageStudioRunStatusRunning}
	deleted := &model.ImageStudioRun{UserID: 7, IdempotencyKey: "batch-deleted", Status: model.ImageStudioRunStatusDeleted}
	for _, run := range []*model.ImageStudioRun{own, foreign, deleted} {
		require.NoError(t, db.Create(run).Error)
	}
	response := studioRequest(PollImageStudioRuns, 7, common.RoleCommonUser, "/?ids="+own.PublicID+","+foreign.PublicID+","+deleted.PublicID, nil)
	require.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), own.PublicID)
	assert.NotContains(t, response.Body.String(), foreign.PublicID)
	assert.NotContains(t, response.Body.String(), deleted.PublicID)
	assert.Contains(t, response.Body.String(), `"total":1`)
	for _, path := range []string{"/", "/?ids=" + strings.Repeat("run-x,", 50) + "run-x"} {
		response = studioRequest(PollImageStudioRuns, 7, common.RoleRootUser, path, nil)
		assert.Equal(t, http.StatusBadRequest, response.Code)
	}
	response = studioRequest(PollImageStudioRuns, 0, common.RoleCommonUser, "/?ids="+own.PublicID, nil)
	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestImageStudioReviewListDoesNotAuditButDetailStillDoes(t *testing.T) {
	db := setupManageUserTestDB(t)
	previousMaster := common.IsMasterNode
	common.IsMasterNode = true
	t.Cleanup(func() { common.IsMasterNode = previousMaster })
	require.NoError(t, authz.Init(db))
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}))
	run := &model.ImageStudioRun{UserID: 7, IdempotencyKey: "list-no-audit", Status: model.ImageStudioRunStatusFailed}
	require.NoError(t, db.Create(run).Error)
	for _, path := range []string{"/?scope=all", "/?scope=all&user_id=7", "/?scope=all&p=2"} {
		response := studioRequest(ListImageStudioRuns, 8, common.RoleRootUser, path, nil)
		require.Equal(t, http.StatusOK, response.Code)
	}
	var logs []model.Log
	require.NoError(t, db.Where("type = ?", model.LogTypeManage).Find(&logs).Error)
	assert.Empty(t, logs, "refreshing, filtering or paging the review list must not write audit records")
	response := studioRequest(GetImageStudioRun, 8, common.RoleRootUser, "/?scope=all", gin.Params{{Key: "id", Value: run.PublicID}})
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, db.Where("type = ?", model.LogTypeManage).Find(&logs).Error)
	require.Len(t, logs, 1)
	assert.Contains(t, logs[0].Other, "image_studio.review_read")
}

func TestImageStudioTaskLogsPreserveTrackingAndScopeWithoutImageContents(t *testing.T) {
	db := setupManageUserTestDB(t)
	previousMaster := common.IsMasterNode
	common.IsMasterNode = true
	t.Cleanup(func() { common.IsMasterNode = previousMaster })
	require.NoError(t, authz.Init(db))
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}))
	own := &model.ImageStudioRun{UserID: 7, IdempotencyKey: "log-own", ChannelID: 17, Model: "image2", Status: model.ImageStudioRunStatusCompleted,
		Prompt: "private prompt", RequestID: "request-own", CreatedAt: 150, DurationMS: 1234}
	deleted := &model.ImageStudioRun{UserID: 7, IdempotencyKey: "log-deleted", Status: model.ImageStudioRunStatusDeleted, CreatedAt: 170}
	foreign := &model.ImageStudioRun{UserID: 8, IdempotencyKey: "log-foreign", ChannelID: 23, Status: model.ImageStudioRunStatusFailed, CreatedAt: 160, ErrorMessage: "provider unavailable"}
	for _, run := range []*model.ImageStudioRun{own, deleted, foreign} {
		require.NoError(t, db.Create(run).Error)
	}
	// No image results/storage tables are needed to read task metadata.
	response := studioRequest(ListImageStudioLogs, 7, common.RoleCommonUser, "/?start_timestamp=100&end_timestamp=200", nil)
	require.Equal(t, http.StatusOK, response.Code)
	var result struct {
		Data struct {
			Items []struct {
				ID        string `json:"id"`
				RequestID string `json:"request_id"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
	assert.Equal(t, 2, result.Data.Total)
	require.Len(t, result.Data.Items, 2)
	assert.Equal(t, deleted.PublicID, result.Data.Items[0].ID)
	assert.Equal(t, "request-own", result.Data.Items[1].RequestID)
	assert.NotContains(t, response.Body.String(), foreign.PublicID)
	for _, field := range []string{"prompt", "results", "preview_url", "download_url", "quota", "channel_id"} {
		assert.NotContains(t, response.Body.String(), `"`+field+`"`)
	}
	for _, filter := range []string{own.PublicID, own.RequestID} {
		response = studioRequest(ListImageStudioLogs, 7, common.RoleCommonUser, "/?filter="+filter, nil)
		require.Equal(t, http.StatusOK, response.Code)
		assert.Contains(t, response.Body.String(), `"total":1`)
		assert.Contains(t, response.Body.String(), own.PublicID)
	}
	for _, tc := range []struct {
		role   int
		grant  bool
		status int
	}{
		{common.RoleCommonUser, true, http.StatusForbidden}, {common.RoleAdminUser, false, http.StatusForbidden}, {common.RoleAdminUser, true, http.StatusOK},
	} {
		require.NoError(t, authz.SetUserPermissions(9, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: tc.grant}}))
		response = studioRequest(ListImageStudioLogs, 9, tc.role, "/?scope=all", nil)
		assert.Equal(t, tc.status, response.Code)
		if tc.status == http.StatusOK {
			assert.Contains(t, response.Body.String(), foreign.PublicID)
		}
	}
	response = studioRequest(ListImageStudioLogs, 9, common.RoleRootUser, "/?scope=all&channel_id=23", nil)
	require.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), `"total":1`)
	assert.Contains(t, response.Body.String(), `"channel_id":23`)
	assert.Contains(t, response.Body.String(), `"error_message":"provider unavailable"`)
	assert.NotContains(t, response.Body.String(), own.PublicID)
	response = studioRequest(ListImageStudioLogs, 7, common.RoleCommonUser, "/?start_timestamp=180&end_timestamp=200", nil)
	assert.Contains(t, response.Body.String(), `"total":0`)
	response = studioRequest(ListImageStudioLogs, 7, common.RoleCommonUser, "/?page_size=-1&p=-1", nil)
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
	assert.Len(t, result.Data.Items, 1, "negative page size must never disable the SQL limit")
	for _, path := range []string{"/?start_timestamp=bad", "/?channel_id=-1", "/?channel_id=23", "/?start_timestamp=200&end_timestamp=100", "/?filter=" + strings.Repeat("x", 65)} {
		assert.Equal(t, http.StatusBadRequest, studioRequest(ListImageStudioLogs, 7, common.RoleCommonUser, path, nil).Code)
	}
	var count int64
	require.NoError(t, db.Model(&model.Log{}).Where("type = ?", model.LogTypeManage).Count(&count).Error)
	assert.Zero(t, count, "task metadata reads do not create audit noise")
}
