package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamSupplierStorageIsPrivatePermanentAndKeepsSeparateTwentyMiBLimit(t *testing.T) {
	savedDB := model.DB
	t.Cleanup(func() { model.DB = savedDB })
	setupStorageDatabase(t)
	require.NoError(t, model.DB.AutoMigrate(&model.UpstreamBillingStatement{}, &model.UpstreamBillingStatementEvent{}, &model.UpstreamBillingAttachment{}))
	profile := &model.StorageProfile{Name: "supplier-evidence", ProviderType: model.StorageProviderAliyunOSS, Status: model.StorageProfileStatusEnabled, Bucket: "test-bucket"}
	require.NoError(t, model.DB.Create(profile).Error)
	snapshot := `{"supplier":"fixture"}`
	snapshotHash := sha256.Sum256([]byte(snapshot))
	statement := &model.UpstreamBillingStatement{ID: "supplier-review", GroupID: 7, Month: "2026-09", Status: model.StatementDraft, StorageProfileID: profile.ID, CreatedBy: 42, Snapshot: model.UpstreamBillingSnapshotData(snapshot), SnapshotSHA256: hex.EncodeToString(snapshotHash[:])}
	require.NoError(t, model.DB.Create(statement).Error)
	transport := &upstreamArchiveTransport{t: t, objects: map[string][]byte{}}
	config := oss.LoadDefaultConfig().WithRegion("cn-hangzhou").WithEndpoint("https://oss-cn-hangzhou.aliyuncs.com").WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test-id", "test-secret")).WithHttpClient(&http.Client{Transport: transport})
	store := &BillingArchiveStore{driver: &aliyunOSSDriver{client: oss.NewClient(config), bucket: profile.Bucket}, profileID: profile.ID, profile: profile}
	body := bytes.Repeat([]byte{'x'}, 16*1024*1024+1)
	hash := sha256.Sum256(body)
	attachment := &model.UpstreamBillingAttachment{StatementID: statement.ID, OriginalFilename: "supplier.csv", ContentType: "text/csv; charset=utf-8", Size: int64(len(body)), SHA256: hex.EncodeToString(hash[:]), UploadedBy: 5}
	require.NoError(t, model.ReserveUpstreamBillingAttachment(attachment))
	require.NoError(t, store.PutUpstreamAttachment(context.Background(), attachment, body))
	require.NoError(t, model.CompleteUpstreamBillingAttachment(attachment.ID, 5, true))
	_, err := store.Read(context.Background(), attachment.ObjectKey, attachment.Size, attachment.SHA256)
	assert.ErrorContains(t, err, "size", "generated statement archives must retain their original sixteen-MiB bound")
	recovered, err := store.ReadUpstreamAttachment(context.Background(), attachment.ObjectKey, attachment.Size, attachment.SHA256)
	require.NoError(t, err)
	assert.Equal(t, body, recovered)
	var object model.StorageObject
	require.NoError(t, model.DB.Where("object_id = ?", attachment.ObjectID).First(&object).Error)
	assert.Equal(t, model.StorageObjectPurposeBillingArchive, object.Purpose)
	assert.Zero(t, object.DeleteAfter)
	assert.Equal(t, 5, object.OwnerUserID)
	assert.Contains(t, object.ObjectKey, "billing/upstream-7/2026-09/supplier-review/attachments/")
	assert.NotContains(t, object.ObjectKey, "supplier.csv")
	payload := []byte("model,cost\nA,7\n")
	smallHash := sha256.Sum256(payload)
	second := &model.UpstreamBillingAttachment{StatementID: statement.ID, OriginalFilename: "second.csv", ContentType: "text/csv; charset=utf-8", Size: int64(len(payload)), SHA256: hex.EncodeToString(smallHash[:]), UploadedBy: 5}
	require.NoError(t, model.ReserveUpstreamBillingAttachment(second))
	transport.corruptRead = true
	assert.ErrorContains(t, store.PutUpstreamAttachment(context.Background(), second, payload), "integrity")
	_, err = model.GetUpstreamBillingAttachment(statement.ID, second.ID)
	assert.Error(t, err, "an upload without a successful full read-back is not reviewable")
	transport.corruptRead = false
	require.NoError(t, store.PutUpstreamAttachment(context.Background(), second, payload), "an interrupted upload reuses the immutable existing object after read-back")
	require.NoError(t, model.CompleteUpstreamBillingAttachment(second.ID, 5, true))
}
