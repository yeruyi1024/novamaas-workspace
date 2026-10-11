package imagestudio

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

type PromptAssistInput struct {
	TokenID int    `json:"token_id"`
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
}

type PromptAssistView struct {
	SuggestedPrompt string `json:"suggested_prompt"`
	RequestID       string `json:"request_id"`
	Quota           int    `json:"quota"`
}

// AssistPrompt uses normal text relay billing and creates no image or asset records.
func AssistPrompt(execution RelayContext, userID int, input PromptAssistInput, relay RelayExecutor) (*PromptAssistView, error) {
	settings := operation_setting.GetImageStudioSetting()
	if !settings.Enabled {
		return nil, &RequestError{http.StatusServiceUnavailable, "image_studio_disabled", errors.New("image generation is disabled")}
	}
	input.Prompt = strings.TrimSpace(input.Prompt)
	if input.Prompt == "" || len([]rune(input.Prompt)) > operation_setting.DefaultImageStudioMaxPromptCharacters {
		return nil, &RequestError{http.StatusBadRequest, "invalid_prompt", fmt.Errorf("prompt must contain 1 to %d characters", operation_setting.DefaultImageStudioMaxPromptCharacters)}
	}
	token, err := resolveToken(userID, input.TokenID)
	if err != nil {
		return nil, err
	}
	if err = validateModel(userID, token, input.Model, constant.EndpointTypeOpenAI); err != nil {
		return nil, err
	}
	body, err := common.Marshal(dto.GeneralOpenAIRequest{
		Model: input.Model, Stream: common.GetPointer(false), MaxTokens: common.GetPointer(uint(2048)),
		Messages: []dto.Message{
			{Role: "system", Content: fmt.Sprintf("You are an image-generation prompt editor. Return only a concise improved prompt in the user's language, no explanations or markdown. Preserve the user's intent and named subjects. Clarify composition, lighting, materials, colors and style where useful without inventing private facts. Keep the result within %d characters.", operation_setting.DefaultImageStudioMaxPromptCharacters)},
			{Role: "user", Content: input.Prompt},
		},
	})
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(execution.Context, 2*time.Minute)
	defer cancel()
	execution.Context = ctx
	result := relay(execution, token, "/v1/chat/completions", body, types.RelayFormatOpenAI)
	if result.StatusCode >= http.StatusBadRequest {
		return nil, &RequestError{result.StatusCode, "assistant_failed", errors.New(relayErrorMessage(result.Body))}
	}
	var response dto.OpenAITextResponse
	if err = common.Unmarshal(result.Body, &response); err != nil || len(response.Choices) == 0 {
		return nil, &RequestError{http.StatusBadGateway, "invalid_assistant_response", errors.New("text model returned an invalid response")}
	}
	suggested := strings.TrimSpace(response.Choices[0].Message.StringContent())
	if suggested == "" || len([]rune(suggested)) > operation_setting.DefaultImageStudioMaxPromptCharacters || response.Choices[0].FinishReason == "length" {
		return nil, &RequestError{http.StatusBadGateway, "invalid_assistant_response", errors.New("text model returned an empty, incomplete, or oversized prompt")}
	}
	return &PromptAssistView{SuggestedPrompt: suggested, RequestID: result.RequestID, Quota: QuotaForRequest(result.RequestID, userID)}, nil
}
