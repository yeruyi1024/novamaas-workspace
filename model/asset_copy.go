package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

var ErrAssetCopyConflict = errors.New("asset copy conflict")
var ErrAssetCopyPermissionDenied = errors.New("asset copy permission denied")

const assetCopyInsertBatchSize = 200

// AssetCopyBatch records the exact committed source snapshot. Cleanup never
// expands this manifest to include files uploaded after the copy.
type AssetCopyBatch struct {
	ID               string `json:"id" gorm:"type:varchar(64);primaryKey"`
	OperatorUserID   int    `json:"-"`
	SourceUserID     int    `json:"source_user_id" gorm:"index"`
	TargetUserID     int    `json:"target_user_id" gorm:"index"`
	RequestHash      string `json:"-" gorm:"type:char(64)"`
	Manifest         string `json:"-" gorm:"size:4194304"`
	CreatedAt        int64  `json:"created_at"`
	SourcesDeletedAt int64  `json:"sources_deleted_at"`
}

type AssetCopyPair struct {
	SourceID int64 `json:"source_id"`
	TargetID int64 `json:"target_id"`
}

type AssetCopiedGroup struct {
	SourceID       int64                  `json:"source_id"`
	TargetPublicID string                 `json:"target_public_id"`
	Name           string                 `json:"name"`
	Assets         []AssetCopyPair        `json:"assets"`
	Destinations   []AssetCopyDestination `json:"destinations,omitempty"`
}

type AssetCopyDestination struct {
	LibraryID int `json:"library_id,omitempty"`
	ChannelID int `json:"channel_id,omitempty"`
}

type AssetCopyProgress struct {
	ReadyCount       int  `json:"ready_count"`
	ProcessingCount  int  `json:"processing_count"`
	FailedCount      int  `json:"failed_count"`
	UnavailableCount int  `json:"unavailable_count"`
	CanDeleteSources bool `json:"can_delete_sources"`
}

// A successful asset is usable on every required account. Terminal failures
// also finish processing; this read never locks assets or replicas.
func GetAssetCopyProgress(tx *gorm.DB, ownerID int, groups []AssetCopiedGroup) (AssetCopyProgress, error) {
	progress := AssetCopyProgress{}
	var ids []int64
	for _, group := range groups {
		for _, pair := range group.Assets {
			ids = append(ids, pair.TargetID)
		}
	}
	if len(ids) == 0 {
		progress.CanDeleteSources = true
		return progress, nil
	}
	var assets []MediaAsset
	if err := tx.Where("id IN ? AND owner_user_id = ?", ids, ownerID).Find(&assets).Error; err != nil {
		return progress, err
	}
	var replicas []AssetReplica
	if err := tx.Where("asset_id IN ?", ids).Find(&replicas).Error; err != nil {
		return progress, err
	}
	statuses := make(map[int64]string, len(assets))
	for _, asset := range assets {
		statuses[asset.ID] = asset.Status
	}
	byAsset := make(map[int64][]AssetReplica)
	for _, replica := range replicas {
		byAsset[replica.AssetID] = append(byAsset[replica.AssetID], replica)
	}
	for _, group := range groups {
		for _, pair := range group.Assets {
			status := statuses[pair.TargetID]
			if status == "" || status == AssetStatusDeleted || status == AssetStatusUnavailable {
				progress.UnavailableCount++
				continue
			}
			current := byAsset[pair.TargetID]
			ready := status == AssetStatusReady && len(current) > 0
			pending := false
			active := make(map[AssetCopyDestination]bool, len(current))
			for _, replica := range current {
				usable := replica.Operation == AssetReplicaOperationSync && replica.Status == AssetReplicaStatusActive && replica.UpstreamAssetID != ""
				ready = ready && usable
				if replica.Operation == AssetReplicaOperationSync {
					switch replica.Status {
					case AssetReplicaStatusPending, AssetReplicaStatusSyncing, AssetReplicaStatusProcessing:
						pending = true
					case AssetReplicaStatusFailed:
						pending = pending || replica.NextSyncAt > 0
					}
				}
				destination := AssetCopyDestination{ChannelID: replica.ChannelID}
				if replica.LibraryID != nil {
					destination = AssetCopyDestination{LibraryID: *replica.LibraryID}
				}
				active[destination] = usable
			}
			for _, destination := range group.Destinations {
				ready = ready && active[destination]
			}
			switch {
			case ready:
				progress.ReadyCount++
			case pending:
				progress.ProcessingCount++
			default:
				progress.FailedCount++
			}
		}
	}
	progress.CanDeleteSources = progress.ProcessingCount == 0
	return progress, nil
}

// MigrateAssetCopyReferences relaxes legacy one-asset-per-file and global
// public_id uniqueness so copied groups and assets can preserve their original
// public IDs under the target owner while sharing the permanent file.
func MigrateAssetCopyReferences(db *gorm.DB) error {
	if db.Migrator().HasTable(&MediaAsset{}) && db.Migrator().HasIndex(&MediaAsset{}, "idx_media_assets_storage_object_id") {
		if err := db.Migrator().DropIndex(&MediaAsset{}, "idx_media_assets_storage_object_id"); err != nil {
			return err
		}
	}
	for _, target := range []struct {
		model any
		name  string
	}{
		{model: &AssetGroup{}, name: "AssetGroup"},
		{model: &MediaAsset{}, name: "MediaAsset"},
	} {
		if !db.Migrator().HasTable(target.model) {
			continue
		}
		if err := db.Model(target.model).
			Where("status = ? AND deleted_slot = 0", AssetStatusDeleted).
			Update("deleted_slot", gorm.Expr("id")).Error; err != nil {
			return err
		}
		if db.Dialector.Name() == "sqlite" {
			table := db.NamingStrategy.TableName(target.name)
			var list []struct {
				Name   string `gorm:"column:name"`
				Unique int    `gorm:"column:unique"`
				Origin string `gorm:"column:origin"`
			}
			if err := db.Raw("PRAGMA index_list(`" + table + "`)").Scan(&list).Error; err != nil {
				return err
			}
			for _, item := range list {
				if item.Unique != 1 || item.Origin != "c" {
					continue
				}
				var cols []struct {
					Name string `gorm:"column:name"`
				}
				if err := db.Raw("PRAGMA index_info(`" + item.Name + "`)").Scan(&cols).Error; err != nil {
					return err
				}
				if len(cols) == 1 && strings.EqualFold(cols[0].Name, "public_id") {
					if err := db.Migrator().DropIndex(target.model, item.Name); err != nil {
						return err
					}
				}
			}
			continue
		}
		indexes, err := db.Migrator().GetIndexes(target.model)
		if err != nil {
			return err
		}
		for _, index := range indexes {
			unique, ok := index.Unique()
			cols := index.Columns()
			if !ok || !unique || len(cols) != 1 || !strings.EqualFold(cols[0], "public_id") {
				continue
			}
			if err := db.Migrator().DropIndex(target.model, index.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}

// CheckAssetCopyUsers rereads roles and statuses inside the copy transaction
// without row locks. Locking user rows would stall quota updates for the
// source, target and operator until a large copy commits.
func CheckAssetCopyUsers(tx *gorm.DB, sourceID, targetID, operatorID int) error {
	var users []User
	err := tx.Select("id", "status", "role").Where("id IN ?", []int{sourceID, targetID, operatorID}).Find(&users).Error
	if err != nil {
		return err
	}
	var operator User
	participants := make([]User, 0, 2)
	for _, user := range users {
		if user.Id == operatorID {
			operator = user
		}
		if user.Id == sourceID || user.Id == targetID {
			participants = append(participants, user)
		}
	}
	// An earlier HTTP permission check may predate a promotion or demotion.
	if operator.Id == 0 || operator.Status != common.UserStatusEnabled || operator.Role < common.RoleAdminUser {
		return ErrAssetCopyPermissionDenied
	}
	if len(participants) != 2 {
		return gorm.ErrRecordNotFound
	}
	for _, user := range participants {
		if user.Id == targetID && user.Status != common.UserStatusEnabled {
			return fmt.Errorf("%w: target user is disabled", ErrAssetCopyConflict)
		}
		if operator.Role != common.RoleRootUser && operator.Role <= user.Role {
			return ErrAssetCopyPermissionDenied
		}
	}
	return nil
}

func LockAssetGroup(tx *gorm.DB, groupID int64) (*AssetGroup, error) {
	var group AssetGroup
	err := lockForUpdate(tx).First(&group, groupID).Error
	return &group, err
}

func NewAssetPublicID(prefix string) (string, error) {
	value, err := common.GenerateRandomCharsKey(16)
	if err != nil {
		return "", err
	}
	return prefix + time.Now().UTC().Format("20060102150405") + "-" + value, nil
}

// CopyAssetGroupsWithTx gives the target independent groups and assets that
// reuse the source files. Each copy is synchronized like a fresh upload to
// channelIDs; copies of rejected assets stay unavailable and are not resent.
// Lock order matches deletion: source groups, source assets, then files.
func CopyAssetGroupsWithTx(tx *gorm.DB, batch *AssetCopyBatch, publicIDs []string, channelIDs []int) ([]MediaAsset, error) {
	var bindings []AssetChannelConfig
	if len(channelIDs) > 0 {
		if err := tx.Select("channel_id", "library_id").Where("channel_id IN ?", channelIDs).Find(&bindings).Error; err != nil {
			return nil, err
		}
	}
	libraries := make(map[int]int, len(bindings))
	for _, binding := range bindings {
		libraries[binding.ChannelID] = binding.LibraryID
	}
	// Channels sharing one asset-library account receive one upstream copy.
	var destinations []AssetReplica
	queued := map[string]bool{}
	for _, channelID := range channelIDs {
		destination, scope := AssetReplica{ChannelID: channelID}, fmt.Sprintf("channel:%d", channelID)
		if libraryID := libraries[channelID]; libraryID > 0 {
			destination = AssetReplica{LibraryID: &libraryID}
			scope = fmt.Sprintf("library:%d", libraryID)
		}
		if !queued[scope] {
			queued[scope] = true
			destinations = append(destinations, destination)
		}
	}

	var groups []AssetGroup
	if err := lockForUpdate(tx).Where("public_id IN ? AND owner_user_id = ? AND status <> ?", publicIDs, batch.SourceUserID, AssetStatusDeleted).
		Order("id asc").Find(&groups).Error; err != nil {
		return nil, err
	}
	if len(groups) != len(publicIDs) {
		return nil, fmt.Errorf("%w: a source group is missing or belongs to another user", ErrAssetCopyConflict)
	}
	groupIDs := make([]int64, 0, len(groups))
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID)
	}
	var sources []MediaAsset
	if err := lockForUpdate(tx).Where("group_id IN ? AND owner_user_id = ? AND status <> ?", groupIDs, batch.SourceUserID, AssetStatusDeleted).
		Order("id asc").Find(&sources).Error; err != nil {
		return nil, err
	}
	if len(sources) > 0 {
		objectIDs := make([]int64, 0, len(sources))
		for _, source := range sources {
			objectIDs = append(objectIDs, source.StorageObjectID)
		}
		var uploaded []int64
		if err := lockForUpdate(tx).Model(&StorageObject{}).Where("id IN ? AND status = ?", objectIDs, StorageObjectStatusUploaded).
			Order("id asc").Pluck("id", &uploaded).Error; err != nil {
			return nil, err
		}
		available := make(map[int64]bool, len(uploaded))
		for _, id := range uploaded {
			available[id] = true
		}
		for _, source := range sources {
			if !available[source.StorageObjectID] {
				return nil, fmt.Errorf("%w: a source file is unavailable; refresh and retry", ErrAssetCopyConflict)
			}
		}
	}

	var existingGroups []AssetGroup
	if err := tx.Select("id").Where("owner_user_id = ? AND public_id IN ? AND status <> ?", batch.TargetUserID, publicIDs, AssetStatusDeleted).
		Limit(1).Find(&existingGroups).Error; err != nil {
		return nil, err
	}
	if len(existingGroups) > 0 {
		return nil, fmt.Errorf("%w: target user already has one or more of the selected asset groups", ErrAssetCopyConflict)
	}
	if len(sources) > 0 {
		sourcePublicIDs := make([]string, 0, len(sources))
		for _, source := range sources {
			sourcePublicIDs = append(sourcePublicIDs, source.PublicID)
		}
		var existingAssets []MediaAsset
		if err := tx.Select("id").Where("owner_user_id = ? AND public_id IN ? AND status <> ?", batch.TargetUserID, sourcePublicIDs, AssetStatusDeleted).
			Limit(1).Find(&existingAssets).Error; err != nil {
			return nil, err
		}
		if len(existingAssets) > 0 {
			return nil, fmt.Errorf("%w: target user already has one or more assets with the same asset ID", ErrAssetCopyConflict)
		}
	}

	targets := make([]AssetGroup, 0, len(groups))
	for _, group := range groups {
		targets = append(targets, AssetGroup{PublicID: group.PublicID, OwnerUserID: batch.TargetUserID, Name: group.Name, Description: group.Description, Status: group.Status})
	}
	if err := tx.CreateInBatches(&targets, assetCopyInsertBatchSize).Error; err != nil {
		return nil, err
	}
	targetGroups := make(map[int64]int64, len(groups))
	for index, group := range groups {
		targetGroups[group.ID] = targets[index].ID
	}
	copies := make([]MediaAsset, 0, len(sources))
	for _, source := range sources {
		copy := source
		copy.ID, copy.DeletedSlot, copy.CreatedAt, copy.UpdatedAt = 0, 0, 0, 0
		copy.OwnerUserID, copy.GroupID = batch.TargetUserID, targetGroups[source.GroupID]
		copy.Status, copy.UnavailableReason = AssetStatusProcessing, ""
		if len(destinations) == 0 {
			copy.Status = AssetStatusReady
		}
		if source.Status == AssetStatusUnavailable {
			copy.Status, copy.UnavailableReason = AssetStatusUnavailable, source.UnavailableReason
		}
		copies = append(copies, copy)
	}
	if len(copies) > 0 {
		if err := tx.CreateInBatches(&copies, assetCopyInsertBatchSize).Error; err != nil {
			return nil, err
		}
	}
	now := common.GetTimestamp()
	var replicas []AssetReplica
	for _, copy := range copies {
		if copy.Status != AssetStatusProcessing {
			continue
		}
		for _, destination := range destinations {
			replica := destination
			replica.AssetID, replica.Operation, replica.Status, replica.NextSyncAt = copy.ID, AssetReplicaOperationSync, AssetReplicaStatusPending, now
			replicas = append(replicas, replica)
		}
	}
	if len(replicas) > 0 {
		if err := tx.CreateInBatches(&replicas, assetCopyInsertBatchSize).Error; err != nil {
			return nil, err
		}
	}

	manifest := make([]AssetCopiedGroup, 0, len(groups))
	var required []AssetCopyDestination
	for _, destination := range destinations {
		scope := AssetCopyDestination{ChannelID: destination.ChannelID}
		if destination.LibraryID != nil {
			scope = AssetCopyDestination{LibraryID: *destination.LibraryID}
		}
		required = append(required, scope)
	}
	entries := make(map[int64]int, len(groups))
	for index, group := range groups {
		entries[group.ID] = len(manifest)
		manifest = append(manifest, AssetCopiedGroup{SourceID: group.ID, TargetPublicID: targets[index].PublicID, Name: group.Name, Assets: []AssetCopyPair{}, Destinations: required})
	}
	for index, source := range sources {
		entry := &manifest[entries[source.GroupID]]
		entry.Assets = append(entry.Assets, AssetCopyPair{SourceID: source.ID, TargetID: copies[index].ID})
	}
	encoded, err := common.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	batch.Manifest, batch.CreatedAt = string(encoded), common.GetTimestamp()
	return copies, tx.Create(batch).Error
}

// DeleteAssetCopySources cleans the source snapshot uniformly after processing
// finishes. Later uploads and sources moved out of the copied groups stay.
func DeleteAssetCopySources(batchID string, operatorID int) (*AssetCopyBatch, error) {
	var batch AssetCopyBatch
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := lockForUpdate(tx).First(&batch, "id = ?", batchID).Error; err != nil {
			return err
		}
		if err := CheckAssetCopyUsers(tx, batch.SourceUserID, batch.TargetUserID, operatorID); err != nil {
			return err
		}
		if batch.SourcesDeletedAt != 0 {
			return nil
		}
		var manifest []AssetCopiedGroup
		if err := common.UnmarshalJsonStr(batch.Manifest, &manifest); err != nil {
			return err
		}
		progress, err := GetAssetCopyProgress(tx, batch.TargetUserID, manifest)
		if err != nil {
			return err
		}
		if !progress.CanDeleteSources {
			return fmt.Errorf("%w: wait for the batch to finish before deleting sources", ErrAssetCopyConflict)
		}
		groupIDs := make([]int64, 0, len(manifest))
		var sourceIDs []int64
		copiedFrom := map[int64]int64{}
		for _, group := range manifest {
			groupIDs = append(groupIDs, group.SourceID)
			for _, pair := range group.Assets {
				sourceIDs = append(sourceIDs, pair.SourceID)
				copiedFrom[pair.SourceID] = group.SourceID
			}
		}
		var groups []AssetGroup
		if err := lockForUpdate(tx).Where("id IN ?", groupIDs).Order("id asc").Find(&groups).Error; err != nil {
			return err
		}
		var sources []MediaAsset
		if len(sourceIDs) > 0 {
			if err := lockForUpdate(tx).Where("id IN ?", sourceIDs).Order("id asc").Find(&sources).Error; err != nil {
				return err
			}
		}
		for _, source := range sources {
			if source.OwnerUserID != batch.SourceUserID || source.Status == AssetStatusDeleted ||
				source.GroupID != copiedFrom[source.ID] {
				continue
			}
			if err := RequestAssetDeletion(tx, source.ID); err != nil {
				return err
			}
		}
		for _, group := range groups {
			// A current read keeps groups that received uploads after the copy.
			var remaining []MediaAsset
			if err := lockForUpdate(tx).Select("id").Where("group_id = ? AND status <> ?", group.ID, AssetStatusDeleted).Limit(1).Find(&remaining).Error; err != nil {
				return err
			}
			if len(remaining) == 0 {
				if err := tx.Model(&group).Updates(map[string]any{"status": AssetStatusDeleted, "deleted_slot": group.ID, "updated_at": common.GetTimestamp()}).Error; err != nil {
					return err
				}
			}
		}
		batch.SourcesDeletedAt = common.GetTimestamp()
		return tx.Model(&batch).Update("sources_deleted_at", batch.SourcesDeletedAt).Error
	}, AssetDeletionTxOptions())
	return &batch, err
}
