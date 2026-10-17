package openai

import (
	"io"
	"strings"
	"testing"

	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectStreamFunctionCallNamesDedupesSameIndex(t *testing.T) {
	seen := make(map[string]struct{})
	var names []string

	chunks := []string{
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"c1","type":"function","function":{"name":"get_weather","arguments":""}}]}}]}`,
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"q\":"}}]}}]}`,
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"x\"}"}}]}}]}`,
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"id":"c2","type":"function","function":{"name":"get_time","arguments":""}}]}}]}`,
		`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":1,"function":{"arguments":"{}"}}]}}]}`,
	}
	for _, chunk := range chunks {
		collectStreamFunctionCallNames(chunk, seen, &names)
	}

	require.Len(t, names, 2)
	assert.Equal(t, []string{"get_weather", "get_time"}, names)
}

func TestChatStreamRetryDoesNotCarryToolChargesFromDiscardedAttempt(t *testing.T) {
	const toolName = "stream_audit_priced_tool"
	operation_setting.SetToolPriceForTest(toolName, 5)
	t.Cleanup(func() { operation_setting.DeleteToolPriceForTest(toolName) })
	chunk := `{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"` + toolName + `","arguments":"{}"}}]}}]}`
	c, _, info, resp := setupResponsesStreamReliabilityTest(t, io.NopCloser(strings.NewReader("data: "+chunk+"\n\n")))
	info.RelayMode = relayconstant.RelayModeChatCompletions
	info.RelayFormat = types.RelayFormatOpenAI
	info.OriginModelName = "stream-billing-fixture"
	_, apiErr := OaiStreamHandler(c, info, resp)
	require.NotNil(t, apiErr, "the first attempt has no downstream output or terminal and can retry")
	if info.ResponsesUsageInfo != nil {
		assert.NotContains(t, info.ResponsesUsageInfo.BuiltInTools, toolName, "a discarded attempt must not leave a per-call charge")
	}
	resp.Body = io.NopCloser(strings.NewReader("data: " + chunk + "\n\ndata: " + streamUsageEmptyChunk + "\n\ndata: [DONE]\n\n"))
	_, apiErr = OaiStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, info.ResponsesUsageInfo)
	require.Contains(t, info.ResponsesUsageInfo.BuiltInTools, toolName)
	assert.Equal(t, 1, info.ResponsesUsageInfo.BuiltInTools[toolName].CallCount)
}
