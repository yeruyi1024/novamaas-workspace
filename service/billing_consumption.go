package service

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"
)

type BillingCheck struct {
	Code   string `json:"code"`
	Passed bool   `json:"passed"`
}

type BillingReadiness struct {
	Ready             bool           `json:"ready"`
	Status            string         `json:"status"`
	Checks            []BillingCheck `json:"checks"`
	AccountingStartAt int64          `json:"accounting_start_at"`
	PeriodStartAt     int64          `json:"period_start_at"`
	PeriodEndAt       int64          `json:"period_end_at"`
	PrepareAt         int64          `json:"prepare_at"`
	Pending           int64          `json:"pending"`
	ExistingStatement string         `json:"existing_statement"`
}

type BillingUsageMonth struct {
	*BillingSnapshot
	Source    string `json:"source"`
	UpdatedAt int64  `json:"updated_at"`
	LastLogAt int64  `json:"last_log_at"`
}

type BillingMonthPreview struct {
	Reference *BillingUsageMonth `json:"reference"`
	Formal    *BillingSnapshot   `json:"formal"`
	Readiness BillingReadiness   `json:"readiness"`
}

var ErrBillingHistoricalDataUnreconciled = errors.New("historical usage exists without formal billing entries; reconcile history before issuing a statement")

func GetBillingMonthPreview(ctx context.Context, userID int, month string, storageProfileID int) (*BillingMonthPreview, error) {
	start, end, err := model.BillingMonthBounds(month)
	if err != nil || start < 0 || start > common.GetTimestamp() {
		return nil, errors.New("invalid billing month")
	}
	window, err := model.GetBillingUsageHours(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}
	hours := make([]model.BillingHour, 0, len(window.Hours))
	var lastLogAt int64
	for _, hour := range window.Hours {
		hours = append(hours, model.BillingHour{Hour: hour.Hour, Charge: hour.Charge, Refund: hour.Refund, Count: hour.Count})
		if hour.LastLogAt > lastLogAt {
			lastLogAt = hour.LastLogAt
		}
	}
	currency, err := currentBillingCurrency(false)
	if err != nil {
		return nil, err
	}
	reference, err := buildBillingSnapshot(&model.BillingAccount{UserID: userID}, month, hours, currency)
	if err != nil {
		return nil, err
	}
	for i := range reference.Days {
		day, _ := time.ParseInLocation("2006-01-02", reference.Days[i].Label, billingLocation)
		if day.Unix() > window.CheckedAt {
			reference.Days[i].State = "future"
		} else if day.AddDate(0, 0, 1).Unix() > window.CheckedAt {
			reference.Days[i].State = "in_progress"
		}
	}
	state, err := model.GetBillingPreparationState(ctx, userID, start, end)
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	account := &state.Account
	configured := account.AccountingStartAt > 0
	covered := configured && account.AccountingStartAt < end && account.AccountingStartAt <= now
	readiness := BillingReadiness{Ready: true, Status: "ready", AccountingStartAt: account.AccountingStartAt, PeriodStartAt: start, PeriodEndAt: end, PrepareAt: end + 86400, Pending: state.Pending, ExistingStatement: state.ExistingStatement}
	if account.AccountingStartAt > start {
		readiness.PeriodStartAt = account.AccountingStartAt
	}
	result := &BillingMonthPreview{Reference: &BillingUsageMonth{BillingSnapshot: reference, Source: "usage_logs", UpdatedAt: window.CheckedAt, LastLogAt: lastLogAt}}
	validSummary := !state.MissingHours
	if covered && validSummary {
		result.Formal, err = BuildBillingSnapshot(account, month, state.Hours)
		if err != nil {
			validSummary = false
			common.SysError("billing preview summary validation failed: " + err.Error())
		}
	}
	historyReconciled := true
	if result.Formal != nil && result.Formal.Total.Count == 0 {
		hasHistory, queryErr := model.HasBillingUsageRecords(ctx, userID, readiness.PeriodStartAt, end)
		if queryErr != nil {
			return nil, queryErr
		}
		historyReconciled = !hasHistory
		if !historyReconciled {
			result.Formal = nil
		}
	}
	storageCode, storageReady := "storage_missing", false
	if storageProfileID > 0 {
		storageCode = "storage_ready"
		_, storageErr := storageService.OpenBillingArchiveStore(storageProfileID, true)
		storageReady = storageErr == nil
		if !storageReady {
			storageCode = "storage_unavailable"
		}
	}
	readiness.Checks = []BillingCheck{
		{Code: "accounting_configured", Passed: configured},
		{Code: "period_covered", Passed: covered},
		{Code: "identity_complete", Passed: account.CompanyTitle != "" && account.TaxID != ""},
		{Code: "month_closed", Passed: now >= readiness.PrepareAt},
		{Code: "operations_settled", Passed: state.Pending == 0},
		{Code: "summary_valid", Passed: validSummary},
		{Code: "history_reconciled", Passed: historyReconciled},
		{Code: "no_active_statement", Passed: state.ExistingStatement == ""},
		{Code: storageCode, Passed: storageReady},
	}
	for _, check := range readiness.Checks {
		if !check.Passed {
			readiness.Ready = false
		}
	}
	switch {
	case !configured:
		readiness.Status = "not_configured"
	case !covered:
		readiness.Status = "outside_period"
	case !validSummary:
		readiness.Status = "data_error"
	case !historyReconciled:
		readiness.Status = "historical_data_unreconciled"
	case result.Formal.Total.Count == 0:
		readiness.Status = "no_consumption"
	case !readiness.Ready:
		readiness.Status = "blocked"
	}
	result.Readiness = readiness
	return result, nil
}

// Used again at issue/confirmation, including for drafts created by older
// versions. Never repairs or overwrites the statement's frozen evidence.
func ValidateBillingStatementSource(ctx context.Context, statement *model.BillingStatement) error {
	if err := statement.VerifySnapshot(); err != nil {
		return err
	}
	var snapshot BillingSnapshot
	if err := common.UnmarshalJsonStr(statement.Snapshot, &snapshot); err != nil {
		return err
	}
	if snapshot.Total.Count > 0 {
		return nil
	}
	hasHistory, err := model.HasBillingUsageRecords(ctx, statement.UserID, statement.StartAt, statement.EndAt)
	if err != nil {
		return err
	}
	if hasHistory {
		return ErrBillingHistoricalDataUnreconciled
	}
	return nil
}

type BillingUsageRecord struct {
	ID        int    `json:"id"`
	CreatedAt int64  `json:"created_at"`
	RequestID string `json:"request_id"`
	ModelName string `json:"model_name"`
	Kind      string `json:"kind"`
	Amount    string `json:"amount"`
}

type BillingUsageDetails struct {
	Items      []BillingUsageRecord `json:"items"`
	NextCursor string               `json:"next_cursor"`
	Currency   BillingCurrency      `json:"currency"`
}

func GetBillingUsageDetails(ctx context.Context, userID int, date string, hour int, cursor string) (*BillingUsageDetails, error) {
	start, err := time.ParseInLocation("2006-01-02", date, billingLocation)
	if err != nil || start.Format("2006-01-02") != date || start.Unix() < 0 || start.After(time.Now()) || hour < 0 || hour > 23 {
		return nil, errors.New("invalid usage billing hour")
	}
	start = start.Add(time.Duration(hour) * time.Hour)
	var before *model.BillingUsageCursor
	if cursor != "" {
		if len(cursor) > 512 {
			return nil, errors.New("invalid usage billing cursor")
		}
		body, err := base64.RawURLEncoding.DecodeString(cursor)
		before = &model.BillingUsageCursor{}
		if err != nil || common.Unmarshal(body, before) != nil || before.At < start.Unix() || before.At >= start.Add(time.Hour).Unix() || before.ID < 0 || len(before.RequestID) > 128 {
			return nil, errors.New("invalid usage billing cursor")
		}
	}
	logs, err := model.GetBillingUsageRecords(ctx, userID, start.Unix(), start.Add(time.Hour).Unix(), before)
	if err != nil {
		return nil, err
	}
	currency, err := currentBillingCurrency(false)
	if err != nil {
		return nil, err
	}
	result := &BillingUsageDetails{Items: make([]BillingUsageRecord, 0), Currency: currency}
	if len(logs) > 100 {
		logs = logs[:100]
		last := logs[len(logs)-1]
		body, err := common.Marshal(model.BillingUsageCursor{At: last.CreatedAt, ID: last.Id, RequestID: last.RequestId})
		if err != nil {
			return nil, err
		}
		result.NextCursor = base64.RawURLEncoding.EncodeToString(body)
	}
	for _, log := range logs {
		if log.Quota < 0 || int64(log.Quota) > int64(common.MaxWalletQuota) {
			return nil, errors.New("usage billing data contains invalid quota")
		}
		kind, charge, refund := "consumption", int64(log.Quota), int64(0)
		if log.Type == model.LogTypeRefund {
			kind, charge, refund = "refund", 0, int64(log.Quota)
		}
		row, err := billingRow("", charge, refund, 1, currency)
		if err != nil {
			return nil, err
		}
		result.Items = append(result.Items, BillingUsageRecord{ID: log.Id, CreatedAt: log.CreatedAt, RequestID: log.RequestId, ModelName: log.ModelName, Kind: kind, Amount: row.Amount})
	}
	return result, nil
}
