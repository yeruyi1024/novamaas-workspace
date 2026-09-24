package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAKSKAssetUploadAppearsInUserUsageLogs(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Log{}))
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousMainDBType, previousLogDBType := common.MainDatabaseType(), common.LogDatabaseType()
	previousRedisEnabled := common.RedisEnabled
	model.DB, model.LOG_DB = db, db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.RedisEnabled = false
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(previousMainDBType, previousLogDBType)
		common.RedisEnabled = previousRedisEnabled
	})

	user := model.User{Username: "asset-uploader", Password: "password", Status: common.UserStatusEnabled}
	require.NoError(t, db.Create(&user).Error)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v3/?Action=CreateAsset&Version=2024-01-01", nil)
	c.Request.RemoteAddr = "203.0.113.10:4321"
	recordVolcAssetUploadAudit(c, &assetService.VolcActionPrincipal{
		OwnerUserID:   user.Id,
		AccessKeyID:   "AKNMEXAMPLE",
		AccessKeyName: "production uploader",
	}, &assetService.AssetView{
		ID:      "asset-123",
		GroupID: "group-456",
		Name:    "product photo",
		Type:    model.AssetTypeImage,
	})

	logs, total, err := model.GetUserLogs(user.Id, model.LogTypeManage, 0, 0, "", "", 0, 10, "", "", "")
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, logs, 1)
	assert.Equal(t, "asset-uploader", logs[0].Username)
	assert.Equal(t, "203.0.113.10", logs[0].Ip)
	assert.Equal(t, "Uploaded asset product photo via AK/SK access key production uploader (ID: asset-123)", logs[0].Content)

	var other map[string]interface{}
	require.NoError(t, common.UnmarshalJsonStr(logs[0].Other, &other))
	assert.NotContains(t, other, "audit_info")
	op, ok := other["op"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "asset.upload_aksk", op["action"])
	params, ok := op["params"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "asset-123", params["id"])
	assert.Equal(t, "group-456", params["groupId"])
	assert.Equal(t, "image", params["assetType"])
	assert.Equal(t, "AKNMEXAMPLE", params["accessKeyId"])
	assert.Equal(t, "production uploader", params["accessKeyName"])
}

func TestAKSKAssetResponseExposesContentRejectionWithoutRawProviderMessage(t *testing.T) {
	asset := &assetService.AssetView{
		ID: "asset-rejected", Name: "portrait", Type: model.AssetTypeImage,
		Status: model.AssetStatusUnavailable, UnavailableReason: model.AssetUnavailableRealPerson,
	}
	response := volcAssetResponse(asset, "https://example.com/preview")
	assert.Equal(t, "Failed", response["Status"])
	assert.Equal(t, model.AssetUnavailableRealPerson, response["FailureReason"])
	assert.NotContains(t, response, "LastError")
}
