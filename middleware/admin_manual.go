package middleware

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// AdminManualAuth deliberately uses the application's ordinary missing-route
// response for every denial, including invalid credentials and disabled policy.
func AdminManualAuth(notFound gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Vary", "Authorization")
		c.Header("X-Robots-Tag", "noindex, nofollow, noarchive")
		enabled, err := model.AdminManualEnabled(c.Request.Context())
		if err != nil {
			logger.LogError(c.Request.Context(), "admin manual policy lookup failed: "+err.Error())
		}
		if !enabled || err != nil {
			notFound(c)
			c.Abort()
			return
		}
		user, identity, useAccessToken, err := authenticateDashboardRequest(c)
		if err != nil || user == nil || user.Status != common.UserStatusEnabled ||
			user.Role < common.RoleAdminUser || !validUserInfo(user.Username, user.Role) {
			notFound(c)
			c.Abort()
			return
		}
		setDashboardAuthContext(c, user, identity, useAccessToken)
		c.Next()
	}
}
