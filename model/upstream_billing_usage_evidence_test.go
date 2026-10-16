package model

import (
	"context"
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamBillingAliFinalDurationUsesPersistedProviderAndReportedUsage(t *testing.T) {
	for _, test := range []struct {
		name, platform, data string
		status               TaskStatus
		seconds              int64
		missing              bool
	}{
		{name: "reported Ali seconds", platform: "17", status: TaskStatusSuccess, data: `{"usage":{"duration":5}}`, seconds: 5},
		{name: "integral numeric Ali seconds", platform: "17", status: TaskStatusSuccess, data: `{"usage":{"duration":5.0}}`, seconds: 5},
		{name: "nested reported Ali seconds", platform: "17", status: TaskStatusSuccess, data: `{"data":{"usage":{"duration":8}}}`, seconds: 8},
		{name: "Ali maximum seconds", platform: "17", status: TaskStatusSuccess, data: fmt.Sprintf(`{"usage":{"duration":%d}}`, relaycommon.MaxTaskDurationSeconds), seconds: relaycommon.MaxTaskDurationSeconds},
		{name: "Ali request-like top-level duration is insufficient", platform: "17", status: TaskStatusSuccess, data: `{"duration":5}`, seconds: 5, missing: true},
		{name: "Ali omitted usage duration", platform: "17", status: TaskStatusSuccess, data: `{"usage":{}}`, missing: true},
		{name: "Ali zero reported seconds", platform: "17", status: TaskStatusSuccess, data: `{"usage":{"duration":0}}`, missing: true},
		{name: "Ali negative reported seconds", platform: "17", status: TaskStatusSuccess, data: `{"usage":{"duration":-1}}`, missing: true},
		{name: "Ali fractional seconds cannot be truncated", platform: "17", status: TaskStatusSuccess, data: `{"usage":{"duration":5.5}}`, missing: true},
		{name: "Ali unbounded reported seconds", platform: "17", status: TaskStatusSuccess, data: fmt.Sprintf(`{"usage":{"duration":%d}}`, relaycommon.MaxTaskDurationSeconds+1), missing: true},
		{name: "Ali failed duration does not prove free supplier cost", platform: "17", status: TaskStatusFailure, data: `{"usage":{"duration":5}}`, missing: true},
		{name: "Doubao duration-only remains incomplete", platform: fmt.Sprint(constant.ChannelTypeDoubaoVideo), status: TaskStatusSuccess, data: `{"usage":{"duration":5}}`, missing: true},
		{name: "missing platform cannot inherit current channel identity", status: TaskStatusSuccess, data: `{"usage":{"duration":5}}`, missing: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			upstreamBillingModelFixture(t)
			require.NoError(t, DB.AutoMigrate(&Log{}, &Task{}, &BillingEntry{}, &BillingCorrectionRow{}, &CostAccountingSnapshot{}, &CostAccountingAdjustment{}))
			start, end, err := BillingMonthBounds("2026-08")
			require.NoError(t, err)
			// Today's channel configuration and per-call pricing cannot establish
			// either provider identity or measured final usage for a historical task.
			require.NoError(t, DB.Create(&Channel{Id: 19, Type: constant.ChannelTypeAli, Key: "private"}).Error)
			task := Task{TaskID: "usage-task", UserId: 1, ChannelId: 19, Platform: constant.TaskPlatform(test.platform), Status: test.status, Data: []byte(test.data), PrivateData: TaskPrivateData{BillingContext: &TaskBillingContext{PerCallBilling: true, OtherRatios: map[string]float64{"seconds": 50}}}}
			require.NoError(t, DB.Create(&task).Error)
			log := Log{UserId: 1, ChannelId: 19, CreatedAt: start, Type: LogTypeConsume, ModelName: "video", Quota: 100, RequestId: "usage-submit", Other: `{"is_task":true,"task_id":"usage-task"}`}
			require.NoError(t, LOG_DB.Create(&log).Error)
			cost, err := BuildCostAccountingSnapshot(&log, &CostAccountingInput{CostBasisQuota: "100", CostDiscount: "0.8", CostQuota: 80})
			require.NoError(t, err)
			require.NoError(t, DB.Create(cost).Error)
			var records []UpstreamBillingStatementRecord
			require.NoError(t, VisitUpstreamBillingSourceRecords(context.Background(), []int{19}, start, end, func(record UpstreamBillingStatementRecord) error { records = append(records, record); return nil }))
			require.Len(t, records, 1)
			assert.Equal(t, test.missing, records[0].UsageMissing)
			assert.Equal(t, test.seconds, records[0].VideoSeconds)
			assert.Zero(t, records[0].VideoTokens)
			assert.Equal(t, int64(1), records[0].Requests)
			assert.False(t, records[0].CostMissing)
			if !test.missing {
				assert.Equal(t, "task_duration_usage", records[0].UsageSource)
			}
		})
	}
}

func TestUpstreamBillingFailedRefundProofReadsLaterPagesAndHonorsSourceWatermarks(t *testing.T) {
	for _, late := range []bool{false, true} {
		t.Run(fmt.Sprintf("refund_inserted_after_watermark_%v", late), func(t *testing.T) {
			upstreamBillingModelFixture(t)
			require.NoError(t, DB.AutoMigrate(&Log{}, &Task{}, &BillingEntry{}, &BillingCorrectionRow{}, &CostAccountingSnapshot{}, &CostAccountingAdjustment{}))
			start, end, err := BillingMonthBounds("2026-08")
			require.NoError(t, err)
			task := Task{TaskID: "failed-late-page", UserId: 1, ChannelId: 19, Status: TaskStatusFailure, Data: []byte(`{"status":"failed","duration":8}`)}
			require.NoError(t, DB.Create(&task).Error)
			original := Log{Id: 1, UserId: 1, ChannelId: 19, CreatedAt: start, Type: LogTypeConsume, ModelName: "video", Quota: 100, RequestId: "original", Other: `{"is_task":true,"task_id":"failed-late-page"}`}
			require.NoError(t, LOG_DB.Create(&original).Error)
			cost, err := BuildCostAccountingSnapshot(&original, &CostAccountingInput{CostBasisQuota: "100", CostDiscount: "0.8", CostQuota: 80})
			require.NoError(t, err)
			require.NoError(t, DB.Create(cost).Error)
			// Exact page-boundary fixture: the matching refund belongs to page 2.
			filler := make([]Log, 999)
			for i := range filler {
				filler[i] = Log{Id: i + 2, UserId: 1, ChannelId: 19, CreatedAt: start + 1, Type: LogTypeConsume, ModelName: "text", RequestId: fmt.Sprintf("unrelated-%d", i), PromptTokens: 1}
			}
			require.NoError(t, LOG_DB.CreateInBatches(&filler, 500).Error)
			refund := Log{Id: 1001, UserId: 1, ChannelId: 19, CreatedAt: start + 2, Type: LogTypeRefund, ModelName: "video", Quota: 100, RequestId: "refund", Other: original.Other}
			refundCost, err := BuildCostAccountingSnapshot(&refund, &CostAccountingInput{CostBasisQuota: "-100", CostDiscount: "0.8", CostQuota: -80})
			require.NoError(t, err)
			if !late {
				require.NoError(t, LOG_DB.Create(&refund).Error)
				require.NoError(t, DB.Create(refundCost).Error)
			}
			var originalRecord UpstreamBillingStatementRecord
			refundCount := 0
			require.NoError(t, VisitUpstreamBillingSourceRecords(context.Background(), []int{19}, start, end, func(record UpstreamBillingStatementRecord) error {
				if record.SourceLogID == original.Id {
					originalRecord = record
					if late {
						require.NoError(t, LOG_DB.Create(&refund).Error)
						require.NoError(t, DB.Create(refundCost).Error)
					} else {
						// An adjustment committed after this traversal's watermark
						// cannot change either the proof or the later refund record.
						require.NoError(t, DB.Create(&CostAccountingAdjustment{SnapshotID: refundCost.ID, DeltaCostQuota: 1, NewCostQuota: -79}).Error)
					}
				}
				if record.SourceLogID == refund.Id {
					refundCount++
					assert.Equal(t, int64(-80), record.CostQuota)
					assert.False(t, record.UsageMissing)
				}
				return nil
			}))
			assert.Equal(t, late, originalRecord.UsageMissing)
			assert.Zero(t, originalRecord.VideoTokens)
			if late {
				assert.Zero(t, refundCount)
			} else {
				assert.Zero(t, originalRecord.VideoSeconds, "a fully refunded failure has no measured output usage to report")
				assert.Equal(t, "task_failed_refunded", originalRecord.UsageSource)
				assert.Equal(t, 1, refundCount)
			}
		})
	}
}

func TestUpstreamBillingMissingTaskIDsRemainSeparateIncompleteRequests(t *testing.T) {
	seen := make(map[string]bool)
	for index := 0; index < 2; index++ {
		log := Log{UserId: 1, ChannelId: 19, Type: LogTypeConsume, Other: `{"is_task":true}`}
		record := UpstreamBillingStatementRecord{}
		populateUpstreamBillingUsage(&log, &record, nil, seen, nil)
		assert.True(t, record.UsageMissing)
		assert.Equal(t, "task_missing_id", record.UsageSource)
		assert.Equal(t, int64(1), record.Requests)
	}
	log := Log{UserId: 1, ChannelId: 19, Type: LogTypeConsume, Other: `{"is_task":true,"task_id":"known-task"}`}
	tasks := map[string][]Task{fmtUpstreamTaskIdentity(1, 19, "known-task"): {{Status: TaskStatusSuccess, Data: []byte(`{"usage":{"total_tokens":123}}`)}}}
	first, duplicate := UpstreamBillingStatementRecord{}, UpstreamBillingStatementRecord{}
	populateUpstreamBillingUsage(&log, &first, tasks, seen, nil)
	populateUpstreamBillingUsage(&log, &duplicate, tasks, seen, nil)
	assert.Equal(t, int64(1), first.Requests)
	assert.Equal(t, int64(123), first.VideoTokens)
	assert.False(t, first.UsageMissing)
	assert.Zero(t, duplicate.Requests)
	assert.Zero(t, duplicate.VideoTokens)
	assert.False(t, duplicate.UsageMissing)
}

func TestUpstreamBillingFailedRefundProofCannotUseCostAdjustmentsAfterReportWatermark(t *testing.T) {
	upstreamBillingModelFixture(t)
	require.NoError(t, DB.AutoMigrate(&Log{}, &Task{}, &BillingEntry{}, &BillingCorrectionRow{}, &CostAccountingSnapshot{}, &CostAccountingAdjustment{}))
	start, end, err := BillingMonthBounds("2026-08")
	require.NoError(t, err)
	filler := make([]Log, 1000)
	for i := range filler {
		filler[i] = Log{Id: i + 1, UserId: 1, ChannelId: 19, CreatedAt: start, Type: LogTypeConsume, ModelName: "text", RequestId: fmt.Sprintf("before-failure-%d", i), PromptTokens: 1}
	}
	require.NoError(t, LOG_DB.CreateInBatches(&filler, 500).Error)
	require.NoError(t, DB.Create(&Task{TaskID: "late-proof", UserId: 1, ChannelId: 19, Status: TaskStatusFailure, Data: []byte(`{"status":"failed"}`)}).Error)
	logs := []Log{
		{Id: 1001, UserId: 1, ChannelId: 19, CreatedAt: start + 1, Type: LogTypeConsume, ModelName: "video", RequestId: "late-proof-consume", Quota: 100, Other: `{"is_task":true,"task_id":"late-proof"}`},
		{Id: 1002, UserId: 1, ChannelId: 19, CreatedAt: start + 2, Type: LogTypeRefund, ModelName: "video", RequestId: "late-proof-refund", Quota: 100, Other: `{"is_task":true,"task_id":"late-proof"}`},
	}
	require.NoError(t, LOG_DB.Create(&logs).Error)
	consume, err := BuildCostAccountingSnapshot(&logs[0], &CostAccountingInput{CostBasisQuota: "100", CostDiscount: "0.8", CostQuota: 80})
	require.NoError(t, err)
	refund, err := BuildCostAccountingSnapshot(&logs[1], &CostAccountingInput{CostBasisQuota: "-100", CostDiscount: "0.8", CostQuota: -79})
	require.NoError(t, err)
	require.NoError(t, DB.Create(consume).Error)
	require.NoError(t, DB.Create(refund).Error)
	var failed UpstreamBillingStatementRecord
	require.NoError(t, VisitUpstreamBillingSourceRecords(context.Background(), []int{19}, start, end, func(record UpstreamBillingStatementRecord) error {
		if record.SourceLogID == 1 {
			// The failed task will first be discovered on the next page. This
			// adjustment makes today's costs close, but postdates this report.
			require.NoError(t, DB.Create(&CostAccountingAdjustment{SnapshotID: refund.ID, DeltaCostQuota: -1, NewCostQuota: -80}).Error)
		}
		if record.SourceLogID == logs[0].Id {
			failed = record
		}
		return nil
	}))
	assert.True(t, failed.UsageMissing)
	assert.Equal(t, "task_usage_missing", failed.UsageSource)
	assert.Equal(t, int64(80), failed.CostQuota)
}

func TestUpstreamBillingFailedRefundExemptionPreservesReportedUsageAndRejectsInvalidEvidence(t *testing.T) {
	for _, test := range []struct {
		name, data, source string
		tokens, seconds    int64
		missing            bool
	}{
		{name: "only missing tokens are exempt", data: `{"status":"failed","duration":8}`, source: "task_failed_refunded"},
		{name: "reported tokens remain real usage", data: `{"usage":{"total_tokens":123},"duration":8}`, source: "task_final_usage", tokens: 123, seconds: 8},
		{name: "negative tokens cannot be exempt", data: `{"usage":{"total_tokens":-1},"duration":8}`, source: "task_invalid_usage", missing: true},
		{name: "invalid response cannot be exempt", data: `{"usage":`, source: "task_invalid_usage", missing: true},
		{name: "invalid seconds cannot be exempt", data: `{"duration":-1}`, source: "task_invalid_usage", missing: true},
		{name: "invalid reported duration cannot be exempt", data: `{"usage":{"duration":-1}}`, source: "task_invalid_usage", missing: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			key := fmtUpstreamTaskIdentity(1, 19, "closed-failure")
			tasks := map[string][]Task{key: {{Status: TaskStatusFailure, Data: []byte(test.data)}}}
			log := Log{UserId: 1, ChannelId: 19, Type: LogTypeConsume, Other: `{"is_task":true,"task_id":"closed-failure"}`}
			record := UpstreamBillingStatementRecord{}
			populateUpstreamBillingUsage(&log, &record, tasks, make(map[string]bool), map[string]bool{key: true})
			assert.Equal(t, test.missing, record.UsageMissing)
			assert.Equal(t, test.source, record.UsageSource)
			assert.Equal(t, test.tokens, record.VideoTokens)
			assert.Equal(t, test.seconds, record.VideoSeconds)
		})
	}
}
