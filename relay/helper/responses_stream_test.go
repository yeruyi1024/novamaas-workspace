package helper

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type responsesFailingWriter struct {
	gin.ResponseWriter
	err    error
	writes int
}

type responsesCancelOnFlushWriter struct {
	gin.ResponseWriter
	cancel context.CancelFunc
}

func (w *responsesCancelOnFlushWriter) Flush() {
	w.ResponseWriter.Flush()
	w.cancel()
}

func (w *responsesFailingWriter) Write([]byte) (int, error) {
	w.writes++
	return 0, w.err
}

func TestResponseChunkDataPropagatesWriteError(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	writeErr := errors.New("client disconnected")
	writer := &responsesFailingWriter{ResponseWriter: c.Writer, err: writeErr}
	c.Writer = writer
	err := ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.output_text.delta"}, `{"type":"response.output_text.delta","delta":"hello"}`)
	assert.ErrorIs(t, err, writeErr)
	assert.Equal(t, 1, writer.writes)
}

func TestResponseChunkDataReportsCancelDuringFlush(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil).WithContext(ctx)
	c.Writer = &responsesCancelOnFlushWriter{ResponseWriter: c.Writer, cancel: cancel}
	err := ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.completed"}, `{"type":"response.completed"}`)
	assert.ErrorIs(t, err, context.Canceled)
	assert.True(t, w.Flushed)
	assert.Contains(t, w.Body.String(), "event: response.completed\n")
}

func TestStringDataPropagatesClientWriteFailure(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	writeErr := errors.New("client disconnected")
	writer := &responsesFailingWriter{ResponseWriter: c.Writer, err: writeErr}
	c.Writer = writer
	assert.ErrorIs(t, StringData(c, `{"choices":[]}`), writeErr)
	assert.Equal(t, 1, writer.writes)
}

func TestStringDataReportsCancelDuringFlush(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil).WithContext(ctx)
	c.Writer = &responsesCancelOnFlushWriter{ResponseWriter: c.Writer, cancel: cancel}
	assert.ErrorIs(t, StringData(c, "[DONE]"), context.Canceled)
}

func TestStringDataPreservesEventFraming(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	require.NoError(t, StringData(c, "first\rsecond"))
	assert.Equal(t, "data: first\\rsecond\n\n", w.Body.String())
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
}

func TestClaudeDataPreservesEventFramingAndReportsFlushCancellation(t *testing.T) {
	for _, cancelDuringFlush := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal", true: "cancel during flush"}[cancelDuringFlush], func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil).WithContext(ctx)
			if cancelDuringFlush {
				c.Writer = &responsesCancelOnFlushWriter{ResponseWriter: c.Writer, cancel: cancel}
			}
			err := ClaudeData(c, dto.ClaudeResponse{Type: "message_stop"})
			if cancelDuringFlush {
				assert.ErrorIs(t, err, context.Canceled)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
			assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
			assert.True(t, strings.HasPrefix(w.Body.String(), "event: message_stop\ndata: "))
			assert.Equal(t, 1, strings.Count(w.Body.String(), "\n\n"))
			var payload dto.ClaudeResponse
			require.NoError(t, common.UnmarshalJsonStr(strings.TrimSpace(strings.TrimPrefix(w.Body.String(), "event: message_stop\ndata: ")), &payload))
			assert.Equal(t, "message_stop", payload.Type)
		})
	}
}

func TestWriteResponsesStreamErrorUsesResponsesSchema(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	require.NoError(t, WriteResponsesStreamError(c, "upstream_stream_incomplete", "The upstream response stream ended before completion.", 9))
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.True(t, strings.HasPrefix(w.Body.String(), "event: error\ndata: "))
	assert.True(t, strings.HasSuffix(w.Body.String(), "\n\n"))
	var payload map[string]any
	require.NoError(t, common.UnmarshalJsonStr(strings.TrimSpace(strings.TrimPrefix(w.Body.String(), "event: error\ndata: ")), &payload))
	assert.Equal(t, map[string]any{
		"type":            "error",
		"code":            "upstream_stream_incomplete",
		"message":         "The upstream response stream ended before completion.",
		"param":           nil,
		"sequence_number": float64(9),
	}, payload)
}

func TestWriteResponsesStreamErrorSkipsCanceledClient(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil).WithContext(ctx)
	err := WriteResponsesStreamError(c, "upstream_stream_incomplete", "The upstream response stream ended before completion.", 9)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, w.Body.String())
}

func TestResetEventStreamHeadersPreservesCommittedStream(t *testing.T) {
	for _, committed := range []bool{false, true} {
		t.Run(map[bool]string{false: "unstarted", true: "committed"}[committed], func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			c.Writer.Header().Set("X-Request-Id", "request-123")
			SetEventStreamHeaders(c)
			if committed {
				require.NoError(t, ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "response.created"}, `{"type":"response.created"}`))
			}
			ResetEventStreamHeaders(c)
			assert.Equal(t, "request-123", w.Header().Get("X-Request-Id"))
			_, headersSet := c.Get("event_stream_headers_set")
			assert.Equal(t, committed, headersSet)
			if committed {
				assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
				return
			}
			for _, name := range []string{"Content-Type", "Cache-Control", "Connection", "Transfer-Encoding", "X-Accel-Buffering"} {
				assert.Empty(t, w.Header().Get(name))
			}
			SetEventStreamHeaders(c)
			assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
		})
	}
}
