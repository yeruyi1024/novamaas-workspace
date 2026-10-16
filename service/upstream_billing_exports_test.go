package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

type memoryUpstreamBillingArchive struct {
	files     map[string][]byte
	artifacts []model.BillingArtifact
}

func (store *memoryUpstreamBillingArchive) PutUpstream(_ context.Context, id, kind string, ordinal, owner int, _ string, body []byte, rows int64) (*model.BillingArtifact, error) {
	key := fmt.Sprintf("%s-%d", kind, ordinal)
	if saved, ok := store.files[key]; ok {
		if !bytes.Equal(saved, body) {
			return nil, model.ErrBillingConflict
		}
		for _, artifact := range store.artifacts {
			if artifact.Kind == kind && artifact.Ordinal == ordinal {
				return &artifact, nil
			}
		}
	}
	store.files[key] = append([]byte(nil), body...)
	hash := sha256.Sum256(body)
	artifact := model.BillingArtifact{StatementID: id, Kind: kind, Ordinal: ordinal, SHA256: hex.EncodeToString(hash[:]), Size: int64(len(body)), Rows: rows}
	store.artifacts = append(store.artifacts, artifact)
	return &artifact, nil
}

func frozenUpstreamBillingFixture(t *testing.T) (*model.UpstreamBillingStatement, *UpstreamBillingSnapshot, []model.UpstreamBillingStatementRecord) {
	t.Helper()
	start, end, err := model.BillingMonthBounds("2026-09")
	require.NoError(t, err)
	statement := &model.UpstreamBillingStatement{ID: "DEMO-UPSTREAM-202609", GroupID: 7, GroupVersion: 1, Month: "2026-09", Revision: 1, CreatedBy: 1, CreatedAt: end + 3600, StorageProfileID: 1, StartAt: start, EndAt: end, RecordCount: 4, Status: model.StatementPreparing}
	records := []model.UpstreamBillingStatementRecord{
		{StatementID: statement.ID, UserID: 42, Sequence: 1, SourceLogID: 1, CostSnapshotID: 1, OccurredAt: start + 3600, LogType: model.LogTypeConsume, ChannelID: 19, ModelName: "gpt-5", UpstreamModelName: "gpt-5-2026-09", BillingGroup: "default", CostQuota: 1000000, CostDiscount: "0.800000", CostBasisQuota: "1250000.000000", RequestID: "demo-request-1", UpstreamRequestID: "provider-request-1", UsageSource: "log", UpstreamBillingUsage: model.UpstreamBillingUsage{Requests: 1, InputTokens: 1600, OutputTokens: 200, CacheReadTokens: 100}},
		{StatementID: statement.ID, UserID: 42, Sequence: 2, SourceLogID: 2, CostSnapshotID: 2, OccurredAt: start + 7200, LogType: model.LogTypeRefund, ChannelID: 19, ModelName: "gpt-5", UpstreamModelName: "gpt-5-2026-09", BillingGroup: "default", CostQuota: -250000, CostDiscount: "0.800000", CostBasisQuota: "-312500.000000", RequestID: "demo-refund-1", UsageSource: "refund"},
		{StatementID: statement.ID, UserID: 42, Sequence: 3, SourceLogID: 3, CostSnapshotID: 3, OccurredAt: start + 86400 + 3600, LogType: model.LogTypeConsume, ChannelID: 26, ModelName: "claude-sonnet-enterprise-context-analysis-2026", UpstreamModelName: "claude-sonnet-4-6", BillingGroup: "preferred", CostQuota: 900000, CostDiscount: "0.900000", CostBasisQuota: "1000000.000000", RequestID: "demo-request-2", UsageSource: "log", UpstreamBillingUsage: model.UpstreamBillingUsage{Requests: 1, InputTokens: 4800, OutputTokens: 600, CacheReadTokens: 1000, CacheWriteTokens: 500}},
		{StatementID: statement.ID, UserID: 42, Sequence: 4, SourceLogID: 4, CostSnapshotID: 4, OccurredAt: start + 2*86400 + 3600, LogType: model.LogTypeConsume, ChannelID: 27, ModelName: "video-pro", UpstreamModelName: "video-pro", BillingGroup: "default", CostQuota: 0, CostDiscount: "0.000000", CostBasisQuota: "500000.000000", RequestID: "demo-request-3", TaskID: "demo-video-task", UsageSource: "task", UpstreamBillingUsage: model.UpstreamBillingUsage{Requests: 1, VideoTokens: 10800, VideoSeconds: 6}},
	}
	var source bytes.Buffer
	for _, record := range records {
		body, err := common.Marshal(record)
		require.NoError(t, err)
		source.Write(body)
		source.WriteByte('\n')
	}
	sourceHash := sha256.Sum256(source.Bytes())
	statement.SourceSHA256 = hex.EncodeToString(sourceHash[:])
	currency := BillingCurrency{Code: "CNY", Symbol: "¥", Rate: "7", QuotaPerUnit: "500000"}
	snapshot := &UpstreamBillingSnapshot{SchemaVersion: 1, AccountingBasis: "historical_cost_snapshots", Timezone: "Asia/Shanghai", Month: "2026-09", StartDate: "2026-09-01", EndDate: "2026-09-30", StartAt: start, EndAt: end, Currency: currency, SourceSHA256: statement.SourceSHA256, UpdatedAt: statement.CreatedAt, Issuer: "new-api 演示平台", OperatingName: "演示运营主体", PDFLogoPNG: billingPlatformMarkV2, RoundingDifference: "0.000000", Coverage: UpstreamBillingCoverage{Complete: true}, Group: model.UpstreamBillingGroupView{UpstreamBillingGroup: model.UpstreamBillingGroup{ID: 7, Name: "示例上游供应商（演示）", Version: 1}, Channels: []model.UpstreamBillingChannelView{{ID: 19, Name: "正式渠道 A", Eligible: true, CostDiscount: "0.800000", GroupID: 7}, {ID: 26, Name: "正式渠道 B", Eligible: true, CostDiscount: "0.900000", GroupID: 7}, {ID: 27, Name: "零成本渠道", Eligible: true, CostDiscount: "0.000000", GroupID: 7}, {ID: 28, Name: "未配置成本的测试渠道", GroupID: 7}}}}
	snapshot.Coverage.ExcludedChannels = []model.UpstreamBillingChannelView{snapshot.Group.Channels[3]}
	amounts, err := billingRow("total", 1900000, 250000, 4, currency)
	require.NoError(t, err)
	snapshot.Total = UpstreamBillingRow{BillingRow: amounts, ChargeCount: 3, RefundCount: 1, UpstreamBillingUsage: model.UpstreamBillingUsage{Requests: 3, InputTokens: 6400, OutputTokens: 800, CacheReadTokens: 1100, CacheWriteTokens: 500, VideoTokens: 10800, VideoSeconds: 6}}
	for day := start; day < end; day += 86400 {
		label := time.Unix(day, 0).In(billingLocation).Format("2006-01-02")
		row := UpstreamBillingRow{}
		for _, record := range records {
			if record.OccurredAt >= day && record.OccurredAt < day+86400 {
				require.NoError(t, row.addArchiveRecord(record))
			}
		}
		row.BillingRow, err = billingRow(label, row.ChargeQuota, row.RefundQuota, row.Count, currency)
		require.NoError(t, err)
		snapshot.Days = append(snapshot.Days, row)
	}
	for _, reference := range []struct{ model, upstream, group string }{{"gpt-5", "gpt-5-2026-09", "default"}, {"claude-sonnet-enterprise-context-analysis-2026", "claude-sonnet-4-6", "preferred"}, {"video-pro", "video-pro", "default"}} {
		row := UpstreamBillingModelRow{ModelName: reference.model, UpstreamModelName: reference.upstream, BillingGroup: reference.group}
		for _, record := range records {
			if record.ModelName == reference.model {
				require.NoError(t, row.UpstreamBillingRow.addArchiveRecord(record))
			}
		}
		row.BillingRow, err = billingRow(reference.model, row.ChargeQuota, row.RefundQuota, row.Count, currency)
		require.NoError(t, err)
		row.ActiveDays = 1
		for _, record := range records {
			if record.ModelName != reference.model {
				continue
			}
			if row.FirstPosted == 0 || record.OccurredAt < row.FirstPosted {
				row.FirstPosted = record.OccurredAt
			}
			if record.OccurredAt > row.LastPosted {
				row.LastPosted = record.OccurredAt
			}
		}
		snapshot.Models = append(snapshot.Models, row)
		for _, day := range snapshot.Days {
			if day.Count == 0 {
				continue
			}
			dayRow := UpstreamBillingDayModelRow{Date: day.Label, UpstreamBillingModelRow: UpstreamBillingModelRow{ModelName: reference.model, UpstreamModelName: reference.upstream, BillingGroup: reference.group}}
			for _, record := range records {
				if record.ModelName == reference.model && time.Unix(record.OccurredAt, 0).In(billingLocation).Format("2006-01-02") == day.Label {
					require.NoError(t, dayRow.UpstreamBillingRow.addArchiveRecord(record))
					dayRow.ActiveDays = 1
					if dayRow.FirstPosted == 0 || record.OccurredAt < dayRow.FirstPosted {
						dayRow.FirstPosted = record.OccurredAt
					}
					if record.OccurredAt > dayRow.LastPosted {
						dayRow.LastPosted = record.OccurredAt
					}
				}
			}
			if dayRow.Count == 0 {
				continue
			}
			dayRow.BillingRow, err = billingRow(reference.model, dayRow.ChargeQuota, dayRow.RefundQuota, dayRow.Count, currency)
			require.NoError(t, err)
			snapshot.DayModels = append(snapshot.DayModels, dayRow)
		}
	}
	frozen, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(frozen)
	statement.Snapshot, statement.SnapshotSHA256 = model.UpstreamBillingSnapshotData(frozen), hex.EncodeToString(hash[:])
	return statement, snapshot, records
}

func TestUpstreamArchiveBindsFrozenEvidenceExportsAndIndependentReceipt(t *testing.T) {
	require.NoError(t, model.DB.AutoMigrate(&model.UpstreamBillingStatementRecord{}))
	statement, snapshot, records := frozenUpstreamBillingFixture(t)
	require.NoError(t, model.DB.Create(&records).Error)
	t.Cleanup(func() {
		model.DB.Where("statement_id = ?", statement.ID).Delete(&model.UpstreamBillingStatementRecord{})
	})
	store := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildUpstreamBillingArchive(context.Background(), statement, store))
	require.Len(t, store.files, 5)
	var manifest UpstreamBillingManifest
	require.NoError(t, common.Unmarshal(store.files["manifest-0"], &manifest))
	assert.Equal(t, int64(4), manifest.Rows)
	assert.Equal(t, int64(1900000), manifest.ChargeQuota)
	assert.Equal(t, int64(250000), manifest.RefundQuota)
	assert.Equal(t, statement.SourceSHA256, manifest.SourceSHA256)
	for _, binding := range []struct{ kind, digest string }{{"snapshot", statement.SnapshotSHA256}, {"xlsx", statement.ExcelSHA256}, {"pdf", statement.PDFSHA256}, {"manifest", statement.ManifestSHA256}} {
		hash := sha256.Sum256(store.files[binding.kind+"-0"])
		assert.Equal(t, binding.digest, hex.EncodeToString(hash[:]))
	}
	compressed, err := gzip.NewReader(bytes.NewReader(store.files["details-0"]))
	require.NoError(t, err)
	details, err := io.ReadAll(compressed)
	require.NoError(t, err)
	require.NoError(t, compressed.Close())
	assert.Contains(t, string(details), `"channel_id":19`)
	assert.Contains(t, string(details), `"channel_id":26`)
	assert.NotContains(t, string(details), "api_key")
	assert.NotContains(t, string(details), "request_body")
	assert.Equal(t, 4, bytes.Count(details, []byte{'\n'}))
	originalPDF := append([]byte(nil), store.files["pdf-0"]...)
	originalExcel := append([]byte(nil), store.files["xlsx-0"]...)
	require.NoError(t, BuildUpstreamBillingArchive(context.Background(), statement, store), "retry must reuse the exact frozen data and bytes")
	_, err = BuildUpstreamBillingReceipt(context.Background(), statement, store)
	require.Error(t, err)
	statement.ConfirmedAt, statement.ConfirmedBy = statement.CreatedAt+3600, 9
	_, err = BuildUpstreamBillingReceipt(context.Background(), statement, store)
	require.NoError(t, err)
	receipt := append([]byte(nil), store.files["receipt-0"]...)
	_, err = BuildUpstreamBillingReceipt(context.Background(), statement, store)
	require.NoError(t, err)
	assert.Equal(t, receipt, store.files["receipt-0"])
	assert.Equal(t, originalPDF, store.files["pdf-0"])
	assert.Equal(t, originalExcel, store.files["xlsx-0"])
	if output := os.Getenv("UPSTREAM_BILLING_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.MkdirAll(output, 0700))
		for name, body := range map[string][]byte{"statement.pdf": originalPDF, "statement.xlsx": originalExcel, "receipt.pdf": receipt, "snapshot.json": []byte(statement.Snapshot), "details.jsonl.gz": store.files["details-0"], "manifest.json": store.files["manifest-0"]} {
			require.NoError(t, os.WriteFile(filepath.Join(output, name), body, 0600))
		}
		for name, value := range map[string]any{"statement.json": statement, "records.json": records, "artifacts.json": store.artifacts} {
			body, err := common.Marshal(value)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(filepath.Join(output, name), body, 0600))
		}
	}
	assert.Equal(t, "23.100000", snapshot.Total.Amount)
}

func TestUpstreamArchiveRejectsTamperedFrozenDetailsAndSummaries(t *testing.T) {
	require.NoError(t, model.DB.AutoMigrate(&model.UpstreamBillingStatementRecord{}))
	statement, _, records := frozenUpstreamBillingFixture(t)
	statement.ID = "DEMO-UPSTREAM-TAMPER"
	for i := range records {
		records[i].StatementID = statement.ID
	}
	require.NoError(t, model.DB.Create(&records).Error)
	t.Cleanup(func() {
		model.DB.Where("statement_id = ?", statement.ID).Delete(&model.UpstreamBillingStatementRecord{})
	})
	require.NoError(t, model.DB.Model(&model.UpstreamBillingStatementRecord{}).Where("statement_id = ? AND sequence = ?", statement.ID, 1).Update("cost_quota", 999999).Error)
	store := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
	assert.ErrorIs(t, BuildUpstreamBillingArchive(context.Background(), statement, store), model.ErrBillingEvidenceIntegrity)
	assert.NotContains(t, store.files, "manifest-0")
	assert.NotContains(t, store.files, "pdf-0")
	statement.Snapshot += " "
	assert.ErrorIs(t, BuildUpstreamBillingArchive(context.Background(), statement, store), model.ErrBillingEvidenceIntegrity)
}

func TestUpstreamWorkbookPreservesCostPrecisionAndUntrustedNamesAsText(t *testing.T) {
	statement, snapshot, _ := frozenUpstreamBillingFixture(t)
	snapshot.Models[0].ModelName = `=HYPERLINK("https://example.invalid")`
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement.Snapshot, statement.SnapshotSHA256 = model.UpstreamBillingSnapshotData(body), hex.EncodeToString(hash[:])
	workbook, err := RenderUpstreamBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	retry, err := RenderUpstreamBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	assert.Equal(t, workbook, retry)
	book, err := excelize.OpenReader(bytes.NewReader(workbook))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, book.Close()) })
	assert.Equal(t, []string{"总览", "每日汇总", "模型汇总", "每日模型汇总", "渠道范围"}, book.GetSheetList())
	for _, cell := range []struct{ sheet, axis, value string }{{"总览", "B14", "23.1"}, {"每日汇总", "F2", "14"}, {"每日汇总", "G2", "3.5"}, {"每日汇总", "H2", "10.5"}, {"模型汇总", "A2", snapshot.Models[0].ModelName}, {"渠道范围", "C4", "0.000000"}, {"渠道范围", "D5", "排除（未配置成本）"}} {
		value, err := book.GetCellValue(cell.sheet, cell.axis, excelize.Options{RawCellValue: true})
		require.NoError(t, err)
		assert.Equal(t, cell.value, value, cell.sheet+"!"+cell.axis)
	}
	formula, err := book.GetCellFormula("模型汇总", "A2")
	require.NoError(t, err)
	assert.Empty(t, formula)
	assert.Equal(t, "1234567890.123456", billingExcelAmount("1234567890.123456"))
	snapshot.Total.Amount = "24.000000"
	_, err = RenderUpstreamBillingStatementExcel(statement, snapshot)
	assert.ErrorIs(t, err, model.ErrBillingEvidenceIntegrity, "export must reject a caller's mutable replacement of the frozen snapshot")
}
