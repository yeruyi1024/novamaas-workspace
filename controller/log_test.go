package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLogsSelfStatSeparatesRequestsFromBillingRecords(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Log{}))
	now := time.Now().Unix()
	require.NoError(t, db.Create(&[]model.Log{
		{UserId: 7, Type: model.LogTypeConsume, Quota: 500000, CreatedAt: now},
		{UserId: 7, Type: model.LogTypeRefund, Quota: 100000, CreatedAt: now},
		{UserId: 8, Type: model.LogTypeConsume, Quota: 900000, CreatedAt: now},
	}).Error)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 7)
	ctx.Set("role", common.RoleCommonUser)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/log/self/stat", nil)

	GetLogsSelfStat(ctx)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Quota    int64 `json:"quota"`
			Records  int64 `json:"records"`
			Requests int64 `json:"requests"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	assert.Equal(t, int64(500000), response.Data.Quota)
	assert.Equal(t, int64(2), response.Data.Records)
	assert.Equal(t, int64(1), response.Data.Requests)
}

func TestCommonLogRowsExposeAccountingOnlyToFinancialViewer(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Log{}, &model.CostAccountingSnapshot{}, &model.CostAccountingAdjustment{}))
	log := model.Log{
		UserId: 7, Type: model.LogTypeConsume, Quota: 1000,
		CreatedAt: time.Now().Unix(), RequestId: "financial-log", ModelName: "model-a",
	}
	require.NoError(t, db.Create(&log).Error)
	require.NoError(t, model.RecordCostAccountingSnapshot(&log, &model.CostAccountingInput{
		CostBasisQuota: "1000", CostDiscount: "0.5", CostQuota: 500,
	}))

	for _, test := range []struct {
		name        string
		role        int
		wantFinance bool
	}{
		{name: "root sees accounting", role: common.RoleRootUser, wantFinance: true},
		{name: "ordinary user does not see accounting", role: common.RoleCommonUser},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Set("id", 7)
			ctx.Set("role", test.role)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/log/self?p=1&page_size=10", nil)

			GetUserLogs(ctx)

			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			var response struct {
				Success bool `json:"success"`
				Data    struct {
					Items []map[string]any `json:"items"`
				} `json:"data"`
			}
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
			require.True(t, response.Success)
			require.Len(t, response.Data.Items, 1)
			item := response.Data.Items[0]
			if test.wantFinance {
				assert.Equal(t, float64(1000), item["revenue_quota"])
				assert.Equal(t, float64(500), item["cost_quota"])
				assert.Equal(t, float64(500), item["profit_quota"])
			} else {
				assert.NotContains(t, item, "revenue_quota")
				assert.NotContains(t, item, "cost_quota")
				assert.NotContains(t, item, "profit_quota")
			}
		})
	}
}
