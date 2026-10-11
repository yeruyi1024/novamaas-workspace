package imagestudio

import (
	"path/filepath"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupStudioServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	previousDB := model.DB
	previousRedisEnabled := common.RedisEnabled
	previousSQLitePath := common.SQLitePath
	previousMasterNode := common.IsMasterNode
	previousDBType := common.MainDatabaseType()
	previousLogDBType := common.LogDatabaseType()
	t.Setenv("SQL_DSN", "")
	common.SQLitePath = filepath.Join(t.TempDir(), "image-studio.db")
	common.IsMasterNode = false
	common.RedisEnabled = false
	require.NoError(t, model.InitDB())
	db := model.DB
	require.NoError(t, db.AutoMigrate(&model.Token{}))
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		model.DB = previousDB
		common.RedisEnabled = previousRedisEnabled
		common.SQLitePath = previousSQLitePath
		common.IsMasterNode = previousMasterNode
		common.SetMainDatabaseType(previousDBType)
		common.SetLogDatabaseType(previousLogDBType)
	})
	return db
}

func TestTokenSelectionRejectsForeignAndUnavailableKeys(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	for _, tc := range []struct {
		name          string
		owner, status int
		expires       int64
		unlimited     bool
		quota         int
	}{
		{"foreign", 8, common.TokenStatusEnabled, -1, true, 0},
		{"disabled", 7, common.TokenStatusDisabled, -1, true, 0},
		{"expired", 7, common.TokenStatusEnabled, 1, true, 0},
		{"exhausted", 7, common.TokenStatusEnabled, -1, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token := &model.Token{UserId: tc.owner, Key: tc.name, Status: tc.status, ExpiredTime: tc.expires, UnlimitedQuota: tc.unlimited, RemainQuota: tc.quota}
			require.NoError(t, db.Create(token).Error)
			_, err := resolveToken(7, token.Id)
			require.Error(t, err)
			assert.Equal(t, "token_unavailable", ErrorCode(err))
		})
	}
}

func TestQuotaForRequestReadsDedicatedLogDatabase(t *testing.T) {
	mainDB := setupStudioServiceTestDB(t)
	require.NoError(t, mainDB.AutoMigrate(&model.Log{}))
	logDB, err := gorm.Open(sqlite.Open("file:studio-logs?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, logDB.AutoMigrate(&model.Log{}))
	previousLogDB := model.LOG_DB
	model.LOG_DB = logDB
	t.Cleanup(func() { model.LOG_DB = previousLogDB })

	require.NoError(t, mainDB.Create(&model.Log{
		UserId: 11, Type: model.LogTypeConsume, RequestId: "request-1", Quota: 7,
	}).Error)
	require.NoError(t, logDB.Create(&model.Log{
		UserId: 11, Type: model.LogTypeConsume, RequestId: "request-1", Quota: 29,
	}).Error)

	assert.Equal(t, 29, QuotaForRequest("request-1", 11))
}

func TestCleanErrorMessageRedactsCredentialsAndSignedURLs(t *testing.T) {
	message := `upstream rejected https://example.test/image.png?X-Amz-Signature=url-secret&width=512 authorization=plain-secret Bearer bearer-secret sk-openai-secret`

	cleaned := CleanErrorMessage(message)

	assert.NotContains(t, cleaned, "url-secret")
	assert.NotContains(t, cleaned, "plain-secret")
	assert.NotContains(t, cleaned, "bearer-secret")
	assert.NotContains(t, cleaned, "sk-openai-secret")
	assert.Contains(t, cleaned, "width=512")
	assert.Contains(t, cleaned, "[redacted]")
}

func TestInterruptedImageGenerationIsFailedWithoutResubmission(t *testing.T) {
	db := setupStudioServiceTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.ImageStudioRun{}))
	run := &model.ImageStudioRun{PublicID: "expired", UserID: 7, IdempotencyKey: "expired", Status: model.ImageStudioRunStatusRunning, LeaseUntil: common.GetTimestamp() - 1}
	require.NoError(t, db.Create(run).Error)
	require.NoError(t, model.FailExpiredImageStudioRun(run, common.GetTimestamp()))
	require.NoError(t, db.First(run, run.ID).Error)
	assert.Equal(t, model.ImageStudioRunStatusFailed, run.Status)
	assert.Zero(t, run.LeaseUntil)
}
