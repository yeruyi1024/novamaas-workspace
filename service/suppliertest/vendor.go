package suppliertest

import (
	"fmt"
	"strings"
)

const (
	VendorGeneric  = "generic"
	VendorGLM      = "glm"
	VendorKimi     = "kimi"
	VendorDeepSeek = "deepseek"

	DefaultCacheWarmUser = "请用一个词回复：ping"
)

type vendorProfile struct {
	id                string
	cachePrefixRole   string
	requireUsage      bool
	requireCacheField bool
	requireJSON       bool
	requireTools      bool
}

func ResolveVendor(requested string) (string, error) {
	requested = strings.ToLower(strings.TrimSpace(requested))
	if requested == "" {
		return VendorGeneric, nil
	}
	switch requested {
	case VendorGeneric, VendorGLM, VendorKimi, VendorDeepSeek:
		return requested, nil
	default:
		return "", fmt.Errorf("unknown vendor %q", requested)
	}
}

func profileFor(vendor string) vendorProfile {
	switch vendor {
	case VendorGLM, VendorKimi, VendorDeepSeek:
		return vendorProfile{
			id:                vendor,
			cachePrefixRole:   "system",
			requireUsage:      true,
			requireCacheField: true,
			requireJSON:       true,
			requireTools:      true,
		}
	default:
		return vendorProfile{
			id:              VendorGeneric,
			cachePrefixRole: "system",
		}
	}
}

func vendorTitle(vendor string) string {
	switch vendor {
	case VendorGLM:
		return "GLM"
	case VendorKimi:
		return "Kimi"
	case VendorDeepSeek:
		return "DeepSeek"
	default:
		return "供应商"
	}
}

func thinkingRequired(vendor string) bool {
	switch vendor {
	case VendorGLM, VendorKimi, VendorDeepSeek:
		return true
	default:
		return false
	}
}

func applyThinking(req chatRequest, vendor string) chatRequest {
	req.Messages = []chatMessage{{Role: "user", Content: "What is 17 times 19? Think step by step."}}
	switch vendor {
	case VendorDeepSeek:
		// DeepSeek (如 DeepSeek-V3/V4, R1 等) 原生自带推理，严禁传 thinking 参数 (传了会报 400)
		req.Thinking = nil
		return req
	case VendorKimi:
		// Kimi 新一代模型 (如 K3, K2 等) 遵循 reasoning_effort 规范
		req.Thinking = nil
		req.ReasoningEffort = "low"
		return req
	case VendorGLM:
		// 智谱 GLM 思考协议
		req.Thinking = map[string]any{"type": "enabled"}
		return req
	default:
		req.Thinking = map[string]any{"type": "enabled"}
		return req
	}
}

func cacheMessages(profile vendorProfile, prefix, followUp string, warm bool) []chatMessage {
	if profile.cachePrefixRole == "system" {
		user := followUp
		if warm {
			user = DefaultCacheWarmUser
		}
		return []chatMessage{
			{Role: "system", Content: prefix},
			{Role: "user", Content: user},
		}
	}
	if warm {
		return []chatMessage{{Role: "user", Content: prefix}}
	}
	return []chatMessage{
		{Role: "user", Content: prefix},
		{Role: "assistant", Content: "ok"},
		{Role: "user", Content: followUp},
	}
}

func checkStatus(required bool, skipMsg, failMsg string) (string, string) {
	if required {
		return "fail", failMsg
	}
	return "skip", skipMsg
}
