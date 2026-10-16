package imagestudio

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	_ "golang.org/x/image/webp"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	coreService "github.com/QuantumNous/new-api/service"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"gorm.io/gorm"
)

const maxStudioImageBytes = 64 << 20

// uploadImage is replaced in tests; production stores images in private OSS.
var uploadImage = storageService.UploadImageStudioObject

type RunRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	N      *uint  `json:"n,omitempty"`
	Size   string `json:"size,omitempty"`
}

type RunInput struct {
	TokenID        int        `json:"token_id"`
	IdempotencyKey string     `json:"idempotency_key"`
	Request        RunRequest `json:"request"`
}

// CreateRun generates images once per (user, idempotency key). A repeated key
// returns the existing run, including a run that is still generating, and
// never calls the relay again. The second return value reports a replay.
func CreateRun(execution RelayContext, userID int, input RunInput, original []byte, relay RelayExecutor) (*RunView, bool, error) {
	key := strings.TrimSpace(input.IdempotencyKey)
	if key == "" || len(key) > 128 {
		return nil, false, &RequestError{http.StatusBadRequest, "invalid_idempotency_key", errors.New("idempotency_key must contain 1 to 128 characters")}
	}
	if existing, err := model.FindImageStudioRunByIdempotencyKey(userID, key); err == nil {
		return replayRun(execution.Context, existing)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	if !operation_setting.GetImageStudioSetting().Enabled {
		return nil, false, &RequestError{http.StatusServiceUnavailable, "image_studio_disabled", errors.New("image generation is disabled")}
	}
	request := input.Request
	request.Model = strings.TrimSpace(request.Model)
	request.Prompt = strings.TrimSpace(request.Prompt)
	request.Size = strings.TrimSpace(request.Size)
	if request.Model == "" {
		return nil, false, &RequestError{http.StatusBadRequest, "model_required", errors.New("model is required")}
	}
	if request.Prompt == "" || len([]rune(request.Prompt)) > operation_setting.ImageStudioMaxPromptCharacters {
		return nil, false, &RequestError{http.StatusBadRequest, "invalid_prompt", fmt.Errorf("prompt must contain 1 to %d characters", operation_setting.ImageStudioMaxPromptCharacters)}
	}
	if request.N != nil && (*request.N == 0 || *request.N > operation_setting.ImageStudioMaxImages) {
		return nil, false, &RequestError{http.StatusBadRequest, "invalid_image_count", fmt.Errorf("n must be between 1 and %d", operation_setting.ImageStudioMaxImages)}
	}
	token, err := resolveToken(userID, input.TokenID)
	if err != nil {
		return nil, false, err
	}
	if err = validateModel(userID, token, request.Model, types.EndpointTypeImageGeneration); err != nil {
		return nil, false, err
	}
	if _, _, _, err = storageService.ResolveEnabledPolicy(model.StoragePolicyAssetLibrary); err != nil {
		return nil, false, &RequestError{http.StatusServiceUnavailable, "image_storage_unavailable", errors.New("image storage is unavailable; enable permanent asset storage before generating")}
	}
	run := &model.ImageStudioRun{
		UserID: userID, IdempotencyKey: key, TokenID: token.Id,
		Model: request.Model, Prompt: request.Prompt, Status: model.ImageStudioRunStatusRunning,
	}
	if err = model.DB.Create(run).Error; err != nil {
		// Only the insert winner calls the relay; a concurrent duplicate reads it.
		if existing, findErr := model.FindImageStudioRunByIdempotencyKey(userID, key); findErr == nil {
			return replayRun(execution.Context, existing)
		}
		return nil, false, err
	}
	body, err := common.Marshal(dto.ImageRequest{Model: request.Model, Prompt: request.Prompt, N: request.N, Size: request.Size})
	if err != nil {
		return nil, false, err
	}
	ctx, cancel := context.WithTimeout(execution.Context, 4*time.Minute)
	defer cancel()
	execution.Context = ctx
	if len(original) == 0 {
		original = body
	}
	result := relay(execution, token, "/v1/images/generations", body, types.RelayFormatOpenAIImage, &RequestSnapshot{RunID: run.ID, Original: original})
	run.RequestID = result.RequestID
	run.DurationMS = result.DurationMS
	run.Quota = quotaForRequest(result.RequestID, userID)
	results, generationErr := uploadImages(execution.Context, run, result)
	run.Status = model.ImageStudioRunStatusCompleted
	if len(results) == 0 {
		run.Status = model.ImageStudioRunStatusFailed
		run.ErrorMessage = CleanErrorMessage(generationErr.Error())
	} else if generationErr != nil {
		run.ErrorMessage = CleanErrorMessage(generationErr.Error())
	}
	if err = model.FinishImageStudioRun(run, results); err != nil {
		// The results were not recorded, so their objects go to the cleanup queue.
		for _, record := range results {
			_ = storageService.DeleteAssetObject(record.StorageObjectID, "record generated image failed")
		}
		if errors.Is(err, model.ErrImageStudioRunFinished) {
			var current model.ImageStudioRun
			if findErr := model.DB.First(&current, run.ID).Error; findErr != nil {
				return nil, false, findErr
			}
			return replayRun(execution.Context, &current)
		}
		return nil, false, err
	}
	if run.Status == model.ImageStudioRunStatusFailed {
		status := http.StatusBadGateway
		if result.StatusCode >= http.StatusBadRequest && result.StatusCode < http.StatusInternalServerError {
			status = result.StatusCode
		}
		return nil, false, &RequestError{status, "generation_failed", errors.New(run.ErrorMessage)}
	}
	view, err := buildRunView(execution.Context, run)
	return view, false, err
}

func replayRun(ctx context.Context, run *model.ImageStudioRun) (*RunView, bool, error) {
	if err := model.FailStaleImageStudioRun(run); err != nil {
		return nil, true, err
	}
	if run.Status == model.ImageStudioRunStatusDeleted {
		return nil, true, &RequestError{http.StatusGone, "run_deleted", errors.New("this image generation was deleted")}
	}
	view, err := buildRunView(ctx, run)
	return view, true, err
}

// uploadImages copies every returned image into private OSS storage and returns
// the result records to commit with the run. Partial success keeps the
// uploaded images; the error describes the images that were lost.
func uploadImages(ctx context.Context, run *model.ImageStudioRun, result RelayResult) ([]*model.ImageStudioResult, error) {
	if result.StatusCode >= http.StatusBadRequest {
		return nil, errors.New(relayErrorMessage(result.Body))
	}
	var response dto.ImageResponse
	if err := common.Unmarshal(result.Body, &response); err != nil {
		return nil, fmt.Errorf("decode image response: %w", err)
	}
	if len(response.Data) == 0 {
		return nil, errors.New("image provider returned no results")
	}
	records := make([]*model.ImageStudioResult, 0, len(response.Data))
	var lastErr error
	for index, output := range response.Data {
		data, err := readGeneratedImage(ctx, output)
		if err != nil {
			lastErr = err
			continue
		}
		// The stored content type, object key and download name follow the
		// decoded format, so JPEG and WebP outputs keep their own extension.
		config, format, err := image.DecodeConfig(bytes.NewReader(data))
		contentType := map[string]string{"png": "image/png", "jpeg": "image/jpeg", "webp": "image/webp"}[format]
		if err != nil || contentType == "" || config.Width <= 0 || config.Height <= 0 {
			lastErr = errors.New("image provider returned an invalid image")
			continue
		}
		record := &model.ImageStudioResult{PublicID: "result_" + common.GetUUID(), RunID: run.ID, ResultIndex: index, RevisedPrompt: output.RevisedPrompt}
		object, err := uploadImage(ctx, run.UserID, run.PublicID, record.PublicID, contentType, data)
		if err != nil {
			lastErr = fmt.Errorf("store generated image: %w", err)
			continue
		}
		record.StorageObjectID = object.ID
		records = append(records, record)
	}
	if len(records) < len(response.Data) && lastErr != nil {
		return records, fmt.Errorf("%d of %d images could not be stored: %w", len(response.Data)-len(records), len(response.Data), lastErr)
	}
	return records, nil
}

// readGeneratedImage decodes base64 output or downloads a URL output through
// the SSRF-protected client, bounded by maxStudioImageBytes.
func readGeneratedImage(ctx context.Context, output dto.ImageData) ([]byte, error) {
	payload := strings.TrimSpace(output.B64Json)
	if payload != "" {
		if strings.HasPrefix(payload, "data:") {
			_, payload, _ = strings.Cut(payload, ",")
		}
		if len(payload) > base64.StdEncoding.EncodedLen(maxStudioImageBytes) {
			return nil, errors.New("generated image exceeds the storage limit")
		}
		data, err := base64.StdEncoding.DecodeString(payload)
		if err != nil || len(data) == 0 {
			return nil, errors.New("generated image contains invalid base64 data")
		}
		return data, nil
	}
	parsed, err := url.Parse(strings.TrimSpace(output.Url))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, errors.New("image provider returned an invalid image URL")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, errors.New("invalid generated image URL")
	}
	response, err := coreService.GetSSRFProtectedHTTPClient().Do(request)
	if err != nil {
		return nil, errors.New("download generated image failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("generated image download returned HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxStudioImageBytes+1))
	if err != nil || len(data) == 0 || len(data) > maxStudioImageBytes {
		return nil, errors.New("generated image download is empty or oversized")
	}
	return data, nil
}
