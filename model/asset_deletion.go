package model

import (
	"database/sql"
	"errors"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// File-lock waiters must see the previous holder's committed cleanup without
// locking sibling assets or replicas. SQLite keeps its default isolation.
func AssetDeletionTxOptions() *sql.TxOptions {
	if DB.Dialector.Name() == "sqlite" {
		return nil
	}
	return &sql.TxOptions{Isolation: sql.LevelReadCommitted}
}

func lockAssetReplica(tx *gorm.DB, id int64) (*MediaAsset, *AssetReplica, error) {
	var reference AssetReplica
	if err := tx.Select("asset_id").First(&reference, id).Error; err != nil {
		return nil, nil, err
	}
	var asset MediaAsset
	if err := lockForUpdate(tx).First(&asset, reference.AssetID).Error; err != nil {
		return nil, nil, err
	}
	var replica AssetReplica
	if err := lockForUpdate(tx).First(&replica, id).Error; err != nil {
		return nil, nil, err
	}
	return &asset, &replica, nil
}

func ownsAssetReplica(replica *AssetReplica, owner string) bool {
	return owner != "" && replica.LockedBy == owner &&
		(replica.Status == AssetReplicaStatusSyncing || replica.Status == AssetReplicaStatusActive)
}

func assetDeletionRequested(asset *MediaAsset, replica *AssetReplica) bool {
	return asset.Status == AssetStatusDeleted || replica.Operation == AssetReplicaOperationDelete
}

func assetReplicaDeletionUpdates(now int64) map[string]any {
	return map[string]any{
		"operation": AssetReplicaOperationDelete, "status": AssetReplicaStatusDeleting, "next_sync_at": now,
		"locked_by": "", "lease_until": 0, "updated_at": now,
	}
}

// The asset lock serializes this final check with late submission results.
func queueDeletedAssetStorage(tx *gorm.DB, asset *MediaAsset, now int64) error {
	if asset.Status != AssetStatusDeleted || asset.StorageObjectID <= 0 {
		return nil
	}
	var replicas []AssetReplica
	if err := lockForUpdate(tx).Select("status").Where("asset_id = ?", asset.ID).Find(&replicas).Error; err != nil {
		return err
	}
	for _, replica := range replicas {
		if replica.Status != AssetReplicaStatusDeleted {
			return nil
		}
	}
	// Copied assets share the file. The file row serializes this check with
	// copying and other final deletions. Reading the sharing assets without
	// locks avoids deadlocks; a stale snapshot can only show an asset as still
	// live, because a copy needs a live source and deletion is never undone.
	var object StorageObject
	if err := lockForUpdate(tx).Select("id").Where("id = ?", asset.StorageObjectID).Limit(1).Find(&object).Error; err != nil || object.ID == 0 {
		return err
	}
	var sharing []MediaAsset
	if err := tx.Select("id").Where("storage_object_id = ? AND id <> ? AND status <> ?", asset.StorageObjectID, asset.ID, AssetStatusDeleted).
		Limit(1).Find(&sharing).Error; err != nil {
		return err
	}
	if len(sharing) > 0 {
		return nil
	}
	// Other deleted assets may still be cleaning their upstream copies. Read
	// without sibling row locks so cleanup retains the existing lock order.
	var unfinished []AssetReplica
	sharedIDs := tx.Model(&MediaAsset{}).Select("id").Where("storage_object_id = ?", asset.StorageObjectID)
	if err := tx.Select("id").Where("asset_id IN (?) AND status <> ?", sharedIDs, AssetReplicaStatusDeleted).
		Limit(1).Find(&unfinished).Error; err != nil {
		return err
	}
	if len(unfinished) > 0 {
		return nil
	}
	return tx.Model(&StorageObject{}).Where("id = ? AND status NOT IN ?", asset.StorageObjectID,
		[]string{StorageObjectStatusDeletePending, StorageObjectStatusDeleting, StorageObjectStatusDeleted}).
		Updates(map[string]any{"status": StorageObjectStatusDeletePending, "delete_after": now, "last_error": "", "updated_at": now}).Error
}

// RequestAssetDeletion must run inside the caller's transaction. Live creation
// leases remain owned until their returned IDs are saved for upstream deletion.
func RequestAssetDeletion(tx *gorm.DB, assetID int64) error {
	var asset MediaAsset
	if err := lockForUpdate(tx).First(&asset, assetID).Error; err != nil {
		return err
	}
	now := common.GetTimestamp()
	if err := tx.Model(&asset).Updates(map[string]any{"status": AssetStatusDeleted, "deleted_slot": asset.ID, "updated_at": now}).Error; err != nil {
		return err
	}
	asset.Status = AssetStatusDeleted
	asset.DeletedSlot = asset.ID
	var replicas []AssetReplica
	if err := lockForUpdate(tx).Where("asset_id = ?", assetID).Order("id asc").Find(&replicas).Error; err != nil {
		return err
	}
	for index := range replicas {
		replica := &replicas[index]
		updates := assetReplicaDeletionUpdates(now)
		if replica.LockedBy != "" && replica.LeaseUntil > now && ownsAssetReplica(replica, replica.LockedBy) {
			delete(updates, "status")
			delete(updates, "locked_by")
			delete(updates, "lease_until")
		} else if replica.Status == AssetReplicaStatusDeleted || replica.UpstreamAssetID == "" {
			updates["status"], updates["next_sync_at"], updates["progress"] = AssetReplicaStatusDeleted, 0, 100
		}
		updates["last_error"] = ""
		if err := tx.Model(replica).Updates(updates).Error; err != nil {
			return err
		}
	}
	return queueDeletedAssetStorage(tx, &asset, now)
}

type UserAssetPurgeResult struct {
	OwnerUserID       int `json:"owner_user_id"`
	DeletedAssetCount int `json:"deleted_asset_count"`
	DeletedGroupCount int `json:"deleted_group_count"`
}

const assetPurgeBatchSize = 100

func checkUserAssetPurgePermission(tx *gorm.DB, targetUserID, operatorID int) error {
	var users []User
	if err := tx.Select("id", "status", "role").Where("id IN ?", []int{targetUserID, operatorID}).Find(&users).Error; err != nil {
		return err
	}
	var operator, target User
	for _, user := range users {
		if user.Id == operatorID {
			operator = user
		}
		if user.Id == targetUserID {
			target = user
		}
	}
	if operator.Id == 0 || operator.Status != common.UserStatusEnabled || operator.Role < common.RoleAdminUser {
		return ErrAssetCopyPermissionDenied
	}
	if target.Id == 0 {
		return gorm.ErrRecordNotFound
	}
	if operator.Role != common.RoleRootUser && operator.Role <= target.Role {
		return ErrAssetCopyPermissionDenied
	}
	return nil
}

// purgeUserAssetsInBatches clears a user's active assets in bounded short
// transactions (100 assets per batch) followed by one bulk update for empty
// groups. On partial failure it returns the already-committed counts alongside
// the error so callers can still record an accurate audit log.
func purgeUserAssetsInBatches(ownerUserID int, beforeBatch func(tx *gorm.DB) error) (*UserAssetPurgeResult, error) {
	result := &UserAssetPurgeResult{OwnerUserID: ownerUserID}
	if ownerUserID <= 0 {
		return result, nil
	}
	if DB.Migrator().HasTable(&MediaAsset{}) {
		for {
			batchDeleted := 0
			err := DB.Transaction(func(tx *gorm.DB) error {
				if beforeBatch != nil {
					if err := beforeBatch(tx); err != nil {
						return err
					}
				}
				var assets []MediaAsset
				if err := lockForUpdate(tx).Select("id").
					Where("owner_user_id = ? AND status <> ?", ownerUserID, AssetStatusDeleted).
					Order("id asc").Limit(assetPurgeBatchSize).Find(&assets).Error; err != nil {
					return err
				}
				for _, asset := range assets {
					if err := RequestAssetDeletion(tx, asset.ID); err != nil {
						return err
					}
				}
				batchDeleted = len(assets)
				return nil
			}, AssetDeletionTxOptions())
			if err != nil {
				return result, err
			}
			result.DeletedAssetCount += batchDeleted
			if batchDeleted < assetPurgeBatchSize {
				break
			}
		}
	}
	if DB.Migrator().HasTable(&AssetGroup{}) {
		err := DB.Transaction(func(tx *gorm.DB) error {
			if beforeBatch != nil {
				if err := beforeBatch(tx); err != nil {
					return err
				}
			}
			var groups []AssetGroup
			if err := lockForUpdate(tx).Select("id").
				Where("owner_user_id = ? AND status <> ?", ownerUserID, AssetStatusDeleted).
				Order("id asc").Find(&groups).Error; err != nil {
				return err
			}
			if len(groups) == 0 {
				return nil
			}
			groupIDs := make([]int64, 0, len(groups))
			for _, group := range groups {
				groupIDs = append(groupIDs, group.ID)
			}
			// Keep any group that received a concurrent upload while assets were being purged.
			var nonEmptyGroupIDs []int64
			if tx.Migrator().HasTable(&MediaAsset{}) {
				if err := tx.Model(&MediaAsset{}).Distinct("group_id").
					Where("group_id IN ? AND status <> ?", groupIDs, AssetStatusDeleted).
					Pluck("group_id", &nonEmptyGroupIDs).Error; err != nil {
					return err
				}
			}
			nonEmpty := make(map[int64]bool, len(nonEmptyGroupIDs))
			for _, id := range nonEmptyGroupIDs {
				nonEmpty[id] = true
			}
			emptyGroupIDs := make([]int64, 0, len(groupIDs))
			for _, id := range groupIDs {
				if !nonEmpty[id] {
					emptyGroupIDs = append(emptyGroupIDs, id)
				}
			}
			if len(emptyGroupIDs) == 0 {
				return nil
			}
			now := common.GetTimestamp()
			if err := tx.Model(&AssetGroup{}).Where("id IN ?", emptyGroupIDs).Updates(map[string]any{
				"status":       AssetStatusDeleted,
				"deleted_slot": gorm.Expr("id"),
				"updated_at":   now,
			}).Error; err != nil {
				return err
			}
			result.DeletedGroupCount = len(emptyGroupIDs)
			return nil
		}, AssetDeletionTxOptions())
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

// PurgeUserAssets clears all active asset groups and assets owned by targetUserID
// in short bounded transactions (100 assets per batch) to avoid long locks on
// large accounts, while keeping the account, AK/SK keys, and webhook settings intact.
func PurgeUserAssets(targetUserID, operatorID int) (*UserAssetPurgeResult, error) {
	if targetUserID <= 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if err := checkUserAssetPurgePermission(DB, targetUserID, operatorID); err != nil {
		return nil, err
	}
	return purgeUserAssetsInBatches(targetUserID, func(tx *gorm.DB) error {
		return checkUserAssetPurgePermission(tx, targetUserID, operatorID)
	})
}

func CanSubmitAssetReplica(id int64, owner string) (allowed bool, err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		asset, replica, err := lockAssetReplica(tx, id)
		if err != nil {
			return err
		}
		allowed = ownsAssetReplica(replica, owner) && replica.LeaseUntil > common.GetTimestamp() &&
			replica.Status == AssetReplicaStatusSyncing && replica.UpstreamAssetID == "" &&
			!assetDeletionRequested(asset, replica) &&
			(asset.Status == AssetStatusProcessing || asset.Status == AssetStatusReady)
		return nil
	})
	return allowed, err
}

// FinishAssetReplicaSubmission preserves a returned ID when deletion overtakes
// an in-flight submission; the original lease stays owned until this point.
func FinishAssetReplicaSubmission(id int64, owner string, upstreamID string, nextSyncAt int64) error {
	if upstreamID == "" {
		return errors.New("upstream asset ID is required")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		asset, replica, err := lockAssetReplica(tx, id)
		if err != nil {
			return err
		}
		now := common.GetTimestamp()
		owned := ownsAssetReplica(replica, owner)
		if owned && !assetDeletionRequested(asset, replica) {
			return tx.Model(replica).Updates(map[string]any{
				"status": AssetReplicaStatusProcessing, "progress": 50, "upstream_asset_id": upstreamID,
				"next_sync_at": nextSyncAt, "locked_by": "", "lease_until": 0,
				"last_error": "", "last_synced_at": now, "updated_at": now,
			}).Error
		}
		if !assetDeletionRequested(asset, replica) {
			return errors.New("asset submission lease was lost before its returned ID could be saved")
		}
		updates := assetReplicaDeletionUpdates(now)
		updates["upstream_asset_id"] = upstreamID
		if replica.LockedBy != "" && replica.LockedBy != owner && replica.LeaseUntil > now {
			delete(updates, "status")
			delete(updates, "locked_by")
			delete(updates, "lease_until")
		}
		return tx.Model(replica).Updates(updates).Error
	})
}

// AcknowledgeAssetReplicaDeletion checks the current ID before final cleanup.
func AcknowledgeAssetReplicaDeletion(id int64, owner, upstreamID string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		asset, replica, err := lockAssetReplica(tx, id)
		if err != nil {
			return err
		}
		if !ownsAssetReplica(replica, owner) || !assetDeletionRequested(asset, replica) {
			return errors.New("asset replica deletion lease was lost")
		}
		now := common.GetTimestamp()
		updates := assetReplicaDeletionUpdates(now)
		updates["last_error"] = ""
		if upstreamID != "" {
			updates["last_synced_at"] = now
		}
		// If a late ID replaced the one just deleted, release the lease and
		// delete that ID next. Never clear it with the older HTTP response.
		if replica.UpstreamAssetID == "" || replica.UpstreamAssetID == upstreamID {
			updates["status"], updates["next_sync_at"], updates["progress"] = AssetReplicaStatusDeleted, 0, 100
		}
		if err := tx.Model(replica).Updates(updates).Error; err != nil {
			return err
		}
		return queueDeletedAssetStorage(tx, asset, now)
	}, AssetDeletionTxOptions())
}

func RetryAssetReplicaSynchronization(id int64) (retried bool, err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		asset, replica, err := lockAssetReplica(tx, id)
		if err != nil {
			return err
		}
		now := common.GetTimestamp()
		if assetDeletionRequested(asset, replica) || replica.Status != AssetReplicaStatusFailed ||
			(replica.LockedBy != "" && replica.LeaseUntil > now) || asset.Status == AssetStatusUnavailable {
			return nil
		}
		status, progress := AssetReplicaStatusPending, 0
		if replica.UpstreamAssetID != "" {
			status, progress = AssetReplicaStatusProcessing, 50
		}
		if err := tx.Model(replica).Updates(map[string]any{"status": status, "progress": progress, "created_at": now,
			"attempts": 0, "last_error": "", "next_sync_at": now, "locked_by": "", "lease_until": 0, "updated_at": now}).Error; err != nil {
			return err
		}
		if asset.Status == AssetStatusSyncFailed {
			if err := tx.Model(asset).Updates(map[string]any{"status": AssetStatusProcessing, "updated_at": now}).Error; err != nil {
				return err
			}
		}
		retried = true
		return nil
	})
	return retried, err
}
