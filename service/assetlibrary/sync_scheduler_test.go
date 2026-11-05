package assetlibrary

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type dueReplicaSeed struct {
	label      string
	status     string
	upstreamID string
}

func seedDueReplicas(t *testing.T, db *gorm.DB, channelID int, seeds []dueReplicaSeed) map[int64]string {
	t.Helper()
	group := model.AssetGroup{PublicID: fmt.Sprintf("group-%d", channelID), OwnerUserID: 42, Status: model.AssetStatusReady}
	require.NoError(t, db.Create(&group).Error)
	due := common.GetTimestamp() - 10
	labels := make(map[int64]string, len(seeds))
	for index, seed := range seeds {
		asset := model.MediaAsset{
			PublicID: fmt.Sprintf("asset-%d-%d", channelID, index), OwnerUserID: 42, GroupID: group.ID,
			StorageObjectID: int64(channelID*100 + index + 1), AssetType: model.AssetTypeImage, Status: model.AssetStatusProcessing,
		}
		require.NoError(t, db.Create(&asset).Error)
		replica := model.AssetReplica{
			AssetID: asset.ID, ChannelID: channelID, Operation: model.AssetReplicaOperationSync,
			Status: seed.status, UpstreamAssetID: seed.upstreamID, NextSyncAt: due,
		}
		require.NoError(t, db.Create(&replica).Error)
		labels[replica.ID] = seed.label
	}
	return labels
}

func TestAssetSchedulerRotatesWorkloadsAndLeasesOnlyTheDispatchedReplica(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	labels := seedDueReplicas(t, db, 7101, []dueReplicaSeed{
		{"submit-1", model.AssetReplicaStatusPending, ""},
		{"submit-2", model.AssetReplicaStatusPending, ""},
		{"poll-1", model.AssetReplicaStatusProcessing, "upstream-1"},
		{"poll-2", model.AssetReplicaStatusProcessing, "upstream-2"},
		{"poll-3", model.AssetReplicaStatusProcessing, "upstream-3"},
		{"review-1", model.AssetReplicaStatusActive, "upstream-4"},
	})
	scheduler := newAssetSyncScheduler("scheduler-test")

	first, err := scheduler.claimNext(7101)
	require.NoError(t, err)
	require.NotNil(t, first)
	var leased int64
	require.NoError(t, db.Model(&model.AssetReplica{}).Where("locked_by <> ?", "").Count(&leased).Error)
	assert.EqualValues(t, 1, leased, "waiting work must stay unleased in the database")

	order := []string{labels[first.ID]}
	for {
		replica, err := scheduler.claimNext(7101)
		require.NoError(t, err)
		if replica == nil {
			break
		}
		order = append(order, labels[replica.ID])
	}
	// Polls take two turns; once a kind runs out, its turns go to the others.
	assert.Equal(t, []string{"submit-1", "poll-1", "poll-2", "review-1", "submit-2", "poll-3"}, order)
}

func TestAssetSchedulerGivesIdleTurnsToWaitingWork(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	labels := seedDueReplicas(t, db, 7102, []dueReplicaSeed{
		{"submit-1", model.AssetReplicaStatusPending, ""},
		{"submit-2", model.AssetReplicaStatusPending, ""},
		{"submit-3", model.AssetReplicaStatusPending, ""},
	})
	scheduler := newAssetSyncScheduler("scheduler-test")

	var order []string
	for range 3 {
		replica, err := scheduler.claimNext(7102)
		require.NoError(t, err)
		require.NotNil(t, replica)
		order = append(order, labels[replica.ID])
	}
	assert.Equal(t, []string{"submit-1", "submit-2", "submit-3"}, order)
}

func TestAssetSchedulerLeavesWorkUnleasedWhileChannelSlotIsTaken(t *testing.T) {
	db := setupAssetLibraryTestDB(t)
	const channelID = 7103
	require.NoError(t, db.Create(&model.AssetChannelConfig{ChannelID: channelID, QPM: 1}).Error)
	seedDueReplicas(t, db, channelID, []dueReplicaSeed{{"submit-1", model.AssetReplicaStatusPending, ""}})
	reserveChannelRate(channelID, 1)
	t.Cleanup(func() {
		channelRateState.Lock()
		delete(channelRateState.next, channelID)
		channelRateState.Unlock()
	})

	wait := newAssetSyncScheduler("scheduler-test").dispatch()

	assert.Greater(t, wait, time.Duration(0))
	assert.LessOrEqual(t, wait, assetSyncScanInterval)
	var replica model.AssetReplica
	require.NoError(t, db.Where("channel_id = ?", channelID).First(&replica).Error)
	assert.Empty(t, replica.LockedBy)
	assert.Equal(t, model.AssetReplicaStatusPending, replica.Status)
}

func TestDispatchedJobUsesItsReservedRequestSlotOnce(t *testing.T) {
	const channelID = 7104
	t.Cleanup(func() {
		channelRateState.Lock()
		delete(channelRateState.next, channelID)
		channelRateState.Unlock()
	})
	reserveChannelRate(channelID, 1)
	ctx := withChannelRateReservation(context.Background(), channelID)

	require.NoError(t, waitForChannelRate(ctx, channelID, 1), "the first call uses the slot taken at dispatch")

	// A further call in the same job needs a new slot, one minute later at QPM 1.
	waitCtx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	assert.ErrorIs(t, waitForChannelRate(waitCtx, channelID, 1), context.DeadlineExceeded)
}

func TestProcessingPollDelayBacksOffAsUploadAges(t *testing.T) {
	const now int64 = 10_000
	tests := []struct {
		name      string
		createdAt int64
		want      time.Duration
	}{
		{"first minute", now - 10, 3 * time.Second},
		{"second minute", now - 90, 5 * time.Second},
		{"within five minutes", now - 200, 10 * time.Second},
		{"older", now - 600, 15 * time.Second},
		{"unknown age", 0, 15 * time.Second},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, processingPollDelay(&model.AssetReplica{CreatedAt: test.createdAt}, now))
		})
	}
}

func TestAssetSyncLeadershipAllowsOneDispatcherAndFailsOver(t *testing.T) {
	server := miniredis.RunT(t)
	previousClient, previousEnabled := common.RDB, common.RedisEnabled
	common.RDB = redis.NewClient(&redis.Options{Addr: server.Addr()})
	common.RedisEnabled = true
	t.Cleanup(func() {
		_ = common.RDB.Close()
		common.RDB, common.RedisEnabled = previousClient, previousEnabled
	})
	ctx := context.Background()

	held, err := holdsAssetSyncLeadership(ctx, "node-a")
	require.NoError(t, err)
	assert.True(t, held, "the first node takes the lock")
	held, err = holdsAssetSyncLeadership(ctx, "node-b")
	require.NoError(t, err)
	assert.False(t, held, "a second node waits while the leader holds the lock")

	server.FastForward(assetSyncLeaderTTL - time.Second)
	held, err = holdsAssetSyncLeadership(ctx, "node-a")
	require.NoError(t, err)
	assert.True(t, held, "the leader renews before its lock expires")
	server.FastForward(assetSyncLeaderTTL - time.Second)
	held, err = holdsAssetSyncLeadership(ctx, "node-b")
	require.NoError(t, err)
	assert.False(t, held, "renewal keeps the lock with the leader")

	server.FastForward(2 * time.Second)
	held, err = holdsAssetSyncLeadership(ctx, "node-b")
	require.NoError(t, err)
	assert.True(t, held, "another node takes over once the leader stops renewing")
	held, err = holdsAssetSyncLeadership(ctx, "node-a")
	require.NoError(t, err)
	assert.False(t, held, "the former leader cannot renew a lock it no longer owns")
}

func TestAssetSyncWithoutRedisDispatchesLocally(t *testing.T) {
	previousEnabled := common.RedisEnabled
	common.RedisEnabled = false
	t.Cleanup(func() { common.RedisEnabled = previousEnabled })

	held, err := holdsAssetSyncLeadership(context.Background(), "node-a")
	require.NoError(t, err)
	assert.True(t, held)
}
