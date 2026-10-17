package helper

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/bytedance/gopkg/util/gopool"

	"github.com/gin-gonic/gin"
)

type streamReadResult struct {
	reason relaycommon.StreamEndReason
	err    error
}

const (
	InitialScannerBufferSize    = 64 << 10  // 64KB (64*1024)
	DefaultMaxScannerBufferSize = 128 << 20 // 64MB (64*1024*1024) default SSE buffer size
	DefaultPingInterval         = 10 * time.Second
	// streamWriteTimeout bounds a single blocked write to a slow client so the
	// unconditional wg.Wait() in cleanup can always finish. Without it, a slow
	// but connected client (full TCP buffer, no server WriteTimeout) could hang
	// the handler forever.
	streamWriteTimeout = 30 * time.Second
)

func getScannerBufferSize() int {
	if constant.StreamScannerMaxBufferMB > 0 {
		return constant.StreamScannerMaxBufferMB << 20
	}
	return DefaultMaxScannerBufferSize
}

func NewStreamScanner(reader io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, InitialScannerBufferSize), getScannerBufferSize())
	return scanner
}

func copyCodexSSEHeaders(c *gin.Context, resp *http.Response) {
	if c == nil || c.Writer == nil || resp == nil {
		return
	}
	// codex
	for _, name := range []string{"X-Reasoning-Included", "X-Codex-Turn-State"} {
		values := resp.Header.Values(name)
		if !service.ShouldCopyUpstreamHeader(c, name, values) {
			continue
		}
		for _, value := range values {
			if value != "" {
				c.Writer.Header().Add(name, value)
			}
		}
	}
}

// ExtendWriteDeadline pushes the connection write deadline forward before each
// stream write. Best-effort: writers that don't support deadlines (e.g.
// httptest recorders) are silently ignored.
func ExtendWriteDeadline(c *gin.Context) {
	if c == nil || c.Writer == nil {
		return
	}
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Now().Add(streamWriteTimeout))
}

func StreamScannerHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, dataHandler func(data string, sr *StreamResult)) {

	if resp == nil || dataHandler == nil {
		return
	}

	// 无条件新建 StreamStatus
	info.StreamStatus = relaycommon.NewStreamStatus()

	ctx, cancel := context.WithCancel(context.Background())

	streamingTimeout := time.Duration(constant.StreamingTimeout) * time.Second

	var (
		stopChan    = make(chan bool, 3) // 增加缓冲区避免阻塞
		scanner     = NewStreamScanner(resp.Body)
		ticker      = time.NewTicker(streamingTimeout)
		pingTicker  *time.Ticker
		writeMutex  sync.Mutex     // Mutex to protect concurrent writes
		wg          sync.WaitGroup // 用于等待所有 goroutine 退出
		cleanupOnce sync.Once
		stopOnce    sync.Once
	)

	stop := func() {
		stopOnce.Do(func() {
			close(stopChan)
		})
	}

	generalSettings := operation_setting.GetGeneralSetting()
	pingEnabled := generalSettings.PingIntervalEnabled && !info.DisablePing
	pingInterval := time.Duration(generalSettings.PingIntervalSeconds) * time.Second
	if pingInterval <= 0 {
		pingInterval = DefaultPingInterval
	}

	if pingEnabled {
		pingTicker = time.NewTicker(pingInterval)
	}

	logger.LogDebug(c, "relay timeout seconds: %d", common.RelayTimeout)
	logger.LogDebug(c, "relay max idle conns: %d", common.RelayMaxIdleConns)
	logger.LogDebug(c, "relay max idle conns per host: %d", common.RelayMaxIdleConnsPerHost)
	logger.LogDebug(c, "streaming timeout seconds: %d", int64(streamingTimeout.Seconds()))
	logger.LogDebug(c, "ping interval seconds: %d", int64(pingInterval.Seconds()))

	cleanup := func() {
		cleanupOnce.Do(func() {
			cancel()
			stop()
			if resp.Body != nil {
				_ = resp.Body.Close()
			}

			ticker.Stop()
			if pingTicker != nil {
				pingTicker.Stop()
			}

			wg.Wait()
		})
	}
	// Ensure gin.Context is not returned to Gin's pool while any stream goroutine can still use it.
	defer cleanup()

	scanner.Split(bufio.ScanLines)
	copyCodexSSEHeaders(c, resp)
	SetEventStreamHeaders(c)

	ctx = context.WithValue(ctx, "stop_chan", stopChan)

	// Handle ping data sending with improved error handling
	if pingEnabled && pingTicker != nil {
		wg.Add(1)
		gopool.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					logger.LogError(c, fmt.Sprintf("ping goroutine panic: %v", r))
					info.StreamStatus.SetEndReason(relaycommon.StreamEndReasonPanic, fmt.Errorf("ping panic: %v", r))
					stop()
				}
				logger.LogDebug(c, "ping goroutine exited")
				wg.Done()
			}()

			// 添加超时保护，防止 goroutine 无限运行
			maxPingDuration := 30 * time.Minute // 最大 ping 持续时间
			pingTimeout := time.NewTimer(maxPingDuration)
			defer pingTimeout.Stop()

			for {
				select {
				case <-pingTicker.C:
					var err error
					func() {
						writeMutex.Lock()
						defer writeMutex.Unlock()
						if ctx.Err() != nil || c.Request.Context().Err() != nil {
							return
						}
						ExtendWriteDeadline(c)
						err = PingData(c)
					}()
					if err != nil {
						logger.LogError(c, "ping data error: "+err.Error())
						info.StreamStatus.SetEndReason(relaycommon.StreamEndReasonPingFail, err)
						stop()
						return
					}
					logger.LogDebug(c, "ping data sent")
				case <-ctx.Done():
					return
				case <-stopChan:
					return
				case <-c.Request.Context().Done():
					// 监听客户端断开连接
					return
				case <-pingTimeout.C:
					logger.LogError(c, "ping goroutine max duration reached")
					return
				}
			}
		})
	}

	dataChan := make(chan string, 10)
	scannerFinished := make(chan struct{})
	// Closing dataChan publishes the transport result. The handler applies it
	// after buffered events, which may contain a protocol terminal, are consumed.
	scannerResult := &streamReadResult{reason: relaycommon.StreamEndReasonEOF}

	wg.Add(1)
	gopool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				logger.LogError(c, fmt.Sprintf("data handler goroutine panic: %v", r))
				info.StreamStatus.SetEndReason(relaycommon.StreamEndReasonPanic, fmt.Errorf("handler panic: %v", r))
			}
			stop()
			wg.Done()
		}()
		sr := newStreamResult(info.StreamStatus)
		for data := range dataChan {
			sr.reset()
			func() {
				writeMutex.Lock()
				defer writeMutex.Unlock()
				// A closed channel can still contain buffered events. Abort them
				// when the client leaves instead of forwarding abandoned output.
				if ctx.Err() != nil || c.Request.Context().Err() != nil {
					return
				}
				ExtendWriteDeadline(c)
				dataHandler(data, sr)
			}()
			if sr.IsStopped() || ctx.Err() != nil || c.Request.Context().Err() != nil {
				return
			}
		}
		if ctx.Err() != nil || c.Request.Context().Err() != nil {
			return
		}
		if scannerResult.reason == relaycommon.StreamEndReasonScannerErr {
			logger.LogError(c, "scanner error: "+scannerResult.err.Error())
		}
		info.StreamStatus.SetEndReason(scannerResult.reason, scannerResult.err)
	})

	// Scanner goroutine with improved error handling
	wg.Add(1)
	common.RelayCtxGo(ctx, func() {
		defer func() {
			if r := recover(); r != nil {
				logger.LogError(c, fmt.Sprintf("scanner goroutine panic: %v", r))
				scannerResult.reason = relaycommon.StreamEndReasonPanic
				scannerResult.err = fmt.Errorf("scanner panic: %v", r)
			}
			close(dataChan)
			close(scannerFinished)
			logger.LogDebug(c, "scanner goroutine exited")
			wg.Done()
		}()

		for scanner.Scan() {
			// 检查是否需要停止
			select {
			case <-stopChan:
				return
			case <-ctx.Done():
				return
			default:
			}

			ticker.Reset(streamingTimeout)
			data := scanner.Text()
			logger.LogDebug(c, "stream scanner data: %s", data)

			if len(data) < 6 {
				continue
			}
			if data[:5] != "data:" && data[:6] != "[DONE]" {
				continue
			}
			data = data[5:]
			data = strings.TrimSpace(data)
			if data == "" {
				continue
			}
			if !strings.HasPrefix(data, "[DONE]") {
				info.SetFirstResponseTime()
				info.ReceivedResponseCount++

				select {
				case dataChan <- data:
				case <-ctx.Done():
					return
				case <-stopChan:
					return
				}
			} else {
				scannerResult.reason = relaycommon.StreamEndReasonDone
				logger.LogDebug(c, "received [DONE], stopping scanner")
				return
			}
		}

		if err := scanner.Err(); err != nil {
			if err != io.EOF {
				scannerResult.reason = relaycommon.StreamEndReasonScannerErr
				scannerResult.err = err
			}
		}
	})

	endReason, endError := waitForStreamEnd(c.Request.Context(), ticker, scannerFinished, stopChan, scannerResult)

	// Close the upstream immediately, then let any in-flight callback finish.
	// A terminal event already forwarded by that callback must win over a
	// concurrent disconnect or idle timeout before its Done/Stop is recorded.
	cleanup()
	if endReason == relaycommon.StreamEndReasonNone && c.Request.Context().Err() != nil {
		endReason = relaycommon.StreamEndReasonClientGone
		endError = c.Request.Context().Err()
	}
	if endReason != relaycommon.StreamEndReasonNone {
		info.StreamStatus.SetEndReason(endReason, endError)
	}
	if info.StreamStatus.IsNormalEnd() && !info.StreamStatus.HasErrors() {
		logger.LogInfo(c, fmt.Sprintf("stream ended: %s", info.StreamStatus.Summary()))
	} else {
		logger.LogError(c, fmt.Sprintf("stream ended: %s, received=%d", info.StreamStatus.Summary(), info.ReceivedResponseCount))
	}
}

// waitForStreamEnd distinguishes upstream idleness from buffered output still
// being delivered. A finished reader no longer needs the upstream idle watchdog,
// while handler completion and client cancellation remain active.
func waitForStreamEnd(requestContext context.Context, ticker *time.Ticker, scannerFinished <-chan struct{}, handlerStopped <-chan bool, scannerResult *streamReadResult) (relaycommon.StreamEndReason, error) {
	idleTimeout := ticker.C
	upstreamFinished := scannerFinished
	for {
		select {
		case <-upstreamFinished:
			// Closing scannerFinished guarantees no further scanner Reset calls.
			ticker.Stop()
			idleTimeout = nil
			upstreamFinished = nil
		case <-idleTimeout:
			// Both notifications may be ready; completed upstream reads take
			// precedence over their now-irrelevant idle deadline.
			select {
			case <-scannerFinished:
				continue
			default:
			}
			return relaycommon.StreamEndReasonTimeout, nil
		case <-handlerStopped:
			if requestContext.Err() == nil {
				return relaycommon.StreamEndReasonNone, nil
			}
		case <-requestContext.Done():
		}
		if err := requestContext.Err(); err != nil {
			// Only a result published before cleanup is evidence of an upstream
			// failure. Closing Body on cancellation can itself cause a read error;
			// that later error must not disguise a real client disconnect.
			// Likewise, a request cancellation propagated to Read is not a fault.
			select {
			case <-scannerFinished:
				if (scannerResult.reason == relaycommon.StreamEndReasonScannerErr || scannerResult.reason == relaycommon.StreamEndReasonPanic) &&
					!errors.Is(scannerResult.err, err) {
					return scannerResult.reason, scannerResult.err
				}
			default:
			}
			return relaycommon.StreamEndReasonClientGone, err
		}
	}
}
