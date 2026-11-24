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
	"sync"
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
	if contentType == "audio/wave" {
		contentType = "audio/wav"
	}
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
	userHash := permanentStorageUserHash(ownerUserID)
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
	AssetObjectURLOriginal    AssetObjectURLMode = "original"
	AssetObjectURLThumbnail   AssetObjectURLMode = "thumbnail"
	AssetObjectURLVideoPoster AssetObjectURLMode = "video_poster"
	AssetObjectURLDownload    AssetObjectURLMode = "download"

	assetThumbnailCacheControl = "private, max-age=1800, immutable"
	maxThumbnailCacheWindowSec = 1800
	maxCachedThumbnailURLs     = 10000
)

type PresignedThumbnail struct {
	URL       string
	PosterURL string
	ExpiresAt int64
}

type signedThumbnailCacheKey struct {
	endpoint         string
	bucket           string
	objectKey        string
	credentialID     int64
	credentialVer    int
	credentialDigest [32]byte
	encryptionDigest [32]byte
	mode             AssetObjectURLMode
	process          string
	cacheControl     string
}

type signedThumbnailCacheEntry struct {
	url          string
	expiresAt    int64
	refreshAfter int64
}

type assetDriverIdentity struct {
	provider         string
	profileID        int
	endpoint         string
	region           string
	bucket           string
	credentialID     int64
	credentialVer    int
	credentialDigest [32]byte
	encryptionDigest [32]byte
}

const maxCachedAssetDrivers = 64

var (
	thumbnailCacheMu sync.RWMutex
	thumbnailCache   = make(map[signedThumbnailCacheKey]signedThumbnailCacheEntry)

	assetDriverCacheMu sync.RWMutex
	assetDriverCache   = make(map[assetDriverIdentity]ObjectDriver)
)

func assetEncryptionDigest(credential *model.StorageCredential) ([32]byte, error) {
	if credential.AuthType != model.StorageCredentialAuthStatic {
		return [32]byte{}, nil
	}
	key, err := storageCredentialEncryptionKey()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(key), nil
}

func cachedAssetDriver(profile *model.StorageProfile, credential *model.StorageCredential) (ObjectDriver, error) {
	if credential.AuthType != model.StorageCredentialAuthStatic {
		return driverForProfile(profile, credential)
	}
	encDigest, err := assetEncryptionDigest(credential)
	if err != nil {
		return nil, err
	}
	id := assetDriverIdentity{
		provider:         profile.ProviderType,
		profileID:        profile.ID,
		endpoint:         profile.Endpoint,
		region:           profile.Region,
		bucket:           profile.Bucket,
		credentialID:     credential.ID,
		credentialVer:    credential.Version,
		credentialDigest: sha256.Sum256([]byte(credential.EncryptedPayload)),
		encryptionDigest: encDigest,
	}
	assetDriverCacheMu.RLock()
	cached, ok := assetDriverCache[id]
	assetDriverCacheMu.RUnlock()
	if ok {
		return cached, nil
	}
	driver, err := driverForProfile(profile, credential)
	if err != nil {
		return nil, err
	}
	assetDriverCacheMu.Lock()
	if len(assetDriverCache) >= maxCachedAssetDrivers {
		assetDriverCache = make(map[assetDriverIdentity]ObjectDriver)
	}
	assetDriverCache[id] = driver
	assetDriverCacheMu.Unlock()
	return driver, nil
}

func presignAssetCached(ctx context.Context, driver ObjectDriver, profile *model.StorageProfile, credential *model.StorageCredential, object *model.StorageObject, ttl time.Duration, mode AssetObjectURLMode, options ObjectGetOptions) (string, int64, error) {
	now := common.GetTimestamp()
	ttlSec := int64(ttl / time.Second)
	if ttlSec <= 0 {
		ttlSec = 3600
		ttl = time.Hour
	}
	cacheable := (mode == AssetObjectURLThumbnail || mode == AssetObjectURLVideoPoster) &&
		credential.AuthType == model.StorageCredentialAuthStatic
	var key signedThumbnailCacheKey
	if cacheable {
		encDigest, err := assetEncryptionDigest(credential)
		if err != nil {
			return "", 0, err
		}
		key = signedThumbnailCacheKey{
			endpoint:         profile.Endpoint,
			bucket:           profile.Bucket,
			objectKey:        object.ObjectKey,
			credentialID:     credential.ID,
			credentialVer:    credential.Version,
			credentialDigest: sha256.Sum256([]byte(credential.EncryptedPayload)),
			encryptionDigest: encDigest,
			mode:             mode,
			process:          options.Process,
			cacheControl:     options.ResponseCacheControl,
		}
		thumbnailCacheMu.RLock()
		entry, ok := thumbnailCache[key]
		thumbnailCacheMu.RUnlock()
		if ok && now < entry.refreshAfter && now < entry.expiresAt {
			return entry.url, entry.expiresAt, nil
		}
	}
	signed, err := driver.PresignGet(ctx, object.ObjectKey, ttl, options)
	if err != nil {
		return "", 0, err
	}
	expiresAt := now + ttlSec
	if cacheable {
		window := ttlSec / 2
		if window > maxThumbnailCacheWindowSec {
			window = maxThumbnailCacheWindowSec
		}
		if window > 0 {
			thumbnailCacheMu.Lock()
			if len(thumbnailCache) >= maxCachedThumbnailURLs {
				thumbnailCache = make(map[signedThumbnailCacheKey]signedThumbnailCacheEntry)
			}
			thumbnailCache[key] = signedThumbnailCacheEntry{
				url:          signed,
				expiresAt:    expiresAt,
				refreshAfter: now + window,
			}
			thumbnailCacheMu.Unlock()
		}
	}
	return signed, expiresAt, nil
}

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
	driver, err := cachedAssetDriver(profile, credential)
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
	options, err := assetObjectGetOptions(&object, mode, downloadName)
	if err != nil {
		return "", err
	}
	signed, _, err := presignAssetCached(ctx, driver, profile, credential, &object, ttl, mode, options)
	return signed, err
}

// PresignAssetThumbnailsWithExpiry signs thumbnail (and video poster) URLs for
// many asset objects with one object query, cached drivers, and process-local
// deterministic URL caching so browsers hit disk cache across reloads/filters.
func PresignAssetThumbnailsWithExpiry(ctx context.Context, objectIDs []int64, ttl time.Duration) (map[int64]PresignedThumbnail, error) {
	results := make(map[int64]PresignedThumbnail, len(objectIDs))
	if len(objectIDs) == 0 {
		return results, nil
	}
	if ttl <= 0 {
		policy, policyErr := model.GetStoragePolicyByKey(model.StoragePolicyAssetLibrary)
		if policyErr != nil || policy == nil {
			return nil, fmt.Errorf("asset storage policy is unavailable")
		}
		ttl = policySignedURLTTL(policy)
	}
	var objects []model.StorageObject
	if err := model.DB.Where("id IN ? AND purpose = ? AND status = ?", objectIDs, model.StorageObjectPurposeAssetLibrary, model.StorageObjectStatusUploaded).
		Find(&objects).Error; err != nil {
		return nil, err
	}
	type profileBundle struct {
		profile    *model.StorageProfile
		credential *model.StorageCredential
		driver     ObjectDriver
	}
	bundles := make(map[int]profileBundle)
	failedProfiles := make(map[int]bool)
	for index := range objects {
		object := &objects[index]
		if failedProfiles[object.StorageProfileID] {
			continue
		}
		bundle, ok := bundles[object.StorageProfileID]
		if !ok {
			profile, err := model.GetStorageProfileByID(object.StorageProfileID)
			if err != nil || profile == nil {
				failedProfiles[object.StorageProfileID] = true
				common.SysError(fmt.Sprintf("asset thumbnail skipped: storage profile %d unavailable", object.StorageProfileID))
				continue
			}
			credential, err := model.GetActiveStorageCredential(profile.ID)
			if err != nil || credential == nil {
				failedProfiles[object.StorageProfileID] = true
				common.SysError(fmt.Sprintf("asset thumbnail skipped: storage credential for profile %d unavailable", profile.ID))
				continue
			}
			driver, err := cachedAssetDriver(profile, credential)
			if err != nil {
				failedProfiles[object.StorageProfileID] = true
				common.SysError(fmt.Sprintf("asset thumbnail skipped: storage driver for profile %d failed: %v", profile.ID, err))
				continue
			}
			bundle = profileBundle{profile: profile, credential: credential, driver: driver}
			bundles[object.StorageProfileID] = bundle
		}
		options, err := assetObjectGetOptions(object, AssetObjectURLThumbnail, "")
		if err != nil {
			continue
		}
		signed, expiresAt, err := presignAssetCached(ctx, bundle.driver, bundle.profile, bundle.credential, object, ttl, AssetObjectURLThumbnail, options)
		if err != nil {
			common.SysError(fmt.Sprintf("asset thumbnail presign skipped for object %d: %v", object.ID, err))
			continue
		}
		item := PresignedThumbnail{URL: signed, ExpiresAt: expiresAt}
		if strings.HasPrefix(object.ContentType, "video/") {
			posterOpts, err := assetObjectGetOptions(object, AssetObjectURLVideoPoster, "")
			if err == nil && posterOpts.Process != "" {
				if posterURL, posterExp, posterErr := presignAssetCached(ctx, bundle.driver, bundle.profile, bundle.credential, object, ttl, AssetObjectURLVideoPoster, posterOpts); posterErr == nil {
					item.PosterURL = posterURL
					if posterExp < item.ExpiresAt {
						item.ExpiresAt = posterExp
					}
				}
			}
		}
		results[object.ID] = item
	}
	return results, nil
}

func assetObjectGetOptions(object *model.StorageObject, mode AssetObjectURLMode, downloadName string) (ObjectGetOptions, error) {
	options := ObjectGetOptions{}
	switch mode {
	case AssetObjectURLOriginal:
	case AssetObjectURLThumbnail:
		options.ResponseCacheControl = assetThumbnailCacheControl
		if strings.HasPrefix(object.ContentType, "image/") {
			options.Process = "image/resize,m_lfit,w_512,h_384/auto-orient,1/format,webp/quality,q_75"
		}
	case AssetObjectURLVideoPoster:
		options.ResponseCacheControl = assetThumbnailCacheControl
		if strings.HasPrefix(object.ContentType, "video/") {
			options.Process = "video/snapshot,t_100,f_jpg,w_512,h_384,m_fast"
		}
	case AssetObjectURLDownload:
		if downloadName == "" || strings.ContainsAny(downloadName, "\"\\\r\n/") {
			return options, errors.New("invalid asset download name")
		}
		extension := assetExtension(object.ContentType)
		if extension == "" {
			extension = "bin"
		}
		options.ResponseContentDisposition = fmt.Sprintf("attachment; filename=\"%s.%s\"", downloadName, extension)
	default:
		return options, errors.New("invalid asset object URL mode")
	}
	return options, nil
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

// Permanent assets and billing evidence use the same opaque account directory.
// Keep the asset-user namespace to preserve existing asset paths.
func permanentStorageUserHash(userID int) string {
	return common.GenerateHMAC("asset-user:" + strconv.Itoa(userID))[:24]
}
