package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const volcNativeCancelReason = "cancelled"

var volcNativeTaskPlatform = constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeVolcNative))

// RelayVolcNativeImage forwards Fire Ark native image requests without parsing
// or rebuilding their JSON body. The only accepted channel type is Volc Native;
// the existing VolcEngine channel remains responsible for OpenAI-compatible
// image requests.
func RelayVolcNativeImage(c *gin.Context) {
	info, err := relaycommon.GenRelayInfo(c, types.RelayFormatTask, nil, nil)
	if err != nil {
		respondVolcNativeError(c, http.StatusInternalServerError, "gen_relay_info_failed", "failed to initialise relay")
		return
	}
	info.InitChannelMeta(c)
	if info.ChannelType != constant.ChannelTypeVolcNative {
		respondVolcNativeError(c, http.StatusBadRequest, "invalid_channel_type", "this endpoint requires a Volc Native channel")
		return
	}

	body, err := volcNativeRequestBody(c)
	if err != nil {
		respondVolcNativeError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	modelName := gjson.GetBytes(body, "model")
	if !modelName.Exists() || modelName.Type != gjson.String || strings.TrimSpace(modelName.String()) == "" {
		respondVolcNativeError(c, http.StatusBadRequest, "invalid_request", "model is required")
		return
	}
	info.OriginModelName = modelName.String()
	info.Action = "volc_native_image_generation"

	if err := helper.ModelMappedHelper(c, info, nil); err != nil {
		respondVolcNativeError(c, http.StatusBadRequest, "model_mapping_failed", err.Error())
		return
	}
	if info.IsModelMapped {
		respondVolcNativeError(c, http.StatusBadRequest, "model_mapping_not_supported", "Volc Native channels require the upstream model id directly")
		return
	}
	if len(info.ParamOverride) != 0 {
		respondVolcNativeError(c, http.StatusBadRequest, "param_override_not_supported", "Volc Native channels do not support parameter overrides")
		return
	}

	priceData, err := helper.ModelPriceHelperPerCall(c, info)
	if err != nil {
		respondVolcNativeError(c, http.StatusBadRequest, "model_price_error", err.Error())
		return
	}
	info.PriceData = priceData
	if !priceData.FreeModel {
		info.ForcePreConsume = true
		if billingErr := service.PreConsumeBilling(c, priceData.Quota, info); billingErr != nil {
			respondVolcNativeError(c, billingErr.StatusCode, string(billingErr.GetErrorCode()), billingErr.MaskSensitiveError())
			return
		}
	}

	succeeded := false
	defer func() {
		if !succeeded && info.Billing != nil {
			info.Billing.Refund(c)
		}
	}()

	baseURL := strings.TrimRight(info.ChannelBaseUrl, "/")
	if baseURL == "" {
		baseURL = constant.ChannelBaseURLs[constant.ChannelTypeVolcNative]
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, baseURL+"/api/v3/images/generations", bytes.NewReader(body))
	if err != nil {
		respondVolcNativeError(c, http.StatusInternalServerError, "build_request_failed", "failed to build upstream request")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+info.ApiKey)

	client, err := service.GetHttpClientWithProxy(info.ChannelSetting.Proxy)
	if err != nil {
		respondVolcNativeError(c, http.StatusInternalServerError, "http_client_failed", "failed to initialise upstream client")
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		respondVolcNativeError(c, http.StatusBadGateway, "upstream_request_failed", "upstream request failed")
		return
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		respondVolcNativeError(c, http.StatusBadGateway, "read_response_failed", "failed to read upstream response")
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		c.Data(resp.StatusCode, contentType, responseBody)
		return
	}

	if err := service.SettleBilling(c, info, priceData.Quota); err != nil {
		respondVolcNativeError(c, http.StatusInternalServerError, "billing_settlement_failed", "failed to settle usage")
		return
	}
	service.LogTaskConsumption(c, info)
	succeeded = true
	c.Data(resp.StatusCode, contentType, responseBody)
}

// RelayVolcNativeTaskFetch returns the caller's own task in Fire Ark's native
// response shape. It never forwards an upstream task id supplied by a client.
func RelayVolcNativeTaskFetch(c *gin.Context) {
	task, ok := getVolcNativeTask(c)
	if !ok {
		return
	}
	c.Data(http.StatusOK, "application/json", buildVolcNativeTaskResponse(task))
}

// RelayVolcNativeTaskList deliberately lists locally owned tasks rather than
// relaying Fire Ark's account-level list endpoint, preventing one user from
// enumerating another user's work on a shared upstream credential.
func RelayVolcNativeTaskList(c *gin.Context) {
	page := parseVolcNativePositiveInt(c.DefaultQuery("page_num", "1"), 1)
	pageSize := parseVolcNativePositiveInt(c.DefaultQuery("page_size", "10"), 10)
	if pageSize > 100 {
		pageSize = 100
	}
	if page > 500 {
		page = 500
	}
	offset := (page - 1) * pageSize
	query := model.DB.WithContext(c.Request.Context()).Model(&model.Task{}).
		Where("user_id = ? AND platform = ?", c.GetInt("id"), volcNativeTaskPlatform).Order("id desc")
	rows, err := query.Rows()
	if err != nil {
		respondVolcNativeError(c, http.StatusInternalServerError, "task_lookup_failed", "failed to list tasks")
		return
	}
	defer rows.Close()
	items := make([]json.RawMessage, 0, pageSize)
	total := 0
	for rows.Next() {
		task := &model.Task{}
		if err := model.DB.ScanRows(rows, task); err != nil {
			respondVolcNativeError(c, http.StatusInternalServerError, "task_lookup_failed", "failed to list tasks")
			return
		}
		if !volcNativeTaskAllowed(c, task) {
			continue
		}
		if filter := c.Query("filter.model"); filter != "" && volcNativeTaskModel(task) != filter {
			continue
		}
		if filter := c.Query("filter.status"); filter != "" && volcNativeTaskStatus(task) != filter {
			continue
		}
		if filter := c.Query("filter.service_tier"); filter != "" && gjson.GetBytes(task.Data, "service_tier").String() != filter {
			continue
		}
		if total >= offset && len(items) < pageSize {
			items = append(items, buildVolcNativeTaskResponse(task))
		}
		total++
	}
	if err := rows.Err(); err != nil {
		respondVolcNativeError(c, http.StatusInternalServerError, "task_lookup_failed", "failed to list tasks")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// RelayVolcNativeTaskDelete cancels only a task owned by the current user. The
// upstream credential stays inside the server process and is never copied into
// a response or log entry.
func RelayVolcNativeTaskDelete(c *gin.Context) {
	task, ok := getVolcNativeTask(c)
	if !ok {
		return
	}
	if isVolcNativeTaskTerminal(task.Status) {
		c.Data(http.StatusOK, "application/json", buildVolcNativeTaskResponse(task))
		return
	}

	channel, err := model.CacheGetChannel(task.ChannelId)
	if err != nil || channel == nil || channel.Type != constant.ChannelTypeVolcNative {
		respondVolcNativeError(c, http.StatusBadRequest, "invalid_task_channel", "task does not belong to a Volc Native channel")
		return
	}
	key := task.PrivateData.Key
	if key == "" {
		selectedKey, _, keyErr := channel.GetNextEnabledKey()
		if keyErr != nil {
			respondVolcNativeError(c, keyErr.StatusCode, "channel_no_available_key", "no upstream credential is available")
			return
		}
		key = selectedKey
	}
	baseURL := strings.TrimRight(channel.GetBaseURL(), "/")
	if baseURL == "" {
		baseURL = constant.ChannelBaseURLs[constant.ChannelTypeVolcNative]
	}
	upstreamURL := baseURL + "/api/v3/contents/generations/tasks/" + url.PathEscape(task.GetUpstreamTaskID())
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodDelete, upstreamURL, nil)
	if err != nil {
		respondVolcNativeError(c, http.StatusInternalServerError, "build_request_failed", "failed to build upstream request")
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	client, err := service.GetHttpClientWithProxy(channel.GetSetting().Proxy)
	if err != nil {
		respondVolcNativeError(c, http.StatusInternalServerError, "http_client_failed", "failed to initialise upstream client")
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		respondVolcNativeError(c, http.StatusBadGateway, "upstream_request_failed", "upstream request failed")
		return
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		respondVolcNativeError(c, http.StatusBadGateway, "read_response_failed", "failed to read upstream response")
		return
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		c.Data(resp.StatusCode, contentType, responseBody)
		return
	}

	previousStatus := task.Status
	task.Status = model.TaskStatusFailure
	task.Progress = "100%"
	task.FailReason = volcNativeCancelReason
	task.FinishTime = time.Now().Unix()
	won, err := task.UpdateWithStatus(previousStatus)
	if err != nil {
		respondVolcNativeError(c, http.StatusInternalServerError, "task_update_failed", "failed to update local task")
		return
	}
	if won {
		service.RefundTaskQuota(c.Request.Context(), task, volcNativeCancelReason)
	} else {
		// A concurrent poll completed the task. Return the durable state and do
		// not perform a second refund.
		refreshed, exists, refreshErr := model.GetByTaskId(c.GetInt("id"), task.TaskID)
		if refreshErr != nil || !exists || refreshed == nil {
			respondVolcNativeError(c, http.StatusInternalServerError, "task_refresh_failed", "failed to refresh local task")
			return
		}
		task = refreshed
	}
	c.Data(http.StatusOK, "application/json", buildVolcNativeTaskResponse(task))
}

func getVolcNativeTask(c *gin.Context) (*model.Task, bool) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" || strings.Contains(taskID, "/") || strings.Contains(taskID, "..") {
		respondVolcNativeError(c, http.StatusBadRequest, "invalid_task_id", "invalid task id")
		return nil, false
	}
	task, exists, err := model.GetByTaskId(c.GetInt("id"), taskID)
	if err != nil {
		respondVolcNativeError(c, http.StatusInternalServerError, "task_lookup_failed", "failed to read task")
		return nil, false
	}
	if !exists || task == nil || task.Platform != volcNativeTaskPlatform {
		respondVolcNativeError(c, http.StatusNotFound, "task_not_found", "task not found")
		return nil, false
	}
	if !volcNativeTaskAllowed(c, task) {
		respondVolcNativeError(c, http.StatusForbidden, "task_access_denied", "the token cannot access this task")
		return nil, false
	}
	return task, true
}

func volcNativeRequestBody(c *gin.Context) ([]byte, error) {
	storage, err := common.GetBodyStorage(c)
	if err != nil {
		return nil, err
	}
	body, err := storage.Bytes()
	if err != nil {
		return nil, err
	}
	if !gjson.ValidBytes(body) {
		return nil, fmt.Errorf("request body must be valid JSON")
	}
	return body, nil
}

func buildVolcNativeTaskResponse(task *model.Task) []byte {
	if status := gjson.GetBytes(task.Data, "status"); status.Exists() {
		if body, err := sjson.SetBytes(task.Data, "id", task.TaskID); err == nil {
			if body, err = sjson.SetBytes(body, "status", volcNativeTaskStatus(task)); err == nil {
				return body
			}
		}
	}
	body, err := common.Marshal(gin.H{
		"id":         task.TaskID,
		"model":      volcNativeTaskModel(task),
		"status":     volcNativeTaskStatus(task),
		"created_at": task.CreatedAt,
		"updated_at": task.UpdatedAt,
	})
	if err != nil {
		return []byte(`{"status":"queued"}`)
	}
	return body
}

func volcNativeTaskModel(task *model.Task) string {
	if task.Properties.OriginModelName != "" {
		return task.Properties.OriginModelName
	}
	return task.Properties.UpstreamModelName
}

func volcNativeTaskStatus(task *model.Task) string {
	switch task.Status {
	case model.TaskStatusSuccess:
		return "succeeded"
	case model.TaskStatusFailure:
		if task.FailReason == volcNativeCancelReason || task.FailReason == "expired" {
			return task.FailReason
		}
		return "failed"
	case model.TaskStatusInProgress:
		return "running"
	default:
		return "queued"
	}
}

func isVolcNativeTaskTerminal(status model.TaskStatus) bool {
	return status == model.TaskStatusSuccess || status == model.TaskStatusFailure
}

func parseVolcNativePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func respondVolcNativeError(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

// Authenticated task operations use stored ownership and token model/channel
// restrictions, without selecting an unrelated channel from an empty GET body.
func volcNativeTaskAllowed(c *gin.Context, task *model.Task) bool {
	if value := common.GetContextKeyString(c, constant.ContextKeyTokenSpecificChannelId); value != "" && value != strconv.Itoa(task.ChannelId) {
		return false
	}
	if !common.GetContextKeyBool(c, constant.ContextKeyTokenModelLimitEnabled) {
		return true
	}
	value, exists := common.GetContextKey(c, constant.ContextKeyTokenModelLimit)
	if !exists {
		return false
	}
	allowed, ok := value.(map[string]bool)
	if !ok {
		return false
	}
	_, exists = allowed[ratio_setting.FormatMatchingModelName(volcNativeTaskModel(task))]
	return exists
}
