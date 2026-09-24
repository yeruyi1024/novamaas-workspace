package controller

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSupplierTestModelsAllowsPrivateUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service.InitHttpClient()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/models", r.URL.Path)
		_, _ = io.WriteString(w, `{"data":[{"id":"demo-model"}]}`)
	}))
	defer upstream.Close()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	body := `{"base_url":"` + upstream.URL + `","api_key":"test-key"}`
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/supplier-test/models", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	ListSupplierTestModels(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "demo-model")
}

func TestRunSupplierTestAllowsPrivateUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service.InitHttpClient()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"id\":\"chatcmpl-1\",\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer upstream.Close()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	body := `{"base_url":"` + upstream.URL + `","api_key":"test-key","model":"demo","modules":["basic"],"basic":{"prompt":"hi","max_tokens":8,"stream":true,"checks":["connectivity"]}}`
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/supplier-test/runs", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	RunSupplierTest(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"type":"check"`)
	assert.Contains(t, recorder.Body.String(), "connectivity")
	assert.Contains(t, recorder.Body.String(), "[DONE]")
}

func TestRunSupplierTestInvalidURLStreamsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service.InitHttpClient()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	body := `{"base_url":"ftp://example.com","api_key":"k","model":"demo","modules":["basic"]}`
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/supplier-test/runs", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	RunSupplierTest(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"type":"error"`)
	assert.Contains(t, recorder.Body.String(), "[DONE]")
}
