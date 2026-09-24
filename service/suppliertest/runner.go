package suppliertest

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/tidwall/gjson"
)

type RunRequest struct {
	BaseURL string       `json:"base_url"`
	APIKey  string       `json:"api_key"`
	Model   string       `json:"model"`
	Vendor  string       `json:"vendor,omitempty"`
	Modules []string     `json:"modules"`
	Basic   BasicConfig  `json:"basic"`
	Cache   CacheConfig  `json:"cache"`
	Stress  StressConfig `json:"stress"`
	Video   VideoConfig  `json:"video"`
}

type BasicConfig struct {
	Prompt      string   `json:"prompt"`
	MaxTokens   int      `json:"max_tokens"`
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	Stream      *bool    `json:"stream"`
	Checks      []string `json:"checks,omitempty"`
}

type CacheConfig struct {
	Prompt      string `json:"prompt"`
	FollowUp    string `json:"follow_up"`
	WaitSeconds int    `json:"wait_seconds"`
	MaxTokens   int    `json:"max_tokens"`
	Rounds      int    `json:"rounds"`
	Stream      *bool  `json:"stream"`
	Mode        string `json:"mode"`
}

type StressConfig struct {
	Concurrency int    `json:"concurrency"`
	Rounds      int    `json:"rounds"`
	MaxTokens   int    `json:"max_tokens"`
	Prompt      string `json:"prompt"`
	BreakCache  bool   `json:"break_cache"`
	Stream      *bool  `json:"stream"`
}

type VideoConfig struct {
	Prompt          string   `json:"prompt"`
	UploadMode      string   `json:"upload_mode,omitempty"`
	ImageURL        string   `json:"image_url,omitempty"`
	Base64Data      string   `json:"base64_data,omitempty"`
	Role            *string  `json:"role,omitempty"`
	LastFrameMode   string   `json:"last_frame_mode,omitempty"`
	LastFrameURL    string   `json:"last_frame_url,omitempty"`
	LastFrameBase64 string   `json:"last_frame_base64,omitempty"`
	Resolution      *string  `json:"resolution,omitempty"`
	Ratio           *string  `json:"ratio,omitempty"`
	Duration        *int     `json:"duration,omitempty"`
	Watermark       *bool    `json:"watermark,omitempty"`
	Seed            *int     `json:"seed,omitempty"`
	GenerateAudio   *bool    `json:"generate_audio,omitempty"`
	ReturnLastFrame *bool    `json:"return_last_frame,omitempty"`
	CustomJSON      string   `json:"custom_json,omitempty"`
	CustomPath      string   `json:"custom_path,omitempty"`
	RawPayload      string   `json:"raw_payload,omitempty"`
	TaskID          string   `json:"task_id,omitempty"`
	Checks          []string `json:"checks,omitempty"`
}

type Event struct {
	Type      string         `json:"type"`
	Module    string         `json:"module,omitempty"`
	CheckID   string         `json:"check_id,omitempty"`
	Status    string         `json:"status,omitempty"`
	Title     string         `json:"title,omitempty"`
	Message   string         `json:"message,omitempty"`
	Summary   string         `json:"summary,omitempty"`
	Worker    int            `json:"worker,omitempty"`
	Text      string         `json:"text,omitempty"`
	Completed int            `json:"completed,omitempty"`
	Total     int            `json:"total,omitempty"`
	Metrics   *StressMetrics `json:"metrics,omitempty"`
	Cache     *CacheMetrics  `json:"cache,omitempty"`
	Video     *VideoMetrics  `json:"video,omitempty"`
}

type VideoMetrics struct {
	TaskID                string  `json:"task_id"`
	Status                string  `json:"status"`
	VideoURL              string  `json:"video_url,omitempty"`
	EndpointURL           string  `json:"endpoint_url,omitempty"`
	ElapsedMS             float64 `json:"elapsed_ms"`
	FailReason            string  `json:"fail_reason,omitempty"`
	RawRequestJSON        string  `json:"raw_request_json,omitempty"`
	RawSubmitResponseJSON string  `json:"raw_submit_response_json,omitempty"`
	RawPollResponseJSON   string  `json:"raw_poll_response_json,omitempty"`
}

type StressMetrics struct {
	Total               int           `json:"total"`
	Attempted           int           `json:"attempted"`
	NotRun              int           `json:"not_run"`
	Succeeded           int           `json:"succeeded"`
	Failed              int           `json:"failed"`
	ErrorRate           float64       `json:"error_rate"`
	ElapsedMS           float64       `json:"elapsed_ms"`
	TokensPerSec        float64       `json:"tokens_per_sec"`
	RequestTokensPerSec float64       `json:"request_tokens_per_sec"`
	RequestAvgMS        float64       `json:"request_avg_ms"`
	RequestP50MS        float64       `json:"request_p50_ms"`
	RequestP90MS        float64       `json:"request_p90_ms"`
	UsageN              int           `json:"usage_n"`
	PromptTokens        int           `json:"prompt_tokens"`
	CompletionTokens    int           `json:"completion_tokens"`
	TTFTAvgMS           float64       `json:"ttft_avg_ms"`
	TTFTP50MS           float64       `json:"ttft_p50_ms"`
	TTFTP90MS           float64       `json:"ttft_p90_ms"`
	TTFTN               int           `json:"ttft_n"`
	TPOTAvgMS           float64       `json:"tpot_avg_ms"`
	TPOTP50MS           float64       `json:"tpot_p50_ms"`
	TPOTP90MS           float64       `json:"tpot_p90_ms"`
	TPOTN               int           `json:"tpot_n"`
	RPM                 float64       `json:"rpm"`
	TPM                 float64       `json:"tpm"`
	Issues              []StressIssue `json:"issues,omitempty"`
	OtherIssueCount     int           `json:"other_issue_count,omitempty"`
}

type StressIssue struct {
	StatusCode int     `json:"status_code"`
	Message    string  `json:"message"`
	Count      int     `json:"count"`
	Worker     int     `json:"worker"`
	Round      int     `json:"round"`
	ElapsedMS  float64 `json:"elapsed_ms"`
}

type CacheMetrics struct {
	WarmPromptTokens int     `json:"warm_prompt_tokens"`
	AvgHitRate       float64 `json:"avg_hit_rate"`
	MinHitRate       float64 `json:"min_hit_rate"`
	LastCachedTokens int     `json:"last_cached_tokens"`
	LastPromptTokens int     `json:"last_prompt_tokens"`
	WaitSeconds      int     `json:"wait_seconds"`
	Rounds           int     `json:"rounds"`
	HasCachedTokens  bool    `json:"has_cached_tokens"`
	HitCount         int     `json:"hit_count"`
	AvgDepthRate     float64 `json:"avg_depth_rate"`
	Mode             string  `json:"mode,omitempty"`
}

type Emitter func(Event)

func NormalizeRunRequest(req *RunRequest) error {
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	req.Model = strings.TrimSpace(req.Model)
	if req.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	if len(req.Modules) == 0 {
		req.Modules = []string{ModuleBasic}
	}
	seen := map[string]bool{}
	normalized := make([]string, 0, len(req.Modules))
	for _, module := range req.Modules {
		module = strings.TrimSpace(strings.ToLower(module))
		if module == "" || seen[module] {
			continue
		}
		if module != ModuleBasic && module != ModuleStress && module != ModuleCache && module != ModuleVideo {
			return fmt.Errorf("unknown module %q", module)
		}
		seen[module] = true
		normalized = append(normalized, module)
	}
	if len(normalized) == 0 {
		return fmt.Errorf("select at least one module")
	}
	req.Modules = normalized
	hasBasic := seen[ModuleBasic]
	hasStress := seen[ModuleStress]
	hasCache := seen[ModuleCache]
	if strings.TrimSpace(req.Video.Prompt) == "" {
		req.Video.Prompt = DefaultVideoPrompt
	}
	if req.Stress.Concurrency == 0 {
		req.Stress.Concurrency = 10
	}
	if req.Stress.Rounds == 0 {
		req.Stress.Rounds = 1
	}
	if req.Stress.MaxTokens == 0 {
		req.Stress.MaxTokens = 256
	}
	if hasStress {
		if req.Stress.Concurrency < 1 || req.Stress.Concurrency > maxConcurrency {
			return fmt.Errorf("concurrency must be between 1 and %d", maxConcurrency)
		}
		if req.Stress.Rounds < 1 || req.Stress.Rounds > maxRounds {
			return fmt.Errorf("rounds must be between 1 and %d", maxRounds)
		}
		if req.Stress.Concurrency > maxStressRequests/req.Stress.Rounds {
			return fmt.Errorf("concurrency × rounds must be at most %d requests", maxStressRequests)
		}
		if req.Stress.MaxTokens < 1 || req.Stress.MaxTokens > maxTokensCap {
			return fmt.Errorf("max_tokens must be between 1 and %d", maxTokensCap)
		}
	}
	if strings.TrimSpace(req.Stress.Prompt) == "" {
		req.Stress.Prompt = DefaultStressPrompt
	}
	if req.Stress.Stream == nil {
		req.Stress.Stream = ptrBool(true)
	}
	if strings.TrimSpace(req.Basic.Prompt) == "" {
		req.Basic.Prompt = DefaultBasicPrompt
	}
	if req.Basic.MaxTokens == 0 {
		req.Basic.MaxTokens = 64
	}
	if hasBasic && (req.Basic.MaxTokens < 1 || req.Basic.MaxTokens > maxTokensCap) {
		return fmt.Errorf("basic max_tokens must be between 1 and %d", maxTokensCap)
	}
	if req.Basic.Stream == nil {
		req.Basic.Stream = ptrBool(true)
	}
	if hasBasic {
		checks, err := filterBasicChecks(req.Basic.Checks)
		if err != nil {
			return err
		}
		req.Basic.Checks = checks
	} else if len(req.Basic.Checks) == 0 {
		req.Basic.Checks = allBasicChecks
	} else {
		checks, err := filterBasicChecks(req.Basic.Checks)
		if err != nil {
			req.Basic.Checks = allBasicChecks
		} else {
			req.Basic.Checks = checks
		}
	}
	if req.Cache.MaxTokens == 0 {
		req.Cache.MaxTokens = 16
	}
	if req.Cache.Rounds == 0 {
		req.Cache.Rounds = 5
	}
	if hasCache {
		if req.Cache.WaitSeconds < 0 || req.Cache.WaitSeconds > 600 {
			return fmt.Errorf("cache wait must be between 0 and 600 seconds")
		}
		if req.Cache.MaxTokens < 1 || req.Cache.MaxTokens > maxTokensCap {
			return fmt.Errorf("cache max_tokens must be between 1 and %d", maxTokensCap)
		}
		if req.Cache.Rounds < 1 || req.Cache.Rounds > maxCacheRounds {
			return fmt.Errorf("cache rounds must be between 1 and %d", maxCacheRounds)
		}
	}
	if strings.TrimSpace(req.Cache.FollowUp) == "" {
		req.Cache.FollowUp = DefaultCacheFollowUp
	}
	if req.Cache.Stream == nil {
		req.Cache.Stream = ptrBool(true)
	}
	if hasCache {
		cacheMode := strings.ToLower(strings.TrimSpace(req.Cache.Mode))
		if cacheMode == "" || cacheMode == CacheModeStatic {
			req.Cache.Mode = CacheModeStatic
		} else if cacheMode == CacheModeCumulative {
			req.Cache.Mode = CacheModeCumulative
		} else {
			return fmt.Errorf("invalid cache mode %q", req.Cache.Mode)
		}
	}
	vendor, err := ResolveVendor(req.Vendor)
	if err != nil {
		return err
	}
	req.Vendor = vendor
	return nil
}

func filterBasicChecks(requested []string) ([]string, error) {
	if len(requested) == 0 {
		out := make([]string, len(allBasicChecks))
		copy(out, allBasicChecks)
		return out, nil
	}
	allowed := make(map[string]bool, len(allBasicChecks))
	for _, id := range allBasicChecks {
		allowed[id] = true
	}
	seen := make(map[string]bool, len(requested))
	for _, id := range requested {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if !allowed[id] {
			return nil, fmt.Errorf("unknown check %q", id)
		}
		seen[id] = true
	}
	if len(seen) == 0 {
		return nil, fmt.Errorf("select at least one check")
	}
	out := make([]string, 0, len(seen))
	for _, id := range allBasicChecks {
		if seen[id] {
			out = append(out, id)
		}
	}
	return out, nil
}

func Run(ctx context.Context, httpClient *http.Client, req RunRequest, emit Emitter) error {
	if httpClient == nil {
		return fmt.Errorf("http client is required")
	}
	if emit == nil {
		emit = func(Event) {}
	}
	if err := NormalizeRunRequest(&req); err != nil {
		return err
	}
	var endpoint string
	hasChatModule := false
	for _, m := range req.Modules {
		if m == ModuleBasic || m == ModuleStress || m == ModuleCache {
			hasChatModule = true
			break
		}
	}
	if hasChatModule {
		var err error
		endpoint, err = ChatCompletionsURL(req.BaseURL)
		if err != nil {
			return err
		}
	}
	ctx = withStreamOptionsTracker(ctx)
	client := NewHTTPClient(httpClient)
	var emitMu sync.Mutex
	safeEmit := func(event Event) {
		emitMu.Lock()
		defer emitMu.Unlock()
		emit(event)
	}

	for _, module := range req.Modules {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		switch module {
		case ModuleBasic:
			runBasic(ctx, client, endpoint, req, safeEmit)
		case ModuleStress:
			runStress(ctx, client, endpoint, req, safeEmit)
		case ModuleCache:
			runCache(ctx, client, endpoint, req, safeEmit)
		case ModuleVideo:
			runVideo(ctx, client, req, safeEmit)
		}
	}
	safeEmit(Event{Type: "done"})
	return nil
}

func applyStream(req chatRequest, stream bool) chatRequest {
	req.Stream = stream
	if stream {
		req.StreamOptions = &streamOptions{IncludeUsage: true}
	} else {
		req.StreamOptions = nil
	}
	return req
}

func runBasic(ctx context.Context, httpClient *http.Client, endpoint string, req RunRequest, emit Emitter) {
	wanted := make(map[string]bool, len(req.Basic.Checks))
	for _, id := range req.Basic.Checks {
		wanted[id] = true
	}
	counts := map[string]int{}
	emitCheck := func(id, status, message string) {
		emit(Event{
			Type:    "check",
			Module:  ModuleBasic,
			CheckID: id,
			Title:   basicCheckTitles[id],
			Status:  status,
			Message: message,
		})
		if status != "running" {
			counts[status]++
		}
	}
	profile := profileFor(req.Vendor)
	stream := boolVal(req.Basic.Stream, true)
	chat := applyStream(chatRequest{
		Model:       req.Model,
		Messages:    []chatMessage{{Role: "user", Content: req.Basic.Prompt}},
		MaxTokens:   ptrInt(req.Basic.MaxTokens),
		Temperature: req.Basic.Temperature,
		TopP:        req.Basic.TopP,
	}, stream)

	needShared := wanted[CheckConnectivity] || wanted[CheckStream] || wanted[CheckUsage] || wanted[CheckRequestID]
	var connected StreamResult
	sharedFailed := false
	sharedFailMsg := ""
	if needShared {
		if wanted[CheckConnectivity] {
			emitCheck(CheckConnectivity, "running", "")
		}
		var onDelta func(StreamDelta)
		if wanted[CheckConnectivity] && stream {
			onDelta = func(delta StreamDelta) {
				text := delta.Content
				if text == "" {
					text = delta.Reasoning
				}
				if text != "" {
					emit(Event{Type: "stream", Module: ModuleBasic, Text: text})
				}
			}
		}
		connected = streamChat(ctx, httpClient, endpoint, req.APIKey, chat, basicChatTimeout, onDelta)
		gotOutput := connected.Content != "" || connected.Reasoning != "" || connected.ToolName != ""
		httpFailed := connected.StatusCode != http.StatusOK || connected.ErrorMessage != ""
		if httpFailed {
			sharedFailed = true
			sharedFailMsg = firstNonEmpty(connected.ErrorMessage, fmt.Sprintf("HTTP %d", connected.StatusCode))
			if wanted[CheckConnectivity] {
				emitCheck(CheckConnectivity, "fail", "连通失败："+sharedFailMsg)
			}
			for _, id := range []string{CheckStream, CheckUsage, CheckRequestID} {
				if wanted[id] {
					emitCheck(id, "skip", "连通失败，这项无法对照："+sharedFailMsg)
				}
			}
		} else {
			if wanted[CheckConnectivity] {
				if !gotOutput {
					emitCheck(CheckConnectivity, "fail", "HTTP 200，但没有返回正文、思考内容或工具调用")
				} else if connected.SSE {
					emitCheck(CheckConnectivity, "pass", "HTTP 200，流式输出正常")
				} else {
					emitCheck(CheckConnectivity, "pass", "HTTP 200，非流式 JSON 正常")
				}
			}
			if wanted[CheckStream] {
				if connected.FinishReason != "" {
					emitCheck(CheckStream, "pass", "finish_reason="+connected.FinishReason)
				} else if connected.SSE {
					emitCheck(CheckStream, "skip", "流式正常，但供应商没返回 finish_reason")
				} else {
					emitCheck(CheckStream, "skip", "这次是非流式 JSON，没有 SSE 帧可检查")
				}
			}
			if wanted[CheckUsage] {
				if connected.HasUsage && connected.HasPromptTokens && connected.HasCompletionTokens {
					msg := fmt.Sprintf("prompt=%d，completion=%d", connected.PromptTokens, connected.CompletionTokens)
					if connected.HasReasoningTokens || connected.ReasoningTokens > 0 {
						msg += fmt.Sprintf("（reasoning=%d）", connected.ReasoningTokens)
					}
					if connected.HasCachedTokens && connected.CachedTokens > 0 {
						msg += fmt.Sprintf("（cached=%d）", connected.CachedTokens)
					}
					emitCheck(CheckUsage, "pass", msg)
				} else {
					status, message := checkStatus(
						profile.requireUsage,
						"供应商返回的 usage 缺少 prompt_tokens 或 completion_tokens，已跳过",
						vendorTitle(profile.id)+" 文档要求返回完整 usage（prompt_tokens 和 completion_tokens），这次没有",
					)
					emitCheck(CheckUsage, status, message)
				}
			}
			if wanted[CheckRequestID] {
				if id := requestIDFrom(connected); id != "" {
					emitCheck(CheckRequestID, "pass", id)
				} else {
					emitCheck(CheckRequestID, "skip", "响应里没有 id，请求头也没有 request-id")
				}
			}
		}
	}

	if wanted[CheckSampling] {
		emitCheck(CheckSampling, "running", "")
		if sharedFailed {
			emitCheck(CheckSampling, "skip", "连通失败，这项无法对照："+sharedFailMsg)
		} else {
			sampled := connected
			if !needShared {
				sampled = streamChat(ctx, httpClient, endpoint, req.APIKey, chat, basicChatTimeout, nil)
			}
			samplingNote := "供应商接受了当前 max_tokens"
			if req.Basic.Temperature != nil || req.Basic.TopP != nil {
				samplingNote = "供应商接受了当前采样参数"
			}
			if sampled.StatusCode == http.StatusOK && sampled.ErrorMessage == "" {
				emitCheck(CheckSampling, "pass", samplingNote)
			} else if sampled.StatusCode >= 400 && sampled.StatusCode < 500 {
				emitCheck(CheckSampling, "skip", "供应商拒绝了采样字段："+firstNonEmpty(sampled.ErrorMessage, fmt.Sprintf("HTTP %d", sampled.StatusCode)))
			} else {
				emitCheck(CheckSampling, "fail", firstNonEmpty(sampled.ErrorMessage, fmt.Sprintf("HTTP %d", sampled.StatusCode)))
			}
		}
	}

	if wanted[CheckJSONMode] {
		emitCheck(CheckJSONMode, "running", "")
		jsonReq := chat
		// JSON mode is a response-format contract check, not a streaming check.
		// Keep it non-streaming so providers such as Kimi are compared against
		// their documented JSON response behavior without inheriting the basic
		// stream_options/include_usage fields. Streaming is covered separately by
		// CheckStream.
		jsonReq.Stream = false
		jsonReq.StreamOptions = nil
		jsonReq.ResponseFormat = map[string]any{"type": "json_object"}
		jsonReq.Messages = []chatMessage{{Role: "user", Content: "Return a JSON object with key ping and value pong."}}
		jsonResult := streamChat(ctx, httpClient, endpoint, req.APIKey, jsonReq, basicChatTimeout, nil)
		if jsonResult.StatusCode != http.StatusOK || jsonResult.ErrorMessage != "" {
			status, message := checkStatus(
				profile.requireJSON,
				"供应商不接受 JSON 模式："+firstNonEmpty(jsonResult.ErrorMessage, fmt.Sprintf("HTTP %d", jsonResult.StatusCode)),
				vendorTitle(profile.id)+" 文档标注支持 JSON 模式，但请求失败："+firstNonEmpty(jsonResult.ErrorMessage, fmt.Sprintf("HTTP %d", jsonResult.StatusCode)),
			)
			emitCheck(CheckJSONMode, status, message)
		} else if looksLikeJSON(jsonResult.Content) {
			emitCheck(CheckJSONMode, "pass", "返回内容可以解析成 JSON")
		} else {
			status, message := checkStatus(
				profile.requireJSON,
				"接口成功了，但正文不是 JSON",
				vendorTitle(profile.id)+" 接受了 JSON 模式，但正文不是 JSON",
			)
			emitCheck(CheckJSONMode, status, message)
		}
	}

	if wanted[CheckToolCall] {
		emitCheck(CheckToolCall, "running", "")
		toolReq := chat
		toolReq.Messages = []chatMessage{{Role: "user", Content: "What is the weather in Beijing? Use the get_weather tool."}}
		toolReq.Tools = []map[string]any{
			{
				"type": "function",
				"function": map[string]any{
					"name":        "get_weather",
					"description": "Get the weather for a city",
					"parameters": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"city": map[string]any{"type": "string"},
						},
						"required": []string{"city"},
					},
				},
			},
		}
		toolResult := streamChat(ctx, httpClient, endpoint, req.APIKey, toolReq, basicChatTimeout, nil)
		if toolResult.StatusCode != http.StatusOK || toolResult.ErrorMessage != "" {
			status, message := checkStatus(
				profile.requireTools,
				"供应商不接受工具调用："+firstNonEmpty(toolResult.ErrorMessage, fmt.Sprintf("HTTP %d", toolResult.StatusCode)),
				vendorTitle(profile.id)+" 文档标注支持工具调用，但请求失败："+firstNonEmpty(toolResult.ErrorMessage, fmt.Sprintf("HTTP %d", toolResult.StatusCode)),
			)
			emitCheck(CheckToolCall, status, message)
		} else if toolResult.ToolName != "" {
			emitCheck(CheckToolCall, "pass", "调用了工具 "+toolResult.ToolName)
		} else {
			status, message := checkStatus(
				profile.requireTools,
				"这次响应里没有 tool_calls",
				vendorTitle(profile.id)+" 接受了 tools，但响应里没有 tool_calls",
			)
			emitCheck(CheckToolCall, status, message)
		}
	}

	if wanted[CheckThinking] {
		emitCheck(CheckThinking, "running", "")
		mustThink := thinkingRequired(profile.id, req.Model)
		thinkReq := applyThinking(chat, profile.id)
		thinking := streamChat(ctx, httpClient, endpoint, req.APIKey, thinkReq, basicChatTimeout, nil)
		if (thinking.StatusCode != http.StatusOK || thinking.ErrorMessage != "") && (thinkReq.Thinking != nil || thinkReq.ReasoningEffort != "") {
			fallbackReq := chat
			fallbackReq.Messages = []chatMessage{{Role: "user", Content: "What is 17 times 19? Think step by step."}}
			fallbackReq.Thinking = nil
			fallbackReq.ReasoningEffort = ""
			thinking = streamChat(ctx, httpClient, endpoint, req.APIKey, fallbackReq, basicChatTimeout, nil)
		}
		if thinking.StatusCode != http.StatusOK || thinking.ErrorMessage != "" {
			status, message := checkStatus(
				mustThink,
				"供应商不接受思考控制参数："+firstNonEmpty(thinking.ErrorMessage, fmt.Sprintf("HTTP %d", thinking.StatusCode)),
				vendorTitle(profile.id)+" 该模型应按文档返回思考内容，请求失败："+firstNonEmpty(thinking.ErrorMessage, fmt.Sprintf("HTTP %d", thinking.StatusCode)),
			)
			emitCheck(CheckThinking, status, message)
		} else if thinking.Reasoning != "" || thinking.ReasoningTokens > 0 {
			msg := "返回了 reasoning 内容"
			if thinking.Reasoning == "" && thinking.ReasoningTokens > 0 {
				msg = fmt.Sprintf("返回了 reasoning_tokens=%d", thinking.ReasoningTokens)
			} else if thinking.ReasoningTokens > 0 {
				msg = fmt.Sprintf("返回了 reasoning 内容（reasoning_tokens=%d）", thinking.ReasoningTokens)
			}
			emitCheck(CheckThinking, "pass", msg)
		} else {
			status, message := checkStatus(
				mustThink,
				"这次没有 reasoning 字段",
				vendorTitle(profile.id)+" 该模型应按文档返回 reasoning_content，这次没有",
			)
			emitCheck(CheckThinking, status, message)
		}
	}

	if wanted[CheckKimiKVV] {
		emitCheck(CheckKimiKVV, "running", "")
		status, message := "skip", "KVV 预检只适用于 Kimi K3"
		if profile.id == VendorKimi && strings.Contains(modelKey(req.Model), "kimik3") {
			status, message = runKimiKVV(ctx, httpClient, endpoint, req.APIKey, chat)
		}
		emitCheck(CheckKimiKVV, status, message)
	}

	if wanted[CheckAuthError] {
		emitCheck(CheckAuthError, "running", "")
		unauthorized := streamChat(ctx, httpClient, endpoint, "invalid-supplier-test-key", applyStream(chatRequest{
			Model:     req.Model,
			Messages:  []chatMessage{{Role: "user", Content: "ping"}},
			MaxTokens: ptrInt(8),
		}, stream), 30*time.Second, nil)
		if unauthorized.StatusCode == http.StatusUnauthorized || unauthorized.StatusCode == http.StatusForbidden {
			emitCheck(CheckAuthError, "pass", fmt.Sprintf("无效 Key 返回了 HTTP %d，符合预期", unauthorized.StatusCode))
		} else {
			emitCheck(CheckAuthError, "skip", fmt.Sprintf("无效 Key 返回了 HTTP %d，不是 401/403", unauthorized.StatusCode))
		}
	}

	if wanted[CheckBadRequest] {
		emitCheck(CheckBadRequest, "running", "")
		bad := streamChat(ctx, httpClient, endpoint, req.APIKey, applyStream(chatRequest{
			Model: req.Model,
		}, stream), 30*time.Second, nil)
		if bad.StatusCode >= 400 && bad.StatusCode < 500 {
			emitCheck(CheckBadRequest, "pass", fmt.Sprintf("缺字段返回了 HTTP %d，符合预期", bad.StatusCode))
		} else {
			emitCheck(CheckBadRequest, "skip", fmt.Sprintf("缺字段返回了 HTTP %d，不是 4xx", bad.StatusCode))
		}
	}

	summary := fmt.Sprintf("基础检查结束：通过 %d，跳过 %d，失败 %d。跳过通常表示供应商没提供这个能力或字段，不是失败。", counts["pass"], counts["skip"], counts["fail"])
	if profile.id != VendorGeneric {
		summary = fmt.Sprintf("按 %s 字段规则检查。通过 %d，跳过 %d，失败 %d。", vendorTitle(profile.id), counts["pass"], counts["skip"], counts["fail"])
	}
	emit(Event{
		Type:    "summary",
		Module:  ModuleBasic,
		Summary: summary,
	})
}

func runKimiKVV(
	ctx context.Context,
	httpClient *http.Client,
	endpoint string,
	apiKey string,
	baseChat chatRequest,
) (string, string) {
	return runOfficialKimiKVV(ctx, httpClient, endpoint, apiKey, baseChat.Model)
}

func runCache(ctx context.Context, httpClient *http.Client, endpoint string, req RunRequest, emit Emitter) {
	emitCheck := func(id, title, status, message string) {
		emit(Event{Type: "check", Module: ModuleCache, CheckID: id, Title: title, Status: status, Message: message})
	}
	profile := profileFor(req.Vendor)
	stream := boolVal(req.Cache.Stream, true)
	prefix := strings.TrimSpace(req.Cache.Prompt)
	if prefix == "" {
		prefix = DefaultCachePrefix
	}
	followUp := strings.TrimSpace(req.Cache.FollowUp)
	if followUp == "" {
		followUp = DefaultCacheFollowUp
	}

	emitCheck(CheckCacheWarm, "Cache warm", "running", "")
	warmReq := applyStream(chatRequest{
		Model:     req.Model,
		Messages:  cacheMessages(profile, prefix, followUp, true),
		MaxTokens: ptrInt(req.Cache.MaxTokens),
	}, stream)
	warm := streamChat(ctx, httpClient, endpoint, req.APIKey, warmReq, cacheChatTimeout, nil)
	if warm.StatusCode != http.StatusOK || warm.ErrorMessage != "" {
		msg := firstNonEmpty(warm.ErrorMessage, fmt.Sprintf("HTTP %d", warm.StatusCode))
		emitCheck(CheckCacheWarm, "Cache warm", "fail", "预热失败："+msg)
		emit(Event{Type: "summary", Module: ModuleCache, Summary: "缓存测试中断：预热请求失败。"})
		return
	}
	emitCheck(CheckCacheWarm, "Cache warm", "pass", fmt.Sprintf("预热完成，prompt_tokens=%d，cached_tokens=%d", warm.PromptTokens, warm.CachedTokens))

	wait := time.Duration(req.Cache.WaitSeconds) * time.Second
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			emitCheck(CheckCacheProbe, "Cache probe", "skip", "等待时被取消")
			emit(Event{Type: "summary", Module: ModuleCache, Summary: "缓存测试中断：等待间隔时被取消。"})
			return
		case <-timer.C:
		}
	}

	rounds := req.Cache.Rounds
	if rounds < 1 {
		rounds = 1
	}
	emitCheck(CheckCacheProbe, "Cache probe", "running", "")
	probeMessages := cacheMessages(profile, prefix, followUp, false)
	probeReq := applyStream(chatRequest{
		Model:     req.Model,
		Messages:  probeMessages,
		MaxTokens: ptrInt(req.Cache.MaxTokens),
	}, stream)

	var (
		hitRates     []float64
		depthRates   []float64
		hitCount     int
		lastCached   int
		lastPrompt   int
		sawCached    bool
		probeDetails []string
		failedRound  int
	)
	for i := 0; i < rounds; i++ {
		if ctx.Err() != nil {
			emitCheck(CheckCacheProbe, "Cache probe", "skip", "探测时被取消")
			return
		}
		if req.Cache.Mode == CacheModeCumulative {
			probeReq.Messages = probeMessages
		}
		probe := streamChat(ctx, httpClient, endpoint, req.APIKey, probeReq, cacheChatTimeout, nil)
		if probe.StatusCode != http.StatusOK || probe.ErrorMessage != "" {
			failedRound++
			probeDetails = append(probeDetails, fmt.Sprintf("第%d轮失败：%s", i+1, firstNonEmpty(probe.ErrorMessage, fmt.Sprintf("HTTP %d", probe.StatusCode))))
			continue
		}
		effectivePrompt := probe.PromptTokens
		if probe.HasCachedTokens && probe.CachedTokens > 0 && effectivePrompt < probe.CachedTokens {
			effectivePrompt = probe.PromptTokens + probe.CachedTokens
		}
		lastPrompt = effectivePrompt
		lastCached = probe.CachedTokens
		if probe.HasCachedTokens {
			sawCached = true
		}
		if probe.HasCachedTokens && effectivePrompt > 0 {
			rate := float64(probe.CachedTokens) / float64(effectivePrompt)
			if rate > 1.0 {
				rate = 1.0
			}
			hitRates = append(hitRates, rate)
			if probe.CachedTokens > 0 {
				hitCount++
				depthRates = append(depthRates, rate)
			}
			probeDetails = append(probeDetails, fmt.Sprintf("第%d轮 %.1f%%", i+1, rate*100))
		} else {
			probeDetails = append(probeDetails, fmt.Sprintf("第%d轮 prompt=%d", i+1, effectivePrompt))
		}

		if req.Cache.Mode == CacheModeCumulative {
			replyText := strings.TrimSpace(probe.Content)
			if replyText == "" {
				replyText = "ok"
			}
			probeMessages = append(probeMessages,
				chatMessage{Role: "assistant", Content: replyText},
				chatMessage{Role: "user", Content: followUp},
			)
		}
	}

	emitCache := func(avgHit, minHit float64, hits int, avgDepth float64) {
		emit(Event{
			Type:   "metrics",
			Module: ModuleCache,
			Cache: &CacheMetrics{
				WarmPromptTokens: warm.PromptTokens,
				AvgHitRate:       avgHit,
				MinHitRate:       minHit,
				LastCachedTokens: lastCached,
				LastPromptTokens: lastPrompt,
				WaitSeconds:      req.Cache.WaitSeconds,
				Rounds:           rounds,
				HasCachedTokens:  sawCached,
				HitCount:         hits,
				AvgDepthRate:     avgDepth,
				Mode:             req.Cache.Mode,
			},
		})
	}

	if failedRound == rounds {
		emitCheck(CheckCacheProbe, "Cache probe", "fail", strings.Join(probeDetails, "；"))
		emitCheck(CheckCacheTTL, "Cache TTL", "skip", "探测全部失败，不能判定缓存存活")
		emitCache(0, 0, 0, 0)
		emit(Event{Type: "summary", Module: ModuleCache, Summary: fmt.Sprintf("缓存测试失败：探测 %d 轮全部失败。", rounds)})
		return
	}
	probeStatus := "pass"
	probeMessage := fmt.Sprintf("探测 %d 轮完成。%s", rounds, strings.Join(probeDetails, "，"))
	if failedRound > 0 {
		probeStatus = "fail"
		probeMessage = fmt.Sprintf("探测 %d 轮：成功 %d，失败 %d。%s", rounds, rounds-failedRound, failedRound, strings.Join(probeDetails, "，"))
	}
	emitCheck(CheckCacheProbe, "Cache probe", probeStatus, probeMessage)

	if !sawCached {
		status, message := checkStatus(
			profile.requireCacheField,
			"供应商没返回 cached_tokens，无法确认真实命中",
			vendorTitle(profile.id)+" 文档要求返回 cached_tokens，这次没有",
		)
		emitCheck(CheckCacheTokens, "Cached tokens", status, message)
		emitCheck(CheckCacheHitRate, "Cache hit rate", "skip", "没有 cached_tokens，算不出命中率")
		emitCheck(CheckCacheTTL, "Cache TTL", "skip", "没有 cached_tokens，不能判定缓存存活")
		emitCache(0, 0, 0, 0)
		summary := fmt.Sprintf("缓存测试结束：预热 1 次，探测 %d 轮。供应商没返回 cached_tokens，只能确认请求成功，不能确认缓存命中。", rounds)
		if failedRound > 0 {
			summary += fmt.Sprintf(" 另有 %d 轮探测失败。", failedRound)
		}
		emit(Event{
			Type:    "summary",
			Module:  ModuleCache,
			Summary: summary,
		})
		return
	}
	emitCheck(CheckCacheTokens, "Cached tokens", "pass", fmt.Sprintf("最后一轮 cached_tokens=%d", lastCached))
	if len(hitRates) == 0 {
		emitCheck(CheckCacheHitRate, "Cache hit rate", "skip", "prompt_tokens 缺失，算不出命中率")
		emitCheck(CheckCacheTTL, "Cache TTL", "skip", "没有命中率，不能判定缓存存活")
		emitCache(0, 0, 0, 0)
		summary := fmt.Sprintf("缓存测试结束：预热 1 次，探测 %d 轮。有 cached_tokens=%d，但没有 prompt_tokens。", rounds, lastCached)
		if failedRound > 0 {
			summary += fmt.Sprintf(" 另有 %d 轮探测失败。", failedRound)
		}
		emit(Event{Type: "summary", Module: ModuleCache, Summary: summary})
		return
	}
	avgHit := average(hitRates)
	minHit := hitRates[0]
	for _, rate := range hitRates[1:] {
		if rate < minHit {
			minHit = rate
		}
	}
	var avgDepth float64
	var minDepth float64
	if len(depthRates) > 0 {
		avgDepth = average(depthRates)
		minDepth = depthRates[0]
		for _, rate := range depthRates[1:] {
			if rate < minDepth {
				minDepth = rate
			}
		}
	}

	modeLabel := "固定前缀"
	if req.Cache.Mode == CacheModeCumulative {
		modeLabel = "多轮累加"
	}

	var hitMsg string
	if hitCount > 0 {
		hitMsg = fmt.Sprintf("[%s] 命中频次 %d/%d 轮 (%.1f%%)，有效命中深度均值 %.1f%%（最低 %.1f%%，最后一轮 %d/%d）。是否达标看页面尺子。",
			modeLabel, hitCount, len(hitRates), float64(hitCount)/float64(len(hitRates))*100,
			avgDepth*100, minDepth*100, lastCached, lastPrompt)
	} else {
		hitMsg = fmt.Sprintf("[%s] 命中频次 0/%d 轮 (0.0%%)（最后一轮 %d/%d）。是否达标看页面尺子。",
			modeLabel, len(hitRates), lastCached, lastPrompt)
	}
	emitCheck(CheckCacheHitRate, "Cache hit rate", "pass", hitMsg)
	if req.Cache.WaitSeconds <= 0 {
		emitCheck(CheckCacheTTL, "Cache TTL", "skip", fmt.Sprintf("等待 %ds，TTL 是否达标看页面尺子。", req.Cache.WaitSeconds))
	} else {
		emitCheck(CheckCacheTTL, "Cache TTL", "pass", fmt.Sprintf("等待 %ds 后平均命中率 %.1f%%。是否达标看页面尺子。", req.Cache.WaitSeconds, avgHit*100))
	}
	emitCache(avgHit, minHit, hitCount, avgDepth)
	summary := fmt.Sprintf("缓存测试结束：预热 1 次，探测 %d 轮。[%s] 命中频次 %d/%d 轮 (%.1f%%)，平均命中率 %.1f%%。", rounds, modeLabel, hitCount, len(hitRates), float64(hitCount)/float64(len(hitRates))*100, avgHit*100)
	if failedRound > 0 {
		summary += fmt.Sprintf(" 另有 %d 轮探测失败。", failedRound)
	}
	emit(Event{
		Type:    "summary",
		Module:  ModuleCache,
		Summary: summary,
	})
}

func runStress(ctx context.Context, httpClient *http.Client, endpoint string, req RunRequest, emit Emitter) {
	total := req.Stress.Concurrency * req.Stress.Rounds
	emit(Event{Type: "progress", Module: ModuleStress, Completed: 0, Total: total})
	stream := boolVal(req.Stress.Stream, true)

	var (
		completed    atomic.Int64
		succeeded    atomic.Int64
		failed       atomic.Int64
		inTokens     atomic.Int64
		outTokens    atomic.Int64
		usageCount   atomic.Int64
		usageElapsed atomic.Int64
		ttftMu       sync.Mutex
		ttfts        []float64
		tpots        []float64
		requestTimes []float64
		issueMu      sync.Mutex
		issues       = make(map[string]*StressIssue)
		otherIssues  int
	)

	stressPrompt := strings.TrimSpace(req.Stress.Prompt)
	if stressPrompt == "" {
		stressPrompt = DefaultStressPrompt
	}
	baseReq := applyStream(chatRequest{
		Model:     req.Model,
		Messages:  []chatMessage{{Role: "user", Content: stressPrompt}},
		MaxTokens: ptrInt(req.Stress.MaxTokens),
	}, stream)

	started := time.Now()
	var wg sync.WaitGroup
	for worker := 0; worker < req.Stress.Concurrency; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for round := 0; round < req.Stress.Rounds; round++ {
				if ctx.Err() != nil {
					return
				}
				var onDelta func(StreamDelta)
				if stream && workerID == 0 && round == 0 {
					onDelta = func(delta StreamDelta) {
						text := delta.Content
						if text == "" {
							text = delta.Reasoning
						}
						if text != "" {
							emit(Event{Type: "stream", Module: ModuleStress, Worker: workerID, Text: text})
						}
					}
				}
				stressReq := baseReq
				if req.Stress.BreakCache {
					stressReq.Messages = []chatMessage{{
						Role:    "user",
						Content: cacheBustPrefix() + stressPrompt,
					}}
				}
				result := streamChat(ctx, httpClient, endpoint, req.APIKey, stressReq, stressChatTimeout, onDelta)
				ttftMu.Lock()
				requestTimes = append(requestTimes, float64(result.Elapsed)/float64(time.Millisecond))
				ttftMu.Unlock()
				ok := result.StatusCode == http.StatusOK && result.ErrorMessage == "" && (result.Content != "" || result.Reasoning != "")
				if ok {
					succeeded.Add(1)
					if result.HasUsage && result.HasPromptTokens && result.HasCompletionTokens {
						usageCount.Add(1)
						usageElapsed.Add(result.Elapsed.Nanoseconds())
						inTokens.Add(int64(result.PromptTokens))
						outTokens.Add(int64(result.CompletionTokens))
					}
					ttftMu.Lock()
					if stream && result.SSE && result.TTFT > 0 {
						ttfts = append(ttfts, float64(result.TTFT)/float64(time.Millisecond))
						if result.TPOT > 0 {
							tpots = append(tpots, float64(result.TPOT)/float64(time.Millisecond))
						}
					}
					ttftMu.Unlock()
				} else {
					failed.Add(1)
					failure := firstNonEmpty(result.ErrorMessage, fmt.Sprintf("HTTP %d", result.StatusCode))
					if result.StatusCode == http.StatusOK && result.ErrorMessage == "" {
						failure = "HTTP 200，但没有返回正文或思考内容"
					}
					failure = strings.Join(strings.Fields(failure), " ")
					if chars := []rune(failure); len(chars) > 300 {
						failure = string(chars[:300]) + "…"
					}
					key := fmt.Sprintf("%d\x00%s", result.StatusCode, failure)
					issueMu.Lock()
					if issue, found := issues[key]; found {
						issue.Count++
					} else if len(issues) < 50 {
						issues[key] = &StressIssue{
							StatusCode: result.StatusCode,
							Message:    failure,
							Count:      1,
							Worker:     workerID + 1,
							Round:      round + 1,
							ElapsedMS:  float64(result.Elapsed) / float64(time.Millisecond),
						}
					} else {
						otherIssues++
					}
					issueMu.Unlock()
				}
				done := int(completed.Add(1))
				if done == total || done%max(1, total/20) == 0 {
					emit(Event{Type: "progress", Module: ModuleStress, Completed: done, Total: total})
				}
			}
		}(worker)
	}
	wg.Wait()

	elapsed := time.Since(started)
	ok := int(succeeded.Load())
	bad := int(failed.Load())
	promptTotal := int(inTokens.Load())
	completionTotal := int(outTokens.Load())
	metrics := &StressMetrics{
		Total:            total,
		Attempted:        int(completed.Load()),
		NotRun:           total - int(completed.Load()),
		Succeeded:        ok,
		Failed:           bad,
		ElapsedMS:        float64(elapsed) / float64(time.Millisecond),
		PromptTokens:     promptTotal,
		CompletionTokens: completionTotal,
		UsageN:           int(usageCount.Load()),
		TTFTN:            len(ttfts),
		TPOTN:            len(tpots),
		OtherIssueCount:  otherIssues,
	}
	for _, issue := range issues {
		metrics.Issues = append(metrics.Issues, *issue)
	}
	sort.Slice(metrics.Issues, func(i, j int) bool {
		if metrics.Issues[i].Count != metrics.Issues[j].Count {
			return metrics.Issues[i].Count > metrics.Issues[j].Count
		}
		if metrics.Issues[i].StatusCode != metrics.Issues[j].StatusCode {
			return metrics.Issues[i].StatusCode < metrics.Issues[j].StatusCode
		}
		return metrics.Issues[i].Message < metrics.Issues[j].Message
	})
	if metrics.Attempted > 0 {
		metrics.ErrorRate = float64(bad) / float64(metrics.Attempted)
	}
	if metrics.UsageN == ok && elapsed.Seconds() > 0 {
		metrics.TokensPerSec = float64(completionTotal) / elapsed.Seconds()
		if usageElapsed.Load() > 0 {
			metrics.RequestTokensPerSec = float64(completionTotal) / (float64(usageElapsed.Load()) / float64(time.Second))
		}
	}
	if elapsed.Minutes() > 0 {
		metrics.RPM = float64(ok) / elapsed.Minutes()
		if metrics.UsageN == ok {
			metrics.TPM = float64(promptTotal+completionTotal) / elapsed.Minutes()
		}
	}
	metrics.RequestAvgMS = average(requestTimes)
	metrics.RequestP50MS = percentile(requestTimes, 50)
	metrics.RequestP90MS = percentile(requestTimes, 90)
	metrics.TTFTAvgMS = average(ttfts)
	metrics.TTFTP50MS = percentile(ttfts, 50)
	metrics.TTFTP90MS = percentile(ttfts, 90)
	metrics.TPOTAvgMS = average(tpots)
	metrics.TPOTP50MS = percentile(tpots, 50)
	metrics.TPOTP90MS = percentile(tpots, 90)

	mode := "流式"
	if !stream {
		mode = "非流式"
	}
	cacheMode := "相同语料（可走缓存）"
	if req.Stress.BreakCache {
		cacheMode = "随机前缀（打断缓存）"
	}
	summary := fmt.Sprintf("%s压测结束（%s）：%d 并发 × %d 轮，计划 %d 次，实际完成 %d 次，成功 %d，失败 %d，未执行 %d 次，整批耗时 %.0f ms。",
		mode, cacheMode, req.Stress.Concurrency, req.Stress.Rounds, metrics.Total, metrics.Attempted, metrics.Succeeded, metrics.Failed, metrics.NotRun, metrics.ElapsedMS)
	if metrics.RequestAvgMS > 0 {
		summary += fmt.Sprintf(" 单请求总耗时: 均值 %.0f ms / P50 %.0f ms / P90 %.0f ms。", metrics.RequestAvgMS, metrics.RequestP50MS, metrics.RequestP90MS)
	}
	if metrics.TTFTAvgMS > 0 {
		summary += fmt.Sprintf(" TTFT首字: 均值 %.0f ms / P50 %.0f ms / P90 %.0f ms（n=%d）。", metrics.TTFTAvgMS, metrics.TTFTP50MS, metrics.TTFTP90MS, metrics.TTFTN)
	}
	if metrics.TPOTAvgMS > 0 {
		summary += fmt.Sprintf(" TPOT每Token: 均值 %.1f ms / P50 %.1f ms / P90 %.1f ms（n=%d）。", metrics.TPOTAvgMS, metrics.TPOTP50MS, metrics.TPOTP90MS, metrics.TPOTN)
	}
	if metrics.TokensPerSec > 0 {
		summary += fmt.Sprintf(" 整批吞吐: %.1f tok/s；按耗时加权的单请求输出速率: %.1f tok/s。", metrics.TokensPerSec, metrics.RequestTokensPerSec)
	} else if metrics.UsageN != ok {
		summary += fmt.Sprintf(" Token 用量仅 %d/%d 条成功请求完整返回，Token 速率与 TPM 不计算。", metrics.UsageN, ok)
	}
	if metrics.RPM > 0 {
		summary += fmt.Sprintf(" 短测推算 RPM %.0f。", metrics.RPM)
		if metrics.UsageN == ok {
			summary += fmt.Sprintf(" 短测推算 TPM %.0f。", metrics.TPM)
		}
	}
	if bad > 0 {
		summary += fmt.Sprintf(" 失败分为 %d 类", len(metrics.Issues))
		if metrics.OtherIssueCount > 0 {
			summary += fmt.Sprintf("，另有 %d 次失败未单独列出", metrics.OtherIssueCount)
		}
		summary += "；详细原因见下方问题汇总。"
	}
	emit(Event{Type: "metrics", Module: ModuleStress, Completed: metrics.Attempted, Total: total, Metrics: metrics, Summary: summary})
	emit(Event{Type: "summary", Module: ModuleStress, Summary: summary})
}

func looksLikeJSON(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false
	}
	var value map[string]any
	return common.Unmarshal([]byte(trimmed), &value) == nil && value != nil
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	copied := append([]float64(nil), values...)
	sort.Float64s(copied)
	if p <= 0 {
		return copied[0]
	}
	if p >= 100 {
		return copied[len(copied)-1]
	}
	rank := int(math.Ceil(p/100*float64(len(copied)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(copied) {
		rank = len(copied) - 1
	}
	return copied[rank]
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func extractTaskID(raw []byte) string {
	for _, path := range []string{"id", "task_id", "data.id", "data.task_id", "Result.Id", "Result.ID", "Result.TaskId", "Result.TaskID"} {
		if val := gjson.GetBytes(raw, path).String(); strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func ParseVideoTaskState(raw []byte) (status string, videoURL string, failReason string) {
	for _, path := range []string{"status", "task_status", "data.status", "data.task_status", "data.data.status", "Result.Status"} {
		if val := gjson.GetBytes(raw, path).String(); strings.TrimSpace(val) != "" {
			status = strings.TrimSpace(val)
			break
		}
	}
	if status == "" {
		status = "unknown"
	}

	for _, path := range []string{
		"content.video_url",
		"data.content.video_url",
		"data.data.content.video_url",
		"result_url",
		"data.result_url",
		"data.data.result_url",
		"video_url",
		"data.video_url",
	} {
		if val := gjson.GetBytes(raw, path).String(); strings.TrimSpace(val) != "" {
			videoURL = strings.TrimSpace(val)
			break
		}
	}

	for _, path := range []string{
		"error.message",
		"error.code",
		"data.error.message",
		"data.data.error.message",
		"fail_reason",
		"data.fail_reason",
		"data.data.fail_reason",
		"message",
		"data.message",
		"ResponseMetadata.Error.Message",
	} {
		if val := gjson.GetBytes(raw, path).String(); strings.TrimSpace(val) != "" {
			failReason = strings.TrimSpace(val)
			break
		}
	}

	return status, videoURL, failReason
}

func runVideo(ctx context.Context, httpClient *http.Client, req RunRequest, emit Emitter) {
	started := time.Now()
	metrics := &VideoMetrics{}

	runSubmit := true
	runPoll := true
	runResult := true
	if len(req.Video.Checks) > 0 {
		runSubmit = false
		runPoll = false
		runResult = false
		for _, c := range req.Video.Checks {
			switch c {
			case CheckVideoSubmit:
				runSubmit = true
			case CheckVideoPoll:
				runPoll = true
			case CheckVideoResult:
				runResult = true
			}
		}
	}

	var taskID string

	if !runSubmit {
		taskID = strings.TrimSpace(req.Video.TaskID)
		if taskID == "" {
			failedCheckID := CheckVideoPoll
			failedTitle := "状态轮询"
			if !runPoll && runResult {
				failedCheckID = CheckVideoResult
				failedTitle = "视频结果"
			}
			emit(Event{
				Type:    "check",
				Module:  ModuleVideo,
				CheckID: failedCheckID,
				Status:  "fail",
				Title:   failedTitle,
				Message: "未提供任务 ID，无法执行状态查询。请先执行任务提交或输入任务 ID",
				Video:   metrics,
			})
			return
		}
		metrics.TaskID = taskID
	} else {
		emit(Event{
			Type:    "check",
			Module:  ModuleVideo,
			CheckID: CheckVideoSubmit,
			Status:  "running",
			Title:   "任务提交",
			Message: "正在组装请求并向火山方舟提交视频生成任务...",
			Video:   metrics,
		})

		var payloadBytes []byte
		rawPayloadTrimmed := strings.TrimSpace(req.Video.RawPayload)
		if rawPayloadTrimmed != "" {
			var rawMap map[string]any
			if err := common.Unmarshal([]byte(rawPayloadTrimmed), &rawMap); err != nil {
				emit(Event{
					Type:    "check",
					Module:  ModuleVideo,
					CheckID: CheckVideoSubmit,
					Status:  "fail",
					Title:   "任务提交",
					Message: "原生请求 JSON 解析失败: " + err.Error(),
					Video:   metrics,
				})
				return
			}
			if _, hasModel := rawMap["model"]; !hasModel && strings.TrimSpace(req.Model) != "" {
				rawMap["model"] = strings.TrimSpace(req.Model)
			}
			b, err := common.Marshal(rawMap)
			if err != nil {
				emit(Event{
					Type:    "check",
					Module:  ModuleVideo,
					CheckID: CheckVideoSubmit,
					Status:  "fail",
					Title:   "任务提交",
					Message: "序列化原生请求 JSON 失败: " + err.Error(),
					Video:   metrics,
				})
				return
			}
			payloadBytes = b
		} else {
			contentSlice := make([]map[string]any, 0, 3)
			if prompt := strings.TrimSpace(req.Video.Prompt); prompt != "" {
				contentSlice = append(contentSlice, map[string]any{
					"type": "text",
					"text": prompt,
				})
			}

			uploadMode := strings.ToLower(strings.TrimSpace(req.Video.UploadMode))
			if uploadMode == "url" {
				rawURL := strings.TrimSpace(req.Video.ImageURL)
				if rawURL == "" {
					emit(Event{
						Type:    "check",
						Module:  ModuleVideo,
						CheckID: CheckVideoSubmit,
						Status:  "fail",
						Title:   "任务提交",
						Message: "选择了公网 URL 模式但未提供图片地址",
						Video:   metrics,
					})
					return
				}
				role := "reference_image"
				if req.Video.Role != nil && strings.TrimSpace(*req.Video.Role) != "" {
					role = strings.TrimSpace(*req.Video.Role)
				}
				contentSlice = append(contentSlice, map[string]any{
					"type": "image_url",
					"image_url": map[string]any{
						"url": rawURL,
					},
					"role": role,
				})
			} else if uploadMode == "base64" {
				rawB64 := strings.TrimSpace(req.Video.Base64Data)
				if rawB64 == "" {
					emit(Event{
						Type:    "check",
						Module:  ModuleVideo,
						CheckID: CheckVideoSubmit,
						Status:  "fail",
						Title:   "任务提交",
						Message: "选择了 Base64 模式但未提供数据",
						Video:   metrics,
					})
					return
				}
				if !strings.HasPrefix(rawB64, "data:image/") || !strings.Contains(rawB64, ";base64,") {
					emit(Event{
						Type:    "check",
						Module:  ModuleVideo,
						CheckID: CheckVideoSubmit,
						Status:  "fail",
						Title:   "任务提交",
						Message: "Base64 数据格式不规范，需为 data:image/<type>;base64,... 格式",
						Video:   metrics,
					})
					return
				}
				role := "reference_image"
				if req.Video.Role != nil && strings.TrimSpace(*req.Video.Role) != "" {
					role = strings.TrimSpace(*req.Video.Role)
				}
				contentSlice = append(contentSlice, map[string]any{
					"type": "image_url",
					"image_url": map[string]any{
						"url": rawB64,
					},
					"role": role,
				})
			}

			lastFrameMode := strings.ToLower(strings.TrimSpace(req.Video.LastFrameMode))
			if lastFrameMode == "url" {
				lastURL := strings.TrimSpace(req.Video.LastFrameURL)
				if lastURL != "" {
					contentSlice = append(contentSlice, map[string]any{
						"type": "image_url",
						"image_url": map[string]any{
							"url": lastURL,
						},
						"role": "last_frame",
					})
				}
			} else if lastFrameMode == "base64" {
				lastB64 := strings.TrimSpace(req.Video.LastFrameBase64)
				if lastB64 != "" {
					if !strings.HasPrefix(lastB64, "data:image/") || !strings.Contains(lastB64, ";base64,") {
						emit(Event{
							Type:    "check",
							Module:  ModuleVideo,
							CheckID: CheckVideoSubmit,
							Status:  "fail",
							Title:   "任务提交",
							Message: "尾帧 Base64 数据格式不规范，需为 data:image/<type>;base64,... 格式",
							Video:   metrics,
						})
						return
					}
					contentSlice = append(contentSlice, map[string]any{
						"type": "image_url",
						"image_url": map[string]any{
							"url": lastB64,
						},
						"role": "last_frame",
					})
				}
			}

			payloadMap := map[string]any{
				"model": req.Model,
			}
			if len(contentSlice) > 0 {
				payloadMap["content"] = contentSlice
			}
			if req.Video.Resolution != nil && strings.TrimSpace(*req.Video.Resolution) != "" {
				payloadMap["resolution"] = strings.TrimSpace(*req.Video.Resolution)
			}
			if req.Video.Ratio != nil && strings.TrimSpace(*req.Video.Ratio) != "" {
				payloadMap["ratio"] = strings.TrimSpace(*req.Video.Ratio)
			}
			if req.Video.Duration != nil && *req.Video.Duration != 0 {
				payloadMap["duration"] = *req.Video.Duration
			}
			if req.Video.Watermark != nil {
				payloadMap["watermark"] = *req.Video.Watermark
			}
			if req.Video.Seed != nil {
				payloadMap["seed"] = *req.Video.Seed
			}
			if req.Video.GenerateAudio != nil {
				payloadMap["generate_audio"] = *req.Video.GenerateAudio
			}
			if req.Video.ReturnLastFrame != nil {
				payloadMap["return_last_frame"] = *req.Video.ReturnLastFrame
			}
			if strings.TrimSpace(req.Video.CustomJSON) != "" {
				var extra map[string]any
				if err := common.Unmarshal([]byte(req.Video.CustomJSON), &extra); err != nil {
					emit(Event{
						Type:    "check",
						Module:  ModuleVideo,
						CheckID: CheckVideoSubmit,
						Status:  "fail",
						Title:   "任务提交",
						Message: "自定义参数 JSON 解析失败: " + err.Error(),
						Video:   metrics,
					})
					return
				}
				for k, v := range extra {
					payloadMap[k] = v
				}
			}

			b, err := common.Marshal(payloadMap)
			if err != nil {
				emit(Event{
					Type:    "check",
					Module:  ModuleVideo,
					CheckID: CheckVideoSubmit,
					Status:  "fail",
					Title:   "任务提交",
					Message: "构建请求 JSON 失败: " + err.Error(),
					Video:   metrics,
				})
				return
			}
			payloadBytes = b
		}

		metrics.RawRequestJSON = string(payloadBytes)

		endpoint, _ := VideoTasksURL(req.BaseURL, req.Video.CustomPath)
		metrics.EndpointURL = endpoint

		status, submitResp, err := CreateVideoTaskRaw(ctx, httpClient, req.BaseURL, req.Video.CustomPath, req.APIKey, payloadBytes)
		metrics.ElapsedMS = float64(time.Since(started).Milliseconds())
		metrics.RawSubmitResponseJSON = string(submitResp)
		if err != nil {
			emit(Event{
				Type:    "check",
				Module:  ModuleVideo,
				CheckID: CheckVideoSubmit,
				Status:  "fail",
				Title:   "任务提交",
				Message: "提交任务网络请求失败: " + err.Error(),
				Video:   metrics,
			})
			return
		}

		if status != http.StatusOK && status != http.StatusCreated && status != http.StatusAccepted {
			errMsg := ExtractAPIError(submitResp, fmt.Sprintf("HTTP %d", status))
			emit(Event{
				Type:    "check",
				Module:  ModuleVideo,
				CheckID: CheckVideoSubmit,
				Status:  "fail",
				Title:   "任务提交",
				Message: fmt.Sprintf("上游返回错误 (HTTP %d): %s", status, errMsg),
				Video:   metrics,
			})
			return
		}

		parsedID := extractTaskID(submitResp)
		if parsedID == "" {
			emit(Event{
				Type:    "check",
				Module:  ModuleVideo,
				CheckID: CheckVideoSubmit,
				Status:  "fail",
				Title:   "任务提交",
				Message: "提交成功但响应中未返回有效任务 ID",
				Video:   metrics,
			})
			return
		}

		taskID = parsedID
		metrics.TaskID = taskID
		emit(Event{
			Type:    "check",
			Module:  ModuleVideo,
			CheckID: CheckVideoSubmit,
			Status:  "pass",
			Title:   "任务提交",
			Message: fmt.Sprintf("任务提交成功，任务 ID: %s", taskID),
			Video:   metrics,
		})

		if !runPoll && !runResult {
			return
		}
	}

	if runPoll {
		emit(Event{
			Type:    "check",
			Module:  ModuleVideo,
			CheckID: CheckVideoPoll,
			Status:  "running",
			Title:   "状态轮询",
			Message: fmt.Sprintf("任务 [%s] 已进入队列，开始轮询状态...", taskID),
			Video:   metrics,
		})
	} else if runResult {
		emit(Event{
			Type:    "check",
			Module:  ModuleVideo,
			CheckID: CheckVideoResult,
			Status:  "running",
			Title:   "视频结果",
			Message: fmt.Sprintf("正在查询任务 [%s] 的视频生成结果...", taskID),
			Video:   metrics,
		})
	}

	pollCtx, cancelPoll := context.WithTimeout(ctx, videoPollTotalTimeout)
	defer cancelPoll()
	emitPollFailure := func(message string) {
		metrics.ElapsedMS = float64(time.Since(started).Milliseconds())
		if runPoll {
			emit(Event{Type: "check", Module: ModuleVideo, CheckID: CheckVideoPoll, Status: "fail", Title: "状态轮询", Message: message, Video: metrics})
		}
		if runResult {
			emit(Event{Type: "check", Module: ModuleVideo, CheckID: CheckVideoResult, Status: "fail", Title: "视频结果", Message: message, Video: metrics})
		}
	}

	queryTask := func() (done bool) {
		pollStatus, pollResp, pollErr := GetVideoTaskRaw(pollCtx, httpClient, req.BaseURL, req.Video.CustomPath, req.APIKey, taskID)
		metrics.RawPollResponseJSON = string(pollResp)
		elapsed := time.Since(started)
		metrics.ElapsedMS = float64(elapsed.Milliseconds())

		activeCheckID := CheckVideoPoll
		activeTitle := "状态轮询"
		if !runPoll && runResult {
			activeCheckID = CheckVideoResult
			activeTitle = "视频结果"
		}

		if pollErr != nil {
			if pollCtx.Err() != nil {
				return false
			}
			emit(Event{
				Type:    "progress",
				Module:  ModuleVideo,
				CheckID: activeCheckID,
				Status:  "running",
				Title:   activeTitle,
				Message: fmt.Sprintf("轮询异常 (%s)，5 秒后重试...", pollErr.Error()),
				Video:   metrics,
			})
			return false
		}

		if pollStatus != http.StatusOK {
			if pollStatus >= http.StatusBadRequest && pollStatus < http.StatusInternalServerError && pollStatus != http.StatusRequestTimeout && pollStatus != http.StatusTooManyRequests {
				message := fmt.Sprintf("轮询失败 (HTTP %d): %s", pollStatus, ExtractAPIError(pollResp, http.StatusText(pollStatus)))
				emitPollFailure(message)
				return true
			}
			emit(Event{
				Type:    "progress",
				Module:  ModuleVideo,
				CheckID: activeCheckID,
				Status:  "running",
				Title:   activeTitle,
				Message: fmt.Sprintf("轮询返回 HTTP %d，5 秒后重试...", pollStatus),
				Video:   metrics,
			})
			return false
		}

		taskState, videoURL, failReason := ParseVideoTaskState(pollResp)
		metrics.Status = taskState
		metrics.VideoURL = videoURL
		metrics.FailReason = failReason

		normalizedState := strings.ToLower(taskState)
		switch normalizedState {
		case "queued", "pending":
			emit(Event{
				Type:    "progress",
				Module:  ModuleVideo,
				CheckID: activeCheckID,
				Status:  "running",
				Title:   activeTitle,
				Message: fmt.Sprintf("排队中 (queued)，已等待 %.1f 秒...", elapsed.Seconds()),
				Video:   metrics,
			})
			return false
		case "running", "processing", "in_progress":
			emit(Event{
				Type:    "progress",
				Module:  ModuleVideo,
				CheckID: activeCheckID,
				Status:  "running",
				Title:   activeTitle,
				Message: fmt.Sprintf("生成渲染中 (running)，已耗时 %.1f 秒...", elapsed.Seconds()),
				Video:   metrics,
			})
			return false
		case "succeeded", "success":
			if runPoll {
				emit(Event{
					Type:    "check",
					Module:  ModuleVideo,
					CheckID: CheckVideoPoll,
					Status:  "pass",
					Title:   "状态轮询",
					Message: fmt.Sprintf("任务生成完成，总耗时 %.1f 秒", elapsed.Seconds()),
					Video:   metrics,
				})
			}

			if runResult {
				if videoURL != "" {
					emit(Event{
						Type:    "check",
						Module:  ModuleVideo,
						CheckID: CheckVideoResult,
						Status:  "pass",
						Title:   "视频结果",
						Message: fmt.Sprintf("成功获取视频播放地址: %s", videoURL),
						Video:   metrics,
					})
				} else {
					emit(Event{
						Type:    "check",
						Module:  ModuleVideo,
						CheckID: CheckVideoResult,
						Status:  "fail",
						Title:   "视频结果",
						Message: "任务状态成功，但响应中未提取到 video_url",
						Video:   metrics,
					})
				}
			}
			return true
		case "failed", "failure", "cancelled", "expired":
			errMsg := failReason
			if errMsg == "" {
				errMsg = ExtractAPIError(pollResp, "任务执行失败")
			}
			if runPoll {
				emit(Event{
					Type:    "check",
					Module:  ModuleVideo,
					CheckID: CheckVideoPoll,
					Status:  "fail",
					Title:   "状态轮询",
					Message: fmt.Sprintf("任务执行失败 (%s): %s", taskState, errMsg),
					Video:   metrics,
				})
			}
			if runResult {
				emit(Event{
					Type:    "check",
					Module:  ModuleVideo,
					CheckID: CheckVideoResult,
					Status:  "fail",
					Title:   "视频结果",
					Message: fmt.Sprintf("由于任务失败 (%s)，未能生成视频结果: %s", taskState, errMsg),
					Video:   metrics,
				})
			}
			return true
		default:
			emit(Event{
				Type:    "progress",
				Module:  ModuleVideo,
				CheckID: activeCheckID,
				Status:  "running",
				Title:   activeTitle,
				Message: fmt.Sprintf("当前状态: %s，已耗时 %.1f 秒...", taskState, elapsed.Seconds()),
				Video:   metrics,
			})
			return false
		}
	}

	if queryTask() {
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			emitPollFailure("测试已取消或请求已超时")
			return
		case <-pollCtx.Done():
			message := "任务轮询超时（超过 40 分钟），上游未在预期时间内完成"
			if ctx.Err() != nil {
				message = "测试已取消或请求已超时"
			}
			emitPollFailure(message)
			return
		case <-ticker.C:
			if queryTask() {
				return
			}
		}
	}
}
