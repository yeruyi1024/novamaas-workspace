package controller

import (
	"net/http"

	adminmanual "github.com/QuantumNous/new-api/docs/admin-manual"
	"github.com/gin-gonic/gin"
)

func GetAdminManualAccess(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"available": true}})
}

func GetAdminManual(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": adminmanual.Documents})
}
