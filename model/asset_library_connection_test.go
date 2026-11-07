package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Recreate the channel-only schema that existed before shared libraries.
type legacyAssetReplicaSchema struct {
	ID              int64  `gorm:"primaryKey"`
	AssetID         int64  `gorm:"uniqueIndex:idx_asset_channel,priority:1"`
	ChannelID       int    `gorm:"uniqueIndex:idx_asset_channel,priority:2;index"`
	UpstreamAssetID string `gorm:"type:varchar(255)"`
	Status          string `gorm:"type:varchar(32)"`
}

func TestSharedLibrarySchemaUpgradePreservesLegacyRowsAndAllowsMultipleNullChannelCopies(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db := setupAssetSchedulerDB(t, dialect)
			t.Cleanup(func() {
				assert.NoError(t, db.Migrator().DropTable(&AssetLibraryConnection{}, &AssetGroupReplica{}, &Channel{}))
			})
			require.NoError(t, db.Migrator().DropTable(&AssetReplica{}))
			table := db.NamingStrategy.TableName("AssetReplica")
			require.NoError(t, db.Table(table).AutoMigrate(&legacyAssetReplicaSchema{}))
			require.NoError(t, db.Table(table).Create(&legacyAssetReplicaSchema{ID: 91, AssetID: 81, ChannelID: 11, UpstreamAssetID: "original-upstream-id", Status: AssetReplicaStatusActive}).Error)
			require.NoError(t, db.AutoMigrate(&AssetReplica{}, &AssetLibraryConnection{}, &AssetGroupReplica{}, &Channel{}))
			require.NoError(t, db.Create(&Channel{Id: 11, Name: "legacy pricing channel"}).Error)
			require.NoError(t, db.Create(&AssetChannelConfig{ChannelID: 11, Enabled: true, Protocol: AssetChannelProtocolVolcAction, EncryptedCredential: "retained-ciphertext"}).Error)
			require.NoError(t, MigrateAssetLibraryConnections(db))
			require.NoError(t, MigrateAssetLibraryConnections(db))
			var migrated AssetReplica
			require.NoError(t, db.First(&migrated, 91).Error)
			assert.Equal(t, "original-upstream-id", migrated.UpstreamAssetID)
			assert.Equal(t, AssetReplicaStatusActive, migrated.Status)
			require.NotNil(t, migrated.LibraryID)
			var roots int64
			require.NoError(t, db.Model(&AssetLibraryConnection{}).Count(&roots).Error)
			assert.EqualValues(t, 1, roots)
			var first AssetLibraryConnection
			require.NoError(t, db.First(&first, *migrated.LibraryID).Error)
			second := AssetLibraryFromConfig("same account / second pricing channel", first.ProviderConfig(0))
			require.NoError(t, db.Create(&second).Error)
			asset := MediaAsset{PublicID: "asset-after-upgrade", StorageObjectID: 91, Status: AssetStatusReady}
			require.NoError(t, db.Create(&asset).Error)
			for _, id := range []int{*migrated.LibraryID, second.ID} {
				require.NoError(t, db.Create(&AssetReplica{AssetID: asset.ID, LibraryID: &id, Status: AssetReplicaStatusPending}).Error)
			}
			var copies []struct{ ChannelID *int }
			require.NoError(t, db.Model(&AssetReplica{}).Where("asset_id = ?", asset.ID).Select("channel_id").Find(&copies).Error)
			require.Len(t, copies, 2)
			assert.Nil(t, copies[0].ChannelID)
			assert.Nil(t, copies[1].ChannelID)
		})
	}
}
