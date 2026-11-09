package service

import (
	"archive/zip"
	"bytes"
	"cmp"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

const billingExcelContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

// The workbook is another view of the frozen ledger snapshot. It deliberately
// excludes provider costs and inferred historical list prices; the detail sheet
// adds the usage figures of its requests, read when the workbook is built.
func RenderBillingStatementExcel(statement *model.BillingStatement, snapshot *BillingSnapshot) ([]byte, error) {
	return RenderBillingStatementExcelWithDetails(statement, snapshot, nil)
}

// BillingExcelDetail is one frozen ledger record listed in the 详细日志 sheet.
type BillingExcelDetail struct {
	PostedAt     int64
	Kind         string
	Amount       string
	RequestID    string
	ModelName    string
	TokenID      int
	TokenName    string
	BillingGroup string
	BillingRate  string
	// SourceLogID and TaskID link a request's entries; they are not listed.
	// Usage is what the usage log behind the entry recorded when the workbook
	// was made, so it is not part of the frozen ledger.
	SourceLogID int
	TaskID      string
	Usage       model.LogLink
	// EventKey finds the reservation behind a text request; ReservedAmount is
	// what was held for it, in the statement currency, empty when unknown.
	EventKey       string
	ReservedAmount string
	// Task is what the task record reports for a video request: its tokens and
	// run time, which the usage log does not carry.
	Task      model.TaskUsage
	TaskFound bool
	// Refund follows the ledger entry's own sign, the rule the day and model
	// summaries use, so the sheet sums to the same totals.
	Refund bool
}

// billingExcelDetailSheetRows keeps each sheet below Excel's 1,048,576 row limit.
const billingExcelDetailSheetRows = 1_000_000

// RenderBillingStatementExcelWithDetails renders the workbook; snapshots that
// enable DetailSheet also get the per-record 详细日志 sheets from details.
func RenderBillingStatementExcelWithDetails(statement *model.BillingStatement, snapshot *BillingSnapshot, details []BillingExcelDetail) ([]byte, error) {
	if err := statement.VerifySnapshot(); err != nil {
		return nil, err
	}
	book := excelize.NewFile()
	defer book.Close()
	if err := book.SetSheetName("Sheet1", "总览"); err != nil {
		return nil, err
	}
	for _, name := range []string{"每日汇总", "模型汇总"} {
		if _, err := book.NewSheet(name); err != nil {
			return nil, err
		}
	}
	created := time.Unix(statement.CreatedAt, 0).UTC().Format(time.RFC3339)
	if err := book.SetDocProps(&excelize.DocProperties{Creator: snapshot.Issuer, Title: "月度对账单 " + statement.Month, Created: created, Modified: created, Description: "new-api frozen billing statement"}); err != nil {
		return nil, err
	}
	header, err := book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"19335E"}}, Alignment: &excelize.Alignment{Vertical: "center", WrapText: true}})
	if err != nil {
		return nil, err
	}
	moneyFormat := "#,##0.000000;[Red](#,##0.000000);0.000000"
	if snapshot.PDFTemplateVersion >= 13 {
		symbol := "\"" + strings.ReplaceAll(snapshot.Currency.Symbol, "\"", "\\\"") + "\""
		moneyFormat = `_(` + symbol + `* #,##0.000000_);_(` + symbol + `* (#,##0.000000);_(` + symbol + `* "-"??????_);_(@_)`
	}
	if snapshot.PDFTemplateVersion >= 14 {
		moneyFormat = `_(* #,##0.000000_);_(* (#,##0.000000);_(* "-"??????_);_(@_)`
	}
	moneyStyle := &excelize.Style{CustomNumFmt: &moneyFormat}
	if snapshot.PDFTemplateVersion >= 13 {
		moneyStyle.Alignment = &excelize.Alignment{Horizontal: "right"}
	}
	money, err := book.NewStyle(moneyStyle)
	if err != nil {
		return nil, err
	}
	percent, err := book.NewStyle(&excelize.Style{NumFmt: 10})
	if err != nil {
		return nil, err
	}
	sourceID := statement.SourceStatementID
	if sourceID == "" {
		sourceID = "无（首次生成）"
	}
	rows := [][]any{
		{"月度对账单", "冻结账本金额"},
		{"企业抬头", snapshot.CompanyTitle}, {"税号", snapshot.TaxID},
		{"用户名", snapshot.Username}, {"用户编号", statement.UserID}, {"账期", statement.Month},
		{"运营主体", snapshot.OperatingName}, {"服务平台", snapshot.Issuer},
		{"币种", snapshot.Currency.Code}, {"时区", snapshot.Timezone},
		{"正式记账起点", time.Unix(snapshot.AccountingStartAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")},
		{"本期起点", time.Unix(statement.StartAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")},
		{"本期终点（不含）", time.Unix(statement.EndAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")},
		{"消费金额", billingStatementExcelAmount(snapshot.Total.Charge, snapshot)}, {"退款金额", billingStatementExcelAmount(snapshot.Total.Refund, snapshot)},
		{"本期净额", billingStatementExcelAmount(snapshot.Total.Amount, snapshot)}, {"正式账本记录数", snapshot.Total.Count},
		{"模型数量", billingModelCount(snapshot.Models)}, {"舍入差额", billingStatementExcelAmount(snapshot.RoundingDifference, snapshot)},
		{"计费费率参考时间", time.Unix(snapshot.RateReferenceAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")},
		{"计费费率说明", "生成时用户可用的模型计费组倍率，包含用户专属倍率；多组逐一列示。不是历史实际折扣，不用于倒推原价。"},
		{"原价与令牌说明", "正式账本未保存完整的历史原价、优惠、输入/输出/缓存令牌或成功请求数；未记录不代表零。"},
		{"记录数说明", "消费和退款均为正式入账笔数，不等同于成功请求数；资金充值和临时预扣不计入消费。"},
		{"对账单编号", statement.ID}, {"版本", statement.Revision}, {"来源对账单", sourceID},
	}
	if snapshot.PDFTemplateVersion < 18 {
		rows = append(rows, []any{"账单数据 SHA-256", statement.SnapshotSHA256})
	}
	rows = append(rows,
		[]any{"文件绑定校验", "PDF 列示本 Excel 的 SHA-256；校验清单同时绑定 PDF、Excel 与账单数据。客户确认该清单对应的两个文件。"},
		[]any{"金额精度", "金额保留六位小数；超过 Excel 15 位有效数字的金额以文本保存，避免精度丢失。"},
	)
	if snapshot.PDFTemplateVersion >= 10 {
		rows[19] = []any{"计费口径", "消费发生时的实际计费组与费率"}
		rows[20] = []any{"计费费率说明", "根据历史消费日志或任务提交时的计费快照列示；同一模型按实际计费组与费率拆分。未记录表示缺少可核验历史证据，不使用当前配置替代。"}
		rows[21] = []any{"令牌说明", "正式账本未完整冻结输入、输出及缓存令牌或成功请求数；本表仅列示正式入账明细汇总。"}
	}
	if snapshot.PDFTemplateVersion >= 12 {
		rows = append(rows, []any{"折扣金额说明", billingDiscountAmountNote})
	}
	for i, row := range rows {
		if err := book.SetSheetRow("总览", fmt.Sprintf("A%d", i+1), &row); err != nil {
			return nil, err
		}
	}
	if err := book.SetCellStyle("总览", "A1", "B1", header); err != nil {
		return nil, err
	}
	if err := book.SetCellStyle("总览", "B14", "B16", money); err != nil {
		return nil, err
	}
	if snapshot.PDFTemplateVersion >= 13 {
		if err := book.SetCellStyle("总览", "B19", "B19", money); err != nil {
			return nil, err
		}
	}
	wrap, err := book.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{WrapText: true, Vertical: "center"}})
	if err != nil {
		return nil, err
	}
	if err := book.SetCellStyle("总览", "B20", fmt.Sprintf("B%d", len(rows)), wrap); err != nil {
		return nil, err
	}
	if err := book.SetColWidth("总览", "A", "A", 27); err != nil {
		return nil, err
	}
	if err := book.SetColWidth("总览", "B", "B", 95); err != nil {
		return nil, err
	}
	noteRows := []int{21, 22, 23, 28, 29}
	if snapshot.PDFTemplateVersion >= 18 {
		noteRows = []int{21, 22, 23, 27, 28}
	}
	for _, row := range noteRows {
		if err := book.SetRowHeight("总览", row, 44); err != nil {
			return nil, err
		}
	}

	if snapshot.PDFTemplateVersion >= 12 {
		if err := book.SetRowHeight("总览", len(rows), 44); err != nil {
			return nil, err
		}
	}

	daily := [][]any{{"日期", "消费金额", "退款金额", "本期净额", "正式账本记录数", "记账范围"}}
	for _, row := range snapshot.Days {
		scope := "正式记账范围"
		if row.State == "outside_period" {
			scope = "正式记账起点之前"
		}
		daily = append(daily, []any{row.Label, billingStatementExcelAmount(row.Charge, snapshot), billingStatementExcelAmount(row.Refund, snapshot), billingStatementExcelAmount(row.Amount, snapshot), row.Count, scope})
	}
	daily = append(daily, []any{"合计", billingStatementExcelAmount(snapshot.Total.Charge, snapshot), billingStatementExcelAmount(snapshot.Total.Refund, snapshot), billingStatementExcelAmount(snapshot.Total.Amount, snapshot), snapshot.Total.Count, ""})
	models := [][]any{{"模型名称", "使用账单天数", "消费笔数", "退款笔数", "正式账本记录数", "消费金额", "退款金额", "本期净额", "计费费率（当前参考）", "标准原价", "消费金额占比", "计费组费率（当前参考）", "首笔入账时间", "末笔入账时间"}}
	if snapshot.PDFTemplateVersion >= 10 {
		models[0] = []any{"模型名称", "使用账单天数", "消费笔数", "退款笔数", "正式账本记录数", "消费金额", "退款金额", "本期净额", "计费费率", "消费金额占比", "实际计费组", "首笔入账时间", "末笔入账时间"}
	}
	for _, row := range snapshot.Models {
		var rate any
		groupLabel := row.BillingGroup
		if snapshot.PDFTemplateVersion >= 10 {
			rate = billingHistoricalRateLabel(row)
			if value, err := decimal.NewFromString(row.BillingRate); err == nil && !value.IsNegative() {
				rate = value.InexactFloat64()
			}
			if groupLabel == "" {
				groupLabel = "未记录"
			}
		} else {
			var groups []string
			for _, reference := range row.CurrentRates {
				value, err := decimal.NewFromString(reference.Ratio)
				if err != nil {
					return nil, err
				}
				groups = append(groups, reference.Group+": "+value.Mul(decimal.NewFromInt(100)).StringFixed(2)+"%")
			}
			groupLabel = strings.Join(groups, "; ")
			rate = billingRateLabel(row)
			if len(row.CurrentRates) == 1 {
				value, err := decimal.NewFromString(row.CurrentRates[0].Ratio)
				if err != nil {
					return nil, err
				}
				rate = value.InexactFloat64()
			}
		}
		share, err := decimal.NewFromString(strings.TrimSuffix(row.ChargeShare, "%"))
		if err != nil {
			return nil, err
		}
		name := row.ModelName
		if name == "" {
			name = "未标注模型"
		}
		values := []any{name, row.ActiveDays, row.ChargeCount, row.RefundCount, row.Count, billingStatementExcelAmount(row.Charge, snapshot), billingStatementExcelAmount(row.Refund, snapshot), billingStatementExcelAmount(row.Amount, snapshot), rate}
		if snapshot.PDFTemplateVersion < 10 {
			values = append(values, "未记录")
		}
		values = append(values, share.Div(decimal.NewFromInt(100)).InexactFloat64(), groupLabel, time.Unix(row.FirstPosted, 0).In(billingLocation).Format("2006-01-02 15:04:05"), time.Unix(row.LastPosted, 0).In(billingLocation).Format("2006-01-02 15:04:05"))
		models = append(models, values)
	}
	var chargeCount, refundCount int64
	for _, row := range snapshot.Models {
		chargeCount += row.ChargeCount
		refundCount += row.RefundCount
	}
	var shareTotal float64
	if snapshot.ChargeQuota > 0 {
		shareTotal = 1
	}
	models = append(models, []any{"合计", "—", chargeCount, refundCount, snapshot.Total.Count, billingStatementExcelAmount(snapshot.Total.Charge, snapshot), billingStatementExcelAmount(snapshot.Total.Refund, snapshot), billingStatementExcelAmount(snapshot.Total.Amount, snapshot), "—", "未记录", shareTotal})
	modelEnd, groupColumn, shareColumn := "N", "L", "K"
	if snapshot.PDFTemplateVersion >= 10 {
		models[len(models)-1] = []any{"合计", "—", chargeCount, refundCount, snapshot.Total.Count, billingStatementExcelAmount(snapshot.Total.Charge, snapshot), billingStatementExcelAmount(snapshot.Total.Refund, snapshot), billingStatementExcelAmount(snapshot.Total.Amount, snapshot), "—", shareTotal}
		modelEnd, groupColumn, shareColumn = "M", "K", "J"
	}
	if snapshot.PDFTemplateVersion >= 18 {
		daily[0][1], daily[0][2], daily[0][3] = "预扣金额", "实际消费金额", "退款金额"
		for i := 1; i < len(daily); i++ {
			daily[i][2], daily[i][3] = daily[i][3], daily[i][2]
		}
		models[0][5], models[0][6], models[0][7] = "预扣金额", "实际消费金额", "退款金额"
		for i := 1; i < len(models); i++ {
			models[i][6], models[i][7] = models[i][7], models[i][6]
		}
	}
	for _, table := range []struct {
		name                      string
		rows                      [][]any
		end, moneyStart, moneyEnd string
	}{
		{"每日汇总", daily, "F", "B", "D"}, {"模型汇总", models, modelEnd, "F", "H"},
	} {
		for i, row := range table.rows {
			if err := book.SetSheetRow(table.name, fmt.Sprintf("A%d", i+1), &row); err != nil {
				return nil, err
			}
		}
		if err := book.SetCellStyle(table.name, "A1", table.end+"1", header); err != nil {
			return nil, err
		}
		if err := book.SetCellStyle(table.name, table.moneyStart+"2", fmt.Sprintf("%s%d", table.moneyEnd, len(table.rows)), money); err != nil {
			return nil, err
		}
		if err := book.SetColWidth(table.name, "A", table.end, 20); err != nil {
			return nil, err
		}
		if err := book.SetColWidth(table.name, "A", "A", 42); err != nil {
			return nil, err
		}
		if snapshot.PDFTemplateVersion >= 13 {
			if err := book.SetColWidth(table.name, table.moneyStart, table.moneyEnd, 30); err != nil {
				return nil, err
			}
		}
		if err := book.SetRowHeight(table.name, 1, 32); err != nil {
			return nil, err
		}
		if err := book.SetPanes(table.name, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
			return nil, err
		}
	}
	for _, col := range []string{"I", shareColumn} {
		if err := book.SetCellStyle("模型汇总", col+"2", fmt.Sprintf("%s%d", col, len(models)), percent); err != nil {
			return nil, err
		}
	}
	if err := book.SetColWidth("模型汇总", groupColumn, groupColumn, 60); err != nil {
		return nil, err
	}
	if err := book.SetCellStyle("模型汇总", groupColumn+"2", fmt.Sprintf("%s%d", groupColumn, len(models)), wrap); err != nil {
		return nil, err
	}
	for i := 2; i <= len(models); i++ {
		height := 30.0
		if snapshot.PDFTemplateVersion >= 10 {
			height = 45
		}
		if i-2 < len(snapshot.Models) {
			height = max(height, float64(len(snapshot.Models[i-2].CurrentRates))*30)
		}
		if err := book.SetRowHeight("模型汇总", i, min(409, height)); err != nil {
			return nil, err
		}
	}
	if snapshot.DetailSheet {
		if err := writeBillingExcelDetailSheets(book, snapshot, details, header, money, percent, len(rows)+1); err != nil {
			return nil, err
		}
	}
	buffer, err := book.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	// Excelize iterates package maps. Canonical ZIP ordering and fixed timestamps
	// keep retries byte-identical for immutable OSS objects and PDF fingerprints.
	archive, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		return nil, err
	}
	sort.Slice(archive.File, func(i, j int) bool { return archive.File[i].Name < archive.File[j].Name })
	var result bytes.Buffer
	writer := zip.NewWriter(&result)
	for _, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			return nil, err
		}
		body, readErr := io.ReadAll(reader)
		reader.Close()
		if readErr != nil {
			return nil, readErr
		}
		entry, err := writer.CreateHeader(&zip.FileHeader{Name: file.Name, Method: zip.Deflate})
		if err != nil {
			return nil, err
		}
		if _, err := entry.Write(body); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

func billingExcelAmount(value string) any {
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return value
	}
	significant := strings.TrimLeft(strings.ReplaceAll(strings.TrimPrefix(amount.String(), "-"), ".", ""), "0")
	if len(significant) > 15 {
		return value
	}
	return amount.InexactFloat64()
}

// billingExcelRequestRow is every ledger entry of one request, kept on the
// ledger's own terms so that each column adds up to the statement's totals:
// first is the first charge (for a video task, what was held when it was
// submitted), extra the charges after it, refund what was returned, and actual
// what the request cost: first + extra - refund. The charges add up to the
// statement's consumption and the refunds to its refunds.
//
// reserved is what a text request had frozen before it was settled. A temporary
// freeze is not consumption, so it is shown apart and is in no total.
type billingExcelRequestRow struct {
	BillingExcelDetail
	first, extra, refund decimal.Decimal
	charged              bool
	reserved             decimal.Decimal
	hasReserved          bool
}

// actual is what the request cost.
func (r billingExcelRequestRow) actual() decimal.Decimal {
	return r.first.Add(r.extra).Sub(r.refund)
}

// mergeBillingExcelDetails folds the entries of one request into one row, in
// the order requests first appear. A task's refund carries its own request ID,
// so entries of one task merge by task. Entries with neither stay apart.
func mergeBillingExcelDetails(details []BillingExcelDetail) ([]billingExcelRequestRow, error) {
	// Imports append old events after live entries, so sequence order cannot
	// identify the first charge. Prefer the original charge on timestamp ties.
	sort.SliceStable(details, func(i, j int) bool {
		if details[i].PostedAt != details[j].PostedAt {
			return details[i].PostedAt < details[j].PostedAt
		}
		return details[i].Kind != "task_adjustment" && details[j].Kind == "task_adjustment"
	})
	rows := make([]billingExcelRequestRow, 0, len(details))
	byRequest := make(map[string]int, len(details))
	for _, detail := range details {
		amount, err := decimal.NewFromString(detail.Amount)
		if err != nil {
			return nil, err
		}
		key := ""
		switch {
		case detail.TaskID != "":
			key = "task:" + detail.TaskID
		case detail.RequestID != "":
			key = "request:" + detail.RequestID
		}
		index, found := byRequest[key]
		if !found || key == "" {
			index = len(rows)
			rows = append(rows, billingExcelRequestRow{BillingExcelDetail: detail})
			if key != "" {
				byRequest[key] = index
			}
		}
		row := &rows[index]
		// The first charge of a request is what it was charged up front; a later
		// one, such as the extra charge when a task costs more than was held, is an
		// extra charge. A refund carries the ledger's negative sign.
		switch {
		case detail.Refund:
			row.refund = row.refund.Sub(amount)
		case !row.charged:
			row.first, row.charged = amount, true
		default:
			row.extra = row.extra.Add(amount)
		}
		if row.BillingGroup == "" {
			row.BillingGroup = detail.BillingGroup
		}
		if !row.Usage.Found {
			row.Usage = detail.Usage
		}
		if row.ReservedAmount == "" {
			row.ReservedAmount = detail.ReservedAmount
		}
		if !row.TaskFound {
			row.Task, row.TaskFound = detail.Task, detail.TaskFound
		}
		if row.BillingRate == "" {
			row.BillingRate = detail.BillingRate
		}
	}
	for i := range rows {
		// Round the displayed components after merging the request, so their
		// arithmetic agrees with its displayed net amount.
		rows[i].first = rows[i].first.Round(6)
		rows[i].extra = rows[i].extra.Round(6)
		rows[i].refund = rows[i].refund.Round(6)
		if rows[i].BillingRate == "" {
			rows[i].BillingRate = rows[i].Usage.GroupRatio
		}
		if rows[i].ReservedAmount == "" {
			continue
		}
		reserved, err := decimal.NewFromString(rows[i].ReservedAmount)
		if err != nil {
			return nil, err
		}
		rows[i].reserved, rows[i].hasReserved = reserved, true
	}
	return rows, nil
}

// figures are the usage columns of one request: a text request's own log, or
// for a video task what the task record reports.
func (r billingExcelRequestRow) figures() billingExcelFigures {
	figures := billingExcelFigures{
		input: r.Usage.PromptTokens, output: r.Usage.CompletionTokens, cacheRead: r.Usage.CacheTokens, cacheWrite: r.Usage.CacheWrite,
		seconds: r.Usage.UseTime, ip: r.Usage.Ip, stream: r.Usage.IsStream, text: r.Usage.Found && r.TaskID == "" && !r.Usage.IsTask,
	}
	if r.TaskFound {
		if figures.output == 0 {
			figures.output = r.Task.TotalTokens
		}
		if figures.seconds == 0 {
			figures.seconds = r.Task.Seconds
		}
	}
	return figures
}

// billingExcelFigures are the usage columns of a request; text marks a text
// request, the only kind with a stream flag.
type billingExcelFigures struct {
	input, output, cacheRead, cacheWrite, seconds int
	ip                                            string
	stream, text                                  bool
}

// billingExcelColumn is one column of the detail sheet. A column with nothing
// to show in any row is left out rather than filled with zeros.
type billingExcelColumn struct {
	title string
	width float64
	show  bool
	value func(r billingExcelRequestRow) any
}

func writeBillingExcelDetailSheets(book *excelize.File, snapshot *BillingSnapshot, details []BillingExcelDetail, header, money, percent, overviewNextRow int) error {
	requests, err := mergeBillingExcelDetails(details)
	if err != nil {
		return err
	}
	var charge, refund decimal.Decimal
	for _, request := range requests {
		charge = charge.Add(request.first).Add(request.extra)
		refund = refund.Add(request.refund)
	}
	// The overview rounds the month's integer quota once. Disclose the
	// difference from rounded request rows without inventing charge records.
	var roundingRows [][]any
	for _, total := range []struct {
		label, amount string
		detail        decimal.Decimal
	}{
		{"详细日志消费舍入差额", snapshot.Total.Charge, charge},
		{"详细日志退款舍入差额", snapshot.Total.Refund, refund},
		{"详细日志净额舍入差额", snapshot.Total.Amount, charge.Sub(refund)},
	} {
		value, err := decimal.NewFromString(total.amount)
		if err != nil {
			return err
		}
		difference := value.Sub(total.detail)
		if !difference.IsZero() {
			roundingRows = append(roundingRows, []any{total.label, billingStatementExcelAmount(difference.StringFixed(6), snapshot)})
		}
	}
	if len(roundingRows) > 0 {
		roundingRows = append(roundingRows, []any{"详细日志舍入说明", "明细合计加对应舍入差额等于总览；差额不改变账本或扣费。"})
		for i, row := range roundingRows {
			if err := book.SetSheetRow("总览", fmt.Sprintf("A%d", overviewNextRow+i), &row); err != nil {
				return err
			}
			if i < len(roundingRows)-1 {
				cell := fmt.Sprintf("B%d", overviewNextRow+i)
				if err := book.SetCellStyle("总览", cell, cell, money); err != nil {
					return err
				}
			}
		}
	}
	var hasStream, hasInput, hasOutput, hasCacheRead, hasCacheWrite, hasSeconds, hasIP, hasRate, hasReserved bool
	for _, request := range requests {
		hasReserved = hasReserved || request.hasReserved
		figures := request.figures()
		hasStream = hasStream || figures.text
		hasInput, hasOutput = hasInput || figures.input > 0, hasOutput || figures.output > 0
		hasCacheRead, hasCacheWrite = hasCacheRead || figures.cacheRead > 0, hasCacheWrite || figures.cacheWrite > 0
		hasSeconds, hasIP = hasSeconds || figures.seconds > 0, hasIP || figures.ip != ""
		if value, err := decimal.NewFromString(request.BillingRate); err == nil && !value.IsNegative() {
			hasRate = true
		}
	}
	// Zero shows as a dash, as in the amount columns.
	countFormat := `#,##0;-#,##0;"-"`
	count, err := book.NewStyle(&excelize.Style{CustomNumFmt: &countFormat})
	if err != nil {
		return err
	}
	number := func(value int) any { return excelize.Cell{StyleID: count, Value: value} }
	amount := func(value decimal.Decimal) any {
		return excelize.Cell{StyleID: money, Value: billingStatementExcelAmount(value.StringFixed(6), snapshot)}
	}
	all := []billingExcelColumn{
		{"时间", 21, true, func(r billingExcelRequestRow) any {
			return time.Unix(r.PostedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")
		}},
		{"用户", 24, true, func(billingExcelRequestRow) any { return snapshot.Username }},
		{"令牌", 28, true, func(r billingExcelRequestRow) any {
			switch {
			case r.TokenID == 0:
				return "未记录"
			case r.TokenName == "":
				return fmt.Sprintf("#%d", r.TokenID)
			}
			return r.TokenName
		}},
		{"所选分组", 14, true, func(r billingExcelRequestRow) any {
			if group := cmp.Or(r.BillingGroup, r.Usage.Group); group != "" {
				return group
			}
			return "未记录"
		}},
		{"模型", 36, true, func(r billingExcelRequestRow) any {
			if r.ModelName == "" {
				return "未标注模型"
			}
			return r.ModelName
		}},
		{"流", 6, hasStream, func(r billingExcelRequestRow) any {
			figures := r.figures()
			switch {
			case !figures.text:
				return ""
			case figures.stream:
				return "是"
			}
			return "否"
		}},
		{"输入 Token", 13, hasInput, func(r billingExcelRequestRow) any { return number(r.figures().input) }},
		{"输出 Token", 13, hasOutput, func(r billingExcelRequestRow) any { return number(r.figures().output) }},
		{"缓存读取", 13, hasCacheRead, func(r billingExcelRequestRow) any { return number(r.figures().cacheRead) }},
		{"缓存写入", 13, hasCacheWrite, func(r billingExcelRequestRow) any { return number(r.figures().cacheWrite) }},
		{"计费费率", 14, hasRate, func(r billingExcelRequestRow) any {
			if value, err := decimal.NewFromString(r.BillingRate); err == nil && !value.IsNegative() {
				return excelize.Cell{StyleID: percent, Value: value.InexactFloat64()}
			}
			return "未记录"
		}},
		// 首次扣费 + 补扣金额 - 退款金额 = 实际扣费金额. The first two add up to the
		// statement's consumption, the refunds to its refunds and the last to its
		// net amount. 预冻结金额 is a text request's temporary freeze, in no total.
		{"预冻结金额", 22, hasReserved, func(r billingExcelRequestRow) any {
			if !r.hasReserved {
				return ""
			}
			return amount(r.reserved)
		}},
		{"首次扣费", 22, true, func(r billingExcelRequestRow) any { return amount(r.first) }},
		{"补扣金额", 22, true, func(r billingExcelRequestRow) any { return amount(r.extra) }},
		{"退款金额", 22, true, func(r billingExcelRequestRow) any { return amount(r.refund) }},
		{"实际扣费金额", 22, true, func(r billingExcelRequestRow) any { return amount(r.actual()) }},
		{"耗时(秒)", 10, hasSeconds, func(r billingExcelRequestRow) any { return number(r.figures().seconds) }},
		{"IP", 16, hasIP, func(r billingExcelRequestRow) any { return r.figures().ip }},
		{"请求ID", 44, true, func(r billingExcelRequestRow) any { return r.RequestID }},
		{"扣费说明", 60, true, func(r billingExcelRequestRow) any { return r.explanation(snapshot.Currency.Code) }},
	}
	columns := make([]billingExcelColumn, 0, len(all))
	for _, column := range all {
		if column.show {
			columns = append(columns, column)
		}
	}
	lastColumn, err := excelize.ColumnNumberToName(len(columns))
	if err != nil {
		return err
	}
	sheetCount := max(1, (len(requests)+billingExcelDetailSheetRows-1)/billingExcelDetailSheetRows)
	for sheet := 0; sheet < sheetCount; sheet++ {
		name := "详细日志"
		if sheet > 0 {
			name = fmt.Sprintf("详细日志%d", sheet+1)
		}
		if _, err := book.NewSheet(name); err != nil {
			return err
		}
		writer, err := book.NewStreamWriter(name)
		if err != nil {
			return err
		}
		if err := writer.SetPanes(&excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
			return err
		}
		headerRow := make([]any, len(columns))
		for i, column := range columns {
			if err := writer.SetColWidth(i+1, i+1, column.width); err != nil {
				return err
			}
			headerRow[i] = excelize.Cell{StyleID: header, Value: column.title}
		}
		if err := writer.SetRow("A1", headerRow, excelize.RowOpts{Height: 32}); err != nil {
			return err
		}
		from := sheet * billingExcelDetailSheetRows
		to := min(len(requests), from+billingExcelDetailSheetRows)
		for i, request := range requests[from:to] {
			row := make([]any, len(columns))
			for j, column := range columns {
				row[j] = column.value(request)
			}
			cell, err := excelize.CoordinatesToCellName(1, i+2)
			if err != nil {
				return err
			}
			if err := writer.SetRow(cell, row); err != nil {
				return err
			}
		}
		if err := writer.Flush(); err != nil {
			return err
		}
		if err := book.AutoFilter(name, fmt.Sprintf("A1:%s%d", lastColumn, to-from+1), nil); err != nil {
			return err
		}
	}
	return nil
}

// explanation says in words how the request's charge came about.
func (r billingExcelRequestRow) explanation(currency string) string {
	text := func(value decimal.Decimal) string { return value.StringFixed(6) + " " + currency }
	switch {
	case !r.charged && r.refund.IsPositive():
		return "本账期内无对应扣费，退回 " + text(r.refund)
	case r.extra.IsZero() && r.refund.IsZero():
		switch {
		case r.hasReserved && r.reserved.IsZero():
			return "未预扣，按实际用量扣费 " + text(r.actual())
		case r.hasReserved:
			return "预冻结 " + text(r.reserved) + "（临时冻结，不计入消费），按实际用量扣费 " + text(r.actual())
		}
		return "按 " + text(r.actual()) + " 一次计费"
	}
	first := "首次扣费 "
	if r.TaskID != "" || r.Usage.IsTask {
		first = "提交时预扣 "
	}
	parts := []string{first + text(r.first)}
	if r.extra.IsPositive() {
		parts = append(parts, "完成后补扣 "+text(r.extra))
	}
	if r.refund.IsPositive() {
		parts = append(parts, "完成后退回 "+text(r.refund))
	}
	return strings.Join(append(parts, "实际扣费 "+text(r.actual())), "，")
}

// linkBillingExcelDetails finds each entry's task and, when the ledger did not
// record one, its group from the usage log behind it. A task refund's request
// ID is the task ID itself.
func linkBillingExcelDetails(userID int, details []BillingExcelDetail) error {
	var logIDs []int
	var requestIDs []string
	for _, detail := range details {
		switch {
		case detail.Kind == "task_adjustment":
		case detail.SourceLogID > 0 && !common.UsingLogDatabase(common.DatabaseTypeClickHouse):
			logIDs = append(logIDs, detail.SourceLogID)
		case detail.RequestID != "":
			requestIDs = append(requestIDs, detail.RequestID)
		}
	}
	byID, err := model.GetLogLinksByIDs(userID, logIDs)
	if err != nil {
		return err
	}
	byRequest, err := model.GetLogLinksByRequestIDs(userID, requestIDs)
	if err != nil {
		return err
	}
	for i := range details {
		detail := &details[i]
		var link model.LogLink
		switch {
		case detail.Kind == "task_adjustment":
			detail.TaskID = detail.RequestID
			continue
		case detail.SourceLogID > 0 && !common.UsingLogDatabase(common.DatabaseTypeClickHouse):
			link = byID[detail.SourceLogID]
		default:
			link = byRequest[detail.RequestID]
		}
		detail.TaskID, detail.Usage = link.TaskID, link
		if detail.BillingGroup == "" {
			detail.BillingGroup = link.Group
		}
	}

	// The older video logs recorded no task ID, so their refund, which does,
	// would stay a row of its own. A task is submitted when it is charged: find
	// it by that time.
	var chargeTimes []int64
	for _, detail := range details {
		if detail.TaskID == "" && detail.Usage.IsTask {
			chargeTimes = append(chargeTimes, detail.Usage.CreatedAt)
		}
	}
	taskIDs, err := model.GetTaskIDsBySubmitTimes(userID, chargeTimes)
	if err != nil {
		return err
	}
	for i := range details {
		if details[i].TaskID == "" && details[i].Usage.IsTask {
			details[i].TaskID = taskIDs[details[i].Usage.CreatedAt]
		}
	}
	return nil
}

// attachBillingExcelReservations records what each text request held up front,
// from the reservation its settlement entry belongs to.
func attachBillingExcelReservations(userID int, details []BillingExcelDetail, rate, unit decimal.Decimal) error {
	eventKeys := make([]string, 0, len(details))
	for _, detail := range details {
		eventKeys = append(eventKeys, detail.EventKey)
	}
	operations, err := model.GetBillingOperationsByEventKeys(userID, eventKeys)
	if err != nil {
		return err
	}
	for i := range details {
		if operation, found := operations[details[i].EventKey]; found {
			details[i].ReservedAmount = decimal.NewFromInt(operation.Reserved).Mul(rate).Div(unit).StringFixed(6)
		}
	}
	return nil
}

// enrichBillingExcelDetails adds to the frozen entries what the ledger does not
// hold: token names, the usage log and task behind each entry, and the
// reservation a text request settled. The links come first, since the task
// usage is looked up by the task they find.
func enrichBillingExcelDetails(userID int, details []BillingExcelDetail, rate, unit decimal.Decimal) error {
	tokenIDs := lo.Uniq(lo.FilterMap(details, func(detail BillingExcelDetail, _ int) (int, bool) { return detail.TokenID, detail.TokenID > 0 }))
	tokenNames, err := model.GetTokenNamesByIDs(tokenIDs)
	if err != nil {
		return err
	}
	for i := range details {
		details[i].TokenName = tokenNames[details[i].TokenID]
	}
	if err := linkBillingExcelDetails(userID, details); err != nil {
		return err
	}
	if err := attachBillingExcelTaskUsage(userID, details); err != nil {
		return err
	}
	return attachBillingExcelReservations(userID, details, rate, unit)
}

// attachBillingExcelTaskUsage adds what the task records report, for the video
// requests whose usage logs carry no tokens or run time.
func attachBillingExcelTaskUsage(userID int, details []BillingExcelDetail) error {
	taskIDs := lo.Uniq(lo.FilterMap(details, func(detail BillingExcelDetail, _ int) (string, bool) { return detail.TaskID, detail.TaskID != "" }))
	usage, err := model.GetTaskUsageByTaskIDs(userID, taskIDs)
	if err != nil {
		return err
	}
	for i := range details {
		details[i].Task, details[i].TaskFound = usage[details[i].TaskID]
	}
	return nil
}
