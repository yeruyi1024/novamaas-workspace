package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamBillingWarningsSurvivePreparationArchiveAndFinancialConfirmation(t *testing.T) {
	group := upstreamBillingServiceFixture(t)
	require.NoError(t, model.DB.AutoMigrate(&model.UpstreamBillingAttachment{}))
	start, _, err := model.BillingMonthBounds("2026-09")
	require.NoError(t, err)
	ctx := context.Background()
	missingCost := model.Log{UserId: 42, Type: model.LogTypeConsume, ChannelId: 19, CreatedAt: start + 1, ModelName: "legacy-cost", PromptTokens: 12, Quota: 100, RequestId: "warning-missing-cost", Other: `{"admin_info":{"usage_billing_path":"upstream"}}`}
	require.NoError(t, model.LOG_DB.Create(&missingCost).Error)
	upstreamBillingCostLog(t, model.Log{UserId: 42, Type: model.LogTypeConsume, ChannelId: 19, CreatedAt: start + 2, ModelName: "video-missing-task", Quota: 100, RequestId: "warning-missing-task", Other: `{"is_task":true}`}, "100", "0.8", 80)
	upstreamBillingCostLog(t, model.Log{UserId: 42, Type: model.LogTypeConsume, ChannelId: 19, CreatedAt: start + 3, ModelName: "estimated-text", PromptTokens: 100, CompletionTokens: 20, Quota: 300, RequestId: "warning-estimated", Other: `{"admin_info":{"local_count_tokens":true,"usage_billing_path":"local"}}`}, "300", "0.8", 240)
	task := model.Task{UserId: 42, ChannelId: 19, TaskID: "warning-pending-video", Status: model.TaskStatusInProgress, SubmitTime: start + 4}
	require.NoError(t, model.DB.Create(&task).Error)
	upstreamBillingCostLog(t, model.Log{UserId: 42, Type: model.LogTypeConsume, ChannelId: 19, CreatedAt: start + 4, ModelName: "pending-video", Quota: 200, RequestId: "warning-pending", Other: `{"is_task":true,"task_id":"warning-pending-video"}`}, "200", "0.8", 160)
	// A missing member is another frozen quality warning, not a permission to
	// silently drop it from the supplier's historical membership.
	require.NoError(t, model.DB.Delete(&model.Channel{}, 27).Error)

	statement, err := PrepareUpstreamBillingStatement(ctx, group.ID, 1, 1, "2026-09")
	require.NoError(t, err, "all source quality warnings must allow preparation")
	var snapshot UpstreamBillingSnapshot
	require.NoError(t, common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot))
	assert.False(t, snapshot.Coverage.Complete)
	assert.Equal(t, int64(1), snapshot.Coverage.MissingChannels)
	assert.Equal(t, int64(1), snapshot.Coverage.MissingCostRecords)
	assert.Equal(t, int64(2), snapshot.Coverage.MissingUsageRecords)
	assert.Equal(t, int64(1), snapshot.Coverage.EstimatedUsageRecords)
	assert.Equal(t, int64(1), snapshot.Coverage.PendingTasks)
	assert.Equal(t, int64(480), snapshot.Total.ChargeQuota, "unknown cost contributes no invented charge")
	frozenRows, err := model.GetUpstreamBillingStatementRecords(statement, 0, 1000)
	require.NoError(t, err)
	require.Len(t, frozenRows, 4)
	assert.True(t, frozenRows[0].CostMissing)
	assert.True(t, frozenRows[1].UsageMissing)
	assert.True(t, frozenRows[2].UsageEstimated)
	assert.True(t, frozenRows[3].TaskPending)
	assert.True(t, frozenRows[3].UsageMissing)

	// Later recovery changes live evidence; the archive must keep the exact
	// quality warnings and amounts that the reviewer was originally shown.
	recovered, err := model.BuildCostAccountingSnapshot(&missingCost, &model.CostAccountingInput{CostBasisQuota: "100", CostDiscount: "0.8", CostQuota: 80})
	require.NoError(t, err)
	require.NoError(t, model.DB.Create(recovered).Error)
	task.SetData(map[string]any{"usage": map[string]any{"total_tokens": 1000}, "duration": 5})
	require.NoError(t, model.DB.Model(&task).Updates(map[string]any{"status": model.TaskStatusSuccess, "data": task.Data, "finish_time": start + 10}).Error)
	live, err := GetUpstreamBillingReport(ctx, group.ID, "2026-09", "", "")
	require.NoError(t, err)
	assert.Zero(t, live.Coverage.MissingCostRecords)
	assert.Zero(t, live.Coverage.PendingTasks)
	assert.Equal(t, int64(560), live.Total.ChargeQuota)

	store := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildUpstreamBillingArchive(ctx, statement, store), "the archive worker must accept the frozen warnings")
	var manifest struct {
		Coverage *UpstreamBillingCoverage `json:"coverage"`
	}
	require.NoError(t, common.Unmarshal(store.files["manifest-0"], &manifest))
	require.NotNil(t, manifest.Coverage)
	assert.Equal(t, snapshot.Coverage, *manifest.Coverage)
	for _, artifact := range store.artifacts {
		require.NoError(t, model.DB.Create(&artifact).Error)
	}
	owner := "warning-flow-archive-worker"
	require.NoError(t, model.DB.Model(&model.UpstreamBillingStatement{}).Where("id = ?", statement.ID).Update("lease_owner", owner).Error)
	require.NoError(t, model.CompleteUpstreamBillingStatementArchive(statement, owner, true))
	review, err := model.GetUpstreamBillingReview(statement.ID)
	require.NoError(t, err)
	confirmed, err := model.ActUpstreamBillingStatementWithReview(statement.ID, "confirm", statement.ManifestSHA256, review.EvidenceSHA256, "已核对上游附件，保留差异待后续核实", "financial-warning-review", 9)
	require.NoError(t, err, "financial review must be possible with all frozen quality warnings")
	assert.Equal(t, model.StatementConfirmed, confirmed.Status)
	assert.Positive(t, confirmed.ArchivedAt)
	assert.Equal(t, statement.Snapshot, confirmed.Snapshot)
	assert.Equal(t, statement.SourceSHA256, confirmed.SourceSHA256)
	originalPDF, originalExcel := append([]byte(nil), store.files["pdf-0"]...), append([]byte(nil), store.files["xlsx-0"]...)
	_, err = BuildUpstreamBillingReceipt(ctx, confirmed, store)
	require.NoError(t, err)
	assert.Equal(t, originalPDF, store.files["pdf-0"])
	assert.Equal(t, originalExcel, store.files["xlsx-0"])
	after, err := model.GetUpstreamBillingStatementRecords(confirmed, 0, 1000)
	require.NoError(t, err)
	assert.Equal(t, frozenRows, after, "later source recovery and review never rewrite frozen evidence")

	if output := os.Getenv("UPSTREAM_BILLING_WARNING_OUTPUT"); output != "" {
		require.NoError(t, os.MkdirAll(output, 0700))
		for name, body := range map[string][]byte{"warning-statement.pdf": originalPDF, "warning-statement.xlsx": originalExcel, "warning-receipt.pdf": store.files["receipt-0"], "warning-snapshot.json": []byte(confirmed.Snapshot), "warning-manifest.json": store.files["manifest-0"]} {
			require.NoError(t, os.WriteFile(filepath.Join(output, name), body, 0600))
		}
	}
}
