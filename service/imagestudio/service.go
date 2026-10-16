// Package imagestudio implements the PC image generation studio. Every model
// call is replayed through the public relay chain, so authorization, channel
// selection and billing are identical to the public API.
package imagestudio

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/types"
	baseService "github.com/QuantumNous/new-api/service"

	"gorm.io/gorm"
)

type RequestError struct {
	StatusCode int
	Code       string
	Err        error
}

func (err *RequestError) Error() string { return err.Err.Error() }
func (err *RequestError) Unwrap() error { return err.Err }

type RelayContext struct {
	Context  context.Context
	ClientIP string
	Language string
}

type RelayResult struct {
	StatusCode int
	Body       []byte
	RequestID  string
	DurationMS int64
}

// RequestSnapshot asks the relay to archive the client request next to the
// final upstream body. The relay request ID is linked to RunID before the
// archive is written, so snapshot readers can always resolve the owner run.
type RequestSnapshot struct {
	RunID    int64
	Original []byte
}

// RelayExecutor sends body to path through the public relay chain with token.
// A nil snapshot disables request snapshots for that call.
type RelayExecutor func(execution RelayContext, token *model.Token, path string, body []byte, format types.RelayFormat, snapshot *RequestSnapshot) RelayResult

func resolveToken(userID, tokenID int) (*model.Token, error) {
	token, err := model.GetTokenByIds(tokenID, userID)
	if err != nil {
		return nil, &RequestError{http.StatusBadRequest, "token_unavailable", errors.New("the selected API key is unavailable")}
	}
	valid, err := model.ValidateUserToken(token.Key)
	if err != nil || valid == nil || valid.Id != token.Id {
		return nil, &RequestError{http.StatusBadRequest, "token_unavailable", errors.New("the selected API key is disabled, expired, or out of quota")}
	}
	return valid, nil
}

// ModelsForToken lists the models the token may call on endpointType.
func ModelsForToken(userID, tokenID int, endpointType constant.EndpointType) ([]string, error) {
	token, err := resolveToken(userID, tokenID)
	if err != nil {
		return nil, err
	}
	return modelsForToken(userID, token, endpointType)
}

func modelsForToken(userID int, token *model.Token, endpointType constant.EndpointType) ([]string, error) {
	user, err := model.GetUserCache(userID)
	if err != nil {
		return nil, err
	}
	group := strings.TrimSpace(token.Group)
	if group == "" {
		group = user.Group
	}
	groups := []string{group}
	if group == "auto" {
		configured, parseErr := token.GetAutoGroups()
		if parseErr != nil {
			return nil, parseErr
		}
		groups = baseService.GetUserAutoGroup(user.Group)
		if token.AutoGroups != "" {
			groups = baseService.FilterUserTokenAutoGroups(user.Group, configured)
		}
	}
	model.GetPricing()
	limits := token.GetModelLimitsMap()
	result := make([]string, 0)
	for _, name := range baseService.GetGroupsEnabledModels(groups) {
		if token.ModelLimitsEnabled && !limits[name] {
			continue
		}
		if endpointType == constant.EndpointTypeOpenAI && common.IsImageGenerationModel(name) {
			continue
		}
		if slices.Contains(model.GetModelSupportEndpointTypes(name), endpointType) {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result, nil
}

func validateModel(userID int, token *model.Token, name string, endpointType constant.EndpointType) error {
	models, err := modelsForToken(userID, token, endpointType)
	if err != nil {
		return err
	}
	if !slices.Contains(models, name) {
		return &RequestError{http.StatusForbidden, "model_unavailable", fmt.Errorf("model %s is unavailable for this API key", name)}
	}
	return nil
}

// quotaForRequest reads the amount the relay charged from its consume log.
func quotaForRequest(requestID string, userID int) int {
	if requestID == "" || model.LOG_DB == nil {
		return 0
	}
	var log model.Log
	if model.LOG_DB.Where("request_id = ? AND user_id = ? AND type = ?", requestID, userID, model.LogTypeConsume).
		Order("id desc").First(&log).Error != nil {
		return 0
	}
	return log.Quota
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
		if value, ok := payload.Error.(map[string]any); ok {
			if message, ok := value["message"].(string); ok {
				return CleanErrorMessage(message)
			}
		}
		if payload.Error != nil {
			return CleanErrorMessage(fmt.Sprint(payload.Error))
		}
	}
	if len(body) > 0 {
		return CleanErrorMessage(string(body))
	}
	return "model provider request failed"
}

var (
	embeddedURLPattern      = regexp.MustCompile(`https?://[^\s"'<>]+`)
	bearerCredentialPattern = regexp.MustCompile(`(?i)(bearer\s+)[^\s"'<>]+`)
	openAIKeyPattern        = regexp.MustCompile(`\bsk-[A-Za-z0-9_-]+`)
)

// CleanErrorMessage strips URLs and credentials before an error is stored or shown.
func CleanErrorMessage(message string) string {
	message = embeddedURLPattern.ReplaceAllStringFunc(strings.TrimSpace(message), relaycommon.SanitizeURLForLog)
	message = bearerCredentialPattern.ReplaceAllString(message, "${1}[redacted]")
	message = openAIKeyPattern.ReplaceAllString(message, "sk-[redacted]")
	characters := []rune(message)
	return string(characters[:min(len(characters), 500)])
}

func StatusCode(err error) int {
	var requestErr *RequestError
	if errors.As(err, &requestErr) {
		return requestErr.StatusCode
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

func ErrorCode(err error) string {
	var requestErr *RequestError
	if errors.As(err, &requestErr) {
		return requestErr.Code
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "not_found"
	}
	return "image_studio_error"
}
