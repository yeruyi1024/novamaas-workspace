package doubao

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relaykitdto "github.com/QuantumNous/new-api/relaykit/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoubaoVideoBuildRequestBodyUsesCachedBase64Staging(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", nil)
	ctx.Set("task_request", relaycommon.TaskSubmitReq{
		Model:  "seedance-public",
		Prompt: "animate the reference",
		Metadata: map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "video_url",
					"video_url": map[string]interface{}{
						"url": "data:video/mp4;base64,AAAA",
					},
				},
			},
		},
	})
	staged := []byte(`{"model":"seedance-public","content":[{"type":"video_url","video_url":{"url":"https://storage.example.com/reference.mp4"}},{"type":"text","text":"animate the reference"}]}`)
	ctx.Set("doubao_video_base64_staging:relay_media_temp", stagedRequestBody{body: staged, convertedCount: 1})
	info := &relaycommon.RelayInfo{
		UserId:          19,
		RequestId:       "request-doubao-video",
		ChannelMeta:     &relaycommon.ChannelMeta{ChannelOtherSettings: relaykitdto.ChannelOtherSettings{Base64Staging: &relaykitdto.Base64StagingSettings{Enabled: true}}},
		TaskRelayInfo:   &relaycommon.TaskRelayInfo{PublicTaskID: "task_doubao_video"},
		OriginModelName: "seedance-public",
	}

	body, err := (&TaskAdaptor{}).BuildRequestBody(ctx, info)
	require.NoError(t, err)
	got, err := io.ReadAll(body)
	require.NoError(t, err)

	assert.JSONEq(t, string(staged), string(got))
	assert.True(t, common.GetContextKeyBool(ctx, constant.ContextKeyTemporaryMediaConverted))
	assert.Equal(t, 1, common.GetContextKeyInt(ctx, constant.ContextKeyTemporaryMediaConvertedCount))
	assert.JSONEq(t, string(staged), common.GetContextKeyString(ctx, constant.ContextKeyVideoTaskUpstreamRequestBody))
}

func TestDoubaoVideoBuildRequestBodyAppliesModelMappingAfterBase64Staging(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/video/generations", nil)
	ctx.Set("task_request", relaycommon.TaskSubmitReq{Model: "seedance-public", Prompt: "hello"})
	ctx.Set("doubao_video_base64_staging:relay_media_temp", stagedRequestBody{
		body:           []byte(`{"model":"seedance-public","content":[{"type":"text","text":"hello"}]}`),
		convertedCount: 1,
	})
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			IsModelMapped:     true,
			UpstreamModelName: "doubao-seedance-2-0-260128",
			ChannelOtherSettings: relaykitdto.ChannelOtherSettings{
				Base64Staging: &relaykitdto.Base64StagingSettings{Enabled: true, StoragePolicy: "relay_media_temp"},
			},
		},
		TaskRelayInfo:   &relaycommon.TaskRelayInfo{PublicTaskID: "task_doubao_video"},
		OriginModelName: "seedance-public",
	}

	body, err := (&TaskAdaptor{}).BuildRequestBody(ctx, info)
	require.NoError(t, err)
	got, err := io.ReadAll(body)
	require.NoError(t, err)

	assert.JSONEq(t, `{"model":"doubao-seedance-2-0-260128","content":[{"type":"text","text":"hello"}]}`, string(got))
	assert.JSONEq(t, string(got), common.GetContextKeyString(ctx, constant.ContextKeyVideoTaskUpstreamRequestBody))
}
