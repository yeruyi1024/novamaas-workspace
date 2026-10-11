package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var ErrImageStudioResultBusy = errors.New("image is being generated or saved")

const (
	ImageStudioRunStatusRunning   = "running"
	ImageStudioRunStatusCompleted = "completed"
	ImageStudioRunStatusPartial   = "partial"
	ImageStudioRunStatusFailed    = "failed"
	ImageStudioRunStatusCanceled  = "canceled"
)

type ImageStudioRun struct {
	ID             int64  `json:"-" gorm:"primaryKey"`
	PublicID       string `json:"id" gorm:"type:varchar(64);uniqueIndex"`
	UserID         int    `json:"owner_user_id" gorm:"uniqueIndex:idx_image_studio_runs_user_idempotency,priority:1;index:idx_image_studio_runs_user_created,priority:1"`
	IdempotencyKey string `json:"-" gorm:"type:varchar(128);uniqueIndex:idx_image_studio_runs_user_idempotency,priority:2"`
	Model          string `json:"model" gorm:"type:varchar(255)"`
	TokenID        int    `json:"-"`
	RequestID      string `json:"request_id" gorm:"type:varchar(64)"`
	Quota          int    `json:"quota"`
	DurationMS     int64  `json:"duration_ms"`
	Status         string `json:"status" gorm:"type:varchar(16);index"`
	Prompt         string `json:"prompt,omitempty" gorm:"type:text"`
	LeaseUntil     int64  `json:"-" gorm:"bigint;index"`
	ErrorCode      string `json:"error_code,omitempty" gorm:"type:varchar(64)"`
	ErrorMessage   string `json:"error_message,omitempty" gorm:"type:text"`
	StartedAt      int64  `json:"started_at" gorm:"bigint"`
	CompletedAt    int64  `json:"completed_at" gorm:"bigint"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint;index:idx_image_studio_runs_user_created,priority:2"`
	UpdatedAt      int64  `json:"updated_at" gorm:"bigint"`
}

type ImageStudioResult struct {
	SaveStatus      string             `json:"save_status" gorm:"type:varchar(16)"`
	SaveError       string             `json:"save_error,omitempty" gorm:"type:text"`
	SaveGroupID     string             `json:"save_group_id,omitempty" gorm:"type:varchar(64)"`
	SaveAttemptID   string             `json:"-" gorm:"type:varchar(64)"`
	SaveStartedAt   int64              `json:"-" gorm:"bigint"`
	PreviewURL      ImageStudioPreview `json:"-"`
	ThumbnailURL    ImageStudioPreview `json:"-"`
	ID              int64              `json:"-" gorm:"primaryKey"`
	PublicID        string             `json:"id" gorm:"type:varchar(64);uniqueIndex"`
	RunID           int64              `json:"-" gorm:"index"`
	StorageObjectID int64              `json:"-" gorm:"index"`
	MediaAssetID    int64              `json:"-" gorm:"index"`
	ResultIndex     int                `json:"index"`
	RevisedPrompt   string             `json:"revised_prompt,omitempty" gorm:"type:text"`
	CreatedAt       int64              `json:"created_at" gorm:"bigint;index"`
}

type ImageStudioResultWithAsset struct {
	ImageStudioResult
	AssetPublicID string
}

// History reads small previews for a whole page; originals are only loaded for
// a selected task. Deleted or foreign library assets are never exposed.
func ListImageStudioResults(runIDs []int64, originals bool) ([]ImageStudioResultWithAsset, error) {
	rows := make([]ImageStudioResultWithAsset, 0)
	if len(runIDs) == 0 {
		return rows, nil
	}
	columns := "image_studio_results.id, image_studio_results.public_id, image_studio_results.run_id, image_studio_results.result_index, image_studio_results.revised_prompt, image_studio_results.thumbnail_url, image_studio_results.save_status, image_studio_results.save_error, image_studio_results.media_asset_id, image_studio_results.storage_object_id"
	if originals {
		columns = "image_studio_results.*"
	}
	err := DB.Table("image_studio_results").Select(columns+", media_assets.public_id AS asset_public_id").
		Joins("JOIN image_studio_runs ON image_studio_runs.id = image_studio_results.run_id").
		Joins("LEFT JOIN media_assets ON media_assets.id = image_studio_results.media_asset_id AND media_assets.owner_user_id = image_studio_runs.user_id AND media_assets.status <> ?", AssetStatusDeleted).
		Where("image_studio_results.run_id IN ?", runIDs).Order("image_studio_results.result_index asc").Scan(&rows).Error
	return rows, err
}

func (value *ImageStudioRun) BeforeCreate(_ *gorm.DB) error {
	if value.CreatedAt == 0 {
		value.CreatedAt = common.GetTimestamp()
	}
	value.UpdatedAt = value.CreatedAt
	return nil
}

func (value *ImageStudioResult) BeforeCreate(_ *gorm.DB) error {
	if value.PublicID == "" {
		value.PublicID = "result_" + common.GetUUID()
	}
	if value.CreatedAt == 0 {
		value.CreatedAt = common.GetTimestamp()
	}
	return nil
}

// Interrupted calls are failed, never resubmitted or billed again.
func FailExpiredImageStudioRun(run *ImageStudioRun, now int64) error {
	if run.Status != ImageStudioRunStatusRunning || run.LeaseUntil == 0 || run.LeaseUntil > now {
		return nil
	}
	update := DB.Model(&ImageStudioRun{}).Where("id = ? AND status = ? AND lease_until = ?", run.ID, run.Status, run.LeaseUntil).Updates(map[string]any{
		"status": ImageStudioRunStatusFailed, "error_code": "generation_interrupted",
		"error_message": "generation was interrupted; it will not be resubmitted automatically",
		"completed_at":  now, "lease_until": 0, "updated_at": now,
	})
	if update.Error != nil {
		return update.Error
	}
	return DB.First(run, run.ID).Error
}

// ClaimImageStudioResultSave prevents duplicate concurrent imports. A timed-out
// attempt can be retried; the attempt ID prevents its late completion taking over.
func ClaimImageStudioResultSave(id int64, groupID, attemptID string, now int64) (bool, error) {
	result := DB.Model(&ImageStudioResult{}).Where("id = ? AND media_asset_id = 0", id).
		Where("save_status IS NULL OR save_status <> ? OR save_started_at < ?", "saving", now-360).
		Updates(map[string]any{"save_status": "saving", "save_error": "", "save_group_id": groupID, "save_attempt_id": attemptID, "save_started_at": now})
	return result.RowsAffected == 1, result.Error
}

// DiscardImageStudioRun removes failed outputs, retaining only retry identity
// after an accepted request. Its objects use the existing storage cleanup job.
func DiscardImageStudioRun(run *ImageStudioRun) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		objects := tx.Model(&ImageStudioResult{}).Select("storage_object_id").Where("run_id = ? AND storage_object_id > 0", run.ID)
		now := common.GetTimestamp()
		if err := tx.Model(&StorageObject{}).Where("id IN (?) AND owner_user_id = ? AND purpose = ?", objects, run.UserID, StorageObjectPurposeImageStudio).
			Updates(map[string]any{"status": StorageObjectStatusDeletePending, "delete_after": now, "updated_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Where("run_id = ?", run.ID).Delete(&ImageStudioResult{}).Error; err != nil {
			return err
		}
		if run.Status != ImageStudioRunStatusCanceled {
			return tx.Delete(run).Error
		}
		run.Prompt, run.LeaseUntil = "", 0
		run.CompletedAt, run.UpdatedAt = now, now
		return tx.Save(run).Error
	})
}

// DeleteImageStudioResult removes one owned image. The empty run keeps only its
// billing and idempotency metadata so a delayed retry cannot generate again.
func DeleteImageStudioResult(userID int, publicID string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var result ImageStudioResult
		if err := tx.Table("image_studio_results").Select("image_studio_results.*").
			Joins("JOIN image_studio_runs ON image_studio_runs.id = image_studio_results.run_id").
			Where("image_studio_results.public_id = ? AND image_studio_runs.user_id = ?", publicID, userID).
			First(&result).Error; err != nil {
			return err
		}
		var run ImageStudioRun
		if err := lockForUpdate(tx).Where("id = ? AND user_id = ?", result.RunID, userID).First(&run).Error; err != nil {
			return err
		}
		if run.Status != ImageStudioRunStatusCompleted && run.Status != ImageStudioRunStatusPartial {
			return ErrImageStudioResultBusy
		}
		deleted := tx.Where("id = ?", result.ID).
			Where("save_status IS NULL OR save_status <> ? OR save_started_at < ?", "saving", common.GetTimestamp()-360).
			Delete(&ImageStudioResult{})
		if deleted.Error != nil {
			return deleted.Error
		}
		if deleted.RowsAffected == 0 {
			return ErrImageStudioResultBusy
		}
		if result.StorageObjectID > 0 {
			if err := tx.Model(&StorageObject{}).Where("id = ? AND owner_user_id = ? AND purpose = ?", result.StorageObjectID, userID, StorageObjectPurposeImageStudio).
				Updates(map[string]any{"status": StorageObjectStatusDeletePending, "delete_after": common.GetTimestamp(), "updated_at": common.GetTimestamp()}).Error; err != nil {
				return err
			}
		}
		var remaining int64
		if err := tx.Model(&ImageStudioResult{}).Where("run_id = ?", run.ID).Count(&remaining).Error; err != nil {
			return err
		}
		if remaining > 0 {
			return nil
		}
		return tx.Model(&run).Updates(map[string]any{
			"status": ImageStudioRunStatusCanceled, "prompt": "",
			"error_code": "", "error_message": "", "updated_at": common.GetTimestamp(),
		}).Error
	})
}

// Unsaved Base64 images require more than MySQL TEXT's 64 KiB limit.
type ImageStudioPreview string

func (ImageStudioPreview) GormDataType() string { return "text" }
func (ImageStudioPreview) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	if db.Dialector.Name() == "mysql" {
		return "longtext"
	}
	return "text"
}
