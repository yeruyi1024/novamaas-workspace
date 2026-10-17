package helper

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"testing/synctest"
	"time"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
	if constant.StreamingTimeout == 0 {
		constant.StreamingTimeout = 30
	}
}

func setupStreamTest(t *testing.T, body io.Reader) (*gin.Context, *http.Response, *relaycommon.RelayInfo) {
	t.Helper()

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	resp := &http.Response{
		Body: io.NopCloser(body),
	}

	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{},
	}

	return c, resp, info
}

func buildSSEBody(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "data: {\"id\":%d,\"choices\":[{\"delta\":{\"content\":\"token_%d\"}}]}\n", i, i)
	}
	b.WriteString("data: [DONE]\n")
	return b.String()
}

// ---------- Basic correctness ----------

func TestStreamScannerHandler_NilInputs(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{}}

	StreamScannerHandler(c, nil, info, func(data string, sr *StreamResult) {})
	StreamScannerHandler(c, &http.Response{Body: io.NopCloser(strings.NewReader(""))}, info, nil)
}

func TestNewStreamScanner_AllowsLargeStreamLine(t *testing.T) {
	oldBufferMB := constant.StreamScannerMaxBufferMB
	constant.StreamScannerMaxBufferMB = 1
	t.Cleanup(func() {
		constant.StreamScannerMaxBufferMB = oldBufferMB
	})

	payload := strings.Repeat("x", 128<<10)
	scanner := NewStreamScanner(strings.NewReader("data: " + payload + "\n"))
	scanner.Split(bufio.ScanLines)

	require.True(t, scanner.Scan())
	assert.Equal(t, "data: "+payload, scanner.Text())
	require.NoError(t, scanner.Err())
}

func TestStreamScannerHandler_EmptyBody(t *testing.T) {
	t.Parallel()

	c, resp, info := setupStreamTest(t, strings.NewReader(""))

	var called atomic.Bool
	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		called.Store(true)
	})

	assert.False(t, called.Load(), "handler should not be called for empty body")
}

func TestStreamScannerHandler_1000Chunks(t *testing.T) {
	t.Parallel()

	const numChunks = 1000
	body := buildSSEBody(numChunks)
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	var count atomic.Int64
	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		count.Add(1)
	})

	assert.Equal(t, int64(numChunks), count.Load())
	assert.Equal(t, numChunks, info.ReceivedResponseCount)
}

func TestStreamScannerHandler_OrderPreserved(t *testing.T) {
	t.Parallel()

	const numChunks = 500
	body := buildSSEBody(numChunks)
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	var mu sync.Mutex
	received := make([]string, 0, numChunks)

	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		mu.Lock()
		received = append(received, data)
		mu.Unlock()
	})

	require.Equal(t, numChunks, len(received))
	for i := 0; i < numChunks; i++ {
		expected := fmt.Sprintf("{\"id\":%d,\"choices\":[{\"delta\":{\"content\":\"token_%d\"}}]}", i, i)
		assert.Equal(t, expected, received[i], "chunk %d out of order", i)
	}
}

func TestStreamScannerHandler_DoneStopsScanner(t *testing.T) {
	t.Parallel()

	body := buildSSEBody(50) + "data: should_not_appear\n"
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	var count atomic.Int64
	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		count.Add(1)
	})

	assert.Equal(t, int64(50), count.Load(), "data after [DONE] must not be processed")
}

func TestStreamScannerHandler_StopStopsStream(t *testing.T) {
	t.Parallel()

	const numChunks = 200
	body := buildSSEBody(numChunks)
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	const stopAt int64 = 50
	var count atomic.Int64
	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		n := count.Add(1)
		if n >= stopAt {
			sr.Stop(fmt.Errorf("fatal at %d", n))
		}
	})

	assert.Equal(t, stopAt, count.Load())
	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonHandlerStop, info.StreamStatus.EndReason)
}

func TestStreamScannerHandler_SkipsNonDataLines(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	b.WriteString(": comment line\n")
	b.WriteString("event: message\n")
	b.WriteString("id: 12345\n")
	b.WriteString("retry: 5000\n")
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&b, "data: payload_%d\n", i)
		b.WriteString(": interleaved comment\n")
	}
	b.WriteString("data: [DONE]\n")

	c, resp, info := setupStreamTest(t, strings.NewReader(b.String()))

	var count atomic.Int64
	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		count.Add(1)
	})

	assert.Equal(t, int64(100), count.Load())
}

func TestStreamScannerHandler_DataWithExtraSpaces(t *testing.T) {
	t.Parallel()

	body := "data:   {\"trimmed\":true}  \ndata: [DONE]\n"
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	var got string
	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		got = data
	})

	assert.Equal(t, "{\"trimmed\":true}", got)
}

// TestStreamScannerHandler_ClientCancelAbortsUpstreamAndReturns pins the
// disconnect contract: when the client goes away, the handler must return
// promptly (all goroutines joined, so the gin.Context can never leak into a
// pooled reuse), the upstream body must be closed to stop token generation,
// and no data received after the disconnect may be processed or written.
func TestStreamScannerHandler_ClientCancelAbortsUpstreamAndReturns(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pr, pw := io.Pipe()
	t.Cleanup(func() {
		_ = pr.Close()
		_ = pw.Close()
	})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil).WithContext(ctx)

	resp := &http.Response{Body: pr}
	info := &relaycommon.RelayInfo{
		DisablePing: true,
		ChannelMeta: &relaycommon.ChannelMeta{},
	}

	var count atomic.Int64
	firstHandled := make(chan struct{})
	done := make(chan struct{})
	go func() {
		StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
			count.Add(1)
			_ = StringData(c, data)
			if data == "first" {
				close(firstHandled)
			}
		})
		close(done)
	}()

	_, err := fmt.Fprint(pw, "data: first\n")
	require.NoError(t, err)

	select {
	case <-firstHandled:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first chunk")
	}

	cancel()

	// The handler must return without any further upstream input: cleanup
	// closes resp.Body, which unblocks the scanner goroutine.
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after client disconnect")
	}

	// Upstream read side must be closed so the provider stops generating
	// (and billing) for a request nobody is listening to.
	_, err = fmt.Fprint(pw, "data: second\n")
	require.ErrorIs(t, err, io.ErrClosedPipe, "upstream body should be closed after client disconnect")

	assert.Equal(t, int64(1), count.Load(), "no chunk after disconnect should be processed")
	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonClientGone, info.StreamStatus.EndReason)

	body := recorder.Body.String()
	assert.Contains(t, body, "first")
	assert.NotContains(t, body, "second")
}

// ---------- Ping tests ----------

func TestStreamScannerHandler_PingSentDuringSlowUpstream(t *testing.T) {
	setting := operation_setting.GetGeneralSetting()
	oldEnabled := setting.PingIntervalEnabled
	oldSeconds := setting.PingIntervalSeconds
	setting.PingIntervalEnabled = true
	setting.PingIntervalSeconds = 1
	t.Cleanup(func() {
		setting.PingIntervalEnabled = oldEnabled
		setting.PingIntervalSeconds = oldSeconds
	})

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		for i := 0; i < 4; i++ {
			fmt.Fprintf(pw, "data: chunk_%d\n", i)
			time.Sleep(400 * time.Millisecond)
		}
		fmt.Fprint(pw, "data: [DONE]\n")
	}()

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	resp := &http.Response{Body: pr}
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{}}

	var count atomic.Int64
	done := make(chan struct{})
	go func() {
		StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
			count.Add(1)
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for stream to finish")
	}

	assert.Equal(t, int64(4), count.Load())

	body := recorder.Body.String()
	pingCount := strings.Count(body, ": PING")
	assert.GreaterOrEqual(t, pingCount, 1,
		"expected at least 1 ping during slow stream with 1s interval; got %d", pingCount)
}

func TestStreamScannerHandler_PingDisabledByRelayInfo(t *testing.T) {
	setting := operation_setting.GetGeneralSetting()
	oldEnabled := setting.PingIntervalEnabled
	oldSeconds := setting.PingIntervalSeconds
	setting.PingIntervalEnabled = true
	setting.PingIntervalSeconds = 1
	t.Cleanup(func() {
		setting.PingIntervalEnabled = oldEnabled
		setting.PingIntervalSeconds = oldSeconds
	})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	resp := &http.Response{Body: io.NopCloser(strings.NewReader(buildSSEBody(5)))}
	info := &relaycommon.RelayInfo{
		DisablePing: true,
		ChannelMeta: &relaycommon.ChannelMeta{},
	}

	var count atomic.Int64
	done := make(chan struct{})
	go func() {
		StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
			count.Add(1)
		})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out")
	}

	assert.Equal(t, int64(5), count.Load())

	body := recorder.Body.String()
	pingCount := strings.Count(body, ": PING")
	assert.Equal(t, 0, pingCount, "pings should be disabled when DisablePing=true")
}

// ---------- StreamStatus integration ----------

func TestStreamScannerHandler_StreamStatus_DoneReason(t *testing.T) {
	t.Parallel()

	body := buildSSEBody(10)
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {})

	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonDone, info.StreamStatus.EndReason)
	assert.Nil(t, info.StreamStatus.EndError)
	assert.True(t, info.StreamStatus.IsNormalEnd())
	assert.False(t, info.StreamStatus.HasErrors())
}

func TestStreamScannerHandler_StreamStatus_EOFWithoutDone(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	for i := 0; i < 5; i++ {
		fmt.Fprintf(&b, "data: {\"id\":%d}\n", i)
	}
	c, resp, info := setupStreamTest(t, strings.NewReader(b.String()))

	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {})

	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonEOF, info.StreamStatus.EndReason)
	assert.True(t, info.StreamStatus.IsNormalEnd())
}

func TestStreamScannerHandler_StreamStatus_HandlerStop(t *testing.T) {
	t.Parallel()

	body := buildSSEBody(100)
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	var count atomic.Int64
	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		n := count.Add(1)
		if n >= 10 {
			sr.Stop(fmt.Errorf("stop at 10"))
		}
	})

	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonHandlerStop, info.StreamStatus.EndReason)
	assert.True(t, info.StreamStatus.HasErrors())
}

func TestStreamScannerHandler_StreamStatus_HandlerDone(t *testing.T) {
	t.Parallel()

	body := buildSSEBody(20)
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	var count atomic.Int64
	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		n := count.Add(1)
		if n >= 5 {
			sr.Done()
		}
	})

	assert.Equal(t, int64(5), count.Load())
	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonDone, info.StreamStatus.EndReason)
	assert.False(t, info.StreamStatus.HasErrors())
}

func TestStreamScannerHandler_StreamStatus_Timeout(t *testing.T) {
	// Not parallel: modifies global constant.StreamingTimeout
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 1
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	pr, pw := io.Pipe()
	go func() {
		fmt.Fprint(pw, "data: {\"id\":1}\n")
		time.Sleep(2 * time.Second)
		pw.Close()
	}()

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	resp := &http.Response{Body: pr}
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{}}

	done := make(chan struct{})
	go func() {
		StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for stream timeout")
	}

	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonTimeout, info.StreamStatus.EndReason)
	assert.False(t, info.StreamStatus.IsNormalEnd())
}

func TestStreamScannerHandler_StreamStatus_SoftErrors(t *testing.T) {
	t.Parallel()

	body := buildSSEBody(10)
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		sr.Error(fmt.Errorf("soft error for chunk"))
	})

	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonDone, info.StreamStatus.EndReason)
	assert.True(t, info.StreamStatus.HasErrors())
	assert.Equal(t, 10, info.StreamStatus.TotalErrorCount())
}

func TestStreamScannerHandler_StreamStatus_MultipleErrorsPerChunk(t *testing.T) {
	t.Parallel()

	body := buildSSEBody(5)
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		sr.Error(fmt.Errorf("error A"))
		sr.Error(fmt.Errorf("error B"))
	})

	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonDone, info.StreamStatus.EndReason)
	assert.Equal(t, 10, info.StreamStatus.TotalErrorCount())
}

func TestStreamScannerHandler_StreamStatus_ErrorThenStop(t *testing.T) {
	t.Parallel()

	// A fatal handler result takes precedence over buffered upstream events.
	var b strings.Builder
	for i := 0; i < 100; i++ {
		fmt.Fprintf(&b, "data: {\"id\":%d}\n", i)
	}
	c, resp, info := setupStreamTest(t, strings.NewReader(b.String()))

	var count atomic.Int64
	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
		count.Add(1)
		sr.Error(fmt.Errorf("soft error"))
		sr.Stop(fmt.Errorf("fatal"))
	})

	assert.Equal(t, int64(1), count.Load())
	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonHandlerStop, info.StreamStatus.EndReason)
	assert.Equal(t, 2, info.StreamStatus.TotalErrorCount())
}

func TestStreamScannerHandler_StreamStatus_InitializedIfNil(t *testing.T) {
	t.Parallel()

	body := buildSSEBody(1)
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	assert.Nil(t, info.StreamStatus)

	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {})

	assert.NotNil(t, info.StreamStatus)
}

func TestStreamScannerHandler_StreamStatus_ReplacesPreInitialized(t *testing.T) {
	t.Parallel()

	body := buildSSEBody(5)
	c, resp, info := setupStreamTest(t, strings.NewReader(body))

	info.StreamStatus = relaycommon.NewStreamStatus()
	info.StreamStatus.RecordError("pre-existing error")

	StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {})

	assert.Equal(t, relaycommon.StreamEndReasonDone, info.StreamStatus.EndReason)
	assert.Equal(t, 0, info.StreamStatus.TotalErrorCount())
}

func TestStreamScannerHandler_DrainsEventsBeforeTransportResult(t *testing.T) {
	for _, test := range []struct {
		name     string
		tail     string
		readErr  error
		terminal bool
		reason   relaycommon.StreamEndReason
	}{
		{name: "done_after_final_event", tail: "data: [DONE]\n", reason: relaycommon.StreamEndReasonDone},
		{name: "clean_eof", reason: relaycommon.StreamEndReasonEOF},
		{name: "truncated_stream", readErr: io.ErrUnexpectedEOF, reason: relaycommon.StreamEndReasonScannerErr},
		{name: "terminal_before_read_error", readErr: io.ErrUnexpectedEOF, terminal: true, reason: relaycommon.StreamEndReasonDone},
	} {
		t.Run(test.name, func(t *testing.T) {
			body := "data: first\n\ndata: last\n\n" + test.tail
			var reader io.Reader = strings.NewReader(body)
			if test.readErr != nil {
				reader = io.MultiReader(reader, iotest.ErrReader(test.readErr))
			}
			c, resp, info := setupStreamTest(t, reader)
			var received []string
			StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
				received = append(received, data)
				if data == "last" && test.terminal {
					sr.Done()
				}
			})
			assert.Equal(t, []string{"first", "last"}, received)
			require.NotNil(t, info.StreamStatus)
			assert.Equal(t, test.reason, info.StreamStatus.EndReason)
			if test.reason == relaycommon.StreamEndReasonScannerErr {
				assert.ErrorIs(t, info.StreamStatus.EndError, test.readErr)
			} else {
				assert.NoError(t, info.StreamStatus.EndError)
			}
		})
	}
}

type observedStreamBody struct {
	io.ReadCloser
	closed chan struct{}
	once   sync.Once
}

func (b *observedStreamBody) Close() error {
	err := b.ReadCloser.Close()
	b.once.Do(func() { close(b.closed) })
	return err
}

func TestStreamScannerHandler_PingWriteFailureClosesUpstream(t *testing.T) {
	setting := operation_setting.GetGeneralSetting()
	oldEnabled, oldSeconds := setting.PingIntervalEnabled, setting.PingIntervalSeconds
	oldTimeout := constant.StreamingTimeout
	setting.PingIntervalEnabled, setting.PingIntervalSeconds = true, 1
	constant.StreamingTimeout = 30
	t.Cleanup(func() {
		setting.PingIntervalEnabled, setting.PingIntervalSeconds = oldEnabled, oldSeconds
		constant.StreamingTimeout = oldTimeout
	})
	reader, writer := io.Pipe()
	t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() })
	c, resp, info := setupStreamTest(t, reader)
	body := &observedStreamBody{ReadCloser: reader, closed: make(chan struct{})}
	resp.Body = body
	writeErr := errors.New("client socket disconnected")
	failingWriter := &responsesFailingWriter{ResponseWriter: c.Writer, err: writeErr}
	c.Writer = failingWriter
	done := make(chan struct{})
	go func() {
		StreamScannerHandler(c, resp, info, func(string, *StreamResult) {})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = writer.Close()
		<-done
		t.Fatal("failed heartbeat did not stop the stream before the upstream idle timeout")
	}
	assert.Equal(t, relaycommon.StreamEndReasonPingFail, info.StreamStatus.EndReason)
	assert.ErrorIs(t, info.StreamStatus.EndError, writeErr)
	assert.Equal(t, 1, failingWriter.writes)
	select {
	case <-body.closed:
	default:
		t.Fatal("failed heartbeat left the upstream connection open")
	}
}

func TestStreamScannerHandler_ForwardedTerminalWinsConcurrentClientCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pr, pw := io.Pipe()
	t.Cleanup(func() {
		_ = pr.Close()
		_ = pw.Close()
	})
	c, resp, info := setupStreamTest(t, pr)
	c.Request = c.Request.WithContext(ctx)
	body := &observedStreamBody{ReadCloser: pr, closed: make(chan struct{})}
	resp.Body = body
	forwarded := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	defer close(release)
	go func() {
		StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
			_ = StringData(c, data)
			close(forwarded)
			<-release
			sr.Done()
		})
		close(done)
	}()

	_, err := fmt.Fprint(pw, "data: completed\n")
	require.NoError(t, err)
	select {
	case <-forwarded:
	case <-time.After(2 * time.Second):
		t.Fatal("terminal event was not forwarded")
	}
	cancel()
	select {
	case <-body.closed:
	case <-time.After(2 * time.Second):
		t.Fatal("client cancellation did not close the upstream")
	}
	release <- struct{}{}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not finish after its terminal callback")
	}

	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonDone, info.StreamStatus.EndReason)
	assert.NoError(t, info.StreamStatus.EndError)
}

func TestStreamScannerHandler_ClientCancelDiscardsBufferedEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pr, pw := io.Pipe()
	t.Cleanup(func() {
		_ = pr.Close()
		_ = pw.Close()
	})
	c, resp, info := setupStreamTest(t, pr)
	body := &observedStreamBody{ReadCloser: pr, closed: make(chan struct{})}
	resp.Body = body
	c.Request = c.Request.WithContext(ctx)
	firstHandled := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	defer close(release)
	var received []string
	go func() {
		StreamScannerHandler(c, resp, info, func(data string, sr *StreamResult) {
			received = append(received, data)
			if data == "first" {
				close(firstHandled)
				<-release
			}
		})
		close(done)
	}()

	_, err := fmt.Fprint(pw, "data: first\n")
	require.NoError(t, err)
	select {
	case <-firstHandled:
	case <-time.After(2 * time.Second):
		t.Fatal("first event was not handled")
	}
	// The pipe write completes only after Scanner has read these bytes. They
	// remain buffered while the first callback is deliberately in flight.
	_, err = fmt.Fprint(pw, "data: buffered\n\ndata: completed\n\n")
	require.NoError(t, err)
	// A further read proves both preceding events have entered dataChan.
	_, err = fmt.Fprint(pw, ": queued\n")
	require.NoError(t, err)
	cancel()
	select {
	case <-body.closed:
	case <-time.After(2 * time.Second):
		t.Fatal("client cancellation did not close the upstream")
	}
	release <- struct{}{}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not finish after client cancellation")
	}

	assert.Equal(t, []string{"first"}, received)
	require.NotNil(t, info.StreamStatus)
	assert.Equal(t, relaycommon.StreamEndReasonClientGone, info.StreamStatus.EndReason)
	assert.ErrorIs(t, info.StreamStatus.EndError, context.Canceled)
}

func TestWaitForStreamEnd_FinishedUpstreamDisablesIdleTimeoutWhileDraining(t *testing.T) {
	for _, test := range []struct {
		name            string
		deadlineExpired bool
		cancelDrain     bool
	}{
		{name: "handler_completes"},
		{name: "deadline_and_finished_reader_both_ready", deadlineExpired: true},
		{name: "client_cancels", cancelDrain: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			// Keep this bubble limited to the waiting logic. StreamScannerHandler
			// uses global goroutine pools whose workers may live outside a bubble.
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()
				if test.deadlineExpired {
					// Make the deadline ready before delivering the read-complete signal.
					time.Sleep(time.Second)
				}
				scannerFinished := make(chan struct{})
				close(scannerFinished)
				handlerStopped := make(chan bool)
				var reason relaycommon.StreamEndReason
				var endError error
				done := make(chan struct{})
				go func() {
					reason, endError = waitForStreamEnd(ctx, ticker, scannerFinished, handlerStopped, &streamReadResult{})
					close(done)
				}()
				synctest.Wait()
				// Only the virtual clock advances; buffered output remains pending
				// beyond the former upstream idle deadline without a wall-clock wait.
				time.Sleep(2 * time.Second)
				synctest.Wait()
				select {
				case <-done:
					require.FailNowf(t, "stream stopped before handler completion or client cancellation", "reason=%s", reason)
				default:
				}
				if test.cancelDrain {
					cancel()
				} else {
					close(handlerStopped)
				}
				<-done
				if test.cancelDrain {
					assert.Equal(t, relaycommon.StreamEndReasonClientGone, reason)
					assert.ErrorIs(t, endError, context.Canceled)
				} else {
					assert.Equal(t, relaycommon.StreamEndReasonNone, reason)
					assert.NoError(t, endError)
				}
			})
		})
	}
}

func TestWaitForStreamEnd_UnfinishedUpstreamStillTimesOut(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		reason, err := waitForStreamEnd(context.Background(), ticker, make(chan struct{}), make(chan bool), &streamReadResult{})
		assert.Equal(t, relaycommon.StreamEndReasonTimeout, reason)
		assert.NoError(t, err)
	})
}

func TestWaitForStreamEnd_ReadFailurePrecedesLaterClientCancel(t *testing.T) {
	for _, tc := range []struct {
		name           string
		readerFinished bool
		readError      error
		wantReason     relaycommon.StreamEndReason
		stopHandler    bool
	}{
		{"failed reader", true, io.ErrUnexpectedEOF, relaycommon.StreamEndReasonScannerErr, false},
		{"failed reader and stopped handler", true, io.ErrUnexpectedEOF, relaycommon.StreamEndReasonScannerErr, true},
		{"finished reader with pending output", true, nil, relaycommon.StreamEndReasonClientGone, false},
		{"reader still active", false, nil, relaycommon.StreamEndReasonClientGone, false},
		{"unpublished read error", false, io.ErrUnexpectedEOF, relaycommon.StreamEndReasonClientGone, false},
		{"request cancellation reached reader", true, fmt.Errorf("read stopped: %w", context.Canceled), relaycommon.StreamEndReasonClientGone, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()
				finished := make(chan struct{})
				readResult := &streamReadResult{reason: relaycommon.StreamEndReasonEOF}
				if tc.readError != nil {
					readResult.reason, readResult.err = relaycommon.StreamEndReasonScannerErr, tc.readError
				}
				if tc.readerFinished {
					close(finished)
				}
				var reason relaycommon.StreamEndReason
				var endError error
				done := make(chan struct{})
				handlerStopped := make(chan bool)
				go func() {
					reason, endError = waitForStreamEnd(ctx, ticker, finished, handlerStopped, readResult)
					close(done)
				}()
				synctest.Wait()
				cancel()
				if tc.stopHandler {
					close(handlerStopped)
				}
				<-done
				assert.Equal(t, tc.wantReason, reason)
				if tc.wantReason == relaycommon.StreamEndReasonScannerErr {
					assert.ErrorIs(t, endError, tc.readError)
				} else {
					assert.ErrorIs(t, endError, context.Canceled)
				}
			})
		})
	}
}
