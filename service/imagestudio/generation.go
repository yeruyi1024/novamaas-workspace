package imagestudio

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"
	"github.com/QuantumNous/new-api/service/authz"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"gorm.io/gorm"
)

type RunInput struct {
	TokenID        int              `json:"token_id"`
	IdempotencyKey string           `json:"idempotency_key"`
	Request        dto.ImageRequest `json:"request"`
	InputAssetIDs  []string         `json:"input_asset_ids"`
	ExtraFields    json.RawMessage  `json:"extra_fields,omitempty"`
}

type RelayResult struct {
	StatusCode int
	Body       []byte
	RequestID  string
	DurationMS int64
}

type RelayContext struct {
	Context             context.Context
	ClientIP            string
	Language            string
	OriginalRequestBody []byte
	uploadImage         func(context.Context, int, string, string, string, int64, io.Reader) (*model.StorageObject, error)
}

type RelayExecutor func(RelayContext, *model.Token, string, []byte, types.RelayFormat) RelayResult
type Actor struct{ UserID, Role int }

type ResultView struct {
	ID           string `json:"id"`
	AssetID      string `json:"asset_id"`
	PreviewURL   string `json:"preview_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Description  string `json:"description,omitempty"`
	Index        int    `json:"index"`
	SaveStatus   string `json:"save_status"`
	SaveError    string `json:"save_error,omitempty"`
}

type RunView struct {
	model.ImageStudioRun
	ID      string       `json:"id"`
	Results []ResultView `json:"results"`
}

type RunList struct {
	Items []*RunView `json:"items"`
	Total int64      `json:"total"`
	Users []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"users,omitempty"`
}

func CreateRun(execution RelayContext, userID int, input RunInput, relay RelayExecutor) (*RunView, bool, error) {
	key := strings.TrimSpace(input.IdempotencyKey)
	if key == "" || len(key) > 128 {
		return nil, false, &RequestError{http.StatusBadRequest, "invalid_idempotency_key", errors.New("idempotency_key must contain 1 to 128 characters")}
	}
	existing, err := findImageStudioRunByIdempotency(userID, key)
	if err == nil {
		view, viewErr := buildImageStudioRunView(execution.Context, existing, userID, true)
		return view, true, viewErr
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	if err = validateStudioImageRequest(&input.Request, input.ExtraFields, len(input.InputAssetIDs)); err != nil {
		return nil, false, err
	}
	token, err := resolveToken(userID, input.TokenID)
	if err != nil {
		return nil, false, err
	}
	if err = validateModel(userID, token, input.Request.Model, types.EndpointTypeImageGeneration); err != nil {
		return nil, false, err
	}
	assetURLs, err := resolveStudioAssetURLs(execution.Context, userID, input.InputAssetIDs)
	if err != nil {
		return nil, false, err
	}
	if len(assetURLs) > 0 {
		input.Request.Images, _ = common.Marshal(assetURLs)
		input.Request.Image = input.Request.Images
		if len(assetURLs) == 1 {
			input.Request.Image, _ = common.Marshal(assetURLs[0])
		}
	}
	if err := storageService.CheckImageStudioStorage(); err != nil {
		return nil, false, &RequestError{http.StatusServiceUnavailable, "image_storage_unavailable", errors.New("image storage is unavailable; configure permanent OSS storage before generating")}
	}
	now := common.GetTimestamp()
	run := &model.ImageStudioRun{
		PublicID: "run_" + common.GetUUID(), UserID: userID, IdempotencyKey: key,
		Status: model.ImageStudioRunStatusRunning, StartedAt: now, LeaseUntil: now + 300, Model: input.Request.Model, TokenID: token.Id,
		Prompt: input.Request.Prompt,
	}
	err = model.DB.Create(run).Error
	if err != nil {
		// Only the insert winner calls the relay. A duplicate request reads its result.
		if existing, findErr := findImageStudioRunByIdempotency(userID, key); findErr == nil {
			view, viewErr := buildImageStudioRunView(execution.Context, existing, userID, true)
			return view, true, viewErr
		}
		return nil, false, err
	}
	ctx, cancel := context.WithTimeout(execution.Context, 4*time.Minute)
	defer cancel()
	execution.Context = ctx
	path := "/v1/images/generations"
	if len(assetURLs) > 0 {
		path = "/v1/images/edits"
	}
	generationErr := generateImages(execution, run, token, input.Request, path, relay)
	if run.Status == model.ImageStudioRunStatusFailed || run.Status == model.ImageStudioRunStatusCanceled {
		run.ErrorMessage = CleanErrorMessage(generationErr.Error())
		if err = model.DiscardImageStudioRun(run); err != nil {
			common.SysError("failed to discard image studio run " + run.PublicID + ": " + err.Error())
			return nil, false, err
		}
		var requestErr *RequestError
		if errors.As(generationErr, &requestErr) {
			return nil, false, requestErr
		}
		return nil, false, &RequestError{http.StatusBadGateway, "generation_failed", errors.New(CleanErrorMessage(generationErr.Error()))}
	}
	if generationErr != nil {
		run.ErrorCode = ErrorCode(generationErr)
		run.ErrorMessage = CleanErrorMessage(generationErr.Error())
	}
	now = common.GetTimestamp()
	run.CompletedAt = now
	run.LeaseUntil = 0
	run.UpdatedAt = now
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(run).Error; err != nil {
			return err
		}
		objects := tx.Model(&model.ImageStudioResult{}).Select("storage_object_id").Where("run_id = ? AND storage_object_id > 0", run.ID)
		return tx.Model(&model.StorageObject{}).Where("id IN (?) AND owner_user_id = ? AND purpose = ? AND status = ?", objects, run.UserID, model.StorageObjectPurposeImageStudio, model.StorageObjectStatusUploaded).Update("delete_after", 0).Error
	})
	if err != nil {
		common.SysError("failed to finalize image studio run " + run.PublicID + ": " + err.Error())
		run.Status = model.ImageStudioRunStatusCanceled
		run.ErrorCode, run.ErrorMessage = "image_storage_failed", "record generated images failed"
		if cleanupErr := model.DiscardImageStudioRun(run); cleanupErr != nil {
			common.SysError("failed to discard image studio outputs " + run.PublicID + ": " + cleanupErr.Error())
		}
		return nil, false, &RequestError{http.StatusInternalServerError, run.ErrorCode, errors.New(run.ErrorMessage)}
	}
	view, err := buildImageStudioRunView(execution.Context, run, userID, true)
	return view, false, err
}

func ListRuns(ctx context.Context, actor Actor, review bool, ownerUserID int, page *common.PageInfo) (*RunList, bool, error) {
	if review && !authz.Can(actor.UserID, actor.Role, authz.ImageStudioRead) {
		return nil, false, &RequestError{http.StatusForbidden, "permission_denied", errors.New("image review permission is required")}
	}
	query := model.DB.Model(&model.ImageStudioRun{}).Where("status IN ?", []string{model.ImageStudioRunStatusCompleted, model.ImageStudioRunStatusPartial})
	if !review {
		query = query.Where("user_id = ?", actor.UserID)
	} else if ownerUserID > 0 {
		query = query.Where("user_id = ?", ownerUserID)
	}
	result := &RunList{Items: make([]*RunView, 0)}
	if review {
		// Keep the selector complete when the result page is filtered or empty.
		err := model.DB.Table("image_studio_runs").Select("image_studio_runs.user_id AS id, users.username AS name").
			Joins("LEFT JOIN users ON users.id = image_studio_runs.user_id").
			Where("image_studio_runs.status IN ?", []string{model.ImageStudioRunStatusCompleted, model.ImageStudioRunStatusPartial}).
			Group("image_studio_runs.user_id, users.username").Order("image_studio_runs.user_id asc").Scan(&result.Users).Error
		if err != nil {
			return nil, false, err
		}
	}
	if err := query.Count(&result.Total).Error; err != nil {
		return nil, false, err
	}
	var runs []*model.ImageStudioRun
	if err := query.Order("id desc").Limit(page.GetPageSize()).Offset(page.GetStartIdx()).Find(&runs).Error; err != nil {
		return nil, false, err
	}
	contentRead := !review || authz.Can(actor.UserID, actor.Role, authz.ImageStudioContentRead)
	runIDs := make([]int64, 0, len(runs))
	views := make(map[int64]*RunView, len(runs))
	for _, run := range runs {
		view := &RunView{ImageStudioRun: *run, ID: run.PublicID, Results: make([]ResultView, 0)}
		if !contentRead {
			view.Prompt, view.ErrorMessage = "", ""
		}
		runIDs = append(runIDs, run.ID)
		views[run.ID] = view
		result.Items = append(result.Items, view)
	}
	if !contentRead {
		return result, false, nil
	}
	rows, err := model.ListImageStudioResults(runIDs, false)
	if err != nil {
		return nil, false, err
	}
	for _, row := range rows {
		view, viewErr := studioResultView(ctx, row, views[row.RunID].UserID, false)
		if viewErr != nil {
			return nil, false, viewErr
		}
		views[row.RunID].Results = append(views[row.RunID].Results, view)
	}
	return result, review && contentRead, nil
}

func GetRun(ctx context.Context, actor Actor, publicID string, review bool) (*RunView, bool, error) {
	if review && !authz.Can(actor.UserID, actor.Role, authz.ImageStudioRead) {
		return nil, false, &RequestError{http.StatusForbidden, "permission_denied", errors.New("image review permission is required")}
	}
	query := model.DB.Where("public_id = ? AND status NOT IN ?", publicID, []string{model.ImageStudioRunStatusFailed, model.ImageStudioRunStatusCanceled})
	if !review {
		query = query.Where("user_id = ?", actor.UserID)
	}
	var run model.ImageStudioRun
	if err := query.First(&run).Error; err != nil {
		return nil, false, err
	}
	contentRead := !review || authz.Can(actor.UserID, actor.Role, authz.ImageStudioContentRead)
	view, err := buildImageStudioRunView(ctx, &run, actor.UserID, contentRead)
	return view, review && contentRead, err
}

func findImageStudioRunByIdempotency(userID int, key string) (*model.ImageStudioRun, error) {
	var run model.ImageStudioRun
	if err := model.DB.Where("user_id = ? AND idempotency_key = ?", userID, key).First(&run).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func buildImageStudioRunView(ctx context.Context, run *model.ImageStudioRun, viewerUserID int, contentRead bool) (*RunView, error) {
	if err := model.FailExpiredImageStudioRun(run, common.GetTimestamp()); err != nil {
		return nil, err
	}
	if run.Status == model.ImageStudioRunStatusFailed || run.Status == model.ImageStudioRunStatusCanceled {
		return nil, &RequestError{http.StatusGone, "run_unavailable", errors.New("image generation record is no longer available")}
	}
	view := &RunView{ImageStudioRun: *run, ID: run.PublicID, Results: make([]ResultView, 0)}
	if !contentRead {
		view.Prompt = ""
		view.ErrorMessage = ""
		return view, nil
	}
	if run.Status == model.ImageStudioRunStatusRunning {
		return view, nil
	}
	rows, err := model.ListImageStudioResults([]int64{run.ID}, true)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result, resultErr := studioResultView(ctx, row, run.UserID, true)
		if resultErr != nil {
			return nil, resultErr
		}
		if result.PreviewURL == "" && row.AssetPublicID != "" {
			if value, _, err := assetService.PreviewURL(ctx, row.AssetPublicID, viewerUserID, viewerUserID != run.UserID); err == nil {
				result.PreviewURL = value
			}
		}
		view.Results = append(view.Results, result)
	}
	return view, nil
}

func studioResultView(ctx context.Context, row model.ImageStudioResultWithAsset, ownerUserID int, originals bool) (ResultView, error) {
	result := ResultView{ID: row.PublicID, AssetID: row.AssetPublicID, PreviewURL: string(row.PreviewURL), ThumbnailURL: string(row.ThumbnailURL), Description: row.RevisedPrompt, Index: row.ResultIndex, SaveStatus: row.SaveStatus, SaveError: row.SaveError}
	if row.MediaAssetID != 0 && row.AssetPublicID == "" {
		result.SaveStatus, result.SaveError = "not_requested", ""
	}
	if row.StorageObjectID > 0 {
		var err error
		result.ThumbnailURL, err = storageService.PresignImageStudioObject(ctx, row.StorageObjectID, ownerUserID, storageService.AssetObjectURLThumbnail)
		if err != nil {
			return result, err
		}
		if originals {
			result.PreviewURL, err = storageService.PresignImageStudioObject(ctx, row.StorageObjectID, ownerUserID, storageService.AssetObjectURLOriginal)
			if err != nil {
				return result, err
			}
		}
	}
	return result, nil
}
