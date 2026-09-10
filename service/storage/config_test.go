package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupStorageDatabase(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.StorageProfile{},
		&model.StorageCredential{},
		&model.StoragePolicy{},
		&model.StorageObject{},
	))
	model.DB = db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
}

func TestSaveProfileEncryptsStaticCredentialAndReturnsOnlyHint(t *testing.T) {
	setupStorageDatabase(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "test-master-key-with-sufficient-entropy")

	view, err := SaveProfile(0, ProfileInput{
		Name:            "Primary OSS",
		ProviderType:    model.StorageProviderAliyunOSS,
		Status:          model.StorageProfileStatusEnabled,
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		Region:          "cn-hangzhou",
		Bucket:          "test-media-bucket",
		AuthType:        model.StorageCredentialAuthStatic,
		AccessKeyID:     "LTAI1234567890",
		AccessKeySecret: "super-secret-value",
	})

	require.NoError(t, err)
	assert.Equal(t, "****7890", view.AccessKeyHint)
	assert.True(t, view.CredentialConfigured)
	credential, err := model.GetActiveStorageCredential(view.ID)
	require.NoError(t, err)
	require.NotNil(t, credential)
	assert.NotContains(t, credential.EncryptedPayload, "LTAI1234567890")
	assert.NotContains(t, credential.EncryptedPayload, "super-secret-value")
	secret, err := decryptCredential(credential.EncryptedPayload)
	require.NoError(t, err)
	assert.Equal(t, "LTAI1234567890", secret.AccessKeyID)
	assert.Equal(t, "super-secret-value", secret.AccessKeySecret)
}

func TestRelayMediaPolicyReferencesEnabledProfile(t *testing.T) {
	setupStorageDatabase(t)
	t.Setenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY", "test-master-key-with-sufficient-entropy")
	view, err := SaveProfile(0, ProfileInput{
		Name:            "Primary OSS",
		ProviderType:    model.StorageProviderAliyunOSS,
		Status:          model.StorageProfileStatusEnabled,
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		Region:          "cn-hangzhou",
		Bucket:          "test-media-bucket",
		AuthType:        model.StorageCredentialAuthStatic,
		AccessKeyID:     "LTAI1234567890",
		AccessKeySecret: "super-secret-value",
	})
	require.NoError(t, err)

	policy, err := SaveRelayMediaPolicy(PolicyInput{
		StorageProfileID:    view.ID,
		ObjectPrefix:        "/temporary/relay-media/",
		SignedURLTTLSeconds: 3600,
		RetentionSeconds:    7200,
		MaxFileBytes:        1024,
		MaxTotalBytes:       2048,
		MaxFiles:            2,
		AllowedMIMETypes:    "image/webp, image/png, image/webp",
		Enabled:             true,
	})

	require.NoError(t, err)
	assert.Equal(t, "temporary/relay-media", policy.ObjectPrefix)
	assert.Equal(t, "image/webp,image/png", policy.AllowedMIMETypes)
	assert.True(t, policy.Enabled)
	assert.False(t, strings.Contains(policy.ObjectPrefix, ".."))
}

func TestStorageObjectDeletionLeasePreventsConcurrentCleanup(t *testing.T) {
	setupStorageDatabase(t)
	now := time.Now().Unix()
	object := &model.StorageObject{
		ObjectID:         "obj_cleanup_test",
		StorageProfileID: 1,
		StoragePolicyID:  1,
		Purpose:          model.StorageObjectPurposeRelayMediaTemp,
		ObjectKey:        "temporary/relay-media/test.webp",
		ContentType:      "image/webp",
		Status:           model.StorageObjectStatusUploaded,
		DeleteAfter:      now,
	}
	require.NoError(t, model.CreateStorageObject(object))

	firstClaim, err := model.ClaimStorageObjectsForDeletion(now, now+120, "runner-a", 10)
	require.NoError(t, err)
	require.Len(t, firstClaim, 1)
	secondClaim, err := model.ClaimStorageObjectsForDeletion(now, now+120, "runner-b", 10)
	require.NoError(t, err)
	assert.Empty(t, secondClaim)

	require.NoError(t, model.FailStorageObjectDeletion(object.ID, "runner-a", now+60, "temporary failure"))
	retryClaim, err := model.ClaimStorageObjectsForDeletion(now+60, now+180, "runner-b", 10)
	require.NoError(t, err)
	require.Len(t, retryClaim, 1)
	require.NoError(t, model.FinishStorageObjectDeletion(object.ID, "runner-b"))

	var stored model.StorageObject
	require.NoError(t, model.DB.First(&stored, object.ID).Error)
	assert.Equal(t, model.StorageObjectStatusDeleted, stored.Status)
	assert.Equal(t, 1, stored.DeleteAttempts)
	assert.NotZero(t, stored.DeletedAt)

	deleted, err := model.PurgeDeletedStorageObjects(stored.DeletedAt, 10)
	require.NoError(t, err)
	assert.EqualValues(t, 1, deleted)
	assert.ErrorIs(t, model.DB.First(&model.StorageObject{}, object.ID).Error, gorm.ErrRecordNotFound)
}

func TestMarkStorageObjectUploadedUsesMappedETagColumn(t *testing.T) {
	setupStorageDatabase(t)
	object := &model.StorageObject{
		ObjectID:         "obj_upload_test",
		StorageProfileID: 1,
		StoragePolicyID:  1,
		Purpose:          model.StorageObjectPurposeRelayMediaTemp,
		ObjectKey:        "temporary/relay-media/test.webp",
		ContentType:      "image/webp",
		Status:           model.StorageObjectStatusUploading,
	}
	require.NoError(t, model.CreateStorageObject(object))

	require.NoError(t, model.MarkStorageObjectUploaded(object.ID, "aliyun-etag"))

	var stored model.StorageObject
	require.NoError(t, model.DB.First(&stored, object.ID).Error)
	assert.Equal(t, model.StorageObjectStatusUploaded, stored.Status)
	assert.Equal(t, "aliyun-etag", stored.ETag)
	assert.NotZero(t, stored.UpdatedAt)
}

func TestStoragePolicyRejectsUnsupportedMediaType(t *testing.T) {
	setupStorageDatabase(t)

	_, err := SaveRelayMediaPolicy(PolicyInput{
		ObjectPrefix:        "temporary/relay-media",
		SignedURLTTLSeconds: 3600,
		RetentionSeconds:    7200,
		MaxFileBytes:        1024,
		MaxTotalBytes:       2048,
		MaxFiles:            2,
		AllowedMIMETypes:    "image/gif",
		Enabled:             false,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported storage policy MIME type")
}

func TestStoragePolicyAcceptsVideoMediaTypes(t *testing.T) {
	setupStorageDatabase(t)

	policy, err := SaveRelayMediaPolicy(PolicyInput{
		ObjectPrefix:        "temporary/relay-media",
		SignedURLTTLSeconds: 3600,
		RetentionSeconds:    7200,
		MaxFileBytes:        1024,
		MaxTotalBytes:       2048,
		MaxFiles:            2,
		AllowedMIMETypes:    "video/mp4, video/webm, video/quicktime",
		Enabled:             false,
	})

	require.NoError(t, err)
	assert.Equal(t, "video/mp4,video/webm,video/quicktime", policy.AllowedMIMETypes)
}

func TestAliyunProfileRejectsCustomEndpointPort(t *testing.T) {
	err := validateAliyunProfile(&model.StorageProfile{
		Endpoint: "https://oss-cn-hangzhou.aliyuncs.com:444",
		Region:   "cn-hangzhou",
		Bucket:   "private-media",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTPS URL")
}
