package controller

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

func TestBuildVolcNativeTaskResponseReplacesUpstreamTaskID(t *testing.T) {
	task := &model.Task{
		TaskID: "task_public_id",
		Data:   []byte(`{"id":"upstream-task-id","status":"running","watermark":false}`),
	}

	body := buildVolcNativeTaskResponse(task)

	require.Equal(t, "task_public_id", gjson.GetBytes(body, "id").String())
	require.NotContains(t, string(body), "upstream-task-id")
	require.Equal(t, false, gjson.GetBytes(body, "watermark").Bool())
}

func TestBuildVolcNativeTaskResponseSynthesizesPendingTask(t *testing.T) {
	task := &model.Task{
		TaskID: "task_public_id",
		Status: model.TaskStatusQueued,
		Properties: model.Properties{
			OriginModelName: "doubao-seedance-2-0-260128",
		},
	}

	body := buildVolcNativeTaskResponse(task)

	require.Equal(t, "task_public_id", gjson.GetBytes(body, "id").String())
	require.Equal(t, "doubao-seedance-2-0-260128", gjson.GetBytes(body, "model").String())
	require.Equal(t, "queued", gjson.GetBytes(body, "status").String())
}

func TestBuildVolcNativeTaskResponseRestoresMappedAlias(t *testing.T) {
	task := &model.Task{
		TaskID: "task_public_id",
		Status: model.TaskStatusInProgress,
		Properties: model.Properties{
			OriginModelName:   "public-seedance",
			UpstreamModelName: "doubao-seedance-2-0-260128",
		},
		Data: []byte(`{"id":"upstream-task-id","status":"running","model":"doubao-seedance-2-0-260128","watermark":false}`),
	}

	body := buildVolcNativeTaskResponse(task)

	require.Equal(t, "task_public_id", gjson.GetBytes(body, "id").String())
	require.Equal(t, "public-seedance", gjson.GetBytes(body, "model").String())
	require.NotContains(t, string(body), "doubao-seedance-2-0-260128")
	require.Contains(t, string(body), `"watermark":false`)
}

func TestVolcNativeResponseUsesDurableTerminalState(t *testing.T) {
	for _, status := range []string{"cancelled", "expired"} {
		task := &model.Task{TaskID: "task_public", Status: model.TaskStatusFailure, FailReason: status, Data: []byte(`{"id":"upstream-id","status":"running","watermark":false}`)}
		body := buildVolcNativeTaskResponse(task)
		require.Equal(t, status, gjson.GetBytes(body, "status").String())
		require.Equal(t, "task_public", gjson.GetBytes(body, "id").String())
		require.Contains(t, string(body), `"watermark":false`)
	}
}

func setupVolcNativeControllerTest(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}, &model.Channel{}, &model.User{}, &model.Token{}, &model.Log{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousCache, previousRedis, previousBatch := common.MemoryCacheEnabled, common.RedisEnabled, common.BatchUpdateEnabled
	previousLog := common.LogConsumeEnabled
	previousDBType := common.MainDatabaseType()
	model.DB, model.LOG_DB = db, db
	common.SetMainDatabaseType(common.DatabaseTypeSQLite)
	common.MemoryCacheEnabled, common.RedisEnabled, common.BatchUpdateEnabled, common.LogConsumeEnabled = false, false, false, false
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.SetMainDatabaseType(previousDBType)
		common.MemoryCacheEnabled, common.RedisEnabled, common.BatchUpdateEnabled, common.LogConsumeEnabled = previousCache, previousRedis, previousBatch, previousLog
		_ = sqlDB.Close()
	})
	return db
}

func TestVolcNativeFetchAndListRespectOwnershipAndTokenModels(t *testing.T) {
	db := setupVolcNativeControllerTest(t)
	for _, task := range []*model.Task{
		{TaskID: "task_allowed", UserId: 1, Platform: volcNativeTaskPlatform, Status: model.TaskStatusInProgress, Properties: model.Properties{OriginModelName: "seedance"}, Data: []byte(`{"id":"upstream-secret-id","status":"running","resolution":"720p"}`)},
		{TaskID: "task_denied_model", UserId: 1, Platform: volcNativeTaskPlatform, Properties: model.Properties{OriginModelName: "other-model"}},
		{TaskID: "task_other_user", UserId: 2, Platform: volcNativeTaskPlatform, Properties: model.Properties{OriginModelName: "seedance"}},
		{TaskID: "task_other_platform", UserId: 1, Platform: "54", Properties: model.Properties{OriginModelName: "seedance"}},
	} {
		require.NoError(t, db.Create(task).Error)
	}
	for _, tc := range []struct {
		id     string
		status int
	}{
		{"task_allowed", 200}, {"task_denied_model", 403}, {"task_other_user", 404}, {"task_other_platform", 404},
	} {
		t.Run(tc.id, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v3/contents/generations/tasks/"+tc.id, nil)
			ctx.Params = gin.Params{{Key: "task_id", Value: tc.id}}
			ctx.Set("id", 1)
			common.SetContextKey(ctx, constant.ContextKeyTokenModelLimitEnabled, true)
			common.SetContextKey(ctx, constant.ContextKeyTokenModelLimit, map[string]bool{"seedance": true})
			RelayVolcNativeTaskFetch(ctx)
			assert.Equal(t, tc.status, recorder.Code)
			assert.NotContains(t, recorder.Body.String(), "upstream-secret-id")
		})
	}
	adminRecorder := httptest.NewRecorder()
	adminContext, _ := gin.CreateTestContext(adminRecorder)
	adminContext.Request = httptest.NewRequest(http.MethodGet, "/api/v3/contents/generations/tasks/task_other_user", nil)
	adminContext.Params = gin.Params{{Key: "task_id", Value: "task_other_user"}}
	adminContext.Set("id", 99)
	adminContext.Set("role", common.RoleAdminUser)
	RelayVolcNativeTaskFetch(adminContext)
	assert.Equal(t, http.StatusOK, adminRecorder.Code)
	assert.Equal(t, "task_other_user", gjson.Get(adminRecorder.Body.String(), "id").String())

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v3/contents/generations/tasks?page_size=1", nil)
	ctx.Set("id", 1)
	common.SetContextKey(ctx, constant.ContextKeyTokenModelLimitEnabled, true)
	common.SetContextKey(ctx, constant.ContextKeyTokenModelLimit, map[string]bool{ratio_setting.FormatMatchingModelName("seedance"): true})
	RelayVolcNativeTaskList(ctx)
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, int64(1), gjson.Get(recorder.Body.String(), "total").Int())
	assert.Equal(t, "task_allowed", gjson.Get(recorder.Body.String(), "items.0.id").String())
	assert.Equal(t, "720p", gjson.Get(recorder.Body.String(), "items.0.resolution").String())
}

func TestRelayVolcNativeImageMapsOnlyModelAndRestoresAlias(t *testing.T) {
	db := setupVolcNativeControllerTest(t)
	savedModelPrices := ratio_setting.ModelPrice2JSONString()
	require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(`{"public-seedream":0}`))
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(savedModelPrices))
	})

	requestBody := `{"model":"public-seedream","prompt":"hello","watermark":false,"seed":0,"extra":9007199254740993}`
	expectedUpstreamBody := `{"model":"doubao-seedream-4-0-250828","prompt":"hello","watermark":false,"seed":0,"extra":9007199254740993}`
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedUpstreamBody, string(body))
		assert.Equal(t, "Bearer upstream-key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(`{"model":"doubao-seedream-4-0-250828","data":[{"url":"https://example.com/image.png"}],"watermark":false}`))
		assert.NoError(t, err)
	}))
	defer upstream.Close()

	baseURL := upstream.URL
	channel := &model.Channel{Type: constant.ChannelTypeVolcNative, Key: "upstream-key", BaseURL: &baseURL}
	require.NoError(t, db.Create(channel).Error)
	user := &model.User{Username: "volc-image-mapping", Quota: 1000}
	require.NoError(t, db.Create(user).Error)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v3/images/generations", strings.NewReader(requestBody))
	common.SetContextKey(c, constant.ContextKeyUserId, user.Id)
	common.SetContextKey(c, constant.ContextKeyUserQuota, user.Quota)
	common.SetContextKey(c, constant.ContextKeyUserSetting, dto.UserSetting{BillingPreference: "wallet_only"})
	common.SetContextKey(c, constant.ContextKeyUsingGroup, "default")
	common.SetContextKey(c, constant.ContextKeyUserGroup, "default")
	common.SetContextKey(c, constant.ContextKeyOriginalModel, "public-seedream")
	common.SetContextKey(c, constant.ContextKeyChannelId, channel.Id)
	common.SetContextKey(c, constant.ContextKeyChannelType, constant.ChannelTypeVolcNative)
	common.SetContextKey(c, constant.ContextKeyChannelBaseUrl, upstream.URL)
	common.SetContextKey(c, constant.ContextKeyChannelKey, "upstream-key")
	common.SetContextKey(c, constant.ContextKeyChannelModelMapping, `{"public-seedream":"doubao-seedream-4-0-250828"}`)

	RelayVolcNativeImage(c)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "public-seedream", gjson.Get(recorder.Body.String(), "model").String())
	assert.Equal(t, false, gjson.Get(recorder.Body.String(), "watermark").Bool())
}

func TestVolcNativeCancelReusesSubmissionKeyAndRefundsOnce(t *testing.T) {
	db := setupVolcNativeControllerTest(t)
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v3/contents/generations/tasks/upstream-task", r.URL.Path)
		assert.Equal(t, "Bearer submission-key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()
	url := upstream.URL
	ch := &model.Channel{Type: constant.ChannelTypeVolcNative, Key: "wrong-key", BaseURL: &url, UsedQuota: 100}
	require.NoError(t, db.Create(ch).Error)
	user := &model.User{Username: "volc-cancel", Quota: 1000, UsedQuota: 100}
	require.NoError(t, db.Create(user).Error)
	task := &model.Task{TaskID: "task_cancel", UserId: user.Id, ChannelId: ch.Id, Platform: volcNativeTaskPlatform, Quota: 100, Status: model.TaskStatusInProgress, Data: []byte(`{"status":"running","id":"upstream-task"}`), PrivateData: model.TaskPrivateData{Key: "submission-key", UpstreamTaskID: "upstream-task"}}
	require.NoError(t, db.Create(task).Error)
	for i := 0; i < 2; i++ {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/v3/contents/generations/tasks/task_cancel", nil)
		ctx.Params = gin.Params{{Key: "task_id", Value: "task_cancel"}}
		ctx.Set("id", user.Id)
		RelayVolcNativeTaskDelete(ctx)
		assert.Equal(t, 200, recorder.Code)
		assert.Equal(t, "cancelled", gjson.Get(recorder.Body.String(), "status").String())
		assert.NotContains(t, recorder.Body.String(), "submission-key")
	}
	assert.Equal(t, 1, calls)
	var stored model.User
	require.NoError(t, db.First(&stored, user.Id).Error)
	assert.Equal(t, 1100, stored.Quota)
	var storedTask model.Task
	require.NoError(t, db.First(&storedTask, task.ID).Error)
	assert.Zero(t, storedTask.Quota)
}
