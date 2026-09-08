package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	storageCleanupInterval  = time.Minute
	storageCleanupLease     = 2 * time.Minute
	storageCleanupBatch     = 50
	storageLedgerRetention  = 30 * 24 * time.Hour
	storageLedgerPurgeBatch = 500
)

var storageCleanupOnce sync.Once

func StartCleanupTask() {
	storageCleanupOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		runnerID := fmt.Sprintf("%s-storage-%s", common.NodeName, common.GetRandomString(8))
		gopool.Go(func() {
			runStorageCleanupPass(runnerID)
			ticker := time.NewTicker(storageCleanupInterval)
			defer ticker.Stop()
			for range ticker.C {
				runStorageCleanupPass(runnerID)
			}
		})
	})
}

func ScheduleTaskCleanup(taskID string, deleteAfter int64) {
	if err := model.MarkStorageObjectsForTaskDeletion(taskID, deleteAfter); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("schedule temporary storage cleanup failed: task_id=%s error=%v", taskID, err))
	}
}

func runStorageCleanupPass(runnerID string) {
	now := common.GetTimestamp()
	objects, err := model.ClaimStorageObjectsForDeletion(now, now+int64(storageCleanupLease/time.Second), runnerID, storageCleanupBatch)
	if err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("claim temporary storage cleanup failed: %v", err))
		return
	}
	for _, object := range objects {
		deleteStorageObject(runnerID, object)
	}
	if _, err = model.PurgeDeletedStorageObjects(now-int64(storageLedgerRetention/time.Second), storageLedgerPurgeBatch); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("purge deleted temporary storage ledgers failed: %v", err))
	}
}

func deleteStorageObject(runnerID string, object *model.StorageObject) {
	profile, err := model.GetStorageProfileByID(object.StorageProfileID)
	if err == nil && profile == nil {
		err = fmt.Errorf("storage profile %d does not exist", object.StorageProfileID)
	}
	var credential *model.StorageCredential
	if err == nil {
		credential, err = model.GetActiveStorageCredential(object.StorageProfileID)
		if err == nil && credential == nil {
			err = fmt.Errorf("storage credential for profile %d does not exist", object.StorageProfileID)
		}
	}
	var driver ObjectDriver
	if err == nil {
		driver, err = driverForProfile(profile, credential)
	}
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err = driver.DeleteObject(ctx, object.ObjectKey)
		cancel()
	}
	if err != nil {
		nextAttempt := common.GetTimestamp() + int64(model.StorageObjectDeletionBackoff(object.DeleteAttempts)/time.Second)
		if markErr := model.FailStorageObjectDeletion(object.ID, runnerID, nextAttempt, err.Error()); markErr != nil {
			logger.LogWarn(context.Background(), fmt.Sprintf("record temporary storage deletion failure: object_id=%s error=%v", object.ObjectID, markErr))
		}
		logger.LogWarn(context.Background(), fmt.Sprintf("temporary storage deletion failed: object_id=%s profile_id=%d error=%v", object.ObjectID, object.StorageProfileID, err))
		return
	}
	if err = model.FinishStorageObjectDeletion(object.ID, runnerID); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("finish temporary storage deletion failed: object_id=%s error=%v", object.ObjectID, err))
	}
}
