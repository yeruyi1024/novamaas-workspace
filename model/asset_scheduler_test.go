package model

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func TestAssetSyncLeadershipUsesOneExistingLockAndSurvivesFailover(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db := setupAssetSchedulerDB(t, dialect)
			ctx := context.Background()
			results := make(chan string, 4)
			for _, owner := range []string{"node-a", "node-b", "node-c", "node-d"} {
				go func() {
					held, err := TryHoldAssetSyncLeadership(ctx, owner, 30)
					assert.NoError(t, err)
					if held {
						results <- owner
					} else {
						results <- ""
					}
				}()
			}
			var owners []string
			for range 4 {
				if owner := <-results; owner != "" {
					owners = append(owners, owner)
				}
			}
			require.Len(t, owners, 1, "four API nodes must elect only one background dispatcher")
			winner := owners[0]
			held, err := TryHoldAssetSyncLeadership(ctx, winner, 30)
			require.NoError(t, err)
			assert.True(t, held, "the owner can renew")
			held, err = TryHoldAssetSyncLeadership(ctx, "replacement", 30)
			require.NoError(t, err)
			assert.False(t, held)
			require.NoError(t, ExpireStaleSystemTaskLocks(common.GetTimestamp()+3600))
			var lock SystemTaskLock
			require.NoError(t, db.Where("type = ?", AssetSyncLeaderLockType).First(&lock).Error)
			assert.Equal(t, winner, lock.LockedBy, "ordinary job cleanup must not remove the scheduler's lease")
			// Expire the isolated fixture lease to exercise failover.
			require.NoError(t, db.Model(&lock).Update("locked_until", 0).Error)
			held, err = TryHoldAssetSyncLeadership(ctx, "replacement", 30)
			require.NoError(t, err)
			assert.True(t, held)
		})
	}
}

func setupAssetSchedulerDB(t *testing.T, dialect string) *gorm.DB {
	t.Helper()
	prefix := fmt.Sprintf("test_as_%d_", time.Now().UnixNano())
	var driver gorm.Dialector = sqlite.Open("file:" + prefix + "?mode=memory&cache=shared")
	databaseType := common.DatabaseTypeSQLite
	if dialect == "mysql" {
		dsn := os.Getenv("TEST_MYSQL_DSN")
		if dsn == "" {
			t.Skip("TEST_MYSQL_DSN is not configured")
		}
		config, err := mysqldriver.ParseDSN(dsn)
		require.NoError(t, err)
		require.Equal(t, "novamaas_local", config.DBName)
		driver, databaseType = mysql.Open(dsn), common.DatabaseTypeMySQL
	}
	db, err := gorm.Open(driver, &gorm.Config{NamingStrategy: schema.NamingStrategy{TablePrefix: prefix}, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	connection, err := db.DB()
	require.NoError(t, err)
	connection.SetMaxOpenConns(1)
	if dialect == "mysql" {
		connection.SetMaxOpenConns(4)
	}
	previous, previousType := DB, common.MainDatabaseType()
	DB = db
	common.SetMainDatabaseType(databaseType)
	t.Cleanup(func() {
		assert.NoError(t, db.Migrator().DropTable(&SystemTaskLock{}, &AssetReplica{}, &MediaAsset{}, &StorageObject{}, &AssetChannelConfig{}))
		DB = previous
		common.SetMainDatabaseType(previousType)
		assert.NoError(t, connection.Close())
	})
	require.NoError(t, db.AutoMigrate(&MediaAsset{}, &AssetReplica{}, &StorageObject{}, &SystemTaskLock{}, &AssetChannelConfig{}))
	return db
}

func TestAssetRequeuePreservesLiveAndCompletedCopies(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db := setupAssetSchedulerDB(t, dialect)
			now := common.GetTimestamp()
			asset := MediaAsset{PublicID: "requeue-material", StorageObjectID: 1, Status: AssetStatusReady}
			require.NoError(t, db.Create(&asset).Error)
			replicas := []AssetReplica{
				{AssetID: asset.ID, ChannelID: 1, Operation: AssetReplicaOperationSync, Status: AssetReplicaStatusSyncing, LockedBy: "creating", LeaseUntil: now + 120},
				{AssetID: asset.ID, ChannelID: 2, Operation: AssetReplicaOperationSync, Status: AssetReplicaStatusActive, UpstreamAssetID: "ready-id", Progress: 100, NextSyncAt: now + 1800},
				{AssetID: asset.ID, ChannelID: 3, Operation: AssetReplicaOperationSync, Status: AssetReplicaStatusFailed, UpstreamAssetID: "retry-id", Attempts: 3, CreatedAt: now - 301},
				{AssetID: asset.ID, ChannelID: 4, Operation: AssetReplicaOperationDelete, Status: AssetReplicaStatusDeleting, UpstreamAssetID: "deleting-id"},
			}
			require.NoError(t, db.Create(&replicas).Error)
			require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return QueueAssetReplicas(tx, asset.ID, []int{1, 2, 3, 4, 5}) }))
			var stored []AssetReplica
			require.NoError(t, db.Where("asset_id = ?", asset.ID).Order("channel_id").Find(&stored).Error)
			require.Len(t, stored, 5)
			assert.Equal(t, AssetReplicaStatusSyncing, stored[0].Status)
			assert.Equal(t, "creating", stored[0].LockedBy)
			assert.Equal(t, AssetReplicaStatusActive, stored[1].Status)
			assert.Equal(t, "ready-id", stored[1].UpstreamAssetID)
			assert.Equal(t, now+1800, stored[1].NextSyncAt)
			assert.Equal(t, AssetReplicaStatusProcessing, stored[2].Status)
			assert.Equal(t, "retry-id", stored[2].UpstreamAssetID)
			assert.Zero(t, stored[2].Attempts)
			assert.GreaterOrEqual(t, stored[2].CreatedAt, now)
			assert.Equal(t, AssetReplicaOperationDelete, stored[3].Operation)
			assert.Equal(t, AssetReplicaStatusDeleting, stored[3].Status)
			assert.Equal(t, AssetReplicaStatusPending, stored[4].Status)
			require.NoError(t, FinishAssetReplicaSubmission(stored[0].ID, "creating", "returned-id", now+8))
			require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return RequestAssetDeletion(tx, asset.ID) }))
			require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return QueueAssetReplicas(tx, asset.ID, []int{6}) }))
			var added int64
			require.NoError(t, db.Model(&AssetReplica{}).Where("asset_id = ? AND channel_id = 6", asset.ID).Count(&added).Error)
			assert.Zero(t, added, "requeue must not revive a deleted material")
		})
	}
}

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
