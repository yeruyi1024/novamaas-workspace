package model

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TaskRequestBody keeps large administrator-only request payloads out of the
// hot task and log rows. ReferenceID is a task ID for normal task requests and
// "request:<request-id>" for legacy error logs that have no task linkage.
type TaskRequestBody struct {
	ID               int64           `json:"id" gorm:"primary_key"`
	ReferenceID      string          `json:"reference_id" gorm:"type:varchar(191);uniqueIndex"`
	TaskID           string          `json:"task_id" gorm:"type:varchar(191);index"`
	RequestID        string          `json:"request_id" gorm:"type:varchar(64);index"`
	Body             json.RawMessage `json:"body" gorm:"type:json"`
	BodySize         int64           `json:"body_size"`
	BodySHA256       string          `json:"body_sha256" gorm:"type:varchar(64)"`
	RequestCreatedAt int64           `json:"request_created_at" gorm:"bigint;index"`
	CreatedAt        int64           `json:"created_at" gorm:"bigint;index"`
	UpdatedAt        int64           `json:"updated_at" gorm:"bigint"`
}

func (body *TaskRequestBody) BeforeCreate(_ *gorm.DB) error {
	now := common.GetTimestamp()
	if body.CreatedAt == 0 {
		body.CreatedAt = now
	}
	if body.UpdatedAt == 0 {
		body.UpdatedAt = now
	}
	return nil
}

func taskRequestBodyReference(taskID string, requestID string) (string, error) {
	if taskID != "" {
		return taskID, nil
	}
	if requestID != "" {
		return "request:" + requestID, nil
	}
	return "", errors.New("task id or request id is required")
}

func saveTaskRequestBody(db *gorm.DB, taskID string, requestID string, body []byte, requestCreatedAt int64) error {
	if len(body) == 0 {
		return errors.New("request body is required")
	}
	var decoded any
	if err := common.Unmarshal(body, &decoded); err != nil {
		return fmt.Errorf("invalid request body JSON: %w", err)
	}
	referenceID, err := taskRequestBodyReference(taskID, requestID)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(body)
	record := TaskRequestBody{
		ReferenceID:      referenceID,
		TaskID:           taskID,
		RequestID:        requestID,
		Body:             append(json.RawMessage(nil), body...),
		BodySize:         int64(len(body)),
		BodySHA256:       fmt.Sprintf("%x", digest),
		RequestCreatedAt: requestCreatedAt,
		UpdatedAt:        common.GetTimestamp(),
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "reference_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"task_id", "request_id", "body", "body_size", "body_sha256", "request_created_at", "updated_at",
		}),
	}).Create(&record).Error
}

func SaveTaskRequestBody(taskID string, requestID string, body []byte) error {
	return saveTaskRequestBody(DB, taskID, requestID, body, common.GetTimestamp())
}

func SaveTaskRequestBodyAt(ctx context.Context, taskID string, requestID string, body []byte, requestCreatedAt int64) error {
	if requestCreatedAt <= 0 {
		requestCreatedAt = common.GetTimestamp()
	}
	return saveTaskRequestBody(DB.WithContext(ctx), taskID, requestID, body, requestCreatedAt)
}

func GetTaskRequestBody(taskID string) (json.RawMessage, bool, error) {
	return GetArchivedRequestBody(taskID, "")
}

func GetArchivedRequestBody(taskID string, requestID string) (json.RawMessage, bool, error) {
	referenceID, err := taskRequestBodyReference(taskID, requestID)
	if err != nil {
		return nil, false, err
	}
	var record TaskRequestBody
	err = DB.Select("body").Where("reference_id = ?", referenceID).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return record.Body, true, nil
}

type LegacyTaskRequestBodyRow struct {
	ID         int64
	TaskID     string
	CreatedAt  int64
	SubmitTime int64
	Properties Properties
}

func GetTaskRequestBodyArchiveUpperID(ctx context.Context) (int64, error) {
	var upperID int64
	err := DB.WithContext(ctx).Model(&Task{}).Select("COALESCE(MAX(id), 0)").Scan(&upperID).Error
	return upperID, err
}

func CountTasksThroughID(ctx context.Context, upperID int64) (int64, error) {
	var total int64
	err := DB.WithContext(ctx).Model(&Task{}).Where("id <= ?", upperID).Count(&total).Error
	return total, err
}

func FindLegacyTaskRequestBodyBatch(ctx context.Context, cursor int64, upperID int64, limit int) ([]LegacyTaskRequestBodyRow, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []LegacyTaskRequestBodyRow
	err := DB.WithContext(ctx).Model(&Task{}).
		Select("id, task_id, created_at, submit_time, properties").
		Where("id > ? AND id <= ?", cursor, upperID).
		Order("id ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// ArchiveLegacyTaskRequestBody writes the payload first and only then removes
// it from the hot task row. Re-running after interruption is therefore safe.
func ArchiveLegacyTaskRequestBody(ctx context.Context, row LegacyTaskRequestBodyRow) (bool, error) {
	if len(row.Properties.RequestBody) == 0 {
		return false, nil
	}
	err := DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		requestCreatedAt := row.SubmitTime
		if requestCreatedAt <= 0 {
			requestCreatedAt = row.CreatedAt
		}
		if err := saveTaskRequestBody(tx, row.TaskID, "", row.Properties.RequestBody, requestCreatedAt); err != nil {
			return err
		}
		row.Properties.RequestBody = nil
		return tx.Model(&Task{}).Where("id = ?", row.ID).Updates(map[string]any{
			"properties":             row.Properties,
			"request_body_available": true,
		}).Error
	})
	return err == nil, err
}

func DeleteOrphanTaskRequestBodiesBefore(ctx context.Context, targetTimestamp int64) (int64, error) {
	result := DB.WithContext(ctx).
		Where("request_created_at > 0 AND request_created_at < ?", targetTimestamp).
		Where("task_id = ? OR NOT EXISTS (SELECT 1 FROM tasks WHERE tasks.task_id = task_request_bodies.task_id)", "").
		Delete(&TaskRequestBody{})
	return result.RowsAffected, result.Error
}

type LogRequestBodyArchiveRow struct {
	ID          int
	CreatedAt   int64
	RequestID   string
	Other       string
	OtherSHA256 string `gorm:"column:other_sha256"`
}

type LogRequestBodyArchiveCursor struct {
	ID          int
	CreatedAt   int64
	RequestID   string
	OtherSHA256 string
}

func GetLogRequestBodyArchiveUpperCursor(ctx context.Context) (LogRequestBodyArchiveCursor, error) {
	var row LogRequestBodyArchiveRow
	order := "id DESC"
	selectColumns := "id, created_at, request_id"
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		selectColumns += ", lower(hex(SHA256(other))) AS other_sha256"
		order = "created_at DESC, request_id DESC, other_sha256 DESC"
	}
	err := LOG_DB.WithContext(ctx).Model(&Log{}).
		Select(selectColumns).
		Order(order).
		Limit(1).
		Scan(&row).Error
	return LogRequestBodyArchiveCursor{
		ID: row.ID, CreatedAt: row.CreatedAt, RequestID: row.RequestID, OtherSHA256: row.OtherSHA256,
	}, err
}

func CountLogsThroughCursor(ctx context.Context, upper LogRequestBodyArchiveCursor) (int64, error) {
	var total int64
	tx := LOG_DB.WithContext(ctx).Model(&Log{})
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		tx = tx.Where(
			"created_at < ? OR (created_at = ? AND request_id < ?) OR (created_at = ? AND request_id = ? AND lower(hex(SHA256(other))) <= ?)",
			upper.CreatedAt, upper.CreatedAt, upper.RequestID, upper.CreatedAt, upper.RequestID, upper.OtherSHA256,
		)
	} else {
		tx = tx.Where("id <= ?", upper.ID)
	}
	return total, tx.Count(&total).Error
}

func FindLogRequestBodyArchiveBatch(ctx context.Context, cursor LogRequestBodyArchiveCursor, upper LogRequestBodyArchiveCursor, limit int) ([]LogRequestBodyArchiveRow, error) {
	if limit <= 0 {
		limit = 100
	}
	selectColumns := "id, created_at, request_id, other"
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		selectColumns += ", lower(hex(SHA256(other))) AS other_sha256"
	}
	tx := LOG_DB.WithContext(ctx).Model(&Log{}).Select(selectColumns)
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		tx = tx.
			Where(
				"created_at > ? OR (created_at = ? AND request_id > ?) OR (created_at = ? AND request_id = ? AND lower(hex(SHA256(other))) > ?)",
				cursor.CreatedAt, cursor.CreatedAt, cursor.RequestID, cursor.CreatedAt, cursor.RequestID, cursor.OtherSHA256,
			).
			Where(
				"created_at < ? OR (created_at = ? AND request_id < ?) OR (created_at = ? AND request_id = ? AND lower(hex(SHA256(other))) <= ?)",
				upper.CreatedAt, upper.CreatedAt, upper.RequestID, upper.CreatedAt, upper.RequestID, upper.OtherSHA256,
			).
			Order("created_at ASC, request_id ASC, other_sha256 ASC")
	} else {
		tx = tx.Where("id > ? AND id <= ?", cursor.ID, upper.ID).Order("id ASC")
	}
	var rows []LogRequestBodyArchiveRow
	err := tx.Limit(limit).Scan(&rows).Error
	return rows, err
}

type LogRequestBodyArchiveUpdate struct {
	ID            int
	CreatedAt     int64
	RequestID     string
	OriginalOther string
	Other         string
}

func UpdateArchivedLogRequestBodies(ctx context.Context, updates []LogRequestBodyArchiveUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		args := make([]any, 0, len(updates)*6)
		assignment := "multiIf("
		where := ""
		for i, update := range updates {
			if i > 0 {
				assignment += ", "
				where += " OR "
			}
			originalDigest := sha256.Sum256([]byte(update.OriginalOther))
			originalSHA256 := fmt.Sprintf("%x", originalDigest)
			assignment += "(created_at = ? AND request_id = ? AND lower(hex(SHA256(other))) = ?), ?"
			args = append(args, update.CreatedAt, update.RequestID, originalSHA256, update.Other)
			where += "(created_at = ? AND request_id = ? AND lower(hex(SHA256(other))) = ?)"
			updates[i].OriginalOther = originalSHA256
		}
		assignment += ", other)"
		for _, update := range updates {
			args = append(args, update.CreatedAt, update.RequestID, update.OriginalOther)
		}
		query := "ALTER TABLE logs UPDATE other = " + assignment + " WHERE " + where + " SETTINGS mutations_sync = 1"
		return LOG_DB.WithContext(ctx).Exec(query, args...).Error
	}

	return LOG_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, update := range updates {
			if err := tx.Model(&Log{}).Where("id = ?", update.ID).Update("other", update.Other).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
