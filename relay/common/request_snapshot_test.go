package common

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	rootcommon "github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecordVideoTaskUpstreamRequestBodyRejectsOversizedSnapshotBeforeSend(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	body := bytes.Repeat([]byte(" "), constant.MaxVideoTaskRequestBodyBytes+1)

	err := RecordVideoTaskUpstreamRequestBody(c, body)

	require.Error(t, err)
	var statusError interface{ HTTPStatusCode() int }
	require.ErrorAs(t, err, &statusError)
	assert.Equal(t, http.StatusRequestEntityTooLarge, statusError.HTTPStatusCode())
	assert.Empty(t, rootcommon.GetContextKeyString(c, constant.ContextKeyVideoTaskUpstreamRequestBody))
}
