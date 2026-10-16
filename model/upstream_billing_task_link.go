package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const UpstreamBillingTaskLinkRule = "compound_task_link_exact_second_v1"

// A reviewed historical association is separate from original request archives
// and financial facts. Neither the source log nor the task is rewritten.
type UpstreamBillingTaskLink struct {
	ID             int64                       `json:"id" gorm:"primaryKey"`
	SourceLogID    int                         `json:"source_log_id" gorm:"uniqueIndex"`
	TaskRowID      int64                       `json:"task_row_id" gorm:"uniqueIndex"`
	TaskID         string                      `json:"task_id" gorm:"type:varchar(191)"`
	UserID         int                         `json:"user_id" gorm:"index"`
	ChannelID      int                         `json:"channel_id" gorm:"index"`
	Algorithm      string                      `json:"algorithm" gorm:"type:varchar(64)"`
	SourceSHA256   string                      `json:"source_sha256" gorm:"type:char(64)"`
	EvidenceSHA256 string                      `json:"evidence_sha256" gorm:"type:char(64)"`
	Evidence       UpstreamBillingSnapshotData `json:"-"`
	ReviewSHA256   string                      `json:"review_sha256" gorm:"type:char(64)"`
	BatchID        string                      `json:"batch_id" gorm:"type:varchar(64);index"`
	CreatedBy      int                         `json:"created_by"`
	CreatedAt      int64                       `json:"created_at" gorm:"bigint"`
}

func (*UpstreamBillingTaskLink) BeforeUpdate(*gorm.DB) error { return ErrBillingEvidenceImmutable }
func (*UpstreamBillingTaskLink) BeforeDelete(*gorm.DB) error { return ErrBillingEvidenceImmutable }

type UpstreamBillingTaskLinkManifest struct {
	SchemaVersion int                                `json:"schema_version"`
	Rule          string                             `json:"rule"`
	Candidates    []UpstreamBillingTaskLinkCandidate `json:"candidates"`
}

type UpstreamBillingTaskLinkPrice struct {
	Log   *float64 `json:"log"`
	Task  *float64 `json:"task"`
	Equal bool     `json:"equal"`
}

type UpstreamBillingTaskLinkAdjustment struct {
	LogID             int    `json:"log_id"`
	Type              int    `json:"type"`
	SignedDelta       int64  `json:"signed_delta"`
	OccurredAt        int64  `json:"occurred_at"`
	PreConsumedQuota  *int64 `json:"pre_consumed_quota,omitempty"`
	ActualQuota       *int64 `json:"actual_quota,omitempty"`
	FullFailureRefund bool   `json:"full_failure_refund,omitempty"`
}

type UpstreamBillingTaskLinkCandidate struct {
	LogID             int                                     `json:"log_id"`
	TaskRowID         int64                                   `json:"task_row_id"`
	TaskID            string                                  `json:"task_id"`
	UpstreamTaskID    string                                  `json:"upstream_task_id"`
	Decision          string                                  `json:"decision"`
	Status            string                                  `json:"status"`
	SubmitTime        int64                                   `json:"submit_time"`
	InitialQuota      int64                                   `json:"initial_quota_from_final_and_linked_events"`
	DataSHA256        string                                  `json:"data_sha256"`
	PropertiesSHA256  string                                  `json:"properties_sha256"`
	ContentSHA256     string                                  `json:"content_sha256"`
	UserID            int                                     `json:"user_id"`
	ChannelID         int                                     `json:"channel_id"`
	TokenID           int                                     `json:"token_id"`
	Model             string                                  `json:"model"`
	Group             string                                  `json:"group"`
	CreatedAt         int64                                   `json:"created_at"`
	Quota             int64                                   `json:"quota"`
	RequestID         string                                  `json:"request_id"`
	UpstreamRequestID string                                  `json:"upstream_request_id"`
	AdjustmentChain   []UpstreamBillingTaskLinkAdjustment     `json:"adjustment_chain"`
	PricingChecks     map[string]UpstreamBillingTaskLinkPrice `json:"pricing_checks"`
}

type upstreamBillingTaskLinkEvent struct {
	UpstreamBillingTaskLinkAdjustment
	OtherSHA256 string `json:"other_sha256"`
}

type upstreamBillingTaskLinkEvidence struct {
	SchemaVersion int                              `json:"schema_version"`
	Algorithm     string                           `json:"algorithm"`
	SourceSHA256  string                           `json:"source_sha256"`
	Candidate     UpstreamBillingTaskLinkCandidate `json:"reviewed_candidate"`
	Events        []upstreamBillingTaskLinkEvent   `json:"events"`
	Platform      string                           `json:"platform"`
	FinalQuota    int                              `json:"final_quota"`
}

func upstreamBillingTaskLinkHash(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// The fingerprint includes exact original metadata without duplicating prompts,
// credentials, or request/response bodies in the association evidence.
func UpstreamBillingTaskLinkSourceFingerprint(log *Log) (string, error) {
	body, err := common.Marshal(struct {
		ID, UserID, ChannelID, Type, TokenID                                   int
		CreatedAt                                                              int64
		Quota                                                                  int
		Model, Group, RequestID, UpstreamRequestID, ContentSHA256, OtherSHA256 string
	}{log.Id, log.UserId, log.ChannelId, log.Type, log.TokenId, log.CreatedAt, log.Quota, log.ModelName, log.Group, log.RequestId, log.UpstreamRequestId, upstreamBillingTaskLinkHash([]byte(log.Content)), upstreamBillingTaskLinkHash([]byte(log.Other))})
	if err != nil {
		return "", err
	}
	return upstreamBillingTaskLinkHash(body), nil
}

// ApplyUpstreamBillingTaskLinks consumes the exact reviewed file digest. All
// database evidence and global uniqueness are checked again inside one task-
// locking transaction; manifest assertions alone never authorize a link.
func ApplyUpstreamBillingTaskLinks(ctx context.Context, manifest []byte, expectedSHA256, batchID string, actorID int) ([]UpstreamBillingTaskLink, error) {
	if actorID < 0 || batchID == "" || len(batchID) > 64 || len(expectedSHA256) != 64 || upstreamBillingTaskLinkHash(manifest) != expectedSHA256 {
		return nil, ErrBillingConflict
	}
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		return nil, ErrBillingEvidenceIntegrity
	}
	var review UpstreamBillingTaskLinkManifest
	if err := common.Unmarshal(manifest, &review); err != nil {
		return nil, err
	}
	if review.SchemaVersion != 1 || review.Rule != UpstreamBillingTaskLinkRule || len(review.Candidates) == 0 || len(review.Candidates) > 1000 {
		return nil, ErrBillingEvidenceIntegrity
	}
	selected := make([]UpstreamBillingTaskLinkCandidate, 0)
	users, channels, taskIDs := map[int]bool{}, map[int]bool{}, map[int64]bool{}
	sourceIDs := map[int]bool{}
	for _, candidate := range review.Candidates {
		if candidate.Decision == "keep_missing_task_id" {
			continue
		}
		if candidate.Decision != "review_compound_exact" || candidate.LogID <= 0 || candidate.TaskRowID <= 0 || candidate.TaskID == "" || candidate.UserID <= 0 || candidate.ChannelID <= 0 || sourceIDs[candidate.LogID] || taskIDs[candidate.TaskRowID] {
			return nil, ErrBillingEvidenceIntegrity
		}
		sourceIDs[candidate.LogID], taskIDs[candidate.TaskRowID] = true, true
		users[candidate.UserID], channels[candidate.ChannelID] = true, true
		selected = append(selected, candidate)
	}
	if len(selected) == 0 {
		return nil, ErrBillingEvidenceIntegrity
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].TaskRowID < selected[j].TaskRowID })
	userIDs, channelIDs, rowIDs := []int{}, []int{}, []int64{}
	for id := range users {
		userIDs = append(userIDs, id)
	}
	for id := range channels {
		channelIDs = append(channelIDs, id)
	}
	for id := range taskIDs {
		rowIDs = append(rowIDs, id)
	}
	links := make([]UpstreamBillingTaskLink, 0, len(selected))
	err := DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		logDB := LOG_DB.WithContext(ctx)
		if DB == LOG_DB {
			logDB = tx
		}
		var locked []Task
		if err := lockForUpdate(tx).Where("id IN ?", rowIDs).Order("id asc").Find(&locked).Error; err != nil {
			return err
		}
		if len(locked) != len(selected) {
			return ErrBillingConflict
		}
		// Read the full ownership scope, including explicitly linked submissions,
		// so uniqueness is established globally rather than only within the plan.
		var logs []Log
		if err := logDB.Select("id", "user_id", "channel_id", "created_at", "type", "quota", "token_id", "group", "model_name", "request_id", "upstream_request_id", "content", "other").Where("user_id IN ? AND channel_id IN ? AND type IN ?", userIDs, channelIDs, []int{LogTypeConsume, LogTypeRefund}).Order("id asc").Limit(100001).Find(&logs).Error; err != nil {
			return err
		}
		if len(logs) > 100000 {
			return ErrBillingEvidenceIntegrity
		}
		byLog := map[int]*Log{}
		for i := range logs {
			byLog[logs[i].Id] = &logs[i]
		}
		submitTimes := []int64{}
		for _, candidate := range selected {
			submitTimes = append(submitTimes, candidate.CreatedAt)
		}
		var candidates []Task
		if err := tx.Where("user_id IN ? AND channel_id IN ? AND submit_time IN ?", userIDs, channelIDs, submitTimes).Find(&candidates).Error; err != nil {
			return err
		}
		byTask := map[int64]*Task{}
		for i := range candidates {
			byTask[candidates[i].ID] = &candidates[i]
		}
		var rawProperties []struct {
			ID         int64
			Properties []byte
		}
		if err := tx.Table("tasks").Select("id", "properties").Where("id IN ?", rowIDs).Find(&rawProperties).Error; err != nil {
			return err
		}
		properties := map[int64][]byte{}
		for _, raw := range rawProperties {
			properties[raw.ID] = raw.Properties
		}
		for _, candidate := range selected {
			log, task := byLog[candidate.LogID], byTask[candidate.TaskRowID]
			if log == nil || task == nil {
				return fmt.Errorf("historical task link log %d: missing source or task: %w", candidate.LogID, ErrBillingConflict)
			}
			var metadata upstreamBillingLogMetadata
			if common.UnmarshalJsonStr(log.Other, &metadata) != nil || !metadata.IsTask || metadata.TaskID != "" || metadata.PreConsumedQuota != nil || log.Type != LogTypeConsume {
				return fmt.Errorf("historical task link log %d: original submission metadata: %w", candidate.LogID, ErrBillingEvidenceIntegrity)
			}
			if log.UserId != candidate.UserID || log.ChannelId != candidate.ChannelID || log.TokenId != candidate.TokenID || log.ModelName != candidate.Model || log.Group != candidate.Group || log.CreatedAt != candidate.CreatedAt || int64(log.Quota) != candidate.Quota || log.RequestId != candidate.RequestID || log.UpstreamRequestId != candidate.UpstreamRequestID || upstreamBillingTaskLinkHash([]byte(log.Content)) != candidate.ContentSHA256 {
				return fmt.Errorf("historical task link log %d: original submission identity: %w", candidate.LogID, ErrBillingEvidenceIntegrity)
			}
			if task.TaskID != candidate.TaskID || task.GetUpstreamTaskID() != candidate.UpstreamTaskID || string(task.Status) != candidate.Status || task.SubmitTime != candidate.SubmitTime || task.SubmitTime != log.CreatedAt || task.Status != TaskStatusSuccess && task.Status != TaskStatusFailure || upstreamBillingTaskLinkHash(task.Data) != candidate.DataSHA256 || upstreamBillingTaskLinkHash(properties[task.ID]) != candidate.PropertiesSHA256 {
				return fmt.Errorf("historical task link log %d: reviewed terminal task: %w", candidate.LogID, ErrBillingEvidenceIntegrity)
			}
			var response struct {
				ID string `json:"id"`
			}
			if common.Unmarshal(task.Data, &response) != nil || task.PrivateData.UpstreamTaskID == "" || response.ID != task.PrivateData.UpstreamTaskID {
				return fmt.Errorf("historical task link log %d: explicit upstream identity: %w", candidate.LogID, ErrBillingEvidenceIntegrity)
			}
			if !upstreamBillingTaskLinkDimensions(log, task) || !upstreamBillingTaskLinkPrices(log, task, candidate.PricingChecks) {
				return fmt.Errorf("historical task link log %d: original billing dimensions: %w", candidate.LogID, ErrBillingEvidenceIntegrity)
			}
			initial, events, known, valid := upstreamBillingTaskLinkChain(task, logs)
			if !valid || initial == nil || *initial != candidate.InitialQuota || *initial != int64(log.Quota) || known != 0 || len(events) == 0 || len(events) != len(candidate.AdjustmentChain) {
				return fmt.Errorf("historical task link log %d: independent settlement chain: %w", candidate.LogID, ErrBillingEvidenceIntegrity)
			}
			for i, event := range events {
				left, err := common.Marshal(event.UpstreamBillingTaskLinkAdjustment)
				if err != nil {
					return err
				}
				right, err := common.Marshal(candidate.AdjustmentChain[i])
				if err != nil {
					return err
				}
				if string(left) != string(right) {
					return fmt.Errorf("historical task link log %d: reviewed settlement event %d: %w", candidate.LogID, event.LogID, ErrBillingEvidenceIntegrity)
				}
			}
			matches := 0
			for i := range candidates {
				other := &candidates[i]
				if !upstreamBillingTaskLinkDimensions(log, other) || !upstreamBillingTaskLinkPrices(log, other, nil) {
					continue
				}
				amount, _, _, _ := upstreamBillingTaskLinkChain(other, logs)
				if amount == nil || *amount == int64(log.Quota) {
					matches++
				}
			}
			if matches != 1 {
				return ErrBillingConflict
			}
			reverse := 0
			for i := range logs {
				original := &logs[i]
				if original.Type != LogTypeConsume || !upstreamBillingTaskLinkDimensions(original, task) || int64(original.Quota) != *initial || !upstreamBillingTaskLinkPrices(original, task, nil) {
					continue
				}
				var meta upstreamBillingLogMetadata
				if common.UnmarshalJsonStr(original.Other, &meta) != nil {
					return ErrBillingEvidenceIntegrity
				}
				if (meta.IsTask || meta.TaskID != "") && meta.PreConsumedQuota == nil {
					reverse++
				}
			}
			if reverse != 1 {
				return ErrBillingConflict
			}
			var sameID int64
			if err := tx.Model(&Task{}).Where("user_id = ? AND channel_id = ? AND task_id = ?", task.UserId, task.ChannelId, task.TaskID).Count(&sameID).Error; err != nil {
				return err
			}
			if sameID != 1 {
				return ErrBillingConflict
			}
			sourceHash, err := UpstreamBillingTaskLinkSourceFingerprint(log)
			if err != nil {
				return err
			}
			proof := upstreamBillingTaskLinkEvidence{SchemaVersion: 1, Algorithm: UpstreamBillingTaskLinkRule, SourceSHA256: sourceHash, Candidate: candidate, Events: events, Platform: string(task.Platform), FinalQuota: task.Quota}
			body, err := common.Marshal(proof)
			if err != nil {
				return err
			}
			link := UpstreamBillingTaskLink{SourceLogID: log.Id, TaskRowID: task.ID, TaskID: task.TaskID, UserID: log.UserId, ChannelID: log.ChannelId, Algorithm: UpstreamBillingTaskLinkRule, SourceSHA256: sourceHash, EvidenceSHA256: upstreamBillingTaskLinkHash(body), Evidence: UpstreamBillingSnapshotData(body), ReviewSHA256: expectedSHA256, BatchID: batchID, CreatedBy: actorID, CreatedAt: common.GetTimestamp()}
			var existing []UpstreamBillingTaskLink
			if err := tx.Where("source_log_id = ? OR task_row_id = ?", log.Id, task.ID).Find(&existing).Error; err != nil {
				return err
			}
			if len(existing) > 0 {
				if len(existing) != 1 || existing[0].SourceLogID != link.SourceLogID || existing[0].TaskRowID != link.TaskRowID || existing[0].ReviewSHA256 != link.ReviewSHA256 || existing[0].EvidenceSHA256 != link.EvidenceSHA256 || existing[0].BatchID != link.BatchID {
					return ErrBillingConflict
				}
				links = append(links, existing[0])
				continue
			}
			if err := tx.Create(&link).Error; err != nil {
				return err
			}
			links = append(links, link)
		}
		// Source logs can live in another database. Recheck their fingerprints
		// before publishing the batch; no distributed writes are attempted.
		var sourceLogs []Log
		ids := make([]int, 0, len(selected))
		for _, candidate := range selected {
			ids = append(ids, candidate.LogID)
		}
		if err := logDB.Select("id", "user_id", "channel_id", "created_at", "type", "quota", "token_id", "group", "model_name", "request_id", "upstream_request_id", "content", "other").Where("id IN ?", ids).Find(&sourceLogs).Error; err != nil {
			return err
		}
		if len(sourceLogs) != len(selected) {
			return ErrBillingConflict
		}
		for i := range sourceLogs {
			before, err := UpstreamBillingTaskLinkSourceFingerprint(byLog[sourceLogs[i].Id])
			if err != nil {
				return err
			}
			after, err := UpstreamBillingTaskLinkSourceFingerprint(&sourceLogs[i])
			if err != nil {
				return err
			}
			if before != after {
				return ErrBillingConflict
			}
		}
		return nil
	})
	return links, err
}

func upstreamBillingTaskLinkDimensions(log *Log, task *Task) bool {
	return log.UserId == task.UserId && log.ChannelId == task.ChannelId && log.TokenId == task.PrivateData.TokenId && log.ModelName == task.Properties.OriginModelName && log.Group == task.Group && log.CreatedAt == task.SubmitTime
}

func upstreamBillingTaskLinkPrices(log *Log, task *Task, expected map[string]UpstreamBillingTaskLinkPrice) bool {
	bc := task.PrivateData.BillingContext
	if bc == nil || bc.OriginModelName != log.ModelName {
		return false
	}
	var metadata map[string]any
	if common.UnmarshalJsonStr(log.Other, &metadata) != nil {
		return false
	}
	for key, value := range map[string]float64{"model_price": bc.ModelPrice, "model_ratio": bc.ModelRatio, "group_ratio": bc.GroupRatio} {
		raw, ok := metadata[key].(float64)
		if !ok || math.IsNaN(value) || math.IsInf(value, 0) || raw != value {
			return false
		}
		if expected != nil {
			review, present := expected[key]
			if !present || review.Log == nil || review.Task == nil || *review.Log != raw || *review.Task != value {
				return false
			}
		}
	}
	return true
}

// An initial amount comes from an independent explicit pre-consumption anchor
// or an explicit full failure refund, never from finalQuota minus summed deltas.
func upstreamBillingTaskLinkChain(task *Task, logs []Log) (*int64, []upstreamBillingTaskLinkEvent, int, bool) {
	events := []upstreamBillingTaskLinkEvent{}
	known := 0
	var initial *int64
	for i := range logs {
		log := &logs[i]
		if log.UserId != task.UserId || log.ChannelId != task.ChannelId {
			continue
		}
		var meta struct {
			IsTask json.RawMessage `json:"is_task"`
			TaskID string          `json:"task_id"`
			Pre    *int64          `json:"pre_consumed_quota"`
			Actual *int64          `json:"actual_quota"`
		}
		if common.UnmarshalJsonStr(log.Other, &meta) != nil {
			if strings.Contains(log.Other, task.TaskID) {
				return nil, nil, known, false
			}
			continue
		}
		if meta.TaskID != task.TaskID {
			continue
		}
		// Before 2026-10-09, settlement/refund writers omitted is_task.
		// An explicit task ID, full dimensions and independent amount anchors
		// still identify those events; a contradictory marker is never accepted.
		if len(meta.IsTask) > 0 {
			var markedTask bool
			if common.Unmarshal(meta.IsTask, &markedTask) != nil || !markedTask {
				return nil, nil, known, false
			}
		}
		if log.TokenId != task.PrivateData.TokenId || log.ModelName != task.Properties.OriginModelName || log.Group != task.Group || log.Quota < 0 || log.Quota > common.MaxQuota {
			return nil, nil, known, false
		}
		if log.Type == LogTypeConsume && meta.Pre == nil {
			known++
			amount := int64(log.Quota)
			if initial != nil && *initial != amount {
				return nil, nil, known, false
			}
			initial = &amount
			continue
		}
		// Polling persists FinishTime before it writes settlement/refund logs.
		// Delayed bookkeeping may therefore legitimately follow FinishTime.
		if log.CreatedAt < task.SubmitTime || task.FinishTime < task.SubmitTime || !upstreamBillingTaskLinkPrices(log, task, nil) {
			return nil, nil, known, false
		}
		delta := int64(log.Quota)
		if log.Type == LogTypeRefund {
			delta = -delta
		} else if log.Type != LogTypeConsume {
			return nil, nil, known, false
		}
		full := meta.Pre == nil && log.Type == LogTypeRefund && task.Status == TaskStatusFailure
		if initial == nil {
			if meta.Pre != nil {
				amount := *meta.Pre
				initial = &amount
			} else if full {
				amount := int64(log.Quota)
				initial = &amount
			} else {
				return nil, nil, known, false
			}
		}
		events = append(events, upstreamBillingTaskLinkEvent{UpstreamBillingTaskLinkAdjustment: UpstreamBillingTaskLinkAdjustment{LogID: log.Id, Type: log.Type, SignedDelta: delta, OccurredAt: log.CreatedAt, PreConsumedQuota: meta.Pre, ActualQuota: meta.Actual, FullFailureRefund: full}, OtherSHA256: upstreamBillingTaskLinkHash([]byte(log.Other))})
	}
	if initial == nil || *initial < 0 || *initial > int64(common.MaxQuota) || task.Quota < 0 || task.Quota > common.MaxQuota {
		return initial, events, known, false
	}
	current := *initial
	for _, event := range events {
		if event.PreConsumedQuota != nil {
			if event.ActualQuota == nil || *event.PreConsumedQuota != current || *event.ActualQuota < 0 || *event.ActualQuota > int64(common.MaxQuota) || *event.ActualQuota-*event.PreConsumedQuota != event.SignedDelta {
				return initial, events, known, false
			}
		} else if !event.FullFailureRefund || -event.SignedDelta != current {
			return initial, events, known, false
		}
		current += event.SignedDelta
	}
	return initial, events, known, current == int64(task.Quota)
}

// Compatibility reads on databases not yet migrated remain incomplete; they do
// not fabricate associations. The application migration creates the new table.
func upstreamBillingResolveTaskLogs(ctx context.Context, logs []Log) ([]Log, map[int]UpstreamBillingTaskLink, error) {
	ids := []int{}
	for _, log := range logs {
		var metadata upstreamBillingLogMetadata
		if common.UnmarshalJsonStr(log.Other, &metadata) == nil && metadata.IsTask && metadata.TaskID == "" && log.Type == LogTypeConsume && metadata.PreConsumedQuota == nil {
			ids = append(ids, log.Id)
		}
	}
	resolved := append([]Log(nil), logs...)
	links := map[int]UpstreamBillingTaskLink{}
	if len(ids) == 0 || common.UsingLogDatabase(common.DatabaseTypeClickHouse) || !DB.Migrator().HasTable(&UpstreamBillingTaskLink{}) {
		return resolved, links, nil
	}
	var rows []UpstreamBillingTaskLink
	if err := DB.WithContext(ctx).Where("source_log_id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	if len(rows) == 0 {
		return resolved, links, nil
	}
	taskIDs := []int64{}
	for _, link := range rows {
		links[link.SourceLogID] = link
		taskIDs = append(taskIDs, link.TaskRowID)
	}
	type taskState struct {
		ID                int64
		TaskID            string
		UserID, ChannelID int
		Platform, Status  string
		SubmitTime        int64
		Data, Properties  []byte
	}
	var tasks []taskState
	if err := DB.WithContext(ctx).Table("tasks").Select("id", "task_id", "user_id", "channel_id", "platform", "status", "submit_time", "data", "properties").Where("id IN ?", taskIDs).Find(&tasks).Error; err != nil {
		return nil, nil, err
	}
	byTask := map[int64]*taskState{}
	for i := range tasks {
		byTask[tasks[i].ID] = &tasks[i]
	}
	for i := range resolved {
		link, exists := links[resolved[i].Id]
		if !exists {
			continue
		}
		body := []byte(link.Evidence)
		if upstreamBillingTaskLinkHash(body) != link.EvidenceSHA256 {
			return nil, nil, ErrBillingEvidenceIntegrity
		}
		var proof upstreamBillingTaskLinkEvidence
		if err := common.Unmarshal(body, &proof); err != nil {
			return nil, nil, err
		}
		fingerprint, err := UpstreamBillingTaskLinkSourceFingerprint(&logs[i])
		if err != nil {
			return nil, nil, err
		}
		candidate := proof.Candidate
		task := byTask[link.TaskRowID]
		if proof.SchemaVersion != 1 || proof.Algorithm != UpstreamBillingTaskLinkRule || link.Algorithm != proof.Algorithm || proof.SourceSHA256 != fingerprint || link.SourceSHA256 != fingerprint || candidate.LogID != link.SourceLogID || candidate.TaskRowID != link.TaskRowID || candidate.TaskID != link.TaskID || candidate.UserID != link.UserID || candidate.ChannelID != link.ChannelID || link.UserID != logs[i].UserId || link.ChannelID != logs[i].ChannelId || task == nil || task.ID != candidate.TaskRowID || task.TaskID != candidate.TaskID || task.UserID != link.UserID || task.ChannelID != link.ChannelID || task.Status != candidate.Status || task.Platform != proof.Platform || task.SubmitTime != candidate.SubmitTime || upstreamBillingTaskLinkHash(task.Data) != candidate.DataSHA256 || upstreamBillingTaskLinkHash(task.Properties) != candidate.PropertiesSHA256 {
			return nil, nil, ErrBillingEvidenceIntegrity
		}
		var metadata map[string]any
		if err := common.UnmarshalJsonStr(resolved[i].Other, &metadata); err != nil {
			return nil, nil, err
		}
		if metadata == nil {
			return nil, nil, errors.New("missing upstream task metadata")
		}
		metadata["task_id"] = link.TaskID
		encoded, err := common.Marshal(metadata)
		if err != nil {
			return nil, nil, err
		}
		resolved[i].Other = string(encoded)
	}
	return resolved, links, nil
}
