package doubao

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskAdaptorAppliesResolutionOverrideBeforeBillingAndConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		keepOrigin        bool
		wantResolution    string
		wantBillingRatio  float64
		wantBillingExists bool
	}{
		{
			name:           "replace existing resolution",
			wantResolution: "720p",
		},
		{
			name:              "keep original resolution",
			keepOrigin:        true,
			wantResolution:    "1080p",
			wantBillingRatio:  51.0 / 46.0,
			wantBillingExists: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(`{"model":"doubao-seedance-2-0-260128","prompt":"a flying cat","metadata":{"resolution":"1080p"}}`)
			request := httptest.NewRequest(http.MethodPost, "/v1/video/generations", body)
			request.Header.Set("Content-Type", "application/json")
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = request
			defer common.CleanupBodyStorage(context)

			info := &relaycommon.RelayInfo{
				OriginModelName: "doubao-seedance-2-0-260128",
				ChannelMeta: &relaycommon.ChannelMeta{
					ParamOverride: map[string]interface{}{
						"operations": []interface{}{
							map[string]interface{}{
								"mode":        "set",
								"description": "resolution",
								"path":        "metadata.resolution",
								"value":       "720p",
								"keep_origin": test.keepOrigin,
							},
						},
					},
				},
				TaskRelayInfo: &relaycommon.TaskRelayInfo{},
			}
			adaptor := &TaskAdaptor{}

			taskErr := adaptor.ValidateRequestAndSetAction(context, info)
			require.Nil(t, taskErr)

			ratios := adaptor.EstimateBilling(context, info)
			billingRatio, billingExists := ratios["video_input"]
			assert.Equal(t, test.wantBillingExists, billingExists)
			if test.wantBillingExists {
				assert.InDelta(t, test.wantBillingRatio, billingRatio, 1e-12)
			}

			requestBody, err := adaptor.BuildRequestBody(context, info)
			require.NoError(t, err)
			requestJSON, err := io.ReadAll(requestBody)
			require.NoError(t, err)
			var upstreamBody requestPayload
			require.NoError(t, common.Unmarshal(requestJSON, &upstreamBody))
			assert.Equal(t, test.wantResolution, upstreamBody.Resolution)
		})
	}
}
