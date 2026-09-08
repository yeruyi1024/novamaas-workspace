package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	storageService "github.com/QuantumNous/new-api/service/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListStorageProfiles(c *gin.Context) {
	profiles, err := storageService.ListProfiles()
	if err != nil {
		storageAPIError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": profiles})
}

func GetStorageProfile(c *gin.Context) {
	id, ok := storageProfileID(c)
	if !ok {
		return
	}
	profile, err := storageService.GetProfile(id)
	if err != nil {
		storageAPIError(c, http.StatusInternalServerError, err)
		return
	}
	if profile == nil {
		storageAPIError(c, http.StatusNotFound, errors.New("storage profile not found"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": profile})
}

func CreateStorageProfile(c *gin.Context) {
	var input storageService.ProfileInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		storageAPIError(c, http.StatusBadRequest, errors.New("invalid storage profile request"))
		return
	}
	profile, err := storageService.SaveProfile(0, input)
	if err != nil {
		storageAPIError(c, http.StatusBadRequest, err)
		return
	}
	recordManageAudit(c, "storage.profile_create", map[string]interface{}{"id": profile.ID, "name": profile.Name, "type": profile.ProviderType})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": profile})
}

func UpdateStorageProfile(c *gin.Context) {
	id, ok := storageProfileID(c)
	if !ok {
		return
	}
	var input storageService.ProfileInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		storageAPIError(c, http.StatusBadRequest, errors.New("invalid storage profile request"))
		return
	}
	profile, err := storageService.SaveProfile(id, input)
	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, gorm.ErrRecordNotFound) {
			statusCode = http.StatusNotFound
		}
		storageAPIError(c, statusCode, err)
		return
	}
	recordManageAudit(c, "storage.profile_update", map[string]interface{}{"id": profile.ID, "name": profile.Name})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": profile})
}

func DeleteStorageProfile(c *gin.Context) {
	id, ok := storageProfileID(c)
	if !ok {
		return
	}
	if err := storageService.ArchiveProfile(id); err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, gorm.ErrRecordNotFound) {
			statusCode = http.StatusNotFound
		}
		storageAPIError(c, statusCode, err)
		return
	}
	recordManageAudit(c, "storage.profile_delete", map[string]interface{}{"id": id})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func TestStorageProfileInput(c *gin.Context) {
	var input storageService.ProfileInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		storageAPIError(c, http.StatusBadRequest, errors.New("invalid storage profile request"))
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	if err := storageService.TestProfile(ctx, input); err != nil {
		storageAPIError(c, http.StatusBadGateway, err)
		return
	}
	recordManageAudit(c, "storage.profile_test", map[string]interface{}{"id": "unsaved"})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func TestSavedStorageProfile(c *gin.Context) {
	id, ok := storageProfileID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	if err := storageService.TestSavedProfile(ctx, id); err != nil {
		statusCode := http.StatusBadGateway
		if errors.Is(err, gorm.ErrRecordNotFound) {
			statusCode = http.StatusNotFound
		}
		storageAPIError(c, statusCode, err)
		return
	}
	recordManageAudit(c, "storage.profile_test", map[string]interface{}{"id": id})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetRelayMediaStoragePolicy(c *gin.Context) {
	policy, err := storageService.GetRelayMediaPolicy()
	if err != nil {
		storageAPIError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": policy})
}

func UpdateRelayMediaStoragePolicy(c *gin.Context) {
	var input storageService.PolicyInput
	if err := common.DecodeJson(c.Request.Body, &input); err != nil {
		storageAPIError(c, http.StatusBadRequest, errors.New("invalid storage policy request"))
		return
	}
	policy, err := storageService.SaveRelayMediaPolicy(input)
	if err != nil {
		storageAPIError(c, http.StatusBadRequest, err)
		return
	}
	recordManageAudit(c, "storage.policy_update", map[string]interface{}{"key": policy.Key, "profile_id": policy.StorageProfileID, "enabled": policy.Enabled})
	c.JSON(http.StatusOK, gin.H{"success": true, "data": policy})
}

func storageProfileID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		storageAPIError(c, http.StatusBadRequest, errors.New("invalid storage profile ID"))
		return 0, false
	}
	return id, true
}

func storageAPIError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"message": err.Error(),
	})
}
