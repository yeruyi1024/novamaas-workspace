package assetlibrary

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSharedLibraryQueuesOneReplicaForMultipleChannels(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	library := model.AssetLibraryConnection{Name: "shared account", Enabled: true}
	require.NoError(t, db.Create(&library).Error)
	asset := model.MediaAsset{PublicID: "asset-shared", OwnerUserID: 7, StorageObjectID: 7001, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	for _, channelID := range []int{11, 12} {
		require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: channelID, LibraryID: library.ID, Enabled: true}).Error)
	}
	require.NoError(t, model.QueueAssetReplicas(db, asset.ID, []int{11, 12}))
	var replicas []model.AssetReplica
	require.NoError(t, db.Where("asset_id = ?", asset.ID).Find(&replicas).Error)
	require.Len(t, replicas, 1)
	require.NotNil(t, replicas[0].LibraryID)
	assert.Equal(t, library.ID, *replicas[0].LibraryID)
	assert.Zero(t, replicas[0].ChannelID)
}

func TestSharedLibraryChannelOwnershipRequiresExplicitUnlink(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "exclusive-binding-fixture")
	require.NoError(t, db.Create(&model.Channel{Id: 91, Name: "video channel", Type: constant.ChannelTypeVolcNative}).Error)
	input := LibraryConfigInput{Name: "account A", ProviderName: "provider", ChannelConfigInput: ChannelConfigInput{Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://provider.example", Credential: "fixture-account-A", QPM: 60}}
	first, err := SaveLibraryConnection(0, input)
	require.NoError(t, err)
	input.Name, input.Credential = "account B", "fixture-account-B"
	second, err := SaveLibraryConnection(0, input)
	require.NoError(t, err)
	_, err = BindChannelLibrary(91, first.ID, false)
	require.NoError(t, err)
	var conflict *RequestError
	_, err = BindChannelLibrary(91, second.ID, true)
	require.ErrorAs(t, err, &conflict)
	assert.Equal(t, http.StatusConflict, conflict.HTTPStatusCode())
	var binding model.AssetChannelConfig
	require.NoError(t, db.Where("channel_id = ?", 91).First(&binding).Error)
	assert.Equal(t, first.ID, binding.LibraryID)
	assert.False(t, binding.Enabled, "disabled associations still belong to their account")
	_, err = BindChannelLibrary(91, 0, false)
	require.NoError(t, err)
	_, err = BindChannelLibrary(91, second.ID, true)
	require.NoError(t, err)
	require.NoError(t, db.Where("channel_id = ?", 91).First(&binding).Error)
	assert.Equal(t, second.ID, binding.LibraryID)
}

func TestSharedLibraryAssociationsOnlyAllowSupportedVideoChannelTypes(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	library := model.AssetLibraryConnection{Name: "account", Enabled: true}
	require.NoError(t, db.Create(&library).Error)
	for _, channel := range []model.Channel{
		{Id: 81, Type: constant.ChannelTypeDoubaoVideo, Name: "DoubaoVideo"},
		{Id: 82, Type: constant.ChannelTypeVolcNative, Name: "Volc Native"},
		{Id: 83, Type: constant.ChannelTypeVolcEngine, Name: "VolcEngine"},
		{Id: 84, Type: constant.ChannelTypeOpenAI, Name: "OpenAI"},
	} {
		require.NoError(t, db.Create(&channel).Error)
		_, err := BindChannelLibrary(channel.Id, library.ID, true)
		if channel.Id <= 82 {
			require.NoError(t, err)
		} else {
			var invalid *RequestError
			require.ErrorAs(t, err, &invalid)
			assert.Equal(t, http.StatusBadRequest, invalid.HTTPStatusCode())
		}
	}
	views, err := ListChannelConfigs()
	require.NoError(t, err)
	require.Len(t, views, 2)
	assert.Equal(t, 81, views[0].ChannelID)
	assert.Equal(t, 82, views[1].ChannelID)
}

func TestLegacyChannelConfigurationReusesAccountWithoutUploadingAndKeepsOtherChannelEnabled(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "shared-config-fixture")
	input := ChannelConfigInput{Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://provider.example", Credential: "same-fixture-secret", QPM: 60}
	var libraryID int
	for _, id := range []int{51, 52} {
		require.NoError(t, db.Create(&model.Channel{Id: id, Name: "discount channel", Type: 61}).Error)
		view, err := SaveChannelConfig(id, input)
		require.NoError(t, err)
		if libraryID == 0 {
			libraryID = view.LibraryID
		}
		assert.Equal(t, libraryID, view.LibraryID)
	}
	var libraries, replicas int64
	require.NoError(t, db.Model(&model.AssetLibraryConnection{}).Count(&libraries).Error)
	require.NoError(t, db.Model(&model.AssetReplica{}).Count(&replicas).Error)
	assert.EqualValues(t, 1, libraries)
	assert.Zero(t, replicas)
	input.Enabled, input.Credential = false, ""
	_, err := SaveChannelConfig(51, input)
	require.NoError(t, err)
	var library model.AssetLibraryConnection
	require.NoError(t, db.First(&library, libraryID).Error)
	assert.True(t, library.Enabled)
	var other model.AssetChannelConfig
	require.NoError(t, db.Where("channel_id = ?", 52).First(&other).Error)
	assert.True(t, other.Enabled)
	_, err = SaveLibraryConnection(libraryID, LibraryConfigInput{Name: "renamed account", ProviderName: "provider", ChannelConfigInput: ChannelConfigInput{Enabled: true, Protocol: input.Protocol, AuthType: input.AuthType, BaseURL: input.BaseURL, QPM: input.QPM}})
	require.NoError(t, err)
	var disabled model.AssetChannelConfig
	require.NoError(t, db.Where("channel_id = ?", 51).First(&disabled).Error)
	assert.False(t, disabled.Enabled, "editing a resource account must preserve disabled channel associations")
	assert.Equal(t, libraryID, disabled.LibraryID)
}

func TestLegacyMigrationPreservesReplicaIDsAndDoesNotMergeResourceAccounts(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	group := model.AssetGroup{PublicID: "group-legacy", OwnerUserID: 7, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	asset := model.MediaAsset{PublicID: "asset-legacy", OwnerUserID: 7, GroupID: group.ID, StorageObjectID: 7003, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	for _, channelID := range []int{21, 22} {
		require.NoError(t, db.Create(&model.Channel{Id: channelID, Name: "same provider"}).Error)
		require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: channelID, Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, EncryptedCredential: "kept-encrypted"}).Error)
		require.NoError(t, db.Create(&model.AssetGroupReplica{GroupID: group.ID, ChannelID: channelID, UpstreamGroupID: "original-group", Status: model.AssetReplicaStatusActive}).Error)
		require.NoError(t, db.Create(&model.AssetReplica{AssetID: asset.ID, ChannelID: channelID, UpstreamAssetID: "original-asset", Status: model.AssetReplicaStatusActive, Attempts: 3, LastError: "historical message"}).Error)
	}
	var before []model.AssetReplica
	require.NoError(t, db.Order("id").Find(&before).Error)
	require.NoError(t, model.MigrateAssetLibraryConnections(db))
	require.NoError(t, model.MigrateAssetLibraryConnections(db))
	var libraries []model.AssetLibraryConnection
	require.NoError(t, db.Find(&libraries).Error)
	require.Len(t, libraries, 2, "a provider name does not establish shared account permissions")
	var after []model.AssetReplica
	require.NoError(t, db.Order("id").Find(&after).Error)
	require.Len(t, after, 2)
	for index := range before {
		require.NotNil(t, after[index].LibraryID)
		assert.Equal(t, before[index].ID, after[index].ID)
		assert.Equal(t, before[index].UpstreamAssetID, after[index].UpstreamAssetID)
		assert.Equal(t, before[index].Status, after[index].Status)
		assert.Equal(t, before[index].Attempts, after[index].Attempts)
		assert.Equal(t, before[index].LastError, after[index].LastError)
	}
	assert.NotEqual(t, *after[0].LibraryID, *after[1].LibraryID)
	var migratedAsset model.MediaAsset
	require.NoError(t, db.First(&migratedAsset, asset.ID).Error)
	assert.Equal(t, asset.PublicID, migratedAsset.PublicID)
	assert.Equal(t, asset.OwnerUserID, migratedAsset.OwnerUserID)
}

func TestSharedLibraryIdentityChangeIsRejectedOnceAccountOwnsUpstreamAssets(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "credential-change-fixture")
	input := LibraryConfigInput{Name: "resource account", ProviderName: "provider", ChannelConfigInput: ChannelConfigInput{Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://provider.example", Credential: "original-fixture-key", QPM: 60}}
	first, err := SaveLibraryConnection(0, input)
	require.NoError(t, err)
	input.Name, input.Credential = "other account", "other-fixture-key"
	other, err := SaveLibraryConnection(0, input)
	require.NoError(t, err)

	// Another account's identity is never adopted, even before any upload.
	input.Credential = "original-fixture-key"
	var conflict *RequestError
	_, err = SaveLibraryConnection(other.ID, input)
	require.ErrorAs(t, err, &conflict)
	assert.Equal(t, "library_identity_exists", conflict.Code)

	// Without upstream state an account may still take a new credential.
	input.Credential = "rotated-fixture-key"
	_, err = SaveLibraryConnection(other.ID, input)
	require.NoError(t, err)

	require.NoError(t, db.Create(&model.AssetReplica{AssetID: 81, LibraryID: &first.ID, UpstreamAssetID: "existing-resource", Status: model.AssetReplicaStatusProcessing}).Error)
	input.Name, input.Credential, input.ProjectName = "resource account", "", "default"
	_, err = SaveLibraryConnection(first.ID, input)
	require.NoError(t, err, "an empty legacy project and the default project share a namespace")
	var before model.AssetLibraryConnection
	require.NoError(t, db.First(&before, first.ID).Error)
	input.Credential = "another-account-fixture-key"
	_, err = SaveLibraryConnection(first.ID, input)
	require.ErrorAs(t, err, &conflict)
	assert.Equal(t, "library_identity_locked", conflict.Code)
	var after model.AssetLibraryConnection
	require.NoError(t, db.First(&after, first.ID).Error)
	assert.Equal(t, before.EncryptedCredential, after.EncryptedCredential)
}

func TestBackfilledResourceKeyLetsNewAccountReuseMigratedAccount(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "backfill-fixture")
	secret, err := encryptChannelCredential("shared-fixture-secret")
	require.NoError(t, err)
	account := model.AssetLibraryConnection{Name: "migrated", Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://provider.example", QPM: 60, EncryptedCredential: secret, CredentialKeyVersion: channelCredentialVersion}
	unreadable := model.AssetLibraryConnection{Name: "unreadable", Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://provider.example", QPM: 60, EncryptedCredential: "unreadable-ciphertext"}
	require.NoError(t, db.Create(&account).Error)
	require.NoError(t, db.Create(&unreadable).Error)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: 1, LibraryID: account.ID, Enabled: true}).Error)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: 2, LibraryID: unreadable.ID, Enabled: true}).Error)
	require.NoError(t, RecordLibraryIdentities())
	require.NoError(t, db.First(&account, account.ID).Error)
	require.NoError(t, db.First(&unreadable, unreadable.ID).Error)
	assert.NotNil(t, account.ResourceKey)
	assert.Nil(t, unreadable.ResourceKey, "an account that cannot be decrypted is never guessed into a shared identity")

	view, err := SaveLibraryConnection(0, LibraryConfigInput{Name: "duplicate", ProviderName: "provider", ChannelConfigInput: ChannelConfigInput{Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://provider.example/", Credential: "shared-fixture-secret", QPM: 60}})
	require.NoError(t, err)
	assert.Equal(t, account.ID, view.ID)
	assert.True(t, view.ReusedExisting)
	var libraries int64
	require.NoError(t, db.Model(&model.AssetLibraryConnection{}).Count(&libraries).Error)
	assert.EqualValues(t, 2, libraries)
}

func TestLinkingChannelsPreservesExistingReplicasWithoutSync(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "shared-link-fixture")
	for _, channelID := range []int{31, 32} {
		require.NoError(t, db.Create(&model.Channel{Id: channelID, Name: "discount channel", Type: 61}).Error)
	}
	input := LibraryConfigInput{Name: "account one", ProviderName: "provider", ChannelConfigInput: ChannelConfigInput{
		Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://provider.example", Credential: "fixture-secret", QPM: 60,
	}}
	first, err := SaveLibraryConnection(0, input)
	require.NoError(t, err)
	_, err = BindChannelLibrary(31, first.ID, true)
	require.NoError(t, err)
	asset := model.MediaAsset{PublicID: "asset-existing", OwnerUserID: 7, StorageObjectID: 7004, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	replica := model.AssetReplica{AssetID: asset.ID, LibraryID: &first.ID, UpstreamAssetID: "active-upstream", Status: model.AssetReplicaStatusActive}
	require.NoError(t, db.Create(&replica).Error)
	input.Name = "account two"
	input.Credential = "different-resource-account-secret"
	second, err := SaveLibraryConnection(0, input)
	require.NoError(t, err)
	_, err = BindChannelLibrary(32, second.ID, true)
	require.NoError(t, err)
	var before int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Count(&before).Error)
	_, err = BindChannelLibrary(32, 0, false)
	require.NoError(t, err)
	_, err = BindChannelLibrary(32, first.ID, true)
	require.NoError(t, err)
	var after int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Count(&after).Error)
	assert.Equal(t, before, after)
	var kept model.AssetReplica
	require.NoError(t, db.First(&kept, replica.ID).Error)
	assert.Equal(t, "active-upstream", kept.UpstreamAssetID)
	assert.Equal(t, model.AssetReplicaStatusActive, kept.Status)
	views, err := ListLibraryConnections()
	require.NoError(t, err)
	require.Len(t, views, 2)
	assert.Len(t, views[0].Channels, 2)
	assert.Empty(t, views[1].Channels)
	assert.Equal(t, second.ID, views[1].ID)
	assert.EqualValues(t, 1, views[0].Summary.Active)
}

func TestSharedLibraryPollsOnceAndUsesOneRateBudgetAcrossChannels(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "shared-poll-fixture")
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GetAsset", r.URL.Query().Get("Action"))
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"Result":{"Status":"Active"}}`)
	}))
	t.Cleanup(upstream.Close)
	credential, err := encryptChannelCredential("fixture-secret")
	require.NoError(t, err)
	library := model.AssetLibraryConnection{Name: "shared", Enabled: true, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, BaseURL: upstream.URL, QPM: 1000000, EncryptedCredential: credential}
	require.NoError(t, db.Create(&library).Error)
	for _, channelID := range []int{41, 42} {
		require.NoError(t, db.Create(&model.Channel{Id: channelID, Name: "shared channel"}).Error)
		require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: channelID, LibraryID: library.ID, Enabled: true}).Error)
	}
	group := model.AssetGroup{PublicID: "group-poll", OwnerUserID: 7, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	asset := model.MediaAsset{PublicID: "asset-poll", OwnerUserID: 7, GroupID: group.ID, StorageObjectID: 7005, Status: model.AssetStatusProcessing}
	require.NoError(t, db.Create(&asset).Error)
	replica := model.AssetReplica{AssetID: asset.ID, LibraryID: &library.ID, UpstreamAssetID: "shared-upload", Status: model.AssetReplicaStatusProcessing, Operation: model.AssetReplicaOperationSync, NextSyncAt: common.GetTimestamp()}
	require.NoError(t, db.Create(&replica).Error)
	scheduler := newAssetSyncScheduler("shared-poll")
	scheduler.dispatch(context.Background())
	scheduler.jobs.Wait()
	assert.EqualValues(t, 1, calls.Load())
	require.NoError(t, db.First(&replica, replica.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusActive, replica.Status)
	for _, channelID := range []int{41, 42} {
		config := library.ProviderConfig(channelID)
		assert.Equal(t, -library.ID, config.RateLimitID())
	}
}

func TestSharedLibraryResolvesSameUpstreamAssetThroughBothChannels(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	library := model.AssetLibraryConnection{Name: "shared account", Enabled: true}
	require.NoError(t, db.Create(&library).Error)
	asset := model.MediaAsset{PublicID: "asset-shared", OwnerUserID: 7, StorageObjectID: 7002, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	for _, channelID := range []int{11, 12} {
		require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: channelID, LibraryID: library.ID, Enabled: true}).Error)
	}
	require.NoError(t, db.Create(&model.AssetReplica{AssetID: asset.ID, LibraryID: &library.ID, UpstreamAssetID: "upstream-shared", Status: model.AssetReplicaStatusActive}).Error)
	body := []byte(`{"content":[{"image_url":{"url":"asset://asset-shared"}}]}`)
	for _, channelID := range []int{11, 12} {
		mapped, count, err := ResolveRequestAssetIDs(body, 7, channelID)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, "asset://upstream-shared", gjson.GetBytes(mapped, "content.0.image_url.url").String())
	}
	_, _, err := ResolveRequestAssetIDs(body, 8, 11)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not accessible")
	other := model.AssetLibraryConnection{Name: "other account", Enabled: true}
	require.NoError(t, db.Create(&other).Error)
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: 13, LibraryID: other.ID, Enabled: true}).Error)
	_, _, err = ResolveRequestAssetIDs(body, 7, 13)
	require.Error(t, err, "another account cannot fall back to the first account's upstream ID")
}

func TestPausedProcessingResumesAfterReenable(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "pr89-review-fixture")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"Result":{"Status":"Active"}}`)
	}))
	t.Cleanup(upstream.Close)
	input := LibraryConfigInput{Name: "paused account", ChannelConfigInput: ChannelConfigInput{
		Enabled: false, Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer,
		BaseURL: upstream.URL, Credential: "fixture-secret", QPM: 60,
	}}
	library, err := SaveLibraryConnection(0, input)
	require.NoError(t, err)
	require.NoError(t, db.Create(&model.Channel{Id: 1, Name: "channel", Type: 61}).Error)
	_, err = BindChannelLibrary(1, library.ID, true)
	require.NoError(t, err)
	group := model.AssetGroup{PublicID: "group-paused", OwnerUserID: 7, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	asset := model.MediaAsset{PublicID: "asset-paused", OwnerUserID: 7, GroupID: group.ID, StorageObjectID: 200, Status: model.AssetStatusProcessing}
	require.NoError(t, db.Create(&asset).Error)
	now := common.GetTimestamp()
	replica := model.AssetReplica{AssetID: asset.ID, LibraryID: &library.ID, UpstreamAssetID: "working-upstream",
		Operation: model.AssetReplicaOperationSync, Status: model.AssetReplicaStatusProcessing, Progress: 75,
		CreatedAt: now - 600, NextSyncAt: now,
	}
	require.NoError(t, db.Create(&replica).Error)
	scheduler := newAssetSyncScheduler("pr89-review")
	paused, err := scheduler.claimNextConfig(model.AssetChannelConfig{LibraryID: library.ID, Enabled: false})
	require.NoError(t, err)
	require.Nil(t, paused)
	for index, state := range []string{model.AssetReplicaStatusFailed, model.AssetReplicaStatusActive, model.AssetReplicaStatusSyncing} {
		other := model.AssetReplica{AssetID: asset.ID + int64(index) + 1, LibraryID: &library.ID, Operation: model.AssetReplicaOperationSync,
			Status: state, CreatedAt: now - 600, NextSyncAt: now + 3600}
		if state == model.AssetReplicaStatusSyncing {
			other.LockedBy, other.LeaseUntil = "live-worker", now+120
		}
		require.NoError(t, db.Create(&other).Error)
	}
	input.Enabled, input.Credential = true, ""
	_, err = SaveLibraryConnection(library.ID, input)
	require.NoError(t, err)
	var untouched []model.AssetReplica
	require.NoError(t, db.Where("library_id = ? AND asset_id <> ?", library.ID, asset.ID).Find(&untouched).Error)
	require.Len(t, untouched, 3)
	for _, other := range untouched {
		assert.Equal(t, now-600, other.CreatedAt, "failed, active and currently leased work retains its budget")
		assert.Equal(t, now+3600, other.NextSyncAt)
	}
	claimed, err := scheduler.claimNextConfig(model.AssetChannelConfig{LibraryID: library.ID, Enabled: true})
	require.NoError(t, err)
	require.NotNil(t, claimed)
	syncReplica(context.Background(), claimed.LockedBy, claimed)
	require.NoError(t, db.First(&replica, replica.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusActive, replica.Status, "an account pause must not exhaust the processing timeout: %s", replica.LastError)
}

func TestMergedHistoricalDeletionSharesCurrentAccountBudget(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "shared-delete-fixture")
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DeleteAsset", r.URL.Query().Get("Action"))
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"Result":{}}`)
	}))
	t.Cleanup(upstream.Close)
	for _, channelID := range []int{1, 2} {
		seedLegacyChannel(t, db, channelID, upstream.URL, "fixture-ak", "fixture-secret", true)
	}
	asset := model.MediaAsset{PublicID: "asset-delete-budget", OwnerUserID: 7, StorageObjectID: 201, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&asset).Error)
	seedLegacyCopy(t, db, asset.ID, 1, "first-copy", model.AssetReplicaStatusActive)
	seedLegacyCopy(t, db, asset.ID, 2, "historical-copy", model.AssetReplicaStatusActive)
	require.NoError(t, model.MigrateAssetLibraryConnections(db))
	report, err := MergeIdenticalLibraries()
	require.NoError(t, err)
	require.Equal(t, 1, report.MergedGroups)
	var library model.AssetLibraryConnection
	require.NoError(t, db.First(&library, libraryOfChannel(t, db, 1)).Error)
	require.NoError(t, model.RequestAssetDeletion(db, asset.ID))
	// Occupy the account's next request slot without relying on wall-clock sleeps.
	reserveChannelRate(library.ProviderConfig(0).RateLimitID(), 1)
	scheduler := newAssetSyncScheduler("historical-delete-budget")
	scheduler.dispatch(context.Background())
	scheduler.jobs.Wait()
	assert.Zero(t, calls.Load(), "neither current nor historical copies may bypass the occupied account budget")
	for range 2 {
		channelRateState.Lock()
		delete(channelRateState.next, library.ProviderConfig(0).RateLimitID())
		delete(channelRateState.sentNext, library.ProviderConfig(0).RateLimitID())
		channelRateState.Unlock()
		scheduler.dispatch(context.Background())
		scheduler.jobs.Wait()
	}
	assert.EqualValues(t, 2, calls.Load(), "both accounts retain their own deletion queue while sharing one budget")
	var remaining int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("asset_id = ? AND status <> ?", asset.ID, model.AssetReplicaStatusDeleted).Count(&remaining).Error)
	assert.Zero(t, remaining)
}

func TestFirstEnabledChannelResumesPausedWorkButAdditionalChannelsDoNot(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	library := model.AssetLibraryConnection{Name: "unlinked", Enabled: true}
	require.NoError(t, db.Create(&library).Error)
	for _, id := range []int{1, 2} {
		require.NoError(t, db.Create(&model.Channel{Id: id, Name: "video channel", Type: constant.ChannelTypeVolcNative}).Error)
	}
	asset := model.MediaAsset{PublicID: "asset-relink", OwnerUserID: 7, StorageObjectID: 202, Status: model.AssetStatusProcessing}
	require.NoError(t, db.Create(&asset).Error)
	now := common.GetTimestamp()
	replica := model.AssetReplica{AssetID: asset.ID, LibraryID: &library.ID, Operation: model.AssetReplicaOperationSync,
		Status: model.AssetReplicaStatusProcessing, UpstreamAssetID: "already-submitted", CreatedAt: now - 600}
	require.NoError(t, db.Create(&replica).Error)
	_, err := BindChannelLibrary(1, library.ID, true)
	require.NoError(t, err)
	require.NoError(t, db.First(&replica, replica.ID).Error)
	assert.GreaterOrEqual(t, replica.CreatedAt, now)
	assert.Equal(t, "already-submitted", replica.UpstreamAssetID)
	require.NoError(t, db.Model(&replica).Update("created_at", now-60).Error)
	_, err = BindChannelLibrary(2, library.ID, true)
	require.NoError(t, err)
	require.NoError(t, db.First(&replica, replica.ID).Error)
	assert.Equal(t, now-60, replica.CreatedAt, "adding a channel must not extend an account that was already synchronizing")
}
