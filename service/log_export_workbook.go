package service

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

// LogExportTypeNames are the log types a log export may list, with their names
// in the file: what was consumed and what was refunded, the same two a
// statement's detail sheet carries.
var LogExportTypeNames = map[int]string{
	model.LogTypeConsume: "消耗",
	model.LogTypeRefund:  "退款",
}

// logExportRow is one exported log, reduced to what the sheet shows so that a
// large export does not keep whole logs in memory.
type logExportRow struct {
	createdAt                                                         int64
	userID, logType                                                   int
	username, token, group, modelName, ip, requestID, content, taskID string
	stream, isTask                                                    bool
	prompt, completion, cacheRead, cacheWrite, quota, seconds         int
	// extraQuota and refundQuota are what the later logs of the same task
	// changed: an extra charge when it cost more than was held, a refund when it
	// cost less.
	extraQuota, refundQuota int
	// merged marks a log folded into the first consumption row of its task.
	merged bool
}

// LogExportWorkbook writes usage logs as a spreadsheet with readable column
// widths, real numbers and dates, so the file opens ready to use. It never
// lists the platform's cost, revenue, profit or supplier channels. A video
// task leaves several logs: the charge at submission, then an extra charge or a
// refund when the real cost is known. They are folded into one row: what was
// charged, the settlement adjustment, and what the task cost.
type LogExportWorkbook struct {
	book                 *excelize.File
	withUser             bool
	rate, unit           decimal.Decimal
	symbol, currencyCode string
	rows                 []logExportRow
}

// NewLogExportWorkbook starts a workbook; withUser adds the user column that
// only administrators see.
func NewLogExportWorkbook(withUser bool) (*LogExportWorkbook, error) {
	currency, err := currentBillingCurrency(false)
	if err != nil {
		return nil, err
	}
	rate, err := decimal.NewFromString(currency.Rate)
	if err != nil {
		return nil, err
	}
	unit, err := decimal.NewFromString(currency.QuotaPerUnit)
	if err != nil {
		return nil, err
	}
	return &LogExportWorkbook{book: excelize.NewFile(), withUser: withUser, rate: rate, unit: unit, symbol: currency.Symbol, currencyCode: currency.Code}, nil
}

// Add records one log; the sheet is written by Write.
func (w *LogExportWorkbook) Add(log *model.Log) {
	link := model.LogLinkOf(log)
	w.rows = append(w.rows, logExportRow{
		createdAt: log.CreatedAt, userID: log.UserId, logType: log.Type, isTask: link.IsTask, username: log.Username, token: log.TokenName, group: log.Group,
		modelName: log.ModelName, ip: log.Ip, requestID: log.RequestId, content: log.Content, taskID: link.TaskID,
		stream: log.IsStream, prompt: log.PromptTokens, completion: log.CompletionTokens, cacheRead: link.CacheTokens,
		cacheWrite: link.CacheWrite, quota: log.Quota, seconds: log.UseTime,
	})
}

// foldTaskAdjustments folds the later logs of a task into its first
// consumption log and reports whether any was, in which case the sheet gets the
// adjustment columns. A refund with no consumption log in the export stays a row
// of its own.
func (w *LogExportWorkbook) foldTaskAdjustments() (bool, error) {
	// The older video logs recorded no task ID, so the logs that follow them could
	// not be folded in. A task is submitted when it is charged: find it by that time.
	legacy := make(map[int][]int)
	for i, row := range w.rows {
		if row.logType == model.LogTypeConsume && row.isTask && row.taskID == "" {
			legacy[row.userID] = append(legacy[row.userID], i)
		}
	}
	for userID, rows := range legacy {
		taskIDs, err := model.GetTaskIDsBySubmitTimes(userID, lo.Map(rows, func(i int, _ int) int64 { return w.rows[i].createdAt }))
		if err != nil {
			return false, err
		}
		for _, i := range rows {
			w.rows[i].taskID = taskIDs[w.rows[i].createdAt]
		}
	}

	first := make(map[string]int)
	for i, row := range w.rows {
		if row.logType != model.LogTypeConsume || row.taskID == "" {
			continue
		}
		if base, found := first[row.taskID]; !found || row.createdAt < w.rows[base].createdAt {
			first[row.taskID] = i
		}
	}
	folded := false
	for i := range w.rows {
		row := &w.rows[i]
		base, found := first[row.taskID]
		if !found || i == base || (row.logType != model.LogTypeConsume && row.logType != model.LogTypeRefund) {
			continue
		}
		if row.logType == model.LogTypeRefund {
			w.rows[base].refundQuota += row.quota
		} else {
			w.rows[base].extraQuota += row.quota
		}
		row.merged, folded = true, true
	}
	return folded, nil
}

// details is the detail text of a row: the log's own text, or for a task whose
// later logs were folded in, how its charge came about.
func (w *LogExportWorkbook) details(row logExportRow) string {
	if row.extraQuota == 0 && row.refundQuota == 0 {
		return row.content
	}
	text := func(quota int) string { return w.symbol + decimal.NewFromFloat(w.amount(quota)).StringFixed(6) }
	parts := []string{"提交时扣 " + text(row.quota)}
	if row.extraQuota > 0 {
		parts = append(parts, "完成后补扣 "+text(row.extraQuota))
	}
	if row.refundQuota > 0 {
		parts = append(parts, "完成后退回 "+text(row.refundQuota))
	}
	return strings.Join(append(parts, "实际 "+text(row.quota+row.extraQuota-row.refundQuota)), "，")
}

func (w *LogExportWorkbook) amount(quota int) float64 {
	return decimal.NewFromInt(int64(quota)).Mul(w.rate).Div(w.unit).InexactFloat64()
}

// Write finishes the workbook and returns the file.
func (w *LogExportWorkbook) Write(startTimestamp, endTimestamp int64) (*bytes.Buffer, error) {
	withAdjustments, err := w.foldTaskAdjustments()
	if err != nil {
		return nil, err
	}
	// A refund whose charge is not in the export is a row of its own, and its
	// amount shows as an adjustment, so the adjustment columns serve it too.
	logCount := len(w.rows)
	var charge, refunds decimal.Decimal
	for _, row := range w.rows {
		withAdjustments = withAdjustments || row.logType == model.LogTypeRefund
		switch row.logType {
		case model.LogTypeConsume:
			charge = charge.Add(decimal.NewFromInt(int64(row.quota)))
		case model.LogTypeRefund:
			refunds = refunds.Add(decimal.NewFromInt(int64(row.quota)))
		}
	}
	if err := w.book.SetSheetName("Sheet1", "汇总"); err != nil {
		return nil, err
	}
	if _, err := w.book.NewSheet("使用日志"); err != nil {
		return nil, err
	}
	writer, err := w.book.NewStreamWriter("使用日志")
	if err != nil {
		return nil, err
	}
	header, err := w.book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"19335E"}}, Alignment: &excelize.Alignment{Vertical: "center", WrapText: true}})
	if err != nil {
		return nil, err
	}
	timeFormat := "yyyy-mm-dd hh:mm:ss"
	timeStyle, err := w.book.NewStyle(&excelize.Style{CustomNumFmt: &timeFormat, Alignment: &excelize.Alignment{Horizontal: "left"}})
	if err != nil {
		return nil, err
	}
	countFormat := `#,##0;-#,##0;"-"`
	count, err := w.book.NewStyle(&excelize.Style{CustomNumFmt: &countFormat})
	if err != nil {
		return nil, err
	}
	moneyFormat := `#,##0.000000`
	if w.symbol != "TOKENS" {
		moneyFormat = `"` + strings.ReplaceAll(w.symbol, `"`, `\"`) + `"#,##0.000000`
	}
	// A refund shows in parentheses and zero as a dash, as in the statement.
	signedFormat := moneyFormat + `;(` + moneyFormat + `);"-"`
	signedMoney, err := w.book.NewStyle(&excelize.Style{CustomNumFmt: &signedFormat})
	if err != nil {
		return nil, err
	}

	_, offset := time.Now().In(billingLocation).Zone()
	titles := []any{fmt.Sprintf("时间(UTC%+d)", offset/3600), "类型"}
	widths := []float64{21, 8}
	if w.withUser {
		titles, widths = append(titles, "用户"), append(widths, 18)
	}
	titles = append(titles, "令牌", "分组", "模型", "流", "输入 Token", "输出 Token", "缓存读取", "缓存写入")
	widths = append(widths, 22, 14, 30, 6, 13, 13, 13, 13)
	// 首次扣费 + 补扣金额 - 退款金额 = 实际扣费金额, as in the statement's detail
	// sheet; without a task's later logs the one amount is what was charged.
	if withAdjustments {
		titles, widths = append(titles, "首次扣费", "补扣金额", "退款金额", "实际扣费金额"), append(widths, 16, 16, 16, 16)
	} else {
		titles, widths = append(titles, "实际扣费金额"), append(widths, 16)
	}
	titles = append(titles, "耗时(秒)", "IP", "请求 ID", "详情")
	widths = append(widths, 10, 16, 44, 60)
	for i, width := range widths {
		if err := writer.SetColWidth(i+1, i+1, width); err != nil {
			return nil, err
		}
	}
	if err := writer.SetPanes(&excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
		return nil, err
	}
	headerRow := make([]any, len(titles))
	for i, title := range titles {
		headerRow[i] = excelize.Cell{StyleID: header, Value: title}
	}
	if err := writer.SetRow("A1", headerRow, excelize.RowOpts{Height: 28}); err != nil {
		return nil, err
	}

	signed := func(value float64) any { return excelize.Cell{StyleID: signedMoney, Value: value} }
	number := func(value int) any { return excelize.Cell{StyleID: count, Value: value} }
	zero := signed(0)
	next := 2
	for _, row := range w.rows {
		if row.merged {
			continue
		}
		// A spreadsheet date has no zone, so it carries the wall clock of the
		// statement time zone.
		local := time.Unix(row.createdAt, 0).In(billingLocation)
		wallClock := time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), local.Second(), 0, time.UTC)
		cells := []any{excelize.Cell{StyleID: timeStyle, Value: wallClock}, LogExportTypeNames[row.logType]}
		if w.withUser {
			cells = append(cells, row.username)
		}
		// Usage and amounts belong to a consumption log; a refund on its own row has
		// only its amount: no charge, a refund, and a negative net.
		var stream, prompt, completion, cacheRead, cacheWrite, seconds, first, extra, refund, actual any = "", "", "", "", "", "", "", "", "", ""
		switch row.logType {
		case model.LogTypeConsume:
			stream = "否"
			if row.stream {
				stream = "是"
			}
			prompt, completion, cacheRead, cacheWrite, seconds = number(row.prompt), number(row.completion), number(row.cacheRead), number(row.cacheWrite), number(row.seconds)
			first = signed(w.amount(row.quota))
			actual, extra, refund = first, zero, zero
			if row.extraQuota != 0 || row.refundQuota != 0 {
				extra, refund = signed(w.amount(row.extraQuota)), signed(w.amount(row.refundQuota))
				actual = signed(w.amount(row.quota + row.extraQuota - row.refundQuota))
			}
		case model.LogTypeRefund:
			refund, actual = signed(w.amount(row.quota)), signed(-w.amount(row.quota))
		}
		cells = append(cells, row.token, row.group, row.modelName, stream, prompt, completion, cacheRead, cacheWrite)
		if withAdjustments {
			cells = append(cells, first, extra, refund)
		}
		cells = append(cells, actual, seconds, row.ip, row.requestID, w.details(row))
		cell, err := excelize.CoordinatesToCellName(1, next)
		if err != nil {
			return nil, err
		}
		next++
		if err := writer.SetRow(cell, cells); err != nil {
			return nil, err
		}
	}
	w.rows = nil
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	last, err := excelize.ColumnNumberToName(len(titles))
	if err != nil {
		return nil, err
	}
	if err := w.book.AutoFilter("使用日志", fmt.Sprintf("A1:%s%d", last, max(next-1, 1)), nil); err != nil {
		return nil, err
	}
	// Count each source log once, including logs folded into a task's row.
	// Convert only after summing quota so the total is not a sum of rounded fees.
	amount := func(quota decimal.Decimal) any {
		return billingExcelAmount(quota.Mul(w.rate).Div(w.unit).StringFixed(6))
	}
	wallTimes := make([]time.Time, 2)
	for i, timestamp := range []int64{startTimestamp, endTimestamp} {
		local := time.Unix(timestamp, 0).In(billingLocation)
		wallTimes[i] = time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), local.Second(), 0, time.UTC)
	}
	summary := [][]any{
		{"本次导出汇总"},
		{},
		{fmt.Sprintf("开始时间(UTC%+d)", offset/3600), wallTimes[0]},
		{fmt.Sprintf("结束时间(UTC%+d)", offset/3600), wallTimes[1]},
		{"币种", w.currencyCode},
		{},
		{"消费金额", amount(charge)},
		{"退款金额", amount(refunds)},
		{"净消费金额", amount(charge.Sub(refunds))},
		{},
		{"原始日志条数", logCount},
		{"导出明细行数", next - 2},
		{},
		{"统计范围", "仅统计本次筛选导出的消费和退款。"},
	}
	for i, row := range summary {
		if err := w.book.SetSheetRow("汇总", fmt.Sprintf("A%d", i+1), &row); err != nil {
			return nil, err
		}
		if err := w.book.SetRowHeight("汇总", i+1, 22); err != nil {
			return nil, err
		}
	}
	if err := w.book.SetColWidth("汇总", "A", "A", 24); err != nil {
		return nil, err
	}
	if err := w.book.SetColWidth("汇总", "B", "B", 62); err != nil {
		return nil, err
	}
	title, err := w.book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14, Color: "19335E"}})
	if err != nil {
		return nil, err
	}
	for _, style := range []struct {
		from, to string
		id       int
	}{{"A1", "B1", title}, {"B3", "B4", timeStyle}, {"B7", "B9", signedMoney}, {"B11", "B12", count}} {
		if err := w.book.SetCellStyle("汇总", style.from, style.to, style.id); err != nil {
			return nil, err
		}
	}
	showGridLines := false
	if err := w.book.SetSheetView("汇总", 0, &excelize.ViewOptions{ShowGridLines: &showGridLines}); err != nil {
		return nil, err
	}
	return w.book.WriteToBuffer()
}

// Close releases the workbook's temporary files.
func (w *LogExportWorkbook) Close() {
	_ = w.book.Close()
}
