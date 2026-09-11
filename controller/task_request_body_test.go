package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/types"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCaptureVideoTaskRequestBodyChannels(t *testing.T) {
	body := []byte(`{"model":"video-model","prompt":"hello"}`)
	tests := []struct {
		name        string
		channelType int
		wantBody    bool
	}{
		{name: "Ali Bailian", channelType: constant.ChannelTypeAli, wantBody: true},
		{name: "DoubaoVideo", channelType: constant.ChannelTypeDoubaoVideo, wantBody: true},
		{name: "native Ark", channelType: constant.ChannelTypeVolcNative, wantBody: true},
		{name: "other video channel", channelType: constant.ChannelTypeKling, wantBody: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage, err := common.CreateBodyStorage(body)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, storage.Close()) })

			got, err := captureVideoTaskRequestBody(test.channelType, storage)
			require.NoError(t, err)
			if test.wantBody {
				assert.JSONEq(t, string(body), string(got))
				return
			}
			assert.Nil(t, got)
		})
	}
}

func TestCaptureVideoTaskRequestBodyRejectsOversizedPayload(t *testing.T) {
	body := bytes.Repeat([]byte(" "), model.MaxTaskRequestBodyBytes+1)
	storage, err := common.CreateBodyStorage(body)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })

	got, err := captureVideoTaskRequestBody(constant.ChannelTypeDoubaoVideo, storage)

	assert.Nil(t, got)
	assert.True(t, errors.Is(err, model.ErrTaskRequestBodyTooLarge))
}

func TestCaptureVideoTaskRequestBodyRequiresJSON(t *testing.T) {
	storage, err := common.CreateBodyStorage([]byte("not-json"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })

	got, err := captureVideoTaskRequestBody(constant.ChannelTypeAli, storage)

	assert.Nil(t, got)
	assert.Error(t, err)
}

func TestProcessChannelErrorStoresVideoTaskRequestBody(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Log{}, &model.User{}, &model.TaskRequestBody{}))
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousErrorLogEnabled := constant.ErrorLogEnabled
	previousRedisEnabled := common.RedisEnabled
	model.DB, model.LOG_DB = db, db
	constant.ErrorLogEnabled = true
	common.RedisEnabled = false
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		constant.ErrorLogEnabled = previousErrorLogEnabled
		common.RedisEnabled = previousRedisEnabled
	})

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v3/contents/generations/tasks", nil)
	c.Set("id", 41)
	c.Set("username", "video-user")
	c.Set("token_name", "video-token")
	c.Set("original_model", "video-model")
	c.Set("channel_id", 8)
	c.Set("channel_name", "video-channel")
	c.Set("channel_type", constant.ChannelTypeVolcNative)
	c.Set("group", "default")
	c.Set(string(constant.ContextKeyRequestStartTime), time.Now())
	common.SetContextKey(c, constant.ContextKeyVideoTaskOriginalRequestBody, `{"prompt":"failed request"}`)
	common.SetContextKey(c, constant.ContextKeyVideoTaskRequestBodyStored, true)
	common.SetContextKey(c, constant.ContextKeyVideoTaskPublicID, "task_failed")
	common.SetContextKey(c, constant.ContextKeyTemporaryMediaConverted, true)
	common.SetContextKey(c, constant.ContextKeyTemporaryMediaConvertedCount, 1)

	processChannelError(
		c,
		types.ChannelError{ChannelId: 8, ChannelType: constant.ChannelTypeVolcNative},
		types.NewErrorWithStatusCode(errors.New("upstream failed"), types.ErrorCodeBadResponseStatusCode, http.StatusBadGateway),
	)

	var log model.Log
	require.NoError(t, db.First(&log).Error)
	var other map[string]any
	require.NoError(t, common.Unmarshal([]byte(log.Other), &other))
	assert.Equal(t, true, other["is_task"])
	assert.Equal(t, true, other["request_body_available"])
	assert.Equal(t, "task_failed", other["task_id"])
	assert.NotContains(t, other, "request_body")
	assert.Equal(t, true, other["temporary_media_converted"])
	assert.Equal(t, float64(1), other["temporary_media_converted_count"])
}

func TestGetTaskRequestBodyAllowsAdminLookup(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}, &model.TaskRequestBody{}))
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previousDB })

	require.NoError(t, model.SaveTaskRequestBody("task_admin", "request_admin", []byte(`{"prompt":"admin-visible"}`)))

	response := runTaskRequestBodyHandler(1, "task_admin", GetTaskRequestBody)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"success":true,"message":"","data":{"prompt":"admin-visible"}}`, response.Body.String())
}

func TestGetTaskRequestSnapshotsReturnsOriginalAndUpstreamBodies(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}, &model.TaskRequestBody{}))
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previousDB })
	require.NoError(t, model.SaveTaskRequestSnapshots(
		"task_snapshots",
		"request_snapshots",
		[]byte(`{"model":"public-model","prompt":"hello"}`),
		[]byte(`{"model":"upstream-model","content":[{"type":"text","text":"hello"}]}`),
	))

	response := runTaskRequestBodyHandler(1, "task_snapshots", GetTaskRequestSnapshots)

	assert.Equal(t, http.StatusOK, response.Code)
	var payload struct {
		Success bool                       `json:"success"`
		Data    model.TaskRequestSnapshots `json:"data"`
	}
	require.NoError(t, common.DecodeJson(response.Body, &payload))
	assert.True(t, payload.Success)
	assert.JSONEq(t, `{"model":"public-model","prompt":"hello"}`, string(payload.Data.Original))
	assert.JSONEq(t, `{"model":"upstream-model","content":[{"type":"text","text":"hello"}]}`, string(payload.Data.Upstream))
}

func TestGetTaskRequestBodyFallsBackToLegacyTaskProperty(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}, &model.TaskRequestBody{}))
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previousDB })

	require.NoError(t, db.Create(&model.Task{
		TaskID: "task_legacy",
		UserId: 73,
		Properties: model.Properties{
			RequestBody: json.RawMessage(`{"prompt":"legacy"}`),
		},
	}).Error)

	response := runTaskRequestBodyHandler(1, "task_legacy", GetTaskRequestBody)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"success":true,"message":"","data":{"prompt":"legacy"}}`, response.Body.String())
}

func TestGetLogRequestBodySupportsLegacyErrorRequestID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.TaskRequestBody{}))
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previousDB })
	require.NoError(t, model.SaveTaskRequestBody("", "request_error", []byte(`{"prompt":"failed"}`)))

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/log/request-body?request_id=request_error", nil)

	GetLogRequestBody(c)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"success":true,"message":"","data":{"prompt":"failed"}}`, recorder.Body.String())
}

func runTaskRequestBodyHandler(userID int, taskID string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/task/request-body", nil)
	c.Params = gin.Params{{Key: "task_id", Value: taskID}}
	c.Set("id", userID)
	handler(c)
	return recorder
}
