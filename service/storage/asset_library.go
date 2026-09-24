package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

func UploadAssetObject(ctx context.Context, ownerUserID int, groupPublicID string, assetPublicID string, fileName string, contentType string, size int64, body io.Reader) (*model.StorageObject, error) {
	if ownerUserID <= 0 || strings.TrimSpace(groupPublicID) == "" || strings.TrimSpace(assetPublicID) == "" {
		return nil, &RequestError{StatusCode: 400, Err: errors.New("asset storage identity is incomplete")}
	}
	policy, profile, credential, err := ResolveEnabledPolicy(model.StoragePolicyAssetLibrary)
	if err != nil {
		return nil, &RequestError{StatusCode: 503, Err: fmt.Errorf("asset library storage is unavailable: %w", err)}
	}
	contentType = strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	if !policyAllowsMIMEType(policy, contentType) {
		return nil, &RequestError{StatusCode: 400, Err: fmt.Errorf("asset MIME type %s is not allowed", contentType)}
	}
	if size <= 0 || size > policy.MaxFileBytes {
		return nil, &RequestError{StatusCode: 400, Err: fmt.Errorf("asset size must be between 1 and %d bytes", policy.MaxFileBytes)}
	}
	extension := assetExtension(contentType)
	if extension == "" {
		extension = strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	}
	if extension == "" || len(extension) > 10 {
		extension = "bin"
	}
	suffix, err := randomObjectSuffix()
	if err != nil {
		return nil, err
	}
	userHash := common.GenerateHMAC("asset-user:" + strconv.Itoa(ownerUserID))
	if len(userHash) > 24 {
		userHash = userHash[:24]
	}
	objectKey := fmt.Sprintf("%s/user-%s/%s/%s-%s.%s",
		strings.Trim(policy.ObjectPrefix, "/"), userHash, groupPublicID, assetPublicID, suffix, extension)
	if len(objectKey) > 1024 {
		return nil, errors.New("asset storage object key is too long")
	}
	object := &model.StorageObject{
		ObjectID:         "obj_" + suffix,
		OwnerUserID:      ownerUserID,
		StorageProfileID: profile.ID,
		StoragePolicyID:  policy.ID,
		Purpose:          model.StorageObjectPurposeAssetLibrary,
		ObjectKey:        objectKey,
		ContentType:      contentType,
		Size:             size,
		Status:           model.StorageObjectStatusUploading,
		DeleteAfter:      0,
	}
	if err = model.CreateStorageObject(object); err != nil {
		return nil, err
	}
	driver, err := driverForProfile(profile, credential)
	if err != nil {
		_ = model.MarkStorageObjectForDeletion(object.ID, common.GetTimestamp(), "initialize permanent asset storage failed")
		return nil, err
	}
	hash := sha256.New()
	etag, err := driver.PutObject(ctx, objectKey, contentType, size, io.TeeReader(body, hash))
	if err != nil {
		_ = model.MarkStorageObjectForDeletion(object.ID, common.GetTimestamp(), "permanent asset upload failed")
		return nil, &RequestError{StatusCode: 502, Err: fmt.Errorf("upload asset: %w", err)}
	}
	object.SHA256 = hex.EncodeToString(hash.Sum(nil))
	object.ETag = etag
	object.Status = model.StorageObjectStatusUploaded
	object.UpdatedAt = common.GetTimestamp()
	if err = model.MarkStorageObjectUploaded(object.ID, etag, object.SHA256); err != nil {
		_ = model.MarkStorageObjectForDeletion(object.ID, common.GetTimestamp(), "persist permanent asset upload state failed")
		return nil, err
	}
	return object, nil
}

type AssetObjectURLMode string

const (
	AssetObjectURLOriginal  AssetObjectURLMode = "original"
	AssetObjectURLThumbnail AssetObjectURLMode = "thumbnail"
	AssetObjectURLDownload  AssetObjectURLMode = "download"
)

func PresignAssetObject(ctx context.Context, objectID int64, ttl time.Duration, mode AssetObjectURLMode, downloadName string) (string, error) {
	var object model.StorageObject
	if err := model.DB.Where("id = ? AND purpose = ? AND status = ?", objectID, model.StorageObjectPurposeAssetLibrary, model.StorageObjectStatusUploaded).First(&object).Error; err != nil {
		return "", err
	}
	profile, err := model.GetStorageProfileByID(object.StorageProfileID)
	if err != nil || profile == nil {
		return "", fmt.Errorf("asset storage profile is unavailable")
	}
	credential, err := model.GetActiveStorageCredential(profile.ID)
	if err != nil || credential == nil {
		return "", fmt.Errorf("asset storage credential is unavailable")
	}
	driver, err := driverForProfile(profile, credential)
	if err != nil {
		return "", err
	}
	if ttl <= 0 {
		policy, policyErr := model.GetStoragePolicyByKey(model.StoragePolicyAssetLibrary)
		if policyErr != nil || policy == nil {
			return "", fmt.Errorf("asset storage policy is unavailable")
		}
		ttl = policySignedURLTTL(policy)
	}
	options := ObjectGetOptions{}
	switch mode {
	case AssetObjectURLOriginal:
	case AssetObjectURLThumbnail:
		if strings.HasPrefix(object.ContentType, "image/") {
			options.Process = "image/resize,m_lfit,w_640,h_360"
			if object.ContentType == "image/jpeg" || object.ContentType == "image/webp" {
				options.Process += "/quality,q_80"
			}
		}
	case AssetObjectURLDownload:
		if downloadName == "" || strings.ContainsAny(downloadName, "\"\\\r\n/") {
			return "", errors.New("invalid asset download name")
		}
		extension := assetExtension(object.ContentType)
		if extension == "" {
			extension = "bin"
		}
		options.ResponseContentDisposition = fmt.Sprintf("attachment; filename=\"%s.%s\"", downloadName, extension)
	default:
		return "", errors.New("invalid asset object URL mode")
	}
	return driver.PresignGet(ctx, object.ObjectKey, ttl, options)
}

func DeleteAssetObject(objectID int64, reason string) error {
	return model.MarkStorageObjectForDeletion(objectID, common.GetTimestamp(), reason)
}

func policyAllowsMIMEType(policy *model.StoragePolicy, contentType string) bool {
	for _, candidate := range strings.Split(policy.AllowedMIMETypes, ",") {
		if strings.TrimSpace(candidate) == contentType {
			return true
		}
	}
	return false
}

func assetExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	case "video/quicktime":
		return "mov"
	case "audio/mpeg":
		return "mp3"
	case "audio/wav", "audio/x-wav":
		return "wav"
	case "audio/mp4":
		return "m4a"
	case "audio/aac":
		return "aac"
	case "audio/ogg":
		return "ogg"
	default:
		return ""
	}
}
