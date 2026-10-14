package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	relaytypes "github.com/QuantumNous/new-api/relaykit/types"
	hosttypes "github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestChannelErrorPersistsTransportDiagnosisOnlyForAdministrators(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Log{}, &model.BillingEntry{}))
	require.NoError(t, db.Create(&model.User{Id: 7, Username: "transport-test", Setting: "{}"}).Error)
	previousLogging, previousDisable := constant.ErrorLogEnabled, common.AutomaticDisableChannelEnabled
	constant.ErrorLogEnabled, common.AutomaticDisableChannelEnabled = true, false
	t.Cleanup(func() {
		constant.ErrorLogEnabled, common.AutomaticDisableChannelEnabled = previousLogging, previousDisable
	})
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", nil)
	ctx.Set("id", 7)
	ctx.Set("token_id", 41)
	ctx.Set("channel_id", 27)
	ctx.Set(common.RequestIdKey, "local-transport-request")
	common.SetContextKey(ctx, constant.ContextKeyUpstreamTransportError, hosttypes.UpstreamTransportError{Category: "timeout", Timeout: true})
	apiError := relaytypes.NewOpenAIError(errors.New("upstream error: do request failed"), relaytypes.ErrorCodeDoRequestFailed, http.StatusInternalServerError)
	processChannelError(ctx, relaytypes.ChannelError{ChannelId: 27}, apiError)

	var stored model.Log
	require.NoError(t, db.Where("request_id = ?", "local-transport-request").First(&stored).Error)
	assert.Equal(t, "status_code=500, upstream error: do request failed", stored.Content)
	assert.Equal(t, "timeout", gjson.Get(stored.Other, "admin_info.upstream_transport_error.category").String())
	assert.True(t, gjson.Get(stored.Other, "admin_info.upstream_transport_error.timeout").Bool())
	assert.Equal(t, "do_request_failed", gjson.Get(stored.Other, "error_code").String())

	userLogs, err := model.GetLogByTokenId(41)
	require.NoError(t, err)
	require.Len(t, userLogs, 1)
	assert.False(t, gjson.Get(userLogs[0].Other, "admin_info").Exists())
	assert.NotContains(t, userLogs[0].Other, "upstream_transport_error")
	assert.Equal(t, stored.Content, userLogs[0].Content)
}
