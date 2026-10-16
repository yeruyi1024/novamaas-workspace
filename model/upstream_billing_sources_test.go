package model

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamFailedTaskUsageRequiresSamePeriodKnownConsumptionAndFullRefund(t *testing.T) {
	for _, tc := range []struct {
		name             string
		refundQuota      int
		consumeCost      int64
		adjustedConsume  int64
		refundCost       int64
		refundUser       int
		refundChannel    int
		refundOtherTask  bool
		refundAfterMonth bool
		missingSnapshot  string
		zeroDiscount     bool
		wantComplete     bool
	}{
		{name: "full customer and supplier refund", refundQuota: 100, consumeCost: 80, refundCost: -80, wantComplete: true},
		{name: "effective supplier adjustment closes the refund", refundQuota: 100, consumeCost: 80, adjustedConsume: 70, refundCost: -70, wantComplete: true},
		{name: "configured zero supplier cost with full customer refund", refundQuota: 100, zeroDiscount: true, wantComplete: true},
		{name: "task quota zero without any refund", consumeCost: 80},
		{name: "partial customer and supplier refund", refundQuota: 50, consumeCost: 80, refundCost: -40},
		{name: "supplier refunded but customer remainder", refundQuota: 50, consumeCost: 80, refundCost: -80},
		{name: "customer refunded but supplier remainder", refundQuota: 100, consumeCost: 80, refundCost: -70},
		{name: "consume cost missing", refundQuota: 100, consumeCost: 80, refundCost: -80, missingSnapshot: "consume"},
		{name: "refund cost missing", refundQuota: 100, consumeCost: 80, refundCost: -80, missingSnapshot: "refund"},
		{name: "unconfigured refund cost evidence", refundQuota: 100, consumeCost: 100, refundCost: -100, missingSnapshot: "unconfigured"},
		{name: "another user cannot refund this task", refundQuota: 100, consumeCost: 80, refundCost: -80, refundUser: 2},
		{name: "another channel cannot refund this task", refundQuota: 100, consumeCost: 80, refundCost: -80, refundChannel: 26},
		{name: "another task cannot refund this task", refundQuota: 100, consumeCost: 80, refundCost: -80, refundOtherTask: true},
		{name: "a later month cannot close this month", refundQuota: 100, consumeCost: 80, refundCost: -80, refundAfterMonth: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstreamBillingModelFixture(t)
			require.NoError(t, DB.AutoMigrate(&Log{}, &Task{}, &BillingEntry{}, &BillingCorrectionRow{}, &CostAccountingSnapshot{}, &CostAccountingAdjustment{}))
			start, end, err := BillingMonthBounds("2026-09")
			require.NoError(t, err)
			task := Task{TaskID: "failed-cost-proof", UserId: 1, ChannelId: 19, Platform: "54", Status: TaskStatusFailure, Quota: 0}
			task.SetData(map[string]any{"duration": 8, "status": "failed"})
			require.NoError(t, DB.Create(&task).Error)
			metadata := `{"is_task":true,"task_id":"failed-cost-proof"}`
			consume := Log{UserId: 1, Type: LogTypeConsume, CreatedAt: start + 1, ChannelId: 19, ModelName: "video", RequestId: "failed-consume", Quota: 100, Other: metadata}
			refund := Log{UserId: 1, Type: LogTypeRefund, CreatedAt: start + 2, ChannelId: 19, ModelName: "video", RequestId: "failed-refund", Quota: tc.refundQuota, Other: metadata}
			if tc.refundUser != 0 {
				refund.UserId = tc.refundUser
			}
			if tc.refundChannel != 0 {
				refund.ChannelId = tc.refundChannel
			}
			if tc.refundOtherTask {
				refund.Other = `{"is_task":true,"task_id":"different-failed-task"}`
			}
			if tc.refundAfterMonth {
				refund.CreatedAt = end
			}
			for _, event := range []struct {
				name string
				log  *Log
				cost int64
			}{{"consume", &consume, tc.consumeCost}, {"refund", &refund, tc.refundCost}} {
				if event.name == "refund" && tc.refundQuota == 0 {
					continue
				}
				require.NoError(t, LOG_DB.Create(event.log).Error)
				if tc.missingSnapshot == event.name {
					continue
				}
				discount := "0.8"
				if tc.zeroDiscount {
					discount = "0"
				}
				if tc.missingSnapshot == "unconfigured" && event.name == "refund" {
					discount = ""
				}
				basis := "100"
				if event.name == "refund" {
					basis = strconv.FormatInt(-int64(event.log.Quota), 10)
				}
				snapshot, err := BuildCostAccountingSnapshot(event.log, &CostAccountingInput{CostBasisQuota: basis, CostDiscount: discount, CostQuota: event.cost})
				require.NoError(t, err)
				require.NoError(t, DB.Create(snapshot).Error)
				if event.name == "consume" && tc.adjustedConsume > 0 {
					_, err := AdjustCostAccountingSnapshots([]CostAccountingAdjustmentTarget{{SnapshotID: snapshot.ID, NewCostQuota: tc.adjustedConsume}}, "Supplier refund reconciliation", 1, "refund-proof")
					require.NoError(t, err)
				}
			}
			var records []UpstreamBillingStatementRecord
			require.NoError(t, VisitUpstreamBillingSourceRecords(context.Background(), []int{19, 26}, start, end, func(record UpstreamBillingStatementRecord) error {
				records = append(records, record)
				return nil
			}))
			require.NotEmpty(t, records)
			var original *UpstreamBillingStatementRecord
			var requests, tokens int64
			for i := range records {
				if records[i].SourceLogID == consume.Id {
					original = &records[i]
				}
				requests += records[i].Requests
				tokens += records[i].VideoTokens
			}
			require.NotNil(t, original)
			assert.Equal(t, !tc.wantComplete, original.UsageMissing)
			assert.False(t, original.TaskPending, "a terminal failure is distinct from missing final usage")
			assert.Equal(t, int64(1), requests, "refunds never create another supplier request")
			assert.Zero(t, tokens)
			if tc.wantComplete {
				assert.Equal(t, "task_failed_refunded", original.UsageSource)
				assert.Zero(t, original.VideoSeconds, "failed response duration is not measured billable usage")
			} else {
				assert.NotEqual(t, "task_failed_refunded", original.UsageSource)
			}
		})
	}
}

func TestUpstreamSuccessfulTokenVideoDoesNotUseDurationAsFinalTokenEvidence(t *testing.T) {
	upstreamBillingModelFixture(t)
	require.NoError(t, DB.AutoMigrate(&Log{}, &Task{}, &BillingEntry{}, &BillingCorrectionRow{}, &CostAccountingSnapshot{}, &CostAccountingAdjustment{}))
	start, end, err := BillingMonthBounds("2026-09")
	require.NoError(t, err)
	task := Task{TaskID: "doubao-duration-only", UserId: 1, ChannelId: 19, Platform: "54", Status: TaskStatusSuccess, Quota: 0}
	task.SetData(map[string]any{"duration": 8, "usage": map[string]any{"duration": 8}})
	require.NoError(t, DB.Create(&task).Error)
	log := Log{UserId: 1, Type: LogTypeConsume, CreatedAt: start + 1, ChannelId: 19, ModelName: "video", RequestId: "doubao-duration-only", Quota: 100, Other: `{"is_task":true,"task_id":"doubao-duration-only"}`}
	require.NoError(t, LOG_DB.Create(&log).Error)
	snapshot, err := BuildCostAccountingSnapshot(&log, &CostAccountingInput{CostBasisQuota: "100", CostDiscount: "0.8", CostQuota: 80})
	require.NoError(t, err)
	require.NoError(t, DB.Create(snapshot).Error)
	var records []UpstreamBillingStatementRecord
	require.NoError(t, VisitUpstreamBillingSourceRecords(context.Background(), []int{19}, start, end, func(record UpstreamBillingStatementRecord) error {
		records = append(records, record)
		return nil
	}))
	require.Len(t, records, 1)
	assert.True(t, records[0].UsageMissing)
	assert.False(t, records[0].CostMissing)
	assert.Equal(t, int64(8), records[0].VideoSeconds)
	assert.Zero(t, records[0].VideoTokens)
	assert.Equal(t, "task_usage_missing", records[0].UsageSource)
}

func TestUpstreamBillingTaskIdentityCannotBeInferredFromNearbyTasks(t *testing.T) {
	upstreamBillingModelFixture(t)
	require.NoError(t, DB.AutoMigrate(&Log{}, &Task{}, &BillingEntry{}, &BillingCorrectionRow{}, &CostAccountingSnapshot{}, &CostAccountingAdjustment{}))
	start, end, err := BillingMonthBounds("2026-09")
	require.NoError(t, err)
	for index, id := range []string{"nearby-task-a", "nearby-task-b", "known-task"} {
		task := Task{TaskID: id, UserId: 1, ChannelId: 19, Platform: "54", Status: TaskStatusSuccess, Quota: 100, SubmitTime: start + int64(index+1)}
		task.SetData(map[string]any{"usage": map[string]any{"total_tokens": 100}, "duration": 4})
		require.NoError(t, DB.Create(&task).Error)
	}
	for index, metadata := range []string{
		`{"is_task":true}`,
		`{"is_task":true}`,
		`{"is_task":true,"task_id":"known-task"}`,
		`{"is_task":true,"task_id":"known-task"}`,
		`{"is_task":true}`,
		`{"is_task":true,"pre_consumed_quota":100}`,
	} {
		kind, basis, cost := LogTypeConsume, "100", int64(80)
		if index == 4 {
			kind, basis, cost = LogTypeRefund, "-100", -80
		}
		log := Log{UserId: 1, Type: kind, CreatedAt: start + int64(index+1), ChannelId: 19, ModelName: "video", RequestId: "identity-" + strconv.Itoa(index), Quota: 100, Other: metadata}
		require.NoError(t, LOG_DB.Create(&log).Error)
		snapshot, err := BuildCostAccountingSnapshot(&log, &CostAccountingInput{CostBasisQuota: basis, CostDiscount: "0.8", CostQuota: cost})
		require.NoError(t, err)
		require.NoError(t, DB.Create(snapshot).Error)
	}
	var records []UpstreamBillingStatementRecord
	require.NoError(t, VisitUpstreamBillingSourceRecords(context.Background(), []int{19}, start, end, func(record UpstreamBillingStatementRecord) error {
		records = append(records, record)
		return nil
	}))
	require.Len(t, records, 6)
	var requests, missing, tokens int64
	for _, record := range records {
		requests += record.Requests
		tokens += record.VideoTokens
		if record.UsageMissing {
			missing++
		}
		assert.False(t, record.CostMissing)
		if record.RequestID == "identity-0" || record.RequestID == "identity-1" {
			assert.Empty(t, record.TaskID, "nearby matching tasks do not establish a source identity")
			assert.Equal(t, int64(1), record.Requests)
			assert.True(t, record.UsageMissing)
			assert.Equal(t, "task_missing_id", record.UsageSource)
			assert.Zero(t, record.VideoTokens)
			assert.Zero(t, record.VideoSeconds)
		}
	}
	assert.Equal(t, int64(3), requests, "two unidentified original submissions plus one deduplicated identified task")
	assert.Equal(t, int64(2), missing, "each unidentified submission stays independently incomplete")
	assert.Equal(t, int64(100), tokens, "only the task with an explicit identity supplies final usage")
}
