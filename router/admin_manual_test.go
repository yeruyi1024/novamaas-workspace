package router

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	adminmanual "github.com/QuantumNous/new-api/docs/admin-manual"
	apidocs "github.com/QuantumNous/new-api/docs/platform-api"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type adminManualFixture struct {
	engine *gin.Engine
	tokens map[int]string
	pats   map[int]string
	users  map[int]model.User
}

func setupAdminManualFixture(t *testing.T) adminManualFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.UserSession{}, &model.Token{}, &model.Option{}, &model.Log{}))
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousRedis, previousSecret := common.RedisEnabled, common.SessionSecret
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	common.OptionMapRWMutex.Lock()
	previousOptions := common.OptionMap
	common.OptionMap = map[string]string{"HeaderNavModules": `{"docs":true}`, model.AdminManualEnabledOptionKey: "false"}
	common.OptionMapRWMutex.Unlock()
	model.DB, model.LOG_DB = db, db
	common.RedisEnabled, common.SessionSecret = false, "isolated-admin-manual-test"
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.RedisEnabled, common.SessionSecret = previousRedis, previousSecret
		common.SetDatabaseTypes(previousMainType, previousLogType)
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previousOptions
		common.OptionMapRWMutex.Unlock()
		require.NoError(t, sqlDB.Close())
	})
	fixture := adminManualFixture{engine: gin.New(), tokens: map[int]string{}, pats: map[int]string{}, users: map[int]model.User{}}
	for _, role := range []int{common.RoleCommonUser, common.RoleAdminUser, common.RoleRootUser} {
		username := fmt.Sprintf("manual-%d", role)
		user := model.User{Username: username, AffCode: username, Password: "unused-test-hash", Role: role, Status: common.UserStatusEnabled, Group: "default", AuthVersion: 1}
		pat := username + "-panel-pat"
		user.SetAccessToken(pat)
		require.NoError(t, db.Create(&user).Error)
		bundle, err := service.CreateLoginSession(user.Id, "password", "127.0.0.1", "manual-tests")
		require.NoError(t, err)
		fixture.tokens[role], fixture.pats[role], fixture.users[role] = bundle.AccessToken, pat, user
	}
	SetApiRouter(fixture.engine)
	SetWebRouter(fixture.engine, WebAssets{IndexPage: []byte("<html>application shell</html>")})
	return fixture
}

func (fixture adminManualFixture) request(method, path, credential, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if credential != "" {
		request.Header.Set("Authorization", "Bearer "+credential)
	}
	response := httptest.NewRecorder()
	fixture.engine.ServeHTTP(response, request)
	return response
}

func TestAdminManualPermissionMatrix(t *testing.T) {
	fixture := setupAdminManualFixture(t)
	for _, enabled := range []bool{false, true} {
		require.NoError(t, model.UpdateOptionsBulk(map[string]string{model.AdminManualEnabledOptionKey: fmt.Sprint(enabled)}))
		for _, role := range []int{0, common.RoleCommonUser, common.RoleAdminUser, common.RoleRootUser} {
			for _, path := range []string{"/api/admin-manual", "/api/admin-manual/", "/api/admin-manual/access", "/api/admin-manual/access/"} {
				t.Run(fmt.Sprintf("enabled=%t/role=%d%s", enabled, role, path), func(t *testing.T) {
					response := fixture.request(http.MethodGet, path, fixture.tokens[role], "")
					assert.Contains(t, response.Header().Get("Cache-Control"), "no-store")
					assert.Contains(t, response.Header().Get("Cache-Control"), "private")
					assert.Contains(t, response.Header().Get("Vary"), "Authorization")
					assert.Contains(t, response.Header().Get("X-Robots-Tag"), "noindex")
					if !enabled || role < common.RoleAdminUser {
						require.Equal(t, http.StatusNotFound, response.Code, response.Body.String())
						missing := fixture.request(http.MethodGet, "/api/nonexistent-manual", "", "")
						assert.JSONEq(t, strings.ReplaceAll(missing.Body.String(), "/api/nonexistent-manual", path), response.Body.String())
						assert.NotContains(t, response.Body.String(), "AUTH_")
						assert.NotContains(t, response.Body.String(), "available")
						return
					}
					require.Equal(t, http.StatusOK, response.Code, response.Body.String())
					if strings.HasSuffix(strings.TrimSuffix(path, "/"), "/access") {
						assert.JSONEq(t, `{"success":true,"data":{"available":true}}`, response.Body.String())
						assert.NotContains(t, response.Body.String(), "content")
						return
					}
					var body struct {
						Data []adminmanual.Document `json:"data"`
					}
					require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
					assert.Equal(t, adminmanual.Documents, body.Data)
					require.Len(t, body.Data, 9)
					ids := make([]string, 0, len(body.Data))
					for _, document := range body.Data {
						ids = append(ids, document.ID)
					}
					assert.Equal(t, []string{"overview", "platform", "common", "plans", "models", "changes", "extensions", "delivery", "roadmap"}, ids)
					assert.Equal(t, "手册概览", body.Data[0].Title)
				})
			}
		}
	}
}

func TestAdminManualRejectsInvalidIdentityAndStalePolicy(t *testing.T) {
	fixture := setupAdminManualFixture(t)
	require.NoError(t, model.UpdateOptionsBulk(map[string]string{model.AdminManualEnabledOptionKey: "true"}))
	for _, role := range []int{common.RoleAdminUser, common.RoleRootUser} {
		assert.Equal(t, http.StatusOK, fixture.request(http.MethodGet, "/api/admin-manual", fixture.pats[role], "").Code)
	}
	admin := fixture.users[common.RoleAdminUser]
	require.NoError(t, model.DB.Create(&model.Token{UserId: admin.Id, Key: "model-api-key-only", Status: common.TokenStatusEnabled, ExpiredTime: -1, UnlimitedQuota: true}).Error)
	for _, credential := range []string{"invalid", fixture.tokens[common.RoleAdminUser] + "tampered", "model-api-key-only", fixture.pats[common.RoleCommonUser]} {
		assert.Equal(t, http.StatusNotFound, fixture.request(http.MethodGet, "/api/admin-manual", credential, "").Code)
	}
	require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", admin.Id).Update("status", common.UserStatusDisabled).Error)
	assert.Equal(t, http.StatusNotFound, fixture.request(http.MethodGet, "/api/admin-manual", fixture.tokens[common.RoleAdminUser], "").Code)
	assert.Equal(t, http.StatusNotFound, fixture.request(http.MethodGet, "/api/admin-manual", fixture.pats[common.RoleAdminUser], "").Code)
	require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", admin.Id).Updates(map[string]any{"status": common.UserStatusEnabled, "role": 99}).Error)
	assert.Equal(t, http.StatusNotFound, fixture.request(http.MethodGet, "/api/admin-manual", fixture.pats[common.RoleAdminUser], "").Code)
	root := fixture.users[common.RoleRootUser]
	_, err := model.RevokeAllUserSessions(root.Id, "manual-test")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, fixture.request(http.MethodGet, "/api/admin-manual", fixture.tokens[common.RoleRootUser], "").Code)
	for _, persisted := range []string{"false", "1", "invalid"} {
		require.NoError(t, model.DB.Model(&model.Option{}).Where("key = ?", model.AdminManualEnabledOptionKey).Update("value", persisted).Error)
		assert.Equal(t, http.StatusNotFound, fixture.request(http.MethodGet, "/api/admin-manual", fixture.pats[common.RoleRootUser], "").Code, "persisted closure must override stale true OptionMap")
	}
	require.NoError(t, model.DB.Where("key = ?", model.AdminManualEnabledOptionKey).Delete(&model.Option{}).Error)
	assert.Equal(t, http.StatusNotFound, fixture.request(http.MethodGet, "/api/admin-manual", fixture.pats[common.RoleRootUser], "").Code)
	require.NoError(t, model.DB.Migrator().DropTable(&model.Option{}))
	assert.Equal(t, http.StatusNotFound, fixture.request(http.MethodGet, "/api/admin-manual", fixture.pats[common.RoleRootUser], "").Code)
}

func TestAdminManualSwitchRootOnlyPersistentAndAudited(t *testing.T) {
	fixture := setupAdminManualFixture(t)
	for _, role := range []int{0, common.RoleCommonUser, common.RoleAdminUser} {
		response := fixture.request(http.MethodPut, "/api/option/", fixture.tokens[role], `{"key":"AdminManualEnabled","value":true}`)
		if role == 0 {
			assert.Equal(t, http.StatusUnauthorized, response.Code)
		} else {
			assert.Equal(t, http.StatusForbidden, response.Code)
		}
	}
	root := fixture.users[common.RoleRootUser]
	for index, enabled := range []bool{true, false} {
		response := fixture.request(http.MethodPut, "/api/option/", fixture.tokens[common.RoleRootUser], fmt.Sprintf(`{"key":"AdminManualEnabled","value":%t}`, enabled))
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		assert.Contains(t, response.Body.String(), `"success":true`)
		stored, err := model.AdminManualEnabled(t.Context())
		require.NoError(t, err)
		assert.Equal(t, enabled, stored)
		read := fixture.request(http.MethodGet, "/api/option/", fixture.tokens[common.RoleRootUser], "")
		assert.Contains(t, read.Body.String(), `"key":"AdminManualEnabled","value":"`+fmt.Sprint(enabled)+`"`)
		var logs []model.Log
		require.NoError(t, model.LOG_DB.Where("type = ?", model.LogTypeManage).Order("id").Find(&logs).Error)
		require.Len(t, logs, index+1, "explicit and fallback audit must not duplicate successful writes")
		assert.Equal(t, root.Id, logs[index].UserId)
		var other map[string]any
		require.NoError(t, common.UnmarshalJsonStr(logs[index].Other, &other))
		assert.Equal(t, map[string]any{"action": "option.update", "params": map[string]any{"key": model.AdminManualEnabledOptionKey}}, other["op"])
		assert.Contains(t, other, "admin_info")
		assert.NotContains(t, logs[index].Other, fixture.tokens[common.RoleRootUser])
		assert.NotContains(t, logs[index].Other, `"value"`)
	}
}

func TestAdminManualSwitchFailuresPreservePolicyAndAuditFailure(t *testing.T) {
	fixture := setupAdminManualFixture(t)
	require.NoError(t, model.UpdateOptionsBulk(map[string]string{model.AdminManualEnabledOptionKey: "false"}))
	for index, value := range []string{`"yes"`, `1`, `null`, `{}`, `[]`} {
		response := fixture.request(http.MethodPut, "/api/option/", fixture.tokens[common.RoleRootUser], `{"key":"AdminManualEnabled","value":`+value+`}`)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Eventually(t, func() bool {
			var count int64
			return model.LOG_DB.Model(&model.Log{}).Where("type = ?", model.LogTypeManage).Count(&count).Error == nil && count == int64(index+1)
		}, time.Second, time.Millisecond, "failed authorized writes must finish their asynchronous fallback audit")
	}
	const callback = "test:admin-manual-save-error"
	require.NoError(t, model.DB.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "options" {
			tx.AddError(errors.New("isolated test persistence failure"))
		}
	}))
	response := fixture.request(http.MethodPut, "/api/option/", fixture.tokens[common.RoleRootUser], `{"key":"AdminManualEnabled","value":true}`)
	require.NoError(t, model.DB.Callback().Update().Remove(callback))
	assert.Contains(t, response.Body.String(), `"success":false`)
	stored, err := model.AdminManualEnabled(t.Context())
	require.NoError(t, err)
	assert.False(t, stored)
	common.OptionMapRWMutex.RLock()
	assert.Equal(t, "false", common.OptionMap[model.AdminManualEnabledOptionKey])
	common.OptionMapRWMutex.RUnlock()
	require.Eventually(t, func() bool {
		var logs []model.Log
		if model.LOG_DB.Where("type = ?", model.LogTypeManage).Find(&logs).Error != nil || len(logs) != 6 {
			return false
		}
		for _, log := range logs {
			var other struct {
				Audit struct{ Success bool } `json:"audit_info"`
			}
			if common.UnmarshalJsonStr(log.Other, &other) != nil || other.Audit.Success || strings.Contains(log.Other, `"option.update"`) {
				return false
			}
		}
		return true
	}, time.Second, time.Millisecond)
}

func TestAdminManualNeverAppearsInPublicDocumentsOrOpenAPI(t *testing.T) {
	fixture := setupAdminManualFixture(t)
	for _, enabled := range []bool{false, true} {
		require.NoError(t, model.UpdateOptionsBulk(map[string]string{model.AdminManualEnabledOptionKey: fmt.Sprint(enabled)}))
		status := fixture.request(http.MethodGet, "/api/status", "", "")
		require.Equal(t, http.StatusOK, status.Code, status.Body.String())
		assert.NotContains(t, status.Body.String(), model.AdminManualEnabledOptionKey)
		assert.NotContains(t, status.Body.String(), "/api/admin-manual")
		for _, credential := range []string{"", fixture.tokens[common.RoleAdminUser]} {
			response := fixture.request(http.MethodGet, "/api/docs", credential, "")
			require.Equal(t, http.StatusOK, response.Code)
			var body struct {
				Data []apidocs.Document `json:"data"`
			}
			require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
			assert.Equal(t, apidocs.Documents, body.Data)
			assert.Len(t, body.Data, 7)
			for _, manual := range adminmanual.Documents {
				encoded, err := common.Marshal(manual.Content)
				require.NoError(t, err)
				assert.NotContains(t, response.Body.String(), string(encoded))
				for _, doc := range body.Data {
					assert.NotEqual(t, manual.ID, doc.ID)
				}
			}
		}
	}
	for _, filename := range []string{"../docs/openapi/api.json", "../docs/openapi/relay.json"} {
		body, err := os.ReadFile(filename)
		require.NoError(t, err)
		var document map[string]any
		require.NoError(t, common.Unmarshal(body, &document))
		assert.NotContains(t, string(body), "/api/admin-manual")
		assert.NotContains(t, string(body), model.AdminManualEnabledOptionKey)
		for _, manual := range adminmanual.Documents {
			encoded, err := common.Marshal(manual.Content)
			require.NoError(t, err)
			assert.NotContains(t, string(body), string(encoded))
		}
	}
	for _, path := range []string{"/docs/admin-manual/starpath-integration.zh_CN.md", "/admin-manual/../docs/admin-manual/starpath-integration.zh_CN.md", "/docs?article=account-login", "/api/admin-manual/missing"} {
		response := fixture.request(http.MethodGet, path, fixture.tokens[common.RoleRootUser], "")
		assert.NotContains(t, response.Body.String(), adminmanual.Documents[0].Content)
		assert.NotContains(t, response.Body.String(), "/api/apps/starpath/managed-key")
		if strings.HasPrefix(path, "/api/") {
			assert.Equal(t, http.StatusNotFound, response.Code)
		} else {
			assert.Equal(t, "<html>application shell</html>", response.Body.String())
		}
	}
	// Use the exact existing missing-route handler as the response contract.
	missing := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(missing)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin-manual", nil)
	controller.RelayNotFound(context)
	denied := fixture.request(http.MethodGet, "/api/admin-manual", "", "")
	assert.JSONEq(t, missing.Body.String(), denied.Body.String())
}
