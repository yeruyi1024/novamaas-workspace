package ali

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQwenImageEditConvertsStudioReferencesToMessageContent(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/edits", nil)
	c.Request.Header.Set("Content-Type", gin.MIMEJSON)
	info := &relaycommon.RelayInfo{RelayMode: relayconstant.RelayModeImagesEdits, ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "qwen-image-edit-plus"}}
	images, err := common.Marshal([]string{"https://assets.example.test/first.png", "https://assets.example.test/second.png"})
	require.NoError(t, err)

	converted, err := (&Adaptor{}).ConvertImageRequest(c, info, dto.ImageRequest{Model: info.UpstreamModelName, Prompt: "combine", Images: images, Image: images})

	require.NoError(t, err)
	body, err := common.Marshal(converted)
	require.NoError(t, err)
	assert.JSONEq(t, `{"model":"qwen-image-edit-plus","input":{"messages":[{"role":"user","content":[{"image":"https://assets.example.test/first.png"},{"image":"https://assets.example.test/second.png"},{"text":"combine"}]}]},"parameters":{"n":1}}`, string(body))
}

func TestAliImageExtraParametersPreserveStandardRequestFields(t *testing.T) {
	extra, err := common.Marshal(map[string]any{"prompt_extend": false, "seed": 0})
	require.NoError(t, err)
	request := dto.ImageRequest{
		Model: "qwen-image-plus", Prompt: "paper sculpture", Size: "1536x1024",
		N: common.GetPointer(uint(3)), Extra: map[string]json.RawMessage{"parameters": extra},
	}
	info := &relaycommon.RelayInfo{}

	converted, err := oaiImage2AliImageRequest(info, request, true)

	require.NoError(t, err)
	assert.Equal(t, "1536*1024", converted.Parameters.Size)
	assert.Equal(t, 3, converted.Parameters.N)
	require.NotNil(t, converted.Parameters.PromptExtend)
	assert.False(t, *converted.Parameters.PromptExtend)
	require.NotNil(t, converted.Parameters.Seed)
	assert.Zero(t, *converted.Parameters.Seed)
	assert.Equal(t, 3.0, info.PriceData.OtherRatios()["n"])
}
