package storage

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, "image/resize,m_lfit,w_512,h_384/auto-orient,1/format,webp/quality,q_75", thumbnailURL.Query().Get("x-oss-process"))
	assert.Equal(t, assetThumbnailCacheControl, thumbnailURL.Query().Get("response-cache-control"))
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
	assert.Equal(t, "image/resize,m_lfit,w_512,h_384/auto-orient,1/format,webp/quality,q_75", pngURL.Query().Get("x-oss-process"))
	assert.Equal(t, assetThumbnailCacheControl, pngURL.Query().Get("response-cache-control"))

	download, err := PresignAssetObject(context.Background(), image.ID, time.Hour, AssetObjectURLDownload, "asset-public-id")
	require.NoError(t, err)
	downloadURL, err := url.Parse(download)
	require.NoError(t, err)
	assert.Equal(t, `attachment; filename="asset-public-id.jpg"`, downloadURL.Query().Get("response-content-disposition"))
	assert.NotEmpty(t, downloadURL.Query().Get("x-oss-signature"))
	assert.Empty(t, downloadURL.Query().Get("x-oss-process"), "download must retain original bytes")
	assert.Empty(t, downloadURL.Query().Get("response-cache-control"), "download keeps object no-store policy")

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
	assert.Equal(t, assetThumbnailCacheControl, videoURL.Query().Get("response-cache-control"))
}

func TestAssetThumbnailBatchSignsLikeSinglePreviewsAndCachesStableURLs(t *testing.T) {
	setupStorageDatabase(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "asset-preview-test-encryption-key")
	profile, err := SaveProfile(0, ProfileInput{
		Name: "Asset OSS", ProviderType: model.StorageProviderAliyunOSS,
		Status:   model.StorageProfileStatusEnabled,
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", Region: "cn-hangzhou", Bucket: "test-assets-bucket",
		AuthType: model.StorageCredentialAuthStatic, AccessKeyID: "test-id", AccessKeySecret: "test-secret",
	})
	require.NoError(t, err)
	objects := []model.StorageObject{
		{ObjectID: "batch-image", ObjectKey: "assets/a.jpg", ContentType: "image/jpeg", Status: model.StorageObjectStatusUploaded},
		{ObjectID: "batch-video", ObjectKey: "assets/b.mp4", ContentType: "video/mp4", Status: model.StorageObjectStatusUploaded},
		{ObjectID: "batch-pending", ObjectKey: "assets/c.png", ContentType: "image/png", Status: model.StorageObjectStatusDeletePending},
	}
	ids := make([]int64, 0, len(objects))
	for index := range objects {
		objects[index].StorageProfileID = profile.ID
		objects[index].Purpose = model.StorageObjectPurposeAssetLibrary
		require.NoError(t, model.DB.Create(&objects[index]).Error)
		ids = append(ids, objects[index].ID)
	}

	first, err := PresignAssetThumbnailsWithExpiry(context.Background(), ids, time.Hour)
	require.NoError(t, err)
	require.Len(t, first, 2)
	assert.NotContains(t, first, objects[2].ID, "files that are not uploaded have no preview")
	assert.Empty(t, first[objects[0].ID].PosterURL)
	posterURL, err := url.Parse(first[objects[1].ID].PosterURL)
	require.NoError(t, err)
	assert.Equal(t, "video/snapshot,t_100,f_jpg,w_512,h_384,m_fast", posterURL.Query().Get("x-oss-process"))
	assert.Equal(t, assetThumbnailCacheControl, posterURL.Query().Get("response-cache-control"))

	second, err := PresignAssetThumbnailsWithExpiry(context.Background(), ids, time.Hour)
	require.NoError(t, err)
	assert.Equal(t, first[objects[0].ID].URL, second[objects[0].ID].URL, "stable cached URL enables browser disk cache hits")
	assert.Equal(t, first[objects[1].ID].PosterURL, second[objects[1].ID].PosterURL)

	for _, object := range objects[:2] {
		single, err := PresignAssetObject(context.Background(), object.ID, time.Hour, AssetObjectURLThumbnail, "")
		require.NoError(t, err)
		assert.Equal(t, single, first[object.ID].URL)
	}
}

func TestAssetThumbnailBatchSkipsBrokenLegacyProfileWithoutBreakingHealthyPage(t *testing.T) {
	setupStorageDatabase(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "asset-preview-test-encryption-key")
	healthyProfile, err := SaveProfile(0, ProfileInput{
		Name: "Healthy OSS", ProviderType: model.StorageProviderAliyunOSS,
		Status:   model.StorageProfileStatusEnabled,
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", Region: "cn-hangzhou", Bucket: "healthy-bucket",
		AuthType: model.StorageCredentialAuthStatic, AccessKeyID: "healthy-id", AccessKeySecret: "healthy-secret",
	})
	require.NoError(t, err)

	healthyObj := model.StorageObject{
		ObjectID: "healthy-obj", StorageProfileID: healthyProfile.ID,
		Purpose: model.StorageObjectPurposeAssetLibrary, ObjectKey: "assets/ok.png",
		ContentType: "image/png", Status: model.StorageObjectStatusUploaded,
	}
	brokenObj := model.StorageObject{
		ObjectID: "broken-obj", StorageProfileID: 99999,
		Purpose: model.StorageObjectPurposeAssetLibrary, ObjectKey: "assets/legacy.png",
		ContentType: "image/png", Status: model.StorageObjectStatusUploaded,
	}
	require.NoError(t, model.DB.Create(&healthyObj).Error)
	require.NoError(t, model.DB.Create(&brokenObj).Error)

	result, err := PresignAssetThumbnailsWithExpiry(context.Background(), []int64{brokenObj.ID, healthyObj.ID}, time.Hour)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Contains(t, result, healthyObj.ID)
	assert.NotContains(t, result, brokenObj.ID)
}
