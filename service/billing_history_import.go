package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type BillingHistoryReview struct {
	UserID            int                          `json:"user_id"`
	Month             string                       `json:"month"`
	Customer          model.BillingUserIdentity    `json:"customer"`
	Ready             bool                         `json:"ready"`
	Checks            []BillingCheck               `json:"checks"`
	SourceSHA256      string                       `json:"source_sha256"`
	SourceCount       int                          `json:"source_count"`
	RecordLimit       int                          `json:"record_limit"`
	ExistingStatement string                       `json:"existing_statement"`
	OldStartAt        int64                        `json:"old_start_at"`
	NewStartAt        int64                        `json:"new_start_at"`
	Snapshot          *BillingSnapshot             `json:"snapshot"`
	Records           []model.BillingHistoryRecord `json:"-"`
	Account           model.BillingAccount         `json:"-"`
	Source            []byte                       `json:"-"`
}

// This is a read-only review. No ledger, account, import batch or object-store
// writes occur until the administrator explicitly confirms this exact digest.
func ReviewBillingHistory(ctx context.Context, userID int, month string) (*BillingHistoryReview, error) {
	start, end, err := model.BillingMonthBounds(month)
	if err != nil || start < 0 || userID <= 0 {
		return nil, errors.New("invalid historical billing period")
	}
	state, err := model.GetBillingHistoryState(ctx, userID, month)
	if err != nil {
		return nil, err
	}
	identities, err := model.GetBillingUserIdentities(ctx, []int{userID})
	if err != nil {
		return nil, err
	}
	customer, exists := identities[userID]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	records, err := model.GetBillingHistoryRecords(ctx, userID, month)
	if err != nil {
		return nil, err
	}
	review := &BillingHistoryReview{UserID: userID, Month: month, Customer: customer, Ready: true, SourceCount: len(records), RecordLimit: model.MaxBillingHistoryRecords, Account: state.Account, ExistingStatement: state.ActiveStatement, OldStartAt: state.Account.AccountingStartAt, NewStartAt: state.Account.AccountingStartAt}
	if review.NewStartAt == 0 || review.NewStartAt > start {
		review.NewStartAt = start
	}
	review.Checks = []BillingCheck{
		{Code: "month_closed", Passed: common.GetTimestamp() >= end+86400},
		{Code: "identity_complete", Passed: state.Account.CompanyTitle != "" && state.Account.TaxID != ""},
		{Code: "history_within_limit", Passed: len(records) > 0 && len(records) <= model.MaxBillingHistoryRecords},
		{Code: "history_empty_ledger", Passed: state.Entries == 0 && state.Hours == 0},
		{Code: "operations_settled", Passed: state.Pending == 0},
		{Code: "no_active_statement", Passed: state.ActiveStatement == ""},
		{Code: "history_start_safe", Passed: state.CanExtendStart},
	}
	for _, check := range review.Checks {
		if !check.Passed {
			review.Ready = false
		}
	}
	if len(records) > model.MaxBillingHistoryRecords {
		return review, nil // never present a truncated set as the complete month
	}
	byHour := make(map[int64]*model.BillingHour)
	seen := make(map[int]bool, len(records))
	for _, record := range records {
		if record.ID <= 0 || seen[record.ID] || record.UserID != userID || record.Quota < 0 || record.Quota > int64(common.MaxWalletQuota) || len(record.RequestID) > 128 || len([]rune(record.ModelName)) > 255 {
			return nil, model.ErrBillingEvidenceIntegrity
		}
		seen[record.ID] = true
		at := record.CreatedAt / 3600 * 3600
		hour := byHour[at]
		if hour == nil {
			hour = &model.BillingHour{UserID: userID, Hour: at}
			byHour[at] = hour
		}
		if record.Type == model.LogTypeRefund {
			if record.Quota > int64(common.MaxWalletQuota)-hour.Refund {
				return nil, model.ErrWalletQuotaLimitExceeded
			}
			hour.Refund += record.Quota
		} else {
			if record.Quota > int64(common.MaxWalletQuota)-hour.Charge {
				return nil, model.ErrWalletQuotaLimitExceeded
			}
			hour.Charge += record.Quota
		}
		hour.Count++
	}
	hours := make([]model.BillingHour, 0, len(byHour))
	for _, hour := range byHour {
		hours = append(hours, *hour)
	}
	sort.Slice(hours, func(i, j int) bool { return hours[i].Hour < hours[j].Hour })
	account := state.Account
	account.AccountingStartAt = review.NewStartAt
	review.Snapshot, err = BuildBillingSnapshot(&account, month, hours)
	if err != nil {
		return nil, err
	}
	review.Snapshot.Username, review.Snapshot.DisplayName = customer.Username, customer.DisplayName
	// Commit-order sequence is deliberately included even though it is not a
	// public account field. A concurrent settlement invalidates the review.
	review.Source, err = common.Marshal(struct {
		SchemaVersion int                          `json:"schema_version"`
		Source        string                       `json:"source"`
		Account       model.BillingAccount         `json:"account"`
		Sequence      int64                        `json:"sequence"`
		StartSequence int64                        `json:"start_sequence"`
		Snapshot      *BillingSnapshot             `json:"snapshot"`
		Records       []model.BillingHistoryRecord `json:"records"`
	}{1, "administrator_verified_usage_logs", state.Account, state.Account.Sequence, state.Account.StartSequence, review.Snapshot, records})
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(review.Source)
	review.SourceSHA256, review.Records = hex.EncodeToString(hash[:]), records
	return review, nil
}

// Financial mutation is confined to the local ledger transaction. In
// particular, this NEVER invokes wallet deduction, token quota, or user usage
// counters. Cloud failure, stale reviews and commit races leave the ledger alone.
func ConfirmBillingHistoryImport(ctx context.Context, userID, actorID, profileID int, month, id, digest, note, sessionID string, store BillingArchiveWriter) (*model.BillingHistoryImport, error) {
	note = strings.TrimSpace(note)
	if userID <= 0 || actorID <= 0 || profileID <= 0 || sessionID == "" || len(id) != 32 || len(digest) != 64 || note == "" || len([]rune(note)) > 2000 {
		return nil, model.ErrBillingHistoryBlocked
	}
	if _, err := hex.DecodeString(id); err != nil {
		return nil, model.ErrBillingHistoryBlocked
	}
	batchID := "history-" + id
	existing, err := model.GetBillingHistoryImport(userID, batchID)
	if err == nil {
		if existing.SourceSHA256 != digest || existing.Month != month {
			return nil, model.ErrBillingConflict
		}
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	review, err := ReviewBillingHistory(ctx, userID, month)
	if err != nil {
		return nil, err
	}
	if review.SourceSHA256 != digest {
		return nil, model.ErrBillingConflict
	}
	if !review.Ready {
		return nil, model.ErrBillingHistoryBlocked
	}
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(review.Source); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	artifact, err := store.Put(ctx, batchID, "source", 0, userID, "application/gzip", compressed.Bytes(), int64(len(review.Records)))
	if err != nil {
		return nil, err
	}
	batch := &model.BillingHistoryImport{ID: batchID, UserID: userID, Month: month, SourceSHA256: digest, ArtifactID: artifact.ID, StorageProfileID: profileID, ConfirmedBy: actorID, Note: note, SessionID: sessionID}
	if err := model.ApplyBillingHistoryImport(ctx, batch, review.Account, review.Records); err != nil {
		return nil, err
	}
	return batch, nil
}
