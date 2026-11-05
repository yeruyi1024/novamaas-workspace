package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

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
	if err := tx.Model(&asset).Updates(map[string]any{"status": AssetStatusDeleted, "updated_at": now}).Error; err != nil {
		return err
	}
	asset.Status = AssetStatusDeleted
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
	})
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
