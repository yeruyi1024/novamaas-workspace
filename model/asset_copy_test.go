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
