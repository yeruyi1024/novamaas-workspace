package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type upstreamArchiveTransport struct {
	t           *testing.T
	objects     map[string][]byte
	puts        int
	corruptRead bool
}

func (transport *upstreamArchiveTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.t.Helper()
	assert.Contains(transport.t, request.Header.Get("Authorization"), "OSS4-HMAC-SHA256", "archives still use the OSS SDK's authenticated requests")
	status := http.StatusOK
	var body []byte
	switch request.Method {
	case http.MethodPut:
		assert.Equal(transport.t, "true", request.Header.Get("x-oss-forbid-overwrite"))
		assert.Equal(transport.t, "private", request.Header.Get("x-oss-object-acl"), "all internal cost evidence must stay private even in a public bucket")
		transport.puts++
		if _, exists := transport.objects[request.URL.Path]; exists {
			status = http.StatusConflict
			body = []byte(`<Error><Code>FileAlreadyExists</Code><Message>immutable object exists</Message></Error>`)
		} else {
			var err error
			body, err = io.ReadAll(request.Body)
			require.NoError(transport.t, err)
			transport.objects[request.URL.Path] = body
			body = nil
		}
	case http.MethodGet:
		body = transport.objects[request.URL.Path]
		if transport.corruptRead {
			body = []byte("corrupt read back")
		}
	default:
		transport.t.Errorf("unexpected OSS archive method %s", request.Method)
	}
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/octet-stream"}}, Body: io.NopCloser(bytes.NewReader(body)), ContentLength: int64(len(body)), Request: request}, nil
}

func TestUpstreamArchiveUsesPermanentIsolatedStorageAndImmutableRetries(t *testing.T) {
	savedDB := model.DB
	t.Cleanup(func() { model.DB = savedDB })
	setupStorageDatabase(t)
	require.NoError(t, model.DB.AutoMigrate(&model.UpstreamBillingStatement{}, &model.BillingArtifact{}))
	profile := &model.StorageProfile{Name: "upstream-finance", ProviderType: model.StorageProviderAliyunOSS, Status: model.StorageProfileStatusEnabled, Bucket: "test-bucket"}
	require.NoError(t, model.DB.Create(profile).Error)
	statement := &model.UpstreamBillingStatement{ID: "upstream-statement-uuid", GroupID: 7, Month: "2026-09", Revision: 1, CreatedBy: 42, StorageProfileID: profile.ID}
	require.NoError(t, model.DB.Create(statement).Error)
	transport := &upstreamArchiveTransport{t: t, objects: map[string][]byte{}}
	config := oss.LoadDefaultConfig().WithRegion("cn-hangzhou").WithEndpoint("https://oss-cn-hangzhou.aliyuncs.com").WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test-id", "test-secret")).WithHttpClient(&http.Client{Transport: transport})
	store := &BillingArchiveStore{driver: &aliyunOSSDriver{client: oss.NewClient(config), bucket: profile.Bucket}, profileID: profile.ID, profile: profile}
	for _, kind := range []string{"details", "snapshot", "xlsx", "pdf", "manifest", "receipt"} {
		payload := []byte("frozen upstream " + kind)
		artifact, err := store.PutUpstream(context.Background(), statement.ID, kind, 0, statement.CreatedBy, "application/octet-stream", payload, 0)
		require.NoError(t, err)
		expected := fmt.Sprintf("billing/upstream-7/2026-09/upstream-statement-uuid/%s/000000-%s", kind, artifact.SHA256)
		assert.Equal(t, expected, artifact.ObjectKey)
		assert.Equal(t, payload, transport.objects["/"+expected])
		var object model.StorageObject
		require.NoError(t, model.DB.Where("object_id = ?", artifact.ObjectID).First(&object).Error)
		assert.Equal(t, model.StorageObjectPurposeBillingArchive, object.Purpose)
		assert.Equal(t, model.StorageObjectStatusUploaded, object.Status)
		assert.Equal(t, 42, object.OwnerUserID)
		assert.Zero(t, object.DeleteAfter)
		puts := transport.puts
		retry, err := store.PutUpstream(context.Background(), statement.ID, kind, 0, statement.CreatedBy, "application/octet-stream", payload, 0)
		require.NoError(t, err)
		assert.Equal(t, artifact, retry)
		assert.Equal(t, puts, transport.puts)
		_, err = store.PutUpstream(context.Background(), statement.ID, kind, 0, statement.CreatedBy, "application/octet-stream", []byte("changed"), 0)
		assert.ErrorIs(t, err, model.ErrBillingConflict)
		_, err = store.PutUpstream(context.Background(), statement.ID, kind, 0, 6, "application/octet-stream", payload, 0)
		assert.ErrorContains(t, err, "owner")
	}
	assert.ErrorContains(t, ArchiveProfile(profile.ID), "still has objects")
	transport.corruptRead = true
	_, err := store.PutUpstream(context.Background(), statement.ID, "details", 1, 42, "application/gzip", []byte("frozen new detail"), 1)
	require.ErrorContains(t, err, "integrity")
	var published int64
	require.NoError(t, model.DB.Model(&model.BillingArtifact{}).Where("statement_id = ? AND kind = ? AND ordinal = ?", statement.ID, "details", 1).Count(&published).Error)
	assert.Zero(t, published, "a failed read-back must never publish artifact metadata")
	// An upload can finish before its metadata transaction. A retry must verify
	// that existing object and finish publication without replacing its bytes.
	transport.corruptRead = false
	recovered, err := store.PutUpstream(context.Background(), statement.ID, "details", 1, statement.CreatedBy, "application/gzip", []byte("frozen new detail"), 1)
	require.NoError(t, err)
	assert.Equal(t, []byte("frozen new detail"), transport.objects["/"+recovered.ObjectKey])
	_, err = store.PutUpstream(context.Background(), statement.ID, "../escape", 0, 42, "application/json", []byte("payload"), 0)
	assert.ErrorContains(t, err, "kind")
}
