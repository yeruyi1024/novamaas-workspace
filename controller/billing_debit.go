package controller

import (
	"context"
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// No confirmation amount means preview. A collection must name the exact
// account and the outstanding total displayed by that preview.
func ReviewedBillingDebits(c *gin.Context) {
	identity, ok := middleware.GetSessionAuthIdentity(c)
	if !ok || c.GetInt("role") < common.RoleAdminUser || identity.UserID != c.GetInt("id") {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	var input struct {
		Month        string `json:"month"`
		Account      int    `json:"account"`
		ConfirmUser  int    `json:"confirm_user_id"`
		ConfirmQuota *int64 `json:"confirm_quota"`
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 2048), &input); err != nil {
		billingError(c, model.ErrBillingConflict)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	result, err := service.RunReviewedDebits(ctx, input.Month, input.Account, identity.UserID, input.ConfirmUser, input.ConfirmQuota)
	if err != nil {
		billingError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}
