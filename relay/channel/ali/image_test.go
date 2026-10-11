package ali

import (
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
