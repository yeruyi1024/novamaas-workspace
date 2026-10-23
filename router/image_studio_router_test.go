package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	studioService "github.com/QuantumNous/new-api/service/imagestudio"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageStudioPublicSurfaceKeepsModelsAndPersonalDataAuthenticated(t *testing.T) {
	previousLimit := common.GlobalApiRateLimitEnable
	common.GlobalApiRateLimitEnable = false
	t.Cleanup(func() { common.GlobalApiRateLimitEnable = previousLimit })
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	SetApiRouter(engine)
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/image-studio/settings", nil))
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var settings struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &settings))
	assert.True(t, settings.Success)
	assert.Equal(t, map[string]any{
		"enabled":               operation_setting.GetImageStudioSetting().Enabled,
		"max_images":            float64(operation_setting.ImageStudioMaxImages),
		"max_running_runs":      float64(operation_setting.ImageStudioMaxRunningPerUser),
		"max_prompt_characters": float64(operation_setting.ImageStudioMaxPromptCharacters),
		"max_reference_images":  float64(studioService.MaxReferenceImages),
		"max_reference_bytes":   float64(studioService.MaxReferenceImageBytes),
	}, settings.Data)
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/models"},
		{http.MethodGet, "/runs"},
		{http.MethodGet, "/runs/poll"},
		{http.MethodGet, "/references"},
		{http.MethodGet, "/logs"},
		{http.MethodGet, "/runs/private-run"},
		{http.MethodPost, "/runs"},
		{http.MethodPost, "/prompt-assist"},
		{http.MethodPost, "/references"},
		{http.MethodDelete, "/runs/private-run"},
		{http.MethodPost, "/results/private-result/save"},
		{http.MethodDelete, "/results/private-result"},
	} {
		t.Run(route.method+route.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest(route.method, "/api/image-studio"+route.path, strings.NewReader(`{}`)))
			assert.Equal(t, http.StatusUnauthorized, recorder.Code, recorder.Body.String())
		})
	}
}
