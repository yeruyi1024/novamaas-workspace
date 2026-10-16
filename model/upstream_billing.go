package model

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Upstream groups describe accounting ownership only. They do not participate
// in relay routing or modify a channel's routing group or tag.
type UpstreamBillingGroup struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"type:varchar(200)"`
	Version   int64  `json:"version" gorm:"bigint"`
	CreatedBy int    `json:"created_by"`
	CreatedAt int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt int64  `json:"updated_at" gorm:"bigint"`
}

type UpstreamBillingGroupChannel struct {
	ChannelID int `json:"channel_id" gorm:"primaryKey;autoIncrement:false"`
	GroupID   int `json:"group_id" gorm:"index"`
}

type UpstreamBillingGroupEvent struct {
	ID        int64  `json:"id" gorm:"primaryKey"`
	GroupID   int    `json:"group_id" gorm:"index"`
	ActorID   int    `json:"actor_id"`
	Version   int64  `json:"version" gorm:"bigint"`
	Snapshot  string `json:"snapshot" gorm:"type:text"`
	CreatedAt int64  `json:"created_at" gorm:"bigint"`
}

type UpstreamBillingChannelView struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Status       int    `json:"status"`
	CostDiscount string `json:"cost_discount"`
	GroupID      int    `json:"group_id,omitempty"`
	Eligible     bool   `json:"eligible"`
	Tag          string `json:"tag"`
	GroupName    string `json:"group_name,omitempty"`
	Missing      bool   `json:"missing"`
}

type UpstreamBillingGroupView struct {
	UpstreamBillingGroup
	Channels []UpstreamBillingChannelView `json:"channels"`
}

type UpstreamBillingUsage struct {
	Requests           int64 `json:"requests"`
	InputTokens        int64 `json:"input_tokens"`
	OutputTokens       int64 `json:"output_tokens"`
	CacheReadTokens    int64 `json:"cache_read_tokens"`
	CacheWriteTokens   int64 `json:"cache_write_tokens"`
	CacheWrite5mTokens int64 `json:"cache_write_5m_tokens"`
	CacheWrite1hTokens int64 `json:"cache_write_1h_tokens"`
	VideoTokens        int64 `json:"video_tokens"`
	VideoSeconds       int64 `json:"video_seconds"`
}

// Frozen model/day summaries and logos may exceed MySQL TEXT's 64 KiB limit.
// The API retains a JSON string; each database receives its native text type.
type UpstreamBillingSnapshotData string

func (UpstreamBillingSnapshotData) GormDataType() string { return "text" }

func (UpstreamBillingSnapshotData) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	if db.Dialector.Name() == "mysql" {
		return "LONGTEXT"
	}
	return "TEXT"
}

func (data UpstreamBillingSnapshotData) Value() (driver.Value, error) { return string(data), nil }

func (data *UpstreamBillingSnapshotData) Scan(value any) error {
	switch value := value.(type) {
	case nil:
		*data = ""
	case string:
		*data = UpstreamBillingSnapshotData(value)
	case []byte:
		*data = UpstreamBillingSnapshotData(value)
	default:
		return errors.New("invalid upstream billing snapshot data")
	}
	return nil
}

type UpstreamBillingStatement struct {
	ID                     string                      `json:"id" gorm:"type:varchar(64);primaryKey"`
	GroupID                int                         `json:"group_id" gorm:"uniqueIndex:idx_upstream_statement_revision,priority:1;index:idx_upstream_statement_list,priority:1"`
	Month                  string                      `json:"month" gorm:"type:varchar(7);uniqueIndex:idx_upstream_statement_revision,priority:2"`
	Revision               int                         `json:"revision" gorm:"uniqueIndex:idx_upstream_statement_revision,priority:3"`
	ActiveKey              *string                     `json:"-" gorm:"type:varchar(64);uniqueIndex"`
	Status                 string                      `json:"status" gorm:"type:varchar(16);index:idx_upstream_statement_jobs,priority:1"`
	StartAt                int64                       `json:"start_at" gorm:"bigint"`
	EndAt                  int64                       `json:"end_at" gorm:"bigint"`
	GroupVersion           int64                       `json:"group_version" gorm:"bigint"`
	Snapshot               UpstreamBillingSnapshotData `json:"snapshot"`
	SnapshotSHA256         string                      `json:"snapshot_sha256" gorm:"type:char(64)"`
	SourceSHA256           string                      `json:"source_sha256" gorm:"type:char(64)"`
	ManifestSHA256         string                      `json:"manifest_sha256" gorm:"type:char(64)"`
	PDFSHA256              string                      `json:"pdf_sha256" gorm:"type:char(64)"`
	ExcelSHA256            string                      `json:"excel_sha256" gorm:"type:char(64)"`
	ReviewEvidenceSHA256   string                      `json:"review_evidence_sha256,omitempty" gorm:"type:char(64)"`
	ReviewEvidenceSnapshot UpstreamBillingSnapshotData `json:"-"`
	RecordCount            int64                       `json:"record_count" gorm:"bigint"`
	StorageProfileID       int                         `json:"storage_profile_id"`
	CreatedBy              int                         `json:"created_by"`
	CreatedAt              int64                       `json:"created_at" gorm:"bigint;index:idx_upstream_statement_list,priority:2"`
	ConfirmedBy            int                         `json:"confirmed_by,omitempty"`
	ConfirmedAt            int64                       `json:"confirmed_at,omitempty" gorm:"bigint"`
	ArchivedAt             int64                       `json:"archived_at,omitempty" gorm:"bigint"`
	LeaseOwner             string                      `json:"-" gorm:"type:varchar(64)"`
	LeaseUntil             int64                       `json:"-" gorm:"bigint;index:idx_upstream_statement_jobs,priority:2"`
	LastError              string                      `json:"last_error,omitempty" gorm:"type:text"`
}

type UpstreamBillingStatementEvent struct {
	ID             int64  `json:"id" gorm:"primaryKey"`
	StatementID    string `json:"statement_id" gorm:"type:varchar(64);index"`
	ActorID        int    `json:"actor_id"`
	Action         string `json:"action" gorm:"type:varchar(32)"`
	Note           string `json:"note,omitempty" gorm:"type:text"`
	ManifestSHA256 string `json:"manifest_sha256,omitempty" gorm:"type:char(64)"`
	EvidenceSHA256 string `json:"evidence_sha256,omitempty" gorm:"type:char(64)"`
	SessionID      string `json:"-" gorm:"type:varchar(128)"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint"`
}

// Staging rows freeze safe per-event facts before cloud I/O. They never contain
// API keys, request/response bodies, customer balances, or mutable pricing.
type UpstreamBillingStatementRecord struct {
	ID                   int64  `json:"-" gorm:"primaryKey"`
	CreatedAt            int64  `json:"-" gorm:"bigint;index"`
	StatementID          string `json:"-" gorm:"type:varchar(64);uniqueIndex:idx_upstream_record_sequence,priority:1"`
	Sequence             int64  `json:"sequence" gorm:"bigint;uniqueIndex:idx_upstream_record_sequence,priority:2"`
	SourceLogID          int    `json:"source_log_id"`
	CostSnapshotID       int64  `json:"cost_snapshot_id" gorm:"bigint"`
	LatestAdjustmentID   int64  `json:"latest_adjustment_id" gorm:"bigint"`
	OccurredAt           int64  `json:"occurred_at" gorm:"bigint"`
	LogType              int    `json:"log_type"`
	UserID               int    `json:"user_id"`
	ChannelID            int    `json:"channel_id"`
	ModelName            string `json:"model_name" gorm:"type:varchar(255)"`
	UpstreamModelName    string `json:"upstream_model_name,omitempty" gorm:"type:varchar(255)"`
	BillingGroup         string `json:"billing_group,omitempty" gorm:"type:varchar(64)"`
	RequestID            string `json:"request_id" gorm:"type:varchar(128)"`
	UpstreamRequestID    string `json:"upstream_request_id,omitempty" gorm:"type:varchar(128)"`
	TaskID               string `json:"task_id,omitempty" gorm:"type:varchar(191)"`
	CostQuota            int64  `json:"cost_quota" gorm:"bigint"`
	CostDiscount         string `json:"cost_discount" gorm:"type:varchar(16)"`
	CostBasisQuota       string `json:"cost_basis_quota" gorm:"type:varchar(64)"`
	UpstreamBillingUsage `gorm:"embedded;embeddedPrefix:usage_"`
	CostMissing          bool   `json:"cost_missing"`
	UsageMissing         bool   `json:"usage_missing"`
	UsageEstimated       bool   `json:"usage_estimated"`
	TaskPending          bool   `json:"task_pending"`
	UsageSource          string `json:"usage_source" gorm:"type:varchar(64)"`
}

// Only abandoned collection rows are disposable. The NOT EXISTS predicate is
// repeated in the delete itself, so a publication after selection wins safely.
func PurgeUnpublishedUpstreamBillingRecords(ctx context.Context, olderThan int64, limit int) (int64, error) {
	if olderThan <= 0 || olderThan > common.GetTimestamp()-86400 || limit <= 0 || limit > 1000 {
		return 0, errors.New("invalid unpublished upstream record cleanup range")
	}
	var ids []int64
	condition := "created_at > 0 AND created_at < ? AND NOT EXISTS (SELECT 1 FROM upstream_billing_statements WHERE upstream_billing_statements.id = upstream_billing_statement_records.statement_id)"
	if err := DB.WithContext(ctx).Model(&UpstreamBillingStatementRecord{}).Where(condition, olderThan).Order("id asc").Limit(limit).Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	result := DB.WithContext(ctx).Where("id IN ?", ids).Where(condition, olderThan).Delete(&UpstreamBillingStatementRecord{})
	return result.RowsAffected, result.Error
}

func ListUpstreamBillingGroups(ctx context.Context) ([]UpstreamBillingGroupView, error) {
	var groups []UpstreamBillingGroup
	if err := DB.WithContext(ctx).Order("id asc").Find(&groups).Error; err != nil {
		return nil, err
	}
	views := make([]UpstreamBillingGroupView, 0, len(groups))
	for _, group := range groups {
		view, err := GetUpstreamBillingGroup(ctx, group.ID)
		if err != nil {
			return nil, err
		}
		views = append(views, *view)
	}
	return views, nil
}

func GetUpstreamBillingGroup(ctx context.Context, id int) (*UpstreamBillingGroupView, error) {
	if id <= 0 {
		return nil, errors.New("invalid upstream accounting group")
	}
	var group UpstreamBillingGroup
	if err := DB.WithContext(ctx).First(&group, id).Error; err != nil {
		return nil, err
	}
	var members []UpstreamBillingGroupChannel
	if err := DB.WithContext(ctx).Where("group_id = ?", id).Order("channel_id asc").Find(&members).Error; err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.ChannelID)
	}
	var channels []Channel
	if len(ids) > 0 {
		if err := DB.WithContext(ctx).Select("id", "name", "status", "cost_discount", "tag").Where("id IN ?", ids).Order("id asc").Find(&channels).Error; err != nil {
			return nil, err
		}
	}
	byID := make(map[int]Channel, len(channels))
	for _, channel := range channels {
		byID[channel.Id] = channel
	}
	view := &UpstreamBillingGroupView{UpstreamBillingGroup: group, Channels: make([]UpstreamBillingChannelView, 0, len(members))}
	for _, member := range members {
		channel, found := byID[member.ChannelID]
		discount, err := NormalizeCostDiscount(channel.CostDiscount)
		if err != nil {
			return nil, err
		}
		view.Channels = append(view.Channels, UpstreamBillingChannelView{ID: member.ChannelID, Name: channel.Name, Status: channel.Status, CostDiscount: discount, Eligible: found && discount != "", Tag: channel.GetTag(), GroupID: id, GroupName: group.Name, Missing: !found})
	}
	return view, nil
}

func SearchUpstreamBillingChannels(ctx context.Context, q string) ([]UpstreamBillingChannelView, error) {
	q = strings.TrimSpace(q)
	if len([]rune(q)) > 100 {
		return nil, errors.New("channel search is too long")
	}
	query := DB.WithContext(ctx).Table("channels").
		Select("channels.id, channels.name, channels.status, channels.cost_discount, channels.tag, COALESCE(members.group_id, 0) AS group_id, COALESCE(accounting.name, '') AS group_name").
		Joins("LEFT JOIN upstream_billing_group_channels AS members ON members.channel_id = channels.id").
		Joins("LEFT JOIN upstream_billing_groups AS accounting ON accounting.id = members.group_id")
	if q != "" {
		pattern := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(q) + "%"
		query = query.Where("channels.name LIKE ? ESCAPE '!' OR channels.tag LIKE ? ESCAPE '!'", pattern, pattern)
	}
	views := make([]UpstreamBillingChannelView, 0)
	if err := query.Order("channels.id asc").Limit(100).Scan(&views).Error; err != nil {
		return nil, err
	}
	for i := range views {
		discount, err := NormalizeCostDiscount(views[i].CostDiscount)
		if err != nil {
			return nil, err
		}
		views[i].CostDiscount, views[i].Eligible = discount, discount != ""
	}
	return views, nil
}

func SaveUpstreamBillingGroup(ctx context.Context, id int, expectedVersion int64, name string, channelIDs []int, actorID int) (*UpstreamBillingGroupView, error) {
	name = strings.TrimSpace(name)
	if id < 0 || expectedVersion < 0 || actorID <= 0 || name == "" || len([]rune(name)) > 200 || strings.ContainsAny(name, "\x00\r\n") || len(channelIDs) == 0 || len(channelIDs) > 100 {
		return nil, errors.New("invalid upstream accounting group")
	}
	seen := make(map[int]bool, len(channelIDs))
	for _, channelID := range channelIDs {
		if channelID <= 0 || seen[channelID] {
			return nil, errors.New("invalid upstream accounting channel membership")
		}
		seen[channelID] = true
	}
	var saved UpstreamBillingGroup
	err := DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		saved = UpstreamBillingGroup{ID: id}
		if id == 0 {
			if expectedVersion != 0 {
				return ErrBillingConflict
			}
			saved.CreatedBy, saved.CreatedAt = actorID, common.GetTimestamp()
			if err := tx.Create(&saved).Error; err != nil {
				return err
			}
		} else if err := lockForUpdate(tx).First(&saved, id).Error; err != nil {
			return err
		}
		if saved.Version != expectedVersion {
			return ErrBillingConflict
		}
		var channels []Channel
		if err := lockForUpdate(tx).Select("id", "cost_discount").Where("id IN ?", channelIDs).Order("id asc").Find(&channels).Error; err != nil {
			return err
		}
		if len(channels) != len(channelIDs) {
			return errors.New("an upstream accounting channel no longer exists")
		}
		for _, channel := range channels {
			if _, err := NormalizeCostDiscount(channel.CostDiscount); err != nil {
				return err
			}
		}
		var conflicts int64
		if err := tx.Model(&UpstreamBillingGroupChannel{}).Where("channel_id IN ? AND group_id <> ?", channelIDs, saved.ID).Count(&conflicts).Error; err != nil {
			return err
		}
		if conflicts > 0 {
			return ErrBillingConflict
		}
		if err := tx.Where("group_id = ?", saved.ID).Delete(&UpstreamBillingGroupChannel{}).Error; err != nil {
			return err
		}
		members := make([]UpstreamBillingGroupChannel, 0, len(channelIDs))
		for _, channel := range channels {
			members = append(members, UpstreamBillingGroupChannel{ChannelID: channel.Id, GroupID: saved.ID})
		}
		if err := tx.Create(&members).Error; err != nil {
			return err
		}
		saved.Name, saved.Version, saved.UpdatedAt = name, saved.Version+1, common.GetTimestamp()
		if err := tx.Save(&saved).Error; err != nil {
			return err
		}
		body, err := common.Marshal(struct {
			Name     string                        `json:"name"`
			Channels []UpstreamBillingGroupChannel `json:"channels"`
		}{name, members})
		if err != nil {
			return err
		}
		return tx.Create(&UpstreamBillingGroupEvent{GroupID: saved.ID, ActorID: actorID, Version: saved.Version, Snapshot: string(body), CreatedAt: saved.UpdatedAt}).Error
	})
	if err != nil {
		return nil, err
	}
	return GetUpstreamBillingGroup(ctx, saved.ID)
}
