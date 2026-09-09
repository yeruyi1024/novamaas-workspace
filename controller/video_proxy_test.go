package controller

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupVideoProxyControllerTest(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}, &model.Channel{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)

	previousDB := model.DB
	previousCache := common.MemoryCacheEnabled
	previousRelayTimeout := common.RelayTimeout
	previousFetchSetting := *system_setting.GetFetchSetting()
	model.DB = db
	common.MemoryCacheEnabled = false
	common.RelayTimeout = 0
	system_setting.GetFetchSetting().EnableSSRFProtection = false
	service.InitHttpClient()

	t.Cleanup(func() {
		model.DB = previousDB
		common.MemoryCacheEnabled = previousCache
		common.RelayTimeout = previousRelayTimeout
		*system_setting.GetFetchSetting() = previousFetchSetting
		service.InitHttpClient()
		_ = sqlDB.Close()
	})
	return db
}

func insertVideoProxyTask(t *testing.T, db *gorm.DB, mode dto.VideoContentDeliveryMode, resultURL string) {
	insertVideoProxyTaskForChannel(t, db, constant.ChannelTypeDoubaoVideo, mode, resultURL)
}

func insertVideoProxyTaskForChannel(t *testing.T, db *gorm.DB, channelType int, mode dto.VideoContentDeliveryMode, resultURL string) {
	t.Helper()
	channel := &model.Channel{
		Id:     channelType,
		Type:   channelType,
		Name:   "video-channel",
		Status: common.ChannelStatusEnabled,
	}
	channel.SetOtherSettings(dto.ChannelOtherSettings{VideoContentDeliveryMode: mode})
	require.NoError(t, db.Create(channel).Error)
	require.NoError(t, db.Create(&model.Task{
		TaskID:    "task_public",
		UserId:    1,
		ChannelId: channel.Id,
		Status:    model.TaskStatusSuccess,
		PrivateData: model.TaskPrivateData{
			ResultURL: resultURL,
		},
	}).Error)
}

func runVideoProxyRequest() *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task_public/content", nil)
	ctx.Params = gin.Params{{Key: "task_id", Value: "task_public"}}
	ctx.Set("id", 1)
	VideoProxy(ctx)
	return recorder
}

func runVideoContentInfoRequest() *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task_public/content-info", nil)
	ctx.Params = gin.Params{{Key: "task_id", Value: "task_public"}}
	ctx.Set("id", 1)
	GetVideoContentInfo(ctx)
	return recorder
}

func TestVideoProxyAllowsAdminToDownloadAnotherUsersTask(t *testing.T) {
	db := setupVideoProxyControllerTest(t)
	channel := &model.Channel{
		Id:     54,
		Type:   constant.ChannelTypeDoubaoVideo,
		Name:   "doubao-video",
		Status: common.ChannelStatusEnabled,
	}
	require.NoError(t, db.Create(channel).Error)
	require.NoError(t, db.Create(&model.Task{
		TaskID:    "task_other_user",
		UserId:    99,
		ChannelId: channel.Id,
		Status:    model.TaskStatusSuccess,
		PrivateData: model.TaskPrivateData{
			ResultURL: "data:video/mp4;base64,dmlkZW8=",
		},
	}).Error)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task_other_user/content", nil)
	c.Params = gin.Params{{Key: "task_id", Value: "task_other_user"}}
	c.Set("id", 1)
	c.Set("role", common.RoleAdminUser)

	VideoProxy(c)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "video", recorder.Body.String())
}

func TestVideoProxyRedirectsDoubaoContentWithoutFetchingIt(t *testing.T) {
	db := setupVideoProxyControllerTest(t)
	var upstreamRequests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstreamRequests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)
	resultURL := upstream.URL + "/video.mp4?signature=test"
	insertVideoProxyTask(t, db, dto.VideoContentDeliveryModeRedirect, resultURL)

	recorder := runVideoProxyRequest()

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, resultURL, recorder.Header().Get("Location"))
	assert.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
	assert.Zero(t, upstreamRequests.Load())
}

func TestVideoProxyRedirectsPublicVideoByDefault(t *testing.T) {
	db := setupVideoProxyControllerTest(t)
	var upstreamRequests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstreamRequests.Add(1)
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = w.Write([]byte("video-data"))
	}))
	t.Cleanup(upstream.Close)
	insertVideoProxyTask(t, db, "", upstream.URL+"/video.mp4")

	recorder := runVideoProxyRequest()

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, upstream.URL+"/video.mp4", recorder.Header().Get("Location"))
	assert.Zero(t, upstreamRequests.Load())
}

func TestVideoProxyPreservesExplicitServerProxyCompatibility(t *testing.T) {
	db := setupVideoProxyControllerTest(t)
	var upstreamRequests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstreamRequests.Add(1)
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = w.Write([]byte("video-data"))
	}))
	t.Cleanup(upstream.Close)
	insertVideoProxyTask(t, db, dto.VideoContentDeliveryModeProxy, upstream.URL+"/video.mp4")

	recorder := runVideoProxyRequest()

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "video/mp4", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "video-data", recorder.Body.String())
	assert.EqualValues(t, 1, upstreamRequests.Load())
}

func TestVideoContentInfoExposesDoubaoRedirectURLAfterAuthorization(t *testing.T) {
	db := setupVideoProxyControllerTest(t)
	resultURL := "https://cdn.example.com/video.mp4?signature=test"
	insertVideoProxyTask(t, db, dto.VideoContentDeliveryModeRedirect, resultURL)

	recorder := runVideoContentInfoRequest()

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{
		"success": true,
		"message": "",
		"data": {
			"delivery_mode": "redirect",
			"url": "https://cdn.example.com/video.mp4?signature=test"
		}
	}`, recorder.Body.String())
}

func TestVideoContentInfoDoesNotExposeURLForProxyDelivery(t *testing.T) {
	db := setupVideoProxyControllerTest(t)
	insertVideoProxyTask(t, db, dto.VideoContentDeliveryModeProxy, "https://cdn.example.com/video.mp4")

	recorder := runVideoContentInfoRequest()

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{
		"success": true,
		"message": "",
		"data": {
			"delivery_mode": "proxy"
		}
	}`, recorder.Body.String())
}

func TestVideoContentInfoRedirectsAliPublicResourceByDefault(t *testing.T) {
	db := setupVideoProxyControllerTest(t)
	insertVideoProxyTaskForChannel(t, db, constant.ChannelTypeAli, "", "https://cdn.example.com/ali.mp4")

	recorder := runVideoContentInfoRequest()

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{
		"success": true,
		"message": "",
		"data": {
			"delivery_mode": "redirect",
			"url": "https://cdn.example.com/ali.mp4"
		}
	}`, recorder.Body.String())
}

func TestVideoContentInfoKeepsCredentialedProvidersOnProxy(t *testing.T) {
	db := setupVideoProxyControllerTest(t)
	insertVideoProxyTaskForChannel(t, db, constant.ChannelTypeGemini, "", "https://cdn.example.com/gemini.mp4")

	recorder := runVideoContentInfoRequest()

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{
		"success": true,
		"message": "",
		"data": {"delivery_mode": "proxy"}
	}`, recorder.Body.String())
}
