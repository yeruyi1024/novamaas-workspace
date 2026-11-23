package assetlibrary

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

func setupAssetCopy(t *testing.T, statuses ...string) (*gorm.DB, AssetCopyInput, []model.MediaAsset) {
	t.Helper()
	db := setupAssetLibraryTestDB(t)
	previousRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() { common.RedisEnabled = previousRedisEnabled })
	for _, user := range []model.User{
		{Id: 1, Username: "customer-a", AffCode: "customer-a", Status: common.UserStatusEnabled},
		{Id: 2, Username: "customer-b", AffCode: "customer-b", Status: common.UserStatusEnabled},
		{Id: 100, Username: "copy-operator", AffCode: "copy-operator", Role: common.RoleRootUser, Status: common.UserStatusEnabled},
	} {
		require.NoError(t, db.Create(&user).Error)
	}
	require.NoError(t, db.Create(&model.Channel{Id: 91, Name: "shared channel", Status: common.ChannelStatusEnabled}).Error)
	library := model.AssetLibraryConnection{Name: "shared upstream", Enabled: true}
	require.NoError(t, db.Create(&library).Error)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: 91, LibraryID: library.ID, Enabled: true}).Error)
	group := model.AssetGroup{PublicID: "group-source", OwnerUserID: 1, Name: "Campaign", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	var assets []model.MediaAsset
	for index, status := range statuses {
		object := model.StorageObject{ObjectID: fmt.Sprintf("object-%d", index), OwnerUserID: 1, ObjectKey: fmt.Sprintf("assets/%d", index), Status: model.StorageObjectStatusUploaded, Purpose: model.StorageObjectPurposeAssetLibrary}
		require.NoError(t, db.Create(&object).Error)
		asset := model.MediaAsset{PublicID: fmt.Sprintf("asset-source-%d", index), OwnerUserID: 1, GroupID: group.ID, StorageObjectID: object.ID, Name: fmt.Sprintf("Photo %d", index), AssetType: "image", ContentType: "image/png", Size: 24, Status: status}
		if status == model.AssetStatusUnavailable {
			asset.UnavailableReason = model.AssetUnavailablePolicyRejected
		}
		require.NoError(t, db.Create(&asset).Error)
		replica := model.AssetReplica{AssetID: asset.ID, LibraryID: &library.ID, Operation: model.AssetReplicaOperationSync, UpstreamAssetID: fmt.Sprintf("asset-upstream-%d", index), Status: model.AssetReplicaStatusActive, Progress: 100}
		switch status {
		case model.AssetStatusSyncFailed:
			replica.Status, replica.Progress, replica.LastError = model.AssetReplicaStatusFailed, 0, "source timeout"
		case model.AssetStatusUnavailable:
			replica.Status = model.AssetReplicaStatusRejected
		case model.AssetStatusProcessing:
			replica.Status, replica.Progress = model.AssetReplicaStatusProcessing, 50
		}
		require.NoError(t, db.Create(&replica).Error)
		assets = append(assets, asset)
	}
	return db, AssetCopyInput{RequestID: "copy-request-test-001", SourceUserID: 1, TargetUserID: 2, GroupIDs: []string{group.PublicID}}, assets
}

func targetCopies(t *testing.T, db *gorm.DB) []model.MediaAsset {
	t.Helper()
	var copies []model.MediaAsset
	require.NoError(t, db.Where("owner_user_id = ?", 2).Order("id asc").Find(&copies).Error)
	return copies
}

func TestAssetCopyCreatesIndependentAssetsSynchronizedLikeUploads(t *testing.T) {
	db, input, originals := setupAssetCopy(t, model.AssetStatusReady, model.AssetStatusSyncFailed, model.AssetStatusProcessing, model.AssetStatusUnavailable)
	unselected := model.AssetGroup{PublicID: "group-not-selected", OwnerUserID: 1, Name: "Private", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&unselected).Error)
	require.NoError(t, db.Create(&model.MediaAsset{PublicID: "asset-not-selected", OwnerUserID: 1, GroupID: unselected.ID, StorageObjectID: originals[0].StorageObjectID, Status: model.AssetStatusReady}).Error)
	result, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	assert.Equal(t, 4, result.AssetCount)
	assert.Equal(t, 3, result.ProcessingCount)
	assert.Equal(t, 1, result.UnavailableCount)
	require.Len(t, result.Groups, 1)
	assert.Equal(t, input.GroupIDs[0], result.Groups[0].ID)
	var library model.AssetLibraryConnection
	require.NoError(t, db.First(&library).Error)
	copies := targetCopies(t, db)
	require.Len(t, copies, len(originals), "only the selected group is copied")
	for index, copy := range copies {
		source := originals[index]
		assert.Equal(t, source.PublicID, copy.PublicID)
		assert.Equal(t, source.StorageObjectID, copy.StorageObjectID)
		assert.Equal(t, source.Name, copy.Name)
		var replicas []model.AssetReplica
		require.NoError(t, db.Where("asset_id = ?", copy.ID).Find(&replicas).Error)
		if source.Status == model.AssetStatusUnavailable {
			assert.Equal(t, model.AssetStatusUnavailable, copy.Status)
			assert.Equal(t, source.UnavailableReason, copy.UnavailableReason)
			assert.Empty(t, replicas, "a rejected file is not resubmitted for review")
			continue
		}
		assert.Equal(t, model.AssetStatusProcessing, copy.Status)
		require.Len(t, replicas, 1)
		require.NotNil(t, replicas[0].LibraryID)
		assert.Equal(t, library.ID, *replicas[0].LibraryID)
		assert.Equal(t, model.AssetReplicaStatusPending, replicas[0].Status)
		assert.Empty(t, replicas[0].UpstreamAssetID, "copies never borrow the source's upstream ID")
	}
	for _, source := range originals {
		var stored model.MediaAsset
		require.NoError(t, db.First(&stored, source.ID).Error)
		assert.Equal(t, source.Status, stored.Status)
		var replicas int64
		require.NoError(t, db.Model(&model.AssetReplica{}).Where("asset_id = ?", source.ID).Count(&replicas).Error)
		assert.Equal(t, int64(1), replicas)
	}
}

func TestAssetCopyQueuesOneTaskPerEnabledAccount(t *testing.T) {
	db, input, _ := setupAssetCopy(t, model.AssetStatusReady)
	var shared model.AssetLibraryConnection
	require.NoError(t, db.First(&shared).Error)
	second := model.AssetLibraryConnection{Name: "second account", Enabled: true}
	require.NoError(t, db.Create(&second).Error)
	paused := model.AssetLibraryConnection{Name: "paused account", Enabled: true}
	require.NoError(t, db.Create(&paused).Error)
	require.NoError(t, db.Model(&paused).Update("enabled", false).Error)
	for channelID, libraryID := range map[int]int{92: shared.ID, 93: second.ID, 94: paused.ID} {
		require.NoError(t, db.Create(&model.Channel{Id: channelID, Name: fmt.Sprintf("channel %d", channelID), Status: common.ChannelStatusEnabled}).Error)
		require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: channelID, LibraryID: libraryID, Enabled: true}).Error)
	}
	_, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	copies := targetCopies(t, db)
	require.Len(t, copies, 1)
	var libraries []int
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("asset_id = ?", copies[0].ID).Order("library_id asc").Pluck("library_id", &libraries).Error)
	assert.Equal(t, []int{shared.ID, second.ID}, libraries, "channels sharing an account get one task; paused accounts get none")
}

func TestAssetCopyWithoutSynchronizationAllowsCleanupAndNotifiesOnlyTarget(t *testing.T) {
	db, input, _ := setupAssetCopy(t, model.AssetStatusReady)
	require.NoError(t, db.Model(&model.AssetChannelConfig{}).Where("channel_id = ?", 91).Update("enabled", false).Error)
	require.NoError(t, db.Create(&model.AssetWebhookEndpoint{PublicID: model.SystemAssetWebhookEndpointID, Status: model.AssetWebhookEndpointStatusEnabled}).Error)
	for _, owner := range []int{1, 2} {
		require.NoError(t, db.Create(&model.AssetWebhookEndpoint{PublicID: fmt.Sprintf("we_copy_%d", owner), OwnerUserID: owner,
			Name: "customer", URL: "https://customer.example.com/assets", EventTypes: `["asset.active","asset.failed"]`, Status: model.AssetWebhookEndpointStatusEnabled}).Error)
	}
	result, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	assert.Zero(t, result.ReadyCount)
	assert.True(t, result.CanDeleteSources)
	_, err = DeleteCopiedSources(result.ID, 100)
	require.NoError(t, err)
	copies := targetCopies(t, db)
	require.Len(t, copies, 1)
	assert.Equal(t, model.AssetStatusReady, copies[0].Status)
	var replicas int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("asset_id = ?", copies[0].ID).Count(&replicas).Error)
	assert.Zero(t, replicas)
	var deliveries []model.AssetWebhookDelivery
	require.NoError(t, db.Find(&deliveries).Error)
	require.NotEmpty(t, deliveries)
	for _, delivery := range deliveries {
		assert.Equal(t, 2, delivery.OwnerUserID)
		assert.Equal(t, copies[0].PublicID, gjson.Get(delivery.Payload, "data.id").String())
	}
}

func TestAssetCopyTargetRejectionDoesNotAffectSource(t *testing.T) {
	db, input, originals := setupAssetCopy(t, model.AssetStatusReady)
	_, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	copies := targetCopies(t, db)
	require.Len(t, copies, 1)
	var replica model.AssetReplica
	require.NoError(t, db.Where("asset_id = ?", copies[0].ID).First(&replica).Error)
	require.NoError(t, db.Model(&replica).Updates(map[string]any{"status": model.AssetReplicaStatusSyncing, "locked_by": "worker", "upstream_asset_id": "asset-target-review"}).Error)
	require.NoError(t, model.RejectAssetReplica(replica.ID, "worker", model.AssetUnavailablePolicyRejected, "asset-target-review"))
	var copy, source model.MediaAsset
	require.NoError(t, db.First(&copy, copies[0].ID).Error)
	assert.Equal(t, model.AssetStatusUnavailable, copy.Status)
	require.NoError(t, db.First(&source, originals[0].ID).Error)
	assert.Equal(t, model.AssetStatusReady, source.Status)
	mapped, _, err := ResolveRequestAssetIDs([]byte(fmt.Sprintf(`{"content":[{"image_url":{"url":"asset://%s"}}]}`, source.PublicID)), 1, 91)
	require.NoError(t, err)
	assert.Equal(t, "asset://asset-upstream-0", gjson.GetBytes(mapped, "content.0.image_url.url").String())
}

func TestAssetCopySynchronizesWithItsOwnUpstreamIDBeforeSourceCleanup(t *testing.T) {
	db, input, originals := setupAssetCopy(t, model.AssetStatusReady)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "copy-provider-fixture")
	require.NoError(t, db.AutoMigrate(&model.StorageProfile{}, &model.StorageCredential{}))
	profile, err := storageService.SaveProfile(0, storageService.ProfileInput{Name: "test files", ProviderType: model.StorageProviderAliyunOSS, Status: model.StorageProfileStatusEnabled,
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", Region: "cn-hangzhou", Bucket: "test-assets-bucket", AuthType: model.StorageCredentialAuthStatic, AccessKeyID: "fixture-id", AccessKeySecret: "fixture-secret"})
	require.NoError(t, err)
	require.NoError(t, db.Model(&model.StorageObject{}).Where("id = ?", originals[0].StorageObjectID).Update("storage_profile_id", profile.ID).Error)
	var submits, polls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer target-fixture-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("Action") {
		case "CreateAssetGroup":
			_, _ = io.WriteString(w, `{"Result":{"Id":"target-group"}}`)
		case "CreateAsset":
			submits.Add(1)
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, "target-group", gjson.GetBytes(body, "GroupId").String())
			originalURL, err := url.Parse(gjson.GetBytes(body, "URL").String())
			assert.NoError(t, err)
			assert.Equal(t, "/assets/0", originalURL.Path, "the copy is submitted from the shared OSS file")
			_, _ = io.WriteString(w, `{"Result":{"Id":"target-new-id"}}`)
		case "GetAsset":
			polls.Add(1)
			_, _ = io.WriteString(w, `{"Result":{"Status":"Active"}}`)
		default:
			t.Errorf("unexpected provider request: %s", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(upstream.Close)
	credential, err := encryptChannelCredential("target-fixture-token")
	require.NoError(t, err)
	target := model.AssetLibraryConnection{Name: "target", Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: upstream.URL, QPM: 1000000, EncryptedCredential: credential}
	require.NoError(t, db.Create(&target).Error)
	require.NoError(t, db.Model(&model.AssetChannelConfig{}).Where("channel_id = ?", 91).Update("library_id", target.ID).Error)
	result, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	assert.Equal(t, 1, result.ProcessingCount)
	assert.False(t, result.CanDeleteSources)
	_, err = DeleteCopiedSources(result.ID, 100)
	require.ErrorIs(t, err, model.ErrAssetCopyConflict, "sources stay until the batch finishes")
	require.NoError(t, db.Create(&model.AssetWebhookEndpoint{PublicID: model.SystemAssetWebhookEndpointID, Status: model.AssetWebhookEndpointStatusEnabled}).Error)
	for _, owner := range []int{1, 2} {
		require.NoError(t, db.Create(&model.AssetWebhookEndpoint{PublicID: fmt.Sprintf("we_copy_%d", owner), OwnerUserID: owner,
			Name: "customer", URL: "https://customer.example.com/assets", EventTypes: `["asset.active","asset.failed"]`, Status: model.AssetWebhookEndpointStatusEnabled}).Error)
	}
	copies := targetCopies(t, db)
	require.Len(t, copies, 1)
	for _, workload := range []model.AssetReplicaWorkload{model.AssetReplicaWorkloadSubmit, model.AssetReplicaWorkloadPoll} {
		var replica model.AssetReplica
		require.NoError(t, db.Where("asset_id = ? AND library_id = ?", copies[0].ID, target.ID).First(&replica).Error)
		now := replica.NextSyncAt
		candidates, err := model.ListDueAssetLibraryReplicas(now, target.ID, workload, 10)
		require.NoError(t, err)
		require.Len(t, candidates, 1)
		claimed, err := model.ClaimAssetReplica(candidates[0], now, now+120, "copy-worker")
		require.NoError(t, err)
		require.True(t, claimed)
		syncReplica(context.Background(), "copy-worker", candidates[0])
	}
	assert.EqualValues(t, 1, submits.Load())
	assert.EqualValues(t, 1, polls.Load())
	var copy model.MediaAsset
	require.NoError(t, db.First(&copy, copies[0].ID).Error)
	assert.Equal(t, model.AssetStatusReady, copy.Status)
	mapped, _, err := ResolveRequestAssetIDs([]byte(fmt.Sprintf(`{"content":[{"image_url":{"url":"asset://%s"}}]}`, copy.PublicID)), 2, 91)
	require.NoError(t, err)
	assert.Equal(t, "asset://target-new-id", gjson.GetBytes(mapped, "content.0.image_url.url").String())
	var deliveries []model.AssetWebhookDelivery
	require.NoError(t, db.Find(&deliveries).Error)
	require.NotEmpty(t, deliveries)
	for _, delivery := range deliveries {
		assert.Equal(t, 2, delivery.OwnerUserID, "the deleted source receives no event")
		assert.Equal(t, copy.PublicID, gjson.Get(delivery.Payload, "data.id").String())
	}
	receipt, err := GetAssetCopy(result.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, receipt.ReadyCount)
	require.True(t, receipt.CanDeleteSources)
	_, err = DeleteCopiedSources(result.ID, 100)
	require.NoError(t, err)
	mapped, _, err = ResolveRequestAssetIDs([]byte(fmt.Sprintf(`{"content":[{"image_url":{"url":"asset://%s"}}]}`, copy.PublicID)), 2, 91)
	require.NoError(t, err, "the copy keeps working after its source is deleted")
	assert.Equal(t, "asset://target-new-id", gjson.GetBytes(mapped, "content.0.image_url.url").String())
	var object model.StorageObject
	require.NoError(t, db.First(&object, originals[0].StorageObjectID).Error)
	assert.Equal(t, model.StorageObjectStatusUploaded, object.Status)
}

func TestAssetCopyCompletedCleanupKeepsMovedSourcesLaterUploadsAndSharedFiles(t *testing.T) {
	db, input, originals := setupAssetCopy(t, model.AssetStatusReady, model.AssetStatusReady)
	result, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	copies := targetCopies(t, db)
	markCopiesReady(t, db)
	moved := model.AssetGroup{PublicID: "group-moved", OwnerUserID: 1, Name: "Moved", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&moved).Error)
	require.NoError(t, db.Model(&originals[1]).Update("group_id", moved.ID).Error)
	later := model.MediaAsset{PublicID: "asset-later", OwnerUserID: 1, GroupID: originals[0].GroupID, StorageObjectID: 999, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&later).Error)
	for range 2 {
		_, err = DeleteCopiedSources(result.ID, 100)
		require.NoError(t, err)
	}
	_, err = model.FindMediaAsset(originals[0].PublicID, 1, false)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	for _, id := range []string{later.PublicID, originals[1].PublicID} {
		_, err = model.FindMediaAsset(id, 1, false)
		require.NoError(t, err)
	}
	_, err = model.FindAssetGroup(input.GroupIDs[0], 1, false)
	require.NoError(t, err)
	var replica model.AssetReplica
	require.NoError(t, db.Where("asset_id = ?", originals[0].ID).First(&replica).Error)
	assert.Equal(t, model.AssetReplicaOperationDelete, replica.Operation)
	var object model.StorageObject
	require.NoError(t, db.First(&object, originals[0].StorageObjectID).Error)
	assert.Equal(t, model.StorageObjectStatusUploaded, object.Status)
	mapped, _, err := ResolveRequestAssetIDs([]byte(fmt.Sprintf(`{"content":[{"image_url":{"url":"asset://%s"}}]}`, copies[0].PublicID)), 2, 91)
	require.NoError(t, err)
	assert.Equal(t, "asset://target-own-upstream-id", gjson.GetBytes(mapped, "content.0.image_url.url").String())
}

func markCopiesReady(t *testing.T, db *gorm.DB) {
	t.Helper()
	copies := targetCopies(t, db)
	for _, copy := range copies {
		require.NoError(t, db.Model(&copy).Update("status", model.AssetStatusReady).Error)
		require.NoError(t, db.Model(&model.AssetReplica{}).Where("asset_id = ?", copy.ID).Updates(map[string]any{"status": model.AssetReplicaStatusActive, "upstream_asset_id": "target-own-upstream-id"}).Error)
	}
}

func TestAssetCopyRechecksCurrentRolesBeforeWriting(t *testing.T) {
	db, input, _ := setupAssetCopy(t, model.AssetStatusReady)
	require.NoError(t, db.Model(&model.User{}).Where("id = ?", 100).Update("role", common.RoleAdminUser).Error)
	// The target was promoted after the API layer inspected its old role.
	require.NoError(t, db.Model(&model.User{}).Where("id = ?", input.TargetUserID).Update("role", common.RoleAdminUser).Error)
	_, err := CopyAssetGroups(100, input)
	require.ErrorIs(t, err, model.ErrAssetCopyPermissionDenied)
	var copies int64
	require.NoError(t, db.Model(&model.MediaAsset{}).Where("owner_user_id = ?", input.TargetUserID).Count(&copies).Error)
	assert.Zero(t, copies)
}

func TestAssetCopyCleanupRechecksParticipantsAndOperatorBeforeDeletion(t *testing.T) {
	for _, changedUser := range []string{"source", "target", "operator"} {
		t.Run(changedUser, func(t *testing.T) {
			db, input, originals := setupAssetCopy(t, model.AssetStatusReady)
			require.NoError(t, db.Model(&model.User{}).Where("id = ?", 100).Update("role", common.RoleAdminUser).Error)
			result, err := CopyAssetGroups(100, input)
			require.NoError(t, err)
			userID, role := input.SourceUserID, common.RoleAdminUser
			if changedUser == "target" {
				userID = input.TargetUserID
			} else if changedUser == "operator" {
				userID, role = 100, common.RoleCommonUser
			}
			require.NoError(t, db.Model(&model.User{}).Where("id = ?", userID).Update("role", role).Error)
			_, err = DeleteCopiedSources(result.ID, 100)
			require.ErrorIs(t, err, model.ErrAssetCopyPermissionDenied)
			_, err = model.FindMediaAsset(originals[0].PublicID, input.SourceUserID, false)
			require.NoError(t, err)
			var batch model.AssetCopyBatch
			require.NoError(t, db.First(&batch, "id = ?", result.ID).Error)
			assert.Zero(t, batch.SourcesDeletedAt)
		})
	}
}

func TestAssetCopyInvalidTargetRollsBackEveryWrite(t *testing.T) {
	for _, scenario := range []string{"disabled user", "database failure"} {
		t.Run(scenario, func(t *testing.T) {
			db, input, _ := setupAssetCopy(t, model.AssetStatusReady)
			if scenario == "disabled user" {
				require.NoError(t, db.Model(&model.User{}).Where("id = ?", 2).Update("status", common.UserStatusDisabled).Error)
			} else {
				require.NoError(t, db.Callback().Create().Before("gorm:create").Register("fail_copy_batch", func(tx *gorm.DB) {
					if _, ok := tx.Statement.Dest.(*model.AssetCopyBatch); ok {
						tx.AddError(errors.New("database unavailable"))
					}
				}))
				t.Cleanup(func() { db.Callback().Create().Remove("fail_copy_batch") })
			}
			_, err := CopyAssetGroups(100, input)
			require.Error(t, err)
			for _, fixture := range []struct {
				entity any
				count  int64
			}{{&model.MediaAsset{}, 1}, {&model.AssetGroup{}, 1}, {&model.AssetReplica{}, 1}, {&model.AssetCopyBatch{}, 0}} {
				var count int64
				require.NoError(t, db.Model(fixture.entity).Count(&count).Error)
				assert.Equal(t, fixture.count, count, "%T cannot be partially committed", fixture.entity)
			}
		})
	}
}

func TestAssetCopyConcurrentReplayCreatesOneBatchAndOneTargetTask(t *testing.T) {
	db, input, _ := setupAssetCopy(t, model.AssetStatusReady)
	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() { <-start; _, err := CopyAssetGroups(100, input); results <- err }()
	}
	close(start)
	require.NoError(t, <-results)
	require.NoError(t, <-results)
	for _, fixture := range []struct {
		entity any
		count  int64
	}{{&model.MediaAsset{}, 2}, {&model.AssetGroup{}, 2}, {&model.AssetReplica{}, 2}, {&model.AssetCopyBatch{}, 1}} {
		var count int64
		require.NoError(t, db.Model(fixture.entity).Count(&count).Error)
		assert.Equal(t, fixture.count, count)
	}
}

func TestAssetCopyRacingSourceDeletionNeverKeepsACopyOfAReclaimedFile(t *testing.T) {
	db, input, originals := setupAssetCopy(t, model.AssetStatusReady)
	// Without an upstream copy, deleting the source finishes and reclaims at once.
	require.NoError(t, db.Where("asset_id = ?", originals[0].ID).Delete(&model.AssetReplica{}).Error)
	start := make(chan struct{})
	copyResult, deletionResult := make(chan error, 1), make(chan error, 1)
	go func() { <-start; _, err := CopyAssetGroups(100, input); copyResult <- err }()
	go func() { <-start; deletionResult <- DeleteAsset(originals[0].PublicID, 1, false, 0) }()
	close(start)
	require.NoError(t, <-deletionResult)
	require.NoError(t, <-copyResult)
	var object model.StorageObject
	require.NoError(t, db.First(&object, originals[0].StorageObjectID).Error)
	if copies := targetCopies(t, db); len(copies) > 0 {
		assert.Equal(t, model.StorageObjectStatusUploaded, object.Status)
		return
	}
	assert.Equal(t, model.StorageObjectStatusDeletePending, object.Status)
}

func TestAssetCopyCleanupKeepsUploadsCommittedAfterItsInitialSnapshot(t *testing.T) {
	db, input, assets := setupAssetCopy(t, model.AssetStatusReady)
	if db.Dialector.Name() != "mysql" {
		t.Skip("requires MySQL repeatable-read snapshots and a second connection")
	}
	result, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	markCopiesReady(t, db)
	newUpload := model.MediaAsset{PublicID: "asset-concurrent-upload", OwnerUserID: 1, GroupID: assets[0].GroupID, StorageObjectID: 999, Status: model.AssetStatusReady}
	inserted := false
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("upload_after_cleanup_snapshot", func(tx *gorm.DB) {
		if _, isBatch := tx.Statement.Dest.(*model.AssetCopyBatch); !isBatch || inserted {
			return
		}
		inserted = true
		// A second connection commits after cleanup's first consistent read,
		// before it obtains the group lock. No sleeps or timing assumptions.
		tx.AddError(db.Create(&newUpload).Error)
	}))
	t.Cleanup(func() { db.Callback().Query().Remove("upload_after_cleanup_snapshot") })
	_, err = DeleteCopiedSources(result.ID, 100)
	require.NoError(t, err)
	require.True(t, inserted)
	_, err = model.FindMediaAsset(newUpload.PublicID, 1, false)
	require.NoError(t, err)
	_, err = model.FindAssetGroup(input.GroupIDs[0], 1, false)
	require.NoError(t, err, "a source group containing a concurrent later upload must remain visible")
}

func TestAssetCopyReplayReturnsCommittedReceiptAfterTargetChanges(t *testing.T) {
	db, input, _ := setupAssetCopy(t, model.AssetStatusReady)
	first, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	require.NoError(t, db.Model(&model.User{}).Where("id = ?", input.TargetUserID).Update("status", common.UserStatusDisabled).Error)
	replayed, err := CopyAssetGroups(100, input)
	require.NoError(t, err, "a lost response is answered with the committed batch")
	assert.Equal(t, first.ID, replayed.ID)
	assert.Equal(t, first.AssetCount, replayed.AssetCount)
	var copies int64
	require.NoError(t, db.Model(&model.MediaAsset{}).Where("owner_user_id = ?", input.TargetUserID).Count(&copies).Error)
	assert.Equal(t, int64(1), copies)
	changed := input
	changed.TargetUserID = 100
	_, err = CopyAssetGroups(100, changed)
	var requestErr *RequestError
	require.ErrorAs(t, err, &requestErr)
	assert.Equal(t, http.StatusConflict, requestErr.StatusCode)
}

func TestAssetCopyDoesNotLockRowsUpdatedByRelayBilling(t *testing.T) {
	db, input, _ := setupAssetCopy(t, model.AssetStatusReady)
	if db.Dialector.Name() != "mysql" {
		t.Skip("requires MySQL row locks and a second connection")
	}
	checked := false
	require.NoError(t, db.Callback().Create().After("gorm:create").Register("quota_update_during_copy", func(tx *gorm.DB) {
		switch tx.Statement.Dest.(type) {
		case []model.MediaAsset, *[]model.MediaAsset:
		default:
			return
		}
		if checked {
			return
		}
		checked = true
		// Relay billing updates user quota and channel usage directly. A second
		// connection must not wait on the copy transaction for these rows.
		tx.AddError(db.Transaction(func(other *gorm.DB) error {
			if err := other.Exec("SET SESSION innodb_lock_wait_timeout = 1").Error; err != nil {
				return err
			}
			if err := other.Model(&model.User{}).Where("id IN ?", []int{input.SourceUserID, input.TargetUserID, 100}).
				Update("quota", gorm.Expr("quota + ?", 1)).Error; err != nil {
				return err
			}
			return other.Model(&model.Channel{}).Where("id = ?", 91).Update("used_quota", gorm.Expr("used_quota + ?", 1)).Error
		}))
	}))
	t.Cleanup(func() { require.NoError(t, db.Callback().Create().Remove("quota_update_during_copy")) })
	_, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	require.True(t, checked)
}

func TestAssetCopyWaitsForEveryAccountToFinishProcessing(t *testing.T) {
	db, input, sources := setupAssetCopy(t, model.AssetStatusReady)
	second := model.AssetLibraryConnection{Name: "second account", Enabled: true}
	require.NoError(t, db.Create(&second).Error)
	require.NoError(t, db.Create(&model.Channel{Id: 92, Status: common.ChannelStatusEnabled}).Error)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: 92, LibraryID: second.ID, Enabled: true}).Error)
	result, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	copies := targetCopies(t, db)
	require.Len(t, copies, 1)
	var replicas []model.AssetReplica
	require.NoError(t, db.Where("asset_id = ?", copies[0].ID).Order("id").Find(&replicas).Error)
	require.Len(t, replicas, 2)
	require.NoError(t, db.Model(&copies[0]).Update("status", model.AssetStatusReady).Error)
	require.NoError(t, db.Model(&replicas[0]).Updates(map[string]any{"status": model.AssetReplicaStatusActive, "upstream_asset_id": "copy-first-account"}).Error)
	receipt, err := GetAssetCopy(result.ID)
	require.NoError(t, err)
	assert.Zero(t, receipt.ReadyCount, "one account is not a completed copy")
	_, err = DeleteCopiedSources(result.ID, 100)
	require.ErrorIs(t, err, model.ErrAssetCopyConflict)
	require.NoError(t, db.First(&sources[0], sources[0].ID).Error)
	assert.Equal(t, model.AssetStatusReady, sources[0].Status)
	for _, status := range []string{model.AssetReplicaStatusFailed, model.AssetReplicaStatusRejected, model.AssetReplicaStatusProcessing} {
		require.NoError(t, db.Model(&replicas[1]).Update("status", status).Error)
		receipt, err = GetAssetCopy(result.ID)
		require.NoError(t, err)
		if status == model.AssetReplicaStatusRejected {
			assert.True(t, receipt.CanDeleteSources)
			assert.Equal(t, 1, receipt.FailedCount)
		} else {
			assert.False(t, receipt.CanDeleteSources)
			_, err = DeleteCopiedSources(result.ID, 100)
			require.ErrorIs(t, err, model.ErrAssetCopyConflict)
		}
	}
	require.NoError(t, db.Delete(&replicas[1]).Error)
	receipt, err = GetAssetCopy(result.ID)
	require.NoError(t, err)
	assert.Zero(t, receipt.ReadyCount, "a removed required account cannot become completed")
	assert.True(t, receipt.CanDeleteSources)
	require.NoError(t, db.Create(&replicas[1]).Error)
	require.NoError(t, db.Model(&replicas[1]).Update("status", model.AssetReplicaStatusActive).Error)
	receipt, err = GetAssetCopy(result.ID)
	require.NoError(t, err)
	assert.Zero(t, receipt.ReadyCount, "active without an upstream ID is not successful")
	assert.True(t, receipt.CanDeleteSources)
	require.NoError(t, db.Model(&replicas[1]).Updates(map[string]any{"status": model.AssetReplicaStatusActive, "upstream_asset_id": "copy-second-account"}).Error)
	receipt, err = GetAssetCopy(result.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, receipt.ReadyCount)
	_, err = DeleteCopiedSources(result.ID, 100)
	require.NoError(t, err)
}

func TestAssetCopyFinishedBatchCleanupIncludesUnavailableSources(t *testing.T) {
	for _, status := range []string{model.AssetStatusDeleted, model.AssetStatusUnavailable, "missing"} {
		t.Run(status, func(t *testing.T) {
			db, input, originals := setupAssetCopy(t, model.AssetStatusReady, model.AssetStatusReady)
			result, err := CopyAssetGroups(100, input)
			require.NoError(t, err)
			markCopiesReady(t, db)
			if status == "missing" {
				require.NoError(t, db.Where("owner_user_id = ?", 2).Delete(&model.MediaAsset{}).Error)
			} else {
				require.NoError(t, db.Model(&model.MediaAsset{}).Where("owner_user_id = ?", 2).Update("status", status).Error)
			}
			receipt, err := GetAssetCopy(result.ID)
			require.NoError(t, err)
			assert.Zero(t, receipt.ReadyCount)
			assert.True(t, receipt.CanDeleteSources)
			_, err = DeleteCopiedSources(result.ID, 100)
			require.NoError(t, err)
			_, err = model.FindMediaAsset(originals[0].PublicID, 1, false)
			require.ErrorIs(t, err, gorm.ErrRecordNotFound)
			_, err = model.FindMediaAsset(originals[1].PublicID, 1, false)
			require.ErrorIs(t, err, gorm.ErrRecordNotFound)
		})
	}
}

func TestAssetCopyPartialCompletionDeletesTheSourceBatch(t *testing.T) {
	db, input, originals := setupAssetCopy(t, model.AssetStatusReady, model.AssetStatusReady)
	result, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	markCopiesReady(t, db)
	copies := targetCopies(t, db)
	for _, status := range []string{model.AssetReplicaStatusPending, model.AssetReplicaStatusFailed} {
		require.NoError(t, db.Model(&model.AssetReplica{}).Where("asset_id = ?", copies[1].ID).Updates(map[string]any{"status": status, "next_sync_at": common.GetTimestamp() + 60}).Error)
		receipt, err := GetAssetCopy(result.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, receipt.ProcessingCount)
		assert.False(t, receipt.CanDeleteSources)
		_, err = DeleteCopiedSources(result.ID, 100)
		require.ErrorIs(t, err, model.ErrAssetCopyConflict)
	}
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("asset_id = ?", copies[1].ID).Updates(map[string]any{"status": model.AssetReplicaStatusFailed, "next_sync_at": 0, "upstream_asset_id": ""}).Error)
	receipt, err := GetAssetCopy(result.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, receipt.ReadyCount)
	assert.Equal(t, 1, receipt.FailedCount)
	assert.Zero(t, receipt.ProcessingCount)
	assert.True(t, receipt.CanDeleteSources)
	_, err = DeleteCopiedSources(result.ID, 100)
	require.NoError(t, err)
	_, err = model.FindMediaAsset(originals[0].PublicID, 1, false)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = model.FindMediaAsset(originals[1].PublicID, 1, false)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = model.FindAssetGroup(input.GroupIDs[0], 1, false)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	for _, copy := range copies {
		_, err = model.FindMediaAsset(copy.PublicID, 2, false)
		require.NoError(t, err, "failed destination copies remain available for retry")
		var object model.StorageObject
		require.NoError(t, db.First(&object, copy.StorageObjectID).Error)
		assert.Equal(t, model.StorageObjectStatusUploaded, object.Status)
	}
}

func TestAssetCopyPreservesPublicIDsAcrossTenantsAndRejectsActiveDuplicates(t *testing.T) {
	db, input, originals := setupAssetCopy(t, model.AssetStatusReady)
	require.NoError(t, db.Create(&model.User{Id: 3, Username: "customer-c", AffCode: "customer-c", Status: common.UserStatusEnabled}).Error)

	result, err := CopyAssetGroups(100, input)
	require.NoError(t, err)
	require.Len(t, result.Groups, 1)
	assert.Equal(t, input.GroupIDs[0], result.Groups[0].ID)

	copies := targetCopies(t, db)
	require.Len(t, copies, 1)
	assert.Equal(t, originals[0].PublicID, copies[0].PublicID)

	// Activate target copy with its own upstream ID.
	require.NoError(t, db.Model(&copies[0]).Update("status", model.AssetStatusReady).Error)
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("asset_id = ?", copies[0].ID).
		Updates(map[string]any{"status": model.AssetReplicaStatusActive, "upstream_asset_id": "target-preserved-upstream"}).Error)

	// Both source and target resolve the exact same asset:// reference to their own upstream copies.
	requestBody := []byte(fmt.Sprintf(`{"content":[{"image_url":{"url":"asset://%s"}}]}`, originals[0].PublicID))
	sourceMapped, _, err := ResolveRequestAssetIDs(requestBody, 1, 91)
	require.NoError(t, err)
	assert.Equal(t, "asset://asset-upstream-0", gjson.GetBytes(sourceMapped, "content.0.image_url.url").String())

	targetMapped, _, err := ResolveRequestAssetIDs(requestBody, 2, 91)
	require.NoError(t, err)
	assert.Equal(t, "asset://target-preserved-upstream", gjson.GetBytes(targetMapped, "content.0.image_url.url").String())

	// An unrelated customer still gets 404 and never passes through the managed asset ID.
	_, _, err = ResolveRequestAssetIDs(requestBody, 3, 91)
	var reqErr *RequestError
	require.ErrorAs(t, err, &reqErr)
	assert.Equal(t, http.StatusNotFound, reqErr.StatusCode)

	// Copying the same group again with a new request_id is rejected while active copies exist.
	duplicateInput := input
	duplicateInput.RequestID = "copy-request-test-duplicate"
	_, err = CopyAssetGroups(100, duplicateInput)
	require.ErrorIs(t, err, model.ErrAssetCopyConflict)

	// Admin operations without owner_user_id refuse ambiguous shared IDs; with owner_user_id they succeed.
	err = DeleteAsset(originals[0].PublicID, 100, true, 0)
	require.ErrorIs(t, err, model.ErrAssetAmbiguousPublicID)
	err = DeleteGroup(input.GroupIDs[0], 100, true, 0)
	require.ErrorIs(t, err, model.ErrAssetAmbiguousPublicID)

	// Deleting the target's copy and group frees the active slot so the same group/asset IDs can be copied again.
	require.NoError(t, DeleteAsset(copies[0].PublicID, 100, true, 2))
	require.NoError(t, DeleteGroup(input.GroupIDs[0], 100, true, 2))

	// Source remains intact and resolvable; target gets 404 until re-copied.
	_, _, err = ResolveRequestAssetIDs(requestBody, 2, 91)
	require.ErrorAs(t, err, &reqErr)
	assert.Equal(t, http.StatusNotFound, reqErr.StatusCode)

	recopyInput := input
	recopyInput.RequestID = "copy-request-test-recopy"
	recopied, err := CopyAssetGroups(100, recopyInput)
	require.NoError(t, err)
	require.Len(t, recopied.Groups, 1)
	assert.Equal(t, input.GroupIDs[0], recopied.Groups[0].ID)
}
