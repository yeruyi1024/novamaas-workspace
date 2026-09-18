package controller

import (
	"fmt"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/service/suppliertest"

	"github.com/gin-gonic/gin"
)

func ListSupplierTestModels(c *gin.Context) {
	var req struct {
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
	}
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request body"})
		return
	}
	if _, err := suppliertest.ModelsURL(req.BaseURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	models, err := suppliertest.ListModels(c.Request.Context(), suppliertest.NewHTTPClient(service.GetHttpClient()), req.BaseURL, req.APIKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    models,
	})
}

func RunSupplierTest(c *gin.Context) {
	var req suppliertest.RunRequest
	if err := common.DecodeJson(c.Request.Body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request body"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	flusher, _ := c.Writer.(http.Flusher)
	flush := func() {
		if flusher != nil {
			flusher.Flush()
		}
	}
	flush()

	emit := func(event suppliertest.Event) {
		payload, err := common.Marshal(event)
		if err != nil {
			return
		}
		fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
		flush()
	}

	if err := suppliertest.NormalizeRunRequest(&req); err != nil {
		emit(suppliertest.Event{Type: "error", Message: err.Error()})
		fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
		flush()
		return
	}
	if _, err := suppliertest.ChatCompletionsURL(req.BaseURL); err != nil {
		emit(suppliertest.Event{Type: "error", Message: err.Error()})
		fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
		flush()
		return
	}

	err := suppliertest.Run(c.Request.Context(), suppliertest.NewHTTPClient(service.GetHttpClient()), req, emit)
	if err != nil && c.Request.Context().Err() == nil {
		emit(suppliertest.Event{Type: "error", Message: err.Error()})
	}
	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	flush()
}
