package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/console_setting"

	"github.com/gin-gonic/gin"
)

type acknowledgeLoginNoticeRequest struct {
	DeviceFingerprint string `json:"device_fingerprint"`
}

func GetLoginNotice(c *gin.Context) {
	identity, ok := middleware.GetSessionAuthIdentity(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "dashboard session required"})
		return
	}

	stats, err := service.GetLoginNoticeStats(identity.UserID, time.Now())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	acknowledged, err := model.HasLoginNoticeAcknowledgement(identity.UserID, identity.SessionID)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	announcements := []map[string]interface{}{}
	if console_setting.GetConsoleSetting().AnnouncementsEnabled {
		announcements = console_setting.GetAnnouncements()
	}
	c.Header("Cache-Control", "private, no-store")
	common.ApiSuccess(c, gin.H{
		"announcements":            announcements,
		"statistics":               stats,
		"requires_acknowledgement": stats.SevenDays.Violations > 0,
		"acknowledged":             acknowledged,
	})
}

func AcknowledgeLoginNotice(c *gin.Context) {
	identity, ok := middleware.GetSessionAuthIdentity(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "dashboard session required"})
		return
	}

	var request acknowledgeLoginNoticeRequest
	if err := common.DecodeJson(c.Request.Body, &request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid request"})
		return
	}
	fingerprint := strings.ToLower(strings.TrimSpace(request.DeviceFingerprint))
	decoded, err := hex.DecodeString(fingerprint)
	if err != nil || len(decoded) != sha256.Size {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid device fingerprint"})
		return
	}

	now := time.Now()
	stats, err := service.GetLoginNoticeStats(identity.UserID, now)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	announcements := []map[string]interface{}{}
	if console_setting.GetConsoleSetting().AnnouncementsEnabled {
		announcements = console_setting.GetAnnouncements()
	}
	announcementJSON, err := common.Marshal(announcements)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	announcementHash := sha256.Sum256(announcementJSON)

	acknowledgement := &model.LoginNoticeAcknowledgement{
		UserID:                  identity.UserID,
		SessionID:               identity.SessionID,
		DeviceFingerprint:       fingerprint,
		IP:                      truncateLoginNoticeMetadata(c.ClientIP(), 64),
		UserAgent:               truncateLoginNoticeMetadata(c.Request.UserAgent(), 512),
		AnnouncementFingerprint: hex.EncodeToString(announcementHash[:]),
		RequiredAcknowledgement: stats.SevenDays.Violations > 0,
		GeneratedToday:          stats.Today.Generated,
		GeneratedSevenDays:      stats.SevenDays.Generated,
		GeneratedThirtyDays:     stats.ThirtyDays.Generated,
		ViolationsToday:         stats.Today.Violations,
		ViolationsSevenDays:     stats.SevenDays.Violations,
		ViolationsThirtyDays:    stats.ThirtyDays.Violations,
		AcknowledgedAt:          now.Unix(),
	}
	if err := model.SaveLoginNoticeAcknowledgement(acknowledgement); err != nil {
		common.ApiError(c, err)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	common.ApiSuccess(c, gin.H{"acknowledged_at": acknowledgement.AcknowledgedAt})
}

func truncateLoginNoticeMetadata(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
