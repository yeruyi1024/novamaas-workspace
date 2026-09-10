// Package volcnative handles Fire Ark's native asynchronous content-generation
// API. It preserves the submitted JSON bytes except for an optional top-level
// model mapping, so provider-specific fields are not discarded by the OpenAI
// task conversion path.
package volcnative

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relay/channel/task/doubao"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	storageService "github.com/QuantumNous/new-api/service/storage"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var ModelList = []string{
	"doubao-seedance-1-0-pro-250528",
	"doubao-seedance-1-0-lite-t2v",
	"doubao-seedance-1-0-lite-i2v",
	"doubao-seedance-1-5-pro-251215",
	"doubao-seedance-2-0-260128",
	"doubao-seedance-2-0-fast-260128",
}

type TaskAdaptor struct {
	doubao.TaskAdaptor
}

type stagedRequestBody struct {
	body           []byte
	convertedCount int
}

func (a *TaskAdaptor) GetChannelName() string { return "volc-native-task" }

func (a *TaskAdaptor) GetModelList() []string { return ModelList }

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.RelayInfo) *dto.TaskError {
	body, err := rawRequestBody(c)
	if err != nil {
		return taskError(err, "invalid_request", http.StatusBadRequest)
	}
	model := gjson.GetBytes(body, "model")
	if !model.Exists() || model.Type != gjson.String || strings.TrimSpace(model.String()) == "" {
		return taskError(fmt.Errorf("model is required"), "invalid_request", http.StatusBadRequest)
	}
	content := gjson.GetBytes(body, "content")
	if !content.IsArray() || len(content.Array()) == 0 {
		return taskError(fmt.Errorf("content must be a non-empty array"), "invalid_request", http.StatusBadRequest)
	}
	if duration := gjson.GetBytes(body, "duration"); duration.Exists() {
		value := duration.Float()
		// -1 is the provider's adaptive-duration sentinel, not a billing multiplier.
		if duration.Type != gjson.Number || math.Trunc(value) != value || value < -1 || value > relaycommon.MaxTaskDurationSeconds {
			return taskError(fmt.Errorf("duration is outside the supported bounds"), "invalid_request", http.StatusBadRequest)
		}
	}
	info.OriginModelName = model.String()
	if info.ChannelMeta != nil && len(info.ParamOverride) != 0 {
		return taskError(fmt.Errorf("Volc Native does not support parameter overrides"), "invalid_request", http.StatusBadRequest)
	}
	var metadata map[string]interface{}
	if err := common.Unmarshal(body, &metadata); err != nil {
		return taskError(err, "invalid_request", http.StatusBadRequest)
	}
	// Only the billing context is decoded. BuildRequestBody still sends raw bytes.
	c.Set("task_request", relaycommon.TaskSubmitReq{Model: model.String(), Metadata: metadata})
	info.Action = constant.TaskActionGenerate
	return nil
}

// BuildRequestBody preserves the native JSON body and applies the channel model
// mapping only to its top-level model field.
func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.RelayInfo) (io.Reader, error) {
	if info != nil && info.ChannelMeta != nil && len(info.ParamOverride) != 0 {
		return nil, fmt.Errorf("volc native channels do not support parameter overrides")
	}
	body, err := rawRequestBody(c)
	if err != nil {
		return nil, err
	}
	if info.ChannelOtherSettings.IsBase64StagingEnabled(info.OriginModelName) {
		policyKey := strings.TrimSpace(info.ChannelOtherSettings.Base64Staging.StoragePolicy)
		if policyKey == "" {
			policyKey = model.StoragePolicyRelayMediaTemp
		}
		cacheKey := "volc_native_base64_staging:" + policyKey
		if cached, exists := c.Get(cacheKey); exists {
			if staged, ok := cached.(stagedRequestBody); ok {
				body = staged.body
				common.SetContextKey(c, constant.ContextKeyTemporaryMediaConvertedCount, staged.convertedCount)
				common.SetContextKey(c, constant.ContextKeyTemporaryMediaConverted, staged.convertedCount > 0)
			}
		} else {
			var convertedCount int
			body, convertedCount, err = storageService.MaterializeVideoTaskBase64(
				c.Request.Context(),
				body,
				info.UserId,
				info.RequestId,
				info.PublicTaskID,
				policyKey,
				storageService.Base64StagingSourceVolcNative,
			)
			if err != nil {
				return nil, err
			}
			c.Set(cacheKey, stagedRequestBody{body: body, convertedCount: convertedCount})
			common.SetContextKey(c, constant.ContextKeyTemporaryMediaConvertedCount, convertedCount)
			common.SetContextKey(c, constant.ContextKeyTemporaryMediaConverted, convertedCount > 0)
		}
	}
	body, err = helper.ApplyModelMappingToJSONBody(info, body)
	if err != nil {
		return nil, err
	}
	common.SetContextKey(c, constant.ContextKeyVideoTaskUpstreamRequestBody, string(body))
	return bytes.NewReader(body), nil
}

// DoResponse replaces the upstream task id with NewAPI's public id and restores
// the public model alias after a mapping. The upstream id remains private in
// Task.PrivateData and is never returned to the caller; all other response
// fields are forwarded as received.
func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (string, []byte, *dto.TaskError) {
	if resp == nil || resp.Body == nil {
		return "", nil, taskError(fmt.Errorf("upstream response is empty"), "invalid_response", http.StatusBadGateway)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, taskError(err, "read_response_body_failed", http.StatusInternalServerError)
	}

	upstreamID := gjson.GetBytes(body, "id")
	if !upstreamID.Exists() || upstreamID.Type != gjson.String || upstreamID.String() == "" {
		return "", nil, taskError(fmt.Errorf("upstream task id is empty"), "invalid_response", http.StatusBadGateway)
	}
	clientBody, err := sjson.SetBytes(body, "id", info.PublicTaskID)
	if err != nil {
		return "", nil, taskError(err, "patch_response_failed", http.StatusInternalServerError)
	}
	if info.ChannelMeta != nil && info.IsModelMapped {
		clientBody, err = helper.RestoreOriginalModelInJSONBody(clientBody, info.OriginModelName)
		if err != nil {
			return "", nil, taskError(err, "response_model_restore_failed", http.StatusInternalServerError)
		}
	}
	c.Data(resp.StatusCode, responseContentType(resp), clientBody)
	return upstreamID.String(), body, nil
}

func rawRequestBody(c *gin.Context) ([]byte, error) {
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

func responseContentType(resp *http.Response) string {
	if value := resp.Header.Get("Content-Type"); value != "" {
		return value
	}
	return "application/json"
}

func taskError(err error, code string, statusCode int) *dto.TaskError {
	return &dto.TaskError{Error: err, Code: code, Message: err.Error(), StatusCode: statusCode, LocalError: true}
}

func (a *TaskAdaptor) ParseTaskResult(body []byte) (*relaycommon.TaskInfo, error) {
	result, err := a.TaskAdaptor.ParseTaskResult(body)
	if err != nil {
		return nil, err
	}
	status := gjson.GetBytes(body, "status").String()
	if status == "cancelled" || status == "expired" {
		result.Status = string(model.TaskStatusFailure)
		result.Progress = "100%"
		result.Reason = status
	}
	return result, nil
}
