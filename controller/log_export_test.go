package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

const xlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

func runLogExport(t *testing.T, handler gin.HandlerFunc, userId int, role int, query string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", userId)
	ctx.Set("role", role)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/log/export?"+query, nil)
	handler(ctx)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	return recorder
}

// exportedSheet returns the rows of the exported workbook, header included.
func exportedSheet(t *testing.T, recorder *httptest.ResponseRecorder) [][]string {
	t.Helper()
	require.Equal(t, xlsxContentType, recorder.Header().Get("Content-Type"), recorder.Body.String())
	book, err := excelize.OpenReader(bytes.NewReader(recorder.Body.Bytes()))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("使用日志")
	require.NoError(t, err)
	return rows
}

// exportFailure returns the API error an export answered with instead of a file.
func exportFailure(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	require.NotEqual(t, xlsxContentType, recorder.Header().Get("Content-Type"), "a failed export must not send a file")
	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &result))
	require.False(t, result.Success)
	return result.Message
}

func TestExportUserLogsListsOnlyOwnLogsNewestFirst(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Log{}, &model.BillingEntry{}))
	now := time.Now().Unix()
	require.NoError(t, db.Create(&[]model.Log{
		{UserId: 7, Type: model.LogTypeConsume, RequestId: "old", TokenName: "prod-key", ModelName: "model-a", Quota: 500000, PromptTokens: 12, CreatedAt: now - 20},
		{UserId: 8, Type: model.LogTypeConsume, RequestId: "other-user", CreatedAt: now - 15},
		{UserId: 7, Type: model.LogTypeConsume, RequestId: "new", CreatedAt: now - 10},
	}).Error)

	rows := exportedSheet(t, runLogExport(t, ExportUserLogs, 7, common.RoleCommonUser, ""))

	require.Len(t, rows, 3)
	assert.NotContains(t, rows[0], "用户", "only administrators see the user column")
	requestIdColumn := len(rows[0]) - 2
	assert.Equal(t, "new", rows[1][requestIdColumn])
	assert.Equal(t, "old", rows[2][requestIdColumn])
	assert.Equal(t, "prod-key", rows[2][2])
	assert.Equal(t, "消耗", rows[2][1])
}

func TestExportAllLogsAddsUserColumnForAdministrators(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Log{}, &model.BillingEntry{}))
	require.NoError(t, db.Create(&model.Log{UserId: 7, Username: "alice", Type: model.LogTypeConsume, RequestId: "a", CreatedAt: time.Now().Unix() - 5}).Error)

	rows := exportedSheet(t, runLogExport(t, ExportAllLogs, 1, common.RoleRootUser, ""))

	require.Len(t, rows, 2)
	assert.Equal(t, "用户", rows[0][2])
	assert.Equal(t, "alice", rows[1][2])
}

func TestExportLogsRejectsRangeOverLimit(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Log{}, &model.BillingEntry{}))
	end := time.Now().Unix()
	start := end - (maxLogExportDays+1)*24*3600

	recorder := runLogExport(t, ExportUserLogs, 7, common.RoleCommonUser,
		"start_timestamp="+strconv.FormatInt(start, 10)+"&end_timestamp="+strconv.FormatInt(end, 10))

	assert.Contains(t, exportFailure(t, recorder), "366")
}

func TestExportLogsStopsAtRowLimitWithoutSendingAFile(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Log{}, &model.BillingEntry{}))
	now := time.Now().Unix()
	require.NoError(t, db.Create(&[]model.Log{
		{UserId: 7, Type: model.LogTypeConsume, RequestId: "a", CreatedAt: now - 3},
		{UserId: 7, Type: model.LogTypeConsume, RequestId: "b", CreatedAt: now - 2},
		{UserId: 7, Type: model.LogTypeConsume, RequestId: "c", CreatedAt: now - 1},
	}).Error)
	previous := maxLogExportRows
	maxLogExportRows = 2
	t.Cleanup(func() { maxLogExportRows = previous })

	recorder := runLogExport(t, ExportUserLogs, 7, common.RoleCommonUser, "")

	assert.Contains(t, exportFailure(t, recorder), "上限")
}

func TestExportLogsNeverCarriesPlatformCostFigures(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Log{}, &model.CostAccountingSnapshot{}, &model.CostAccountingAdjustment{}, &model.BillingEntry{}))
	log := model.Log{
		UserId: 7, Type: model.LogTypeConsume, Quota: 1000, ChannelId: 5, ChannelName: "supplier-a",
		CreatedAt: time.Now().Unix(), RequestId: "export-financial", ModelName: "model-a",
	}
	require.NoError(t, db.Create(&log).Error)
	require.NoError(t, model.RecordCostAccountingSnapshot(&log, &model.CostAccountingInput{
		CostBasisQuota: "1000", CostDiscount: "0.5", CostQuota: 500,
	}))

	// A root user sees cost, revenue and profit on screen, yet the file stays shareable.
	rows := exportedSheet(t, runLogExport(t, ExportAllLogs, 1, common.RoleRootUser, ""))

	require.Len(t, rows, 2)
	for _, cell := range append(append([]string{}, rows[0]...), rows[1]...) {
		assert.NotContains(t, cell, "成本")
		assert.NotContains(t, cell, "利润")
		assert.NotContains(t, cell, "收入")
		assert.NotContains(t, cell, "渠道")
		assert.NotContains(t, cell, "supplier-a")
	}
}

func TestExportLogsListsOnlyBillingRecords(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Log{}, &model.BillingEntry{}))
	now := time.Now().Unix()
	require.NoError(t, db.Create(&[]model.Log{
		{UserId: 7, Type: model.LogTypeTopup, RequestId: "topup", Content: "订阅购买", CreatedAt: now - 6},
		{UserId: 7, Type: model.LogTypeConsume, RequestId: "consume", CreatedAt: now - 5},
		{UserId: 7, Type: model.LogTypeRefund, RequestId: "refund", CreatedAt: now - 4},
		{UserId: 7, Type: model.LogTypeManage, RequestId: "manage", Content: "POST /api/channel/:id/key", CreatedAt: now - 3},
		{UserId: 7, Type: model.LogTypeError, RequestId: "error", Content: "upstream supplier failed", CreatedAt: now - 2},
		{UserId: 7, Type: model.LogTypeLogin, RequestId: "login", CreatedAt: now - 1},
	}).Error)

	rows := exportedSheet(t, runLogExport(t, ExportUserLogs, 7, common.RoleCommonUser, ""))
	require.Len(t, rows, 3)
	requestIdColumn := len(rows[0]) - 2
	assert.Equal(t, []string{"refund", "consume"}, []string{rows[1][requestIdColumn], rows[2][requestIdColumn]})

	for _, logType := range []int{model.LogTypeTopup, model.LogTypeManage} {
		rejected := runLogExport(t, ExportUserLogs, 7, common.RoleCommonUser, "type="+strconv.Itoa(logType))
		assert.Contains(t, exportFailure(t, rejected), "不支持导出")
	}
}
