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

// RecordVideoTaskUpstreamRequestBody captures the final JSON bytes after all
// provider conversion, temporary-media materialization, and model mapping.
// The limit is enforced before the HTTP request is sent so every accepted
// audited request can be persisted without an oversized database write.
func RecordVideoTaskUpstreamRequestBody(c *gin.Context, body []byte) error {
	if len(body) > constant.MaxVideoTaskRequestBodyBytes {
		rootcommon.SetContextKey(c, constant.ContextKeyVideoTaskUpstreamRequestBody, "")
		return &upstreamRequestSnapshotTooLargeError{size: len(body)}
	}
	rootcommon.SetContextKey(c, constant.ContextKeyVideoTaskUpstreamRequestBody, string(body))
	return nil
}
