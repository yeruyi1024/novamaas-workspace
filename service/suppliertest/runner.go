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
}

type StressConfig struct {
	Concurrency int    `json:"concurrency"`
	Rounds      int    `json:"rounds"`
	MaxTokens   int    `json:"max_tokens"`
	Prompt      string `json:"prompt"`
	BreakCache  bool   `json:"break_cache"`
	Stream      *bool  `json:"stream"`
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
}

type StressMetrics struct {
	Total            int     `json:"total"`
	Succeeded        int     `json:"succeeded"`
	Failed           int     `json:"failed"`
	ErrorRate        float64 `json:"error_rate"`
	ElapsedMS        float64 `json:"elapsed_ms"`
	TokensPerSec     float64 `json:"tokens_per_sec"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TTFTAvgMS        float64 `json:"ttft_avg_ms"`
	TTFTP50MS        float64 `json:"ttft_p50_ms"`
	TTFTP90MS        float64 `json:"ttft_p90_ms"`
	TTFTN            int     `json:"ttft_n"`
	TPOTAvgMS        float64 `json:"tpot_avg_ms"`
	TPOTP50MS        float64 `json:"tpot_p50_ms"`
	TPOTP90MS        float64 `json:"tpot_p90_ms"`
	TPOTN            int     `json:"tpot_n"`
	RPM              float64 `json:"rpm"`
	TPM              float64 `json:"tpm"`
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
		if module != ModuleBasic && module != ModuleStress && module != ModuleCache {
			return fmt.Errorf("unknown module %q", module)
		}
		seen[module] = true
		normalized = append(normalized, module)
	}
	if len(normalized) == 0 {
		return fmt.Errorf("select at least one module")
	}
	req.Modules = normalized
	if req.Stress.Concurrency == 0 {
		req.Stress.Concurrency = 10
	}
	if req.Stress.Rounds == 0 {
		req.Stress.Rounds = 1
	}
	if req.Stress.MaxTokens == 0 {
		req.Stress.MaxTokens = 256
	}
	if req.Stress.Concurrency < 1 || req.Stress.Concurrency > maxConcurrency {
		return fmt.Errorf("concurrency must be between 1 and %d", maxConcurrency)
	}
	if req.Stress.Rounds < 1 || req.Stress.Rounds > maxRounds {
		return fmt.Errorf("rounds must be between 1 and %d", maxRounds)
	}
	if req.Stress.MaxTokens < 1 || req.Stress.MaxTokens > maxTokensCap {
		return fmt.Errorf("max_tokens must be between 1 and %d", maxTokensCap)
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
	if req.Basic.MaxTokens < 1 || req.Basic.MaxTokens > maxTokensCap {
		return fmt.Errorf("basic max_tokens must be between 1 and %d", maxTokensCap)
	}
	if req.Basic.Stream == nil {
		req.Basic.Stream = ptrBool(true)
	}
	checks, err := filterBasicChecks(req.Basic.Checks)
	if err != nil {
		return err
	}
	req.Basic.Checks = checks
	if req.Cache.MaxTokens == 0 {
		req.Cache.MaxTokens = 16
	}
	if req.Cache.Rounds == 0 {
		req.Cache.Rounds = 5
	}
	if req.Cache.WaitSeconds < 0 || req.Cache.WaitSeconds > 600 {
		return fmt.Errorf("cache wait must be between 0 and 600 seconds")
	}
	if req.Cache.MaxTokens < 1 || req.Cache.MaxTokens > maxTokensCap {
		return fmt.Errorf("cache max_tokens must be between 1 and %d", maxTokensCap)
	}
	if req.Cache.Rounds < 1 || req.Cache.Rounds > maxCacheRounds {
		return fmt.Errorf("cache rounds must be between 1 and %d", maxCacheRounds)
	}
	if strings.TrimSpace(req.Cache.FollowUp) == "" {
		req.Cache.FollowUp = DefaultCacheFollowUp
	}
	if req.Cache.Stream == nil {
		req.Cache.Stream = ptrBool(true)
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
	endpoint, err := ChatCompletionsURL(req.BaseURL)
	if err != nil {
		return err
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
		connected = streamChat(ctx, httpClient, endpoint, req.APIKey, chat, 60*time.Second, onDelta)
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
				if connected.HasUsage {
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
						"供应商没返回 usage，已跳过",
						vendorTitle(profile.id)+" 文档要求返回 usage，这次没有",
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
				sampled = streamChat(ctx, httpClient, endpoint, req.APIKey, chat, 60*time.Second, nil)
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
		jsonReq.ResponseFormat = map[string]any{"type": "json_object"}
		jsonReq.Messages = []chatMessage{{Role: "user", Content: "Return a JSON object with key ping and value pong."}}
		jsonResult := streamChat(ctx, httpClient, endpoint, req.APIKey, jsonReq, 60*time.Second, nil)
		if jsonResult.StatusCode != http.StatusOK || jsonResult.ErrorMessage != "" {
			emitCheck(CheckJSONMode, "skip", "供应商不接受 JSON 模式："+firstNonEmpty(jsonResult.ErrorMessage, fmt.Sprintf("HTTP %d", jsonResult.StatusCode)))
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
		toolResult := streamChat(ctx, httpClient, endpoint, req.APIKey, toolReq, 60*time.Second, nil)
		if toolResult.StatusCode != http.StatusOK || toolResult.ErrorMessage != "" {
			emitCheck(CheckToolCall, "skip", "供应商不接受工具调用："+firstNonEmpty(toolResult.ErrorMessage, fmt.Sprintf("HTTP %d", toolResult.StatusCode)))
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
		mustThink := thinkingRequired(profile.id)
		thinkReq := applyThinking(chat, profile.id)
		thinking := streamChat(ctx, httpClient, endpoint, req.APIKey, thinkReq, 60*time.Second, nil)
		if (thinking.StatusCode != http.StatusOK || thinking.ErrorMessage != "") && thinkReq.Thinking != nil {
			fallbackReq := chat
			fallbackReq.Messages = []chatMessage{{Role: "user", Content: "What is 17 times 19? Think step by step."}}
			fallbackReq.Thinking = nil
			thinking = streamChat(ctx, httpClient, endpoint, req.APIKey, fallbackReq, 60*time.Second, nil)
		}
		if thinking.StatusCode != http.StatusOK || thinking.ErrorMessage != "" {
			status, message := checkStatus(
				mustThink,
				"供应商不接受 thinking 参数："+firstNonEmpty(thinking.ErrorMessage, fmt.Sprintf("HTTP %d", thinking.StatusCode)),
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
	warm := streamChat(ctx, httpClient, endpoint, req.APIKey, warmReq, 180*time.Second, nil)
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
	probeReq := applyStream(chatRequest{
		Model:     req.Model,
		Messages:  cacheMessages(profile, prefix, followUp, false),
		MaxTokens: ptrInt(req.Cache.MaxTokens),
	}, stream)

	var (
		hitRates     []float64
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
		probe := streamChat(ctx, httpClient, endpoint, req.APIKey, probeReq, 180*time.Second, nil)
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
			probeDetails = append(probeDetails, fmt.Sprintf("第%d轮 %.1f%%", i+1, rate*100))
		} else {
			probeDetails = append(probeDetails, fmt.Sprintf("第%d轮 prompt=%d", i+1, effectivePrompt))
		}
	}

	emitCache := func(avgHit, minHit float64) {
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
			},
		})
	}

	if failedRound == rounds {
		emitCheck(CheckCacheProbe, "Cache probe", "fail", strings.Join(probeDetails, "；"))
		emitCheck(CheckCacheTTL, "Cache TTL", "skip", "探测全部失败，不能判定缓存存活")
		emitCache(0, 0)
		emit(Event{Type: "summary", Module: ModuleCache, Summary: fmt.Sprintf("缓存测试失败：探测 %d 轮全部失败。", rounds)})
		return
	}
	emitCheck(CheckCacheProbe, "Cache probe", "pass", fmt.Sprintf("探测 %d 轮完成。%s", rounds, strings.Join(probeDetails, "，")))

	if !sawCached {
		status, message := checkStatus(
			profile.requireCacheField,
			"供应商没返回 cached_tokens，无法确认真实命中",
			vendorTitle(profile.id)+" 文档要求返回 cached_tokens，这次没有",
		)
		emitCheck(CheckCacheTokens, "Cached tokens", status, message)
		emitCheck(CheckCacheHitRate, "Cache hit rate", "skip", "没有 cached_tokens，算不出命中率")
		emitCheck(CheckCacheTTL, "Cache TTL", "skip", "没有 cached_tokens，不能判定缓存存活")
		emitCache(0, 0)
		emit(Event{
			Type:    "summary",
			Module:  ModuleCache,
			Summary: fmt.Sprintf("缓存测试结束：预热 1 次，探测 %d 轮。供应商没返回 cached_tokens，只能确认请求成功，不能确认缓存命中。", rounds),
		})
		return
	}
	emitCheck(CheckCacheTokens, "Cached tokens", "pass", fmt.Sprintf("最后一轮 cached_tokens=%d", lastCached))
	if len(hitRates) == 0 {
		emitCheck(CheckCacheHitRate, "Cache hit rate", "skip", "prompt_tokens 缺失，算不出命中率")
		emitCheck(CheckCacheTTL, "Cache TTL", "skip", "没有命中率，不能判定缓存存活")
		emitCache(0, 0)
		emit(Event{Type: "summary", Module: ModuleCache, Summary: fmt.Sprintf("缓存测试结束：预热 1 次，探测 %d 轮。有 cached_tokens=%d，但没有 prompt_tokens。", rounds, lastCached)})
		return
	}
	avgHit := average(hitRates)
	minHit := hitRates[0]
	for _, rate := range hitRates[1:] {
		if rate < minHit {
			minHit = rate
		}
	}
	emitCheck(CheckCacheHitRate, "Cache hit rate", "pass", fmt.Sprintf("平均命中率 %.1f%%，最低 %.1f%%（%d 轮，最后一轮 %d/%d）。是否达标看页面尺子。", avgHit*100, minHit*100, len(hitRates), lastCached, lastPrompt))
	if req.Cache.WaitSeconds <= 0 {
		emitCheck(CheckCacheTTL, "Cache TTL", "skip", fmt.Sprintf("等待 %ds，TTL 是否达标看页面尺子。", req.Cache.WaitSeconds))
	} else {
		emitCheck(CheckCacheTTL, "Cache TTL", "pass", fmt.Sprintf("等待 %ds 后平均命中率 %.1f%%。是否达标看页面尺子。", req.Cache.WaitSeconds, avgHit*100))
	}
	emitCache(avgHit, minHit)
	emit(Event{
		Type:    "summary",
		Module:  ModuleCache,
		Summary: fmt.Sprintf("缓存测试结束：预热 1 次，探测 %d 轮。平均命中率 %.1f%%，最低 %.1f%%。", rounds, avgHit*100, minHit*100),
	})
}

func runStress(ctx context.Context, httpClient *http.Client, endpoint string, req RunRequest, emit Emitter) {
	total := req.Stress.Concurrency * req.Stress.Rounds
	emit(Event{Type: "progress", Module: ModuleStress, Completed: 0, Total: total})
	stream := boolVal(req.Stress.Stream, true)

	var (
		completed atomic.Int64
		succeeded atomic.Int64
		failed    atomic.Int64
		inTokens  atomic.Int64
		outTokens atomic.Int64
		ttftMu    sync.Mutex
		ttfts     []float64
		tpots     []float64
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
				result := streamChat(ctx, httpClient, endpoint, req.APIKey, stressReq, 180*time.Second, onDelta)
				ok := result.StatusCode == http.StatusOK && result.ErrorMessage == "" && (result.Content != "" || result.Reasoning != "" || result.FinishReason != "")
				if ok {
					succeeded.Add(1)
					if result.PromptTokens > 0 {
						inTokens.Add(int64(result.PromptTokens))
					}
					if result.CompletionTokens > 0 {
						outTokens.Add(int64(result.CompletionTokens))
					} else if result.Content != "" {
						outTokens.Add(int64(max(1, len([]rune(result.Content))/2)))
					}
					if result.TTFT > 0 {
						ttftMu.Lock()
						ttfts = append(ttfts, float64(result.TTFT)/float64(time.Millisecond))
						if result.TPOT > 0 {
							tpots = append(tpots, float64(result.TPOT)/float64(time.Millisecond))
						}
						ttftMu.Unlock()
					}
				} else {
					failed.Add(1)
					if workerID == 0 {
						emit(Event{
							Type:    "check",
							Module:  ModuleStress,
							CheckID: fmt.Sprintf("worker-%d-%d", workerID, round),
							Title:   "Stress request",
							Status:  "fail",
							Message: firstNonEmpty(result.ErrorMessage, fmt.Sprintf("HTTP %d", result.StatusCode)),
						})
					}
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
		Succeeded:        ok,
		Failed:           bad,
		ElapsedMS:        float64(elapsed.Milliseconds()),
		PromptTokens:     promptTotal,
		CompletionTokens: completionTotal,
		TTFTN:            len(ttfts),
		TPOTN:            len(tpots),
	}
	if total > 0 {
		metrics.ErrorRate = float64(bad) / float64(total)
	}
	if elapsed.Seconds() > 0 {
		metrics.TokensPerSec = float64(completionTotal) / elapsed.Seconds()
	}
	if elapsed.Minutes() > 0 {
		metrics.RPM = float64(ok) / elapsed.Minutes()
		metrics.TPM = float64(promptTotal+completionTotal) / elapsed.Minutes()
	}
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
	summary := fmt.Sprintf("%s压测结束（%s）：%d 并发 × %d 轮，共 %d 次，成功 %d，失败 %d，耗时 %.0f ms。",
		mode, cacheMode, req.Stress.Concurrency, req.Stress.Rounds, metrics.Total, metrics.Succeeded, metrics.Failed, metrics.ElapsedMS)
	if metrics.TTFTAvgMS > 0 {
		summary += fmt.Sprintf(" TTFT首字: 均值 %.0f ms / P50 %.0f ms / P90 %.0f ms（n=%d）。", metrics.TTFTAvgMS, metrics.TTFTP50MS, metrics.TTFTP90MS, metrics.TTFTN)
	}
	if metrics.TPOTAvgMS > 0 {
		summary += fmt.Sprintf(" TPOT每Token: 均值 %.1f ms / P50 %.1f ms / P90 %.1f ms（n=%d）。", metrics.TPOTAvgMS, metrics.TPOTP50MS, metrics.TPOTP90MS, metrics.TPOTN)
	}
	if metrics.TokensPerSec > 0 {
		summary += fmt.Sprintf(" 吞吐: 约 %.1f tok/s。", metrics.TokensPerSec)
	}
	if metrics.RPM > 0 || metrics.TPM > 0 {
		summary += fmt.Sprintf(" 短测推算 RPM %.0f，TPM %.0f。", metrics.RPM, metrics.TPM)
	}
	emit(Event{Type: "metrics", Module: ModuleStress, Completed: total, Total: total, Metrics: metrics, Summary: summary})
	emit(Event{Type: "summary", Module: ModuleStress, Summary: summary})
}

func looksLikeJSON(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false
	}
	var value any
	return common.Unmarshal([]byte(trimmed), &value) == nil
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
