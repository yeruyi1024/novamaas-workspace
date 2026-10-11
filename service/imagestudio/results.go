package imagestudio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"gorm.io/gorm"
)

func DeleteResult(userID int, publicID string) error {
	err := model.DeleteImageStudioResult(userID, publicID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &RequestError{http.StatusNotFound, "result_not_found", errors.New("image result not found")}
	}
	if errors.Is(err, model.ErrImageStudioResultBusy) {
		return &RequestError{http.StatusConflict, "image_busy", err}
	}
	return err
}

// SaveResult imports an owned result without generating or billing again.
func SaveResult(ctx context.Context, userID int, publicID, groupID string) error {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return &RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_group", Err: errors.New("select an asset group")}
	}
	if !operation_setting.GetImageStudioSetting().AssetLibraryEnabled {
		return &RequestError{http.StatusServiceUnavailable, "asset_library_disabled", errors.New("asset library integration is disabled")}
	}

	var result model.ImageStudioResult
	err := model.DB.Table("image_studio_results").Select("image_studio_results.*").
		Joins("JOIN image_studio_runs ON image_studio_runs.id = image_studio_results.run_id").
		Where("image_studio_results.public_id = ? AND image_studio_runs.user_id = ?", publicID, userID).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &RequestError{StatusCode: http.StatusNotFound, Code: "result_not_found", Err: errors.New("image result not found")}
	}
	if err != nil {
		return err
	}
	var run model.ImageStudioRun
	if err := model.DB.First(&run, result.RunID).Error; err != nil {
		return err
	}
	if run.Status != model.ImageStudioRunStatusCompleted && run.Status != model.ImageStudioRunStatusPartial {
		return &RequestError{http.StatusConflict, "image_busy", model.ErrImageStudioResultBusy}
	}
	if _, err := model.FindAssetGroup(groupID, userID, false); err != nil {
		return &RequestError{http.StatusBadRequest, "output_group_unavailable", errors.New("the selected asset group is unavailable")}
	}
	return saveImageStudioResult(ctx, &result, userID, groupID, run.Model)
}

func saveImageStudioResult(ctx context.Context, result *model.ImageStudioResult, userID int, groupID, modelName string) (err error) {
	if result.MediaAssetID != 0 {
		var asset model.MediaAsset
		err := model.DB.Select("id").Where("id = ? AND owner_user_id = ? AND status <> ?", result.MediaAssetID, userID, model.AssetStatusDeleted).First(&asset).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// Only detach the stale link. A concurrent retry must keep its save lease.
		if err := model.DB.Model(result).Where("media_asset_id = ?", result.MediaAssetID).
			Updates(map[string]any{"media_asset_id": 0, "save_status": "not_requested", "save_error": ""}).Error; err != nil {
			return err
		}
		if err := model.DB.First(result, result.ID).Error; err != nil {
			return err
		}
	}
	attemptID := common.GetUUID()
	claimed, err := model.ClaimImageStudioResultSave(result.ID, groupID, attemptID, common.GetTimestamp())
	if err != nil {
		return err
	}
	if !claimed {
		var current model.ImageStudioResult
		if findErr := model.DB.First(&current, "id = ?", result.ID).Error; findErr != nil {
			return findErr
		}
		if current.MediaAssetID != 0 {
			return nil
		}
		return &RequestError{StatusCode: http.StatusConflict, Code: "save_in_progress", Err: errors.New("image is already being saved")}
	}
	defer func() {
		if err != nil {
			_ = model.DB.Model(&model.ImageStudioResult{}).Where("id = ? AND save_attempt_id = ?", result.ID, attemptID).
				Updates(map[string]any{"save_status": "failed", "save_error": CleanErrorMessage(err.Error())}).Error
		}
	}()
	if result.StorageObjectID > 0 {
		value, signErr := storageService.PresignImageStudioObject(ctx, result.StorageObjectID, userID, storageService.AssetObjectURLOriginal)
		if signErr != nil {
			return signErr
		}
		result.PreviewURL = model.ImageStudioPreview(value)
	}
	image := dto.ImageData{Url: string(result.PreviewURL)}
	if strings.HasPrefix(string(result.PreviewURL), "data:image/") {
		image.Url = ""
		image.B64Json = string(result.PreviewURL)
	}
	saveCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	asset, err := storeStudioImage(saveCtx, userID, groupID, modelName, result.ResultIndex, image)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = assetService.DeleteAsset(asset.ID, userID, false)
		}
	}()
	media, err := model.FindMediaAsset(asset.ID, userID, false)
	if err != nil {
		return err
	}
	update := model.DB.Model(&model.ImageStudioResult{}).
		Where("id = ? AND save_attempt_id = ? AND media_asset_id = 0", result.ID, attemptID).
		Updates(map[string]any{"media_asset_id": media.ID, "save_status": "saved", "save_error": ""})
	if update.Error != nil {
		return update.Error
	}
	if update.RowsAffected != 1 {
		return errors.New("image save attempt expired; retry saving")
	}
	result.MediaAssetID = media.ID
	result.SaveStatus = "saved"
	result.SaveGroupID = groupID
	return nil
}

func storeStudioImage(ctx context.Context, userID int, groupID string, modelName string, index int, image dto.ImageData) (*assetService.AssetView, error) {
	name := fmt.Sprintf("%s-%d-%d", strings.ReplaceAll(modelName, "/", "-"), time.Now().Unix(), index+1)
	data, contentType, err := decodeStudioImage(image)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return assetService.CreateAssetFromURL(ctx, userID, groupID, name, model.AssetTypeImage, image.Url)
	}
	return assetService.CreateAsset(ctx, userID, assetService.AssetInput{
		GroupPublicID: groupID, Name: name, AssetType: model.AssetTypeImage,
		FileName: name + "." + strings.TrimPrefix(contentType, "image/"), ContentType: contentType,
		Size: int64(len(data)), Body: bytes.NewReader(data),
	})
}
