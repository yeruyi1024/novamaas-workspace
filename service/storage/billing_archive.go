package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Billing archives deliberately do not use the temporary media policy/TTL.
type BillingArchiveStore struct {
	driver    *aliyunOSSDriver
	profileID int
	profile   *model.StorageProfile
}

func OpenBillingArchiveStore(profileID int, writing bool) (*BillingArchiveStore, error) {
	profile, err := model.GetStorageProfileByID(profileID)
	if err != nil {
		return nil, err
	}
	if profile == nil || (writing && profile.Status != model.StorageProfileStatusEnabled) {
		return nil, errors.New("an enabled storage profile is required")
	}
	if profile.ProviderType != model.StorageProviderAliyunOSS {
		return nil, errors.New("billing archives currently require Aliyun OSS")
	}
	credential, err := model.GetActiveStorageCredential(profileID)
	if err != nil {
		return nil, err
	}
	if credential == nil {
		return nil, errors.New("storage credentials are not configured")
	}
	driver, err := newAliyunOSSDriver(profile, credential)
	if err != nil {
		return nil, err
	}
	return &BillingArchiveStore{driver: driver, profileID: profileID, profile: profile}, nil
}

func (store *BillingArchiveStore) Read(ctx context.Context, key string, size int64, digest string) ([]byte, error) {
	return store.readBillingObject(ctx, key, size, digest, 16*1024*1024)
}

func (store *BillingArchiveStore) ReadUpstreamAttachment(ctx context.Context, key string, size int64, digest string) ([]byte, error) {
	return store.readBillingObject(ctx, key, size, digest, model.MaxUpstreamBillingAttachmentBytes)
}

func (store *BillingArchiveStore) readBillingObject(ctx context.Context, key string, size int64, digest string, limit int64) ([]byte, error) {
	if size <= 0 || size > limit {
		return nil, errors.New("invalid billing archive size")
	}
	result, err := store.driver.client.GetObject(ctx, &oss.GetObjectRequest{Bucket: oss.Ptr(store.driver.bucket), Key: oss.Ptr(key)})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	body, err := io.ReadAll(io.LimitReader(result.Body, size+1))
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(body)
	if int64(len(body)) != size || hex.EncodeToString(hash[:]) != digest {
		return nil, errors.New("billing archive integrity check failed")
	}
	return body, nil
}

// Put verifies a read-back before publishing metadata. A retry can reuse the
// exact existing bytes, but can never overwrite an earlier immutable object.
func (store *BillingArchiveStore) Put(ctx context.Context, statementID, kind string, ordinal, userID int, contentType string, body []byte, rows int64) (*model.BillingArtifact, error) {
	if len(body) == 0 || len(body) > 16*1024*1024 {
		return nil, errors.New("billing archive chunk exceeds 16 MiB")
	}
	hash := sha256.Sum256(body)
	digest := hex.EncodeToString(hash[:])
	var artifact model.BillingArtifact
	err := model.DB.Where("statement_id = ? AND kind = ? AND ordinal = ?", statementID, kind, ordinal).First(&artifact).Error
	if err == nil {
		if artifact.SHA256 != digest {
			return nil, model.ErrBillingConflict
		}
		_, err = store.Read(ctx, artifact.ObjectKey, artifact.Size, artifact.SHA256)
		return &artifact, err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	statement, err := model.GetBillingStatement(statementID)
	if err != nil {
		return nil, err
	}
	if userID <= 0 || statement.UserID != userID {
		return nil, errors.New("invalid billing archive owner")
	}
	if _, err := time.Parse("2006-01", statement.Month); err != nil {
		return nil, errors.New("invalid billing archive month")
	}
	key := fmt.Sprintf("billing/user-%s/%s/%s/%s/%06d-%s", permanentStorageUserHash(userID), statement.Month, statementID, kind, ordinal, digest)
	return store.putBillingArtifact(ctx, statementID, kind, ordinal, userID, contentType, body, rows, key, digest, false)
}

// Upstream evidence belongs to internal financial accounting. It has a
// separate namespace and can only be attached to an upstream statement.
func (store *BillingArchiveStore) PutUpstream(ctx context.Context, statementID, kind string, ordinal, ownerID int, contentType string, body []byte, rows int64) (*model.BillingArtifact, error) {
	if len(body) == 0 || len(body) > 16*1024*1024 {
		return nil, errors.New("billing archive chunk exceeds 16 MiB")
	}
	if ordinal < 0 || rows < 0 || statementID == "" || strings.ContainsAny(statementID, "/\\") {
		return nil, errors.New("invalid upstream billing artifact")
	}
	switch kind {
	case "details", "snapshot", "pdf", "xlsx", "manifest", "receipt":
	default:
		return nil, errors.New("invalid upstream billing artifact kind")
	}
	var statement model.UpstreamBillingStatement
	if err := model.DB.First(&statement, "id = ?", statementID).Error; err != nil {
		return nil, err
	}
	if ownerID <= 0 || statement.CreatedBy != ownerID || statement.GroupID <= 0 || statement.StorageProfileID != store.profileID {
		return nil, errors.New("invalid upstream billing archive owner or profile")
	}
	month, err := time.Parse("2006-01", statement.Month)
	if err != nil || month.Format("2006-01") != statement.Month {
		return nil, errors.New("invalid upstream billing archive month")
	}
	hash := sha256.Sum256(body)
	digest := hex.EncodeToString(hash[:])
	var artifact model.BillingArtifact
	err = model.DB.Where("statement_id = ? AND kind = ? AND ordinal = ?", statementID, kind, ordinal).First(&artifact).Error
	if err == nil {
		if artifact.SHA256 != digest || artifact.Size != int64(len(body)) || artifact.Rows != rows {
			return nil, model.ErrBillingConflict
		}
		_, err = store.Read(ctx, artifact.ObjectKey, artifact.Size, artifact.SHA256)
		return &artifact, err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	key := fmt.Sprintf("billing/upstream-%d/%s/%s/%s/%06d-%s", statement.GroupID, statement.Month, statementID, kind, ordinal, digest)
	return store.putBillingArtifact(ctx, statementID, kind, ordinal, ownerID, contentType, body, rows, key, digest, true)
}

// putBillingArtifact registers permanent storage before I/O and verifies the
// complete bytes before publishing immutable artifact metadata. Both upstream
// and customer statements use the same storage identity and retention policy.
func (store *BillingArchiveStore) putBillingArtifact(ctx context.Context, statementID, kind string, ordinal, ownerID int, contentType string, body []byte, rows int64, key, digest string, private bool) (*model.BillingArtifact, error) {
	objectHash := sha256.Sum256([]byte(fmt.Sprintf("%d:%s", store.profileID, key)))
	object := model.StorageObject{ObjectID: hex.EncodeToString(objectHash[:]), OwnerUserID: ownerID, StorageProfileID: store.profileID, Purpose: model.StorageObjectPurposeBillingArchive, ObjectKey: key, ContentType: contentType, Size: int64(len(body)), SHA256: digest, Status: model.StorageObjectStatusUploading}
	// Register the object before I/O to freeze the profile's bucket/endpoint.
	if err := model.RegisterBillingArchiveObject(store.profile, &object); err != nil {
		return nil, err
	}
	object.ID = 0
	if err := model.DB.Where("object_id = ?", object.ObjectID).First(&object).Error; err != nil {
		return nil, err
	}
	var putErr error
	if private {
		// Every upstream artifact is internal financial data, including PDFs and
		// workbooks. Explicit ACLs preserve authorization even in public buckets.
		_, putErr = store.driver.client.PutObject(ctx, &oss.PutObjectRequest{
			Bucket: oss.Ptr(store.driver.bucket), Key: oss.Ptr(key), Body: bytes.NewReader(body),
			ContentLength: oss.Ptr(object.Size), ContentType: oss.Ptr(contentType),
			CacheControl: oss.Ptr("private, no-store"), ForbidOverwrite: oss.Ptr("true"), Acl: oss.ObjectACLPrivate,
		})
	} else {
		_, putErr = store.driver.PutObject(ctx, key, contentType, object.Size, bytes.NewReader(body))
	}
	if _, err := store.Read(ctx, key, object.Size, digest); err != nil {
		if putErr != nil {
			return nil, fmt.Errorf("upload billing archive: %w", putErr)
		}
		return nil, err
	}
	if err := model.DB.Model(&object).Update("status", model.StorageObjectStatusUploaded).Error; err != nil {
		return nil, err
	}
	artifact := model.BillingArtifact{StatementID: statementID, Kind: kind, Ordinal: ordinal, ObjectID: object.ObjectID, ObjectKey: key, SHA256: digest, Size: object.Size, Rows: rows}
	if err := model.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&artifact).Error; err != nil {
		return nil, err
	}
	artifact.ID = 0
	if err := model.DB.Where("statement_id = ? AND kind = ? AND ordinal = ?", statementID, kind, ordinal).First(&artifact).Error; err != nil {
		return nil, err
	}
	if artifact.SHA256 != digest || artifact.ObjectKey != key {
		return nil, model.ErrBillingConflict
	}
	return &artifact, nil
}
