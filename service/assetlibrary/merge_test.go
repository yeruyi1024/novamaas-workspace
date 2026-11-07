package assetlibrary

import (
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

// seedLegacyChannel stores a per-channel configuration the way the version
// before shared accounts did.
func seedLegacyChannel(t *testing.T, db *gorm.DB, channelID int, endpoint, accessKey, secret string, enabled bool) {
	t.Helper()
	encrypted, err := encryptChannelCredential(secret)
	require.NoError(t, err)
	require.NoError(t, db.Create(&model.Channel{Id: channelID, Name: fmt.Sprintf("channel %d", channelID), Type: constant.ChannelTypeVolcNative}).Error)
	require.NoError(t, db.Create(&model.AssetChannelConfig{
		ChannelID: channelID, Enabled: enabled, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthAKSK,
		BaseURL: endpoint, Region: "cn-beijing", Service: "ark", APIVersion: "2024-01-01", QPM: 60,
		AccessKeyID: accessKey, EncryptedCredential: encrypted, CredentialKeyVersion: channelCredentialVersion,
	}).Error)
}

func seedLegacyCopy(t *testing.T, db *gorm.DB, assetID int64, channelID int, upstreamID, status string) {
	t.Helper()
	require.NoError(t, db.Create(&model.AssetReplica{AssetID: assetID, ChannelID: channelID, UpstreamAssetID: upstreamID, Status: status, Operation: model.AssetReplicaOperationSync}).Error)
}

func libraryOfChannel(t *testing.T, db *gorm.DB, channelID int) int {
	t.Helper()
	var binding model.AssetChannelConfig
	require.NoError(t, db.Where("channel_id = ?", channelID).First(&binding).Error)
	return binding.LibraryID
}

func TestMergeIdenticalLibrariesFoldsPerChannelConfigsWithoutUploadingAgain(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "merge-fixture")
	// The shape of the 2026-10-16 production backup: eight per-channel
	// configurations that are really four upstream accounts.
	for _, channel := range []struct {
		id                          int
		endpoint, accessKey, secret string
		enabled                     bool
	}{
		{6, "https://flaios.example/api/v3/", "AK-FLAIOS", "secret-flaios", false},
		{29, "https://flaios.example/api/v3", "AK-FLAIOS", "secret-flaios", true},
		{31, "https://flaios.example/api/v3/", "AK-FLAIOS", "secret-flaios", true},
		{13, "https://yoofang.example", "", "secret-yoofang", false},
		{32, "https://yoofang.example", "", "secret-yoofang", false},
		{38, "https://yoofang.example", "", "secret-yoofang", false},
		{27, "https://matrix.example", "AK-MATRIX", "secret-matrix", true},
		{33, "https://mintel.example", "AK-MINTEL", "secret-mintel", true},
	} {
		seedLegacyChannel(t, db, channel.id, channel.endpoint, channel.accessKey, channel.secret, channel.enabled)
	}
	var assets []model.MediaAsset
	for index, name := range []string{"asset-one", "asset-two", "asset-three"} {
		asset := model.MediaAsset{PublicID: name, OwnerUserID: 7, StorageObjectID: int64(100 + index), Status: model.AssetStatusReady}
		require.NoError(t, db.Create(&asset).Error)
		assets = append(assets, asset)
	}
	for _, asset := range assets {
		// Both enabled Flaios channels uploaded their own copy; the others did not.
		seedLegacyCopy(t, db, asset.ID, 29, "flaios-29-"+asset.PublicID, model.AssetReplicaStatusActive)
		seedLegacyCopy(t, db, asset.ID, 31, "flaios-31-"+asset.PublicID, model.AssetReplicaStatusActive)
		seedLegacyCopy(t, db, asset.ID, 32, "", model.AssetReplicaStatusFailed)
		seedLegacyCopy(t, db, asset.ID, 38, "", model.AssetReplicaStatusFailed)
		seedLegacyCopy(t, db, asset.ID, 27, "matrix-"+asset.PublicID, model.AssetReplicaStatusActive)
		seedLegacyCopy(t, db, asset.ID, 33, "mintel-"+asset.PublicID, model.AssetReplicaStatusActive)
	}
	var copiesBefore int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Count(&copiesBefore).Error)

	require.NoError(t, model.MigrateAssetLibraryConnections(db))
	require.NoError(t, RecordLibraryIdentities())
	var perChannel int64
	require.NoError(t, db.Model(&model.AssetLibraryConnection{}).Count(&perChannel).Error)
	require.EqualValues(t, 8, perChannel, "starting up only records identities and merges nothing")
	assert.NotEqual(t, libraryOfChannel(t, db, 29), libraryOfChannel(t, db, 31))
	var recorded int64
	require.NoError(t, db.Model(&model.AssetLibraryConnection{}).Where("resource_key IS NOT NULL").Count(&recorded).Error)
	assert.EqualValues(t, 4, recorded, "one account per upstream identity records it")

	report, err := MergeIdenticalLibraries()
	require.NoError(t, err)
	assert.Equal(t, 2, report.MergedGroups, "Flaios and YooFang were split across several accounts")
	assert.Zero(t, report.DeferredGroups)

	flaios := libraryOfChannel(t, db, 29)
	assert.Equal(t, flaios, libraryOfChannel(t, db, 31))
	assert.Equal(t, flaios, libraryOfChannel(t, db, 6))
	yoofang := libraryOfChannel(t, db, 13)
	assert.Equal(t, yoofang, libraryOfChannel(t, db, 32))
	assert.Equal(t, yoofang, libraryOfChannel(t, db, 38))
	assert.NotEqual(t, libraryOfChannel(t, db, 27), libraryOfChannel(t, db, 33))
	linked := map[int]bool{}
	for _, channelID := range []int{6, 13, 27, 29, 31, 32, 33, 38} {
		linked[libraryOfChannel(t, db, channelID)] = true
	}
	assert.Len(t, linked, 4)

	// The kept Flaios account is the one holding copies, not the lowest ID
	// (channel 6 had none), and it is enabled because two of its channels were.
	var kept model.AssetLibraryConnection
	require.NoError(t, db.First(&kept, flaios).Error)
	assert.True(t, kept.Enabled)
	assert.Equal(t, "channel 6", kept.Name, "titled after the channel with the smallest ID")
	assert.NotNil(t, kept.ResourceKey)
	var keyed int64
	require.NoError(t, db.Model(&model.AssetLibraryConnection{}).Where("resource_key IS NOT NULL").Count(&keyed).Error)
	assert.EqualValues(t, 4, keyed, "exactly one account per upstream identity carries it")

	// Nothing was uploaded or deleted: every copy is still there, and each asset
	// has one active copy under the kept account.
	var copiesAfter int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("status <> ?", model.AssetReplicaStatusDeleted).Count(&copiesAfter).Error)
	assert.Equal(t, copiesBefore, copiesAfter)
	for _, asset := range assets {
		var active int64
		require.NoError(t, db.Model(&model.AssetReplica{}).Where("asset_id = ? AND library_id = ? AND status = ?", asset.ID, flaios, model.AssetReplicaStatusActive).Count(&active).Error)
		assert.EqualValues(t, 1, active)
	}
	var retired int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("library_id <> ? AND status = ? AND upstream_asset_id LIKE ?", flaios, model.AssetReplicaStatusActive, "flaios-%").Count(&retired).Error)
	assert.EqualValues(t, 3, retired, "the surplus Flaios copies stay with their old account")

	// Both enabled channels reach the same upstream copy.
	body := []byte(`{"content":[{"image_url":{"url":"asset://asset-one"}}]}`)
	first, _, err := ResolveRequestAssetIDs(body, 7, 29)
	require.NoError(t, err)
	second, _, err := ResolveRequestAssetIDs(body, 7, 31)
	require.NoError(t, err)
	assert.Equal(t, gjson.GetBytes(first, "content.0.image_url.url").String(), gjson.GetBytes(second, "content.0.image_url.url").String())

	// Channel 6's account held nothing and is gone; the accounts that still hold
	// surplus copies (channel 31, 32 and 38) remain as retired accounts.
	var accounts int64
	require.NoError(t, db.Model(&model.AssetLibraryConnection{}).Count(&accounts).Error)
	assert.EqualValues(t, 7, accounts)

	// Edits made after the merge survive later startups and a second merge.
	require.NoError(t, db.Model(&model.AssetLibraryConnection{}).Where("id = ?", flaios).Updates(map[string]any{"name": "Flaios production", "enabled": false, "qpm": 30}).Error)
	require.NoError(t, RecordLibraryIdentities())
	again, err := MergeIdenticalLibraries()
	require.NoError(t, err)
	assert.Zero(t, again.MergedGroups)
	require.NoError(t, db.First(&kept, flaios).Error)
	assert.Equal(t, "Flaios production", kept.Name)
	assert.False(t, kept.Enabled)
	assert.Equal(t, 30, kept.QPM)
	require.NoError(t, db.Model(&model.AssetLibraryConnection{}).Count(&accounts).Error)
	assert.EqualValues(t, 7, accounts)
	assert.Equal(t, flaios, libraryOfChannel(t, db, 31))
}

func TestMergeIdenticalLibrariesMovesOnlyCopiesTheKeptAccountLacks(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "merge-move-fixture")
	seedLegacyChannel(t, db, 1, "https://provider.example", "AK", "same-secret", true)
	seedLegacyChannel(t, db, 2, "https://provider.example", "AK", "same-secret", true)
	require.NoError(t, model.MigrateAssetLibraryConnections(db))
	first, second := libraryOfChannel(t, db, 1), libraryOfChannel(t, db, 2)
	require.Less(t, first, second)
	group := model.AssetGroup{PublicID: "group-a", OwnerUserID: 7, Status: model.AssetStatusReady}
	otherGroup := model.AssetGroup{PublicID: "group-b", OwnerUserID: 7, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	require.NoError(t, db.Create(&otherGroup).Error)
	ids := map[string]int64{}
	for index, name := range []string{"one", "two", "three"} {
		asset := model.MediaAsset{PublicID: "asset-" + name, OwnerUserID: 7, GroupID: group.ID, StorageObjectID: int64(200 + index), Status: model.AssetStatusReady}
		require.NoError(t, db.Create(&asset).Error)
		ids[name] = asset.ID
	}
	copyFor := func(name string, library int, upstream string) {
		id := library
		require.NoError(t, db.Create(&model.AssetReplica{AssetID: ids[name], LibraryID: &id, UpstreamAssetID: upstream, Status: model.AssetReplicaStatusActive, Operation: model.AssetReplicaOperationSync}).Error)
	}
	// Equal counts, so the lower account is kept. Asset "one" is duplicated;
	// "two" exists only on the second account and "three" only on the first.
	copyFor("one", first, "first-one")
	copyFor("three", first, "first-three")
	copyFor("one", second, "second-one")
	copyFor("two", second, "second-two")
	for _, library := range []int{first, second} {
		id := library
		groupID := group.ID
		if library == second {
			groupID = otherGroup.ID
		}
		require.NoError(t, db.Create(&model.AssetGroupReplica{GroupID: groupID, LibraryID: &id, UpstreamGroupID: "upstream-group", Status: model.AssetReplicaStatusActive}).Error)
	}

	laterAsset := model.MediaAsset{PublicID: "asset-four", OwnerUserID: 7, GroupID: group.ID, StorageObjectID: 299, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&laterAsset).Error)

	_, err := MergeIdenticalLibraries()
	require.NoError(t, err)

	assert.Equal(t, first, libraryOfChannel(t, db, 2))
	upstream := map[int64]string{}
	var kept []model.AssetReplica
	require.NoError(t, db.Where("library_id = ?", first).Find(&kept).Error)
	for _, replica := range kept {
		upstream[replica.AssetID] = replica.UpstreamAssetID
	}
	assert.Equal(t, map[int64]string{ids["one"]: "first-one", ids["two"]: "second-two", ids["three"]: "first-three", laterAsset.ID: ""}, upstream)
	var queued model.AssetReplica
	require.NoError(t, db.Where("library_id = ? AND asset_id = ?", first, laterAsset.ID).First(&queued).Error)
	assert.Equal(t, model.AssetReplicaStatusPending, queued.Status, "an asset no account holds is queued for the merged account")
	var duplicate model.AssetReplica
	require.NoError(t, db.Where("library_id = ? AND asset_id = ?", second, ids["one"]).First(&duplicate).Error)
	assert.Equal(t, "second-one", duplicate.UpstreamAssetID, "a surplus copy is retained, not deleted")
	var groups int64
	require.NoError(t, db.Model(&model.AssetGroupReplica{}).Where("library_id = ?", first).Count(&groups).Error)
	assert.EqualValues(t, 2, groups, "the kept account takes the platform group it lacked")
}

func TestMergeIdenticalLibrariesWaitsForRequestsInFlightAndSkipsUnreadableCredentials(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "merge-wait-fixture")
	seedLegacyChannel(t, db, 1, "https://provider.example", "AK", "same-secret", true)
	seedLegacyChannel(t, db, 2, "https://provider.example", "AK", "same-secret", true)
	seedLegacyChannel(t, db, 3, "https://other.example", "AK", "unreadable", true)
	seedLegacyChannel(t, db, 4, "https://other.example", "AK", "unreadable", true)
	require.NoError(t, db.Model(&model.AssetChannelConfig{}).Where("channel_id IN ?", []int{3, 4}).Update("encrypted_credential", "unreadable-ciphertext").Error)
	require.NoError(t, model.MigrateAssetLibraryConnections(db))
	second := libraryOfChannel(t, db, 2)
	asset := model.MediaAsset{PublicID: "asset-busy", OwnerUserID: 7, StorageObjectID: 1, Status: model.AssetStatusProcessing}
	require.NoError(t, db.Create(&asset).Error)
	require.NoError(t, db.Create(&model.AssetReplica{AssetID: asset.ID, LibraryID: &second, Status: model.AssetReplicaStatusSyncing, Operation: model.AssetReplicaOperationSync,
		LockedBy: "worker", LeaseUntil: common.GetTimestamp() + 120}).Error)

	report, err := MergeIdenticalLibraries()
	require.NoError(t, err)
	assert.Equal(t, 1, report.DeferredGroups)
	assert.Equal(t, 2, report.UnreadableAccounts, "accounts whose credential cannot be read are counted so they are not mistaken for no duplicates")
	assert.NotEqual(t, libraryOfChannel(t, db, 1), libraryOfChannel(t, db, 2), "no merge while an upload holds a lease")
	assert.NotEqual(t, libraryOfChannel(t, db, 3), libraryOfChannel(t, db, 4), "unreadable credentials are never guessed equal")

	require.NoError(t, db.Model(&model.AssetReplica{}).Where("library_id = ?", second).Updates(map[string]any{"locked_by": "", "lease_until": 0}).Error)
	report, err = MergeIdenticalLibraries()
	require.NoError(t, err)
	assert.Equal(t, 1, report.MergedGroups)
	assert.Equal(t, libraryOfChannel(t, db, 1), libraryOfChannel(t, db, 2), "the next run merges once the lease is gone")
}

func TestNewlyActiveAccountReceivesAssetsUploadedBeforeIt(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "backfill-fixture")
	for _, id := range []int{51, 52, 53} {
		require.NoError(t, db.Create(&model.Channel{Id: id, Name: "video", Type: constant.ChannelTypeVolcNative}).Error)
	}
	var assets []model.MediaAsset
	for index, status := range []string{model.AssetStatusReady, model.AssetStatusProcessing, model.AssetStatusDeleted} {
		asset := model.MediaAsset{PublicID: "asset-" + status, OwnerUserID: 7, StorageObjectID: int64(300 + index), Status: status}
		require.NoError(t, db.Create(&asset).Error)
		assets = append(assets, asset)
	}
	account, err := SaveLibraryConnection(0, LibraryConfigInput{Name: "new account", ChannelConfigInput: ChannelConfigInput{Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://provider.example", Credential: "backfill-secret", QPM: 60}})
	require.NoError(t, err)
	copiesOf := func(libraryID int) map[int64]string {
		var rows []model.AssetReplica
		require.NoError(t, db.Where("library_id = ?", libraryID).Find(&rows).Error)
		result := map[int64]string{}
		for _, row := range rows {
			result[row.AssetID] = row.Status
		}
		return result
	}
	assert.Empty(t, copiesOf(account.ID), "creating an account alone queues nothing")

	// A failed copy is never retried by a backfill, and a deleted asset is skipped.
	id := account.ID
	require.NoError(t, db.Create(&model.AssetReplica{AssetID: assets[1].ID, LibraryID: &id, Status: model.AssetReplicaStatusFailed, Operation: model.AssetReplicaOperationSync}).Error)
	_, err = BindChannelLibrary(51, account.ID, true)
	require.NoError(t, err)
	assert.Equal(t, map[int64]string{assets[0].ID: model.AssetReplicaStatusPending, assets[1].ID: model.AssetReplicaStatusFailed}, copiesOf(account.ID))

	// Assets follow the account, so linking another channel adds no copies.
	_, err = BindChannelLibrary(52, account.ID, true)
	require.NoError(t, err)
	assert.Len(t, copiesOf(account.ID), 2)
}

func TestDisabledAccountReceivesNoCopiesAndIsBackfilledWhenSwitchedOn(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "disabled-fixture")
	require.NoError(t, db.Create(&model.Channel{Id: 61, Name: "video", Type: constant.ChannelTypeVolcNative}).Error)
	before := model.MediaAsset{PublicID: "asset-before", OwnerUserID: 7, StorageObjectID: 400, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&before).Error)
	input := LibraryConfigInput{Name: "account", ChannelConfigInput: ChannelConfigInput{Enabled: false, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://provider.example", Credential: "disabled-secret", QPM: 60}}
	account, err := SaveLibraryConnection(0, input)
	require.NoError(t, err)
	_, err = BindChannelLibrary(61, account.ID, true)
	require.NoError(t, err)

	count := func() int64 {
		var total int64
		require.NoError(t, db.Model(&model.AssetReplica{}).Where("library_id = ?", account.ID).Count(&total).Error)
		return total
	}
	assert.Zero(t, count(), "binding a channel to a disabled account queues nothing")
	channelIDs, err := enabledChannelIDs(db)
	require.NoError(t, err)
	assert.NotContains(t, channelIDs, 61, "new uploads skip a disabled account")
	var conflict *RequestError
	require.ErrorAs(t, QueueLibrarySync(account.ID), &conflict)
	assert.Equal(t, 409, conflict.HTTPStatusCode())
	configs, err := assetSyncConfigs()
	require.NoError(t, err)
	require.Len(t, configs, 1)
	assert.False(t, configs[0].Enabled, "the scheduler only runs maintenance for a disabled account")

	input.Enabled = true
	_, err = SaveLibraryConnection(account.ID, input)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count(), "switching the account on queues the assets it missed")
	channelIDs, err = enabledChannelIDs(db)
	require.NoError(t, err)
	assert.Contains(t, channelIDs, 61)

	input.Enabled = false
	_, err = SaveLibraryConnection(account.ID, input)
	require.NoError(t, err)
	channelIDs, err = enabledChannelIDs(db)
	require.NoError(t, err)
	assert.NotContains(t, channelIDs, 61, "the account can be switched off again")
}

func TestDisabledAccountStopsWritingButKeepsServingSynchronizedAssets(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	library := model.AssetLibraryConnection{Name: "paused account", Enabled: false}
	require.NoError(t, db.Create(&library).Error)
	require.NoError(t, db.Create(&model.Channel{Id: 71, Name: "video", Type: constant.ChannelTypeVolcNative}).Error)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: 71, LibraryID: library.ID, Enabled: true}).Error)
	synchronized := model.MediaAsset{PublicID: "asset-synchronized", OwnerUserID: 7, StorageObjectID: 500, Status: model.AssetStatusReady}
	missed := model.MediaAsset{PublicID: "asset-missed", OwnerUserID: 7, StorageObjectID: 501, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&synchronized).Error)
	require.NoError(t, db.Create(&missed).Error)
	require.NoError(t, db.Create(&model.AssetReplica{AssetID: synchronized.ID, LibraryID: &library.ID, UpstreamAssetID: "upstream-synchronized", Status: model.AssetReplicaStatusActive, Operation: model.AssetReplicaOperationSync}).Error)

	mapped, count, err := ResolveRequestAssetIDs([]byte(`{"content":[{"image_url":{"url":"asset://asset-synchronized"}}]}`), 7, 71)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, "asset://upstream-synchronized", gjson.GetBytes(mapped, "content.0.image_url.url").String(), "a paused account still serves what it already holds")

	_, _, err = ResolveRequestAssetIDs([]byte(`{"content":[{"image_url":{"url":"asset://asset-missed"}}]}`), 7, 71)
	var conflict *RequestError
	require.ErrorAs(t, err, &conflict)
	assert.Contains(t, conflict.Error(), "has not been synchronized", "what was missed while paused is not available until it is synchronized")

	backfillMissingAssets(library.ID)
	var queued int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("library_id = ? AND asset_id = ?", library.ID, missed.ID).Count(&queued).Error)
	assert.Zero(t, queued, "a paused account is not backfilled")
	require.NoError(t, db.Model(&model.AssetLibraryConnection{}).Where("id = ?", library.ID).Update("enabled", true).Error)
	backfillMissingAssets(library.ID)
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("library_id = ? AND asset_id = ?", library.ID, missed.ID).Count(&queued).Error)
	assert.EqualValues(t, 1, queued, "turning it on queues what was missed meanwhile")
}
