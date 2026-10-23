package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminManualAuthHidesExpiredAccessTokenWithoutPATFallback(t *testing.T) {
	setupDashboardAuthMiddlewareTest(t)
	require.NoError(t, model.DB.AutoMigrate(&model.Option{}))
	require.NoError(t, model.DB.Create(&model.Option{Key: model.AdminManualEnabledOptionKey, Value: "true"}).Error)
	expired := issueExpiredDashboardAccessToken(t, service.AuthIdentity{UserID: 42, SessionID: "expired-manual-session", UserAuthVersion: 1, SessionVersion: 1})
	user := createMiddlewarePATUser(t, "manual-expired-fallback", expired)
	require.NoError(t, model.DB.Model(user).Update("role", common.RoleAdminUser).Error)
	engine := gin.New()
	engine.GET("/manual", DisableCache(), AdminManualAuth(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "missing"})
	}), func(c *gin.Context) {
		c.String(http.StatusOK, "private manual content")
	})
	request := httptest.NewRequest(http.MethodGet, "/manual", nil)
	request.Header.Set("Authorization", "Bearer "+expired)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.JSONEq(t, `{"error":"missing"}`, response.Body.String())
	assert.NotContains(t, response.Body.String(), "AUTH_TOKEN_EXPIRED")
	assert.NotContains(t, response.Body.String(), "private manual content")
}
