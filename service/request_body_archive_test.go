package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestBodyArchiveMovesLegacyPayloadsAndPreservesAdminLookup(t *testing.T) {
	truncate(t)

	require.NoError(t, model.DB.Create(&model.Task{
		TaskID: "task_legacy",
		Properties: model.Properties{
			RequestBody: json.RawMessage(`{"source":"task-property"}`),
		},
	}).Error)
	require.NoError(t, model.DB.Create(&model.Log{
		CreatedAt: 100,
		Type:      model.LogTypeConsume,
		RequestId: "request_task",
		Other: common.MapToJsonStr(map[string]any{
			"is_task":      true,
			"task_id":      "task_legacy",
			"request_body": `{"source":"usage-log"}`,
		}),
	}).Error)
	require.NoError(t, model.DB.Create(&model.Log{
		CreatedAt: 101,
		Type:      model.LogTypeError,
		RequestId: "request_orphan",
		Other: common.MapToJsonStr(map[string]any{
			"is_task":      true,
			"request_body": `{"source":"orphan-error"}`,
		}),
	}).Error)

	task, err := model.CreateSystemTask(
		model.SystemTaskTypeRequestBodyArchive,
		RequestBodyArchivePayload{BatchSize: 1},
		RequestBodyArchiveState{},
	)
	require.NoError(t, err)
	claimed, ok, err := model.ClaimSystemTask(task.ID, task.Type, "archive-runner", common.GetTimestamp()+60)
	require.NoError(t, err)
	require.True(t, ok)

	runRequestBodyArchiveTask(context.Background(), claimed, "archive-runner")

	finished, err := model.GetSystemTaskByTaskID(task.TaskID)
	require.NoError(t, err)
	require.NotNil(t, finished)
	assert.Equal(t, model.SystemTaskStatusSucceeded, finished.Status)

	var storedTask model.Task
	require.NoError(t, model.DB.Where("task_id = ?", "task_legacy").First(&storedTask).Error)
	assert.Empty(t, storedTask.Properties.RequestBody)
	assert.True(t, storedTask.RequestBodyAvailable)

	var logs []*model.Log
	require.NoError(t, model.DB.Order("id ASC").Find(&logs).Error)
	require.Len(t, logs, 2)
	for _, log := range logs {
		var other map[string]any
		require.NoError(t, common.Unmarshal([]byte(log.Other), &other))
		assert.NotContains(t, other, "request_body")
		assert.Equal(t, true, other["request_body_available"])
	}

	body, exists, err := model.GetArchivedRequestBody("task_legacy", "")
	require.NoError(t, err)
	require.True(t, exists)
	assert.JSONEq(t, `{"source":"usage-log"}`, string(body))
	body, exists, err = model.GetArchivedRequestBody("", "request_orphan")
	require.NoError(t, err)
	require.True(t, exists)
	assert.JSONEq(t, `{"source":"orphan-error"}`, string(body))

	var archiveRows int64
	require.NoError(t, model.DB.Model(&model.TaskRequestBody{}).Count(&archiveRows).Error)
	assert.Equal(t, int64(2), archiveRows)
}

func TestLogCleanupRemovesOnlyExpiredOrphanRequestBodies(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	require.NoError(t, model.DB.Create(&model.Log{CreatedAt: 100, RequestId: "request_old"}).Error)
	require.NoError(t, model.DB.Create(&model.Log{CreatedAt: 200, RequestId: "request_recent"}).Error)
	require.NoError(t, model.DB.Create(&model.Task{TaskID: "task_retained"}).Error)
	require.NoError(t, model.SaveTaskRequestBodyAt(ctx, "", "request_old", []byte(`{"old":true}`), 100))
	require.NoError(t, model.SaveTaskRequestBodyAt(ctx, "", "request_recent", []byte(`{"recent":true}`), 200))
	require.NoError(t, model.SaveTaskRequestBodyAt(ctx, "task_retained", "request_task", []byte(`{"task":true}`), 100))
	require.NoError(t, model.SaveTaskRequestBodyAt(ctx, "task_failed", "request_failed", []byte(`{"failed":true}`), 100))

	task, err := model.CreateSystemTask(
		model.SystemTaskTypeLogCleanup,
		LogCleanupPayload{TargetTimestamp: 150, BatchSize: 10},
		LogCleanupState{},
	)
	require.NoError(t, err)
	claimed, ok, err := model.ClaimSystemTask(task.ID, task.Type, "cleanup-runner", common.GetTimestamp()+60)
	require.NoError(t, err)
	require.True(t, ok)

	runLogCleanupTask(ctx, claimed, "cleanup-runner")

	_, exists, err := model.GetArchivedRequestBody("", "request_old")
	require.NoError(t, err)
	assert.False(t, exists)
	_, exists, err = model.GetArchivedRequestBody("", "request_recent")
	require.NoError(t, err)
	assert.True(t, exists)
	_, exists, err = model.GetArchivedRequestBody("task_retained", "")
	require.NoError(t, err)
	assert.True(t, exists)
	_, exists, err = model.GetArchivedRequestBody("task_failed", "")
	require.NoError(t, err)
	assert.False(t, exists)
}
