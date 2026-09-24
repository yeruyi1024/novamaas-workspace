package controller

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxVolcAssetActionBodyBytes = 1024 * 1024

type volcAssetFilter struct {
	GroupIDs []string `json:"GroupIds"`
	Statuses []string `json:"Statuses"`
	Name     string   `json:"Name"`
}

type volcAssetActionRequest struct {
	ID          string          `json:"Id"`
	Name        string          `json:"Name"`
	Description string          `json:"Description"`
	GroupType   string          `json:"GroupType"`
	GroupID     string          `json:"GroupId"`
	AssetType   string          `json:"AssetType"`
	URL         string          `json:"URL"`
	ProjectName string          `json:"ProjectName"`
	Filter      volcAssetFilter `json:"Filter"`
	PageNumber  int             `json:"PageNumber"`
	PageSize    int             `json:"PageSize"`
	MaxResults  int             `json:"MaxResults"`
	NextToken   string          `json:"NextToken"`
	SortBy      string          `json:"SortBy"`
	SortOrder   string          `json:"SortOrder"`
}

func ListAssetAccessKeys(c *gin.Context) {
	keys, err := assetService.ListAccessKeys(c.GetInt("id"))
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": keys})
}

func CreateAssetAccessKey(c *gin.Context) {
	var input struct {
		Name string `json:"name"`
	}
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid access key request")})
		return
	}
	key, err := assetService.CreateAccessKey(c.GetInt("id"), input.Name)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": key})
}

func DeleteAssetAccessKey(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid access key ID")})
		return
	}
	if err = assetService.DeleteAccessKey(c.GetInt("id"), id); err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// HandleVolcAssetAction exposes the same Action API shape used by the Fire Ark
// asset library. It intentionally owns only tenant-local assets; upstream
// replicas stay private and are resolved later when a generation is routed.
func HandleVolcAssetAction(c *gin.Context) {
	action := strings.TrimSpace(c.Query("Action"))
	version := strings.TrimSpace(c.Query("Version"))
	if action == "" || version != assetService.DefaultAPIVersion {
		volcAssetError(c, http.StatusBadRequest, action, version, "InvalidParameter", "Action and Version=2024-01-01 are required")
		return
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxVolcAssetActionBodyBytes+1))
	if err != nil || len(body) > maxVolcAssetActionBodyBytes {
		volcAssetError(c, http.StatusRequestEntityTooLarge, action, version, "InvalidRequest", "request body exceeds the allowed size")
		return
	}
	principal, err := assetService.AuthenticateVolcActionRequestPrincipal(c.Request, body)
	if err != nil {
		volcAssetError(c, http.StatusUnauthorized, action, version, "AuthFailure", "request authentication failed")
		return
	}
	ownerUserID := principal.OwnerUserID
	var input volcAssetActionRequest
	if len(body) > 0 {
		if err = common.Unmarshal(body, &input); err != nil {
			volcAssetError(c, http.StatusBadRequest, action, version, "InvalidRequest", "request body must be valid JSON")
			return
		}
	}

	var result any
	switch action {
	case "CreateAssetGroup":
		if input.GroupType != "" && !strings.EqualFold(input.GroupType, "AIGC") {
			err = &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("GroupType must be AIGC")}
			break
		}
		var group *model.AssetGroup
		group, err = assetService.CreateGroup(ownerUserID, assetService.GroupInput{Name: input.Name, Description: input.Description})
		if err == nil {
			result = gin.H{"Id": group.PublicID}
		}
	case "ListAssetGroups":
		result, err = listVolcAssetGroups(ownerUserID, input)
	case "GetAssetGroup":
		var group *model.AssetGroup
		group, err = assetService.GetGroup(input.ID, ownerUserID)
		if err == nil {
			result = volcAssetGroup(group)
		}
	case "UpdateAssetGroup":
		var group *model.AssetGroup
		group, err = assetService.UpdateGroup(input.ID, ownerUserID, false, assetService.GroupInput{Name: input.Name, Description: input.Description})
		if err == nil {
			result = gin.H{"Id": group.PublicID}
		}
	case "DeleteAssetGroup":
		err = assetService.DeleteGroup(input.ID, ownerUserID, false)
		result = gin.H{}
	case "CreateAsset":
		var asset *assetService.AssetView
		asset, err = assetService.CreateAssetFromURL(c.Request.Context(), ownerUserID, input.GroupID, input.Name, input.AssetType, input.URL)
		if err == nil {
			result = gin.H{"Id": asset.ID}
			recordVolcAssetUploadAudit(c, principal, asset)
		}
	case "ListAssets":
		result, err = listVolcAssets(c, ownerUserID, input)
	case "GetAsset":
		var asset *assetService.AssetView
		asset, err = assetService.GetAsset(input.ID, ownerUserID)
		if err == nil {
			result, err = volcAsset(c, asset)
		}
	case "UpdateAsset":
		var asset *assetService.AssetView
		asset, err = assetService.UpdateAsset(input.ID, ownerUserID, input.Name)
		if err == nil {
			result = gin.H{"Id": asset.ID}
		}
	case "DeleteAsset":
		err = assetService.DeleteAsset(input.ID, ownerUserID, false)
		result = gin.H{}
	default:
		volcAssetError(c, http.StatusBadRequest, action, version, "InvalidAction", "unsupported asset Action")
		return
	}
	if err != nil {
		statusCode, code := volcAssetErrorStatus(err)
		volcAssetError(c, statusCode, action, version, code, err.Error())
		return
	}
	volcAssetSuccess(c, action, version, result)
}

func recordVolcAssetUploadAudit(c *gin.Context, principal *assetService.VolcActionPrincipal, asset *assetService.AssetView) {
	params := map[string]interface{}{
		"id":            asset.ID,
		"name":          asset.Name,
		"groupId":       asset.GroupID,
		"assetType":     asset.Type,
		"accessKeyId":   principal.AccessKeyID,
		"accessKeyName": principal.AccessKeyName,
	}
	auditInfo := map[string]interface{}{
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"status":      http.StatusOK,
		"success":     true,
		"auth_method": "asset_access_key",
	}
	model.RecordOperationAuditLog(
		principal.OwnerUserID,
		auditContentEN("asset.upload_aksk", params),
		c.ClientIP(),
		"asset.upload_aksk",
		params,
		nil,
		auditInfo,
	)
}

func listVolcAssetGroups(ownerUserID int, input volcAssetActionRequest) (gin.H, error) {
	page, pageSize := volcAssetPage(input)
	groups, err := assetService.ListGroupsPage(assetService.GroupListInput{
		OwnerUserID: ownerUserID, GroupIDs: input.Filter.GroupIDs, Search: input.Filter.Name,
		Page: page, PageSize: pageSize, SortBy: input.SortBy, SortOrder: input.SortOrder,
	})
	if err != nil {
		return nil, err
	}
	items := make([]gin.H, 0, len(groups.Items))
	for index := range groups.Items {
		items = append(items, volcAssetGroup(&groups.Items[index]))
	}
	result := gin.H{"Items": items, "TotalCount": groups.Total, "PageNumber": groups.Page, "PageSize": groups.PageSize}
	attachVolcNextToken(result, input, groups.Page, groups.PageSize, groups.Total)
	return result, nil
}

func listVolcAssets(c *gin.Context, ownerUserID int, input volcAssetActionRequest) (gin.H, error) {
	page, pageSize := volcAssetPage(input)
	statuses := make([]string, 0, 2)
	if len(input.Filter.Statuses) > 0 {
		for _, status := range input.Filter.Statuses {
			if strings.EqualFold(status, "Active") {
				statuses = append(statuses, model.AssetStatusReady)
			} else if strings.EqualFold(status, "Failed") || strings.EqualFold(status, "Rejected") {
				statuses = append(statuses, model.AssetStatusUnavailable)
			}
		}
		if len(statuses) == 0 {
			return gin.H{"Items": []any{}, "TotalCount": 0, "PageNumber": page, "PageSize": pageSize}, nil
		}
	}
	assets, err := assetService.ListAssets(assetService.AssetListInput{
		OwnerUserID: ownerUserID, GroupPublicIDs: input.Filter.GroupIDs, Statuses: statuses, Search: input.Filter.Name,
		Page: page, PageSize: pageSize, SortBy: input.SortBy, SortOrder: input.SortOrder,
	})
	if err != nil {
		return nil, err
	}
	items := make([]gin.H, 0, len(assets.Items))
	for index := range assets.Items {
		item, itemErr := volcAsset(c, &assets.Items[index])
		if itemErr != nil {
			return nil, itemErr
		}
		items = append(items, item)
	}
	result := gin.H{"Items": items, "TotalCount": assets.Total, "PageNumber": assets.Page, "PageSize": assets.PageSize}
	attachVolcNextToken(result, input, assets.Page, assets.PageSize, assets.Total)
	return result, nil
}

func volcAssetPage(input volcAssetActionRequest) (int, int) {
	page := input.PageNumber
	if page < 1 {
		page = 1
	}
	if input.NextToken != "" {
		if decoded, err := base64.RawURLEncoding.DecodeString(input.NextToken); err == nil {
			if nextPage, parseErr := strconv.Atoi(string(decoded)); parseErr == nil && nextPage > 0 {
				page = nextPage
			}
		}
	}
	pageSize := input.PageSize
	if input.MaxResults > 0 {
		pageSize = input.MaxResults
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func attachVolcNextToken(result gin.H, input volcAssetActionRequest, page int, pageSize int, total int64) {
	if input.MaxResults > 0 && int64(page*pageSize) < total {
		result["NextToken"] = base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(page + 1)))
	}
}

func volcAssetGroup(group *model.AssetGroup) gin.H {
	return gin.H{
		"Id": group.PublicID, "Name": group.Name, "Description": group.Description, "GroupType": "AIGC",
		"ProjectName": "default", "CreateTime": volcAssetTime(group.CreatedAt), "UpdateTime": volcAssetTime(group.UpdatedAt),
	}
}

func volcAsset(c *gin.Context, asset *assetService.AssetView) (gin.H, error) {
	previewURL, _, err := assetService.PreviewURL(c.Request.Context(), asset.ID, asset.OwnerUserID, false)
	if err != nil {
		return nil, err
	}
	return volcAssetResponse(asset, previewURL), nil
}

func volcAssetResponse(asset *assetService.AssetView, previewURL string) gin.H {
	result := gin.H{
		"Id": asset.ID, "Name": asset.Name, "URL": previewURL, "GroupId": asset.GroupID,
		"AssetType": strings.ToUpper(asset.Type[:1]) + strings.ToLower(asset.Type[1:]), "Status": "Active",
		"Moderation": gin.H{"Strategy": "Default"}, "ProjectName": "default",
		"CreateTime": volcAssetTime(asset.CreatedAt), "UpdateTime": volcAssetTime(asset.UpdatedAt),
	}
	if asset.Status == model.AssetStatusUnavailable {
		result["Status"] = "Failed"
		result["FailureReason"] = asset.UnavailableReason
	}
	return result
}

func volcAssetTime(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
}

func volcAssetSuccess(c *gin.Context, action string, version string, result any) {
	c.JSON(http.StatusOK, gin.H{"ResponseMetadata": volcAssetMetadata(c, action, version), "Result": result})
}

func volcAssetError(c *gin.Context, statusCode int, action string, version string, code string, message string) {
	metadata := volcAssetMetadata(c, action, version)
	metadata["Error"] = gin.H{"Code": code, "Message": message}
	c.JSON(statusCode, gin.H{"ResponseMetadata": metadata})
}

func volcAssetMetadata(c *gin.Context, action string, version string) gin.H {
	requestID := c.GetString(common.RequestIdKey)
	if requestID == "" {
		requestID = common.NewRequestId()
	}
	return gin.H{"RequestId": requestID, "Action": action, "Version": version, "Service": assetService.DefaultService, "Region": assetService.DefaultRegion}
}

func volcAssetErrorStatus(err error) (int, string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound, "ResourceNotFound"
	}
	var statusError interface{ HTTPStatusCode() int }
	if errors.As(err, &statusError) {
		statusCode := statusError.HTTPStatusCode()
		if statusCode >= http.StatusInternalServerError {
			return statusCode, "InternalError"
		}
		return statusCode, "InvalidParameter"
	}
	return http.StatusInternalServerError, "InternalError"
}
