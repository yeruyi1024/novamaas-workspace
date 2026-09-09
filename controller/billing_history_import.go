package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/gin-gonic/gin"
)

func ReviewBillingHistory(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	data, err := service.ReviewBillingHistory(c.Request.Context(), id, c.Query("month"))
	if err != nil {
		billingError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func ConfirmBillingHistoryImport(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	identity, live := middleware.GetSessionAuthIdentity(c)
	if !live || c.GetInt("role") < common.RoleAdminUser {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	var input struct {
		ID               string `json:"id"`
		Month            string `json:"month"`
		SourceSHA256     string `json:"source_sha256"`
		StorageProfileID int    `json:"storage_profile_id"`
		Acknowledged     bool   `json:"acknowledged"`
		Note             string `json:"note"`
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 16384), &input); err != nil || !input.Acknowledged {
		billingError(c, model.ErrBillingHistoryBlocked)
		return
	}
	store, err := storageService.OpenBillingArchiveStore(input.StorageProfileID, true)
	if err != nil {
		billingError(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()
	data, err := service.ConfirmBillingHistoryImport(ctx, id, identity.UserID, input.StorageProfileID, input.Month, input.ID, input.SourceSHA256, input.Note, identity.SessionID, store)
	if err != nil {
		billingError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func ListBillingHistoryImports(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	if _, _, err := model.BillingMonthBounds(c.Query("month")); err != nil {
		billingError(c, err)
		return
	}
	data, err := model.ListBillingHistoryImports(id, c.Query("month"))
	if err != nil {
		billingError(c, err)
		return
	}
	ids := make([]int, 0, len(data))
	for _, batch := range data {
		ids = append(ids, batch.ConfirmedBy)
	}
	identities, err := model.GetBillingUserIdentities(c.Request.Context(), ids)
	if err != nil {
		billingError(c, err)
		return
	}
	type importView struct {
		model.BillingHistoryImport
		ConfirmedByUsername string `json:"confirmed_by_username"`
	}
	views := make([]importView, 0, len(data))
	for _, batch := range data {
		views = append(views, importView{batch, identities[batch.ConfirmedBy].Username})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": views})
}

func DownloadBillingHistorySource(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	batch, err := model.GetBillingHistoryImport(id, c.Param("import_id"))
	if err != nil {
		billingError(c, err)
		return
	}
	var artifact model.BillingArtifact
	if err := model.DB.Where("id = ? AND statement_id = ? AND kind = ?", batch.ArtifactID, batch.ID, "source").First(&artifact).Error; err != nil {
		billingError(c, err)
		return
	}
	store, err := storageService.OpenBillingArchiveStore(batch.StorageProfileID, false)
	if err != nil {
		billingError(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	body, err := store.Read(ctx, artifact.ObjectKey, artifact.Size, artifact.SHA256)
	if err != nil {
		billingError(c, errors.New("historical billing source is unavailable"))
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"billing-history-%s.json.gz\"", batch.Month))
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Content-SHA256", artifact.SHA256)
	c.Data(http.StatusOK, "application/gzip", body)
}
