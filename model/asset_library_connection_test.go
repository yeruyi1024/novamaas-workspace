package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"

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

func TestAssetLibraryDeletionKeepsPlatformAssetsAndInFlightSubmission(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db := setupAssetSchedulerDB(t, dialect)
			require.NoError(t, db.AutoMigrate(&AssetLibraryConnection{}, &AssetGroupReplica{}, &Channel{}))
			t.Cleanup(func() {
				assert.NoError(t, db.Migrator().DropTable(&AssetLibraryConnection{}, &AssetGroupReplica{}, &Channel{}))
			})
			library := AssetLibraryConnection{Name: "retired", Enabled: true}
			other := AssetLibraryConnection{Name: "current", Enabled: true}
			require.NoError(t, db.Create(&library).Error)
			require.NoError(t, db.Create(&other).Error)
			require.NoError(t, db.Create(&AssetChannelConfig{ChannelID: 999, LibraryID: library.ID}).Error)
			object := StorageObject{ObjectID: "preserved-source", Status: StorageObjectStatusUploaded}
			require.NoError(t, db.Create(&object).Error)
			asset := MediaAsset{PublicID: "kept-platform-asset", StorageObjectID: object.ID, Status: AssetStatusReady}
			require.NoError(t, db.Create(&asset).Error)
			now := common.GetTimestamp()
			pending := AssetReplica{AssetID: asset.ID, LibraryID: &library.ID, Operation: AssetReplicaOperationSync,
				Status: AssetReplicaStatusSyncing, LockedBy: "upload-owner", LeaseUntil: now + 120}
			kept := AssetReplica{AssetID: asset.ID, LibraryID: &other.ID, Operation: AssetReplicaOperationSync, Status: AssetReplicaStatusActive, UpstreamAssetID: "kept-upstream"}
			require.NoError(t, db.Create(&pending).Error)
			require.NoError(t, db.Create(&kept).Error)
			require.NoError(t, RequestAssetLibraryDeletion(library.ID))
			finished, err := FinalizeAssetLibraryDeletion(library.ID)
			require.NoError(t, err)
			assert.False(t, finished, "keep credentials while an upload may still return an upstream ID")
			require.NoError(t, db.First(&pending, pending.ID).Error)
			assert.Equal(t, "upload-owner", pending.LockedBy)
			assert.Equal(t, AssetReplicaOperationDelete, pending.Operation)
			require.NoError(t, FinishAssetReplicaSubmission(pending.ID, "upload-owner", "late-upstream", now))
			require.NoError(t, db.First(&pending, pending.ID).Error)
			assert.Equal(t, "late-upstream", pending.UpstreamAssetID)
			now = common.GetTimestamp()
			claimed, err := ClaimAssetReplica(&pending, now, now+120, "delete-owner")
			require.NoError(t, err)
			require.True(t, claimed)
			require.NoError(t, AcknowledgeAssetReplicaDeletion(pending.ID, "delete-owner", "late-upstream"))
			finished, err = FinalizeAssetLibraryDeletion(library.ID)
			require.NoError(t, err)
			assert.True(t, finished)
			var bindings int64
			require.NoError(t, db.Model(&AssetChannelConfig{}).Where("library_id = ?", library.ID).Count(&bindings).Error)
			assert.Zero(t, bindings, "bindings to channels that were already deleted do not block cleanup")
			require.NoError(t, db.First(&asset, asset.ID).Error)
			require.NoError(t, db.First(&object, object.ID).Error)
			require.NoError(t, db.First(&kept, kept.ID).Error)
			assert.Equal(t, AssetStatusReady, asset.Status)
			assert.Equal(t, StorageObjectStatusUploaded, object.Status)
			assert.Equal(t, AssetReplicaStatusActive, kept.Status)
		})
	}
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

func TestMergeAssetLibraryGroupSwapsCopiesWithoutLosingUpstreamState(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db := setupAssetSchedulerDB(t, dialect)
			require.NoError(t, db.AutoMigrate(&AssetLibraryConnection{}, &AssetGroupReplica{}))
			t.Cleanup(func() {
				assert.NoError(t, db.Migrator().DropTable(&AssetLibraryConnection{}, &AssetGroupReplica{}))
			})
			first := AssetLibraryConnection{Name: "kept", Enabled: true, QPM: 30}
			second := AssetLibraryConnection{Name: "retired", Enabled: true, QPM: 60}
			require.NoError(t, db.Create(&first).Error)
			require.NoError(t, db.Create(&second).Error)
			failed := AssetReplica{AssetID: 1, LibraryID: &first.ID, Status: AssetReplicaStatusFailed, Operation: AssetReplicaOperationSync,
				UpstreamAssetID: "failed-upstream", LastError: "provider failed", Attempts: 3}
			working := AssetReplica{AssetID: 1, LibraryID: &second.ID, Status: AssetReplicaStatusActive, Operation: AssetReplicaOperationSync, UpstreamAssetID: "working-upstream"}
			require.NoError(t, db.Create(&failed).Error)
			require.NoError(t, db.Create(&working).Error)
			for _, id := range []int64{2, 3} {
				require.NoError(t, db.Create(&AssetReplica{AssetID: id, LibraryID: &first.ID, Status: AssetReplicaStatusActive, Operation: AssetReplicaOperationSync}).Error)
			}
			for _, library := range []AssetLibraryConnection{first, second} {
				status := AssetReplicaStatusActive
				if library.ID == first.ID {
					status = AssetReplicaStatusFailed
				}
				require.NoError(t, db.Create(&AssetGroupReplica{GroupID: 1, LibraryID: &library.ID, Status: status, UpstreamGroupID: library.Name}).Error)
			}
			keptID, err := MergeAssetLibraryGroup([]int{first.ID, second.ID}, "fixture-resource")
			require.NoError(t, err)
			require.Equal(t, first.ID, keptID)
			require.NoError(t, db.First(&working, working.ID).Error)
			require.NoError(t, db.First(&failed, failed.ID).Error)
			assert.Equal(t, first.ID, *working.LibraryID)
			assert.Equal(t, "working-upstream", working.UpstreamAssetID)
			assert.Equal(t, second.ID, *failed.LibraryID)
			assert.Equal(t, "failed-upstream", failed.UpstreamAssetID)
			assert.Equal(t, "provider failed", failed.LastError)
			assert.Equal(t, 3, failed.Attempts)
			var group AssetGroupReplica
			require.NoError(t, db.Where("library_id = ?", first.ID).First(&group).Error)
			assert.Equal(t, "retired", group.UpstreamGroupID)
			assert.Equal(t, AssetReplicaStatusActive, group.Status)
		})
	}
}
