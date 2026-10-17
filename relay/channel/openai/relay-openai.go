package openai

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/relay/channel/openrouter"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	moonshotfacade "github.com/QuantumNous/new-api/relay/facade/moonshot"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/relayconvert"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func sendStreamData(c *gin.Context, info *relaycommon.RelayInfo, data string, forceFormat bool, thinkToContent bool) error {
	if data == "" {
		return nil
	}

	if !forceFormat && !thinkToContent {
		return helper.StringData(c, data)
	}

	var lastStreamResponse dto.ChatCompletionsStreamResponse
	if err := common.UnmarshalJsonStr(data, &lastStreamResponse); err != nil {
		return err
	}

	if !thinkToContent {
		return helper.ObjectData(c, lastStreamResponse)
	}

	hasThinkingContent := false
	hasContent := false
	var thinkingContent strings.Builder
	for _, choice := range lastStreamResponse.Choices {
		if len(choice.Delta.GetReasoningContent()) > 0 {
			hasThinkingContent = true
			thinkingContent.WriteString(choice.Delta.GetReasoningContent())
		}
		if len(choice.Delta.GetContentString()) > 0 {
			hasContent = true
		}
	}

	// Handle think to content conversion
	if info.ThinkingContentInfo.IsFirstThinkingContent {
		if hasThinkingContent {
			response := lastStreamResponse.Copy()
			for i := range response.Choices {
				// send `think` tag with thinking content
				response.Choices[i].Delta.SetContentString("<think>\n" + thinkingContent.String())
				response.Choices[i].Delta.ReasoningContent = nil
				response.Choices[i].Delta.Reasoning = nil
			}
			info.ThinkingContentInfo.IsFirstThinkingContent = false
			info.ThinkingContentInfo.HasSentThinkingContent = true
			return helper.ObjectData(c, response)
		}
	}

	if lastStreamResponse.Choices == nil || len(lastStreamResponse.Choices) == 0 {
		return helper.ObjectData(c, lastStreamResponse)
	}

	// Process each choice
	for i, choice := range lastStreamResponse.Choices {
		// Handle transition from thinking to content
		// only send `</think>` tag when previous thinking content has been sent
		if hasContent && !info.ThinkingContentInfo.SendLastThinkingContent && info.ThinkingContentInfo.HasSentThinkingContent {
			response := lastStreamResponse.Copy()
			for j := range response.Choices {
				response.Choices[j].Delta.SetContentString("\n</think>\n")
				response.Choices[j].Delta.ReasoningContent = nil
				response.Choices[j].Delta.Reasoning = nil
			}
			info.ThinkingContentInfo.SendLastThinkingContent = true
			if err := helper.ObjectData(c, response); err != nil {
				return err
			}
		}

		// Convert reasoning content to regular content if any
		if len(choice.Delta.GetReasoningContent()) > 0 {
			lastStreamResponse.Choices[i].Delta.SetContentString(choice.Delta.GetReasoningContent())
			lastStreamResponse.Choices[i].Delta.ReasoningContent = nil
			lastStreamResponse.Choices[i].Delta.Reasoning = nil
		} else if !hasThinkingContent && !hasContent {
			// flush thinking content
			lastStreamResponse.Choices[i].Delta.ReasoningContent = nil
			lastStreamResponse.Choices[i].Delta.Reasoning = nil
		}
	}

	return helper.ObjectData(c, lastStreamResponse)
}

func OaiStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (resultUsage *dto.Usage, apiErr *types.NewAPIError) {
	if resp == nil || resp.Body == nil {
		logger.LogError(c, "invalid response or response body")
		return nil, types.NewOpenAIError(fmt.Errorf("invalid response"), types.ErrorCodeBadResponse, http.StatusInternalServerError)
	}

	defer service.CloseResponseBodyGracefully(resp)
	kimiPassthrough := moonshotfacade.KimiPassthroughEnabled(c)
	forceFormat := info.ChannelSetting.ForceFormat && !kimiPassthrough
	thinkingToContent := info.ChannelSetting.ThinkingToContent && !kimiPassthrough

	model := info.UpstreamModelName
	var responseId string
	var createAt int64 = 0
	var systemFingerprint string
	var containStreamUsage bool
	var responseTextBuilder strings.Builder
	var toolCount int
	var usage = &dto.Usage{}
	var usageStreamData string
	var lastStreamData string
	seenStreamToolCalls := make(map[string]struct{})
	var streamFunctionCallNames []string
	defer func() {
		// A retryable attempt has not been settled. Its tool observations must
		// not carry into the next channel and charge the user twice.
		if apiErr == nil {
			for _, name := range streamFunctionCallNames {
				info.CountBillableToolCall(dto.BuildInCallFunctionCall, name)
			}
		}
	}()
	finishedChoices := make(map[int]bool)
	clientWriteFailed := false

	helper.StreamScannerHandler(c, resp, info, func(data string, sr *helper.StreamResult) {
		var streamResponse struct {
			Usage   *dto.Usage         `json:"usage"`
			Error   *types.OpenAIError `json:"error"`
			Choices []struct {
				Index        int     `json:"index"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}
		parseErr := common.UnmarshalJsonStr(data, &streamResponse)
		if parseErr == nil && service.ValidUsage(streamResponse.Usage) {
			// The current frame has already arrived from upstream. Retain its
			// authoritative usage even if writing the pending frame fails.
			usage = streamResponse.Usage
			usageStreamData = data
			containStreamUsage = true
		}
		if lastStreamData != "" {
			if err := HandleStreamFormat(c, info, lastStreamData, forceFormat, thinkingToContent); err != nil {
				common.SysLog("error handling stream format: " + err.Error())
				clientWriteFailed = true
				if c.Request.Context().Err() != nil {
					info.StreamStatus.SetEndReason(relaycommon.StreamEndReasonClientGone, c.Request.Context().Err())
				}
				sr.Stop(err)
				return
			}
			lastStreamData = ""
		}
		if len(data) > 0 {
			if parseErr != nil {
				sr.Stop(parseErr)
				return
			}
			if streamResponse.Error != nil {
				sr.Stop(fmt.Errorf("upstream returned an error inside the OpenAI stream"))
				return
			}
			for _, choice := range streamResponse.Choices {
				finishedChoices[choice.Index] = finishedChoices[choice.Index] ||
					(choice.FinishReason != nil && strings.TrimSpace(*choice.FinishReason) != "")
			}

			lastStreamData = data
			collectStreamFunctionCallNames(data, seenStreamToolCalls, &streamFunctionCallNames)
			if err := processTokenData(info.RelayMode, data, &responseTextBuilder, &toolCount); err != nil {
				logger.LogError(c, "error processing stream token data: "+err.Error())
				sr.Stop(err)
			}
		}
	})

	// Chat accepts an explicit [DONE], or a graceful EOF after every requested
	// choice supplied finish_reason. EOF/timeout alone cannot prove completion.
	streamCompleted := info.StreamStatus.EndReason == relaycommon.StreamEndReasonDone
	if info.StreamStatus.EndReason == relaycommon.StreamEndReasonEOF && len(finishedChoices) > 0 {
		expectedChoices := 1
		if request, ok := info.Request.(*dto.GeneralOpenAIRequest); ok && request != nil && request.N != nil && *request.N > 1 {
			expectedChoices = *request.N
		}
		streamCompleted = len(finishedChoices) >= expectedChoices
		for _, finished := range finishedChoices {
			if !finished {
				streamCompleted = false
				break
			}
		}
	}
	streamCompleted = streamCompleted && !info.StreamStatus.HasErrors()

	// 处理最后的响应
	shouldSendLastResp := true
	if streamCompleted {
		if err := handleLastResponse(lastStreamData, &responseId, &createAt, &systemFingerprint, &model, &usage,
			&containStreamUsage, info, &shouldSendLastResp); err != nil {
			logger.LogError(c, fmt.Sprintf("error handling last response: %s, lastStreamData: [%s]", err.Error(), lastStreamData))
		}
	}
	if kimiPassthrough {
		shouldSendLastResp = true
	}

	if streamCompleted && c.Request.Context().Err() == nil && info.RelayFormat == types.RelayFormatOpenAI {
		if shouldSendLastResp {
			if err := sendStreamData(c, info, lastStreamData, forceFormat, thinkingToContent); err != nil {
				clientWriteFailed = true
				info.StreamStatus.RecordError(err.Error())
			}
		}
	}

	if !containStreamUsage {
		usage = service.ResponseText2Usage(c, responseTextBuilder.String(), info.UpstreamModelName, info.GetEstimatePromptTokens())
		usage.CompletionTokens += toolCount * 7
	}

	if containStreamUsage {
		applyUsagePostProcessing(info, usage, common.StringToByteSlice(usageStreamData))
	}
	applyUsagePostProcessing(info, usage, common.StringToByteSlice(lastStreamData))

	if c.Request.Context().Err() != nil || clientWriteFailed || info.StreamStatus.EndReason == relaycommon.StreamEndReasonPingFail {
		return usage, nil
	}
	if !streamCompleted {
		info.StreamStatus.RecordError("OpenAI stream ended without a valid terminal signal")
		logger.LogError(c, "OpenAI stream ended without a valid terminal signal: "+info.StreamStatus.Summary())
		message := "The upstream stream ended before the response was complete."
		code := "upstream_stream_incomplete"
		if info.StreamStatus.EndReason == relaycommon.StreamEndReasonTimeout {
			code = "upstream_stream_timeout"
			message = "The upstream response stream timed out before completion."
		}
		if !c.Writer.Written() {
			helper.ResetEventStreamHeaders(c)
			return nil, types.NewOpenAIError(fmt.Errorf("%s", message), types.ErrorCodeBadResponseBody, http.StatusBadGateway)
		}
		if err := writeTextStreamError(c, info, code, message); err != nil {
			info.StreamStatus.RecordError("failed to write OpenAI stream error: " + err.Error())
		}
		return usage, nil
	}

	if err := HandleFinalResponse(c, info, lastStreamData, responseId, createAt, model, systemFingerprint, usage, containStreamUsage); err != nil {
		info.StreamStatus.RecordError("failed to write OpenAI stream terminal: " + err.Error())
	}

	return usage, nil
}

func collectStreamFunctionCallNames(data string, seen map[string]struct{}, names *[]string) {
	var streamResponse dto.ChatCompletionsStreamResponse
	if err := common.UnmarshalJsonStr(data, &streamResponse); err != nil {
		return
	}
	for _, choice := range streamResponse.Choices {
		for i, tc := range choice.Delta.ToolCalls {
			name := tc.Function.Name
			if name == "" {
				continue
			}
			toolIdx := i
			if tc.Index != nil {
				toolIdx = *tc.Index
			}
			key := fmt.Sprintf("%d-%d", choice.Index, toolIdx)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			*names = append(*names, name)
		}
	}
}

func OpenaiHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)

	var simpleResponse dto.OpenAITextResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeReadResponseBodyFailed, http.StatusInternalServerError)
	}
	logger.LogDebug(c, "upstream response body: %s", responseBody)
	// Unmarshal to simpleResponse
	if info.ChannelType == constant.ChannelTypeOpenRouter && info.ChannelOtherSettings.IsOpenRouterEnterprise() {
		// 尝试解析为 openrouter enterprise
		var enterpriseResponse openrouter.OpenRouterEnterpriseResponse
		err = common.Unmarshal(responseBody, &enterpriseResponse)
		if err != nil {
			return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
		}
		if enterpriseResponse.Success {
			responseBody = enterpriseResponse.Data
		} else {
			logger.LogError(c, fmt.Sprintf("openrouter enterprise response success=false, data: %s", enterpriseResponse.Data))
			return nil, types.NewOpenAIError(fmt.Errorf("openrouter response success=false"), types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
		}
	}

	err = common.Unmarshal(responseBody, &simpleResponse)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	if oaiError := simpleResponse.GetOpenAIError(); oaiError != nil && oaiError.Type != "" {
		return nil, types.WithOpenAIError(*oaiError, resp.StatusCode)
	}

	for _, choice := range simpleResponse.Choices {
		if choice.FinishReason == constant.FinishReasonContentFilter {
			common.SetContextKey(c, constant.ContextKeyAdminRejectReason, "openai_finish_reason=content_filter")
			break
		}
	}

	for _, choice := range simpleResponse.Choices {
		for _, tc := range choice.Message.ParseToolCalls() {
			info.CountBillableToolCall(dto.BuildInCallFunctionCall, tc.Function.Name)
		}
	}

	forceFormat := info.ChannelSetting.ForceFormat && !moonshotfacade.KimiPassthroughEnabled(c)

	usageModified := false
	if simpleResponse.Usage.PromptTokens == 0 && allowsSyntheticUsage(c) {
		completionTokens := simpleResponse.Usage.CompletionTokens
		if completionTokens == 0 {
			for _, choice := range simpleResponse.Choices {
				ctkm := service.CountTextToken(choice.Message.StringContent()+choice.Message.GetReasoningContent(), info.UpstreamModelName)
				completionTokens += ctkm
			}
		}
		simpleResponse.Usage = dto.Usage{
			PromptTokens:     info.GetEstimatePromptTokens(),
			CompletionTokens: completionTokens,
			TotalTokens:      info.GetEstimatePromptTokens() + completionTokens,
		}
		usageModified = true
	}

	applyUsagePostProcessing(info, &simpleResponse.Usage, responseBody)

	switch info.RelayFormat {
	case types.RelayFormatOpenAI:
		if usageModified {
			var bodyMap map[string]interface{}
			err = common.Unmarshal(responseBody, &bodyMap)
			if err != nil {
				return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
			}
			bodyMap["usage"] = simpleResponse.Usage
			responseBody, _ = common.Marshal(bodyMap)
		}
		if forceFormat {
			responseBody, err = common.Marshal(simpleResponse)
			if err != nil {
				return nil, types.NewError(err, types.ErrorCodeBadResponseBody)
			}
		} else {
			break
		}
	case types.RelayFormatClaude:
		convertResult, err := relayconvert.ConvertResponse(c, info, types.RelayFormatClaude, &simpleResponse)
		if err != nil {
			return nil, types.NewError(err, types.ErrorCodeBadResponseBody)
		}
		claudeRespStr, err := common.Marshal(convertResult.Value)
		if err != nil {
			return nil, types.NewError(err, types.ErrorCodeBadResponseBody)
		}
		responseBody = claudeRespStr
	case types.RelayFormatGemini:
		convertResult, err := relayconvert.ConvertResponse(c, info, types.RelayFormatGemini, &simpleResponse)
		if err != nil {
			return nil, types.NewError(err, types.ErrorCodeBadResponseBody)
		}
		geminiRespStr, err := common.Marshal(convertResult.Value)
		if err != nil {
			return nil, types.NewError(err, types.ErrorCodeBadResponseBody)
		}
		responseBody = geminiRespStr
	}

	service.IOCopyBytesGracefully(c, resp, responseBody)

	return &simpleResponse.Usage, nil
}
