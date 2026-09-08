package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

const (
	StorageProviderAliyunOSS  = "aliyun_oss"
	StorageProviderTencentCOS = "tencent_cos"
	StorageProviderS3         = "s3_compatible"

	StorageProfileStatusDisabled = 0
	StorageProfileStatusEnabled  = 1
	StorageProfileStatusArchived = 2

	StorageCredentialStatusInactive = 0
	StorageCredentialStatusActive   = 1

	StorageCredentialAuthStatic      = "static_access_key"
	StorageCredentialAuthEnvironment = "environment"

	StoragePolicyRelayMediaTemp = "relay_media_temp"

	StorageObjectPurposeRelayMediaTemp = "relay_media_temp"
	StorageObjectStatusUploading       = "uploading"
	StorageObjectStatusUploaded        = "uploaded"
	StorageObjectStatusDeletePending   = "delete_pending"
	StorageObjectStatusDeleting        = "deleting"
	StorageObjectStatusDeleteFailed    = "delete_failed"
	StorageObjectStatusDeleted         = "deleted"
)

type StorageProfile struct {
	ID           int    `json:"id" gorm:"primaryKey"`
	Name         string `json:"name" gorm:"type:varchar(128);uniqueIndex"`
	ProviderType string `json:"provider_type" gorm:"type:varchar(32);index"`
	Status       int    `json:"status" gorm:"index"`
	Endpoint     string `json:"endpoint" gorm:"type:varchar(512)"`
	Region       string `json:"region" gorm:"type:varchar(64)"`
	Bucket       string `json:"bucket" gorm:"type:varchar(255)"`
	Config       string `json:"-" gorm:"type:text"`
	CreatedAt    int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt    int64  `json:"updated_at" gorm:"bigint"`
}

type StorageCredential struct {
	ID               int64  `json:"id" gorm:"primaryKey"`
	StorageProfileID int    `json:"storage_profile_id" gorm:"uniqueIndex:idx_storage_credentials_profile_version,priority:1"`
	Version          int    `json:"version" gorm:"uniqueIndex:idx_storage_credentials_profile_version,priority:2"`
	AuthType         string `json:"auth_type" gorm:"type:varchar(32)"`
	EncryptedPayload string `json:"-" gorm:"type:text"`
	AccessKeyHint    string `json:"access_key_hint" gorm:"type:varchar(16)"`
	KeyVersion       string `json:"-" gorm:"type:varchar(32)"`
	Status           int    `json:"status" gorm:"index"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt        int64  `json:"updated_at" gorm:"bigint"`
}

type StoragePolicy struct {
	ID                  int64  `json:"id" gorm:"primaryKey"`
	Key                 string `json:"key" gorm:"type:varchar(64);uniqueIndex"`
	Name                string `json:"name" gorm:"type:varchar(128)"`
	Purpose             string `json:"purpose" gorm:"type:varchar(64);index"`
	StorageProfileID    int    `json:"storage_profile_id" gorm:"index"`
	ObjectPrefix        string `json:"object_prefix" gorm:"type:varchar(255)"`
	SignedURLTTLSeconds int64  `json:"signed_url_ttl_seconds" gorm:"bigint"`
	RetentionSeconds    int64  `json:"retention_seconds" gorm:"bigint"`
	MaxFileBytes        int64  `json:"max_file_bytes" gorm:"bigint"`
	MaxTotalBytes       int64  `json:"max_total_bytes" gorm:"bigint"`
	MaxFiles            int    `json:"max_files"`
	AllowedMIMETypes    string `json:"allowed_mime_types" gorm:"type:text"`
	Enabled             bool   `json:"enabled"`
	CreatedAt           int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt           int64  `json:"updated_at" gorm:"bigint"`
}

type StorageObject struct {
	ID               int64  `json:"id" gorm:"primaryKey"`
	ObjectID         string `json:"object_id" gorm:"type:varchar(64);uniqueIndex"`
	OwnerUserID      int    `json:"owner_user_id" gorm:"index"`
	RequestID        string `json:"request_id" gorm:"type:varchar(128);index"`
	TaskID           string `json:"task_id" gorm:"type:varchar(64);index"`
	StorageProfileID int    `json:"storage_profile_id" gorm:"index"`
	StoragePolicyID  int64  `json:"storage_policy_id" gorm:"index"`
	Purpose          string `json:"purpose" gorm:"type:varchar(64);index"`
	ObjectKey        string `json:"object_key" gorm:"type:varchar(1024)"`
	ContentType      string `json:"content_type" gorm:"type:varchar(128)"`
	Size             int64  `json:"size" gorm:"bigint"`
	SHA256           string `json:"sha256" gorm:"type:char(64);index"`
	ETag             string `json:"etag" gorm:"type:varchar(255)"`
	Status           string `json:"status" gorm:"type:varchar(32);index:idx_storage_objects_cleanup,priority:1"`
	DeleteAfter      int64  `json:"delete_after" gorm:"bigint;index:idx_storage_objects_cleanup,priority:2"`
	DeleteAttempts   int    `json:"delete_attempts"`
	DeleteLockedBy   string `json:"-" gorm:"type:varchar(128);index"`
	DeleteLeaseUntil int64  `json:"-" gorm:"bigint;index"`
	LastError        string `json:"-" gorm:"type:text"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt        int64  `json:"updated_at" gorm:"bigint"`
	DeletedAt        int64  `json:"deleted_at" gorm:"bigint;index"`
}

func (profile *StorageProfile) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if profile.CreatedAt == 0 {
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now
	return nil
}

func (profile *StorageProfile) BeforeUpdate(_ *gorm.DB) error {
	profile.UpdatedAt = common.GetTimestamp()
	return nil
}

func (credential *StorageCredential) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if credential.CreatedAt == 0 {
		credential.CreatedAt = now
	}
	credential.UpdatedAt = now
	return nil
}

func (credential *StorageCredential) BeforeUpdate(_ *gorm.DB) error {
	credential.UpdatedAt = common.GetTimestamp()
	return nil
}

func (policy *StoragePolicy) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if policy.CreatedAt == 0 {
		policy.CreatedAt = now
	}
	policy.UpdatedAt = now
	return nil
}

func (policy *StoragePolicy) BeforeUpdate(_ *gorm.DB) error {
	policy.UpdatedAt = common.GetTimestamp()
	return nil
}

func (object *StorageObject) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if object.CreatedAt == 0 {
		object.CreatedAt = now
	}
	object.UpdatedAt = now
	return nil
}

func DefaultRelayMediaStoragePolicy() *StoragePolicy {
	return &StoragePolicy{
		Key:                 StoragePolicyRelayMediaTemp,
		Name:                "Relay media temporary storage",
		Purpose:             StorageObjectPurposeRelayMediaTemp,
		ObjectPrefix:        "temporary/relay-media",
		SignedURLTTLSeconds: 72 * 60 * 60,
		RetentionSeconds:    72 * 60 * 60,
		MaxFileBytes:        10 * 1024 * 1024,
		MaxTotalBytes:       20 * 1024 * 1024,
		MaxFiles:            10,
		AllowedMIMETypes:    "image/jpeg,image/png,image/webp",
		Enabled:             false,
	}
}

func GetStorageProfileByID(id int) (*StorageProfile, error) {
	var profile StorageProfile
	if err := DB.First(&profile, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func ListStorageProfiles() ([]*StorageProfile, error) {
	var profiles []*StorageProfile
	err := DB.Where("status <> ?", StorageProfileStatusArchived).Order("id asc").Find(&profiles).Error
	return profiles, err
}

func GetActiveStorageCredential(profileID int) (*StorageCredential, error) {
	var credential StorageCredential
	err := DB.Where("storage_profile_id = ? AND status = ?", profileID, StorageCredentialStatusActive).
		Order("version desc").First(&credential).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &credential, nil
}

func GetStoragePolicyByKey(key string) (*StoragePolicy, error) {
	var policy StoragePolicy
	err := DB.Where(map[string]any{"key": key}).First(&policy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &policy, nil
}

func SaveStoragePolicy(policy *StoragePolicy) error {
	if policy == nil {
		return errors.New("storage policy is required")
	}
	existing, err := GetStoragePolicyByKey(policy.Key)
	if err != nil {
		return err
	}
	if existing == nil {
		return DB.Create(policy).Error
	}
	policy.ID = existing.ID
	policy.CreatedAt = existing.CreatedAt
	return DB.Save(policy).Error
}

func CreateStorageObject(object *StorageObject) error {
	return DB.Create(object).Error
}

func MarkStorageObjectUploaded(id int64, etag string) error {
	return DB.Model(&StorageObject{}).
		Where("id = ? AND status = ?", id, StorageObjectStatusUploading).
		Select("Status", "ETag", "UpdatedAt").
		Updates(&StorageObject{
			Status:    StorageObjectStatusUploaded,
			ETag:      etag,
			UpdatedAt: common.GetTimestamp(),
		}).Error
}

func MarkStorageObjectForDeletion(id int64, deleteAfter int64, message string) error {
	if len(message) > 2000 {
		message = message[:2000]
	}
	return DB.Model(&StorageObject{}).
		Where("id = ? AND status <> ?", id, StorageObjectStatusDeleted).
		Updates(map[string]any{
			"status":       StorageObjectStatusDeletePending,
			"delete_after": deleteAfter,
			"last_error":   message,
			"updated_at":   common.GetTimestamp(),
		}).Error
}

func MarkStorageObjectsForTaskDeletion(taskID string, deleteAfter int64) error {
	if taskID == "" {
		return nil
	}
	return DB.Model(&StorageObject{}).
		Where("task_id = ? AND status IN ?", taskID, []string{StorageObjectStatusUploaded, StorageObjectStatusDeleteFailed}).
		Updates(map[string]any{
			"status":       StorageObjectStatusDeletePending,
			"delete_after": deleteAfter,
			"updated_at":   common.GetTimestamp(),
		}).Error
}

func ClaimStorageObjectsForDeletion(now int64, leaseUntil int64, lockedBy string, limit int) ([]*StorageObject, error) {
	if limit <= 0 {
		limit = 50
	}
	var candidates []*StorageObject
	err := DB.Where(
		"((status IN ? AND delete_after <= ?) OR (status = ? AND delete_lease_until <= ?))",
		[]string{StorageObjectStatusUploading, StorageObjectStatusUploaded, StorageObjectStatusDeletePending, StorageObjectStatusDeleteFailed}, now,
		StorageObjectStatusDeleting, now,
	).Order("delete_after asc").Limit(limit).Find(&candidates).Error
	if err != nil {
		return nil, err
	}

	claimed := make([]*StorageObject, 0, len(candidates))
	for _, candidate := range candidates {
		result := DB.Model(&StorageObject{}).
			Where("id = ? AND ((status IN ? AND delete_after <= ?) OR (status = ? AND delete_lease_until <= ?))",
				candidate.ID,
				[]string{StorageObjectStatusUploading, StorageObjectStatusUploaded, StorageObjectStatusDeletePending, StorageObjectStatusDeleteFailed}, now,
				StorageObjectStatusDeleting, now,
			).
			Updates(map[string]any{
				"status":             StorageObjectStatusDeleting,
				"delete_locked_by":   lockedBy,
				"delete_lease_until": leaseUntil,
				"updated_at":         now,
			})
		if result.Error != nil {
			return nil, result.Error
		}
		if result.RowsAffected == 1 {
			candidate.Status = StorageObjectStatusDeleting
			candidate.DeleteLockedBy = lockedBy
			candidate.DeleteLeaseUntil = leaseUntil
			claimed = append(claimed, candidate)
		}
	}
	return claimed, nil
}

func FinishStorageObjectDeletion(id int64, lockedBy string) error {
	now := common.GetTimestamp()
	result := DB.Model(&StorageObject{}).
		Where("id = ? AND status = ? AND delete_locked_by = ?", id, StorageObjectStatusDeleting, lockedBy).
		Updates(map[string]any{
			"status":             StorageObjectStatusDeleted,
			"deleted_at":         now,
			"updated_at":         now,
			"delete_locked_by":   "",
			"delete_lease_until": 0,
			"last_error":         "",
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("storage object deletion lease lost: %d", id)
	}
	return nil
}

func FailStorageObjectDeletion(id int64, lockedBy string, deleteAfter int64, message string) error {
	if len(message) > 2000 {
		message = message[:2000]
	}
	now := common.GetTimestamp()
	result := DB.Model(&StorageObject{}).
		Where("id = ? AND status = ? AND delete_locked_by = ?", id, StorageObjectStatusDeleting, lockedBy).
		Updates(map[string]any{
			"status":             StorageObjectStatusDeleteFailed,
			"delete_attempts":    gorm.Expr("delete_attempts + ?", 1),
			"delete_after":       deleteAfter,
			"delete_locked_by":   "",
			"delete_lease_until": 0,
			"last_error":         message,
			"updated_at":         now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("storage object deletion lease lost: %d", id)
	}
	return nil
}

func StorageProfileHasObjects(profileID int) (bool, error) {
	var count int64
	err := DB.Model(&StorageObject{}).
		Where("storage_profile_id = ? AND status <> ?", profileID, StorageObjectStatusDeleted).
		Count(&count).Error
	return count > 0, err
}

func StorageProfileIsReferenced(profileID int) (bool, error) {
	var count int64
	err := DB.Model(&StoragePolicy{}).Where("storage_profile_id = ?", profileID).Count(&count).Error
	return count > 0, err
}

func StorageObjectDeletionBackoff(attempts int) time.Duration {
	if attempts < 0 {
		attempts = 0
	}
	minutes := 1 << min(attempts, 8)
	return time.Duration(minutes) * time.Minute
}

func PurgeDeletedStorageObjects(deletedBefore int64, limit int) (int64, error) {
	if limit <= 0 {
		limit = 500
	}
	var ids []int64
	if err := DB.Model(&StorageObject{}).
		Where("status = ? AND deleted_at > 0 AND deleted_at <= ?", StorageObjectStatusDeleted, deletedBefore).
		Order("deleted_at asc").Limit(limit).Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	result := DB.Where("id IN ?", ids).Delete(&StorageObject{})
	return result.RowsAffected, result.Error
}
