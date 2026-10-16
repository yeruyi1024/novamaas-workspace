package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// This fixture carries genuine missing and estimated evidence through the same
// report aggregates and independent archive verifier as a generated statement.
func frozenUpstreamBillingWarningFixture(t *testing.T) (*model.UpstreamBillingStatement, *UpstreamBillingSnapshot, []model.UpstreamBillingStatementRecord) {
	t.Helper()
	statement, snapshot, records := frozenUpstreamBillingFixture(t)
	statement.ID = "DEMO-UPSTREAM-WARNING-202609"
	records[2].CostMissing, records[2].UsageEstimated = true, true
	records[2].CostQuota, records[2].CostSnapshotID = 0, 0
	records[2].CostDiscount, records[2].CostBasisQuota = "", ""
	records[3].UsageMissing, records[3].TaskPending = true, true
	records[3].VideoTokens, records[3].VideoSeconds = 0, 0
	records[3].UsageSource = "task_pending"
	var total upstreamBillingAggregate
	days := make(map[string]*upstreamBillingAggregate)
	models := make(map[upstreamArchiveModelKey]*upstreamBillingAggregate)
	var source bytes.Buffer
	for i := range records {
		records[i].StatementID = statement.ID
		record := records[i]
		require.NoError(t, total.Add(record))
		day := time.Unix(record.OccurredAt, 0).In(billingLocation).Format("2006-01-02")
		if days[day] == nil {
			days[day] = &upstreamBillingAggregate{}
		}
		require.NoError(t, days[day].Add(record))
		key := upstreamArchiveModelKey{record.ModelName, record.UpstreamModelName, record.BillingGroup}
		if models[key] == nil {
			models[key] = &upstreamBillingAggregate{}
		}
		require.NoError(t, models[key].Add(record))
		body, err := common.Marshal(record)
		require.NoError(t, err)
		source.Write(body)
		source.WriteByte('\n')
	}
	var err error
	snapshot.Total, err = total.Row("合计", snapshot.Currency)
	require.NoError(t, err)
	for i, day := range snapshot.Days {
		aggregate := days[day.Label]
		if aggregate == nil {
			aggregate = &upstreamBillingAggregate{}
		}
		snapshot.Days[i], err = aggregate.Row(day.Label, snapshot.Currency)
		require.NoError(t, err)
	}
	for i := range snapshot.Models {
		row := &snapshot.Models[i]
		row.UpstreamBillingRow, err = models[upstreamArchiveModelKey{row.ModelName, row.UpstreamModelName, row.BillingGroup}].Row(row.Label, snapshot.Currency)
		require.NoError(t, err)
	}
	for i := range snapshot.DayModels {
		row := &snapshot.DayModels[i]
		row.UpstreamBillingRow, err = models[upstreamArchiveModelKey{row.ModelName, row.UpstreamModelName, row.BillingGroup}].Row(row.Label, snapshot.Currency)
		require.NoError(t, err)
	}
	snapshot.Coverage.Complete = false
	snapshot.Coverage.MissingCostRecords = snapshot.Total.MissingCostRecords
	snapshot.Coverage.MissingUsageRecords = snapshot.Total.MissingUsageRecords
	snapshot.Coverage.EstimatedUsageRecords = snapshot.Total.EstimatedUsageRecords
	snapshot.Coverage.PendingTasks = snapshot.Total.PendingTasks
	snapshot.SchemaVersion = 2
	for _, channel := range snapshot.Group.Channels {
		if !channel.Eligible {
			continue
		}
		var aggregate upstreamBillingAggregate
		for _, record := range records {
			if record.ChannelID == channel.ID {
				require.NoError(t, aggregate.Add(record))
			}
		}
		row, err := aggregate.Row("", snapshot.Currency)
		require.NoError(t, err)
		net := row.ChargeQuota - row.RefundQuota
		snapshot.ChannelUsers = append(snapshot.ChannelUsers, UpstreamBillingChannelUserRow{ChannelID: channel.ID, ChannelName: channel.Name, UserID: 42, Username: "demo-user", DisplayName: "示例客户", ChannelAmount: row.Amount, SharePercent: upstreamBillingCostShare(net, net, row.MissingCostRecords), UpstreamBillingRow: row})
	}
	sourceHash := sha256.Sum256(source.Bytes())
	statement.SourceSHA256 = hex.EncodeToString(sourceHash[:])
	snapshot.SourceSHA256 = statement.SourceSHA256
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement.Snapshot, statement.SnapshotSHA256 = model.UpstreamBillingSnapshotData(body), hex.EncodeToString(hash[:])
	return statement, snapshot, records
}

func TestUpstreamArchiveRetainsWarningsAndExportsUnknownAmounts(t *testing.T) {
	require.NoError(t, model.DB.AutoMigrate(&model.UpstreamBillingStatementRecord{}))
	statement, _, records := frozenUpstreamBillingWarningFixture(t)
	require.NoError(t, model.DB.Create(&records).Error)
	t.Cleanup(func() {
		model.DB.Where("statement_id = ?", statement.ID).Delete(&model.UpstreamBillingStatementRecord{})
	})
	store := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildUpstreamBillingArchive(context.Background(), statement, store), "quality warnings must not prevent a verifiable frozen archive")
	require.NotEmpty(t, store.files["pdf-0"])
	book, err := excelize.OpenReader(bytes.NewReader(store.files["xlsx-0"]))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, book.Close()) })
	for _, cell := range []struct{ sheet, axis, value string }{
		{"总览", "B1", "数据质量告警：用于核对，缺口待核实"},
		{"总览", "A14", "已知净成本小计"}, {"总览", "B14", "10.5"},
		{"总览", "B34", "2026-10-01 01:00:00 Asia/Shanghai"},
		{"模型汇总", "K3", "未知"}, {"模型汇总", "L3", "未知"}, {"模型汇总", "M3", "未知"},
		{"模型汇总", "Z3", "成本未知"}, {"模型汇总", "AA3", "估算用量 1 条"},
		{"模型汇总", "M4", "0"}, {"模型汇总", "Z4", "成本完整"}, {"模型汇总", "AA4", "用量未知 1 条 / 待结算 1 条"},
		{"渠道用户成本", "F3", "未知"}, {"渠道用户成本", "G3", "未知"}, {"渠道用户成本", "H3", "未知"}, {"渠道用户成本", "I3", "未知"},
		{"渠道用户成本", "J3", "—"}, {"渠道用户成本", "L3", "成本未知"}, {"渠道用户成本", "M3", "成本未知"},
		{"渠道用户成本", "N3", "估算用量 1 条"}, {"渠道用户成本", "H4", "0"}, {"渠道用户成本", "N4", "用量未知 1 条 / 待结算 1 条"},
	} {
		value, err := book.GetCellValue(cell.sheet, cell.axis, excelize.Options{RawCellValue: true})
		require.NoError(t, err)
		assert.Equal(t, cell.value, value, cell.sheet+"!"+cell.axis)
	}
	warning, err := book.GetCellValue("总览", "B35")
	require.NoError(t, err)
	assert.Contains(t, warning, "缺少成本 1，用量未知 1，估算用量 1，待结算 1")
	compressed, err := gzip.NewReader(bytes.NewReader(store.files["details-0"]))
	require.NoError(t, err)
	details, err := io.ReadAll(compressed)
	require.NoError(t, err)
	require.NoError(t, compressed.Close())
	for _, flag := range []string{`"cost_missing":true`, `"usage_missing":true`, `"usage_estimated":true`, `"task_pending":true`} {
		assert.Contains(t, string(details), flag)
	}
	statement.ConfirmedAt, statement.ConfirmedBy = statement.CreatedAt+3600, 9
	_, err = BuildUpstreamBillingReceipt(context.Background(), statement, store)
	require.NoError(t, err, "a confirmed warning-bearing statement also has a receipt")
	if output := os.Getenv("UPSTREAM_BILLING_WARNING_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.MkdirAll(output, 0700))
		for name, body := range map[string][]byte{"statement.pdf": store.files["pdf-0"], "statement.xlsx": store.files["xlsx-0"], "receipt.pdf": store.files["receipt-0"], "snapshot.json": []byte(statement.Snapshot)} {
			require.NoError(t, os.WriteFile(filepath.Join(output, name), body, 0600))
		}
	}
}

func TestUpstreamExportsAllMissingCostsAsUnknown(t *testing.T) {
	statement, snapshot, _ := frozenUpstreamBillingWarningFixture(t)
	snapshot.Total.MissingCostRecords = snapshot.Total.Count
	snapshot.Coverage.MissingCostRecords = snapshot.Total.Count
	snapshot.Total.Charge, snapshot.Total.Refund, snapshot.Total.Amount = "0.000000", "0.000000", "0.000000"
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement.Snapshot, statement.SnapshotSHA256 = model.UpstreamBillingSnapshotData(body), hex.EncodeToString(hash[:])
	xlsx, err := RenderUpstreamBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	book, err := excelize.OpenReader(bytes.NewReader(xlsx))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, book.Close()) })
	for _, axis := range []string{"B12", "B13", "B14"} {
		value, err := book.GetCellValue("总览", axis, excelize.Options{RawCellValue: true})
		require.NoError(t, err)
		assert.Equal(t, "未知", value)
	}
	label, err := book.GetCellValue("总览", "A14")
	require.NoError(t, err)
	assert.Equal(t, "净成本未知", label)
	pdf, err := RenderUpstreamBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(pdf, []byte("%PDF-")))
}

func TestUpstreamExportsMissingAndExcludedChannelScope(t *testing.T) {
	statement, snapshot, _ := frozenUpstreamBillingFixture(t)
	snapshot.Group.Channels = append(snapshot.Group.Channels, model.UpstreamBillingChannelView{ID: 99, Name: "缺失渠道", Missing: true})
	snapshot.Coverage.Complete, snapshot.Coverage.MissingChannels = false, 1
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement.Snapshot, statement.SnapshotSHA256 = model.UpstreamBillingSnapshotData(body), hex.EncodeToString(hash[:])
	partialXLSX, err := RenderUpstreamBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	partialBook, err := excelize.OpenReader(bytes.NewReader(partialXLSX))
	require.NoError(t, err)
	partialLabel, err := partialBook.GetCellValue("总览", "A14")
	require.NoError(t, err)
	assert.Equal(t, "已知净成本小计", partialLabel, "an absent member channel leaves the whole supplier amount incomplete")
	partialAmount, err := partialBook.GetCellValue("总览", "B14", excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	assert.Equal(t, "23.1", partialAmount, "known remaining channels keep their numeric subtotal")
	require.NoError(t, partialBook.Close())
	for i := range snapshot.Group.Channels {
		snapshot.Group.Channels[i].Eligible = false
	}
	snapshot.Group.Channels[0].Missing = true
	snapshot.Coverage.MissingChannels = 2
	body, err = common.Marshal(snapshot)
	require.NoError(t, err)
	hash = sha256.Sum256(body)
	statement.Snapshot, statement.SnapshotSHA256 = model.UpstreamBillingSnapshotData(body), hex.EncodeToString(hash[:])
	xlsx, err := RenderUpstreamBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	book, err := excelize.OpenReader(bytes.NewReader(xlsx))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, book.Close()) })
	for _, cell := range []struct{ sheet, axis, expected string }{
		{"总览", "B14", "未知"}, {"渠道范围", "C2", "未知"}, {"渠道范围", "D2", "渠道不存在（缺失）"},
		{"渠道范围", "D3", "排除（未配置成本）"}, {"每日汇总", "H2", "未知"}, {"每日汇总", "U2", "无可纳入成本统计的渠道"},
	} {
		value, err := book.GetCellValue(cell.sheet, cell.axis, excelize.Options{RawCellValue: true})
		require.NoError(t, err)
		assert.Equal(t, cell.expected, value)
	}
	warning, err := book.GetCellValue("总览", "B35")
	require.NoError(t, err)
	assert.Contains(t, warning, "无可纳入成本统计的渠道")
}
