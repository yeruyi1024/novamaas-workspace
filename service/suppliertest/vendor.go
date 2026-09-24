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
	requireKVV        bool
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
	case VendorKimi:
		return vendorProfile{
			id:                vendor,
			cachePrefixRole:   "system",
			requireUsage:      true,
			requireCacheField: true,
			requireJSON:       true,
			requireTools:      true,
			requireKVV:        true,
		}
	case VendorGLM, VendorDeepSeek:
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

func modelKey(model string) string {
	model = strings.ToLower(strings.TrimSpace(model))
	replacer := strings.NewReplacer("-", "", "_", "", ".", "", "/", "")
	return replacer.Replace(model)
}

func thinkingRequired(vendor, model string) bool {
	model = modelKey(model)
	switch vendor {
	case VendorGLM:
		return strings.Contains(model, "thinking") ||
			strings.Contains(model, "glm53") ||
			strings.Contains(model, "glm52") ||
			strings.Contains(model, "glm51") ||
			strings.Contains(model, "glm50") ||
			strings.Contains(model, "glm5") ||
			strings.Contains(model, "glm47") ||
			strings.Contains(model, "glm46") ||
			strings.Contains(model, "glm45")
	case VendorKimi:
		return strings.Contains(model, "kimik3") ||
			strings.Contains(model, "kimik27") ||
			strings.Contains(model, "kimik26") ||
			strings.Contains(model, "thinking")
	case VendorDeepSeek:
		return strings.Contains(model, "deepseekflash") ||
			strings.Contains(model, "deepseekv4") ||
			strings.Contains(model, "reasoner") ||
			strings.Contains(model, "deepseekr1") ||
			strings.HasPrefix(model, "r1")
	default:
		return false
	}
}

func applyThinking(req chatRequest, vendor string) chatRequest {
	req.Messages = []chatMessage{{Role: "user", Content: "What is 17 times 19? Think step by step."}}
	req.Thinking = nil
	req.ReasoningEffort = ""
	model := modelKey(req.Model)
	switch vendor {
	case VendorDeepSeek:
		// DeepSeek V4 的 OpenAI Chat Completions 协议支持显式 thinking 和 reasoning_effort。
		// 旧版 reasoner/R1 本身就是思考模型，额外参数在部分兼容端点会被拒绝。
		if strings.Contains(model, "deepseekflash") || strings.Contains(model, "deepseekv4") {
			req.Thinking = map[string]any{"type": "enabled"}
			req.ReasoningEffort = "low"
		}
		return req
	case VendorKimi:
		// K3 只接受 reasoning_effort；K2.7 不应传思考控制参数；K2.6 接受 thinking.type。
		switch {
		case strings.Contains(model, "kimik3"):
			req.ReasoningEffort = "low"
		case strings.Contains(model, "kimik26"):
			req.Thinking = map[string]any{"type": "enabled"}
		}
		return req
	case VendorGLM:
		// GLM 思考协议；5.3 系列支持 reasoning_effort，测试使用 low 控制耗时。
		req.Thinking = map[string]any{"type": "enabled"}
		if strings.Contains(model, "glm53") {
			req.ReasoningEffort = "low"
		}
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
