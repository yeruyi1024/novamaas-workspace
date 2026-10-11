package volcengine

import (
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertImageEditUsesSeedreamGenerationPayload(t *testing.T) {
	request := dto.ImageRequest{Model: "doubao-seedream-4-5", Prompt: "restyle", Image: []byte(`"https://example.test/reference.png"`)}
	info := &relaycommon.RelayInfo{RelayMode: relayconstant.RelayModeImagesEdits}

	converted, err := (&Adaptor{}).ConvertImageRequest(nil, info, request)

	require.NoError(t, err)
	assert.Equal(t, request, converted)
}
