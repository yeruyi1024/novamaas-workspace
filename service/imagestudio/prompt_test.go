package imagestudio

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptOptimizationValidatesPermissionsAndResponseWithoutCreatingImages(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Channel{}, &model.Ability{}, &model.Model{}, &model.Vendor{}, &model.Log{}, &model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}))
	settings := config.GlobalConfig.Get("image_studio_setting").(*operation_setting.ImageStudioSetting)
	previousSettings, previousLogDB := *settings, model.LOG_DB
	model.LOG_DB = db
	t.Cleanup(func() { *settings = previousSettings; model.LOG_DB = previousLogDB })
	require.NoError(t, db.Create(&model.User{Id: 611, Username: "prompt-owner", Group: "default", Status: common.UserStatusEnabled}).Error)
	channel := model.Channel{Type: constant.ChannelTypeOpenAI, Name: "prompt-fixture", Status: common.ChannelStatusEnabled, Models: "gpt-4o-mini,gpt-4o,dall-e-3", Group: "vip"}
	require.NoError(t, db.Create(&channel).Error)
	for _, name := range []string{"gpt-4o-mini", "gpt-4o", "dall-e-3"} {
		require.NoError(t, db.Create(&model.Ability{Group: "vip", Model: name, ChannelId: channel.Id, Enabled: true}).Error)
	}
	token := model.Token{UserId: 611, Key: "prompt-fixture-key", Status: common.TokenStatusEnabled, ExpiredTime: -1, UnlimitedQuota: true, Group: "vip", ModelLimitsEnabled: true, ModelLimits: "gpt-4o-mini,dall-e-3"}
	foreign := model.Token{UserId: 612, Key: "foreign-prompt-key", Status: common.TokenStatusEnabled, ExpiredTime: -1, UnlimitedQuota: true, Group: "vip"}
	require.NoError(t, db.Create(&token).Error)
	require.NoError(t, db.Create(&foreign).Error)
	model.InvalidatePricingCache()
	t.Cleanup(model.InvalidatePricingCache)
	textModels, err := ModelsForToken(611, token.Id, constant.EndpointTypeOpenAI)
	require.NoError(t, err)
	assert.Equal(t, []string{"gpt-4o-mini"}, textModels)
	imageModels, err := ModelsForToken(611, token.Id, constant.EndpointTypeImageGeneration)
	require.NoError(t, err)
	assert.Equal(t, []string{"dall-e-3"}, imageModels)
	require.NoError(t, db.Create(&model.Log{UserId: 611, Type: model.LogTypeConsume, RequestId: "prompt-request", Quota: 29}).Error)

	for _, tc := range []struct {
		name, prompt, model, body, code string
		tokenID, responseStatus         int
		disabled                        bool
	}{
		{name: "successful text optimization", prompt: " A mountain ", model: "gpt-4o-mini", tokenID: token.Id, body: `{"choices":[{"message":{"role":"assistant","content":"A sunlit mountain"},"finish_reason":"stop"}]}`},
		{name: "root disabled creation", prompt: "A mountain", model: "gpt-4o-mini", tokenID: token.Id, disabled: true, code: "image_studio_disabled"},
		{name: "empty prompt", prompt: " ", model: "gpt-4o-mini", tokenID: token.Id, code: "invalid_prompt"},
		{name: "oversized prompt", prompt: strings.Repeat("山", operation_setting.DefaultImageStudioMaxPromptCharacters+1), model: "gpt-4o-mini", tokenID: token.Id, code: "invalid_prompt"},
		{name: "foreign key", prompt: "A mountain", model: "gpt-4o-mini", tokenID: foreign.Id, code: "token_unavailable"},
		{name: "image model cannot optimize", prompt: "A mountain", model: "dall-e-3", tokenID: token.Id, code: "model_unavailable"},
		{name: "model blocked by key", prompt: "A mountain", model: "gpt-4o", tokenID: token.Id, code: "model_unavailable"},
		{name: "upstream failure", prompt: "A mountain", model: "gpt-4o-mini", tokenID: token.Id, responseStatus: http.StatusBadGateway, body: `{"error":{"message":"Provider unavailable"}}`, code: "assistant_failed"},
		{name: "missing choices", prompt: "A mountain", model: "gpt-4o-mini", tokenID: token.Id, body: `{}`, code: "invalid_assistant_response"},
		{name: "empty result", prompt: "A mountain", model: "gpt-4o-mini", tokenID: token.Id, body: `{"choices":[{"message":{"content":" "}}]}`, code: "invalid_assistant_response"},
		{name: "truncated result", prompt: "A mountain", model: "gpt-4o-mini", tokenID: token.Id, body: `{"choices":[{"message":{"content":"Incomplete"},"finish_reason":"length"}]}`, code: "invalid_assistant_response"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			settings.Enabled = !tc.disabled
			called := false
			relay := func(_ RelayContext, selected *model.Token, path string, body []byte, format types.RelayFormat) RelayResult {
				called = true
				assert.Equal(t, token.Id, selected.Id)
				assert.Equal(t, "/v1/chat/completions", path)
				assert.Equal(t, types.RelayFormat(types.RelayFormatOpenAI), format)
				var request dto.GeneralOpenAIRequest
				require.NoError(t, common.Unmarshal(body, &request))
				assert.Equal(t, "gpt-4o-mini", request.Model)
				require.Len(t, request.Messages, 2)
				assert.Equal(t, "A mountain", request.Messages[1].StringContent())
				require.NotNil(t, request.Stream)
				assert.False(t, *request.Stream)
				status := tc.responseStatus
				if status == 0 {
					status = http.StatusOK
				}
				return RelayResult{StatusCode: status, Body: []byte(tc.body), RequestID: "prompt-request"}
			}
			view, err := AssistPrompt(RelayContext{Context: context.Background()}, 611, PromptAssistInput{TokenID: tc.tokenID, Model: tc.model, Prompt: tc.prompt}, relay)
			if tc.code == "" {
				require.NoError(t, err)
				assert.Equal(t, "A sunlit mountain", view.SuggestedPrompt)
				assert.Equal(t, 29, view.Quota)
			} else {
				require.Error(t, err)
				assert.Equal(t, tc.code, ErrorCode(err))
				assert.Equal(t, tc.code == "assistant_failed" || tc.code == "invalid_assistant_response", called)
			}
		})
	}
	for _, table := range []string{"image_studio_runs", "image_studio_results", "media_assets"} {
		var count int64
		require.NoError(t, db.Table(table).Count(&count).Error)
		assert.Zero(t, count)
	}
}
