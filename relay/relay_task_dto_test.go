package relay

import (
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskModel2DtoReportsRequestBodyWithoutIncludingItInListPayload(t *testing.T) {
	task := &model.Task{
		Properties: model.Properties{
			Input:       "legacy-input",
			RequestBody: json.RawMessage(`{"prompt":"private prompt"}`),
		},
	}

	result := TaskModel2Dto(task)

	assert.True(t, result.RequestBodyAvailable)
	properties, ok := result.Properties.(model.Properties)
	require.True(t, ok)
	assert.Empty(t, properties.RequestBody)
	assert.Equal(t, "legacy-input", properties.Input)
}
