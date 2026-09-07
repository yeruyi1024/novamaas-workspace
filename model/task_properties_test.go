package model

import (
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropertiesDatabaseValueRoundTripPreservesRequestBody(t *testing.T) {
	properties := Properties{
		OriginModelName:   "public-model",
		UpstreamModelName: "upstream-model",
		RequestBody:       json.RawMessage(`{"model":"public-model","duration":5}`),
	}

	value, err := properties.Value()
	require.NoError(t, err)
	stored, ok := value.([]byte)
	require.True(t, ok)

	var restored Properties
	require.NoError(t, restored.Scan(stored))
	assert.Equal(t, properties.OriginModelName, restored.OriginModelName)
	assert.Equal(t, properties.UpstreamModelName, restored.UpstreamModelName)
	assert.JSONEq(t, string(properties.RequestBody), string(restored.RequestBody))
}

func TestTaskJSONDoesNotExposeStoredRequestBody(t *testing.T) {
	task := Task{
		Properties: Properties{
			OriginModelName: "public-model",
			RequestBody:     json.RawMessage(`{"prompt":"private prompt"}`),
		},
	}

	body, err := common.Marshal(task)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "private prompt")
	assert.NotContains(t, string(body), "request_body")
}
