package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func performUserMutationRequest(t *testing.T, method string, body string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, "/api/user/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("id", 9999)
	c.Set("role", common.RoleAdminUser)
	c.Set("username", "admin-operator")
	handler(c)
	return recorder
}

func TestCreateUserPersistsNormalizedPhone(t *testing.T) {
	db := setupManageUserTestDB(t)
	recorder := performUserMutationRequest(
		t,
		http.MethodPost,
		`{"username":"phone-create-user","password":"NewPassword123","display_name":"Phone Create","phone":" 13800138000 ","role":1}`,
		CreateUser,
	)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"success":true`)

	var stored model.User
	require.NoError(t, db.Where("username = ?", "phone-create-user").First(&stored).Error)
	assert.Equal(t, "13800138000", stored.Phone)
}

func TestUpdateUserPersistsNormalizedPhone(t *testing.T) {
	db := setupManageUserTestDB(t)
	stored := model.User{
		Username:    "phone-update-user",
		Password:    "stored-password",
		DisplayName: "Before",
		Phone:       "13900139000",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AuthVersion: 1,
	}
	require.NoError(t, db.Create(&stored).Error)

	recorder := performUserMutationRequest(
		t,
		http.MethodPut,
		fmt.Sprintf(
			`{"id":%d,"username":"phone-update-user","display_name":"After","phone":" 13800138000 ","group":"default"}`,
			stored.Id,
		),
		UpdateUser,
	)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"success":true`)

	var updated model.User
	require.NoError(t, db.First(&updated, stored.Id).Error)
	assert.Equal(t, "13800138000", updated.Phone)
	assert.Equal(t, "After", updated.DisplayName)
}

func TestUpdateUserPreservesPhoneWhenLegacyClientOmitsField(t *testing.T) {
	db := setupManageUserTestDB(t)
	stored := model.User{
		Username:    "legacy-phone-user",
		Password:    "stored-password",
		DisplayName: "Before",
		Phone:       "13800138000",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AuthVersion: 1,
	}
	require.NoError(t, db.Create(&stored).Error)

	recorder := performUserMutationRequest(
		t,
		http.MethodPut,
		fmt.Sprintf(
			`{"id":%d,"username":"legacy-phone-user","display_name":"After","group":"default"}`,
			stored.Id,
		),
		UpdateUser,
	)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"success":true`)

	var updated model.User
	require.NoError(t, db.First(&updated, stored.Id).Error)
	assert.Equal(t, "13800138000", updated.Phone)
}
