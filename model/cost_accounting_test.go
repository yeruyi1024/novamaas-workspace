package model

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeCostDiscountPreservesExactSupportedPrecision(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "unconfigured", input: "", want: ""},
		{name: "integer one", input: "1", want: "1.000000"},
		{name: "six decimals", input: "0.876543", want: "0.876543"},
		{name: "zero", input: "0", want: "0.000000"},
		{name: "too precise", input: "0.1234567", wantErr: true},
		{name: "trailing seventh decimal", input: "0.8000000", wantErr: true},
		{name: "scientific notation", input: "8e-1", wantErr: true},
		{name: "negative", input: "-0.1", wantErr: true},
		{name: "above one", input: "1.000001", wantErr: true},
		{name: "not a number", input: "discount", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeCostDiscount(test.input)
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestAttachLogAccountingUsesAdjustedSnapshotAndDoesNotMatchAnotherUsersRequest(t *testing.T) {
	truncateTables(t)
	log := &Log{
		Id: 101, UserId: 7, Type: LogTypeConsume, CreatedAt: 123,
		Quota: 1000, RequestId: "shared-request", ModelName: "model-a", ChannelId: 3,
	}
	snapshot, err := BuildCostAccountingSnapshot(log, &CostAccountingInput{
		CostBasisQuota: "1000", CostDiscount: "0.5", CostQuota: 500,
	})
	require.NoError(t, err)
	require.NoError(t, DB.Create(snapshot).Error)
	require.NoError(t, DB.Create(&CostAccountingAdjustment{
		SnapshotID: snapshot.ID, DeltaCostQuota: -100, NewCostQuota: 400,
	}).Error)
	otherUser := &Log{
		Id: 202, UserId: 8, Type: LogTypeConsume, CreatedAt: 123,
		Quota: 1000, RequestId: "shared-request", ModelName: "model-a", ChannelId: 3,
	}
	refund := &Log{
		Id: 303, UserId: 7, Type: LogTypeRefund, CreatedAt: 124,
		Quota: 100, RequestId: "refund-request", ModelName: "model-a", ChannelId: 3,
	}

	require.NoError(t, AttachLogAccounting([]*Log{log, otherUser, refund}))
	require.NotNil(t, log.RevenueQuota)
	require.NotNil(t, log.CostQuota)
	require.NotNil(t, log.ProfitQuota)
	assert.Equal(t, int64(1000), *log.RevenueQuota)
	assert.Equal(t, int64(400), *log.CostQuota)
	assert.Equal(t, int64(600), *log.ProfitQuota)
	require.NotNil(t, otherUser.RevenueQuota)
	assert.Equal(t, int64(1000), *otherUser.RevenueQuota)
	assert.Nil(t, otherUser.CostQuota)
	assert.Nil(t, otherUser.ProfitQuota)
	require.NotNil(t, refund.RevenueQuota)
	assert.Equal(t, int64(-100), *refund.RevenueQuota)
	assert.Nil(t, refund.CostQuota)

	// ClickHouse and imported logs can have display IDs that differ from the
	// snapshot source ID. Match their accounting evidence by request metadata.
	logByRequest := &Log{
		Id: 999, UserId: 7, Type: LogTypeConsume, CreatedAt: 123,
		Quota: 1000, RequestId: "shared-request", ModelName: "model-a", ChannelId: 3,
	}
	require.NoError(t, AttachLogAccounting([]*Log{logByRequest}))
	require.NotNil(t, logByRequest.CostQuota)
	assert.Equal(t, int64(400), *logByRequest.CostQuota)
}

func TestCostAccountingSnapshotsAggregateRevenueCostProfitAndAdjustments(t *testing.T) {
	truncateTables(t)

	consume, err := BuildCostAccountingSnapshot(&Log{
		Id: 11, UserId: 7, Username: "finance", TokenName: "token", ModelName: "model-a",
		ChannelId: 3, Group: "default", Type: LogTypeConsume, CreatedAt: 100, Quota: 1000,
		RequestId: "request-finance-1", UpstreamRequestId: "upstream-finance-1",
	}, &CostAccountingInput{
		EventKey:       "task:abc:initial",
		CostBasisQuota: "1250",
		CostDiscount:   "0.6",
		CostQuota:      750,
		Source:         CostSnapshotSourceRealtime,
	})
	require.NoError(t, err)
	refund, err := BuildCostAccountingSnapshot(&Log{
		Id: 12, UserId: 7, Username: "finance", TokenName: "token", ModelName: "model-a",
		ChannelId: 3, Group: "default", Type: LogTypeRefund, CreatedAt: 200, Quota: 200,
	}, &CostAccountingInput{
		CostBasisQuota: "-250",
		CostDiscount:   "0.6",
		CostQuota:      -150,
		Source:         CostSnapshotSourceRealtime,
	})
	require.NoError(t, err)
	require.NoError(t, DB.Create(consume).Error)
	require.NoError(t, DB.Create(refund).Error)

	totals, err := SumCostAccounting(CostAccountingFilter{UserID: 7, StartTimestamp: 1, EndTimestamp: 300})
	require.NoError(t, err)
	assert.Equal(t, int64(1000), totals.ConsumptionQuota)
	assert.Equal(t, int64(200), totals.RefundQuota)
	assert.Equal(t, int64(800), totals.RevenueQuota)
	assert.Equal(t, int64(600), totals.CostQuota)
	assert.Equal(t, int64(200), totals.ProfitQuota)
	assert.Equal(t, int64(2), totals.Records)
	filtered, err := SumCostAccounting(CostAccountingFilter{
		RequestID: "request-finance-1", UpstreamRequestID: "upstream-finance-1",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), filtered.Records)
	assert.Equal(t, int64(750), filtered.CostQuota)

	assert.Equal(t, "task:abc:initial", consume.EventKey)
	adjustments, err := AdjustCostAccountingSnapshots([]CostAccountingAdjustmentTarget{{
		SnapshotID: consume.ID, NewCostQuota: 700,
	}}, "supplier reconciliation", 1, "batch-1")
	require.NoError(t, err)
	require.Len(t, adjustments, 1)
	adjustment := adjustments[0]
	assert.Equal(t, int64(-50), adjustment.DeltaCostQuota)
	assert.Equal(t, int64(700), adjustment.NewCostQuota)

	totals, err = SumCostAccounting(CostAccountingFilter{UserID: 7})
	require.NoError(t, err)
	assert.Equal(t, int64(550), totals.CostQuota)
	assert.Equal(t, int64(250), totals.ProfitQuota)

	views, total, err := ListCostAccountingSnapshots(CostAccountingFilter{UserID: 7}, 0, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, views, 2)
	assert.Equal(t, refund.ID, views[0].ID)
	assert.Equal(t, consume.ID, views[1].ID)
	assert.Equal(t, int64(-150), views[0].EffectiveCostQuota)
	assert.Equal(t, int64(-50), views[0].ProfitQuota)
	assert.Equal(t, int64(-50), views[1].AdjustmentQuota)
	assert.Equal(t, int64(700), views[1].EffectiveCostQuota)
	assert.Equal(t, int64(300), views[1].ProfitQuota)

	var storedAdjustments []CostAccountingAdjustment
	require.NoError(t, DB.Where("snapshot_id = ?", consume.ID).Order("id DESC").Find(&storedAdjustments).Error)
	require.Len(t, storedAdjustments, 1)
	assert.Equal(t, "supplier reconciliation", storedAdjustments[0].Reason)
}

func TestCostAccountingTotalsIgnoreLaterDuplicateForSameSourceLog(t *testing.T) {
	truncateTables(t)
	log := &Log{
		Id: 21, UserId: 7, ModelName: "task-model", ChannelId: 3, Group: "default",
		Type: LogTypeConsume, CreatedAt: 100, Quota: 1000, RequestId: "request-duplicate",
	}
	original, err := BuildCostAccountingSnapshot(log, &CostAccountingInput{
		EventKey:       "task:duplicate:initial",
		CostBasisQuota: "1000",
		CostQuota:      1000,
		Source:         CostSnapshotSourceRealtime,
	})
	require.NoError(t, err)
	duplicate, err := BuildCostAccountingSnapshot(log, &CostAccountingInput{
		CostBasisQuota: "1000",
		CostDiscount:   "0.8",
		CostQuota:      800,
		Source:         CostSnapshotSourceBackfill,
	})
	require.NoError(t, err)
	require.NoError(t, DB.Create(original).Error)
	require.NoError(t, DB.Create(duplicate).Error)

	totals, err := SumCostAccounting(CostAccountingFilter{UserID: 7})
	require.NoError(t, err)
	assert.Equal(t, int64(1), totals.Records)
	assert.Equal(t, int64(1), totals.LinkedRecords)
	assert.Equal(t, int64(1000), totals.RevenueQuota)
	assert.Equal(t, int64(1000), totals.CostQuota)
	assert.Zero(t, totals.ProfitQuota)

	views, total, err := ListCostAccountingSnapshots(CostAccountingFilter{UserID: 7}, 0, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, views, 1)
	assert.Equal(t, original.ID, views[0].ID)
}

func TestExistingCostSnapshotLogIndexesMatchesLogsWithoutDatabaseIDs(t *testing.T) {
	truncateTables(t)
	log := Log{
		UserId: 7, ModelName: "task-model", ChannelId: 3, Group: "default",
		Type: LogTypeConsume, CreatedAt: 100, Quota: 1000, RequestId: "request-without-id",
	}
	snapshot, err := BuildCostAccountingSnapshot(&log, &CostAccountingInput{
		EventKey:       "task:without-id:initial",
		CostBasisQuota: "1000",
		CostQuota:      1000,
		Source:         CostSnapshotSourceRealtime,
	})
	require.NoError(t, err)
	require.NoError(t, DB.Create(snapshot).Error)

	existing, err := ExistingCostSnapshotLogIndexes([]Log{log})
	require.NoError(t, err)
	assert.Contains(t, existing, 0)
}

func TestReconcileCostAccountingTotalsUsesLogTurnoverAsProfitBasis(t *testing.T) {
	totals := ReconcileCostAccountingTotals(
		LogStatistics{Quota: 1000, RefundQuota: 200, RevenueQuota: 999, Records: 2},
		CostAccountingTotals{RevenueQuota: 900, CostQuota: 600, ProfitQuota: 300},
	)

	assert.Equal(t, int64(1000), totals.ConsumptionQuota)
	assert.Equal(t, int64(200), totals.RefundQuota)
	assert.Equal(t, int64(800), totals.RevenueQuota)
	assert.Equal(t, int64(600), totals.CostQuota)
	assert.Equal(t, int64(200), totals.ProfitQuota)
	assert.Equal(t, int64(2), totals.UsageRecords)
	assert.Equal(t, int64(2), totals.MissingCostRecords)
	assert.False(t, totals.AccountingComplete)

	complete := ReconcileCostAccountingTotals(
		LogStatistics{Quota: 1000, RefundQuota: 200, Records: 2},
		CostAccountingTotals{CostQuota: 600, Records: 2, LinkedRecords: 2},
	)
	assert.True(t, complete.AccountingComplete)
	assert.Zero(t, complete.MissingCostRecords)
}

func TestUnconfiguredChannelDefaultsOnlyMissingSnapshotCostToTurnover(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.Create(&Channel{Id: 12, Name: "break-even", Key: "key", CostDiscount: ""}).Error)
	logs := []Log{
		{UserId: 7, ChannelId: 12, Type: LogTypeConsume, CreatedAt: 100, Quota: 1000, RequestId: "break-even-consume"},
		{UserId: 7, ChannelId: 12, Type: LogTypeRefund, CreatedAt: 200, Quota: 200, RequestId: "break-even-refund"},
	}
	require.NoError(t, LOG_DB.Create(&logs).Error)
	consumeSnapshot, err := BuildCostAccountingSnapshot(&logs[0], &CostAccountingInput{
		CostBasisQuota: "1000",
		CostQuota:      1000,
	})
	require.NoError(t, err)
	require.NoError(t, DB.Create(consumeSnapshot).Error)

	statistics, err := SumLogStatistics(CostAccountingFilter{UserID: 7})
	require.NoError(t, err)
	totals, err := SumCostAccounting(CostAccountingFilter{UserID: 7})
	require.NoError(t, err)
	totals = ReconcileCostAccountingTotals(statistics, totals)

	assert.Equal(t, int64(800), totals.RevenueQuota)
	assert.Equal(t, int64(800), totals.CostQuota)
	assert.Zero(t, totals.ProfitQuota)
	assert.Equal(t, int64(1), totals.DefaultedRecords)
	assert.Equal(t, int64(-200), totals.DefaultedCostQuota)
	assert.Zero(t, totals.MissingCostRecords)
	assert.True(t, totals.AccountingComplete)

	buckets, err := SumCostAccountingBuckets(CostAccountingFilter{UserID: 7}, 3600, 0)
	require.NoError(t, err)
	require.Len(t, buckets, 1)
	assert.Equal(t, int64(800), buckets[0].RevenueQuota)
	assert.Equal(t, int64(800), buckets[0].CostQuota)
	assert.Zero(t, buckets[0].ProfitQuota)
}

func TestChannelUpdateClearingCostDiscountCanClearConfiguredValue(t *testing.T) {
	truncateTables(t)
	channel := &Channel{Id: 13, Name: "clear-cost", Key: "key", CostDiscount: "0.8"}
	require.NoError(t, DB.Create(channel).Error)

	channel.Name = "renamed-without-cost-change"
	channel.CostDiscount = ""
	require.NoError(t, channel.Update())

	var stored Channel
	require.NoError(t, DB.First(&stored, channel.Id).Error)
	assert.Equal(t, "0.8", stored.CostDiscount)

	channel.CostDiscount = ""
	require.NoError(t, channel.UpdateClearingCostDiscount())

	require.NoError(t, DB.First(&stored, channel.Id).Error)
	assert.Empty(t, stored.CostDiscount)
}

func TestBuildCostAccountingSnapshotRejectsInvalidEvidence(t *testing.T) {
	log := &Log{Id: 1, Type: LogTypeConsume, Quota: 100, CreatedAt: 1}

	_, err := BuildCostAccountingSnapshot(log, &CostAccountingInput{
		CostBasisQuota: "invalid",
		CostDiscount:   "1",
		CostQuota:      100,
	})
	require.ErrorContains(t, err, "cost basis quota")

	_, err = BuildCostAccountingSnapshot(log, &CostAccountingInput{
		CostBasisQuota: "100",
		CostDiscount:   "0.1234567",
		CostQuota:      100,
	})
	require.ErrorContains(t, err, "6 decimal places")

	_, err = BuildCostAccountingSnapshot(log, &CostAccountingInput{
		CostBasisQuota: "-100",
		CostDiscount:   "1",
		CostQuota:      -100,
	})
	require.ErrorContains(t, err, "consumption")

	_, err = BuildCostAccountingSnapshot(&Log{Id: 2, Type: LogTypeRefund, Quota: 100, CreatedAt: 1}, &CostAccountingInput{
		CostBasisQuota: "100",
		CostDiscount:   "1",
		CostQuota:      100,
	})
	require.ErrorContains(t, err, "refund")
}

func TestBuildCostAccountingSnapshotAcceptsBreakEvenEvidenceWithoutDiscount(t *testing.T) {
	snapshot, err := BuildCostAccountingSnapshot(&Log{
		Id: 1, Type: LogTypeConsume, Quota: 8600, CreatedAt: 1,
	}, &CostAccountingInput{
		CostBasisQuota: "8600",
		CostQuota:      8600,
	})
	require.NoError(t, err)
	require.NotNil(t, snapshot)
	assert.Empty(t, snapshot.CostDiscount)
	assert.Equal(t, int64(8600), snapshot.CostQuota)

	_, err = BuildCostAccountingSnapshot(&Log{
		Id: 2, Type: LogTypeConsume, Quota: 8600, CreatedAt: 1,
	}, &CostAccountingInput{
		CostBasisQuota: "10000",
		CostQuota:      10000,
	})
	require.ErrorContains(t, err, "must use revenue as cost")
}

func TestCostSnapshotRespectsDisabledUsageLogging(t *testing.T) {
	truncateTables(t)
	previous := common.LogConsumeEnabled
	common.LogConsumeEnabled = false
	t.Cleanup(func() { common.LogConsumeEnabled = previous })

	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Set("username", "finance")
	context.Set(common.RequestIdKey, "request-without-log")
	RecordConsumeLog(context, 8, RecordConsumeLogParams{
		ChannelId: 2,
		ModelName: "model-a",
		Quota:     1000,
		Group:     "default",
		CostAccounting: &CostAccountingInput{
			CostBasisQuota: "500",
			CostDiscount:   "0.8",
			CostQuota:      400,
		},
	})

	var logCount int64
	require.NoError(t, LOG_DB.Model(&Log{}).Count(&logCount).Error)
	assert.Zero(t, logCount)
	totals, err := SumCostAccounting(CostAccountingFilter{RequestID: "request-without-log"})
	require.NoError(t, err)
	assert.Zero(t, totals.Records)
	assert.Zero(t, totals.RevenueQuota)
	assert.Zero(t, totals.CostQuota)
}
