package controller

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service/authz"
	studioService "github.com/QuantumNous/new-api/service/imagestudio"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Studio HTTP fixtures keep transport/auth assertions separate from service tests.
func studioRequest(handler gin.HandlerFunc, user, role int, method, path, id, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	c.Set("id", user)
	c.Set("role", role)
	c.Params = gin.Params{{Key: "id", Value: id}, {Key: "task_id", Value: "request:image-private"}}
	handler(c)
	return recorder
}

func TestStudioReadPermissionsAndAudit(t *testing.T) {
	db := setupManageUserTestDB(t)
	previousMaster := common.IsMasterNode
	common.IsMasterNode = true
	t.Cleanup(func() { common.IsMasterNode = previousMaster })
	require.NoError(t, authz.Init(db))
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}, &model.TaskRequestBody{}))
	run := model.ImageStudioRun{PublicID: "private", UserID: 7, IdempotencyKey: "private", Prompt: "private prompt", ErrorMessage: "private error", Status: model.ImageStudioRunStatusCompleted}
	require.NoError(t, db.Create(&run).Error)
	require.NoError(t, db.Create(&model.ImageStudioResult{PublicID: "image", RunID: run.ID}).Error)
	require.NoError(t, model.SaveRequestSnapshots("", "image-private", model.RequestSnapshotSourceImageStudio, []byte(`{"prompt":"private prompt"}`), []byte(`{"model":"upstream"}`)))
	for _, tc := range []struct {
		name                  string
		user, role            int
		review, read, content bool
		status                int
	}{
		{"owner", 7, common.RoleCommonUser, false, false, false, 200},
		{"foreign workspace", 8, common.RoleRootUser, false, false, false, 404},
		{"ordinary admin", 8, common.RoleAdminUser, true, false, false, 403},
		{"metadata only", 8, common.RoleAdminUser, true, true, false, 200},
		{"authorized admin", 8, common.RoleAdminUser, true, true, true, 200},
		{"revoked content", 8, common.RoleAdminUser, true, true, false, 200},
		{"ordinary user", 8, common.RoleCommonUser, true, true, true, 403},
		{"root review", 8, common.RoleRootUser, true, false, false, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, authz.SetUserPermissions(tc.user, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: tc.read, authz.ActionContentRead: tc.content}}))
			path := "/?request_id=image-private"
			if tc.review {
				path += "&scope=all"
			}
			response := studioRequest(GetImageStudioRun, tc.user, tc.role, http.MethodGet, path, run.PublicID, "")
			require.Equal(t, tc.status, response.Code)
			hasContent := tc.status == 200 && (!tc.review || tc.content || tc.role == common.RoleRootUser)
			assert.Equal(t, hasContent, strings.Contains(response.Body.String(), "private prompt"))
			if !hasContent {
				assert.NotContains(t, response.Body.String(), "private.png")
			}
			for _, handler := range []gin.HandlerFunc{GetLogRequestSnapshots, GetLogRequestBody, GetTaskRequestBody} {
				snapshot := studioRequest(handler, tc.user, tc.role, http.MethodGet, path, "", "")
				allowed := tc.role == common.RoleRootUser || (tc.role == common.RoleAdminUser && tc.read && tc.content)
				assert.Equal(t, allowed, snapshot.Code == 200)
				assert.Equal(t, allowed, strings.Contains(snapshot.Body.String(), "private prompt"))
			}
		})
	}
	var logs []model.Log
	require.NoError(t, db.Where("type = ?", model.LogTypeManage).Find(&logs).Error)
	require.NotEmpty(t, logs)
	assert.Contains(t, logs[0].Other, "image_studio.content_read")
	// Read grants never allow changes to another user's result, including Root.
	for _, role := range []int{common.RoleAdminUser, common.RoleRootUser} {
		assert.Equal(t, 404, studioRequest(SaveImageStudioResult, 8, role, http.MethodPost, "/", "image", `{"group_id":"foreign"}`).Code)
		assert.Equal(t, 404, studioRequest(DeleteImageStudioResult, 8, role, http.MethodDelete, "/", "image", "").Code)
	}
}

func TestStudioListOwnerFilterAndThumbnails(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}))
	for _, owner := range []int{7, 8} {
		name := "artist-" + strconv.Itoa(owner)
		require.NoError(t, db.Create(&model.User{Id: owner, Username: name, AffCode: name}).Error)
		run := model.ImageStudioRun{PublicID: name, UserID: owner, IdempotencyKey: name, Status: model.ImageStudioRunStatusCompleted}
		require.NoError(t, db.Create(&run).Error)
		require.NoError(t, db.Create(&model.ImageStudioResult{RunID: run.ID}).Error)
	}
	for _, tc := range []struct {
		query                      string
		role, status, count, owner int
	}{
		{"scope=all", common.RoleRootUser, 200, 2, 0},
		{"scope=all&user_id=8", common.RoleRootUser, 200, 1, 8},
		{"scope=all&user_id=99", common.RoleRootUser, 200, 0, 0},
		{"user_id=8", common.RoleCommonUser, 200, 1, 7},
		{"scope=all", common.RoleCommonUser, 403, 0, 0},
		{"scope=all&user_id=-1", common.RoleRootUser, 400, 0, 0},
	} {
		response := studioRequest(ListImageStudioRuns, 7, tc.role, http.MethodGet, "/?"+tc.query, "", "")
		require.Equal(t, tc.status, response.Code)
		if tc.status != 200 {
			continue
		}
		var body struct{ Data studioService.RunList }
		require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
		assert.EqualValues(t, tc.count, body.Data.Total)
		require.Len(t, body.Data.Items, tc.count)
		assert.NotContains(t, response.Body.String(), "original")
		if tc.owner > 0 {
			assert.Equal(t, tc.owner, body.Data.Items[0].UserID)
		}
		if tc.role == common.RoleRootUser {
			assert.Len(t, body.Data.Users, 2)
		} else {
			assert.Empty(t, body.Data.Users)
		}
	}
}

func TestImageStudioRelayUsesCompanyTokenAuthentication(t *testing.T) {
	previousDB := model.DB
	initModelListColumnNames(t)
	model.DB = previousDB
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Token{}))
	for _, tc := range []struct {
		name, allowedIP, clientIP string
		status, want              int
	}{
		{"disabled key", "", "192.168.1.5", common.TokenStatusDisabled, http.StatusUnauthorized},
		{"foreign IP", "192.168.1.4", "192.168.1.5", common.TokenStatusEnabled, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token := &model.Token{UserId: 7, Key: common.GetUUID(), Status: tc.status, UnlimitedQuota: true, ExpiredTime: -1, AllowIps: &tc.allowedIP}
			require.NoError(t, db.Create(token).Error)
			result := RelayImageStudio(studioService.RelayContext{Context: context.Background(), ClientIP: tc.clientIP}, token, "/v1/images/generations", []byte(`{"model":"dall-e-3","prompt":"Mountain"}`), types.RelayFormatOpenAIImage)
			assert.Equal(t, tc.want, result.StatusCode)
			assert.NotEmpty(t, result.RequestID)
		})
	}
}

func TestImageStudioDefaultRequestUsesSelectedProviderProtocol(t *testing.T) {
	for _, tc := range []struct {
		name, model, path, body, response string
		override                          bool
		channelType                       int
	}{
		{name: "mapped override", override: true, channelType: constant.ChannelTypeOpenAI, model: "mapped-image", path: "/v1/images/generations", body: `{"model":"mapped-image","prompt":"upstream mountain","n":1}`},
		{name: "OpenAI compatible", channelType: constant.ChannelTypeOpenAI, model: "gpt-image-1", path: "/v1/images/generations", body: `{"model":"gpt-image-1","prompt":"mountain","n":1}`},
		{name: "Seedream", channelType: constant.ChannelTypeVolcEngine, model: "doubao-seedream-4-5", path: "/api/v3/images/generations", body: `{"model":"doubao-seedream-4-5","prompt":"mountain","n":1}`},
		{name: "Baidu", channelType: constant.ChannelTypeBaiduV2, model: "qwen-image", path: "/v2/images/generations", body: `{"model":"qwen-image","prompt":"mountain","n":1}`},
		{name: "Ali Qwen", channelType: constant.ChannelTypeAli, model: "qwen-image-plus", path: "/api/v1/services/aigc/multimodal-generation/generation", body: `{"model":"qwen-image-plus","input":{"messages":[{"role":"user","content":[{"text":"mountain"}]}]},"parameters":{"n":1}}`, response: `{"output":{"results":[{"url":"https://example.test/result.png"}]},"usage":{"image_count":1}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupVolcNativeControllerTest(t)
			require.NoError(t, db.AutoMigrate(&model.TaskRequestBody{}, &model.CostAccountingSnapshot{}))
			prices, previousLog := ratio_setting.ModelPrice2JSONString(), common.LogConsumeEnabled
			common.LogConsumeEnabled = true
			require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(`{"studio-default":0}`))
			t.Cleanup(func() {
				require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(prices))
				common.LogConsumeEnabled = previousLog
			})
			var sent []byte
			var sentPath string
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var err error
				sent, err = io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				sentPath = r.URL.Path
				w.Header().Set("Content-Type", gin.MIMEJSON)
				response := tc.response
				if response == "" {
					response = `{"data":[{"url":"https://example.test/result.png"}]}`
				}
				_, err = io.WriteString(w, response)
				assert.NoError(t, err)
			}))
			t.Cleanup(upstream.Close)
			channel := &model.Channel{Type: tc.channelType, Key: "test-key", BaseURL: &upstream.URL}
			require.NoError(t, db.Create(channel).Error)
			user := &model.User{Username: "studio-provider", Quota: 1000}
			require.NoError(t, db.Create(user).Error)
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", strings.NewReader(`{"model":"studio-default","prompt":"mountain"}`))
			c.Request.Header.Set("Content-Type", gin.MIMEJSON)
			c.Set(common.RequestIdKey, "studio-provider-default")
			common.SetContextKey(c, constant.ContextKeyUserId, user.Id)
			common.SetContextKey(c, constant.ContextKeyUserQuota, user.Quota)
			common.SetContextKey(c, constant.ContextKeyUserSetting, dto.UserSetting{BillingPreference: "wallet_only"})
			common.SetContextKey(c, constant.ContextKeyUsingGroup, "default")
			common.SetContextKey(c, constant.ContextKeyUserGroup, "default")
			common.SetContextKey(c, constant.ContextKeyOriginalModel, "studio-default")
			common.SetContextKey(c, constant.ContextKeyTokenUnlimited, true)
			common.SetContextKey(c, constant.ContextKeyChannelId, channel.Id)
			common.SetContextKey(c, constant.ContextKeyChannelType, channel.Type)
			common.SetContextKey(c, constant.ContextKeyChannelBaseUrl, upstream.URL)
			common.SetContextKey(c, constant.ContextKeyChannelKey, channel.Key)
			mapping, err := common.Marshal(map[string]string{"studio-default": tc.model})
			require.NoError(t, err)
			common.SetContextKey(c, constant.ContextKeyChannelModelMapping, string(mapping))
			original := `{"request":{"model":"studio-default","prompt":"mountain"}}`
			common.SetContextKey(c, constant.ContextKeyImageStudioOriginalRequestBody, original)
			if tc.override {
				common.SetContextKey(c, constant.ContextKeyChannelParamOverride, map[string]interface{}{"prompt": "upstream mountain"})
			}

			Relay(c, types.RelayFormatOpenAIImage)

			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			assert.Equal(t, tc.path, sentPath)
			assert.JSONEq(t, tc.body, string(sent))
			snapshots, exists, err := model.GetArchivedRequestSnapshots("", "studio-provider-default")
			require.NoError(t, err)
			require.True(t, exists)
			assert.JSONEq(t, original, string(snapshots.Original))
			assert.JSONEq(t, string(sent), string(snapshots.Upstream))
			assert.NotContains(t, string(snapshots.Upstream), "test-key")
			var log model.Log
			require.NoError(t, db.Where("request_id = ? AND type = ?", "studio-provider-default", model.LogTypeConsume).First(&log).Error)
			assert.Contains(t, log.Other, `"request_body_available":true`)
			assert.NotContains(t, log.Other, "mountain")
			var response dto.ImageResponse
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
			require.Len(t, response.Data, 1)
			assert.Equal(t, "https://example.test/result.png", response.Data[0].Url)
		})
	}
}
