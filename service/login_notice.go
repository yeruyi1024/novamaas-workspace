package service

import (
	"strings"
	"time"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
)

type LoginNoticePeriodStats struct {
	Generated  int64 `json:"generated"`
	Violations int64 `json:"violations"`
}

type LoginNoticeStats struct {
	Today      LoginNoticePeriodStats `json:"today"`
	SevenDays  LoginNoticePeriodStats `json:"seven_days"`
	ThirtyDays LoginNoticePeriodStats `json:"thirty_days"`
}

type loginNoticeTaskRow struct {
	SubmitTime int64
	FailReason string
}

var videoTaskActions = []string{
	constant.TaskActionGenerate,
	constant.TaskActionTextGenerate,
	constant.TaskActionFirstTailGenerate,
	constant.TaskActionReferenceGenerate,
	constant.TaskActionRemix,
}

var videoViolationMarkers = []string{
	"violation",
	"prohibited",
	"inappropriate",
	"sensitive content",
	"content policy",
	"usage guideline",
	"safety filter",
	"safety policy",
	"safety_check_type",
	"moderation",
	"risk control",
	"ip infringement",
	"data inspection",
	"violation_fee.",
	"违规",
	"违禁",
	"敏感内容",
	"内容政策",
	"安全审核",
	"审核不通过",
	"风控",
}

func GetLoginNoticeStats(userID int, now time.Time) (LoginNoticeStats, error) {
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	sevenDayStart := todayStart.AddDate(0, 0, -6)
	thirtyDayStart := todayStart.AddDate(0, 0, -29)

	var stats LoginNoticeStats
	periods := []struct {
		start time.Time
		count *int64
	}{
		{start: todayStart, count: &stats.Today.Generated},
		{start: sevenDayStart, count: &stats.SevenDays.Generated},
		{start: thirtyDayStart, count: &stats.ThirtyDays.Generated},
	}
	for _, period := range periods {
		if err := model.DB.Model(&model.Task{}).
			Where("user_id = ? AND submit_time >= ? AND submit_time <= ?", userID, period.start.Unix(), now.Unix()).
			Where("action IN ?", videoTaskActions).
			Count(period.count).Error; err != nil {
			return LoginNoticeStats{}, err
		}
	}

	var rows []loginNoticeTaskRow
	if err := model.DB.Model(&model.Task{}).
		Select("submit_time", "fail_reason").
		Where("user_id = ? AND submit_time >= ? AND submit_time <= ?", userID, thirtyDayStart.Unix(), now.Unix()).
		Where("action IN ? AND status = ?", videoTaskActions, model.TaskStatusFailure).
		Find(&rows).Error; err != nil {
		return LoginNoticeStats{}, err
	}
	for _, row := range rows {
		if !IsVideoTaskViolation(row.FailReason) {
			continue
		}
		stats.ThirtyDays.Violations++
		if row.SubmitTime >= sevenDayStart.Unix() {
			stats.SevenDays.Violations++
		}
		if row.SubmitTime >= todayStart.Unix() {
			stats.Today.Violations++
		}
	}
	return stats, nil
}

func IsVideoTaskViolation(failureReason string) bool {
	reason := strings.ToLower(strings.TrimSpace(failureReason))
	if reason == "" {
		return false
	}
	for _, marker := range videoViolationMarkers {
		if strings.Contains(reason, marker) {
			return true
		}
	}
	return false
}
