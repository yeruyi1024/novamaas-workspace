package imagestudio

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"
	storageService "github.com/QuantumNous/new-api/service/storage"

	"gorm.io/gorm"
)

type ResultView struct {
	ID           string `json:"id"`
	Index        int    `json:"index"`
	Description  string `json:"description,omitempty"`
	PreviewURL   string `json:"preview_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	DownloadURL  string `json:"download_url,omitempty"`
	AssetID      string `json:"asset_id,omitempty"`
	Expired      bool   `json:"expired,omitempty"`
}

// RunView is the owner and reviewer view of a run. RequestID links to the
// archived request snapshots and is only filled for authorized review.
type RunView struct {
	model.ImageStudioRun
	RequestID string       `json:"request_id,omitempty"`
	Username  string       `json:"username,omitempty"`
	Results   []ResultView `json:"results"`
}

type UserOption struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type RunList struct {
	Items []*RunView   `json:"items"`
	Total int64        `json:"total"`
	Users []UserOption `json:"users,omitempty"`
}

// ListFilter selects runs. OwnerUserID 0 with Review lists every user.
type ListFilter struct {
	PublicIDs   []string
	Review      bool
	OwnerUserID int
	Page        *common.PageInfo
}

// ListRuns lists visible runs. Callers must authorize Review before calling.
func ListRuns(ctx context.Context, filter ListFilter) (*RunList, error) {
	statuses := []string{model.ImageStudioRunStatusRunning, model.ImageStudioRunStatusCompleted, model.ImageStudioRunStatusFailed}
	query := model.DB.Model(&model.ImageStudioRun{}).Where("status IN ?", statuses)
	if len(filter.PublicIDs) > 0 {
		query = query.Where("public_id IN ?", filter.PublicIDs)
	}
	if filter.OwnerUserID > 0 {
		query = query.Where("user_id = ?", filter.OwnerUserID)
	}
	list := &RunList{Items: make([]*RunView, 0)}
	if err := query.Count(&list.Total).Error; err != nil {
		return nil, err
	}
	var runs []*model.ImageStudioRun
	if err := query.Order("id desc").Limit(filter.Page.GetPageSize()).Offset(filter.Page.GetStartIdx()).Find(&runs).Error; err != nil {
		return nil, err
	}
	runIDs := make([]int64, 0, len(runs))
	views := make(map[int64]*RunView, len(runs))
	userIDs := make([]int, 0, len(runs))
	for _, run := range runs {
		if err := model.FailStaleImageStudioRun(run); err != nil {
			return nil, err
		}
		view := &RunView{ImageStudioRun: *run, Results: make([]ResultView, 0)}
		runIDs = append(runIDs, run.ID)
		views[run.ID] = view
		userIDs = append(userIDs, run.UserID)
		list.Items = append(list.Items, view)
	}
	results, err := model.ListImageStudioResults(runIDs)
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		view := views[result.RunID]
		resultView, err := buildResultView(ctx, result, view.UserID)
		if err != nil {
			return nil, err
		}
		view.Results = append(view.Results, resultView)
	}
	if !filter.Review {
		return list, nil
	}
	if err = model.DB.Model(&model.ImageStudioRun{}).Select("image_studio_runs.user_id AS id, users.username AS name").
		Joins("LEFT JOIN users ON users.id = image_studio_runs.user_id").
		Where("image_studio_runs.status IN ?", statuses).
		Group("image_studio_runs.user_id, users.username").Order("image_studio_runs.user_id asc").
		Scan(&list.Users).Error; err != nil {
		return nil, err
	}
	usernames, err := model.GetUsernamesByIDs(userIDs)
	if err != nil {
		return nil, err
	}
	for _, view := range list.Items {
		view.Username = usernames[view.UserID]
		view.RequestID = view.ImageStudioRun.RequestID
	}
	return list, nil
}

// TaskLog contains tracking metadata only, never prompts, signed URLs or snapshots.
type TaskLog struct {
	ChannelID    int    `json:"channel_id,omitempty"`
	ID           string `json:"id"`
	OwnerUserID  int    `json:"owner_user_id"`
	Model        string `json:"model"`
	Status       string `json:"status"`
	RequestID    string `json:"request_id,omitempty"`
	DurationMS   int64  `json:"duration_ms"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type TaskLogList struct {
	Items []TaskLog `json:"items"`
	Total int64     `json:"total"`
}

type TaskLogFilter struct {
	OwnerUserID int
	ChannelID   int
	ID          string
	Start, End  int64
	Page        *common.PageInfo
}

// ListTaskLogs reuses the run ledger, including deleted runs, without signing images.
// filter.OwnerUserID 0 is reserved for authorized cross-user review.
func ListTaskLogs(ctx context.Context, filter TaskLogFilter) (*TaskLogList, error) {
	query := model.DB.WithContext(ctx).Model(&model.ImageStudioRun{})
	if filter.ChannelID > 0 {
		query = query.Where("channel_id = ?", filter.ChannelID)
	}
	if filter.OwnerUserID > 0 {
		query = query.Where("user_id = ?", filter.OwnerUserID)
	}
	if filter.ID != "" {
		query = query.Where("(public_id = ? OR request_id = ?)", filter.ID, filter.ID)
	}
	if filter.Start > 0 {
		query = query.Where("created_at >= ?", filter.Start)
	}
	if filter.End > 0 {
		query = query.Where("created_at <= ?", filter.End)
	}
	list := &TaskLogList{Items: make([]TaskLog, 0)}
	if err := query.Count(&list.Total).Error; err != nil {
		return nil, err
	}
	var runs []model.ImageStudioRun
	if err := query.Select("id", "public_id", "user_id", "model", "status", "request_id", "channel_id", "duration_ms", "created_at", "updated_at", "error_message").
		Order("id DESC").Limit(filter.Page.GetPageSize()).Offset(filter.Page.GetStartIdx()).Find(&runs).Error; err != nil {
		return nil, err
	}
	for i := range runs {
		run := &runs[i]
		if err := model.FailStaleImageStudioRun(run); err != nil {
			return nil, err
		}
		if filter.OwnerUserID > 0 {
			run.ChannelID = 0
		}
		list.Items = append(list.Items, TaskLog{ChannelID: run.ChannelID, ID: run.PublicID, OwnerUserID: run.UserID, Model: run.Model, Status: run.Status,
			RequestID: run.RequestID, DurationMS: run.DurationMS, CreatedAt: run.CreatedAt, UpdatedAt: run.UpdatedAt,
			ErrorMessage: CleanErrorMessage(run.ErrorMessage)})
	}
	return list, nil
}

// PollRuns batches owner-scoped status checks so concurrent generations do not
// multiply the client's API request rate. Missing or foreign IDs are omitted.
func PollRuns(ctx context.Context, userID int, ids []string) (*RunList, error) {
	if userID <= 0 {
		return nil, &RequestError{http.StatusForbidden, "permission_denied", errors.New("image generation owner is required")}
	}
	if len(ids) == 0 || len(ids) > 50 {
		return nil, &RequestError{http.StatusBadRequest, "invalid_run_ids", errors.New("poll between 1 and 50 image generations")}
	}
	for _, id := range ids {
		if strings.TrimSpace(id) == "" || len(id) > 64 {
			return nil, &RequestError{http.StatusBadRequest, "invalid_run_ids", errors.New("invalid image generation ID")}
		}
	}
	return ListRuns(ctx, ListFilter{OwnerUserID: userID, PublicIDs: ids, Page: &common.PageInfo{Page: 1, PageSize: len(ids)}})
}

// GetRun returns a visible run; ownerUserID 0 is only for authorized review.
func GetRun(ctx context.Context, ownerUserID int, publicID string) (*RunView, error) {
	query := model.DB.Where("public_id = ? AND status <> ?", publicID, model.ImageStudioRunStatusDeleted)
	if ownerUserID > 0 {
		query = query.Where("user_id = ?", ownerUserID)
	}
	var run model.ImageStudioRun
	if err := query.First(&run).Error; err != nil {
		return nil, err
	}
	if err := model.FailStaleImageStudioRun(&run); err != nil {
		return nil, err
	}
	view, err := buildRunView(ctx, &run)
	if err == nil && ownerUserID == 0 {
		view.RequestID = run.RequestID
	}
	return view, err
}

func buildRunView(ctx context.Context, run *model.ImageStudioRun) (*RunView, error) {
	view := &RunView{ImageStudioRun: *run, Results: make([]ResultView, 0)}
	results, err := model.ListImageStudioResults([]int64{run.ID})
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		resultView, err := buildResultView(ctx, result, run.UserID)
		if err != nil {
			return nil, err
		}
		view.Results = append(view.Results, resultView)
	}
	return view, nil
}

func buildResultView(ctx context.Context, result model.ImageStudioResult, ownerUserID int) (ResultView, error) {
	view := ResultView{ID: result.PublicID, Index: result.ResultIndex, Description: result.RevisedPrompt}
	if result.MediaAssetID > 0 {
		var asset model.MediaAsset
		err := model.DB.Select("public_id").Where("id = ? AND owner_user_id = ? AND status <> ?", result.MediaAssetID, ownerUserID, model.AssetStatusDeleted).First(&asset).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return view, err
		}
		view.AssetID = asset.PublicID
	}
	var err error
	view.PreviewURL, err = storageService.PresignImageStudioObject(ctx, result.StorageObjectID, ownerUserID, storageService.AssetObjectURLOriginal, "")
	if err == nil {
		view.ThumbnailURL, err = storageService.PresignImageStudioObject(ctx, result.StorageObjectID, ownerUserID, storageService.AssetObjectURLThumbnail, "")
	}
	if err == nil {
		view.DownloadURL, err = storageService.PresignImageStudioObject(ctx, result.StorageObjectID, ownerUserID, storageService.AssetObjectURLDownload, result.PublicID)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		view.PreviewURL, view.ThumbnailURL, view.DownloadURL = "", "", ""
		view.Expired = true
		return view, nil
	}
	return view, err
}

// Linked library asset handling when deleting generated images.
const (
	LinkedAssetsAsk    = ""
	LinkedAssetsKeep   = "keep"
	LinkedAssetsDelete = "delete"
)

// DeleteRun hides an owned run and removes its images; see DeleteResult for
// linkedAssets.
func DeleteRun(userID int, publicID string, linkedAssets string) error {
	run, err := model.FindDeletableImageStudioRun(userID, publicID)
	if err != nil {
		return err
	}
	results, err := model.ListImageStudioResults([]int64{run.ID})
	if err != nil {
		return err
	}
	if err = handleLinkedAssets(userID, results, linkedAssets); err != nil {
		return err
	}
	return model.DeleteImageStudioRun(userID, run, results)
}

// DeleteResult removes one owned image. When it was saved to the asset
// library and the asset still exists, linkedAssets must say whether the asset
// is kept or deleted through the asset library's own delete flow.
func DeleteResult(userID int, publicID string, linkedAssets string) error {
	result, _, err := model.FindOwnedImageStudioResult(userID, publicID)
	if err != nil {
		return err
	}
	results := []model.ImageStudioResult{*result}
	if err = handleLinkedAssets(userID, results, linkedAssets); err != nil {
		return err
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		return model.DeleteImageStudioResults(tx, userID, results)
	})
}

// ReviewDeleteRun removes another user's run from image review. The owner's
// asset library copies are always kept; only the generated images are removed.
func ReviewDeleteRun(publicID string) (int, error) {
	run, err := model.FindDeletableImageStudioRun(0, publicID)
	if err != nil {
		return 0, err
	}
	results, err := model.ListImageStudioResults([]int64{run.ID})
	if err != nil {
		return 0, err
	}
	return run.UserID, model.DeleteImageStudioRun(run.UserID, run, results)
}

// handleLinkedAssets applies the user's choice for library assets saved from
// results. Assets already deleted from the library need no choice.
func handleLinkedAssets(userID int, results []model.ImageStudioResult, linkedAssets string) error {
	if linkedAssets != LinkedAssetsAsk && linkedAssets != LinkedAssetsKeep && linkedAssets != LinkedAssetsDelete {
		return &RequestError{http.StatusBadRequest, "invalid_linked_assets", errors.New("linked_assets must be keep or delete")}
	}
	assetIDs := make([]int64, 0, len(results))
	for _, result := range results {
		if result.MediaAssetID > 0 {
			assetIDs = append(assetIDs, result.MediaAssetID)
		}
	}
	if len(assetIDs) == 0 {
		return nil
	}
	var assets []model.MediaAsset
	if err := model.DB.Select("public_id").Where("id IN ? AND owner_user_id = ? AND status <> ?", assetIDs, userID, model.AssetStatusDeleted).Find(&assets).Error; err != nil {
		return err
	}
	if len(assets) == 0 || linkedAssets == LinkedAssetsKeep {
		return nil
	}
	if linkedAssets == LinkedAssetsAsk {
		return &RequestError{http.StatusConflict, "linked_assets_exist", errors.New("this image was saved to the asset library; choose whether to keep the library copy")}
	}
	for _, asset := range assets {
		if err := assetService.DeleteAsset(asset.PublicID, userID, false); err != nil {
			return err
		}
	}
	return nil
}

// SaveResult imports an owned image into the asset library for video
// generation. It never generates or bills again; concurrent saves of the same
// image are rejected by an atomic claim so only one asset is created.
func SaveResult(ctx context.Context, userID int, publicID, groupID string) (string, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return "", &RequestError{http.StatusBadRequest, "invalid_group", errors.New("select an asset group")}
	}
	result, run, err := model.FindOwnedImageStudioResult(userID, publicID)
	if err != nil {
		return "", err
	}
	if result.MediaAssetID > 0 {
		view, err := buildResultView(ctx, *result, userID)
		if err != nil {
			return "", err
		}
		if view.AssetID != "" {
			return view.AssetID, nil
		}
		if err = model.DetachImageStudioResultAsset(result.ID, result.MediaAssetID); err != nil {
			return "", err
		}
	}
	if _, err = model.FindAssetGroup(groupID, userID, false); err != nil {
		return "", &RequestError{http.StatusBadRequest, "group_unavailable", errors.New("the selected asset group is unavailable")}
	}
	claimedAt, err := model.ClaimImageStudioResultSave(result.ID)
	if errors.Is(err, model.ErrImageStudioSaveInProgress) {
		return "", &RequestError{http.StatusConflict, "save_in_progress", err}
	}
	if err != nil {
		return "", err
	}
	sourceURL, err := storageService.PresignImageStudioObject(ctx, result.StorageObjectID, userID, storageService.AssetObjectURLOriginal, "")
	if err != nil {
		_ = model.ReleaseImageStudioResultSave(result.ID, claimedAt)
		return "", err
	}
	saveCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	name := fmt.Sprintf("%s-%d-%d", strings.ReplaceAll(run.Model, "/", "-"), run.CreatedAt, result.ResultIndex+1)
	asset, err := assetService.CreateAssetFromURL(saveCtx, userID, groupID, name, model.AssetTypeImage, sourceURL)
	if err != nil {
		_ = model.ReleaseImageStudioResultSave(result.ID, claimedAt)
		return "", err
	}
	// From here on the asset exists; any failure to link it deletes it through
	// the library flow so no unlinked asset is left behind.
	linked := false
	media, err := model.FindMediaAsset(asset.ID, userID, false)
	if err == nil {
		linked, err = model.CompleteImageStudioResultSave(result.ID, claimedAt, media.ID)
	}
	if err != nil || !linked {
		if deleteErr := assetService.DeleteAsset(asset.ID, userID, false); deleteErr != nil {
			common.SysError(fmt.Sprintf("image studio: delete unlinked asset %s failed: %v", asset.ID, deleteErr))
		}
		_ = model.ReleaseImageStudioResultSave(result.ID, claimedAt)
		if err == nil {
			err = &RequestError{http.StatusConflict, "save_in_progress", model.ErrImageStudioSaveInProgress}
		}
		return "", err
	}
	return asset.ID, nil
}
