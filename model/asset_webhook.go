package model

import (
	"context"
	"errors"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

const (
	AssetWebhookEndpointStatusEnabled  = "enabled"
	AssetWebhookEndpointStatusDisabled = "disabled"

	AssetWebhookDeliveryStatusPending    = "pending"
	AssetWebhookDeliveryStatusDelivering = "delivering"
	AssetWebhookDeliveryStatusFailed     = "failed"
	AssetWebhookDeliveryStatusSucceeded  = "succeeded"
	AssetWebhookDeliveryStatusExhausted  = "exhausted"
	AssetWebhookDeliveryStatusSuperseded = "superseded"
)

var ErrAssetWebhookEndpointLimit = errors.New("asset webhook endpoint limit reached")

// AssetWebhookEndpoint stores one tenant-owned downstream callback.
type AssetWebhookEndpoint struct {
	ID          int64  `json:"-" gorm:"primaryKey"`
	PublicID    string `json:"id" gorm:"type:varchar(64);uniqueIndex"`
	OwnerUserID int    `json:"owner_user_id" gorm:"index:idx_asset_webhook_owner_status,priority:1"`
	Name        string `json:"name" gorm:"type:varchar(128)"`
	URL         string `json:"url" gorm:"type:varchar(2048)"`
	EventTypes  string `json:"-" gorm:"type:text"`
	Status      string `json:"status" gorm:"type:varchar(16);index:idx_asset_webhook_owner_status,priority:2"`
	CreatedAt   int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt   int64  `json:"updated_at" gorm:"bigint"`
}

// AssetWebhookDelivery is both the durable outbox row and delivery audit.
// Payload is immutable so retries always carry the original state transition.
type AssetWebhookDelivery struct {
	ID               int64  `json:"-" gorm:"primaryKey"`
	EventID          string `json:"event_id" gorm:"type:varchar(64);index"`
	WebhookID        string `json:"webhook_id" gorm:"type:varchar(64);uniqueIndex"`
	EndpointID       int64  `json:"-" gorm:"index"`
	EndpointPublicID string `json:"endpoint_id" gorm:"type:varchar(64);index"`
	OwnerUserID      int    `json:"owner_user_id" gorm:"index"`
	AssetID          int64  `json:"-" gorm:"index"`
	EventType        string `json:"event_type" gorm:"type:varchar(64);index"`
	Payload          string `json:"-" gorm:"type:text"`
	Status           string `json:"status" gorm:"type:varchar(16);index:idx_asset_webhook_delivery_work,priority:1;index:idx_asset_webhook_delivery_retention,priority:1"`
	Attempts         int    `json:"attempts"`
	NextAttemptAt    int64  `json:"next_attempt_at" gorm:"bigint;index:idx_asset_webhook_delivery_work,priority:2"`
	LockedBy         string `json:"-" gorm:"type:varchar(128);index"`
	LeaseUntil       int64  `json:"-" gorm:"bigint;index"`
	ResponseStatus   int    `json:"response_status"`
	LastError        string `json:"last_error,omitempty" gorm:"type:text"`
	DeliveredAt      int64  `json:"delivered_at" gorm:"bigint"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint;index;index:idx_asset_webhook_delivery_retention,priority:2"`
	UpdatedAt        int64  `json:"updated_at" gorm:"bigint"`
}

func (value *AssetWebhookEndpoint) BeforeCreate(_ *gorm.DB) error {
	setAssetTimestamps(&value.CreatedAt, &value.UpdatedAt)
	return nil
}

func (value *AssetWebhookDelivery) BeforeCreate(_ *gorm.DB) error {
	setAssetTimestamps(&value.CreatedAt, &value.UpdatedAt)
	return nil
}

// CreateAssetWebhookEndpointWithinLimit serializes creation before counting so
// concurrent requests cannot exceed the per-owner limit.
func CreateAssetWebhookEndpointWithinLimit(endpoint *AssetWebhookEndpoint, limit int) error {
	if common.UsingMainDatabase(common.DatabaseTypeSQLite) {
		return DB.Connection(func(connection *gorm.DB) error {
			if err := connection.Exec("BEGIN IMMEDIATE").Error; err != nil {
				return err
			}
			committed := false
			defer func() {
				if !committed {
					_ = connection.Exec("ROLLBACK").Error
				}
			}()
			transaction := connection.Session(&gorm.Session{SkipDefaultTransaction: true})
			if err := createAssetWebhookEndpointWithinLimit(transaction, endpoint, limit, false); err != nil {
				return err
			}
			if err := connection.Exec("COMMIT").Error; err != nil {
				return err
			}
			committed = true
			return nil
		})
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		return createAssetWebhookEndpointWithinLimit(tx, endpoint, limit, true)
	})
}

func createAssetWebhookEndpointWithinLimit(tx *gorm.DB, endpoint *AssetWebhookEndpoint, limit int, lockOwner bool) error {
	ownerQuery := tx
	if lockOwner {
		ownerQuery = lockForUpdate(ownerQuery)
	}
	var owner User
	if err := ownerQuery.Select("id").First(&owner, endpoint.OwnerUserID).Error; err != nil {
		return err
	}
	var count int64
	if err := tx.Model(&AssetWebhookEndpoint{}).
		Where("owner_user_id = ? AND status = ?", endpoint.OwnerUserID, AssetWebhookEndpointStatusEnabled).
		Count(&count).Error; err != nil {
		return err
	}
	if count >= int64(limit) {
		return ErrAssetWebhookEndpointLimit
	}
	return tx.Create(endpoint).Error
}

// PurgeAssetWebhookDeliveriesBefore removes only terminal deliveries in a
// bounded batch. Retrying and in-flight notifications remain untouched.
func PurgeAssetWebhookDeliveriesBefore(ctx context.Context, statuses []string, cutoff int64, limit int) (int64, error) {
	if limit <= 0 || cutoff <= 0 || len(statuses) == 0 {
		return 0, errors.New("invalid asset webhook retention request")
	}
	for _, status := range statuses {
		if status != AssetWebhookDeliveryStatusSucceeded && status != AssetWebhookDeliveryStatusSuperseded && status != AssetWebhookDeliveryStatusExhausted {
			return 0, errors.New("only terminal asset webhook deliveries may be purged")
		}
	}
	var ids []int64
	if err := DB.WithContext(ctx).Model(&AssetWebhookDelivery{}).
		Where("status IN ? AND created_at < ?", statuses, cutoff).
		Order("created_at asc").Order("id asc").Limit(limit).Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	result := DB.WithContext(ctx).Where("id IN ? AND status IN ? AND created_at < ?", ids, statuses, cutoff).
		Delete(&AssetWebhookDelivery{})
	return result.RowsAffected, result.Error
}

func ClaimAssetWebhookDeliveries(now int64, leaseUntil int64, lockedBy string, limit int) ([]*AssetWebhookDelivery, error) {
	if limit <= 0 {
		limit = 20
	}
	var candidates []*AssetWebhookDelivery
	if err := DB.Where(
		"((status IN ? AND next_attempt_at <= ?) OR (status = ? AND lease_until <= ?))",
		[]string{AssetWebhookDeliveryStatusPending, AssetWebhookDeliveryStatusFailed}, now,
		AssetWebhookDeliveryStatusDelivering, now,
	).Order("next_attempt_at asc").Order("id asc").Limit(limit).Find(&candidates).Error; err != nil {
		return nil, err
	}
	claimed := make([]*AssetWebhookDelivery, 0, len(candidates))
	for _, candidate := range candidates {
		result := DB.Model(&AssetWebhookDelivery{}).
			Where("id = ? AND ((status IN ? AND next_attempt_at <= ?) OR (status = ? AND lease_until <= ?))",
				candidate.ID,
				[]string{AssetWebhookDeliveryStatusPending, AssetWebhookDeliveryStatusFailed}, now,
				AssetWebhookDeliveryStatusDelivering, now,
			).
			Updates(map[string]any{
				"status":      AssetWebhookDeliveryStatusDelivering,
				"locked_by":   lockedBy,
				"lease_until": leaseUntil,
				"updated_at":  now,
			})
		if result.Error != nil {
			return nil, result.Error
		}
		if result.RowsAffected == 1 {
			candidate.Status = AssetWebhookDeliveryStatusDelivering
			candidate.LockedBy = lockedBy
			candidate.LeaseUntil = leaseUntil
			claimed = append(claimed, candidate)
		}
	}
	return claimed, nil
}

// OwnsAssetWebhookDeliveryLease checks the persisted claim before an external POST.
func OwnsAssetWebhookDeliveryLease(ctx context.Context, id int64, lockedBy string, now int64) (bool, error) {
	var count int64
	err := DB.WithContext(ctx).Model(&AssetWebhookDelivery{}).
		Where("id = ? AND status = ? AND locked_by = ? AND lease_until > ?", id, AssetWebhookDeliveryStatusDelivering, lockedBy, now).
		Count(&count).Error
	return count == 1, err
}

func SupersedeAssetWebhookDelivery(ctx context.Context, id int64, lockedBy string, reason string) error {
	now := common.GetTimestamp()
	result := DB.WithContext(ctx).Model(&AssetWebhookDelivery{}).
		Where("id = ? AND status = ? AND locked_by = ? AND lease_until > ?", id, AssetWebhookDeliveryStatusDelivering, lockedBy, now).
		Updates(map[string]any{
			"status":          AssetWebhookDeliveryStatusSuperseded,
			"next_attempt_at": 0,
			"locked_by":       "",
			"lease_until":     0,
			"last_error":      reason,
			"updated_at":      now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("asset webhook delivery lease was lost")
	}
	return nil
}

func FinishAssetWebhookDelivery(ctx context.Context, id int64, lockedBy string, status string, responseStatus int, nextAttemptAt int64, lastError string) error {
	now := common.GetTimestamp()
	updates := map[string]any{
		"status":          status,
		"attempts":        gorm.Expr("attempts + ?", 1),
		"next_attempt_at": nextAttemptAt,
		"locked_by":       "",
		"lease_until":     0,
		"response_status": responseStatus,
		"last_error":      lastError,
		"updated_at":      now,
	}
	if status == AssetWebhookDeliveryStatusSucceeded {
		updates["delivered_at"] = now
	}
	result := DB.WithContext(ctx).Model(&AssetWebhookDelivery{}).
		Where("id = ? AND status = ? AND locked_by = ? AND lease_until > ?", id, AssetWebhookDeliveryStatusDelivering, lockedBy, now).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("asset webhook delivery lease was lost")
	}
	return nil
}
