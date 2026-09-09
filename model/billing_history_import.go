package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// Online bootstrap imports are intentionally bounded. Large or mixed-ledger
// months require a separately reviewed migration, not an unbounded wallet lock.
const MaxBillingHistoryRecords = 10000

var ErrBillingHistoryBlocked = errors.New("historical billing import is blocked; refresh the review and resolve its checks")

type BillingHistoryRecord struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	CreatedAt int64  `json:"created_at"`
	Type      int    `json:"type"`
	Quota     int64  `json:"quota"`
	RequestID string `json:"request_id"`
	ModelName string `json:"model_name"`
	TokenID   int    `json:"token_id"`
}

// Source records live in a verified private object, not an unbounded TEXT cell.
// A completed import is append-only and unique for the entire customer/month.
type BillingHistoryImport struct {
	ID               string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	UserID           int    `json:"user_id" gorm:"uniqueIndex:idx_billing_history_month,priority:1"`
	Month            string `json:"month" gorm:"type:varchar(7);uniqueIndex:idx_billing_history_month,priority:2"`
	SourceSHA256     string `json:"source_sha256" gorm:"type:char(64)"`
	ArtifactID       int64  `json:"artifact_id"`
	StorageProfileID int    `json:"storage_profile_id"`
	Records          int64  `json:"records"`
	ChargeQuota      int64  `json:"charge_quota"`
	RefundQuota      int64  `json:"refund_quota"`
	FromSequence     int64  `json:"from_sequence"`
	ToSequence       int64  `json:"to_sequence"`
	OldStartAt       int64  `json:"old_start_at"`
	NewStartAt       int64  `json:"new_start_at"`
	ConfirmedBy      int    `json:"confirmed_by"`
	ConfirmedAt      int64  `json:"confirmed_at"`
	SessionID        string `json:"-" gorm:"type:varchar(128)"`
	Note             string `json:"note" gorm:"type:text"`
}

func (*BillingHistoryImport) BeforeUpdate(*gorm.DB) error { return ErrBillingEvidenceImmutable }
func (*BillingHistoryImport) BeforeDelete(*gorm.DB) error { return ErrBillingEvidenceImmutable }

type BillingHistoryState struct {
	Account         BillingAccount
	Entries         int64
	Hours           int64
	Pending         int64
	ActiveStatement string
	CanExtendStart  bool
}

func GetBillingHistoryState(ctx context.Context, userID int, month string) (*BillingHistoryState, error) {
	start, end, err := BillingMonthBounds(month)
	if err != nil {
		return nil, err
	}
	state := &BillingHistoryState{Account: BillingAccount{UserID: userID}}
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = DB.WithContext(queryCtx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Limit(1).Find(&state.Account).Error; err != nil {
			return err
		}
		return readBillingHistoryState(tx, state, start, end, month)
	})
	return state, err
}

// The same checks run under the account lock at commit, so a concurrent
// settlement or new draft cannot race an otherwise valid preview.
func readBillingHistoryState(tx *gorm.DB, state *BillingHistoryState, start, end int64, month string) error {
	userID := state.Account.UserID
	var entry BillingEntry
	result := tx.Select("id").Where("user_id = ? AND posted_at >= ? AND posted_at < ? AND kind <> ?", userID, start, end, "funding").Limit(1).Find(&entry)
	if result.Error != nil {
		return result.Error
	}
	state.Entries = result.RowsAffected
	if err := tx.Model(&BillingHour{}).Where("user_id = ? AND hour >= ? AND hour < ?", userID, start, end).Count(&state.Hours).Error; err != nil {
		return err
	}
	if err := tx.Model(&BillingOperation{}).Where("user_id = ? AND state = ? AND created_at < ?", userID, "reserved", end).Count(&state.Pending).Error; err != nil {
		return err
	}
	var statement BillingStatement
	if err := tx.Select("id").Where("user_id = ? AND month = ? AND status <> ?", userID, month, StatementVoid).Limit(1).Find(&statement).Error; err != nil {
		return err
	}
	state.ActiveStatement = statement.ID
	state.CanExtendStart = true
	if state.Account.AccountingStartAt == 0 || state.Account.AccountingStartAt > start {
		var statements int64
		if err := tx.Model(&BillingStatement{}).Where("user_id = ? AND status <> ?", userID, StatementVoid).Count(&statements).Error; err != nil {
			return err
		}
		// Do not silently change coverage of existing ledger entries or files.
		state.CanExtendStart = statements == 0 && state.Account.Sequence == 0
	}
	return nil
}

func GetBillingHistoryRecords(ctx context.Context, userID int, month string) ([]BillingHistoryRecord, error) {
	start, end, err := BillingMonthBounds(month)
	if err != nil || userID <= 0 || start < 0 {
		return nil, errors.New("invalid historical billing period")
	}
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		return nil, errors.New("online historical import requires a relational log database with stable IDs")
	}
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	records := make([]BillingHistoryRecord, 0)
	err = LOG_DB.WithContext(queryCtx).Model(&Log{}).
		Select("id", "user_id", "created_at", "type", "quota", "request_id", "model_name", "token_id").
		Where("user_id = ? AND created_at >= ? AND created_at < ? AND type IN ?", userID, start, end, []int{LogTypeConsume, LogTypeRefund}).
		Order("created_at asc").Order("id asc").Limit(MaxBillingHistoryRecords + 1).Find(&records).Error
	return records, err
}

func GetBillingHistoryImport(userID int, id string) (*BillingHistoryImport, error) {
	var batch BillingHistoryImport
	err := DB.Where("id = ? AND user_id = ?", id, userID).First(&batch).Error
	return &batch, err
}

func ListBillingHistoryImports(userID int, month string) ([]BillingHistoryImport, error) {
	imports := make([]BillingHistoryImport, 0)
	err := DB.Where("user_id = ? AND month = ?", userID, month).Order("confirmed_at desc").Limit(10).Find(&imports).Error
	return imports, err
}

func ApplyBillingHistoryImport(ctx context.Context, batch *BillingHistoryImport, expected BillingAccount, records []BillingHistoryRecord) error {
	start, end, err := BillingMonthBounds(batch.Month)
	if err != nil || len(records) == 0 || len(records) > MaxBillingHistoryRecords || batch.ID == "" || batch.SourceSHA256 == "" || batch.ArtifactID <= 0 || batch.SessionID == "" || batch.Note == "" {
		return ErrBillingHistoryBlocked
	}
	if common.GetTimestamp() < end+86400 {
		return ErrBillingHistoryBlocked
	}
	queryCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return DB.WithContext(queryCtx).Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).Select("id", "quota").First(&user, batch.UserID).Error; err != nil {
			return err
		}
		account, err := lockBillingAccount(tx, batch.UserID)
		if err != nil {
			return err
		}
		var existing BillingHistoryImport
		if err := tx.Where("id = ?", batch.ID).Limit(1).Find(&existing).Error; err != nil {
			return err
		}
		if existing.ID != "" {
			if existing.UserID != batch.UserID || existing.SourceSHA256 != batch.SourceSHA256 || existing.Month != batch.Month {
				return ErrBillingConflict
			}
			*batch = existing
			return nil
		}
		if account.UserID != expected.UserID || account.AccountingStartAt != expected.AccountingStartAt || account.ProfileVersion != expected.ProfileVersion || account.Sequence != expected.Sequence || account.StartSequence != expected.StartSequence {
			return ErrBillingConflict
		}
		state := BillingHistoryState{Account: *account}
		if err := readBillingHistoryState(tx, &state, start, end, batch.Month); err != nil {
			return err
		}
		if state.Entries != 0 || state.Hours != 0 || state.Pending != 0 || state.ActiveStatement != "" || !state.CanExtendStart || account.CompanyTitle == "" || account.TaxID == "" {
			return ErrBillingHistoryBlocked
		}
		batch.OldStartAt = account.AccountingStartAt
		if account.AccountingStartAt == 0 || account.AccountingStartAt > start {
			account.AccountingStartAt, account.StartSequence = start, account.Sequence+1
		}
		batch.NewStartAt, batch.FromSequence = account.AccountingStartAt, account.Sequence+1
		entries := make([]BillingEntry, 0, len(records))
		byHour := make(map[int64]*BillingHour)
		seen := make(map[int]bool, len(records))
		var charge, refund int64
		for _, record := range records {
			if record.ID <= 0 || seen[record.ID] || record.UserID != batch.UserID || record.CreatedAt < start || record.CreatedAt >= end || record.Quota < 0 || record.Quota > int64(common.MaxWalletQuota) || (record.Type != LogTypeConsume && record.Type != LogTypeRefund) {
				return ErrBillingEvidenceIntegrity
			}
			seen[record.ID] = true
			body, err := common.Marshal(record)
			if err != nil {
				return err
			}
			hash := sha256.Sum256(body)
			digest := hex.EncodeToString(hash[:])
			account.Sequence++
			quota, kind := record.Quota, "history_usage"
			if record.Type == LogTypeRefund {
				if record.Quota > int64(common.MaxWalletQuota)-refund {
					return ErrWalletQuotaLimitExceeded
				}
				refund += record.Quota
				quota, kind = -quota, "history_refund"
			} else {
				if record.Quota > int64(common.MaxWalletQuota)-charge {
					return ErrWalletQuotaLimitExceeded
				}
				charge += record.Quota
			}
			entries = append(entries, BillingEntry{EventKey: "history:" + digest, UserID: batch.UserID, Sequence: account.Sequence, PostedAt: record.CreatedAt, Kind: kind, Quota: quota, RequestID: record.RequestID, ModelName: record.ModelName, TokenID: record.TokenID, ActorID: batch.ConfirmedBy, ImportID: batch.ID, SourceLogID: record.ID, SourceSHA256: digest, BalanceAfter: int64(user.Quota), BalanceBasis: "unchanged_at_import"})
			at := record.CreatedAt / 3600 * 3600
			hour := byHour[at]
			if hour == nil {
				hour = &BillingHour{UserID: batch.UserID, Hour: at, FirstSequence: account.Sequence}
				byHour[at] = hour
			}
			if record.Type == LogTypeRefund {
				hour.Refund += record.Quota
			} else {
				hour.Charge += record.Quota
			}
			hour.Count++
			hour.LastSequence = account.Sequence
		}
		hours := make([]BillingHour, 0, len(byHour))
		for _, hour := range byHour {
			hours = append(hours, *hour)
		}
		sort.Slice(hours, func(i, j int) bool { return hours[i].Hour < hours[j].Hour })
		if err := tx.CreateInBatches(&entries, 100).Error; err != nil {
			return err
		}
		if err := tx.CreateInBatches(&hours, 100).Error; err != nil {
			return err
		}
		account.ProfileVersion++
		account.UpdatedAt = common.GetTimestamp()
		if err := tx.Save(account).Error; err != nil {
			return err
		}
		batch.Records, batch.ChargeQuota, batch.RefundQuota = int64(len(records)), charge, refund
		batch.ToSequence, batch.ConfirmedAt = account.Sequence, account.UpdatedAt
		if err := tx.Create(batch).Error; err != nil {
			return err
		}
		body, err := common.Marshal(struct {
			Account  *BillingAccount `json:"account"`
			ImportID string          `json:"import_id"`
		}{account, batch.ID})
		if err != nil {
			return err
		}
		return tx.Create(&BillingAccountEvent{UserID: batch.UserID, ActorID: batch.ConfirmedBy, CreatedAt: batch.ConfirmedAt, Snapshot: string(body)}).Error
	})
}
