package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"
	"github.com/gin-gonic/gin"
)

func CopyAssetGroups(c *gin.Context) {
	var input assetService.AssetCopyInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid asset copy request")})
		return
	}
	if !requireAssetCopyUsers(c, input.SourceUserID, input.TargetUserID) {
		return
	}
	result, err := assetService.CopyAssetGroups(c.GetInt("id"), input)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.copy", map[string]interface{}{"batch_id": result.ID, "source_user_id": result.SourceUserID, "target_user_id": result.TargetUserID, "asset_count": result.AssetCount})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func GetAssetCopy(c *gin.Context) {
	result, err := assetService.GetAssetCopy(c.Param("id"))
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	if !requireAssetCopyUsers(c, result.SourceUserID, result.TargetUserID) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func DeleteAssetCopySources(c *gin.Context) {
	result, err := assetService.GetAssetCopy(c.Param("id"))
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	if !requireAssetCopyUsers(c, result.SourceUserID, result.TargetUserID) {
		return
	}
	result, err = assetService.DeleteCopiedSources(c.Param("id"), c.GetInt("id"))
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.copy_sources_delete", map[string]interface{}{"batch_id": result.ID, "source_user_id": result.SourceUserID, "target_user_id": result.TargetUserID, "asset_count": result.AssetCount})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func PurgeUserAssets(c *gin.Context) {
	targetUserID, err := strconv.Atoi(c.Param("id"))
	if err != nil || targetUserID <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid user ID")})
		return
	}
	if c.GetInt("role") != common.RoleRootUser {
		var user model.User
		if err := model.DB.Select("id", "role").First(&user, targetUserID).Error; err != nil {
			assetLibraryError(c, err)
			return
		}
		if !canManageTargetRole(c.GetInt("role"), user.Role) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "cannot manage assets of users with an equal or higher role"})
			return
		}
	}
	result, err := assetService.PurgeUserAssets(targetUserID, c.GetInt("id"))
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	recordManageAudit(c, "asset_library.user_purge", map[string]interface{}{
		"owner_user_id":       result.OwnerUserID,
		"deleted_asset_count": result.DeletedAssetCount,
		"deleted_group_count": result.DeletedGroupCount,
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// Apply the same target-role boundary as user management to both sides of a
// copy, including later receipt reads and source cleanup.
func requireAssetCopyUsers(c *gin.Context, sourceID, targetID int) bool {
	if c.GetInt("role") == common.RoleRootUser {
		return true
	}
	if sourceID <= 0 || targetID <= 0 || sourceID == targetID {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "select two different users"})
		return false
	}
	var users []model.User
	if err := model.DB.Select("id", "role").Where("id IN ?", []int{sourceID, targetID}).Find(&users).Error; err != nil {
		assetLibraryError(c, err)
		return false
	}
	if len(users) != 2 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "source or target user does not exist"})
		return false
	}
	for _, user := range users {
		if !canManageTargetRole(c.GetInt("role"), user.Role) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "cannot manage assets of users with an equal or higher role"})
			return false
		}
	}
	return true
}
