package model

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// AssetRequestLog contains only allowlisted request metadata. Diagnostic payloads
// live in a separate, short-lived table so list queries stay inexpensive.
type AssetRequestLog struct {
	ID            int64  `json:"id" gorm:"primaryKey"`
	CreatedAt     int64  `json:"created_at" gorm:"bigint;index:idx_asset_request_time,priority:1;index:idx_asset_request_channel_time,priority:2;index:idx_asset_request_replica_time,priority:2"`
	ChannelID     int    `json:"channel_id" gorm:"index:idx_asset_request_channel_time,priority:1"`
	ReplicaID     int64  `json:"replica_id" gorm:"index:idx_asset_request_replica_time,priority:1"`
	AssetID       int64  `json:"asset_id"`
	RequestID     string `json:"request_id" gorm:"type:varchar(64);index"`
	Source        string `json:"source" gorm:"type:varchar(16)"`
	Protocol      string `json:"protocol" gorm:"type:varchar(32)"`
	Operation     string `json:"operation" gorm:"type:varchar(32)"`
	Method        string `json:"method" gorm:"type:varchar(8)"`
	PathTemplate  string `json:"path_template" gorm:"type:varchar(64)"`
	HTTPStatus    int    `json:"http_status"`
	DurationMS    int64  `json:"duration_ms"`
	RequestBytes  int64  `json:"request_bytes"`
	ResponseBytes int64  `json:"response_bytes"`
	Result        string `json:"result" gorm:"type:varchar(16)"`
	ErrorKind     string `json:"error_kind" gorm:"type:varchar(32)"`
}

type AssetRequestLogDetail struct {
	LogID               int64  `json:"log_id" gorm:"primaryKey"`
	CreatedAt           int64  `json:"-" gorm:"bigint;index"`
	RequestURL          string `json:"request_url" gorm:"type:text"`
	RequestBody         string `json:"request_body" gorm:"type:text"`
	ResponseBody        string `json:"response_body" gorm:"type:text"`
	RequestBodyOmitted  bool   `json:"request_body_omitted"`
	ResponseBodyOmitted bool   `json:"response_body_omitted"`
}

type AssetRequestLogEntry struct {
	Log    AssetRequestLog
	Detail AssetRequestLogDetail
}

type AssetRequestLogWithDetail struct {
	Log    AssetRequestLog        `json:"log"`
	Detail *AssetRequestLogDetail `json:"detail"`
}

type AssetRequestLogFilter struct {
	ChannelID int
	ReplicaID int64
	RequestID string
	Source    string
	Result    string
	StartMS   int64
	EndMS     int64
	CursorMS  int64
	CursorID  int64
	Limit     int
}

func assetRequestLogDB() *gorm.DB {
	if LOG_DB != nil && !common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		return LOG_DB
	}
	return DB
}

func MigrateAssetRequestLogs() error {
	db := assetRequestLogDB()
	if db == nil {
		return errors.New("asset request log database is unavailable")
	}
	return db.AutoMigrate(&AssetRequestLog{}, &AssetRequestLogDetail{})
}

func InsertAssetRequestLogEntries(ctx context.Context, entries []AssetRequestLogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	db := assetRequestLogDB()
	if db == nil {
		return errors.New("asset request log database is unavailable")
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		logs := make([]AssetRequestLog, len(entries))
		for i := range entries {
			logs[i] = entries[i].Log
		}
		if err := tx.CreateInBatches(&logs, 50).Error; err != nil {
			return err
		}
		details := make([]AssetRequestLogDetail, len(entries))
		for i := range entries {
			details[i] = entries[i].Detail
			details[i].LogID = logs[i].ID
			details[i].CreatedAt = logs[i].CreatedAt
		}
		return tx.CreateInBatches(&details, 50).Error
	})
}

func GetAssetRequestLogWithDetail(ctx context.Context, id int64) (*AssetRequestLogWithDetail, error) {
	db := assetRequestLogDB()
	if db == nil {
		return nil, errors.New("asset request log database is unavailable")
	}
	var log AssetRequestLog
	if err := db.WithContext(ctx).First(&log, id).Error; err != nil {
		return nil, err
	}
	var detail AssetRequestLogDetail
	result := db.WithContext(ctx).Where("log_id = ?", id).Limit(1).Find(&detail)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return &AssetRequestLogWithDetail{Log: log}, nil
	}
	return &AssetRequestLogWithDetail{Log: log, Detail: &detail}, nil
}

func InsertAssetRequestLogs(ctx context.Context, logs []AssetRequestLog) error {
	if len(logs) == 0 {
		return nil
	}
	db := assetRequestLogDB()
	if db == nil {
		return errors.New("asset request log database is unavailable")
	}
	return db.WithContext(ctx).CreateInBatches(&logs, 100).Error
}

func ListAssetRequestLogs(ctx context.Context, filter AssetRequestLogFilter) ([]AssetRequestLog, string, error) {
	db := assetRequestLogDB()
	if db == nil {
		return nil, "", errors.New("asset request log database is unavailable")
	}
	query := db.WithContext(ctx).Model(&AssetRequestLog{}).
		Where("created_at >= ? AND created_at <= ?", filter.StartMS, filter.EndMS)
	if filter.ChannelID > 0 {
		query = query.Where("channel_id = ?", filter.ChannelID)
	}
	if filter.ReplicaID > 0 {
		query = query.Where("replica_id = ?", filter.ReplicaID)
	}
	if filter.RequestID != "" {
		query = query.Where("request_id = ?", filter.RequestID)
	}
	if filter.Source != "" {
		query = query.Where("source = ?", filter.Source)
	}
	if filter.Result != "" {
		query = query.Where("result = ?", filter.Result)
	}
	if filter.CursorID > 0 {
		query = query.Where("created_at < ? OR (created_at = ? AND id < ?)", filter.CursorMS, filter.CursorMS, filter.CursorID)
	}
	var logs []AssetRequestLog
	if err := query.Order("created_at DESC, id DESC").Limit(filter.Limit + 1).Find(&logs).Error; err != nil {
		return nil, "", err
	}
	if len(logs) <= filter.Limit {
		return logs, "", nil
	}
	logs = logs[:filter.Limit]
	last := logs[len(logs)-1]
	return logs, fmt.Sprintf("%d:%d", last.CreatedAt, last.ID), nil
}

func ParseAssetRequestLogCursor(value string) (int64, int64, error) {
	if value == "" {
		return 0, 0, nil
	}
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid request log cursor")
	}
	createdAt, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || createdAt <= 0 {
		return 0, 0, errors.New("invalid request log cursor")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return 0, 0, errors.New("invalid request log cursor")
	}
	return createdAt, id, nil
}

// PurgeAssetRequestLogsBefore deletes one small batch to avoid long locks.
func PurgeAssetRequestLogsBefore(ctx context.Context, cutoffMS int64, limit int) (int, error) {
	db := assetRequestLogDB()
	if db == nil {
		return 0, errors.New("asset request log database is unavailable")
	}
	var ids []int64
	if err := db.WithContext(ctx).Model(&AssetRequestLog{}).
		Where("created_at < ?", cutoffMS).Order("created_at ASC").Limit(limit).Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	var deleted int64
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("log_id IN ?", ids).Delete(&AssetRequestLogDetail{}).Error; err != nil {
			return err
		}
		result := tx.Where("id IN ?", ids).Delete(&AssetRequestLog{})
		deleted = result.RowsAffected
		return result.Error
	})
	return int(deleted), err
}

// PurgeAssetRequestLogDetailsBefore retains compact metadata longer than payloads.
func PurgeAssetRequestLogDetailsBefore(ctx context.Context, cutoffMS int64, limit int) (int, error) {
	db := assetRequestLogDB()
	if db == nil {
		return 0, errors.New("asset request log database is unavailable")
	}
	var ids []int64
	if err := db.WithContext(ctx).Model(&AssetRequestLogDetail{}).
		Where("created_at < ?", cutoffMS).Order("created_at ASC").Limit(limit).Pluck("log_id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	result := db.WithContext(ctx).Where("log_id IN ?", ids).Delete(&AssetRequestLogDetail{})
	return int(result.RowsAffected), result.Error
}
