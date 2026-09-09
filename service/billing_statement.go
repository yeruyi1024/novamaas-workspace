package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
)

type BillingCurrency struct {
	Code         string `json:"code"`
	Symbol       string `json:"symbol"`
	Rate         string `json:"rate"`
	QuotaPerUnit string `json:"quota_per_unit"`
}
type BillingRow struct {
	Label       string `json:"label"`
	Charge      string `json:"charge"`
	Refund      string `json:"refund"`
	Amount      string `json:"amount"`
	Count       int64  `json:"count"`
	ChargeQuota int64  `json:"charge_quota"`
	RefundQuota int64  `json:"refund_quota"`
	State       string `json:"state,omitempty"`
}
type BillingSnapshot struct {
	SchemaVersion       int             `json:"schema_version"`
	PDFTemplateVersion  int             `json:"pdf_template_version,omitempty"`
	PDFLogoPNG          []byte          `json:"pdf_logo_png,omitempty"`
	PDFFooter           string          `json:"pdf_footer,omitempty"`
	Month               string          `json:"month"`
	Timezone            string          `json:"timezone"`
	UserID              int             `json:"user_id"`
	Username            string          `json:"username,omitempty"`
	DisplayName         string          `json:"display_name,omitempty"`
	HistoryImportID     string          `json:"history_import_id,omitempty"`
	HistorySourceSHA256 string          `json:"history_source_sha256,omitempty"`
	CompanyTitle        string          `json:"company_title"`
	TaxID               string          `json:"tax_id"`
	Issuer              string          `json:"issuer"`
	AccountingStartAt   int64           `json:"accounting_start_at"`
	Currency            BillingCurrency `json:"currency"`
	Days                []BillingRow    `json:"days"`
	Total               BillingRow      `json:"total"`
	ChargeQuota         int64           `json:"charge_quota"`
	RefundQuota         int64           `json:"refund_quota"`
	RoundingDifference  string          `json:"rounding_difference"`
}
type BillingDayView struct {
	Date               string                `json:"date"`
	Timezone           string                `json:"timezone"`
	Account            *model.BillingAccount `json:"account"`
	Currency           BillingCurrency       `json:"currency"`
	Hours              []BillingRow          `json:"hours"`
	Total              BillingRow            `json:"total"`
	UpdatedAt          int64                 `json:"updated_at"`
	Available          bool                  `json:"available"`
	RoundingDifference string                `json:"rounding_difference"`
	Source             string                `json:"source"`
	LastLogAt          int64                 `json:"last_log_at"`
}

var billingLocation = time.FixedZone("Asia/Shanghai", 8*3600)

func currentBillingCurrency(formal bool) (BillingCurrency, error) {
	code := operation_setting.GetQuotaDisplayType()
	symbol := operation_setting.GetCurrencySymbol()
	rate := operation_setting.GetUsdToCurrencyRate(operation_setting.USDExchangeRate)
	unit := common.QuotaPerUnit
	if code == operation_setting.QuotaDisplayTypeTokens {
		if formal {
			code, symbol, rate = "USD", "$", 1
		} else {
			symbol, rate, unit = "TOKENS", 1, 1
		}
	}
	if rate <= 0 || unit <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) || math.IsNaN(unit) || math.IsInf(unit, 0) {
		return BillingCurrency{}, errors.New("invalid billing currency configuration")
	}
	return BillingCurrency{Code: code, Symbol: symbol, Rate: decimal.NewFromFloat(rate).String(), QuotaPerUnit: decimal.NewFromFloat(unit).String()}, nil
}

func billingRow(label string, charge, refund, count int64, currency BillingCurrency) (BillingRow, error) {
	rate, err := decimal.NewFromString(currency.Rate)
	if err != nil {
		return BillingRow{}, err
	}
	unit, err := decimal.NewFromString(currency.QuotaPerUnit)
	if err != nil || !unit.IsPositive() {
		return BillingRow{}, errors.New("invalid billing currency divisor")
	}
	chargeAmount := decimal.NewFromInt(charge).Mul(rate).Div(unit)
	refundAmount := decimal.NewFromInt(refund).Mul(rate).Div(unit)
	return BillingRow{Label: label, Charge: chargeAmount.StringFixed(6), Refund: refundAmount.StringFixed(6), Amount: chargeAmount.Sub(refundAmount).StringFixed(6), Count: count, ChargeQuota: charge, RefundQuota: refund}, nil
}

// The displayed total is rounded once from integer quota, not summed from
// already rounded rows. Surface the difference so readers can reconcile both.
func billingRoundingDifference(rows []BillingRow, total BillingRow) (string, error) {
	amount, err := decimal.NewFromString(total.Amount)
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		value, err := decimal.NewFromString(row.Amount)
		if err != nil {
			return "", err
		}
		amount = amount.Sub(value)
	}
	return amount.StringFixed(6), nil
}

func GetBillingDay(userID int, day string) (*BillingDayView, error) {
	return GetBillingDayContext(context.Background(), userID, day)
}

func GetBillingDayContext(ctx context.Context, userID int, day string) (*BillingDayView, error) {
	if day == "" {
		day = time.Now().In(billingLocation).Format("2006-01-02")
	}
	start, err := time.ParseInLocation("2006-01-02", day, billingLocation)
	if err != nil || start.Format("2006-01-02") != day || start.Unix() < 0 || start.After(time.Now()) {
		return nil, errors.New("invalid billing date")
	}
	account, err := model.GetBillingAccount(userID)
	if err != nil {
		return nil, err
	}
	currency, err := currentBillingCurrency(false)
	if err != nil {
		return nil, err
	}
	window, err := model.GetBillingUsageHours(ctx, userID, start.Unix(), start.AddDate(0, 0, 1).Unix())
	if err != nil {
		return nil, err
	}
	byHour := make(map[int64]model.BillingUsageHour, len(window.Hours))
	var charge, refund, count, lastLogAt int64
	for _, hour := range window.Hours {
		if hour.Charge > int64(common.MaxWalletQuota)-charge || hour.Refund > int64(common.MaxWalletQuota)-refund {
			return nil, errors.New("daily usage quota exceeds exact limit")
		}
		byHour[hour.Hour] = hour
		charge += hour.Charge
		refund += hour.Refund
		count += hour.Count
		if hour.LastLogAt > lastLogAt {
			lastLogAt = hour.LastLogAt
		}
	}
	view := BillingDayView{Date: day, Timezone: "Asia/Shanghai", Account: account, Currency: currency, UpdatedAt: window.CheckedAt, Hours: make([]BillingRow, 0, 24), Source: "usage_logs", LastLogAt: lastLogAt, Available: true}
	for hour := 0; hour < 24; hour++ {
		at := start.Add(time.Duration(hour) * time.Hour)
		value := byHour[at.Unix()]
		row, err := billingRow(at.Format("15:04"), value.Charge, value.Refund, value.Count, currency)
		if err != nil {
			return nil, err
		}
		if at.Unix() > window.CheckedAt {
			row.State = "future"
		} else if at.Add(time.Hour).Unix() > window.CheckedAt {
			row.State = "in_progress"
		}
		view.Hours = append(view.Hours, row)
	}
	view.Total, err = billingRow(day, charge, refund, count, currency)
	if err != nil {
		return nil, err
	}
	view.RoundingDifference, err = billingRoundingDifference(view.Hours, view.Total)
	return &view, err
}

func BuildBillingSnapshot(account *model.BillingAccount, month string, hours []model.BillingHour) (*BillingSnapshot, error) {
	currency, err := currentBillingCurrency(true)
	if err != nil {
		return nil, err
	}
	return buildBillingSnapshot(account, month, hours, currency)
}

func buildBillingSnapshot(account *model.BillingAccount, month string, hours []model.BillingHour, currency BillingCurrency) (*BillingSnapshot, error) {
	start, end, err := model.BillingMonthBounds(month)
	if err != nil {
		return nil, err
	}
	snapshot := BillingSnapshot{SchemaVersion: 1, Month: month, Timezone: "Asia/Shanghai", UserID: account.UserID, CompanyTitle: account.CompanyTitle, TaxID: account.TaxID, Issuer: common.SystemName, AccountingStartAt: account.AccountingStartAt, Currency: currency, Days: make([]BillingRow, 0, 31)}
	type totals struct{ charge, refund, count int64 }
	byDay := make(map[string]totals)
	var all totals
	for _, hour := range hours {
		if hour.Charge < 0 || hour.Refund < 0 || hour.Charge > int64(common.MaxWalletQuota)-all.charge || hour.Refund > int64(common.MaxWalletQuota)-all.refund {
			return nil, errors.New("monthly billing quota exceeds exact limit")
		}
		day := time.Unix(hour.Hour, 0).In(billingLocation).Format("2006-01-02")
		sum := byDay[day]
		sum.charge += hour.Charge
		sum.refund += hour.Refund
		sum.count += hour.Count
		byDay[day] = sum
		all.charge += hour.Charge
		all.refund += hour.Refund
		all.count += hour.Count
	}
	for at := time.Unix(start, 0).In(billingLocation); at.Unix() < end; at = at.AddDate(0, 0, 1) {
		day := at.Format("2006-01-02")
		sum := byDay[day]
		row, err := billingRow(day, sum.charge, sum.refund, sum.count, currency)
		if err != nil {
			return nil, err
		}
		if at.AddDate(0, 0, 1).Unix() <= account.AccountingStartAt {
			row.State = "outside_period"
		}
		snapshot.Days = append(snapshot.Days, row)
	}
	snapshot.ChargeQuota, snapshot.RefundQuota = all.charge, all.refund
	snapshot.Total, err = billingRow(month, all.charge, all.refund, all.count, currency)
	if err != nil {
		return nil, err
	}
	snapshot.RoundingDifference, err = billingRoundingDifference(snapshot.Days, snapshot.Total)
	return &snapshot, err
}

func PrepareBillingStatement(userID, actorID, storageProfileID int, month string) (*model.BillingStatement, error) {
	return PrepareBillingStatementContext(context.Background(), userID, actorID, storageProfileID, month)
}

func PrepareBillingStatementContext(ctx context.Context, userID, actorID, storageProfileID int, month string) (*model.BillingStatement, error) {
	if storageProfileID <= 0 {
		return nil, errors.New("billing archive storage profile is required")
	}
	start, end, err := model.BillingMonthBounds(month)
	if err != nil {
		return nil, err
	}
	account, err := model.GetBillingAccount(userID)
	if err != nil {
		return nil, err
	}
	if account.AccountingStartAt > start {
		start = account.AccountingStartAt
	}
	hasHistory := false
	if start < end {
		hasHistory, err = model.HasBillingUsageRecords(ctx, userID, start, end)
		if err != nil {
			return nil, err
		}
	}
	identities, err := model.GetBillingUserIdentities(ctx, []int{userID})
	if err != nil {
		return nil, err
	}
	customer, exists := identities[userID]
	if !exists {
		return nil, errors.New("billing customer does not exist")
	}
	imports, err := model.ListBillingHistoryImports(userID, month)
	if err != nil {
		return nil, err
	}
	branding, err := captureBillingDocumentBranding(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBillingDocumentBranding, err)
	}
	statement := &model.BillingStatement{ID: common.GetUUID(), UserID: userID, CreatedBy: actorID, StorageProfileID: storageProfileID, Month: month}
	err = model.CreateBillingStatement(statement, func(locked *model.BillingAccount, hours []model.BillingHour) (string, string, error) {
		if locked.AccountingStartAt != account.AccountingStartAt || locked.ProfileVersion != account.ProfileVersion {
			return "", "", model.ErrBillingConflict
		}
		snapshot, err := BuildBillingSnapshot(locked, month, hours)
		if err != nil {
			return "", "", err
		}
		if snapshot.Total.Count == 0 && hasHistory {
			return "", "", ErrBillingHistoricalDataUnreconciled
		}
		snapshot.PDFTemplateVersion = 3
		snapshot.Issuer, snapshot.PDFLogoPNG, snapshot.PDFFooter = branding.Issuer, branding.LogoPNG, branding.Footer
		snapshot.Username, snapshot.DisplayName = customer.Username, customer.DisplayName
		if len(imports) > 0 {
			snapshot.HistoryImportID = imports[0].ID
			snapshot.HistorySourceSHA256 = imports[0].SourceSHA256
		}
		body, err := common.Marshal(snapshot)
		if err != nil {
			return "", "", err
		}
		if len(body) > 60*1024 {
			return "", "", errors.New("billing document snapshot exceeds the safe database TEXT limit")
		}
		hash := sha256.Sum256(body)
		return string(body), hex.EncodeToString(hash[:]), nil
	})
	return statement, err
}
