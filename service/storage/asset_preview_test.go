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
