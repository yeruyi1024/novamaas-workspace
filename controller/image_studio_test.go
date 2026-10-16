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

func TestImageStudioSnapshotsAllowAdminsWithoutImageReviewGrant(t *testing.T) {
	db := setupManageUserTestDB(t)
	previousMaster := common.IsMasterNode
	common.IsMasterNode = true
	t.Cleanup(func() { common.IsMasterNode = previousMaster })
	require.NoError(t, authz.Init(db))
	require.NoError(t, authz.SetUserPermissions(8, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: false}}))
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.TaskRequestBody{}, &model.Task{}))
	require.NoError(t, db.Create(&model.ImageStudioRun{UserID: 7, IdempotencyKey: "k", Prompt: "private prompt", Status: model.ImageStudioRunStatusCompleted, RequestID: "studio-request"}).Error)
	require.NoError(t, model.SaveTaskRequestSnapshots("", "studio-request", []byte(`{"prompt":"private prompt"}`), []byte(`{"prompt":"private prompt","model":"upstream"}`)))
	require.NoError(t, model.SaveTaskRequestSnapshots("", "chat-request", []byte(`{"prompt":"chat"}`), nil))
	var upstreamRow model.TaskRequestBody
	require.NoError(t, db.Where("request_id = ? AND reference_id LIKE ?", "studio-request", "upstream:%").First(&upstreamRow).Error)
	upstreamReference := upstreamRow.ReferenceID

	readers := []struct {
		name    string
		handler gin.HandlerFunc
		path    string
		params  gin.Params
	}{
		{"log snapshots", GetLogRequestSnapshots, "/?request_id=studio-request", nil},
		{"log body", GetLogRequestBody, "/?request_id=studio-request", nil},
		{"log snapshots by reference", GetLogRequestSnapshots, "/?task_id=request:studio-request", nil},
		{"task snapshots by reference", GetTaskRequestSnapshots, "/", gin.Params{{Key: "task_id", Value: "request:studio-request"}}},
		{"task body by reference", GetTaskRequestBody, "/", gin.Params{{Key: "task_id", Value: "request:studio-request"}}},
		{"upstream body by archive reference", GetTaskRequestBody, "/", gin.Params{{Key: "task_id", Value: upstreamReference}}},
		{"upstream log body by archive reference", GetLogRequestBody, "/?task_id=" + upstreamReference, nil},
		{"log original download", DownloadRequestSnapshot, "/?request_id=studio-request", gin.Params{{Key: "kind", Value: "original"}}},
		{"log upstream download", DownloadRequestSnapshot, "/?request_id=studio-request", gin.Params{{Key: "kind", Value: "upstream"}}},
		{"log download by reference", DownloadRequestSnapshot, "/?task_id=request:studio-request", gin.Params{{Key: "kind", Value: "original"}}},
		{"task original download", DownloadRequestSnapshot, "/", gin.Params{{Key: "task_id", Value: "request:studio-request"}, {Key: "kind", Value: "original"}}},
		{"task upstream download", DownloadRequestSnapshot, "/", gin.Params{{Key: "task_id", Value: "request:studio-request"}, {Key: "kind", Value: "upstream"}}},
		{"download by archive reference", DownloadRequestSnapshot, "/", gin.Params{{Key: "task_id", Value: upstreamReference}, {Key: "kind", Value: "original"}}},
	}
	for _, reader := range readers {
		t.Run(reader.name, func(t *testing.T) {
			response := studioRequest(reader.handler, 8, common.RoleAdminUser, reader.path, reader.params)
			assert.Equal(t, http.StatusOK, response.Code)
			assert.Contains(t, response.Body.String(), "private prompt")
		})
	}
	response := studioRequest(GetLogRequestSnapshots, 8, common.RoleAdminUser, "/?request_id=chat-request", nil)
	assert.Equal(t, http.StatusOK, response.Code)
}

// The relay links its request ID to the running run before archiving, so a
// snapshot never exists without an owner run, and a finished run cannot be
// relinked to another request.
func TestImageStudioRunRequestLinkPrecedesArchive(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}))
	run := &model.ImageStudioRun{UserID: 7, IdempotencyKey: "link", Status: model.ImageStudioRunStatusRunning}
	require.NoError(t, db.Create(run).Error)
	require.NoError(t, model.LinkImageStudioRunRequest(run.ID, "first"))
	require.NoError(t, model.LinkImageStudioRunRequest(run.ID, "first"), "channel retries reuse the request ID")
	assert.Error(t, model.LinkImageStudioRunRequest(run.ID, "other"))
	var linked model.ImageStudioRun
	require.NoError(t, db.First(&linked, run.ID).Error)
	assert.Equal(t, "first", linked.RequestID)
	require.NoError(t, db.Model(run).Update("status", model.ImageStudioRunStatusCompleted).Error)
	assert.Error(t, model.LinkImageStudioRunRequest(run.ID, "first"))
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
	serve := func(path string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		return recorder
	}

	setting.Enabled = false
	assert.Equal(t, http.StatusOK, serve("/settings").Code, "the page reads the switch to block its route")
	assert.Equal(t, http.StatusServiceUnavailable, serve("/runs").Code)
	assert.Equal(t, http.StatusServiceUnavailable, serve("/runs?scope=all").Code, "review must also be closed for root")

	setting.Enabled = true
	list := serve("/runs")
	require.Equal(t, http.StatusOK, list.Code)
	assert.Contains(t, list.Body.String(), `"total":1`)
}
