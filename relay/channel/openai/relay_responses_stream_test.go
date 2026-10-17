package openai

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupResponsesStreamReliabilityTest(t *testing.T, body io.ReadCloser) (*gin.Context, *httptest.ResponseRecorder, *relaycommon.RelayInfo, *http.Response) {
	t.Helper()
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	info := &relaycommon.RelayInfo{
		DisablePing: true,
		ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "kimi-k3"},
	}
	return c, w, info, &http.Response{StatusCode: http.StatusOK, Body: body}
}

func TestResponsesStreamTerminalBeforeTransportError(t *testing.T) {
	for _, eventType := range []string{"response.completed", "response.done", "response.in_progress"} {
		t.Run(eventType, func(t *testing.T) {
			event := `{"type":"` + eventType + `","sequence_number":7,"response":{"usage":{"input_tokens":30976,"output_tokens":113,"total_tokens":31089,"input_tokens_details":{"cached_tokens":30000}}}}`
			// Reproduce an incomplete chunked HTTP/1.1 response, rather than a
			// graceful EOF. A complete protocol response still takes precedence.
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, _, err := w.(http.Hijacker).Hijack()
				if err != nil {
					return
				}
				defer conn.Close()
				payload := "data: " + event + "\n\n"
				_, _ = fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\nTransfer-Encoding: chunked\r\n\r\n%x\r\n%s\r\n", len(payload), payload)
			}))
			t.Cleanup(upstream.Close)
			resp, err := upstream.Client().Get(upstream.URL)
			require.NoError(t, err)
			c, w, info, _ := setupResponsesStreamReliabilityTest(t, resp.Body)

			usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
			assert.Contains(t, w.Body.String(), "data: "+event+"\n\n")
			assert.Equal(t, 30976, usage.PromptTokens)
			assert.Equal(t, 113, usage.CompletionTokens)
			assert.Equal(t, 31089, usage.TotalTokens)
			assert.Equal(t, 30000, usage.PromptTokensDetails.CachedTokens)
			if eventType == "response.in_progress" {
				assert.Equal(t, relaycommon.StreamEndReasonScannerErr, info.StreamStatus.EndReason)
				assert.ErrorIs(t, info.StreamStatus.EndError, io.ErrUnexpectedEOF)
				assert.Contains(t, w.Body.String(), "event: error\n")
				assert.Contains(t, w.Body.String(), `"code":"upstream_stream_incomplete"`)
				assert.Contains(t, w.Body.String(), `"sequence_number":8`)
				assert.NotContains(t, w.Body.String(), "unexpected EOF")
				assert.NotContains(t, w.Body.String(), "response.completed")
				return
			}
			assert.Equal(t, relaycommon.StreamEndReasonDone, info.StreamStatus.EndReason)
			assert.NoError(t, info.StreamStatus.EndError)
			assert.False(t, info.StreamStatus.HasErrors())
			assert.NotContains(t, w.Body.String(), "event: error\n")
		})
	}
}

func TestResponsesStreamReturnsAtEveryTerminalWithoutHTTPEOF(t *testing.T) {
	tests := []struct {
		eventType string
		status    string
		failed    bool
	}{
		{eventType: "response.completed", status: "completed"},
		{eventType: "response.done", status: "completed"},
		{eventType: "response.done", status: "incomplete"},
		{eventType: "response.done", status: "failed", failed: true},
		{eventType: "response.failed", status: "failed", failed: true},
		{eventType: "response.incomplete", status: "incomplete"},
		{eventType: "response.cancelled", status: "cancelled", failed: true},
		{eventType: "response.canceled", status: "canceled", failed: true},
		{eventType: "response.error", failed: true},
		{eventType: "error", failed: true},
	}
	for _, tt := range tests {
		t.Run(tt.eventType+"/"+tt.status, func(t *testing.T) {
			reader, writer := io.Pipe()
			t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() })
			c, w, info, resp := setupResponsesStreamReliabilityTest(t, reader)
			event := `{"type":"` + tt.eventType + `","response":{"status":"` + tt.status + `","usage":{"input_tokens":11,"output_tokens":0,"total_tokens":11}}}`
			done := make(chan struct{})
			go func() {
				defer close(done)
				usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
				assert.Nil(t, apiErr)
				if assert.NotNil(t, usage) {
					assert.Equal(t, 11, usage.PromptTokens)
					assert.Zero(t, usage.CompletionTokens)
				}
			}()
			_, err := io.WriteString(writer, "data: "+event+"\n\n")
			require.NoError(t, err)
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				_ = writer.Close()
				<-done
				t.Fatal("terminal Responses event waited for HTTP EOF")
			}
			assert.Contains(t, w.Body.String(), "event: "+tt.eventType+"\n")
			assert.Equal(t, tt.failed, info.StreamStatus.HasErrors())
			if tt.failed {
				assert.Equal(t, relaycommon.StreamEndReasonHandlerStop, info.StreamStatus.EndReason)
			} else {
				assert.Equal(t, relaycommon.StreamEndReasonDone, info.StreamStatus.EndReason)
			}
			_, err = io.WriteString(writer, "data: unexpected\n\n")
			assert.ErrorIs(t, err, io.ErrClosedPipe)
		})
	}
}

func TestResponsesStreamMissingTerminalEmitsError(t *testing.T) {
	for _, ending := range []string{"", "data: [DONE]\n\n"} {
		t.Run(fmt.Sprintf("done_marker=%t", ending != ""), func(t *testing.T) {
			body := "data: {\"type\":\"response.in_progress\",\"sequence_number\":12,\"response\":{\"usage\":{\"input_tokens\":10,\"output_tokens\":0,\"total_tokens\":10}}}\n\n" +
				"data: {\"type\":\"response.output_text.delta\",\"sequence_number\":13,\"delta\":\"partial output\"}\n\n" + ending
			c, w, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader(body)))
			usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
			assert.Equal(t, 10, usage.PromptTokens)
			assert.Positive(t, usage.CompletionTokens, "retain the existing partial-output token fallback")
			assert.True(t, info.StreamStatus.HasErrors())
			expectedReason := relaycommon.StreamEndReasonEOF
			if ending != "" {
				expectedReason = relaycommon.StreamEndReasonDone
			}
			assert.Equal(t, expectedReason, info.StreamStatus.EndReason)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `event: error`)
			assert.Contains(t, w.Body.String(), `"type":"error","code":"upstream_stream_incomplete"`)
			assert.Contains(t, w.Body.String(), `"param":null,"sequence_number":14`)
			assert.NotContains(t, w.Body.String(), "response.completed")
		})
	}
}

func TestResponsesStreamTimeoutEmitsErrorAfterPartialResponse(t *testing.T) {
	reader, writer := io.Pipe()
	t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() })
	c, w, info, resp := setupResponsesStreamReliabilityTest(t, reader)
	constant.StreamingTimeout = 1
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, apiErr := OaiResponsesStreamHandler(c, info, resp)
		assert.Nil(t, apiErr)
	}()
	_, err := io.WriteString(writer, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial output\"}\n\n")
	require.NoError(t, err)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = writer.Close()
		<-done
		t.Fatal("idle upstream was not stopped by the stream timeout")
	}
	assert.Equal(t, relaycommon.StreamEndReasonTimeout, info.StreamStatus.EndReason)
	assert.Contains(t, w.Body.String(), `"code":"upstream_stream_timeout"`)
	assert.NotContains(t, w.Body.String(), "response.completed")
}

type responsesCancelWriter struct {
	gin.ResponseWriter
	cancel context.CancelFunc
}

type responsesDisconnectedWriter struct {
	gin.ResponseWriter
	err      error
	attempts int
}

type responsesCancelOnFlushWriter struct {
	gin.ResponseWriter
	cancel  context.CancelFunc
	flushes int
}

func (w *responsesCancelOnFlushWriter) Flush() {
	w.flushes++
	w.ResponseWriter.Flush()
	w.cancel()
}

func (w *responsesDisconnectedWriter) Write([]byte) (int, error) {
	w.attempts++
	return 0, w.err
}

func (w *responsesCancelWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.cancel()
	return n, err
}

func TestResponsesStreamClientCancelDoesNotAppendError(t *testing.T) {
	body := "data: {\"type\":\"response.in_progress\",\"response\":{\"usage\":{\"input_tokens\":10,\"output_tokens\":1}}}\n\n"
	c, w, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader(body)))
	ctx, cancel := context.WithCancel(c.Request.Context())
	t.Cleanup(cancel)
	c.Request = c.Request.WithContext(ctx)
	c.Writer = &responsesCancelWriter{ResponseWriter: c.Writer, cancel: cancel}
	usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 10, usage.PromptTokens)
	assert.Equal(t, relaycommon.StreamEndReasonClientGone, info.StreamStatus.EndReason)
	assert.NotContains(t, w.Body.String(), "event: error\n")
	assert.NotContains(t, w.Body.String(), "response.completed")
}

func TestResponsesStreamFailedTerminalWriteIsNotSuccessful(t *testing.T) {
	body := "data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":10,\"output_tokens\":1}}}\n\n"
	c, w, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader(body)))
	writeErr := fmt.Errorf("client socket closed")
	writer := &responsesDisconnectedWriter{ResponseWriter: c.Writer, err: writeErr}
	c.Writer = writer
	usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 10, usage.PromptTokens)
	assert.Equal(t, 1, usage.CompletionTokens)
	assert.Equal(t, relaycommon.StreamEndReasonHandlerStop, info.StreamStatus.EndReason)
	assert.ErrorIs(t, info.StreamStatus.EndError, writeErr)
	assert.True(t, info.StreamStatus.HasErrors())
	assert.Empty(t, w.Body.String())
	assert.Equal(t, 1, writer.attempts, "do not append an error to an unwritable connection")
}

func TestResponsesStreamCancelDuringTerminalFlushIsNotSuccessful(t *testing.T) {
	body := "data: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":10,\"output_tokens\":1}}}\n\n"
	c, w, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader(body)))
	ctx, cancel := context.WithCancel(c.Request.Context())
	t.Cleanup(cancel)
	c.Request = c.Request.WithContext(ctx)
	writer := &responsesCancelOnFlushWriter{ResponseWriter: c.Writer, cancel: cancel}
	c.Writer = writer
	usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 10, usage.PromptTokens)
	assert.Equal(t, 1, usage.CompletionTokens)
	assert.Equal(t, relaycommon.StreamEndReasonClientGone, info.StreamStatus.EndReason)
	assert.ErrorIs(t, info.StreamStatus.EndError, context.Canceled)
	assert.True(t, info.StreamStatus.HasErrors())
	assert.NotContains(t, w.Body.String(), "event: error\n")
	assert.Equal(t, 1, writer.flushes, "do not append another event after a failed flush")
}

func TestResponsesStreamFailureBeforeOutputCanRetry(t *testing.T) {
	for _, body := range []string{"", "data: malformed internal-provider.example\n\n"} {
		t.Run(body, func(t *testing.T) {
			c, w, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader(body)))
			usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
			require.NotNil(t, apiErr)
			assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
			assert.Nil(t, usage)
			assert.False(t, c.Writer.Written())
			assert.Empty(t, w.Body.String())
			assert.Empty(t, w.Header().Get("Content-Type"))
			_, headersSet := c.Get("event_stream_headers_set")
			assert.False(t, headersSet)
			assert.NotContains(t, apiErr.Error(), "internal-provider.example")
		})
	}
}

func TestResponsesStreamPreservesUsageAcrossZeroTerminalSnapshot(t *testing.T) {
	for _, terminalType := range []string{"response.completed", "response.incomplete"} {
		t.Run(terminalType, func(t *testing.T) {
			body := "data: {\"type\":\"response.in_progress\",\"response\":{\"usage\":{\"input_tokens\":10,\"output_tokens\":3,\"total_tokens\":13,\"input_tokens_details\":{\"cached_tokens\":8}}}}\n\n" +
				"data: {\"type\":\"" + terminalType + "\",\"response\":{\"usage\":{\"input_tokens\":0,\"output_tokens\":0,\"total_tokens\":0,\"input_tokens_details\":{\"cached_tokens\":0}}}}\n\n"
			c, _, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader(body)))
			usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
			assert.Equal(t, 10, usage.PromptTokens)
			assert.Equal(t, 3, usage.CompletionTokens)
			assert.Equal(t, 13, usage.TotalTokens)
			assert.Equal(t, 8, usage.PromptTokensDetails.CachedTokens)
		})
	}
}

func TestResponsesStreamZeroPlaceholderUsageAllowsTokenFallback(t *testing.T) {
	body := "data: {\"type\":\"response.created\",\"response\":{\"usage\":{\"input_tokens\":0,\"output_tokens\":0,\"total_tokens\":0}}}\n\n" +
		"data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial output\"}\n\n"
	c, _, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader(body)))
	usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Positive(t, usage.CompletionTokens)
}

func TestResponsesStreamRetainsInputUsageFromEarlierSnapshot(t *testing.T) {
	body := "data: {\"type\":\"response.in_progress\",\"response\":{\"usage\":{\"input_tokens\":10,\"output_tokens\":0,\"total_tokens\":10}}}\n\n" +
		"data: {\"type\":\"response.incomplete\",\"response\":{\"usage\":{\"input_tokens\":0,\"output_tokens\":3,\"total_tokens\":3}}}\n\n"
	c, _, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader(body)))
	usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 10, usage.PromptTokens)
	assert.Equal(t, 3, usage.CompletionTokens)
	assert.Equal(t, 13, usage.TotalTokens)
}

func TestResponsesStreamMalformedFrameAfterOutputEmitsPublicError(t *testing.T) {
	body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial output\"}\n\n" +
		"data: malformed internal-provider.example\n\n" +
		"data: {\"type\":\"response.completed\"}\n\n"
	c, w, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader(body)))
	usage, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Positive(t, usage.CompletionTokens)
	assert.Equal(t, relaycommon.StreamEndReasonHandlerStop, info.StreamStatus.EndReason)
	assert.Contains(t, w.Body.String(), `"code":"upstream_stream_incomplete"`)
	assert.NotContains(t, w.Body.String(), "internal-provider.example")
	assert.NotContains(t, w.Body.String(), "response.completed")
}

func TestResponsesStreamErrorSequenceDoesNotOverflow(t *testing.T) {
	body := fmt.Sprintf("data: {\"type\":\"response.in_progress\",\"sequence_number\":%d}\n\n", int64(math.MaxInt64))
	c, w, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader(body)))
	_, apiErr := OaiResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	parts := strings.Split(w.Body.String(), "data: ")
	require.Len(t, parts, 3)
	var event struct {
		Type           string `json:"type"`
		SequenceNumber int64  `json:"sequence_number"`
	}
	require.NoError(t, common.UnmarshalJsonStr(strings.TrimSpace(parts[2]), &event))
	assert.Equal(t, "error", event.Type)
	assert.GreaterOrEqual(t, event.SequenceNumber, int64(0))
}
