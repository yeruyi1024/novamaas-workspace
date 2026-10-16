package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service/authz"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func upstreamControllerDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	savedDB, savedLogDB := model.DB, model.LOG_DB
	model.DB, model.LOG_DB = db, db
	t.Cleanup(func() { model.DB, model.LOG_DB = savedDB, savedLogDB })
	require.NoError(t, db.AutoMigrate(&model.Channel{}, &model.User{}, &model.UpstreamBillingGroup{}, &model.UpstreamBillingGroupChannel{}, &model.UpstreamBillingGroupEvent{}, &model.UpstreamBillingStatement{}, &model.UpstreamBillingStatementEvent{}, &model.UpstreamBillingStatementRecord{}, &model.UpstreamBillingAttachment{}, &model.BillingArtifact{}, &model.BillingStatement{}))
	setupBillingPermissions(t, db)
	require.NoError(t, authz.SetUserPermissions(2, authz.PermissionsMap{authz.ResourceFinancialAccounting: {authz.ActionFinancialView: false}}))
	require.NoError(t, authz.SetUserPermissions(4, authz.PermissionsMap{authz.ResourceFinancialAccounting: {authz.ActionFinancialView: true}}))
	require.NoError(t, authz.SetUserPermissions(1, authz.PermissionsMap{authz.ResourceFinancialManagement: {authz.ActionFinancialManagementView: true}}))
	require.NoError(t, authz.SetUserPermissions(7, authz.PermissionsMap{authz.ResourceFinancialAccounting: {authz.ActionFinancialView: false}, authz.ResourceFinancialManagement: {authz.ActionFinancialManagementView: true}}))
	require.NoError(t, authz.SetUserPermissions(8, authz.PermissionsMap{authz.ResourceFinancialManagement: {authz.ActionFinancialManagementView: true}}))
	return db
}

func TestUpstreamBillingAllEndpointsRequireAdministratorAndFinancialPermission(t *testing.T) {
	upstreamControllerDatabase(t)
	endpoints := []struct {
		name    string
		handler gin.HandlerFunc
	}{
		{"groups", UpstreamBillingGroups}, {"save group", SaveUpstreamBillingGroup},
		{"channels", UpstreamBillingChannels}, {"report", UpstreamBillingReport},
		{"usage details", UpstreamBillingUsageDetails}, {"storage", UpstreamBillingStorageProfiles},
		{"list statements", ListUpstreamBillingStatements}, {"create statement", CreateUpstreamBillingStatement},
		{"statement detail", GetUpstreamBillingStatement}, {"statement actions", ActOnUpstreamBillingStatement},
		{"download", DownloadUpstreamBillingArtifact},
		{"upload attachment", UploadUpstreamBillingAttachment}, {"download attachment", DownloadUpstreamBillingAttachment},
	}
	for _, actor := range []struct {
		name     string
		id, role int
	}{
		{"customer", 5, common.RoleCommonUser},
		{"customer with financial grant", 4, common.RoleCommonUser},
		{"customer with financial management", 8, common.RoleCommonUser},
		{"administrator with financial access revoked", 2, common.RoleAdminUser},
		{"anonymous", 0, 0},
	} {
		t.Run(actor.name, func(t *testing.T) {
			for _, endpoint := range endpoints {
				t.Run(endpoint.name, func(t *testing.T) {
					recorder := httptest.NewRecorder()
					ctx, _ := gin.CreateTestContext(recorder)
					ctx.Set("id", actor.id)
					ctx.Set("role", actor.role)
					ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
					endpoint.handler(ctx)
					assert.Equal(t, http.StatusForbidden, recorder.Code)
					assert.NotContains(t, recorder.Body.String(), "cost_quota")
				})
			}
		})
	}
}

func TestUpstreamChannelOptionsNeverExposeCredentials(t *testing.T) {
	db := upstreamControllerDatabase(t)
	require.NoError(t, db.Create(&[]model.Channel{
		{Id: 19, Name: "Production text", Key: "upstream-private-key", CostDiscount: "0.5", OtherInfo: "private-provider-account"},
		{Id: 26, Name: "Production video", Key: "another-private-key", CostDiscount: "0"},
		{Id: 27, Name: "Test only", Key: "test-private-key"},
	}).Error)
	for _, role := range []int{common.RoleAdminUser, common.RoleRootUser} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Set("id", 1)
		ctx.Set("role", role)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/channels", nil)
		UpstreamBillingChannels(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)
		var response struct {
			Success bool                               `json:"success"`
			Data    []model.UpstreamBillingChannelView `json:"data"`
		}
		require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
		require.True(t, response.Success)
		require.Len(t, response.Data, 3)
		eligible := map[int]bool{}
		for _, channel := range response.Data {
			eligible[channel.ID] = channel.Eligible
		}
		assert.True(t, eligible[19])
		assert.True(t, eligible[26], "explicit zero discount is a configured upstream")
		assert.False(t, eligible[27], "unconfigured testing channel is excluded")
		for _, secret := range []string{"upstream-private-key", "another-private-key", "test-private-key", "private-provider-account"} {
			assert.NotContains(t, recorder.Body.String(), secret)
		}
	}
}

func TestUpstreamConfirmationRequiresSessionAcknowledgementAndFrozenManifest(t *testing.T) {
	db := upstreamControllerDatabase(t)
	snapshot := `{"schema_version":1,"source_sha256":"source","total":{"count":0}}`
	hash := sha256.Sum256([]byte(snapshot))
	require.NoError(t, db.Create(&model.UpstreamBillingStatement{
		ID: "upstream-document", GroupID: 1, Month: "2020-01", Revision: 1, Status: model.StatementDraft,
		Snapshot: model.UpstreamBillingSnapshotData(snapshot), SnapshotSHA256: hex.EncodeToString(hash[:]), SourceSHA256: "source",
		ManifestSHA256: strings.Repeat("a", 64), PDFSHA256: strings.Repeat("b", 64), ExcelSHA256: strings.Repeat("c", 64),
	}).Error)
	for _, scenario := range []struct {
		name, session, body string
		status              int
	}{
		{"personal token cannot confirm", "", `{"action":"confirm","acknowledged":true,"manifest_sha256":"wrong"}`, http.StatusForbidden},
		{"missing explicit acknowledgement", "session", `{"action":"confirm","manifest_sha256":"wrong"}`, http.StatusBadRequest},
		{"stale manifest cannot confirm", "session", `{"action":"confirm","acknowledged":true,"manifest_sha256":"wrong"}`, http.StatusConflict},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Set("id", 1)
			ctx.Set("role", common.RoleAdminUser)
			ctx.Set("session_id", scenario.session)
			ctx.Set("auth_version", int64(1))
			ctx.Set("session_version", int64(1))
			ctx.Params = gin.Params{{Key: "statement_id", Value: "upstream-document"}}
			ctx.Request = httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(scenario.body))
			ActOnUpstreamBillingStatement(ctx)
			assert.Equal(t, scenario.status, recorder.Code, recorder.Body.String())
		})
	}
	var retained model.UpstreamBillingStatement
	require.NoError(t, db.First(&retained, "id = ?", "upstream-document").Error)
	assert.Equal(t, model.StatementDraft, retained.Status)
	assert.Zero(t, retained.ConfirmedAt)
}

func TestCustomerAndUpstreamStatementsUseIndependentLookupBoundaries(t *testing.T) {
	db := upstreamControllerDatabase(t)
	require.NoError(t, db.Create(&model.BillingStatement{ID: "customer-only", UserID: 1}).Error)
	require.NoError(t, db.Create(&model.UpstreamBillingStatement{ID: "upstream-only", GroupID: 1}).Error)
	for _, scenario := range []struct {
		id      string
		handler gin.HandlerFunc
	}{
		{"customer-only", GetUpstreamBillingStatement},
		{"upstream-only", GetBillingStatement},
	} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Set("id", 1)
		ctx.Set("role", common.RoleRootUser)
		ctx.Params = gin.Params{{Key: "statement_id", Value: scenario.id}}
		ctx.Request = httptest.NewRequest(http.MethodGet, "/statements/"+scenario.id, nil)
		scenario.handler(ctx)
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	}
}

func TestUpstreamBillingEitherFinanceGrantAllowsAdministratorReadAccess(t *testing.T) {
	upstreamControllerDatabase(t)
	for _, actorID := range []int{6, 7} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Set("id", actorID)
		ctx.Set("role", common.RoleAdminUser)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/groups", nil)
		UpstreamBillingGroups(ctx)
		assert.Equal(t, http.StatusOK, recorder.Code)
	}
}

func TestAccountingAdministratorCannotReviewOrUploadUpstreamEvidence(t *testing.T) {
	upstreamControllerDatabase(t)
	for _, action := range []string{"confirm", "void"} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Set("id", 6)
		ctx.Set("role", common.RoleAdminUser)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(`{"action":"`+action+`","acknowledged":true,"reason":"review"}`))
		ActOnUpstreamBillingStatement(ctx)
		assert.Equal(t, http.StatusForbidden, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "UPSTREAM_BILLING_REVIEW_FORBIDDEN")
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 6)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/attachments", nil)
	ctx.Request.Body = upstreamUnreadBody{t: t}
	UploadUpstreamBillingAttachment(ctx)
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

type upstreamUnreadBody struct{ t *testing.T }

func (body upstreamUnreadBody) Read([]byte) (int, error) {
	body.t.Fatal("request body must not be read before authorization and statement-state checks")
	return 0, io.EOF
}

func (body upstreamUnreadBody) Close() error { return nil }

func TestTeamPermissionEditorReturnsIndependentFinanceSwitches(t *testing.T) {
	db := upstreamControllerDatabase(t)
	require.NoError(t, db.Create(&model.User{Id: 7, Username: "finance-manager", Role: common.RoleAdminUser}).Error)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 1)
	ctx.Set("role", common.RoleRootUser)
	ctx.Params = gin.Params{{Key: "id", Value: "7"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/user/7", nil)
	GetUser(ctx)
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			AdminPermissions authz.PermissionsMap `json:"admin_permissions"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	assert.False(t, response.Data.AdminPermissions[authz.ResourceFinancialAccounting][authz.ActionFinancialView])
	assert.True(t, response.Data.AdminPermissions[authz.ResourceFinancialManagement][authz.ActionFinancialManagementView])
	assert.True(t, authz.Can(7, common.RoleAdminUser, authz.FinancialAccountingView), "an independent off switch does not hide inherited views")
}

func TestUpstreamAttachmentDownloadCannotCrossStatementsOrReadUnpublishedUploads(t *testing.T) {
	db := upstreamControllerDatabase(t)
	snapshot := `{"schema_version":1}`
	hash := sha256.Sum256([]byte(snapshot))
	require.NoError(t, db.Create(&model.UpstreamBillingStatement{ID: "current", GroupID: 1, Status: model.StatementDraft, Snapshot: model.UpstreamBillingSnapshotData(snapshot), SnapshotSHA256: hex.EncodeToString(hash[:]), StorageProfileID: 1}).Error)
	require.NoError(t, db.Create(&[]model.UpstreamBillingAttachment{
		{ID: "foreign-file", StatementID: "other", Status: "ready", OriginalFilename: "foreign.pdf"},
		{ID: "pending-file", StatementID: "current", Status: "uploading", OriginalFilename: "pending.pdf"},
	}).Error)
	for _, id := range []string{"foreign-file", "pending-file", "missing-file"} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Set("id", 6)
		ctx.Set("role", common.RoleAdminUser)
		ctx.Params = gin.Params{{Key: "statement_id", Value: "current"}, {Key: "attachment_id", Value: id}}
		ctx.Request = httptest.NewRequest(http.MethodGet, "/attachments/"+id, nil)
		DownloadUpstreamBillingAttachment(ctx)
		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.NotContains(t, recorder.Body.String(), ".pdf")
	}
}

func TestUpstreamAttachmentUploadChecksFrozenStateBeforeReadingBody(t *testing.T) {
	db := upstreamControllerDatabase(t)
	snapshot := `{"schema_version":1}`
	hash := sha256.Sum256([]byte(snapshot))
	require.NoError(t, db.Create(&model.UpstreamBillingStatement{ID: "confirmed", GroupID: 1, Status: model.StatementConfirmed, Snapshot: model.UpstreamBillingSnapshotData(snapshot), SnapshotSHA256: hex.EncodeToString(hash[:])}).Error)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 7)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Params = gin.Params{{Key: "statement_id", Value: "confirmed"}}
	ctx.Request = httptest.NewRequest(http.MethodPost, "/attachments", nil)
	ctx.Request.Body = upstreamUnreadBody{t: t}
	UploadUpstreamBillingAttachment(ctx)
	assert.Equal(t, http.StatusConflict, recorder.Code)
}

func TestFinancialManagerConfirmsOnlyTheReviewedAttachmentDigest(t *testing.T) {
	db := upstreamControllerDatabase(t)
	snapshot := `{"schema_version":2,"source_sha256":"source","total":{"count":0}}`
	hash := sha256.Sum256([]byte(snapshot))
	statement := model.UpstreamBillingStatement{
		ID: "reviewed-draft", GroupID: 1, Month: "2020-01", Revision: 1, Status: model.StatementDraft,
		Snapshot: model.UpstreamBillingSnapshotData(snapshot), SnapshotSHA256: hex.EncodeToString(hash[:]), SourceSHA256: "source",
		ManifestSHA256: strings.Repeat("a", 64), PDFSHA256: strings.Repeat("b", 64), ExcelSHA256: strings.Repeat("c", 64),
	}
	require.NoError(t, db.Create(&statement).Error)
	for _, file := range []struct{ kind, digest string }{{"snapshot", statement.SnapshotSHA256}, {"pdf", statement.PDFSHA256}, {"xlsx", statement.ExcelSHA256}, {"manifest", statement.ManifestSHA256}} {
		require.NoError(t, db.Create(&model.BillingArtifact{StatementID: statement.ID, Kind: file.kind, SHA256: file.digest, Size: 1}).Error)
	}
	review, err := model.GetUpstreamBillingReview(statement.ID)
	require.NoError(t, err)
	for _, scenario := range []struct {
		name, digest string
		status       int
	}{
		{"unbound attachment set", "", http.StatusConflict},
		{"stale attachment set", strings.Repeat("d", 64), http.StatusConflict},
		{"reviewed empty optional attachment set", review.EvidenceSHA256, http.StatusOK},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			body, err := common.Marshal(map[string]any{"action": "confirm", "acknowledged": true, "manifest_sha256": statement.ManifestSHA256, "evidence_sha256": scenario.digest, "reason": "核对上游金额无误"})
			require.NoError(t, err)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Set("id", 7)
			ctx.Set("role", common.RoleAdminUser)
			ctx.Set("session_id", "financial-management-review")
			ctx.Set("auth_version", int64(1))
			ctx.Set("session_version", int64(1))
			ctx.Params = gin.Params{{Key: "statement_id", Value: statement.ID}}
			ctx.Request = httptest.NewRequest(http.MethodPost, "/actions", strings.NewReader(string(body)))
			ActOnUpstreamBillingStatement(ctx)
			require.Equal(t, scenario.status, recorder.Code, recorder.Body.String())
		})
	}
	confirmed, err := model.GetUpstreamBillingStatement(statement.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatementConfirmed, confirmed.Status)
	assert.Equal(t, 7, confirmed.ConfirmedBy)
	assert.Equal(t, review.EvidenceSHA256, confirmed.ReviewEvidenceSHA256)
	var event model.UpstreamBillingStatementEvent
	require.NoError(t, db.Where("statement_id = ? AND action = ?", statement.ID, "confirm").First(&event).Error)
	assert.Equal(t, "financial-management-review", event.SessionID)
	assert.Equal(t, "核对上游金额无误", event.Note)
}
