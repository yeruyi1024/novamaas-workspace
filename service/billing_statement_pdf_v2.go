package service

import (
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/signintech/gopdf"
)

// This versioned copy of the repository's platform mark is permanent. Do not
// fetch a mutable remote logo while rendering an immutable financial document.
//
//go:embed billingassets/platform-mark-v2.png
var billingPlatformMarkV2 []byte

type billingPDFDocument struct {
	pdf *gopdf.GoPdf
	err error
}

// text wraps identity and legal copy; monetary values use right-aligned cells.
func (doc *billingPDFDocument) text(x, y, width, size float64, text string) float64 {
	if doc.err != nil {
		return y
	}
	if doc.err = doc.pdf.SetFont("billing", "", size); doc.err != nil {
		return y
	}
	lines, err := doc.pdf.SplitText(text, width)
	if err != nil {
		doc.err = err
		return y
	}
	for _, line := range lines {
		if y+size > 820 {
			doc.err = errors.New("billing PDF content exceeds page boundary")
			return y
		}
		doc.pdf.SetXY(x, y)
		if doc.err = doc.pdf.Cell(nil, line); doc.err != nil {
			return y
		}
		y += size * 1.45
	}
	return y
}

func (doc *billingPDFDocument) right(x, y, width, size float64, text string) {
	if doc.err != nil {
		return
	}
	for size >= 6 {
		if doc.err = doc.pdf.SetFont("billing", "", size); doc.err != nil {
			return
		}
		measured, err := doc.pdf.MeasureTextWidth(text)
		if err != nil {
			doc.err = err
			return
		}
		if measured <= width {
			doc.pdf.SetXY(x+width-measured, y)
			doc.err = doc.pdf.Cell(nil, text)
			return
		}
		size -= 0.5
	}
	doc.err = errors.New("billing value exceeds PDF column width")
}

func (doc *billingPDFDocument) header(issuer string, page, pages int) {
	if doc.err != nil {
		return
	}
	doc.pdf.AddPage()
	mark, err := gopdf.ImageHolderByBytes(billingPlatformMarkV2)
	if err != nil {
		doc.err = err
		return
	}
	if doc.err = doc.pdf.ImageByHolder(mark, 40, 33, &gopdf.Rect{W: 38, H: 38}); doc.err != nil {
		return
	}
	doc.pdf.SetTextColor(25, 51, 94)
	end := doc.text(90, 34, 465, 15, issuer)
	if end > 80 {
		doc.err = errors.New("platform name exceeds PDF letterhead")
		return
	}
	doc.pdf.SetTextColor(90, 106, 129)
	doc.text(90, end+1, 465, 8, "CUSTOMER RECONCILIATION  /  客户服务消费对账")
	doc.pdf.SetStrokeColor(39, 71, 122)
	doc.pdf.SetLineWidth(1.5)
	doc.pdf.Line(40, 100, 555, 100)
	doc.pdf.SetStrokeColor(221, 228, 237)
	doc.pdf.SetLineWidth(0.5)
	doc.pdf.Line(40, 784, 555, 784)
	doc.pdf.SetTextColor(99, 113, 132)
	doc.text(40, 794, 425, 7.5, "new-api  ·  电子对账存档 / Electronic reconciliation archive  ·  v2")
	doc.right(500, 794, 55, 8, fmt.Sprintf("%02d / %02d", page, pages))
	doc.pdf.SetTextColor(28, 43, 64)
}

func renderBillingStatementPDFV2(statement *model.BillingStatement, snapshot *BillingSnapshot, receipt bool) ([]byte, error) {
	if receipt && (statement.ConfirmedAt <= 0 || statement.PDFSHA256 == "" || statement.ManifestSHA256 == "") {
		return nil, errors.New("statement is not confirmed with archived evidence")
	}
	pages := 2
	if receipt {
		pages = 3
	}
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.SetInfo(gopdf.PdfInfo{Title: "Monthly statement of account / " + statement.Month, Author: snapshot.Issuer, Creator: "new-api billing statements", CreationDate: time.Unix(statement.CreatedAt, 0).UTC()})
	missingGlyph := false
	if err := pdf.AddTTFFontDataWithOption("billing", billingPDFFont, gopdf.TtfOption{OnGlyphNotFound: func(r rune) { missingGlyph = true }, OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute}); err != nil {
		return nil, err
	}
	doc := billingPDFDocument{pdf: pdf}
	doc.header(snapshot.Issuer, 1, pages)
	pdf.SetTextColor(89, 105, 127)
	doc.text(40, 121, 360, 8.5, "STATEMENT OF ACCOUNT")
	pdf.SetTextColor(25, 51, 94)
	doc.text(40, 141, 350, 26, "月度对账单")
	doc.right(410, 144, 145, 22, statement.Month)
	pdf.SetTextColor(88, 104, 126)
	doc.text(40, 185, 515, 8.5, fmt.Sprintf("文件编号  %s   /   版本 %02d", statement.ID, statement.Revision))
	doc.text(40, 202, 515, 8.5, "生成时间  "+time.Unix(statement.CreatedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")+"   /   Asia/Shanghai (UTC+08:00)")

	pdf.SetTextColor(89, 105, 127)
	doc.text(40, 242, 232, 8.5, "服务平台 / SERVICE PROVIDER")
	doc.text(308, 242, 247, 8.5, "对账客户 / BILL TO")
	pdf.SetTextColor(28, 43, 64)
	left := doc.text(40, 264, 232, 12, snapshot.Issuer)
	left = doc.text(40, left+14, 232, 9, "账期起点  "+time.Unix(statement.StartAt, 0).In(billingLocation).Format("2006-01-02 15:04:05"))
	left = doc.text(40, left+5, 232, 9, "账期终点  "+time.Unix(statement.EndAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")+" (不含)")
	left = doc.text(40, left+5, 232, 9, "计价币种  "+snapshot.Currency.Code+"  /  "+snapshot.Currency.Symbol)
	right := doc.text(308, 264, 247, 10, snapshot.CompanyTitle)
	right = doc.text(308, right+12, 247, 9, fmt.Sprintf("客户账户  %s (#%d)", snapshot.Username, statement.UserID))
	if snapshot.DisplayName != "" && snapshot.DisplayName != snapshot.CompanyTitle {
		right = doc.text(308, right+4, 247, 9, "账户名称  "+snapshot.DisplayName)
	}
	right = doc.text(308, right+4, 247, 9, "纳税人识别号  "+snapshot.TaxID)
	y := max(left, right) + 27
	for i, item := range []struct{ label, value string }{
		{"消费总额 / CHARGES", snapshot.Total.Charge},
		{"退款总额 / REFUNDS", snapshot.Total.Refund},
		{"本期净额 / NET TOTAL", snapshot.Total.Amount},
	} {
		x := 40 + float64(i)*176
		pdf.SetFillColor(241, 245, 251)
		if i == 2 {
			pdf.SetFillColor(25, 51, 94)
		}
		pdf.RectFromUpperLeftWithStyle(x, y, 163, 79, "F")
		pdf.SetTextColor(83, 102, 129)
		if i == 2 {
			pdf.SetTextColor(220, 231, 247)
		}
		doc.text(x+12, y+11, 139, 8, item.label)
		pdf.SetTextColor(25, 51, 94)
		if i == 2 {
			pdf.SetTextColor(255, 255, 255)
		}
		doc.right(x+12, y+34, 139, 16, item.value)
		doc.right(x+12, y+60, 139, 8, snapshot.Currency.Code)
	}
	y += 96
	pdf.SetTextColor(28, 43, 64)
	y = doc.text(40, y, 515, 9, fmt.Sprintf("本期明细记录 %d 条  ·  按原始整数额度汇总，金额保留六位小数。", snapshot.Total.Count))
	y = doc.text(40, y+5, 515, 8.5, "换算口径  1 USD = "+snapshot.Currency.Rate+" "+snapshot.Currency.Code+"；1 USD = "+snapshot.Currency.QuotaPerUnit+" quota。展示舍入差额："+snapshot.RoundingDifference+" "+snapshot.Currency.Code+"。")
	y += 20
	pdf.SetTextColor(89, 105, 127)
	y = doc.text(40, y, 515, 8.5, "对账说明 / RECONCILIATION NOTES")
	y = doc.text(40, y+8, 515, 8.5, "本文件用于平台与客户核对服务消费，不是银行凭证、发票或支付收据。管理员充值及临时预扣不计入消费；退款抵减净额，不代表再次支付。")
	if snapshot.HistoryImportID != "" {
		y = doc.text(40, y+6, 515, 8.5, "本期数据为经管理员核验的历史日志补录，按原日志时间归属账期，未改变钱包余额。导入批次："+snapshot.HistoryImportID+"。")
	}
	y += 18
	doc.text(40, y, 515, 8, "冻结数据快照 / SHA-256")
	y = doc.text(40, y+17, 515, 8, statement.SnapshotSHA256)
	if y > 767 {
		return nil, errors.New("billing cover identity exceeds available page space")
	}

	doc.header(snapshot.Issuer, 2, pages)
	doc.text(40, 119, 360, 18, "逐日消费明细")
	doc.right(380, 126, 175, 9, statement.Month+"  /  "+snapshot.Currency.Code)
	doc.text(40, 154, 515, 8.5, fmt.Sprintf("客户 %s (#%d)  ·  版本 %02d  ·  Asia/Shanghai", snapshot.Username, statement.UserID, statement.Revision))
	pdf.SetFillColor(25, 51, 94)
	pdf.RectFromUpperLeftWithStyle(40, 179, 515, 26, "F")
	pdf.SetTextColor(255, 255, 255)
	columns := []float64{50, 147, 257, 367, 477}
	widths := []float64{85, 98, 98, 98, 68}
	for i, title := range []string{"日期", "消费金额", "退款金额", "净消费金额", "记录数"} {
		if i == 0 {
			doc.text(columns[i], 186, widths[i], 9, title)
		} else {
			doc.right(columns[i], 186, widths[i], 9, title)
		}
	}
	y = 205
	if len(snapshot.Days) > 31 {
		return nil, errors.New("billing month exceeds 31 rows")
	}
	for i, row := range snapshot.Days {
		if i%2 == 0 {
			pdf.SetFillColor(245, 247, 251)
			pdf.RectFromUpperLeftWithStyle(40, y, 515, 15, "F")
		}
		pdf.SetTextColor(35, 49, 69)
		values := []string{row.Label, row.Charge, row.Refund, row.Amount, fmt.Sprint(row.Count)}
		if row.State == "outside_period" {
			values = []string{row.Label, "—", "—", "未纳入", "—"}
		}
		for col, value := range values {
			if col == 0 {
				doc.text(columns[col], y+3.5, widths[col], 8.2, value)
			} else {
				doc.right(columns[col], y+3.5, widths[col], 8.2, value)
			}
		}
		y += 15
	}
	pdf.SetFillColor(227, 234, 245)
	pdf.RectFromUpperLeftWithStyle(40, y, 515, 25, "F")
	pdf.SetTextColor(25, 51, 94)
	for col, value := range []string{"合计", snapshot.Total.Charge, snapshot.Total.Refund, snapshot.Total.Amount, fmt.Sprint(snapshot.Total.Count)} {
		if col == 0 {
			doc.text(columns[col], y+7, widths[col], 8.8, value)
		} else {
			doc.right(columns[col], y+7, widths[col], 8.8, value)
		}
	}
	pdf.SetTextColor(89, 105, 127)
	y = doc.text(40, y+37, 515, 8, "未纳入：正式记账起点之前。明细记录数包含消费及退款；完整逐笔明细可从平台下载经校验的归档文件。")
	y = doc.text(40, y+5, 515, 8, "下发、异议及确认状态以平台记录为准。原始文件不随状态改变；确认后另附回执。逐日净额之和加舍入差额等于本期净额。")
	if y > 767 {
		return nil, errors.New("billing daily notes exceed available page space")
	}
	if receipt {
		doc.header(snapshot.Issuer, 3, pages)
		doc.text(40, 122, 515, 23, "对账确认回执")
		pdf.SetTextColor(89, 105, 127)
		doc.text(40, 163, 515, 9, "CONFIRMATION RECEIPT  /  原始对账单附页")
		pdf.SetFillColor(237, 246, 241)
		pdf.RectFromUpperLeftWithStyle(40, 204, 515, 63, "F")
		pdf.SetTextColor(30, 102, 74)
		doc.text(56, 218, 483, 16, "客户账户已执行确认")
		doc.text(56, 244, 483, 9, "确认时间  "+time.Unix(statement.ConfirmedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05")+"  Asia/Shanghai")
		pdf.SetTextColor(28, 43, 64)
		y = doc.text(40, 297, 515, 10, fmt.Sprintf("确认账户  %s (#%d)", snapshot.Username, statement.UserID))
		y = doc.text(40, y+12, 515, 10, "对账主体  "+snapshot.CompanyTitle)
		y = doc.text(40, y+12, 515, 10, "纳税人识别号  "+snapshot.TaxID)
		y = doc.text(40, y+12, 515, 10, fmt.Sprintf("关联文件  %s  /  %s  /  版本 %02d", statement.ID, statement.Month, statement.Revision))
		y = doc.text(40, y+12, 515, 13, "确认净额  "+snapshot.Currency.Code+" "+snapshot.Total.Amount)
		y += 28
		pdf.SetTextColor(89, 105, 127)
		for _, item := range []struct{ label, digest string }{{"原始 PDF / SHA-256", statement.PDFSHA256}, {"明细归档清单 / SHA-256", statement.ManifestSHA256}, {"冻结数据快照 / SHA-256", statement.SnapshotSHA256}} {
			y = doc.text(40, y, 515, 8.5, item.label)
			y = doc.text(40, y+6, 515, 8, item.digest) + 16
		}
		y = doc.text(40, y+6, 515, 9, "本回执记录客户通过平台真实登录会话执行的确认操作。它不替代电子签章或法定数字签名；原始 PDF 与明细归档保持不变，可通过以上指纹核验关联文件。")
		if y > 767 {
			return nil, errors.New("billing receipt identity exceeds available page space")
		}
	}
	if doc.err != nil {
		return nil, doc.err
	}
	if missingGlyph {
		return nil, errors.New("company information contains characters unsupported by the PDF font")
	}
	return pdf.GetBytesPdfReturnErr()
}
