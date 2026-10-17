package openai

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relay/helper"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func OaiResponsesHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	defer service.CloseResponseBodyGracefully(resp)

	// read response body
	var responsesResponse dto.OpenAIResponsesResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeReadResponseBodyFailed, http.StatusInternalServerError)
	}
	err = common.Unmarshal(responseBody, &responsesResponse)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	if oaiError := responsesResponse.GetOpenAIError(); oaiError != nil && oaiError.Type != "" {
		return nil, types.WithOpenAIError(*oaiError, resp.StatusCode)
	}

	// 写入新的 response body
	service.IOCopyBytesGracefully(c, resp, responseBody)

	// compute usage
	usage := dto.Usage{}
	if responsesResponse.Usage != nil {
		usage.PromptTokens = responsesResponse.Usage.InputTokens
		usage.CompletionTokens = responsesResponse.Usage.OutputTokens
		usage.TotalTokens = responsesResponse.Usage.TotalTokens
		if responsesResponse.Usage.InputTokensDetails != nil {
			usage.PromptTokensDetails.CachedTokens = responsesResponse.Usage.InputTokensDetails.CachedTokens
			usage.PromptTokensDetails.CacheWriteTokens = responsesResponse.Usage.InputTokensDetails.CacheWriteTokens
		}
	}
	// Count actual tool invocations from Output (not tool declarations).
	for _, output := range responsesResponse.Output {
		switch output.Type {
		case dto.BuildInCallWebSearchCall:
			info.CountBillableToolCall(dto.BuildInCallWebSearchCall, "")
		case dto.BuildInCallFileSearchCall:
			info.CountBillableToolCall(dto.BuildInCallFileSearchCall, "")
		case dto.BuildInCallFunctionCall:
			info.CountBillableToolCall(dto.BuildInCallFunctionCall, output.Name)
		}
	}

	imageCounter := &relaycommon.ImageGenerationCallCounter{}
	if !relaycommon.IsNonBillableResponsesStatus(responsesResponse.Status) {
		for i := range responsesResponse.Output {
			idx := i
			imageCounter.Observe(&responsesResponse.Output[i], &idx)
		}
	}
	imageCounter.Commit(info)

	return &usage, nil
}

func OaiResponsesStreamHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	if resp == nil || resp.Body == nil {
		logger.LogError(c, "invalid response or response body")
		return nil, types.NewError(fmt.Errorf("invalid response"), types.ErrorCodeBadResponse)
	}

	defer service.CloseResponseBodyGracefully(resp)

	var usage = &dto.Usage{}
	var responseTextBuilder strings.Builder
	imageCounter := &relaycommon.ImageGenerationCallCounter{}
	imageCommitted := false
	terminalReceived := false
	clientWriteFailed := false
	var nextSequenceNumber int64

	helper.StreamScannerHandler(c, resp, info, func(data string, sr *helper.StreamResult) {
		var streamResponse struct {
			dto.ResponsesStreamResponse
			SequenceNumber *int64 `json:"sequence_number"`
		}
		if err := common.UnmarshalJsonStr(data, &streamResponse); err != nil {
			logger.LogError(c, "failed to unmarshal stream response: "+err.Error())
			sr.Stop(err)
			return
		}
		if streamResponse.Type == "" {
			sr.Stop(fmt.Errorf("Responses stream event is missing its type"))
			return
		}
		if streamResponse.SequenceNumber != nil && *streamResponse.SequenceNumber >= nextSequenceNumber {
			nextSequenceNumber = *streamResponse.SequenceNumber
		}
		if nextSequenceNumber < math.MaxInt64 {
			nextSequenceNumber++
		}

		// Failed/incomplete responses may also carry authoritative usage. Keep it
		// even if the stream ends before a completed event can arrive.
		if streamResponse.Response != nil && streamResponse.Response.Usage != nil {
			responseUsage := streamResponse.Response.Usage
			if responseUsage.InputTokens != 0 || responseUsage.OutputTokens != 0 || responseUsage.TotalTokens != 0 {
				if responseUsage.InputTokens != 0 {
					usage.PromptTokens = responseUsage.InputTokens
				}
				if responseUsage.OutputTokens != 0 {
					usage.CompletionTokens = responseUsage.OutputTokens
				}
				if responseUsage.TotalTokens != 0 {
					usage.TotalTokens = responseUsage.TotalTokens
				}
				if responseUsage.InputTokensDetails != nil {
					usage.PromptTokensDetails.CachedTokens = responseUsage.InputTokensDetails.CachedTokens
					usage.PromptTokensDetails.CacheWriteTokens = responseUsage.InputTokensDetails.CacheWriteTokens
				}
			}
		}
		var terminalError error
		switch streamResponse.Type {
		case "response.completed", "response.done":
			terminalReceived = true
			if streamResponse.Response != nil {
				var status string
				if err := common.Unmarshal(streamResponse.Response.Status, &status); err == nil {
					switch strings.ToLower(strings.TrimSpace(status)) {
					case "failed", "cancelled", "canceled":
						terminalError = fmt.Errorf("Responses stream ended with status %s", status)
					}
				}
				if !imageCommitted {
					if relaycommon.IsNonBillableResponsesStatus(streamResponse.Response.Status) {
						imageCounter.Reset()
						imageCounter.Commit(info)
						imageCommitted = true
					} else {
						for i := range streamResponse.Response.Output {
							idx := i
							imageCounter.Observe(&streamResponse.Response.Output[i], &idx)
						}
						imageCounter.Commit(info)
						imageCommitted = true
					}
				}
			} else if !imageCommitted {
				imageCounter.Commit(info)
				imageCommitted = true
			}
		case "response.failed", "response.incomplete", "response.cancelled", "response.canceled", "response.error", "error":
			terminalReceived = true
			// Incomplete is a protocol terminal (e.g. max_output_tokens), not
			// a transport failure. Preserve its details and finish the stream.
			if streamResponse.Type != "response.incomplete" {
				terminalError = fmt.Errorf("Responses stream ended with %s", streamResponse.Type)
			}
			if !imageCommitted {
				imageCounter.Reset()
				imageCounter.Commit(info)
				imageCommitted = true
			}
		case "response.output_text.delta":
			// 处理输出文本
			responseTextBuilder.WriteString(streamResponse.Delta)
		case dto.ResponsesOutputTypeItemDone:
			if streamResponse.Item != nil {
				switch streamResponse.Item.Type {
				case dto.BuildInCallWebSearchCall:
					info.CountBillableToolCall(dto.BuildInCallWebSearchCall, "")
				case dto.BuildInCallFileSearchCall:
					info.CountBillableToolCall(dto.BuildInCallFileSearchCall, "")
				case dto.BuildInCallFunctionCall:
					info.CountBillableToolCall(dto.BuildInCallFunctionCall, streamResponse.Item.Name)
				case dto.ResponsesOutputTypeImageGenerationCall:
					if !imageCommitted {
						imageCounter.Observe(streamResponse.Item, streamResponse.OutputIndex)
					}
				}
			}
		}
		if err := helper.ResponseChunkData(c, streamResponse.ResponsesStreamResponse, data); err != nil {
			clientWriteFailed = true
			if c.Request.Context().Err() != nil {
				info.StreamStatus.SetEndReason(relaycommon.StreamEndReasonClientGone, c.Request.Context().Err())
			}
			sr.Stop(err)
			return
		}
		if terminalError != nil {
			sr.Stop(terminalError)
		} else if terminalReceived {
			sr.Done()
		}
	})

	if !terminalReceived && !clientWriteFailed && c.Request.Context().Err() == nil &&
		info.StreamStatus.EndReason != relaycommon.StreamEndReasonClientGone &&
		info.StreamStatus.EndReason != relaycommon.StreamEndReasonPingFail {
		// Transport EOF/[DONE] alone cannot complete a Responses response. Keep
		// the transport reason for diagnostics and expose protocol failure in SSE.
		message := "The upstream response stream ended before completion."
		code := "upstream_stream_incomplete"
		if info.StreamStatus.EndReason == relaycommon.StreamEndReasonTimeout {
			code = "upstream_stream_timeout"
			message = "The upstream response stream timed out before completion."
		}
		info.StreamStatus.RecordError("Responses stream ended without a terminal response event")
		logger.LogError(c, "Responses stream ended without a terminal response event: "+info.StreamStatus.Summary())
		if !c.Writer.Written() {
			// A retry is safe only before any response bytes (including pings)
			// have been sent. Restore the headers for a retry or JSON error.
			helper.ResetEventStreamHeaders(c)
			return nil, types.NewOpenAIError(fmt.Errorf("%s", message), types.ErrorCodeBadResponseBody, http.StatusBadGateway)
		}
		if err := helper.WriteResponsesStreamError(c, code, message, nextSequenceNumber); err != nil {
			info.StreamStatus.RecordError("failed to write Responses stream error: " + err.Error())
		}
	}

	if usage.CompletionTokens == 0 {
		// 计算输出文本的 token 数量
		tempStr := responseTextBuilder.String()
		if len(tempStr) > 0 {
			// 非正常结束，使用输出文本的 token 数量
			completionTokens := service.CountTextToken(tempStr, info.UpstreamModelName)
			usage.CompletionTokens = completionTokens
		}
	}

	if usage.PromptTokens == 0 && usage.CompletionTokens != 0 {
		usage.PromptTokens = info.GetEstimatePromptTokens()
	}

	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens

	return usage, nil
}
