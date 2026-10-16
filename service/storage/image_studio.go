package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// UploadImageStudioObject uses the asset library's storage connection for a
// generated image, with a separate 30-day lifetime and no library asset.
func UploadImageStudioObject(ctx context.Context, ownerUserID int, runID string, resultID string, contentType string, data []byte) (*model.StorageObject, error) {
	policy, profile, credential, err := ResolveEnabledPolicy(model.StoragePolicyAssetLibrary)
	if err != nil {
		return nil, &RequestError{StatusCode: 503, Err: fmt.Errorf("permanent storage is unavailable: %w", err)}
	}
	extension := assetExtension(contentType)
	if !strings.HasPrefix(contentType, "image/") || extension == "" || !policyAllowsMIMEType(policy, contentType) {
		return nil, &RequestError{StatusCode: 400, Err: fmt.Errorf("image type %s is not allowed by permanent storage", contentType)}
	}
	size := int64(len(data))
	if size <= 0 || size > policy.MaxFileBytes {
		return nil, &RequestError{StatusCode: 400, Err: fmt.Errorf("image size must be between 1 and %d bytes", policy.MaxFileBytes)}
	}
	suffix, err := randomObjectSuffix()
	if err != nil {
		return nil, err
	}
	objectKey := fmt.Sprintf("%s/user-%s/image-studio/%s/%s-%s.%s",
		strings.Trim(policy.ObjectPrefix, "/"), permanentStorageUserHash(ownerUserID), runID, resultID, suffix, extension)
	object := &model.StorageObject{
		ObjectID:         "obj_" + suffix,
		OwnerUserID:      ownerUserID,
		StorageProfileID: profile.ID,
		StoragePolicyID:  policy.ID,
		Purpose:          model.StorageObjectPurposeImageStudio,
		ObjectKey:        objectKey,
		ContentType:      contentType,
		Size:             size,
		Status:           model.StorageObjectStatusUploading,
	}
	if err = model.CreateStorageObject(object); err != nil {
		return nil, err
	}
	driver, err := driverForProfile(profile, credential)
	if err != nil {
		_ = model.MarkStorageObjectForDeletion(object.ID, common.GetTimestamp(), "initialize image studio storage failed")
		return nil, err
	}
	etag, err := driver.PutObject(ctx, objectKey, contentType, size, bytes.NewReader(data))
	if err != nil {
		_ = model.MarkStorageObjectForDeletion(object.ID, common.GetTimestamp(), "image studio upload failed")
		return nil, &RequestError{StatusCode: 502, Err: fmt.Errorf("upload image: %w", err)}
	}
	digest := sha256.Sum256(data)
	if err = model.MarkStorageObjectUploaded(object.ID, etag, hex.EncodeToString(digest[:])); err != nil {
		_ = model.MarkStorageObjectForDeletion(object.ID, common.GetTimestamp(), "persist image studio upload state failed")
		return nil, err
	}
	object.Status = model.StorageObjectStatusUploaded
	return object, nil
}

// PresignImageStudioObject signs a generated image owned by ownerUserID.
// Download URLs name the file after downloadName with the extension of the
// stored content type, matching asset library downloads.
func PresignImageStudioObject(ctx context.Context, objectID int64, ownerUserID int, mode AssetObjectURLMode, downloadName string) (string, error) {
	var object model.StorageObject
	if err := model.DB.Where("id = ? AND owner_user_id = ? AND purpose = ? AND status = ?",
		objectID, ownerUserID, model.StorageObjectPurposeImageStudio, model.StorageObjectStatusUploaded).
		Where("created_at > ?", common.GetTimestamp()-model.ImageStudioRetentionSeconds).First(&object).Error; err != nil {
		return "", err
	}
	expiresAt := object.CreatedAt + model.ImageStudioRetentionSeconds
	if object.DeleteAfter == 0 || object.DeleteAfter > expiresAt {
		object.DeleteAfter = expiresAt
	}
	if mode == AssetObjectURLDownload && assetExtension(object.ContentType) == "" {
		return "", errors.New("invalid image download type")
	}
	return presignPermanentObject(ctx, &object, 0, mode, downloadName)
}
