package volcengine

import (
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func imageInfo(mode int, model string) *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{RelayMode: mode, ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: model}}
}

func TestSeedreamEditMapsOpenAIReferenceImagesToImageField(t *testing.T) {
	request := dto.ImageRequest{
		Model:  "doubao-seedream-4-0-250828",
		Prompt: "restyle",
		Images: []byte(`[{"image_url":"https://example.test/a.png"},{"image_url":"https://example.test/b.png"}]`),
	}
	converted, err := (&Adaptor{}).ConvertImageRequest(nil, imageInfo(constant.RelayModeImagesEdits, request.Model), request)
	require.NoError(t, err)
	got := converted.(dto.ImageRequest)
	assert.JSONEq(t, `["https://example.test/a.png","https://example.test/b.png"]`, string(got.Image))
	assert.Empty(t, got.Images)
}

func TestSeedreamSizesUseTwoKPresetsForGenerationAndEdits(t *testing.T) {
	for _, mode := range []int{constant.RelayModeImagesGenerations, constant.RelayModeImagesEdits} {
		for size, want := range map[string]string{
			"1024x1024": "2048x2048", "1536x1024": "2496x1664", "1024x1536": "1664x2496",
			"1792x1024": "2816x1584", "1024x1792": "1584x2816", "2048x2048": "2048x2048", "": "",
		} {
			converted, err := (&Adaptor{}).ConvertImageRequest(nil, imageInfo(mode, "seedream-4-0-250828"), dto.ImageRequest{Model: "seedream-4-0-250828", Size: size})
			require.NoError(t, err)
			assert.Equal(t, want, converted.(dto.ImageRequest).Size)
		}
	}
}

func TestNonSeedreamImageRequestsAreUntouched(t *testing.T) {
	request := dto.ImageRequest{
		Model: "other-image-model", Size: "1024x1024",
		Images: []byte(`[{"image_url":"https://example.test/a.png"}]`),
	}
	for _, info := range []*relaycommon.RelayInfo{
		imageInfo(constant.RelayModeImagesEdits, request.Model),
		{RelayMode: constant.RelayModeImagesEdits},
	} {
		converted, err := (&Adaptor{}).ConvertImageRequest(nil, info, request)
		require.NoError(t, err)
		assert.Equal(t, request, converted)
	}
}

func TestSeedreamEditKeepsExplicitImageField(t *testing.T) {
	request := dto.ImageRequest{Model: "seedream", Image: []byte(`"https://example.test/a.png"`),
		Images: []byte(`[{"image_url":"https://example.test/b.png"}]`)}
	converted, err := (&Adaptor{}).ConvertImageRequest(nil, imageInfo(constant.RelayModeImagesEdits, "seedream"), request)
	require.NoError(t, err)
	assert.JSONEq(t, `"https://example.test/a.png"`, string(converted.(dto.ImageRequest).Image))
}
