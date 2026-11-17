package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSharedAssetFileIsReclaimedOnlyAfterItsLastAssetIsDeleted(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db := setupAssetSchedulerDB(t, dialect)
			object := StorageObject{ObjectID: "copy-object", ObjectKey: "assets/shared", Status: StorageObjectStatusUploaded, Purpose: StorageObjectPurposeAssetLibrary}
			require.NoError(t, db.Create(&object).Error)
			// Exercise the upgrade from the former one-asset-per-file schema.
			table := db.NamingStrategy.TableName("MediaAsset")
			require.NoError(t, db.Exec("CREATE UNIQUE INDEX idx_media_assets_storage_object_id ON "+table+" (storage_object_id)").Error)
			require.NoError(t, MigrateAssetCopyReferences(db))
			require.NoError(t, MigrateAssetCopyReferences(db))
			var ids []int64
			for _, publicID := range []string{"asset-original", "asset-copy-b", "asset-copy-c"} {
				asset := MediaAsset{PublicID: publicID, OwnerUserID: len(ids) + 1, StorageObjectID: object.ID, Status: AssetStatusReady}
				require.NoError(t, db.Create(&asset).Error)
				ids = append(ids, asset.ID)
			}
			require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return RequestAssetDeletion(tx, ids[0]) }))
			require.NoError(t, db.First(&object, object.ID).Error)
			assert.Equal(t, StorageObjectStatusUploaded, object.Status, "copies still use the file")
			// The two remaining owners delete concurrently; exactly one finishes last.
			start := make(chan struct{})
			results := make(chan error, 2)
			for _, id := range ids[1:] {
				go func() {
					<-start
					results <- db.Transaction(func(tx *gorm.DB) error { return RequestAssetDeletion(tx, id) })
				}()
			}
			close(start)
			require.NoError(t, <-results)
			require.NoError(t, <-results)
			require.NoError(t, db.First(&object, object.ID).Error)
			assert.Equal(t, StorageObjectStatusDeletePending, object.Status)
		})
	}
}

func TestSharedAssetFileWaitsForEveryUpstreamCopyDeletion(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db := setupAssetSchedulerDB(t, dialect)
			object := StorageObject{ObjectID: "shared-upstream-cleanup", ObjectKey: "assets/shared-upstream", Status: StorageObjectStatusUploaded, Purpose: StorageObjectPurposeAssetLibrary}
			require.NoError(t, db.Create(&object).Error)
			var replicas []AssetReplica
			for _, publicID := range []string{"asset-source", "asset-destination"} {
				asset := MediaAsset{PublicID: publicID, StorageObjectID: object.ID, Status: AssetStatusReady}
				require.NoError(t, db.Create(&asset).Error)
				replica := AssetReplica{AssetID: asset.ID, ChannelID: 1, Operation: AssetReplicaOperationSync, Status: AssetReplicaStatusActive, UpstreamAssetID: publicID + "-upstream"}
				require.NoError(t, db.Create(&replica).Error)
				replicas = append(replicas, replica)
				require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return RequestAssetDeletion(tx, asset.ID) }))
			}
			for index := len(replicas) - 1; index >= 0; index-- {
				replica := replicas[index]
				require.NoError(t, db.Model(&replica).Updates(map[string]any{"status": AssetReplicaStatusSyncing, "locked_by": "delete-worker"}).Error)
				require.NoError(t, AcknowledgeAssetReplicaDeletion(replica.ID, "delete-worker", replica.UpstreamAssetID))
				require.NoError(t, db.First(&object, object.ID).Error)
				if index > 0 {
					assert.Equal(t, StorageObjectStatusUploaded, object.Status, "another deleted asset still has an upstream copy")
				} else {
					assert.Equal(t, StorageObjectStatusDeletePending, object.Status)
				}
			}
		})
	}
}

func TestSharedAssetConcurrentUpstreamCleanupReclaimsFile(t *testing.T) {
	db := setupAssetSchedulerDB(t, "mysql")
	object := StorageObject{ObjectID: "concurrent-cleanup", Status: StorageObjectStatusUploaded, Purpose: StorageObjectPurposeAssetLibrary}
	require.NoError(t, db.Create(&object).Error)
	var replicas []AssetReplica
	for _, publicID := range []string{"concurrent-a", "concurrent-b"} {
		asset := MediaAsset{PublicID: publicID, StorageObjectID: object.ID, Status: AssetStatusDeleted}
		require.NoError(t, db.Create(&asset).Error)
		replica := AssetReplica{AssetID: asset.ID, ChannelID: 1, Operation: AssetReplicaOperationDelete, Status: AssetReplicaStatusSyncing, LockedBy: "worker", UpstreamAssetID: publicID}
		require.NoError(t, db.Create(&replica).Error)
		replicas = append(replicas, replica)
	}
	// Both acknowledgements read before either commits; file-lock serialization
	// must still let the last acknowledgement observe the first one's cleanup.
	ready, release := make(chan struct{}, 2), make(chan struct{})
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("cleanup_snapshot_barrier", func(tx *gorm.DB) {
		if len(tx.Statement.Selects) == 1 && tx.Statement.Selects[0] == "asset_id" {
			ready <- struct{}{}
			<-release
		}
	}))
	t.Cleanup(func() { db.Callback().Query().Remove("cleanup_snapshot_barrier") })
	results := make(chan error, 2)
	for _, replica := range replicas {
		go func() { results <- AcknowledgeAssetReplicaDeletion(replica.ID, "worker", replica.UpstreamAssetID) }()
	}
	<-ready
	<-ready
	close(release)
	require.NoError(t, <-results)
	require.NoError(t, <-results)
	require.NoError(t, db.First(&object, object.ID).Error)
	assert.Equal(t, StorageObjectStatusDeletePending, object.Status)
}
