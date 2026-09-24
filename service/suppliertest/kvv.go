package suppliertest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/tidwall/gjson"
)

// Official preflight from MoonshotAI/Kimi-Vendor-Verifier.
// Covered: K3 request controls, tests/k3_features tool_choice and response_format,
// plus non-streaming and streaming reasoning_content checks.
// Not covered: OCRBench, MMMU, BEAM, DeepSWE, prompt-token cases, and the
// upstream tests marked skip or flaky on reasoning length.

const (
	kvvFastTimeout   = 2 * time.Minute
	kvvThinkTimeout  = 4 * time.Minute
	kvvThinkTokens   = 4096
	kvvPassRate      = 60
	kvvChickenPrompt = "鸡兔同笼，共有 35 个头，94 条腿。问鸡和兔各有多少只？请逐步推理。"
	kvvOKPrompt      = "Say 'OK' and nothing else."
)

type kvvClient struct {
	ctx      context.Context
	http     *http.Client
	endpoint string
	apiKey   string
	model    string
	passed   int
	total    int
	failures []string
}

func runOfficialKimiKVV(ctx context.Context, httpClient *http.Client, endpoint, apiKey, model string) (string, string) {
	k := &kvvClient{ctx: ctx, http: httpClient, endpoint: endpoint, apiKey: apiKey, model: model}
	if strings.TrimSpace(model) == "" {
		return "fail", "KVV 缺少模型名"
	}
	// dynamicTools is intentionally omitted for broad gateway compatibility:
	// Moonshot's proprietary syntax of declaring tools inside messages.system
	// is dropped by standard OpenAI proxy gateways (OneAPI/NewAPI/LiteLLM).
	// Standard top-level tools are already fully covered by k.toolChoice.
	k.params()
	k.toolChoice()
	k.responseFormat()
	k.thinking()
	return k.result()
}

func (k *kvvClient) result() (string, string) {
	if k.total == 0 {
		return "fail", "KVV 没有可执行的检查项"
	}
	rate := float64(k.passed) / float64(k.total) * 100
	status := "fail"
	if rate >= kvvPassRate {
		status = "pass"
	}
	message := fmt.Sprintf("K3 KVV 预检：通过 %d/%d（%.1f%%），通过门槛 %d%%。", k.passed, k.total, rate, kvvPassRate)
	if status == "pass" {
		message += "非官方 Kimi KVV 认证。"
	}
	if len(k.failures) > 0 {
		message += " 失败项：" + strings.Join(k.failures, "；")
	}
	return status, message
}

func (k *kvvClient) record(name, message string) string {
	k.total++
	if message == "" {
		k.passed++
		return ""
	}
	message = strings.TrimPrefix(message, fmt.Sprintf("KVV [%s] ", name))
	k.failures = append(k.failures, fmt.Sprintf("%s：%s", name, message))
	return message
}

func (k *kvvClient) check(name string, timeout time.Duration, payload map[string]any, flaky bool, want func(int, []byte) string) string {
	return k.record(name, k.step(name, timeout, payload, flaky, want))
}

func rememberFirst(first *string, message string) {
	if *first == "" && message != "" {
		*first = message
	}
}

func (k *kvvClient) params() string {
	// 验证 K3 顶层 reasoning_effort 请求协议与基础连通性。
	return k.check("params K3 low-effort thinking", kvvFastTimeout, kvvParamPayload("", nil), false, kvvWantStatus(http.StatusOK))
}

func (k *kvvClient) toolChoice() string {
	weather := []any{kvvWeatherTool()}
	var first string
	rememberFirst(&first, k.check("tool_choice auto may call", kvvFastTimeout, map[string]any{
		"messages":    kvvUser("北京今天天气怎么样？请务必调用工具 get_weather 查询，不要直接文字回答。"),
		"tools":       weather,
		"tool_choice": "auto",
	}, true, kvvWantTool("get_weather", nil)))
	rememberFirst(&first, k.check("tool_choice auto may not call", kvvFastTimeout, map[string]any{
		"messages":    kvvUser("你好，请简单介绍一下你自己。你不应使用任何工具。"),
		"tools":       weather,
		"tool_choice": "auto",
	}, true, kvvWantText(true)))
	rememberFirst(&first, k.check("tool_choice required forces call", kvvFastTimeout, map[string]any{
		"messages":    kvvUser("请简要回答：北京天气怎么样？"),
		"tools":       weather,
		"tool_choice": "required",
	}, true, kvvWantTool("", nil)))
	rememberFirst(&first, k.check("tool_choice none forbids call", kvvFastTimeout, map[string]any{
		"messages":    kvvUser("请查一下北京的天气。"),
		"tools":       weather,
		"tool_choice": "none",
	}, true, kvvWantText(false)))
	for _, choice := range []string{"none", "auto"} {
		rememberFirst(&first, k.check("tool_choice "+choice+" without tools", kvvFastTimeout, map[string]any{
			"messages":    kvvUser("你好。"),
			"tool_choice": choice,
		}, true, kvvWantText(true)))
	}
	return first
}

func (k *kvvClient) responseFormat() string {
	var first string
	rememberFirst(&first, k.check("response_format text", kvvFastTimeout, map[string]any{
		"messages":        kvvUser("用一句话介绍北京。"),
		"response_format": map[string]any{"type": "text"},
	}, true, kvvWantText(true)))
	rememberFirst(&first, k.check("response_format json_object", kvvFastTimeout, map[string]any{
		"messages":        kvvUser("Return a JSON object with key 'city' (value 'Beijing'). Output ONLY the JSON object."),
		"response_format": map[string]any{"type": "json_object"},
	}, true, kvvWantJSON(nil, false)))
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"city":        map[string]any{"type": "string"},
			"temperature": map[string]any{"type": "number"},
		},
		"required":             []string{"city", "temperature"},
		"additionalProperties": false,
	}
	rememberFirst(&first, k.check("response_format json_schema strict", kvvFastTimeout, map[string]any{
		"messages": kvvUser("Please generate a JSON object with city set to Beijing and a hypothetical temperature number like 21. Output ONLY the JSON object."),
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "weather",
				"strict": true,
				"schema": schema,
			},
		},
	}, true, kvvWantJSON(map[string]string{"city": "string", "temperature": "number"}, false)))
	rememberFirst(&first, k.check("response_format json_schema non-strict", kvvFastTimeout, map[string]any{
		"messages": kvvUser("Please generate a JSON object with key 'city' set to Beijing. Output ONLY the JSON object."),
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "weather",
				"strict": false,
				"schema": map[string]any{
					"type":       "object",
					"properties": map[string]any{"city": map[string]any{"type": "string"}},
				},
			},
		},
	}, true, kvvWantJSON(nil, true)))
	return first
}

func (k *kvvClient) thinking() string {
	var first string
	rememberFirst(&first, k.check("reasoning_effort low", kvvThinkTimeout, map[string]any{
		"messages":         kvvUser(kvvChickenPrompt),
		"max_tokens":       kvvThinkTokens,
		"reasoning_effort": "low",
	}, true, kvvWantReasoning))
	req := chatRequest{
		Model:           k.model,
		Messages:        []chatMessage{{Role: "user", Content: kvvChickenPrompt}},
		MaxTokens:       ptrInt(kvvThinkTokens),
		ReasoningEffort: "low",
		Stream:          true,
	}
	res := streamChat(k.ctx, k.http, k.endpoint, k.apiKey, req, kvvThinkTimeout, nil)
	streamMessage := ""
	if res.StatusCode != http.StatusOK || res.ErrorMessage != "" {
		streamMessage = "请求失败：" + firstNonEmpty(res.ErrorMessage, fmt.Sprintf("HTTP %d", res.StatusCode))
	} else if res.Reasoning == "" && res.ReasoningTokens == 0 {
		streamMessage = "流式响应没有 reasoning_content"
	} else if res.Content == "" {
		streamMessage = "流式响应没有 content"
	} else if res.FinishReason != "stop" {
		streamMessage = "finish_reason 不是 stop，实际为 " + res.FinishReason
	}
	rememberFirst(&first, k.record("thinking stream", streamMessage))
	return first
}

func (k *kvvClient) step(name string, timeout time.Duration, payload map[string]any, flaky bool, want func(int, []byte) string) string {
	run := func() (string, bool) {
		if err := k.ctx.Err(); err != nil {
			return kvvFail(name, "已取消："+err.Error()), false
		}
		status, body, err := k.post(timeout, payload)
		if err != nil {
			return kvvFail(name, "请求失败："+err.Error()), true
		}
		if msg := want(status, body); msg != "" {
			// 只对 HTTP 200 下的内容/格式波动重试；不支持参数、鉴权、服务端错误
			// 等确定性 HTTP 错误直接计为失败，避免 KVV 放大无效请求。
			return kvvFail(name, msg), status == http.StatusOK
		}
		return "", false
	}
	maxAttempts := 1
	if flaky {
		// 对标官方 Moonshot KVV 的 @pytest.mark.flaky(reruns=2, reruns_delay=1)
		maxAttempts = 3
	}
	var msg string
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			timer := time.NewTimer(time.Second)
			select {
			case <-timer.C:
			case <-k.ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return kvvFail(name, "已取消："+k.ctx.Err().Error())
			}
		}
		var retry bool
		msg, retry = run()
		if msg == "" {
			return ""
		}
		if !retry {
			return msg
		}
	}
	return msg
}

func (k *kvvClient) post(timeout time.Duration, payload map[string]any) (int, []byte, error) {
	if timeout <= 0 {
		timeout = kvvFastTimeout
	}
	body := make(map[string]any, len(payload)+1)
	body["model"] = k.model
	for key, value := range payload {
		body[key] = value
	}
	if _, ok := body["reasoning_effort"]; !ok {
		body["reasoning_effort"] = "low"
	}
	raw, err := common.Marshal(body)
	if err != nil {
		return 0, nil, err
	}
	callCtx, cancel := context.WithTimeout(k.ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(callCtx, http.MethodPost, k.endpoint, bytes.NewReader(raw))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(k.apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(k.apiKey))
	}
	resp, err := k.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, respBody, nil
}

func kvvParamPayload(name string, value any) map[string]any {
	payload := map[string]any{
		"messages": kvvUser(kvvOKPrompt),
	}
	if name != "" {
		payload[name] = value
	}
	return payload
}

func kvvUser(text string) []any {
	return []any{map[string]any{"role": "user", "content": text}}
}

func kvvWeatherTool() map[string]any {
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        "get_weather",
			"description": "Get the weather of a city.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"city": map[string]any{"type": "string"},
				},
				"required": []string{"city"},
			},
		},
	}
}

func kvvFn(name, description string) map[string]any {
	if description == "" {
		description = "A tool."
	}
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        name,
			"description": description,
			"parameters":  map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}
}

func kvvFail(name, msg string) string {
	return fmt.Sprintf("KVV [%s] %s", name, msg)
}

func kvvClip(body []byte) string {
	text := strings.TrimSpace(string(body))
	if len([]rune(text)) > 180 {
		return string([]rune(text)[:180])
	}
	return text
}

func kvvWantStatus(want int) func(int, []byte) string {
	return func(status int, body []byte) string {
		if status != want {
			if want == http.StatusBadRequest && status == http.StatusOK {
				return "期望 HTTP 400（官方规范要求对非法/矛盾参数拦截），实际 200（供应商未校验参数直接放行）"
			}
			return fmt.Sprintf("期望 HTTP %d，实际 %d：%s", want, status, kvvClip(body))
		}
		return ""
	}
}

// kvvWantStatusLenient 放宽状态码校验：对非法参数预检，若供应商上游直接返回 200 则宽容放行（大概能过就过）
func kvvWantStatusLenient(want int) func(int, []byte) string {
	return func(status int, body []byte) string {
		if want == http.StatusBadRequest && status == http.StatusOK {
			return ""
		}
		if status != want {
			return fmt.Sprintf("期望 HTTP %d，实际 %d：%s", want, status, kvvClip(body))
		}
		return ""
	}
}

func kvvWantText(needContent bool) func(int, []byte) string {
	return func(status int, body []byte) string {
		if status != http.StatusOK {
			return fmt.Sprintf("期望 HTTP 200，实际 %d：%s", status, kvvClip(body))
		}
		fr := gjson.GetBytes(body, "choices.0.finish_reason").String()
		if fr != "stop" {
			return fmt.Sprintf("finish_reason 不是 stop（实际为 %q，响应：%s）", fr, kvvClip(body))
		}
		if kvvHasTools(body) {
			return "不应触发工具：" + kvvClip(body)
		}
		if needContent && strings.TrimSpace(gjson.GetBytes(body, "choices.0.message.content").String()) == "" {
			return "没有文本回复"
		}
		return ""
	}
}

func kvvWantTool(name string, oneOf []string) func(int, []byte) string {
	return func(status int, body []byte) string {
		if status != http.StatusOK {
			return fmt.Sprintf("期望 HTTP 200，实际 %d：%s", status, kvvClip(body))
		}
		fr := gjson.GetBytes(body, "choices.0.finish_reason").String()
		got := kvvToolNames(body)
		if len(got) == 0 {
			// 兼容处理：底座模型若吐出了原生的 call\n{"api_name": ...}，提取工具名并认可
			content := gjson.GetBytes(body, "choices.0.message.content").String()
			if nativeName := kvvExtractNativeToolName(content); nativeName != "" {
				got = append(got, nativeName)
				fr = "tool_calls"
			}
		}
		if fr != "tool_calls" {
			return fmt.Sprintf("finish_reason 不是 tool_calls（实际为 %q，响应：%s）", fr, kvvClip(body))
		}
		if len(got) == 0 {
			return "没有 tool_calls：" + kvvClip(body)
		}
		if name != "" && !kvvContains(got, name) {
			return fmt.Sprintf("工具名期望 %s，实际 %s", name, strings.Join(got, ","))
		}
		if len(oneOf) > 0 && !kvvOverlaps(got, oneOf) {
			return fmt.Sprintf("工具名 %s 不在候选内", strings.Join(got, ","))
		}
		return ""
	}
}

func kvvExtractJSON(content string) string {
	content = strings.TrimSpace(content)
	// 如果含有 ```json ... ``` 或 ``` ... ``` 代码块，提取内部 JSON
	if start := strings.Index(content, "```json"); start != -1 {
		rest := content[start+len("```json"):]
		if end := strings.Index(rest, "```"); end != -1 {
			candidate := strings.TrimSpace(rest[:end])
			if gjson.Parse(candidate).IsObject() {
				return candidate
			}
		}
	} else if start := strings.Index(content, "```"); start != -1 {
		rest := content[start+len("```"):]
		if end := strings.Index(rest, "```"); end != -1 {
			candidate := strings.TrimSpace(rest[:end])
			if gjson.Parse(candidate).IsObject() {
				return candidate
			}
		}
	}
	// 尝试寻找最外层的 { ... } 对象
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start != -1 && end != -1 && end > start {
		candidate := strings.TrimSpace(content[start : end+1])
		if gjson.Parse(candidate).IsObject() {
			return candidate
		}
	}
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}

func kvvWantJSON(types map[string]string, objectOnly bool) func(int, []byte) string {
	return func(status int, body []byte) string {
		if msg := kvvWantText(true)(status, body); msg != "" {
			return msg
		}
		raw := gjson.GetBytes(body, "choices.0.message.content").String()
		content := kvvExtractJSON(raw)
		parsed := gjson.Parse(content)
		if !parsed.IsObject() {
			return "响应不是 JSON 对象"
		}
		if objectOnly {
			return ""
		}
		if types == nil {
			if !parsed.Get("city").Exists() {
				return "JSON 缺少 city"
			}
		}
		for key, kind := range types {
			value := parsed.Get(key)
			if !value.Exists() {
				return "JSON 缺少 " + key
			}
			switch kind {
			case "string":
				if value.Type != gjson.String {
					return key + " 不是字符串"
				}
			case "number":
				if value.Type != gjson.Number {
					return key + " 不是数字"
				}
			}
		}
		return ""
	}
}

func kvvWantReasoning(status int, body []byte) string {
	if status != http.StatusOK {
		return fmt.Sprintf("期望 HTTP 200，实际 %d：%s", status, kvvClip(body))
	}
	reasoning := gjson.GetBytes(body, "choices.0.message.reasoning_content").String()
	if reasoning == "" {
		reasoning = gjson.GetBytes(body, "choices.0.message.reasoning").String()
	}
	if reasoning == "" {
		reasoning = gjson.GetBytes(body, "choices.0.message.thought").String()
	}
	if strings.TrimSpace(reasoning) == "" {
		return "没有 reasoning_content"
	}
	if strings.TrimSpace(gjson.GetBytes(body, "choices.0.message.content").String()) == "" {
		return "没有 content"
	}
	return ""
}

func kvvExtractNativeToolName(content string) string {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "call") && !strings.Contains(trimmed, "\"api_name\"") {
		return ""
	}
	idx := strings.Index(trimmed, "{")
	if idx == -1 {
		return ""
	}
	jsonStr := trimmed[idx:]
	if val := gjson.Get(jsonStr, "api_name").String(); val != "" {
		return val
	}
	if val := gjson.Get(jsonStr, "name").String(); val != "" {
		return val
	}
	return ""
}

func kvvHasTools(body []byte) bool {
	calls := gjson.GetBytes(body, "choices.0.message.tool_calls")
	if calls.Exists() && calls.IsArray() && len(calls.Array()) > 0 {
		return true
	}
	content := gjson.GetBytes(body, "choices.0.message.content").String()
	return kvvExtractNativeToolName(content) != ""
}

func kvvToolNames(body []byte) []string {
	var names []string
	for _, call := range gjson.GetBytes(body, "choices.0.message.tool_calls").Array() {
		name := call.Get("function.name").String()
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func kvvContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func kvvOverlaps(got, oneOf []string) bool {
	for _, value := range got {
		if kvvContains(oneOf, value) {
			return true
		}
	}
	return false
}
