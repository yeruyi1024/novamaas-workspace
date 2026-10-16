package controller

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/service/authz"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Upstream statements contain internal costs. Even explicit financial grants
// do not allow a customer role to enter this administrative workflow.
func requireUpstreamBillingAccess(c *gin.Context) bool {
	if c.GetInt("id") <= 0 || !canManageBillingAccounts(c) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "code": "UPSTREAM_BILLING_FORBIDDEN"})
		return false
	}
	return true
}

func requireUpstreamBillingReviewAccess(c *gin.Context) bool {
	if !requireUpstreamBillingAccess(c) {
		return false
	}
	if !authz.Can(c.GetInt("id"), c.GetInt("role"), authz.FinancialManagementView) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "code": "UPSTREAM_BILLING_REVIEW_FORBIDDEN"})
		return false
	}
	return true
}

func upstreamBillingError(c *gin.Context, err error) {
	status, code := http.StatusBadRequest, "UPSTREAM_BILLING_OPERATION_FAILED"
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		status, code = http.StatusNotFound, "UPSTREAM_BILLING_NOT_FOUND"
	case errors.Is(err, model.ErrBillingConflict):
		status, code = http.StatusConflict, "UPSTREAM_BILLING_CONFLICT"
	case errors.Is(err, model.ErrBillingEvidenceIntegrity):
		status, code = http.StatusConflict, "UPSTREAM_BILLING_INTEGRITY_FAILED"
	case errors.Is(err, model.ErrUpstreamBillingInvalidAttachment):
		status, code = http.StatusBadRequest, "UPSTREAM_BILLING_INVALID_ATTACHMENT"
	case errors.Is(err, model.ErrUpstreamBillingAttachmentLimit):
		status, code = http.StatusConflict, "UPSTREAM_BILLING_ATTACHMENT_LIMIT"
	case errors.Is(err, model.ErrUpstreamBillingUploadPending):
		status, code = http.StatusConflict, "UPSTREAM_BILLING_UPLOAD_PENDING"
	case errors.Is(err, service.ErrUpstreamBillingIncomplete):
		status, code = http.StatusConflict, "UPSTREAM_BILLING_INCOMPLETE"
	case errors.Is(err, service.ErrUpstreamBillingMonthOpen):
		status, code = http.StatusConflict, "UPSTREAM_BILLING_MONTH_OPEN"
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		status, code = http.StatusGatewayTimeout, "UPSTREAM_BILLING_TIMEOUT"
	default:
		common.SysError("upstream billing request failed: " + err.Error())
	}
	c.JSON(status, gin.H{"success": false, "code": code, "message": "Upstream billing operation could not be completed."})
}

func UpstreamBillingGroups(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	groups, err := model.ListUpstreamBillingGroups(c.Request.Context())
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	common.ApiSuccess(c, groups)
}

func SaveUpstreamBillingGroup(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	id := 0
	if raw := c.Param("group_id"); raw != "" {
		var err error
		id, err = strconv.Atoi(raw)
		if err != nil || id <= 0 {
			upstreamBillingError(c, errors.New("invalid upstream group"))
			return
		}
	}
	var input struct {
		Name       string `json:"name"`
		Version    int64  `json:"version"`
		ChannelIDs []int  `json:"channel_ids"`
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 32*1024), &input); err != nil {
		upstreamBillingError(c, errors.New("invalid upstream group request"))
		return
	}
	group, err := model.SaveUpstreamBillingGroup(c.Request.Context(), id, input.Version, input.Name, input.ChannelIDs, c.GetInt("id"))
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	common.ApiSuccess(c, group)
}

func UpstreamBillingChannels(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	channels, err := model.SearchUpstreamBillingChannels(c.Request.Context(), c.Query("q"))
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	common.ApiSuccess(c, channels)
}

func UpstreamBillingReport(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	id, err := strconv.Atoi(c.Query("group_id"))
	if err != nil || id <= 0 {
		upstreamBillingError(c, errors.New("invalid upstream group"))
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	report, err := service.GetUpstreamBillingReport(ctx, id, c.Query("month"), c.Query("start_date"), c.Query("end_date"))
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	common.ApiSuccess(c, report)
}

func UpstreamBillingUsageDetails(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	id, err := strconv.Atoi(c.Query("group_id"))
	if err != nil || id <= 0 {
		upstreamBillingError(c, errors.New("invalid upstream group"))
		return
	}
	cursor := int64(0)
	if raw := c.Query("cursor"); raw != "" {
		cursor, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || cursor < 0 {
			upstreamBillingError(c, errors.New("invalid upstream usage cursor"))
			return
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	details, err := service.GetUpstreamBillingUsageDetails(ctx, id, c.Query("month"), c.Query("start_date"), c.Query("end_date"), c.Query("day"), c.Query("model"), cursor)
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	common.ApiSuccess(c, details)
}

func UpstreamBillingStorageProfiles(c *gin.Context) {
	if requireUpstreamBillingAccess(c) {
		BillingStorageProfiles(c)
	}
}

func CreateUpstreamBillingStatement(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	var input struct {
		GroupID          int    `json:"group_id"`
		Month            string `json:"month"`
		StorageProfileID int    `json:"storage_profile_id"`
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 4096), &input); err != nil || input.GroupID <= 0 || input.StorageProfileID <= 0 {
		upstreamBillingError(c, errors.New("invalid upstream statement request"))
		return
	}
	if _, err := storageService.OpenBillingArchiveStore(input.StorageProfileID, true); err != nil {
		upstreamBillingError(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	statement, err := service.PrepareUpstreamBillingStatement(ctx, input.GroupID, c.GetInt("id"), input.StorageProfileID, input.Month)
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"success": true, "data": statement})
}

func ListUpstreamBillingStatements(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	groupID, before := 0, int64(0)
	if raw := c.Query("group_id"); raw != "" {
		var err error
		groupID, err = strconv.Atoi(raw)
		if err != nil || groupID <= 0 {
			upstreamBillingError(c, errors.New("invalid upstream group"))
			return
		}
	}
	if raw := c.Query("before"); raw != "" {
		var err error
		before, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || before < 0 {
			upstreamBillingError(c, errors.New("invalid upstream statement cursor"))
			return
		}
	}
	statements, err := model.ListUpstreamBillingStatements(groupID, before, c.Query("before_id"), c.Query("status"))
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	common.ApiSuccess(c, statements)
}

func GetUpstreamBillingStatement(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	statement, err := model.GetUpstreamBillingStatement(c.Param("statement_id"))
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	if err := statement.VerifySnapshot(); err != nil {
		upstreamBillingError(c, err)
		return
	}
	var snapshot service.UpstreamBillingSnapshot
	if err := common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot); err != nil {
		upstreamBillingError(c, err)
		return
	}
	var events []model.UpstreamBillingStatementEvent
	if err := model.DB.WithContext(c.Request.Context()).Where("statement_id = ?", statement.ID).Order("id asc").Find(&events).Error; err != nil {
		upstreamBillingError(c, err)
		return
	}
	review, err := model.GetUpstreamBillingReview(statement.ID)
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	actorIDs := []int{statement.CreatedBy, statement.ConfirmedBy}
	for _, event := range events {
		actorIDs = append(actorIDs, event.ActorID)
	}
	for _, attachment := range review.Attachments {
		actorIDs = append(actorIDs, attachment.UploadedBy)
	}
	identities, err := model.GetBillingUserIdentities(c.Request.Context(), actorIDs)
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	type eventView struct {
		model.UpstreamBillingStatementEvent
		ActorUsername string `json:"actor_username"`
	}
	eventViews := make([]eventView, 0, len(events))
	for _, event := range events {
		eventViews = append(eventViews, eventView{event, identities[event.ActorID].Username})
	}
	artifacts := make([]model.BillingArtifact, 0)
	if err := model.DB.WithContext(c.Request.Context()).Where("statement_id = ? AND kind <> ?", statement.ID, "details").Order("kind asc").Find(&artifacts).Error; err != nil {
		upstreamBillingError(c, err)
		return
	}
	var detailChunks []model.BillingArtifact
	if err := model.DB.WithContext(c.Request.Context()).Where("statement_id = ? AND kind = ?", statement.ID, "details").Order("ordinal asc").Limit(200).Find(&detailChunks).Error; err != nil {
		upstreamBillingError(c, err)
		return
	}
	var detailCount int64
	if err := model.DB.WithContext(c.Request.Context()).Model(&model.BillingArtifact{}).Where("statement_id = ? AND kind = ?", statement.ID, "details").Count(&detailCount).Error; err != nil {
		upstreamBillingError(c, err)
		return
	}
	artifacts = append(artifacts, detailChunks...)
	type attachmentView struct {
		model.UpstreamBillingAttachment
		UploadedUsername string `json:"uploaded_username"`
	}
	attachments := make([]attachmentView, 0, len(review.Attachments))
	for _, attachment := range review.Attachments {
		attachments = append(attachments, attachmentView{attachment, identities[attachment.UploadedBy].Username})
	}
	common.ApiSuccess(c, gin.H{"statement": statement, "snapshot": snapshot, "events": eventViews, "artifacts": artifacts, "detail_count": detailCount, "attachments": attachments, "evidence_sha256": review.EvidenceSHA256, "pending_uploads": review.PendingUploads})
}

func ActOnUpstreamBillingStatement(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	var input struct {
		Action         string `json:"action"`
		ManifestSHA256 string `json:"manifest_sha256"`
		EvidenceSHA256 string `json:"evidence_sha256"`
		Reason         string `json:"reason"`
		Acknowledged   bool   `json:"acknowledged"`
	}
	if err := common.DecodeJson(http.MaxBytesReader(c.Writer, c.Request.Body, 8192), &input); err != nil {
		upstreamBillingError(c, errors.New("invalid upstream statement action"))
		return
	}
	if (input.Action == "confirm" || input.Action == "void") && !requireUpstreamBillingReviewAccess(c) {
		return
	}
	sessionID := ""
	if input.Action == "confirm" {
		identity, ok := middleware.GetSessionAuthIdentity(c)
		if !ok || identity.UserID != c.GetInt("id") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "code": "UPSTREAM_BILLING_SESSION_REQUIRED"})
			return
		}
		if !input.Acknowledged {
			upstreamBillingError(c, errors.New("upstream statement confirmation requires acknowledgement"))
			return
		}
		sessionID = identity.SessionID
	}
	statement, err := model.ActUpstreamBillingStatementWithReview(c.Param("statement_id"), input.Action, input.ManifestSHA256, input.EvidenceSHA256, input.Reason, sessionID, c.GetInt("id"))
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	common.ApiSuccess(c, statement)
}

func UploadUpstreamBillingAttachment(c *gin.Context) {
	if !requireUpstreamBillingReviewAccess(c) {
		return
	}
	statement, err := model.GetUpstreamBillingStatement(c.Param("statement_id"))
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	if statement.Status != model.StatementDraft {
		upstreamBillingError(c, model.ErrBillingConflict)
		return
	}
	const maxAttachmentBytes int64 = 20 * 1024 * 1024
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAttachmentBytes+64*1024)
	err = c.Request.ParseMultipartForm(1024 * 1024)
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	if err != nil || c.Request.MultipartForm == nil || len(c.Request.MultipartForm.File) != 1 || len(c.Request.MultipartForm.File["file"]) != 1 {
		upstreamBillingError(c, model.ErrUpstreamBillingInvalidAttachment)
		return
	}
	header := c.Request.MultipartForm.File["file"][0]
	if header.Size <= 0 || header.Size > maxAttachmentBytes {
		upstreamBillingError(c, model.ErrUpstreamBillingInvalidAttachment)
		return
	}
	file, err := header.Open()
	if err != nil {
		upstreamBillingError(c, model.ErrUpstreamBillingInvalidAttachment)
		return
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, maxAttachmentBytes+1))
	if err != nil || len(body) == 0 || int64(len(body)) > maxAttachmentBytes {
		upstreamBillingError(c, model.ErrUpstreamBillingInvalidAttachment)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	attachment, err := service.UploadUpstreamBillingAttachment(ctx, statement.ID, c.GetInt("id"), header.Filename, body)
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": attachment})
}

func DownloadUpstreamBillingAttachment(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	statement, err := model.GetUpstreamBillingStatement(c.Param("statement_id"))
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	review, err := model.GetUpstreamBillingReview(statement.ID)
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	var attachment *model.UpstreamBillingAttachment
	for i := range review.Attachments {
		if review.Attachments[i].ID == c.Param("attachment_id") {
			attachment = &review.Attachments[i]
			break
		}
	}
	if attachment == nil {
		upstreamBillingError(c, gorm.ErrRecordNotFound)
		return
	}
	if attachment.StorageProfileID != statement.StorageProfileID {
		upstreamBillingError(c, model.ErrBillingEvidenceIntegrity)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	store, err := storageService.OpenBillingArchiveStore(attachment.StorageProfileID, false)
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	body, err := store.ReadUpstreamAttachment(ctx, attachment.ObjectKey, attachment.Size, attachment.SHA256)
	if err != nil {
		common.SysError("upstream billing attachment read failed: " + err.Error())
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": attachment.OriginalFilename}))
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Content-SHA256", attachment.SHA256)
	c.Data(http.StatusOK, attachment.ContentType, body)
}

func DownloadUpstreamBillingArtifact(c *gin.Context) {
	if !requireUpstreamBillingAccess(c) {
		return
	}
	statement, err := model.GetUpstreamBillingStatement(c.Param("statement_id"))
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	if err := statement.VerifySnapshot(); err != nil {
		upstreamBillingError(c, err)
		return
	}
	kind, ordinal := c.Param("kind"), 0
	switch kind {
	case "pdf", "xlsx", "details", "manifest", "receipt", "snapshot":
	default:
		upstreamBillingError(c, gorm.ErrRecordNotFound)
		return
	}
	ordinal, err = strconv.Atoi(c.Param("ordinal"))
	if err != nil || ordinal < 0 || (kind != "details" && ordinal != 0) || (kind == "receipt" && statement.Status != model.StatementConfirmed) {
		upstreamBillingError(c, gorm.ErrRecordNotFound)
		return
	}
	var snapshot service.UpstreamBillingSnapshot
	if err := common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot); err != nil {
		upstreamBillingError(c, err)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	store, err := storageService.OpenBillingArchiveStore(statement.StorageProfileID, false)
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	var artifact model.BillingArtifact
	err = model.DB.WithContext(ctx).Where("statement_id = ? AND kind = ? AND ordinal = ?", statement.ID, kind, ordinal).First(&artifact).Error
	if kind == "receipt" && errors.Is(err, gorm.ErrRecordNotFound) {
		pdf, renderErr := service.RenderUpstreamBillingStatementPDF(statement, &snapshot, true)
		if renderErr != nil {
			upstreamBillingError(c, renderErr)
			return
		}
		result, putErr := store.PutUpstream(ctx, statement.ID, kind, 0, statement.CreatedBy, "application/pdf", pdf, 0)
		if putErr != nil {
			upstreamBillingError(c, putErr)
			return
		}
		artifact, err = *result, nil
	}
	if err != nil {
		upstreamBillingError(c, err)
		return
	}
	body, err := store.Read(ctx, artifact.ObjectKey, artifact.Size, artifact.SHA256)
	if err != nil {
		common.SysError("upstream billing archive read failed: " + err.Error())
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}
	contentType := "application/pdf"
	switch kind {
	case "details":
		contentType = "application/gzip"
	case "manifest", "snapshot":
		contentType = "application/json"
	case "xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": service.UpstreamBillingArtifactFilename(statement, &snapshot, kind, ordinal)}))
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Content-SHA256", artifact.SHA256)
	c.Data(http.StatusOK, contentType, body)
}
