package model

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type BillingPreparationState struct {
	Account           BillingAccount
	Hours             []BillingHour
	Pending           int64
	ExistingStatement string
	MissingHours      bool
}

// A read-only snapshot of prerequisites; CreateBillingStatement rechecks the
// mutable conditions under the account lock when the administrator submits.
func GetBillingPreparationState(ctx context.Context, userID int, start, end int64) (*BillingPreparationState, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	state := &BillingPreparationState{Account: BillingAccount{UserID: userID}}
	err := DB.WithContext(queryCtx).Transaction(func(tx *gorm.DB) error {
		err := tx.Where("user_id = ?", userID).First(&state.Account).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if state.Account.AccountingStartAt == 0 || state.Account.AccountingStartAt >= end {
			return nil
		}
		if start < state.Account.AccountingStartAt {
			start = state.Account.AccountingStartAt
		}
		if err := tx.Where("user_id = ? AND hour >= ? AND hour < ?", userID, start/3600*3600, end).Order("hour asc").Find(&state.Hours).Error; err != nil {
			return err
		}
		if len(state.Hours) == 0 {
			var entry BillingEntry
			err := tx.Select("id").Where("user_id = ? AND posted_at >= ? AND posted_at < ? AND sequence >= ? AND kind <> ?", userID, start, end, state.Account.StartSequence, "funding").Limit(1).Find(&entry).Error
			if err != nil {
				return err
			}
			state.MissingHours = entry.ID != 0
		}
		if err := tx.Model(&BillingOperation{}).Where("user_id = ? AND state = ? AND created_at < ?", userID, "reserved", end).Count(&state.Pending).Error; err != nil {
			return err
		}
		var existing BillingStatement
		if err := tx.Select("id").Where("user_id = ? AND month = ? AND status <> ?", userID, time.Unix(end-1, 0).In(time.FixedZone("Asia/Shanghai", 8*3600)).Format("2006-01"), StatementVoid).Order("revision desc").Limit(1).Find(&existing).Error; err != nil {
			return err
		}
		state.ExistingStatement = existing.ID
		return nil
	})
	return state, err
}
