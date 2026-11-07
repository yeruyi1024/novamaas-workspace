package service

import (
	"bytes"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestLogExportWorkbookWritesTypedCellsWithReadableFormats(t *testing.T) {
	createdAt := int64(1790000000)
	workbook, err := NewLogExportWorkbook(true)
	require.NoError(t, err)
	defer workbook.Close()
	workbook.Add(&model.Log{
		CreatedAt: createdAt, Type: model.LogTypeConsume, Username: "alice", TokenName: "prod-key", Group: "standard", ModelName: "model-a",
		IsStream: true, PromptTokens: 1200, CompletionTokens: 30, Quota: 500000, UseTime: 4, Ip: "10.0.0.1", RequestId: "req-1",
		Other: `{"cache_tokens":800,"cache_creation_tokens_5m":50,"cache_creation_tokens_1h":25}`, Content: "=HYPERLINK(\"https://example.invalid\")",
	})
	workbook.Add(&model.Log{CreatedAt: createdAt - 60, Type: model.LogTypeConsume, Username: "alice", RequestId: "req-2", ModelName: "model-b"})
	out, err := workbook.Write(createdAt-60, createdAt)
	require.NoError(t, err)

	book, err := excelize.OpenReader(bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("使用日志")
	require.NoError(t, err)

	require.Len(t, rows, 3)
	assert.Equal(t, "用户", rows[0][2])
	assert.Equal(t, time.Unix(createdAt, 0).In(billingLocation).Format(time.DateTime), rows[1][0], "the time shows the statement time zone's wall clock")
	assert.Equal(t, []string{"消耗", "alice", "prod-key", "standard", "model-a", "是"}, rows[1][1:7])
	assert.Equal(t, []string{"1,200", "30", "800", "75"}, rows[1][7:11], "tokens are numbers with thousands separators")
	assert.Equal(t, "$1.000000", rows[1][11], "the fee is a number in the display currency")
	assert.Equal(t, "4", rows[1][12])
	assert.Equal(t, "消耗", rows[2][1])
	formula, err := book.GetCellFormula("使用日志", "P2")
	require.NoError(t, err)
	assert.Empty(t, formula, "log text must never become a formula")
	text, err := book.GetCellValue("使用日志", "P2")
	require.NoError(t, err)
	assert.Equal(t, "=HYPERLINK(\"https://example.invalid\")", text)
	for column, width := range map[string]float64{"A": 21, "L": 16, "O": 44} {
		got, err := book.GetColWidth("使用日志", column)
		require.NoError(t, err)
		assert.Equal(t, width, got, column)
	}
}

func TestLogExportWorkbookFoldsTheLaterLogsOfATaskIntoItsFirstCharge(t *testing.T) {
	createdAt := int64(1790000000)
	workbook, err := NewLogExportWorkbook(false)
	require.NoError(t, err)
	defer workbook.Close()
	// Newest first, as the export reads them. Task 1 cost less than was held;
	// task 2 cost more, so a second consumption log charged the difference.
	workbook.Add(&model.Log{CreatedAt: createdAt + 400, Type: model.LogTypeConsume, Quota: 100000, ModelName: "model-b", RequestId: "extra-2", Other: `{"task_id":"task_2","actual_quota":600000}`})
	workbook.Add(&model.Log{CreatedAt: createdAt + 180, Type: model.LogTypeRefund, Quota: 200000, ModelName: "model-a", RequestId: "refund-1", Other: `{"task_id":"task_1"}`})
	workbook.Add(&model.Log{CreatedAt: createdAt + 10, Type: model.LogTypeConsume, Quota: 500000, ModelName: "model-b", RequestId: "submit-2", Other: `{"task_id":"task_2","is_task":true}`})
	workbook.Add(&model.Log{CreatedAt: createdAt, Type: model.LogTypeConsume, Quota: 500000, ModelName: "model-a", RequestId: "submit-1", Other: `{"task_id":"task_1","is_task":true}`})
	workbook.Add(&model.Log{CreatedAt: createdAt - 60, Type: model.LogTypeRefund, Quota: 100000, RequestId: "orphan-refund", Other: `{"task_id":"task_other"}`})
	out, err := workbook.Write(createdAt-60, createdAt+400)
	require.NoError(t, err)

	book, err := excelize.OpenReader(bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("使用日志")
	require.NoError(t, err)

	require.Len(t, rows, 4, "a task is one row; a refund whose charge is outside the export stays its own row")
	assert.Equal(t, []string{"首次扣费", "补扣金额", "退款金额", "实际扣费金额"}, rows[0][10:14])
	assert.Equal(t, "submit-2", rows[1][16])
	assert.Equal(t, []string{"$1.000000", "$0.200000", "$0.000000", "$1.200000"}, rows[1][10:14], "the hold, the extra charge and what the task cost")
	assert.Equal(t, "提交时扣 $1.000000，完成后补扣 $0.200000，实际 $1.200000", rows[1][17], "how the charge came about")
	assert.Equal(t, "提交时扣 $1.000000，完成后退回 $0.400000，实际 $0.600000", rows[2][17])
	assert.Equal(t, "submit-1", rows[2][16])
	assert.Equal(t, []string{"$1.000000", "$0.000000", "$0.400000", "$0.600000"}, rows[2][10:14], "the hold, the refund and what the task cost")
	assert.Equal(t, "退款", rows[3][1])
	assert.Equal(t, "orphan-refund", rows[3][16])
	assert.Equal(t, []string{"", "", "$0.200000", "($0.200000)"}, rows[3][10:14], "a refund on its own row has no charge; it is the refund and the net is negative")
	assert.Equal(t, []string{"汇总", "使用日志"}, book.GetSheetList())
	summary, err := book.GetRows("汇总")
	require.NoError(t, err)
	assert.Equal(t, []string{"消费金额", "$2.200000"}, summary[6])
	assert.Equal(t, []string{"退款金额", "$0.600000"}, summary[7])
	assert.Equal(t, []string{"净消费金额", "$1.600000"}, summary[8])
	assert.Equal(t, []string{"原始日志条数", "5"}, summary[10])
	assert.Equal(t, []string{"导出明细行数", "3"}, summary[11])
	assert.Equal(t, time.Unix(createdAt-60, 0).In(billingLocation).Format(time.DateTime), summary[2][1])
	assert.Equal(t, time.Unix(createdAt+400, 0).In(billingLocation).Format(time.DateTime), summary[3][1])
	assert.Equal(t, []string{"币种", "USD"}, summary[4])
	modelRows, err := book.GetRows("汇总", excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	assert.Equal(t, [][]string{
		{"模型", "消费金额", "退款金额", "净消费金额"},
		{"model-b", "1.2", "0", "1.2"},
		{"model-a", "1", "0.4", "0.6"},
		{"未标注模型", "0", "0.2", "-0.2"},
		{"合计", "2.2", "0.6", "1.6"},
	}, modelRows[16:])
}

func TestLogExportModelTotalsDiscloseRoundingAndKeepCaseDistinctModels(t *testing.T) {
	settings := operation_setting.GetGeneralSetting()
	savedSettings, savedRate, savedUnit := *settings, operation_setting.USDExchangeRate, common.QuotaPerUnit
	t.Cleanup(func() {
		*settings, operation_setting.USDExchangeRate, common.QuotaPerUnit = savedSettings, savedRate, savedUnit
	})
	settings.QuotaDisplayType, operation_setting.USDExchangeRate, common.QuotaPerUnit = "CNY", 7.1, 500000
	workbook, err := NewLogExportWorkbook(false)
	require.NoError(t, err)
	defer workbook.Close()
	for _, name := range []string{"Model", "model", "other"} {
		workbook.Add(&model.Log{CreatedAt: 1790000000, Type: model.LogTypeConsume, Quota: 1, ModelName: name})
	}
	out, err := workbook.Write(1790000000, 1790000100)
	require.NoError(t, err)
	book, err := excelize.OpenReader(bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("汇总", excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	assert.Equal(t, [][]string{
		{"Model", "0.000014", "0", "0.000014"},
		{"model", "0.000014", "0", "0.000014"},
		{"other", "0.000014", "0", "0.000014"},
		{"舍入差额", "0.000001", "0", "0.000001"},
		{"合计", "0.000043", "0", "0.000043"},
	}, rows[17:])
}

func TestLogExportWorkbookFoldsAnOlderTaskLogThatRecordedNoTaskID(t *testing.T) {
	truncate(t)
	createdAt := int64(1790000000)
	require.NoError(t, model.DB.AutoMigrate(&model.Task{}))
	require.NoError(t, model.DB.Create(&model.Task{TaskID: "task_old", UserId: 9, SubmitTime: createdAt}).Error)
	workbook, err := NewLogExportWorkbook(false)
	require.NoError(t, err)
	defer workbook.Close()
	// Newest first, as the export reads them: only the refund's log has a task ID.
	workbook.Add(&model.Log{UserId: 9, CreatedAt: createdAt + 120, Type: model.LogTypeRefund, Quota: 200000, RequestId: "refund", Other: `{"task_id":"task_old"}`})
	workbook.Add(&model.Log{UserId: 9, CreatedAt: createdAt, Type: model.LogTypeConsume, Quota: 500000, RequestId: "hold", Other: `{"is_task":true}`})
	out, err := workbook.Write(createdAt, createdAt+120)
	require.NoError(t, err)

	book, err := excelize.OpenReader(bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("使用日志")
	require.NoError(t, err)

	require.Len(t, rows, 2, "the hold and its refund are one row")
	assert.Equal(t, "hold", rows[1][16])
	assert.Equal(t, []string{"$1.000000", "$0.000000", "$0.400000", "$0.600000"}, rows[1][10:14], "charged, refunded, and what the task cost")
}

func TestLogExportSummaryHandlesNoLogsAndRefundOnlySelections(t *testing.T) {
	for _, tc := range []struct {
		name    string
		refund  int
		wantNet string
	}{
		{"empty selection", 0, "0"},
		{"refund only", 100000, "-0.2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			workbook, err := NewLogExportWorkbook(false)
			require.NoError(t, err)
			defer workbook.Close()
			if tc.refund > 0 {
				workbook.Add(&model.Log{CreatedAt: 1790000000, Type: model.LogTypeRefund, Quota: tc.refund})
			}
			out, err := workbook.Write(1790000000, 1790000100)
			require.NoError(t, err)
			book, err := excelize.OpenReader(bytes.NewReader(out.Bytes()))
			require.NoError(t, err)
			defer book.Close()
			value, err := book.GetCellValue("汇总", "B9", excelize.Options{RawCellValue: true})
			require.NoError(t, err)
			assert.Equal(t, tc.wantNet, value)
			cellType, err := book.GetCellType("汇总", "B9")
			require.NoError(t, err)
			assert.Contains(t, []excelize.CellType{excelize.CellTypeNumber, excelize.CellTypeUnset}, cellType, "Excel also permits omitting the type attribute for numeric cells")
		})
	}
}
