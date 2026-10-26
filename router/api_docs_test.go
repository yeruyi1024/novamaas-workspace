package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	apidocs "github.com/QuantumNous/new-api/docs/platform-api"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type documentationFixture struct {
	engine *gin.Engine
	tokens map[int]string
	pats   map[int]string
	users  map[int]model.User
}

func setupDocumentationFixture(t *testing.T) documentationFixture {
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
	common.OptionMap = map[string]string{"HeaderNavModules": `{"docs":true}`}
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
	fixture := documentationFixture{engine: gin.New(), tokens: map[int]string{}, pats: map[int]string{}, users: map[int]model.User{}}
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

func (fixture documentationFixture) request(path, credential string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("Content-Type", "application/json")
	if credential != "" {
		request.Header.Set("Authorization", "Bearer "+credential)
	}
	response := httptest.NewRecorder()
	fixture.engine.ServeHTTP(response, request)
	return response
}

func TestDocumentationAdminAccess(t *testing.T) {
	fixture := setupDocumentationFixture(t)
	for _, tt := range []struct {
		name       string
		credential string
		status     int
	}{
		{"visitor", "", http.StatusUnauthorized},
		{"invalid token", "invalid", http.StatusUnauthorized},
		{"ordinary user", fixture.tokens[common.RoleCommonUser], http.StatusForbidden},
		{"administrator", fixture.tokens[common.RoleAdminUser], http.StatusOK},
		{"root", fixture.tokens[common.RoleRootUser], http.StatusOK},
		{"administrator PAT", fixture.pats[common.RoleAdminUser], http.StatusOK},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response := fixture.request("/api/docs/admin", tt.credential)
			require.Equal(t, tt.status, response.Code)
			assert.Contains(t, response.Header().Get("Cache-Control"), "no-store")
			var body struct {
				Success bool               `json:"success"`
				Data    []apidocs.Document `json:"data"`
			}
			require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
			if tt.status != http.StatusOK {
				assert.Empty(t, body.Data)
				assert.NotContains(t, response.Body.String(), "/api/apps/starpath/managed-key")
				return
			}
			assert.True(t, body.Success)
			assert.Equal(t, apidocs.AdminDocuments, body.Data)
			require.Len(t, body.Data, 9)
			assert.Equal(t, "overview", body.Data[0].ID)
			assert.Contains(t, body.Data[2].Content, "/api/user/auth/refresh")
		})
	}
	admin := fixture.users[common.RoleAdminUser]
	require.NoError(t, model.DB.Model(&model.User{}).Where("id = ?", admin.Id).Update("role", common.RoleCommonUser).Error)
	assert.Equal(t, http.StatusForbidden, fixture.request("/api/docs/admin", fixture.tokens[common.RoleAdminUser]).Code)
}

func TestDocumentationPublicResponseExcludesAdminContent(t *testing.T) {
	fixture := setupDocumentationFixture(t)
	for _, credential := range []string{"", fixture.tokens[common.RoleCommonUser], fixture.tokens[common.RoleRootUser]} {
		response := fixture.request("/api/docs?section=admin", credential)
		require.Equal(t, http.StatusOK, response.Code)
		var body struct {
			Data []apidocs.Document `json:"data"`
		}
		require.NoError(t, common.Unmarshal(response.Body.Bytes(), &body))
		assert.Equal(t, apidocs.Documents, body.Data)
		assert.Len(t, body.Data, 7)
		for _, manual := range apidocs.AdminDocuments {
			encoded, err := common.Marshal(manual.Content)
			require.NoError(t, err)
			assert.NotContains(t, response.Body.String(), string(encoded))
		}
	}
	common.OptionMapRWMutex.Lock()
	common.OptionMap["HeaderNavModules"] = `{"docs":false}`
	common.OptionMapRWMutex.Unlock()
	assert.Equal(t, http.StatusForbidden, fixture.request("/api/docs/admin", fixture.tokens[common.RoleRootUser]).Code)
}
