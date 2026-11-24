package assetlibrary

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type bulkAssetSeed struct {
	publicID  string
	owner     int
	groupID   int64
	assetType string
	status    string
}

func seedBulkAssets(t *testing.T, db *gorm.DB, seeds []bulkAssetSeed) {
	t.Helper()
	for index, seed := range seeds {
		object := model.StorageObject{
			ObjectID: fmt.Sprintf("bulk-object-%d-%s-%d", seed.owner, seed.publicID, index), OwnerUserID: seed.owner, ObjectKey: fmt.Sprintf("assets/bulk-%d-%s", seed.owner, seed.publicID),
			Status: model.StorageObjectStatusUploaded, Purpose: model.StorageObjectPurposeAssetLibrary,
		}
		require.NoError(t, db.Create(&object).Error)
		asset := model.MediaAsset{
			PublicID: seed.publicID, OwnerUserID: seed.owner, GroupID: seed.groupID, StorageObjectID: object.ID,
			Name: seed.publicID, AssetType: seed.assetType, Status: seed.status, CreatedAt: int64(index + 1),
		}
		require.NoError(t, db.Create(&asset).Error)
	}
}

func setupBulkAssetLibrary(t *testing.T) (*gorm.DB, model.AssetGroup, model.AssetGroup, model.AssetGroup) {
	t.Helper()
	db := setupAssetLibraryTestDB(t)
	for _, user := range []model.User{
		{Id: 1, Username: "root-admin", AffCode: "root-admin", Status: common.UserStatusEnabled, Role: common.RoleRootUser},
		{Id: 42, Username: "owner-a", AffCode: "owner-a", Status: common.UserStatusEnabled, Role: common.RoleCommonUser},
		{Id: 43, Username: "owner-b", AffCode: "owner-b", Status: common.UserStatusEnabled, Role: common.RoleCommonUser},
	} {
		require.NoError(t, db.Create(&user).Error)
	}
	campaign := model.AssetGroup{PublicID: "group-campaign", OwnerUserID: 42, Name: "Campaign", Status: model.AssetStatusReady}
	archive := model.AssetGroup{PublicID: "group-archive", OwnerUserID: 42, Name: "Archive", Status: model.AssetStatusReady}
	foreign := model.AssetGroup{PublicID: "group-campaign", OwnerUserID: 43, Name: "Campaign", Status: model.AssetStatusReady}
	for _, group := range []*model.AssetGroup{&campaign, &archive, &foreign} {
		require.NoError(t, db.Create(group).Error)
	}
	seedBulkAssets(t, db, []bulkAssetSeed{
		{"asset-photo", 42, campaign.ID, model.AssetTypeImage, model.AssetStatusReady},
		{"asset-clip", 42, campaign.ID, model.AssetTypeVideo, model.AssetStatusReady},
		{"asset-voice", 42, campaign.ID, model.AssetTypeAudio, model.AssetStatusProcessing},
		{"asset-rejected", 42, campaign.ID, model.AssetTypeImage, model.AssetStatusUnavailable},
		{"asset-sync-failed", 42, campaign.ID, model.AssetTypeVideo, model.AssetStatusSyncFailed},
		{"asset-removed", 42, campaign.ID, model.AssetTypeVideo, model.AssetStatusDeleted},
		{"asset-kept", 42, archive.ID, model.AssetTypeImage, model.AssetStatusReady},
		{"asset-photo", 43, foreign.ID, model.AssetTypeImage, model.AssetStatusReady},
	})
	return db, campaign, archive, foreign
}

func listedAssetIDs(t *testing.T, input AssetListInput) []string {
	t.Helper()
	list, err := ListAssets(input)
	require.NoError(t, err)
	ids := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		ids = append(ids, item.ID)
	}
	return ids
}

func assetStatus(t *testing.T, db *gorm.DB, owner int, publicID string) string {
	t.Helper()
	var asset model.MediaAsset
	require.NoError(t, db.Where("owner_user_id = ? AND public_id = ?", owner, publicID).First(&asset).Error)
	return asset.Status
}

func TestAssetListSummaryAndOwners(t *testing.T) {
	db, _, archive, _ := setupBulkAssetLibrary(t)
	scope := AssetListInput{OwnerUserID: 42, GroupPublicID: "group-campaign"}

	videos := scope
	videos.AssetTypes = []string{model.AssetTypeVideo}
	assert.Equal(t, []string{"asset-sync-failed", "asset-clip"}, listedAssetIDs(t, videos))

	unavailable := scope
	unavailable.Statuses = []string{model.AssetStatusUnavailable, model.AssetStatusSyncFailed}
	assert.Equal(t, []string{"asset-sync-failed", "asset-rejected"}, listedAssetIDs(t, unavailable))

	// Oldest sort uses ID ascending as a stable tie-breaker for identical timestamps.
	for _, publicID := range []string{"same-ts-first", "same-ts-second"} {
		obj := model.StorageObject{ObjectID: "obj-" + publicID, OwnerUserID: 42, ObjectKey: "assets/" + publicID, Status: model.StorageObjectStatusUploaded, Purpose: model.StorageObjectPurposeAssetLibrary}
		require.NoError(t, db.Create(&obj).Error)
		require.NoError(t, db.Create(&model.MediaAsset{PublicID: publicID, OwnerUserID: 42, GroupID: archive.ID, StorageObjectID: obj.ID, Name: publicID, AssetType: model.AssetTypeImage, Status: model.AssetStatusReady, CreatedAt: 999}).Error)
	}
	assert.Equal(t, []string{"asset-kept", "same-ts-first", "same-ts-second"}, listedAssetIDs(t, AssetListInput{OwnerUserID: 42, GroupPublicID: archive.PublicID, SortOrder: "Asc"}))

	summary, err := SummarizeAssets(AssetListInput{
		OwnerUserID: 42, GroupPublicID: "group-campaign",
		AssetTypes: []string{model.AssetTypeAudio}, Statuses: []string{model.AssetStatusProcessing},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(5), summary.Total)
	assert.Equal(t, map[string]int64{model.AssetTypeImage: 2, model.AssetTypeVideo: 2, model.AssetTypeAudio: 1}, summary.ByType)
	assert.Equal(t, map[string]int64{model.AssetStatusReady: 2, model.AssetStatusProcessing: 1, model.AssetStatusUnavailable: 2}, summary.ByStatus)

	owners, err := ListAssetOwners("43", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), owners.Total)
	assert.Equal(t, []AssetOwnerView{{UserID: 43, Username: "owner-b", GroupCount: 1, AssetCount: 1}}, owners.Items)
}

func TestDeleteAssetsBatchAtomicityBoundsAndTenantScope(t *testing.T) {
	db, _, _, _ := setupBulkAssetLibrary(t)

	oversized := make([]AssetRef, MaxAssetBatchSize+1)
	for index := range oversized {
		oversized[index] = AssetRef{ID: fmt.Sprintf("asset-%d", index)}
	}
	for _, refs := range [][]AssetRef{nil, oversized} {
		_, err := DeleteAssets(refs, 42, false)
		var requestError *RequestError
		require.True(t, errors.As(err, &requestError))
		assert.Equal(t, http.StatusBadRequest, requestError.HTTPStatusCode())
	}

	_, err := DeleteAssets([]AssetRef{{ID: "asset-clip"}, {ID: "asset-missing"}}, 42, false)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Equal(t, model.AssetStatusReady, assetStatus(t, db, 42, "asset-clip"))

	deleted, err := DeleteAssets([]AssetRef{{ID: "asset-clip"}, {ID: "asset-photo", OwnerUserID: 43}}, 42, false)
	require.NoError(t, err)
	assert.Len(t, deleted, 2)
	assert.Equal(t, model.AssetStatusDeleted, assetStatus(t, db, 42, "asset-photo"))
	assert.Equal(t, model.AssetStatusReady, assetStatus(t, db, 43, "asset-photo"))

	deleted, err = DeleteAssets([]AssetRef{{ID: "asset-photo", OwnerUserID: 43}}, 1, true)
	require.NoError(t, err)
	assert.Len(t, deleted, 1)
	assert.Equal(t, model.AssetStatusDeleted, assetStatus(t, db, 43, "asset-photo"))
}

func TestDeleteGroupPermissionsAndCascade(t *testing.T) {
	db, campaign, _, _ := setupBulkAssetLibrary(t)
	adminA := model.User{Id: 61, Username: "admin-a", AffCode: "admin-a", Status: common.UserStatusEnabled, Role: common.RoleAdminUser}
	adminB := model.User{Id: 62, Username: "admin-b", AffCode: "admin-b", Status: common.UserStatusEnabled, Role: common.RoleAdminUser}
	require.NoError(t, db.Create(&[]model.User{adminA, adminB}).Error)

	// Plain delete refuses non-empty groups; cascade delete clears only the targeted group.
	var requestError *RequestError
	require.True(t, errors.As(DeleteGroup(campaign.PublicID, 42, false, 0), &requestError))
	assert.Equal(t, http.StatusConflict, requestError.HTTPStatusCode())
	result, err := DeleteGroupWithAssets(campaign.PublicID, 42, false, 0)
	require.NoError(t, err)
	assert.Equal(t, 5, result.DeletedAssetCount)
	assert.Equal(t, model.AssetStatusReady, assetStatus(t, db, 42, "asset-kept"))

	// Both empty-group and cascade deletes enforce cross-owner role hierarchy.
	rootGroup := model.AssetGroup{PublicID: "group-root-e", OwnerUserID: 1, Name: "Root", Status: model.AssetStatusReady}
	peerEmpty := model.AssetGroup{PublicID: "group-peer-e", OwnerUserID: adminB.Id, Name: "Peer", Status: model.AssetStatusReady}
	memberEmpty := model.AssetGroup{PublicID: "group-member-e", OwnerUserID: 42, Name: "Member", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&[]model.AssetGroup{rootGroup, peerEmpty, memberEmpty}).Error)
	seedBulkAssets(t, db, []bulkAssetSeed{{"root-asset-1", 1, rootGroup.ID, model.AssetTypeImage, model.AssetStatusReady}})

	_, err = DeleteGroupWithAssets(rootGroup.PublicID, adminA.Id, true, 1)
	require.ErrorIs(t, err, model.ErrAssetCopyPermissionDenied)
	require.ErrorIs(t, DeleteGroup(peerEmpty.PublicID, adminA.Id, true, adminB.Id), model.ErrAssetCopyPermissionDenied)
	require.NoError(t, DeleteGroup(memberEmpty.PublicID, adminA.Id, true, 42))
	require.NoError(t, DeleteGroup(peerEmpty.PublicID, 1, true, adminB.Id))
}

func TestDeleteAssetsByFilterAndStatusLockGuard(t *testing.T) {
	db, campaign, _, _ := setupBulkAssetLibrary(t)
	admin := model.User{Id: 60, Username: "admin-filter", AffCode: "admin-filter", Status: common.UserStatusEnabled, Role: common.RoleAdminUser}
	require.NoError(t, db.Create(&admin).Error)

	res, err := DeleteAssetsByFilter(AssetListInput{
		GroupPublicID: campaign.PublicID,
		Statuses:      []string{model.AssetStatusUnavailable, model.AssetStatusSyncFailed},
	}, []AssetRef{{ID: "asset-rejected", OwnerUserID: 42}}, 42, false)
	require.NoError(t, err)
	assert.Equal(t, 1, res.DeletedCount)
	assert.Equal(t, model.AssetStatusDeleted, assetStatus(t, db, 42, "asset-sync-failed"))
	assert.Equal(t, model.AssetStatusUnavailable, assetStatus(t, db, 42, "asset-rejected"))
	assert.Equal(t, model.AssetStatusReady, assetStatus(t, db, 43, "asset-photo"))

	// Status-filtered row lock skips an asset that transitioned to ready before locking.
	var ready model.MediaAsset
	require.NoError(t, db.Where("owner_user_id = ? AND public_id = ?", 42, "asset-photo").First(&ready).Error)
	deleted, err := model.RequestAssetDeletionIfStatusIn(db, ready.ID, []string{model.AssetStatusUnavailable, model.AssetStatusSyncFailed})
	require.NoError(t, err)
	assert.False(t, deleted)
	assert.Equal(t, model.AssetStatusReady, assetStatus(t, db, 42, "asset-photo"))

	// Regular admin cannot filter-delete a root user's assets.
	rootGroup := model.AssetGroup{PublicID: "group-root-f", OwnerUserID: 1, Name: "Root", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&rootGroup).Error)
	seedBulkAssets(t, db, []bulkAssetSeed{{"root-filter-asset", 1, rootGroup.ID, model.AssetTypeImage, model.AssetStatusReady}})
	_, err = DeleteAssetsByFilter(AssetListInput{IncludeAllOwners: true, TargetOwnerUserID: 1}, nil, admin.Id, true)
	require.ErrorIs(t, err, model.ErrAssetCopyPermissionDenied)
	assert.Equal(t, model.AssetStatusReady, assetStatus(t, db, 1, "root-filter-asset"))
}

func TestInlineThumbnailsAndBatchPreviewFallback(t *testing.T) {
	db, campaign, _, _ := setupBulkAssetLibrary(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "inline-preview-test-encryption-key")
	require.NoError(t, db.AutoMigrate(&model.StorageProfile{}, &model.StorageCredential{}, &model.StoragePolicy{}))
	profile, err := storageService.SaveProfile(0, storageService.ProfileInput{
		Name: "Inline OSS", ProviderType: model.StorageProviderAliyunOSS, Status: model.StorageProfileStatusEnabled,
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", Region: "cn-hangzhou", Bucket: "inline-assets-bucket",
		AuthType: model.StorageCredentialAuthStatic, AccessKeyID: "test-id", AccessKeySecret: "test-secret",
	})
	require.NoError(t, err)
	require.NoError(t, db.Model(&model.StorageObject{}).Where("1 = 1").Updates(map[string]any{"storage_profile_id": profile.ID, "content_type": "image/png"}).Error)
	var clip model.MediaAsset
	require.NoError(t, db.Where("owner_user_id = ? AND public_id = ?", 42, "asset-clip").First(&clip).Error)
	require.NoError(t, db.Model(&model.StorageObject{}).Where("id = ?", clip.StorageObjectID).Update("content_type", "video/mp4").Error)
	_, err = storageService.SaveAssetLibraryPolicy(storageService.PolicyInput{
		Enabled: true, StorageProfileID: profile.ID, ObjectPrefix: "asset-library",
		SignedURLTTLSeconds: 3600, MaxFileBytes: 1024, MaxFiles: 100, AllowedMIMETypes: "image/jpeg,image/png,video/mp4,audio/mpeg",
	})
	require.NoError(t, err)

	list, err := ListAssets(AssetListInput{Context: context.Background(), OwnerUserID: 42, GroupPublicID: campaign.PublicID, IncludePreviews: true})
	require.NoError(t, err)
	byID := make(map[string]AssetView, len(list.Items))
	for _, item := range list.Items {
		byID[item.ID] = item
	}
	assert.Contains(t, byID["asset-photo"].ThumbnailURL, "format%2Cwebp")
	assert.Contains(t, byID["asset-photo"].ThumbnailURL, "response-cache-control=")
	assert.Contains(t, byID["asset-clip"].PosterURL, "video%2Fsnapshot")

	previews, err := PreviewThumbnails(context.Background(), []AssetRef{{ID: "asset-photo", OwnerUserID: 42}, {ID: "asset-deleted-just-now", OwnerUserID: 42}}, 42, false)
	require.NoError(t, err)
	require.Len(t, previews, 1)
	assert.Contains(t, previews, AssetRefKey(42, "asset-photo"))
}
