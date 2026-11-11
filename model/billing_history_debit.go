package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ReviewedDebit is read exclusively from the approved evidence file. A supplier
// row is a manually attributed source, never a live gateway observation.
type ReviewedDebit struct {
	UserID      int                    `json:"user_id"`
	Username    string                 `json:"username"`
	Month       string                 `json:"month"`
	OperationID string                 `json:"operation_id"`
	RequestID   string                 `json:"request_id"`
	TokenID     int                    `json:"token_id"`
	Model       string                 `json:"model"`
	LogID       int                    `json:"log_id"`
	Quota       int64                  `json:"quota"`
	EndedAt     int64                  `json:"ended_at"`
	UpstreamID  string                 `json:"upstream_request_id"`
	Supplier    *SupplierDebitEvidence `json:"supplier,omitempty"`
}

type SupplierDebitEvidence struct {
	FileSHA256 string `json:"file_sha256"`
	Sheet      string `json:"sheet"`
	Row        int    `json:"row"`
	LogID      string `json:"log_id"`
	BaseCNY    string `json:"base_cny"`
	CostCNY    string `json:"cost_cny"`
	BuyRatio   string `json:"buy_ratio"`
	SaleRatio  string `json:"sale_ratio"`
	Group      string `json:"group"`
	Input      int    `json:"input_tokens"`
	Output     int    `json:"output_tokens"`
	Seconds    int    `json:"seconds"`
	Stream     bool   `json:"stream"`
}

type DebitReceipt struct {
	Status  string `json:"status"`
	EntryID int64  `json:"entry_id,omitempty"`
}

// RecoverReviewedDebits reconciles the two approved operations under one wallet
// lock. A stale total or a failure on either operation rolls back both.
func RecoverReviewedDebits(ctx context.Context, cases []ReviewedDebit, actorID int, confirmQuota *int64) ([]DebitReceipt, int64, error) {
	if len(cases) != 2 || DB != LOG_DB {
		return nil, 0, ErrBillingHistoryBlocked
	}
	a, b := cases[0], cases[1]
	if a.UserID != b.UserID || a.Username != b.Username || a.Month != b.Month || a.TokenID != b.TokenID || a.OperationID == b.OperationID || a.RequestID == b.RequestID {
		return nil, 0, ErrBillingConflict
	}
	start, end, err := BillingMonthBounds(a.Month)
	if err != nil || common.GetTimestamp() < end+86400 {
		return nil, 0, ErrBillingHistoryBlocked
	}
	for _, c := range cases {
		if c.Quota <= 0 || c.Quota > common.MaxQuota || c.EndedAt < start || c.EndedAt >= end || c.UpstreamID == "" || (c.LogID == 0) != (c.Supplier != nil) {
			return nil, 0, ErrBillingHistoryBlocked
		}
	}
	if confirmQuota != nil && (*confirmQuota <= 0 || *confirmQuota > a.Quota+b.Quota) {
		return nil, 0, ErrBillingConflict
	}
	receipts := make([]DebitReceipt, len(cases))
	var total int64
	var tokenKey string
	err = DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var actor, user User
		if err := tx.Select("id", "role").Where("id = ? AND role >= ? AND status = ?", actorID, common.RoleAdminUser, common.UserStatusEnabled).Take(&actor).Error; err != nil {
			return err
		}
		if err := lockForUpdate(tx).Select("id", "role", "username", "quota").First(&user, a.UserID).Error; err != nil {
			return err
		}
		if user.Username != a.Username || (actor.Role < common.RoleRootUser && actor.Id != user.Id && user.Role >= actor.Role) {
			return ErrBillingConflict
		}
		originalBalance := int64(user.Quota)
		for i, c := range cases {
			if err := c.recover(tx, &user, actorID, start, end, confirmQuota != nil, &receipts[i], &tokenKey); err != nil {
				return err
			}
			if receipts[i].Status != "already_applied" {
				total += c.Quota
			}
		}
		if total > originalBalance && total > 0 {
			return ErrBillingInsufficientQuota
		}
		if confirmQuota != nil && total > 0 && *confirmQuota != total {
			return ErrBillingConflict
		}
		if confirmQuota != nil && tokenKey == "" {
			var token Token
			if err := tx.Unscoped().Select("key").Where("id = ? AND user_id = ?", a.TokenID, a.UserID).Take(&token).Error; err != nil {
				return err
			}
			tokenKey = token.Key
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	if confirmQuota != nil {
		if err := errors.Join(invalidateUserCache(a.UserID), invalidateTokenCacheForMutation(tokenKey)); err != nil {
			common.SysError(fmt.Sprintf("reviewed debits committed; cache refresh failed: %v", err))
		}
	}
	return receipts, total, nil
}

func (c ReviewedDebit) recover(tx *gorm.DB, user *User, actorID int, start, end int64, apply bool, receipt *DebitReceipt, tokenKey *string) error {
	var ops []BillingOperation
	if err := lockForUpdate(tx).Where("user_id = ? AND request_id = ?", user.Id, c.RequestID).Limit(2).Find(&ops).Error; err != nil {
		return err
	}
	if len(ops) != 1 {
		return ErrBillingConflict
	}
	op := ops[0]
	if op.ID != c.OperationID || op.RequestID != c.RequestID || op.ModelName != c.Model || op.TokenID != c.TokenID || op.CreatedAt < start || op.CreatedAt > c.EndedAt || op.Reserved != 0 {
		return ErrBillingConflict
	}

	var logs []Log
	if err := lockForUpdate(tx).Where("(user_id = ? AND request_id = ? OR upstream_request_id = ?) AND type IN ?", user.Id, op.RequestID, c.UpstreamID, []int{LogTypeConsume, LogTypeRefund}).Limit(2).Find(&logs).Error; err != nil {
		return err
	}
	var log Log
	if len(logs) > 1 {
		return ErrBillingConflict
	}
	if len(logs) == 1 {
		log = logs[0]
		var meta struct {
			Source string `json:"billing_source"`
		}
		if log.Type != LogTypeConsume || log.UserId != c.UserID || log.RequestId != c.RequestID || log.UpstreamRequestId != c.UpstreamID || log.TokenId != c.TokenID || log.ModelName != c.Model || log.CreatedAt != c.EndedAt || int64(log.Quota) != c.Quota || common.UnmarshalJsonStr(log.Other, &meta) != nil || meta.Source != "wallet" || (c.LogID != 0 && log.Id != c.LogID) {
			return ErrBillingConflict
		}
	} else if c.Supplier == nil {
		return ErrBillingConflict
	}

	key := billingSettleKeyPrefix + op.ID
	var entry BillingEntry
	if err := tx.Where("event_key = ?", key).Limit(1).Find(&entry).Error; err != nil {
		return err
	}
	if entry.ID != 0 {
		if op.State != "settled" || op.Actual != c.Quota || log.Id == 0 || entry.UserID != user.Id || entry.RequestID != op.RequestID || entry.TokenID != op.TokenID || entry.SourceLogID != log.Id || entry.Kind != "funding" || entry.BalanceBasis != "historical_usage_debit" || entry.Quota != 0 || entry.WalletDelta != -c.Quota {
			return ErrBillingConflict
		}
		*receipt = DebitReceipt{Status: "already_applied", EntryID: entry.ID}
		return nil
	}
	if op.State != "reserved" || op.Actual != 0 || (c.Supplier != nil && log.Id != 0) {
		return ErrBillingConflict
	}
	conflicts := tx.Model(&BillingEntry{}).Where("user_id = ?", user.Id).Where("(request_id = ? OR (source_log_id > 0 AND source_log_id = ?) OR (kind = 'funding' AND wallet_delta < 0 AND posted_at >= ? AND (request_id IS NULL OR request_id = '')))", op.RequestID, log.Id, op.CreatedAt)
	var count int64
	if err := conflicts.Count(&count).Error; err != nil {
		return err
	}
	if count != 0 {
		return ErrBillingConflict
	}
	var account BillingAccount
	if err := lockForUpdate(tx).Where("user_id = ?", user.Id).Take(&account).Error; err != nil {
		return err
	}
	state := BillingHistoryState{Account: account}
	if err := readBillingHistoryState(tx, &state, start, end, c.Month); err != nil {
		return err
	}
	if !state.PeriodCovered || account.CompanyTitle == "" || account.TaxID == "" || c.EndedAt < account.AccountingStartAt || c.EndedAt >= state.Boundary || state.Entries != 0 || state.Hours != 0 || state.ActiveStatement != "" {
		return ErrBillingHistoryBlocked
	}
	if err := tx.Model(&BillingHistoryImport{}).Where("user_id = ? AND month = ?", user.Id, c.Month).Count(&count).Error; err != nil {
		return err
	}
	if count != 0 {
		return ErrBillingHistoryBlocked
	}
	if err := tx.Model(&Log{}).Where("user_id = ? AND created_at >= ? AND created_at < ? AND type IN ?", user.Id, max(start, account.AccountingStartAt), state.Boundary, []int{LogTypeConsume, LogTypeRefund}).Count(&count).Error; err != nil {
		return err
	}
	if c.Supplier != nil {
		count++
	}
	if count > MaxBillingHistoryRecords {
		return ErrBillingHistoryBlocked
	}
	var token Token
	if err := lockForUpdate(tx).Select("id", "name", "key", "used_quota", "remain_quota").Where("id = ? AND user_id = ?", op.TokenID, user.Id).Take(&token).Error; err != nil {
		return err
	}
	if int64(user.Quota) < c.Quota {
		return ErrBillingInsufficientQuota
	}
	if int64(token.UsedQuota) > 2147483647-c.Quota || int64(token.RemainQuota) < -2147483648+c.Quota {
		return ErrBillingConflict
	}
	receipt.Status = "ready"
	if !apply {
		return nil
	}
	if c.Supplier != nil {
		if err := c.recordSupplierUsage(tx, &log, token.Name, actorID); err != nil {
			return err
		}
	}
	*tokenKey = token.Key
	if err := invalidateTokenCacheForMutation(token.Key); err != nil {
		return err
	}
	if err := changeBillingWallet(tx, user, -c.Quota); err != nil {
		return err
	}
	entry = BillingEntry{EventKey: key, Kind: "funding", WalletDelta: -c.Quota, ActorID: actorID, RequestID: op.RequestID, ModelName: c.Model, TokenID: op.TokenID, SourceLogID: log.Id, BalanceBasis: "historical_usage_debit"}
	if err := appendBillingEntry(tx, user, &entry); err != nil {
		return err
	}
	if err := tx.Model(&token).Updates(map[string]any{"remain_quota": gorm.Expr("remain_quota - ?", c.Quota), "used_quota": gorm.Expr("used_quota + ?", c.Quota)}).Error; err != nil {
		return err
	}
	r := tx.Model(&op).Where("state = ? AND actual = 0 AND reserved = 0", "reserved").Updates(map[string]any{"state": "settled", "actual": c.Quota, "updated_at": common.GetTimestamp()})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected != 1 {
		return ErrBillingConflict
	}

	*receipt = DebitReceipt{Status: "applied", EntryID: entry.ID}
	return nil
}

func (c ReviewedDebit) recordSupplierUsage(tx *gorm.DB, log *Log, tokenName string, actor int) error {
	s := c.Supplier
	base, err := decimal.NewFromString(s.BaseCNY)
	if err != nil {
		return err
	}
	factor, err := decimal.NewFromString(s.SaleRatio)
	if err != nil {
		return err
	}
	cost, err := decimal.NewFromString(s.CostCNY)
	if err != nil {
		return err
	}
	unit := decimal.NewFromInt(500000)
	rate := decimal.NewFromInt(7)
	basis := base.Mul(unit).Div(rate)
	saleQuota, err := common.QuotaFromDecimalStrict(basis.Mul(factor))
	if err != nil || int64(saleQuota) != c.Quota {
		return ErrBillingConflict
	}
	costQuota, err := common.QuotaFromDecimalStrict(cost.Mul(unit).Div(rate))
	if err != nil {
		return err
	}
	// Do not invent a channel, IP, cache breakdown or original gateway response.
	// Source documents and purchase prices are administrator-only.
	meta, err := common.Marshal(map[string]any{"billing_source": "wallet", "group_ratio": factor.InexactFloat64(), "admin_info": map[string]any{"reconstructed_from": "supplier_statement", "approved_by": actor, "supplier_evidence": s}})
	if err != nil {
		return err
	}
	*log = Log{Type: LogTypeConsume, CreatedAt: c.EndedAt, UserId: c.UserID, Username: c.Username, RequestId: c.RequestID, UpstreamRequestId: c.UpstreamID, TokenId: c.TokenID, TokenName: tokenName, ModelName: c.Model, Group: s.Group, PromptTokens: s.Input, CompletionTokens: s.Output, IsStream: s.Stream, UseTime: s.Seconds, Quota: int(c.Quota), Content: "根据已核验的上游账单补录", Other: string(meta)}
	var used int64
	if err := tx.Model(&CostAccountingSnapshot{}).Where("request_id = ? OR upstream_request_id = ?", c.RequestID, c.UpstreamID).Count(&used).Error; err != nil {
		return err
	}
	if used != 0 {
		return ErrBillingConflict
	}
	if err := tx.Create(log).Error; err != nil {
		return err
	}
	snap, err := BuildCostAccountingSnapshot(log, &CostAccountingInput{CostBasisQuota: basis.String(), CostQuota: int64(costQuota), CostDiscount: s.BuyRatio, Source: CostSnapshotSourceBackfill})
	if err != nil {
		return err
	}
	return tx.Create(snap).Error
}
