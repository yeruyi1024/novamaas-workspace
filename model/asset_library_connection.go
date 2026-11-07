package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// AssetLibraryConnection identifies one upstream resource namespace. Multiple
// routing channels can share its credentials, replicas and request budget.
type AssetLibraryConnection struct {
	ID                   int     `json:"id" gorm:"primaryKey"`
	Name                 string  `json:"name" gorm:"type:varchar(128)"`
	ProviderName         string  `json:"provider_name" gorm:"type:varchar(128)"`
	Enabled              bool    `json:"enabled"`
	Protocol             string  `json:"protocol" gorm:"type:varchar(32)"`
	AuthType             string  `json:"auth_type" gorm:"type:varchar(16)"`
	BaseURL              string  `json:"base_url" gorm:"type:varchar(1024)"`
	Region               string  `json:"region" gorm:"type:varchar(64)"`
	Service              string  `json:"service" gorm:"type:varchar(64)"`
	APIVersion           string  `json:"api_version" gorm:"type:varchar(32)"`
	ProjectName          string  `json:"project_name" gorm:"type:varchar(128)"`
	QPM                  int     `json:"qpm"`
	AccessKeyID          string  `json:"-" gorm:"type:varchar(255)"`
	AccessKeyHint        string  `json:"access_key_hint" gorm:"type:varchar(16)"`
	EncryptedCredential  string  `json:"-" gorm:"type:text"`
	CredentialKeyVersion string  `json:"-" gorm:"type:varchar(32)"`
	LegacyChannelID      *int    `json:"legacy_channel_id,omitempty" gorm:"uniqueIndex"`
	ResourceKey          *string `json:"-" gorm:"type:char(64);uniqueIndex"`
	CreatedAt            int64   `json:"created_at" gorm:"bigint;index"`
	UpdatedAt            int64   `json:"updated_at" gorm:"bigint;index"`
}

func (value *AssetLibraryConnection) BeforeCreate(_ *gorm.DB) error {
	setAssetTimestamps(&value.CreatedAt, &value.UpdatedAt)
	return nil
}

func LockAssetLibraryConnection(tx *gorm.DB, libraryID int) (*AssetLibraryConnection, error) {
	var library AssetLibraryConnection
	if err := lockForUpdate(tx).First(&library, libraryID).Error; err != nil {
		return nil, err
	}
	return &library, nil
}

// LockAssetLibraryChannels serializes ownership changes even when a channel has
// no binding row yet. Read only the public fields required for association.
func LockAssetLibraryChannels(tx *gorm.DB, channelIDs []int) ([]Channel, error) {
	var channels []Channel
	err := lockForUpdate(tx).Select("id", "type", "name").Where("id IN ?", channelIDs).Order("id asc").Find(&channels).Error
	return channels, err
}

func (value AssetLibraryConnection) ProviderConfig(channelID int) AssetChannelConfig {
	return AssetChannelConfig{
		ChannelID: channelID, LibraryID: value.ID, Enabled: value.Enabled,
		Protocol: value.Protocol, AuthType: value.AuthType, BaseURL: value.BaseURL,
		Region: value.Region, Service: value.Service, APIVersion: value.APIVersion,
		ProjectName: value.ProjectName, QPM: value.QPM, AccessKeyID: value.AccessKeyID,
		AccessKeyHint: value.AccessKeyHint, EncryptedCredential: value.EncryptedCredential,
		CredentialKeyVersion: value.CredentialKeyVersion,
		CreatedAt:            value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
}

func AssetLibraryFromConfig(name string, config AssetChannelConfig) AssetLibraryConnection {
	return AssetLibraryConnection{
		Name: name, ProviderName: name, Enabled: config.Enabled, Protocol: config.Protocol, AuthType: config.AuthType,
		BaseURL: config.BaseURL, Region: config.Region, Service: config.Service,
		APIVersion: config.APIVersion, ProjectName: config.ProjectName, QPM: config.QPM,
		AccessKeyID: config.AccessKeyID, AccessKeyHint: config.AccessKeyHint,
		EncryptedCredential: config.EncryptedCredential, CredentialKeyVersion: config.CredentialKeyVersion,
		CreatedAt: config.CreatedAt, UpdatedAt: config.UpdatedAt,
	}
}

// RateLimitID separates shared-library budgets from legacy channel budgets.
// These keys are internal rate-limiter buckets, never routing-channel IDs.
func (value AssetChannelConfig) RateLimitID() int {
	if value.LibraryID > 0 {
		return -value.LibraryID
	}
	return value.ChannelID
}

// MigrateAssetLibraryConnections only rehomes local metadata. It neither
// contacts upstream providers nor merges, reuploads or deletes old replicas.
func MigrateAssetLibraryConnections(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var configs []AssetChannelConfig
		if err := lockForUpdate(tx).Where("(library_id = ? OR library_id IS NULL) AND (protocol <> ? OR encrypted_credential <> ?)", 0, "", "").Order("channel_id asc").Find(&configs).Error; err != nil {
			return err
		}
		for _, config := range configs {
			if config.ChannelID <= 0 {
				return errors.New("legacy asset configuration has an invalid channel ID")
			}
			var library AssetLibraryConnection
			err := tx.Where("legacy_channel_id = ?", config.ChannelID).First(&library).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				name := fmt.Sprintf("Asset library #%d", config.ChannelID)
				var channel Channel
				if result := tx.Select("id", "name").Where("id = ?", config.ChannelID).Find(&channel); result.Error != nil {
					return result.Error
				} else if channel.Name != "" {
					name = channel.Name
				}
				library = AssetLibraryFromConfig(name, config)
				library.LegacyChannelID = new(int)
				*library.LegacyChannelID = config.ChannelID
				if err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).Create(&library).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			if err := tx.Model(&config).Update("library_id", library.ID).Error; err != nil {
				return err
			}
			for _, replicaModel := range []any{&AssetReplica{}, &AssetGroupReplica{}} {
				if err := tx.Model(replicaModel).Where("channel_id = ? AND library_id IS NULL", config.ChannelID).
					Update("library_id", library.ID).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// MergeAssetLibraryGroup folds accounts that point at exactly the same upstream
// account into the one holding the most active copies (lowest ID on a tie).
// Channel bindings move to it and it takes the name of the lowest-ID account. A copy moves only when the kept account has none
// for that asset or platform group, so nothing is uploaded again and nothing
// upstream is deleted; surplus copies stay with their old account, which keeps
// its credentials so those copies can still be removed with the asset; an old
// account holding no copies at all is deleted. It
// returns 0, changing nothing, while any of the accounts has a request in
// flight, so a later run retries; otherwise it returns the kept account's ID.
func MergeAssetLibraryGroup(ids []int, resourceKey string) (int, error) {
	keptID := 0
	err := DB.Transaction(func(tx *gorm.DB) error {
		var libraries []AssetLibraryConnection
		if err := lockForUpdate(tx).Where("id IN ?", ids).Order("id asc").Find(&libraries).Error; err != nil {
			return err
		}
		now := common.GetTimestamp()
		var inFlight int64
		if err := tx.Model(&AssetReplica{}).Where("library_id IN ? AND locked_by <> ? AND lease_until > ?", ids, "", now).Count(&inFlight).Error; err != nil {
			return err
		}
		if inFlight > 0 || len(libraries) < 2 {
			return nil
		}
		var counts []struct {
			LibraryID int
			Count     int64
		}
		if err := tx.Model(&AssetReplica{}).Where("library_id IN ? AND status = ? AND operation = ?", ids, AssetReplicaStatusActive, AssetReplicaOperationSync).
			Select("library_id, count(*) as count").Group("library_id").Find(&counts).Error; err != nil {
			return err
		}
		usable := make(map[int]int64, len(counts))
		for _, count := range counts {
			usable[count.LibraryID] = count.Count
		}
		keep, enabled, qpm := libraries[0], false, 0
		for _, library := range libraries {
			if usable[library.ID] > usable[keep.ID] {
				keep = library
			}
			enabled = enabled || library.Enabled
			if library.QPM > 0 && (qpm == 0 || library.QPM < qpm) {
				qpm = library.QPM
			}
		}
		var others []int
		for _, library := range libraries {
			if library.ID != keep.ID {
				others = append(others, library.ID)
			}
		}
		if err := tx.Model(&AssetChannelConfig{}).Where("library_id IN ?", others).Updates(map[string]any{"library_id": keep.ID, "updated_at": now}).Error; err != nil {
			return err
		}
		for _, table := range []struct {
			model  any
			column string
		}{{&AssetReplica{}, "asset_id"}, {&AssetGroupReplica{}, "group_id"}} {
			var held []int64
			if err := tx.Model(table.model).Where("library_id = ?", keep.ID).Pluck(table.column, &held).Error; err != nil {
				return err
			}
			taken := make(map[int64]bool, len(held))
			for _, id := range held {
				taken[id] = true
			}
			var candidates []struct {
				ID    int64
				Owner int64
			}
			query := tx.Model(table.model).Where("library_id IN ? AND status = ?", others, AssetReplicaStatusActive).Order("id asc").Select("id, " + table.column + " as owner")
			if _, isAsset := table.model.(*AssetReplica); isAsset {
				query = query.Where("operation = ?", AssetReplicaOperationSync)
			}
			if err := query.Find(&candidates).Error; err != nil {
				return err
			}
			var moving []int64
			for _, candidate := range candidates {
				if !taken[candidate.Owner] {
					taken[candidate.Owner] = true
					moving = append(moving, candidate.ID)
				}
			}
			for start := 0; start < len(moving); start += 500 {
				end := min(start+500, len(moving))
				if err := tx.Model(table.model).Where("id IN ?", moving[start:end]).Update("library_id", keep.ID).Error; err != nil {
					return err
				}
			}
		}
		// An account left with no copies and no channels has nothing to keep.
		for _, id := range others {
			var held int64
			if err := tx.Model(&AssetReplica{}).Where("library_id = ?", id).Count(&held).Error; err != nil {
				return err
			}
			if held == 0 {
				if err := tx.Model(&AssetGroupReplica{}).Where("library_id = ?", id).Count(&held).Error; err != nil {
					return err
				}
			}
			if held == 0 {
				if err := tx.Delete(&AssetLibraryConnection{}, id).Error; err != nil {
					return err
				}
			}
		}
		if err := tx.Model(&AssetLibraryConnection{}).Where("resource_key = ? AND id <> ?", resourceKey, keep.ID).Update("resource_key", nil).Error; err != nil {
			return err
		}
		keptID = keep.ID
		// The merged account is titled after the lowest-ID account of the group,
		// which for migrated accounts is the channel with the smallest ID.
		return tx.Model(&AssetLibraryConnection{}).Where("id = ?", keep.ID).
			Updates(map[string]any{"resource_key": resourceKey, "enabled": enabled, "qpm": qpm, "name": libraries[0].Name, "provider_name": libraries[0].ProviderName, "updated_at": now}).Error
	})
	return keptID, err
}
