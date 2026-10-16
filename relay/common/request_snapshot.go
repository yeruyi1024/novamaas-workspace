package common

import (
	"fmt"
	"net/http"

	rootcommon "github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"

	"github.com/gin-gonic/gin"
)

type upstreamRequestSnapshotTooLargeError struct {
	size int
}

func (err *upstreamRequestSnapshotTooLargeError) Error() string {
	return fmt.Sprintf("upstream request snapshot is %d bytes and exceeds the %d byte persistence limit", err.size, constant.MaxVideoTaskRequestBodyBytes)
}

func (err *upstreamRequestSnapshotTooLargeError) HTTPStatusCode() int {
	return http.StatusRequestEntityTooLarge
}

// UpstreamRequestRecorder archives the final upstream body of a synchronous
// relay request. It is registered under constant.ContextKeyUpstreamRequestRecorder
// and runs before the request is sent; an error prevents the upstream call.
type UpstreamRequestRecorder func(c *gin.Context, upstreamBody []byte) error

// RecordVideoTaskUpstreamRequestBody captures the final JSON bytes after all
// provider conversion, temporary-media materialization, and model mapping.
// External snapshots are durably handed to the node's queue. The legacy
// database path retains its size limit before the upstream request is sent.
func RecordVideoTaskUpstreamRequestBody(c *gin.Context, body []byte) error {
	rootcommon.SetContextKey(c, constant.ContextKeyVideoTaskUpstreamBodySize, int64(len(body)))
	if rootcommon.GetContextKeyBool(c, constant.ContextKeyVideoTaskExternalSnapshots) {
		if record, ok := rootcommon.GetContextKeyType[func([]byte)](c, constant.ContextKeyVideoTaskSnapshotRecorder); ok {
			record(body)
		}
		return nil
	}
	if len(body) > constant.MaxVideoTaskRequestBodyBytes {
		rootcommon.SetContextKey(c, constant.ContextKeyVideoTaskUpstreamRequestBody, "")
		return &upstreamRequestSnapshotTooLargeError{size: len(body)}
	}
	rootcommon.SetContextKey(c, constant.ContextKeyVideoTaskUpstreamRequestBody, string(body))
	return nil
}
