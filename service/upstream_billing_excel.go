package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/xuri/excelize/v2"
)

// Each sheet presents the frozen evidence. Names are stored as text, never
// formulas, and exact monetary precision follows the customer export policy.
func RenderUpstreamBillingStatementExcel(statement *model.UpstreamBillingStatement, snapshot *UpstreamBillingSnapshot) ([]byte, error) {
	if snapshot == nil {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	if err := statement.VerifySnapshot(); err != nil {
		return nil, err
	}
	frozen, err := common.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(frozen, []byte(statement.Snapshot)) || (snapshot.SchemaVersion != 1 && snapshot.SchemaVersion != 2) {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	book := excelize.NewFile()
	defer book.Close()
	if err := book.SetSheetName("Sheet1", "总览"); err != nil {
		return nil, err
	}
	for _, name := range []string{"每日汇总", "模型汇总", "每日模型汇总", "渠道范围"} {
		if _, err := book.NewSheet(name); err != nil {
			return nil, err
		}
	}
	created := time.Unix(statement.CreatedAt, 0).UTC().Format(time.RFC3339)
	if err := book.SetDocProps(&excelize.DocProperties{Creator: snapshot.Issuer, Title: "上游月度对账单 " + statement.Month, Created: created, Modified: created, Description: "new-api frozen upstream cost statement"}); err != nil {
		return nil, err
	}
	header, err := book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"19335E"}}, Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true}})
	if err != nil {
		return nil, err
	}
	moneyFormat := `_(* #,##0.000000_);_(* (#,##0.000000);_(* "-"??????_);_(@_)`
	money, err := book.NewStyle(&excelize.Style{CustomNumFmt: &moneyFormat, Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"}})
	if err != nil {
		return nil, err
	}
	countFormat := "#,##0"
	count, err := book.NewStyle(&excelize.Style{CustomNumFmt: &countFormat, Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"}})
	if err != nil {
		return nil, err
	}
	wrap, err := book.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{WrapText: true, Vertical: "center"}})
	if err != nil {
		return nil, err
	}
	warningStyle, err := book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "9A5800"}, Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FFF4DA"}}, Alignment: &excelize.Alignment{WrapText: true, Vertical: "center"}})
	if err != nil {
		return nil, err
	}
	dateFormat := "yyyy-mm-dd"
	dateStyle, err := book.NewStyle(&excelize.Style{CustomNumFmt: &dateFormat, Alignment: &excelize.Alignment{Vertical: "center"}})
	if err != nil {
		return nil, err
	}
	rows := [][]any{
		{"上游月度对账单", "冻结成本与用量"}, {"上游", snapshot.Group.Name}, {"上游编号", statement.GroupID}, {"渠道归属版本", snapshot.Group.Version}, {"账期", statement.Month},
		{"服务平台", snapshot.Issuer}, {"运营主体", snapshot.OperatingName}, {"币种", snapshot.Currency.Code}, {"时区", snapshot.Timezone},
		{"本期起点", time.Unix(statement.StartAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")}, {"本期终点（不含）", time.Unix(statement.EndAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")},
		{"成本消费金额", billingExcelAmount(upstreamBillingExportTotalAmount(snapshot.Total.Charge, snapshot))}, {"成本退款金额", billingExcelAmount(upstreamBillingExportTotalAmount(snapshot.Total.Refund, snapshot))}, {upstreamBillingExportNetLabel(snapshot), billingExcelAmount(upstreamBillingExportTotalAmount(snapshot.Total.Amount, snapshot))},
		{"成本记录数", snapshot.Total.Count}, {"请求数", snapshot.Total.Requests}, {"输入令牌", snapshot.Total.InputTokens}, {"输出令牌", snapshot.Total.OutputTokens}, {"缓存读取令牌", snapshot.Total.CacheReadTokens}, {"缓存写入令牌", snapshot.Total.CacheWriteTokens},
		{"缺少成本记录", snapshot.Coverage.MissingCostRecords}, {"缺少用量记录", snapshot.Coverage.MissingUsageRecords}, {"估算用量记录", snapshot.Coverage.EstimatedUsageRecords}, {"待结算任务", snapshot.Coverage.PendingTasks},
		{"舍入差额", billingExcelAmount(snapshot.RoundingDifference)}, {"对账单编号", statement.ID}, {"版本", statement.Revision},
		{"成本口径", "消费发生时冻结的历史成本快照与已审计调整；净成本 = 成本消费金额 - 成本退款金额。未配置成本的渠道排除，合法零成本仍纳入。"},
		{"用量口径", "消费与退款按发生时间记账。仅消费记录计入请求与用量；退款令牌不重复扣减。未记录用量不表示零，估算及缺失数量单独列示。"},
		{"金额精度", "金额保留六位小数；超过 Excel 15 位有效数字的金额及用量以文本保存，避免精度丢失。"},
		{"数据快照 SHA-256", statement.SnapshotSHA256}, {"使用明细 SHA-256", statement.SourceSHA256},
		{"文件绑定校验", "PDF 列示本 Excel SHA-256；校验清单绑定 PDF、Excel、数据快照与分块使用明细。财务管理员确认该校验清单对应的文件。"},
		{"数据截至", time.Unix(snapshot.UpdatedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05") + " Asia/Shanghai"},
		{"数据质量", upstreamBillingExportQualityWarning(snapshot)},
		{"金额范围", upstreamBillingExportCostScope(snapshot.Total) + "。缺少成本时，仅列示可确认的金额；未知金额不作为零成本。"},
		{"对账说明", "数据质量告警不限制生成、审核或归档；请结合逐笔缺口及上游附件核对。本单保留生成时的告警和统计截止，确认不代表缺口已恢复。"},
	}
	if !snapshot.Coverage.Complete {
		rows[0][1] = "数据质量告警：用于核对，缺口待核实"
	}
	if !upstreamBillingExportHasEligibleChannel(snapshot) {
		rows[35][1] = "无可纳入成本统计的渠道。当前范围不能提供完整的上游成本金额。"
	} else if snapshot.Coverage.MissingChannels > 0 {
		rows[35][1] = "已知成本小计。缺失渠道未纳入完整统计，未知金额不作为零成本。"
	}
	for i, row := range rows {
		if value, ok := row[1].(int64); ok && len(strings.TrimPrefix(fmt.Sprint(value), "-")) > 15 {
			row[1] = fmt.Sprint(value)
		}
		if err := book.SetSheetRow("总览", fmt.Sprintf("A%d", i+1), &row); err != nil {
			return nil, err
		}
	}
	if err := book.SetCellStyle("总览", "A1", "B1", header); err != nil {
		return nil, err
	}
	if err := book.SetRowHeight("总览", 1, 30); err != nil {
		return nil, err
	}
	if err := book.SetCellStyle("总览", "B12", "B14", money); err != nil {
		return nil, err
	}
	if err := book.SetCellStyle("总览", "B25", "B25", money); err != nil {
		return nil, err
	}
	if err := book.SetCellStyle("总览", "B28", fmt.Sprintf("B%d", len(rows)), wrap); err != nil {
		return nil, err
	}
	if !snapshot.Coverage.Complete {
		if err := book.SetCellStyle("总览", "A35", "B35", warningStyle); err != nil {
			return nil, err
		}
	}
	if err := book.SetColWidth("总览", "A", "A", 27); err != nil {
		return nil, err
	}
	if err := book.SetColWidth("总览", "B", "B", 93); err != nil {
		return nil, err
	}
	for _, row := range []int{28, 29, 30, 33, 35, 36, 37} {
		if err := book.SetRowHeight("总览", row, 44); err != nil {
			return nil, err
		}
	}
	usageHeaders := []any{"成本记录数", "消费记录数", "退款记录数", "请求数", "成本消费金额", "成本退款金额", "净成本", "输入令牌", "输出令牌", "缓存读取令牌", "缓存写入令牌", "缓存写入 5 分钟令牌", "缓存写入 1 小时令牌", "视频令牌", "视频秒", "缺少成本记录", "缺少用量记录", "估算用量记录", "待结算任务", "金额范围", "用量状态"}
	daily := [][]any{append([]any{"日期"}, usageHeaders...)}
	for _, row := range snapshot.Days {
		day, err := time.ParseInLocation("2006-01-02", row.Label, billingLocation)
		if err != nil {
			return nil, err
		}
		daily = append(daily, append([]any{day}, upstreamExcelRowValues(row, upstreamBillingExportHasEligibleChannel(snapshot))...))
	}
	daily = append(daily, append([]any{"合计"}, upstreamExcelRowValues(snapshot.Total, upstreamBillingExportHasEligibleChannel(snapshot))...))
	models := [][]any{append([]any{"模型名称", "上游模型名称", "计费组", "使用账单天数", "首笔时间", "末笔时间"}, usageHeaders...)}
	for _, row := range snapshot.Models {
		models = append(models, append([]any{row.ModelName, row.UpstreamModelName, row.BillingGroup, row.ActiveDays, time.Unix(row.FirstPosted, 0).In(billingLocation), time.Unix(row.LastPosted, 0).In(billingLocation)}, upstreamExcelRowValues(row.UpstreamBillingRow, upstreamBillingExportHasEligibleChannel(snapshot))...))
	}
	models = append(models, append([]any{"合计", nil, nil, nil, nil, nil}, upstreamExcelRowValues(snapshot.Total, upstreamBillingExportHasEligibleChannel(snapshot))...))
	dayModels := [][]any{append([]any{"日期", "模型名称", "上游模型名称", "计费组"}, usageHeaders...)}
	for _, row := range snapshot.DayModels {
		day, err := time.ParseInLocation("2006-01-02", row.Date, billingLocation)
		if err != nil {
			return nil, err
		}
		dayModels = append(dayModels, append([]any{day, row.ModelName, row.UpstreamModelName, row.BillingGroup}, upstreamExcelRowValues(row.UpstreamBillingRow, upstreamBillingExportHasEligibleChannel(snapshot))...))
	}
	dayModels = append(dayModels, append([]any{"合计", nil, nil, nil}, upstreamExcelRowValues(snapshot.Total, upstreamBillingExportHasEligibleChannel(snapshot))...))
	for _, table := range []struct {
		name   string
		rows   [][]any
		labels int
	}{{"每日汇总", daily, 1}, {"模型汇总", models, 6}, {"每日模型汇总", dayModels, 4}} {
		end, _ := excelize.ColumnNumberToName(table.labels + len(usageHeaders))
		for i, row := range table.rows {
			if err := book.SetSheetRow(table.name, fmt.Sprintf("A%d", i+1), &row); err != nil {
				return nil, err
			}
		}
		if err := book.SetCellStyle(table.name, "A1", end+"1", header); err != nil {
			return nil, err
		}
		first, _ := excelize.ColumnNumberToName(table.labels + 1)
		if err := book.SetCellStyle(table.name, first+"2", fmt.Sprintf("%s%d", end, len(table.rows)), count); err != nil {
			return nil, err
		}
		moneyFirst, _ := excelize.ColumnNumberToName(table.labels + 5)
		moneyLast, _ := excelize.ColumnNumberToName(table.labels + 7)
		if err := book.SetCellStyle(table.name, moneyFirst+"2", fmt.Sprintf("%s%d", moneyLast, len(table.rows)), money); err != nil {
			return nil, err
		}
		if err := book.SetColWidth(table.name, "A", end, 20); err != nil {
			return nil, err
		}
		if err := book.SetColWidth(table.name, moneyFirst, moneyLast, 26); err != nil {
			return nil, err
		}
		statusFirst, _ := excelize.ColumnNumberToName(table.labels + len(usageHeaders) - 1)
		if err := book.SetColWidth(table.name, statusFirst, end, 46); err != nil {
			return nil, err
		}
		if err := book.SetCellStyle(table.name, statusFirst+"2", fmt.Sprintf("%s%d", end, len(table.rows)), wrap); err != nil {
			return nil, err
		}
		if err := book.SetRowHeight(table.name, 1, 36); err != nil {
			return nil, err
		}
		if table.labels > 1 {
			start := "A"
			last := "B"
			if table.labels == 4 {
				start = "B"
				last = "C"
			}
			if err := book.SetColWidth(table.name, start, last, 44); err != nil {
				return nil, err
			}
			if err := book.SetCellStyle(table.name, start+"2", fmt.Sprintf("%s%d", last, len(table.rows)), wrap); err != nil {
				return nil, err
			}
			for i, row := range table.rows[1:] {
				height := max(32, float64(len([]rune(fmt.Sprint(row[0])))/30+1)*16)
				if err := book.SetRowHeight(table.name, i+2, min(height, 409)); err != nil {
					return nil, err
				}
			}
		}
		if table.name != "模型汇总" && len(table.rows) > 2 {
			if err := book.SetCellStyle(table.name, "A2", fmt.Sprintf("A%d", len(table.rows)-1), dateStyle); err != nil {
				return nil, err
			}
		}
		if table.name == "模型汇总" && len(table.rows) > 2 {
			datetimeFormat := "yyyy-mm-dd hh:mm:ss"
			datetimeStyle, err := book.NewStyle(&excelize.Style{CustomNumFmt: &datetimeFormat, Alignment: &excelize.Alignment{Vertical: "center"}})
			if err != nil {
				return nil, err
			}
			if err := book.SetCellStyle(table.name, "E2", fmt.Sprintf("F%d", len(table.rows)-1), datetimeStyle); err != nil {
				return nil, err
			}
			if err := book.SetColWidth(table.name, "E", "F", 24); err != nil {
				return nil, err
			}
		}
		left, _ := excelize.ColumnNumberToName(table.labels + 1)
		if err := book.SetPanes(table.name, &excelize.Panes{Freeze: true, XSplit: table.labels, YSplit: 1, TopLeftCell: left + "2", ActivePane: "bottomRight"}); err != nil {
			return nil, err
		}
	}
	channels := [][]any{{"渠道编号", "渠道名称", "成本折扣配置", "统计范围"}}
	for _, channel := range snapshot.Group.Channels {
		scope := "纳入"
		discount := channel.CostDiscount
		if channel.Missing {
			scope = "渠道不存在（缺失）"
			discount = "未知"
		} else if !channel.Eligible {
			scope = "排除（未配置成本）"
			discount = "未配置"
		}
		channels = append(channels, []any{channel.ID, channel.Name, discount, scope})
	}
	for i, row := range channels {
		if err := book.SetSheetRow("渠道范围", fmt.Sprintf("A%d", i+1), &row); err != nil {
			return nil, err
		}
	}
	if err := book.SetCellStyle("渠道范围", "A1", "D1", header); err != nil {
		return nil, err
	}
	if err := book.SetColWidth("渠道范围", "A", "A", 18); err != nil {
		return nil, err
	}
	if err := book.SetColWidth("渠道范围", "B", "B", 50); err != nil {
		return nil, err
	}
	if err := book.SetColWidth("渠道范围", "C", "D", 28); err != nil {
		return nil, err
	}
	if snapshot.SchemaVersion >= 2 {
		const sheet = "渠道用户成本"
		if _, err := book.NewSheet(sheet); err != nil {
			return nil, err
		}
		rows := [][]any{{"渠道编号", "渠道名称", "用户编号", "用户名", "显示名称", "成本消费金额", "成本退款金额", "用户净成本", "渠道净成本", "占渠道净成本比例", "请求数", "用户金额范围", "渠道金额范围", "用量状态"}}
		channelCoverage := make(map[int]UpstreamBillingRow)
		for _, row := range snapshot.ChannelUsers {
			coverage := channelCoverage[row.ChannelID]
			coverage.Count += row.Count
			coverage.MissingCostRecords += row.MissingCostRecords
			channelCoverage[row.ChannelID] = coverage
		}
		for _, row := range snapshot.ChannelUsers {
			share := "—"
			if row.SharePercent != nil {
				share = *row.SharePercent + "%"
			}
			var requests any = row.Requests
			if len(fmt.Sprint(row.Requests)) > 15 {
				requests = fmt.Sprint(row.Requests)
			}
			rows = append(rows, []any{row.ChannelID, row.ChannelName, row.UserID, row.Username, row.DisplayName, billingExcelAmount(upstreamBillingExportAmount(row.Charge, row.UpstreamBillingRow)), billingExcelAmount(upstreamBillingExportAmount(row.Refund, row.UpstreamBillingRow)), billingExcelAmount(upstreamBillingExportAmount(row.Amount, row.UpstreamBillingRow)), billingExcelAmount(upstreamBillingExportAmount(row.ChannelAmount, channelCoverage[row.ChannelID])), share, requests, upstreamBillingExportCostScope(row.UpstreamBillingRow), upstreamBillingExportCostScope(channelCoverage[row.ChannelID]), upstreamBillingExportUsageStatus(row.UpstreamBillingRow)})
		}
		for i, row := range rows {
			if err := book.SetSheetRow(sheet, fmt.Sprintf("A%d", i+1), &row); err != nil {
				return nil, err
			}
		}
		if err := book.SetCellStyle(sheet, "A1", "N1", header); err != nil {
			return nil, err
		}
		if err := book.SetRowHeight(sheet, 1, 36); err != nil {
			return nil, err
		}
		if len(rows) > 1 {
			if err := book.SetCellStyle(sheet, "F2", fmt.Sprintf("I%d", len(rows)), money); err != nil {
				return nil, err
			}
			if err := book.SetCellStyle(sheet, "B2", fmt.Sprintf("E%d", len(rows)), wrap); err != nil {
				return nil, err
			}
			if err := book.SetCellStyle(sheet, "L2", fmt.Sprintf("N%d", len(rows)), wrap); err != nil {
				return nil, err
			}
			for i := 2; i <= len(rows); i++ {
				if err := book.SetRowHeight(sheet, i, 44); err != nil {
					return nil, err
				}
			}
		}
		for _, widths := range []struct {
			first, last string
			width       float64
		}{{"A", "A", 12}, {"B", "B", 32}, {"C", "C", 12}, {"D", "E", 26}, {"F", "I", 24}, {"J", "J", 24}, {"K", "K", 14}, {"L", "N", 46}} {
			if err := book.SetColWidth(sheet, widths.first, widths.last, widths.width); err != nil {
				return nil, err
			}
		}
		row := len(rows) + 2
		if err := book.MergeCell(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("N%d", row)); err != nil {
			return nil, err
		}
		if err := book.SetCellStr(sheet, fmt.Sprintf("A%d", row), "分摊口径：用户净成本 ÷ 本渠道净成本 × 100%；使用未舍入额度计算并保留四位百分数。渠道净成本不为正或存在成本缺口时不计算比例。渠道净成本在每个用户行重复显示，不可重复加总；退款可能使用户占比为负或超过 100%。"); err != nil {
			return nil, err
		}
		if err := book.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("N%d", row), wrap); err != nil {
			return nil, err
		}
		if err := book.SetRowHeight(sheet, row, 44); err != nil {
			return nil, err
		}
		if err := book.SetPanes(sheet, &excelize.Panes{Freeze: true, XSplit: 3, YSplit: 1, TopLeftCell: "D2", ActivePane: "bottomRight"}); err != nil {
			return nil, err
		}
	}
	showGridLines := false
	for _, sheet := range book.GetSheetList() {
		if err := book.SetSheetView(sheet, 0, &excelize.ViewOptions{ShowGridLines: &showGridLines}); err != nil {
			return nil, err
		}
	}
	buffer, err := book.WriteToBuffer()
	if err != nil {
		return nil, err
	}
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

func upstreamExcelRowValues(row UpstreamBillingRow, hasEligibleChannel bool) []any {
	values := []any{row.Count, row.ChargeCount, row.RefundCount, row.Requests, billingExcelAmount(upstreamBillingExportAmount(row.Charge, row)), billingExcelAmount(upstreamBillingExportAmount(row.Refund, row)), billingExcelAmount(upstreamBillingExportAmount(row.Amount, row)), row.InputTokens, row.OutputTokens, row.CacheReadTokens, row.CacheWriteTokens, row.CacheWrite5mTokens, row.CacheWrite1hTokens, row.VideoTokens, row.VideoSeconds, row.MissingCostRecords, row.MissingUsageRecords, row.EstimatedUsageRecords, row.PendingTasks, upstreamBillingExportCostScope(row), upstreamBillingExportUsageStatus(row)}
	if !hasEligibleChannel {
		values[4], values[5], values[6], values[19] = "未知", "未知", "未知", "无可纳入成本统计的渠道"
	}
	for i, value := range values {
		if number, ok := value.(int64); ok && len(strings.TrimPrefix(fmt.Sprint(number), "-")) > 15 {
			values[i] = fmt.Sprint(number)
		}
	}
	return values
}
