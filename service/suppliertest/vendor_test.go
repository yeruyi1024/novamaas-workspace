package suppliertest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveVendor(t *testing.T) {
	t.Parallel()
	got, err := ResolveVendor("")
	require.NoError(t, err)
	assert.Equal(t, VendorGeneric, got)

	got, err = ResolveVendor("GLM")
	require.NoError(t, err)
	assert.Equal(t, VendorGLM, got)

	got, err = ResolveVendor("kimi")
	require.NoError(t, err)
	assert.Equal(t, VendorKimi, got)

	got, err = ResolveVendor("deepseek")
	require.NoError(t, err)
	assert.Equal(t, VendorDeepSeek, got)

	_, err = ResolveVendor("claude")
	require.Error(t, err)
}

func TestThinkingRequired(t *testing.T) {
	t.Parallel()
	assert.True(t, thinkingRequired(VendorGLM, "glm-5.3-flash"))
	assert.True(t, thinkingRequired(VendorKimi, "kimi-k2.7-code"))
	assert.True(t, thinkingRequired(VendorDeepSeek, "deepseek-v4-pro"))
	assert.True(t, thinkingRequired(VendorDeepSeek, "deepseek-reasoner"))
	assert.False(t, thinkingRequired(VendorGLM, "glm-4-flash"))
	assert.False(t, thinkingRequired(VendorKimi, "moonshot-v1-8k"))
	assert.False(t, thinkingRequired(VendorDeepSeek, "deepseek-chat"))
	assert.False(t, thinkingRequired(VendorGeneric, "glm-5.3"))
}

func TestApplyThinkingVendorBased(t *testing.T) {
	t.Parallel()

	// Kimi K3 只使用 reasoning_effort。
	got := applyThinking(chatRequest{Model: "kimik3"}, VendorKimi)
	assert.Nil(t, got.Thinking)
	assert.Equal(t, "low", got.ReasoningEffort)

	// Kimi K2.7 始终思考，不发送额外控制字段。
	got = applyThinking(chatRequest{Model: "kimi-k2.7-code"}, VendorKimi)
	assert.Nil(t, got.Thinking)
	assert.Empty(t, got.ReasoningEffort)

	// Kimi K2.6 通过 thinking.type 显式开启，不支持 reasoning_effort。
	got = applyThinking(chatRequest{Model: "kimi-k2.6"}, VendorKimi)
	require.NotNil(t, got.Thinking)
	assert.Equal(t, "enabled", got.Thinking["type"])
	assert.Empty(t, got.ReasoningEffort)

	// GLM 5.3 使用 thinking enabled，并用 low 档控制测试耗时。
	got = applyThinking(chatRequest{Model: "glm-5.3-flash"}, VendorGLM)
	require.NotNil(t, got.Thinking)
	assert.Equal(t, "enabled", got.Thinking["type"])
	assert.Equal(t, "low", got.ReasoningEffort)

	// DeepSeek V4 支持显式 thinking 与 reasoning_effort。
	got = applyThinking(chatRequest{Model: "deepseek-v4-pro"}, VendorDeepSeek)
	require.NotNil(t, got.Thinking)
	assert.Equal(t, "enabled", got.Thinking["type"])
	assert.Equal(t, "low", got.ReasoningEffort)

	// 旧版 DeepSeek reasoner 本身就是思考模型，兼容请求不附加控制字段。
	got = applyThinking(chatRequest{Model: "deepseek-reasoner"}, VendorDeepSeek)
	assert.Nil(t, got.Thinking)
	assert.Empty(t, got.ReasoningEffort)
}

func TestRequiredVendorCapabilityRejectionFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"message":"unsupported field"}}`)
	}))
	defer server.Close()

	for _, checkID := range []string{CheckJSONMode, CheckToolCall} {
		checkID := checkID
		t.Run(checkID, func(t *testing.T) {
			statusFor := func(vendor string) string {
				var status string
				err := Run(context.Background(), server.Client(), RunRequest{
					BaseURL: server.URL,
					APIKey:  "any",
					Model:   "vendor-model",
					Vendor:  vendor,
					Modules: []string{ModuleBasic},
					Basic:   BasicConfig{Checks: []string{checkID}},
				}, func(event Event) {
					if event.Type == "check" && event.CheckID == checkID && event.Status != "running" {
						status = event.Status
					}
				})
				require.NoError(t, err)
				return status
			}

			assert.Equal(t, "fail", statusFor(VendorGLM))
			assert.Equal(t, "skip", statusFor(VendorGeneric))
		})
	}
}

func TestCacheMessagesSystemPrefix(t *testing.T) {
	t.Parallel()
	profile := profileFor(VendorGLM)
	warm := cacheMessages(profile, "long-prefix", "pong", true)
	require.Len(t, warm, 2)
	assert.Equal(t, "system", warm[0].Role)
	assert.Equal(t, "long-prefix", warm[0].Content)

	generic := cacheMessages(profileFor(VendorGeneric), "long-prefix", "pong", true)
	require.Len(t, generic, 2)
	assert.Equal(t, "system", generic[0].Role)
}

func TestJoinOpenAIPathV4(t *testing.T) {
	t.Parallel()
	got, err := ChatCompletionsURL("https://open.bigmodel.cn/api/paas/v4")
	require.NoError(t, err)
	assert.Equal(t, "https://open.bigmodel.cn/api/paas/v4/chat/completions", got)
}

func TestGLMThinkingMissingReasoningFails(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"id\":\"c1\",\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":8,\"completion_tokens\":2}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "glm-5.3",
		Vendor:  VendorGLM,
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Checks: []string{CheckThinking}},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)
	statusByCheck := map[string]string{}
	for _, event := range events {
		if event.Type == "check" && event.CheckID != "" {
			statusByCheck[event.CheckID] = event.Status
		}
	}
	assert.Equal(t, "fail", statusByCheck[CheckThinking])
}

func TestKimiK3ThinkingSendsReasoningEffort(t *testing.T) {
	t.Parallel()
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"step\",\"content\":\"323\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":8,\"completion_tokens\":4,\"cached_tokens\":0}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "kimi-k3",
		Vendor:  VendorKimi,
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Checks: []string{CheckThinking}},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)
	assert.Contains(t, body, `"reasoning_effort":"low"`)
	assert.NotContains(t, body, `"thinking"`)
	statusByCheck := map[string]string{}
	for _, event := range events {
		if event.Type == "check" && event.CheckID != "" {
			statusByCheck[event.CheckID] = event.Status
		}
	}
	assert.Equal(t, "pass", statusByCheck[CheckThinking])
}

func TestGLMCacheUsesSystemPrefix(t *testing.T) {
	t.Parallel()
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if body == "" {
			body = string(raw)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":20,\"completion_tokens\":1}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "glm-5.3",
		Vendor:  VendorGLM,
		Modules: []string{ModuleCache},
		Cache:   CacheConfig{Prompt: "glm-prefix", FollowUp: "pong", WaitSeconds: 0, MaxTokens: 8, Rounds: 1},
	}, func(Event) {})
	require.NoError(t, err)
	assert.Contains(t, body, `"role":"system"`)
	assert.Contains(t, body, "glm-prefix")
}

func TestThinkingFallbackWhenRejected(t *testing.T) {
	t.Parallel()

	var attempts []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		attempts = append(attempts, string(body))
		if strings.Contains(string(body), `"thinking"`) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":{"message":"Unrecognized request argument: thinking"}}`)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"step 1\",\"content\":\"323\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":15,\"completion_tokens\":10,\"completion_tokens_details\":{\"reasoning_tokens\":8}}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "key",
		Model:   "custom-relay-model",
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Checks: []string{CheckThinking}},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)
	require.Len(t, attempts, 2)
	assert.Contains(t, attempts[0], `"thinking"`)
	assert.NotContains(t, attempts[1], `"thinking"`)

	var thinkingStatus, thinkingMsg string
	for _, event := range events {
		if event.Type == "check" && event.CheckID == CheckThinking {
			thinkingStatus = event.Status
			thinkingMsg = event.Message
		}
	}
	assert.Equal(t, "pass", thinkingStatus)
	assert.Contains(t, thinkingMsg, "reasoning")
}

func TestStreamOptionsFallbackWhenRejected(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		if strings.Contains(string(body), `"stream_options"`) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":{"message":"Unrecognized request argument: stream_options"}}`)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"},\"finish_reason\":\"stop\"}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "key",
		Model:   "strict-proxy-model",
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Checks: []string{CheckConnectivity}},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, calls, 2)

	var connStatus string
	for _, event := range events {
		if event.Type == "check" && event.CheckID == CheckConnectivity {
			connStatus = event.Status
		}
	}
	assert.Equal(t, "pass", connStatus)
}

func TestApplyUsageCachedTokensExtraction(t *testing.T) {
	t.Parallel()

	// 1. DeepSeek 风格：prompt_tokens_details.cached_tokens 虽为 0，但 prompt_cache_hit_tokens 为 2500，应正确提取 2500
	var res StreamResult
	zero := 0.0
	deepseekHit := 2500.0
	promptTokens := 3000.0
	applyUsage(&usageFields{
		PromptTokens:         &promptTokens,
		PromptCacheHitTokens: &deepseekHit,
		PromptTokensDetails: &struct {
			CachedTokens *float64 `json:"cached_tokens"`
		}{
			CachedTokens: &zero,
		},
	}, &res)
	assert.True(t, res.HasCachedTokens)
	assert.Equal(t, 2500, res.CachedTokens)
	assert.Equal(t, 3000, res.PromptTokens)

	// 2. 后续 Chunk 发送 cached_tokens=0，不应覆盖已捕获的有效值 2500
	applyUsage(&usageFields{
		CachedTokens: &zero,
	}, &res)
	assert.True(t, res.HasCachedTokens)
	assert.Equal(t, 2500, res.CachedTokens)

	// 3. Moonshot / Kimi 风格：流式 choices[0].usage
	var moonshotRes StreamResult
	moonshotChunk := []byte(`{"choices":[{"delta":{"content":"ok"},"usage":{"prompt_tokens":1200,"cached_tokens":1024}}]}`)
	applyChunk(moonshotChunk, time.Now(), &moonshotRes, nil)
	assert.True(t, moonshotRes.HasCachedTokens)
	assert.Equal(t, 1024, moonshotRes.CachedTokens)
	assert.Equal(t, 1200, moonshotRes.PromptTokens)

	// 4. InputTokensDetails 风格
	var inputDetailsRes StreamResult
	inputDetailsHit := 888.0
	applyUsage(&usageFields{
		InputTokensDetails: &struct {
			CachedTokens *float64 `json:"cached_tokens"`
		}{
			CachedTokens: &inputDetailsHit,
		},
	}, &inputDetailsRes)
	assert.True(t, inputDetailsRes.HasCachedTokens)
	assert.Equal(t, 888, inputDetailsRes.CachedTokens)
}

func TestApplyUsageTracksZeroTokenFieldsAsPresent(t *testing.T) {
	t.Parallel()

	var result StreamResult
	zero := 0.0
	applyUsage(&usageFields{PromptTokens: &zero, CompletionTokens: &zero}, &result)

	assert.True(t, result.HasUsage)
	assert.True(t, result.HasPromptTokens)
	assert.True(t, result.HasCompletionTokens)
	assert.Zero(t, result.PromptTokens)
	assert.Zero(t, result.CompletionTokens)
}

func TestStreamOptionsDoesNotRetryUnrelatedBadRequest(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":{"message":"invalid response_format"}}`)
	}))
	defer server.Close()

	result := streamChat(context.Background(), server.Client(), server.URL, "key", chatRequest{
		Model:         "demo",
		Stream:        true,
		StreamOptions: &streamOptions{IncludeUsage: true},
	}, 5*time.Second, nil)

	assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	assert.Equal(t, 1, calls)
	assert.Contains(t, result.ErrorMessage, "invalid response_format")
}
