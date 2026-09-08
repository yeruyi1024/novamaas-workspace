package relay

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

func TestTaskModel2DtoReportsRequestBodyWithoutIncludingItInListPayload(t *testing.T) {
	task := &model.Task{
		Properties: model.Properties{
			Input:       "legacy-input",
			RequestBody: json.RawMessage(`{"prompt":"private prompt"}`),
		},
	}

	result := TaskModel2Dto(task)

	assert.True(t, result.RequestBodyAvailable)
	properties, ok := result.Properties.(model.Properties)
	require.True(t, ok)
	assert.Empty(t, properties.RequestBody)
	assert.Equal(t, "legacy-input", properties.Input)
}

func TestVideoFetchByIDAllowsAdministratorTaskLogLookup(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Task{}))
	previousDB := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previousDB })
	require.NoError(t, db.Create(&model.Task{
		TaskID:   "task_other_user",
		UserId:   41,
		Platform: "54",
		Status:   model.TaskStatusSuccess,
	}).Error)

	adminContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	adminContext.Request = httptest.NewRequest(http.MethodGet, "/v1/video/generations/task_other_user", nil)
	adminContext.Params = gin.Params{{Key: "task_id", Value: "task_other_user"}}
	adminContext.Set("id", 99)
	adminContext.Set("role", common.RoleAdminUser)
	body, taskErr := videoFetchByIDRespBodyBuilder(adminContext)
	require.Nil(t, taskErr)
	assert.Equal(t, "task_other_user", gjson.GetBytes(body, "data.task_id").String())

	userContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	userContext.Request = httptest.NewRequest(http.MethodGet, "/v1/video/generations/task_other_user", nil)
	userContext.Params = gin.Params{{Key: "task_id", Value: "task_other_user"}}
	userContext.Set("id", 99)
	body, taskErr = videoFetchByIDRespBodyBuilder(userContext)
	assert.Nil(t, body)
	require.NotNil(t, taskErr)
	assert.Equal(t, "task_not_exist", taskErr.Code)
}
