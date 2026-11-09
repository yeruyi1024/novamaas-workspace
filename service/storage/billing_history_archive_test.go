package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type billingHistoryArchiveFixture struct {
	store   *BillingArchiveStore
	body    []byte
	putKeys []string
	calls   int
}

func (fixture *billingHistoryArchiveFixture) RoundTrip(request *http.Request) (*http.Response, error) {
	fixture.calls++
	return (billingArchiveTransport{data: fixture.body, putKeys: &fixture.putKeys}).RoundTrip(request)
}

func setupBillingHistoryArchive(t *testing.T, body []byte) *billingHistoryArchiveFixture {
	t.Helper()
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	previousSecret := common.CryptoSecret
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
		common.CryptoSecret = previousSecret
	})
	common.CryptoSecret = "history-archive-test-key"
	setupStorageDatabase(t)
	connection, err := model.DB.DB()
	require.NoError(t, err)
	connection.SetMaxOpenConns(1)
	t.Cleanup(func() { assert.NoError(t, connection.Close()) })
	require.NoError(t, model.DB.AutoMigrate(&model.BillingStatement{}, &model.BillingArtifact{}, &model.BillingHistoryImport{}))
	profile := &model.StorageProfile{
		Name: "history archive", ProviderType: model.StorageProviderAliyunOSS, Status: model.StorageProfileStatusEnabled,
		Bucket: "history-test-bucket", Region: "cn-hangzhou", Endpoint: "https://oss-cn-hangzhou.aliyuncs.com",
	}
	require.NoError(t, model.DB.Create(profile).Error)
	fixture := &billingHistoryArchiveFixture{body: body}
	config := oss.LoadDefaultConfig().WithRegion(profile.Region).WithEndpoint(profile.Endpoint).
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test-id", "test-secret")).
		WithHttpClient(&http.Client{Transport: fixture})
	fixture.store = &BillingArchiveStore{
		driver: &aliyunOSSDriver{client: oss.NewClient(config), bucket: profile.Bucket}, profileID: profile.ID, profile: profile,
	}
	return fixture
}

func TestBillingHistorySourceArchivesWithoutAStatementOrImportBatch(t *testing.T) {
	body := []byte("immutable compressed history source")
	fixture := setupBillingHistoryArchive(t, body)
	const importID = "history-0123456789abcdef0123456789abcdef"
	ctx := context.Background()

	artifact, err := fixture.store.PutHistorySource(ctx, importID, 42, "2026-10", body, 2)
	require.NoError(t, err)
	require.NotNil(t, artifact)
	assert.Equal(t, importID, artifact.StatementID)
	assert.Equal(t, "source", artifact.Kind)
	assert.Zero(t, artifact.Ordinal)
	assert.EqualValues(t, 2, artifact.Rows)
	expected := fmt.Sprintf("billing/user-%s/2026-10/%s/source/000000-%s", permanentStorageUserHash(42), importID, artifact.SHA256)
	assert.Equal(t, expected, artifact.ObjectKey)
	assert.Equal(t, []string{"/" + expected}, fixture.putKeys)
	var object model.StorageObject
	require.NoError(t, model.DB.Where("object_id = ?", artifact.ObjectID).First(&object).Error)
	assert.Equal(t, 42, object.OwnerUserID)
	assert.Equal(t, fixture.store.profileID, object.StorageProfileID)
	assert.Equal(t, model.StorageObjectPurposeBillingArchive, object.Purpose)
	assert.Equal(t, model.StorageObjectStatusUploaded, object.Status)
	assert.Equal(t, "application/gzip", object.ContentType)
	assert.Equal(t, artifact.ObjectKey, object.ObjectKey)
	assert.Zero(t, object.DeleteAfter)
	for _, value := range []any{&model.BillingStatement{}, &model.BillingHistoryImport{}} {
		var count int64
		require.NoError(t, model.DB.Model(value).Count(&count).Error)
		assert.Zero(t, count, "source archival must precede ledger import without creating a fake statement")
	}

	retry, err := fixture.store.PutHistorySource(ctx, importID, 42, "2026-10", body, 2)
	require.NoError(t, err)
	assert.Equal(t, artifact.ID, retry.ID)
	assert.Equal(t, artifact.ObjectKey, retry.ObjectKey)
	assert.Len(t, fixture.putKeys, 1, "an identical retry must only verify the archived source")
	_, err = fixture.store.PutHistorySource(ctx, importID, 42, "2026-10", []byte("changed source"), 2)
	assert.ErrorIs(t, err, model.ErrBillingConflict)
	assert.Len(t, fixture.putKeys, 1)
	for _, value := range []any{&model.StorageObject{}, &model.BillingArtifact{}} {
		var count int64
		require.NoError(t, model.DB.Model(value).Count(&count).Error)
		assert.EqualValues(t, 1, count)
	}
}

func TestBillingHistorySourceReusesLegacyObjectsAfterSigningKeyRotation(t *testing.T) {
	for _, layout := range []string{"original statement directory", "account and month directory"} {
		t.Run(layout, func(t *testing.T) {
			body := []byte("legacy frozen source")
			fixture := setupBillingHistoryArchive(t, body)
			const importID = "history-0123456789ABCDEF0123456789ABCDEF"
			hash := sha256.Sum256(body)
			digest := hex.EncodeToString(hash[:])
			previousUserDirectory := permanentStorageUserHash(42)
			key := fmt.Sprintf("billing/user-%s/2026-09/%s/source/000000-%s", previousUserDirectory, importID, digest)
			if layout == "original statement directory" {
				key = fmt.Sprintf("billing/statements/%s/source/000000-%s", importID, digest)
			}
			objectHash := sha256.Sum256([]byte(fmt.Sprintf("%d:%s", fixture.store.profileID, key)))
			object := model.StorageObject{
				ObjectID: hex.EncodeToString(objectHash[:]), OwnerUserID: 42, StorageProfileID: fixture.store.profileID,
				Purpose: model.StorageObjectPurposeBillingArchive, ObjectKey: key, ContentType: "application/gzip",
				Size: int64(len(body)), SHA256: digest, Status: model.StorageObjectStatusUploaded,
			}
			require.NoError(t, model.DB.Create(&object).Error)
			previous := model.BillingArtifact{StatementID: importID, Kind: "source", ObjectID: object.ObjectID, ObjectKey: key, SHA256: digest, Size: int64(len(body)), Rows: 2}
			require.NoError(t, model.DB.Create(&previous).Error)
			common.CryptoSecret = "rotated-history-archive-test-key"
			require.NotEqual(t, previousUserDirectory, permanentStorageUserHash(42))

			artifact, err := fixture.store.PutHistorySource(context.Background(), importID, 42, "2026-09", body, 2)

			require.NoError(t, err)
			assert.Equal(t, previous.ID, artifact.ID)
			assert.Equal(t, key, artifact.ObjectKey)
			assert.Empty(t, fixture.putKeys, "a persisted source path must survive signing-key changes")
			assert.Positive(t, fixture.calls, "reuse must verify the original object")
		})
	}
}

func TestBillingHistorySourceCannotReuseAnotherOwnerOrProfile(t *testing.T) {
	for _, mismatch := range []string{"owner", "profile"} {
		t.Run(mismatch, func(t *testing.T) {
			body := []byte("private source")
			fixture := setupBillingHistoryArchive(t, body)
			const importID = "history-0123456789abcdef0123456789abcdef"
			_, err := fixture.store.PutHistorySource(context.Background(), importID, 42, "2026-10", body, 1)
			require.NoError(t, err)
			calls := fixture.calls
			store, userID := fixture.store, 42
			if mismatch == "owner" {
				userID = 7
			} else {
				profile := *fixture.store.profile
				profile.ID, profile.Name = 0, "another history archive"
				require.NoError(t, model.DB.Create(&profile).Error)
				otherStore := *fixture.store
				otherStore.profile, otherStore.profileID = &profile, profile.ID
				store = &otherStore
			}

			artifact, err := store.PutHistorySource(context.Background(), importID, userID, "2026-10", body, 1)

			assert.ErrorIs(t, err, model.ErrBillingConflict)
			assert.Nil(t, artifact)
			assert.Equal(t, calls, fixture.calls, "owner and profile must be checked before cloud I/O")
		})
	}
}

func TestBillingHistorySourceRejectsInvalidIdentityBeforeStorage(t *testing.T) {
	const importID = "history-0123456789abcdef0123456789abcdef"
	for _, test := range []struct {
		name, id, month string
		userID          int
		rows            int64
	}{
		{"missing history prefix", "0123456789abcdef0123456789abcdef", "2026-10", 42, 1},
		{"short history ID", "history-short", "2026-10", 42, 1},
		{"non-hex history ID", "history-zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", "2026-10", 42, 1},
		{"missing owner", importID, "2026-10", 0, 1},
		{"negative owner", importID, "2026-10", -1, 1},
		{"missing month", importID, "", 42, 1},
		{"invalid month", importID, "2026-13", 42, 1},
		{"noncanonical month", importID, "2026-9", 42, 1},
		{"pre-epoch month", importID, "1969-12", 42, 1},
		{"empty records", importID, "2026-10", 42, 0},
		{"too many records", importID, "2026-10", 42, model.MaxBillingHistoryRecords + 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := setupBillingHistoryArchive(t, []byte("frozen source"))

			artifact, err := fixture.store.PutHistorySource(context.Background(), test.id, test.userID, test.month, fixture.body, test.rows)

			require.Error(t, err)
			assert.Nil(t, artifact)
			assert.Zero(t, fixture.calls)
			for _, value := range []any{&model.StorageObject{}, &model.BillingArtifact{}} {
				var count int64
				require.NoError(t, model.DB.Model(value).Count(&count).Error)
				assert.Zero(t, count)
			}
		})
	}
}

func TestBillingArchiveExistingArtifactStillRequiresTheStatementOwner(t *testing.T) {
	body := []byte("formal statement artifact")
	fixture := setupBillingHistoryArchive(t, body)
	statement := &model.BillingStatement{ID: "owner-checked-statement", UserID: 42, Month: "2026-10", Revision: 1}
	require.NoError(t, model.DB.Create(statement).Error)
	_, err := fixture.store.Put(context.Background(), statement.ID, "pdf", 0, 42, "application/pdf", body, 0)
	require.NoError(t, err)
	calls := fixture.calls

	artifact, err := fixture.store.Put(context.Background(), statement.ID, "pdf", 0, 7, "application/pdf", body, 0)

	require.ErrorContains(t, err, "owner")
	assert.Nil(t, artifact)
	assert.Equal(t, calls, fixture.calls)
}
