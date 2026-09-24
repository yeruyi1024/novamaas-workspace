package suppliertest

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/tidwall/gjson"
)

type streamOptionsKey struct{}

func withStreamOptionsTracker(ctx context.Context) context.Context {
	return context.WithValue(ctx, streamOptionsKey{}, &atomic.Bool{})
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type chatRequest struct {
	Model           string           `json:"model"`
	Messages        []chatMessage    `json:"messages,omitempty"`
	Stream          bool             `json:"stream"`
	MaxTokens       *int             `json:"max_tokens,omitempty"`
	Temperature     *float64         `json:"temperature,omitempty"`
	TopP            *float64         `json:"top_p,omitempty"`
	StreamOptions   *streamOptions   `json:"stream_options,omitempty"`
	Tools           []map[string]any `json:"tools,omitempty"`
	ResponseFormat  map[string]any   `json:"response_format,omitempty"`
	Thinking        map[string]any   `json:"thinking,omitempty"`
	ReasoningEffort string           `json:"reasoning_effort,omitempty"`
}

type usageFields struct {
	PromptTokens         *float64 `json:"prompt_tokens"`
	CompletionTokens     *float64 `json:"completion_tokens"`
	TotalTokens          *float64 `json:"total_tokens"`
	InputTokens          *float64 `json:"input_tokens"`
	OutputTokens         *float64 `json:"output_tokens"`
	CachedTokens         *float64 `json:"cached_tokens"`
	PromptCacheHitTokens *float64 `json:"prompt_cache_hit_tokens"`
	CacheReadInputTokens *float64 `json:"cache_read_input_tokens"`
	ReasoningTokens      *float64 `json:"reasoning_tokens"`
	PromptTokensDetails  *struct {
		CachedTokens *float64 `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	InputTokensDetails *struct {
		CachedTokens *float64 `json:"cached_tokens"`
	} `json:"input_tokens_details"`
	CompletionTokensDetails *struct {
		ReasoningTokens *float64 `json:"reasoning_tokens"`
	} `json:"completion_tokens_details"`
}

type streamChunk struct {
	ID        string `json:"id"`
	RequestID string `json:"request_id"`
	Error     *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error"`
	Usage   *usageFields `json:"usage"`
	Choices []struct {
		FinishReason string       `json:"finish_reason"`
		Usage        *usageFields `json:"usage"`
		Delta        struct {
			Content          string `json:"content"`
			Reasoning        string `json:"reasoning"`
			ReasoningContent string `json:"reasoning_content"`
			ToolCalls        []struct {
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		Message *struct {
			Content          string `json:"content"`
			Reasoning        string `json:"reasoning"`
			ReasoningContent string `json:"reasoning_content"`
			ToolCalls        []struct {
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
}

type StreamDelta struct {
	Content   string
	Reasoning string
}

type StreamResult struct {
	StatusCode          int
	Header              http.Header
	ID                  string
	Content             string
	Reasoning           string
	FinishReason        string
	ToolName            string
	ToolArgs            string
	PromptTokens        int
	CompletionTokens    int
	CachedTokens        int
	ReasoningTokens     int
	HasUsage            bool
	HasPromptTokens     bool
	HasCompletionTokens bool
	HasCachedTokens     bool
	HasReasoningTokens  bool
	TTFT                time.Duration
	TPOT                time.Duration
	Elapsed             time.Duration
	lastOutput          time.Duration
	requestStarted      time.Time
	ErrorMessage        string
	SSE                 bool
}

func NewHTTPClient(base *http.Client) *http.Client {
	if base == nil {
		return &http.Client{Timeout: 0}
	}
	clone := *base
	clone.Timeout = 0
	clone.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
		return nil
	}
	return &clone
}

func ChatCompletionsURL(raw string) (string, error) {
	return joinOpenAIPath(raw, "chat/completions")
}

func ModelsURL(raw string) (string, error) {
	return joinOpenAIPath(raw, "models")
}

func VideoTasksURL(raw string, customEndpoint ...string) (string, error) {
	parsed, err := parseSupplierBase(raw)
	if err != nil {
		return "", err
	}
	path := strings.TrimRight(parsed.Path, "/")
	for _, suffix := range []string{"/supplier-test", "/console", "/dashboard"} {
		path = strings.TrimSuffix(path, suffix)
	}
	path = strings.TrimRight(path, "/")

	// If a custom endpoint path is explicitly provided (e.g. from VideoConfig.CustomPath)
	if len(customEndpoint) > 0 && strings.TrimSpace(customEndpoint[0]) != "" {
		custom := strings.TrimSpace(customEndpoint[0])
		if !strings.HasPrefix(custom, "/") {
			custom = "/" + custom
		}
		trimmedPath := strings.TrimRight(path, "/")
		if trimmedPath == "" || trimmedPath == "/" {
			parsed.Path = custom
		} else if strings.HasSuffix(trimmedPath, custom) {
			parsed.Path = trimmedPath
		} else if strings.HasSuffix(trimmedPath, "/api") && strings.HasPrefix(custom, "/api/") {
			parsed.Path = trimmedPath + strings.TrimPrefix(custom, "/api")
		} else if strings.HasSuffix(trimmedPath, "/v1") && strings.HasPrefix(custom, "/v1/") {
			parsed.Path = trimmedPath + strings.TrimPrefix(custom, "/v1")
		} else {
			parsed.Path = trimmedPath + custom
		}
		return parsed.String(), nil
	}

	leaf := "contents/generations/tasks"
	switch {
	// Case 1: User filled in full path ending with /contents/generations/tasks
	case strings.HasSuffix(path, "/"+leaf) || path == leaf || path == "/"+leaf:
		parsed.Path = path
	// Case 2: User filled path ending with /contents/generations
	case strings.HasSuffix(path, "/contents/generations"):
		parsed.Path = path + "/tasks"
	// Case 3: User filled path ending with /contents (e.g. /api/v3/contents or /v1/contents or /contents)
	case strings.HasSuffix(path, "/contents"):
		parsed.Path = path + "/generations/tasks"
	// Case 4: Path ends with /api/v3 or /v3
	case strings.HasSuffix(path, "/api/v3") || strings.HasSuffix(path, "/v3"):
		parsed.Path = path + "/" + leaf
	// Case 5: Path ends with /api (e.g. https://domain.com/api)
	case strings.HasSuffix(path, "/api"):
		parsed.Path = path + "/v3/" + leaf
	// Case 6: Path is root or empty
	case path == "" || path == "/":
		parsed.Path = "/api/v3/" + leaf
	// Case 7: Custom proxy path
	default:
		cleanPath := strings.TrimSuffix(path, "/v1")
		cleanPath = strings.TrimRight(cleanPath, "/")
		if strings.HasSuffix(cleanPath, "/api") {
			parsed.Path = cleanPath + "/v3/" + leaf
		} else if strings.HasSuffix(cleanPath, "/api/v3") || strings.HasSuffix(cleanPath, "/v3") {
			parsed.Path = cleanPath + "/" + leaf
		} else {
			parsed.Path = cleanPath + "/api/v3/" + leaf
		}
	}
	return parsed.String(), nil
}

func VideoTaskGetURL(raw string, taskID string, customEndpoint ...string) (string, error) {
	tasksURL, err := VideoTasksURL(raw, customEndpoint...)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(tasksURL, "/") + "/" + url.PathEscape(taskID), nil
}

func CreateVideoTaskRaw(ctx context.Context, httpClient *http.Client, baseURL, customPath, apiKey string, payload []byte) (int, []byte, error) {
	if httpClient == nil {
		return 0, nil, fmt.Errorf("http client is required")
	}
	endpoint, err := VideoTasksURL(baseURL, customPath)
	if err != nil {
		return 0, nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, videoSubmitTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	rawResp, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, rawResp, nil
}

func GetVideoTaskRaw(ctx context.Context, httpClient *http.Client, baseURL, customPath, apiKey, taskID string) (int, []byte, error) {
	if httpClient == nil {
		return 0, nil, fmt.Errorf("http client is required")
	}
	endpoint, err := VideoTaskGetURL(baseURL, taskID, customPath)
	if err != nil {
		return 0, nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, videoPollRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	rawResp, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, rawResp, nil
}

func parseSupplierBase(raw string) (*url.URL, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("base URL must use http or https")
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("base URL host is required")
	}
	parsed.Fragment = ""
	parsed.RawQuery = ""
	return parsed, nil
}

func joinOpenAIPath(raw, leaf string) (string, error) {
	parsed, err := parseSupplierBase(raw)
	if err != nil {
		return "", err
	}
	path := strings.TrimRight(parsed.Path, "/")
	for _, suffix := range []string{"/supplier-test", "/console", "/dashboard"} {
		path = strings.TrimSuffix(path, suffix)
	}
	path = strings.TrimRight(path, "/")
	switch {
	case strings.HasSuffix(path, "/"+leaf):
		parsed.Path = path
	case strings.HasSuffix(path, "/v1"), strings.HasSuffix(path, "/v4"):
		parsed.Path = path + "/" + leaf
	default:
		parsed.Path = path + "/v1/" + leaf
	}
	return parsed.String(), nil
}

func ListModels(ctx context.Context, httpClient *http.Client, baseURL, apiKey string) ([]string, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("http client is required")
	}
	endpoint, err := ModelsURL(baseURL)
	if err != nil {
		return nil, err
	}
	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(callCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s", ExtractAPIError(raw, fmt.Sprintf("HTTP %d", resp.StatusCode)))
	}
	ids, err := parseModelIDs(raw)
	if err != nil {
		snippet := strings.ToLower(strings.TrimSpace(string(raw)))
		if strings.HasPrefix(snippet, "<!doctype") || strings.HasPrefix(snippet, "<html") {
			return nil, fmt.Errorf("base URL should be the API origin, not the web console")
		}
		return nil, err
	}
	return ids, nil
}

func parseModelIDs(raw []byte) ([]string, error) {
	var payload struct {
		Data   []any `json:"data"`
		Models []any `json:"models"`
	}
	if err := common.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("upstream model list is not JSON")
	}
	items := payload.Data
	if len(items) == 0 {
		items = payload.Models
	}
	ids := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		id := modelIDFromItem(item)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids, nil
}

func modelIDFromItem(item any) string {
	switch value := item.(type) {
	case string:
		return strings.TrimSpace(value)
	case map[string]any:
		for _, key := range []string{"id", "model", "name"} {
			text, _ := value[key].(string)
			text = strings.TrimSpace(text)
			if text != "" {
				return text
			}
		}
	}
	return ""
}

func streamChat(
	ctx context.Context,
	httpClient *http.Client,
	endpoint string,
	apiKey string,
	req chatRequest,
	timeout time.Duration,
	onDelta func(StreamDelta),
) StreamResult {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	tracker, _ := ctx.Value(streamOptionsKey{}).(*atomic.Bool)
	if tracker != nil && tracker.Load() {
		req.StreamOptions = nil
	}

	body, err := common.Marshal(req)
	if err != nil {
		return StreamResult{ErrorMessage: err.Error()}
	}

	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return StreamResult{ErrorMessage: err.Error()}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	accept := "application/json"
	if req.Stream {
		accept = "text/event-stream"
	}
	httpReq.Header.Set("Accept", accept)
	if strings.TrimSpace(apiKey) != "" {
		httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}

	started := time.Now()
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return StreamResult{ErrorMessage: err.Error(), Elapsed: time.Since(started)}
	}
	defer resp.Body.Close()

	result := StreamResult{
		StatusCode:     resp.StatusCode,
		Header:         resp.Header.Clone(),
		Elapsed:        time.Since(started),
		requestStarted: started,
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		result.ErrorMessage = ExtractAPIError(raw, resp.Status)
		if req.StreamOptions != nil && (resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity) && streamOptionsUnsupported(result.ErrorMessage) {
			noOptReq := req
			noOptReq.StreamOptions = nil
			retryResult := streamChat(ctx, httpClient, endpoint, apiKey, noOptReq, timeout, onDelta)
			if retryResult.StatusCode == http.StatusOK && retryResult.ErrorMessage == "" {
				retryResult.Elapsed = time.Since(started)
				if retryResult.TTFT > 0 {
					beforeRetry := retryResult.requestStarted.Sub(started)
					retryResult.TTFT += beforeRetry
					retryResult.lastOutput += beforeRetry
				}
				if tracker != nil {
					tracker.Store(true)
				}
				return retryResult
			}
			result.ErrorMessage += "; retry without stream_options: " + firstNonEmpty(retryResult.ErrorMessage, fmt.Sprintf("HTTP %d", retryResult.StatusCode))
		}
		result.Elapsed = time.Since(started)
		return result
	}

	if strings.Contains(contentType, "text/event-stream") {
		result.SSE = true
		parseSSE(resp.Body, started, &result, onDelta)
		result.Elapsed = time.Since(started)
		finalizeStreamTiming(&result)
		return result
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		result.ErrorMessage = err.Error()
		result.Elapsed = time.Since(started)
		return result
	}
	var parsed streamChunk
	if err := common.Unmarshal(raw, &parsed); err != nil {
		result.ErrorMessage = "invalid JSON response: " + err.Error()
		result.Elapsed = time.Since(started)
		return result
	}
	applyChunk(raw, started, &result, nil)
	// A complete JSON response has no observable first output token.
	result.TTFT = 0
	result.TPOT = 0
	result.lastOutput = 0
	result.Elapsed = time.Since(started)
	return result
}

func finalizeStreamTiming(result *StreamResult) {
	if result.TTFT <= 0 || result.lastOutput <= result.TTFT || result.CompletionTokens <= 1 {
		return
	}
	visibleTokens := result.CompletionTokens
	if result.ReasoningTokens > 0 && result.Reasoning == "" {
		visibleTokens -= result.ReasoningTokens
	}
	if visibleTokens > 1 {
		result.TPOT = (result.lastOutput - result.TTFT) / time.Duration(visibleTokens-1)
	}
}

func parseSSE(body io.Reader, started time.Time, result *StreamResult, onDelta func(StreamDelta)) {
	reader := bufio.NewReaderSize(body, 1<<20)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "data:") {
				payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				if payload == "[DONE]" {
					return
				}
				applyChunk([]byte(payload), started, result, onDelta)
			}
		}
		if err != nil {
			if err != io.EOF && result.ErrorMessage == "" {
				result.ErrorMessage = err.Error()
			}
			return
		}
	}
}

func applyChunk(raw []byte, started time.Time, result *StreamResult, onDelta func(StreamDelta)) {
	if len(strings.TrimSpace(string(raw))) == 0 {
		return
	}
	var chunk streamChunk
	if err := common.Unmarshal(raw, &chunk); err != nil {
		if result.ErrorMessage == "" {
			result.ErrorMessage = "invalid JSON stream chunk: " + err.Error()
		}
		return
	}
	if chunk.ID != "" && result.ID == "" {
		result.ID = chunk.ID
	}
	if result.ID == "" && chunk.RequestID != "" {
		result.ID = chunk.RequestID
	}
	if chunk.Error != nil && chunk.Error.Message != "" && result.ErrorMessage == "" {
		result.ErrorMessage = chunk.Error.Message
	}
	applyUsage(chunk.Usage, result)
	for _, choice := range chunk.Choices {
		if choice.Usage != nil {
			applyUsage(choice.Usage, result)
		}
	}
	delta := StreamDelta{}
	for _, choice := range chunk.Choices {
		if choice.FinishReason != "" {
			result.FinishReason = choice.FinishReason
		}
		content := choice.Delta.Content
		reasoning := choice.Delta.Reasoning
		if reasoning == "" {
			reasoning = choice.Delta.ReasoningContent
		}
		if choice.Message != nil {
			if content == "" {
				content = choice.Message.Content
			}
			if reasoning == "" {
				reasoning = choice.Message.Reasoning
			}
			if reasoning == "" {
				reasoning = choice.Message.ReasoningContent
			}
			if result.ToolName == "" {
				for _, call := range choice.Message.ToolCalls {
					if call.Function.Name != "" {
						result.ToolName = call.Function.Name
						result.ToolArgs += call.Function.Arguments
					}
				}
			}
		}
		for _, call := range choice.Delta.ToolCalls {
			if call.Function.Name != "" {
				result.ToolName = call.Function.Name
			}
			result.ToolArgs += call.Function.Arguments
		}
		if content != "" {
			if result.Content == "" && result.Reasoning == "" && result.TTFT == 0 {
				result.TTFT = time.Since(started)
			}
			result.Content += content
			delta.Content += content
			result.lastOutput = time.Since(started)
		}
		if reasoning != "" {
			if result.Content == "" && result.Reasoning == "" && result.TTFT == 0 {
				result.TTFT = time.Since(started)
			}
			result.Reasoning += reasoning
			delta.Reasoning += reasoning
			result.lastOutput = time.Since(started)
		}
	}
	if (delta.Content != "" || delta.Reasoning != "") && onDelta != nil {
		onDelta(delta)
	}
}

func applyUsage(usage *usageFields, result *StreamResult) {
	if usage == nil {
		return
	}
	result.HasUsage = true
	if usage.PromptTokens != nil {
		result.HasPromptTokens = true
		result.PromptTokens = nonNegativeTokenCount(*usage.PromptTokens)
	} else if usage.InputTokens != nil {
		result.HasPromptTokens = true
		result.PromptTokens = nonNegativeTokenCount(*usage.InputTokens)
	}
	if usage.CompletionTokens != nil {
		result.HasCompletionTokens = true
		result.CompletionTokens = nonNegativeTokenCount(*usage.CompletionTokens)
	} else if usage.OutputTokens != nil {
		result.HasCompletionTokens = true
		result.CompletionTokens = nonNegativeTokenCount(*usage.OutputTokens)
	}

	// 缓存 Token 提取：多路探测，优先取 > 0 的有效值，防止空值或 0 覆盖真实命中数
	cachedCandidates := []*float64{}
	if usage.CachedTokens != nil {
		cachedCandidates = append(cachedCandidates, usage.CachedTokens)
	}
	if usage.PromptTokensDetails != nil && usage.PromptTokensDetails.CachedTokens != nil {
		cachedCandidates = append(cachedCandidates, usage.PromptTokensDetails.CachedTokens)
	}
	if usage.InputTokensDetails != nil && usage.InputTokensDetails.CachedTokens != nil {
		cachedCandidates = append(cachedCandidates, usage.InputTokensDetails.CachedTokens)
	}
	if usage.PromptCacheHitTokens != nil {
		cachedCandidates = append(cachedCandidates, usage.PromptCacheHitTokens)
	}
	if usage.CacheReadInputTokens != nil {
		cachedCandidates = append(cachedCandidates, usage.CacheReadInputTokens)
	}

	foundPositive := false
	for _, cand := range cachedCandidates {
		if cand != nil && *cand > 0 {
			result.CachedTokens = int(*cand)
			result.HasCachedTokens = true
			foundPositive = true
			break
		}
	}
	if !foundPositive && !result.HasCachedTokens && len(cachedCandidates) > 0 {
		for _, cand := range cachedCandidates {
			if cand != nil {
				result.CachedTokens = int(*cand)
				result.HasCachedTokens = true
				break
			}
		}
	}

	// 推理 Token 提取：多路探测
	reasoningCandidates := []*float64{}
	if usage.ReasoningTokens != nil {
		reasoningCandidates = append(reasoningCandidates, usage.ReasoningTokens)
	}
	if usage.CompletionTokensDetails != nil && usage.CompletionTokensDetails.ReasoningTokens != nil {
		reasoningCandidates = append(reasoningCandidates, usage.CompletionTokensDetails.ReasoningTokens)
	}
	for _, cand := range reasoningCandidates {
		if cand != nil && *cand > 0 {
			result.ReasoningTokens = int(*cand)
			result.HasReasoningTokens = true
			break
		}
	}
	if !result.HasReasoningTokens && len(reasoningCandidates) > 0 {
		for _, cand := range reasoningCandidates {
			if cand != nil {
				result.ReasoningTokens = int(*cand)
				result.HasReasoningTokens = true
				break
			}
		}
	}
}

func nonNegativeTokenCount(value float64) int {
	if value <= 0 {
		return 0
	}
	return int(value)
}

func streamOptionsUnsupported(message string) bool {
	lower := strings.ToLower(message)
	return strings.Contains(lower, "stream_options") ||
		strings.Contains(lower, "include_usage") ||
		(strings.Contains(lower, "unknown field") && strings.Contains(lower, "stream")) ||
		(strings.Contains(lower, "unrecognized request argument") && strings.Contains(lower, "stream"))
}

func ExtractAPIError(raw []byte, fallback string) string {
	for _, p := range []struct{ msg, code string }{
		{"error.message", "error.code"},
		{"ResponseMetadata.Error.Message", "ResponseMetadata.Error.Code"},
		{"response_metadata.error.message", "response_metadata.error.code"},
		{"data.error.message", "data.error.code"},
	} {
		msg := strings.TrimSpace(gjson.GetBytes(raw, p.msg).String())
		if msg != "" {
			code := strings.TrimSpace(gjson.GetBytes(raw, p.code).String())
			if code != "" {
				return fmt.Sprintf("[%s] %s", code, msg)
			}
			return msg
		}
	}
	if msg := strings.TrimSpace(gjson.GetBytes(raw, "message").String()); msg != "" {
		return msg
	}
	text := strings.TrimSpace(string(raw))
	if text != "" {
		if len(text) > 300 {
			return text[:300]
		}
		return text
	}
	return fallback
}

func requestIDFrom(result StreamResult) string {
	if result.ID != "" {
		return result.ID
	}
	if result.Header == nil {
		return ""
	}
	keys := []string{
		"X-Request-Id",
		"X-Request-ID",
		"Request-Id",
		"Request-ID",
		"X-Dashscope-Request-Id",
		"Cf-Ray",
	}
	for _, key := range keys {
		if value := strings.TrimSpace(result.Header.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func cacheBustPrefix() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("cache-bust %d\n\n", time.Now().UnixNano())
	}
	return fmt.Sprintf("cache-bust %x\n\n", buf)
}

func ptrInt(v int) *int { return &v }

func ptrBool(v bool) *bool { return &v }

func boolVal(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}
