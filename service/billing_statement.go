package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
)

const billingDiscountAmountNote = "本对账单中消费金额、退款金额、净消费金额已包含计费折扣，各项金额为折扣后的费用"

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
	Cost        string `json:"cost,omitempty"`
	Profit      string `json:"profit,omitempty"`
	CostQuota   int64  `json:"cost_quota,omitempty"`
	ProfitQuota int64  `json:"profit_quota,omitempty"`
	State       string `json:"state,omitempty"`
}
type BillingModelRow struct {
	ModelName    string             `json:"model_name"`
	BillingGroup string             `json:"billing_group,omitempty"`
	BillingRate  string             `json:"billing_rate,omitempty"`
	Charge       string             `json:"charge"`
	Refund       string             `json:"refund"`
	Amount       string             `json:"amount"`
	ChargeShare  string             `json:"charge_share,omitempty"`
	Count        int64              `json:"count"`
	ChargeCount  int64              `json:"charge_count,omitempty"`
	RefundCount  int64              `json:"refund_count,omitempty"`
	ActiveDays   int64              `json:"active_days,omitempty"`
	FirstPosted  int64              `json:"first_posted_at,omitempty"`
	LastPosted   int64              `json:"last_posted_at,omitempty"`
	ChargeQuota  int64              `json:"charge_quota"`
	RefundQuota  int64              `json:"refund_quota"`
	CurrentRates []BillingGroupRate `json:"current_rates,omitempty"`
}
type BillingGroupRate struct {
	Group string `json:"group"`
	Ratio string `json:"ratio"`
}
type BillingSnapshot struct {
	AccountingBasis     string            `json:"accounting_basis,omitempty"`
	SchemaVersion       int               `json:"schema_version"`
	PDFTemplateVersion  int               `json:"pdf_template_version,omitempty"`
	PDFLogoPNG          []byte            `json:"pdf_logo_png,omitempty"`
	PDFOperatingLogoPNG []byte            `json:"pdf_operating_logo_png,omitempty"`
	OperatingName       string            `json:"operating_name,omitempty"`
	PDFFooter           string            `json:"pdf_footer,omitempty"`
	Month               string            `json:"month"`
	Timezone            string            `json:"timezone"`
	UserID              int               `json:"user_id"`
	Username            string            `json:"username,omitempty"`
	DisplayName         string            `json:"display_name,omitempty"`
	HistoryImportID     string            `json:"history_import_id,omitempty"`
	HistorySourceSHA256 string            `json:"history_source_sha256,omitempty"`
	CompanyTitle        string            `json:"company_title"`
	TaxID               string            `json:"tax_id"`
	Issuer              string            `json:"issuer"`
	AccountingStartAt   int64             `json:"accounting_start_at"`
	Currency            BillingCurrency   `json:"currency"`
	Days                []BillingRow      `json:"days"`
	Models              []BillingModelRow `json:"models,omitempty"`
	Total               BillingRow        `json:"total"`
	ChargeQuota         int64             `json:"charge_quota"`
	RefundQuota         int64             `json:"refund_quota"`
	RoundingDifference  string            `json:"rounding_difference"`
	RateReferenceAt     int64             `json:"rate_reference_at,omitempty"`
	// DetailSheet adds the per-record 详细日志 sheet to the Excel workbook.
	// Snapshots frozen before it existed keep rendering byte-identical files.
	DetailSheet bool `json:"detail_sheet,omitempty"`
}

func attachBillingModelRows(snapshot *BillingSnapshot, totals []model.BillingModelTotal) error {
	if snapshot == nil {
		return errors.New("billing snapshot is required")
	}
	rows := make([]BillingModelRow, 0, len(totals))
	var charge, refund, count int64
	for _, total := range totals {
		if total.Charge < 0 || total.Refund < 0 || total.Count < 0 ||
			total.ChargeCount < 0 || total.RefundCount < 0 || total.ChargeCount > math.MaxInt64-total.RefundCount ||
			total.ChargeCount+total.RefundCount != total.Count ||
			total.ActiveDays <= 0 || total.ActiveDays > total.Count || total.FirstPosted <= 0 || total.LastPosted < total.FirstPosted ||
			total.Charge > int64(common.MaxWalletQuota)-charge ||
			total.Refund > int64(common.MaxWalletQuota)-refund ||
			total.Count > math.MaxInt64-count {
			return errors.New("model billing quota exceeds exact limit")
		}
		amounts, err := billingRow(total.ModelName, total.Charge, total.Refund, total.Count, snapshot.Currency)
		if err != nil {
			return err
		}
		chargeShare := "0.00%"
		if snapshot.ChargeQuota > 0 {
			chargeShare = decimal.NewFromInt(total.Charge).Mul(decimal.NewFromInt(100)).Div(decimal.NewFromInt(snapshot.ChargeQuota)).StringFixed(2) + "%"
		}
		rows = append(rows, BillingModelRow{
			ModelName: total.ModelName, BillingGroup: total.BillingGroup, BillingRate: total.BillingRate, Charge: amounts.Charge, Refund: amounts.Refund,
			Amount: amounts.Amount, ChargeShare: chargeShare, Count: amounts.Count,
			ChargeCount: total.ChargeCount, RefundCount: total.RefundCount,
			ActiveDays: total.ActiveDays, FirstPosted: total.FirstPosted, LastPosted: total.LastPosted,
			ChargeQuota: amounts.ChargeQuota, RefundQuota: amounts.RefundQuota,
		})
		charge += total.Charge
		refund += total.Refund
		count += total.Count
	}
	if charge != snapshot.ChargeQuota || refund != snapshot.RefundQuota || count != snapshot.Total.Count {
		return errors.New("model billing totals do not reconcile with hourly snapshot")
	}
	sort.Slice(rows, func(i, j int) bool {
		left := rows[i].ChargeQuota - rows[i].RefundQuota
		right := rows[j].ChargeQuota - rows[j].RefundQuota
		if left != right {
			return left > right
		}
		if rows[i].ModelName != rows[j].ModelName {
			return rows[i].ModelName < rows[j].ModelName
		}
		if rows[i].BillingGroup != rows[j].BillingGroup {
			return rows[i].BillingGroup < rows[j].BillingGroup
		}
		return rows[i].BillingRate < rows[j].BillingRate
	})
	snapshot.Models = rows
	return nil
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

func attachBillingAccounting(row *BillingRow, costQuota int64, currency BillingCurrency) error {
	rate, err := decimal.NewFromString(currency.Rate)
	if err != nil {
		return err
	}
	unit, err := decimal.NewFromString(currency.QuotaPerUnit)
	if err != nil || !unit.IsPositive() {
		return errors.New("invalid billing currency divisor")
	}
	row.CostQuota = costQuota
	row.ProfitQuota = row.ChargeQuota - row.RefundQuota - costQuota
	row.Cost = decimal.NewFromInt(costQuota).Mul(rate).Div(unit).StringFixed(6)
	row.Profit = decimal.NewFromInt(row.ProfitQuota).Mul(rate).Div(unit).StringFixed(6)
	return nil
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
	return GetBillingDayContextWithAccounting(ctx, userID, day, false)
}

func GetBillingDayContextWithAccounting(ctx context.Context, userID int, day string, includeAccounting bool) (*BillingDayView, error) {
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
	if err != nil || !includeAccounting {
		return &view, err
	}
	accountingFilter := model.CostAccountingFilter{
		StartTimestamp: start.Unix(),
		EndTimestamp:   start.AddDate(0, 0, 1).Unix() - 1,
		UserID:         userID,
	}
	buckets, err := model.SumCostAccountingBuckets(accountingFilter, 3600, 8*3600)
	if err != nil {
		return nil, err
	}
	byAccountingHour := make(map[int64]model.CostAccountingBucket, len(buckets))
	for _, bucket := range buckets {
		byAccountingHour[bucket.Bucket] = bucket
	}
	for hour := range view.Hours {
		bucket := byAccountingHour[start.Add(time.Duration(hour)*time.Hour).Unix()]
		if err := attachBillingAccounting(&view.Hours[hour], bucket.CostQuota, currency); err != nil {
			return nil, err
		}
	}
	totals, err := model.SumCostAccounting(accountingFilter)
	if err != nil {
		return nil, err
	}
	if err := attachBillingAccounting(&view.Total, totals.CostQuota, currency); err != nil {
		return nil, err
	}
	return &view, err
}

func EnrichBillingSnapshotAccounting(snapshot *BillingSnapshot, userID int) error {
	if snapshot == nil || userID <= 0 {
		return nil
	}
	start, end, err := model.BillingMonthBounds(snapshot.Month)
	if err != nil {
		return err
	}
	filter := model.CostAccountingFilter{StartTimestamp: start, EndTimestamp: end - 1, UserID: userID}
	buckets, err := model.SumCostAccountingBuckets(filter, 86400, 8*3600)
	if err != nil {
		return err
	}
	byDay := make(map[int64]model.CostAccountingBucket, len(buckets))
	for _, bucket := range buckets {
		byDay[bucket.Bucket] = bucket
	}
	for i := range snapshot.Days {
		day, parseErr := time.ParseInLocation("2006-01-02", snapshot.Days[i].Label, billingLocation)
		if parseErr != nil {
			return parseErr
		}
		bucket := byDay[day.Unix()]
		if err := attachBillingAccounting(&snapshot.Days[i], bucket.CostQuota, snapshot.Currency); err != nil {
			return err
		}
	}
	totals, err := model.SumCostAccounting(filter)
	if err != nil {
		return err
	}
	return attachBillingAccounting(&snapshot.Total, totals.CostQuota, snapshot.Currency)
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
		return nil, fmt.Errorf("%w: %w", ErrBillingDocumentBranding, err)
	}
	statement := &model.BillingStatement{ID: common.GetUUID(), UserID: userID, CreatedBy: actorID, StorageProfileID: storageProfileID, Month: month}
	err = model.CreateBillingStatement(statement, func(locked *model.BillingAccount, hours []model.BillingHour, modelTotals []model.BillingModelTotal) (string, string, error) {
		if locked.AccountingStartAt != account.AccountingStartAt || locked.ProfileVersion != account.ProfileVersion {
			return "", "", model.ErrBillingConflict
		}
		snapshot, err := BuildBillingSnapshot(locked, month, hours)
		if err != nil {
			return "", "", err
		}
		if err := attachBillingModelRows(snapshot, modelTotals); err != nil {
			return "", "", err
		}
		if snapshot.Total.Count == 0 && hasHistory {
			return "", "", ErrBillingHistoricalDataUnreconciled
		}
		snapshot.PDFTemplateVersion = 18
		snapshot.DetailSheet = true
		snapshot.AccountingBasis = model.BillingSourcePeriodCorrections
		snapshot.Issuer, snapshot.PDFLogoPNG, snapshot.PDFFooter = branding.Issuer, branding.LogoPNG, branding.Footer
		snapshot.OperatingName, snapshot.PDFOperatingLogoPNG = branding.OperatingName, branding.OperatingLogoPNG
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
	}, true)
	return statement, err
}
