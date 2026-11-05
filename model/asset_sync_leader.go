package model

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const AssetSyncLeaderLockType = "asset_sync_scheduler"

// TryHoldAssetSyncLeadership uses one row in the existing task-lock table.
// Database time keeps election independent of node clocks and Redis databases.
func TryHoldAssetSyncLeadership(ctx context.Context, owner string, seconds int64) (held bool, err error) {
	if DB == nil || owner == "" || seconds <= 0 {
		return false, errors.New("asset synchronization leadership is not configured")
	}
	err = DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lock := SystemTaskLock{Type: AssetSyncLeaderLockType, TaskID: AssetSyncLeaderLockType}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&lock).Error; err != nil {
			return err
		}
		if err := lockForUpdate(tx).Where("type = ?", AssetSyncLeaderLockType).First(&lock).Error; err != nil {
			return err
		}
		query := "SELECT CAST(strftime('%s', 'now') AS INTEGER)"
		switch tx.Dialector.Name() {
		case "mysql":
			query = "SELECT CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP) AS SIGNED)"
		case "postgres":
			query = "SELECT CAST(FLOOR(EXTRACT(EPOCH FROM clock_timestamp())) AS BIGINT)"
		}
		var now int64
		if err := tx.Raw(query).Scan(&now).Error; err != nil {
			return err
		}
		if now <= 0 {
			return errors.New("asset synchronization database returned an invalid time")
		}
		if lock.LockedBy != owner && lock.LockedUntil > now {
			return nil
		}
		if err := tx.Model(&lock).Updates(map[string]any{"locked_by": owner, "locked_until": now + seconds, "updated_at": now}).Error; err != nil {
			return err
		}
		held = true
		return nil
	})
	return held && err == nil, err
}
