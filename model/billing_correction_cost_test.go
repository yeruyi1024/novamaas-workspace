package model

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBillingCorrectionReversalRestoresAmountOnlySupplierAdjustment(t *testing.T) {
	fixture := seedBillingCorrectionCosts(t, "doubao-seedance-2-5", "1080p", true, false)
	_, err := AdjustCostAccountingSnapshots([]CostAccountingAdjustmentTarget{{SnapshotID: fixture.Snapshots[0].ID, NewCostQuota: fixture.Snapshots[0].CostQuota - 7}}, "Reconcile supplier amount before customer correction", 1, "supplier-amount")
	require.NoError(t, err)
	batch, err := PreviewBillingCorrection(fixture.Input, 1, "", "customer_group")
	require.NoError(t, err)
	require.True(t, batch.CanApply)
	for _, phase := range []struct {
		name    string
		apply   bool
		reverse bool
		want    int64
	}{
		{name: "amount-only supplier correction", want: fixture.OriginalCost - 7},
		{name: "customer model correction", apply: true, want: 26943},
		{name: "customer correction reversed", apply: true, reverse: true, want: fixture.OriginalCost - 7},
	} {
		if phase.apply {
			_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "customer_group", 1, phase.reverse, "Restore reconciled supplier amount")
			require.NoError(t, err)
		}
		records := make([]UpstreamBillingStatementRecord, 0)
		require.NoError(t, VisitUpstreamBillingSourceRecords(context.Background(), []int{904}, fixture.MonthStart, fixture.MonthEnd, func(record UpstreamBillingStatementRecord) error {
			records = append(records, record)
			return nil
		}))
		require.Len(t, records, 2, phase.name)
		var cost, requests, videoTokens int64
		for _, record := range records {
			assert.False(t, record.CostMissing, phase.name)
			assert.False(t, record.UsageMissing, phase.name)
			assert.Equal(t, "0.820000", record.CostDiscount, phase.name)
			assert.Contains(t, []int64{fixture.Snapshots[0].ID, fixture.Snapshots[1].ID}, record.CostSnapshotID)
			cost += record.CostQuota
			requests += record.Requests
			videoTokens += record.VideoTokens
		}
		assert.Equal(t, phase.want, cost, phase.name)
		assert.Equal(t, int64(1), requests, phase.name)
		assert.Equal(t, int64(10000), videoTokens, phase.name)
	}
	var originals []CostAccountingSnapshot
	require.NoError(t, DB.Where("id IN ?", []int64{fixture.Snapshots[0].ID, fixture.Snapshots[1].ID}).Order("id asc").Find(&originals).Error)
	assert.Equal(t, fixture.Snapshots, originals, "supplier and customer corrections append evidence without changing the originals")
}

type billingCorrectionCostFixture struct {
	Input                BillingCorrectionInput
	Task                 Task
	Sources              []BillingEntry
	Snapshots            []CostAccountingSnapshot
	MonthStart, MonthEnd int64
	ExecutionMonthStart  int64
	OriginalCost         int64
}

type billingCorrectionCostFixtureOptions struct {
	Discount       string
	NoSourceLogIDs bool
}

func seedBillingCorrectionCosts(t *testing.T, modelName, resolution string, hasVideo, legacyCost bool, options ...billingCorrectionCostFixtureOptions) billingCorrectionCostFixture {
	t.Helper()
	truncateTables(t)
	location := time.FixedZone("Asia/Shanghai", 8*3600)
	now := time.Now().In(location)
	executionStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
	monthStart := executionStart.AddDate(0, -1, 0)
	posted := monthStart.Add(49 * time.Hour).Unix()
	const submissionQuota = 1075000
	const finalQuota = 43000
	const wallet = 2000000
	var option billingCorrectionCostFixtureOptions
	if len(options) > 0 {
		option = options[0]
	}
	discount := "0.820000"
	if option.Discount != "" {
		discount = option.Discount
	}
	ratio := 5.0
	configureBillingCorrectionPricing(t, modelName, nil, &ratio)
	require.NoError(t, DB.Create(&User{Id: 901, Username: "cost_correction_customer", Group: "customer_group", Quota: wallet, UsedQuota: finalQuota, Status: common.UserStatusEnabled}).Error)
	require.NoError(t, DB.Create(&BillingAccount{UserID: 901, AccountingStartAt: monthStart.Unix(), Sequence: 2}).Error)
	require.NoError(t, DB.Create(&Token{Id: 903, UserId: 901, Name: "cost-test-token", Key: "cost-test-key", RemainQuota: wallet, UsedQuota: finalQuota}).Error)
	require.NoError(t, DB.Create(&Channel{Id: 904, Name: "cost-test-provider", CostDiscount: discount, UsedQuota: finalQuota}).Error)
	task := Task{TaskID: "cost-correction-task", UserId: 901, Group: "spe_sd_86", ChannelId: 904, Quota: finalQuota, Status: TaskStatusSuccess,
		PrivateData: TaskPrivateData{TokenId: 903, BillingContext: &TaskBillingContext{OriginModelName: modelName, ModelPrice: -1, ModelRatio: ratio, GroupRatio: .86}},
	}
	contentType := "text"
	if hasVideo {
		contentType = "video_url"
	}
	body, err := common.Marshal(map[string]any{"resolution": resolution, "content": []map[string]any{{"type": contentType}}})
	require.NoError(t, err)
	task.Properties.RequestBody = body
	task.SetData(map[string]any{"resolution": resolution, "usage": map[string]any{"total_tokens": 10000}})
	require.NoError(t, DB.Create(&task).Error)
	fixture := billingCorrectionCostFixture{Task: task, MonthStart: monthStart.Unix(), MonthEnd: executionStart.Unix(), ExecutionMonthStart: executionStart.Unix(),
		Input: BillingCorrectionInput{UserID: 901, StartAt: posted - 60, EndAt: posted + 3600, Models: []string{modelName}, Mode: BillingCorrectionModelPricing, Reason: "Reprice sales and historical supplier cost"},
	}
	for i, amount := range []int64{submissionQuota, finalQuota - submissionQuota} {
		kind, logType, absolute := "usage", LogTypeConsume, amount
		if amount < 0 {
			kind, logType, absolute = "task_adjustment", LogTypeRefund, -amount
		}
		meta, err := common.Marshal(map[string]any{"task_id": task.TaskID, "group_ratio": .86, "model_ratio": ratio, "model_price": -1})
		require.NoError(t, err)
		log := Log{UserId: 901, Username: "cost_correction_customer", Type: logType, CreatedAt: posted + int64(i), ModelName: modelName, Quota: int(absolute), Group: task.Group, TokenId: 903, ChannelId: 904, Other: string(meta), RequestId: fmt.Sprintf("cost-source-%d", i)}
		require.NoError(t, LOG_DB.Create(&log).Error)
		entry := BillingEntry{EventKey: fmt.Sprintf("cost-source-%d", i), UserID: 901, Sequence: int64(i + 1), PostedAt: log.CreatedAt, Kind: kind, ModelName: modelName, Quota: amount, TokenID: 903, RequestID: task.TaskID, SourceLogID: log.Id}
		if option.NoSourceLogIDs {
			entry.SourceLogID = 0
			if entry.Kind == "usage" {
				entry.RequestID = log.RequestId
			}
		}
		require.NoError(t, DB.Create(&entry).Error)
		fixture.Sources = append(fixture.Sources, entry)
		basis, snapshotDiscount := decimal.NewFromInt(amount).Div(decimal.RequireFromString("0.86")), discount
		cost := basis.Mul(decimal.RequireFromString(snapshotDiscount))
		if legacyCost {
			basis, cost, snapshotDiscount = decimal.NewFromInt(amount), decimal.NewFromInt(amount), ""
		}
		quota, err := common.QuotaFromDecimalStrict(cost)
		require.NoError(t, err)
		snapshot, err := BuildCostAccountingSnapshot(&log, &CostAccountingInput{EventKey: fmt.Sprintf("immutable-cost-%d", i), CostBasisQuota: basis.StringFixed(6), CostDiscount: snapshotDiscount, CostQuota: int64(quota)})
		require.NoError(t, err)
		require.NotNil(t, snapshot)
		require.NoError(t, DB.Create(snapshot).Error)
		fixture.Snapshots = append(fixture.Snapshots, *snapshot)
		fixture.OriginalCost += snapshot.CostQuota
	}
	return fixture
}

func TestBillingModelCorrectionRepricesSeedanceSupplierCostAndKeepsSnapshotsImmutable(t *testing.T) {
	for _, tc := range []struct {
		name, model, resolution string
		video, legacyCost       bool
		finalSales, finalCost   int64
	}{
		{"720p_without_video_missing_discount", "doubao-seedance-2-5", "720p", false, true, 43000, 41000},
		{"720p_with_video", "doubao-seedance-2-5-260628", "720p", true, false, 25800, 24600},
		{"1080p_without_video", "doubao-seedance-2-5", "1080p", false, false, 47300, 45100},
		{"1080p_with_video", "doubao-seedance-2-5-260628", "1080p", true, false, 28257, 26943},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := seedBillingCorrectionCosts(t, tc.model, tc.resolution, tc.video, tc.legacyCost)
			filter := CostAccountingFilter{UserID: 901, ModelName: tc.model, StartTimestamp: fixture.MonthStart, EndTimestamp: fixture.MonthEnd - 1}
			before, err := SumCostAccounting(filter)
			require.NoError(t, err)
			assert.Equal(t, fixture.OriginalCost, before.CostQuota)
			batch, err := PreviewBillingCorrection(fixture.Input, 1, "", "customer_group")
			require.NoError(t, err)
			require.True(t, batch.CanApply, "a cost-only correction must remain reviewable when sales already use the correct price")
			assert.Equal(t, tc.finalSales-43000, batch.NetDelta)
			previewCosts, err := SumCostAccounting(filter)
			require.NoError(t, err)
			assert.Equal(t, before.CostQuota, previewCosts.CostQuota, "preview must not change supplier cost")
			for i := 0; i < 2; i++ {
				_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "customer_group", 1, false, "")
				require.NoError(t, err)
			}
			var user User
			require.NoError(t, DB.First(&user, 901).Error)
			assert.Equal(t, int64(2000000)-batch.NetDelta, int64(user.Quota), "supplier cost never changes the customer's wallet")
			assert.Equal(t, tc.finalSales, int64(user.UsedQuota))
			costs, err := SumCostAccounting(filter)
			require.NoError(t, err)
			assert.Equal(t, tc.finalCost, costs.CostQuota)
			assert.Equal(t, tc.finalSales, costs.RevenueQuota)
			assert.Equal(t, tc.finalSales-tc.finalCost, costs.ProfitQuota)
			var originals []CostAccountingSnapshot
			require.NoError(t, DB.Where("id IN ?", []int64{fixture.Snapshots[0].ID, fixture.Snapshots[1].ID}).Order("id asc").Find(&originals).Error)
			assert.Equal(t, fixture.Snapshots, originals, "original supplier evidence must remain immutable")
			var corrections []CostAccountingAdjustment
			require.NoError(t, DB.Where("snapshot_id IN ?", []int64{fixture.Snapshots[0].ID, fixture.Snapshots[1].ID}).Find(&corrections).Error)
			require.NotEmpty(t, corrections, "cost corrections must append audit entries")
			var costDelta int64
			for _, correction := range corrections {
				costDelta += correction.DeltaCostQuota
			}
			assert.Equal(t, tc.finalCost-fixture.OriginalCost, costDelta)
			execution, err := SumCostAccounting(CostAccountingFilter{UserID: 901, StartTimestamp: fixture.ExecutionMonthStart, EndTimestamp: common.GetTimestamp() + 3600})
			require.NoError(t, err)
			assert.Zero(t, execution.CostQuota, "cost corrections remain attributed to the original source month")
			for i := 0; i < 2; i++ {
				_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "", 1, true, "Reverse supplier and customer repricing")
				require.NoError(t, err)
			}
			restored, err := SumCostAccounting(filter)
			require.NoError(t, err)
			assert.Equal(t, before.CostQuota, restored.CostQuota)
			assert.Equal(t, before.RevenueQuota, restored.RevenueQuota)
			require.NoError(t, DB.First(&user, 901).Error)
			assert.Equal(t, 2000000, user.Quota)
			assert.Equal(t, 43000, user.UsedQuota)
		})
	}
}

func TestBillingModelCorrectionRejectsChangedSupplierEvidenceBeforeApply(t *testing.T) {
	for _, change := range []string{"cost_adjustment", "channel_discount", "model_pricing"} {
		t.Run(change, func(t *testing.T) {
			fixture := seedBillingCorrectionCosts(t, "doubao-seedance-2-5", "1080p", true, false)
			batch, err := PreviewBillingCorrection(fixture.Input, 1, "", "customer_group")
			require.NoError(t, err)
			require.True(t, batch.CanApply)
			switch change {
			case "cost_adjustment":
				_, err = AdjustCostAccountingSnapshots([]CostAccountingAdjustmentTarget{{SnapshotID: fixture.Snapshots[0].ID, NewCostQuota: fixture.Snapshots[0].CostQuota + 7}}, "Concurrent supplier correction", 2, "concurrent-cost")
				require.NoError(t, err)
			case "channel_discount":
				require.NoError(t, DB.Model(&Channel{}).Where("id = ?", 904).Update("cost_discount", "0.750000").Error)
			case "model_pricing":
				ratio := 6.0
				configureBillingCorrectionPricing(t, fixture.Task.PrivateData.BillingContext.OriginModelName, nil, &ratio)
			}
			_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "customer_group", 1, false, "")
			assert.ErrorIs(t, err, ErrBillingConflict)
			var user User
			require.NoError(t, DB.First(&user, 901).Error)
			assert.Equal(t, 2000000, user.Quota)
			assert.Equal(t, 43000, user.UsedQuota)
		})
	}
}

func TestBillingModelCorrectionBlocksConfiguredSupplierWithoutOriginalSnapshot(t *testing.T) {
	fixture := seedBillingCorrectionCosts(t, "doubao-seedance-2-5", "1080p", true, false)
	require.NoError(t, DB.Delete(&CostAccountingSnapshot{}, fixture.Snapshots[1].ID).Error)
	batch, err := PreviewBillingCorrection(fixture.Input, 1, "", "customer_group")
	require.NoError(t, err)
	assert.False(t, batch.CanApply, "a configured supplier cannot synthesize missing historical cost evidence")
	_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "customer_group", 1, false, "")
	assert.ErrorIs(t, err, ErrBillingCorrectionBlocked)
}

// Reconstruct a batch applied by the sales-only implementation. It contains
// neither supplier evidence nor supplier correction entries, just as existing
// deployed batches did before combined sales/cost repricing was introduced.
func seedAppliedLegacySalesCorrection(t *testing.T, fixture billingCorrectionCostFixture) *BillingCorrection {
	t.Helper()
	var user User
	require.NoError(t, DB.First(&user, fixture.Input.UserID).Error)
	account, err := GetBillingAccount(fixture.Input.UserID)
	require.NoError(t, err)
	batch, err := PreviewBillingCorrection(fixture.Input, 1, "", "customer_group")
	require.NoError(t, err)
	require.True(t, batch.CanApply)
	batch.CurrentCostQuota, batch.CorrectedCostQuota, batch.CostDelta = nil, nil, 0
	for i := range batch.Rows {
		row := &batch.Rows[i]
		row.TargetCost, row.CostDiscount, row.CostBlocked = "", "", ""
		row.CurrentCostQuota, row.CorrectedCostQuota = nil, nil
		row.CostDelta, row.CostChanged = 0, false
		require.NoError(t, DB.Save(row).Error)
		require.NoError(t, DB.Create(&BillingCorrectionClaim{SourceEntryID: row.SourceEntryID, BatchID: batch.ID}).Error)
		eventKey := fmt.Sprintf("correction:apply:%s:%d", batch.ID, row.SourceEntryID)
		require.NoError(t, DB.Create(&BillingEntry{EventKey: eventKey, UserID: user.Id, Sequence: account.Sequence + int64(i) + 1, PostedAt: common.GetTimestamp(),
			Kind: "rate_correction", Quota: row.Delta, WalletDelta: -row.Delta, ActorID: 1, CorrectionID: batch.ID, SourceEntryID: row.SourceEntryID,
			ModelName: row.ModelName, TokenID: row.TokenID, BillingGroup: row.TargetGroup, BillingRate: row.TargetRate,
			RequestID: billingCorrectionRequestID(eventKey)}).Error)
	}
	batch.SHA256, err = billingCorrectionDigest(batch, batch.Rows)
	require.NoError(t, err)
	batch.Status, batch.AppliedAt = "applied", common.GetTimestamp()
	require.NoError(t, DB.Save(batch).Error)
	require.NoError(t, DB.Model(&BillingAccount{}).Where("user_id = ?", user.Id).Update("sequence", account.Sequence+int64(len(batch.Rows))).Error)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{"quota": int64(user.Quota) - batch.NetDelta, "used_quota": int64(user.UsedQuota) + batch.NetDelta}).Error)
	require.NoError(t, DB.Model(&Token{}).Where("id = ?", 903).Updates(map[string]any{"remain_quota": int64(user.Quota) - batch.NetDelta, "used_quota": int64(user.UsedQuota) + batch.NetDelta}).Error)
	require.NoError(t, DB.Model(&Channel{}).Where("id = ?", 904).Update("used_quota", int64(user.UsedQuota)+batch.NetDelta).Error)
	return batch
}

func TestBillingModelCorrectionRepairsCostAfterLegacySalesOnlyAdjustment(t *testing.T) {
	fixture := seedBillingCorrectionCosts(t, "doubao-seedance-2-5", "1080p", true, false)
	legacy := seedAppliedLegacySalesCorrection(t, fixture)
	filter := CostAccountingFilter{UserID: 901, StartTimestamp: fixture.MonthStart, EndTimestamp: fixture.MonthEnd - 1}
	before, err := SumCostAccounting(filter)
	require.NoError(t, err)
	assert.Equal(t, int64(41000), before.CostQuota)
	assert.Equal(t, int64(28257), before.RevenueQuota)
	batch, err := PreviewBillingCorrection(fixture.Input, 1, "", "customer_group")
	require.NoError(t, err)
	require.True(t, batch.CanApply, "correct sales must not prevent repairing an obsolete supplier model basis")
	assert.Zero(t, batch.NetDelta)
	require.NotNil(t, batch.CurrentCostQuota)
	require.NotNil(t, batch.CorrectedCostQuota)
	assert.Equal(t, int64(41000), *batch.CurrentCostQuota)
	assert.Equal(t, int64(26943), *batch.CorrectedCostQuota)
	assert.Equal(t, int64(-14057), batch.CostDelta)
	for _, row := range batch.Rows {
		assert.Equal(t, legacy.ID, row.PreviousBatchID)
		assert.Zero(t, row.Delta)
		assert.True(t, row.CostChanged)
	}
	for i := 0; i < 2; i++ {
		_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "customer_group", 1, false, "")
		require.NoError(t, err)
	}
	var user User
	require.NoError(t, DB.First(&user, 901).Error)
	assert.Equal(t, int64(2000000)-legacy.NetDelta, int64(user.Quota), "a supplier-only repair must not charge or credit the already-correct customer again")
	assert.Equal(t, 28257, user.UsedQuota)
	corrected, err := SumCostAccounting(filter)
	require.NoError(t, err)
	assert.Equal(t, int64(26943), corrected.CostQuota)
	assert.Equal(t, int64(28257), corrected.RevenueQuota)
	_, err = ApplyBillingCorrection(legacy.ID, legacy.SHA256, "", "", 1, true, "Restore earlier sales pricing")
	assert.ErrorIs(t, err, ErrBillingCorrectionDependency, "new supplier corrections must be reversed before their parent sales batch")
	_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "", 1, true, "Restore previous supplier basis")
	require.NoError(t, err)
	restored, err := SumCostAccounting(filter)
	require.NoError(t, err)
	assert.Equal(t, before.CostQuota, restored.CostQuota)
	assert.Equal(t, before.RevenueQuota, restored.RevenueQuota)
	_, err = ApplyBillingCorrection(legacy.ID, legacy.SHA256, "", "", 1, true, "Restore earlier sales pricing")
	require.NoError(t, err, "legacy batches without supplier fields must remain reversible")
	restored, err = SumCostAccounting(filter)
	require.NoError(t, err)
	assert.Equal(t, int64(41000), restored.CostQuota)
	assert.Equal(t, int64(43000), restored.RevenueQuota)
	require.NoError(t, DB.First(&user, 901).Error)
	assert.Equal(t, 2000000, user.Quota)
	assert.Equal(t, 43000, user.UsedQuota)
}

func TestBillingModelCorrectionPersistsNewSupplierBasisEvenWithZeroDiscount(t *testing.T) {
	fixture := seedBillingCorrectionCosts(t, "doubao-seedance-2-5", "1080p", true, false, billingCorrectionCostFixtureOptions{Discount: "0.000000"})
	batch, err := PreviewBillingCorrection(fixture.Input, 1, "", "customer_group")
	require.NoError(t, err)
	require.True(t, batch.CanApply)
	assert.Zero(t, batch.CostDelta)
	for _, row := range batch.Rows {
		assert.True(t, row.CostChanged, "a model basis change remains auditable when the current supplier discount makes its cash delta zero")
	}
	_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "customer_group", 1, false, "")
	require.NoError(t, err)
	filter := CostAccountingFilter{UserID: 901, StartTimestamp: fixture.MonthStart, EndTimestamp: fixture.MonthEnd - 1}
	views, total, err := ListCostAccountingSnapshots(filter, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	basis := decimal.Zero
	for _, view := range views {
		basis = basis.Add(decimal.RequireFromString(view.EffectiveCostBasisQuota))
		assert.Zero(t, view.EffectiveCostQuota)
		assert.Equal(t, "0.000000", view.EffectiveCostDiscount)
		assert.Equal(t, CostAccountingCorrectedBasisVersion, view.EffectiveBasisVersion)
		assert.NotEmpty(t, view.EffectiveCostBasisBefore)
		assert.NotEmpty(t, view.EffectiveCostBasisAfter)
	}
	expectedBasis := decimal.NewFromInt(50000).Mul(decimal.NewFromInt(46)).Div(decimal.NewFromInt(70))
	assert.True(t, expectedBasis.Equal(basis), "later supplier discount repricing must use the exact corrected 46/70 model basis")
	_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "", 1, true, "Restore original zero-discount model basis")
	require.NoError(t, err)
	views, _, err = ListCostAccountingSnapshots(filter, 0, 10)
	require.NoError(t, err)
	basis = decimal.Zero
	for _, view := range views {
		basis = basis.Add(decimal.RequireFromString(view.EffectiveCostBasisQuota))
		assert.Zero(t, view.EffectiveCostQuota)
		assert.Equal(t, CostAccountingOriginalPriceBasisVersion, view.EffectiveBasisVersion)
	}
	assert.True(t, decimal.NewFromInt(50000).Equal(basis), "reversal restores original supplier basis as well as cash cost")
}

func TestBillingModelCorrectionReadsSupplierEvidenceAcrossSeparateLogDatabase(t *testing.T) {
	fixture := seedBillingCorrectionCosts(t, "doubao-seedance-2-5", "1080p", true, false, billingCorrectionCostFixtureOptions{NoSourceLogIDs: true})
	var logs []Log
	require.NoError(t, LOG_DB.Where("user_id = ?", 901).Find(&logs).Error)
	separateLogs, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := separateLogs.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, separateLogs.AutoMigrate(&Log{}))
	require.NoError(t, separateLogs.Create(&logs).Error)
	previousLogs := LOG_DB
	LOG_DB = separateLogs
	t.Cleanup(func() {
		LOG_DB = previousLogs
		assert.NoError(t, sqlDB.Close())
	})
	// Historical imports may have no direct source-log link. Matching the exact
	// task and signed source amount must work through the independent log DB.
	batch, err := PreviewBillingCorrection(fixture.Input, 1, "", "customer_group")
	require.NoError(t, err)
	for _, row := range batch.Rows {
		if row.Blocked != "" {
			t.Logf("blocked=%s cost_blocked=%s", row.Blocked, row.CostBlocked)
		}
	}
	require.True(t, batch.CanApply)
	require.NotNil(t, batch.CorrectedCostQuota)
	assert.Equal(t, int64(26943), *batch.CorrectedCostQuota)
	_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "customer_group", 1, false, "")
	require.NoError(t, err)
	costs, err := SumCostAccounting(CostAccountingFilter{UserID: 901, StartTimestamp: fixture.MonthStart, EndTimestamp: fixture.MonthEnd - 1})
	require.NoError(t, err)
	assert.Equal(t, int64(26943), costs.CostQuota)
	assert.Equal(t, int64(28257), costs.RevenueQuota)
}

func TestBillingModelCorrectionRollsBackSupplierCostWhenTokenChargeFails(t *testing.T) {
	fixture := seedBillingCorrectionCosts(t, "doubao-seedance-2-5", "1080p", false, false)
	batch, err := PreviewBillingCorrection(fixture.Input, 1, "", "customer_group")
	require.NoError(t, err)
	require.True(t, batch.CanApply)
	assert.Equal(t, int64(4300), batch.NetDelta)
	assert.Equal(t, int64(4100), batch.CostDelta)
	require.NoError(t, DB.Model(&Token{}).Where("id = ?", 903).Update("remain_quota", 0).Error)
	_, err = ApplyBillingCorrection(batch.ID, batch.SHA256, "", "customer_group", 1, false, "")
	assert.ErrorIs(t, err, ErrBillingInsufficientQuota)
	costs, err := SumCostAccounting(CostAccountingFilter{UserID: 901, StartTimestamp: fixture.MonthStart, EndTimestamp: fixture.MonthEnd - 1})
	require.NoError(t, err)
	assert.Equal(t, int64(41000), costs.CostQuota, "supplier adjustments and wallet entries must roll back together")
	assert.Equal(t, int64(43000), costs.RevenueQuota)
	var corrections, entries int64
	require.NoError(t, DB.Model(&CostAccountingAdjustment{}).Count(&corrections).Error)
	require.NoError(t, DB.Model(&BillingEntry{}).Where("kind = ?", "rate_correction").Count(&entries).Error)
	assert.Zero(t, corrections)
	assert.Zero(t, entries)
	var user User
	require.NoError(t, DB.First(&user, 901).Error)
	assert.Equal(t, 2000000, user.Quota)
	assert.Equal(t, 43000, user.UsedQuota)
	stored, err := GetBillingCorrection(batch.ID)
	require.NoError(t, err)
	assert.Equal(t, "preview", stored.Status)
}
