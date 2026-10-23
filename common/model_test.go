package common

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/assert"
)

func TestGPTImageModelsExposeImageGenerationEndpoint(t *testing.T) {
	for _, name := range []string{
		"gpt-image-1", "gpt-image-1-mini", "gpt-image-2-1K",
		"gpt-image-2-adobe", "gpt-image-2-sr", "gpt-image-2.5-flare-1K",
	} {
		t.Run(name, func(t *testing.T) {
			assert.True(t, IsImageGenerationModel(name))
			assert.Equal(t, []constant.EndpointType{constant.EndpointTypeImageGeneration, constant.EndpointTypeOpenAI},
				GetEndpointTypesByChannelType(constant.ChannelTypeOpenAI, name))
		})
	}
	assert.False(t, IsImageGenerationModel("gpt-4o-mini"))
}
