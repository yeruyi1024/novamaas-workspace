package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func GetAllLogs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	username := c.Query("username")
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))
	group := c.Query("group")
	requestId := c.Query("request_id")
	upstreamRequestId := c.Query("upstream_request_id")
	logs, total, err := model.GetAllLogs(logType, startTimestamp, endTimestamp, modelName, username, tokenName, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), channel, group, requestId, upstreamRequestId, canViewFinancialAccounting(c))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
	return
}

func GetUserLogs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	userId := c.GetInt("id")
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	group := c.Query("group")
	requestId := c.Query("request_id")
	upstreamRequestId := c.Query("upstream_request_id")
	logs, total, err := model.GetUserLogs(userId, logType, startTimestamp, endTimestamp, modelName, tokenName, pageInfo.GetStartIdx(), pageInfo.GetPageSize(), group, requestId, upstreamRequestId, canViewFinancialAccounting(c))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(logs)
	common.ApiSuccess(c, pageInfo)
	return
}

// Deprecated: SearchAllLogs 已废弃，前端未使用该接口。
func SearchAllLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "该接口已废弃",
	})
}

// Deprecated: SearchUserLogs 已废弃，前端未使用该接口。
func SearchUserLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "该接口已废弃",
	})
}

func GetLogByKey(c *gin.Context) {
	tokenId := c.GetInt("token_id")
	if tokenId == 0 {
		c.JSON(200, gin.H{
			"success": false,
			"message": "无效的令牌",
		})
		return
	}
	logs, err := model.GetLogByTokenId(tokenId)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "",
		"data":    logs,
	})
}

func GetLogsStat(c *gin.Context) {
	filter := costAccountingFilterFromQuery(c)
	stat, err := model.SumLogStatistics(filter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	data := gin.H{
		"quota":    stat.Quota,
		"records":  stat.Records,
		"requests": stat.Requests,
		"rpm":      stat.Rpm,
		"tpm":      stat.Tpm,
	}
	if canViewFinancialAccounting(c) {
		accounting, accountingErr := model.SumCostAccounting(filter)
		if accountingErr != nil {
			common.ApiError(c, accountingErr)
			return
		}
		accounting = model.ReconcileCostAccountingTotals(stat, accounting)
		data["revenue_quota"] = accounting.RevenueQuota
		data["cost_quota"] = accounting.CostQuota
		data["profit_quota"] = accounting.ProfitQuota
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    data,
	})
	return
}

func GetLogsSelfStat(c *gin.Context) {
	filter := costAccountingFilterFromQuery(c)
	filter.UserID = c.GetInt("id")
	filter.Username = ""
	statistics, err := model.SumLogStatistics(filter)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	data := gin.H{
		"quota":    statistics.Quota,
		"records":  statistics.Records,
		"requests": statistics.Requests,
		"rpm":      statistics.Rpm,
		"tpm":      statistics.Tpm,
	}
	if canViewFinancialAccounting(c) {
		accounting, accountingErr := model.SumCostAccounting(filter)
		if accountingErr != nil {
			common.ApiError(c, accountingErr)
			return
		}
		accounting = model.ReconcileCostAccountingTotals(statistics, accounting)
		data["revenue_quota"] = accounting.RevenueQuota
		data["cost_quota"] = accounting.CostQuota
		data["profit_quota"] = accounting.ProfitQuota
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "",
		"data":    data,
	})
	return
}
