package service

import (
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/signintech/gopdf"
)

//go:embed billingassets/NotoSansSC-Regular.ttf
var billingPDFFont []byte

// All dates, amounts and identity data come from the frozen snapshot.
// Confirmation appends a receipt; the original PDF is never re-rendered.
func RenderBillingStatementPDF(statement *model.BillingStatement, snapshot *BillingSnapshot, receipt bool) ([]byte, error) {
	switch snapshot.PDFTemplateVersion {
	case 0, 1:
		return renderBillingStatementPDFV1(statement, snapshot, receipt)
	case 2:
		return renderBillingStatementPDFV2(statement, snapshot, receipt)
	case 3:
		return renderBillingStatementPDFV3(statement, snapshot, receipt)
	default:
		return nil, errors.New("unsupported billing PDF template version")
	}
}

// Retained byte-for-byte for legacy archive retries and confirmation receipts.
func renderBillingStatementPDFV1(statement *model.BillingStatement, snapshot *BillingSnapshot, receipt bool) ([]byte, error) {
	pageCount := 2
	if receipt {
		pageCount = 3
	}
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.SetInfo(gopdf.PdfInfo{Title: "Statement of account", Author: snapshot.Issuer, Creator: "new-api billing statements", CreationDate: time.Unix(statement.CreatedAt, 0).UTC()})
	missingGlyph := false
	err := pdf.AddTTFFontDataWithOption("billing", billingPDFFont, gopdf.TtfOption{OnGlyphNotFound: func(r rune) { missingGlyph = true }, OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute})
	if err != nil {
		return nil, err
	}
	pdf.AddPage()
	pdf.SetTextColor(24, 39, 59)
	y := 45.0
	for _, text := range []string{"STATEMENT OF ACCOUNT", "月度对账单 / " + statement.Month} {
		if err := billingPDFText(pdf, &y, text, 18); err != nil {
			return nil, err
		}
	}
	y += 18
	metadata := []string{
		fmt.Sprintf("对账单编号  %s  /  版本 %d", statement.ID, statement.Revision),
		"客户企业抬头  " + snapshot.CompanyTitle,
		"纳税人识别号  " + snapshot.TaxID,
		fmt.Sprintf("客户账户  %d", statement.UserID),
		"服务方  " + snapshot.Issuer,
		"记账时区  Asia/Shanghai (UTC+08:00)",
		"账期  " + time.Unix(statement.StartAt, 0).In(billingLocation).Format("2006-01-02 15:04:05") + " 至 " + time.Unix(statement.EndAt, 0).In(billingLocation).Format("2006-01-02 15:04:05") + " (不含)",
	}
	for _, text := range metadata {
		if err := billingPDFText(pdf, &y, text, 10); err != nil {
			return nil, err
		}
	}
	y += 24
	pdf.SetStrokeColor(201, 211, 222)
	pdf.Line(40, y, 555, y)
	y += 20
	for _, row := range []string{
		"消费总额  " + snapshot.Currency.Symbol + " " + snapshot.Total.Charge,
		"退款总额  " + snapshot.Currency.Symbol + " " + snapshot.Total.Refund,
		"本期净额  " + snapshot.Currency.Symbol + " " + snapshot.Total.Amount,
		fmt.Sprintf("明细记录  %d 条", snapshot.Total.Count),
	} {
		if err := billingPDFText(pdf, &y, row, 14); err != nil {
			return nil, err
		}
	}
	y += 22
	for _, text := range []string{
		"币种 " + snapshot.Currency.Code + "；1 USD = " + snapshot.Currency.Rate + " " + snapshot.Currency.Code + "；1 USD = " + snapshot.Currency.QuotaPerUnit + " quota。",
		"本单用于管理员与客户核对服务消费，不是银行凭证、发票或支付收据。管理员充值及预扣冻结金额不计入消费。",
		"按实际结算时间记账；跨期退款计入退款发生期。金额保留六位小数，逐日显示的舍入差异不改变本期净额。",
		"逐日净额的展示舍入差额  " + snapshot.RoundingDifference + " " + snapshot.Currency.Code + "；逐日显示净额之和加此差额等于本期净额。",
		"确认时间及确认人由系统独立存档。请在平台查看下发、异议和确认状态；本文件不会因状态变化被覆盖。",
		"数据快照 SHA-256: " + statement.SnapshotSHA256,
	} {
		if err := billingPDFText(pdf, &y, text, 9); err != nil {
			return nil, err
		}
	}
	y = 794
	if err := billingPDFText(pdf, &y, fmt.Sprintf("1 / %d  |  Monthly reconciliation", pageCount), 8); err != nil {
		return nil, err
	}

	pdf.AddPage()
	y = 45
	if err := billingPDFText(pdf, &y, "逐日消费明细 / "+statement.Month, 18); err != nil {
		return nil, err
	}
	if err := billingPDFText(pdf, &y, fmt.Sprintf("账户 %d  /  版本 %d  /  %s", statement.UserID, statement.Revision, snapshot.Currency.Code), 10); err != nil {
		return nil, err
	}
	y += 22
	columns := []float64{40, 158, 267, 376, 494}
	header := []string{"日期", "消费", "退款", "净额", "记录数"}
	if err := pdf.SetFont("billing", "", 9); err != nil {
		return nil, err
	}
	for i, text := range header {
		pdf.SetXY(columns[i], y)
		if err := pdf.Cell(nil, text); err != nil {
			return nil, err
		}
	}
	y += 23
	for _, row := range snapshot.Days {
		values := []string{row.Label, row.Charge, row.Refund, row.Amount, fmt.Sprint(row.Count)}
		for i, text := range values {
			pdf.SetXY(columns[i], y)
			width, err := pdf.MeasureTextWidth(text)
			if err != nil {
				return nil, err
			}
			if width > 105 {
				return nil, errors.New("billing amount exceeds PDF column width")
			}
			if err := pdf.Cell(nil, text); err != nil {
				return nil, err
			}
		}
		pdf.SetStrokeColor(230, 235, 240)
		pdf.Line(40, y+17, 555, y+17)
		y += 18
	}
	y += 18
	if err := billingPDFText(pdf, &y, "合计净额  "+snapshot.Currency.Symbol+" "+snapshot.Total.Amount, 12); err != nil {
		return nil, err
	}
	y = 794
	if err := billingPDFText(pdf, &y, fmt.Sprintf("2 / %d  |  Detailed entries are available as a verified archive.", pageCount), 8); err != nil {
		return nil, err
	}
	if receipt {
		if statement.ConfirmedAt <= 0 {
			return nil, errors.New("statement is not confirmed")
		}
		pdf.AddPage()
		y = 45
		for _, text := range []string{
			"CONFIRMATION RECEIPT / 确认回执",
			"对账单编号  " + statement.ID,
			fmt.Sprintf("确认账户  %d", statement.UserID),
			"确认时间  " + time.Unix(statement.ConfirmedAt, 0).In(billingLocation).Format("2006-01-02 15:04:05") + " Asia/Shanghai",
			"原始 PDF SHA-256  " + statement.PDFSHA256,
			"明细清单 SHA-256  " + statement.ManifestSHA256,
			"该回执证明账户在平台执行了确认操作，不替代电子签章或法定数字签名。",
		} {
			if err := billingPDFText(pdf, &y, text, 11); err != nil {
				return nil, err
			}
			y += 10
		}
		y = 794
		if err := billingPDFText(pdf, &y, "3 / 3  |  Confirmation receipt", 8); err != nil {
			return nil, err
		}
	}
	if missingGlyph {
		return nil, errors.New("company information contains characters unsupported by the PDF font")
	}
	return pdf.GetBytesPdfReturnErr()
}

func billingPDFText(pdf *gopdf.GoPdf, y *float64, text string, size float64) error {
	if err := pdf.SetFont("billing", "", size); err != nil {
		return err
	}
	lines, err := pdf.SplitText(text, 515)
	if err != nil {
		return err
	}
	for _, line := range lines {
		if *y+size > 820 {
			return errors.New("billing PDF content exceeds page boundary")
		}
		pdf.SetXY(40, *y)
		if err := pdf.Cell(nil, line); err != nil {
			return err
		}
		*y += size * 1.6
	}
	*y += 4
	return nil
}
