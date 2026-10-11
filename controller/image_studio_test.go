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
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service/authz"
	studioService "github.com/QuantumNous/new-api/service/imagestudio"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageStudioUsesCompanyRootOptionsAndAudit(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Option{}))
	previousMaster := common.IsMasterNode
	common.IsMasterNode = true
	settings := config.GlobalConfig.Get("image_studio_setting").(*operation_setting.ImageStudioSetting)
	previousSettings := *settings
	common.OptionMapRWMutex.Lock()
	previousOptions := common.OptionMap
	common.OptionMap = make(map[string]string)
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		*settings = previousSettings
		common.IsMasterNode = previousMaster
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previousOptions
		common.OptionMapRWMutex.Unlock()
	})
	router := gin.New()
	router.PUT("/api/option/", middleware.RootAuth(), UpdateOption)
	for _, role := range []int{common.RoleCommonUser, common.RoleAdminUser, common.RoleRootUser} {
		credential := "studio-settings-" + strconv.Itoa(role)
		user := &model.User{Username: credential, AffCode: credential, AccessToken: &credential, Role: role, Status: common.UserStatusEnabled, Group: "default"}
		require.NoError(t, db.Create(user).Error)
		for _, field := range []string{"enabled", "visible", "asset_library_enabled"} {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPut, "/api/option/", strings.NewReader(`{"key":"image_studio_setting.`+field+`","value":false}`))
			request.Header.Set("Authorization", "Bearer "+credential)
			router.ServeHTTP(recorder, request)
			if role != common.RoleRootUser {
				assert.Equal(t, http.StatusForbidden, recorder.Code)
				assert.Equal(t, previousSettings, *settings)
			} else {
				require.Equal(t, http.StatusOK, recorder.Code)
			}
		}
	}
	assert.False(t, settings.Enabled)
	assert.False(t, settings.Visible)
	assert.False(t, settings.AssetLibraryEnabled)
	var logs []model.Log
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 3)
	for _, log := range logs {
		assert.Contains(t, log.Other, "option.update")
	}

}

func TestImageStudioSaveRejectsOtherUsersEvenRoot(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}))
	run := &model.ImageStudioRun{PublicID: "run-owner", UserID: 7, IdempotencyKey: "owner"}
	require.NoError(t, db.Create(run).Error)
	result := &model.ImageStudioResult{RunID: run.ID, PublicID: "result-owner", PreviewURL: "https://example.test/image.png", ThumbnailURL: "https://example.test/image.png"}
	require.NoError(t, db.Create(result).Error)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"group_id":"group-root"}`))
	c.Set("id", 8)
	c.Set("role", common.RoleRootUser)
	c.Params = gin.Params{{Key: "id", Value: result.PublicID}}
	SaveImageStudioResult(c)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestImageStudioRunReadScopeAndContentPermission(t *testing.T) {
	db := setupManageUserTestDB(t)
	previousMaster := common.IsMasterNode
	common.IsMasterNode = true
	t.Cleanup(func() { common.IsMasterNode = previousMaster })
	require.NoError(t, authz.Init(db))
	require.NoError(t, authz.SetUserPermissions(8, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: true, authz.ActionContentRead: true}}))
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}))
	run := &model.ImageStudioRun{PublicID: "run-private", UserID: 7, IdempotencyKey: "private", Prompt: "Private prompt", ErrorMessage: "Private failure", Status: model.ImageStudioRunStatusCompleted}
	require.NoError(t, db.Create(run).Error)
	require.NoError(t, db.Create(&model.ImageStudioResult{RunID: run.ID, PreviewURL: "https://example.test/private.png"}).Error)
	require.NoError(t, authz.SetUserPermissions(10, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: true, authz.ActionContentRead: false}}))
	require.NoError(t, authz.SetUserPermissions(11, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: false}}))

	for _, tc := range []struct {
		name         string
		userID, role int
		scope        string
		status       int
		content      bool
	}{
		{"owner", 7, common.RoleCommonUser, "", http.StatusOK, true},
		{"other user", 9, common.RoleCommonUser, "", http.StatusNotFound, false},
		{"root user workspace stays private", 8, common.RoleRootUser, "", http.StatusNotFound, false},
		{"admin explicit review", 8, common.RoleAdminUser, "all", http.StatusOK, true},
		{"admin metadata only", 10, common.RoleAdminUser, "all", http.StatusOK, false},
		{"revoked review", 11, common.RoleAdminUser, "all", http.StatusForbidden, false},
		{"unassigned administrator", 12, common.RoleAdminUser, "all", http.StatusForbidden, false},
		{"ordinary user review denied", 9, common.RoleCommonUser, "all", http.StatusForbidden, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/image-studio/runs/"+run.PublicID+"?scope="+tc.scope, nil)
			c.Params = gin.Params{{Key: "id", Value: run.PublicID}}
			c.Set("id", tc.userID)
			c.Set("role", tc.role)
			GetImageStudioRun(c)
			require.Equal(t, tc.status, recorder.Code)
			if tc.status != http.StatusOK {
				return
			}
			var response struct{ Data studioService.RunView }
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
			if tc.content {
				assert.Equal(t, run.Prompt, response.Data.Prompt)
				require.Len(t, response.Data.Results, 1)
				assert.Equal(t, "https://example.test/private.png", response.Data.Results[0].PreviewURL)
			} else {
				assert.Empty(t, response.Data.Prompt)
				assert.Empty(t, response.Data.Results)
				assert.Empty(t, response.Data.ErrorMessage)
			}
		})
	}
}

func TestImageStudioReviewUserFilterPreservesScopeAndListsAllOwners(t *testing.T) {
	db := setupManageUserTestDB(t)
	previousMaster := common.IsMasterNode
	common.IsMasterNode = true
	t.Cleanup(func() { common.IsMasterNode = previousMaster })
	require.NoError(t, authz.Init(db))
	require.NoError(t, authz.SetUserPermissions(7, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: true, authz.ActionContentRead: true}}))
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}))
	upload := model.MediaAsset{PublicID: "direct-upload", OwnerUserID: 7, StorageObjectID: 1, Name: "Uploaded image", AssetType: "image"}
	saved := model.MediaAsset{PublicID: "saved-generated", OwnerUserID: 8, StorageObjectID: 2, Name: "Generated image", AssetType: "image"}
	require.NoError(t, db.Create(&upload).Error)
	require.NoError(t, db.Create(&saved).Error)
	for _, id := range []int{7, 8} {
		name := "artist-" + strconv.Itoa(id)
		require.NoError(t, db.Create(&model.User{Id: id, Username: name, AffCode: name}).Error)
		run := model.ImageStudioRun{PublicID: name, UserID: id, IdempotencyKey: name, Model: "seedream", Status: model.ImageStudioRunStatusCompleted}
		require.NoError(t, db.Create(&run).Error)
		result := model.ImageStudioResult{RunID: run.ID, PreviewURL: "https://example.test/generated.png", ThumbnailURL: "https://example.test/generated.png", RevisedPrompt: name + " image"}
		if id == 8 {
			result.MediaAssetID = saved.ID
		}
		require.NoError(t, db.Create(&result).Error)
	}
	require.NoError(t, db.Create(&model.ImageStudioRun{PublicID: "old-failure", UserID: 99, IdempotencyKey: "old-failure", Status: model.ImageStudioRunStatusFailed}).Error)
	for _, tc := range []struct {
		name, query                string
		role, status, count, owner int
	}{
		{"all by default", "scope=all", common.RoleAdminUser, http.StatusOK, 2, 0},
		{"one owner", "scope=all&user_id=8", common.RoleAdminUser, http.StatusOK, 1, 8},
		{"owner without images", "scope=all&user_id=99", common.RoleAdminUser, http.StatusOK, 0, 0},
		{"ordinary history cannot change owner", "user_id=8", common.RoleCommonUser, http.StatusOK, 1, 7},
		{"ordinary review denied", "scope=all&user_id=8", common.RoleCommonUser, http.StatusForbidden, 0, 0},
		{"invalid owner", "scope=all&user_id=-8", common.RoleAdminUser, http.StatusBadRequest, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/image-studio/runs?"+tc.query, nil)
			c.Set("id", 7)
			c.Set("role", tc.role)
			ListImageStudioRuns(c)
			require.Equal(t, tc.status, recorder.Code)
			if tc.status != http.StatusOK {
				return
			}
			var response struct{ Data studioService.RunList }
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
			assert.EqualValues(t, tc.count, response.Data.Total)
			require.Len(t, response.Data.Items, tc.count)
			for _, run := range response.Data.Items {
				require.Len(t, run.Results, 1)
				assert.Empty(t, run.Results[0].PreviewURL)
				assert.Equal(t, "https://example.test/generated.png", run.Results[0].ThumbnailURL)
				assert.Equal(t, "artist-"+strconv.Itoa(run.UserID)+" image", run.Results[0].Description)
				assert.NotEqual(t, upload.PublicID, run.Results[0].AssetID)
				if run.UserID == 8 {
					assert.Equal(t, saved.PublicID, run.Results[0].AssetID)
				} else {
					assert.Empty(t, run.Results[0].AssetID)
				}
			}
			if tc.owner > 0 {
				assert.Equal(t, tc.owner, response.Data.Items[0].UserID)
			}
			if tc.role == common.RoleAdminUser {
				require.Len(t, response.Data.Users, 2)
				assert.Equal(t, "artist-8", response.Data.Users[1].Name)
			} else {
				assert.Empty(t, response.Data.Users)
			}
		})
	}
	var logs []model.Log
	require.NoError(t, db.Where("user_id = ?", 7).Find(&logs).Error)
	require.Len(t, logs, 3)
	for _, log := range logs {
		assert.Contains(t, log.Other, "image_studio.content_read")
	}

}

func TestDeleteImageStudioResultOnlyOwnerCanDeleteAndKeepsAssetsAndRetryIdentity(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}, &model.ImageStudioResult{}, &model.MediaAsset{}))
	asset := &model.MediaAsset{PublicID: "saved-image", OwnerUserID: 7, StorageObjectID: 1, Name: "Saved image", AssetType: model.AssetTypeImage}
	require.NoError(t, db.Create(asset).Error)
	run := &model.ImageStudioRun{PublicID: "run-delete", UserID: 7, IdempotencyKey: "delete-request", Prompt: "Private prompt", Status: model.ImageStudioRunStatusCompleted}
	require.NoError(t, db.Create(run).Error)
	images := []model.ImageStudioResult{
		{PublicID: "image-first", RunID: run.ID, MediaAssetID: asset.ID, SaveStatus: "saving", SaveStartedAt: common.GetTimestamp()},
		{PublicID: "image-second", RunID: run.ID, PreviewURL: "https://example.test/private.png"},
	}
	require.NoError(t, db.Create(&images).Error)
	for _, tc := range []struct{ userID, role, status int }{
		{8, common.RoleCommonUser, http.StatusNotFound},
		{8, common.RoleAdminUser, http.StatusNotFound},
		{8, common.RoleRootUser, http.StatusNotFound},
		{7, common.RoleCommonUser, http.StatusConflict},
	} {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/image-studio/results/image-first", nil)
		c.Set("id", tc.userID)
		c.Set("role", tc.role)
		c.Params = gin.Params{{Key: "id", Value: images[0].PublicID}}
		DeleteImageStudioResult(c)
		assert.Equal(t, tc.status, recorder.Code)
	}
	require.NoError(t, db.Model(&images[0]).Update("save_status", "saved").Error)
	for index, image := range images {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/image-studio/results/"+image.PublicID, nil)
		c.Set("id", 7)
		c.Set("role", common.RoleCommonUser)
		c.Params = gin.Params{{Key: "id", Value: image.PublicID}}
		DeleteImageStudioResult(c)
		require.Equal(t, http.StatusOK, recorder.Code)
		var remaining int64
		require.NoError(t, db.Model(&model.ImageStudioResult{}).Count(&remaining).Error)
		assert.EqualValues(t, 1-index, remaining)
	}
	var persisted model.ImageStudioRun
	require.NoError(t, db.First(&persisted, run.ID).Error)
	assert.Empty(t, persisted.Prompt)
	_, _, err := studioService.GetRun(context.Background(), studioService.Actor{UserID: 7, Role: common.RoleCommonUser}, run.PublicID, false)
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, studioService.StatusCode(err))
	_, _, err = studioService.CreateRun(studioService.RelayContext{Context: context.Background()}, 7, studioService.RunInput{IdempotencyKey: run.IdempotencyKey}, nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusGone, studioService.StatusCode(err))
	require.NoError(t, db.First(&model.MediaAsset{}, asset.ID).Error)
}

func TestImageStudioRelayUsesCompanyTokenAuthentication(t *testing.T) {
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

func TestImageStudioRelayArchivesActualMappedAndOverriddenRequest(t *testing.T) {
	db := setupVolcNativeControllerTest(t)
	require.NoError(t, db.AutoMigrate(&model.TaskRequestBody{}))
	prices := ratio_setting.ModelPrice2JSONString()
	previousLogEnabled := common.LogConsumeEnabled
	require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(`{"public-image":0}`))
	common.LogConsumeEnabled = true
	t.Cleanup(func() {
		require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(prices))
		common.LogConsumeEnabled = previousLogEnabled
	})
	var sent []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		sent, err = io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Bearer upstream-secret", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", gin.MIMEJSON)
		_, err = io.WriteString(w, `{"data":[{"url":"https://example.test/result.png"}]}`)
		assert.NoError(t, err)
	}))
	defer upstream.Close()
	channel := &model.Channel{Type: constant.ChannelTypeOpenAI, Key: "upstream-secret", BaseURL: &upstream.URL}
	require.NoError(t, db.Create(channel).Error)
	user := &model.User{Username: "image-audit", Quota: 1000}
	require.NoError(t, db.Create(user).Error)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", strings.NewReader(`{"model":"public-image","prompt":"mountain","n":2}`))
	c.Request.Header.Set("Content-Type", gin.MIMEJSON)
	c.Set(common.RequestIdKey, "request_image_audit")
	common.SetContextKey(c, constant.ContextKeyUserId, user.Id)
	common.SetContextKey(c, constant.ContextKeyUserQuota, user.Quota)
	common.SetContextKey(c, constant.ContextKeyUserSetting, dto.UserSetting{BillingPreference: "wallet_only"})
	common.SetContextKey(c, constant.ContextKeyUsingGroup, "default")
	common.SetContextKey(c, constant.ContextKeyUserGroup, "default")
	common.SetContextKey(c, constant.ContextKeyOriginalModel, "public-image")
	common.SetContextKey(c, constant.ContextKeyTokenUnlimited, true)
	common.SetContextKey(c, constant.ContextKeyChannelId, channel.Id)
	common.SetContextKey(c, constant.ContextKeyChannelType, channel.Type)
	common.SetContextKey(c, constant.ContextKeyChannelBaseUrl, upstream.URL)
	common.SetContextKey(c, constant.ContextKeyChannelKey, channel.Key)
	common.SetContextKey(c, constant.ContextKeyChannelModelMapping, `{"public-image":"mapped-image"}`)
	common.SetContextKey(c, constant.ContextKeyChannelParamOverride, map[string]interface{}{"prompt": "upstream mountain"})
	original := `{"token_id":7,"idempotency_key":"audit","request":{"model":"public-image","prompt":"mountain","n":2}}`
	common.SetContextKey(c, constant.ContextKeyImageStudioOriginalRequestBody, original)
	Relay(c, types.RelayFormatOpenAIImage)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	snapshots, exists, err := model.GetArchivedRequestSnapshots("", "request_image_audit")
	require.NoError(t, err)
	require.True(t, exists)
	assert.JSONEq(t, original, string(snapshots.Original))
	assert.JSONEq(t, string(sent), string(snapshots.Upstream))
	assert.Contains(t, string(sent), `"model":"mapped-image"`)
	assert.Contains(t, string(sent), `"prompt":"upstream mountain"`)
	assert.NotContains(t, string(snapshots.Upstream), "upstream-secret")
	var log model.Log
	require.NoError(t, db.Where("request_id = ? AND type = ?", "request_image_audit", model.LogTypeConsume).First(&log).Error)
	assert.Contains(t, log.Other, `"request_body_available":true`)
	assert.NotContains(t, log.Other, "mountain")
}

func TestImageStudioDefaultRequestUsesSelectedProviderProtocol(t *testing.T) {
	for _, tc := range []struct {
		name, model, path, body, response string
		channelType                       int
	}{
		{name: "OpenAI compatible", channelType: constant.ChannelTypeOpenAI, model: "gpt-image-1", path: "/v1/images/generations", body: `{"model":"gpt-image-1","prompt":"mountain","n":1}`},
		{name: "Seedream", channelType: constant.ChannelTypeVolcEngine, model: "doubao-seedream-4-5", path: "/api/v3/images/generations", body: `{"model":"doubao-seedream-4-5","prompt":"mountain","n":1}`},
		{name: "Baidu", channelType: constant.ChannelTypeBaiduV2, model: "qwen-image", path: "/v2/images/generations", body: `{"model":"qwen-image","prompt":"mountain","n":1}`},
		{name: "Ali Qwen", channelType: constant.ChannelTypeAli, model: "qwen-image-plus", path: "/api/v1/services/aigc/multimodal-generation/generation", body: `{"model":"qwen-image-plus","input":{"messages":[{"role":"user","content":[{"text":"mountain"}]}]},"parameters":{"n":1}}`, response: `{"output":{"results":[{"url":"https://example.test/result.png"}]},"usage":{"image_count":1}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := setupVolcNativeControllerTest(t)
			require.NoError(t, db.AutoMigrate(&model.TaskRequestBody{}))
			prices := ratio_setting.ModelPrice2JSONString()
			require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(`{"studio-default":0}`))
			t.Cleanup(func() { require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(prices)) })
			var sent []byte
			var sentPath string
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var err error
				sent, err = io.ReadAll(r.Body)
				assert.NoError(t, err)
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
			common.SetContextKey(c, constant.ContextKeyImageStudioOriginalRequestBody, `{"request":{"model":"studio-default","prompt":"mountain"}}`)

			Relay(c, types.RelayFormatOpenAIImage)

			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			assert.Equal(t, tc.path, sentPath)
			assert.JSONEq(t, tc.body, string(sent))
			var response dto.ImageResponse
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
			require.Len(t, response.Data, 1)
			assert.Equal(t, "https://example.test/result.png", response.Data[0].Url)
		})
	}
}

func TestImageRequestSnapshotEndpointsEnforceRootGrantedContentPermission(t *testing.T) {
	db := setupManageUserTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.TaskRequestBody{}))
	previousMaster := common.IsMasterNode
	common.IsMasterNode = true
	t.Cleanup(func() { common.IsMasterNode = previousMaster })
	require.NoError(t, authz.Init(db))
	require.NoError(t, authz.SetUserPermissions(8, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: true, authz.ActionContentRead: true}}))
	require.NoError(t, authz.SetUserPermissions(9, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: true}}))
	require.NoError(t, model.SaveRequestSnapshots("", "image-private", model.RequestSnapshotSourceImageStudio, []byte(`{"request":{"prompt":"private prompt"}}`), []byte(`{"model":"upstream"}`)))
	for _, handler := range []gin.HandlerFunc{GetLogRequestSnapshots, GetLogRequestBody, GetTaskRequestBody} {
		for _, tc := range []struct{ userID, role, status int }{
			{1, common.RoleRootUser, http.StatusOK},
			{8, common.RoleAdminUser, http.StatusOK},
			{9, common.RoleAdminUser, http.StatusForbidden},
			{10, common.RoleAdminUser, http.StatusForbidden},
			{8, common.RoleCommonUser, http.StatusForbidden},
		} {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/log/request-snapshots?request_id=image-private", nil)
			c.Params = gin.Params{{Key: "task_id", Value: "request:image-private"}}
			c.Set("id", tc.userID)
			c.Set("role", tc.role)
			handler(c)
			assert.Equal(t, tc.status, recorder.Code)
			if tc.status == http.StatusOK {
				assert.Contains(t, recorder.Body.String(), "private prompt")
			} else {
				assert.NotContains(t, recorder.Body.String(), "private prompt")
			}
		}
	}
	require.NoError(t, authz.SetUserPermissions(8, authz.PermissionsMap{authz.ResourceImageStudio: {authz.ActionRead: true, authz.ActionContentRead: false}}))
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/log/request-snapshots?request_id=image-private", nil)
	c.Set("id", 8)
	c.Set("role", common.RoleAdminUser)
	GetLogRequestSnapshots(c)
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}
