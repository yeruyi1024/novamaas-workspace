package model

import (
	"errors"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ImageStudioRetentionSeconds int64 = 30 * 24 * 60 * 60

	ImageStudioRunStatusRunning   = "running"
	ImageStudioRunStatusCompleted = "completed"
	ImageStudioRunStatusFailed    = "failed"
	// Deleted runs are hidden from owners and reviewers but keep the request
	// ID and idempotency key so a delayed retry can never generate again.
	ImageStudioRunStatusDeleted = "deleted"

	// A run still marked running after this many seconds was interrupted
	// (for example by a restart) and is failed instead of being resubmitted.
	ImageStudioRunStaleSeconds = 600
	// A save claim older than this many seconds may be taken over by a retry.
	ImageStudioSaveStaleSeconds = 300
)

// InitializeImageStudioRetention schedules existing generated images without
// changing explicit deletion requests or permanent asset library copies.
func InitializeImageStudioRetention(db *gorm.DB) error {
	var legacy StoragePolicy
	err := db.Where(map[string]any{"key": StoragePolicyAssetLibrary}).First(&legacy).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	for _, purpose := range []string{StorageObjectPurposeImageStudio, StorageObjectPurposeImageStudioReference} {
		policy := DefaultImageStudioStoragePolicy(purpose)
		if legacy.ID > 0 {
			policy.StorageProfileID, policy.Enabled = legacy.StorageProfileID, legacy.Enabled
			policy.SignedURLTTLSeconds = min(legacy.SignedURLTTLSeconds, policy.RetentionSeconds)
		}
		// Seed once; never follow later asset-library policy edits.
		if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "key"}}, DoNothing: true}).Create(policy).Error; err != nil {
			return err
		}
	}
	return db.Model(&StorageObject{}).
		Where("purpose = ? AND delete_after = 0 AND status IN ?", StorageObjectPurposeImageStudio,
			[]string{StorageObjectStatusUploading, StorageObjectStatusUploaded}).
		Update("delete_after", gorm.Expr("created_at + ?", ImageStudioRetentionSeconds)).Error
}

func DefaultImageStudioStoragePolicy(purpose string) *StoragePolicy {
	retention, maxBytes := ImageStudioRetentionSeconds, int64(64<<20)
	if purpose == StorageObjectPurposeImageStudioReference {
		retention, maxBytes = 7*24*3600, 20<<20
	}
	return &StoragePolicy{Key: purpose, Name: purpose, Purpose: purpose,
		ObjectPrefix: "images/" + purpose, SignedURLTTLSeconds: 24 * 3600,
		RetentionSeconds: retention, MaxFileBytes: maxBytes, MaxTotalBytes: maxBytes, MaxFiles: 1,
		AllowedMIMETypes: "image/jpeg,image/png,image/webp"}
}

// CountRunningImageStudioRuns counts a user's generations that are still within
// the stale window; interrupted runs no longer occupy a slot.
func CountRunningImageStudioRuns(userID int) (int64, error) {
	var count int64
	err := DB.Model(&ImageStudioRun{}).
		Where("user_id = ? AND status = ? AND created_at > ?", userID, ImageStudioRunStatusRunning, common.GetTimestamp()-ImageStudioRunStaleSeconds).
		Count(&count).Error
	return count, err
}

var (
	ErrImageStudioSaveInProgress = errors.New("image is already being saved")
	ErrImageStudioRunFinished    = errors.New("image generation already finished")
)

type ImageStudioRun struct {
	ID             int64  `json:"-" gorm:"primaryKey"`
	PublicID       string `json:"id" gorm:"type:varchar(64);uniqueIndex"`
	UserID         int    `json:"owner_user_id" gorm:"uniqueIndex:idx_image_studio_runs_user_key,priority:1;index:idx_image_studio_runs_user_created,priority:1"`
	IdempotencyKey string `json:"-" gorm:"type:varchar(128);uniqueIndex:idx_image_studio_runs_user_key,priority:2"`
	TokenID        int    `json:"-"`
	ChannelID      int    `json:"-" gorm:"index;default:0"`
	Model          string `json:"model" gorm:"type:varchar(255)"`
	Prompt         string `json:"prompt" gorm:"type:text"`
	Status         string `json:"status" gorm:"type:varchar(16);index"`
	RequestID      string `json:"-" gorm:"type:varchar(64);index"`
	Quota          int    `json:"quota"`
	DurationMS     int64  `json:"duration_ms"`
	ErrorMessage   string `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint;index:idx_image_studio_runs_user_created,priority:2"`
	UpdatedAt      int64  `json:"updated_at" gorm:"bigint"`
}

type ImageStudioResult struct {
	ID              int64  `json:"-" gorm:"primaryKey"`
	PublicID        string `json:"id" gorm:"type:varchar(64);uniqueIndex"`
	RunID           int64  `json:"-" gorm:"index"`
	ResultIndex     int    `json:"index"`
	RevisedPrompt   string `json:"description,omitempty" gorm:"type:text"`
	StorageObjectID int64  `json:"-"`
	MediaAssetID    int64  `json:"-" gorm:"index"`
	SavingAt        int64  `json:"-" gorm:"bigint"`
	CreatedAt       int64  `json:"created_at" gorm:"bigint"`
}

func (run *ImageStudioRun) BeforeCreate(_ *gorm.DB) error {
	if run.PublicID == "" {
		run.PublicID = "run_" + common.GetUUID()
	}
	if run.CreatedAt == 0 {
		run.CreatedAt = common.GetTimestamp()
	}
	run.UpdatedAt = run.CreatedAt
	return nil
}

func (result *ImageStudioResult) BeforeCreate(_ *gorm.DB) error {
	if result.PublicID == "" {
		result.PublicID = "result_" + common.GetUUID()
	}
	if result.CreatedAt == 0 {
		result.CreatedAt = common.GetTimestamp()
	}
	return nil
}

func FindImageStudioRunByIdempotencyKey(userID int, key string) (*ImageStudioRun, error) {
	var run ImageStudioRun
	if err := DB.Where("user_id = ? AND idempotency_key = ?", userID, key).First(&run).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

// FailStaleImageStudioRun fails an interrupted run once. The status guard makes
// the update a no-op when the generating request finished in the meantime.
func FailStaleImageStudioRun(run *ImageStudioRun) error {
	now := common.GetTimestamp()
	if run.Status != ImageStudioRunStatusRunning || run.CreatedAt > now-ImageStudioRunStaleSeconds {
		return nil
	}
	if err := DB.Model(&ImageStudioRun{}).Where("id = ? AND status = ?", run.ID, ImageStudioRunStatusRunning).
		Updates(map[string]any{"status": ImageStudioRunStatusFailed, "error_message": "generation was interrupted", "updated_at": now}).Error; err != nil {
		return err
	}
	return DB.First(run, run.ID).Error
}

// LinkImageStudioRunRequest binds the relay request ID to a running run before
// its snapshot is archived, so every archived studio snapshot already has an
// owner run when it becomes readable. Channel retries reuse the same ID.
func LinkImageStudioRunRequest(runID int64, requestID string, channelID int) error {
	if requestID == "" {
		return errors.New("relay request ID is missing")
	}
	update := DB.Model(&ImageStudioRun{}).
		Where("id = ? AND status = ? AND request_id IN ?", runID, ImageStudioRunStatusRunning, []string{"", requestID}).
		Updates(map[string]any{"request_id": requestID, "channel_id": channelID})
	if update.Error != nil {
		return update.Error
	}
	if update.RowsAffected == 1 {
		return nil
	}
	// MySQL reports zero affected rows when the value is unchanged.
	var count int64
	if err := DB.Model(&ImageStudioRun{}).Where("id = ? AND status = ? AND request_id = ?", runID, ImageStudioRunStatusRunning, requestID).Count(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return errors.New("image generation is no longer running")
	}
	return nil
}

// FinishImageStudioRun records the results and the final status of a running
// run in one transaction. Nothing is written when the run already finished
// elsewhere (for example failed as interrupted).
func FinishImageStudioRun(run *ImageStudioRun, results []*ImageStudioResult) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		run.UpdatedAt = common.GetTimestamp()
		update := tx.Model(&ImageStudioRun{}).Where("id = ? AND status = ?", run.ID, ImageStudioRunStatusRunning).Updates(map[string]any{
			"status": run.Status, "request_id": run.RequestID, "quota": run.Quota, "duration_ms": run.DurationMS, "channel_id": run.ChannelID,
			"error_message": run.ErrorMessage, "updated_at": run.UpdatedAt,
		})
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected != 1 {
			return ErrImageStudioRunFinished
		}
		if len(results) == 0 {
			return nil
		}
		return tx.Create(results).Error
	})
}

func ListImageStudioResults(runIDs []int64) ([]ImageStudioResult, error) {
	results := make([]ImageStudioResult, 0)
	if len(runIDs) == 0 {
		return results, nil
	}
	err := DB.Where("run_id IN ?", runIDs).Order("run_id desc, result_index asc").Find(&results).Error
	return results, err
}

// FindOwnedImageStudioResult returns a result of a visible run owned by userID.
func FindOwnedImageStudioResult(userID int, publicID string) (*ImageStudioResult, *ImageStudioRun, error) {
	var result ImageStudioResult
	if err := DB.Where("public_id = ?", publicID).First(&result).Error; err != nil {
		return nil, nil, err
	}
	var run ImageStudioRun
	if err := DB.Where("id = ? AND user_id = ? AND status = ?", result.RunID, userID, ImageStudioRunStatusCompleted).First(&run).Error; err != nil {
		return nil, nil, err
	}
	return &result, &run, nil
}

// ClaimImageStudioResultSave atomically marks a result as being saved so two
// concurrent saves cannot both create library assets. It returns the claim
// timestamp that must match when the save is completed or released.
func ClaimImageStudioResultSave(resultID int64) (int64, error) {
	now := common.GetTimestamp()
	update := DB.Model(&ImageStudioResult{}).
		Where("id = ? AND media_asset_id = 0 AND saving_at < ?", resultID, now-ImageStudioSaveStaleSeconds).
		Update("saving_at", now)
	if update.Error != nil {
		return 0, update.Error
	}
	if update.RowsAffected != 1 {
		return 0, ErrImageStudioSaveInProgress
	}
	return now, nil
}

// CompleteImageStudioResultSave links the created asset only if the claim is
// still held; false means the result was deleted or the claim was taken over.
func CompleteImageStudioResultSave(resultID int64, claimedAt int64, mediaAssetID int64) (bool, error) {
	update := DB.Model(&ImageStudioResult{}).
		Where("id = ? AND saving_at = ? AND media_asset_id = 0", resultID, claimedAt).
		Updates(map[string]any{"media_asset_id": mediaAssetID, "saving_at": 0})
	return update.RowsAffected == 1, update.Error
}

func ReleaseImageStudioResultSave(resultID int64, claimedAt int64) error {
	return DB.Model(&ImageStudioResult{}).Where("id = ? AND saving_at = ?", resultID, claimedAt).Update("saving_at", 0).Error
}

// DetachImageStudioResultAsset clears a link to an asset deleted from the
// library so the image can be saved again.
func DetachImageStudioResultAsset(resultID int64, mediaAssetID int64) error {
	return DB.Model(&ImageStudioResult{}).Where("id = ? AND media_asset_id = ?", resultID, mediaAssetID).Update("media_asset_id", 0).Error
}

// DeleteImageStudioResults removes result rows and queues their objects for
// the existing storage cleanup job. Linked library assets are handled by the
// caller, which decides whether they are kept or deleted.
func DeleteImageStudioResults(tx *gorm.DB, userID int, results []ImageStudioResult) error {
	if len(results) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(results))
	objectIDs := make([]int64, 0, len(results))
	for _, result := range results {
		ids = append(ids, result.ID)
		if result.StorageObjectID > 0 {
			objectIDs = append(objectIDs, result.StorageObjectID)
		}
	}
	now := common.GetTimestamp()
	if len(objectIDs) > 0 {
		if err := tx.Model(&StorageObject{}).
			Where("id IN ? AND owner_user_id = ? AND purpose = ? AND status <> ?", objectIDs, userID, StorageObjectPurposeImageStudio, StorageObjectStatusDeleted).
			Updates(map[string]any{"status": StorageObjectStatusDeletePending, "delete_after": now, "updated_at": now}).Error; err != nil {
			return err
		}
	}
	return tx.Where("id IN ?", ids).Delete(&ImageStudioResult{}).Error
}

// FindDeletableImageStudioRun returns a completed or failed run owned by
// userID; userID 0 matches any owner for image review.
func FindDeletableImageStudioRun(userID int, publicID string) (*ImageStudioRun, error) {
	var run ImageStudioRun
	query := DB.Where("public_id = ? AND status IN ?", publicID,
		[]string{ImageStudioRunStatusCompleted, ImageStudioRunStatusFailed})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	err := query.First(&run).Error
	return &run, err
}

// DeleteImageStudioRun hides a run and removes its images. The run keeps its
// request ID and idempotency key so a delayed retry cannot generate again.
func DeleteImageStudioRun(userID int, run *ImageStudioRun, results []ImageStudioResult) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := DeleteImageStudioResults(tx, userID, results); err != nil {
			return err
		}
		return tx.Model(&ImageStudioRun{}).Where("id = ? AND status IN ?", run.ID, []string{ImageStudioRunStatusCompleted, ImageStudioRunStatusFailed}).
			Updates(map[string]any{"status": ImageStudioRunStatusDeleted, "updated_at": common.GetTimestamp()}).Error
	})
}
