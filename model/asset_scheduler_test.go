package model

import (
	"context"
	"fmt"
	"os"
	"sync"
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
			var group sync.WaitGroup
			winners := make(chan string, 4)
			for _, owner := range []string{"node-a", "node-b", "node-c", "node-d"} {
				group.Add(1)
				go func() {
					defer group.Done()
					held, err := TryHoldAssetSyncLeadership(ctx, owner, 30)
					assert.NoError(t, err)
					if held {
						winners <- owner
					}
				}()
			}
			group.Wait()
			close(winners)
			var owners []string
			for owner := range winners {
				owners = append(owners, owner)
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
			// Advance only this isolated fixture, without sleeping or touching a
			// live scheduler's lease in the development database.
			require.NoError(t, db.Model(&lock).Update("locked_until", 0).Error)
			held, err = TryHoldAssetSyncLeadership(ctx, "replacement", 30)
			require.NoError(t, err)
			assert.True(t, held)
			held, err = TryHoldAssetSyncLeadership(ctx, winner, 30)
			require.NoError(t, err)
			assert.False(t, held, "the former owner cannot extend the replacement's lease")
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			held, err = TryHoldAssetSyncLeadership(canceled, "failed-db-check", 30)
			assert.Error(t, err)
			assert.False(t, held, "a failed ownership check must never allow dispatch")
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
		assert.NoError(t, db.Migrator().DropTable(&SystemTaskLock{}, &AssetReplica{}, &MediaAsset{}, &StorageObject{}))
		DB = previous
		common.SetMainDatabaseType(previousType)
		assert.NoError(t, connection.Close())
	})
	require.NoError(t, db.AutoMigrate(&MediaAsset{}, &AssetReplica{}, &StorageObject{}, &SystemTaskLock{}))
	return db
}

func TestAssetSchedulerLeasesAndLateDeletion(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db := setupAssetSchedulerDB(t, dialect)
			now := common.GetTimestamp()
			object := StorageObject{ObjectID: "source", Purpose: StorageObjectPurposeAssetLibrary, Status: StorageObjectStatusUploaded}
			require.NoError(t, db.Create(&object).Error)
			asset := MediaAsset{PublicID: "material", StorageObjectID: object.ID, Status: AssetStatusProcessing}
			require.NoError(t, db.Create(&asset).Error)
			replicas := []AssetReplica{
				{AssetID: asset.ID, ChannelID: 1, Status: AssetReplicaStatusActive, Operation: AssetReplicaOperationSync, UpstreamAssetID: "known-id", Progress: 100},
				{AssetID: asset.ID, ChannelID: 2, Status: AssetReplicaStatusPending, Operation: AssetReplicaOperationSync},
				{AssetID: asset.ID, ChannelID: 3, Status: AssetReplicaStatusFailed, Operation: AssetReplicaOperationSync, LastError: "timed out"},
			}
			require.NoError(t, db.Create(&replicas).Error)
			review, err := ListDueAssetReplicas(now, 1, AssetReplicaWorkloadReview, 5)
			require.NoError(t, err)
			require.Len(t, review, 1)
			claimed, err := ClaimAssetReplica(review[0], now, now+120, "review")
			require.NoError(t, err)
			require.True(t, claimed)
			assert.Equal(t, AssetReplicaStatusActive, review[0].Status)
			claimed, err = ClaimAssetReplica(&replicas[0], now, now+120, "other-review")
			require.NoError(t, err)
			assert.False(t, claimed, "an active review must not be leased twice")
			failed, err := ListDueAssetReplicas(now, 3, AssetReplicaWorkloadSubmit, 5)
			require.NoError(t, err)
			assert.Empty(t, failed, "technical terminal failures must not restart automatically")

			submit, err := ListDueAssetReplicas(now, 2, AssetReplicaWorkloadSubmit, 5)
			require.NoError(t, err)
			require.Len(t, submit, 1)
			claimed, err = ClaimAssetReplica(submit[0], now, now+120, "creating")
			require.NoError(t, err)
			require.True(t, claimed)
			require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return RequestAssetDeletion(tx, asset.ID) }))
			var source StorageObject
			require.NoError(t, db.First(&source, object.ID).Error)
			assert.Equal(t, StorageObjectStatusUploaded, source.Status, "live creation still needs the source")
			require.NoError(t, FinishAssetReplica(review[0].ID, "review", AssetReplicaStatusActive, 100, "known-id", now+1800, ""))
			require.NoError(t, FinishAssetReplicaSubmission(submit[0].ID, "creating", "late-id", now+8))
			for _, channelID := range []int{1, 2} {
				deleting, err := ListDueAssetReplicas(now+1, channelID, AssetReplicaWorkloadMaintenance, 5)
				require.NoError(t, err)
				require.Len(t, deleting, 1)
				claimed, err := ClaimAssetReplica(deleting[0], now+1, now+121, "deleting")
				require.NoError(t, err)
				require.True(t, claimed)
				assert.Equal(t, map[int]string{1: "known-id", 2: "late-id"}[channelID], deleting[0].UpstreamAssetID)
				require.NoError(t, AcknowledgeAssetReplicaDeletion(deleting[0].ID, "deleting", deleting[0].UpstreamAssetID))
			}
			require.NoError(t, db.First(&source, object.ID).Error)
			assert.Equal(t, StorageObjectStatusDeletePending, source.Status)
			claimed, err = ClaimAssetReplica(&replicas[1], now+1, now+121, "stale-candidate")
			require.NoError(t, err)
			assert.False(t, claimed, "a prefetched submission must not overwrite deletion")
		})
	}
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
