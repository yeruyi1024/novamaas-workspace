package controller

import (
	"net/http"

	apidocs "github.com/QuantumNous/new-api/docs/platform-api"
	"github.com/gin-gonic/gin"
)

func GetAPIDocumentation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": apidocs.Documents})
}

func GetAdminDocumentation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": apidocs.AdminDocuments})
}
