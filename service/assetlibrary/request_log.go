package assetlibrary

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const (
	requestLogQueueSize     = 1024
	requestLogBatchSize     = 50
	requestLogRetentionDays = 14
	requestLogDetailDays    = 7
)

type requestLogContextKey struct{}

type requestLogContext struct {
	source    string
	replicaID int64
	assetID   int64
	requestID string
}

type assetRequestLogWriter struct {
	mu     sync.RWMutex
	queue  chan model.AssetRequestLogEntry
	done   chan struct{}
	closed bool
}

var activeRequestLogWriter atomic.Pointer[assetRequestLogWriter]
var droppedRequestLogs atomic.Uint64
var lastRequestLogErrorAt atomic.Int64

func withRequestLogContext(ctx context.Context, source string, replicaID int64, assetID int64) context.Context {
	requestID, _ := ctx.Value(common.RequestIdKey).(string)
	if requestID == "" {
		requestID = common.NewRequestId()
	}
	return context.WithValue(ctx, requestLogContextKey{}, requestLogContext{
		source: source, replicaID: replicaID, assetID: assetID, requestID: requestID,
	})
}

func recordRequestAttempt(ctx context.Context, channelID int, protocol string, operation string, method string, pathTemplate string, status int, elapsed time.Duration, requestBytes int64, responseBytes int64, errorKind string, requestURL string, requestBody []byte, responseBody []byte) {
	metadata, ok := ctx.Value(requestLogContextKey{}).(requestLogContext)
	if !ok {
		return
	}
	result := "success"
	if errorKind != "" {
		result = "failure"
	}
	log := model.AssetRequestLog{
		CreatedAt: time.Now().UnixMilli(), ChannelID: channelID, ReplicaID: metadata.replicaID,
		AssetID: metadata.assetID, RequestID: metadata.requestID, Source: metadata.source,
		Protocol: protocol, Operation: operation, Method: method, PathTemplate: pathTemplate,
		HTTPStatus: status, DurationMS: elapsed.Milliseconds(), RequestBytes: requestBytes,
		ResponseBytes: responseBytes, Result: result, ErrorKind: errorKind,
	}
	writer := activeRequestLogWriter.Load()
	if writer == nil {
		droppedRequestLogs.Add(1)
		return
	}
	writer.mu.RLock()
	if writer.closed {
		writer.mu.RUnlock()
		droppedRequestLogs.Add(1)
		return
	}
	entry := model.AssetRequestLogEntry{Log: log, Detail: requestLogDetail(requestURL, requestBody, responseBody)}
	select {
	case writer.queue <- entry:
	default:
		droppedRequestLogs.Add(1)
	}
	writer.mu.RUnlock()
}

func StartAssetRequestLogWriter() {
	writer := &assetRequestLogWriter{
		queue: make(chan model.AssetRequestLogEntry, requestLogQueueSize),
		done:  make(chan struct{}),
	}
	if !activeRequestLogWriter.CompareAndSwap(nil, writer) {
		return
	}
	go writer.run()
	if common.IsMasterNode {
		go purgeOldAssetRequestLogs()
	}
}

func StopAssetRequestLogWriter(ctx context.Context) error {
	writer := activeRequestLogWriter.Swap(nil)
	if writer == nil {
		return nil
	}
	writer.mu.Lock()
	writer.closed = true
	close(writer.queue)
	writer.mu.Unlock()
	select {
	case <-writer.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func DroppedAssetRequestLogs() uint64 {
	return droppedRequestLogs.Load()
}

func (writer *assetRequestLogWriter) run() {
	defer close(writer.done)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	batch := make([]model.AssetRequestLogEntry, 0, requestLogBatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := model.InsertAssetRequestLogEntries(ctx, batch)
		cancel()
		if err != nil {
			droppedRequestLogs.Add(uint64(len(batch)))
			now := time.Now().Unix()
			previous := lastRequestLogErrorAt.Load()
			if now-previous >= 60 && lastRequestLogErrorAt.CompareAndSwap(previous, now) {
				common.SysError(fmt.Sprintf("asset request log batch dropped: count=%d error=%v", len(batch), err))
			}
		}
		batch = batch[:0]
	}
	for {
		select {
		case entry, ok := <-writer.queue:
			if !ok {
				flush()
				return
			}
			batch = append(batch, entry)
			if len(batch) >= requestLogBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func purgeOldAssetRequestLogs() {
	purge := func() {
		detailCutoff := time.Now().AddDate(0, 0, -requestLogDetailDays).UnixMilli()
		for range 10 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			count, err := model.PurgeAssetRequestLogDetailsBefore(ctx, detailCutoff, 1000)
			cancel()
			if err != nil {
				common.SysError(fmt.Sprintf("purge asset request log details failed: %v", err))
				break
			}
			if count < 1000 {
				break
			}
		}
		cutoff := time.Now().AddDate(0, 0, -requestLogRetentionDays).UnixMilli()
		for range 10 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			count, err := model.PurgeAssetRequestLogsBefore(ctx, cutoff, 1000)
			cancel()
			if err != nil {
				common.SysError(fmt.Sprintf("purge asset request logs failed: %v", err))
				return
			}
			if count < 1000 {
				return
			}
		}
	}
	purge()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		purge()
	}
}
