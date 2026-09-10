package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageProfileTestReportsMissingCredentialEncryptionKey(t *testing.T) {
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "")
	t.Setenv("CRYPTO_SECRET", "")
	t.Setenv("SESSION_SECRET", "")

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/storage/profiles/test", strings.NewReader(`{
		"name":"Primary OSS",
		"provider_type":"aliyun_oss",
		"status":1,
		"endpoint":"https://oss-cn-hangzhou.aliyuncs.com",
		"region":"cn-hangzhou",
		"bucket":"test-media-bucket",
		"auth_type":"static_access_key",
		"access_key_id":"LTAI1234567890",
		"access_key_secret":"super-secret-value"
	}`))

	TestStorageProfileInput(c)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "STORAGE_CREDENTIAL_ENCRYPTION_KEY")
}
