package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatUserLogsStripsQuotaSaturation verifies the admin-only quota
// saturation marker (nested under other.admin_info) is removed for non-admin
// log views, since formatUserLogs strips the whole admin_info object.
func TestFormatUserLogsStripsQuotaSaturation(t *testing.T) {
	other := common.MapToJsonStr(map[string]interface{}{
		"model_price": 0.004,
		"admin_info": map[string]interface{}{
			"quota_saturation": map[string]interface{}{
				"op":      "QuotaFromDecimal",
				"kind":    "overflow",
				"clamped": common.MaxQuota,
			},
		},
	})
	logs := []*Log{{Other: other}}

	formatUserLogs(logs, 0)

	parsed, err := common.StrToMap(logs[0].Other)
	require.NoError(t, err)
	_, hasAdminInfo := parsed["admin_info"]
	require.False(t, hasAdminInfo, "admin_info (and nested quota_saturation) must be stripped for non-admin views")
	// Non-admin billing fields remain visible.
	require.Contains(t, parsed, "model_price")
}

func TestFormatUserLogsStripsRequestBodyButKeepsConversionMarker(t *testing.T) {
	logs := []*Log{{Other: common.MapToJsonStr(map[string]interface{}{
		"request_body":                    `{"content":"data:image/webp;base64,secret"}`,
		"request_body_available":          true,
		"request_body_ref":                "request:req_secret",
		"temporary_media_converted":       true,
		"temporary_media_converted_count": 1,
	})}}

	formatUserLogs(logs, 0)

	parsed, err := common.StrToMap(logs[0].Other)
	require.NoError(t, err)
	require.NotContains(t, parsed, "request_body")
	require.NotContains(t, parsed, "request_body_available")
	require.NotContains(t, parsed, "request_body_ref")
	require.Equal(t, true, parsed["temporary_media_converted"])
	require.Equal(t, float64(1), parsed["temporary_media_converted_count"])
}

func TestGetUserLogsKeepsOwnerUsageDataWhileStrippingSensitiveMetadata(t *testing.T) {
	truncateTables(t)
	now := time.Now().Unix()
	require.NoError(t, LOG_DB.Create(&Log{
		UserId:           2,
		CreatedAt:        now,
		Type:             LogTypeConsume,
		Content:          "generate video",
		Username:         "regular-user",
		TokenName:        "user-token",
		ModelName:        "video-model",
		Quota:            12345,
		PromptTokens:     100,
		CompletionTokens: 20,
		UseTime:          5,
		Group:            "default",
		RequestId:        "request-owner",
		Other: common.MapToJsonStr(map[string]interface{}{
			"model_price":            0.004,
			"request_body":           `{"prompt":"secret"}`,
			"request_body_available": true,
			"admin_info": map[string]interface{}{
				"use_channel": []int{3},
			},
		}),
	}).Error)
	require.NoError(t, LOG_DB.Create(&Log{
		UserId:    3,
		CreatedAt: now,
		Type:      LogTypeConsume,
		Username:  "another-user",
		Quota:     999,
	}).Error)

	logs, total, err := GetUserLogs(
		2,
		LogTypeUnknown,
		now-60,
		now+60,
		"",
		"",
		0,
		20,
		"",
		"",
		"",
	)

	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.EqualValues(t, 1, total)
	assert.Equal(t, "generate video", logs[0].Content)
	assert.Equal(t, "user-token", logs[0].TokenName)
	assert.Equal(t, "video-model", logs[0].ModelName)
	assert.Equal(t, 12345, logs[0].Quota)
	assert.Equal(t, 100, logs[0].PromptTokens)
	assert.Equal(t, 20, logs[0].CompletionTokens)
	assert.Equal(t, "request-owner", logs[0].RequestId)

	other, err := common.StrToMap(logs[0].Other)
	require.NoError(t, err)
	assert.Equal(t, 0.004, other["model_price"])
	assert.NotContains(t, other, "request_body")
	assert.NotContains(t, other, "request_body_available")
	assert.NotContains(t, other, "admin_info")
}
