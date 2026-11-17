package assetlibrary

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

type AssetCopyInput struct {
	RequestID    string   `json:"request_id"`
	SourceUserID int      `json:"source_user_id"`
	TargetUserID int      `json:"target_user_id"`
	GroupIDs     []string `json:"group_ids"`
}

type AssetCopyGroupView struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	AssetCount int    `json:"asset_count"`
}

type AssetCopyResult struct {
	ID               string               `json:"id"`
	SourceUserID     int                  `json:"source_user_id"`
	TargetUserID     int                  `json:"target_user_id"`
	ReadyCount       int                  `json:"ready_count"`
	ProcessingCount  int                  `json:"processing_count"`
	FailedCount      int                  `json:"failed_count"`
	UnavailableCount int                  `json:"unavailable_count"`
	Groups           []AssetCopyGroupView `json:"groups"`
	AssetCount       int                  `json:"asset_count"`
	SourcesDeletedAt int64                `json:"sources_deleted_at"`
}

// A retried request returns its committed batch before any current-state
// check, so a lost response is answered even if the target changed since.
func committedAssetCopy(batch *model.AssetCopyBatch) (*model.AssetCopyBatch, error) {
	var existing model.AssetCopyBatch
	err := model.DB.First(&existing, "id = ?", batch.ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if existing.OperatorUserID != batch.OperatorUserID || existing.RequestHash != batch.RequestHash {
		return nil, &RequestError{StatusCode: http.StatusConflict, Err: errors.New("request ID was already used for a different copy")}
	}
	return &existing, nil
}

// CopyAssetGroups gives the target independent assets that share only the
// source files. Copies synchronize like fresh uploads; sources are untouched.
func CopyAssetGroups(operatorID int, input AssetCopyInput) (*AssetCopyResult, error) {
	input.RequestID = strings.TrimSpace(input.RequestID)
	if len(input.RequestID) < 16 || len(input.RequestID) > 64 || input.SourceUserID <= 0 || input.TargetUserID <= 0 || input.SourceUserID == input.TargetUserID || len(input.GroupIDs) == 0 || len(input.GroupIDs) > 100 {
		return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid copy request; select different users and 1–100 groups")}
	}
	input.GroupIDs = append([]string(nil), input.GroupIDs...)
	sort.Strings(input.GroupIDs)
	for index, id := range input.GroupIDs {
		if id == "" || len(id) > 64 || (index > 0 && input.GroupIDs[index-1] == id) {
			return nil, &RequestError{StatusCode: http.StatusBadRequest, Err: errors.New("invalid or duplicate group ID")}
		}
	}
	encoded, err := common.Marshal(input)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(encoded)
	batch := model.AssetCopyBatch{ID: input.RequestID, OperatorUserID: operatorID, SourceUserID: input.SourceUserID, TargetUserID: input.TargetUserID, RequestHash: hex.EncodeToString(digest[:])}
	if existing, err := committedAssetCopy(&batch); err != nil || existing != nil {
		if err != nil {
			return nil, err
		}
		return assetCopyResult(existing)
	}
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := model.CheckAssetCopyUsers(tx, input.SourceUserID, input.TargetUserID, operatorID); err != nil {
			return err
		}
		channelIDs, err := enabledChannelIDs(tx)
		if err != nil {
			return err
		}
		copies, err := model.CopyAssetGroupsWithTx(tx, &batch, input.GroupIDs, channelIDs)
		if err != nil {
			return err
		}
		for index := range copies {
			if copies[index].Status != model.AssetStatusReady {
				continue
			}
			if err := queueAssetWebhookDeliveries(tx, &copies[index], WebhookEventAssetActive); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		// A concurrent replay loses on the batch primary key after waiting for
		// the same source group locks; answer it with the committed batch.
		if existing, lookupErr := committedAssetCopy(&batch); lookupErr != nil || existing != nil {
			if lookupErr != nil {
				return nil, lookupErr
			}
			return assetCopyResult(existing)
		}
		return nil, err
	}
	return assetCopyResult(&batch)
}

func GetAssetCopy(batchID string) (*AssetCopyResult, error) {
	var batch model.AssetCopyBatch
	if err := model.DB.First(&batch, "id = ?", batchID).Error; err != nil {
		return nil, err
	}
	return assetCopyResult(&batch)
}

func DeleteCopiedSources(batchID string, operatorID int) (*AssetCopyResult, error) {
	batch, err := model.DeleteAssetCopySources(batchID, operatorID)
	if err != nil {
		return nil, err
	}
	return assetCopyResult(batch)
}

func assetCopyResult(batch *model.AssetCopyBatch) (*AssetCopyResult, error) {
	var manifest []model.AssetCopiedGroup
	if err := common.UnmarshalJsonStr(batch.Manifest, &manifest); err != nil {
		return nil, err
	}
	result := &AssetCopyResult{ID: batch.ID, SourceUserID: batch.SourceUserID, TargetUserID: batch.TargetUserID, SourcesDeletedAt: batch.SourcesDeletedAt, Groups: []AssetCopyGroupView{}}
	var targetIDs []int64
	for _, group := range manifest {
		for _, pair := range group.Assets {
			targetIDs = append(targetIDs, pair.TargetID)
		}
		result.Groups = append(result.Groups, AssetCopyGroupView{ID: group.TargetPublicID, Name: group.Name, AssetCount: len(group.Assets)})
		result.AssetCount += len(group.Assets)
	}
	var copies []model.MediaAsset
	if len(targetIDs) > 0 {
		if err := model.DB.Select("status").Where("id IN ? AND owner_user_id = ?", targetIDs, batch.TargetUserID).Find(&copies).Error; err != nil {
			return nil, err
		}
	}
	// Rejected, deleted and missing copies all count as unavailable.
	result.UnavailableCount = result.AssetCount
	for _, copy := range copies {
		switch copy.Status {
		case model.AssetStatusReady:
			result.ReadyCount++
		case model.AssetStatusProcessing:
			result.ProcessingCount++
		case model.AssetStatusSyncFailed:
			result.FailedCount++
		default:
			continue
		}
		result.UnavailableCount--
	}
	return result, nil
}
