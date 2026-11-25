package assetlibrary

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newClearAccount creates a resource account linked to its own enabled
// channel (none when channelID is 0); a paused account is disabled.
func newClearAccount(t *testing.T, db *gorm.DB, name string, channelID int, paused bool) *model.AssetLibraryConnection {
	library := &model.AssetLibraryConnection{Name: name, Enabled: true}
	require.NoError(t, db.Create(library).Error)
	if paused {
		require.NoError(t, db.Model(library).Update("enabled", false).Error)
	}
	if channelID > 0 {
		require.NoError(t, db.Create(&model.Channel{Id: channelID, Name: fmt.Sprintf("channel-%d", channelID), Type: constant.ChannelTypeVolcNative}).Error)
		require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: channelID, LibraryID: library.ID, Enabled: true}).Error)
	}
	return library
}

// clearFixture creates assets of one owner and their copies on accounts.
type clearFixture struct {
	t     *testing.T
	db    *gorm.DB
	group model.AssetGroup
}

func newClearFixture(t *testing.T, db *gorm.DB, ownerID int) clearFixture {
	group := model.AssetGroup{PublicID: fmt.Sprintf("grp-clear-%d", ownerID), OwnerUserID: ownerID, Name: "Clear", Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	return clearFixture{t: t, db: db, group: group}
}

func (f clearFixture) asset(publicID string, status string) model.MediaAsset {
	asset := model.MediaAsset{PublicID: publicID, OwnerUserID: f.group.OwnerUserID, GroupID: f.group.ID, StorageObjectID: int64(len(publicID)), Status: status}
	require.NoError(f.t, f.db.Create(&asset).Error)
	return asset
}

func (f clearFixture) copy(asset model.MediaAsset, library *model.AssetLibraryConnection, replica model.AssetReplica) model.AssetReplica {
	replica.AssetID, replica.LibraryID, replica.Operation = asset.ID, &library.ID, model.AssetReplicaOperationSync
	require.NoError(f.t, f.db.Create(&replica).Error)
	return replica
}

func TestClearLibraryFailuresDeletesUnusableAssetsAndKeepsUsableOnes(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	cleared := newClearAccount(t, db, "account-with-failures", 77, false)
	other := newClearAccount(t, db, "healthy-account", 78, false)
	f := newClearFixture(t, db, 7)
	failed := func(upstreamID string, nextSyncAt int64) model.AssetReplica {
		return model.AssetReplica{Status: model.AssetReplicaStatusFailed, UpstreamAssetID: upstreamID, NextSyncAt: nextSyncAt, Attempts: 1, LastError: "failed"}
	}

	failedEverywhere := f.asset("failed-everywhere", model.AssetStatusSyncFailed)
	f.copy(failedEverywhere, cleared, failed("", 0))
	rejected := f.asset("rejected", model.AssetStatusUnavailable)
	f.copy(rejected, cleared, model.AssetReplica{Status: model.AssetReplicaStatusRejected})
	usableNeverAccepted := f.asset("usable-never-accepted", model.AssetStatusReady)
	neverAccepted := f.copy(usableNeverAccepted, cleared, failed("", 0))
	healthy := f.copy(usableNeverAccepted, other, model.AssetReplica{Status: model.AssetReplicaStatusActive, UpstreamAssetID: "healthy-copy"})
	usableWithLeftover := f.asset("usable-with-leftover", model.AssetStatusReady)
	leftover := f.copy(usableWithLeftover, cleared, failed("failed-upstream-copy", 0))
	stillRetrying := f.asset("still-retrying", model.AssetStatusProcessing)
	retrying := f.copy(stillRetrying, cleared, failed("", 4102444800))
	syncingElsewhere := f.asset("syncing-elsewhere", model.AssetStatusProcessing)
	waiting := f.copy(syncingElsewhere, cleared, failed("", 0))
	f.copy(syncingElsewhere, other, model.AssetReplica{Status: model.AssetReplicaStatusProcessing, UpstreamAssetID: "still-processing", NextSyncAt: 4102444800})
	otherAccountFailure := f.asset("other-account-failure", model.AssetStatusReady)
	untouched := f.copy(otherAccountFailure, other, failed("", 0))

	before, err := libraryConnectionView(cleared.ID)
	require.NoError(t, err)
	assert.EqualValues(t, 5, before.Summary.Failed, "the account shows only final failures")
	assert.EqualValues(t, 1, before.Summary.Processing, "a failure waiting for its automatic retry is still synchronizing")

	result, err := ClearLibraryFailures(cleared.ID, 7)
	require.NoError(t, err)
	assert.Equal(t, LibraryFailureCleanup{DeletedAssets: 2, DismissedReplicas: 2}, *result)

	for _, asset := range []model.MediaAsset{failedEverywhere, rejected} {
		require.NoError(t, db.First(&asset, asset.ID).Error)
		assert.Equal(t, model.AssetStatusDeleted, asset.Status, "an asset unusable everywhere is deleted: %s", asset.PublicID)
	}
	for _, asset := range []model.MediaAsset{usableNeverAccepted, usableWithLeftover, stillRetrying, otherAccountFailure} {
		require.NoError(t, db.First(&asset, asset.ID).Error)
		assert.NotEqual(t, model.AssetStatusDeleted, asset.Status, "a usable asset is never deleted: %s", asset.PublicID)
	}
	require.NoError(t, db.First(&neverAccepted, neverAccepted.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusDeleted, neverAccepted.Status)
	require.NoError(t, db.First(&leftover, leftover.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusDeleting, leftover.Status, "a copy the upstream still holds is queued for upstream deletion")
	assert.Equal(t, model.AssetReplicaOperationDelete, leftover.Operation)
	require.NoError(t, db.First(&healthy, healthy.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusActive, healthy.Status)
	require.NoError(t, db.First(&retrying, retrying.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusFailed, retrying.Status, "a failure still waiting for its automatic retry is kept")
	require.NoError(t, db.First(&waiting, waiting.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusFailed, waiting.Status, "an asset still syncing on another account is not judged yet")
	require.NoError(t, db.First(&untouched, untouched.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusFailed, untouched.Status, "another account's failures are not touched")

	require.NoError(t, QueueLibrarySync(cleared.ID))
	libraryFills.wg.Wait()
	require.NoError(t, db.First(&neverAccepted, neverAccepted.ID).Error)
	assert.Equal(t, model.AssetReplicaOperationSync, neverAccepted.Operation, "an explicit retry uploads a cleared copy again")
	assert.Equal(t, model.AssetReplicaStatusPending, neverAccepted.Status)
	assert.Empty(t, neverAccepted.UpstreamAssetID)
}

func TestUpstreamDeletionStopsRetryingAfterAboutADay(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, `{"ResponseMetadata":{"Error":{"Code":"ProviderGone","Message":"provider shut down"}}}`)
	}))
	t.Cleanup(upstream.Close)
	seedSyncChannel(t, 7701, 1000000, model.AssetChannelProtocolVolcAction, upstream.URL)
	now := common.GetTimestamp()
	replicas := seedSyncReplicas(t, 7701,
		model.AssetReplica{Status: model.AssetReplicaStatusDeleting, UpstreamAssetID: "gone-copy", Attempts: assetDeleteMaxAttempts - 1, NextSyncAt: now - 1},
		model.AssetReplica{Status: model.AssetReplicaStatusDeleting, UpstreamAssetID: "retry-copy", NextSyncAt: now - 1},
	)
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("channel_id = ?", 7701).Update("operation", model.AssetReplicaOperationDelete).Error)
	require.NoError(t, db.Model(&model.MediaAsset{}).Where("id IN ?", []int64{replicas[0].AssetID, replicas[1].AssetID}).Update("status", model.AssetStatusDeleted).Error)
	// Synchronization failures before a deletion must not use up its retries.
	syncFailures := seedSyncReplicas(t, 7702, model.AssetReplica{Status: model.AssetReplicaStatusFailed,
		UpstreamAssetID: "failed-sync-copy", Attempts: assetDeleteMaxAttempts - 1, NextSyncAt: now - 1})[0]
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return model.RequestAssetDeletion(tx, syncFailures.AssetID) }))
	require.NoError(t, db.First(&syncFailures, syncFailures.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusDeleting, syncFailures.Status)
	assert.Zero(t, syncFailures.Attempts, "a deletion starts with a full retry budget")

	for index := range replicas {
		replica := replicas[index]
		require.NoError(t, db.First(&replica, replica.ID).Error)
		claimed, err := model.ClaimAssetReplica(&replica, now, now+120, "deleter")
		require.NoError(t, err)
		require.True(t, claimed)
		syncReplica(context.Background(), "deleter", &replica)
	}

	var gaveUp, retrying model.AssetReplica
	require.NoError(t, db.First(&gaveUp, replicas[0].ID).Error)
	assert.Equal(t, model.AssetReplicaStatusDeleted, gaveUp.Status, "the last allowed failure closes the deletion")
	require.NoError(t, db.First(&retrying, replicas[1].ID).Error)
	assert.Equal(t, model.AssetReplicaStatusDeleting, retrying.Status)
	assert.Equal(t, 1, retrying.Attempts)
	assert.Greater(t, retrying.NextSyncAt, now, "earlier failures keep retrying with backoff")
}

func TestClearingAPausedAccountSettlesAssetsWaitingOnIt(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	require.NoError(t, db.Create(&model.User{Id: 1, Username: "root-admin", Role: common.RoleRootUser, Status: common.UserStatusEnabled, AffCode: "clear-root"}).Error)
	working := newClearAccount(t, db, "working account", 81, false)
	cancelled := newClearAccount(t, db, "cancelled provider", 82, true)
	alsoCancelled := newClearAccount(t, db, "another cancelled provider", 0, true)
	// Owner 9 no longer exists: an administrator may still clear what it left unusable.
	f := newClearFixture(t, db, 9)
	copyOf := func(asset model.MediaAsset, library *model.AssetLibraryConnection, status string, nextSyncAt int64) model.AssetReplica {
		return f.copy(asset, library, model.AssetReplica{Status: status, NextSyncAt: nextSyncAt})
	}

	readyElsewhere := f.asset("ready-elsewhere", model.AssetStatusReady)
	copyOf(readyElsewhere, working, model.AssetReplicaStatusActive, 0)
	pausedCopy := copyOf(readyElsewhere, cancelled, model.AssetReplicaStatusPending, 1)
	failedElsewhere := f.asset("failed-elsewhere", model.AssetStatusProcessing)
	copyOf(failedElsewhere, working, model.AssetReplicaStatusFailed, 0)
	copyOf(failedElsewhere, cancelled, model.AssetReplicaStatusPending, 1)
	syncingElsewhere := f.asset("syncing-elsewhere", model.AssetStatusProcessing)
	stillUploading := copyOf(syncingElsewhere, working, model.AssetReplicaStatusProcessing, 4102444800)
	releasedCopy := copyOf(syncingElsewhere, cancelled, model.AssetReplicaStatusPending, 1)
	onlyCancelled := f.asset("only-cancelled-providers", model.AssetStatusProcessing)
	copyOf(onlyCancelled, cancelled, model.AssetReplicaStatusProcessing, 1)
	copyOf(onlyCancelled, alsoCancelled, model.AssetReplicaStatusPending, 1)

	result, err := ClearLibraryFailures(cancelled.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, LibraryFailureCleanup{DeletedAssets: 2, DismissedReplicas: 2}, *result)
	for publicID, want := range map[string]string{
		"ready-elsewhere": model.AssetStatusReady, "failed-elsewhere": model.AssetStatusDeleted,
		"syncing-elsewhere": model.AssetStatusProcessing, "only-cancelled-providers": model.AssetStatusDeleted,
	} {
		var asset model.MediaAsset
		require.NoError(t, db.Where("public_id = ?", publicID).First(&asset).Error)
		assert.Equal(t, want, asset.Status, publicID)
	}
	require.NoError(t, db.First(&releasedCopy, releasedCopy.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusDeleted, releasedCopy.Status, "the paused copy is released while the asset keeps uploading on a working account")
	require.NoError(t, db.First(&stillUploading, stillUploading.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusProcessing, stillUploading.Status)

	require.NoError(t, db.First(&pausedCopy, pausedCopy.ID).Error)
	assert.Equal(t, model.AssetReplicaStatusDeleted, pausedCopy.Status, "a copy usable elsewhere is only dropped from the paused account")
}

func TestReleasedDeletionKeepsItsRetryCount(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	now := common.GetTimestamp()
	replica := seedSyncReplicas(t, 7703, model.AssetReplica{Status: model.AssetReplicaStatusDeleting, UpstreamAssetID: "slow-delete", Attempts: 7, NextSyncAt: now - 1})[0]
	require.NoError(t, db.Model(&replica).Update("operation", model.AssetReplicaOperationDelete).Error)
	require.NoError(t, db.Model(&model.MediaAsset{}).Where("id = ?", replica.AssetID).Update("status", model.AssetStatusDeleted).Error)
	require.NoError(t, db.First(&replica, replica.ID).Error)
	claimed, err := model.ClaimAssetReplica(&replica, now, now+120, "restarting")
	require.NoError(t, err)
	require.True(t, claimed)
	// A restart cancels the job, which hands the deletion back unchanged.
	require.NoError(t, model.FinishAssetReplica(replica.ID, "restarting", model.AssetReplicaStatusDeleting, replica.Progress, replica.UpstreamAssetID, now, "", model.AssetReplicaOperationDelete))
	require.NoError(t, db.First(&replica, replica.ID).Error)
	assert.Equal(t, 7, replica.Attempts, "pausing a deletion must not restart its retry budget")
}

func TestClearingAWorkingAccountRespectsOtherAccountsAndLeases(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	working := newClearAccount(t, db, "working account", 91, false)
	paused := newClearAccount(t, db, "paused for maintenance", 92, true)
	healthy := newClearAccount(t, db, "healthy account", 93, false)
	f := newClearFixture(t, db, 7)
	copyOf := func(asset model.MediaAsset, library *model.AssetLibraryConnection, status string, lockedBy string) {
		replica := model.AssetReplica{Status: status, LockedBy: lockedBy}
		if lockedBy != "" {
			replica.LeaseUntil = common.GetTimestamp() + 120
		}
		f.copy(asset, library, replica)
	}

	waitingForResume := f.asset("waiting-for-resume", model.AssetStatusProcessing)
	copyOf(waitingForResume, working, model.AssetReplicaStatusFailed, "")
	copyOf(waitingForResume, paused, model.AssetReplicaStatusPending, "")
	rejectedByReview := f.asset("rejected-by-review", model.AssetStatusUnavailable)
	copyOf(rejectedByReview, working, model.AssetReplicaStatusRejected, "")
	copyOf(rejectedByReview, healthy, model.AssetReplicaStatusActive, "")
	beingProcessed := f.asset("being-processed", model.AssetStatusSyncFailed)
	copyOf(beingProcessed, working, model.AssetReplicaStatusFailed, "worker-holding-it")

	result, err := ClearLibraryFailures(working.ID, 7)
	require.NoError(t, err)
	assert.Equal(t, LibraryFailureCleanup{DeletedAssets: 1}, *result)
	for publicID, want := range map[string]string{
		"waiting-for-resume": model.AssetStatusProcessing, // its upload stays paused on the other account (#89)
		"rejected-by-review": model.AssetStatusDeleted,    // review rejection makes it unusable everywhere
		"being-processed":    model.AssetStatusSyncFailed, // a leased copy is left to its worker
	} {
		var asset model.MediaAsset
		require.NoError(t, db.Where("public_id = ?", publicID).First(&asset).Error)
		assert.Equal(t, want, asset.Status, publicID)
	}
}

func TestDeletionStartedByAFailedUploadGetsAFullRetryBudget(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	now := common.GetTimestamp()
	replica := seedSyncReplicas(t, 7704, model.AssetReplica{Status: model.AssetReplicaStatusProcessing, UpstreamAssetID: "uploading", Attempts: 4, NextSyncAt: now - 1})[0]
	claimed, err := model.ClaimAssetReplica(&replica, now, now+120, "uploader")
	require.NoError(t, err)
	require.True(t, claimed)
	// The owner deletes the asset while the upload is still failing.
	require.NoError(t, db.Model(&model.MediaAsset{}).Where("id = ?", replica.AssetID).Update("status", model.AssetStatusDeleted).Error)
	require.NoError(t, model.FinishAssetReplica(replica.ID, "uploader", model.AssetReplicaStatusFailed, replica.Progress, replica.UpstreamAssetID, now+60, "upload failed", model.AssetReplicaOperationSync))
	require.NoError(t, db.First(&replica, replica.ID).Error)
	assert.Equal(t, model.AssetReplicaOperationDelete, replica.Operation)
	assert.Zero(t, replica.Attempts, "upload failures must not use up the deletion's retries")
}

func TestClearingSkipsCopiesRemovedMeanwhile(t *testing.T) {
	setupAssetLibraryTestDB(t)
	dropped, deleted, err := model.ClearAssetReplicaFailure(987654, 1, nil)
	require.NoError(t, err, "a copy that no longer exists is not an error")
	assert.False(t, dropped || deleted)
}
