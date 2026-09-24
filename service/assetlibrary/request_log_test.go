package assetlibrary

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetProviderRequestLogKeepsOnlyAllowlistedMetadata(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	previousLogDB := model.LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	model.LOG_DB = db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		model.LOG_DB = previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
	})
	require.NoError(t, model.MigrateAssetRequestLogs())
	writer := &assetRequestLogWriter{queue: make(chan model.AssetRequestLogEntry, 4), done: make(chan struct{})}
	require.Nil(t, activeRequestLogWriter.Swap(writer))
	go writer.run()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = StopAssetRequestLogWriter(ctx)
	})

	config := &model.AssetChannelConfig{
		ChannelID: 7, Protocol: model.AssetChannelProtocolVolcAction,
		AuthType:   model.AssetChannelAuthBearer,
		BaseURL:    "https://assets.example.com/private?token=secret-in-url",
		APIVersion: DefaultAPIVersion, QPM: 10000,
	}
	client := &volcActionClient{config: config, secret: "secret-in-header", now: time.Now}
	client.httpClient = &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway, Header: make(http.Header),
			Body: io.NopCloser(strings.NewReader(`{"ResponseMetadata":{"Error":{"Code":"BadRequest","Message":"secret-in-response"}}}`)),
		}, nil
	})}
	ctx := context.WithValue(context.Background(), common.RequestIdKey, "safe-request-id")
	ctx = withRequestLogContext(ctx, "sync", 91, 123)
	_, err := client.createAsset(ctx, "upstream-group", "private-name", model.AssetTypeImage, "https://storage.example.com/signed?token=secret-in-body")
	require.Error(t, err)
	stopCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	require.NoError(t, StopAssetRequestLogWriter(stopCtx))

	logs, _, err := model.ListAssetRequestLogs(context.Background(), model.AssetRequestLogFilter{
		ChannelID: 7, ReplicaID: 91, StartMS: time.Now().Add(-time.Minute).UnixMilli(), EndMS: time.Now().Add(time.Minute).UnixMilli(), Limit: 10,
	})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, "safe-request-id", logs[0].RequestID)
	assert.Equal(t, "CreateAsset", logs[0].Operation)
	assert.Equal(t, "failure", logs[0].Result)
	assert.Equal(t, "upstream", logs[0].ErrorKind)
	assert.Equal(t, http.StatusBadGateway, logs[0].HTTPStatus)
	assert.Equal(t, "/", logs[0].PathTemplate)
	encoded, err := common.Marshal(logs[0])
	require.NoError(t, err)
	for _, secret := range []string{"secret-in-url", "secret-in-header", "secret-in-body", "secret-in-response", "private-name"} {
		assert.NotContains(t, string(encoded), secret)
	}
	entry, err := model.GetAssetRequestLogWithDetail(context.Background(), logs[0].ID)
	require.NoError(t, err)
	require.NotNil(t, entry.Detail)
	assert.Contains(t, entry.Detail.RequestURL, "assets.example.com/private")
	assert.Contains(t, entry.Detail.RequestURL, "Action=CreateAsset")
	assert.NotContains(t, entry.Detail.RequestURL, "secret-in-url")
	assert.Contains(t, entry.Detail.RequestBody, "private-name")
	assert.NotContains(t, entry.Detail.RequestBody, "secret-in-body")
	assert.Contains(t, entry.Detail.ResponseBody, "BadRequest")
	assert.Equal(t, http.StatusBadGateway, entry.Log.HTTPStatus)
}

func TestYoufangConnectionTestRecordsSuccessfulRequest(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	previousLogDB := model.LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	model.LOG_DB = db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		model.LOG_DB = previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
	})
	require.NoError(t, model.MigrateAssetRequestLogs())
	writer := &assetRequestLogWriter{queue: make(chan model.AssetRequestLogEntry, 4), done: make(chan struct{})}
	require.Nil(t, activeRequestLogWriter.Swap(writer))
	go writer.run()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = StopAssetRequestLogWriter(ctx)
	})

	client := &youfangRESTClient{
		config: &model.AssetChannelConfig{ChannelID: 8, Protocol: model.AssetChannelProtocolYoufangREST, BaseURL: "https://assets.example.com", QPM: 10000},
		secret: "private-key",
	}
	client.httpClient = &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":{"items":[]}}`))}, nil
	})}
	ctx := context.WithValue(context.Background(), common.RequestIdKey, "connection-test-id")
	require.NoError(t, client.test(withRequestLogContext(ctx, "test", 0, 0)))
	stopCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	require.NoError(t, StopAssetRequestLogWriter(stopCtx))
	logs, _, err := model.ListAssetRequestLogs(context.Background(), model.AssetRequestLogFilter{
		ChannelID: 8, Source: "test", RequestID: "connection-test-id",
		StartMS: time.Now().Add(-time.Minute).UnixMilli(), EndMS: time.Now().Add(time.Minute).UnixMilli(), Limit: 10,
	})
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, "ListAssets", logs[0].Operation)
	assert.Equal(t, "/api/v1/assets", logs[0].PathTemplate)
	assert.Equal(t, http.StatusOK, logs[0].HTTPStatus)
	assert.Equal(t, "success", logs[0].Result)
	entry, err := model.GetAssetRequestLogWithDetail(context.Background(), logs[0].ID)
	require.NoError(t, err)
	require.NotNil(t, entry.Detail)
	assert.Equal(t, "https://assets.example.com/api/v1/assets?page=1&size=1", entry.Detail.RequestURL)
	assert.Empty(t, entry.Detail.RequestBody)
	assert.Contains(t, entry.Detail.ResponseBody, `"items":[]`)
}

func TestRequestLogDetailRedactsCredentialsAndOmitsOversizeBodies(t *testing.T) {
	detail := requestLogDetail(
		"https://user:password@example.com/path?token=private&Action=ListAssetGroups",
		[]byte(`{"api_key":"private","url":"https://storage.example.com/file?X-Amz-Signature=private&version=1","name":"visible"}`),
		[]byte(`{"error":{"authorization":"Bearer private","message":"invalid request"}}`),
	)
	assert.NotContains(t, detail.RequestURL, "private")
	assert.Contains(t, detail.RequestURL, "Action=ListAssetGroups")
	assert.NotContains(t, detail.RequestBody, "private")
	assert.Contains(t, detail.RequestBody, "visible")
	assert.NotContains(t, detail.ResponseBody, "private")
	assert.Contains(t, detail.ResponseBody, "invalid request")
	_, omitted := sanitizeRequestLogBody([]byte(strings.Repeat("x", requestLogBodyLimit+1)))
	assert.True(t, omitted)
	plain, omitted := sanitizeRequestLogBody([]byte("Method Not Allowed"))
	assert.False(t, omitted)
	assert.Equal(t, "Method Not Allowed", plain)
}

func TestAssetRequestLogStorageFailureDoesNotFailProviderRequest(t *testing.T) {
	previousDB, previousLogDB := model.DB, model.LOG_DB
	model.DB, model.LOG_DB = nil, nil
	t.Cleanup(func() { model.DB, model.LOG_DB = previousDB, previousLogDB })
	writer := &assetRequestLogWriter{queue: make(chan model.AssetRequestLogEntry, 4), done: make(chan struct{})}
	require.Nil(t, activeRequestLogWriter.Swap(writer))
	go writer.run()
	before := DroppedAssetRequestLogs()
	config := &model.AssetChannelConfig{ChannelID: 8, Protocol: model.AssetChannelProtocolYoufangREST, BaseURL: "https://assets.example.com", QPM: 10000}
	client := &youfangRESTClient{config: config, secret: "secret"}
	client.httpClient = &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":{"id":"asset-1"}}`))}, nil
	})}
	ctx := withRequestLogContext(context.Background(), "test", 0, 0)
	assert.NoError(t, client.test(ctx))
	stopCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	require.NoError(t, StopAssetRequestLogWriter(stopCtx))
	assert.Greater(t, DroppedAssetRequestLogs(), before)
}

func TestAssetRequestLogFullQueueDropsWithoutBlockingRequest(t *testing.T) {
	writer := &assetRequestLogWriter{queue: make(chan model.AssetRequestLogEntry, 1), done: make(chan struct{})}
	require.Nil(t, activeRequestLogWriter.Swap(writer))
	t.Cleanup(func() { activeRequestLogWriter.Store(nil) })
	before := DroppedAssetRequestLogs()
	ctx := withRequestLogContext(context.Background(), "sync", 91, 123)
	recordRequestAttempt(ctx, 7, model.AssetChannelProtocolVolcAction, "GetAsset", http.MethodPost, "/", 200, time.Millisecond, 0, 0, "", "", nil, nil)
	recordRequestAttempt(ctx, 7, model.AssetChannelProtocolVolcAction, "GetAsset", http.MethodPost, "/", 200, time.Millisecond, 0, 0, "", "", nil, nil)
	assert.Len(t, writer.queue, 1)
	assert.Equal(t, before+1, DroppedAssetRequestLogs())
}
