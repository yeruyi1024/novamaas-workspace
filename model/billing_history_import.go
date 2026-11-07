package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// nonLiveBillingKinds are the entry kinds that are not a month's own live
// traffic: funding, corrections of other periods and earlier imports.
var nonLiveBillingKinds = []string{"funding", "rate_correction", "history_usage", "history_refund"}

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
	PeriodCovered   bool
	// Boundary is where the month's formal ledger begins: the time of its first
	// formal entry. A month in which formal accounting was switched on part-way
	// is imported only up to it, so the entries already booked are not booked a
	// second time. Without formal entries it is the end of the month.
	Boundary int64
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
	state.PeriodCovered = state.Account.AccountingStartAt > 0 && state.Account.AccountingStartAt < end
	if state.Account.AccountingStartAt > start {
		start = state.Account.AccountingStartAt
	}
	state.Boundary = end
	var first BillingEntry
	result := tx.Select("posted_at").Where("user_id = ? AND posted_at >= ? AND posted_at < ? AND sequence >= ? AND kind NOT IN ?", userID, start, end, state.Account.StartSequence, nonLiveBillingKinds).Order("posted_at asc").Limit(1).Find(&first)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		state.Boundary = first.PostedAt
	}
	// Only the part of the month before the boundary is imported, so only it has
	// to be free of formal entries and summaries. The boundary hour already holds
	// a summary of its own, which the import adds to.
	var entry BillingEntry
	result = tx.Select("id").Where("user_id = ? AND posted_at >= ? AND posted_at < ? AND sequence >= ? AND kind <> ?", userID, start, state.Boundary, state.Account.StartSequence, "funding").Limit(1).Find(&entry)
	if result.Error != nil {
		return result.Error
	}
	state.Entries = result.RowsAffected
	if err := tx.Model(&BillingHour{}).Where("user_id = ? AND hour >= ? AND hour < ?", userID, start/3600*3600, state.Boundary/3600*3600).Count(&state.Hours).Error; err != nil {
		return err
	}
	if err := tx.Model(&BillingOperation{}).Where("user_id = ? AND state = ? AND created_at < ?", userID, "reserved", end).Count(&state.Pending).Error; err != nil {
		return err
	}
	var statement BillingStatement
	if err := effectiveBillingStatements(tx).Select("id").Where("user_id = ? AND month = ?", userID, month).Limit(1).Find(&statement).Error; err != nil {
		return err
	}
	state.ActiveStatement = statement.ID
	return nil
}

// GetBillingHistoryRecords lists the usage logs of a month from the accounting
// start up to the boundary, where the formal ledger takes over. A log whose
// request the formal ledger booked just after the boundary is left out.
func GetBillingHistoryRecords(ctx context.Context, userID int, month string, accountingStartAt, boundary int64) ([]BillingHistoryRecord, error) {
	start, end, err := BillingMonthBounds(month)
	if err != nil || userID <= 0 || start < 0 || accountingStartAt < 0 {
		return nil, errors.New("invalid historical billing period")
	}
	records := make([]BillingHistoryRecord, 0)
	if accountingStartAt == 0 || accountingStartAt >= end {
		return records, nil
	}
	if accountingStartAt > start {
		start = accountingStartAt
	}
	monthEnd := end
	if boundary > 0 && boundary < end {
		end = boundary
	}
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		return nil, errors.New("online historical import requires a relational log database with stable IDs")
	}
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := LOG_DB.WithContext(queryCtx).Model(&Log{}).
		Select("id", "user_id", "created_at", "type", "quota", "request_id", "model_name", "token_id").
		Where("user_id = ? AND created_at >= ? AND created_at < ? AND type IN ?", userID, start, end, []int{LogTypeConsume, LogTypeRefund}).
		Order("created_at asc").Order("id asc")
	var booked []string
	if end != monthEnd {
		if err := DB.WithContext(queryCtx).Model(&BillingEntry{}).Where("user_id = ? AND posted_at >= ? AND posted_at < ? AND kind NOT IN ?", userID, end, end+3600, nonLiveBillingKinds).Pluck("request_id", &booked).Error; err != nil {
			return nil, err
		}
	}
	// A log without a request ID cannot be matched, so it is never left out.
	bookedRequests := lo.Keyify(lo.Without(booked, ""))
	var after BillingHistoryRecord
	for len(records) <= MaxBillingHistoryRecords {
		limit := MaxBillingHistoryRecords + 1 - len(records)
		pageQuery := query.Session(&gorm.Session{})
		if after.ID > 0 {
			pageQuery = pageQuery.Where("created_at > ? OR (created_at = ? AND id > ?)", after.CreatedAt, after.CreatedAt, after.ID)
		}
		var page []BillingHistoryRecord
		if err := pageQuery.Limit(limit).Find(&page).Error; err != nil {
			return nil, err
		}
		for _, record := range page {
			if _, isBooked := bookedRequests[record.RequestID]; !isBooked {
				records = append(records, record)
			}
		}
		if len(page) < limit {
			break
		}
		after = page[len(page)-1]
	}
	return records, nil
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
		if state.Entries != 0 || state.Hours != 0 || state.Pending != 0 || state.ActiveStatement != "" || !state.PeriodCovered || account.CompanyTitle == "" || account.TaxID == "" {
			return ErrBillingHistoryBlocked
		}
		end = state.Boundary
		batch.OldStartAt = account.AccountingStartAt
		if account.AccountingStartAt > start {
			start = account.AccountingStartAt
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
		// The month's first formal hour already has a summary: the import adds to it
		// instead of opening a second one for the same hour.
		hourStarts := make([]int64, 0, len(hours))
		for _, hour := range hours {
			hourStarts = append(hourStarts, hour.Hour)
		}
		var existingHours []BillingHour
		if err := tx.Where("user_id = ? AND hour IN ?", batch.UserID, hourStarts).Find(&existingHours).Error; err != nil {
			return err
		}
		existingByHour := make(map[int64]BillingHour, len(existingHours))
		for _, hour := range existingHours {
			existingByHour[hour.Hour] = hour
		}
		newHours := make([]BillingHour, 0, len(hours))
		for _, hour := range hours {
			existing, found := existingByHour[hour.Hour]
			if !found {
				newHours = append(newHours, hour)
				continue
			}
			if hour.Charge > int64(common.MaxWalletQuota)-existing.Charge || hour.Refund > int64(common.MaxWalletQuota)-existing.Refund {
				return ErrWalletQuotaLimitExceeded
			}
			existing.Charge += hour.Charge
			existing.Refund += hour.Refund
			existing.Count += hour.Count
			existing.FirstSequence = min(existing.FirstSequence, hour.FirstSequence)
			existing.LastSequence = max(existing.LastSequence, hour.LastSequence)
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}
		if err := tx.CreateInBatches(&newHours, 100).Error; err != nil {
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
