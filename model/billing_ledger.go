package model

import (
	"errors"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BillingAccount provides a per-customer commit-order sequence and cutoff.
type BillingAccount struct {
	UserID            int    `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	AccountingStartAt int64  `json:"accounting_start_at" gorm:"bigint"`
	StartSequence     int64  `json:"-" gorm:"bigint"`
	Sequence          int64  `json:"-" gorm:"bigint"`
	CompanyTitle      string `json:"company_title" gorm:"type:varchar(200)"`
	TaxID             string `json:"tax_id" gorm:"type:varchar(64)"`
	ProfileVersion    int64  `json:"profile_version" gorm:"bigint"`
	UpdatedAt         int64  `json:"updated_at" gorm:"bigint"`
}
type BillingAccountEvent struct {
	ID        int64  `json:"id" gorm:"primaryKey"`
	UserID    int    `json:"user_id" gorm:"index"`
	ActorID   int    `json:"actor_id"`
	CreatedAt int64  `json:"created_at" gorm:"bigint"`
	Snapshot  string `json:"snapshot" gorm:"type:text"`
}

// Reservations are not consumption. Unfinished operations survive crashes.
type BillingOperation struct {
	ID        string `json:"id" gorm:"type:varchar(64);primaryKey"`
	UserID    int    `json:"user_id" gorm:"index:idx_billing_operations_open,priority:1"`
	State     string `json:"state" gorm:"type:varchar(16);index:idx_billing_operations_open,priority:2"`
	Reserved  int64  `json:"reserved" gorm:"bigint"`
	Actual    int64  `json:"actual" gorm:"bigint"`
	RequestID string `json:"request_id" gorm:"type:varchar(128)"`
	ModelName string `json:"model_name" gorm:"type:varchar(255)"`
	TokenID   int    `json:"token_id"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index:idx_billing_operations_open,priority:3"`
	UpdatedAt int64  `json:"updated_at" gorm:"bigint"`
}

// Entries are append-only: positive quota is consumption, negative is refund.
// Administrative funding has Kind=funding and Quota=0.
type BillingEntry struct {
	ID           int64  `json:"id" gorm:"primaryKey"`
	EventKey     string `json:"event_key" gorm:"type:varchar(128);uniqueIndex"`
	UserID       int    `json:"user_id" gorm:"uniqueIndex:idx_billing_entry_sequence,priority:1;index:idx_billing_entry_time,priority:1"`
	Sequence     int64  `json:"sequence" gorm:"bigint;uniqueIndex:idx_billing_entry_sequence,priority:2"`
	PostedAt     int64  `json:"posted_at" gorm:"bigint;index:idx_billing_entry_time,priority:2"`
	Kind         string `json:"kind" gorm:"type:varchar(32)"`
	Quota        int64  `json:"quota" gorm:"bigint"`
	WalletDelta  int64  `json:"wallet_delta" gorm:"bigint"`
	BalanceAfter int64  `json:"balance_after" gorm:"bigint"`
	RequestID    string `json:"request_id" gorm:"type:varchar(128)"`
	ModelName    string `json:"model_name" gorm:"type:varchar(255)"`
	TokenID      int    `json:"token_id"`
	ActorID      int    `json:"actor_id"`
	ImportID     string `json:"import_id,omitempty" gorm:"type:varchar(64)"`
	SourceLogID  int    `json:"source_log_id,omitempty"`
	SourceSHA256 string `json:"source_sha256,omitempty" gorm:"type:char(64)"`
	BalanceBasis string `json:"balance_basis,omitempty" gorm:"type:varchar(32)"`
}
type BillingHour struct {
	ID            int64 `json:"-" gorm:"primaryKey"`
	UserID        int   `json:"-" gorm:"uniqueIndex:idx_billing_hour,priority:1"`
	Hour          int64 `json:"hour" gorm:"bigint;uniqueIndex:idx_billing_hour,priority:2"`
	Charge        int64 `json:"charge" gorm:"bigint"`
	Refund        int64 `json:"refund" gorm:"bigint"`
	Count         int64 `json:"count" gorm:"bigint"`
	FirstSequence int64 `json:"-" gorm:"bigint"`
	LastSequence  int64 `json:"-" gorm:"bigint"`
}

var ErrBillingConflict = errors.New("billing state changed; refresh and retry")
var ErrBillingNotConfigured = errors.New("the administrator must configure the accounting start time first")
var ErrBillingInsufficientQuota = errors.New("wallet quota insufficient")

func GetBillingAccount(userID int) (*BillingAccount, error) {
	var account BillingAccount
	err := DB.Where("user_id = ?", userID).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &BillingAccount{UserID: userID}, nil
	}
	return &account, err
}

// Wallet mutations always lock User before BillingAccount.
func lockBillingAccount(tx *gorm.DB, userID int) (*BillingAccount, error) {
	account := BillingAccount{UserID: userID}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&account).Error; err != nil {
		return nil, err
	}
	err := lockForUpdate(tx).Where("user_id = ?", userID).First(&account).Error
	return &account, err
}
func SaveBillingAccount(userID, actorID int, expectedVersion int64, title, taxID string, startAt *int64) (*BillingAccount, error) {
	title, taxID = strings.TrimSpace(title), strings.TrimSpace(taxID)
	if len([]rune(title)) > 200 || len(taxID) > 64 || strings.ContainsAny(title+taxID, "\x00\r\n") {
		return nil, errors.New("invalid company title or tax ID")
	}
	var result *BillingAccount
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).Select("id").First(&user, userID).Error; err != nil {
			return err
		}
		account, err := lockBillingAccount(tx, userID)
		if err != nil {
			return err
		}
		if account.ProfileVersion != expectedVersion {
			return ErrBillingConflict
		}
		if startAt != nil && (*startAt != account.AccountingStartAt || account.AccountingStartAt == 0) {
			now := common.GetTimestamp()
			start := *startAt
			if start == 0 {
				start = now
			}
			if start < now || start > now+366*86400 {
				return errors.New("accounting start must be now or a future time within one year")
			}
			var count int64
			if err := tx.Model(&BillingStatement{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errors.New("accounting start is locked after a statement is created")
			}
			if account.AccountingStartAt > 0 {
				if err := tx.Model(&BillingEntry{}).Where("user_id = ? AND sequence >= ? AND posted_at >= ? AND kind <> ?", userID, account.StartSequence, account.AccountingStartAt, "funding").Count(&count).Error; err != nil {
					return err
				}
				if count > 0 {
					return errors.New("accounting start is locked after the first posted entry")
				}
			}
			account.AccountingStartAt, account.StartSequence = start, account.Sequence+1
		}
		account.CompanyTitle, account.TaxID = title, taxID
		account.ProfileVersion++
		account.UpdatedAt = common.GetTimestamp()
		if err := tx.Save(account).Error; err != nil {
			return err
		}
		snapshot, err := common.Marshal(account)
		if err != nil {
			return err
		}
		if err := tx.Create(&BillingAccountEvent{UserID: userID, ActorID: actorID, CreatedAt: account.UpdatedAt, Snapshot: string(snapshot)}).Error; err != nil {
			return err
		}
		result = account
		return nil
	})
	return result, err
}
func appendBillingEntry(tx *gorm.DB, user *User, entry *BillingEntry) error {
	account, err := lockBillingAccount(tx, user.Id)
	if err != nil {
		return err
	}
	account.Sequence++
	entry.UserID, entry.Sequence, entry.PostedAt = user.Id, account.Sequence, common.GetTimestamp()
	entry.BalanceAfter = int64(user.Quota)
	if err := tx.Create(entry).Error; err != nil {
		return err
	}
	if err := tx.Model(account).Update("sequence", account.Sequence).Error; err != nil {
		return err
	}
	if entry.Kind == "funding" || account.AccountingStartAt == 0 || entry.PostedAt < account.AccountingStartAt || entry.Sequence < account.StartSequence {
		return nil
	}
	hour := BillingHour{UserID: user.Id, Hour: entry.PostedAt / 3600 * 3600}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&hour).Error; err != nil {
		return err
	}
	hour.ID = 0 // a duplicate insert must not leave a driver-specific last insert ID
	if err := tx.Where("user_id = ? AND hour = ?", user.Id, hour.Hour).First(&hour).Error; err != nil {
		return err
	}
	charge, refund := entry.Quota, int64(0)
	if charge < 0 {
		refund, charge = -charge, 0
	}
	if charge > int64(common.MaxWalletQuota)-hour.Charge || refund > int64(common.MaxWalletQuota)-hour.Refund {
		return errors.New("billing aggregate exceeds exact quota limit")
	}
	if hour.FirstSequence == 0 {
		hour.FirstSequence = entry.Sequence
	}
	return tx.Model(&hour).Updates(map[string]interface{}{
		"charge": hour.Charge + charge, "refund": hour.Refund + refund, "count": hour.Count + 1,
		"first_sequence": hour.FirstSequence, "last_sequence": entry.Sequence,
	}).Error
}
func changeBillingWallet(tx *gorm.DB, user *User, delta int64) error {
	if delta > int64(common.MaxWalletQuota) || delta < -int64(common.MaxWalletQuota) || int64(user.Quota) > int64(common.MaxWalletQuota)-delta || int64(user.Quota) < -int64(common.MaxWalletQuota)-delta {
		return ErrWalletQuotaLimitExceeded
	}
	user.Quota += int(delta) // bounded by the exact wallet quota boundary above
	return tx.Model(&User{}).Where("id = ?", user.Id).Update("quota", user.Quota).Error
}
func BeginBillingOperation(op *BillingOperation) error {
	if op.ID == "" || op.Reserved < 0 || op.Reserved > int64(common.MaxWalletQuota) {
		return errors.New("invalid billing reservation")
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).First(&user, op.UserID).Error; err != nil {
			return err
		}
		if int64(user.Quota) < op.Reserved {
			return ErrBillingInsufficientQuota
		}
		op.State, op.CreatedAt, op.UpdatedAt = "reserved", common.GetTimestamp(), common.GetTimestamp()
		if err := tx.Create(op).Error; err != nil {
			return err
		}
		return changeBillingWallet(tx, &user, -op.Reserved)
	})
	if err == nil {
		_ = invalidateUserCache(op.UserID)
	}
	return err
}
func ResizeBillingReservation(id string, userID int, target int64) error {
	if target < 0 || target > int64(common.MaxWalletQuota) {
		return errors.New("invalid billing reservation")
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).First(&user, userID).Error; err != nil {
			return err
		}
		var op BillingOperation
		if err := lockForUpdate(tx).Where("id = ? AND user_id = ?", id, userID).First(&op).Error; err != nil {
			return err
		}
		if op.State != "reserved" {
			return ErrBillingConflict
		}
		if err := changeBillingWallet(tx, &user, op.Reserved-target); err != nil {
			return err
		}
		return tx.Model(&op).Updates(map[string]interface{}{"reserved": target, "updated_at": common.GetTimestamp()}).Error
	})
	if err == nil {
		_ = invalidateUserCache(userID)
	}
	return err
}
func FinishBillingOperation(id string, userID int, actual int64, refund bool) error {
	if actual < 0 || actual > int64(common.MaxWalletQuota) {
		return errors.New("invalid billing charge")
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).First(&user, userID).Error; err != nil {
			return err
		}
		var op BillingOperation
		if err := lockForUpdate(tx).Where("id = ? AND user_id = ?", id, userID).First(&op).Error; err != nil {
			return err
		}
		state := "settled"
		if refund {
			state, actual = "refunded", 0
		}
		if op.State == state && op.Actual == actual {
			return nil
		}
		if op.State != "reserved" {
			return ErrBillingConflict
		}
		if err := changeBillingWallet(tx, &user, op.Reserved-actual); err != nil {
			return err
		}
		if !refund {
			entry := BillingEntry{EventKey: "settle:" + op.ID, Kind: "usage", Quota: actual, WalletDelta: -actual, RequestID: op.RequestID, ModelName: op.ModelName, TokenID: op.TokenID}
			if err := appendBillingEntry(tx, &user, &entry); err != nil {
				return err
			}
		}
		return tx.Model(&op).Updates(map[string]interface{}{"state": state, "actual": actual, "updated_at": common.GetTimestamp()}).Error
	})
	if err == nil {
		_ = invalidateUserCache(userID)
	}
	return err
}

// EventKey must be server-owned. Admin funding never counts as consumption.
func PostBillingAdjustment(entry *BillingEntry, override *int) (bool, error) {
	if entry.Quota > int64(common.MaxWalletQuota) || entry.Quota < -int64(common.MaxWalletQuota) {
		return false, ErrWalletQuotaLimitExceeded
	}
	if entry.Kind == "funding" && entry.Quota != 0 {
		return false, errors.New("funding is not consumption")
	}
	if entry.Kind != "funding" && entry.WalletDelta != -entry.Quota {
		return false, errors.New("billing adjustment does not match wallet delta")
	}
	if entry.EventKey == "" || len(entry.EventKey) > 128 {
		return false, errors.New("invalid billing event key")
	}
	applied := false
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).First(&user, entry.UserID).Error; err != nil {
			return err
		}
		var existing BillingEntry
		err := tx.Where("event_key = ?", entry.EventKey).First(&existing).Error
		if err == nil {
			if existing.UserID != entry.UserID || existing.Kind != entry.Kind || existing.Quota != entry.Quota {
				return ErrBillingConflict
			}
			if (override == nil && existing.WalletDelta != entry.WalletDelta) || (override != nil && existing.BalanceAfter != int64(*override)) {
				return ErrBillingConflict
			}
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if override != nil {
			entry.WalletDelta = int64(*override) - int64(user.Quota)
		}
		if err := changeBillingWallet(tx, &user, entry.WalletDelta); err != nil {
			return err
		}
		if err := appendBillingEntry(tx, &user, entry); err != nil {
			return err
		}
		applied = true
		return nil
	})
	if err == nil {
		_ = invalidateUserCache(entry.UserID)
	}
	return applied, err
}
func GetBillingHours(userID int, start, end int64) ([]BillingHour, error) {
	hours := make([]BillingHour, 0)
	err := DB.Where("user_id = ? AND hour >= ? AND hour < ?", userID, start, end).Order("hour asc").Find(&hours).Error
	return hours, err
}
func BillingMonthBounds(month string) (int64, int64, error) {
	location := time.FixedZone("Asia/Shanghai", 8*3600)
	start, err := time.ParseInLocation("2006-01", month, location)
	if err != nil || start.Format("2006-01") != month {
		return 0, 0, errors.New("invalid billing month")
	}
	return start.Unix(), start.AddDate(0, 1, 0).Unix(), nil
}

// AdjustBillingTaskQuota commits the task's refund/settlement marker and money
// in the same transaction, so a stale poller cannot refund a task twice.
func AdjustBillingTaskQuota(task *Task, target int, modelName string) (bool, error) {
	if target < 0 || target > common.MaxWalletQuota {
		return false, errors.New("invalid task charge")
	}
	applied := false
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).First(&user, task.UserId).Error; err != nil {
			return err
		}
		var current Task
		if err := lockForUpdate(tx).Where("id = ? AND user_id = ?", task.ID, task.UserId).First(&current).Error; err != nil {
			return err
		}
		if current.Quota == target {
			return nil
		}
		if current.Quota != task.Quota {
			return ErrBillingConflict
		}
		delta := int64(target) - int64(current.Quota)
		if err := changeBillingWallet(tx, &user, -delta); err != nil {
			return err
		}
		entry := BillingEntry{EventKey: "task:" + common.GetUUID(), UserID: task.UserId, Kind: "task_adjustment", Quota: delta, WalletDelta: -delta, RequestID: task.TaskID, ModelName: modelName, TokenID: task.PrivateData.TokenId}
		if err := appendBillingEntry(tx, &user, &entry); err != nil {
			return err
		}
		if err := tx.Model(&current).Update("quota", target).Error; err != nil {
			return err
		}
		applied = true
		return nil
	})
	if err == nil {
		_ = invalidateUserCache(task.UserId)
	}
	return applied, err
}

func RefundBillingMidjourney(task *Midjourney, modelName string) (bool, error) {
	applied := false
	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := lockForUpdate(tx).First(&user, task.UserId).Error; err != nil {
			return err
		}
		var current Midjourney
		if err := lockForUpdate(tx).Where("id = ? AND user_id = ?", task.Id, task.UserId).First(&current).Error; err != nil {
			return err
		}
		if current.Quota == 0 {
			return nil
		}
		if current.Quota != task.Quota || current.Quota < 0 {
			return ErrBillingConflict
		}
		if err := changeBillingWallet(tx, &user, int64(current.Quota)); err != nil {
			return err
		}
		entry := BillingEntry{EventKey: "mj-refund:" + common.GetUUID(), UserID: task.UserId, Kind: "task_adjustment", Quota: -int64(current.Quota), WalletDelta: int64(current.Quota), RequestID: task.MjId, ModelName: modelName, TokenID: task.TokenId}
		if err := appendBillingEntry(tx, &user, &entry); err != nil {
			return err
		}
		if err := tx.Model(&current).Update("quota", 0).Error; err != nil {
			return err
		}
		applied = true
		return nil
	})
	if err == nil {
		_ = invalidateUserCache(task.UserId)
	}
	return applied, err
}
