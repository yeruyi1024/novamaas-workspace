package relay

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

func TestTaskModel2DtoReportsRequestBodyWithoutIncludingItInListPayload(t *testing.T) {
	task := &model.Task{
		Properties: model.Properties{
			Input:       "legacy-input",
			RequestBody: json.RawMessage(`{"prompt":"private prompt"}`),
		},
	}

	result := TaskModel2Dto(task)

	assert.True(t, result.RequestBodyAvailable)
	properties, ok := result.Properties.(model.Properties)
	require.True(t, ok)
	assert.Empty(t, properties.RequestBody)
	assert.Equal(t, "legacy-input", properties.Input)
}

func TestVideoFetchByIDAllowsAdministratorTaskLogLookup(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}))
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previousDB })
	require.NoError(t, db.Create(&model.Task{
		TaskID:   "task_other_user",
		UserId:   41,
		Platform: "54",
		Status:   model.TaskStatusSuccess,
	}).Error)

	adminContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	adminContext.Request = httptest.NewRequest(http.MethodGet, "/v1/video/generations/task_other_user", nil)
	adminContext.Params = gin.Params{{Key: "task_id", Value: "task_other_user"}}
	adminContext.Set("id", 99)
	adminContext.Set("role", common.RoleAdminUser)
	body, taskErr := videoFetchByIDRespBodyBuilder(adminContext)
	require.Nil(t, taskErr)
	assert.Equal(t, "task_other_user", gjson.GetBytes(body, "data.task_id").String())

	userContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	userContext.Request = httptest.NewRequest(http.MethodGet, "/v1/video/generations/task_other_user", nil)
	userContext.Params = gin.Params{{Key: "task_id", Value: "task_other_user"}}
	userContext.Set("id", 99)
	body, taskErr = videoFetchByIDRespBodyBuilder(userContext)
	assert.Nil(t, body)
	require.NotNil(t, taskErr)
	assert.Equal(t, "task_not_exist", taskErr.Code)
}

func TestVideoFetchByIDRefreshesAliTaskFromUpstream(t *testing.T) {
	var receivedAuthorization string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthorization = r.Header.Get("Authorization")
		assert.Equal(t, "/api/v1/tasks/upstream-ali-task", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":{"task_id":"upstream-ali-task","task_status":"SUCCEEDED","video_url":"https://cdn.example.com/ali.mp4"},"request_id":"request-ali"}`))
	}))
	t.Cleanup(upstream.Close)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}, &model.Channel{}))
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previousDB })

	baseURL := upstream.URL
	channel := &model.Channel{
		Type:    constant.ChannelTypeAli,
		Name:    "ali-video",
		Key:     "current-channel-key",
		BaseURL: &baseURL,
		Status:  common.ChannelStatusEnabled,
	}
	require.NoError(t, db.Create(channel).Error)
	task := &model.Task{
		TaskID:    "task_public_ali",
		UserId:    7,
		Platform:  "17",
		ChannelId: channel.Id,
		Status:    model.TaskStatusInProgress,
		PrivateData: model.TaskPrivateData{
			Key:            "submission-key",
			UpstreamTaskID: "upstream-ali-task",
		},
		Data: json.RawMessage(`{"output":{"task_status":"RUNNING"}}`),
	}
	require.NoError(t, db.Create(task).Error)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/video/generations/task_public_ali", nil)
	ctx.Params = gin.Params{{Key: "task_id", Value: "task_public_ali"}}
	ctx.Set("id", 7)
	body, taskErr := videoFetchByIDRespBodyBuilder(ctx)

	require.Nil(t, taskErr)
	assert.Equal(t, "Bearer submission-key", receivedAuthorization)
	assert.Equal(t, "succeeded", gjson.GetBytes(body, "data.status").String())
	assert.Equal(t, "https://cdn.example.com/ali.mp4", gjson.GetBytes(body, "data.url").String())

	var reloaded model.Task
	require.NoError(t, db.First(&reloaded, task.ID).Error)
	assert.Equal(t, model.TaskStatus(model.TaskStatusInProgress), reloaded.Status, "the polling service must retain ownership of terminal settlement")
}
