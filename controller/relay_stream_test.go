package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayDoesNotRetryCommittedStream(t *testing.T) {
	for _, errorCode := range []types.ErrorCode{types.ErrorCodeDoRequestFailed, types.ErrorCodeChannelInvalidKey} {
		t.Run(string(errorCode), func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			apiErr := types.NewErrorWithStatusCode(errors.New("upstream failed"), errorCode, http.StatusBadGateway)

			_, err := c.Writer.WriteString("event: response.output_text.delta\ndata: {}\n\n")
			require.NoError(t, err)
			assert.False(t, shouldRetry(c, apiErr, 2), "a retry would append another answer to the committed stream")
		})
	}
}

func TestRelayErrorPreservesCommittedStream(t *testing.T) {
	for _, test := range []struct {
		name       string
		prepareSSE bool
		committed  bool
	}{
		{name: "before output"},
		{name: "prepared stream headers", prepareSSE: true},
		{name: "after output", prepareSSE: true, committed: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader("{"))
			c.Request.Header.Set("Content-Type", "application/json")
			const partial = "event: response.output_text.delta\ndata: {}\n\n"
			if test.prepareSSE {
				helper.SetEventStreamHeaders(c)
			}
			if test.committed {
				_, err := c.Writer.WriteString(partial)
				require.NoError(t, err)
			}

			Relay(c, types.RelayFormatOpenAIResponses)

			if test.committed {
				assert.Equal(t, partial, w.Body.String(), "a late JSON error must not corrupt an existing SSE response")
				assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
				return
			}
			var response map[string]any
			require.NoError(t, common.Unmarshal(w.Body.Bytes(), &response))
			assert.Contains(t, response, "error")
			assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
			assert.Empty(t, w.Header().Get("Transfer-Encoding"))
			assert.GreaterOrEqual(t, w.Code, http.StatusBadRequest)
		})
	}
}
