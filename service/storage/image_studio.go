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

func GetImageStudioPolicy(purpose string) (*model.StoragePolicy, error) {
	if purpose != model.StorageObjectPurposeImageStudio && purpose != model.StorageObjectPurposeImageStudioReference {
		return nil, &RequestError{StatusCode: 400, Err: errors.New("invalid image studio storage purpose")}
	}
	policy, err := model.GetStoragePolicyByKey(purpose)
	if err == nil && policy == nil {
		policy = model.DefaultImageStudioStoragePolicy(purpose)
	}
	return policy, err
}

func SaveImageStudioPolicy(purpose string, input PolicyInput) (*model.StoragePolicy, error) {
	policy, err := GetImageStudioPolicy(purpose)
	if err != nil {
		return nil, err
	}
	policy.StorageProfileID, policy.Enabled = input.StorageProfileID, input.Enabled
	policy.ObjectPrefix = strings.Trim(strings.TrimSpace(input.ObjectPrefix), "/")
	policy.SignedURLTTLSeconds, policy.RetentionSeconds = input.SignedURLTTLSeconds, input.RetentionSeconds
	policy.MaxFileBytes, policy.MaxTotalBytes = input.MaxFileBytes, input.MaxFileBytes
	if input.RetentionSeconds < 3600 || input.MaxFileBytes > model.DefaultImageStudioStoragePolicy(purpose).MaxFileBytes {
		return nil, &RequestError{StatusCode: 400, Err: errors.New("retention must be at least one hour and file size must not exceed the image studio limit")}
	}
	if err := validatePolicy(policy); err != nil {
		return nil, &RequestError{StatusCode: 400, Err: err}
	}
	return policy, model.SaveStoragePolicy(policy)
}

// Generated images and references have independent business storage policies.
func UploadImageStudioObject(ctx context.Context, ownerUserID int, runID string, resultID string, contentType string, data []byte) (*model.StorageObject, error) {
	return uploadImageStudioObject(ctx, ownerUserID, runID, resultID, model.StorageObjectPurposeImageStudio, contentType, data)
}

func UploadImageStudioReference(ctx context.Context, ownerUserID int, contentType string, data []byte) (*model.StorageObject, error) {
	return uploadImageStudioObject(ctx, ownerUserID, "references", "reference", model.StorageObjectPurposeImageStudioReference, contentType, data)
}

func uploadImageStudioObject(ctx context.Context, ownerUserID int, runID, resultID, purpose, contentType string, data []byte) (*model.StorageObject, error) {
	policy, profile, credential, err := ResolveEnabledPolicy(purpose)
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
		Purpose:          purpose,
		ObjectKey:        objectKey,
		ContentType:      contentType,
		Size:             size,
		Status:           model.StorageObjectStatusUploading,
		DeleteAfter:      common.GetTimestamp() + policy.RetentionSeconds,
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

// PresignImageStudioReference only resolves live temporary images owned by the
// requesting user. Generation obtains a fresh URL rather than trusting a client URL.
func PresignImageStudioReference(ctx context.Context, objectID string, ownerUserID int) (string, error) {
	var object model.StorageObject
	if err := model.DB.Where("object_id = ? AND owner_user_id = ? AND purpose = ? AND status = ? AND delete_after > ?",
		objectID, ownerUserID, model.StorageObjectPurposeImageStudioReference, model.StorageObjectStatusUploaded, common.GetTimestamp()).First(&object).Error; err != nil {
		return "", err
	}
	return presignPermanentObject(ctx, &object, 0, AssetObjectURLOriginal, "")
}

// PresignImageStudioObject signs a generated image owned by ownerUserID.
// Download URLs name the file after downloadName with the extension of the
// stored content type, matching asset library downloads.
func PresignImageStudioObject(ctx context.Context, objectID int64, ownerUserID int, mode AssetObjectURLMode, downloadName string) (string, error) {
	var object model.StorageObject
	if err := model.DB.Where("id = ? AND owner_user_id = ? AND purpose = ? AND status = ?",
		objectID, ownerUserID, model.StorageObjectPurposeImageStudio, model.StorageObjectStatusUploaded).First(&object).Error; err != nil {
		return "", err
	}
	if mode == AssetObjectURLDownload && assetExtension(object.ContentType) == "" {
		return "", errors.New("invalid image download type")
	}
	return presignPermanentObject(ctx, &object, 0, mode, downloadName)
}
