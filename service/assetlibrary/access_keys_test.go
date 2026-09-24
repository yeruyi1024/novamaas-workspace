package assetlibrary

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetAccessKeyAuthenticatesOfficialVolcengineSignature(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	previousRedisEnabled := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() { common.RedisEnabled = previousRedisEnabled })
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "asset-access-key-test-master-secret")
	user := model.User{Username: "asset-api-user", Password: "password", Status: common.UserStatusEnabled, Group: "default"}
	require.NoError(t, db.Create(&user).Error)

	created, err := CreateAccessKey(user.Id, "automation")
	require.NoError(t, err)
	require.NotEmpty(t, created.SecretAccessKey)
	var stored model.AssetAccessKey
	require.NoError(t, db.First(&stored, created.ID).Error)
	assert.NotContains(t, stored.EncryptedSecret, created.SecretAccessKey)

	fixedTime := time.Date(2026, time.September, 22, 8, 30, 45, 0, time.UTC)
	body := []byte(`{"PageNumber":1,"PageSize":20}`)
	request, err := http.NewRequest(http.MethodPost, "https://gateway.example/?Action=ListAssets&Version=2024-01-01", bytes.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	signer := volcActionClient{
		config: &model.AssetChannelConfig{
			AccessKeyID: created.AccessKeyID,
			Region:      DefaultRegion,
			Service:     DefaultService,
		},
		secret: created.SecretAccessKey,
		now:    func() time.Time { return fixedTime },
	}
	signer.sign(request, body)

	ownerUserID, err := authenticateVolcActionRequestAt(request, body, fixedTime.Add(time.Minute))
	require.NoError(t, err)
	assert.Equal(t, user.Id, ownerUserID)
	principal, err := authenticateVolcActionRequestPrincipalAt(request, body, fixedTime.Add(time.Minute))
	require.NoError(t, err)
	assert.Equal(t, user.Id, principal.OwnerUserID)
	assert.Equal(t, created.AccessKeyID, principal.AccessKeyID)
	assert.Equal(t, "automation", principal.AccessKeyName)

	_, err = authenticateVolcActionRequestAt(request, append(body, ' '), fixedTime.Add(time.Minute))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "X-Content-Sha256")
}

func TestListAssetsPaginatesAndSearchesUploader(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	alice := model.User{Username: "alice-uploader", Password: "password", Status: common.UserStatusEnabled, AffCode: "asset-alice"}
	bob := model.User{Username: "bob-uploader", Password: "password", Status: common.UserStatusEnabled, AffCode: "asset-bob"}
	require.NoError(t, db.Create(&alice).Error)
	require.NoError(t, db.Create(&bob).Error)
	aliceGroup := model.AssetGroup{PublicID: "group-alice", OwnerUserID: alice.Id, Name: "Alice", Status: model.AssetStatusReady}
	bobGroup := model.AssetGroup{PublicID: "group-bob", OwnerUserID: bob.Id, Name: "Bob", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&aliceGroup).Error)
	require.NoError(t, db.Create(&bobGroup).Error)
	assets := []model.MediaAsset{
		{PublicID: "asset-alice-new", OwnerUserID: alice.Id, GroupID: aliceGroup.ID, StorageObjectID: 3001, Name: "New portrait", AssetType: model.AssetTypeImage, ContentType: "image/png", Size: 10, Status: model.AssetStatusReady, CreatedAt: 30},
		{PublicID: "asset-alice-old", OwnerUserID: alice.Id, GroupID: aliceGroup.ID, StorageObjectID: 3002, Name: "Old portrait", AssetType: model.AssetTypeImage, ContentType: "image/png", Size: 10, Status: model.AssetStatusReady, CreatedAt: 20},
		{PublicID: "asset-bob", OwnerUserID: bob.Id, GroupID: bobGroup.ID, StorageObjectID: 3003, Name: "Landscape", AssetType: model.AssetTypeImage, ContentType: "image/png", Size: 10, Status: model.AssetStatusReady, CreatedAt: 10},
	}
	require.NoError(t, db.Create(&assets).Error)

	firstPage, err := ListAssets(AssetListInput{OwnerUserID: alice.Id, IncludeAllOwners: true, Search: "alice-uploader", Page: 1, PageSize: 1})
	require.NoError(t, err)
	assert.EqualValues(t, 2, firstPage.Total)
	require.Len(t, firstPage.Items, 1)
	assert.Equal(t, "asset-alice-new", firstPage.Items[0].ID)
	assert.Equal(t, "alice-uploader", firstPage.Items[0].OwnerName)

	secondPage, err := ListAssets(AssetListInput{OwnerUserID: alice.Id, IncludeAllOwners: true, Search: "alice-uploader", Page: 2, PageSize: 1})
	require.NoError(t, err)
	require.Len(t, secondPage.Items, 1)
	assert.Equal(t, "asset-alice-old", secondPage.Items[0].ID)
}

func TestListGroupsScopesMembersAndHydratesCreatorsForAdmins(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	alice := model.User{Username: "alice-groups", Password: "password", Status: common.UserStatusEnabled, AffCode: "groups-alice"}
	bob := model.User{Username: "bob-groups", Password: "password", Status: common.UserStatusEnabled, AffCode: "groups-bob"}
	require.NoError(t, db.Create(&alice).Error)
	require.NoError(t, db.Create(&bob).Error)
	require.NoError(t, db.Create(&[]model.AssetGroup{
		{PublicID: "group-alice", OwnerUserID: alice.Id, Name: "Campaign", Status: model.AssetStatusReady},
		{PublicID: "group-bob", OwnerUserID: bob.Id, Name: "Campaign", Status: model.AssetStatusReady},
	}).Error)

	memberGroups, err := ListGroups(alice.Id, false)
	require.NoError(t, err)
	require.Len(t, memberGroups, 1)
	assert.Equal(t, "group-alice", memberGroups[0].ID)
	assert.Equal(t, "alice-groups", memberGroups[0].OwnerName)

	adminGroups, err := ListGroups(alice.Id, true)
	require.NoError(t, err)
	require.Len(t, adminGroups, 2)
	creators := make(map[string]string, len(adminGroups))
	for _, group := range adminGroups {
		creators[group.ID] = group.OwnerName
	}
	assert.Equal(t, map[string]string{
		"group-alice": "alice-groups",
		"group-bob":   "bob-groups",
	}, creators)
}

func TestListGroupsPageSortsByNameAndScopesSearchBeforePagination(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	owner := model.User{Username: "group-page-owner", Password: "password", Status: common.UserStatusEnabled, AffCode: "group-page-owner"}
	other := model.User{Username: "group-page-other", Password: "password", Status: common.UserStatusEnabled, AffCode: "group-page-other"}
	require.NoError(t, db.Create(&owner).Error)
	require.NoError(t, db.Create(&other).Error)
	require.NoError(t, db.Create(&[]model.AssetGroup{
		{PublicID: "group-zeta", OwnerUserID: owner.Id, Name: "Zeta", Status: model.AssetStatusReady},
		{PublicID: "group-alpha", OwnerUserID: owner.Id, Name: "Alpha", Status: model.AssetStatusReady},
		{PublicID: "group-alpine", OwnerUserID: owner.Id, Name: "Alpine", Status: model.AssetStatusReady},
		{PublicID: "group-other", OwnerUserID: other.Id, Name: "Albatross", Status: model.AssetStatusReady},
	}).Error)

	first, err := ListGroupsPage(GroupListInput{OwnerUserID: owner.Id, Search: "al", Page: 1, PageSize: 1, SortBy: "Name", SortOrder: "Asc"})
	require.NoError(t, err)
	assert.EqualValues(t, 2, first.Total)
	require.Len(t, first.Items, 1)
	assert.Equal(t, "Alpha", first.Items[0].Name)

	second, err := ListGroupsPage(GroupListInput{OwnerUserID: owner.Id, Search: "al", Page: 2, PageSize: 1, SortBy: "Name", SortOrder: "Asc"})
	require.NoError(t, err)
	require.Len(t, second.Items, 1)
	assert.Equal(t, "Alpine", second.Items[0].Name)

	admin, err := ListGroupsPage(GroupListInput{OwnerUserID: owner.Id, IncludeAllOwners: true, Search: "al", Page: 1, PageSize: 10, SortBy: "Name", SortOrder: "Asc"})
	require.NoError(t, err)
	assert.EqualValues(t, 3, admin.Total)
	require.Len(t, admin.Items, 3)
	assert.Equal(t, "Albatross", admin.Items[0].Name)
}
