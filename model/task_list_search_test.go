package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskListFiltersUsernameAndVisibleModelConsistentlyWithCount(t *testing.T) {
	truncateTables(t)
	alice := User{Username: "task-search-alice", Password: "password", Status: common.UserStatusEnabled, AffCode: "task-search-alice"}
	bob := User{Username: "task-search-bob", Password: "password", Status: common.UserStatusEnabled, AffCode: "task-search-bob"}
	require.NoError(t, DB.Create(&alice).Error)
	require.NoError(t, DB.Create(&bob).Error)
	tasks := []Task{
		{TaskID: "task-search-1", UserId: alice.Id, SubmitTime: 3, Properties: Properties{OriginModelName: "Video-GPT", UpstreamModelName: "vendor-video"}},
		{TaskID: "task-search-2", UserId: alice.Id, SubmitTime: 2, Properties: Properties{UpstreamModelName: "fallback-video"}},
		{TaskID: "task-search-3", UserId: bob.Id, SubmitTime: 1, Properties: Properties{OriginModelName: "Video-GPT"}},
	}
	require.NoError(t, DB.Create(&tasks).Error)

	adminFilter := SyncTaskQueryParams{Username: "ALICE", ModelName: "video-gpt"}
	listed, err := TaskGetAllTasks(0, 20, adminFilter)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, "task-search-1", listed[0].TaskID)
	total, err := TaskCountAllTasks(adminFilter)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)

	userFilter := SyncTaskQueryParams{ModelName: "FALLBACK"}
	listed, err = TaskGetAllUserTask(alice.Id, 0, 20, userFilter)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, "task-search-2", listed[0].TaskID)
	total, err = TaskCountAllUserTask(alice.Id, userFilter)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
}
