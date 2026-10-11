package imagestudio

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strings"

	_ "golang.org/x/image/webp"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	coreService "github.com/QuantumNous/new-api/service"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

const maxStudioImageBytes = 64 << 20

func generateImages(execution RelayContext, run *model.ImageStudioRun, token *model.Token, request dto.ImageRequest, path string, relay RelayExecutor) error {
	run.Status = model.ImageStudioRunStatusFailed
	if !operation_setting.GetImageStudioSetting().Enabled {
		return &RequestError{http.StatusServiceUnavailable, "image_studio_disabled", errors.New("AI Studio is disabled")}
	}
	body, err := common.Marshal(request)
	if err != nil {
		return err
	}
	result := relay(execution, token, path, body, types.RelayFormatOpenAIImage)
	run.RequestID = result.RequestID
	run.DurationMS = result.DurationMS
	run.Quota = QuotaForRequest(result.RequestID, run.UserID)
	if result.StatusCode >= http.StatusBadRequest {
		message := relayErrorMessage(result.Body)
		return errors.New(message)
	}
	// Keep only hidden request metadata after an accepted, potentially billed response.
	if result.RequestID != "" {
		run.Status = model.ImageStudioRunStatusCanceled
	}
	var response dto.ImageResponse
	if err = common.Unmarshal(result.Body, &response); err != nil {
		return fmt.Errorf("decode image response: %w", err)
	}
	if len(response.Data) == 0 {
		err = errors.New("image provider returned no results")
		return err
	}
	stored := 0
	for index, output := range response.Data {
		resultRecord := &model.ImageStudioResult{
			RunID: run.ID, PublicID: "result_" + common.GetUUID(),
			ResultIndex: index, RevisedPrompt: output.RevisedPrompt,
		}
		data, contentType, decodeErr := decodeStudioImage(output)
		if decodeErr != nil {
			err = decodeErr
			continue
		}
		if data == nil {
			var downloadErr error
			data, contentType, downloadErr = downloadStudioImage(execution.Context, output.Url)
			if downloadErr != nil {
				err = downloadErr
				continue
			}
		}
		config, _, imageErr := image.DecodeConfig(bytes.NewReader(data))
		if imageErr != nil || config.Width <= 0 || config.Height <= 0 || int64(config.Width) > maxStudioImageBytes/4/int64(config.Height) {
			err = errors.New("image provider returned an invalid or oversized image")
			continue
		}
		upload := execution.uploadImage
		if upload == nil {
			upload = storageService.UploadImageStudioObject
		}
		object, uploadErr := upload(execution.Context, run.UserID, run.PublicID, resultRecord.PublicID, contentType, int64(len(data)), bytes.NewReader(data))
		if uploadErr != nil {
			run.Status = model.ImageStudioRunStatusCanceled
			err = &RequestError{http.StatusBadGateway, "image_storage_failed", errors.New("store generated image failed: " + CleanErrorMessage(uploadErr.Error()))}
			continue
		}
		resultRecord.SaveStatus = "not_requested"
		resultRecord.StorageObjectID = object.ID
		if createErr := model.DB.Create(resultRecord).Error; createErr != nil {
			_ = storageService.DeleteAssetObject(object.ID, "record generated image failed")
			run.Status = model.ImageStudioRunStatusCanceled
			err = &RequestError{http.StatusInternalServerError, "image_storage_failed", errors.New("record generated image failed")}
			continue
		}
		stored++
	}
	if stored == 0 {
		if err == nil {
			err = errors.New("generated images could not be recorded")
		}
		return err
	}
	run.Status = model.ImageStudioRunStatusCompleted
	if stored != len(response.Data) {
		run.Status = model.ImageStudioRunStatusPartial
		return err
	}
	return nil
}

func validateStudioImageRequest(request *dto.ImageRequest, extraFields json.RawMessage, inputCount int) error {
	settings := operation_setting.GetImageStudioSetting()
	if !settings.Enabled {
		return &RequestError{StatusCode: http.StatusServiceUnavailable, Code: "image_studio_disabled", Err: errors.New("AI Studio is disabled")}
	}
	request.Model = strings.TrimSpace(request.Model)
	request.Prompt = strings.TrimSpace(request.Prompt)
	if request.Model == "" {
		return &RequestError{StatusCode: http.StatusBadRequest, Code: "model_required", Err: errors.New("model is required")}
	}
	if request.Prompt == "" || len([]rune(request.Prompt)) > operation_setting.DefaultImageStudioMaxPromptCharacters {
		return &RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_prompt", Err: fmt.Errorf("prompt must contain 1 to %d characters", operation_setting.DefaultImageStudioMaxPromptCharacters)}
	}
	if request.N != nil && (*request.N == 0 || *request.N > uint(min(operation_setting.DefaultImageStudioMaxImages, dto.MaxImageN))) {
		return &RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_image_count", Err: fmt.Errorf("n must be between 1 and %d", min(operation_setting.DefaultImageStudioMaxImages, dto.MaxImageN))}
	}
	if request.Stream != nil && *request.Stream {
		return &RequestError{StatusCode: http.StatusBadRequest, Code: "stream_not_supported", Err: errors.New("studio runs do not support streaming; use the public Images API for streamed partial images")}
	}
	for field, value := range map[string]json.RawMessage{
		"image": request.Image, "images": request.Images, "mask": request.Mask,
	} {
		if normalized := strings.TrimSpace(string(value)); normalized != "" && normalized != "null" {
			return &RequestError{StatusCode: http.StatusBadRequest, Code: "asset_ids_required", Err: fmt.Errorf("request.%s is not accepted by AI Studio; use asset IDs", field)}
		}
	}
	if inputCount > operation_setting.DefaultImageStudioMaxReferenceImages {
		return &RequestError{StatusCode: http.StatusBadRequest, Code: "too_many_input_assets", Err: fmt.Errorf("no more than %d input assets are allowed", operation_setting.DefaultImageStudioMaxReferenceImages)}
	}
	if len(extraFields) > 0 && strings.TrimSpace(string(extraFields)) != "null" {
		return &RequestError{http.StatusBadRequest, "unsupported_parameters", errors.New("extra_fields is not supported by image studio")}
	}
	if request.ResponseFormat != "" && request.ResponseFormat != "url" && request.ResponseFormat != "b64_json" {
		return &RequestError{http.StatusBadRequest, "invalid_response_format", errors.New("response_format must be url or b64_json")}
	}
	// Quality follows the provider default; the studio does not send it.
	// The studio only exposes these parameters. Provider-specific payloads must
	// use the public relay API and cannot override the studio's billing count.
	*request = dto.ImageRequest{
		Model: request.Model, Prompt: request.Prompt, N: request.N,
		Size: request.Size, ResponseFormat: request.ResponseFormat,
	}
	return nil
}

func loadInputAssets(userID int, publicIDs []string) ([]*model.MediaAsset, error) {
	if len(publicIDs) == 0 {
		return nil, nil
	}
	if !operation_setting.GetImageStudioSetting().AssetLibraryEnabled {
		return nil, &RequestError{StatusCode: http.StatusServiceUnavailable, Code: "asset_library_disabled", Err: errors.New("asset library integration is disabled")}
	}
	ids := make([]string, len(publicIDs))
	for i, id := range publicIDs {
		ids[i] = strings.TrimSpace(id)
	}
	var assets []*model.MediaAsset
	if err := model.DB.Where("public_id IN ? AND owner_user_id = ? AND asset_type = ? AND status = ?", ids, userID, model.AssetTypeImage, model.AssetStatusReady).Find(&assets).Error; err != nil {
		return nil, err
	}
	byID := make(map[string]*model.MediaAsset, len(assets))
	for _, asset := range assets {
		byID[asset.PublicID] = asset
	}
	result := make([]*model.MediaAsset, 0, len(ids))
	for _, id := range ids {
		asset := byID[id]
		if asset == nil {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Code: "input_asset_unavailable", Err: fmt.Errorf("image asset %s is unavailable", id)}
		}
		result = append(result, asset)
	}
	return result, nil
}

func resolveStudioAssetURLs(ctx context.Context, userID int, publicIDs []string) ([]string, error) {
	assets, err := loadInputAssets(userID, publicIDs)
	if err != nil {
		return nil, err
	}
	urls := make([]string, 0, len(assets))
	for _, asset := range assets {
		value, _, err := assetService.PreviewURL(ctx, asset.PublicID, userID, false)
		if err != nil {
			return nil, err
		}
		urls = append(urls, value)
	}
	return urls, nil
}

func relayErrorMessage(body []byte) string {
	var payload struct {
		Error   any    `json:"error"`
		Message string `json:"message"`
	}
	if err := common.Unmarshal(body, &payload); err == nil {
		if payload.Message != "" {
			return CleanErrorMessage(payload.Message)
		}
		if payload.Error != nil {
			if value, ok := payload.Error.(map[string]any); ok {
				if message, ok := value["message"].(string); ok {
					return CleanErrorMessage(message)
				}
			}
			return CleanErrorMessage(fmt.Sprint(payload.Error))
		}
	}
	if len(body) > 0 {
		return CleanErrorMessage(string(body))
	}
	return "model provider request failed"
}

// The same validation is used when recording an image and importing it later.
func decodeStudioImage(image dto.ImageData) (data []byte, contentType string, err error) {
	payload := strings.TrimSpace(image.B64Json)
	if payload == "" {
		parsed, parseErr := url.Parse(strings.TrimSpace(image.Url))
		if parseErr != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return nil, "", errors.New("image provider returned an invalid image URL")
		}
		return nil, "", nil
	}
	if strings.HasPrefix(payload, "data:image/") {
		_, payload, _ = strings.Cut(payload, ",")
	}
	if len(payload) > base64.StdEncoding.EncodedLen(maxStudioImageBytes) {
		return nil, "", errors.New("generated image exceeds the studio storage limit")
	}
	data, err = base64.StdEncoding.DecodeString(payload)
	if err != nil || len(data) == 0 || len(data) > maxStudioImageBytes {
		return nil, "", errors.New("generated image contains invalid base64 data")
	}
	contentType = http.DetectContentType(data[:min(len(data), 512)])
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", errors.New("generated content is not an image")
	}
	return data, contentType, nil
}

// URL-only providers use the same bounded, SSRF-protected download path as imports.
func downloadStudioImage(ctx context.Context, sourceURL string) ([]byte, string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, "", errors.New("invalid generated image URL")
	}
	response, err := coreService.GetSSRFProtectedHTTPClient().Do(request)
	if err != nil {
		return nil, "", errors.New("download generated image failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 || response.ContentLength > maxStudioImageBytes {
		return nil, "", errors.New("generated image download is unavailable or oversized")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxStudioImageBytes+1))
	if err != nil || len(data) == 0 || len(data) > maxStudioImageBytes {
		return nil, "", errors.New("generated image download is empty or oversized")
	}
	return data, http.DetectContentType(data[:min(512, len(data))]), nil
}
