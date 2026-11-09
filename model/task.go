package model

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	commonRelay "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	hosttypes "github.com/QuantumNous/new-api/types"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type TaskStatus string

func (t TaskStatus) ToVideoStatus() string {
	var status string
	switch t {
	case TaskStatusQueued, TaskStatusSubmitted:
		status = dto.VideoStatusQueued
	case TaskStatusInProgress:
		status = dto.VideoStatusInProgress
	case TaskStatusSuccess:
		status = dto.VideoStatusCompleted
	case TaskStatusFailure:
		status = dto.VideoStatusFailed
	default:
		status = dto.VideoStatusUnknown // Default fallback
	}
	return status
}

const (
	TaskStatusNotStart   TaskStatus = "NOT_START"
	TaskStatusSubmitted             = "SUBMITTED"
	TaskStatusQueued                = "QUEUED"
	TaskStatusInProgress            = "IN_PROGRESS"
	TaskStatusFailure               = "FAILURE"
	TaskStatusSuccess               = "SUCCESS"
	TaskStatusUnknown               = "UNKNOWN"
)

// TaskRefundLegacyCutoff separates tasks created before timeout refunds were
// introduced. Those legacy tasks are failed without an automatic refund.
const TaskRefundLegacyCutoff int64 = 1771718400 // 2026-02-22 00:00:00 UTC

type Task struct {
	ID                   int64                 `json:"id" gorm:"primary_key;AUTO_INCREMENT;index:idx_task_submit_id,priority:2;index:idx_task_user_submit_id,priority:3"`
	CreatedAt            int64                 `json:"created_at" gorm:"index"`
	UpdatedAt            int64                 `json:"updated_at"`
	WebhookVersion       int64                 `json:"-" gorm:"bigint"`
	TaskID               string                `json:"task_id" gorm:"type:varchar(191);index"` // 第三方id，不一定有/ song id\ Task id
	Platform             constant.TaskPlatform `json:"platform" gorm:"type:varchar(30);index"` // 平台
	UserId               int                   `json:"user_id" gorm:"index;index:idx_task_user_submit_id,priority:1"`
	Group                string                `json:"group" gorm:"type:varchar(50)"` // 修正计费用
	ChannelId            int                   `json:"channel_id" gorm:"index"`
	Quota                int                   `json:"quota"`
	Action               string                `json:"action" gorm:"type:varchar(40);index"` // 任务类型, song, lyrics, description-mode
	Status               TaskStatus            `json:"status" gorm:"type:varchar(20);index"` // 任务状态
	FailReason           string                `json:"fail_reason"`
	SubmitTime           int64                 `json:"submit_time" gorm:"index;index:idx_task_submit_id,priority:1;index:idx_task_user_submit_id,priority:2"`
	StartTime            int64                 `json:"start_time" gorm:"index"`
	FinishTime           int64                 `json:"finish_time" gorm:"index"`
	Progress             string                `json:"progress" gorm:"type:varchar(20);index"`
	Properties           Properties            `json:"properties" gorm:"type:json"`
	RequestBodyAvailable bool                  `json:"-"`
	Username             string                `json:"username,omitempty" gorm:"-"`
	// 禁止返回给用户，内部可能包含key等隐私信息
	PrivateData TaskPrivateData `json:"-" gorm:"column:private_data;type:json"`
	Data        json.RawMessage `json:"data" gorm:"type:json"`
}

func (t *Task) SetData(data any) {
	b, _ := common.Marshal(data)
	t.Data = json.RawMessage(b)
}

func (t *Task) GetData(v any) error {
	return common.Unmarshal(t.Data, &v)
}

type Properties struct {
	SnapshotKinds     []string                      `json:"-"`
	RequestID         string                        `json:"-"`
	RequestEvidence   *TaskRequestEvidence          `json:"-"`
	Input             string                        `json:"input"`
	UpstreamModelName string                        `json:"upstream_model_name,omitempty"`
	OriginModelName   string                        `json:"origin_model_name,omitempty"`
	RequestBody       json.RawMessage               `json:"-"`
	RequestMetrics    *hosttypes.TaskRequestMetrics `json:"-"`
}

type propertiesStorage struct {
	SnapshotKinds     []string                      `json:"snapshot_kinds,omitempty"`
	RequestID         string                        `json:"request_id,omitempty"`
	RequestEvidence   *TaskRequestEvidence          `json:"request_evidence,omitempty"`
	Input             string                        `json:"input"`
	UpstreamModelName string                        `json:"upstream_model_name,omitempty"`
	OriginModelName   string                        `json:"origin_model_name,omitempty"`
	RequestBody       json.RawMessage               `json:"request_body,omitempty"`
	RequestMetrics    *hosttypes.TaskRequestMetrics `json:"request_metrics,omitempty"`
}

func (m *Properties) Scan(val interface{}) error {
	bytesValue, _ := val.([]byte)
	if len(bytesValue) == 0 {
		*m = Properties{}
		return nil
	}
	var stored propertiesStorage
	if err := common.Unmarshal(bytesValue, &stored); err != nil {
		return err
	}
	*m = Properties{
		SnapshotKinds:     stored.SnapshotKinds,
		RequestID:         stored.RequestID,
		RequestEvidence:   stored.RequestEvidence,
		Input:             stored.Input,
		UpstreamModelName: stored.UpstreamModelName,
		OriginModelName:   stored.OriginModelName,
		RequestBody:       stored.RequestBody,
		RequestMetrics:    stored.RequestMetrics,
	}
	return nil
}

func (m Properties) Value() (driver.Value, error) {
	if len(m.SnapshotKinds) == 0 && m.RequestID == "" && m.RequestEvidence == nil && m.Input == "" && m.UpstreamModelName == "" && m.OriginModelName == "" && len(m.RequestBody) == 0 && m.RequestMetrics == nil {
		return nil, nil
	}
	return common.Marshal(propertiesStorage{
		SnapshotKinds:     m.SnapshotKinds,
		RequestID:         m.RequestID,
		RequestEvidence:   m.RequestEvidence,
		Input:             m.Input,
		UpstreamModelName: m.UpstreamModelName,
		OriginModelName:   m.OriginModelName,
		RequestBody:       m.RequestBody,
		RequestMetrics:    m.RequestMetrics,
	})
}

type TaskPrivateData struct {
	Key            string `json:"key,omitempty"`
	UpstreamTaskID string `json:"upstream_task_id,omitempty"` // 上游真实 task ID
	ResultURL      string `json:"result_url,omitempty"`       // 任务成功后的结果 URL（视频地址等）
	// 计费上下文：用于异步退款/差额结算（轮询阶段读取）
	BillingSource  string              `json:"billing_source,omitempty"`  // "wallet" 或 "subscription"
	SubscriptionId int                 `json:"subscription_id,omitempty"` // 订阅 ID，用于订阅退款
	TokenId        int                 `json:"token_id,omitempty"`        // 令牌 ID，用于令牌额度退款
	NodeName       string              `json:"node_name,omitempty"`       // 发起任务的节点名，轮询结算阶段据此归属日志而非最后查询节点
	BillingContext *TaskBillingContext `json:"billing_context,omitempty"` // 计费参数快照（用于轮询阶段重新计算）
}

// TaskBillingContext 记录任务提交时的计费参数，以便轮询阶段可以重新计算额度。
type TaskBillingContext struct {
	ModelPrice      float64            `json:"model_price,omitempty"`       // 模型单价
	GroupRatio      float64            `json:"group_ratio,omitempty"`       // 分组倍率
	ModelRatio      float64            `json:"model_ratio,omitempty"`       // 模型倍率
	OtherRatios     map[string]float64 `json:"other_ratios,omitempty"`      // 附加倍率（时长、分辨率等）
	OriginModelName string             `json:"origin_model_name,omitempty"` // 模型名称，必须为OriginModelName
	PerCallBilling  bool               `json:"per_call_billing,omitempty"`  // 按次计费：跳过轮询阶段的差额结算
	CostDiscount    string             `json:"cost_discount,omitempty"`     // 提交时渠道成本折扣快照
}

// GetUpstreamTaskID 获取上游真实 task ID（用于与 provider 通信）
// 旧数据没有 UpstreamTaskID 时，TaskID 本身就是上游 ID
func (t *Task) GetUpstreamTaskID() string {
	if t.PrivateData.UpstreamTaskID != "" {
		return t.PrivateData.UpstreamTaskID
	}
	return t.TaskID
}

// GetResultURL 获取任务结果 URL（视频地址等）
// 新数据存在 PrivateData.ResultURL 中；旧数据回退到 FailReason（历史兼容）
func (t *Task) GetResultURL() string {
	if t.PrivateData.ResultURL != "" {
		return t.PrivateData.ResultURL
	}
	return t.FailReason
}

// GenerateTaskID 生成对外暴露的 task_xxxx 格式 ID
func GenerateTaskID() string {
	key, _ := common.GenerateRandomCharsKey(32)
	return "task_" + key
}

func (p *TaskPrivateData) Scan(val interface{}) error {
	bytesValue, _ := val.([]byte)
	if len(bytesValue) == 0 {
		return nil
	}
	return common.Unmarshal(bytesValue, p)
}

func (p TaskPrivateData) Value() (driver.Value, error) {
	if (p == TaskPrivateData{}) {
		return nil, nil
	}
	return common.Marshal(p)
}

// SyncTaskQueryParams 用于包含所有搜索条件的结构体，可以根据需求添加更多字段
type SyncTaskQueryParams struct {
	Platform       constant.TaskPlatform
	ChannelID      string
	TaskID         string
	UserID         string
	Action         string
	Status         string
	StartTimestamp int64
	EndTimestamp   int64
	UserIDs        []int
	Username       string
	ModelName      string
}

func InitTask(platform constant.TaskPlatform, relayInfo *commonRelay.RelayInfo) *Task {
	properties := Properties{}
	privateData := TaskPrivateData{}
	if relayInfo != nil && relayInfo.ChannelMeta != nil {
		if relayInfo.ChannelMeta.ChannelType == constant.ChannelTypeGemini ||
			relayInfo.ChannelMeta.ChannelType == constant.ChannelTypeVertexAi ||
			relayInfo.ChannelMeta.ChannelType == constant.ChannelTypeVolcNative ||
			relayInfo.ChannelMeta.ChannelType == constant.ChannelTypeAli {
			privateData.Key = relayInfo.ChannelMeta.ApiKey
		}
		if relayInfo.UpstreamModelName != "" {
			properties.UpstreamModelName = relayInfo.UpstreamModelName
		}
		if relayInfo.OriginModelName != "" {
			properties.OriginModelName = relayInfo.OriginModelName
		}
	}
	if relayInfo != nil && relayInfo.TaskRelayInfo != nil && relayInfo.RequestMetrics.HasData() {
		metrics := relayInfo.RequestMetrics
		properties.RequestMetrics = &metrics
	}

	// 使用预生成的公开 ID（如果有），否则新生成
	taskID := ""
	if relayInfo.TaskRelayInfo != nil && relayInfo.TaskRelayInfo.PublicTaskID != "" {
		taskID = relayInfo.TaskRelayInfo.PublicTaskID
	} else {
		taskID = GenerateTaskID()
	}

	t := &Task{
		TaskID:      taskID,
		UserId:      relayInfo.UserId,
		Group:       relayInfo.UsingGroup,
		SubmitTime:  time.Now().Unix(),
		Status:      TaskStatusNotStart,
		Progress:    "0%",
		ChannelId:   relayInfo.ChannelId,
		Platform:    platform,
		Properties:  properties,
		PrivateData: privateData,
	}
	return t
}

func TaskGetAllUserTask(userId int, startIdx int, num int, queryParams SyncTaskQueryParams) ([]*Task, error) {
	var tasks []*Task
	query := filterTaskList(DB.Model(&Task{}).Where("user_id = ?", userId), queryParams)
	err := query.Select(taskListSelectColumns(false)).Order("submit_time desc, id desc").Limit(num).Offset(startIdx).Find(&tasks).Error
	return tasks, err
}

func TaskGetAllTasks(startIdx int, num int, queryParams SyncTaskQueryParams) ([]*Task, error) {
	var tasks []*Task

	query := filterTaskList(DB.Model(&Task{}), queryParams)
	err := query.Select(taskListSelectColumns(true)).Order("submit_time desc, id desc").Limit(num).Offset(startIdx).Find(&tasks).Error
	return tasks, err
}

func filterTaskList(query *gorm.DB, queryParams SyncTaskQueryParams) *gorm.DB {
	if queryParams.ChannelID != "" {
		query = query.Where("channel_id = ?", queryParams.ChannelID)
	}
	if queryParams.Platform != "" {
		query = query.Where("platform = ?", queryParams.Platform)
	}
	if queryParams.UserID != "" {
		query = query.Where("user_id = ?", queryParams.UserID)
	}
	if len(queryParams.UserIDs) != 0 {
		query = query.Where("user_id IN ?", queryParams.UserIDs)
	}
	if keyword := strings.TrimSpace(queryParams.Username); keyword != "" {
		pattern := taskSearchPattern(keyword)
		users := DB.Model(&User{}).Select("id").Where("LOWER(username) LIKE ? ESCAPE '!'", pattern)
		query = query.Where("user_id IN (?)", users)
	}
	if keyword := strings.TrimSpace(queryParams.ModelName); keyword != "" {
		var modelExpression string
		switch common.MainDatabaseType() {
		case common.DatabaseTypePostgreSQL:
			modelExpression = "COALESCE(NULLIF(properties->>'origin_model_name', ''), properties->>'upstream_model_name', '')"
		case common.DatabaseTypeMySQL:
			modelExpression = "COALESCE(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(properties, '$.origin_model_name')), ''), JSON_UNQUOTE(JSON_EXTRACT(properties, '$.upstream_model_name')), '')"
		default:
			modelExpression = "COALESCE(NULLIF(json_extract(properties, '$.origin_model_name'), ''), json_extract(properties, '$.upstream_model_name'), '')"
		}
		query = query.Where("LOWER("+modelExpression+") LIKE ? ESCAPE '!'", taskSearchPattern(keyword))
	}
	if queryParams.TaskID != "" {
		query = query.Where("task_id = ?", queryParams.TaskID)
	}
	if queryParams.Action != "" {
		query = query.Where("action = ?", queryParams.Action)
	}
	if queryParams.Status != "" {
		query = query.Where("status = ?", queryParams.Status)
	}
	if queryParams.StartTimestamp != 0 {
		query = query.Where("submit_time >= ?", queryParams.StartTimestamp)
	}
	if queryParams.EndTimestamp != 0 {
		query = query.Where("submit_time <= ?", queryParams.EndTimestamp)
	}

	return query
}

func taskSearchPattern(keyword string) string {
	keyword = strings.ToLower(keyword)
	keyword = strings.ReplaceAll(keyword, "!", "!!")
	keyword = strings.ReplaceAll(keyword, "%", "!%")
	keyword = strings.ReplaceAll(keyword, "_", "!_")
	return "%" + keyword + "%"
}

func taskListSelectColumns(includeChannel bool) []string {
	columns := []string{
		"id", "created_at", "updated_at", "task_id", "platform", "user_id", "group",
		"quota", "action", "status", "fail_reason", "submit_time", "start_time", "finish_time",
		"progress", "properties", "request_body_available", "private_data", "data",
	}
	if includeChannel {
		columns = append(columns, "channel_id")
	}
	return columns
}

func GetTimedOutUnfinishedTasks(cutoffUnix int64, limit int) []*Task {
	var tasks []*Task
	err := DB.Where("progress != ?", "100%").
		Where("status NOT IN ?", []string{TaskStatusFailure, TaskStatusSuccess}).
		Where("submit_time < ?", cutoffUnix).
		Order("submit_time").
		Limit(limit).
		Find(&tasks).Error
	if err != nil {
		return nil
	}
	return tasks
}

func GetAllUnFinishSyncTasks(limit int) []*Task {
	var tasks []*Task
	var err error
	// get all tasks progress is not 100%
	err = DB.Where("progress != ?", "100%").Where("status != ?", TaskStatusFailure).Where("status != ?", TaskStatusSuccess).Limit(limit).Order("id").Find(&tasks).Error
	if err != nil {
		return nil
	}
	return tasks
}

// HasUnfinishedSyncTasks reports whether at least one async (Suno/video) task is
// still in progress. It is a cheap existence check (LIMIT 1) used to decide
// whether the async_task_poll system task needs to run; when no task is pending
// the scheduler skips creating a row entirely.
func HasUnfinishedSyncTasks() bool {
	var id int64
	err := DB.Model(&Task{}).
		Where("progress != ?", "100%").
		Where("status != ?", TaskStatusFailure).
		Where("status != ?", TaskStatusSuccess).
		Limit(1).
		Pluck("id", &id).Error
	return err == nil && id != 0
}

func GetByTaskId(userId int, taskId string) (*Task, bool, error) {
	if taskId == "" {
		return nil, false, nil
	}
	var task *Task
	var err error
	err = DB.Where("user_id = ? and task_id = ?", userId, taskId).
		First(&task).Error
	exist, err := RecordExist(err)
	if err != nil {
		return nil, false, err
	}
	return task, exist, err
}

func GetByTaskIdForAdmin(taskId string) (*Task, bool, error) {
	if taskId == "" {
		return nil, false, nil
	}
	var task *Task
	err := DB.Where("task_id = ?", taskId).First(&task).Error
	exist, err := RecordExist(err)
	if err != nil {
		return nil, false, err
	}
	return task, exist, nil
}

func GetByTaskIds(userId int, taskIds []any) ([]*Task, error) {
	if len(taskIds) == 0 {
		return nil, nil
	}
	var task []*Task
	var err error
	err = DB.Where("user_id = ? and task_id in (?)", userId, taskIds).
		Find(&task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (Task *Task) Insert() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		Task.WebhookVersion = 1
		if err := tx.Create(Task).Error; err != nil {
			return err
		}
		return Task.queueStatusWebhook(tx)
	})
}

type taskSnapshot struct {
	Status     TaskStatus
	Progress   string
	StartTime  int64
	FinishTime int64
	FailReason string
	ResultURL  string
	Data       json.RawMessage
}

func (s taskSnapshot) Equal(other taskSnapshot) bool {
	return s.Status == other.Status &&
		s.Progress == other.Progress &&
		s.StartTime == other.StartTime &&
		s.FinishTime == other.FinishTime &&
		s.FailReason == other.FailReason &&
		s.ResultURL == other.ResultURL &&
		bytes.Equal(s.Data, other.Data)
}

func (t *Task) Snapshot() taskSnapshot {
	return taskSnapshot{
		Status:     t.Status,
		Progress:   t.Progress,
		StartTime:  t.StartTime,
		FinishTime: t.FinishTime,
		FailReason: t.FailReason,
		ResultURL:  t.PrivateData.ResultURL,
		Data:       t.Data,
	}
}

func (Task *Task) Update() error {
	if Task.ID == 0 {
		return Task.Insert()
	}
	var previous struct{ Status TaskStatus }
	if err := DB.Model(Task).Select("status").First(&previous).Error; err != nil {
		return err
	}
	_, err := Task.UpdateWithStatus(previous.Status)
	return err
}

func (t *Task) UpdateQuota() error {
	return DB.Model(t).Update("quota", t.Quota).Error
}

// UpdateWithStatus performs a conditional UPDATE guarded by fromStatus (CAS).
// Returns (true, nil) if this caller won the update, (false, nil) if
// another process already moved the task out of fromStatus. MySQL commonly
// reports changed rows rather than matched rows, so a same-value no-op update
// can also return false even when the status predicate still matched.
//
// Uses Model().Select("*").Updates() instead of Save() because GORM's Save
// falls back to INSERT ON CONFLICT when the WHERE-guarded UPDATE matches
// zero rows, which silently bypasses the CAS guard.
func (t *Task) UpdateWithStatus(fromStatus TaskStatus) (bool, error) {
	won := false
	err := DB.Transaction(func(tx *gorm.DB) error {
		var previous Task
		if err := lockForUpdate(tx).Select("id", "status", "webhook_version", "request_body_available").First(&previous, t.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if previous.Status != fromStatus {
			return nil
		}
		t.WebhookVersion = previous.WebhookVersion
		changed := t.Status != previous.Status
		if changed {
			t.WebhookVersion++
		}
		t.RequestBodyAvailable = t.RequestBodyAvailable || previous.RequestBodyAvailable
		update := tx.Model(t).Where("status = ?", fromStatus).Select("*")
		if previous.RequestBodyAvailable && len(t.Properties.RequestBody) > 0 {
			var current Task
			if err := tx.Select("properties").First(&current, t.ID).Error; err != nil {
				return err
			}
			if len(current.Properties.RequestBody) == 0 {
				// A poll may have loaded the legacy body before maintenance removed
				// it. Preserve the current properties, including unknown schema keys.
				t.Properties = current.Properties
				update = update.Omit("properties")
			}
		}
		result := update.Updates(t)
		if result.Error != nil {
			return result.Error
		}
		won = result.RowsAffected > 0
		if won && changed {
			return t.queueStatusWebhook(tx)
		}
		return nil
	})
	return won && err == nil, err
}

// TaskBulkUpdateByID performs an unconditional bulk UPDATE by primary key IDs.
// WARNING: This function has NO CAS (Compare-And-Swap) guard — it will overwrite
// any concurrent status changes. DO NOT use in billing/quota lifecycle flows
// (e.g., timeout, success, failure transitions that trigger refunds or settlements).
// For status transitions that involve billing, use Task.UpdateWithStatus() instead.
func TaskBulkUpdateByID(ids []int64, params map[string]any) error {
	if len(ids) == 0 {
		return nil
	}
	if _, changesStatus := params["status"]; !changesStatus {
		return DB.Model(&Task{}).Where("id in (?)", ids).Updates(params).Error
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var tasks []Task
		if err := lockForUpdate(tx).Where("id IN ?", ids).Find(&tasks).Error; err != nil {
			return err
		}
		for _, task := range tasks {
			previousStatus := task.Status
			result := tx.Model(&task).Where("status = ?", previousStatus).Updates(params)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				continue
			}
			if err := tx.First(&task, task.ID).Error; err != nil {
				return err
			}
			if task.Status == previousStatus {
				continue
			}
			task.WebhookVersion++
			if err := tx.Model(&task).Update("webhook_version", task.WebhookVersion).Error; err != nil {
				return err
			}
			if err := task.queueStatusWebhook(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

type TaskQuotaUsage struct {
	Mode  string  `json:"mode"`
	Count float64 `json:"count"`
}

// TaskCountAllTasks returns total tasks that match the given query params (admin usage)
func TaskCountAllTasks(queryParams SyncTaskQueryParams) (int64, error) {
	var total int64
	err := filterTaskList(DB.Model(&Task{}), queryParams).Count(&total).Error
	return total, err
}

// TaskCountAllUserTask returns total tasks for given user
func TaskCountAllUserTask(userId int, queryParams SyncTaskQueryParams) (int64, error) {
	var total int64
	err := filterTaskList(DB.Model(&Task{}).Where("user_id = ?", userId), queryParams).Count(&total).Error
	return total, err
}
func (t *Task) ToOpenAIVideo() *dto.OpenAIVideo {
	openAIVideo := dto.NewOpenAIVideo()
	openAIVideo.ID = t.TaskID
	openAIVideo.Status = t.Status.ToVideoStatus()
	openAIVideo.Model = t.Properties.OriginModelName
	openAIVideo.SetProgressStr(t.Progress)
	openAIVideo.CreatedAt = t.CreatedAt
	openAIVideo.CompletedAt = t.UpdatedAt
	openAIVideo.SetMetadata("url", t.GetResultURL())
	return openAIVideo
}

// TaskUsage is what a finished task consumed: the tokens its platform reported
// and how long it ran.
type TaskUsage struct {
	TotalTokens int
	Seconds     int
}

// GetTaskUsageByTaskIDs returns the usage of the user's tasks by their public
// task ID. Tasks the platform reported no usage for are still listed, so the
// run time shows.
func GetTaskUsageByTaskIDs(userID int, taskIDs []string) (map[string]TaskUsage, error) {
	usage := make(map[string]TaskUsage, len(taskIDs))
	for _, batch := range lo.Chunk(taskIDs, 500) {
		var tasks []Task
		if err := DB.Select("task_id", "submit_time", "finish_time", "data").Where("user_id = ? AND task_id IN ?", userID, batch).Find(&tasks).Error; err != nil {
			return nil, err
		}
		for _, task := range tasks {
			var data struct {
				Usage struct {
					TotalTokens      int `json:"total_tokens"`
					CompletionTokens int `json:"completion_tokens"`
				} `json:"usage"`
			}
			if len(task.Data) > 0 {
				// Tasks without readable results simply report no tokens.
				_ = common.Unmarshal(task.Data, &data)
			}
			item := TaskUsage{TotalTokens: data.Usage.TotalTokens}
			if item.TotalTokens == 0 {
				item.TotalTokens = data.Usage.CompletionTokens
			}
			if task.FinishTime > task.SubmitTime && task.SubmitTime > 0 {
				item.Seconds = int(task.FinishTime - task.SubmitTime)
			}
			usage[task.TaskID] = item
		}
	}
	return usage, nil
}

// GetTaskIDsBySubmitTimes finds the task behind each charge time for the user's
// charges whose usage log recorded no task ID, which the older video logs did
// not. A task is submitted when it is charged, so a time resolves to the task
// the user submitted within a second of it, and only when that is a single one:
// with two candidates neither is guessed.
func GetTaskIDsBySubmitTimes(userID int, times []int64) (map[int64]string, error) {
	taskIDs := make(map[int64]string, len(times))
	unique := lo.Filter(lo.Uniq(times), func(at int64, _ int) bool { return at > 0 })
	for _, batch := range lo.Chunk(unique, 150) {
		window := make([]int64, 0, len(batch)*3)
		for _, at := range batch {
			window = append(window, at-1, at, at+1)
		}
		var tasks []Task
		if err := DB.Select("task_id", "submit_time").Where("user_id = ? AND submit_time IN ?", userID, window).Find(&tasks).Error; err != nil {
			return nil, err
		}
		for _, at := range batch {
			var candidate string
			count := 0
			for _, task := range tasks {
				if task.SubmitTime >= at-1 && task.SubmitTime <= at+1 {
					candidate = task.TaskID
					count++
				}
			}
			if count == 1 {
				taskIDs[at] = candidate
			}
		}
	}
	return taskIDs, nil
}
