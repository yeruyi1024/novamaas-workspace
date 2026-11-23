package model

import (
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
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

func TestMediaAssetPublicIDScopedPerOwnerWithSoftDeleteReentry(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db := setupAssetSchedulerDB(t, dialect)
			require.NoError(t, db.AutoMigrate(&AssetGroup{}))
			t.Cleanup(func() { assert.NoError(t, db.Migrator().DropTable(&AssetGroup{})) })

			assetTable := db.NamingStrategy.TableName("MediaAsset")
			groupTable := db.NamingStrategy.TableName("AssetGroup")
			// Simulate upgrading from a legacy schema with global unique public_id
			// and historical deleted rows where deleted_slot was 0.
			require.NoError(t, db.Exec("CREATE UNIQUE INDEX idx_media_assets_public_id ON "+assetTable+" (public_id)").Error)
			require.NoError(t, db.Exec("CREATE UNIQUE INDEX idx_asset_groups_public_id ON "+groupTable+" (public_id)").Error)
			historicalDeletedGroup := AssetGroup{PublicID: "group-reenter", OwnerUserID: 2, Name: "Old", Status: AssetStatusDeleted, DeletedSlot: 0}
			require.NoError(t, db.Create(&historicalDeletedGroup).Error)
			historicalDeleted := MediaAsset{PublicID: "asset-reenter", OwnerUserID: 2, StorageObjectID: 10, Status: AssetStatusDeleted, DeletedSlot: 0}
			require.NoError(t, db.Create(&historicalDeleted).Error)
			require.NoError(t, MigrateAssetCopyReferences(db))
			require.NoError(t, MigrateAssetCopyReferences(db))

			var reloadedDeletedGroup AssetGroup
			require.NoError(t, db.First(&reloadedDeletedGroup, historicalDeletedGroup.ID).Error)
			assert.Equal(t, historicalDeletedGroup.ID, reloadedDeletedGroup.DeletedSlot)

			var reloadedDeleted MediaAsset
			require.NoError(t, db.First(&reloadedDeleted, historicalDeleted.ID).Error)
			assert.Equal(t, historicalDeleted.ID, reloadedDeleted.DeletedSlot)

			// Same group public_id can coexist across different owners, while active duplicates under one owner fail.
			groupA := AssetGroup{PublicID: "group-shared-id", OwnerUserID: 1, Name: "Campaign", Status: AssetStatusReady}
			groupB := AssetGroup{PublicID: "group-shared-id", OwnerUserID: 2, Name: "Campaign", Status: AssetStatusReady}
			require.NoError(t, db.Create(&groupA).Error)
			require.NoError(t, db.Create(&groupB).Error)
			require.Error(t, db.Create(&AssetGroup{PublicID: "group-shared-id", OwnerUserID: 2, Name: "Dup", Status: AssetStatusReady}).Error)
			require.NoError(t, db.Create(&AssetGroup{PublicID: "group-reenter", OwnerUserID: 2, Name: "New", Status: AssetStatusReady}).Error)

			_, err := FindAssetGroup("group-shared-id", 99, true)
			require.ErrorIs(t, err, ErrAssetAmbiguousPublicID)
			foundGroupB, err := FindAssetGroupForOwner("group-shared-id", 99, true, 2)
			require.NoError(t, err)
			assert.Equal(t, groupB.ID, foundGroupB.ID)

			// Same asset public_id can coexist across different owners.
			ownerA := MediaAsset{PublicID: "asset-shared-id", OwnerUserID: 1, GroupID: groupA.ID, StorageObjectID: 11, Status: AssetStatusReady}
			ownerB := MediaAsset{PublicID: "asset-shared-id", OwnerUserID: 2, GroupID: groupB.ID, StorageObjectID: 11, Status: AssetStatusReady}
			require.NoError(t, db.Create(&ownerA).Error)
			require.NoError(t, db.Create(&ownerB).Error)

			// Same owner cannot hold two active rows with the same public_id.
			duplicate := MediaAsset{PublicID: "asset-shared-id", OwnerUserID: 2, StorageObjectID: 12, Status: AssetStatusReady}
			require.Error(t, db.Create(&duplicate).Error)

			// Historical deleted row does not block creating a new active row for the same owner.
			reentered := MediaAsset{PublicID: "asset-reenter", OwnerUserID: 2, StorageObjectID: 13, Status: AssetStatusReady}
			require.NoError(t, db.Create(&reentered).Error)

			// Admin lookup without targetOwnerUserID rejects ambiguous shared public_id,
			// while specifying owner_user_id or non-admin lookup resolves deterministically.
			foundA, err := FindMediaAsset("asset-shared-id", 1, false)
			require.NoError(t, err)
			assert.Equal(t, ownerA.ID, foundA.ID)

			_, err = FindMediaAsset("asset-shared-id", 99, true)
			require.ErrorIs(t, err, ErrAssetAmbiguousPublicID)

			foundB, err := FindMediaAssetForOwner("asset-shared-id", 99, true, 2)
			require.NoError(t, err)
			assert.Equal(t, ownerB.ID, foundB.ID)

			// Soft-deleting ownerB frees the active slot so the same public_id can be copied again.
			require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
				return RequestAssetDeletion(tx, ownerB.ID)
			}, AssetDeletionTxOptions()))
			var deletedB MediaAsset
			require.NoError(t, db.First(&deletedB, ownerB.ID).Error)
			assert.Equal(t, AssetStatusDeleted, deletedB.Status)
			assert.Equal(t, ownerB.ID, deletedB.DeletedSlot)

			recopiedB := MediaAsset{PublicID: "asset-shared-id", OwnerUserID: 2, StorageObjectID: 14, Status: AssetStatusReady}
			require.NoError(t, db.Create(&recopiedB).Error)
		})
	}
}

func TestPurgeAndUserDeletionCleanUpAssetsWithoutHarmingSharedCopies(t *testing.T) {
	for _, dialect := range []string{"sqlite", "mysql"} {
		t.Run(dialect, func(t *testing.T) {
			db := setupAssetSchedulerDB(t, dialect)
			require.NoError(t, db.AutoMigrate(&User{}, &AssetGroup{}, &AssetAccessKey{}, &AssetWebhookEndpoint{},
				&ExternalIdentityClaim{}, &VerifiedPhone{}, &WeChatMiniAppProfile{}, &SMSChallenge{},
				&TwoFABackupCode{}, &TwoFA{}, &UserSession{}, &LoginNoticeAcknowledgement{}, &AuthFlow{},
				&PasskeyCredential{}, &Token{}, &UserOAuthBinding{}))
			t.Cleanup(func() {
				assert.NoError(t, db.Migrator().DropTable(&User{}, &AssetGroup{}, &AssetAccessKey{}, &AssetWebhookEndpoint{},
					&ExternalIdentityClaim{}, &VerifiedPhone{}, &WeChatMiniAppProfile{}, &SMSChallenge{},
					&TwoFABackupCode{}, &TwoFA{}, &UserSession{}, &LoginNoticeAcknowledgement{}, &AuthFlow{},
					&PasskeyCredential{}, &Token{}, &UserOAuthBinding{}))
			})
			previousRedis := common.RedisEnabled
			common.RedisEnabled = false
			t.Cleanup(func() { common.RedisEnabled = previousRedis })

			root := User{Id: 100, Username: "purge-root", AffCode: "purge-root", Role: common.RoleRootUser, Status: common.UserStatusEnabled}
			userA := User{Id: 1, Username: "purge-a", AffCode: "purge-a", Role: common.RoleCommonUser, Status: common.UserStatusEnabled}
			userB := User{Id: 2, Username: "purge-b", AffCode: "purge-b", Role: common.RoleCommonUser, Status: common.UserStatusEnabled}
			userC := User{Id: 3, Username: "purge-c", AffCode: "purge-c", Role: common.RoleCommonUser, Status: common.UserStatusEnabled}
			for _, u := range []User{root, userA, userB, userC} {
				require.NoError(t, db.Create(&u).Error)
			}

			sharedObj := StorageObject{ObjectID: "purge-shared", ObjectKey: "assets/purge-shared", Status: StorageObjectStatusUploaded, Purpose: StorageObjectPurposeAssetLibrary}
			soloObjB := StorageObject{ObjectID: "purge-solo-b", ObjectKey: "assets/purge-solo-b", Status: StorageObjectStatusUploaded, Purpose: StorageObjectPurposeAssetLibrary}
			soloObjC := StorageObject{ObjectID: "purge-solo-c", ObjectKey: "assets/purge-solo-c", Status: StorageObjectStatusUploaded, Purpose: StorageObjectPurposeAssetLibrary}
			for _, obj := range []*StorageObject{&sharedObj, &soloObjB, &soloObjC} {
				require.NoError(t, db.Create(obj).Error)
			}

			groupA := AssetGroup{PublicID: "group-shared", OwnerUserID: userA.Id, Name: "G", Status: AssetStatusReady}
			groupB := AssetGroup{PublicID: "group-shared", OwnerUserID: userB.Id, Name: "G", Status: AssetStatusReady}
			groupC := AssetGroup{PublicID: "group-c", OwnerUserID: userC.Id, Name: "GC", Status: AssetStatusReady}
			for _, g := range []*AssetGroup{&groupA, &groupB, &groupC} {
				require.NoError(t, db.Create(g).Error)
			}

			assetA := MediaAsset{PublicID: "asset-shared", OwnerUserID: userA.Id, GroupID: groupA.ID, StorageObjectID: sharedObj.ID, Status: AssetStatusReady}
			assetBShared := MediaAsset{PublicID: "asset-shared", OwnerUserID: userB.Id, GroupID: groupB.ID, StorageObjectID: sharedObj.ID, Status: AssetStatusReady}
			assetBSolo := MediaAsset{PublicID: "asset-solo-b", OwnerUserID: userB.Id, GroupID: groupB.ID, StorageObjectID: soloObjB.ID, Status: AssetStatusReady}
			assetCSolo := MediaAsset{PublicID: "asset-solo-c", OwnerUserID: userC.Id, GroupID: groupC.ID, StorageObjectID: soloObjC.ID, Status: AssetStatusReady}
			for _, a := range []*MediaAsset{&assetA, &assetBShared, &assetBSolo, &assetCSolo} {
				require.NoError(t, db.Create(a).Error)
			}
			replicaB := AssetReplica{AssetID: assetBShared.ID, ChannelID: 1, Operation: AssetReplicaOperationSync, Status: AssetReplicaStatusActive, UpstreamAssetID: "up-b"}
			require.NoError(t, db.Create(&replicaB).Error)

			keyB := AssetAccessKey{OwnerUserID: userB.Id, Name: "k-b", AccessKeyID: "ak-b", Status: AssetAccessKeyStatusEnabled}
			keyC := AssetAccessKey{OwnerUserID: userC.Id, Name: "k-c", AccessKeyID: "ak-c", Status: AssetAccessKeyStatusEnabled}
			hookC := AssetWebhookEndpoint{PublicID: "we-c", OwnerUserID: userC.Id, Name: "h-c", URL: "https://example.com", Status: AssetWebhookEndpointStatusEnabled}
			require.NoError(t, db.Create(&keyB).Error)
			require.NoError(t, db.Create(&keyC).Error)
			require.NoError(t, db.Create(&hookC).Error)

			// 1. Admin purge for userB clears assets and groups, keeps AK/SK, preserves shared file for userA, reclaims solo file.
			purged, err := PurgeUserAssets(userB.Id, root.Id)
			require.NoError(t, err)
			assert.Equal(t, 2, purged.DeletedAssetCount)
			assert.Equal(t, 1, purged.DeletedGroupCount)

			var refreshedShared, refreshedSoloB StorageObject
			require.NoError(t, db.First(&refreshedShared, sharedObj.ID).Error)
			assert.Equal(t, StorageObjectStatusUploaded, refreshedShared.Status)
			require.NoError(t, db.First(&refreshedSoloB, soloObjB.ID).Error)
			assert.Equal(t, StorageObjectStatusDeletePending, refreshedSoloB.Status)

			var refreshedReplicaB AssetReplica
			require.NoError(t, db.First(&refreshedReplicaB, replicaB.ID).Error)
			assert.Equal(t, AssetReplicaOperationDelete, refreshedReplicaB.Operation)
			assert.Equal(t, AssetReplicaStatusDeleting, refreshedReplicaB.Status)

			var keyCountB int64
			require.NoError(t, db.Model(&AssetAccessKey{}).Where("owner_user_id = ?", userB.Id).Count(&keyCountB).Error)
			assert.Equal(t, int64(1), keyCountB, "purge keeps user AK/SK keys")

			// 2. Hard deleting userC automatically purges userC's assets and groups while leaving AK/SK and webhooks untouched.
			require.NoError(t, HardDeleteUserById(userC.Id))
			var deletedC MediaAsset
			require.NoError(t, db.First(&deletedC, assetCSolo.ID).Error)
			assert.Equal(t, AssetStatusDeleted, deletedC.Status)
			assert.Equal(t, assetCSolo.ID, deletedC.DeletedSlot)
			var deletedGroupC AssetGroup
			require.NoError(t, db.First(&deletedGroupC, groupC.ID).Error)
			assert.Equal(t, AssetStatusDeleted, deletedGroupC.Status)
			assert.Equal(t, groupC.ID, deletedGroupC.DeletedSlot)
			var refreshedSoloC StorageObject
			require.NoError(t, db.First(&refreshedSoloC, soloObjC.ID).Error)
			assert.Equal(t, StorageObjectStatusDeletePending, refreshedSoloC.Status)
			var keyCountC, hookCountC int64
			require.NoError(t, db.Model(&AssetAccessKey{}).Where("owner_user_id = ?", userC.Id).Count(&keyCountC).Error)
			require.NoError(t, db.Model(&AssetWebhookEndpoint{}).Where("owner_user_id = ?", userC.Id).Count(&hookCountC).Error)
			assert.Equal(t, int64(1), keyCountC, "user deletion does not remove AK/SK keys")
			assert.Equal(t, int64(1), hookCountC, "user deletion does not remove webhook endpoints")

			// 3. Soft deleting userA also purges userA's remaining assets and groups.
			require.NoError(t, DeleteUserById(userA.Id))
			var deletedA MediaAsset
			require.NoError(t, db.First(&deletedA, assetA.ID).Error)
			assert.Equal(t, AssetStatusDeleted, deletedA.Status)
			assert.Equal(t, assetA.ID, deletedA.DeletedSlot)

			// 4. Batched purge handles accounts exceeding one batch (>100 assets) cleanly.
			bulkUser := User{Id: 4, Username: "purge-bulk", AffCode: "purge-bulk", Role: common.RoleCommonUser, Status: common.UserStatusEnabled}
			require.NoError(t, db.Create(&bulkUser).Error)
			bulkGroup := AssetGroup{PublicID: "group-bulk", OwnerUserID: bulkUser.Id, Name: "Bulk", Status: AssetStatusReady}
			require.NoError(t, db.Create(&bulkGroup).Error)
			bulkAssets := make([]MediaAsset, 0, assetPurgeBatchSize+5)
			for i := range assetPurgeBatchSize + 5 {
				bulkAssets = append(bulkAssets, MediaAsset{
					PublicID:        fmt.Sprintf("asset-bulk-%d", i),
					OwnerUserID:     bulkUser.Id,
					GroupID:         bulkGroup.ID,
					StorageObjectID: sharedObj.ID,
					Status:          AssetStatusReady,
				})
			}
			require.NoError(t, db.CreateInBatches(&bulkAssets, 50).Error)
			bulkResult, err := PurgeUserAssets(bulkUser.Id, root.Id)
			require.NoError(t, err)
			assert.Equal(t, assetPurgeBatchSize+5, bulkResult.DeletedAssetCount)
			assert.Equal(t, 1, bulkResult.DeletedGroupCount)
			var remainingBulk int64
			require.NoError(t, db.Model(&MediaAsset{}).Where("owner_user_id = ? AND status <> ?", bulkUser.Id, AssetStatusDeleted).Count(&remainingBulk).Error)
			assert.Zero(t, remainingBulk)
		})
	}
}
