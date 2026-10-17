package helper

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
)

// WriteResponsesStreamError reports failure after an HTTP 200 SSE stream has
// started. Callers supply stable public messages, never raw transport errors.
func WriteResponsesStreamError(c *gin.Context, code, message string, sequenceNumber int64) error {
	payload, err := common.Marshal(struct {
		Type           string  `json:"type"`
		Code           string  `json:"code"`
		Message        string  `json:"message"`
		Param          *string `json:"param"`
		SequenceNumber int64   `json:"sequence_number"`
	}{Type: "error", Code: code, Message: message, SequenceNumber: sequenceNumber})
	if err != nil {
		return err
	}
	ExtendWriteDeadline(c)
	return ResponseChunkData(c, dto.ResponsesStreamResponse{Type: "error"}, string(payload))
}
