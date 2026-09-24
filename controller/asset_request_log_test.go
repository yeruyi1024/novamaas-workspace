package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAssetRequestLogsReturnsScopedCursorPage(t *testing.T) {
	db := setupManageUserTestDB(t)
	previousLogDB := model.LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	model.LOG_DB = db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		model.LOG_DB = previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
	})
	require.NoError(t, model.MigrateAssetRequestLogs())
	now := time.Now().UnixMilli()
	require.NoError(t, db.Create(&[]model.AssetRequestLog{
		{CreatedAt: now - 100, ChannelID: 7, ReplicaID: 91, RequestID: "req-new", Source: "sync", Result: "failure"},
		{CreatedAt: now - 200, ChannelID: 7, ReplicaID: 91, RequestID: "req-old", Source: "sync", Result: "success"},
		{CreatedAt: now - 50, ChannelID: 8, ReplicaID: 92, RequestID: "other-channel", Source: "sync", Result: "success"},
	}).Error)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/asset-library/admin/request-logs?channel_id=7&replica_id=91&page_size=1", nil)
	ListAssetRequestLogs(c)
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Items      []model.AssetRequestLog `json:"items"`
			NextCursor string                  `json:"next_cursor"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Len(t, response.Data.Items, 1)
	assert.Equal(t, "req-new", response.Data.Items[0].RequestID)
	require.NotEmpty(t, response.Data.NextCursor)

	recorder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/asset-library/admin/request-logs?channel_id=7&replica_id=91&page_size=1&cursor="+response.Data.NextCursor, nil)
	ListAssetRequestLogs(c)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data.Items, 1)
	assert.Equal(t, "req-old", response.Data.Items[0].RequestID)
	assert.Empty(t, response.Data.NextCursor)
}

func TestGetAssetRequestLogDetailReturnsPayloadAndMissingLegacyDetail(t *testing.T) {
	db := setupManageUserTestDB(t)
	previousLogDB := model.LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	model.LOG_DB = db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		model.LOG_DB = previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
	})
	require.NoError(t, model.MigrateAssetRequestLogs())
	require.NoError(t, model.InsertAssetRequestLogEntries(context.Background(), []model.AssetRequestLogEntry{{
		Log:    model.AssetRequestLog{CreatedAt: time.Now().UnixMilli(), ChannelID: 7, HTTPStatus: 405},
		Detail: model.AssetRequestLogDetail{RequestURL: "https://example.com/?Action=ListAssetGroups", RequestBody: `{"PageNumber":1}`, ResponseBody: `{"error":"method not allowed"}`},
	}}))
	var log model.AssetRequestLog
	require.NoError(t, db.First(&log).Error)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(log.ID, 10)}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/asset-library/admin/request-logs/1", nil)
	GetAssetRequestLogDetail(c)
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data model.AssetRequestLogWithDetail `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, 405, response.Data.Log.HTTPStatus)
	require.NotNil(t, response.Data.Detail)
	assert.Contains(t, response.Data.Detail.RequestURL, "ListAssetGroups")
	assert.Contains(t, response.Data.Detail.ResponseBody, "method not allowed")
	legacy := model.AssetRequestLog{CreatedAt: time.Now().UnixMilli(), ChannelID: 7}
	require.NoError(t, db.Create(&legacy).Error)
	recorder = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(recorder)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatInt(legacy.ID, 10)}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/asset-library/admin/request-logs/"+strconv.FormatInt(legacy.ID, 10), nil)
	GetAssetRequestLogDetail(c)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Nil(t, response.Data.Detail)
}

func TestListAssetRequestLogsRejectsUnboundedTimeRange(t *testing.T) {
	now := time.Now().UnixMilli()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	url := "/api/asset-library/admin/request-logs?start_ms=" + strconv.FormatInt(now-int64((8*24*time.Hour)/time.Millisecond), 10) + "&end_ms=" + strconv.FormatInt(now, 10)
	c.Request = httptest.NewRequest(http.MethodGet, url, nil)
	ListAssetRequestLogs(c)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}
