package storage

import (
	"context"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAssetSignedURLsKeepThumbnailAndDownloadOptionsInsideOSSSignature(t *testing.T) {
	setupStorageDatabase(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "asset-preview-test-encryption-key")
	profile, err := SaveProfile(0, ProfileInput{
		Name: "Asset OSS", ProviderType: model.StorageProviderAliyunOSS,
		Status:   model.StorageProfileStatusEnabled,
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", Region: "cn-hangzhou", Bucket: "test-assets-bucket",
		AuthType: model.StorageCredentialAuthStatic, AccessKeyID: "test-id", AccessKeySecret: "test-secret",
	})
	require.NoError(t, err)
	image := model.StorageObject{
		ObjectID: "image-object", StorageProfileID: profile.ID, Purpose: model.StorageObjectPurposeAssetLibrary,
		ObjectKey: "assets/portrait.jpg", ContentType: "image/jpeg", Status: model.StorageObjectStatusUploaded,
	}
	require.NoError(t, model.DB.Create(&image).Error)

	thumbnail, err := PresignAssetObject(context.Background(), image.ID, time.Hour, AssetObjectURLThumbnail, "")
	require.NoError(t, err)
	thumbnailURL, err := url.Parse(thumbnail)
	require.NoError(t, err)
	assert.Equal(t, "image/resize,m_lfit,w_640,h_360/quality,q_80", thumbnailURL.Query().Get("x-oss-process"))
	assert.NotEmpty(t, thumbnailURL.Query().Get("x-oss-signature"))
	assert.Empty(t, thumbnailURL.Query().Get("response-content-disposition"))
	png := model.StorageObject{
		ObjectID: "png-object", StorageProfileID: profile.ID, Purpose: model.StorageObjectPurposeAssetLibrary,
		ObjectKey: "assets/logo.png", ContentType: "image/png", Status: model.StorageObjectStatusUploaded,
	}
	require.NoError(t, model.DB.Create(&png).Error)
	pngPreview, err := PresignAssetObject(context.Background(), png.ID, time.Hour, AssetObjectURLThumbnail, "")
	require.NoError(t, err)
	pngURL, err := url.Parse(pngPreview)
	require.NoError(t, err)
	assert.Equal(t, "image/resize,m_lfit,w_640,h_360", pngURL.Query().Get("x-oss-process"))

	download, err := PresignAssetObject(context.Background(), image.ID, time.Hour, AssetObjectURLDownload, "asset-public-id")
	require.NoError(t, err)
	downloadURL, err := url.Parse(download)
	require.NoError(t, err)
	assert.Equal(t, `attachment; filename="asset-public-id.jpg"`, downloadURL.Query().Get("response-content-disposition"))
	assert.NotEmpty(t, downloadURL.Query().Get("x-oss-signature"))
	assert.Empty(t, downloadURL.Query().Get("x-oss-process"), "download must retain original bytes")

	video := model.StorageObject{
		ObjectID: "video-object", StorageProfileID: profile.ID, Purpose: model.StorageObjectPurposeAssetLibrary,
		ObjectKey: "assets/clip.mp4", ContentType: "video/mp4", Status: model.StorageObjectStatusUploaded,
	}
	require.NoError(t, model.DB.Create(&video).Error)
	videoPreview, err := PresignAssetObject(context.Background(), video.ID, time.Hour, AssetObjectURLThumbnail, "")
	require.NoError(t, err)
	videoURL, err := url.Parse(videoPreview)
	require.NoError(t, err)
	assert.Empty(t, videoURL.Query().Get("x-oss-process"), "non-image previews must not use image processing")
}

func TestStudioPoliciesPreserveHistoricalObjectsAndIsolateOtherBusinesses(t *testing.T) {
	setupStorageDatabase(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "studio-policy-test-encryption-key")
	profiles := make([]int, 0, 2)
	for _, bucket := range []string{"old-studio-bucket", "new-studio-bucket"} {
		profile, err := SaveProfile(0, ProfileInput{Name: bucket, ProviderType: model.StorageProviderAliyunOSS,
			Status: model.StorageProfileStatusEnabled, Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", Region: "cn-hangzhou", Bucket: bucket,
			AuthType: model.StorageCredentialAuthStatic, AccessKeyID: "test-id", AccessKeySecret: "test-secret"})
		require.NoError(t, err)
		profiles = append(profiles, profile.ID)
	}
	legacy := model.DefaultAssetLibraryStoragePolicy()
	legacy.StorageProfileID, legacy.Enabled = profiles[0], true
	require.NoError(t, model.SaveStoragePolicy(legacy))
	require.NoError(t, model.InitializeImageStudioRetention(model.DB))
	policy, err := GetImageStudioPolicy(model.StorageObjectPurposeImageStudio)
	require.NoError(t, err)
	assert.Equal(t, profiles[0], policy.StorageProfileID)
	now := time.Now().Unix()
	object := model.StorageObject{ObjectID: "historical-studio", OwnerUserID: 42, StorageProfileID: profiles[0], StoragePolicyID: policy.ID,
		Purpose: model.StorageObjectPurposeImageStudio, ObjectKey: "original/image.png", ContentType: "image/png",
		CreatedAt: now - 31*24*3600, DeleteAfter: now + 3600, Status: model.StorageObjectStatusUploaded}
	require.NoError(t, model.DB.Create(&object).Error)
	input := PolicyInput{StorageProfileID: profiles[1], Enabled: true, ObjectPrefix: "new/studio", SignedURLTTLSeconds: 7200,
		RetentionSeconds: 60 * 24 * 3600, MaxFileBytes: 64 << 20}
	_, err = SaveImageStudioPolicy(model.StorageObjectPurposeImageStudio, input)
	require.NoError(t, err)
	require.NoError(t, model.InitializeImageStudioRetention(model.DB))
	policy, err = GetImageStudioPolicy(model.StorageObjectPurposeImageStudio)
	require.NoError(t, err)
	assert.Equal(t, profiles[1], policy.StorageProfileID, "startup must not overwrite an explicit policy")
	original, err := PresignImageStudioObject(context.Background(), object.ID, 42, AssetObjectURLOriginal, "")
	require.NoError(t, err)
	parsed, err := url.Parse(original)
	require.NoError(t, err)
	assert.Equal(t, "old-studio-bucket.oss-cn-hangzhou.aliyuncs.com", parsed.Host)
	expires, err := strconv.Atoi(parsed.Query().Get("x-oss-expires"))
	require.NoError(t, err)
	assert.LessOrEqual(t, expires, 3600)
	_, err = PresignImageStudioObject(context.Background(), object.ID, 43, AssetObjectURLOriginal, "")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = PresignAssetObject(context.Background(), object.ID, time.Hour, AssetObjectURLOriginal, "")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "asset signing cannot bypass image ownership")
	_, err = PresignImageStudioObject(context.Background(), object.ID, 42, AssetObjectURLDownload, "bad\r\nname")
	assert.Error(t, err)
	reference, err := GetImageStudioPolicy(model.StorageObjectPurposeImageStudioReference)
	require.NoError(t, err)
	assert.Equal(t, profiles[0], reference.StorageProfileID)
	assert.EqualValues(t, 7*24*3600, reference.RetentionSeconds)
	unchanged, err := model.GetStoragePolicyByKey(model.StoragePolicyAssetLibrary)
	require.NoError(t, err)
	assert.Equal(t, profiles[0], unchanged.StorageProfileID)
	assert.Zero(t, unchanged.RetentionSeconds)
	for _, retention := range []int64{0, 3600, 366 * 24 * 3600} {
		input.RetentionSeconds = retention
		_, err = SaveImageStudioPolicy(model.StorageObjectPurposeImageStudio, input)
		assert.Error(t, err)
	}
	input.RetentionSeconds, input.MaxFileBytes = 86400, 65<<20
	_, err = SaveImageStudioPolicy(model.StorageObjectPurposeImageStudio, input)
	assert.Error(t, err, "policy must not exceed the studio file limit")
	_, err = SaveImageStudioPolicy(model.StoragePolicyAssetLibrary, input)
	assert.Error(t, err, "studio endpoint cannot modify another business policy")
	for _, state := range []map[string]any{
		{"purpose": model.StorageObjectPurposeAssetLibrary},
		{"purpose": model.StorageObjectPurposeImageStudio, "status": model.StorageObjectStatusDeletePending},
		{"status": model.StorageObjectStatusUploaded, "delete_after": now - 1},
	} {
		require.NoError(t, model.DB.Model(&object).Updates(state).Error)
		_, err = PresignImageStudioObject(context.Background(), object.ID, 42, AssetObjectURLOriginal, "")
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	}
}
