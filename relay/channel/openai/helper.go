package openai

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	moonshotfacade "github.com/QuantumNous/new-api/relay/facade/moonshot"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/relayconvert"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

// 辅助函数
func HandleStreamFormat(c *gin.Context, info *relaycommon.RelayInfo, data string, forceFormat bool, thinkToContent bool) error {
	info.SendResponseCount++

	switch info.RelayFormat {
	case types.RelayFormatOpenAI:
		return sendStreamData(c, info, data, forceFormat, thinkToContent)
	case types.RelayFormatClaude:
		return handleClaudeFormat(c, data, info)
	case types.RelayFormatGemini:
		return handleGeminiFormat(c, data, info)
	}
	return nil
}

// writeTextStreamError reports a partial text-stream failure in the client's
// protocol, without generating a successful completion or another response.
func writeTextStreamError(c *gin.Context, info *relaycommon.RelayInfo, code, message string) error {
	helper.ExtendWriteDeadline(c)
	switch info.RelayFormat {
	case types.RelayFormatClaude:
		return helper.ClaudeData(c, dto.ClaudeResponse{
			Type: "error", Error: types.ClaudeError{Type: "api_error", Message: message},
		})
	case types.RelayFormatGemini:
		return helper.ObjectData(c, gin.H{"error": gin.H{
			"code": http.StatusBadGateway, "message": message, "status": "UNAVAILABLE",
		}})
	default:
		return helper.ObjectData(c, gin.H{"error": types.OpenAIError{
			Type: "upstream_error", Code: code, Message: message,
		}})
	}
}

func handleClaudeFormat(c *gin.Context, data string, info *relaycommon.RelayInfo) error {
	var streamResponse dto.ChatCompletionsStreamResponse
	if err := common.Unmarshal(common.StringToByteSlice(data), &streamResponse); err != nil {
		return err
	}

	if service.ValidUsage(streamResponse.Usage) {
		applyUsagePostProcessing(info, streamResponse.Usage, common.StringToByteSlice(data))
		info.ClaudeConvertInfo.Usage = streamResponse.Usage
	} else {
		// Empty usage placeholders must not replace a previously reported usage
		// or make the converter emit terminal events with zero token counts.
		streamResponse.Usage = nil
	}
	// Usage snapshots can precede completion. Keep them for settlement, but
	// do not let a usage-only frame close a still-unfinished Claude response.
	// HandleFinalResponse bypasses this gate once [DONE]/finish_reason is known.
	if len(streamResponse.Choices) == 0 && info.EnsureClaudeConvertInfo().FinishReason == "" {
		return nil
	}
	result, err := relayconvert.ConvertStreamResponse(c, info, types.RelayFormatClaude, &streamResponse)
	if err != nil {
		return err
	}
	claudeResponses, ok := result.Value.([]*dto.ClaudeResponse)
	if !ok {
		return fmt.Errorf("expected Claude stream responses, got %T", result.Value)
	}
	for _, resp := range claudeResponses {
		if err := helper.ClaudeData(c, *resp); err != nil {
			return err
		}
	}
	return nil
}

func handleGeminiFormat(c *gin.Context, data string, info *relaycommon.RelayInfo) error {
	var streamResponse dto.ChatCompletionsStreamResponse
	if err := common.Unmarshal(common.StringToByteSlice(data), &streamResponse); err != nil {
		logger.LogError(c, "failed to unmarshal stream response: "+err.Error())
		return err
	}

	result, err := relayconvert.ConvertStreamResponse(c, info, types.RelayFormatGemini, &streamResponse)
	if err != nil {
		return err
	}
	geminiResponse, ok := result.Value.(*dto.GeminiChatResponse)
	if !ok {
		return fmt.Errorf("expected Gemini stream response, got %T", result.Value)
	}

	// 如果返回 nil，表示没有实际内容，跳过发送
	if geminiResponse == nil {
		return nil
	}

	geminiResponseStr, err := common.Marshal(geminiResponse)
	if err != nil {
		logger.LogError(c, "failed to marshal gemini response: "+err.Error())
		return err
	}

	// send gemini format response
	return helper.StringData(c, string(geminiResponseStr))
}

func ProcessStreamResponse(streamResponse dto.ChatCompletionsStreamResponse, responseTextBuilder *strings.Builder, toolCount *int) error {
	for _, choice := range streamResponse.Choices {
		responseTextBuilder.WriteString(choice.Delta.GetContentString())
		responseTextBuilder.WriteString(choice.Delta.GetReasoningContent())
		if choice.Delta.ToolCalls != nil {
			if len(choice.Delta.ToolCalls) > *toolCount {
				*toolCount = len(choice.Delta.ToolCalls)
			}
			for _, tool := range choice.Delta.ToolCalls {
				responseTextBuilder.WriteString(tool.Function.Name)
				responseTextBuilder.WriteString(tool.Function.Arguments)
			}
		}
	}
	return nil
}

func processTokenData(relayMode int, data string, responseTextBuilder *strings.Builder, toolCount *int) error {
	switch relayMode {
	case relayconstant.RelayModeChatCompletions:
		var streamResponse dto.ChatCompletionsStreamResponse
		if err := common.UnmarshalJsonStr(data, &streamResponse); err != nil {
			return err
		}
		return ProcessStreamResponse(streamResponse, responseTextBuilder, toolCount)
	case relayconstant.RelayModeCompletions:
		var streamResponse dto.CompletionsStreamResponse
		if err := common.UnmarshalJsonStr(data, &streamResponse); err != nil {
			return err
		}
		processCompletionsStreamResponse(streamResponse, responseTextBuilder)
	}
	return nil
}

func processCompletionsStreamResponse(streamResponse dto.CompletionsStreamResponse, responseTextBuilder *strings.Builder) {
	for _, choice := range streamResponse.Choices {
		responseTextBuilder.WriteString(choice.Text)
	}
}

func handleLastResponse(lastStreamData string, responseId *string, createAt *int64,
	systemFingerprint *string, model *string, usage **dto.Usage,
	containStreamUsage *bool, info *relaycommon.RelayInfo,
	shouldSendLastResp *bool) error {

	var lastStreamResponse dto.ChatCompletionsStreamResponse
	if err := common.Unmarshal(common.StringToByteSlice(lastStreamData), &lastStreamResponse); err != nil {
		return err
	}

	*responseId = lastStreamResponse.Id
	*createAt = lastStreamResponse.Created
	*systemFingerprint = lastStreamResponse.GetSystemFingerprint()
	*model = lastStreamResponse.Model

	if service.ValidUsage(lastStreamResponse.Usage) {
		*containStreamUsage = true
		*usage = lastStreamResponse.Usage
		if !info.ShouldIncludeUsage {
			// Only a usage-only frame may be suppressed. A mixed frame can
			// contain finish_reason or the final tool arguments without text.
			*shouldSendLastResp = len(lastStreamResponse.Choices) > 0
		}
	}

	return nil
}

func HandleFinalResponse(c *gin.Context, info *relaycommon.RelayInfo, lastStreamData string,
	responseId string, createAt int64, model string, systemFingerprint string,
	usage *dto.Usage, containStreamUsage bool) error {

	switch info.RelayFormat {
	case types.RelayFormatOpenAI:
		if shouldEmitSyntheticStreamUsage(c, info, containStreamUsage) {
			response := helper.GenerateFinalUsageResponse(responseId, createAt, model, *usage)
			response.SetSystemFingerprint(systemFingerprint)
			if err := helper.ObjectData(c, response); err != nil {
				return err
			}
		}
		return helper.StringData(c, "[DONE]")

	case types.RelayFormatClaude:
		var streamResponse dto.ChatCompletionsStreamResponse
		if err := common.Unmarshal(common.StringToByteSlice(lastStreamData), &streamResponse); err != nil {
			return err
		}

		info.ClaudeConvertInfo.Usage = usage
		streamResponse.Usage = usage

		result, err := relayconvert.ConvertStreamResponse(c, info, types.RelayFormatClaude, &streamResponse)
		if err != nil {
			return err
		}
		claudeResponses, ok := result.Value.([]*dto.ClaudeResponse)
		if !ok {
			return fmt.Errorf("expected Claude stream responses, got %T", result.Value)
		}
		for _, resp := range claudeResponses {
			if err := helper.ClaudeData(c, *resp); err != nil {
				return err
			}
		}
		info.ClaudeConvertInfo.Done = true

	case types.RelayFormatGemini:
		var streamResponse dto.ChatCompletionsStreamResponse
		if err := common.Unmarshal(common.StringToByteSlice(lastStreamData), &streamResponse); err != nil {
			return err
		}

		// 这里处理的是 openai 最后一个流响应，其 delta 为空，有 finish_reason 字段
		// 因此相比较于 google 官方的流响应，由 openai 转换而来会多一个 parts 为空，finishReason 为 STOP 的响应
		// 而包含最后一段文本输出的响应（倒数第二个）的 finishReason 为 null
		// 暂不知是否有程序会不兼容。

		result, err := relayconvert.ConvertStreamResponse(c, info, types.RelayFormatGemini, &streamResponse)
		if err != nil {
			return err
		}
		geminiResponse, ok := result.Value.(*dto.GeminiChatResponse)
		if !ok {
			return fmt.Errorf("expected Gemini stream response, got %T", result.Value)
		}

		// openai 流响应开头的空数据
		if geminiResponse == nil {
			return nil
		}

		geminiResponseStr, err := common.Marshal(geminiResponse)
		if err != nil {
			return err
		}

		// 发送最终的 Gemini 响应
		return helper.StringData(c, string(geminiResponseStr))
	}
	return nil
}

func shouldEmitSyntheticStreamUsage(c *gin.Context, info *relaycommon.RelayInfo, containStreamUsage bool) bool {
	return info.ShouldIncludeUsage && !containStreamUsage && allowsSyntheticUsage(c)
}

func allowsSyntheticUsage(c *gin.Context) bool {
	return !moonshotfacade.KimiPassthroughEnabled(c)
}

func sendResponsesStreamData(c *gin.Context, streamResponse dto.ResponsesStreamResponse, data string) {
	if data == "" {
		return
	}
	_ = helper.ResponseChunkData(c, streamResponse, data)
}
