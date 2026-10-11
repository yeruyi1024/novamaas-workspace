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
	"github.com/QuantumNous/new-api/relaykit/types"
	studioService "github.com/QuantumNous/new-api/service/imagestudio"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

func GetImageStudioSettings(c *gin.Context) {
	settings := operation_setting.GetImageStudioSetting()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"enabled": settings.Enabled, "visible": settings.Visible, "asset_library_enabled": settings.AssetLibraryEnabled,
		"max_images":            operation_setting.DefaultImageStudioMaxImages,
		"max_reference_images":  operation_setting.DefaultImageStudioMaxReferenceImages,
		"max_prompt_characters": operation_setting.DefaultImageStudioMaxPromptCharacters,
	}})
}

func ListImageStudioModels(c *gin.Context) {
	tokenID, _ := strconv.Atoi(c.Query("token_id"))
	endpointType := constant.EndpointTypeImageGeneration
	if kind := c.Query("kind"); kind == "text" {
		endpointType = constant.EndpointTypeOpenAI
	} else if kind != "" && kind != "image" {
		imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_model_kind", Err: errors.New("model kind must be image or text")})
		return
	}
	models, err := studioService.ModelsForToken(c.GetInt("id"), tokenID, endpointType)
	if err != nil {
		imageStudioError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": models})
}

func PromptAssistImageStudio(c *gin.Context) {
	var input studioService.PromptAssistInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_request", Err: errors.New("invalid prompt optimization request")})
		return
	}
	execution := studioService.RelayContext{Context: c.Request.Context(), ClientIP: c.ClientIP(), Language: common.GetContextKeyString(c, constant.ContextKeyLanguage)}
	result, err := studioService.AssistPrompt(execution, c.GetInt("id"), input, RelayImageStudio)
	if err != nil {
		imageStudioError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func CreateImageStudioRun(c *gin.Context) {
	body, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, model.MaxTaskRequestBodyBytes))
	if err != nil {
		status := http.StatusBadRequest
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		imageStudioError(c, &studioService.RequestError{StatusCode: status, Code: "invalid_request", Err: errors.New("image request is invalid or exceeds the 2 MiB limit")})
		return
	}
	var input studioService.RunInput
	if err := common.Unmarshal(body, &input); err != nil {
		imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_request", Err: errors.New("invalid image studio run request")})
		return
	}
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if header := strings.TrimSpace(c.GetHeader("Idempotency-Key")); header != "" {
		if input.IdempotencyKey != "" && input.IdempotencyKey != header {
			imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "idempotency_key_mismatch", Err: errors.New("body and header idempotency keys do not match")})
			return
		}
		input.IdempotencyKey = header
	}
	execution := studioService.RelayContext{Context: c.Request.Context(), ClientIP: c.ClientIP(), Language: common.GetContextKeyString(c, constant.ContextKeyLanguage), OriginalRequestBody: body}
	view, replay, err := studioService.CreateRun(execution, c.GetInt("id"), input, RelayImageStudio)
	if err != nil {
		imageStudioError(c, err)
		return
	}
	if replay {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": view, "idempotent_replay": true})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": view})
}

func ListImageStudioRuns(c *gin.Context) {
	ownerUserID := 0
	if value := c.Query("user_id"); value != "" {
		var err error
		ownerUserID, err = strconv.Atoi(value)
		if err != nil || ownerUserID < 0 {
			imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_user_id", Err: errors.New("invalid user filter")})
			return
		}
	}
	actor := studioService.Actor{UserID: c.GetInt("id"), Role: c.GetInt("role")}
	result, audit, err := studioService.ListRuns(c.Request.Context(), actor, c.Query("scope") == "all", ownerUserID, common.GetPageQuery(c))
	if err != nil {
		imageStudioError(c, err)
		return
	}
	if audit {
		recordManageAudit(c, "image_studio.content_read", map[string]interface{}{"scope": "all", "owner_user_id": ownerUserID, "count": len(result.Items)})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func GetImageStudioRun(c *gin.Context) {
	actor := studioService.Actor{UserID: c.GetInt("id"), Role: c.GetInt("role")}
	view, audit, err := studioService.GetRun(c.Request.Context(), actor, c.Param("id"), c.Query("scope") == "all")
	if err != nil {
		imageStudioError(c, err)
		return
	}
	if audit {
		recordManageAudit(c, "image_studio.content_read", map[string]interface{}{"run_id": view.ID, "owner_user_id": view.UserID})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": view})
}

func SaveImageStudioResult(c *gin.Context) {
	var input struct {
		GroupID string `json:"group_id"`
	}
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		imageStudioError(c, &studioService.RequestError{StatusCode: http.StatusBadRequest, Code: "invalid_group", Err: errors.New("select an asset group")})
		return
	}
	if err := studioService.SaveResult(c.Request.Context(), c.GetInt("id"), c.Param("id"), input.GroupID); err != nil {
		imageStudioError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"result_id": c.Param("id")}})
}

func DeleteImageStudioResult(c *gin.Context) {
	if err := studioService.DeleteResult(c.GetInt("id"), c.Param("id")); err != nil {
		imageStudioError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func RelayImageStudio(execution studioService.RelayContext, token *model.Token, path string, body []byte, format types.RelayFormat) studioService.RelayResult {
	startedAt := time.Now()
	recorder := httptest.NewRecorder()
	request, err := http.NewRequestWithContext(execution.Context, http.MethodPost, path, bytes.NewReader(body))
	if err != nil {
		return studioService.RelayResult{StatusCode: http.StatusInternalServerError, Body: []byte(err.Error())}
	}
	request.Header.Set("Content-Type", gin.MIMEJSON)
	request.Header.Set("Authorization", "Bearer sk-"+token.Key)
	clientIP := net.ParseIP(execution.ClientIP)
	if clientIP != nil {
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
			if format == types.RelayFormatOpenAIImage {
				original := execution.OriginalRequestBody
				if len(original) == 0 {
					original = body
				}
				common.SetContextKey(internal, constant.ContextKeyImageStudioOriginalRequestBody, string(original))
			}
			result.RequestID = internal.GetString(common.RequestIdKey)
			internal.Next()
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

func imageStudioError(c *gin.Context, err error) {
	status := studioService.StatusCode(err)
	message := studioService.CleanErrorMessage(err.Error())
	c.JSON(status, gin.H{"success": false, "code": studioService.ErrorCode(err), "message": message})
}
