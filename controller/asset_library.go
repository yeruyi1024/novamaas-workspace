package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"
	storageService "github.com/QuantumNous/new-api/service/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListAssetGroups(c *gin.Context) {
	ownerID := c.GetInt("id")
	includeAll := c.GetInt("role") >= common.RoleAdminUser && c.Query("scope") == "all"
	if requestedOwner, _ := strconv.Atoi(c.Query("owner_user_id")); includeAll && requestedOwner > 0 {
		ownerID, includeAll = requestedOwner, false
	}
	if c.Query("p") != "" || c.Query("page_size") != "" || c.Query("search") != "" || c.Query("group_id") != "" {
		pageInfo := common.GetPageQuery(c)
		input := assetService.GroupListInput{
			OwnerUserID:      ownerID,
			IncludeAllOwners: includeAll,
			Search:           c.Query("search"), Page: pageInfo.GetPage(), PageSize: pageInfo.GetPageSize(),
			SortBy: "Name", SortOrder: "Asc",
		}
		if groupID := strings.TrimSpace(c.Query("group_id")); groupID != "" {
			input.GroupIDs = []string{groupID}
		}
		groups, err := assetService.ListGroupsPage(input)
		if err != nil {
			assetLibraryError(c, err)
			return
		}
		ownerIDs := make([]int, 0, len(groups.Items))
		for _, group := range groups.Items {
			ownerIDs = append(ownerIDs, group.OwnerUserID)
		}
		ownerNames, err := model.GetUsernamesByIDs(ownerIDs)
		if err != nil {
			assetLibraryError(c, err)
			return
		}
		groupIDs := make([]int64, 0, len(groups.Items))
		for _, group := range groups.Items {
			groupIDs = append(groupIDs, group.ID)
		}
		assetCounts, err := assetService.CountGroupAssets(groupIDs)
		if err != nil {
			assetLibraryError(c, err)
			return
		}
		items := make([]assetService.GroupView, 0, len(groups.Items))
		for _, group := range groups.Items {
			assetCount := assetCounts[group.ID]
			items = append(items, assetService.GroupView{
				ID: group.PublicID, OwnerUserID: group.OwnerUserID, OwnerName: ownerNames[group.OwnerUserID], AssetCount: &assetCount,
				Name: group.Name, Description: group.Description, Status: group.Status,
				CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"items": items, "total": groups.Total, "page": groups.Page, "page_size": groups.PageSize,
		}})
		return
	}
	groups, err := assetService.ListGroups(ownerID, includeAll)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": groups})
}

func CreateAssetGroup(c *gin.Context) {
	var input assetService.GroupInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset group request")})
		return
	}
	group, err := assetService.CreateGroup(c.GetInt("id"), input)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": group})
}

func UpdateAssetGroup(c *gin.Context) {
	var input assetService.GroupInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset group request")})
		return
	}
	targetOwner, _ := strconv.Atoi(c.Query("owner_user_id"))
	group, err := assetService.UpdateGroup(c.Param("id"), c.GetInt("id"), c.GetInt("role") >= common.RoleAdminUser, input, targetOwner)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": group})
}

func DeleteAssetGroup(c *gin.Context) {
	targetOwner, _ := strconv.Atoi(c.Query("owner_user_id"))
	isAdmin := c.GetInt("role") >= common.RoleAdminUser
	if c.Query("with_assets") != "true" {
		if err := assetService.DeleteGroup(c.Param("id"), c.GetInt("id"), isAdmin, targetOwner); err != nil {
			assetLibraryError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}
	result, err := assetService.DeleteGroupWithAssets(c.Param("id"), c.GetInt("id"), isAdmin, targetOwner)
	if isAdmin && result != nil && result.DeletedAssetCount > 0 {
		recordManageAudit(c, "asset_library.group_purge", map[string]interface{}{
			"group_id": c.Param("id"), "owner_user_id": result.OwnerUserID,
			"deleted_asset_count": result.DeletedAssetCount, "deleted_group_count": result.DeletedGroupCount,
		})
	}
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// applyAssetTypeAndStatusFilters parses the media type and status filters
// shared across asset listing, summary counts, and filter-based batch deletion
// so what the user sees in the gallery always matches the deletion scope.
func applyAssetTypeAndStatusFilters(input *assetService.AssetListInput, assetTypeRaw string, statusRaw string) error {
	switch assetType := strings.TrimSpace(assetTypeRaw); assetType {
	case "":
	case model.AssetTypeImage, model.AssetTypeVideo, model.AssetTypeAudio:
		input.AssetTypes = []string{assetType}
	default:
		return &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset type filter")}
	}
	switch status := strings.TrimSpace(statusRaw); status {
	case "":
	case model.AssetStatusReady, model.AssetStatusProcessing:
		input.Statuses = []string{status}
	case model.AssetStatusUnavailable:
		input.Statuses = []string{model.AssetStatusUnavailable, model.AssetStatusSyncFailed}
	default:
		return &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset status filter")}
	}
	return nil
}

// assetListInputFromQuery reads the asset list scope and filters shared by
// the list and summary endpoints.
func assetListInputFromQuery(c *gin.Context) (assetService.AssetListInput, error) {
	targetOwner, _ := strconv.Atoi(c.Query("owner_user_id"))
	input := assetService.AssetListInput{
		OwnerUserID:       c.GetInt("id"),
		IncludeAllOwners:  c.GetInt("role") >= common.RoleAdminUser && c.Query("scope") == "all",
		TargetOwnerUserID: targetOwner,
		GroupPublicID:     strings.TrimSpace(c.Query("group_id")),
		Search:            strings.TrimSpace(c.Query("search")),
	}
	if err := applyAssetTypeAndStatusFilters(&input, c.Query("type"), c.Query("status")); err != nil {
		return input, err
	}
	switch sortOrder := strings.TrimSpace(c.Query("sort")); sortOrder {
	case "", "newest":
	case "oldest":
		input.SortOrder = "Asc"
	default:
		return input, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset sort order")}
	}
	return input, nil
}

func ListMediaAssets(c *gin.Context) {
	input, err := assetListInputFromQuery(c)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	pageInfo := common.GetPageQuery(c)
	input.Context = c.Request.Context()
	input.Page, input.PageSize = pageInfo.GetPage(), pageInfo.GetPageSize()
	input.IncludePreviews = c.Query("include_previews") != "false"
	assets, err := assetService.ListAssets(input)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": assets})
}

func SummarizeMediaAssets(c *gin.Context) {
	input, err := assetListInputFromQuery(c)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	summary, err := assetService.SummarizeAssets(input)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": summary})
}

type assetBatchRequest struct {
	Items []assetService.AssetRef `json:"items"`
}

func BatchDeleteMediaAssets(c *gin.Context) {
	var request assetBatchRequest
	if err := common.DecodeJson(c.Request.Body, &request); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset batch request")})
		return
	}
	isAdmin := c.GetInt("role") >= common.RoleAdminUser
	deleted, err := assetService.DeleteAssets(request.Items, c.GetInt("id"), isAdmin)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	if isAdmin {
		ownerSeen := make(map[int]bool, len(deleted))
		ownerIDs := make([]int, 0, len(deleted))
		assetIDs := make([]string, 0, len(deleted))
		for _, item := range deleted {
			assetIDs = append(assetIDs, item.ID)
			if !ownerSeen[item.OwnerUserID] {
				ownerSeen[item.OwnerUserID] = true
				ownerIDs = append(ownerIDs, item.OwnerUserID)
			}
		}
		recordManageAudit(c, "asset_library.assets_batch_delete", map[string]interface{}{
			"deleted_asset_count": len(deleted),
			"owner_user_ids":      ownerIDs,
			"asset_ids":           assetIDs,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"deleted_count": len(deleted)}})
}

type assetFilterDeleteRequest struct {
	Scope       string                  `json:"scope"`
	OwnerUserID int                     `json:"owner_user_id"`
	GroupID     string                  `json:"group_id"`
	Type        string                  `json:"type"`
	Status      string                  `json:"status"`
	Search      string                  `json:"search"`
	Exclude     []assetService.AssetRef `json:"exclude"`
}

func BatchDeleteMediaAssetsByFilter(c *gin.Context) {
	var request assetFilterDeleteRequest
	if err := common.DecodeJson(c.Request.Body, &request); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset batch filter request")})
		return
	}
	isAdmin := c.GetInt("role") >= common.RoleAdminUser
	input := assetService.AssetListInput{
		OwnerUserID:       c.GetInt("id"),
		IncludeAllOwners:  isAdmin && request.Scope == "all",
		TargetOwnerUserID: request.OwnerUserID,
		GroupPublicID:     strings.TrimSpace(request.GroupID),
		Search:            strings.TrimSpace(request.Search),
	}
	if err := applyAssetTypeAndStatusFilters(&input, request.Type, request.Status); err != nil {
		assetLibraryError(c, err)
		return
	}
	result, err := assetService.DeleteAssetsByFilter(input, request.Exclude, c.GetInt("id"), isAdmin)
	if isAdmin && result != nil && result.DeletedCount > 0 {
		recordManageAudit(c, "asset_library.assets_filter_delete", map[string]interface{}{
			"deleted_asset_count": result.DeletedCount,
			"owner_user_ids":      result.OwnerUserIDs,
			"target_owner_id":     request.OwnerUserID,
			"group_id":            input.GroupPublicID,
			"type":                request.Type,
			"status":              request.Status,
			"search":              input.Search,
		})
	}
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func ListAssetOwners(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	owners, err := assetService.ListAssetOwners(c.Query("search"), pageInfo.GetPage(), pageInfo.GetPageSize())
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": owners})
}

func PreviewMediaAssets(c *gin.Context) {
	var request assetBatchRequest
	if err := common.DecodeJson(c.Request.Body, &request); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset batch request")})
		return
	}
	previews, err := assetService.PreviewThumbnails(c.Request.Context(), request.Items, c.GetInt("id"), c.GetInt("role") >= common.RoleAdminUser)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": previews})
}

func UploadMediaAsset(c *gin.Context) {
	policy, err := storageService.GetAssetLibraryPolicy()
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	if policy.MaxFileBytes <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusServiceUnavailable, Err: errors.New("asset library storage is not configured")})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, policy.MaxFileBytes+1024*1024)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		statusCode := http.StatusBadRequest
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			statusCode = http.StatusRequestEntityTooLarge
		}
		assetLibraryError(c, &assetService.RequestError{StatusCode: statusCode, Err: errors.New("asset file is required or exceeds the configured size limit")})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	defer file.Close()
	header := make([]byte, 512)
	read, readErr := io.ReadFull(file, header)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) && !errors.Is(readErr, io.EOF) {
		assetLibraryError(c, readErr)
		return
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		assetLibraryError(c, err)
		return
	}
	detectedType := normalizeUploadedAssetContentType(
		http.DetectContentType(header[:read]),
		fileHeader.Header.Get("Content-Type"),
	)
	asset, err := assetService.CreateAsset(c.Request.Context(), c.GetInt("id"), assetService.AssetInput{
		GroupPublicID: strings.TrimSpace(c.PostForm("group_id")),
		Name:          strings.TrimSpace(c.PostForm("name")),
		AssetType:     strings.TrimSpace(c.PostForm("type")),
		FileName:      fileHeader.Filename,
		ContentType:   detectedType,
		Size:          fileHeader.Size,
		Body:          file,
	})
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": asset})
}

func normalizeUploadedAssetContentType(detectedRaw string, declaredRaw string) string {
	detected := strings.ToLower(strings.TrimSpace(strings.SplitN(detectedRaw, ";", 2)[0]))
	declared := strings.ToLower(strings.TrimSpace(strings.SplitN(declaredRaw, ";", 2)[0]))
	switch declared {
	case "audio/mp3":
		declared = "audio/mpeg"
	case "audio/x-m4a":
		declared = "audio/mp4"
	case "audio/wave":
		declared = "audio/wav"
	}
	switch detected {
	case "audio/wave":
		return "audio/wav"
	case "video/mp4":
		if declared == "audio/mp4" {
			return "audio/mp4"
		}
	case "application/ogg":
		if declared == "audio/ogg" {
			return "audio/ogg"
		}
	case "application/octet-stream":
		if strings.HasPrefix(declared, "image/") || strings.HasPrefix(declared, "video/") || strings.HasPrefix(declared, "audio/") {
			return declared
		}
	}
	return detected
}

func GetMediaAssetPreview(c *gin.Context) {
	targetOwner, _ := strconv.Atoi(c.Query("owner_user_id"))
	value, expiresAt, err := assetService.PreviewURLVariant(c.Request.Context(), c.Param("id"), c.GetInt("id"), c.GetInt("role") >= common.RoleAdminUser, c.DefaultQuery("variant", "original"), targetOwner)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"url": value, "expires_at": expiresAt}})
}

func DeleteMediaAsset(c *gin.Context) {
	targetOwner, _ := strconv.Atoi(c.Query("owner_user_id"))
	err := assetService.DeleteAsset(c.Param("id"), c.GetInt("id"), c.GetInt("role") >= common.RoleAdminUser, targetOwner)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ListAssetChannelConfigs(c *gin.Context) {
	configs, err := assetService.ListChannelConfigs()
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": configs})
}

func ListAssetLibraryConnections(c *gin.Context) {
	libraries, err := assetService.ListLibraryConnections()
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": libraries})
}

func DeleteAssetLibraryConnection(c *gin.Context) {
	libraryID, err := strconv.Atoi(c.Param("library_id"))
	if err != nil || libraryID <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library ID")})
		return
	}
	if err := assetService.DeleteLibraryConnection(libraryID); err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.connection_delete_requested", map[string]interface{}{"library_id": libraryID})
	c.JSON(http.StatusAccepted, gin.H{"success": true})
}

func SaveAssetLibraryConnection(c *gin.Context) {
	libraryID := 0
	if c.Param("library_id") != "" {
		var err error
		libraryID, err = strconv.Atoi(c.Param("library_id"))
		if err != nil || libraryID <= 0 {
			assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library ID")})
			return
		}
	}
	var input assetService.LibraryConfigInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library configuration")})
		return
	}
	library, err := assetService.SaveLibraryConnection(libraryID, input)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.connection_save", map[string]interface{}{"library_id": library.ID, "enabled": library.Enabled})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": library})
}

// MergeAssetLibraries folds accounts that share exactly the same upstream
// account. It runs only when an administrator asks for it.
func MergeAssetLibraries(c *gin.Context) {
	report, err := assetService.MergeIdenticalLibraries()
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.merge", map[string]interface{}{"merged_groups": report.MergedGroups, "deferred_groups": report.DeferredGroups})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": report})
}

func TestAssetLibraryConnection(c *gin.Context) {
	libraryID := 0
	if c.Param("library_id") != "" {
		var err error
		libraryID, err = strconv.Atoi(c.Param("library_id"))
		if err != nil || libraryID <= 0 {
			assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library ID")})
			return
		}
	}
	var input *assetService.LibraryConfigInput
	if c.Request.ContentLength != 0 {
		input = &assetService.LibraryConfigInput{}
		if err := common.DecodeJson(c.Request.Body, input); err != nil {
			assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library test request")})
			return
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	if err := assetService.TestLibraryConnection(ctx, libraryID, input); err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func SyncAssetLibraryConnection(c *gin.Context) {
	libraryID, err := strconv.Atoi(c.Param("library_id"))
	if err != nil || libraryID <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library ID")})
		return
	}
	if err := assetService.QueueLibrarySync(libraryID); err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.connection_sync_missing", map[string]interface{}{"library_id": libraryID})
	c.JSON(http.StatusAccepted, gin.H{"success": true})
}

func ClearAssetLibraryFailures(c *gin.Context) {
	libraryID, err := strconv.Atoi(c.Param("library_id"))
	if err != nil || libraryID <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library ID")})
		return
	}
	result, err := assetService.ClearLibraryFailures(libraryID, c.GetInt("id"))
	if result != nil && result.DeletedAssets+result.DismissedReplicas > 0 {
		recordManageAudit(c, "asset_library.connection_clear_failed", map[string]interface{}{
			"library_id": libraryID, "deleted_assets": result.DeletedAssets, "dismissed_replicas": result.DismissedReplicas,
		})
	}
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func RebuildAssetLibraryConnection(c *gin.Context) {
	libraryID, err := strconv.Atoi(c.Param("library_id"))
	if err != nil || libraryID <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library ID")})
		return
	}
	if err := assetService.RebuildLibrarySync(libraryID); err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.connection_rebuild", map[string]interface{}{"library_id": libraryID})
	c.JSON(http.StatusAccepted, gin.H{"success": true})
}

func UpdateAssetChannelConfig(c *gin.Context) {
	channelID, ok := assetChannelID(c)
	if !ok {
		return
	}
	var input assetService.ChannelConfigInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset channel configuration")})
		return
	}
	config, err := assetService.SaveChannelConfig(channelID, input)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.channel_config_update", map[string]interface{}{
		"channel_id": channelID,
		"enabled":    config.Enabled,
		"protocol":   config.Protocol,
		"auth_type":  config.AuthType,
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": config})
}

func TestAssetChannelConfig(c *gin.Context) {
	channelID, ok := assetChannelID(c)
	if !ok {
		return
	}
	var input *assetService.ChannelConfigInput
	if c.Request.ContentLength != 0 {
		var value assetService.ChannelConfigInput
		if err := common.DecodeJson(c.Request.Body, &value); err != nil {
			assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset channel test request")})
			return
		}
		input = &value
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	if err := assetService.TestChannelConfig(ctx, channelID, input); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadGateway, Err: err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func SyncAssetChannel(c *gin.Context) {
	channelID, ok := assetChannelID(c)
	if !ok {
		return
	}
	if err := assetService.QueueChannelSync(channelID); err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"success": true})
}

func ListAssetSyncJobs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	channelID, err := strconv.Atoi(c.DefaultQuery("channel_id", "0"))
	if err != nil || channelID < 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid channel ID")})
		return
	}
	libraryID, err := strconv.Atoi(c.DefaultQuery("library_id", "0"))
	if err != nil || libraryID < 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library ID")})
		return
	}
	result, err := assetService.ListSyncJobs(pageInfo.GetPage(), pageInfo.GetPageSize(), channelID, c.Query("status"), assetService.SyncJobFilter{LibraryID: libraryID})
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func RetryAssetSyncJob(c *gin.Context) {
	markAuditLogged(c)
	replicaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || replicaID <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid synchronization job ID")})
		return
	}
	if err = assetService.RetrySyncJob(replicaID); err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"success": true})
}

func GetAssetLibraryStoragePolicy(c *gin.Context) {
	policy, err := storageService.GetAssetLibraryPolicy()
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": policy})
}

func UpdateAssetLibraryStoragePolicy(c *gin.Context) {
	var input storageService.PolicyInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset library storage policy")})
		return
	}
	policy, err := storageService.SaveAssetLibraryPolicy(input)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.storage_policy_update", map[string]interface{}{"profile_id": policy.StorageProfileID, "enabled": policy.Enabled})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": policy})
}

func assetChannelID(c *gin.Context) (int, bool) {
	channelID, err := strconv.Atoi(c.Param("channel_id"))
	if err != nil || channelID <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid channel ID")})
		return 0, false
	}
	return channelID, true
}

func assetLibraryError(c *gin.Context, err error) {
	if errors.Is(err, model.ErrAssetCopyConflict) || errors.Is(err, model.ErrAssetAmbiguousPublicID) {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": err.Error()})
		return
	}
	statusCode := http.StatusInternalServerError
	if errors.Is(err, model.ErrAssetCopyPermissionDenied) {
		statusCode = http.StatusForbidden
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		statusCode = http.StatusNotFound
	}
	var statusError interface{ HTTPStatusCode() int }
	if errors.As(err, &statusError) {
		statusCode = statusError.HTTPStatusCode()
	}
	if statusCode >= http.StatusInternalServerError {
		logger.LogError(c.Request.Context(), fmt.Sprintf("asset library request failed method=%s path=%s status=%d error=%q", c.Request.Method, c.Request.URL.Path, statusCode, err.Error()))
	}
	c.JSON(statusCode, gin.H{"success": false, "message": err.Error()})
}
