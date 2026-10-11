package imagestudio

import (
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

func QuotaForRequest(requestID string, userID int) int {
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

var (
	embeddedURLPattern      = regexp.MustCompile(`https?://[^\s"'<>]+`)
	bearerCredentialPattern = regexp.MustCompile(`(?i)(bearer\s+)[^\s"'<>]+`)
	keyValueSecretPattern   = regexp.MustCompile(`(?i)((?:api[_-]?key|access[_-]?key|secret|token|authorization)\s*(?::|=)\s*["']?)[^"',\s}\]<>]+`)
	openAIKeyPattern        = regexp.MustCompile(`\bsk-[A-Za-z0-9_-]+`)
)

func CleanErrorMessage(message string) string {
	message = embeddedURLPattern.ReplaceAllStringFunc(strings.TrimSpace(message), relaycommon.SanitizeURLForLog)
	message = bearerCredentialPattern.ReplaceAllString(message, "${1}[redacted]")
	message = keyValueSecretPattern.ReplaceAllString(message, "${1}[redacted]")
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
	return "image_studio_error"
}
