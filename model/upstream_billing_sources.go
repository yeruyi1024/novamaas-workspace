package model

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/shopspring/decimal"
)

type upstreamBillingLogMetadata struct {
	IsTask            bool   `json:"is_task"`
	TaskID            string `json:"task_id"`
	PreConsumedQuota  *int64 `json:"pre_consumed_quota"`
	UpstreamModelName string `json:"upstream_model_name"`
	CacheRead         int64  `json:"cache_tokens"`
	CacheWrite        int64  `json:"cache_creation_tokens"`
	CacheWrite5m      int64  `json:"cache_creation_tokens_5m"`
	CacheWrite1h      int64  `json:"cache_creation_tokens_1h"`
	TotalTokens       *int64 `json:"total_tokens"`
	Admin             struct {
		UsagePath  string `json:"usage_billing_path"`
		LocalCount bool   `json:"local_count_tokens"`
	} `json:"admin_info"`
}

type upstreamBillingTaskUsage struct {
	TotalTokens      *int64   `json:"total_tokens"`
	CompletionTokens *int64   `json:"completion_tokens"`
	Duration         *float64 `json:"duration"`
}

type upstreamBillingTaskData struct {
	Usage    *upstreamBillingTaskUsage `json:"usage"`
	Duration *int64                    `json:"duration"`
	Data     *struct {
		Usage    *upstreamBillingTaskUsage `json:"usage"`
		Duration *int64                    `json:"duration"`
	} `json:"data"`
}

// Source pages read the log database and main database separately. Only safe,
// normalized evidence reaches the visitor; SQL never joins across databases.
// Corrections move customer groups, while the original supplier stays fixed.
func VisitUpstreamBillingSourceRecords(ctx context.Context, channelIDs []int, start, end int64, visit func(UpstreamBillingStatementRecord) error) error {
	if len(channelIDs) == 0 || start < 0 || end <= start || end-start > 366*86400 {
		return errors.New("invalid upstream billing source range")
	}
	filter := CostAccountingFilter{StartTimestamp: start, EndTimestamp: end - 1, channelIDs: channelIDs}
	var snapshotMax, adjustmentMax int64
	if err := DB.WithContext(ctx).Model(&CostAccountingSnapshot{}).Select("COALESCE(MAX(id), 0)").Scan(&snapshotMax).Error; err != nil {
		return err
	}
	if err := DB.WithContext(ctx).Model(&CostAccountingAdjustment{}).Select("COALESCE(MAX(id), 0)").Scan(&adjustmentMax).Error; err != nil {
		return err
	}
	var logMax int
	if !common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		if err := LOG_DB.WithContext(ctx).Model(&Log{}).Select("COALESCE(MAX(id), 0)").Scan(&logMax).Error; err != nil {
			return err
		}
	}
	corrections, err := billingEffectiveLogs(ctx, filter)
	if err != nil {
		return err
	}
	groups := make(map[billingEffectiveLogKey]string, len(corrections))
	for _, correction := range corrections {
		groups[billingLogIdentity(&correction.Original)] = correction.Group
	}
	usedSnapshots := make(map[int64]bool)
	tasksSeen := make(map[string]bool)
	failedRefunds := make(map[string]bool)
	var cursor int
	var offset int
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := LOG_DB.WithContext(ctx).Model(&Log{}).
			Select("id", "user_id", "created_at", "type", "model_name", "quota", "prompt_tokens", "completion_tokens", "channel_id", "token_id", "request_id", "upstream_request_id", "group", "content", "other").
			Where("channel_id IN ? AND created_at >= ? AND created_at < ? AND type IN ?", channelIDs, start, end, []int{LogTypeConsume, LogTypeRefund})
		if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
			query = query.Order("created_at asc").Order("request_id asc").Order("user_id asc").Order("type asc").Offset(offset)
		} else {
			query = query.Where("id > ? AND id <= ?", cursor, logMax).Order("id asc")
		}
		var logs []Log
		if err := query.Limit(1000).Find(&logs).Error; err != nil {
			return err
		}
		if len(logs) == 0 {
			break
		}
		views, err := upstreamBillingCostEvidence(ctx, filter, logs, snapshotMax, adjustmentMax)
		if err != nil {
			return err
		}
		byID := make(map[int64][]CostAccountingSnapshotView, len(views))
		byRequest := make(map[string][]CostAccountingSnapshotView, len(views))
		for _, view := range views {
			if view.SourceLogID > 0 {
				byID[view.SourceLogID] = append(byID[view.SourceLogID], view)
			}
			if view.RequestID != "" {
				byRequest[view.RequestID] = append(byRequest[view.RequestID], view)
			}
		}
		usageLogs, taskLinks, err := upstreamBillingResolveTaskLogs(ctx, logs)
		if err != nil {
			return err
		}
		taskRows, err := upstreamBillingTasks(ctx, usageLogs, taskLinks)
		if err != nil {
			return err
		}
		if err := upstreamBillingFailedTaskRefunds(ctx, taskRows, filter, snapshotMax, adjustmentMax, logMax, failedRefunds); err != nil {
			return err
		}
		for i := range logs {
			log := &logs[i]
			record := UpstreamBillingStatementRecord{SourceLogID: log.Id, UserID: log.UserId, OccurredAt: log.CreatedAt, LogType: log.Type, ChannelID: log.ChannelId, ModelName: log.ModelName, BillingGroup: log.Group, RequestID: log.RequestId, UpstreamRequestID: log.UpstreamRequestId, CostMissing: true}
			if group, found := groups[billingLogIdentity(log)]; found {
				record.BillingGroup = group
			}
			candidates := append([]CostAccountingSnapshotView(nil), byID[int64(log.Id)]...)
			candidates = append(candidates, byRequest[log.RequestId]...)
			for _, view := range candidates {
				if usedSnapshots[view.ID] {
					continue
				}
				revenue := int64(log.Quota)
				if log.Type == LogTypeRefund {
					revenue = -revenue
				}
				log.RevenueQuota = &revenue
				if !matchesLogAccountingSnapshot(log, view) {
					continue
				}
				usedSnapshots[view.ID] = true
				record.CostSnapshotID, record.LatestAdjustmentID = view.ID, view.LatestAdjustmentID
				record.CostDiscount, record.CostBasisQuota = view.EffectiveCostDiscount, view.EffectiveCostBasisQuota
				if record.CostDiscount != "" {
					if _, err := NormalizeCostDiscount(record.CostDiscount); err != nil {
						return err
					}
					record.CostQuota, record.CostMissing = view.EffectiveCostQuota, false
					if (log.Type == LogTypeConsume && record.CostQuota < 0) || (log.Type == LogTypeRefund && record.CostQuota > 0) {
						return ErrBillingEvidenceIntegrity
					}
				}
				break
			}
			populateUpstreamBillingUsage(&usageLogs[i], &record, taskRows, tasksSeen, failedRefunds)
			if _, linked := taskLinks[log.Id]; linked {
				switch record.UsageSource {
				case "task_final_usage":
					record.UsageSource = "task_historical_final_usage"
				case "task_failed_refunded":
					record.UsageSource = "task_historical_failed_refunded"
				}
			}
			if err := visit(record); err != nil {
				return err
			}
		}
		cursor, offset = logs[len(logs)-1].Id, offset+len(logs)
		if len(logs) < 1000 {
			break
		}
	}
	// A cost fact whose log is missing must remain visible as incomplete evidence,
	// never disappear from the cost total or acquire invented zero usage.
	var after int64
	for {
		query, err := costAccountingQuery(filter)
		if err != nil {
			return err
		}
		var views []CostAccountingSnapshotView
		if err := query.WithContext(ctx).Where("snapshots.id > ? AND snapshots.id <= ? AND snapshots.log_type IN ?", after, snapshotMax, []int{LogTypeConsume, LogTypeRefund}).Select("snapshots.*").Order("snapshots.id asc").Limit(1000).Scan(&views).Error; err != nil {
			return err
		}
		if len(views) == 0 {
			break
		}
		if err := attachCostAccountingBasis(DB.WithContext(ctx).Where("id <= ?", adjustmentMax), views); err != nil {
			return err
		}
		for _, view := range views {
			if usedSnapshots[view.ID] {
				continue
			}
			record := UpstreamBillingStatementRecord{SourceLogID: int(view.SourceLogID), CostSnapshotID: view.ID, LatestAdjustmentID: view.LatestAdjustmentID, OccurredAt: view.OccurredAt, LogType: view.LogType, UserID: view.UserID, ChannelID: view.ChannelID, ModelName: view.ModelName, BillingGroup: view.GroupName, RequestID: view.RequestID, UpstreamRequestID: view.UpstreamRequestID, CostDiscount: view.EffectiveCostDiscount, CostBasisQuota: view.EffectiveCostBasisQuota, UsageMissing: true, UsageSource: "missing_log"}
			record.CostMissing = strings.TrimSpace(view.EffectiveCostDiscount) == ""
			if !record.CostMissing {
				record.CostQuota = view.EffectiveCostQuota
			}
			if err := visit(record); err != nil {
				return err
			}
		}
		after = views[len(views)-1].ID
		if len(views) < 1000 {
			break
		}
	}
	return nil
}

func upstreamBillingCostEvidence(ctx context.Context, filter CostAccountingFilter, logs []Log, snapshotMax, adjustmentMax int64) ([]CostAccountingSnapshotView, error) {
	ids, requestIDs := make([]int64, 0, len(logs)), make([]string, 0, len(logs))
	for _, log := range logs {
		if log.Id > 0 && !common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
			ids = append(ids, int64(log.Id))
		}
		if log.RequestId != "" {
			requestIDs = append(requestIDs, log.RequestId)
		}
	}
	views := make([]CostAccountingSnapshotView, 0)
	if len(ids) == 0 && len(requestIDs) == 0 {
		return views, nil
	}
	query, err := costAccountingQuery(filter)
	if err != nil {
		return nil, err
	}
	query = query.Where("snapshots.id <= ?", snapshotMax)
	if len(ids) > 0 && len(requestIDs) > 0 {
		query = query.Where("snapshots.source_log_id IN ? OR snapshots.request_id IN ?", ids, requestIDs)
	} else if len(ids) > 0 {
		query = query.Where("snapshots.source_log_id IN ?", ids)
	} else {
		query = query.Where("snapshots.request_id IN ?", requestIDs)
	}
	if err := query.WithContext(ctx).Select("snapshots.*").Order("snapshots.id asc").Scan(&views).Error; err != nil {
		return nil, err
	}
	err = attachCostAccountingBasis(DB.WithContext(ctx).Where("id <= ?", adjustmentMax), views)
	return views, err
}

func upstreamBillingTasks(ctx context.Context, logs []Log, links map[int]UpstreamBillingTaskLink) (map[string][]Task, error) {
	ids, users := make(map[string]bool), make(map[int]bool)
	for _, log := range logs {
		var metadata upstreamBillingLogMetadata
		if common.UnmarshalJsonStr(log.Other, &metadata) == nil && metadata.TaskID != "" {
			ids[metadata.TaskID], users[log.UserId] = true, true
		}
	}
	result := make(map[string][]Task)
	if len(ids) == 0 {
		return result, nil
	}
	taskIDs, userIDs := make([]string, 0, len(ids)), make([]int, 0, len(users))
	for id := range ids {
		taskIDs = append(taskIDs, id)
	}
	for id := range users {
		userIDs = append(userIDs, id)
	}
	sort.Strings(taskIDs)
	var tasks []Task
	if err := DB.WithContext(ctx).Select("id", "task_id", "user_id", "channel_id", "platform", "status", "submit_time", "finish_time", "data").Where("task_id IN ? AND user_id IN ?", taskIDs, userIDs).Find(&tasks).Error; err != nil {
		return nil, err
	}
	// The resolver and this usage query are separate database reads. Use only
	// the terminal version whose bytes were approved for the historical link.
	for _, link := range links {
		var proof upstreamBillingTaskLinkEvidence
		if err := common.Unmarshal([]byte(link.Evidence), &proof); err != nil {
			return nil, err
		}
		found := false
		for i := range tasks {
			task := &tasks[i]
			if task.ID != link.TaskRowID {
				continue
			}
			candidate := proof.Candidate
			if task.TaskID != link.TaskID || task.UserId != link.UserID || task.ChannelId != link.ChannelID || string(task.Status) != candidate.Status || string(task.Platform) != proof.Platform || task.SubmitTime != candidate.SubmitTime || upstreamBillingTaskLinkHash(task.Data) != candidate.DataSHA256 {
				return nil, ErrBillingEvidenceIntegrity
			}
			found = true
		}
		if !found {
			return nil, ErrBillingEvidenceIntegrity
		}
	}
	for _, task := range tasks {
		key := fmtUpstreamTaskIdentity(task.UserId, task.ChannelId, task.TaskID)
		result[key] = append(result[key], task)
	}
	return result, nil
}

// Failure is not evidence of a free supplier request. Prove the original
// consumption and refund costs independently, within this statement's period
// and source watermarks. Read all candidate failures together so a refund on a
// later source page is available without rescanning the month per task.
func upstreamBillingFailedTaskRefunds(ctx context.Context, tasks map[string][]Task, filter CostAccountingFilter, snapshotMax, adjustmentMax int64, logMax int, cache map[string]bool) error {
	type refundProof struct {
		complete, consumed, refunded bool
		cost, revenue                decimal.Decimal
	}
	proofs := make(map[string]*refundProof)
	users, channels := make(map[int]bool), make(map[int]bool)
	for key, candidates := range tasks {
		if _, seen := cache[key]; seen || len(candidates) != 1 || candidates[0].Status != TaskStatusFailure {
			continue
		}
		cache[key] = false
		proofs[key] = &refundProof{complete: true}
		users[candidates[0].UserId], channels[candidates[0].ChannelId] = true, true
	}
	if len(proofs) == 0 {
		return nil
	}
	userIDs, channelIDs := make([]int, 0, len(users)), make([]int, 0, len(channels))
	for id := range users {
		userIDs = append(userIDs, id)
	}
	for id := range channels {
		channelIDs = append(channelIDs, id)
	}
	usedSnapshots := make(map[int64]bool)
	var cursor, offset int
	for {
		query := LOG_DB.WithContext(ctx).Model(&Log{}).
			Select("id", "user_id", "created_at", "type", "model_name", "quota", "channel_id", "token_id", "group", "content", "request_id", "upstream_request_id", "other").
			Where("user_id IN ? AND channel_id IN ? AND created_at >= ? AND created_at <= ? AND type IN ?", userIDs, channelIDs, filter.StartTimestamp, filter.EndTimestamp, []int{LogTypeConsume, LogTypeRefund})
		if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
			query = query.Order("created_at asc").Order("request_id asc").Order("user_id asc").Order("type asc").Offset(offset)
		} else {
			query = query.Where("id > ? AND id <= ?", cursor, logMax).Order("id asc")
		}
		var logs []Log
		if err := query.Limit(1000).Find(&logs).Error; err != nil {
			return err
		}
		if len(logs) == 0 {
			break
		}
		matched := make([]Log, 0)
		keys := make([]string, 0)
		usageLogs, _, err := upstreamBillingResolveTaskLogs(ctx, logs)
		if err != nil {
			return err
		}
		for _, log := range usageLogs {
			var metadata upstreamBillingLogMetadata
			if common.UnmarshalJsonStr(log.Other, &metadata) != nil || metadata.TaskID == "" {
				continue
			}
			key := fmtUpstreamTaskIdentity(log.UserId, log.ChannelId, metadata.TaskID)
			if proofs[key] != nil {
				matched, keys = append(matched, log), append(keys, key)
			}
		}
		views, err := upstreamBillingCostEvidence(ctx, filter, matched, snapshotMax, adjustmentMax)
		if err != nil {
			return err
		}
		byID, byRequest := make(map[int64][]CostAccountingSnapshotView), make(map[string][]CostAccountingSnapshotView)
		for _, view := range views {
			if view.SourceLogID > 0 {
				byID[view.SourceLogID] = append(byID[view.SourceLogID], view)
			}
			if view.RequestID != "" {
				byRequest[view.RequestID] = append(byRequest[view.RequestID], view)
			}
		}
		for index := range matched {
			log, proof := &matched[index], proofs[keys[index]]
			if log.Quota < 0 || log.Quota > common.MaxQuota {
				proof.complete = false
				continue
			}
			revenue := int64(log.Quota)
			if log.Type == LogTypeRefund {
				revenue = -revenue
				proof.refunded = true
			} else {
				proof.consumed = true
			}
			log.RevenueQuota = &revenue
			candidates := append([]CostAccountingSnapshotView(nil), byID[int64(log.Id)]...)
			candidates = append(candidates, byRequest[log.RequestId]...)
			matches := make(map[int64]CostAccountingSnapshotView)
			for _, candidate := range candidates {
				if matchesLogAccountingSnapshot(log, candidate) {
					matches[candidate.ID] = candidate
				}
			}
			if len(matches) != 1 {
				proof.complete = false
				continue
			}
			for id, view := range matches {
				discount, err := NormalizeCostDiscount(view.EffectiveCostDiscount)
				cost := view.EffectiveCostQuota
				if err != nil || discount == "" || usedSnapshots[id] || cost < -int64(common.MaxQuota) || cost > int64(common.MaxQuota) || (log.Type == LogTypeConsume && cost < 0) || (log.Type == LogTypeRefund && cost > 0) {
					proof.complete = false
					continue
				}
				usedSnapshots[id] = true
				proof.cost = proof.cost.Add(decimal.NewFromInt(cost))
				proof.revenue = proof.revenue.Add(decimal.NewFromInt(revenue))
			}
		}
		cursor, offset = logs[len(logs)-1].Id, offset+len(logs)
		if len(logs) < 1000 {
			break
		}
	}
	for key, proof := range proofs {
		cache[key] = proof.complete && proof.consumed && proof.refunded && proof.cost.IsZero() && proof.revenue.IsZero()
	}
	return nil
}

func populateUpstreamBillingUsage(log *Log, record *UpstreamBillingStatementRecord, taskRows map[string][]Task, tasksSeen, failedRefunds map[string]bool) {
	var metadata upstreamBillingLogMetadata
	if log.Other != "" && common.UnmarshalJsonStr(log.Other, &metadata) != nil {
		record.UsageMissing, record.UsageSource = true, "invalid_metadata"
		return
	}
	record.UpstreamModelName, record.TaskID = metadata.UpstreamModelName, metadata.TaskID
	if metadata.IsTask || metadata.TaskID != "" {
		record.UsageSource = "task_final_usage"
		key := fmtUpstreamTaskIdentity(log.UserId, log.ChannelId, metadata.TaskID)
		if log.Type == LogTypeRefund || metadata.PreConsumedQuota != nil {
			return
		}
		if metadata.TaskID == "" {
			// There is no identity with which to deduplicate these submissions.
			// Each original log remains a request and an unresolved usage gap.
			record.Requests, record.UsageMissing, record.UsageSource = 1, true, "task_missing_id"
			return
		}
		if tasksSeen[key] {
			return
		}
		tasksSeen[key] = true
		record.Requests = 1
		tasks := taskRows[key]
		if len(tasks) == 0 {
			record.UsageMissing, record.UsageSource = true, "task_not_found"
			return
		}
		if len(tasks) != 1 {
			record.UsageMissing, record.UsageSource = true, "task_ambiguous"
			return
		}
		task := tasks[0]
		if task.Status != TaskStatusSuccess && task.Status != TaskStatusFailure {
			record.TaskPending, record.UsageMissing, record.UsageSource = true, true, "task_pending"
			return
		}
		var data upstreamBillingTaskData
		if common.Unmarshal(task.Data, &data) != nil {
			record.UsageMissing, record.UsageSource = true, "task_invalid_usage"
			return
		}
		usage, duration := data.Usage, data.Duration
		if data.Data != nil {
			if usage == nil {
				usage = data.Data.Usage
			}
			if duration == nil {
				duration = data.Data.Duration
			}
		}
		var tokens *int64
		if usage != nil {
			tokens = usage.TotalTokens
			if tokens == nil {
				tokens = usage.CompletionTokens
			}
		}
		if tokens != nil && (*tokens < 0 || *tokens > int64(common.MaxQuota)) {
			record.UsageMissing, record.UsageSource = true, "task_invalid_usage"
			return
		}
		if duration != nil && (*duration < 0 || *duration > relaycommon.MaxTaskDurationSeconds) {
			record.UsageMissing, record.UsageSource = true, "task_invalid_usage"
			return
		}
		if task.Status == TaskStatusFailure && tokens == nil && failedRefunds[key] && !record.CostMissing {
			if usage != nil && usage.Duration != nil {
				seconds := *usage.Duration
				if seconds <= 0 || seconds > relaycommon.MaxTaskDurationSeconds || math.Trunc(seconds) != seconds {
					record.UsageMissing, record.UsageSource = true, "task_invalid_usage"
					return
				}
			}
			record.UsageSource = "task_failed_refunded"
			return
		}
		// Ali WAN reports measured final seconds under usage, rather than tokens.
		// Only the persisted provider identity and a successful final response
		// authorize this unit. Request duration and per-call pricing do not.
		if task.Status == TaskStatusSuccess && task.Platform == constant.TaskPlatform(strconv.Itoa(constant.ChannelTypeAli)) && usage != nil && usage.Duration != nil {
			seconds := *usage.Duration
			if seconds <= 0 || seconds > relaycommon.MaxTaskDurationSeconds || math.Trunc(seconds) != seconds {
				record.UsageMissing, record.UsageSource = true, "task_invalid_usage"
				return
			}
			record.VideoSeconds, record.UsageSource = int64(seconds), "task_duration_usage"
			return
		}
		// Historical billing contexts do not identify a reliable token-versus-
		// seconds unit. Preserve reported seconds, but never let duration stand
		// in for absent final token evidence on an upstream statement.
		if tokens == nil {
			record.UsageMissing, record.UsageSource = true, "task_usage_missing"
		}
		if tokens != nil {
			record.VideoTokens = *tokens
		}
		if duration != nil {
			record.VideoSeconds = *duration
		}
		return
	}
	if log.Type == LogTypeRefund {
		record.UsageSource = "refund_adjustment"
		return
	}
	record.Requests = 1
	values := []int64{int64(log.PromptTokens), int64(log.CompletionTokens), metadata.CacheRead, metadata.CacheWrite, metadata.CacheWrite5m, metadata.CacheWrite1h}
	for _, value := range values {
		if value < 0 || value > int64(common.MaxQuota) {
			record.UsageMissing, record.UsageSource = true, "invalid_usage"
			return
		}
	}
	record.InputTokens, record.OutputTokens = int64(log.PromptTokens), int64(log.CompletionTokens)
	record.CacheReadTokens, record.CacheWriteTokens = metadata.CacheRead, metadata.CacheWrite
	record.CacheWrite5mTokens, record.CacheWrite1hTokens = metadata.CacheWrite5m, metadata.CacheWrite1h
	if split := metadata.CacheWrite5m + metadata.CacheWrite1h; split > record.CacheWriteTokens {
		record.CacheWriteTokens = split
	}
	record.UsageSource = metadata.Admin.UsagePath
	if record.UsageSource == "" {
		record.UsageSource = "usage_logs"
	}
	record.UsageEstimated = metadata.Admin.LocalCount || metadata.Admin.UsagePath == "local" || strings.Contains(metadata.Admin.UsagePath, "estimated")
	if metadata.Admin.UsagePath == "" && log.PromptTokens == 0 && log.CompletionTokens == 0 {
		record.UsageMissing = true
	}
}

func fmtUpstreamTaskIdentity(userID, channelID int, taskID string) string {
	return fmt.Sprintf("%d:%d:%s", userID, channelID, taskID)
}
