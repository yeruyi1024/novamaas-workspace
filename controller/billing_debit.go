package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type debitReviewInput struct {
	Month        string `json:"month"`
	Account      int    `json:"account"`
	Request      string `json:"request_id"`
	ConfirmUser  int    `json:"confirm_user_id"`
	ConfirmQuota int64  `json:"confirm_quota"`
}

func reviewedDebitInput(c *gin.Context) (debitReviewInput, []model.ReviewedDebit, int, bool) {
	var input debitReviewInput
	identity, ok := middleware.GetSessionAuthIdentity(c)
	if !ok || c.GetInt("role") < common.RoleAdminUser || identity.UserID != c.GetInt("id") {
		c.AbortWithStatus(http.StatusForbidden)
		return input, nil, 0, false
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 2048), &input); err != nil {
		billingError(c, model.ErrBillingEvidenceIntegrity)
		return input, nil, 0, false
	}
	cases, err := service.ReviewedDebits()
	if err != nil {
		billingError(c, err)
		return input, nil, 0, false
	}
	for _, row := range cases {
		if input.Month != row.Month || (input.Account != identity.UserID && input.Account != row.UserID) {
			billingError(c, model.ErrBillingConflict)
			return input, nil, 0, false
		}
	}
	return input, cases, identity.UserID, true
}

func PreviewHistoricalDebits(c *gin.Context) {
	_, cases, actor, ok := reviewedDebitInput(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()
	rows := make([]service.ReviewedDebitRow, 0, len(cases))
	for _, evidence := range cases {
		row, err := service.RunReviewedDebit(ctx, evidence, actor, false)
		if err != nil {
			billingError(c, err)
			return
		}
		rows = append(rows, row)
	}
	common.ApiSuccess(c, rows)
}

func CollectHistoricalDebit(c *gin.Context) {
	input, cases, actor, ok := reviewedDebitInput(c)
	if !ok {
		return
	}
	for _, evidence := range cases {
		if evidence.RequestID != input.Request || evidence.UserID != input.ConfirmUser || evidence.Quota != input.ConfirmQuota {
			continue
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
		defer cancel()
		row, err := service.RunReviewedDebit(ctx, evidence, actor, true)
		if err != nil {
			billingError(c, err)
			return
		}
		common.ApiSuccess(c, row)
		return
	}
	billingError(c, model.ErrBillingConflict)
}
