package openai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type convertedStreamErrorReader struct{}

func (convertedStreamErrorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestChatToResponsesStreamRejectsPrematureEnd(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	for _, transportError := range []bool{false, true} {
		name := "EOF"
		if transportError {
			name = "broken HTTP body"
		}
		t.Run(name, func(t *testing.T) {
			body := "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"partial answer\"}}]}\n\n"
			c, recorder, resp, info := newResponsesChatTestContext(t, body, true)
			info.SetEstimatePromptTokens(11)
			if transportError {
				resp.Body = io.NopCloser(io.MultiReader(strings.NewReader(body), convertedStreamErrorReader{}))
			}

			usage, apiErr := OaiChatToResponsesStreamHandler(c, info, resp)
			require.Nil(t, apiErr, "partial output must not enter controller retry/refund handling")
			require.NotNil(t, usage)
			assert.Equal(t, 11, usage.PromptTokens)
			assert.Positive(t, usage.CompletionTokens)
			assert.Contains(t, recorder.Body.String(), `"delta":"partial answer"`)
			assert.Contains(t, recorder.Body.String(), "event: error\n")
			assert.Contains(t, recorder.Body.String(), `"code":"upstream_stream_incomplete"`)
			assert.NotContains(t, recorder.Body.String(), "event: response.completed\n")
			assert.NotContains(t, recorder.Body.String(), "event: response.output_text.done\n")
			assert.True(t, info.StreamStatus.HasErrors())
		})
	}
}

func TestChatToResponsesStreamPreservesFinishAndTrailingUsage(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	for _, finishReason := range []string{"stop", "length"} {
		t.Run(finishReason, func(t *testing.T) {
			body := "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"answer\"},\"finish_reason\":\"" + finishReason + "\"}]}\n\n" +
				"data: {\"choices\":[],\"usage\":{\"prompt_tokens\":11,\"completion_tokens\":7,\"total_tokens\":18}}\n\n"
			c, recorder, resp, info := newResponsesChatTestContext(t, body, true)
			resp.Body = io.NopCloser(io.MultiReader(strings.NewReader(body), convertedStreamErrorReader{}))

			usage, apiErr := OaiChatToResponsesStreamHandler(c, info, resp)
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
			assert.Equal(t, 11, usage.PromptTokens)
			assert.Equal(t, 7, usage.CompletionTokens)
			assert.Equal(t, 18, usage.TotalTokens)
			assert.Equal(t, relaycommon.StreamEndReasonDone, info.StreamStatus.EndReason)
			assert.False(t, info.StreamStatus.HasErrors())
			terminal := "response.completed"
			if finishReason == "length" {
				terminal = "response.incomplete"
			}
			assert.Contains(t, recorder.Body.String(), "event: "+terminal+"\n")
			assert.Contains(t, recorder.Body.String(), `"total_tokens":18`)
			assert.NotContains(t, recorder.Body.String(), "event: error\n")
		})
	}
}

func TestResponsesToChatStreamRejectsPrematureEndForEachClientProtocol(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	for _, format := range []types.RelayFormat{types.RelayFormatOpenAI, types.RelayFormatClaude, types.RelayFormatGemini} {
		t.Run(string(format), func(t *testing.T) {
			body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial answer\"}\n\n"
			c, recorder, resp, info := newResponsesChatTestContext(t, body, true)
			info.RelayFormat = format
			info.SetEstimatePromptTokens(11)
			resp.Body = io.NopCloser(io.MultiReader(strings.NewReader(body), convertedStreamErrorReader{}))

			usage, apiErr := OaiResponsesToChatStreamHandler(c, info, resp)
			require.Nil(t, apiErr, "partial output must not enter controller retry/refund handling")
			require.NotNil(t, usage)
			assert.Equal(t, 11, usage.PromptTokens)
			assert.Positive(t, usage.CompletionTokens)
			got := recorder.Body.String()
			assert.Contains(t, got, "partial answer")
			assert.Contains(t, got, "The upstream stream ended before the response was complete.")
			assert.NotContains(t, got, `"finish_reason":"stop"`)
			assert.NotContains(t, got, `"finishReason":"STOP"`)
			assert.NotContains(t, got, "event: message_stop\n")
			assert.NotContains(t, got, "data: [DONE]")
			assert.True(t, info.StreamStatus.HasErrors())
		})
	}
}

func TestResponsesToChatStreamTerminalStopsBeforeBodyError(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	for _, terminal := range []string{"response.completed", "response.done", "response.incomplete"} {
		t.Run(terminal, func(t *testing.T) {
			body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"answer\"}\n\n" +
				"data: {\"type\":\"" + terminal + "\",\"response\":{\"status\":\"completed\",\"usage\":{\"input_tokens\":11,\"output_tokens\":7,\"total_tokens\":18}}}\n\n"
			c, recorder, resp, info := newResponsesChatTestContext(t, body, true)
			resp.Body = io.NopCloser(io.MultiReader(strings.NewReader(body), convertedStreamErrorReader{}))

			usage, apiErr := OaiResponsesToChatStreamHandler(c, info, resp)
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
			assert.Equal(t, 18, usage.TotalTokens)
			assert.Equal(t, relaycommon.StreamEndReasonDone, info.StreamStatus.EndReason)
			assert.Contains(t, recorder.Body.String(), "data: [DONE]")
			assert.NotContains(t, recorder.Body.String(), `"upstream_stream_incomplete"`)
		})
	}
}

func TestResponsesBufferedStreamRequiresTerminal(t *testing.T) {
	for _, end := range []string{"", "data: [DONE]\n\n"} {
		body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial answer\"}\n\n" + end
		c, recorder, resp, info := newResponsesChatTestContext(t, body, false)
		usage, apiErr := OaiResponsesToChatBufferedStreamHandler(c, info, resp)
		require.NotNil(t, apiErr)
		assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
		assert.Nil(t, usage)
		assert.Empty(t, recorder.Body.String(), "buffered failures remain retryable before any output")
	}
}

func TestConvertedStreamsBeforeOutputRemainRetryable(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	for _, chatUpstream := range []bool{false, true} {
		c, recorder, resp, info := newResponsesChatTestContext(t, "", true)
		var apiErr *types.NewAPIError
		if chatUpstream {
			_, apiErr = OaiChatToResponsesStreamHandler(c, info, resp)
		} else {
			_, apiErr = OaiResponsesToChatStreamHandler(c, info, resp)
		}
		require.NotNil(t, apiErr)
		assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
		assert.Empty(t, recorder.Body.String())
		assert.Empty(t, recorder.Header().Get("Content-Type"))
		assert.Empty(t, recorder.Header().Get("Transfer-Encoding"))
	}
}

func TestResponsesToChatStreamRejectsStandardErrorEvent(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial answer\"}\n\n" +
		"data: {\"type\":\"error\",\"code\":\"server_error\",\"message\":\"private upstream details\"}\n\n"
	c, recorder, resp, info := newResponsesChatTestContext(t, body, true)
	usage, apiErr := OaiResponsesToChatStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	got := recorder.Body.String()
	assert.Contains(t, got, `"code":"upstream_stream_incomplete"`)
	assert.NotContains(t, got, "private upstream details")
	assert.NotContains(t, got, `"finish_reason":"stop"`)
	assert.NotContains(t, got, "data: [DONE]")
}

func TestResponsesToChatStreamPreservesFailedTerminalUsage(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial answer\"}\n\n" +
		"data: {\"type\":\"response.failed\",\"response\":{\"status\":\"failed\",\"error\":{\"type\":\"server_error\",\"message\":\"private upstream details\"},\"usage\":{\"input_tokens\":11,\"output_tokens\":7,\"total_tokens\":18,\"input_tokens_details\":{\"cached_tokens\":3}}}}\n\n"
	c, recorder, resp, info := newResponsesChatTestContext(t, body, true)
	usage, apiErr := OaiResponsesToChatStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 11, usage.PromptTokens)
	assert.Equal(t, 7, usage.CompletionTokens)
	assert.Equal(t, 18, usage.TotalTokens)
	assert.Equal(t, 3, usage.PromptTokensDetails.CachedTokens)
	got := recorder.Body.String()
	assert.Contains(t, got, `"code":"upstream_stream_incomplete"`)
	assert.NotContains(t, got, "private upstream details")
	assert.NotContains(t, got, `"finish_reason":"stop"`)
	assert.NotContains(t, got, "data: [DONE]")
}

func TestChatToResponsesStreamDoesNotHideMalformedTailAfterFinish(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	body := "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"answer\"},\"finish_reason\":\"stop\"}]}\n\n" +
		"data: {invalid JSON\n\n"
	c, recorder, resp, info := newResponsesChatTestContext(t, body, true)
	usage, apiErr := OaiChatToResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.True(t, info.StreamStatus.HasErrors())
	assert.Contains(t, recorder.Body.String(), "event: error\n")
	assert.NotContains(t, recorder.Body.String(), "event: response.completed\n")
}

func TestResponsesToChatRejectsNestedFailureInCompletionEvent(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	for _, terminal := range []string{"response.completed", "response.done"} {
		t.Run(terminal, func(t *testing.T) {
			body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"partial answer\"}\n\n" +
				"data: {\"type\":\"" + terminal + "\",\"response\":{\"status\":\"failed\",\"usage\":{\"input_tokens\":11,\"output_tokens\":7,\"total_tokens\":18}}}\n\n"
			c, recorder, resp, info := newResponsesChatTestContext(t, body, true)
			usage, apiErr := OaiResponsesToChatStreamHandler(c, info, resp)
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
			assert.Equal(t, 18, usage.TotalTokens)
			assert.Contains(t, recorder.Body.String(), `"code":"upstream_stream_incomplete"`)
			assert.NotContains(t, recorder.Body.String(), `"finish_reason":"stop"`)
			assert.NotContains(t, recorder.Body.String(), "data: [DONE]")

			c, recorder, resp, info = newResponsesChatTestContext(t, body, false)
			usage, apiErr = OaiResponsesToChatBufferedStreamHandler(c, info, resp)
			require.NotNil(t, apiErr)
			assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
			assert.Nil(t, usage)
			assert.Empty(t, recorder.Body.String())
		})
	}
}

func TestConvertedStreamsDoNotRetryOrWriteAfterClientCancellation(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })

	for _, chatUpstream := range []bool{false, true} {
		c, recorder, resp, info := newResponsesChatTestContext(t, "", true)
		ctx, cancel := context.WithCancel(c.Request.Context())
		cancel()
		c.Request = c.Request.WithContext(ctx)
		if chatUpstream {
			usage, apiErr := OaiChatToResponsesStreamHandler(c, info, resp)
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
		} else {
			usage, apiErr := OaiResponsesToChatStreamHandler(c, info, resp)
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
		}
		assert.Empty(t, recorder.Body.String())
	}
}

func TestConvertedStreamsClientWriteFailureStopsImmediately(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })
	for _, tc := range []struct {
		name         string
		chatUpstream bool
		format       types.RelayFormat
	}{
		{"Chat to Responses", true, types.RelayFormatOpenAIResponses},
		{"Responses to Chat", false, types.RelayFormatOpenAI},
		{"Responses to Claude", false, types.RelayFormatClaude},
		{"Responses to Gemini", false, types.RelayFormatGemini},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := "data: {\"type\":\"response.output_text.delta\",\"delta\":\"answer\"}\n\ndata: {\"type\":\"response.completed\"}\n\n"
			if tc.chatUpstream {
				body = "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"answer\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"
			}
			c, recorder, resp, info := newResponsesChatTestContext(t, body, true)
			info.RelayFormat = tc.format
			writeErr := errors.New("client socket disconnected")
			writer := &responsesDisconnectedWriter{ResponseWriter: c.Writer, err: writeErr}
			c.Writer = writer
			var usage *dto.Usage
			var apiErr *types.NewAPIError
			if tc.chatUpstream {
				usage, apiErr = OaiChatToResponsesStreamHandler(c, info, resp)
			} else {
				usage, apiErr = OaiResponsesToChatStreamHandler(c, info, resp)
			}
			require.Nil(t, apiErr, "a client write failure must not retry or refund")
			require.NotNil(t, usage)
			assert.Equal(t, 1, writer.attempts)
			assert.Empty(t, recorder.Body.String())
			assert.True(t, info.StreamStatus.HasErrors())
		})
	}
}

func TestConvertedStreamsZeroUsageTailCannotReplaceBillingSnapshot(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })
	for _, chatUpstream := range []bool{true, false} {
		t.Run(map[bool]string{true: "Chat to Responses", false: "Responses to Chat"}[chatUpstream], func(t *testing.T) {
			body := "data: " + streamUsageTextChunk + "\n\ndata: " + streamUsageEmptyChunk + "\n\ndata: " + streamUsageFinalChunk + "\n\ndata: " + streamUsageZeroChunk + "\n\ndata: [DONE]\n\n"
			if !chatUpstream {
				body = "data: {\"type\":\"response.created\",\"response\":{\"usage\":{\"input_tokens\":5000,\"output_tokens\":60,\"total_tokens\":5060,\"input_tokens_details\":{\"cached_tokens\":4800}}}}\n\n" +
					"data: {\"type\":\"response.output_text.delta\",\"delta\":\"answer\"}\n\n" +
					"data: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\",\"usage\":{\"input_tokens\":0,\"output_tokens\":0,\"total_tokens\":0}}}\n\n"
			}
			c, _, resp, info := newResponsesChatTestContext(t, body, true)
			info.SetEstimatePromptTokens(11)
			var usage *dto.Usage
			var apiErr *types.NewAPIError
			if chatUpstream {
				usage, apiErr = OaiChatToResponsesStreamHandler(c, info, resp)
			} else {
				usage, apiErr = OaiResponsesToChatStreamHandler(c, info, resp)
			}
			require.Nil(t, apiErr)
			require.NotNil(t, usage)
			assert.Equal(t, 5000, usage.PromptTokens)
			assert.Equal(t, 60, usage.CompletionTokens)
			require.NotNil(t, usage.BillingUsage)
			require.NotNil(t, usage.BillingUsage.OpenAIUsage)
			snapshot := usage.BillingUsage.OpenAIUsage
			if chatUpstream {
				assert.Equal(t, 4800, snapshot.PromptTokensDetails.CachedTokens)
			} else {
				require.NotNil(t, snapshot.InputTokensDetails)
				assert.Equal(t, 4800, snapshot.InputTokensDetails.CachedTokens)
			}
		})
	}
}

func TestChatToResponsesStreamKeepsUsageCarriedByError(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })
	errFrame := strings.Replace(streamUsageFinalChunk, `"choices":[]`, `"error":{"type":"server_error","message":"private upstream failure"},"choices":[]`, 1)
	c, w, resp, info := newResponsesChatTestContext(t, "data: "+streamUsageTextChunk+"\n\ndata: "+errFrame+"\n\n", true)
	info.SetEstimatePromptTokens(11)
	usage, apiErr := OaiChatToResponsesStreamHandler(c, info, resp)
	require.Nil(t, apiErr, "a partial stream must settle instead of entering failure refund handling")
	require.NotNil(t, usage)
	assert.Equal(t, 5000, usage.PromptTokens)
	assert.Equal(t, 60, usage.CompletionTokens)
	require.NotNil(t, usage.BillingUsage)
	require.NotNil(t, usage.BillingUsage.OpenAIUsage)
	assert.Equal(t, 4800, usage.BillingUsage.OpenAIUsage.PromptTokensDetails.CachedTokens)
	assert.Contains(t, w.Body.String(), "event: error\n")
	assert.NotContains(t, w.Body.String(), "event: response.completed\n")
}
