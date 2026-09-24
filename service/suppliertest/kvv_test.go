package suppliertest

import (
	"context"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

var kvvTestName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,255}$`)

func TestKimiKVVOfficialPreflightPasses(t *testing.T) {
	t.Parallel()
	upstream := httptest.NewServer(http.HandlerFunc(kvvOfficialFixture))
	defer upstream.Close()

	status, message := runKVVAgainst(t, upstream)
	assert.Equal(t, "pass", status)
	assert.Contains(t, message, "通过 13/13")
	assert.Contains(t, message, "非官方 Kimi KVV 认证")
}

func TestKimiKVVUsesK3ThinkingProtocol(t *testing.T) {
	t.Parallel()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if gjson.GetBytes(body, "thinking").Exists() || gjson.GetBytes(body, "reasoning_effort").String() != "low" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":{"message":"K3 仅接受顶层 reasoning_effort，不接受 thinking","type":"invalid_request_error","param":"thinking","code":"3000"}}`)
			return
		}
		kvvOfficialFixtureBody(w, body)
	}))
	defer upstream.Close()

	status, message := runKVVAgainst(t, upstream)
	assert.Equal(t, "pass", status, message)
}

func TestKimiKVVNativeToolCallFallback(t *testing.T) {
	t.Parallel()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if gjson.GetBytes(body, "tool_choice").String() == "required" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"call\n{\"api_name\": \"get_weather\", \"parameters\": {\"city\": \"Beijing\"}}"}}],"usage":{"prompt_tokens":10,"completion_tokens":20}}`)
			return
		}
		kvvOfficialFixtureBody(w, body)
	}))
	defer upstream.Close()

	status, message := runKVVAgainst(t, upstream)
	assert.Equal(t, "pass", status, message)
}

func TestKimiKVVToleratesLooseSampling(t *testing.T) {
	t.Parallel()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// 模拟上游网关忽略非法参数、直接放行返回 200 OK（大概能过就过）
		if code, ok := kvvFixtureReject(body); ok {
			if code == http.StatusBadRequest {
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"choices":[{"finish_reason":"stop","message":{"content":"OK"}}]}`)
				return
			}
		}
		kvvOfficialFixtureBody(w, body)
	}))
	defer upstream.Close()

	status, message := runKVVAgainst(t, upstream)
	assert.Equal(t, "pass", status, message)
}

func TestKimiKVVPassesWithPartialFeatureSupport(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if gjson.GetBytes(body, "response_format").Exists() {
			w.WriteHeader(http.StatusNotImplemented)
			_, _ = io.WriteString(w, `{"error":{"message":"response format unsupported"}}`)
			return
		}
		kvvOfficialFixtureBody(w, body)
	}))
	defer upstream.Close()

	status, message := runKVVAgainst(t, upstream)
	assert.Equal(t, "pass", status, message)
	assert.Contains(t, message, "通过 9/13")
	assert.Contains(t, message, "69.2%")
	assert.Contains(t, message, "response_format text")
}

func TestKimiKVVRejectsServerError(t *testing.T) {
	t.Parallel()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"message":"internal error"}}`)
	}))
	defer upstream.Close()

	status, message := runKVVAgainst(t, upstream)
	assert.Equal(t, "fail", status)
	assert.Contains(t, message, "params K3 low-effort thinking")
	assert.Contains(t, message, "期望 HTTP 200，实际 500")
}

func TestKimiKVVSkipsNonK3WithoutCallingUpstream(t *testing.T) {
	t.Parallel()

	var calls int
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer upstream.Close()

	var status, message string
	err := Run(context.Background(), upstream.Client(), RunRequest{
		BaseURL: upstream.URL,
		APIKey:  "test-key",
		Model:   "kimi-k2.6",
		Vendor:  VendorKimi,
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Checks: []string{CheckKimiKVV}},
	}, func(e Event) {
		if e.CheckID == CheckKimiKVV && e.Status != "running" {
			status = e.Status
			message = e.Message
		}
	})
	require.NoError(t, err)
	assert.Equal(t, "skip", status)
	assert.Contains(t, message, "只适用于 Kimi K3")
	assert.Zero(t, calls)
}

func runKVVAgainst(t *testing.T, upstream *httptest.Server) (string, string) {
	t.Helper()
	var status, message string
	err := Run(context.Background(), upstream.Client(), RunRequest{
		BaseURL: upstream.URL,
		APIKey:  "test-key",
		Model:   "kimi-k3",
		Vendor:  VendorKimi,
		Modules: []string{ModuleBasic},
		Basic:   BasicConfig{Checks: []string{CheckKimiKVV}},
	}, func(e Event) {
		if e.CheckID == CheckKimiKVV && e.Status != "running" {
			status = e.Status
			message = e.Message
		}
	})
	require.NoError(t, err)
	require.NotEmpty(t, status)
	return status, message
}

func kvvOfficialFixture(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	kvvOfficialFixtureBody(w, body)
}

func kvvOfficialFixtureBody(w http.ResponseWriter, body []byte) {
	if code, ok := kvvFixtureReject(body); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = io.WriteString(w, `{"error":{"message":"rejected"}}`)
		return
	}
	if gjson.GetBytes(body, "stream").Bool() && strings.Contains(string(body), "鸡兔同笼") {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, `data: {"choices":[{"delta":{"reasoning_content":"设鸡x兔y"}}]}`+"\n\n")
		_, _ = io.WriteString(w, `data: {"choices":[{"delta":{"content":"鸡23只，兔12只"}}]}`+"\n\n")
		_, _ = io.WriteString(w, `data: {"choices":[{"finish_reason":"stop"}]}`+"\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, kvvFixtureOK(body))
}

func kvvFixtureReject(body []byte) (int, bool) {
	if gjson.GetBytes(body, "temperature").Exists() {
		value := gjson.GetBytes(body, "temperature").Float()
		if math.Abs(value-1.0) > 0.001 {
			return http.StatusBadRequest, true
		}
	}
	if gjson.GetBytes(body, "top_p").Exists() && math.Abs(gjson.GetBytes(body, "top_p").Float()-0.95) > 0.001 {
		return http.StatusBadRequest, true
	}
	if gjson.GetBytes(body, "presence_penalty").Exists() && gjson.GetBytes(body, "presence_penalty").Float() != 0 {
		return http.StatusBadRequest, true
	}
	if gjson.GetBytes(body, "frequency_penalty").Exists() && gjson.GetBytes(body, "frequency_penalty").Float() != 0 {
		return http.StatusBadRequest, true
	}
	if gjson.GetBytes(body, "n").Exists() && gjson.GetBytes(body, "n").Float() != 1 {
		return http.StatusBadRequest, true
	}
	format := gjson.GetBytes(body, "response_format")
	if format.Exists() {
		switch format.Get("type").String() {
		case "bogus":
			return http.StatusBadRequest, true
		case "json_schema":
			if !format.Get("json_schema.name").Exists() || !format.Get("json_schema.schema").Exists() {
				return http.StatusBadRequest, true
			}
		}
	}
	if rejected := kvvFixtureTools(body); rejected {
		return http.StatusBadRequest, true
	}
	choice := gjson.GetBytes(body, "tool_choice").String()
	if choice == "bogus" || (choice == "required" && !kvvFixtureHasTools(body)) {
		return http.StatusBadRequest, true
	}
	if gjson.GetBytes(body, "thinking.type").String() == "disabled" {
		return http.StatusBadRequest, true
	}
	if gjson.GetBytes(body, "thinking").Exists() {
		return http.StatusBadRequest, true
	}
	if effort := gjson.GetBytes(body, "reasoning_effort"); effort.Exists() && effort.String() != "low" && effort.String() != "high" && effort.String() != "max" {
		return http.StatusBadRequest, true
	}
	return 0, false
}

func kvvFixtureTools(body []byte) bool {
	seen := map[string]bool{}
	if tools := gjson.GetBytes(body, "tools"); tools.IsArray() {
		for _, tool := range tools.Array() {
			name := tool.Get("function.name").String()
			if name != "" {
				seen[name] = true
			}
		}
	}
	for _, msg := range gjson.GetBytes(body, "messages").Array() {
		if msg.Get("role").String() == "tool" && !msg.Get("tool_call_id").Exists() {
			return true
		}
		tools := msg.Get("tools")
		if !tools.Exists() {
			continue
		}
		if msg.Get("role").String() != "system" || strings.TrimSpace(msg.Get("content").String()) != "" || !tools.IsArray() {
			return true
		}
		for _, tool := range tools.Array() {
			if tool.Type == gjson.Null || !tool.IsObject() || !tool.Get("type").Exists() || !tool.Get("function").Exists() {
				return true
			}
			if tool.Get("type").String() != "function" || !tool.Get("function.name").Exists() {
				return true
			}
			name := tool.Get("function.name").String()
			if !kvvTestName.MatchString(name) || seen[name] {
				return true
			}
			seen[name] = true
		}
	}
	return false
}

func kvvFixtureHasTools(body []byte) bool {
	if tools := gjson.GetBytes(body, "tools"); tools.IsArray() && len(tools.Array()) > 0 {
		return true
	}
	for _, msg := range gjson.GetBytes(body, "messages").Array() {
		if msg.Get("role").String() == "system" && msg.Get("tools").IsArray() && len(msg.Get("tools").Array()) > 0 {
			return true
		}
	}
	return false
}

func kvvFixtureOK(body []byte) string {
	if strings.Contains(string(body), "鸡兔同笼") {
		return `{"choices":[{"finish_reason":"stop","message":{"content":"鸡23只，兔12只","reasoning_content":"设鸡x兔y"}}]}`
	}
	format := gjson.GetBytes(body, "response_format.type").String()
	switch format {
	case "text":
		return `{"choices":[{"finish_reason":"stop","message":{"content":"北京是中国的首都。"}}]}`
	case "json_object":
		return `{"choices":[{"finish_reason":"stop","message":{"content":"{\"city\":\"Beijing\"}"}}]}`
	case "json_schema":
		if gjson.GetBytes(body, "response_format.json_schema.strict").Bool() {
			return `{"choices":[{"finish_reason":"stop","message":{"content":"{\"city\":\"Beijing\",\"temperature\":21}"}}]}`
		}
		return `{"choices":[{"finish_reason":"stop","message":{"content":"{\"city\":\"Beijing\"}"}}]}`
	}
	choice := gjson.GetBytes(body, "tool_choice").String()
	text := string(body)
	if choice == "none" || strings.Contains(text, "不应使用任何工具") || !kvvFixtureHasTools(body) {
		return `{"choices":[{"finish_reason":"stop","message":{"content":"OK"}}]}`
	}
	name := "get_weather"
	var names []string
	collect := func(tools gjson.Result) {
		for _, tool := range tools.Array() {
			if got := tool.Get("function.name").String(); got != "" {
				names = append(names, got)
			}
		}
	}
	collect(gjson.GetBytes(body, "tools"))
	for _, msg := range gjson.GetBytes(body, "messages").Array() {
		collect(msg.Get("tools"))
	}
	for _, got := range names {
		if got == "get_weather" {
			name = got
			break
		}
		name = got
	}
	return `{"choices":[{"finish_reason":"tool_calls","message":{"tool_calls":[{"function":{"name":"` + name + `","arguments":"{}"}}]}}]}`
}

func TestLiveKimiK3KVV(t *testing.T) {
	apiKey := strings.TrimSpace(os.Getenv("KIMI_TEST_API_KEY"))
	if apiKey == "" {
		t.Skip("set KIMI_TEST_API_KEY to run the manual live test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	endpoint := strings.TrimSpace(os.Getenv("KIMI_TEST_ENDPOINT"))
	if endpoint == "" {
		endpoint = "https://api.moonshot.ai/v1/chat/completions"
	}
	model := "kimi-k3"

	status, message := runOfficialKimiKVV(ctx, &http.Client{Timeout: 60 * time.Second}, endpoint, apiKey, model)
	t.Logf("LIVE TEST RESULT: status=%s, message=%s", status, message)
	assert.Equal(t, "pass", status, message)
}
