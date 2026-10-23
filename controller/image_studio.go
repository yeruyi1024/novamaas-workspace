package controller

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service/authz"
	studioService "github.com/QuantumNous/new-api/service/imagestudio"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

func GetImageStudioSettings(c *gin.Context) {
	common.ApiSuccess(c, gin.H{
		"enabled":               operation_setting.GetImageStudioSetting().Enabled,
		"max_images":            operation_setting.ImageStudioMaxImages,
		"max_running_runs":      operation_setting.ImageStudioMaxRunningPerUser,
		"max_prompt_characters": operation_setting.ImageStudioMaxPromptCharacters,
		"max_reference_images":  studioService.MaxReferenceImages,
		"max_reference_bytes":   studioService.MaxReferenceImageBytes,
	})
}

func ListImageStudioReferences(c *gin.Context) {
	list, err := studioService.ListReferences(c.Request.Context(), c.GetInt("id"), common.GetPageQuery(c))
	if err != nil {
		imageStudioError(c, err)
		return
	}
	common.ApiSuccess(c, list)
}

func UploadImageStudioReference(c *gin.Context) {
	if strings.HasPrefix(c.GetHeader("Content-Type"), gin.MIMEJSON) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
		var input struct {
			URL string `json:"url"`
		}
		if err := common.DecodeJson(c.Request.Body, &input); err != nil {
			imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_reference_image", Err: errors.New("invalid reference image URL")})
			return
		}
		view, err := studioService.ImportReference(c.Request.Context(), c.GetInt("id"), input.URL)
		if err != nil {
			imageStudioError(c, err)
			return
		}
		common.ApiSuccess(c, view)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, studioService.MaxReferenceImageBytes+1024*1024)
	header, err := c.FormFile("file")
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	if err != nil {
		imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_reference_image", Err: errors.New("Reference images must be PNG, JPEG or WebP and no larger than 20 MB.")})
		return
	}
	file, err := header.Open()
	if err != nil {
		imageStudioError(c, err)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, studioService.MaxReferenceImageBytes+1))
	if err != nil {
		imageStudioError(c, err)
		return
	}
	view, err := studioService.UploadReference(c.Request.Context(), c.GetInt("id"), data)
	if err != nil {
		imageStudioError(c, err)
		return
	}
	common.ApiSuccess(c, view)
}

func ListImageStudioModels(c *gin.Context) {
	tokenID, _ := strconv.Atoi(c.Query("token_id"))
	endpointType := constant.EndpointTypeImageGeneration
	if c.Query("kind") == "text" {
		endpointType = constant.EndpointTypeOpenAI
	}
	models, err := studioService.ModelsForToken(c.GetInt("id"), tokenID, endpointType)
	if err != nil {
		imageStudioError(c, err)
		return
	}
	common.ApiSuccess(c, models)
}

func PromptAssistImageStudio(c *gin.Context) {
	var input studioService.PromptAssistInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_request", Err: errors.New("invalid prompt optimization request")})
		return
	}
	result, err := studioService.AssistPrompt(imageStudioRelayContext(c), c.GetInt("id"), input, RelayImageStudio)
	if err != nil {
		imageStudioError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func CreateImageStudioRun(c *gin.Context) {
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, model.MaxTaskRequestBodyBytes))
	var input studioService.RunInput
	if err == nil {
		err = common.Unmarshal(body, &input)
	}
	if err != nil {
		imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_request", Err: errors.New("invalid image generation request")})
		return
	}
	if header := strings.TrimSpace(c.GetHeader("Idempotency-Key")); header != "" {
		if input.IdempotencyKey != "" && strings.TrimSpace(input.IdempotencyKey) != header {
			imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "idempotency_key_mismatch", Err: errors.New("body and header idempotency keys do not match")})
			return
		}
		input.IdempotencyKey = header
	}
	view, replay, err := studioService.CreateRun(imageStudioRelayContext(c), c.GetInt("id"), input, body, RelayImageStudio)
	if err != nil {
		imageStudioError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": view, "idempotent_replay": replay})
}

// ListImageStudioLogs remains available when generation is disabled.
func ListImageStudioLogs(c *gin.Context) {
	owner := c.GetInt("id")
	if c.Query("scope") == "all" {
		if !canReviewImageStudio(c) {
			return
		}
		owner = 0
	}
	start, startErr := strconv.ParseInt(c.DefaultQuery("start_timestamp", "0"), 10, 64)
	end, endErr := strconv.ParseInt(c.DefaultQuery("end_timestamp", "0"), 10, 64)
	id := strings.TrimSpace(c.Query("filter"))
	channelID, channelErr := strconv.Atoi(c.DefaultQuery("channel_id", "0"))
	if channelErr != nil || channelID < 0 || (channelID > 0 && owner > 0) || startErr != nil || endErr != nil || start < 0 || end < 0 || (end > 0 && start > end) || len(id) > 64 {
		imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_log_filter", Err: errors.New("invalid image task log filter")})
		return
	}
	page := common.GetPageQuery(c)
	page.Page, page.PageSize = max(1, page.Page), max(1, page.PageSize)
	list, err := studioService.ListTaskLogs(c.Request.Context(), studioService.TaskLogFilter{
		OwnerUserID: owner, ChannelID: channelID, ID: id, Start: start, End: end, Page: page,
	})
	if err != nil {
		imageStudioError(c, err)
		return
	}
	common.ApiSuccess(c, list)
}

func ListImageStudioRuns(c *gin.Context) {
	filter := studioService.ListFilter{OwnerUserID: c.GetInt("id"), Page: common.GetPageQuery(c)}
	if c.Query("scope") == "all" {
		if !canReviewImageStudio(c) {
			return
		}
		filter.Review = true
		filter.OwnerUserID, _ = strconv.Atoi(c.Query("user_id"))
	}
	list, err := studioService.ListRuns(c.Request.Context(), filter)
	if err != nil {
		imageStudioError(c, err)
		return
	}
	common.ApiSuccess(c, list)
}

func PollImageStudioRuns(c *gin.Context) {
	list, err := studioService.PollRuns(c.Request.Context(), c.GetInt("id"), strings.Split(c.Query("ids"), ","))
	if err != nil {
		imageStudioError(c, err)
		return
	}
	common.ApiSuccess(c, list)
}

func GetImageStudioRun(c *gin.Context) {
	ownerUserID := c.GetInt("id")
	review := c.Query("scope") == "all"
	if review {
		if !canReviewImageStudio(c) {
			return
		}
		ownerUserID = 0
	}
	view, err := studioService.GetRun(c.Request.Context(), ownerUserID, c.Param("id"))
	if err != nil {
		imageStudioError(c, err)
		return
	}
	if review {
		recordManageAuditFor(c, view.UserID, "image_studio.review_read", map[string]interface{}{"run_id": view.PublicID})
	}
	common.ApiSuccess(c, view)
}

func DeleteImageStudioRun(c *gin.Context) {
	if c.Query("scope") == "all" {
		if !canReviewImageStudio(c) {
			return
		}
		if !authz.Can(c.GetInt("id"), c.GetInt("role"), authz.ImageStudioWrite) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "code": "permission_denied", "message": "image review delete permission is required"})
			return
		}
		ownerUserID, err := studioService.ReviewDeleteRun(c.Param("id"))
		if err != nil {
			imageStudioError(c, err)
			return
		}
		recordManageAuditFor(c, ownerUserID, "image_studio.review_delete", map[string]interface{}{"run_id": c.Param("id")})
		common.ApiSuccess(c, nil)
		return
	}
	if err := studioService.DeleteRun(c.GetInt("id"), c.Param("id"), c.Query("linked_assets")); err != nil {
		imageStudioError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func SaveImageStudioResult(c *gin.Context) {
	var input struct {
		GroupID string `json:"group_id"`
	}
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_group", Err: errors.New("select an asset group")})
		return
	}
	assetID, err := studioService.SaveResult(c.Request.Context(), c.GetInt("id"), c.Param("id"), input.GroupID)
	if err != nil {
		imageStudioError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"result_id": c.Param("id"), "asset_id": assetID})
}

func DeleteImageStudioResult(c *gin.Context) {
	if err := studioService.DeleteResult(c.GetInt("id"), c.Param("id"), c.Query("linked_assets")); err != nil {
		imageStudioError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// RelayImageStudio replays a studio call through the public relay chain
// (TokenAuth, rate limit, distribution, Relay) so authorization, channel
// selection and billing stay identical to the public API. When snapshot is
// set, the relay request ID is linked to the run and then the client request
// and the final upstream body are archived before the upstream call; a link
// or archive failure aborts it.
func RelayImageStudio(execution studioService.RelayContext, token *model.Token, path string, body []byte, format types.RelayFormat, snapshot *studioService.RequestSnapshot) studioService.RelayResult {
	startedAt := time.Now()
	recorder := httptest.NewRecorder()
	request, err := http.NewRequestWithContext(execution.Context, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return studioService.RelayResult{StatusCode: http.StatusInternalServerError, Body: []byte(err.Error())}
	}
	request.Header.Set("Content-Type", gin.MIMEJSON)
	request.Header.Set("Authorization", "Bearer sk-"+token.Key)
	if clientIP := net.ParseIP(execution.ClientIP); clientIP != nil {
		request.RemoteAddr = net.JoinHostPort(clientIP.String(), "0")
	}
	result := studioService.RelayResult{}
	engine := gin.New()
	engine.Use(middleware.RequestId(), middleware.BodyStorageCleanup())
	engine.POST(path,
		func(internal *gin.Context) {
			if execution.Language != "" {
				common.SetContextKey(internal, constant.ContextKeyLanguage, execution.Language)
			}
			if snapshot != nil {
				common.SetContextKey(internal, constant.ContextKeyUpstreamRequestRecorder, relaycommon.UpstreamRequestRecorder(func(c *gin.Context, upstream []byte) error {
					requestID := c.GetString(common.RequestIdKey)
					if err := model.LinkImageStudioRunRequest(snapshot.RunID, requestID, c.GetInt("channel_id")); err != nil {
						return err
					}
					if err := model.SaveTaskRequestSnapshots("", requestID, snapshot.Original, upstream); err != nil {
						return err
					}
					common.SetContextKey(c, constant.ContextKeyRequestSnapshotStored, true)
					return nil
				}))
			}
			result.RequestID = internal.GetString(common.RequestIdKey)
			internal.Next()
			result.ChannelID = internal.GetInt("channel_id")
		},
		middleware.TokenAuth(),
		middleware.ModelRequestRateLimit(),
		middleware.Distribute(),
		func(internal *gin.Context) {
			Relay(internal, format)
		},
	)
	engine.ServeHTTP(recorder, request)
	result.StatusCode = recorder.Code
	result.Body = recorder.Body.Bytes()
	result.DurationMS = time.Since(startedAt).Milliseconds()
	return result
}

// ImageStudioEnabled blocks the studio API while the feature is disabled.
// Runs and stored images are kept and become available again on re-enable.
func ImageStudioEnabled() gin.HandlerFunc {
	return func(c *gin.Context) {
		if operation_setting.GetImageStudioSetting().Enabled {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"success": false, "code": "image_studio_disabled", "message": "image generation is disabled"})
	}
}

func imageStudioRelayContext(c *gin.Context) studioService.RelayContext {
	return studioService.RelayContext{Context: c.Request.Context(), ClientIP: c.ClientIP(), Language: common.GetContextKeyString(c, constant.ContextKeyLanguage)}
}

func canReviewImageStudio(c *gin.Context) bool {
	if c.GetInt("role") >= common.RoleAdminUser && authz.Can(c.GetInt("id"), c.GetInt("role"), authz.ImageStudioRead) {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"success": false, "code": "permission_denied", "message": "image review permission is required"})
	return false
}

func imageStudioError(c *gin.Context, err error) {
	c.JSON(studioService.StatusCode(err), gin.H{"success": false, "code": studioService.ErrorCode(err), "message": studioService.CleanErrorMessage(err.Error())})
}
