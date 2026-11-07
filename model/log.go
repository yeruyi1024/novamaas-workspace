package model

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	relaydto "github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"

	"gorm.io/gorm"
)

func applyExplicitLogTextFilter(tx *gorm.DB, column string, value string) (*gorm.DB, error) {
	if value == "" {
		return tx, nil
	}
	if strings.Contains(value, "%") {
		condition, pattern, err := buildLogLikeCondition(column, value)
		if err != nil {
			return nil, err
		}
		return tx.Where(condition, pattern), nil
	}
	return tx.Where(column+" = ?", value), nil
}

func buildLogLikeCondition(column string, value string) (string, string, error) {
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		pattern, err := sanitizeClickHouseLikePattern(value)
		if err != nil {
			return "", "", err
		}
		return column + " LIKE ?", pattern, nil
	}

	pattern, err := sanitizeLikePattern(value)
	if err != nil {
		return "", "", err
	}
	return column + " LIKE ? ESCAPE '!'", pattern, nil
}

func sanitizeClickHouseLikePattern(input string) (string, error) {
	input = strings.ReplaceAll(input, `\`, `\\`)
	input = strings.ReplaceAll(input, `_`, `\_`)

	if err := validateLikePattern(input); err != nil {
		return "", err
	}
	return input, nil
}

type Log struct {
	Id                int    `json:"id" gorm:"index:idx_created_at_id,priority:2;index:idx_user_id_id,priority:2;index:idx_logs_user_created_id,priority:3;index:idx_logs_type_created_id,priority:3"`
	UserId            int    `json:"user_id" gorm:"index;index:idx_user_id_id,priority:1;index:idx_logs_user_created_id,priority:1"`
	CreatedAt         int64  `json:"created_at" gorm:"bigint;index:idx_created_at_id,priority:1;index:idx_created_at_type;index:idx_logs_user_created_id,priority:2;index:idx_logs_type_created_id,priority:2"`
	Type              int    `json:"type" gorm:"index:idx_created_at_type;index:idx_logs_type_created_id,priority:1"`
	Content           string `json:"content"`
	Username          string `json:"username" gorm:"index;index:index_username_model_name,priority:2;default:''"`
	TokenName         string `json:"token_name" gorm:"index;default:''"`
	ModelName         string `json:"model_name" gorm:"index;index:index_username_model_name,priority:1;default:''"`
	Quota             int    `json:"quota" gorm:"default:0"`
	PromptTokens      int    `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens  int    `json:"completion_tokens" gorm:"default:0"`
	UseTime           int    `json:"use_time" gorm:"default:0"`
	IsStream          bool   `json:"is_stream"`
	ChannelId         int    `json:"channel" gorm:"index"`
	ChannelName       string `json:"channel_name" gorm:"->"`
	TokenId           int    `json:"token_id" gorm:"default:0;index"`
	Group             string `json:"group" gorm:"index"`
	Ip                string `json:"ip" gorm:"index;default:''"`
	RequestId         string `json:"request_id,omitempty" gorm:"type:varchar(64);index:idx_logs_request_id;default:''"`
	UpstreamRequestId string `json:"upstream_request_id,omitempty" gorm:"type:varchar(128);index:idx_logs_upstream_request_id;default:''"`
	Other             string `json:"other"`
	RevenueQuota      *int64 `json:"revenue_quota,omitempty" gorm:"-"`
	CostQuota         *int64 `json:"cost_quota,omitempty" gorm:"-"`
	ProfitQuota       *int64 `json:"profit_quota,omitempty" gorm:"-"`
}

// don't use iota, avoid change log type value
const (
	LogTypeUnknown           = 0
	LogTypeTopup             = 1
	LogTypeConsume           = 2
	LogTypeManage            = 3
	LogTypeSystem            = 4
	LogTypeError             = 5
	LogTypeRefund            = 6
	LogTypeLogin             = 7
	LogTypeBillingCorrection = 8
)

func ensureLogRequestId(log *Log) {
	if log != nil && log.RequestId == "" {
		log.RequestId = common.NewRequestId()
	}
}

func createLog(log *Log) error {
	ensureLogRequestId(log)
	return LOG_DB.Create(log).Error
}

func clickHouseLogOrder(prefix string) string {
	return prefix + "created_at desc, " + prefix + "request_id desc"
}

func assignDisplayLogIds(logs []*Log, startIdx int) {
	for i := range logs {
		logs[i].Id = startIdx + i + 1
	}
}

func formatUserLogs(logs []*Log, startIdx int) {
	for i := range logs {
		logs[i].ChannelName = ""
		otherText := logs[i].Other
		containsAdminOnlyData := strings.Contains(otherText, `"admin_info"`) ||
			strings.Contains(otherText, `"request_body"`) ||
			strings.Contains(otherText, `"request_body_available"`) ||
			strings.Contains(otherText, `"request_body_ref"`) ||
			strings.Contains(otherText, `"audit_info"`)
		if !containsAdminOnlyData {
			continue
		}
		var otherMap map[string]interface{}
		otherMap, _ = common.StrToMap(otherText)
		if otherMap != nil {
			// Remove admin-only debug fields.
			delete(otherMap, "admin_info")
			// Raw request bodies are retained for administrator auditing only.
			// Keep non-sensitive transformation markers visible to the owner.
			delete(otherMap, "request_body")
			delete(otherMap, "request_body_available")
			delete(otherMap, "request_body_ref")
			// Remove operation-audit details (operator/route info), admin-only.
			delete(otherMap, "audit_info")
			// delete(otherMap, "reject_reason")
			// delete(otherMap, "stream_status")
		}
		logs[i].Other = common.MapToJsonStr(otherMap)
	}
	assignDisplayLogIds(logs, startIdx)
}

func GetLogByTokenId(tokenId int) (logs []*Log, err error) {
	order := "id desc"
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		order = clickHouseLogOrder("")
	}
	err = LOG_DB.Model(&Log{}).Where("token_id = ? AND type <> ?", tokenId, LogTypeBillingCorrection).Order(order).Limit(common.MaxRecentItems).Find(&logs).Error
	if err == nil {
		var adjustments []billingEffectiveLog
		adjustments, err = billingEffectiveLogsForPage(context.Background(), logs)
		if err == nil {
			err = attachHistoricalTaskLogPricing(logs)
			if err == nil {
				err = projectBillingEffectiveLogs(logs, adjustments)
			}
		}
	}
	formatUserLogs(logs, 0)
	return logs, err
}

func RecordLog(userId int, logType int, content string) {
	if logType == LogTypeConsume && !common.LogConsumeEnabled {
		return
	}
	username, _ := GetUsernameById(userId, false)
	log := &Log{
		UserId:    userId,
		Username:  username,
		CreatedAt: common.GetTimestamp(),
		Type:      logType,
		Content:   content,
	}
	err := createLog(log)
	if err != nil {
		common.SysLog("failed to record log: " + err.Error())
	}
}

// RecordLogWithAdminInfo 记录操作日志，并将管理员相关信息存入 Other.admin_info，
func RecordLogWithAdminInfo(userId int, logType int, content string, adminInfo map[string]interface{}) {
	if logType == LogTypeConsume && !common.LogConsumeEnabled {
		return
	}
	username, _ := GetUsernameById(userId, false)
	log := &Log{
		UserId:    userId,
		Username:  username,
		CreatedAt: common.GetTimestamp(),
		Type:      logType,
		Content:   content,
	}
	if len(adminInfo) > 0 {
		other := map[string]interface{}{
			"admin_info": adminInfo,
		}
		log.Other = common.MapToJsonStr(other)
	}
	if err := createLog(log); err != nil {
		common.SysLog("failed to record log: " + err.Error())
	}
}

// buildOpField 构建语言无关的操作描述（写入 Other.op）。
// 前端依据 action(稳定操作标识) + params(结构化参数) 在渲染期用 i18n 本地化展示，
// 因此不在数据库中存储自然语言句子。
func buildOpField(action string, params map[string]interface{}) map[string]interface{} {
	op := map[string]interface{}{
		"action": action,
	}
	if len(params) > 0 {
		op["params"] = params
	}
	return op
}

// RecordLoginLog 记录用户登录成功的审计日志（type=LogTypeLogin）。
// username 由调用方传入（登录流程已持有用户对象），避免额外的数据库查询。
// content 为英文兜底文本（用于导出）；action+params 供前端本地化渲染。
// extra 可携带 login_method、user_agent 等附加信息（普通用户可见）。
func RecordLoginLog(userId int, username string, content string, ip string, action string, params map[string]interface{}, extra map[string]interface{}) {
	other := map[string]interface{}{}
	for k, v := range extra {
		other[k] = v
	}
	other["op"] = buildOpField(action, params)
	log := &Log{
		UserId:    userId,
		Username:  username,
		CreatedAt: common.GetTimestamp(),
		Type:      LogTypeLogin,
		Content:   content,
		Ip:        ip,
		Other:     common.MapToJsonStr(other),
	}
	if err := createLog(log); err != nil {
		common.SysLog("failed to record login log: " + err.Error())
	}
}

// RecordOperationAuditLog 记录管理/高危操作审计日志（type=LogTypeManage）。
// logUserId 为日志归属者，管理审计日志应归属实际操作者；目标资源/用户放入
// action params。username 内部按 logUserId 查询。content 为英文兜底文本（供导出使用）。
// action+params 写入 Other.op，供前端本地化渲染（普通用户可见，不含敏感信息）。
// adminInfo 存放操作者身份（写入 Other.admin_info，普通用户查询时剥离）；
// auditInfo 存放路由/方法/结果等中间件兜底信息（写入 Other.audit_info，普通用户查询时剥离）。
func RecordOperationAuditLog(logUserId int, content string, ip string, action string, params map[string]interface{}, adminInfo map[string]interface{}, auditInfo map[string]interface{}) {
	username, _ := GetUsernameById(logUserId, false)
	other := map[string]interface{}{
		"op": buildOpField(action, params),
	}
	if len(adminInfo) > 0 {
		other["admin_info"] = adminInfo
	}
	if len(auditInfo) > 0 {
		other["audit_info"] = auditInfo
	}
	log := &Log{
		UserId:    logUserId,
		Username:  username,
		CreatedAt: common.GetTimestamp(),
		Type:      LogTypeManage,
		Content:   content,
		Ip:        ip,
		Other:     common.MapToJsonStr(other),
	}
	if err := createLog(log); err != nil {
		common.SysLog("failed to record operation audit log: " + err.Error())
	}
}

func RecordTopupLog(userId int, content string, callerIp string, paymentMethod string, callbackPaymentMethod string) {
	username, _ := GetUsernameById(userId, false)
	adminInfo := map[string]interface{}{
		"server_ip":               common.GetIp(),
		"node_name":               common.NodeName,
		"caller_ip":               callerIp,
		"payment_method":          paymentMethod,
		"callback_payment_method": callbackPaymentMethod,
		"version":                 common.Version,
	}
	other := map[string]interface{}{
		"admin_info": adminInfo,
	}
	log := &Log{
		UserId:    userId,
		Username:  username,
		CreatedAt: common.GetTimestamp(),
		Type:      LogTypeTopup,
		Content:   content,
		Ip:        callerIp,
		Other:     common.MapToJsonStr(other),
	}
	err := createLog(log)
	if err != nil {
		common.SysLog("failed to record topup log: " + err.Error())
	}
}

func RecordErrorLog(c *gin.Context, userId int, channelId int, modelName string, tokenName string, content string, tokenId int, useTimeSeconds int,
	isStream bool, group string, other map[string]interface{}) {
	logger.LogInfo(c, fmt.Sprintf("record error log: userId=%d, channelId=%d, modelName=%s, tokenName=%s, content=%s", userId, channelId, modelName, tokenName, common.LocalLogPreview(content)))
	username := c.GetString("username")
	requestId := c.GetString(common.RequestIdKey)
	upstreamRequestId := c.GetString(common.UpstreamRequestIdKey)
	otherStr := common.MapToJsonStr(other)
	// 判断是否需要记录 IP
	needRecordIp := false
	if settingMap, err := GetUserSetting(userId, false); err == nil {
		if settingMap.RecordIpLog {
			needRecordIp = true
		}
	}
	log := &Log{
		UserId:           userId,
		Username:         username,
		CreatedAt:        common.GetTimestamp(),
		Type:             LogTypeError,
		Content:          content,
		PromptTokens:     0,
		CompletionTokens: 0,
		TokenName:        tokenName,
		ModelName:        modelName,
		Quota:            0,
		ChannelId:        channelId,
		TokenId:          tokenId,
		UseTime:          useTimeSeconds,
		IsStream:         isStream,
		Group:            group,
		Ip: func() string {
			if needRecordIp {
				return c.ClientIP()
			}
			return ""
		}(),
		RequestId:         requestId,
		UpstreamRequestId: upstreamRequestId,
		Other:             otherStr,
	}
	err := createLog(log)
	if err != nil {
		logger.LogError(c, "failed to record log: "+err.Error())
	}
}

type RecordConsumeLogParams struct {
	ChannelId        int                    `json:"channel_id"`
	PromptTokens     int                    `json:"prompt_tokens"`
	CompletionTokens int                    `json:"completion_tokens"`
	ModelName        string                 `json:"model_name"`
	TokenName        string                 `json:"token_name"`
	Quota            int                    `json:"quota"`
	Content          string                 `json:"content"`
	TokenId          int                    `json:"token_id"`
	UseTimeSeconds   int                    `json:"use_time_seconds"`
	IsStream         bool                   `json:"is_stream"`
	Group            string                 `json:"group"`
	Other            map[string]interface{} `json:"other"`
	CostAccounting   *CostAccountingInput   `json:"-"`
}

func RecordConsumeLog(c *gin.Context, userId int, params RecordConsumeLogParams) {
	if !common.LogConsumeEnabled {
		return
	}
	logger.LogInfo(c, fmt.Sprintf(
		"record consume log: userId=%d, channelId=%d, model=%s, quota=%d",
		userId, params.ChannelId, params.ModelName, params.Quota,
	))
	username := c.GetString("username")
	requestId := c.GetString(common.RequestIdKey)
	upstreamRequestId := c.GetString(common.UpstreamRequestIdKey)
	createdAt := common.GetTimestamp()
	otherStr := common.MapToJsonStr(params.Other)
	// 判断是否需要记录 IP
	needRecordIp := false
	if userSetting, ok := common.GetContextKeyType[relaydto.UserSetting](c, constant.ContextKeyUserSetting); ok {
		needRecordIp = userSetting.RecordIpLog
	} else if settingMap, err := GetUserSetting(userId, false); err == nil {
		needRecordIp = settingMap.RecordIpLog
	}
	log := &Log{
		UserId:           userId,
		Username:         username,
		CreatedAt:        createdAt,
		Type:             LogTypeConsume,
		Content:          params.Content,
		PromptTokens:     params.PromptTokens,
		CompletionTokens: params.CompletionTokens,
		TokenName:        params.TokenName,
		ModelName:        params.ModelName,
		Quota:            params.Quota,
		ChannelId:        params.ChannelId,
		TokenId:          params.TokenId,
		UseTime:          params.UseTimeSeconds,
		IsStream:         params.IsStream,
		Group:            params.Group,
		Ip: func() string {
			if needRecordIp {
				return c.ClientIP()
			}
			return ""
		}(),
		RequestId:         requestId,
		UpstreamRequestId: upstreamRequestId,
		Other:             otherStr,
	}
	if err := createLog(log); err != nil {
		logger.LogError(c, "failed to record log: "+err.Error())
	} else if err := RecordCostAccountingSnapshot(log, params.CostAccounting); err != nil {
		logger.LogError(c, "failed to record cost accounting snapshot: "+err.Error())
	}
	if common.DataExportEnabled {
		LogQuotaData(QuotaDataLogParams{
			UserID:    userId,
			Username:  username,
			ModelName: params.ModelName,
			Quota:     params.Quota,
			CreatedAt: createdAt,
			TokenUsed: params.PromptTokens + params.CompletionTokens,
			UseGroup:  params.Group,
			TokenID:   params.TokenId,
			ChannelID: params.ChannelId,
			NodeName:  common.NodeName,
		})
	}
}

type RecordTaskBillingLogParams struct {
	UserId         int
	LogType        int
	Content        string
	ChannelId      int
	ModelName      string
	Quota          int
	TokenId        int
	Group          string
	Other          map[string]interface{}
	NodeName       string // 任务发起节点；为空时回退当前节点
	CostAccounting *CostAccountingInput
}

func RecordTaskBillingLog(params RecordTaskBillingLogParams) {
	if params.LogType == LogTypeConsume && !common.LogConsumeEnabled {
		return
	}
	username, _ := GetUsernameById(params.UserId, false)
	tokenName := ""
	if params.TokenId > 0 {
		if token, err := GetTokenById(params.TokenId); err == nil {
			tokenName = token.Name
		}
	}
	createdAt := common.GetTimestamp()
	log := &Log{
		UserId:    params.UserId,
		Username:  username,
		CreatedAt: createdAt,
		Type:      params.LogType,
		Content:   params.Content,
		TokenName: tokenName,
		ModelName: params.ModelName,
		Quota:     params.Quota,
		ChannelId: params.ChannelId,
		TokenId:   params.TokenId,
		Group:     params.Group,
		Other:     common.MapToJsonStr(params.Other),
	}
	if err := createLog(log); err != nil {
		common.SysLog("failed to record task billing log: " + err.Error())
	} else if err := RecordCostAccountingSnapshot(log, params.CostAccounting); err != nil {
		common.SysLog("failed to record task cost accounting snapshot: " + err.Error())
	}
	if params.LogType == LogTypeConsume && common.DataExportEnabled {
		nodeName := params.NodeName
		if nodeName == "" {
			nodeName = common.NodeName
		}
		LogQuotaData(QuotaDataLogParams{
			UserID:    params.UserId,
			Username:  username,
			ModelName: params.ModelName,
			Quota:     params.Quota,
			CreatedAt: createdAt,
			UseGroup:  params.Group,
			TokenID:   params.TokenId,
			ChannelID: params.ChannelId,
			NodeName:  nodeName,
		})
	}
}

func GetAllLogs(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, startIdx int, num int, channel int, group string, requestId string, upstreamRequestId string, includeAccounting bool) (logs []*Log, total int64, err error) {
	return getAllLogs(logType, startTimestamp, endTimestamp, modelName, username, tokenName, startIdx, num, channel, group, requestId, upstreamRequestId, includeAccounting, &logPageOptions{})
}

// logPageOptions tunes one page of the log search for a paged export. The zero
// value is the ordinary list page.
type logPageOptions struct {
	// SkipTotal omits the count, which is expensive on every page of a large result.
	SkipTotal bool
	// SkipDisplay leaves out what only the log page shows: channel names and the
	// task price details, which cost a query and a JSON parse per page.
	SkipDisplay bool
	// After continues behind the last log of the previous page. Unlike an
	// offset, it costs the same on every page.
	After *Log
	// adjustments are the billing corrections of a group-filtered search. The
	// first page loads them and later pages reuse them, since they depend on the
	// search and not on the page.
	adjustments       []billingEffectiveLog
	adjustmentsLoaded bool
	last              *Log
}

// rememberLast keeps the last log as the database returned it. The user view
// later replaces ids with display numbers, which cannot continue a cursor.
func (o *logPageOptions) rememberLast(logs []*Log) {
	if len(logs) == 0 {
		return
	}
	last := *logs[len(logs)-1]
	o.last = &last
}

func (o *logPageOptions) applyCursor(tx *gorm.DB) *gorm.DB {
	if o.After == nil {
		return tx
	}
	return tx.Where("(logs.created_at < ? OR (logs.created_at = ? AND logs.id < ?))", o.After.CreatedAt, o.After.CreatedAt, o.After.Id)
}

// effectiveLogs loads the billing corrections of a group-filtered search once:
// they depend on the search and not on the page, so later pages reuse them.
func (o *logPageOptions) effectiveLogs(filter CostAccountingFilter) ([]billingEffectiveLog, error) {
	if !o.adjustmentsLoaded {
		adjustments, err := billingEffectiveLogs(context.Background(), filter)
		if err != nil {
			return nil, err
		}
		o.adjustments, o.adjustmentsLoaded = adjustments, true
	}
	return o.adjustments, nil
}

// getAllLogs is the administrator log search.
func getAllLogs(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, startIdx int, num int, channel int, group string, requestId string, upstreamRequestId string, includeAccounting bool, opts *logPageOptions) (logs []*Log, total int64, err error) {
	var adjustments []billingEffectiveLog
	if group != "" {
		if adjustments, err = opts.effectiveLogs(CostAccountingFilter{LogType: logType, StartTimestamp: startTimestamp, EndTimestamp: endTimestamp, ModelName: modelName, Username: username, TokenName: tokenName, ChannelID: channel, RequestID: requestId, UpstreamRequestID: upstreamRequestId}); err != nil {
			return nil, 0, err
		}
	}
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = LOG_DB
	} else {
		tx = LOG_DB.Where("logs.type = ?", logType)
	}

	if tx, err = applyExplicitLogTextFilter(tx, "logs.model_name", modelName); err != nil {
		return nil, 0, err
	}
	if tx, err = applyExplicitLogTextFilter(tx, "logs.username", username); err != nil {
		return nil, 0, err
	}
	if tokenName != "" {
		tx = tx.Where("logs.token_name = ?", tokenName)
	}
	if requestId != "" {
		tx = tx.Where("logs.request_id = ?", requestId)
	}
	if upstreamRequestId != "" {
		tx = tx.Where("logs.upstream_request_id = ?", upstreamRequestId)
	}
	if startTimestamp != 0 {
		tx = tx.Where("logs.created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("logs.created_at <= ?", endTimestamp)
	}
	if channel != 0 {
		tx = tx.Where("logs.channel_id = ?", channel)
	}
	tx = applyBillingEffectiveGroup(tx, group, adjustments)
	tx = opts.applyCursor(tx)
	if !opts.SkipTotal {
		err = tx.Model(&Log{}).Count(&total).Error
		if err != nil {
			return nil, 0, err
		}
	}
	order := "logs.created_at desc, logs.id desc"
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		order = clickHouseLogOrder("logs.")
	}
	err = tx.Order(order).Limit(num).Offset(startIdx).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}
	opts.rememberLast(logs)
	if group == "" {
		adjustments, err = billingEffectiveLogsForPage(context.Background(), logs)
		if err != nil {
			return nil, 0, err
		}
	}
	if includeAccounting {
		if err := AttachLogAccounting(logs); err != nil {
			return nil, 0, err
		}
	}
	if !opts.SkipDisplay {
		if err := attachHistoricalTaskLogPricing(logs); err != nil {
			return nil, 0, err
		}
	}
	if err := projectBillingEffectiveLogs(logs, adjustments); err != nil {
		return nil, 0, err
	}
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		assignDisplayLogIds(logs, startIdx)
	}

	if opts.SkipDisplay {
		return logs, total, nil
	}

	channelIds := types.NewSet[int]()
	for _, log := range logs {
		if log.ChannelId != 0 {
			channelIds.Add(log.ChannelId)
		}
	}

	if channelIds.Len() > 0 {
		var channels []struct {
			Id   int    `gorm:"column:id"`
			Name string `gorm:"column:name"`
		}
		if common.MemoryCacheEnabled {
			// Cache get channel
			for _, channelId := range channelIds.Items() {
				if cacheChannel, err := CacheGetChannel(channelId); err == nil {
					channels = append(channels, struct {
						Id   int    `gorm:"column:id"`
						Name string `gorm:"column:name"`
					}{
						Id:   channelId,
						Name: cacheChannel.Name,
					})
				}
			}
		} else {
			// Bulk query channels from DB
			if err = DB.Table("channels").Select("id, name").Where("id IN ?", channelIds.Items()).Find(&channels).Error; err != nil {
				return logs, total, err
			}
		}
		channelMap := make(map[int]string, len(channels))
		for _, channel := range channels {
			channelMap[channel.Id] = channel.Name
		}
		for i := range logs {
			logs[i].ChannelName = channelMap[logs[i].ChannelId]
		}
	}

	return logs, total, err
}

const logSearchCountLimit = 10000

func GetUserLogs(userId int, logType int, startTimestamp int64, endTimestamp int64, modelName string, tokenName string, startIdx int, num int, group string, requestId string, upstreamRequestId string, includeAccounting bool) (logs []*Log, total int64, err error) {
	return getUserLogs(userId, logType, startTimestamp, endTimestamp, modelName, tokenName, startIdx, num, group, requestId, upstreamRequestId, includeAccounting, &logPageOptions{})
}

// getUserLogs is the user log search; see logPageOptions.
func getUserLogs(userId int, logType int, startTimestamp int64, endTimestamp int64, modelName string, tokenName string, startIdx int, num int, group string, requestId string, upstreamRequestId string, includeAccounting bool, opts *logPageOptions) (logs []*Log, total int64, err error) {
	var adjustments []billingEffectiveLog
	if group != "" {
		if adjustments, err = opts.effectiveLogs(CostAccountingFilter{UserID: userId, LogType: logType, StartTimestamp: startTimestamp, EndTimestamp: endTimestamp, ModelName: modelName, TokenName: tokenName, RequestID: requestId, UpstreamRequestID: upstreamRequestId}); err != nil {
			return nil, 0, err
		}
	}
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = LOG_DB.Where("logs.user_id = ?", userId)
	} else {
		tx = LOG_DB.Where("logs.user_id = ? and logs.type = ?", userId, logType)
	}
	// Enforce admin-only correction visibility in the query, including counts
	// and explicit type/request-id filters. Hiding a badge in the UI is insufficient.
	tx = tx.Where("logs.type <> ?", LogTypeBillingCorrection)

	if tx, err = applyExplicitLogTextFilter(tx, "logs.model_name", modelName); err != nil {
		return nil, 0, err
	}
	if tokenName != "" {
		tx = tx.Where("logs.token_name = ?", tokenName)
	}
	if requestId != "" {
		tx = tx.Where("logs.request_id = ?", requestId)
	}
	if upstreamRequestId != "" {
		tx = tx.Where("logs.upstream_request_id = ?", upstreamRequestId)
	}
	if startTimestamp != 0 {
		tx = tx.Where("logs.created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("logs.created_at <= ?", endTimestamp)
	}
	tx = applyBillingEffectiveGroup(tx, group, adjustments)
	tx = opts.applyCursor(tx)
	if !opts.SkipTotal {
		limitedCountQuery := tx.Session(&gorm.Session{}).
			Model(&Log{}).
			Select("logs.id").
			Limit(logSearchCountLimit)
		err = LOG_DB.Table("(?) AS limited_logs", limitedCountQuery).Count(&total).Error
		if err != nil {
			common.SysError("failed to count user logs: " + err.Error())
			return nil, 0, errors.New("查询日志失败")
		}
	}
	order := "logs.created_at desc, logs.id desc"
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		order = clickHouseLogOrder("logs.")
	}
	err = tx.Order(order).Limit(num).Offset(startIdx).Find(&logs).Error
	if err != nil {
		common.SysError("failed to search user logs: " + err.Error())
		return nil, 0, errors.New("查询日志失败")
	}
	opts.rememberLast(logs)
	if group == "" {
		adjustments, err = billingEffectiveLogsForPage(context.Background(), logs)
		if err != nil {
			return nil, 0, err
		}
	}
	if includeAccounting {
		if err := AttachLogAccounting(logs); err != nil {
			return nil, 0, err
		}
	}
	if !opts.SkipDisplay {
		if err := attachHistoricalTaskLogPricing(logs); err != nil {
			return nil, 0, err
		}
	}
	if err := projectBillingEffectiveLogs(logs, adjustments); err != nil {
		return nil, 0, err
	}

	formatUserLogs(logs, startIdx)
	return logs, total, err
}

type LogStatistics struct {
	Quota        int64 `json:"quota"`
	RefundQuota  int64 `json:"refund_quota"`
	RevenueQuota int64 `json:"revenue_quota"`
	Records      int64 `json:"records"`
	Requests     int64 `json:"requests"`
	Rpm          int   `json:"rpm"`
	Tpm          int   `json:"tpm"`
}

func applyLogStatisticsFilter(tx *gorm.DB, filter CostAccountingFilter) (*gorm.DB, error) {
	if !filter.IncludeBillingCorrections {
		tx = tx.Where("type <> ?", LogTypeBillingCorrection)
	}
	if filter.StartTimestamp != 0 {
		tx = tx.Where("created_at >= ?", filter.StartTimestamp)
	}
	if filter.EndTimestamp != 0 {
		tx = tx.Where("created_at <= ?", filter.EndTimestamp)
	}
	if filter.UserID != 0 {
		tx = tx.Where("user_id = ?", filter.UserID)
	}
	var err error
	if tx, err = applyExplicitLogTextFilter(tx, "username", filter.Username); err != nil {
		return nil, err
	}
	if filter.TokenName != "" {
		tx = tx.Where("token_name = ?", filter.TokenName)
	}
	if tx, err = applyExplicitLogTextFilter(tx, "model_name", filter.ModelName); err != nil {
		return nil, err
	}
	if filter.ChannelID != 0 {
		tx = tx.Where("channel_id = ?", filter.ChannelID)
	}
	if len(filter.channelIDs) > 0 {
		tx = tx.Where("channel_id IN ?", filter.channelIDs)
	}
	if filter.Group != "" {
		tx = tx.Where(logGroupCol+" = ?", filter.Group)
	}
	if filter.LogType != LogTypeUnknown {
		tx = tx.Where("type = ?", filter.LogType)
	}
	if filter.RequestID != "" {
		tx = tx.Where("request_id = ?", filter.RequestID)
	}
	if filter.UpstreamRequestID != "" {
		tx = tx.Where("upstream_request_id = ?", filter.UpstreamRequestID)
	}
	return tx, nil
}

func sumLogFinancialStatistics(filter CostAccountingFilter) (LogStatistics, error) {
	var statistics LogStatistics
	totalsQuery, err := applyLogStatisticsFilter(LOG_DB.Table("logs"), filter)
	if err != nil {
		return statistics, err
	}
	if err := totalsQuery.Select(`
		COALESCE(SUM(CASE WHEN type = ? OR (type = ? AND quota > 0) THEN quota ELSE 0 END), 0) AS quota,
		COALESCE(SUM(CASE WHEN type = ? THEN quota WHEN type = ? AND quota < 0 THEN -quota ELSE 0 END), 0) AS refund_quota,
		COALESCE(SUM(CASE WHEN type IN (?, ?, ?) THEN 1 ELSE 0 END), 0) AS records,
		COALESCE(SUM(CASE WHEN type = ? THEN 1 ELSE 0 END), 0) AS requests`,
		LogTypeConsume, LogTypeBillingCorrection, LogTypeRefund, LogTypeBillingCorrection, LogTypeConsume, LogTypeRefund, LogTypeBillingCorrection, LogTypeConsume,
	).Scan(&statistics).Error; err != nil {
		common.SysError("failed to query log stat: " + err.Error())
		return statistics, errors.New("查询统计数据失败")
	}
	statistics.RevenueQuota = statistics.Quota - statistics.RefundQuota
	return statistics, nil
}

func SumLogStatistics(filter CostAccountingFilter) (LogStatistics, error) {
	statistics, err := sumBillingEffectiveLogStatistics(filter)
	if err != nil {
		return statistics, err
	}

	rateFilter := filter
	rateFilter.Group = ""
	rateQuery, err := applyLogStatisticsFilter(LOG_DB.Table("logs"), rateFilter)
	if err != nil {
		return statistics, err
	}
	if filter.Group != "" {
		rateFilter.StartTimestamp = max(filter.StartTimestamp, time.Now().Add(-60*time.Second).Unix())
		adjustments, err := billingEffectiveLogs(context.Background(), rateFilter)
		if err != nil {
			return statistics, err
		}
		rateQuery = applyBillingEffectiveGroup(rateQuery, filter.Group, adjustments)
	}
	rateQuery = rateQuery.
		Select("count(*) rpm, COALESCE(sum(prompt_tokens), 0) + COALESCE(sum(completion_tokens), 0) tpm").
		Where("type = ?", LogTypeConsume).
		Where("created_at >= ?", time.Now().Add(-60*time.Second).Unix())
	var rateStatistics struct {
		Rpm int
		Tpm int
	}
	if err := rateQuery.Scan(&rateStatistics).Error; err != nil {
		common.SysError("failed to query rpm/tpm stat: " + err.Error())
		return statistics, errors.New("查询统计数据失败")
	}
	statistics.Rpm = rateStatistics.Rpm
	statistics.Tpm = rateStatistics.Tpm
	return statistics, nil
}

func sumLogAccountingBuckets(filter CostAccountingFilter, bucketSeconds, offsetSeconds int64) ([]CostAccountingBucket, error) {
	if bucketSeconds <= 0 {
		return nil, errors.New("invalid accounting bucket size")
	}
	bucketExpr := "CAST((created_at + ?) / ? AS BIGINT) * ? - ?"
	if common.UsingLogDatabase(common.DatabaseTypeMySQL) {
		bucketExpr = "((created_at + ?) DIV ?) * ? - ?"
	}
	query, err := applyLogStatisticsFilter(LOG_DB.Table("logs"), filter)
	if err != nil {
		return nil, err
	}
	buckets := make([]CostAccountingBucket, 0)
	err = query.Select(bucketExpr+` AS bucket,
			COALESCE(SUM(CASE WHEN type IN (?, ?) THEN quota WHEN type = ? THEN -quota ELSE 0 END), 0) AS revenue_quota,
			COUNT(*) AS records`, offsetSeconds, bucketSeconds, bucketSeconds, offsetSeconds, LogTypeConsume, LogTypeBillingCorrection, LogTypeRefund).
		Group("bucket").Order("bucket asc").Scan(&buckets).Error
	return buckets, err
}

func CountOldLog(ctx context.Context, targetTimestamp int64) (int64, error) {
	var total int64
	if err := LOG_DB.WithContext(ctx).Model(&Log{}).Where("created_at < ?", targetTimestamp).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func DeleteOldLogBatch(ctx context.Context, targetTimestamp int64, limit int) (int64, error) {
	return 0, ErrUsageLogsRetained
}

// Usage logs are accounting evidence. Archiving payloads never authorizes
// deleting the original usage records, including through legacy cleanup jobs.
var ErrUsageLogsRetained = errors.New("usage logs are retained permanently; deletion is disabled")

// logExportPageSize is the number of logs an export reads per query.
const logExportPageSize = 1000

// logExportPagePause lets other queries through between the pages of an export.
const logExportPagePause = 20 * time.Millisecond

// exportLogPages calls fn with each log fetch returns, newest first, one page
// at a time so a large export never holds the whole result in memory. Pages
// follow a cursor on (created_at, id); ClickHouse ids are display values, so
// it pages by offset there.
func exportLogPages(ctx context.Context, fetch func(offset int, opts *logPageOptions) ([]*Log, error), fn func(*Log) error) error {
	opts := &logPageOptions{SkipTotal: true, SkipDisplay: true}
	byOffset := common.UsingLogDatabase(common.DatabaseTypeClickHouse)
	offset := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		logs, err := fetch(offset, opts)
		if err != nil {
			return err
		}
		for _, log := range logs {
			if err := fn(log); err != nil {
				return err
			}
		}
		if len(logs) < logExportPageSize {
			return nil
		}
		if byOffset {
			offset += logExportPageSize
		} else {
			opts.After = opts.last
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(logExportPagePause):
		}
	}
}

// ExportAllLogs calls fn with each administrator-visible log matching the
// search, newest first, under the same visibility rules as the log list. It
// never attaches the platform's cost and profit figures: a file leaves the
// platform, so an export cannot ask for them.
func ExportAllLogs(ctx context.Context, logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, channel int, group string, requestId string, upstreamRequestId string, fn func(*Log) error) error {
	return exportLogPages(ctx, func(offset int, opts *logPageOptions) ([]*Log, error) {
		logs, _, err := getAllLogs(logType, startTimestamp, endTimestamp, modelName, username, tokenName, offset, logExportPageSize, channel, group, requestId, upstreamRequestId, false, opts)
		return logs, err
	}, fn)
}

// ExportUserLogs is ExportAllLogs for one user's own logs.
func ExportUserLogs(ctx context.Context, userId int, logType int, startTimestamp int64, endTimestamp int64, modelName string, tokenName string, group string, requestId string, upstreamRequestId string, fn func(*Log) error) error {
	return exportLogPages(ctx, func(offset int, opts *logPageOptions) ([]*Log, error) {
		logs, _, err := getUserLogs(userId, logType, startTimestamp, endTimestamp, modelName, tokenName, offset, logExportPageSize, group, requestId, upstreamRequestId, false, opts)
		return logs, err
	}, fn)
}

// LogLink is what a usage log says about the request behind a ledger entry.
type LogLink struct {
	Found                   bool
	Group                   string
	TaskID                  string
	IsStream                bool
	PromptTokens            int
	CompletionTokens        int
	CacheTokens, CacheWrite int
	UseTime                 int
	Ip                      string
	// GroupRatio is the price multiplier the request was billed with.
	GroupRatio string
	// IsTask marks the log of a video or other asynchronous task.
	IsTask bool
	// CreatedAt is when the log was written, which for a task is when it was
	// submitted.
	CreatedAt   int64
	consumption bool
}

// LogLinkOf reads what a usage log says about the request behind it.
func LogLinkOf(log *Log) LogLink {
	var other struct {
		TaskID                string   `json:"task_id"`
		CacheTokens           int      `json:"cache_tokens"`
		CacheCreationTokens   int      `json:"cache_creation_tokens"`
		CacheCreationTokens5m int      `json:"cache_creation_tokens_5m"`
		CacheCreationTokens1h int      `json:"cache_creation_tokens_1h"`
		IsTask                bool     `json:"is_task"`
		GroupRatio            *float64 `json:"group_ratio"`
		UserGroupRatio        *float64 `json:"user_group_ratio"`
	}
	if log.Other != "" {
		// Logs without readable metadata simply carry no task or cache figures.
		_ = common.UnmarshalJsonStr(log.Other, &other)
	}
	cacheWrite := other.CacheCreationTokens5m + other.CacheCreationTokens1h
	if cacheWrite == 0 {
		cacheWrite = other.CacheCreationTokens
	}
	ratio := other.GroupRatio
	if other.UserGroupRatio != nil && *other.UserGroupRatio > 0 {
		ratio = other.UserGroupRatio
	}
	groupRatio := ""
	if ratio != nil && *ratio >= 0 {
		groupRatio = strconv.FormatFloat(*ratio, 'f', -1, 64)
	}
	return LogLink{
		Found: true, Group: log.Group, GroupRatio: groupRatio, TaskID: other.TaskID, IsTask: other.IsTask, CreatedAt: log.CreatedAt, IsStream: log.IsStream,
		PromptTokens: log.PromptTokens, CompletionTokens: log.CompletionTokens,
		CacheTokens: other.CacheTokens, CacheWrite: cacheWrite, UseTime: log.UseTime, Ip: log.Ip,
		consumption: log.Type == LogTypeConsume,
	}
}

// logLinkColumns are the log columns a link reads; the usage figures and the
// task ID live in the metadata.
var logLinkColumns = []string{"id", "request_id", "type", "group", "other", "is_stream", "prompt_tokens", "completion_tokens", "use_time", "ip", "created_at"}

const logLinkBatchSize = 500

// GetLogLinksByIDs returns the usage logs of the user by their id.
func GetLogLinksByIDs(userID int, ids []int) (map[int]LogLink, error) {
	links := make(map[int]LogLink, len(ids))
	for _, batch := range lo.Chunk(ids, logLinkBatchSize) {
		var logs []Log
		if err := LOG_DB.Select(logLinkColumns).Where("user_id = ? AND id IN ?", userID, batch).Find(&logs).Error; err != nil {
			return nil, err
		}
		for i := range logs {
			links[logs[i].Id] = LogLinkOf(&logs[i])
		}
	}
	return links, nil
}

// GetLogLinksByRequestIDs returns the usage logs of the user by request ID.
// A request can leave several logs: its consumption log supplies the usage
// figures, and any of them may supply the group or task.
func GetLogLinksByRequestIDs(userID int, requestIDs []string) (map[string]LogLink, error) {
	links := make(map[string]LogLink, len(requestIDs))
	for _, batch := range lo.Chunk(requestIDs, logLinkBatchSize) {
		var logs []Log
		if err := LOG_DB.Select(logLinkColumns).Where("user_id = ? AND request_id IN ?", userID, batch).Find(&logs).Error; err != nil {
			return nil, err
		}
		for i := range logs {
			best := LogLinkOf(&logs[i])
			if existing, found := links[logs[i].RequestId]; found {
				other := best
				best = existing
				if !existing.consumption && other.consumption {
					best, other = other, existing
				}
				best.Group = cmp.Or(best.Group, other.Group)
				best.TaskID = cmp.Or(best.TaskID, other.TaskID)
			}
			links[logs[i].RequestId] = best
		}
	}
	return links, nil
}
