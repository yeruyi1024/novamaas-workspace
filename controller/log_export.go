package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

// maxLogExportDays is the longest time range one log export may cover.
const maxLogExportDays = 366

// maxLogExportRows bounds one export: the whole file is built in memory, so an
// unbounded one-year export of every user would exhaust it.
var maxLogExportRows = 200000

// logExportSlots bounds the exports one node runs at once, so a handful of
// large exports cannot starve the database of the requests that bill.
var logExportSlots = make(chan struct{}, 2)

// maxLogExportSeconds is maxLogExportDays in seconds.
const maxLogExportSeconds = maxLogExportDays * 24 * 3600

var errLogExportTooLarge = errors.New("导出条数超过上限，请缩小时间范围或增加筛选条件")

// ExportAllLogs streams every user's logs matching the search.
func ExportAllLogs(c *gin.Context) {
	exportLogs(c, true)
}

// ExportUserLogs streams the caller's own logs matching the search.
func ExportUserLogs(c *gin.Context) {
	exportLogs(c, false)
}

// exportLogs answers with a spreadsheet of the matching logs, newest first, or
// with an API error when there is none to send: the file is built before the
// first byte goes out, so a failed export is never a truncated file. Each user
// runs one export at a time. The search and the visibility rules are those of
// the log list.
func exportLogs(c *gin.Context, all bool) {
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	group := c.Query("group")
	requestId := c.Query("request_id")
	upstreamRequestId := c.Query("upstream_request_id")

	// Top-ups, management, system, login, error and billing-correction logs stay
	// on screen and out of the file; the last ones describe the platform's own
	// operations and suppliers.
	if _, listed := service.LogExportTypeNames[logType]; logType != model.LogTypeUnknown && !listed {
		common.ApiError(c, errors.New("该类型的日志不支持导出，请选择消费或退款"))
		return
	}
	if endTimestamp == 0 {
		endTimestamp = time.Now().Unix()
	}
	if startTimestamp == 0 {
		startTimestamp = endTimestamp - maxLogExportSeconds
	}
	if endTimestamp < startTimestamp {
		common.ApiError(c, errors.New("结束时间不能早于开始时间"))
		return
	}
	if endTimestamp-startTimestamp > maxLogExportSeconds {
		common.ApiError(c, errors.New("导出时间跨度不能超过 "+strconv.Itoa(maxLogExportDays)+" 天"))
		return
	}

	select {
	case logExportSlots <- struct{}{}:
		defer func() { <-logExportSlots }()
	default:
		common.ApiError(c, errors.New("当前导出任务较多，请稍后再试"))
		return
	}

	userId := c.GetInt("id")
	release, ok, err := service.AcquireLogExportLock(userId)
	if err != nil {
		common.SysError("failed to lock log export: " + err.Error())
		common.ApiError(c, errors.New("导出日志失败"))
		return
	}
	if !ok {
		common.ApiError(c, errors.New("已有一个日志导出正在进行，请等待完成后再试"))
		return
	}
	defer release()

	workbook, err := service.NewLogExportWorkbook(all)
	if err != nil {
		common.SysError("failed to start log export workbook: " + err.Error())
		common.ApiError(c, errors.New("导出日志失败"))
		return
	}
	defer workbook.Close()

	count := 0
	writeLog := func(log *model.Log) error {
		if _, listed := service.LogExportTypeNames[log.Type]; !listed {
			return nil
		}
		if count >= maxLogExportRows {
			return errLogExportTooLarge
		}
		count++
		workbook.Add(log)
		return nil
	}
	ctx := c.Request.Context()
	if all {
		username := c.Query("username")
		channel, _ := strconv.Atoi(c.Query("channel"))
		err = model.ExportAllLogs(ctx, logType, startTimestamp, endTimestamp, modelName, username, tokenName, channel, group, requestId, upstreamRequestId, writeLog)
	} else {
		err = model.ExportUserLogs(ctx, userId, logType, startTimestamp, endTimestamp, modelName, tokenName, group, requestId, upstreamRequestId, writeLog)
	}
	if ctx.Err() != nil {
		return
	}
	if errors.Is(err, errLogExportTooLarge) {
		common.ApiError(c, err)
		return
	}
	if err != nil {
		common.SysError("failed to export logs: " + err.Error())
		common.ApiError(c, errors.New("导出日志失败"))
		return
	}

	file, err := workbook.Write(startTimestamp, endTimestamp)
	if err != nil {
		common.SysError("failed to build log export: " + err.Error())
		common.ApiError(c, errors.New("导出日志失败"))
		return
	}
	filename := "usage-logs-" + time.Unix(startTimestamp, 0).Format("20060102") + "-" + time.Unix(endTimestamp, 0).Format("20060102") + ".xlsx"
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Header("X-Log-Export-Count", strconv.Itoa(count))
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", file.Bytes())
}
