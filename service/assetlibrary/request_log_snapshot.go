package assetlibrary

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const requestLogBodyLimit = 16 * 1024
const requestLogURLLimit = 8 * 1024

var requestLogURLPattern = regexp.MustCompile(`https?://[^\s"'<>]+`)
var requestLogBearerPattern = regexp.MustCompile(`(?i)\bBearer\s+[^\s"']+`)
var requestLogSecretAssignmentPattern = regexp.MustCompile(`(?i)\b(token|api[_-]?key|secret|signature|authorization|password|credential|access[_-]?key)\s*[=:]\s*[^\s&,;"']+`)

func requestLogSecretKey(key string) bool {
	key = strings.ToLower(strings.NewReplacer("-", "", "_", "", " ", "").Replace(key))
	for _, part := range []string{"token", "secret", "signature", "credential", "password", "authorization", "apikey", "accesskey", "privatekey", "xamz", "xoss", "xtos", "xgoog", "policy"} {
		if strings.Contains(key, part) {
			return true
		}
	}
	return key == "key" || key == "sig" || key == "auth"
}

func sanitizeRequestLogURL(raw string) string {
	if raw == "" {
		return ""
	}
	if len(raw) > requestLogURLLimit {
		return "[address omitted: exceeds 8 KiB]"
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "[invalid address omitted]"
	}
	parsed.User = nil
	query := parsed.Query()
	for key := range query {
		if requestLogSecretKey(key) {
			query.Set(key, "[REDACTED]")
		}
	}
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	return parsed.String()
}

func sanitizeRequestLogString(value string) string {
	value = requestLogURLPattern.ReplaceAllStringFunc(value, sanitizeRequestLogURL)
	value = requestLogBearerPattern.ReplaceAllString(value, "Bearer [REDACTED]")
	return requestLogSecretAssignmentPattern.ReplaceAllStringFunc(value, func(match string) string {
		position := strings.IndexAny(match, "=:")
		if position < 0 {
			return "[REDACTED]"
		}
		return match[:position+1] + "[REDACTED]"
	})
}

func sanitizeRequestLogValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			if requestLogSecretKey(key) {
				typed[key] = "[REDACTED]"
			} else {
				typed[key] = sanitizeRequestLogValue(nested)
			}
		}
		return typed
	case []any:
		for i, nested := range typed {
			typed[i] = sanitizeRequestLogValue(nested)
		}
		return typed
	case string:
		return sanitizeRequestLogString(typed)
	default:
		return value
	}
}

func sanitizeRequestLogBody(body []byte) (string, bool) {
	if len(body) == 0 {
		return "", false
	}
	if len(body) > requestLogBodyLimit {
		return "", true
	}
	var parsed any
	if err := common.Unmarshal(body, &parsed); err == nil {
		encoded, err := common.Marshal(sanitizeRequestLogValue(parsed))
		if err != nil || len(encoded) > requestLogBodyLimit {
			return "", true
		}
		return string(encoded), false
	}
	if !utf8.Valid(body) {
		return "", true
	}
	return sanitizeRequestLogString(string(body)), false
}

func requestLogDetail(requestURL string, requestBody []byte, responseBody []byte) model.AssetRequestLogDetail {
	requestText, requestOmitted := sanitizeRequestLogBody(requestBody)
	responseText, responseOmitted := sanitizeRequestLogBody(responseBody)
	return model.AssetRequestLogDetail{
		RequestURL:  sanitizeRequestLogURL(requestURL),
		RequestBody: requestText, ResponseBody: responseText,
		RequestBodyOmitted: requestOmitted, ResponseBodyOmitted: responseOmitted,
	}
}
