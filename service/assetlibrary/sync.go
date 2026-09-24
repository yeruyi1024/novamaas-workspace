package assetlibrary

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"

	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
)

const (
	assetSyncInterval = 10 * time.Second
	assetSyncLease    = 2 * time.Minute
	assetSyncBatch    = 20
)

var assetSyncOnce sync.Once

func StartSyncTask() {
	assetSyncOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		runnerID := fmt.Sprintf("%s-assets-%s", common.NodeName, common.GetRandomString(8))
		gopool.Go(func() {
			runSyncPass(runnerID)
			ticker := time.NewTicker(assetSyncInterval)
			defer ticker.Stop()
			for range ticker.C {
				runSyncPass(runnerID)
			}
		})
	})
}

func runSyncPass(runnerID string) {
	now := common.GetTimestamp()
	replicas, err := model.ClaimAssetReplicas(now, now+int64(assetSyncLease/time.Second), runnerID, assetSyncBatch)
	if err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("claim asset replicas failed: %v", err))
		return
	}
	for _, replica := range replicas {
		syncReplica(runnerID, replica)
	}
}

func syncReplica(runnerID string, replica *model.AssetReplica) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	ctx = withRequestLogContext(ctx, "sync", replica.ID, replica.AssetID)
	var asset model.MediaAsset
	if err := model.DB.First(&asset, replica.AssetID).Error; err != nil {
		failReplica(runnerID, replica, err)
		return
	}
	if asset.Status == model.AssetStatusUnavailable && replica.Operation != model.AssetReplicaOperationDelete {
		if err := model.FinishAssetReplica(replica.ID, runnerID, model.AssetReplicaStatusRejected, replica.Progress, replica.UpstreamAssetID, 0, asset.UnavailableReason); err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("finish unavailable asset replica failed: replica_id=%d error=%v", replica.ID, err))
		}
		return
	}
	var config model.AssetChannelConfig
	if err := model.DB.Where("channel_id = ?", replica.ChannelID).First(&config).Error; err != nil {
		failReplica(runnerID, replica, err)
		return
	}
	client, err := newProviderClient(&config)
	if err != nil {
		failReplica(runnerID, replica, err)
		return
	}
	if asset.Status == model.AssetStatusDeleted || replica.Operation == model.AssetReplicaOperationDelete {
		if replica.UpstreamAssetID != "" {
			if err = client.deleteAsset(ctx, replica.UpstreamAssetID); err != nil && !isProviderNotFound(err) {
				failReplica(runnerID, replica, err)
				return
			}
		}
		if err = model.FinishAssetReplica(replica.ID, runnerID, model.AssetReplicaStatusDeleted, 100, replica.UpstreamAssetID, 0, ""); err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("finish asset replica deletion failed: replica_id=%d error=%v", replica.ID, err))
			return
		}
		cleanupDeletedAssetStorage(&asset)
		return
	}
	if !config.Enabled {
		finishReplicaLater(runnerID, replica, model.AssetReplicaStatusFailed, replica.Progress, replica.UpstreamAssetID, 24*time.Hour, "asset channel is disabled")
		return
	}
	upstreamGroupID := ""
	if client.requiresGroup() {
		var group model.AssetGroup
		if err = model.DB.First(&group, asset.GroupID).Error; err != nil {
			failReplica(runnerID, replica, err)
			return
		}
		groupReplica, groupErr := ensureGroupReplica(ctx, client, &group, replica.ChannelID)
		if groupErr != nil {
			failReplica(runnerID, replica, groupErr)
			return
		}
		upstreamGroupID = groupReplica.UpstreamGroupID
	}
	if replica.UpstreamAssetID == "" {
		sourceURL, err := storageService.PresignAssetObject(ctx, asset.StorageObjectID, 24*time.Hour, storageService.AssetObjectURLOriginal, "")
		if err != nil {
			failReplica(runnerID, replica, err)
			return
		}
		upstreamID, err := client.createAsset(ctx, upstreamGroupID, asset.Name, asset.AssetType, sourceURL)
		if err != nil {
			if reason := upstreamContentRejectionReason(err); reason != "" {
				rejectReplica(runnerID, replica, reason)
				return
			}
			failReplica(runnerID, replica, err)
			return
		}
		finishReplicaLater(runnerID, replica, model.AssetReplicaStatusProcessing, 50, upstreamID, 10*time.Second, "")
		return
	}
	result, err := client.getAsset(ctx, replica.UpstreamAssetID)
	if err != nil {
		if reason := upstreamContentRejectionReason(err); reason != "" {
			rejectReplica(runnerID, replica, reason)
			return
		}
		failReplica(runnerID, replica, err)
		return
	}
	switch strings.ToLower(result.Status) {
	case "active", "ready", "succeeded", "success":
		if err = model.FinishAssetReplica(replica.ID, runnerID, model.AssetReplicaStatusActive, 100, replica.UpstreamAssetID, 0, ""); err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("finish active asset replica failed: replica_id=%d error=%v", replica.ID, err))
		}
	case "failed", "error", "rejected":
		if reason := processingContentRejectionReason(result); reason != "" {
			rejectReplica(runnerID, replica, reason)
			return
		}
		message := result.Message
		if message == "" {
			message = "upstream asset processing failed"
		}
		failReplica(runnerID, replica, errors.New(message))
	default:
		finishReplicaLater(runnerID, replica, model.AssetReplicaStatusProcessing, 75, replica.UpstreamAssetID, 15*time.Second, "")
	}
}

func rejectReplica(runnerID string, replica *model.AssetReplica, reason string) {
	if err := model.RejectAssetReplica(replica.ID, runnerID, reason, replica.UpstreamAssetID); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("record asset content rejection failed: replica_id=%d error=%v", replica.ID, err))
		return
	}
	logger.LogWarn(context.Background(), fmt.Sprintf("asset content rejected by upstream: replica_id=%d channel_id=%d reason=%s", replica.ID, replica.ChannelID, reason))
}

func ensureGroupReplica(ctx context.Context, client assetProvider, group *model.AssetGroup, channelID int) (*model.AssetGroupReplica, error) {
	var replica model.AssetGroupReplica
	err := model.DB.Where("group_id = ? AND channel_id = ?", group.ID, channelID).First(&replica).Error
	if err == nil && replica.UpstreamGroupID != "" && replica.Status == model.AssetReplicaStatusActive {
		return &replica, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	upstreamID, err := client.createGroup(ctx, group.Name, group.Description)
	if err != nil {
		return nil, err
	}
	now := common.GetTimestamp()
	replica.GroupID = group.ID
	replica.ChannelID = channelID
	replica.UpstreamGroupID = upstreamID
	replica.Status = model.AssetReplicaStatusActive
	replica.Progress = 100
	replica.LastError = ""
	replica.LastSyncedAt = now
	replica.UpdatedAt = now
	if replica.ID == 0 {
		err = model.DB.Create(&replica).Error
	} else {
		err = model.DB.Save(&replica).Error
	}
	return &replica, err
}

func failReplica(runnerID string, replica *model.AssetReplica, err error) {
	next := common.GetTimestamp() + int64(model.AssetReplicaBackoff(replica.Attempts)/time.Second)
	if finishErr := model.FinishAssetReplica(replica.ID, runnerID, model.AssetReplicaStatusFailed, replica.Progress, replica.UpstreamAssetID, next, truncateError(err)); finishErr != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("record asset replica failure: replica_id=%d error=%v", replica.ID, finishErr))
	}
	logger.LogWarn(context.Background(), fmt.Sprintf("asset replica synchronization failed: replica_id=%d channel_id=%d error=%v", replica.ID, replica.ChannelID, err))
}

func finishReplicaLater(runnerID string, replica *model.AssetReplica, status string, progress int, upstreamID string, delay time.Duration, message string) {
	next := common.GetTimestamp() + int64(delay/time.Second)
	if err := model.FinishAssetReplica(replica.ID, runnerID, status, progress, upstreamID, next, message); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("finish asset replica pass failed: replica_id=%d error=%v", replica.ID, err))
	}
}

func cleanupDeletedAssetStorage(asset *model.MediaAsset) {
	var remaining int64
	if err := model.DB.Model(&model.AssetReplica{}).Where("asset_id = ? AND status <> ?", asset.ID, model.AssetReplicaStatusDeleted).Count(&remaining).Error; err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("count deleting asset replicas failed: asset_id=%s error=%v", asset.PublicID, err))
		return
	}
	if remaining == 0 {
		if err := storageService.DeleteAssetObject(asset.StorageObjectID, "asset replicas deleted"); err != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("schedule asset object deletion failed: asset_id=%s error=%v", asset.PublicID, err))
		}
	}
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 2000 {
		message = message[:2000]
	}
	return message
}

func isProviderNotFound(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "notfound") || strings.Contains(message, "not found") || strings.Contains(message, "404")
}
