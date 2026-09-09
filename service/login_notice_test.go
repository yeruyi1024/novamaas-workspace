package service

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsVideoTaskViolation(t *testing.T) {
	testCases := []struct {
		name   string
		reason string
		want   bool
	}{
		{name: "english policy rejection", reason: "Output blocked by content policy", want: true},
		{name: "provider inspection", reason: "InputDataMayContainInappropriateContent", want: true},
		{name: "existing safety marker", reason: "Failed check: SAFETY_CHECK_TYPE", want: true},
		{name: "chinese violation", reason: "内容违规，安全审核不通过", want: true},
		{name: "ordinary upstream error", reason: "upstream request timed out", want: false},
		{name: "empty", reason: "", want: false},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.want, IsVideoTaskViolation(testCase.reason))
		})
	}
}

func TestGetLoginNoticeStatsUsesCalendarWindowsAndVideoActions(t *testing.T) {
	user := setupAuthSessionTestDB(t)
	require.NoError(t, model.DB.AutoMigrate(&model.Task{}))
	location := time.FixedZone("test", 8*60*60)
	now := time.Date(2026, time.September, 9, 15, 0, 0, 0, location)
	todayStart := time.Date(2026, time.September, 9, 0, 0, 0, 0, location)

	tasks := []model.Task{
		{TaskID: "today-ok", UserId: user.Id, Action: constant.TaskActionGenerate, SubmitTime: todayStart.Add(time.Hour).Unix(), Status: model.TaskStatusSuccess},
		{TaskID: "today-violation", UserId: user.Id, Action: constant.TaskActionTextGenerate, SubmitTime: todayStart.Add(2 * time.Hour).Unix(), Status: model.TaskStatusFailure, FailReason: "content policy violation"},
		{TaskID: "seven-day", UserId: user.Id, Action: constant.TaskActionFirstTailGenerate, SubmitTime: todayStart.AddDate(0, 0, -6).Unix(), Status: model.TaskStatusSuccess},
		{TaskID: "thirty-day-violation", UserId: user.Id, Action: constant.TaskActionReferenceGenerate, SubmitTime: todayStart.AddDate(0, 0, -29).Unix(), Status: model.TaskStatusFailure, FailReason: "安全审核不通过"},
		{TaskID: "outside-window", UserId: user.Id, Action: constant.TaskActionGenerate, SubmitTime: todayStart.AddDate(0, 0, -30).Unix(), Status: model.TaskStatusSuccess},
		{TaskID: "image-task", UserId: user.Id, Action: "volc_native_image_generation", SubmitTime: todayStart.Add(time.Hour).Unix(), Status: model.TaskStatusFailure, FailReason: "content policy violation"},
		{TaskID: "other-user", UserId: user.Id + 1, Action: constant.TaskActionGenerate, SubmitTime: todayStart.Add(time.Hour).Unix(), Status: model.TaskStatusFailure, FailReason: "content policy violation"},
	}
	require.NoError(t, model.DB.Create(&tasks).Error)

	stats, err := GetLoginNoticeStats(user.Id, now)
	require.NoError(t, err)
	assert.Equal(t, LoginNoticePeriodStats{Generated: 2, Violations: 1}, stats.Today)
	assert.Equal(t, LoginNoticePeriodStats{Generated: 3, Violations: 1}, stats.SevenDays)
	assert.Equal(t, LoginNoticePeriodStats{Generated: 4, Violations: 2}, stats.ThirtyDays)
}
