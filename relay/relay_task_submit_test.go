package relay

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type taskSubmitErrorResponseBody struct {
	reader   io.Reader
	readErr  error
	closeErr error
	closes   int
}

func (body *taskSubmitErrorResponseBody) Read(buffer []byte) (int, error) {
	n, err := body.reader.Read(buffer)
	if err == io.EOF && body.readErr != nil {
		return n, body.readErr
	}
	return n, err
}

func (body *taskSubmitErrorResponseBody) Close() error {
	body.closes++
	return body.closeErr
}

type taskSubmitErrorTransport struct {
	statusCode int
	body       *taskSubmitErrorResponseBody
	requests   int
}

func (transport *taskSubmitErrorTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.requests++
	if request.Body != nil {
		_, err := io.Copy(io.Discard, request.Body)
		_ = request.Body.Close()
		if err != nil {
			return nil, err
		}
	}
	return &http.Response{StatusCode: transport.statusCode, Body: transport.body, Header: make(http.Header), Request: request}, nil
}

func TestRelayTaskSubmitClosesNon200BodyAndPreservesUpstreamError(t *testing.T) {
	// Configure a free model so this exercises the complete submit/error path
	// without creating a billing session or touching a database.
	previousPrices := ratio_setting.ModelPrice2JSONString()
	require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(`{"sora-2":0}`))
	t.Cleanup(func() { require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(previousPrices)) })
	quotaSetting := operation_setting.GetQuotaSetting()
	previousFreePreConsume := quotaSetting.EnableFreeModelPreConsume
	quotaSetting.EnableFreeModelPreConsume = false
	t.Cleanup(func() { quotaSetting.EnableFreeModelPreConsume = previousFreePreConsume })
	client, err := service.GetHttpClientWithProxySettings("", dto.ChannelSettings{})
	require.NoError(t, err)
	previousTransport := client.Transport
	t.Cleanup(func() { client.Transport = previousTransport })

	tests := []struct {
		name       string
		statusCode int
		payload    string
		readErr    error
		closeErr   error
	}{
		{name: "server error JSON", statusCode: http.StatusInternalServerError, payload: `{"error":{"message":"provider unavailable","request_id":"upstream-test-id"}}`},
		{name: "rate limit status", statusCode: http.StatusTooManyRequests, payload: `{"error":{"message":"rate limited"}}`},
		{name: "interrupted response retains partial body", statusCode: http.StatusBadGateway, payload: `{"error":`, readErr: io.ErrUnexpectedEOF},
		{name: "close error does not replace upstream error", statusCode: http.StatusServiceUnavailable, payload: `{"error":{"message":"try again later"}}`, closeErr: errors.New("response close failed")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := &taskSubmitErrorResponseBody{reader: strings.NewReader(test.payload), readErr: test.readErr, closeErr: test.closeErr}
			transport := &taskSubmitErrorTransport{statusCode: test.statusCode, body: body}
			client.Transport = transport
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", strings.NewReader(`{"model":"sora-2","prompt":"animate a bird"}`))
			ctx.Request.Header.Set("Content-Type", "application/json")
			t.Cleanup(func() { common.CleanupBodyStorage(ctx) })
			common.SetContextKey(ctx, constant.ContextKeyChannelType, constant.ChannelTypeSora)
			common.SetContextKey(ctx, constant.ContextKeyChannelBaseUrl, "https://task-response-test.invalid")
			common.SetContextKey(ctx, constant.ContextKeyChannelKey, "test-task-key")
			info := &relaycommon.RelayInfo{OriginModelName: "sora-2", UserGroup: "default", UsingGroup: "default", TaskRelayInfo: &relaycommon.TaskRelayInfo{}}

			result, taskErr := RelayTaskSubmit(ctx, info)

			assert.Nil(t, result)
			require.NotNil(t, taskErr)
			assert.Equal(t, "fail_to_fetch_task", taskErr.Code)
			assert.Equal(t, test.statusCode, taskErr.StatusCode)
			assert.Equal(t, test.payload, taskErr.Message)
			require.NotNil(t, taskErr.Error)
			assert.Equal(t, test.payload, taskErr.Error.Error())
			assert.False(t, taskErr.LocalError)
			assert.Equal(t, 1, body.closes, "the response must be released before the failed attempt returns")
			assert.Equal(t, 1, transport.requests)
			assert.Empty(t, recorder.Body.String(), "failed submission must not enter the success response writer")
		})
	}
}
