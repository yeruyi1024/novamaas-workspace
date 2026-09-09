package controller

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// An omitted account always means the authenticated account. No client role
// or ownership field is trusted.
func billingUserID(c *gin.Context) (int, bool) {
	id := c.GetInt("id")
	raw := c.Query("user_id")
	if raw == "" {
		return id, id > 0
	}
	target, err := strconv.Atoi(raw)
	if err != nil || target <= 0 {
		billingError(c, errors.New("invalid billing account"))
		return 0, false
	}
	if target != id && c.GetInt("role") < common.RoleAdminUser {
		c.AbortWithStatus(http.StatusForbidden)
		return 0, false
	}
	return target, true
}
func billingError(c *gin.Context, err error) {
	status, code := http.StatusBadRequest, "BILLING_OPERATION_FAILED"
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		status, code = http.StatusNotFound, "BILLING_NOT_FOUND"
	case errors.Is(err, model.ErrBillingConflict):
		status, code = http.StatusConflict, "BILLING_CONFLICT"
	case errors.Is(err, model.ErrBillingNotConfigured):
		code = "BILLING_NOT_CONFIGURED"
	case errors.Is(err, service.ErrBillingHistoricalDataUnreconciled):
		status, code = http.StatusConflict, "BILLING_HISTORY_UNRECONCILED"
	case errors.Is(err, model.ErrBillingHistoryBlocked):
		status, code = http.StatusConflict, "BILLING_HISTORY_BLOCKED"
	case errors.Is(err, service.ErrBillingDocumentBranding):
		code = "BILLING_BRANDING_INVALID"
		common.SysError(err.Error())
	case strings.Contains(err.Error(), "24 hours"):
		code = "BILLING_MONTH_OPEN"
	case strings.Contains(err.Error(), "unfinished billing"):
		code = "BILLING_PENDING"
	case strings.Contains(err.Error(), "company title and tax ID are required"):
		code = "BILLING_IDENTITY_REQUIRED"
	case strings.Contains(err.Error(), "active statement already exists"):
		code = "BILLING_EXISTS"
	case strings.Contains(err.Error(), "accounting start"):
		code = "BILLING_START_LOCKED"
	default:
		common.SysError("billing request failed: " + err.Error())
	}
	c.JSON(status, gin.H{"success": false, "code": code, "message": "Billing operation could not be completed."})
}
func BillingDay(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	data, err := service.GetBillingDayContext(c.Request.Context(), id, c.Query("date"))
	if err != nil {
		billingError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
func BillingAccount(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	data, err := model.GetBillingAccount(id)
	if err != nil {
		billingError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
func UpdateBillingAccount(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	var input struct {
		CompanyTitle      string `json:"company_title"`
		TaxID             string `json:"tax_id"`
		ProfileVersion    int64  `json:"profile_version"`
		AccountingStartAt *int64 `json:"accounting_start_at"`
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 8192), &input); err != nil {
		billingError(c, errors.New("invalid billing profile"))
		return
	}
	if input.AccountingStartAt != nil && c.GetInt("role") < common.RoleAdminUser {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	if id != c.GetInt("id") || input.AccountingStartAt != nil {
		target, err := model.GetUserById(id, false)
		if err != nil {
			billingError(c, err)
			return
		}
		if target.Role > c.GetInt("role") || (target.Role == c.GetInt("role") && id != c.GetInt("id") && c.GetInt("role") != common.RoleRootUser) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
	data, err := model.SaveBillingAccount(id, c.GetInt("id"), input.ProfileVersion, input.CompanyTitle, input.TaxID, input.AccountingStartAt)
	if err != nil {
		billingError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
func BillingMonthPreview(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	profileID, _ := strconv.Atoi(c.Query("storage_profile_id"))
	data, err := service.GetBillingMonthPreview(c.Request.Context(), id, c.Query("month"), profileID)
	if err != nil {
		billingError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func BillingUsageDetails(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	hour, err := strconv.Atoi(c.Query("hour"))
	if err != nil {
		billingError(c, errors.New("invalid usage billing hour"))
		return
	}
	data, err := service.GetBillingUsageDetails(c.Request.Context(), id, c.Query("date"), hour, c.Query("cursor"))
	if err != nil {
		billingError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
func CreateBillingStatement(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	var input struct {
		Month            string `json:"month"`
		StorageProfileID int    `json:"storage_profile_id"`
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 4096), &input); err != nil {
		billingError(c, err)
		return
	}
	if _, err := storageService.OpenBillingArchiveStore(input.StorageProfileID, true); err != nil {
		billingError(c, err)
		return
	}
	data, err := service.PrepareBillingStatementContext(c.Request.Context(), id, c.GetInt("id"), input.StorageProfileID, input.Month)
	if err != nil {
		billingError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"success": true, "data": data})
}
func ListBillingStatements(c *gin.Context) {
	id, ok := billingUserID(c)
	if !ok {
		return
	}
	before, _ := strconv.ParseInt(c.Query("before"), 10, 64)
	data, err := model.ListBillingStatements(id, before, c.Query("before_id"), c.Query("status"), c.GetInt("role") < common.RoleAdminUser)
	if err != nil {
		billingError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
func authorizedBillingStatement(c *gin.Context) (*model.BillingStatement, bool) {
	statement, err := model.GetBillingStatement(c.Param("statement_id"))
	if err != nil {
		billingError(c, err)
		return nil, false
	}
	if c.GetInt("role") < common.RoleAdminUser && (statement.UserID != c.GetInt("id") || statement.IssuedAt == 0) {
		billingError(c, gorm.ErrRecordNotFound)
		return nil, false
	}
	return statement, true
}
func GetBillingStatement(c *gin.Context) {
	statement, ok := authorizedBillingStatement(c)
	if !ok {
		return
	}
	if err := statement.VerifySnapshot(); err != nil {
		billingError(c, err)
		return
	}
	events := make([]model.BillingStatementEvent, 0)
	if err := model.DB.Where("statement_id = ?", statement.ID).Order("id asc").Find(&events).Error; err != nil {
		billingError(c, err)
		return
	}
	artifacts := make([]model.BillingArtifact, 0)
	if err := model.DB.Where("statement_id = ? AND kind <> ?", statement.ID, "details").Order("kind asc").Find(&artifacts).Error; err != nil {
		billingError(c, err)
		return
	}
	var details []model.BillingArtifact
	query := model.DB.Where("statement_id = ? AND kind = ?", statement.ID, "details")
	if err := query.Order("ordinal asc").Limit(200).Find(&details).Error; err != nil {
		billingError(c, err)
		return
	}
	var detailCount int64
	if err := model.DB.Model(&model.BillingArtifact{}).Where("statement_id = ? AND kind = ?", statement.ID, "details").Count(&detailCount).Error; err != nil {
		billingError(c, err)
		return
	}
	artifacts = append(artifacts, details...)
	ids := []int{statement.UserID}
	for _, event := range events {
		ids = append(ids, event.ActorID)
	}
	identities, err := model.GetBillingUserIdentities(c.Request.Context(), ids)
	if err != nil {
		billingError(c, err)
		return
	}
	type eventView struct {
		model.BillingStatementEvent
		ActorUsername string `json:"actor_username"`
	}
	eventViews := make([]eventView, 0, len(events))
	for _, event := range events {
		eventViews = append(eventViews, eventView{event, identities[event.ActorID].Username})
	}
	warning := ""
	if statement.Status != model.StatementConfirmed && statement.Status != model.StatementVoid {
		if err := service.ValidateBillingStatementSource(c.Request.Context(), statement); err != nil {
			if !errors.Is(err, service.ErrBillingHistoricalDataUnreconciled) {
				billingError(c, err)
				return
			}
			warning = "historical_data_unreconciled"
		}
	}
	customer := identities[statement.UserID]
	customer.ID = statement.UserID // retain ownership even if a user was deleted
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"statement": statement, "customer": customer, "events": eventViews, "artifacts": artifacts, "detail_count": detailCount, "source_warning": warning}})
}
func ActOnBillingStatement(c *gin.Context) {
	statement, ok := authorizedBillingStatement(c)
	if !ok {
		return
	}
	var input struct {
		Action         string `json:"action"`
		ManifestSHA256 string `json:"manifest_sha256"`
		Note           string `json:"note"`
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 16384), &input); err != nil {
		billingError(c, err)
		return
	}
	admin := c.GetInt("role") >= common.RoleAdminUser
	identity, session := middleware.GetSessionAuthIdentity(c)
	// All legal acknowledgements require a real owner session, never an admin
	// acting for another customer or a personal access token.
	if input.Action == "confirm" || input.Action == "dispute" {
		if !session || identity.UserID != statement.UserID {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	} else if !admin {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	if (input.Action == "issue" || input.Action == "confirm") && statement.Status != model.StatementConfirmed {
		if err := service.ValidateBillingStatementSource(c.Request.Context(), statement); err != nil {
			billingError(c, err)
			return
		}
	}
	data, err := model.ChangeBillingStatement(statement.ID, input.Action, input.ManifestSHA256, strings.TrimSpace(input.Note), identity.SessionID, c.GetInt("id"), admin)
	if err != nil {
		billingError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
func BillingStorageProfiles(c *gin.Context) {
	profiles, err := model.ListStorageProfiles()
	if err != nil {
		billingError(c, err)
		return
	}
	type profile struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	data := make([]profile, 0)
	for _, item := range profiles {
		if item.Status == model.StorageProfileStatusEnabled && item.ProviderType == model.StorageProviderAliyunOSS {
			data = append(data, profile{item.ID, item.Name})
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
func DownloadBillingArtifact(c *gin.Context) {
	statement, ok := authorizedBillingStatement(c)
	if !ok {
		return
	}
	if err := statement.VerifySnapshot(); err != nil {
		billingError(c, err)
		return
	}
	kind := c.Param("kind")
	if kind != "pdf" && kind != "details" && kind != "manifest" && kind != "receipt" && kind != "snapshot" {
		billingError(c, gorm.ErrRecordNotFound)
		return
	}
	ordinal, err := strconv.Atoi(c.Param("ordinal"))
	if err != nil || ordinal < 0 {
		billingError(c, gorm.ErrRecordNotFound)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	store, err := storageService.OpenBillingArchiveStore(statement.StorageProfileID, false)
	if err != nil {
		billingError(c, err)
		return
	}
	var artifact model.BillingArtifact
	err = model.DB.Where("statement_id = ? AND kind = ? AND ordinal = ?", statement.ID, kind, ordinal).First(&artifact).Error
	if kind == "receipt" && ordinal == 0 && statement.Status == model.StatementConfirmed && errors.Is(err, gorm.ErrRecordNotFound) {
		var snapshot service.BillingSnapshot
		if err = common.UnmarshalJsonStr(statement.Snapshot, &snapshot); err != nil {
			billingError(c, err)
			return
		}
		pdf, renderErr := service.RenderBillingStatementPDF(statement, &snapshot, true)
		if renderErr != nil {
			billingError(c, renderErr)
			return
		}
		result, putErr := store.Put(ctx, statement.ID, kind, 0, statement.UserID, "application/pdf", pdf, 0)
		if putErr != nil {
			billingError(c, putErr)
			return
		}
		artifact, err = *result, nil
	}
	if err != nil {
		billingError(c, err)
		return
	}
	body, err := store.Read(ctx, artifact.ObjectKey, artifact.Size, artifact.SHA256)
	if err != nil {
		common.SysError(err.Error())
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}
	contentType := "application/pdf"
	if kind == "details" {
		contentType = "application/gzip"
	}
	if kind == "manifest" || kind == "snapshot" {
		contentType = "application/json"
	}
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": service.BillingArtifactFilename(statement, kind, ordinal)}))
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Content-SHA256", artifact.SHA256)
	c.Data(http.StatusOK, contentType, body)
}
