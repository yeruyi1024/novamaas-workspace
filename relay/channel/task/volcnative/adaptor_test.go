package volcnative

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskAdaptorDoResponsePreservesNativeFieldsAndHidesUpstreamID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	raw := []byte(`{"id":"upstream-task-id","status":"queued","model":"doubao-seedance-2-0-260128","stream":false,"sequential_image_generation":true,"watermark":false}`)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(raw)),
	}
	info := &relaycommon.RelayInfo{TaskRelayInfo: &relaycommon.TaskRelayInfo{PublicTaskID: "task_public_id"}}

	upstreamID, storedBody, taskErr := (&TaskAdaptor{}).DoResponse(ctx, resp, info)

	require.Nil(t, taskErr)
	require.Equal(t, "upstream-task-id", upstreamID)
	require.Equal(t, raw, storedBody)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"task_public_id"`)
	require.NotContains(t, recorder.Body.String(), "upstream-task-id")
	require.Contains(t, recorder.Body.String(), `"stream":false`)
	require.Contains(t, recorder.Body.String(), `"sequential_image_generation":true`)
	require.Contains(t, recorder.Body.String(), `"watermark":false`)
}

func TestVolcNativeRequestPreservesBodyAndBillingContext(t *testing.T) {
	raw := `{"model":"doubao-seedance-2-0-260128","content":[{"type":"video_url","video_url":{"url":"https://example.com/input.mp4"},"role":"reference_video"}],"resolution":"1080p","duration":5,"watermark":false,"seed":0,"extra":9007199254740993}`
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v3/contents/generations/tasks", strings.NewReader(raw))
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{}, TaskRelayInfo: &relaycommon.TaskRelayInfo{}}
	adaptor := &TaskAdaptor{}
	require.Nil(t, adaptor.ValidateRequestAndSetAction(ctx, info))
	request, err := relaycommon.GetTaskRequest(ctx)
	require.NoError(t, err)
	assert.Equal(t, false, request.Metadata["watermark"])
	assert.InDelta(t, 31.0/46.0, adaptor.EstimateBilling(ctx, info)["video_input"], 1e-12)
	body, err := adaptor.BuildRequestBody(ctx, info)
	require.NoError(t, err)
	got, err := io.ReadAll(body)
	require.NoError(t, err)
	assert.Equal(t, raw, string(got))
}

func TestVolcNativeRejectsInvalidContentAndExcessiveDuration(t *testing.T) {
	for _, raw := range []string{`{"model":"seedance"}`, `{"model":"seedance","content":null}`, `{"model":"seedance","content":[]}`, `{"model":"seedance","content":[{"type":"text","text":"hello"}],"duration":18446744073709551615}`} {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v3/contents/generations/tasks", strings.NewReader(raw))
		info := &relaycommon.RelayInfo{TaskRelayInfo: &relaycommon.TaskRelayInfo{}}
		err := (&TaskAdaptor{}).ValidateRequestAndSetAction(ctx, info)
		require.NotNil(t, err)
		assert.Equal(t, http.StatusBadRequest, err.StatusCode)
	}
}

func TestVolcNativePollingRecognizesTerminalStates(t *testing.T) {
	for _, status := range []string{"expired", "cancelled"} {
		body, err := common.Marshal(map[string]any{"status": status})
		require.NoError(t, err)
		result, err := (&TaskAdaptor{}).ParseTaskResult(body)
		require.NoError(t, err)
		assert.Equal(t, string(model.TaskStatusFailure), result.Status)
		assert.Equal(t, status, result.Reason)
	}
}
