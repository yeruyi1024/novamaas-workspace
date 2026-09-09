package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBillingBrandingFailureReturnsActionableErrorCode(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	billingError(ctx, service.ErrBillingDocumentBranding)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	var response struct {
		Code string `json:"code"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, "BILLING_BRANDING_INVALID", response.Code)
}

func TestBillingControllerOwnershipAndSessionBoundary(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	saved := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = saved })
	require.NoError(t, db.AutoMigrate(&model.BillingStatement{}, &model.BillingStatementEvent{}, &model.BillingArtifact{}))
	statement := &model.BillingStatement{ID: "private", UserID: 2, Month: "2026-02", Revision: 1, Status: model.StatementIssued, IssuedAt: 100, ManifestSHA256: "frozen", PDFSHA256: "pdf"}
	statement.Snapshot = `{"total":{"count":1}}`
	hash := sha256.Sum256([]byte(statement.Snapshot))
	statement.SnapshotSHA256 = hex.EncodeToString(hash[:])
	require.NoError(t, db.Create(statement).Error)
	for _, test := range []struct {
		name, method, path, body, session string
		actor, role, status               int
	}{
		{"cross-account query", "GET", "/day?user_id=2", "", "", 3, common.RoleCommonUser, 403},
		{"cross-account usage details", "GET", "/usage-details?user_id=2&date=2020-02-03&hour=9", "", "", 3, common.RoleCommonUser, 403},
		{"cross-account document", "GET", "/statements/private", "", "", 3, common.RoleCommonUser, 404},
		{"admin on behalf", "POST", "/statements/private/actions", `{"action":"confirm","manifest_sha256":"frozen"}`, "admin", 1, common.RoleAdminUser, 403},
		{"personal access token", "POST", "/statements/private/actions", `{"action":"confirm","manifest_sha256":"frozen"}`, "", 2, common.RoleCommonUser, 403},
		{"owner live session", "POST", "/statements/private/actions", `{"action":"confirm","manifest_sha256":"frozen"}`, "owner", 2, common.RoleCommonUser, 200},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("id", test.actor)
				c.Set("role", test.role)
				c.Set("session_id", test.session)
				c.Set("auth_version", int64(1))
				c.Set("session_version", int64(1))
				c.Next()
			})
			router.GET("/day", BillingDay)
			router.GET("/usage-details", BillingUsageDetails)
			router.GET("/statements/:statement_id", GetBillingStatement)
			router.POST("/statements/:statement_id/actions", ActOnBillingStatement)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(recorder, request)
			assert.Equal(t, test.status, recorder.Code, recorder.Body.String())
		})
	}
	var confirmed model.BillingStatement
	require.NoError(t, db.First(&confirmed, "id = ?", statement.ID).Error)
	assert.Equal(t, model.StatementConfirmed, confirmed.Status)
	var events []model.BillingStatementEvent
	require.NoError(t, db.Find(&events).Error)
	require.Len(t, events, 1)
	assert.Equal(t, 2, events[0].ActorID)
}

func TestBillingControllerUserCannotConfigureAccountingStart(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("id", 2); c.Set("role", common.RoleCommonUser) })
	router.PUT("/account", UpdateBillingAccount)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/account", strings.NewReader(`{"company_title":"Customer","tax_id":"TAX","accounting_start_at":0}`)))
	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestBillingMonthPreviewDistinguishesUnconfiguredFromZeroConsumption(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	savedDB, savedLogDB := model.DB, model.LOG_DB
	model.DB, model.LOG_DB = db, db
	t.Cleanup(func() { model.DB, model.LOG_DB = savedDB, savedLogDB })
	require.NoError(t, db.AutoMigrate(&model.BillingAccount{}, &model.BillingHour{}, &model.BillingOperation{}, &model.BillingStatement{}, &model.Log{}))
	start := time.Date(2020, 2, 3, 12, 0, 0, 0, time.FixedZone("Asia/Shanghai", 8*3600)).Unix()
	require.NoError(t, db.Create(&model.Log{UserId: 4, Type: model.LogTypeConsume, Quota: 500000, CreatedAt: start}).Error)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("id", 4); c.Set("role", common.RoleAdminUser) })
	router.GET("/preview", BillingMonthPreview)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/preview?month=2020-02", nil))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var body struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
	require.Contains(t, body.Data, "readiness")
	assert.Nil(t, body.Data["formal"], "no official zero-amount statement before accounting is configured")
	reference, ok := body.Data["reference"].(map[string]any)
	require.True(t, ok)
	total, ok := reference["total"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(1), total["count"])
}

func TestBillingStatementDetailResolvesCustomerAndOperatorWithoutPrivateUserFields(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	saved := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = saved })
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.BillingStatement{}, &model.BillingStatementEvent{}, &model.BillingArtifact{}))
	require.NoError(t, db.Create(&[]model.User{{Id: 1, Username: "admin", DisplayName: "Administrator", Password: "private-hash", Email: "private@example.com", AffCode: "test-admin"}, {Id: 4, Username: "example_customer", DisplayName: "Customer", AffCode: "test-customer"}}).Error)
	snapshot := `{"user_id":4,"total":{"count":13}}`
	hash := sha256.Sum256([]byte(snapshot))
	require.NoError(t, db.Create(&model.BillingStatement{ID: "identified", UserID: 4, Status: model.StatementIssued, IssuedAt: 100, Snapshot: snapshot, SnapshotSHA256: hex.EncodeToString(hash[:])}).Error)
	require.NoError(t, db.Create(&model.BillingStatementEvent{StatementID: "identified", ActorID: 1, Action: "prepare", SessionID: "private-session"}).Error)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("id", 4); c.Set("role", common.RoleCommonUser) })
	router.GET("/statements/:statement_id", GetBillingStatement)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/statements/identified", nil))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var result struct {
		Data struct {
			Customer model.BillingUserIdentity `json:"customer"`
			Events   []struct {
				ActorID       int    `json:"actor_id"`
				ActorUsername string `json:"actor_username"`
			} `json:"events"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
	assert.Equal(t, model.BillingUserIdentity{ID: 4, Username: "example_customer", DisplayName: "Customer"}, result.Data.Customer)
	require.Len(t, result.Data.Events, 1)
	assert.Equal(t, 1, result.Data.Events[0].ActorID)
	assert.Equal(t, "admin", result.Data.Events[0].ActorUsername)
	assert.NotContains(t, response.Body.String(), "private")
	assert.NotContains(t, response.Body.String(), "password")
	assert.NotContains(t, response.Body.String(), "email")
}

func TestBillingLegacyEmptyDraftCannotBeIssuedWithHistoricalConsumption(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	savedDB, savedLogs := model.DB, model.LOG_DB
	model.DB, model.LOG_DB = db, db
	t.Cleanup(func() { model.DB, model.LOG_DB = savedDB, savedLogs })
	require.NoError(t, db.AutoMigrate(&model.BillingStatement{}, &model.BillingStatementEvent{}, &model.Log{}))
	start, end, err := model.BillingMonthBounds("2020-05")
	require.NoError(t, err)
	snapshot := `{"total":{"count":0}}`
	hash := sha256.Sum256([]byte(snapshot))
	require.NoError(t, db.Create(&model.BillingStatement{ID: "legacy-empty", UserID: 4, Month: "2020-05", Status: model.StatementDraft, StartAt: start, EndAt: end, Snapshot: snapshot, SnapshotSHA256: hex.EncodeToString(hash[:]), ManifestSHA256: "manifest", PDFSHA256: "pdf"}).Error)
	require.NoError(t, db.Create(&model.Log{UserId: 4, Type: model.LogTypeConsume, Quota: 500000, CreatedAt: start + 3600}).Error)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("id", 1); c.Set("role", common.RoleAdminUser) })
	router.POST("/statements/:statement_id/actions", ActOnBillingStatement)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/statements/legacy-empty/actions", strings.NewReader(`{"action":"issue"}`)))
	assert.Equal(t, http.StatusConflict, response.Code)
	assert.Contains(t, response.Body.String(), "BILLING_HISTORY_UNRECONCILED")
	var statement model.BillingStatement
	require.NoError(t, db.First(&statement, "id = ?", "legacy-empty").Error)
	assert.Equal(t, model.StatementDraft, statement.Status)
	assert.Equal(t, snapshot, statement.Snapshot)
}
