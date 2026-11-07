package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An older video log carries no task ID, but the task was submitted when the
// charge was made. A time resolves only to the user's single task around it.
func TestGetTaskIDsBySubmitTimesResolvesOnlyAnUnambiguousTask(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.AutoMigrate(&Task{}))
	require.NoError(t, DB.Create(&[]Task{
		{TaskID: "task_a", UserId: 7, SubmitTime: 1000},
		{TaskID: "task_b", UserId: 7, SubmitTime: 2000},
		{TaskID: "task_c", UserId: 7, SubmitTime: 2001},
		{TaskID: "task_other", UserId: 8, SubmitTime: 3000},
	}).Error)

	got, err := GetTaskIDsBySubmitTimes(7, []int64{1000, 1001, 2000, 3000, 9999})

	require.NoError(t, err)
	assert.Equal(t, map[int64]string{1000: "task_a", 1001: "task_a"}, got, "one task within a second resolves; two candidates, another user's task and no task do not")
}
