package assetlibrary

import (
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/model"
)

type upstreamAssetError struct {
	statusCode int
	code       string
	message    string
}

func (err *upstreamAssetError) Error() string {
	if err.message == "" {
		return err.code
	}
	if err.code == "" {
		return err.message
	}
	return err.code + ": " + err.message
}

// Only recognized content-policy signals are terminal. Unknown failures retain
// the existing retry behavior; transport and authentication failures must not
// make an uploaded asset unusable.
func contentRejectionReason(code string, message string) string {
	text := strings.ToLower(code + " " + message)
	for _, marker := range []string{"timeout", "temporarily unavailable", "service unavailable", "serviceunavailable", "rate limit", "限流", "超时", "服务不可用", "服务异常", "稍后重试"} {
		if strings.Contains(text, marker) {
			return ""
		}
	}
	for _, marker := range []string{"真人", "真实人物", "人脸识别", "活体", "real person", "real-person", "realperson", "realistic person", "human face", "humanface", "face recognition"} {
		if strings.Contains(text, marker) {
			return model.AssetUnavailableRealPerson
		}
	}
	for _, marker := range []string{"敏感", "违规", "涉政", "涉黄", "sensitive content", "sensitivecontent", "sensitive information", "unsafe content", "prohibited content", "content violation"} {
		if strings.Contains(text, marker) {
			return model.AssetUnavailableSensitiveContent
		}
	}
	for _, marker := range []string{"contentrejected", "policyviolation", "moderationrejected", "contentpolicy", "content policy", "审核拒绝", "审核未通过"} {
		if strings.Contains(text, marker) {
			return model.AssetUnavailablePolicyRejected
		}
	}
	return ""
}

func upstreamContentRejectionReason(err error) string {
	var upstreamErr *upstreamAssetError
	if !errors.As(err, &upstreamErr) {
		return ""
	}
	if upstreamErr.statusCode >= 500 || upstreamErr.statusCode == 401 || upstreamErr.statusCode == 408 || upstreamErr.statusCode == 429 {
		return ""
	}
	if upstreamErr.statusCode == 403 {
		text := strings.ToLower(upstreamErr.code + " " + upstreamErr.message)
		for _, marker := range []string{"permission", "accessdenied", "forbidden", "credential", "signature", "authentication", "鉴权", "权限", "签名"} {
			if strings.Contains(text, marker) {
				return ""
			}
		}
	}
	return contentRejectionReason(upstreamErr.code, upstreamErr.message)
}

func processingContentRejectionReason(result providerResult) string {
	if reason := contentRejectionReason(result.Code, result.Message); reason != "" {
		return reason
	}
	if strings.EqualFold(result.Status, "rejected") {
		return model.AssetUnavailablePolicyRejected
	}
	return ""
}
