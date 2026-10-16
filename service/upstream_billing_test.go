package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func upstreamBillingServiceFixture(t *testing.T) *model.UpstreamBillingGroupView {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	oldDB, oldLog := model.DB, model.LOG_DB
	model.DB, model.LOG_DB = db, db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Channel{}, &model.Log{}, &model.Task{}, &model.BillingEntry{}, &model.BillingCorrectionRow{}, &model.CostAccountingSnapshot{}, &model.CostAccountingAdjustment{}, &model.BillingArtifact{}, &model.UpstreamBillingGroup{}, &model.UpstreamBillingGroupChannel{}, &model.UpstreamBillingGroupEvent{}, &model.UpstreamBillingStatement{}, &model.UpstreamBillingStatementEvent{}, &model.UpstreamBillingStatementRecord{}))
	t.Cleanup(func() { model.DB, model.LOG_DB = oldDB, oldLog; require.NoError(t, sqlDB.Close()) })
	require.NoError(t, db.Create(&[]model.Channel{{Id: 19, Name: "primary", Key: "private", CostDiscount: "0.8"}, {Id: 26, Name: "secondary", Key: "private", CostDiscount: "0"}, {Id: 27, Name: "tests only", Key: "private"}}).Error)
	group, err := model.SaveUpstreamBillingGroup(context.Background(), 0, 0, "Merged supplier", []int{19, 26, 27}, 1)
	require.NoError(t, err)
	return group
}

func upstreamBillingCostLog(t *testing.T, log model.Log, basis, discount string, cost int64) *model.CostAccountingSnapshot {
	t.Helper()
	require.NoError(t, model.LOG_DB.Create(&log).Error)
	snapshot, err := model.BuildCostAccountingSnapshot(&log, &model.CostAccountingInput{CostBasisQuota: basis, CostDiscount: discount, CostQuota: cost})
	require.NoError(t, err)
	require.NoError(t, model.DB.Create(snapshot).Error)
	return snapshot
}

func TestUpstreamBillingReportMergesConfiguredChannelsWithoutRepricingHistoricalCosts(t *testing.T) {
	group := upstreamBillingServiceFixture(t)
	start, end, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	text := `{"cache_tokens":40,"admin_info":{"usage_billing_path":"upstream"}}`
	first := upstreamBillingCostLog(t, model.Log{UserId: 7, Type: model.LogTypeConsume, CreatedAt: start + 3600, ModelName: "Model", Group: "default", ChannelId: 19, Quota: 1000, PromptTokens: 100, CompletionTokens: 20, RequestId: "first", Other: text}, "1000", "0.8", 800)
	upstreamBillingCostLog(t, model.Log{UserId: 7, Type: model.LogTypeConsume, CreatedAt: start + 2*86400, ModelName: "model", Group: "default", ChannelId: 26, Quota: 500, PromptTokens: 200, CompletionTokens: 30, RequestId: "zero-cost", Other: `{"admin_info":{"usage_billing_path":"upstream"}}`}, "500", "0", 0)
	upstreamBillingCostLog(t, model.Log{UserId: 7, Type: model.LogTypeRefund, CreatedAt: start + 86400, ModelName: "Model", Group: "default", ChannelId: 19, Quota: 100, RequestId: "refund"}, "-100", "0.8", -80)
	upstreamBillingCostLog(t, model.Log{UserId: 7, Type: model.LogTypeConsume, CreatedAt: end, ModelName: "outside", ChannelId: 19, Quota: 100, RequestId: "outside"}, "100", "0.8", 80)
	require.NoError(t, model.LOG_DB.Create(&model.Log{UserId: 7, Type: model.LogTypeConsume, CreatedAt: start, ModelName: "tests", ChannelId: 27, Quota: 9999, PromptTokens: 9999}).Error)
	require.NoError(t, model.DB.Create(&model.CostAccountingAdjustment{SnapshotID: first.ID, DeltaCostQuota: -50, NewCostQuota: 750, CostDiscount: "0.75", CreatedAt: end + 86400}).Error)
	// A later customer discount changes the visible sales group and amount. It
	// does not create another supplier request or replace the upstream cost.
	source := model.BillingEntry{UserID: 7, Sequence: 1, EventKey: "original", Kind: "consume", Quota: 1000, PostedAt: start + 3600, SourceLogID: int(first.SourceLogID), RequestID: "first", ModelName: "Model"}
	require.NoError(t, model.DB.Create(&source).Error)
	require.NoError(t, model.DB.Create(&model.BillingEntry{UserID: 7, Sequence: 2, EventKey: "correction:apply:customer", Kind: "rate_correction", Quota: -400, PostedAt: end + 86400, SourceEntryID: source.ID, CorrectionID: "customer", BillingGroup: "vip", BillingRate: "0.6", ModelName: "Model"}).Error)
	require.NoError(t, model.DB.Create(&model.BillingCorrectionRow{BatchID: "customer", SourceEntryID: source.ID}).Error)
	report, err := GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "", "")
	require.NoError(t, err)
	assert.True(t, report.Coverage.Complete)
	assert.Equal(t, int64(750), report.Total.ChargeQuota)
	assert.Equal(t, int64(80), report.Total.RefundQuota)
	assert.Equal(t, int64(3), report.Total.Count)
	assert.Equal(t, int64(2), report.Total.Requests)
	assert.Equal(t, int64(300), report.Total.InputTokens)
	assert.Equal(t, int64(50), report.Total.OutputTokens)
	assert.Equal(t, int64(40), report.Total.CacheReadTokens)
	require.Len(t, report.Coverage.ExcludedChannels, 1)
	assert.Equal(t, 27, report.Coverage.ExcludedChannels[0].ID)
	require.Len(t, report.Days, 31)
	assert.Equal(t, int64(750), report.Days[0].ChargeQuota)
	require.Len(t, report.Models, 3, "exact model case and effective customer group remain distinct")
	assert.Equal(t, "default", report.Models[0].BillingGroup)
	assert.Equal(t, int64(80), report.Models[0].RefundQuota)
	assert.Equal(t, "vip", report.Models[1].BillingGroup)
	assert.Equal(t, int64(750), report.Models[1].ChargeQuota)
	assert.Equal(t, "model", report.Models[2].ModelName)
	assert.NotEmpty(t, report.SourceSHA256)
	details, err := GetUpstreamBillingUsageDetails(context.Background(), group.ID, "2026-08", "", "", "2026-08-01", "Model", 0)
	require.NoError(t, err)
	require.Len(t, details.Records, 1)
	assert.Equal(t, int64(750), details.Records[0].CostQuota)
	assert.NotEmpty(t, details.Records[0].Amount)
}

func TestUpstreamBillingHistoricalMissingCostsAndEstimatedUsageRemainWarningsInFrozenStatements(t *testing.T) {
	group := upstreamBillingServiceFixture(t)
	start, _, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	require.NoError(t, model.LOG_DB.Create(&model.Log{UserId: 1, Type: model.LogTypeConsume, CreatedAt: start, ModelName: "legacy", ChannelId: 19, Quota: 100, PromptTokens: 25}).Error)
	upstreamBillingCostLog(t, model.Log{UserId: 1, Type: model.LogTypeConsume, CreatedAt: start + 1, ModelName: "old-default", ChannelId: 19, Quota: 100, PromptTokens: 25, RequestId: "unconfigured-then"}, "100", "", 100)
	upstreamBillingCostLog(t, model.Log{UserId: 1, Type: model.LogTypeConsume, CreatedAt: start + 2, ModelName: "estimated", ChannelId: 19, Quota: 100, PromptTokens: 25, RequestId: "estimate", Other: `{"admin_info":{"usage_billing_path":"billing-usage-openai-estimated"}}`}, "100", "0.8", 80)
	report, err := GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "", "")
	require.NoError(t, err)
	assert.False(t, report.Coverage.Complete)
	assert.Equal(t, int64(2), report.Coverage.MissingCostRecords)
	assert.Equal(t, int64(1), report.Coverage.EstimatedUsageRecords)
	assert.Equal(t, int64(80), report.Total.ChargeQuota, "legacy default revenue must not become supplier cost")
	statement, err := PrepareUpstreamBillingStatement(context.Background(), group.ID, 1, 1, "2026-08")
	require.NoError(t, err)
	var snapshot UpstreamBillingSnapshot
	require.NoError(t, common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot))
	assert.False(t, snapshot.Coverage.Complete)
	assert.Equal(t, report.Coverage, snapshot.Coverage)
	assert.Equal(t, report.Total, snapshot.Total)
	frozen, err := model.GetUpstreamBillingStatementRecords(statement, 0, 1000)
	require.NoError(t, err)
	require.Len(t, frozen, 3)
	assert.True(t, frozen[0].CostMissing)
	assert.True(t, frozen[1].CostMissing)
	assert.True(t, frozen[2].UsageEstimated)
	store := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildUpstreamBillingArchive(context.Background(), statement, store), "warning-bearing statements must reach the archive worker")
	assert.NotEmpty(t, store.files["snapshot-0"])
	assert.NotEmpty(t, store.files["pdf-0"])
	assert.NotEmpty(t, store.files["xlsx-0"])
	assert.Equal(t, []byte(statement.Snapshot), store.files["snapshot-0"])
	var manifest UpstreamBillingManifest
	require.NoError(t, common.Unmarshal(store.files["manifest-0"], &manifest))
	require.NotNil(t, manifest.Coverage)
	assert.Equal(t, snapshot.Coverage, *manifest.Coverage)
}

func TestUpstreamBillingUnknownCostContributesOnlyToWarningAndRecordCounts(t *testing.T) {
	records := []model.UpstreamBillingStatementRecord{
		{LogType: model.LogTypeConsume, CostQuota: 80},
		{LogType: model.LogTypeRefund, CostQuota: -20},
		{LogType: model.LogTypeConsume, CostQuota: 100, CostMissing: true},
		{LogType: model.LogTypeRefund, CostQuota: -50, CostMissing: true},
	}
	aggregate := upstreamBillingAggregate{}
	archived := UpstreamBillingRow{}
	for _, record := range records {
		require.NoError(t, aggregate.Add(record))
		require.NoError(t, archived.addArchiveRecord(record))
	}
	assert.Equal(t, int64(80), aggregate.charge)
	assert.Equal(t, int64(20), aggregate.refund)
	assert.Equal(t, int64(4), aggregate.count)
	assert.Equal(t, int64(2), aggregate.missingCost)
	assert.Equal(t, aggregate.charge, archived.ChargeQuota)
	assert.Equal(t, aggregate.refund, archived.RefundQuota)
	assert.Equal(t, aggregate.count, archived.Count)
	assert.Equal(t, aggregate.missingCost, archived.MissingCostRecords)
}

func TestUpstreamBillingArchiveRetainsQualityIntegrityChecksForWarningSnapshots(t *testing.T) {
	group := upstreamBillingServiceFixture(t)
	start, _, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	require.NoError(t, model.LOG_DB.Create(&model.Log{UserId: 1, Type: model.LogTypeConsume, CreatedAt: start, ModelName: "legacy", ChannelId: 19, Quota: 100, PromptTokens: 25}).Error)
	statement, err := PrepareUpstreamBillingStatement(context.Background(), group.ID, 1, 1, "2026-08")
	require.NoError(t, err)
	var original UpstreamBillingSnapshot
	require.NoError(t, common.UnmarshalJsonStr(string(statement.Snapshot), &original))
	for _, altered := range []UpstreamBillingCoverage{
		{Complete: true, MissingCostRecords: 1, ExcludedChannels: original.Coverage.ExcludedChannels},
		{MissingCostRecords: 2, ExcludedChannels: original.Coverage.ExcludedChannels},
		{MissingCostRecords: 1, MissingChannels: 1, ExcludedChannels: original.Coverage.ExcludedChannels},
	} {
		copy := original
		copy.Coverage = altered
		body, err := common.Marshal(copy)
		require.NoError(t, err)
		digest := sha256.Sum256(body)
		statement.Snapshot, statement.SnapshotSHA256 = model.UpstreamBillingSnapshotData(body), hex.EncodeToString(digest[:])
		store := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
		assert.ErrorIs(t, BuildUpstreamBillingArchive(context.Background(), statement, store), model.ErrBillingEvidenceIntegrity)
		assert.NotContains(t, store.files, "manifest-0")
	}
}

func TestUpstreamBillingVideoFinalUsageIsCountedOnceAndFrozenForArchiveRetries(t *testing.T) {
	group := upstreamBillingServiceFixture(t)
	start, _, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	task := model.Task{UserId: 1, TaskID: "video-public", ChannelId: 19, Status: model.TaskStatusSuccess, SubmitTime: start + 10, FinishTime: start + 20}
	task.SetData(map[string]any{"usage": map[string]any{"total_tokens": 10000}, "duration": 8})
	require.NoError(t, model.DB.Create(&task).Error)
	initial := upstreamBillingCostLog(t, model.Log{UserId: 1, Type: model.LogTypeConsume, CreatedAt: start + 10, ModelName: "video", ChannelId: 19, Quota: 1000, RequestId: "submit", Other: `{"is_task":true,"task_id":"video-public"}`}, "1000", "0.8", 800)
	upstreamBillingCostLog(t, model.Log{UserId: 1, Type: model.LogTypeRefund, CreatedAt: start + 20, ModelName: "video", ChannelId: 19, Quota: 100, RequestId: "settle", Other: `{"is_task":true,"task_id":"video-public","pre_consumed_quota":1000,"actual_quota":900}`}, "-100", "0.8", -80)
	statement, err := PrepareUpstreamBillingStatement(context.Background(), group.ID, 1, 1, "2026-08")
	require.NoError(t, err)
	var snapshot UpstreamBillingSnapshot
	require.NoError(t, common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot))
	assert.Equal(t, int64(1), snapshot.Total.Requests)
	assert.Equal(t, int64(10000), snapshot.Total.VideoTokens)
	assert.Equal(t, int64(8), snapshot.Total.VideoSeconds)
	assert.Equal(t, int64(2), snapshot.Total.Count)
	assert.True(t, snapshot.Coverage.Complete)
	task.SetData(map[string]any{"usage": map[string]any{"total_tokens": 20000}})
	require.NoError(t, model.DB.Model(&task).Update("data", task.Data).Error)
	require.NoError(t, model.DB.Create(&model.CostAccountingAdjustment{SnapshotID: initial.ID, DeltaCostQuota: -100, NewCostQuota: 700}).Error)
	frozen, err := model.GetUpstreamBillingStatementRecords(statement, 0, 1000)
	require.NoError(t, err)
	require.Len(t, frozen, 2)
	assert.Equal(t, int64(10000), frozen[0].VideoTokens)
	assert.Equal(t, int64(800), frozen[0].CostQuota)
	assert.Zero(t, frozen[1].VideoTokens)
	assert.Equal(t, "video-public", frozen[0].TaskID)
	_, err = PrepareUpstreamBillingStatement(context.Background(), group.ID, 1, 1, "2026-08")
	assert.ErrorIs(t, err, model.ErrBillingConflict)
	var count int64
	require.NoError(t, model.DB.Model(&model.UpstreamBillingStatementRecord{}).Where("statement_id = ?", statement.ID).Count(&count).Error)
	assert.Equal(t, int64(2), count)
}

func TestUpstreamBillingPeriodValidationAndEmptyEligibilityWarning(t *testing.T) {
	group := upstreamBillingServiceFixture(t)
	_, err := GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "2026-08-01", "2026-08-31")
	assert.Error(t, err)
	_, err = GetUpstreamBillingReport(context.Background(), group.ID, "", "2025-01-01", "2026-01-02")
	assert.Error(t, err)
	require.NoError(t, model.DB.Model(&model.Channel{}).Where("id IN ?", []int{19, 26}).Update("cost_discount", "").Error)
	report, err := GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "", "")
	require.NoError(t, err)
	assert.False(t, report.Coverage.Complete)
	statement, err := PrepareUpstreamBillingStatement(context.Background(), group.ID, 1, 1, "2026-08")
	require.NoError(t, err)
	var snapshot UpstreamBillingSnapshot
	require.NoError(t, common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot))
	assert.False(t, snapshot.Coverage.Complete, "an empty eligible range is a frozen warning, not verified zero cost")
	assert.Empty(t, snapshot.Models)
	store := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildUpstreamBillingArchive(context.Background(), statement, store))
	assert.NotEmpty(t, store.files["manifest-0"])
	require.NoError(t, model.DB.Model(&model.Channel{}).Where("id = ?", 19).Update("cost_discount", "0").Error)
	report, err = GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "", "")
	require.NoError(t, err)
	assert.True(t, report.Coverage.Complete, "a real configured channel with no usage may have a zero month")
	currentMonth := time.Now().In(billingLocation).Format("2006-01")
	_, err = PrepareUpstreamBillingStatement(context.Background(), group.ID, 1, 1, currentMonth)
	assert.ErrorIs(t, err, ErrUpstreamBillingMonthOpen)
}

func TestUpstreamBillingMissingFinalVideoUsageAndUnfinishedTasksRemainVisibleInDrafts(t *testing.T) {
	for _, test := range []struct {
		name    string
		status  model.TaskStatus
		data    map[string]any
		pending int64
	}{
		{name: "finished without reported usage", status: model.TaskStatusSuccess, data: map[string]any{"status": "succeeded"}},
		{name: "unfinished with provisional usage", status: model.TaskStatusInProgress, data: map[string]any{"usage": map[string]any{"total_tokens": 50}}, pending: 1},
		{name: "unbounded provider tokens", status: model.TaskStatusSuccess, data: map[string]any{"usage": map[string]any{"total_tokens": int64(common.MaxQuota) + 1}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			group := upstreamBillingServiceFixture(t)
			start, _, err := model.BillingMonthBounds("2026-08")
			require.NoError(t, err)
			task := model.Task{UserId: 1, ChannelId: 19, TaskID: "task", Status: test.status}
			task.SetData(test.data)
			require.NoError(t, model.DB.Create(&task).Error)
			upstreamBillingCostLog(t, model.Log{UserId: 1, Type: model.LogTypeConsume, CreatedAt: start, ChannelId: 19, ModelName: "video", Quota: 100, RequestId: "submit", Other: `{"is_task":true,"task_id":"task"}`}, "100", "0.8", 80)
			report, err := GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "", "")
			require.NoError(t, err)
			assert.False(t, report.Coverage.Complete)
			assert.Equal(t, int64(1), report.Coverage.MissingUsageRecords)
			assert.Equal(t, test.pending, report.Coverage.PendingTasks)
			assert.Equal(t, int64(80), report.Total.ChargeQuota)
			statement, err := PrepareUpstreamBillingStatement(context.Background(), group.ID, 1, 1, "2026-08")
			require.NoError(t, err)
			var snapshot UpstreamBillingSnapshot
			require.NoError(t, common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot))
			assert.Equal(t, report.Coverage, snapshot.Coverage)
			store := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
			require.NoError(t, BuildUpstreamBillingArchive(context.Background(), statement, store))
		})
	}
}

func TestUpstreamBillingPreparationFreezesEveryRecordAcrossSourceAndDatabasePageBoundaries(t *testing.T) {
	group := upstreamBillingServiceFixture(t)
	start, _, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	// One more than the public archive/source page protects a real pagination
	// boundary, including SQLite's parameter limit on the wide staging rows.
	logs := make([]model.Log, 1001)
	for i := range logs {
		logs[i] = model.Log{Id: i + 1, UserId: 1, Type: model.LogTypeConsume, CreatedAt: start + int64(i), ChannelId: 19, ModelName: "text", RequestId: fmt.Sprintf("page-source-%d", i), Quota: 100, PromptTokens: 10, Other: `{"admin_info":{"usage_billing_path":"upstream"}}`}
	}
	require.NoError(t, model.LOG_DB.CreateInBatches(&logs, 500).Error)
	snapshots := make([]model.CostAccountingSnapshot, 0, len(logs))
	for i := range logs {
		snapshot, err := model.BuildCostAccountingSnapshot(&logs[i], &model.CostAccountingInput{CostBasisQuota: "100", CostDiscount: "0.8", CostQuota: 80})
		require.NoError(t, err)
		snapshots = append(snapshots, *snapshot)
	}
	require.NoError(t, model.DB.CreateInBatches(&snapshots, 500).Error)
	statement, err := PrepareUpstreamBillingStatement(context.Background(), group.ID, 1, 1, "2026-08")
	require.NoError(t, err)
	assert.Equal(t, int64(1001), statement.RecordCount)
	var snapshot UpstreamBillingSnapshot
	require.NoError(t, common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot))
	assert.Equal(t, int64(80080), snapshot.Total.ChargeQuota)
	assert.Equal(t, int64(1001), snapshot.Total.Requests)
	assert.Equal(t, int64(10010), snapshot.Total.InputTokens)
	first, err := model.GetUpstreamBillingStatementRecords(statement, 0, 1000)
	require.NoError(t, err)
	require.Len(t, first, 1000)
	last, err := model.GetUpstreamBillingStatementRecords(statement, first[len(first)-1].Sequence, 1000)
	require.NoError(t, err)
	require.Len(t, last, 1)
	assert.Equal(t, int64(1001), last[0].Sequence)
	assert.Equal(t, "page-source-1000", last[0].RequestID)
}

func TestUpstreamBillingDurationOnlyVideoPreservesKnownSecondsAndMissingTokenWarning(t *testing.T) {
	for _, response := range []struct {
		name string
		data map[string]any
	}{
		{name: "direct provider response", data: map[string]any{"duration": 8}},
		{name: "nested provider response", data: map[string]any{"data": map[string]any{"duration": 8}}},
	} {
		t.Run(response.name, func(t *testing.T) {
			group := upstreamBillingServiceFixture(t)
			start, _, err := model.BillingMonthBounds("2026-08")
			require.NoError(t, err)
			task := model.Task{UserId: 1, ChannelId: 19, TaskID: "duration-only-video", Status: model.TaskStatusSuccess, PrivateData: model.TaskPrivateData{BillingContext: &model.TaskBillingContext{ModelRatio: 5, CostDiscount: "0.8"}}}
			task.SetData(response.data)
			require.NoError(t, model.DB.Create(&task).Error)
			upstreamBillingCostLog(t, model.Log{UserId: 1, Type: model.LogTypeConsume, CreatedAt: start, ChannelId: 19, ModelName: "video", Quota: 100, RequestId: "duration-submit", Other: `{"is_task":true,"task_id":"duration-only-video"}`}, "100", "0.8", 80)
			report, err := GetUpstreamBillingReport(context.Background(), group.ID, "2026-08", "", "")
			require.NoError(t, err)
			assert.False(t, report.Coverage.Complete, "reported duration cannot prove missing final token usage")
			assert.Equal(t, int64(1), report.Coverage.MissingUsageRecords)
			assert.Equal(t, int64(8), report.Total.VideoSeconds)
			details, err := GetUpstreamBillingUsageDetails(context.Background(), group.ID, "2026-08", "", "", "", "", 0)
			require.NoError(t, err)
			require.Len(t, details.Records, 1)
			assert.True(t, details.Records[0].UsageMissing)
			assert.Equal(t, int64(8), details.Records[0].VideoSeconds)
			statement, err := PrepareUpstreamBillingStatement(context.Background(), group.ID, 1, 1, "2026-08")
			require.NoError(t, err)
			var snapshot UpstreamBillingSnapshot
			require.NoError(t, common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot))
			assert.Equal(t, int64(8), snapshot.Total.VideoSeconds)
			assert.Equal(t, int64(1), snapshot.Coverage.MissingUsageRecords)
			store := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
			require.NoError(t, BuildUpstreamBillingArchive(context.Background(), statement, store))
		})
	}
}
