package helper

import (
	"net/http/httptest"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyModelMappingToJSONBodyChangesOnlyTopLevelModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("model_mapping", `{"public-model":"intermediate-model","intermediate-model":"upstream-model"}`)
	info := &relaycommon.RelayInfo{
		OriginModelName: "public-model",
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "public-model",
		},
	}

	require.NoError(t, ModelMappedHelper(c, info, nil))
	assert.True(t, info.IsModelMapped)
	assert.Equal(t, "upstream-model", info.UpstreamModelName)

	body := []byte(`{ "model": "public-model", "content": [{"type":"text","text":"hello","model":"nested-model"}], "watermark": false, "seed": 0, "extra": 9007199254740993 }`)
	mappedBody, err := ApplyModelMappingToJSONBody(info, body)

	require.NoError(t, err)
	assert.Equal(t, `{ "model": "upstream-model", "content": [{"type":"text","text":"hello","model":"nested-model"}], "watermark": false, "seed": 0, "extra": 9007199254740993 }`, string(mappedBody))
}

func TestApplyModelMappingToJSONBodyLeavesUnmappedPayloadUnchanged(t *testing.T) {
	body := []byte(`{"model":"upstream-model","watermark":false}`)
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "upstream-model"}}

	mappedBody, err := ApplyModelMappingToJSONBody(info, body)

	require.NoError(t, err)
	assert.Equal(t, body, mappedBody)
}

func TestRestoreOriginalModelInJSONBodyDoesNotAddMissingField(t *testing.T) {
	body := []byte(`{"id":"task-id","status":"queued"}`)

	restoredBody, err := RestoreOriginalModelInJSONBody(body, "public-model")

	require.NoError(t, err)
	assert.Equal(t, body, restoredBody)
}
