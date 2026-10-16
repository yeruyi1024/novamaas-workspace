package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"math"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/signintech/gopdf"
)

// The upstream template uses the same letterhead, calendar and summary
// structure as customer statements, with internal cost accounting semantics.
func RenderUpstreamBillingStatementPDF(statement *model.UpstreamBillingStatement, snapshot *UpstreamBillingSnapshot, receipt bool) ([]byte, error) {
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
	if receipt && (statement.ConfirmedAt <= 0 || statement.ConfirmedBy <= 0 || statement.PDFSHA256 == "" || statement.ExcelSHA256 == "" || statement.ManifestSHA256 == "") {
		return nil, errors.New("upstream statement is not confirmed with archived evidence")
	}
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.SetInfo(gopdf.PdfInfo{Title: "Upstream statement of account / " + statement.Month, Author: snapshot.Issuer, Creator: "new-api upstream billing statements", CreationDate: time.Unix(statement.CreatedAt, 0).UTC()})
	missingGlyph := rune(0)
	if err := pdf.AddTTFFontDataWithOption("billing", billingPDFFont, gopdf.TtfOption{OnGlyphNotFound: func(r rune) { missingGlyph = r }, OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute}); err != nil {
		return nil, err
	}
	if receipt && statement.ReviewEvidenceSHA256 != "" {
		return renderUpstreamBillingReviewReceiptPDF(pdf, statement, snapshot, &missingGlyph)
	}
	modelPages := [][]UpstreamBillingModelRow{{}}
	modelHeights := map[upstreamArchiveModelKey]float64{}
	var allocationPages [][]upstreamBillingPDFAllocationRow
	if !receipt {
		if err := pdf.SetFont("billing", "", 8.5); err != nil {
			return nil, err
		}
		modelPages = make([][]UpstreamBillingModelRow, 0, 1)
		page := make([]UpstreamBillingModelRow, 0, 12)
		height := 0.0
		for _, row := range snapshot.Models {
			name := row.ModelName
			if row.UpstreamModelName != "" && row.UpstreamModelName != row.ModelName {
				name += " / 上游: " + row.UpstreamModelName
			}
			if row.BillingGroup != "" {
				name += " / 计费组: " + row.BillingGroup
			}
			lines, err := pdf.SplitText(name, 175)
			if err != nil {
				return nil, err
			}
			rowHeight := math.Max(53, float64(len(lines))*8.5*1.45+47)
			if rowHeight > 490 {
				return nil, errors.New("upstream model identity exceeds PDF page space")
			}
			if height+rowHeight > 490 && len(page) > 0 {
				modelPages = append(modelPages, page)
				page = make([]UpstreamBillingModelRow, 0, 12)
				height = 0
			}
			page = append(page, row)
			height += rowHeight
			modelHeights[upstreamArchiveModelKey{row.ModelName, row.UpstreamModelName, row.BillingGroup}] = rowHeight
		}
		modelPages = append(modelPages, page)
		if snapshot.SchemaVersion >= 2 {
			allocationPages, err = paginateUpstreamBillingAllocationsPDF(pdf, &missingGlyph, snapshot.ChannelUsers)
			if err != nil {
				return nil, err
			}
		}
	}
	pages := 2 + len(modelPages) + len(allocationPages)
	if receipt {
		pages = 1
	}
	doc := billingPDFDocument{pdf: pdf}
	doc.upstreamHeader(snapshot, 1, pages)
	if receipt {
		doc.text(40, 120, 515, 22, "上游对账确认回执")
		y := 177.0
		for _, line := range []string{
			fmt.Sprintf("上游  %s (#%d)", snapshot.Group.Name, statement.GroupID),
			fmt.Sprintf("对账单  %s  /  版本 %02d", statement.ID, statement.Revision),
			"账期  " + statement.Month,
			fmt.Sprintf("财务确认管理员  #%d", statement.ConfirmedBy),
			"确认时间  " + time.Unix(statement.ConfirmedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05") + " Asia/Shanghai",
			upstreamBillingExportNetLabel(snapshot) + "  " + upstreamBillingExportTotalAmount(snapshot.Total.Amount, snapshot) + " " + snapshot.Currency.Code,
			"数据截至  " + time.Unix(snapshot.UpdatedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05") + " Asia/Shanghai",
			upstreamBillingExportQualityWarning(snapshot),
			"原始 PDF SHA-256  " + statement.PDFSHA256,
			"原始 Excel SHA-256  " + statement.ExcelSHA256,
			"确认校验清单 SHA-256  " + statement.ManifestSHA256,
			"数据快照 SHA-256  " + statement.SnapshotSHA256,
			"本回执记录财务管理员对上述原始文件及校验清单的确认。原始对账文件保持不变。",
		} {
			y = doc.text(40, y, 515, 10, line) + 13
		}
	} else {
		doc.text(40, 119, 515, 8.5, "UPSTREAM STATEMENT OF ACCOUNT")
		doc.text(40, 143, 350, 26, "上游月度对账单")
		doc.right(410, 149, 145, 20, statement.Month)
		doc.text(40, 191, 515, 8.5, fmt.Sprintf("文件编号  %s  /  版本 %02d", statement.ID, statement.Revision))
		doc.text(40, 210, 515, 8.5, "生成时间  "+time.Unix(statement.CreatedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")+" Asia/Shanghai")
		doc.text(40, 227, 515, 8.5, "数据截至  "+time.Unix(snapshot.UpdatedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")+" Asia/Shanghai")
		y := 252.0
		for _, line := range []string{
			fmt.Sprintf("上游  %s (#%d)  /  渠道归属版本 %d", snapshot.Group.Name, statement.GroupID, snapshot.Group.Version),
			"账期起点  " + time.Unix(statement.StartAt, 0).In(billingLocation).Format("2006-01-02 15:04:05"),
			"账期终点  " + time.Unix(statement.EndAt, 0).In(billingLocation).Format("2006-01-02 15:04:05") + " (不含)",
			"币种  " + snapshot.Currency.Code + "  /  记账时区  " + snapshot.Timezone,
		} {
			y = doc.text(40, y, 515, 10, line) + 8
		}
		y += 15
		chargeLabel, refundLabel := "成本消费金额", "成本退款金额"
		if snapshot.Total.MissingCostRecords > 0 || snapshot.Coverage.MissingChannels > 0 {
			chargeLabel, refundLabel = "已知消费小计", "已知退款小计"
			if snapshot.Total.MissingCostRecords > 0 && snapshot.Total.Count <= snapshot.Total.MissingCostRecords {
				chargeLabel, refundLabel = "成本消费未知", "成本退款未知"
			}
		}
		if !upstreamBillingExportHasEligibleChannel(snapshot) {
			chargeLabel, refundLabel = "成本消费未知", "成本退款未知"
		}
		for i, item := range []struct{ label, value string }{{chargeLabel, snapshot.Total.Charge}, {refundLabel, snapshot.Total.Refund}, {upstreamBillingExportNetLabel(snapshot), snapshot.Total.Amount}} {
			x := 40 + float64(i)*176
			pdf.SetFillColor(241, 245, 251)
			if i == 2 {
				pdf.SetFillColor(25, 51, 94)
			}
			pdf.RectFromUpperLeftWithStyle(x, y, 163, 77, "F")
			pdf.SetTextColor(25, 51, 94)
			if i == 2 {
				pdf.SetTextColor(255, 255, 255)
			}
			doc.text(x+12, y+12, 139, 9, item.label)
			amount, err := upstreamBillingPDFCostAmount(item.value, snapshot.Total, snapshot)
			if err != nil {
				return nil, err
			}
			doc.right(x+12, y+34, 139, 15, amount)
			doc.right(x+12, y+61, 139, 8, snapshot.Currency.Code)
		}
		pdf.SetTextColor(28, 43, 64)
		y += 101
		y = doc.text(40, y, 515, 10, fmt.Sprintf("本期 %d 条成本记录，%d 次请求。成本退款已从净成本中抵减。", snapshot.Total.Count, snapshot.Total.Requests)) + 15
		y = doc.text(40, y, 515, 9, fmt.Sprintf("输入令牌 %d  /  输出令牌 %d  /  缓存读取 %d  /  缓存写入 %d", snapshot.Total.InputTokens, snapshot.Total.OutputTokens, snapshot.Total.CacheReadTokens, snapshot.Total.CacheWriteTokens)) + 14
		includedChannels := 0
		for _, channel := range snapshot.Group.Channels {
			if channel.Eligible && !channel.Missing {
				includedChannels++
			}
		}
		y = doc.text(40, y, 515, 9, fmt.Sprintf("纳入渠道 %d 个；排除未配置成本渠道 %d 个；缺失渠道 %d 个。逐渠道归属及模型用量见 Excel。", includedChannels, len(snapshot.Coverage.ExcludedChannels), snapshot.Coverage.MissingChannels)) + 14
		y = doc.text(40, y, 515, 8.5, "成本采用消费发生时保存的成本快照与已审计调整；本单冻结渠道归属、币种及历史明细，后续配置变更不会修改本单。") + 12
		if !snapshot.Coverage.Complete {
			pdf.SetTextColor(154, 88, 0)
			y = doc.text(40, y, 515, 10, upstreamBillingExportQualityWarning(snapshot)) + 10
			pdf.SetTextColor(28, 43, 64)
			y = doc.text(40, y, 515, 8.5, "告警不限制生成、审核或归档。未知用量不代表零；存在成本缺口时金额仅为已知小计。确认不会改变本单冻结的缺口。") + 10
		}
		y = doc.text(40, y, 515, 7.8, "Excel SHA-256  "+statement.ExcelSHA256) + 10
		y = doc.text(40, y, 515, 7.8, "数据快照 SHA-256  "+statement.SnapshotSHA256)
		if y > 767 {
			return nil, errors.New("upstream cover exceeds PDF page space")
		}
		doc.upstreamHeader(snapshot, 2, pages)
		doc.text(40, 120, 360, 20, "逐日成本明细")
		doc.right(400, 128, 155, 9, statement.Month+" / "+snapshot.Currency.Code)
		doc.text(40, 156, 515, 9, snapshot.Group.Name+"  /  "+snapshot.Timezone)
		if snapshot.Coverage.MissingCostRecords > 0 {
			doc.text(40, 174, 515, 8, "金额为已知成本小计；全缺成本行显示未知。未列入的未知金额不代表零。")
		}
		y = 201
		columns := []float64{40, 140, 190, 310, 425}
		widths := []float64{94, 44, 114, 109, 130}
		for i, label := range []string{"日期", "请求数", "成本消费", "成本退款", "净成本"} {
			doc.text(columns[i], y, widths[i], 8.5, label)
		}
		pdf.SetStrokeColor(201, 211, 225)
		pdf.Line(40, y+20, 555, y+20)
		y += 30
		for _, row := range snapshot.Days {
			doc.text(columns[0], y, widths[0], 8, row.Label)
			doc.right(columns[1], y, widths[1], 8, fmt.Sprint(row.Requests))
			for i, value := range []string{row.Charge, row.Refund, row.Amount} {
				amount, err := upstreamBillingPDFCostAmount(value, row, snapshot)
				if err != nil {
					return nil, err
				}
				doc.right(columns[i+2], y, widths[i+2], 8, amount)
			}
			pdf.SetStrokeColor(233, 238, 244)
			pdf.Line(40, y+13, 555, y+13)
			y += 15.5
		}
		y += 10
		doc.text(40, y, 94, 9, "合计")
		doc.right(140, y, 44, 8.5, fmt.Sprint(snapshot.Total.Requests))
		for i, value := range []string{snapshot.Total.Charge, snapshot.Total.Refund, snapshot.Total.Amount} {
			amount, err := upstreamBillingPDFCostAmount(value, snapshot.Total, snapshot)
			if err != nil {
				return nil, err
			}
			doc.right(columns[i+2], y, widths[i+2], 8.5, amount)
		}
		y = doc.text(40, y+23, 515, 8, "金额保留六位小数。逐日净成本的展示舍入差额："+snapshot.RoundingDifference+" "+snapshot.Currency.Code+"。")
		if y > 767 {
			return nil, errors.New("upstream daily summary exceeds PDF page space")
		}
		for page, rows := range modelPages {
			doc.upstreamHeader(snapshot, page+3, pages)
			doc.text(40, 120, 360, 20, "模型成本与用量汇总")
			doc.right(400, 128, 155, 9, statement.Month+" / "+snapshot.Currency.Code)
			doc.text(40, 158, 515, 8.5, snapshot.Group.Name+"  /  原请求模型、上游实际模型及计费组分别列示")
			if snapshot.Coverage.MissingCostRecords > 0 {
				doc.text(40, 175, 515, 8, "金额为已知成本小计；全缺成本行显示未知。")
			}
			columns := []float64{40, 231, 270, 365, 460}
			widths := []float64{175, 33, 89, 89, 95}
			y = 196
			for i, label := range []string{"模型", "请求数", "成本消费", "成本退款", "净成本"} {
				doc.text(columns[i], y, widths[i], 8.5, label)
			}
			pdf.SetStrokeColor(201, 211, 225)
			pdf.Line(40, y+20, 555, y+20)
			y += 30
			for _, row := range rows {
				name := row.ModelName
				if name == "" {
					name = "未标注模型"
				}
				if row.UpstreamModelName != "" && row.UpstreamModelName != row.ModelName {
					name += " / 上游: " + row.UpstreamModelName
				}
				if row.BillingGroup != "" {
					name += " / 计费组: " + row.BillingGroup
				}
				bottom := doc.text(40, y, 175, 8.5, name)
				doc.right(columns[1], y, widths[1], 8, fmt.Sprint(row.Requests))
				for i, value := range []string{row.Charge, row.Refund, row.Amount} {
					amount, err := upstreamBillingPDFCostAmount(value, row.UpstreamBillingRow, snapshot)
					if err != nil {
						return nil, err
					}
					doc.right(columns[i+2], y, widths[i+2], 8, amount)
				}
				doc.text(40, bottom+5, 515, 7.5, fmt.Sprintf("使用 %d 天 / 已记录用量：输入 %d / 输出 %d / 缓存读取 %d / 缓存写入 %d / 视频令牌 %d / 视频秒 %d", row.ActiveDays, row.InputTokens, row.OutputTokens, row.CacheReadTokens, row.CacheWriteTokens, row.VideoTokens, row.VideoSeconds))
				doc.text(40, bottom+19, 515, 7.5, upstreamBillingExportCostScope(row.UpstreamBillingRow)+" / "+upstreamBillingExportUsageStatus(row.UpstreamBillingRow))
				y += modelHeights[upstreamArchiveModelKey{row.ModelName, row.UpstreamModelName, row.BillingGroup}]
				pdf.SetStrokeColor(233, 238, 244)
				pdf.Line(40, y-5, 555, y-5)
			}
			if len(rows) == 0 {
				doc.text(40, y, 515, 10, "本期没有符合成本统计范围的使用记录。")
			}
			doc.text(40, 746, 515, 8, "使用明细、每日模型汇总及完整用量单位见 Excel 与校验清单对应的证据文件。")
		}
		channelCoverage := make(map[int]UpstreamBillingRow)
		for _, row := range snapshot.ChannelUsers {
			coverage := channelCoverage[row.ChannelID]
			coverage.Count += row.Count
			coverage.MissingCostRecords += row.MissingCostRecords
			channelCoverage[row.ChannelID] = coverage
		}
		for page, rows := range allocationPages {
			doc.upstreamHeader(snapshot, 3+len(modelPages)+page, pages)
			doc.text(40, 120, 360, 20, "渠道用户成本分配")
			doc.right(400, 128, 155, 9, statement.Month+" / "+snapshot.Currency.Code)
			doc.text(40, 156, 515, 8, "占比 = 用户净成本 ÷ 本渠道净成本，按未舍入额度计算。渠道净额不为正或缺成本时不计算。")
			columns := []float64{40, 195, 228, 309, 390, 471}
			widths := []float64{150, 28, 76, 76, 76, 84}
			y = 196
			for i, label := range []string{"渠道 / 用户", "请求数", "成本消费", "成本退款", "用户净成本", "渠道占比"} {
				doc.text(columns[i], y, widths[i], 8, label)
			}
			pdf.SetStrokeColor(201, 211, 225)
			pdf.Line(40, y+20, 555, y+20)
			y += 30
			for _, item := range rows {
				row := item.row
				bottom := doc.text(40, y, widths[0], 8.5, item.identity)
				doc.right(columns[1], y, widths[1], 8, fmt.Sprint(row.Requests))
				for i, value := range []string{row.Charge, row.Refund, row.Amount} {
					amount, err := upstreamBillingPDFCostAmount(value, row.UpstreamBillingRow, snapshot)
					if err != nil {
						return nil, err
					}
					doc.right(columns[i+2], y, widths[i+2], 8, amount)
				}
				share := "—"
				if row.SharePercent != nil {
					share = *row.SharePercent + "%"
				}
				doc.right(columns[5], y, widths[5], 8, share)
				doc.text(40, bottom+5, 515, 7.5, "本渠道净成本  "+upstreamBillingExportAmount(row.ChannelAmount, channelCoverage[row.ChannelID])+" "+snapshot.Currency.Code+" / "+upstreamBillingExportCostScope(channelCoverage[row.ChannelID])+"（分母，不重复加总）")
				doc.text(40, bottom+19, 515, 7.5, upstreamBillingExportCostScope(row.UpstreamBillingRow)+" / "+upstreamBillingExportUsageStatus(row.UpstreamBillingRow))
				y += item.height
				pdf.SetStrokeColor(233, 238, 244)
				pdf.Line(40, y-5, 555, y-5)
			}
			if len(rows) == 0 {
				doc.text(40, y, 515, 10, "本期没有符合成本统计范围的用户使用记录。")
			}
			doc.text(40, 735, 515, 7.5, "退款可能使用户占比为负或超过 100%；四位百分数的展示合计可能有舍入差异。")
			doc.text(40, 750, 515, 7.5, "未支持的姓名字符以 Unicode 编号显示；原始名称完整保存在冻结快照及 Excel 中。")
		}
	}
	if doc.err != nil {
		return nil, doc.err
	}
	if missingGlyph != 0 {
		return nil, fmt.Errorf("upstream document contains unsupported PDF character %U", missingGlyph)
	}
	return pdf.GetBytesPdfReturnErr()
}

func upstreamBillingPDFCostAmount(value string, row UpstreamBillingRow, snapshot *UpstreamBillingSnapshot) (string, error) {
	if !upstreamBillingExportHasEligibleChannel(snapshot) || upstreamBillingExportAmount(value, row) == "未知" {
		return "未知", nil
	}
	return billingAccountingNumber(value)
}

// The review receipt is an additional document. It binds the original manifest
// and the exact ordered supplier attachments without rewriting either original.
func renderUpstreamBillingReviewReceiptPDF(pdf *gopdf.GoPdf, statement *model.UpstreamBillingStatement, snapshot *UpstreamBillingSnapshot, missingGlyph *rune) ([]byte, error) {
	evidenceBody := []byte(statement.ReviewEvidenceSnapshot)
	hash := sha256.Sum256(evidenceBody)
	if hex.EncodeToString(hash[:]) != statement.ReviewEvidenceSHA256 {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	var evidence model.UpstreamBillingReviewEvidence
	if err := common.Unmarshal(evidenceBody, &evidence); err != nil {
		return nil, err
	}
	if evidence.SchemaVersion != 1 || evidence.StatementID != statement.ID || evidence.ManifestSHA256 != statement.ManifestSHA256 {
		return nil, model.ErrBillingEvidenceIntegrity
	}
	type receiptLine struct {
		text string
		size float64
	}
	lines := []receiptLine{
		{fmt.Sprintf("上游  %s (#%d)", snapshot.Group.Name, statement.GroupID), 10},
		{fmt.Sprintf("对账单  %s / 版本 %02d", statement.ID, statement.Revision), 10},
		{"账期  " + statement.Month, 10},
		{fmt.Sprintf("财务确认管理员  #%d", statement.ConfirmedBy), 10},
		{"确认时间  " + time.Unix(statement.ConfirmedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05") + " Asia/Shanghai", 10},
		{upstreamBillingExportNetLabel(snapshot) + "  " + upstreamBillingExportTotalAmount(snapshot.Total.Amount, snapshot) + " " + snapshot.Currency.Code, 10},
		{"数据截至  " + time.Unix(snapshot.UpdatedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05") + " Asia/Shanghai", 9},
		{upstreamBillingExportQualityWarning(snapshot), 9},
		{"原始 PDF SHA-256  " + statement.PDFSHA256, 8.5},
		{"原始 Excel SHA-256  " + statement.ExcelSHA256, 8.5},
		{"原始校验清单 SHA-256  " + statement.ManifestSHA256, 8.5},
		{"审核证据清单 SHA-256  " + statement.ReviewEvidenceSHA256, 8.5},
		{fmt.Sprintf("供应商附件  %d 份", len(evidence.Attachments)), 10},
	}
	groups := [][]receiptLine{lines}
	if err := pdf.SetFont("billing", "", 9); err != nil {
		return nil, err
	}
	escapedNames := false
	for index, attachment := range evidence.Attachments {
		var filename strings.Builder
		for _, character := range attachment.OriginalFilename {
			*missingGlyph = 0
			if _, err := pdf.MeasureTextWidth(string(character)); err != nil {
				return nil, err
			}
			if *missingGlyph != 0 {
				fmt.Fprintf(&filename, "[U+%04X]", character)
				escapedNames = true
			} else {
				filename.WriteRune(character)
			}
		}
		*missingGlyph = 0
		groups = append(groups, []receiptLine{{fmt.Sprintf("附件 %02d  %s", index+1, filename.String()), 9}, {"SHA-256  " + attachment.SHA256, 8.5}, {fmt.Sprintf("%s / %d bytes / 上传人 #%d / %s", attachment.ContentType, attachment.Size, attachment.UploadedBy, time.Unix(attachment.UploadedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")), 8}})
	}
	notes := make([]receiptLine, 0, 2)
	if escapedNames {
		notes = append(notes, receiptLine{"附件名中未支持的字符以 Unicode 编号显示；原始文件名完整保存在审核证据清单中。", 8})
	}
	notes = append(notes, receiptLine{"本回执记录财务管理员对原始对账文件和本清单附件的审核确认。原始 PDF、Excel 和供应商附件均保持不变。", 9})
	groups = append(groups, notes)
	pages := [][]receiptLine{{}}
	y := 177.0
	for _, group := range groups {
		wrappedGroup := make([]receiptLine, 0, len(group))
		groupHeight := 0.0
		for _, block := range group {
			if err := pdf.SetFont("billing", "", block.size); err != nil {
				return nil, err
			}
			wrapped, err := pdf.SplitText(block.text, 515)
			if err != nil {
				return nil, err
			}
			for _, line := range wrapped {
				wrappedGroup = append(wrappedGroup, receiptLine{line, block.size})
				groupHeight += block.size*1.45 + 7
			}
		}
		if groupHeight > 590 {
			return nil, errors.New("upstream receipt attachment exceeds PDF page space")
		}
		if y+groupHeight > 767 {
			pages = append(pages, []receiptLine{})
			y = 177
		}
		pages[len(pages)-1] = append(pages[len(pages)-1], wrappedGroup...)
		y += groupHeight
	}
	doc := billingPDFDocument{pdf: pdf}
	for page, lines := range pages {
		doc.upstreamHeader(snapshot, page+1, len(pages))
		doc.text(40, 120, 515, 22, "上游对账审核确认回执")
		if page > 0 {
			doc.text(40, 155, 515, 8, statement.ID+" / "+statement.Month)
		}
		y = 177
		for _, line := range lines {
			y = doc.text(40, y, 515, line.size, line.text) + 7
		}
	}
	if doc.err != nil {
		return nil, doc.err
	}
	if *missingGlyph != 0 {
		return nil, fmt.Errorf("upstream receipt contains unsupported PDF character %U", *missingGlyph)
	}
	return pdf.GetBytesPdfReturnErr()
}

func (doc *billingPDFDocument) upstreamHeader(snapshot *UpstreamBillingSnapshot, page, pages int) {
	if doc.err != nil {
		return
	}
	doc.pdf.AddPage()
	config, _, err := image.DecodeConfig(bytes.NewReader(snapshot.PDFLogoPNG))
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		doc.err = errors.New("upstream document has invalid frozen platform logo")
		return
	}
	mark, err := gopdf.ImageHolderByBytes(snapshot.PDFLogoPNG)
	if err != nil {
		doc.err = err
		return
	}
	scale := math.Min(38/float64(config.Width), 38/float64(config.Height))
	width, height := float64(config.Width)*scale, float64(config.Height)*scale
	if doc.err = doc.pdf.ImageByHolder(mark, 40+(38-width)/2, 42+(38-height)/2, &gopdf.Rect{W: width, H: height}); doc.err != nil {
		return
	}
	doc.pdf.SetTextColor(25, 51, 94)
	issuerWidth := 465.0
	if snapshot.OperatingName != "" || len(snapshot.PDFOperatingLogoPNG) > 0 {
		issuerWidth = 335
	}
	end := doc.text(90, 43, issuerWidth, 15, snapshot.Issuer)
	if end > 80 {
		doc.err = errors.New("upstream platform name exceeds PDF letterhead")
		return
	}
	doc.pdf.SetTextColor(90, 106, 129)
	doc.text(90, end+1, issuerWidth, 8, "UPSTREAM RECONCILIATION / 上游成本对账")
	if len(snapshot.PDFOperatingLogoPNG) > 0 {
		config, _, err := image.DecodeConfig(bytes.NewReader(snapshot.PDFOperatingLogoPNG))
		if err != nil || config.Width <= 0 || config.Height <= 0 {
			doc.err = errors.New("upstream operating logo is invalid")
			return
		}
		mark, err := gopdf.ImageHolderByBytes(snapshot.PDFOperatingLogoPNG)
		if err != nil {
			doc.err = err
			return
		}
		scale := math.Min(125/float64(config.Width), 40/float64(config.Height))
		width, height := float64(config.Width)*scale, float64(config.Height)*scale
		if doc.err = doc.pdf.ImageByHolder(mark, 555-width, 42+(40-height)/2, &gopdf.Rect{W: width, H: height}); doc.err != nil {
			return
		}
	} else if snapshot.OperatingName != "" {
		doc.right(430, 54, 125, 9, snapshot.OperatingName)
	}
	doc.pdf.SetStrokeColor(39, 71, 122)
	doc.pdf.SetLineWidth(1.5)
	doc.pdf.Line(40, 100, 555, 100)
	doc.pdf.SetStrokeColor(221, 228, 237)
	doc.pdf.SetLineWidth(0.5)
	doc.pdf.Line(40, 784, 555, 784)
	doc.pdf.SetTextColor(99, 113, 132)
	footer := snapshot.PDFFooter
	if footer == "" {
		footer = "new-api  /  上游成本对账存档"
	}
	end = doc.text(40, 790, 445, 7.5, footer)
	if end > 820 {
		doc.err = errors.New("upstream Footer exceeds PDF page space")
		return
	}
	doc.right(500, 794, 55, 8, fmt.Sprintf("%02d / %02d", page, pages))
	doc.pdf.SetTextColor(28, 43, 64)
}
