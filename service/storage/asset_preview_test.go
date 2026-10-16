package storage

import (
	"context"
	"net/url"
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

func TestImageStudioSignedURLsPreserveOwnershipAndPermanentStorageOptions(t *testing.T) {
	setupStorageDatabase(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "studio-preview-test-encryption-key")
	profile, err := SaveProfile(0, ProfileInput{
		Name: "Studio OSS", ProviderType: model.StorageProviderAliyunOSS,
		Status:   model.StorageProfileStatusEnabled,
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", Region: "cn-hangzhou", Bucket: "test-studio-bucket",
		AuthType: model.StorageCredentialAuthStatic, AccessKeyID: "test-id", AccessKeySecret: "test-secret",
	})
	require.NoError(t, err)
	policy := model.DefaultAssetLibraryStoragePolicy()
	policy.StorageProfileID = profile.ID
	policy.SignedURLTTLSeconds = 3600
	require.NoError(t, model.SaveStoragePolicy(policy))

	for _, format := range []struct {
		contentType string
		extension   string
		processing  string
	}{
		{"image/png", "png", "image/resize,m_lfit,w_640,h_360"},
		{"image/jpeg", "jpg", "image/resize,m_lfit,w_640,h_360/quality,q_80"},
		{"image/webp", "webp", "image/resize,m_lfit,w_640,h_360/quality,q_80"},
	} {
		t.Run(format.extension, func(t *testing.T) {
			object := model.StorageObject{
				ObjectID: "studio-" + format.extension, OwnerUserID: 42, StorageProfileID: profile.ID,
				Purpose: model.StorageObjectPurposeImageStudio, ObjectKey: "studio/image." + format.extension,
				ContentType: format.contentType, Status: model.StorageObjectStatusUploaded,
			}
			require.NoError(t, model.DB.Create(&object).Error)
			for _, mode := range []AssetObjectURLMode{AssetObjectURLOriginal, AssetObjectURLThumbnail, AssetObjectURLDownload} {
				signed, err := PresignImageStudioObject(context.Background(), object.ID, 42, mode, "result-id")
				require.NoError(t, err)
				parsed, err := url.Parse(signed)
				require.NoError(t, err)
				assert.Equal(t, "test-studio-bucket.oss-cn-hangzhou.aliyuncs.com", parsed.Host)
				assert.Equal(t, "/"+object.ObjectKey, parsed.Path)
				assert.NotEmpty(t, parsed.Query().Get("x-oss-signature"))
				assert.Equal(t, "3600", parsed.Query().Get("x-oss-expires"))
				if mode == AssetObjectURLThumbnail {
					assert.Equal(t, format.processing, parsed.Query().Get("x-oss-process"))
				} else {
					assert.Empty(t, parsed.Query().Get("x-oss-process"), "original and download URLs retain original bytes")
				}
				if mode == AssetObjectURLDownload {
					assert.Equal(t, `attachment; filename="result-id.`+format.extension+`"`, parsed.Query().Get("response-content-disposition"))
				} else {
					assert.Empty(t, parsed.Query().Get("response-content-disposition"))
				}
			}
			_, err = PresignImageStudioObject(context.Background(), object.ID, 43, AssetObjectURLOriginal, "")
			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			_, err = PresignAssetObject(context.Background(), object.ID, time.Hour, AssetObjectURLOriginal, "")
			assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "asset signing must not bypass image ownership")
			_, err = PresignImageStudioObject(context.Background(), object.ID, 42, AssetObjectURLDownload, "bad\r\nname")
			assert.Error(t, err)
			require.NoError(t, model.DB.Model(&object).Update("purpose", model.StorageObjectPurposeAssetLibrary).Error)
			_, err = PresignImageStudioObject(context.Background(), object.ID, 42, AssetObjectURLOriginal, "")
			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			require.NoError(t, model.DB.Model(&object).Updates(map[string]any{"purpose": model.StorageObjectPurposeImageStudio, "status": model.StorageObjectStatusDeletePending}).Error)
			_, err = PresignImageStudioObject(context.Background(), object.ID, 42, AssetObjectURLOriginal, "")
			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		})
	}
}
