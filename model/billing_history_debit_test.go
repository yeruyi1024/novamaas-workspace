package model

import (
	"context"
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestReviewedDebitsRestoreSourcesAndCollectExactlyOnce(t *testing.T) {
	id := seedBillingCustomer(t)
	const month = "2025-10"
	start, end, err := BillingMonthBounds(month)
	require.NoError(t, err)
	require.NoError(t, DB.Model(&BillingAccount{}).Where("user_id = ?", id).Update("accounting_start_at", start).Error)
	operator := User{Id: 903, Username: "recovery-admin", AffCode: "test-debit", Role: common.RoleAdminUser, Status: common.UserStatusEnabled}
	require.NoError(t, DB.Create(&operator).Error)
	token := Token{UserId: id, Key: "recovery-test-token", RemainQuota: 10000}
	require.NoError(t, DB.Create(&token).Error)
	log := Log{UserId: id, Username: "billing_customer", Type: LogTypeConsume, Quota: 1234, RequestId: "request-a", UpstreamRequestId: "upstream-a", TokenId: token.Id, ModelName: "test", CreatedAt: start + 100, Other: `{"billing_source":"wallet"}`}
	require.NoError(t, DB.Create(&log).Error)
	a := ReviewedDebit{UserID: id, Username: log.Username, Month: month, Model: log.ModelName, TokenID: token.Id, OperationID: "op-a", RequestID: log.RequestId, UpstreamID: log.UpstreamRequestId, EndedAt: log.CreatedAt, LogID: log.Id, Quota: 1234}
	b := a
	b.OperationID = "op-b"
	b.RequestID = "request-b"
	b.UpstreamID = "upstream-b"
	b.LogID = 0
	b.Quota = 5000
	b.Supplier = &SupplierDebitEvidence{BaseCNY: "0.14", CostCNY: "0.035", SaleRatio: "0.5", BuyRatio: "0.25", Input: 10, Output: 20, Group: "test", Row: 3, LogID: "invoice-line", Sheet: "usage"}
	for _, c := range []ReviewedDebit{a, b} {
		require.NoError(t, DB.Create(&BillingOperation{ID: c.OperationID, UserID: id, State: "reserved", TokenID: token.Id, RequestID: c.RequestID, ModelName: c.Model, CreatedAt: start + 90}).Error)
		r, err := RecoverReviewedDebit(context.Background(), c, operator.Id, false)
		require.NoError(t, err)
		assert.Equal(t, "ready", r.Status)
	}
	var n int64
	require.NoError(t, DB.Model(&Log{}).Where("request_id = ?", b.RequestID).Count(&n).Error)
	assert.Zero(t, n)
	bad := b
	bad.UpstreamID = a.UpstreamID
	_, err = RecoverReviewedDebit(context.Background(), bad, operator.Id, true)
	require.ErrorIs(t, err, ErrBillingConflict)
	_, err = RecoverReviewedDebit(context.Background(), b, id, true)
	require.Error(t, err)
	// Fail after the supplemental log, cost and wallet updates. None may survive.
	require.NoError(t, DB.Callback().Update().Before("gorm:update").Register("test:fail_debit", func(db *gorm.DB) {
		if db.Statement.Table == "billing_operations" {
			db.AddError(errors.New("injected update failure"))
		}
	}))
	_, err = RecoverReviewedDebit(context.Background(), b, operator.Id, true)
	require.NoError(t, DB.Callback().Update().Remove("test:fail_debit"))
	require.Error(t, err)
	require.NoError(t, DB.Model(&CostAccountingSnapshot{}).Where("request_id = ?", b.RequestID).Count(&n).Error)
	assert.Zero(t, n)
	require.NoError(t, DB.Model(&Log{}).Where("request_id = ?", b.RequestID).Count(&n).Error)
	assert.Zero(t, n)
	for _, c := range []ReviewedDebit{b, a} {
		r, err := RecoverReviewedDebit(context.Background(), c, operator.Id, true)
		require.NoError(t, err)
		assert.Equal(t, "applied", r.Status)
		again, err := RecoverReviewedDebit(context.Background(), c, operator.Id, true)
		require.NoError(t, err)
		assert.Equal(t, "already_applied", again.Status)
		assert.Equal(t, r.EntryID, again.EntryID)
	}
	var reconstructed Log
	require.NoError(t, DB.Take(&reconstructed, "request_id = ?", b.RequestID).Error)
	assert.Equal(t, b.UpstreamID, reconstructed.UpstreamRequestId)
	assert.Equal(t, 10, reconstructed.PromptTokens)
	assert.Equal(t, b.EndedAt, reconstructed.CreatedAt)
	assert.Contains(t, reconstructed.Other, "supplier_statement")
	var cost CostAccountingSnapshot
	require.NoError(t, DB.Take(&cost, "source_log_id = ?", reconstructed.Id).Error)
	assert.Equal(t, int64(2500), cost.CostQuota)
	state, err := GetBillingHistoryState(context.Background(), id, month)
	require.NoError(t, err)
	assert.Zero(t, state.Pending)
	assert.Equal(t, end, state.Boundary)
	records, err := GetBillingHistoryRecords(context.Background(), id, month, start, state.Boundary)
	require.NoError(t, err)
	require.Len(t, records, 2)
	require.NoError(t, ApplyBillingHistoryImport(context.Background(), &BillingHistoryImport{ID: "hist", UserID: id, Month: month, SourceSHA256: "sha", ArtifactID: 1, SessionID: "session", Note: "reviewed"}, state.Account, records))
	hours, err := GetBillingHours(id, start, end)
	require.NoError(t, err)
	require.Len(t, hours, 1)
	assert.Equal(t, a.Quota+b.Quota, hours[0].Charge)
	balance, err := GetUserQuota(id, true)
	require.NoError(t, err)
	assert.Equal(t, 3766, balance)
	require.NoError(t, DB.First(&token, token.Id).Error)
	assert.Equal(t, 3766, token.RemainQuota)
	assert.Equal(t, 6234, token.UsedQuota)
}
