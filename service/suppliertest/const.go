package suppliertest

import "time"

const (
	ModuleBasic  = "basic"
	ModuleStress = "stress"
	ModuleCache  = "cache"
	ModuleVideo  = "video"

	CheckConnectivity = "connectivity"
	CheckUsage        = "usage"
	CheckSampling     = "sampling"
	CheckJSONMode     = "json_mode"
	CheckToolCall     = "tool_call"
	CheckAuthError    = "auth_error"
	CheckBadRequest   = "bad_request"
	CheckRequestID    = "request_id"
	CheckThinking     = "thinking"
	CheckKimiKVV      = "kimi_kvv"
	CheckStream       = "stream_format"
	CheckCacheWarm    = "cache_warm"
	CheckCacheProbe   = "cache_probe"
	CheckCacheTokens  = "cache_tokens"
	CheckCacheHitRate = "cache_hit_rate"
	CheckCacheTTL     = "cache_ttl"

	CheckVideoSubmit = "video_submit"
	CheckVideoPoll   = "video_poll"
	CheckVideoResult = "video_result"

	DefaultStressPrompt  = "请用简洁的中文写一段话，说明流式输出在接口压测中的作用。"
	DefaultBasicPrompt   = "请用一句话介绍你自己。"
	DefaultCachePrefix   = "这是一段用于提示缓存的稳定前缀。请记住这段文字，后续问题会基于它。"
	DefaultCacheFollowUp = "根据前面的内容，只用一个词回复：pong"
	DefaultVideoPrompt   = "镜头缓慢向前推进，画面中的人物在雨夜霓虹街道穿梭，4K高清，电影级光影"

	CacheModeStatic     = "static"
	CacheModeCumulative = "cumulative"
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
	CheckKimiKVV,
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
	CheckKimiKVV:      "KVV preflight",
	CheckAuthError:    "Auth error",
	CheckBadRequest:   "Bad request",
}

const (
	maxConcurrency    = 1000
	maxRounds         = 10000
	maxStressRequests = 10000
	maxTokensCap      = 256000
	maxCacheRounds    = 50
	// Supplier models may spend more than a minute before returning the first
	// token. These limits apply only to the isolated supplier-test client.
	basicChatTimeout        = 3 * time.Minute
	stressChatTimeout       = 6 * time.Minute
	cacheChatTimeout        = 6 * time.Minute
	videoSubmitTimeout      = 5 * time.Minute
	videoPollRequestTimeout = 3 * time.Minute
	videoPollTotalTimeout   = 40 * time.Minute
)
