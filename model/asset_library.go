package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	AssetTypeImage = "image"
	AssetTypeVideo = "video"
	AssetTypeAudio = "audio"

	AssetStatusProcessing  = "processing"
	AssetStatusReady       = "ready"
	AssetStatusUnavailable = "unavailable"
	AssetStatusSyncFailed  = "failed"
	AssetStatusDeleted     = "deleted"

	AssetUnavailableRealPerson       = "real_person"
	AssetUnavailableSensitiveContent = "sensitive_content"
	AssetUnavailablePolicyRejected   = "policy_rejected"

	AssetReplicaStatusPending    = "pending"
	AssetReplicaStatusSyncing    = "syncing"
	AssetReplicaStatusProcessing = "processing"
	AssetReplicaStatusActive     = "active"
	AssetReplicaStatusFailed     = "failed"
	AssetReplicaStatusRejected   = "rejected"
	AssetReplicaStatusDeleting   = "deleting"
	AssetReplicaStatusDeleted    = "deleted"

	AssetReplicaOperationSync   = "sync"
	AssetReplicaOperationDelete = "delete"

	AssetChannelProtocolVolcAction  = "volc_action"
	AssetChannelProtocolYoufangREST = "youfang_rest"

	AssetChannelAuthAKSK   = "ak_sk"
	AssetChannelAuthBearer = "bearer"

	AssetAccessKeyStatusEnabled  = "enabled"
	AssetAccessKeyStatusDisabled = "disabled"
)

// AssetGroup is the tenant-owned namespace exposed by this gateway. Upstream
// group identifiers are deliberately kept in AssetGroupReplica.
type AssetGroup struct {
	ID          int64  `json:"-" gorm:"primaryKey"`
	PublicID    string `json:"id" gorm:"type:varchar(64);uniqueIndex"`
	OwnerUserID int    `json:"owner_user_id" gorm:"index:idx_asset_groups_owner_status,priority:1"`
	Name        string `json:"name" gorm:"type:varchar(128)"`
	Description string `json:"description" gorm:"type:varchar(512)"`
	Status      string `json:"status" gorm:"type:varchar(32);index:idx_asset_groups_owner_status,priority:2"`
	CreatedAt   int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt   int64  `json:"updated_at" gorm:"bigint"`
}

// MediaAsset owns a permanent StorageObject. PublicID is the only identifier
// accepted from clients; upstream identifiers never leave replica records.
type MediaAsset struct {
	ID                int64  `json:"-" gorm:"primaryKey"`
	PublicID          string `json:"id" gorm:"type:varchar(64);uniqueIndex"`
	OwnerUserID       int    `json:"owner_user_id" gorm:"index:idx_media_assets_owner_status,priority:1"`
	GroupID           int64  `json:"-" gorm:"index"`
	GroupPublicID     string `json:"group_id" gorm:"-"`
	StorageObjectID   int64  `json:"-" gorm:"uniqueIndex"`
	Name              string `json:"name" gorm:"type:varchar(255)"`
	AssetType         string `json:"type" gorm:"type:varchar(16);index"`
	ContentType       string `json:"content_type" gorm:"type:varchar(128)"`
	Size              int64  `json:"size" gorm:"bigint"`
	SHA256            string `json:"sha256" gorm:"type:char(64);index"`
	Status            string `json:"status" gorm:"type:varchar(32);index:idx_media_assets_owner_status,priority:2"`
	UnavailableReason string `json:"unavailable_reason,omitempty" gorm:"type:varchar(32)"`
	CreatedAt         int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt         int64  `json:"updated_at" gorm:"bigint"`
}

// AssetChannelConfig is a channel-scoped integration configuration. Secrets
// are encrypted and never serialized. It intentionally lives outside Channel
// OtherSettings so normal channel list APIs cannot expose credentials.
type AssetChannelConfig struct {
	ID                   int64  `json:"id" gorm:"primaryKey"`
	ChannelID            int    `json:"channel_id" gorm:"uniqueIndex"`
	Enabled              bool   `json:"enabled"`
	Protocol             string `json:"protocol" gorm:"type:varchar(32)"`
	AuthType             string `json:"auth_type" gorm:"type:varchar(16)"`
	BaseURL              string `json:"base_url" gorm:"type:varchar(1024)"`
	Region               string `json:"region" gorm:"type:varchar(64)"`
	Service              string `json:"service" gorm:"type:varchar(64)"`
	APIVersion           string `json:"api_version" gorm:"type:varchar(32)"`
	ProjectName          string `json:"project_name" gorm:"type:varchar(128)"`
	QPM                  int    `json:"qpm"`
	AccessKeyID          string `json:"-" gorm:"type:varchar(255)"`
	AccessKeyHint        string `json:"access_key_hint" gorm:"type:varchar(16)"`
	EncryptedCredential  string `json:"-" gorm:"type:text"`
	CredentialKeyVersion string `json:"-" gorm:"type:varchar(32)"`
	CreatedAt            int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt            int64  `json:"updated_at" gorm:"bigint"`
}

// AssetAccessKey authenticates downstream clients that use Volcengine's
// HMAC-SHA256 Action API. The secret is encrypted at rest and is returned only
// once when the key is created.
type AssetAccessKey struct {
	ID                   int64  `json:"id" gorm:"primaryKey"`
	OwnerUserID          int    `json:"owner_user_id" gorm:"index:idx_asset_access_keys_owner_status,priority:1"`
	Name                 string `json:"name" gorm:"type:varchar(64)"`
	AccessKeyID          string `json:"access_key_id" gorm:"type:varchar(64);uniqueIndex"`
	SecretHint           string `json:"secret_hint" gorm:"type:varchar(16)"`
	EncryptedSecret      string `json:"-" gorm:"type:text"`
	CredentialKeyVersion string `json:"-" gorm:"type:varchar(32)"`
	Status               string `json:"status" gorm:"type:varchar(16);index:idx_asset_access_keys_owner_status,priority:2"`
	LastUsedAt           int64  `json:"last_used_at" gorm:"bigint"`
	CreatedAt            int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt            int64  `json:"updated_at" gorm:"bigint"`
}

type AssetGroupReplica struct {
	ID              int64  `json:"id" gorm:"primaryKey"`
	GroupID         int64  `json:"-" gorm:"uniqueIndex:idx_asset_group_channel,priority:1"`
	ChannelID       int    `json:"channel_id" gorm:"uniqueIndex:idx_asset_group_channel,priority:2;index"`
	UpstreamGroupID string `json:"upstream_group_id,omitempty" gorm:"type:varchar(255)"`
	Status          string `json:"status" gorm:"type:varchar(32);index"`
	Progress        int    `json:"progress"`
	LastError       string `json:"last_error,omitempty" gorm:"type:text"`
	LastSyncedAt    int64  `json:"last_synced_at" gorm:"bigint"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt       int64  `json:"updated_at" gorm:"bigint"`
}

type AssetReplica struct {
	ID                int64  `json:"id" gorm:"primaryKey"`
	AssetID           int64  `json:"-" gorm:"uniqueIndex:idx_asset_channel,priority:1"`
	ChannelID         int    `json:"channel_id" gorm:"uniqueIndex:idx_asset_channel,priority:2;index"`
	UpstreamAssetID   string `json:"upstream_asset_id,omitempty" gorm:"type:varchar(255)"`
	Operation         string `json:"-" gorm:"type:varchar(16);index"`
	Status            string `json:"status" gorm:"type:varchar(32);index:idx_asset_replicas_work,priority:1"`
	Progress          int    `json:"progress"`
	Attempts          int    `json:"attempts"`
	NextSyncAt        int64  `json:"next_sync_at" gorm:"bigint;index:idx_asset_replicas_work,priority:2"`
	LockedBy          string `json:"-" gorm:"type:varchar(128);index"`
	LeaseUntil        int64  `json:"-" gorm:"bigint;index"`
	LastError         string `json:"last_error,omitempty" gorm:"type:text"`
	LastSyncedAt      int64  `json:"last_synced_at" gorm:"bigint"`
	CreatedAt         int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt         int64  `json:"updated_at" gorm:"bigint"`
	ClaimedFromStatus string `json:"-" gorm:"-"`
}

func setAssetTimestamps(createdAt *int64, updatedAt *int64) {
	now := common.GetTimestamp()
	if *createdAt == 0 {
		*createdAt = now
	}
	*updatedAt = now
}

func (value *AssetGroup) BeforeCreate(_ *gorm.DB) error {
	setAssetTimestamps(&value.CreatedAt, &value.UpdatedAt)
	return nil
}

func (value *MediaAsset) BeforeCreate(_ *gorm.DB) error {
	setAssetTimestamps(&value.CreatedAt, &value.UpdatedAt)
	return nil
}

func (value *AssetChannelConfig) BeforeCreate(_ *gorm.DB) error {
	setAssetTimestamps(&value.CreatedAt, &value.UpdatedAt)
	return nil
}

func (value *AssetAccessKey) BeforeCreate(_ *gorm.DB) error {
	setAssetTimestamps(&value.CreatedAt, &value.UpdatedAt)
	return nil
}

func (value *AssetGroupReplica) BeforeCreate(_ *gorm.DB) error {
	setAssetTimestamps(&value.CreatedAt, &value.UpdatedAt)
	return nil
}

func (value *AssetReplica) BeforeCreate(_ *gorm.DB) error {
	setAssetTimestamps(&value.CreatedAt, &value.UpdatedAt)
	return nil
}

func FindAssetGroup(publicID string, ownerUserID int, includeAllOwners bool) (*AssetGroup, error) {
	query := DB.Where("public_id = ? AND status <> ?", publicID, AssetStatusDeleted)
	if !includeAllOwners {
		query = query.Where("owner_user_id = ?", ownerUserID)
	}
	var group AssetGroup
	if err := query.First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func FindAssetGroupByID(id int64) (*AssetGroup, error) {
	var group AssetGroup
	if err := DB.Where("id = ? AND status <> ?", id, AssetStatusDeleted).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func FindMediaAsset(publicID string, ownerUserID int, includeAllOwners bool) (*MediaAsset, error) {
	query := DB.Where("public_id = ? AND status <> ?", publicID, AssetStatusDeleted)
	if !includeAllOwners {
		query = query.Where("owner_user_id = ?", ownerUserID)
	}
	var asset MediaAsset
	if err := query.First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func QueueAssetReplicas(tx *gorm.DB, assetID int64, channelIDs []int) error {
	now := common.GetTimestamp()
	var asset MediaAsset
	if err := lockForUpdate(tx).Select("id", "status").First(&asset, assetID).Error; err != nil {
		return err
	}
	if asset.Status != AssetStatusProcessing && asset.Status != AssetStatusReady {
		return nil
	}
	for _, channelID := range channelIDs {
		replica := AssetReplica{
			AssetID:    assetID,
			ChannelID:  channelID,
			Operation:  AssetReplicaOperationSync,
			Status:     AssetReplicaStatusPending,
			Progress:   0,
			NextSyncAt: now,
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&replica).Error; err != nil {
			return err
		}
		// Configuration saves and bulk sync may run during an upload. Keep
		// successful copies, deletion intent, and live owners intact; only
		// failed copies receive a new attempt and timeout budget.
		if err := tx.Model(&AssetReplica{}).
			Where("asset_id = ? AND channel_id = ? AND operation = ? AND status = ?", assetID, channelID, AssetReplicaOperationSync, AssetReplicaStatusFailed).
			Where("locked_by = ? OR locked_by IS NULL OR lease_until <= ?", "", now).
			Updates(map[string]any{
				"status":     gorm.Expr("CASE WHEN upstream_asset_id <> ? THEN ? ELSE ? END", "", AssetReplicaStatusProcessing, AssetReplicaStatusPending),
				"progress":   gorm.Expr("CASE WHEN upstream_asset_id <> ? THEN ? ELSE ? END", "", 50, 0),
				"created_at": now, "next_sync_at": now, "attempts": 0, "last_error": "",
				"locked_by": "", "lease_until": 0, "updated_at": now,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

// AssetReplicaWorkload names the kind of upstream request a due replica needs.
// The scheduler shares each channel's request budget between workloads.
type AssetReplicaWorkload string

const (
	AssetReplicaWorkloadSubmit      AssetReplicaWorkload = "submit"
	AssetReplicaWorkloadPoll        AssetReplicaWorkload = "poll"
	AssetReplicaWorkloadMaintenance AssetReplicaWorkload = "maintenance"
	AssetReplicaWorkloadReview      AssetReplicaWorkload = "review"
)

func dueAssetReplicas(query *gorm.DB, now int64) *gorm.DB {
	return query.Where(
		"((status IN ? AND next_sync_at <= ?) OR (status = ? AND next_sync_at <= ? AND lease_until <= ?) OR (status = ? AND lease_until <= ?))",
		[]string{AssetReplicaStatusPending, AssetReplicaStatusProcessing, AssetReplicaStatusFailed, AssetReplicaStatusDeleting}, now,
		AssetReplicaStatusActive, now, now,
		AssetReplicaStatusSyncing, now,
	).Where("locked_by = ? OR locked_by IS NULL OR lease_until <= ?", "", now).
		Where("status <> ? OR next_sync_at > 0 OR operation = ?", AssetReplicaStatusFailed, AssetReplicaOperationDelete)
}

// ListDueAssetReplicas reads due work without leasing it. Callers lease only
// the replica they are about to send, so waiting work stays visible and keeps
// its place in the queue.
func ListDueAssetReplicas(now int64, channelID int, workload AssetReplicaWorkload, limit int) ([]*AssetReplica, error) {
	query := dueAssetReplicas(DB.Where("channel_id = ?", channelID), now)
	switch workload {
	case AssetReplicaWorkloadSubmit:
		query = query.Where("operation <> ? AND status <> ? AND progress < 100 AND upstream_asset_id = ?", AssetReplicaOperationDelete, AssetReplicaStatusActive, "")
	case AssetReplicaWorkloadPoll:
		query = query.Where("operation <> ? AND status <> ? AND progress < 100 AND upstream_asset_id <> ?", AssetReplicaOperationDelete, AssetReplicaStatusActive, "")
	case AssetReplicaWorkloadMaintenance:
		query = query.Where("operation = ?", AssetReplicaOperationDelete)
	case AssetReplicaWorkloadReview:
		query = query.Where("operation <> ? AND (status = ? OR progress >= 100)", AssetReplicaOperationDelete, AssetReplicaStatusActive)
	default:
		return nil, fmt.Errorf("invalid asset replica workload %q", workload)
	}
	var candidates []*AssetReplica
	err := query.Order("next_sync_at asc, id asc").Limit(limit).Find(&candidates).Error
	return candidates, err
}

// ClaimAssetReplica leases one candidate if it is still due. A false result
// means another worker took it or its state changed after it was read.
func ClaimAssetReplica(candidate *AssetReplica, now int64, leaseUntil int64, lockedBy string) (claimed bool, err error) {
	err = DB.Transaction(func(tx *gorm.DB) error {
		var current AssetReplica
		result := lockForUpdate(dueAssetReplicas(tx, now)).Where("id = ? AND status = ? AND operation = ?", candidate.ID, candidate.Status, candidate.Operation).Find(&current)
		if result.Error != nil || result.RowsAffected == 0 {
			return result.Error
		}
		status := AssetReplicaStatusSyncing
		if current.Status == AssetReplicaStatusActive {
			status = AssetReplicaStatusActive
		}
		if err := tx.Model(&current).Updates(map[string]any{"status": status, "locked_by": lockedBy, "lease_until": leaseUntil, "updated_at": now}).Error; err != nil {
			return err
		}
		from := candidate.Status
		*candidate = current
		candidate.ClaimedFromStatus, candidate.Status, candidate.LockedBy = from, status, lockedBy
		candidate.LeaseUntil = leaseUntil
		claimed = true
		return nil
	})
	return claimed, err
}

func ClaimAssetReplicas(now int64, leaseUntil int64, lockedBy string, limit int) ([]*AssetReplica, error) {
	if limit <= 0 {
		limit = 20
	}
	var candidates []*AssetReplica
	if err := dueAssetReplicas(DB, now).Order("next_sync_at asc").Limit(limit).Find(&candidates).Error; err != nil {
		return nil, err
	}
	claimed := make([]*AssetReplica, 0, len(candidates))
	for _, candidate := range candidates {
		ok, err := ClaimAssetReplica(candidate, now, leaseUntil, lockedBy)
		if err != nil {
			return nil, err
		}
		if ok {
			claimed = append(claimed, candidate)
		}
	}
	return claimed, nil
}

func FinishAssetReplica(id int64, lockedBy string, status string, progress int, upstreamAssetID string, nextSyncAt int64, lastError string, operation ...string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		asset, replica, err := lockAssetReplica(tx, id)
		if err != nil {
			return err
		}
		if !ownsAssetReplica(replica, lockedBy) {
			return errors.New("asset replica synchronization lease was lost")
		}
		now := common.GetTimestamp()
		updates := map[string]any{
			"status": status, "progress": progress, "upstream_asset_id": upstreamAssetID,
			"next_sync_at": nextSyncAt, "locked_by": "", "lease_until": 0,
			"last_error": lastError, "last_synced_at": now, "updated_at": now,
		}
		if assetDeletionRequested(asset, replica) {
			updates = assetReplicaDeletionUpdates(now)
			updates["last_synced_at"], updates["last_error"] = now, lastError
			if status == AssetReplicaStatusFailed && replica.Status != AssetReplicaStatusDeleted && replica.UpstreamAssetID != "" && len(operation) > 0 && operation[0] == AssetReplicaOperationDelete {
				updates["next_sync_at"] = nextSyncAt
			}
		} else if status == AssetReplicaStatusDeleted {
			return errors.New("asset deletion was not requested")
		}
		if status == AssetReplicaStatusFailed {
			updates["attempts"] = gorm.Expr("attempts + ?", 1)
		} else if status == AssetReplicaStatusActive && !assetDeletionRequested(asset, replica) {
			updates["attempts"] = 0
		}
		if err := tx.Model(replica).Updates(updates).Error; err != nil {
			return err
		}
		if status == AssetReplicaStatusFailed && nextSyncAt == 0 && asset.Status == AssetStatusProcessing && !assetDeletionRequested(asset, replica) {
			var remaining int64
			if err := tx.Model(&AssetReplica{}).Where("asset_id = ? AND status NOT IN ? AND NOT (status = ? AND next_sync_at = 0)", asset.ID,
				[]string{AssetReplicaStatusDeleted, AssetReplicaStatusRejected}, AssetReplicaStatusFailed).Count(&remaining).Error; err != nil {
				return err
			}
			if remaining == 0 {
				return tx.Model(asset).Updates(map[string]any{"status": AssetStatusSyncFailed, "updated_at": now}).Error
			}
		}
		return nil
	})
}

// ActivateAssetReplicaWithTx marks an upstream replica usable and promotes a
// newly uploaded asset from processing to ready in the same transaction.
func ActivateAssetReplicaWithTx(tx *gorm.DB, id int64, lockedBy string, upstreamAssetID string, nextSyncAt int64) (*MediaAsset, bool, error) {
	now := common.GetTimestamp()
	asset, replica, err := lockAssetReplica(tx, id)
	if err != nil {
		return nil, false, err
	}
	if !ownsAssetReplica(replica, lockedBy) {
		return nil, false, errors.New("asset replica synchronization lease was lost")
	}
	if assetDeletionRequested(asset, replica) {
		return asset, false, tx.Model(replica).Updates(assetReplicaDeletionUpdates(now)).Error
	}
	if asset.Status != AssetStatusProcessing && asset.Status != AssetStatusReady {
		return nil, false, errors.New("asset is no longer available for activation")
	}
	result := tx.Model(&AssetReplica{}).
		Where("id = ? AND status IN ? AND locked_by = ?", id, []string{AssetReplicaStatusSyncing, AssetReplicaStatusActive}, lockedBy).
		Updates(map[string]any{
			"status": AssetReplicaStatusActive, "progress": 100,
			"upstream_asset_id": upstreamAssetID, "next_sync_at": nextSyncAt,
			"locked_by": "", "lease_until": 0, "last_error": "", "attempts": 0,
			"last_synced_at": now, "updated_at": now,
		})
	if result.Error != nil {
		return nil, false, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, false, errors.New("asset replica synchronization lease was lost")
	}
	if asset.Status != AssetStatusProcessing {
		return asset, false, nil
	}
	if err := tx.Model(&MediaAsset{}).Where("id = ? AND status = ?", asset.ID, AssetStatusProcessing).
		Updates(map[string]any{"status": AssetStatusReady, "updated_at": now}).Error; err != nil {
		return nil, false, err
	}
	asset.Status = AssetStatusReady
	asset.UpdatedAt = now
	return asset, true, nil
}

// RejectAssetReplica makes a content-policy rejection terminal for this asset.
// The replica lease and asset state are changed atomically, so mapping cannot
// observe a rejected replica while the asset still appears usable.
func RejectAssetReplica(id int64, lockedBy string, reason string, upstreamAssetID string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		_, _, err := RejectAssetReplicaWithTx(tx, id, lockedBy, reason, upstreamAssetID)
		return err
	})
}

// RejectAssetReplicaWithTx records a terminal content rejection while
// allowing the caller to append durable notifications before commit.
func RejectAssetReplicaWithTx(tx *gorm.DB, id int64, lockedBy string, reason string, upstreamAssetID string) (*MediaAsset, bool, error) {
	if reason != AssetUnavailableRealPerson && reason != AssetUnavailableSensitiveContent && reason != AssetUnavailablePolicyRejected {
		return nil, false, errors.New("invalid asset rejection reason")
	}
	now := common.GetTimestamp()
	asset, replica, err := lockAssetReplica(tx, id)
	if err != nil {
		return nil, false, err
	}
	if !ownsAssetReplica(replica, lockedBy) {
		return nil, false, errors.New("asset replica synchronization lease was lost")
	}
	if assetDeletionRequested(asset, replica) {
		return asset, false, tx.Model(replica).Updates(assetReplicaDeletionUpdates(now)).Error
	}
	if asset.Status != AssetStatusProcessing && asset.Status != AssetStatusReady && asset.Status != AssetStatusUnavailable {
		return nil, false, errors.New("asset is no longer available for synchronization")
	}
	changed := asset.Status != AssetStatusUnavailable
	if changed {
		if err := tx.Model(&MediaAsset{}).Where("id = ?", replica.AssetID).
			Updates(map[string]any{"status": AssetStatusUnavailable, "unavailable_reason": reason, "updated_at": now}).Error; err != nil {
			return nil, false, err
		}
		asset.Status = AssetStatusUnavailable
		asset.UnavailableReason = reason
		asset.UpdatedAt = now
	}
	result := tx.Model(&AssetReplica{}).
		Where("id = ? AND status IN ? AND locked_by = ?", id, []string{AssetReplicaStatusSyncing, AssetReplicaStatusActive}, lockedBy).
		Updates(map[string]any{
			"status": AssetReplicaStatusRejected, "progress": replica.Progress,
			"upstream_asset_id": upstreamAssetID, "next_sync_at": 0,
			"locked_by": "", "lease_until": 0, "last_error": reason,
			"last_synced_at": now, "updated_at": now,
		})
	if result.Error != nil {
		return nil, false, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, false, errors.New("asset replica synchronization lease was lost")
	}
	return asset, changed, nil
}

func AssetReplicaBackoff(attempts int) time.Duration {
	if attempts < 0 {
		attempts = 0
	}
	minutes := 1 << min(attempts, 8)
	return time.Duration(minutes) * time.Minute
}
