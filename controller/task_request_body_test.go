package controller

import (
	"bytes"
	"encoding/json"
	"errors"
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
	body := bytes.Repeat([]byte(" "), maxStoredVideoTaskRequestBodyBytes+1)
	storage, err := common.CreateBodyStorage(body)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })

	got, err := captureVideoTaskRequestBody(constant.ChannelTypeDoubaoVideo, storage)

	assert.Nil(t, got)
	assert.True(t, errors.Is(err, errVideoTaskRequestBodyTooLarge))
}

func TestCaptureVideoTaskRequestBodyRequiresJSON(t *testing.T) {
	storage, err := common.CreateBodyStorage([]byte("not-json"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, storage.Close()) })

	got, err := captureVideoTaskRequestBody(constant.ChannelTypeAli, storage)

	assert.Nil(t, got)
	assert.Error(t, err)
}

func TestGetUserTaskRequestBodyEnforcesOwnership(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}))
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previousDB })

	require.NoError(t, db.Create(&model.Task{
		TaskID: "task_owned",
		UserId: 41,
		Properties: model.Properties{
			RequestBody: json.RawMessage(`{"prompt":"owned"}`),
		},
	}).Error)

	allowed := runTaskRequestBodyHandler(41, "task_owned", GetUserTaskRequestBody)
	assert.Equal(t, http.StatusOK, allowed.Code)
	assert.JSONEq(t, `{"success":true,"message":"","data":{"prompt":"owned"}}`, allowed.Body.String())

	denied := runTaskRequestBodyHandler(42, "task_owned", GetUserTaskRequestBody)
	assert.Equal(t, http.StatusNotFound, denied.Code)
}

func TestGetTaskRequestBodyAllowsAdminLookup(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}))
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previousDB })

	require.NoError(t, db.Create(&model.Task{
		TaskID: "task_admin",
		UserId: 73,
		Properties: model.Properties{
			RequestBody: json.RawMessage(`{"prompt":"admin-visible"}`),
		},
	}).Error)

	response := runTaskRequestBodyHandler(1, "task_admin", GetTaskRequestBody)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"success":true,"message":"","data":{"prompt":"admin-visible"}}`, response.Body.String())
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
