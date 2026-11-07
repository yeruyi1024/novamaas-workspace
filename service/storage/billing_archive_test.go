package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type billingArchiveTransport struct {
	data    []byte
	putKeys *[]string
}

func (transport billingArchiveTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if transport.putKeys != nil && request.Method == http.MethodPut {
		*transport.putKeys = append(*transport.putKeys, request.URL.Path)
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/octet-stream"}}, Body: io.NopCloser(bytes.NewReader(transport.data)), ContentLength: int64(len(transport.data)), Request: request}, nil
}

func TestBillingArchiveUploadsUnderAccountMonthAndStatement(t *testing.T) {
	savedDB, savedSecret := model.DB, common.CryptoSecret
	common.CryptoSecret = "permanent-storage-test-key"
	t.Cleanup(func() { model.DB, common.CryptoSecret = savedDB, savedSecret })
	setupStorageDatabase(t)
	require.NoError(t, model.DB.AutoMigrate(&model.BillingStatement{}, &model.BillingArtifact{}))
	profile := &model.StorageProfile{Name: "billing", ProviderType: model.StorageProviderAliyunOSS, Status: model.StorageProfileStatusEnabled, Bucket: "test-bucket"}
	require.NoError(t, model.DB.Create(profile).Error)
	statement := &model.BillingStatement{ID: "statement-uuid", UserID: 42, Month: "2026-09", Revision: 1}
	require.NoError(t, model.DB.Create(statement).Error)
	payload := []byte("immutable billing artifact")
	var keys []string
	config := oss.LoadDefaultConfig().WithRegion("cn-hangzhou").WithEndpoint("https://oss-cn-hangzhou.aliyuncs.com").WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test-id", "test-secret")).WithHttpClient(&http.Client{Transport: billingArchiveTransport{data: payload, putKeys: &keys}})
	store := &BillingArchiveStore{driver: &aliyunOSSDriver{client: oss.NewClient(config), bucket: profile.Bucket}, profileID: profile.ID, profile: profile}
	missing, err := store.Existing(context.Background(), statement.ID, "xlsx", 0)
	require.NoError(t, err)
	assert.Nil(t, missing)
	for _, kind := range []string{"details", "snapshot", "pdf", "xlsx", "manifest", "receipt"} {
		t.Run(kind, func(t *testing.T) {
			artifact, err := store.Put(context.Background(), statement.ID, kind, 0, 42, "application/octet-stream", payload, 0)
			require.NoError(t, err)
			expected := fmt.Sprintf("billing/user-65dad71240db84f1791cc8ba/2026-09/statement-uuid/%s/000000-%s", kind, artifact.SHA256)
			assert.Equal(t, expected, artifact.ObjectKey)
			assert.Equal(t, "/"+expected, keys[len(keys)-1])
			var object model.StorageObject
			require.NoError(t, model.DB.Where("object_id = ?", artifact.ObjectID).First(&object).Error)
			assert.Equal(t, expected, object.ObjectKey)
			assert.Equal(t, 42, object.OwnerUserID)
			uploads := len(keys)
			retry, err := store.Put(context.Background(), statement.ID, kind, 0, 42, "application/octet-stream", payload, 0)
			require.NoError(t, err)
			assert.Equal(t, artifact.ObjectKey, retry.ObjectKey)
			assert.Len(t, keys, uploads, "immutable retries reuse the archived artifact")
			existing, err := store.Existing(context.Background(), statement.ID, kind, 0)
			require.NoError(t, err)
			assert.Equal(t, artifact.SHA256, existing.SHA256)
			assert.Len(t, keys, uploads, "existing artifacts are verified without uploading")
		})
	}
	common.CryptoSecret = "rotated-permanent-storage-test-key"
	uploads := len(keys)
	receipt, err := store.Put(context.Background(), statement.ID, "receipt", 0, 42, "application/octet-stream", payload, 0)
	require.NoError(t, err)
	assert.Contains(t, receipt.ObjectKey, "billing/user-65dad71240db84f1791cc8ba/2026-09/")
	assert.Len(t, keys, uploads, "saved object paths survive signing key rotation")
	_, err = store.Put(context.Background(), statement.ID, "pdf", 1, 7, "application/pdf", payload, 0)
	require.ErrorContains(t, err, "owner")
	invalid := &model.BillingStatement{ID: "invalid-month", UserID: 42, Month: "../../elsewhere", Revision: 1}
	require.NoError(t, model.DB.Create(invalid).Error)
	_, err = store.Put(context.Background(), invalid.ID, "pdf", 1, 42, "application/pdf", payload, 0)
	require.ErrorContains(t, err, "month")
}

func TestBillingArchiveReadRejectsCorruptionAndTruncation(t *testing.T) {
	payload := []byte("immutable statement detail")
	hash := sha256.Sum256(payload)
	digest := hex.EncodeToString(hash[:])
	for _, test := range []struct {
		name  string
		data  []byte
		valid bool
	}{{"valid", payload, true}, {"corrupt", []byte("changed statement content"), false}, {"truncated", payload[:5], false}, {"oversized", append(append([]byte(nil), payload...), 1), false}} {
		t.Run(test.name, func(t *testing.T) {
			config := oss.LoadDefaultConfig().WithRegion("cn-hangzhou").WithEndpoint("https://oss-cn-hangzhou.aliyuncs.com").WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test-id", "test-secret")).WithHttpClient(&http.Client{Transport: billingArchiveTransport{data: test.data}})
			store := &BillingArchiveStore{driver: &aliyunOSSDriver{client: oss.NewClient(config), bucket: "test-bucket"}}
			body, err := store.Read(context.Background(), "billing/statements/test", int64(len(payload)), digest)
			if test.valid {
				require.NoError(t, err)
				assert.Equal(t, payload, body)
			} else {
				require.Error(t, err)
				assert.Nil(t, body)
			}
		})
	}
}

func TestBillingArchiveIsExcludedFromCleanupEvenWithAnExpiredMarker(t *testing.T) {
	setupStorageDatabase(t)
	now := common.GetTimestamp()
	permanent := &model.StorageObject{ObjectID: "billing-proof", Purpose: model.StorageObjectPurposeBillingArchive, TaskID: "shared-task", Status: model.StorageObjectStatusUploaded, DeleteAfter: now - 1}
	temporary := &model.StorageObject{ObjectID: "temporary-media", Purpose: model.StorageObjectPurposeRelayMediaTemp, TaskID: "shared-task", Status: model.StorageObjectStatusUploaded, DeleteAfter: now - 1}
	require.NoError(t, model.DB.Create(permanent).Error)
	require.NoError(t, model.DB.Create(temporary).Error)
	require.NoError(t, model.MarkStorageObjectsForTaskDeletion("shared-task", now))
	claimed, err := model.ClaimStorageObjectsForDeletion(now, now+60, "worker", 10)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	assert.Equal(t, temporary.ObjectID, claimed[0].ObjectID)
	require.NoError(t, model.DB.Model(permanent).Updates(map[string]interface{}{"status": model.StorageObjectStatusDeleted, "deleted_at": now - 100}).Error)
	_, err = model.PurgeDeletedStorageObjects(now, 100)
	require.NoError(t, err)
	var preserved model.StorageObject
	require.NoError(t, model.DB.First(&preserved, permanent.ID).Error)
	assert.Equal(t, permanent.ObjectID, preserved.ObjectID)
}

func TestBillingStorageRegistrationRejectsChangedIdentityAndPreventsArchival(t *testing.T) {
	setupStorageDatabase(t)
	profile := &model.StorageProfile{Name: "billing", ProviderType: model.StorageProviderAliyunOSS, Status: model.StorageProfileStatusEnabled, Bucket: "old-bucket"}
	require.NoError(t, model.DB.Create(profile).Error)
	object := &model.StorageObject{ObjectID: "permanent-proof", StorageProfileID: profile.ID, Purpose: model.StorageObjectPurposeBillingArchive, Status: model.StorageObjectStatusUploading}
	require.NoError(t, model.DB.Model(&model.StorageProfile{}).Where("id = ?", profile.ID).Update("bucket", "new-bucket").Error)
	require.ErrorContains(t, model.RegisterBillingArchiveObject(profile, object), "identity changed")
	var count int64
	require.NoError(t, model.DB.Model(&model.StorageObject{}).Count(&count).Error)
	assert.Zero(t, count)
	profile.Bucket = "new-bucket"
	require.NoError(t, model.RegisterBillingArchiveObject(profile, object))
	require.ErrorContains(t, ArchiveProfile(profile.ID), "still has objects")
	require.NoError(t, model.DB.Model(object).Update("status", model.StorageObjectStatusDeleted).Error)
	require.ErrorContains(t, ArchiveProfile(profile.ID), "still has objects", "a corrupt deletion marker must not orphan permanent evidence")
}
