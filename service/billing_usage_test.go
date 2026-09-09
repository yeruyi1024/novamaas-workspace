package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBillingDayShowsHistoricalConsumptionWithoutFormalAccounting(t *testing.T) {
	truncate(t)
	start := time.Date(2020, 2, 3, 0, 0, 0, 0, billingLocation).Unix()
	logs := []model.Log{
		{UserId: 81, CreatedAt: start + 9*3600, Type: model.LogTypeConsume, Quota: 500000, RequestId: "historical-charge"},
		{UserId: 81, CreatedAt: start + 9*3600 + 1, Type: model.LogTypeRefund, Quota: 100000, RequestId: "historical-refund"},
		{UserId: 81, CreatedAt: start + 10*3600, Type: model.LogTypeConsume, Quota: 250000, RequestId: "historical-charge-two"},
		{UserId: 82, CreatedAt: start + 9*3600, Type: model.LogTypeConsume, Quota: 999999},
		{UserId: 81, CreatedAt: start + 9*3600, Type: model.LogTypeTopup, Quota: 999999},
		{UserId: 81, CreatedAt: start + 86400, Type: model.LogTypeConsume, Quota: 999999},
	}
	require.NoError(t, model.LOG_DB.Create(&logs).Error)
	view, err := GetBillingDay(81, "2020-02-03")
	require.NoError(t, err)
	assert.True(t, view.Available, "historical consumption must not require formal-account activation")
	require.Len(t, view.Hours, 24)
	assert.Equal(t, int64(500000), view.Hours[9].ChargeQuota)
	assert.Equal(t, int64(100000), view.Hours[9].RefundQuota)
	assert.Equal(t, int64(3), view.Total.Count)
	assert.Equal(t, int64(750000), view.Total.ChargeQuota)
	assert.Equal(t, int64(100000), view.Total.RefundQuota)
	var accounts int64
	require.NoError(t, model.DB.Model(&model.BillingAccount{}).Count(&accounts).Error)
	assert.Zero(t, accounts, "a reference query must not create formal accounting data")
}

func TestBillingUsageDetailsPaginatesEqualTimestampsAndRedactsPrivateFields(t *testing.T) {
	truncate(t)
	start := time.Date(2020, 2, 3, 9, 0, 0, 0, billingLocation).Unix()
	// The 101st record exercises the real 100-row API cursor boundary.
	logs := make([]model.Log, 101)
	for i := range logs {
		logs[i] = model.Log{UserId: 83, CreatedAt: start, Type: model.LogTypeConsume, Quota: 500000, RequestId: fmt.Sprintf("page-%03d", i), ModelName: "example", Content: "private-payload", Other: `{"admin_info":"private-admin"}`, Ip: "192.0.2.1"}
	}
	logs[100].Type = model.LogTypeRefund
	require.NoError(t, model.LOG_DB.Create(&logs).Error)
	first, err := GetBillingUsageDetails(context.Background(), 83, "2020-02-03", 9, "")
	require.NoError(t, err)
	require.Len(t, first.Items, 100)
	assert.Equal(t, "page-100", first.Items[0].RequestID)
	assert.Equal(t, "refund", first.Items[0].Kind)
	assert.True(t, len(first.Items[0].Amount) > 0 && first.Items[0].Amount[0] == '-', "refund must be shown as a negative net amount")
	require.NotEmpty(t, first.NextCursor)
	second, err := GetBillingUsageDetails(context.Background(), 83, "2020-02-03", 9, first.NextCursor)
	require.NoError(t, err)
	require.Len(t, second.Items, 1)
	assert.Equal(t, "page-000", second.Items[0].RequestID)
	assert.Empty(t, second.NextCursor)
	body, err := common.Marshal(first)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "private-")
	assert.NotContains(t, string(body), "192.0.2.1")
	_, err = GetBillingUsageDetails(context.Background(), 83, "2020-02-03", 10, first.NextCursor)
	assert.ErrorContains(t, err, "cursor", "cursor cannot escape its hour")
	other, err := GetBillingUsageDetails(context.Background(), 84, "2020-02-03", 9, first.NextCursor)
	require.NoError(t, err)
	assert.Empty(t, other.Items, "cursor does not override account scope")
}

func TestBillingUsageRejectsInvalidQuotaInsteadOfDisplayingZero(t *testing.T) {
	truncate(t)
	start := time.Date(2020, 2, 3, 9, 0, 0, 0, billingLocation).Unix()
	require.NoError(t, model.LOG_DB.Create(&model.Log{UserId: 85, CreatedAt: start, Type: model.LogTypeConsume, Quota: -1}).Error)
	view, err := GetBillingDay(85, "2020-02-03")
	require.ErrorContains(t, err, "invalid quota")
	assert.Nil(t, view)
	_, err = GetBillingUsageDetails(context.Background(), 85, "2020-02-03", 9, "")
	assert.ErrorContains(t, err, "invalid quota")
}

func TestBillingMonthReadinessKeepsFormalCutoffAndInvalidDataDistinct(t *testing.T) {
	truncate(t)
	start, end, err := model.BillingMonthBounds("2020-02")
	require.NoError(t, err)
	for _, scenario := range []struct {
		name    string
		userID  int
		cutoff  int64
		missing bool
		status  string
	}{
		{"unconfigured", 86, 0, false, "not_configured"},
		{"outside period", 87, end, false, "outside_period"},
		{"missing summaries", 88, start, true, "data_error"},
		{"history is not an empty formal period", 89, start, false, "historical_data_unreconciled"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			require.NoError(t, model.DB.Create(&model.BillingAccount{UserID: scenario.userID, AccountingStartAt: scenario.cutoff, StartSequence: 1, CompanyTitle: "Test", TaxID: "TEST"}).Error)
			require.NoError(t, model.LOG_DB.Create(&model.Log{UserId: scenario.userID, CreatedAt: start + 3600, Type: model.LogTypeConsume, Quota: 500000}).Error)
			if scenario.missing {
				require.NoError(t, model.DB.Create(&model.BillingEntry{UserID: scenario.userID, EventKey: "missing-hour", Sequence: 1, PostedAt: start + 3600, Kind: "usage", Quota: 500000}).Error)
			}
			view, err := GetBillingMonthPreview(context.Background(), scenario.userID, "2020-02", 0)
			require.NoError(t, err)
			assert.Equal(t, scenario.status, view.Readiness.Status)
			assert.False(t, view.Readiness.Ready, "storage has not been selected")
			assert.Equal(t, int64(1), view.Reference.Total.Count)
			if scenario.status == "no_consumption" {
				require.NotNil(t, view.Formal)
				assert.Zero(t, view.Formal.Total.Count)
			} else {
				assert.Nil(t, view.Formal)
			}
		})
	}
}

func TestBillingEmptyPeriodChecksTheExactAccountingCutoff(t *testing.T) {
	truncate(t)
	start, _, err := model.BillingMonthBounds("2020-03")
	require.NoError(t, err)
	cutoff := start + 1800
	require.NoError(t, model.DB.Create(&model.BillingAccount{UserID: 95, AccountingStartAt: cutoff, StartSequence: 1, CompanyTitle: "Test", TaxID: "TEST"}).Error)
	require.NoError(t, model.LOG_DB.Create(&model.Log{UserId: 95, CreatedAt: cutoff - 1, Type: model.LogTypeConsume, Quota: 500000}).Error)
	view, err := GetBillingMonthPreview(context.Background(), 95, "2020-03", 0)
	require.NoError(t, err)
	assert.Equal(t, "no_consumption", view.Readiness.Status, "history before the exact cutoff must not block a genuine empty period")
	require.NotNil(t, view.Formal)
	assert.Zero(t, view.Formal.Total.Count)
}

func TestBillingMonthReadinessPreservesPartialMonthAndBlocksDuplicate(t *testing.T) {
	truncate(t)
	start, _, err := model.BillingMonthBounds("2020-02")
	require.NoError(t, err)
	cutoff := start + 2*86400 + 1800
	require.NoError(t, model.DB.Create(&model.BillingAccount{UserID: 90, AccountingStartAt: cutoff, StartSequence: 1, CompanyTitle: "Test", TaxID: "TEST"}).Error)
	require.NoError(t, model.DB.Create(&model.BillingHour{UserID: 90, Hour: cutoff / 3600 * 3600, Charge: 500000, Count: 1}).Error)
	require.NoError(t, model.DB.Create(&model.BillingStatement{ID: "already-created", UserID: 90, Month: "2020-02", Status: model.StatementDraft}).Error)
	view, err := GetBillingMonthPreview(context.Background(), 90, "2020-02", 0)
	require.NoError(t, err)
	require.NotNil(t, view.Formal)
	assert.Equal(t, "outside_period", view.Formal.Days[0].State)
	assert.Equal(t, int64(1), view.Formal.Days[2].Count)
	assert.Equal(t, cutoff, view.Readiness.PeriodStartAt)
	assert.Equal(t, "already-created", view.Readiness.ExistingStatement)
	assert.Contains(t, view.Readiness.Checks, BillingCheck{Code: "no_active_statement", Passed: false})
}
