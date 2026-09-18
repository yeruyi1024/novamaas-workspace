package suppliertest

const (
	ModuleBasic  = "basic"
	ModuleStress = "stress"
	ModuleCache  = "cache"

	CheckConnectivity = "connectivity"
	CheckUsage        = "usage"
	CheckSampling     = "sampling"
	CheckJSONMode     = "json_mode"
	CheckToolCall     = "tool_call"
	CheckAuthError    = "auth_error"
	CheckBadRequest   = "bad_request"
	CheckRequestID    = "request_id"
	CheckThinking     = "thinking"
	CheckStream       = "stream_format"
	CheckCacheWarm    = "cache_warm"
	CheckCacheProbe   = "cache_probe"
	CheckCacheTokens  = "cache_tokens"
	CheckCacheHitRate = "cache_hit_rate"
	CheckCacheTTL     = "cache_ttl"

	DefaultStressPrompt  = "请用简洁的中文写一段话，说明流式输出在接口压测中的作用。"
	DefaultBasicPrompt   = "请用一句话介绍你自己。"
	DefaultCachePrefix   = "这是一段用于提示缓存的稳定前缀。请记住这段文字，后续问题会基于它。"
	DefaultCacheFollowUp = "根据前面的内容，只用一个词回复：pong"
)

var allBasicChecks = []string{
	CheckConnectivity,
	CheckStream,
	CheckUsage,
	CheckRequestID,
	CheckSampling,
	CheckJSONMode,
	CheckToolCall,
	CheckThinking,
	CheckAuthError,
	CheckBadRequest,
}

var basicCheckTitles = map[string]string{
	CheckConnectivity: "Connectivity",
	CheckStream:       "Stream format",
	CheckUsage:        "Usage fields",
	CheckRequestID:    "Request id",
	CheckSampling:     "Sampling parameters",
	CheckJSONMode:     "JSON mode",
	CheckToolCall:     "Tool call",
	CheckThinking:     "Thinking mode",
	CheckAuthError:    "Auth error",
	CheckBadRequest:   "Bad request",
}

const (
	maxConcurrency = 1000
	maxRounds      = 10000
	maxTokensCap   = 256000
	maxCacheRounds = 50
)
