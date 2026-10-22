package sora

import (
	"testing"

	"github.com/QuantumNous/new-api/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestConvertToOpenAIVideoReplacesStoredUpstreamIDs(t *testing.T) {
	adaptor := &TaskAdaptor{}
	task := &model.Task{TaskID: "task_public_id", Data: []byte(`{"id":"upstream-id","task_id":"upstream-id","status":"queued"}`)}

	body, err := adaptor.ConvertToOpenAIVideo(task)
	require.NoError(t, err)
	assert.Equal(t, "task_public_id", gjson.GetBytes(body, "id").String())
	assert.Equal(t, "task_public_id", gjson.GetBytes(body, "task_id").String())

	idOnly := &model.Task{TaskID: "task_public_id", Data: []byte(`{"id":"upstream-id","status":"queued"}`)}
	body, err = adaptor.ConvertToOpenAIVideo(idOnly)
	require.NoError(t, err)
	assert.Equal(t, "task_public_id", gjson.GetBytes(body, "id").String())
	assert.False(t, gjson.GetBytes(body, "task_id").Exists())
}
