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
	if c.Query("p") != "" || c.Query("page_size") != "" || c.Query("search") != "" {
		pageInfo := common.GetPageQuery(c)
		groups, err := assetService.ListGroupsPage(assetService.GroupListInput{
			OwnerUserID:      c.GetInt("id"),
			IncludeAllOwners: c.GetInt("role") >= common.RoleAdminUser && c.Query("scope") == "all",
			Search:           c.Query("search"), Page: pageInfo.GetPage(), PageSize: pageInfo.GetPageSize(),
			SortBy: "Name", SortOrder: "Asc",
		})
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
		items := make([]assetService.GroupView, 0, len(groups.Items))
		for _, group := range groups.Items {
			items = append(items, assetService.GroupView{
				ID: group.PublicID, OwnerUserID: group.OwnerUserID, OwnerName: ownerNames[group.OwnerUserID],
				Name: group.Name, Description: group.Description, Status: group.Status,
				CreatedAt: group.CreatedAt, UpdatedAt: group.UpdatedAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"items": items, "total": groups.Total, "page": groups.Page, "page_size": groups.PageSize,
		}})
		return
	}
	groups, err := assetService.ListGroups(c.GetInt("id"), c.GetInt("role") >= common.RoleAdminUser && c.Query("scope") == "all")
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
	group, err := assetService.UpdateGroup(c.Param("id"), c.GetInt("id"), c.GetInt("role") >= common.RoleAdminUser, input)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": group})
}

func DeleteAssetGroup(c *gin.Context) {
	err := assetService.DeleteGroup(c.Param("id"), c.GetInt("id"), c.GetInt("role") >= common.RoleAdminUser)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ListMediaAssets(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	assets, err := assetService.ListAssets(assetService.AssetListInput{
		OwnerUserID:      c.GetInt("id"),
		IncludeAllOwners: c.GetInt("role") >= common.RoleAdminUser && c.Query("scope") == "all",
		GroupPublicID:    strings.TrimSpace(c.Query("group_id")),
		Search:           strings.TrimSpace(c.Query("search")),
		Page:             pageInfo.GetPage(),
		PageSize:         pageInfo.GetPageSize(),
	})
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": assets})
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
	detectedType := strings.ToLower(strings.TrimSpace(strings.SplitN(http.DetectContentType(header[:read]), ";", 2)[0]))
	declaredType := strings.ToLower(strings.TrimSpace(strings.SplitN(fileHeader.Header.Get("Content-Type"), ";", 2)[0]))
	if detectedType == "application/octet-stream" && (strings.HasPrefix(declaredType, "image/") || strings.HasPrefix(declaredType, "video/") || strings.HasPrefix(declaredType, "audio/")) {
		detectedType = declaredType
	}
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

func GetMediaAssetPreview(c *gin.Context) {
	value, expiresAt, err := assetService.PreviewURLVariant(c.Request.Context(), c.Param("id"), c.GetInt("id"), c.GetInt("role") >= common.RoleAdminUser, c.DefaultQuery("variant", "original"))
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"url": value, "expires_at": expiresAt}})
}

func DeleteMediaAsset(c *gin.Context) {
	err := assetService.DeleteAsset(c.Param("id"), c.GetInt("id"), c.GetInt("role") >= common.RoleAdminUser)
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
	result, err := assetService.ListSyncJobs(pageInfo.GetPage(), pageInfo.GetPageSize(), channelID, c.Query("status"))
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func RetryAssetSyncJob(c *gin.Context) {
	replicaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || replicaID <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid synchronization job ID")})
		return
	}
	if err = assetService.RetrySyncJob(replicaID); err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.sync_job_retry", map[string]interface{}{"replica_id": replicaID})
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
	statusCode := http.StatusInternalServerError
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
