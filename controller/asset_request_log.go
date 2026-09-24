package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/model"
	assetService "github.com/QuantumNous/new-api/service/assetlibrary"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListAssetRequestLogs(c *gin.Context) {
	now := time.Now().UnixMilli()
	filter := model.AssetRequestLogFilter{
		StartMS:   now - int64((24*time.Hour)/time.Millisecond),
		EndMS:     now,
		Limit:     50,
		RequestID: strings.TrimSpace(c.Query("request_id")),
		Source:    c.Query("source"),
		Result:    c.Query("result"),
	}
	badRequest := func() {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid request log filters")})
	}
	if value := c.Query("channel_id"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			badRequest()
			return
		}
		filter.ChannelID = parsed
	}
	if value := c.Query("replica_id"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed < 0 {
			badRequest()
			return
		}
		filter.ReplicaID = parsed
	}
	if value := c.Query("start_ms"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed <= 0 {
			badRequest()
			return
		}
		filter.StartMS = parsed
	}
	if value := c.Query("end_ms"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed <= 0 {
			badRequest()
			return
		}
		filter.EndMS = parsed
	}
	if value := c.Query("page_size"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			badRequest()
			return
		}
		filter.Limit = parsed
	}
	if filter.StartMS > filter.EndMS || filter.EndMS-filter.StartMS > int64((7*24*time.Hour)/time.Millisecond) ||
		len(filter.RequestID) > 64 ||
		(filter.Source != "" && filter.Source != "test" && filter.Source != "sync") ||
		(filter.Result != "" && filter.Result != "success" && filter.Result != "failure") {
		badRequest()
		return
	}
	var err error
	filter.CursorMS, filter.CursorID, err = model.ParseAssetRequestLogCursor(c.Query("cursor"))
	if err != nil {
		badRequest()
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	items, nextCursor, err := model.ListAssetRequestLogs(ctx, filter)
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"items": items, "next_cursor": nextCursor,
		"dropped_on_this_node": assetService.DroppedAssetRequestLogs(),
	}})
}

func GetAssetRequestLogDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid request log ID")})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	entry, err := model.GetAssetRequestLogWithDetail(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		assetLibraryError(c, &assetService.RequestError{StatusCode: http.StatusNotFound, Err: errors.New("request log not found")})
		return
	}
	if err != nil {
		assetLibraryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": entry})
}
