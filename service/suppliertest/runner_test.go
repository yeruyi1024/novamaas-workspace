package suppliertest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelsURL(t *testing.T) {
	t.Parallel()
	got, err := ModelsURL("https://api.example.com")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/v1/models", got)

	got, err = ModelsURL("https://api.example.com/v1")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/v1/models", got)

	got, err = ModelsURL("https://api.example.com/supplier-test")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/v1/models", got)
}

func TestParseModelIDs(t *testing.T) {
	t.Parallel()
	ids, err := parseModelIDs([]byte(`{"data":[{"id":"alpha"},{"id":"beta"},{"id":"alpha"}]}`))
	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "beta"}, ids)

	ids, err = parseModelIDs([]byte(`{"models":["one","two"]}`))
	require.NoError(t, err)
	assert.Equal(t, []string{"one", "two"}, ids)
}

func TestListModelsAgainstFakeUpstream(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/models", r.URL.Path)
		assert.Equal(t, "Bearer good-key", r.Header.Get("Authorization"))
		_, _ = io.WriteString(w, `{"data":[{"id":"vendor-a"},{"id":"vendor-b"}]}`)
	}))
	defer server.Close()

	ids, err := ListModels(context.Background(), server.Client(), server.URL, "good-key")
	require.NoError(t, err)
	assert.Equal(t, []string{"vendor-a", "vendor-b"}, ids)
}

func TestListModelsAllowsEmptyList(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"data":[]}`)
	}))
	defer server.Close()

	ids, err := ListModels(context.Background(), server.Client(), server.URL, "good-key")
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestListModelsRejectsConsoleHTML(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "<!doctype html><html><body>console</body></html>")
	}))
	defer server.Close()

	_, err := ListModels(context.Background(), server.Client(), server.URL, "good-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API origin")
}

func TestChatCompletionsURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "host only", input: "https://api.example.com", want: "https://api.example.com/v1/chat/completions"},
		{name: "v1 suffix", input: "https://api.example.com/v1", want: "https://api.example.com/v1/chat/completions"},
		{name: "already complete", input: "https://api.example.com/v1/chat/completions", want: "https://api.example.com/v1/chat/completions"},
		{name: "trailing slash", input: "https://api.example.com/v1/", want: "https://api.example.com/v1/chat/completions"},
		{name: "rejects ftp", input: "ftp://api.example.com", wantErr: true},
		{name: "rejects empty", input: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ChatCompletionsURL(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPercentileAndAverage(t *testing.T) {
	t.Parallel()

	values := []float64{10, 20, 30, 40, 50}
	assert.Equal(t, 30.0, average(values))
	assert.Equal(t, 50.0, percentile(values, 90))
	assert.Equal(t, 30.0, percentile(values, 50))
	assert.Equal(t, 10.0, percentile(values, 0))
	assert.Equal(t, 50.0, percentile(values, 100))
	assert.Equal(t, 0.0, percentile(nil, 90))

	single := []float64{42.5}
	assert.Equal(t, 42.5, average(single))
	assert.Equal(t, 42.5, percentile(single, 50))
	assert.Equal(t, 42.5, percentile(single, 90))

	even := []float64{10, 20, 30, 40}
	assert.Equal(t, 25.0, average(even))
	assert.Equal(t, 20.0, percentile(even, 50))
	assert.Equal(t, 40.0, percentile(even, 90))
}

func TestNormalizeRunRequest(t *testing.T) {
	t.Parallel()

	req := RunRequest{
		BaseURL: "https://api.example.com",
		Model:   "demo-model",
		Modules: []string{ModuleStress},
		Stress:  StressConfig{Concurrency: 3, Rounds: 2, MaxTokens: 16},
	}
	require.NoError(t, NormalizeRunRequest(&req))
	assert.Equal(t, DefaultStressPrompt, req.Stress.Prompt)
	assert.Equal(t, allBasicChecks, req.Basic.Checks)
	assert.Nil(t, req.Basic.Temperature)
	assert.Nil(t, req.Basic.TopP)
	assert.Equal(t, 5, req.Cache.Rounds)
	assert.Equal(t, DefaultCacheFollowUp, req.Cache.FollowUp)
	assert.False(t, req.Stress.BreakCache)

	unknownCheck := RunRequest{
		BaseURL: "https://api.example.com",
		Model:   "demo",
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Checks: []string{"not-a-check"}},
	}
	require.Error(t, NormalizeRunRequest(&unknownCheck))

	oneCheck := RunRequest{
		BaseURL: "https://api.example.com",
		Model:   "demo",
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Checks: []string{CheckJSONMode, CheckJSONMode}},
	}
	require.NoError(t, NormalizeRunRequest(&oneCheck))
	assert.Equal(t, []string{CheckJSONMode}, oneCheck.Basic.Checks)

	bad := RunRequest{BaseURL: "https://api.example.com", Model: "demo", Modules: []string{"baseline"}}
	require.Error(t, NormalizeRunRequest(&bad))

	tooWide := RunRequest{
		BaseURL: "https://api.example.com",
		Model:   "demo",
		Modules: []string{ModuleStress},
		Stress:  StressConfig{Concurrency: 1001, Rounds: 1, MaxTokens: 16},
	}
	require.Error(t, NormalizeRunRequest(&tooWide))

	tooMany := RunRequest{
		BaseURL: "https://api.example.com",
		Model:   "demo",
		Modules: []string{ModuleStress},
		Stress:  StressConfig{Concurrency: 101, Rounds: 100, MaxTokens: 16},
	}
	err := NormalizeRunRequest(&tooMany)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at most 10000 requests")

	ignoredModuleConfig := RunRequest{
		BaseURL: "https://api.example.com",
		Model:   "demo",
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{MaxTokens: 8},
		Stress:  StressConfig{Concurrency: maxConcurrency + 1, Rounds: 0, MaxTokens: maxTokensCap + 1},
		Cache:   CacheConfig{WaitSeconds: -1, Rounds: maxCacheRounds + 1, Mode: "invalid"},
	}
	require.NoError(t, NormalizeRunRequest(&ignoredModuleConfig))
}

func TestChatRequestOmitsEmptySampling(t *testing.T) {
	t.Parallel()

	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodies = append(bodies, string(raw))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"chatcmpl-1","choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`)
	}))
	defer server.Close()

	result := streamChat(context.Background(), server.Client(), server.URL+"/v1/chat/completions", "k", chatRequest{
		Model:     "demo",
		Messages:  []chatMessage{{Role: "user", Content: "hi"}},
		MaxTokens: ptrInt(8),
	}, 5*time.Second, nil)
	require.Equal(t, http.StatusOK, result.StatusCode)
	require.Len(t, bodies, 1)
	assert.NotContains(t, bodies[0], `"temperature"`)
	assert.NotContains(t, bodies[0], `"top_p"`)
	assert.Contains(t, bodies[0], `"max_tokens"`)
}

func TestRunBasicOmitsUnsetSampling(t *testing.T) {
	t.Parallel()

	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodies = append(bodies, string(raw))
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "k",
		Model:   "demo",
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Prompt: "hi", MaxTokens: 8, Checks: []string{CheckConnectivity}},
	}, func(Event) {})
	require.NoError(t, err)
	require.NotEmpty(t, bodies)
	assert.NotContains(t, bodies[0], `"temperature"`)
	assert.NotContains(t, bodies[0], `"top_p"`)
}

func TestRunBasicOnlyStreamsConnectivity(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"{\\\"ping\\\":\\\"pong\\\"}\"},\"finish_reason\":\"stop\"}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var sawStream bool
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "k",
		Model:   "demo",
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Prompt: "hi", MaxTokens: 8, Checks: []string{CheckJSONMode}},
	}, func(event Event) {
		if event.Type == "stream" {
			sawStream = true
		}
	})
	require.NoError(t, err)
	assert.False(t, sawStream)
}

func TestStreamChatParsesSSE(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("X-Request-Id", "req-123")
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)
		_, _ = io.WriteString(w, "data: {\"id\":\"chatcmpl-1\",\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n\n")
		flusher.Flush()
		time.Sleep(20 * time.Millisecond)
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":8,\"completion_tokens\":2,\"cached_tokens\":3}}\n\n")
		flusher.Flush()
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var streamed strings.Builder
	result := streamChat(context.Background(), server.Client(), server.URL+"/v1/chat/completions", "test-key", chatRequest{
		Model:  "demo",
		Stream: true,
	}, 5*time.Second, func(delta StreamDelta) {
		streamed.WriteString(delta.Content)
	})

	require.Equal(t, http.StatusOK, result.StatusCode)
	assert.True(t, result.SSE)
	assert.Equal(t, "hello", result.Content)
	assert.Equal(t, "hello", streamed.String())
	assert.Equal(t, "stop", result.FinishReason)
	assert.Equal(t, 8, result.PromptTokens)
	assert.Equal(t, 2, result.CompletionTokens)
	assert.True(t, result.HasCachedTokens)
	assert.Equal(t, 3, result.CachedTokens)
	assert.Equal(t, "chatcmpl-1", requestIDFrom(result))
	assert.Equal(t, "req-123", result.Header.Get("X-Request-Id"))
	assert.Greater(t, result.TTFT, time.Duration(0))
}

func TestStreamChatJSONDoesNotReportFirstTokenTiming(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"pong"},"finish_reason":"stop"}],"usage":{"prompt_tokens":31,"completion_tokens":64}}`)
	}))
	defer server.Close()

	result := streamChat(context.Background(), server.Client(), server.URL, "key", chatRequest{
		Model: "glm-5.3-flash", Stream: true,
	}, 5*time.Second, nil)
	require.Equal(t, http.StatusOK, result.StatusCode)
	assert.False(t, result.SSE)
	assert.Equal(t, "pong", result.Content)
	assert.Positive(t, result.Elapsed)
	assert.Zero(t, result.TTFT)
	assert.Zero(t, result.TPOT)
}

func TestStreamChatReportsMalformedSSEChunk(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {broken-json}\n\ndata: [DONE]\n\n")
	}))
	defer server.Close()

	result := streamChat(context.Background(), server.Client(), server.URL, "key", chatRequest{
		Model: "glm-5.3-flash", Stream: true,
	}, 5*time.Second, nil)
	require.Equal(t, http.StatusOK, result.StatusCode)
	assert.Contains(t, result.ErrorMessage, "invalid JSON stream chunk")
}

func TestFinalizeStreamTimingExcludesUsageTailAndUnstreamedReasoning(t *testing.T) {
	t.Parallel()
	result := StreamResult{
		TTFT:             55 * time.Second,
		lastOutput:       56 * time.Second,
		Elapsed:          60 * time.Second,
		CompletionTokens: 64,
		ReasoningTokens:  14,
	}
	finalizeStreamTiming(&result)
	assert.Equal(t, time.Second/49, result.TPOT)

	result.Reasoning = "streamed reasoning"
	result.TPOT = 0
	finalizeStreamTiming(&result)
	assert.Equal(t, time.Second/63, result.TPOT)
}

func TestRunStressReportsRequestDurationWithoutInventingTokens(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"pong"},"finish_reason":"stop"}]}`)
	}))
	defer server.Close()

	var metrics *StressMetrics
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL, Model: "glm-5.3-flash", Modules: []string{ModuleStress},
		Stress: StressConfig{Concurrency: 1, Rounds: 1, MaxTokens: 64},
	}, func(event Event) {
		if event.Type == "metrics" {
			metrics = event.Metrics
		}
	})
	require.NoError(t, err)
	require.NotNil(t, metrics)
	assert.Equal(t, 1, metrics.Attempted)
	assert.Equal(t, 1, metrics.Succeeded)
	assert.Positive(t, metrics.RequestAvgMS)
	assert.Zero(t, metrics.TTFTN)
	assert.Zero(t, metrics.TPOTN)
	assert.Zero(t, metrics.UsageN)
	assert.Zero(t, metrics.CompletionTokens)
	assert.Zero(t, metrics.TokensPerSec)
	assert.Zero(t, metrics.RequestTokensPerSec)
	assert.Zero(t, metrics.TPM)
}

func TestRunStressSummarizesFailuresFromEveryWorker(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"error":{"message":"rate limit exceeded"}}`)
	}))
	defer server.Close()

	var metrics *StressMetrics
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL, Model: "glm-5.3-flash", Modules: []string{ModuleStress},
		Stress: StressConfig{Concurrency: 2, Rounds: 2, MaxTokens: 64},
	}, func(event Event) {
		if event.Type == "metrics" {
			metrics = event.Metrics
		}
	})
	require.NoError(t, err)
	require.NotNil(t, metrics)
	assert.Equal(t, 4, metrics.Attempted)
	assert.Equal(t, 4, metrics.Failed)
	assert.Greater(t, metrics.RequestAvgMS, 0.0)
	assert.Equal(t, 1.0, metrics.ErrorRate)
	require.Len(t, metrics.Issues, 1)
	assert.Equal(t, 4, metrics.Issues[0].Count)
	assert.Equal(t, http.StatusTooManyRequests, metrics.Issues[0].StatusCode)
	assert.Contains(t, metrics.Issues[0].Message, "rate limit exceeded")
	assert.Contains(t, []int{1, 2}, metrics.Issues[0].Worker)
	assert.Contains(t, []int{1, 2}, metrics.Issues[0].Round)
	assert.Zero(t, metrics.OtherIssueCount)
}

func TestRunStressReportsMalformedJSONResponse(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":`)
	}))
	defer server.Close()

	var metrics *StressMetrics
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL, Model: "glm-5.3-flash", Modules: []string{ModuleStress},
		Stress: StressConfig{Concurrency: 1, Rounds: 1, MaxTokens: 64},
	}, func(event Event) {
		if event.Type == "metrics" {
			metrics = event.Metrics
		}
	})
	require.NoError(t, err)
	require.NotNil(t, metrics)
	assert.Equal(t, 1, metrics.Failed)
	require.Len(t, metrics.Issues, 1)
	assert.Equal(t, http.StatusOK, metrics.Issues[0].StatusCode)
	assert.Contains(t, metrics.Issues[0].Message, "invalid JSON response")
}

func TestRunBasicAndStressAgainstFakeUpstream(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer good-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `{"error":{"message":"invalid api key"}}`)
			return
		}
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		if !strings.Contains(string(body), `"messages"`) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":{"message":"messages required"}}`)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("X-Request-Id", "req-basic")
		if strings.Contains(string(body), "get_weather") {
			_, _ = fmt.Fprintf(w, "data: {\"id\":\"chatcmpl-tool\",\"choices\":[{\"delta\":{\"tool_calls\":[{\"function\":{\"name\":\"get_weather\",\"arguments\":\"{\\\"city\\\":\\\"Beijing\\\"}\"}}]},\"finish_reason\":\"tool_calls\"}],\"usage\":{\"prompt_tokens\":12,\"completion_tokens\":8}}\n\n")
			_, _ = io.WriteString(w, "data: [DONE]\n\n")
			return
		}
		if strings.Contains(string(body), "json_object") {
			_, _ = io.WriteString(w, "data: {\"id\":\"chatcmpl-json\",\"choices\":[{\"delta\":{\"content\":\"{\\\"ping\\\":\\\"pong\\\"}\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":6}}\n\n")
			_, _ = io.WriteString(w, "data: [DONE]\n\n")
			return
		}
		_, _ = io.WriteString(w, "data: {\"id\":\"chatcmpl-ok\",\"choices\":[{\"delta\":{\"content\":\"pong\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":9,\"completion_tokens\":1}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "good-key",
		Model:   "demo-model",
		Modules: []string{ModuleBasic, ModuleStress},
		Stress: StressConfig{
			Concurrency: 2,
			Rounds:      1,
			MaxTokens:   16,
			Prompt:      "hello",
		},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)

	statusByCheck := map[string]string{}
	var metrics *StressMetrics
	var sawStream bool
	var sawDone bool
	for _, event := range events {
		if event.Type == "check" && event.CheckID != "" {
			statusByCheck[event.CheckID] = event.Status
		}
		if event.Type == "stream" && event.Text != "" {
			sawStream = true
		}
		if event.Type == "metrics" {
			metrics = event.Metrics
		}
		if event.Type == "done" {
			sawDone = true
		}
	}

	assert.Equal(t, "pass", statusByCheck[CheckConnectivity])
	assert.Equal(t, "pass", statusByCheck[CheckUsage])
	assert.Equal(t, "pass", statusByCheck[CheckJSONMode])
	assert.Equal(t, "pass", statusByCheck[CheckToolCall])
	assert.Equal(t, "pass", statusByCheck[CheckAuthError])
	assert.Equal(t, "pass", statusByCheck[CheckBadRequest])
	assert.Equal(t, "skip", statusByCheck[CheckThinking])
	require.NotNil(t, metrics)
	assert.Equal(t, 2, metrics.Total)
	assert.Equal(t, 2, metrics.Succeeded)
	assert.Equal(t, 0, metrics.Failed)
	assert.Equal(t, 0.0, metrics.ErrorRate)
	assert.Greater(t, metrics.TTFTAvgMS, 0.0)
	assert.Greater(t, metrics.TTFTP50MS, 0.0)
	assert.Greater(t, metrics.TTFTP90MS, 0.0)
	assert.Equal(t, 18, metrics.PromptTokens)
	assert.Equal(t, 2, metrics.CompletionTokens)
	assert.GreaterOrEqual(t, metrics.TTFTN, 1)
	assert.Greater(t, metrics.RPM, 0.0)
	assert.Greater(t, metrics.TPM, 0.0)
	assert.True(t, sawStream)
	assert.True(t, sawDone)
}

func TestRunBasicJSONModeUsesNonStreamingContract(t *testing.T) {
	t.Parallel()

	var (
		requestMu   sync.Mutex
		requestBody []byte
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		if strings.Contains(string(body), `"response_format":{"type":"json_object"}`) {
			requestMu.Lock()
			requestBody = append([]byte(nil), body...)
			requestMu.Unlock()
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"json","choices":[{"message":{"content":"{\"ping\":\"pong\"}"},"finish_reason":"stop"}]}`)
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "good-key",
		Model:   "kimi-k3",
		Vendor:  VendorKimi,
		Modules: []string{ModuleBasic},
		Basic: BasicConfig{
			Checks: []string{CheckJSONMode},
			Stream: ptrBool(true),
		},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)

	status := ""
	for _, event := range events {
		if event.Type == "check" && event.CheckID == CheckJSONMode && event.Status != "running" {
			status = event.Status
		}
	}
	assert.Equal(t, "pass", status)

	requestMu.Lock()
	body := string(requestBody)
	requestMu.Unlock()
	assert.Contains(t, body, `"stream":false`)
	assert.NotContains(t, body, `"stream_options"`)
}

func TestRunStressSendsCorpusAsIs(t *testing.T) {
	t.Parallel()

	var receivedPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		receivedPrompt = string(body)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"id\":\"c-1\",\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":8,\"completion_tokens\":1}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "key",
		Model:   "demo",
		Modules: []string{ModuleStress},
		Stress: StressConfig{
			Concurrency: 1,
			Rounds:      1,
			MaxTokens:   16,
			Prompt:      "custom-stress-prompt",
		},
	}, func(Event) {})
	require.NoError(t, err)
	assert.Contains(t, receivedPrompt, "custom-stress-prompt")
	assert.NotContains(t, receivedPrompt, "cache-bust")
	assert.NotContains(t, receivedPrompt, "alpha beta gamma")
}

func TestRunStressBreakCachePrefixesDiffer(t *testing.T) {
	t.Parallel()

	var prompts []string
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		mu.Lock()
		prompts = append(prompts, string(body))
		mu.Unlock()
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "key",
		Model:   "demo",
		Modules: []string{ModuleStress},
		Stress: StressConfig{
			Concurrency: 2,
			Rounds:      1,
			MaxTokens:   8,
			Prompt:      "shared-corpus",
			BreakCache:  true,
		},
	}, func(Event) {})
	require.NoError(t, err)
	require.Len(t, prompts, 2)
	assert.Contains(t, prompts[0], "shared-corpus")
	assert.Contains(t, prompts[1], "shared-corpus")
	assert.Contains(t, prompts[0], "cache-bust")
	assert.Contains(t, prompts[1], "cache-bust")
	assert.NotEqual(t, prompts[0], prompts[1])
}

func TestBasicSkipsVendorSpecificFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "demo",
		Modules: []string{ModuleBasic},
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
	assert.Equal(t, "pass", statusByCheck[CheckConnectivity])
	assert.Equal(t, "skip", statusByCheck[CheckStream])
	assert.Equal(t, "skip", statusByCheck[CheckUsage])
	assert.Equal(t, "skip", statusByCheck[CheckRequestID])
	assert.Equal(t, "skip", statusByCheck[CheckJSONMode])
	assert.Equal(t, "skip", statusByCheck[CheckToolCall])
	assert.Equal(t, "skip", statusByCheck[CheckThinking])
	assert.Equal(t, "skip", statusByCheck[CheckAuthError])
	assert.Equal(t, "skip", statusByCheck[CheckBadRequest])
}

func TestBasicRejectsIncompleteVendorUsage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "glm-5.3-flash",
		Vendor:  VendorGLM,
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Checks: []string{CheckConnectivity, CheckUsage}},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)

	statusByCheck := map[string]string{}
	var usageMessage string
	for _, event := range events {
		if event.Type == "check" && event.CheckID != "" {
			statusByCheck[event.CheckID] = event.Status
			if event.CheckID == CheckUsage {
				usageMessage = event.Message
			}
		}
	}
	assert.Equal(t, "pass", statusByCheck[CheckConnectivity])
	assert.Equal(t, "fail", statusByCheck[CheckUsage])
	assert.Contains(t, usageMessage, "完整 usage")
}

func TestBasicRunsOnlyRequestedChecks(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"id\":\"chatcmpl-json\",\"choices\":[{\"delta\":{\"content\":\"{\\\"ping\\\":\\\"pong\\\"}\"},\"finish_reason\":\"stop\"}]}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "demo",
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Checks: []string{CheckJSONMode}},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)
	assert.Equal(t, 1, calls)

	statusByCheck := map[string]string{}
	var summary string
	for _, event := range events {
		if event.Type == "check" && event.CheckID != "" {
			statusByCheck[event.CheckID] = event.Status
		}
		if event.Type == "summary" {
			summary = event.Summary
		}
	}
	assert.Equal(t, "pass", statusByCheck[CheckJSONMode])
	_, hasConnectivity := statusByCheck[CheckConnectivity]
	assert.False(t, hasConnectivity)
	assert.Contains(t, summary, "通过 1")
}

func TestCacheModuleReportsMissingCachedTokens(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":20,\"completion_tokens\":1}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "demo",
		Modules: []string{ModuleCache},
		Cache:   CacheConfig{Prompt: "cache-corpus", WaitSeconds: 0, MaxTokens: 8},
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
	assert.Equal(t, "pass", statusByCheck[CheckCacheWarm])
	assert.Equal(t, "pass", statusByCheck[CheckCacheProbe])
	assert.Equal(t, "skip", statusByCheck[CheckCacheTokens])
	assert.Equal(t, "skip", statusByCheck[CheckCacheHitRate])
	assert.Equal(t, "skip", statusByCheck[CheckCacheTTL])
}

func TestCacheRunsConfiguredProbeRounds(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":20,\"completion_tokens\":1,\"cached_tokens\":16}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "demo",
		Modules: []string{ModuleCache},
		Cache:   CacheConfig{Prompt: "cache-corpus", WaitSeconds: 0, MaxTokens: 8, Rounds: 3},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)
	assert.Equal(t, 4, calls)

	statusByCheck := map[string]string{}
	var summary string
	for _, event := range events {
		if event.Type == "check" && event.CheckID != "" {
			statusByCheck[event.CheckID] = event.Status
		}
		if event.Type == "summary" {
			summary = event.Summary
		}
	}
	assert.Equal(t, "pass", statusByCheck[CheckCacheWarm])
	assert.Equal(t, "pass", statusByCheck[CheckCacheProbe])
	assert.Equal(t, "pass", statusByCheck[CheckCacheTokens])
	assert.Equal(t, "pass", statusByCheck[CheckCacheHitRate])
	assert.Equal(t, "skip", statusByCheck[CheckCacheTTL])
	assert.Contains(t, summary, "探测 3 轮")
}

func TestCacheReportsPartialProbeFailure(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 2 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = io.WriteString(w, `{"error":{"message":"probe unavailable"}}`)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":20,\"completion_tokens\":1,\"cached_tokens\":16}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "demo",
		Modules: []string{ModuleCache},
		Cache:   CacheConfig{Prompt: "cache-corpus", WaitSeconds: 0, MaxTokens: 8, Rounds: 2},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)
	assert.Equal(t, 3, calls)

	statusByCheck := map[string]string{}
	var metrics *CacheMetrics
	for _, event := range events {
		if event.Type == "check" && event.CheckID != "" {
			statusByCheck[event.CheckID] = event.Status
		}
		if event.Type == "metrics" {
			metrics = event.Cache
		}
	}
	assert.Equal(t, "fail", statusByCheck[CheckCacheProbe])
	require.NotNil(t, metrics)
	assert.Equal(t, 1, metrics.HitCount)
}

func TestCacheHitRateDoesNotFailOnLowHit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":20,\"completion_tokens\":1,\"cached_tokens\":4}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "demo",
		Modules: []string{ModuleCache},
		Cache:   CacheConfig{Prompt: "cache-corpus", WaitSeconds: 0, MaxTokens: 8, Rounds: 1},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)

	statusByCheck := map[string]string{}
	var cacheMetrics *CacheMetrics
	for _, event := range events {
		if event.Type == "check" && event.CheckID != "" {
			statusByCheck[event.CheckID] = event.Status
		}
		if event.Type == "metrics" && event.Cache != nil {
			cacheMetrics = event.Cache
		}
	}
	assert.Equal(t, "pass", statusByCheck[CheckCacheHitRate])
	assert.Equal(t, "skip", statusByCheck[CheckCacheTTL])
	require.NotNil(t, cacheMetrics)
	assert.True(t, cacheMetrics.HasCachedTokens)
	assert.InDelta(t, 0.2, cacheMetrics.AvgHitRate, 0.001)
	var hitMessage string
	for _, event := range events {
		if event.Type == "check" && event.CheckID == CheckCacheHitRate {
			hitMessage = event.Message
		}
	}
	assert.NotContains(t, hitMessage, "不正常")
	assert.NotContains(t, hitMessage, "偏弱")
}

func TestBasicSkipsDependentsWhenConnectivityFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, `{"error":{"message":"upstream down"}}`)
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "demo",
		Modules: []string{ModuleBasic},
		Basic: BasicConfig{
			Checks: []string{CheckConnectivity, CheckStream, CheckUsage, CheckRequestID, CheckSampling},
		},
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
	assert.Equal(t, "fail", statusByCheck[CheckConnectivity])
	assert.Equal(t, "skip", statusByCheck[CheckStream])
	assert.Equal(t, "skip", statusByCheck[CheckUsage])
	assert.Equal(t, "skip", statusByCheck[CheckRequestID])
	assert.Equal(t, "skip", statusByCheck[CheckSampling])
}

func TestStreamChatReadsVendorCacheAliases(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":40,\"completion_tokens\":1,\"prompt_cache_hit_tokens\":18}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	result := streamChat(context.Background(), server.Client(), server.URL+"/v1/chat/completions", "test-key", chatRequest{
		Model:  "demo",
		Stream: true,
	}, 5*time.Second, nil)
	require.Equal(t, http.StatusOK, result.StatusCode)
	assert.True(t, result.HasCachedTokens)
	assert.Equal(t, 18, result.CachedTokens)
}

func TestLooksLikeJSON(t *testing.T) {
	t.Parallel()
	assert.True(t, looksLikeJSON(`{"ping":"pong"}`))
	assert.True(t, looksLikeJSON(`{}`))
	assert.False(t, looksLikeJSON(`[]`))
	assert.False(t, looksLikeJSON(`"pong"`))
	assert.False(t, looksLikeJSON(`null`))
	assert.False(t, looksLikeJSON("not json"))
	assert.False(t, looksLikeJSON(""))
}

func TestCacheCumulativeModeAccumulatesMessages(t *testing.T) {
	t.Parallel()

	var recordedBodies []chatRequest
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		body, _ := io.ReadAll(r.Body)
		_ = common.Unmarshal(body, &req)
		mu.Lock()
		recordedBodies = append(recordedBodies, req)
		callIdx := len(recordedBodies)
		mu.Unlock()

		w.Header().Set("Content-Type", "text/event-stream")
		reply := fmt.Sprintf("reply-%d", callIdx)
		_, _ = io.WriteString(w, fmt.Sprintf("data: {\"choices\":[{\"delta\":{\"content\":%q},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":20,\"completion_tokens\":1,\"cached_tokens\":16}}\n\n", reply))
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	var events []Event
	err := Run(context.Background(), server.Client(), RunRequest{
		BaseURL: server.URL,
		APIKey:  "any",
		Model:   "demo",
		Modules: []string{ModuleCache},
		Cache: CacheConfig{
			Prompt:      "cache-corpus",
			FollowUp:    "user-question",
			WaitSeconds: 0,
			MaxTokens:   8,
			Rounds:      2,
			Mode:        CacheModeCumulative,
		},
	}, func(event Event) {
		events = append(events, event)
	})
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 3, len(recordedBodies))

	// Warm-up call (system + user)
	assert.Equal(t, 2, len(recordedBodies[0].Messages))

	// Probe Round 1 (system + user)
	assert.Equal(t, 2, len(recordedBodies[1].Messages))

	// Probe Round 2 (system + user + assistant + user) -> accumulated!
	assert.Equal(t, 4, len(recordedBodies[2].Messages))
	assert.Equal(t, "assistant", recordedBodies[2].Messages[2].Role)
	assert.Equal(t, "reply-2", recordedBodies[2].Messages[2].Content)
	assert.Equal(t, "user", recordedBodies[2].Messages[3].Role)
	assert.Equal(t, "user-question", recordedBodies[2].Messages[3].Content)

	var cacheMetrics *CacheMetrics
	for _, event := range events {
		if event.Type == "metrics" && event.Cache != nil {
			cacheMetrics = event.Cache
		}
	}
	require.NotNil(t, cacheMetrics)
	assert.Equal(t, CacheModeCumulative, cacheMetrics.Mode)
	assert.Equal(t, 2, cacheMetrics.HitCount)
	assert.InDelta(t, 0.8, cacheMetrics.AvgDepthRate, 0.001)
}
