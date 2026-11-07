package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func frozenBillingExportFixture(t *testing.T) (*model.BillingStatement, *BillingSnapshot) {
	t.Helper()
	start, end, err := model.BillingMonthBounds("2026-09")
	require.NoError(t, err)
	snapshot, err := buildBillingSnapshot(&model.BillingAccount{UserID: 42, CompanyTitle: "示例科技有限公司（演示）", TaxID: "DEMO-TAX", AccountingStartAt: start + 27*86400}, "2026-09", []model.BillingHour{{Hour: start + 28*86400, Charge: 500000, Refund: 125000, Count: 2}}, BillingCurrency{Code: "CNY", Symbol: "¥", Rate: "7", QuotaPerUnit: "500000"})
	require.NoError(t, err)
	require.NoError(t, attachBillingModelRows(snapshot, []model.BillingModelTotal{{ModelName: "enterprise-context-analysis-pro-2026-09", Charge: 500000, Refund: 125000, Count: 2, ChargeCount: 1, RefundCount: 1, ActiveDays: 1, FirstPosted: start + 28*86400, LastPosted: start + 28*86400 + 1}}))
	snapshot.PDFTemplateVersion, snapshot.PDFLogoPNG, snapshot.Issuer = 9, billingPlatformMarkV2, "new-api 演示平台"
	snapshot.Username, snapshot.OperatingName = "demo_customer", "演示运营主体"
	snapshot.RateReferenceAt = end + 86400
	snapshot.Models[0].CurrentRates = []BillingGroupRate{{Group: "standard", Ratio: "0.8"}, {Group: "preferred", Ratio: "0.65"}}
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "DEMO-EXPORT-202609", UserID: 42, Month: "2026-09", Revision: 2, CreatedAt: end + 86400, StartAt: start + 27*86400, EndAt: end, FromSequence: 1, ToSequence: 2, Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:])}
	return statement, snapshot
}

func TestBillingExcelPreservesFrozenAmountsRatesAndNoFormulaInjection(t *testing.T) {
	statement, snapshot := frozenBillingExportFixture(t)
	snapshot.Models[0].ModelName = "=HYPERLINK(\"https://example.invalid\")"
	frozen, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(frozen)
	statement.Snapshot, statement.SnapshotSHA256 = string(frozen), hex.EncodeToString(hash[:])
	body, err := RenderBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	retry, err := RenderBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	assert.Equal(t, body, retry, "immutable archives require byte-identical retries")
	book, err := excelize.OpenReader(bytes.NewReader(body))
	require.NoError(t, err)
	defer book.Close()
	assert.Equal(t, []string{"总览", "每日汇总", "模型汇总"}, book.GetSheetList())
	for _, cell := range []struct{ sheet, axis, value string }{
		{"总览", "B16", "5.250000"}, {"总览", "B27", statement.SnapshotSHA256},
		{"每日汇总", "D30", "5.250000"}, {"每日汇总", "F2", "正式记账起点之前"},
		{"模型汇总", "H2", "5.250000"}, {"模型汇总", "I2", "65.00–80.00%"}, {"模型汇总", "J2", "未记录"},
		{"模型汇总", "L2", "standard: 80.00%; preferred: 65.00%"},
		{"模型汇总", "A2", snapshot.Models[0].ModelName},
	} {
		value, err := book.GetCellValue(cell.sheet, cell.axis)
		require.NoError(t, err)
		assert.Equal(t, cell.value, value, cell.sheet+"!"+cell.axis)
	}
	formula, err := book.GetCellFormula("模型汇总", "A2")
	require.NoError(t, err)
	assert.Empty(t, formula, "untrusted model names must remain plain text")
	assert.Equal(t, "1234567890.123456", billingExcelAmount("1234567890.123456"), "large amounts must retain exact digits as text")
	assert.Equal(t, 0.000001, billingExcelAmount("0.000001"))
}

func TestBillingExcelV18ShowsNetConsumptionBeforeRefunds(t *testing.T) {
	statement, snapshot := frozenBillingExportFixture(t)
	snapshot.PDFTemplateVersion = 18
	snapshot.Models[0].BillingGroup, snapshot.Models[0].BillingRate = "standard", "0.8"
	frozen, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(frozen)
	statement.Snapshot, statement.SnapshotSHA256 = string(frozen), hex.EncodeToString(hash[:])
	body, err := RenderBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	retry, err := RenderBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	assert.Equal(t, body, retry, "customer-facing layout changes must keep archive retries deterministic")
	book, err := excelize.OpenReader(bytes.NewReader(body))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, book.Close()) })

	for _, table := range []struct {
		sheet   string
		headers []string
	}{
		{"每日汇总", []string{"日期", "预扣金额", "实际消费金额", "退款金额", "正式账本记录数", "记账范围"}},
		{"模型汇总", []string{"模型名称", "使用账单天数", "消费笔数", "退款笔数", "正式账本记录数", "预扣金额", "实际消费金额", "退款金额", "计费费率", "消费金额占比", "实际计费组", "首笔入账时间", "末笔入账时间"}},
	} {
		rows, err := book.GetRows(table.sheet)
		require.NoError(t, err)
		require.NotEmpty(t, rows)
		assert.Equal(t, table.headers, rows[0], table.sheet)
	}
	for _, cell := range []struct{ sheet, axis, value string }{
		{"每日汇总", "B30", "7"}, {"每日汇总", "C30", "5.25"}, {"每日汇总", "D30", "1.75"},
		{"每日汇总", "B32", "7"}, {"每日汇总", "C32", "5.25"}, {"每日汇总", "D32", "1.75"},
		{"模型汇总", "F2", "7"}, {"模型汇总", "G2", "5.25"}, {"模型汇总", "H2", "1.75"},
		{"模型汇总", "F3", "7"}, {"模型汇总", "G3", "5.25"}, {"模型汇总", "H3", "1.75"},
	} {
		value, err := book.GetCellValue(cell.sheet, cell.axis, excelize.Options{RawCellValue: true})
		require.NoError(t, err)
		assert.Equal(t, cell.value, value, cell.sheet+"!"+cell.axis)
		styleID, err := book.GetCellStyle(cell.sheet, cell.axis)
		require.NoError(t, err)
		style, err := book.GetStyle(styleID)
		require.NoError(t, err)
		require.NotNil(t, style.CustomNumFmt)
		assert.Equal(t, `_(* #,##0.000000_);_(* (#,##0.000000);_(* "-"??????_);_(@_)`, *style.CustomNumFmt, "all reordered amount cells retain accounting precision")
	}
	for _, cell := range []struct{ sheet, axis, value string }{
		{"总览", "A14", "消费金额"}, {"总览", "A15", "退款金额"}, {"总览", "A16", "本期净额"},
		{"总览", "A27", "文件绑定校验"}, {"总览", "A28", "金额精度"}, {"总览", "A29", "折扣金额说明"},
		{"模型汇总", "I2", "80.00%"}, {"模型汇总", "J2", "100.00%"}, {"模型汇总", "K2", "standard"},
	} {
		value, err := book.GetCellValue(cell.sheet, cell.axis)
		require.NoError(t, err)
		assert.Equal(t, cell.value, value, cell.sheet+"!"+cell.axis)
	}
	overview, err := book.GetRows("总览")
	require.NoError(t, err)
	for _, row := range overview {
		assert.NotContains(t, row, "账单数据 SHA-256")
		assert.NotContains(t, row, statement.SnapshotSHA256)
	}
	for _, row := range []int{27, 28, 29} {
		height, err := book.GetRowHeight("总览", row)
		require.NoError(t, err)
		assert.Equal(t, 44.0, height, "the notes must keep enough height after removing the fingerprint row")
	}
	after, err := common.Marshal(snapshot)
	require.NoError(t, err)
	assert.Equal(t, frozen, after, "presentation changes must preserve the frozen accounting source")
	require.NoError(t, statement.VerifySnapshot())
	statement.Snapshot += " "
	corrupted, err := RenderBillingStatementExcel(statement, snapshot)
	assert.ErrorIs(t, err, model.ErrBillingEvidenceIntegrity, "the hidden fingerprint must still protect exported data")
	assert.Empty(t, corrupted)
	if output := os.Getenv("BILLING_EXPORT_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.MkdirAll(output, 0700))
		require.NoError(t, os.WriteFile(filepath.Join(output, "billing-v18.xlsx"), body, 0600))
	}
}

func TestBillingArchiveV9BindsExcelAndPDFAndRetriesUnchanged(t *testing.T) {
	truncate(t)
	statement, snapshot := frozenBillingExportFixture(t)
	posted := snapshot.Models[0].FirstPosted
	entries := []model.BillingEntry{
		{EventKey: "export-charge", UserID: 42, Sequence: 1, PostedAt: posted, Kind: "usage", Quota: 500000, ModelName: snapshot.Models[0].ModelName},
		{EventKey: "export-refund", UserID: 42, Sequence: 2, PostedAt: posted + 1, Kind: "refund", Quota: -125000, ModelName: snapshot.Models[0].ModelName},
	}
	require.NoError(t, model.DB.Create(&entries).Error)
	store := &memoryBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))
	require.Len(t, store.files, 5)
	var manifest BillingManifest
	require.NoError(t, common.Unmarshal(store.files["manifest"], &manifest))
	excelHash := sha256.Sum256(store.files["xlsx"])
	assert.Equal(t, hex.EncodeToString(excelHash[:]), statement.ExcelSHA256)
	assert.Equal(t, statement.ExcelSHA256, manifest.ExcelSHA256)
	assert.Equal(t, statement.PDFSHA256, manifest.PDFSHA256)
	assert.Equal(t, statement.SnapshotSHA256, manifest.SnapshotSHA256)
	frozenPDF, frozenExcel, frozenManifest := store.files["pdf"], store.files["xlsx"], store.files["manifest"]
	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))
	assert.Equal(t, frozenPDF, store.files["pdf"])
	assert.Equal(t, frozenExcel, store.files["xlsx"])
	assert.Equal(t, frozenManifest, store.files["manifest"])
	require.NoError(t, model.DB.Create(&model.BillingEntry{EventKey: "late-entry", UserID: 42, Sequence: 3, PostedAt: posted + 2, Kind: "usage", Quota: 100000, ModelName: "late-model"}).Error)
	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))
	assert.Equal(t, frozenPDF, store.files["pdf"], "later ledger entries must not enter the frozen source")
	if output := os.Getenv("BILLING_EXPORT_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.MkdirAll(output, 0700))
		for _, kind := range []string{"pdf", "xlsx", "manifest"} {
			require.NoError(t, os.WriteFile(filepath.Join(output, BillingArtifactFilename(statement, snapshot.Username, kind, 0, snapshot.CompanyTitle)), store.files[kind], 0600))
		}
	}
}

func TestBillingPDFRendersFrozenOperatingLogo(t *testing.T) {
	statement, snapshot := frozenBillingExportFixture(t)
	statement.ExcelSHA256 = strings.Repeat("b", 64)
	withoutLogo, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	// Use separate raster bytes: image deduplication must not hide the operator.
	transport := configureBillingBranding(t)
	common.OperatingEntityLogo = "https://platform.example/operator.png"
	branding, err := captureBillingDocumentBranding(context.Background())
	require.NoError(t, err)
	require.Len(t, transport.requests, 1)
	snapshot.PDFOperatingLogoPNG = branding.OperatingLogoPNG
	withLogo, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.Greater(t, strings.Count(string(withLogo), "/Subtype /Image"), strings.Count(string(withoutLogo), "/Subtype /Image"))
	if output := os.Getenv("BILLING_EXPORT_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.WriteFile(filepath.Join(output, "operator-preview.pdf"), withLogo, 0600))
	}
}

func TestBillingFilenamePrefersFrozenCompanyAndSanitizesBothFormats(t *testing.T) {
	statement := &model.BillingStatement{UserID: 42, Month: "2026-09", Revision: 2}
	for _, kind := range []string{"pdf", "xlsx"} {
		assert.Equal(t, "示例_企业_2026-09_月度对账单_V02."+kind, BillingArtifactFilename(statement, "username", kind, 0, "示例/企业\r\n"))
		assert.Equal(t, "username_2026-09_月度对账单_V02."+kind, BillingArtifactFilename(statement, "username", kind, 0, "  "))
	}
}

func TestBillingConfirmationPreservesFrozenOriginalAndCorporateIdentity(t *testing.T) {
	for _, test := range []struct{ name, company, taxID string }{
		{"standard", "上海示例科技有限公司", "DEMO-TAX"},
		{"maximum identity", strings.Repeat("企", 200), strings.Repeat("X", 64)},
	} {
		for _, version := range []int{16, 17, 18} {
			t.Run(fmt.Sprintf("%s/v%d", test.name, version), func(t *testing.T) {
				statement, snapshot := frozenBillingExportFixture(t)
				snapshot.PDFTemplateVersion = version
				snapshot.CompanyTitle, snapshot.TaxID = test.company, test.taxID
				frozen, err := common.Marshal(snapshot)
				require.NoError(t, err)
				snapshotHash := sha256.Sum256(frozen)
				statement.Snapshot, statement.SnapshotSHA256 = string(frozen), hex.EncodeToString(snapshotHash[:])
				workbook, err := RenderBillingStatementExcel(statement, snapshot)
				require.NoError(t, err)
				excelHash := sha256.Sum256(workbook)
				statement.ExcelSHA256 = hex.EncodeToString(excelHash[:])
				original, err := RenderBillingStatementPDF(statement, snapshot, false)
				require.NoError(t, err, "accepted corporate identities and complete fingerprints must fit")
				pdfHash := sha256.Sum256(original)
				statement.PDFSHA256 = hex.EncodeToString(pdfHash[:])
				manifest, err := common.Marshal(BillingManifest{StatementID: statement.ID, PDFSHA256: statement.PDFSHA256, ExcelSHA256: statement.ExcelSHA256, SnapshotSHA256: statement.SnapshotSHA256})
				require.NoError(t, err)
				manifestHash := sha256.Sum256(manifest)
				statement.ManifestSHA256 = hex.EncodeToString(manifestHash[:])
				premature, err := RenderBillingStatementPDF(statement, snapshot, true)
				require.ErrorContains(t, err, "statement is not confirmed", "archived evidence alone must never produce confirmation details")
				assert.Empty(t, premature)
				statement.ConfirmedAt = statement.CreatedAt + 3600
				confirmed, err := RenderBillingStatementPDF(statement, snapshot, true)
				require.NoError(t, err, "the confirmation must fit beside the corporate identity")
				expectedPages := strings.Count(string(original), "\n  /Type /Page\n")
				assert.Equal(t, expectedPages, strings.Count(string(confirmed), "\n  /Type /Page\n"), "confirmation appears on the cover without a separate receipt page")
				assert.NotEqual(t, original, confirmed, "the confirmed artifact must include the confirmation")
				retry, err := RenderBillingStatementPDF(statement, snapshot, false)
				require.NoError(t, err)
				assert.Equal(t, original, retry, "confirmation must not change the original PDF or its fingerprint")
				after, err := common.Marshal(snapshot)
				require.NoError(t, err)
				assert.Equal(t, frozen, after)
				require.NoError(t, statement.VerifySnapshot())
			})
		}
	}
}

func TestBillingExcelDetailSheetListsFrozenRecordsOnlyWhenEnabled(t *testing.T) {
	statement, snapshot := frozenBillingExportFixture(t)
	details := []BillingExcelDetail{
		{PostedAt: statement.StartAt + 3600, Amount: "0.800000", RequestID: "req-1", ModelName: "model-a", TokenID: 9, TokenName: "prod-key", BillingGroup: "standard", BillingRate: "0.8", Usage: model.LogLink{Found: true, IsStream: true, PromptTokens: 1200, CompletionTokens: 30, CacheTokens: 800, CacheWrite: 75, UseTime: 4, Ip: "10.0.0.1"}},
		{PostedAt: statement.StartAt + 7200, Amount: "-0.250000", RequestID: "req-2", ModelName: "model-a", TokenID: 0, Refund: true},
		{PostedAt: statement.StartAt + 7300, Amount: "0.000000", RequestID: "req-3", ModelName: "model-a", Refund: true},
	}

	legacy, err := RenderBillingStatementExcelWithDetails(statement, snapshot, details)
	require.NoError(t, err)
	legacyBook, err := excelize.OpenReader(bytes.NewReader(legacy))
	require.NoError(t, err)
	defer legacyBook.Close()
	assert.Equal(t, []string{"总览", "每日汇总", "模型汇总"}, legacyBook.GetSheetList(), "frozen snapshots without the flag keep their workbook")

	snapshot.DetailSheet = true
	body, err := RenderBillingStatementExcelWithDetails(statement, snapshot, details)
	require.NoError(t, err)
	retry, err := RenderBillingStatementExcelWithDetails(statement, snapshot, details)
	require.NoError(t, err)
	assert.Equal(t, body, retry, "immutable archives require byte-identical retries")
	book, err := excelize.OpenReader(bytes.NewReader(body))
	require.NoError(t, err)
	defer book.Close()
	assert.Equal(t, []string{"总览", "每日汇总", "模型汇总", "详细日志"}, book.GetSheetList())
	rows, err := book.GetRows("详细日志")
	require.NoError(t, err)
	require.Len(t, rows, 4)
	assert.Equal(t, []string{"时间", "用户", "令牌", "所选分组", "模型", "流", "输入 Token", "输出 Token", "缓存读取", "缓存写入", "计费费率", "首次扣费", "补扣金额", "退款金额", "实际扣费金额", "耗时(秒)", "IP", "请求ID", "扣费说明"}, rows[0])
	assert.Equal(t, []string{"demo_customer", "prod-key", "standard", "model-a", "是", "1,200", "30", "800", "75"}, rows[1][1:10], "the columns of the log export")
	assert.Equal(t, []string{"0.800000", "0.000000", "0.000000", "0.800000"}, rows[1][11:15], "a request with one charge: charged, no extra charge, no refund, what it cost")
	assert.Equal(t, []string{"4", "10.0.0.1", "req-1"}, rows[1][15:18])
	assert.Equal(t, []string{"0.000000", "0.000000", "0.250000", "(0.250000)"}, rows[2][11:15], "a refund whose charge is outside the statement is a refund, and the net is negative")
	assert.Equal(t, "未记录", rows[2][2])
	assert.Equal(t, "未记录", rows[2][3])
	assert.Equal(t, "0", rows[2][6], "an entry without a usage log has no token figures; the sheet shows the zero as a dash")
	assert.Equal(t, []string{"0.000000", "0.000000", "0.000000", "0.000000"}, rows[3][11:15], "a refund that rounds to zero lists no amount")
}

func TestBillingExcelDetailSheetLeavesOutColumnsWithNothingToShow(t *testing.T) {
	statement, snapshot := frozenBillingExportFixture(t)
	snapshot.DetailSheet = true
	details := []BillingExcelDetail{{PostedAt: statement.StartAt + 3600, Amount: "0.800000", RequestID: "req-1", ModelName: "model-a", TokenID: 9, TokenName: "prod-key", BillingGroup: "standard"}}

	body, err := RenderBillingStatementExcelWithDetails(statement, snapshot, details)
	require.NoError(t, err)
	book, err := excelize.OpenReader(bytes.NewReader(body))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("详细日志")
	require.NoError(t, err)

	require.Len(t, rows, 2)
	assert.Equal(t, []string{"时间", "用户", "令牌", "所选分组", "模型", "首次扣费", "补扣金额", "退款金额", "实际扣费金额", "请求ID", "扣费说明"}, rows[0], "no tokens, run time, IP or rate was recorded, so those columns are not filled with zeros")
	assert.Equal(t, []string{"0.800000", "0.000000", "0.000000", "0.800000"}, rows[1][5:9])
	assert.Equal(t, "req-1", rows[1][9])
}

func TestBillingExcelDetailSheetFillsAVideoRequestFromItsTaskAndRate(t *testing.T) {
	statement, snapshot := frozenBillingExportFixture(t)
	snapshot.DetailSheet = true
	details := []BillingExcelDetail{{
		PostedAt: statement.StartAt + 3600, Amount: "9.889992", RequestID: "req-video", ModelName: "video-model", TokenID: 9, TokenName: "prod-key", BillingGroup: "standard",
		TaskID: "task_1", Task: model.TaskUsage{TotalTokens: 2500000, Seconds: 480}, TaskFound: true,
		Usage: model.LogLink{Found: true, GroupRatio: "0.86", IsTask: true},
	}}

	body, err := RenderBillingStatementExcelWithDetails(statement, snapshot, details)
	require.NoError(t, err)
	book, err := excelize.OpenReader(bytes.NewReader(body))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("详细日志")
	require.NoError(t, err)

	require.Len(t, rows, 2)
	assert.Equal(t, []string{"时间", "用户", "令牌", "所选分组", "模型", "输出 Token", "计费费率", "首次扣费", "补扣金额", "退款金额", "实际扣费金额", "耗时(秒)", "请求ID", "扣费说明"}, rows[0], "a video request has no stream flag, input tokens or IP to list")
	assert.Equal(t, []string{"2,500,000", "86.00%"}, rows[1][5:7], "the task's tokens and the rate the request was billed with")
	assert.Equal(t, "9.889992", rows[1][7])
	assert.Equal(t, "480", rows[1][11])
}

func TestBillingExcelDetailSheetMergesHoldAndRefundOfOneRequest(t *testing.T) {
	statement, snapshot := frozenBillingExportFixture(t)
	snapshot.DetailSheet = true
	at := statement.StartAt + 3600
	details := []BillingExcelDetail{
		{PostedAt: at, Amount: "1.000000", RequestID: "req-a", ModelName: "model-a", TokenID: 9, TokenName: "prod-key"},
		{PostedAt: at + 10, Amount: "2.000000", RequestID: "req-b", ModelName: "model-a", TokenID: 9, TokenName: "prod-key"},
		{PostedAt: at + 600, Amount: "-0.400000", RequestID: "req-a", ModelName: "model-a", TokenID: 9, TokenName: "prod-key", BillingGroup: "standard", Refund: true},
	}

	body, err := RenderBillingStatementExcelWithDetails(statement, snapshot, details)
	require.NoError(t, err)
	book, err := excelize.OpenReader(bytes.NewReader(body))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("详细日志")
	require.NoError(t, err)

	require.Len(t, rows, 3, "one row per request")
	assert.Equal(t, "req-a", rows[1][9])
	assert.Equal(t, "首次扣费 1.000000 CNY，完成后退回 0.400000 CNY，实际扣费 0.600000 CNY", rows[1][10])
	assert.Equal(t, "按 2.000000 CNY 一次计费", rows[2][10])
	assert.Equal(t, []string{"1.000000", "0.000000", "0.400000", "0.600000"}, rows[1][5:9], "首次扣费 + 补扣金额 - 退款金额 = 实际扣费金额")
	assert.Equal(t, "standard", rows[1][3], "a later entry supplies the group the first one lacks")
	assert.Equal(t, "req-b", rows[2][9])
	assert.Equal(t, []string{"2.000000", "0.000000", "0.000000", "2.000000"}, rows[2][5:9])
}

func TestBillingExcelDetailSheetShowsAReservationApartFromTheTotals(t *testing.T) {
	statement, snapshot := frozenBillingExportFixture(t)
	snapshot.DetailSheet = true
	at := statement.StartAt + 3600
	details := []BillingExcelDetail{
		{PostedAt: at, Amount: "0.300000", RequestID: "cost-more", ModelName: "model-a", ReservedAmount: "0.100000"},
		{PostedAt: at + 10, Amount: "0.300000", RequestID: "cost-less", ModelName: "model-a", ReservedAmount: "0.500000"},
		{PostedAt: at + 20, Amount: "0.300000", RequestID: "reserved-nothing", ModelName: "model-a", ReservedAmount: "0.000000"},
	}

	body, err := RenderBillingStatementExcelWithDetails(statement, snapshot, details)
	require.NoError(t, err)
	book, err := excelize.OpenReader(bytes.NewReader(body))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("详细日志")
	require.NoError(t, err)

	require.Len(t, rows, 4)
	assert.Equal(t, []string{"预冻结金额", "首次扣费", "补扣金额", "退款金额", "实际扣费金额"}, rows[0][5:10])
	// The ledger holds one charge for a text request; what was frozen before is
	// not consumption, so it is in no total and the charge is the same either way.
	assert.Equal(t, []string{"0.100000", "0.300000", "0.000000", "0.000000", "0.300000"}, rows[1][5:10], "frozen less than it cost")
	assert.Equal(t, []string{"0.500000", "0.300000", "0.000000", "0.000000", "0.300000"}, rows[2][5:10], "frozen more than it cost: the rest was released, not refunded")
	assert.Equal(t, []string{"0.000000", "0.300000", "0.000000", "0.000000", "0.300000"}, rows[3][5:10], "nothing frozen")
	assert.Equal(t, "预冻结 0.100000 CNY（临时冻结，不计入消费），按实际用量扣费 0.300000 CNY", rows[1][11])
	assert.Equal(t, "未预扣，按实际用量扣费 0.300000 CNY", rows[3][11], "a request that froze nothing says so")
}

func TestBillingExcelDetailSheetTreatsALaterChargeOfATaskAsAnExtraCharge(t *testing.T) {
	statement, snapshot := frozenBillingExportFixture(t)
	snapshot.DetailSheet = true
	at := statement.StartAt + 3600
	details := []BillingExcelDetail{
		// A live extra charge precedes the imported historical charge in the ledger.
		{PostedAt: at + 300, Kind: "task_adjustment", Amount: "0.200000", RequestID: "extra", ModelName: "video-model", TaskID: "task_1"},
		{PostedAt: at, Kind: "history_usage", Amount: "1.000000", RequestID: "submit", ModelName: "video-model", TaskID: "task_1"},
	}

	body, err := RenderBillingStatementExcelWithDetails(statement, snapshot, details)
	require.NoError(t, err)
	book, err := excelize.OpenReader(bytes.NewReader(body))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("详细日志")
	require.NoError(t, err)

	require.Len(t, rows, 2)
	assert.Equal(t, []string{"1.000000", "0.200000", "0.000000", "1.200000"}, rows[1][5:9], "the second charge of a task is an extra charge, not part of the hold")
	assert.Equal(t, "submit", rows[1][9])
	assert.Equal(t, "提交时预扣 1.000000 CNY，完成后补扣 0.200000 CNY，实际扣费 1.200000 CNY", rows[1][10], "the charge in words")
}
