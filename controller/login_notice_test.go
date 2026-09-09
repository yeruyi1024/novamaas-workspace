package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/console_setting"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

func setupLoginNoticeControllerTest(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}, &model.LoginNoticeAcknowledgement{}))
	previousDB := model.DB
	previousConsoleSetting := *console_setting.GetConsoleSetting()
	model.DB = db
	console_setting.GetConsoleSetting().AnnouncementsEnabled = true
	console_setting.GetConsoleSetting().Announcements = `[{"id":1,"content":"Maintenance notice","type":"warning"}]`
	t.Cleanup(func() {
		model.DB = previousDB
		*console_setting.GetConsoleSetting() = previousConsoleSetting
	})
	return db
}

func loginNoticeContext(method, body string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, "/api/user/login-notice", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Request.Header.Set("User-Agent", "notice-test-browser")
	context.Set("id", 12)
	context.Set("session_id", "session-12")
	context.Set("auth_version", int64(1))
	context.Set("session_version", int64(1))
	return context, recorder
}

func TestLoginNoticeRequiresAcknowledgementForRecentViolation(t *testing.T) {
	db := setupLoginNoticeControllerTest(t)
	require.NoError(t, db.Create(&model.Task{
		TaskID:     "recent-violation",
		UserId:     12,
		Action:     constant.TaskActionGenerate,
		Status:     model.TaskStatusFailure,
		FailReason: "content policy violation",
		SubmitTime: time.Now().Unix(),
	}).Error)

	context, recorder := loginNoticeContext(http.MethodGet, "")
	GetLoginNotice(context)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.True(t, gjson.Get(recorder.Body.String(), "data.requires_acknowledgement").Bool())
	assert.False(t, gjson.Get(recorder.Body.String(), "data.acknowledged").Bool())
	assert.Equal(t, int64(1), gjson.Get(recorder.Body.String(), "data.statistics.seven_days.violations").Int())
	assert.Equal(t, "Maintenance notice", gjson.Get(recorder.Body.String(), "data.announcements.0.content").String())
}

func TestAcknowledgeLoginNoticeRecordsHashedDeviceAndSnapshot(t *testing.T) {
	db := setupLoginNoticeControllerTest(t)
	require.NoError(t, db.Create(&model.Task{
		TaskID:     "recent-violation",
		UserId:     12,
		Action:     constant.TaskActionGenerate,
		Status:     model.TaskStatusFailure,
		FailReason: "安全审核不通过",
		SubmitTime: time.Now().Unix(),
	}).Error)
	fingerprint := strings.Repeat("ab", 32)
	context, recorder := loginNoticeContext(http.MethodPost, `{"device_fingerprint":"`+fingerprint+`"}`)
	AcknowledgeLoginNotice(context)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.True(t, gjson.Get(recorder.Body.String(), "success").Bool())
	var acknowledgement model.LoginNoticeAcknowledgement
	require.NoError(t, db.First(&acknowledgement, "user_id = ? AND session_id = ?", 12, "session-12").Error)
	assert.Equal(t, fingerprint, acknowledgement.DeviceFingerprint)
	assert.Equal(t, "notice-test-browser", acknowledgement.UserAgent)
	assert.True(t, acknowledgement.RequiredAcknowledgement)
	assert.Equal(t, int64(1), acknowledgement.ViolationsSevenDays)
	assert.Len(t, acknowledgement.AnnouncementFingerprint, 64)

	getContext, getRecorder := loginNoticeContext(http.MethodGet, "")
	GetLoginNotice(getContext)
	assert.True(t, gjson.Get(getRecorder.Body.String(), "data.acknowledged").Bool())
}

func TestAcknowledgeLoginNoticeRejectsInvalidFingerprint(t *testing.T) {
	setupLoginNoticeControllerTest(t)
	context, recorder := loginNoticeContext(http.MethodPost, `{"device_fingerprint":"raw-device-data"}`)

	AcknowledgeLoginNotice(context)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.False(t, gjson.Get(recorder.Body.String(), "success").Bool())
}

func TestLoginNoticeRequiresDashboardSession(t *testing.T) {
	setupLoginNoticeControllerTest(t)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/user/login-notice", nil)
	context.Set("id", 12)

	GetLoginNotice(context)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Equal(t, "dashboard session required", gjson.Get(recorder.Body.String(), "message").String())
}
