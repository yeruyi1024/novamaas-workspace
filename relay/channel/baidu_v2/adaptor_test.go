package baidu_v2

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertImageRequestPreservesOpenAICompatiblePayload(t *testing.T) {
	request := dto.ImageRequest{Model: "ernie-image", Prompt: "ink landscape", Size: "1024x1024"}

	converted, err := (&Adaptor{}).ConvertImageRequest(nil, nil, request)

	require.NoError(t, err)
	assert.Equal(t, request, converted)
}
