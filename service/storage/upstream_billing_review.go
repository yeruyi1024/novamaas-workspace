package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/QuantumNous/new-api/model"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

// Uploaded supplier evidence follows the same permanent profile protections as
// generated statements, with its own twenty-MiB read/write limit and namespace.
func (store *BillingArchiveStore) PutUpstreamAttachment(ctx context.Context, attachment *model.UpstreamBillingAttachment, body []byte) error {
	if attachment == nil || len(body) == 0 || int64(len(body)) > model.MaxUpstreamBillingAttachmentBytes {
		return model.ErrUpstreamBillingInvalidAttachment
	}
	var reserved model.UpstreamBillingAttachment
	if err := model.DB.Where("id = ?", attachment.ID).First(&reserved).Error; err != nil {
		return err
	}
	if reserved.Status != model.UpstreamBillingAttachmentUploading || reserved.StorageProfileID != store.profileID || reserved.Size != int64(len(body)) || reserved.SHA256 != attachment.SHA256 || reserved.ObjectKey != attachment.ObjectKey || reserved.ObjectID != attachment.ObjectID || reserved.UploadedBy != attachment.UploadedBy {
		return model.ErrBillingConflict
	}
	hash := sha256.Sum256(body)
	if hex.EncodeToString(hash[:]) != reserved.SHA256 {
		return model.ErrBillingEvidenceIntegrity
	}
	object := model.StorageObject{ObjectID: reserved.ObjectID, OwnerUserID: reserved.UploadedBy, StorageProfileID: store.profileID, Purpose: model.StorageObjectPurposeBillingArchive, ObjectKey: reserved.ObjectKey, ContentType: reserved.ContentType, Size: reserved.Size, SHA256: reserved.SHA256, Status: model.StorageObjectStatusUploading}
	if err := model.RegisterBillingArchiveObject(store.profile, &object); err != nil {
		return err
	}
	_, putErr := store.driver.client.PutObject(ctx, &oss.PutObjectRequest{Bucket: oss.Ptr(store.driver.bucket), Key: oss.Ptr(reserved.ObjectKey), Body: bytes.NewReader(body), ContentLength: oss.Ptr(reserved.Size), ContentType: oss.Ptr(reserved.ContentType), CacheControl: oss.Ptr("private, no-store"), ForbidOverwrite: oss.Ptr("true"), Acl: oss.ObjectACLPrivate})
	if _, err := store.ReadUpstreamAttachment(ctx, reserved.ObjectKey, reserved.Size, reserved.SHA256); err != nil {
		if putErr != nil {
			return fmt.Errorf("upload upstream supplier attachment: %w", putErr)
		}
		return err
	}
	return model.DB.Model(&model.StorageObject{}).Where("object_id = ?", reserved.ObjectID).Update("status", model.StorageObjectStatusUploaded).Error
}
