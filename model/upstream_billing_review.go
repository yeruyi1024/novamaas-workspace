package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	UpstreamBillingAttachmentUploading       = "uploading"
	UpstreamBillingAttachmentReady           = "ready"
	UpstreamBillingAttachmentFailed          = "failed"
	MaxUpstreamBillingAttachmentBytes  int64 = 20 * 1024 * 1024
	MaxUpstreamBillingAttachments            = 20
)

var (
	ErrUpstreamBillingInvalidAttachment = errors.New("invalid upstream billing attachment")
	ErrUpstreamBillingAttachmentLimit   = errors.New("upstream billing attachment limit reached")
	ErrUpstreamBillingUploadPending     = errors.New("upstream billing attachment upload is pending")
)

// Reservations freeze identity before cloud I/O. Completion changes only the
// status and upload timestamp; the original bytes and metadata stay immutable.
type UpstreamBillingAttachment struct {
	ID               string `json:"id" gorm:"type:varchar(64);primaryKey"`
	StatementID      string `json:"statement_id" gorm:"type:varchar(64);uniqueIndex:idx_upstream_attachment_order,priority:1"`
	Ordinal          int    `json:"ordinal" gorm:"uniqueIndex:idx_upstream_attachment_order,priority:2"`
	OriginalFilename string `json:"original_filename" gorm:"type:varchar(255)"`
	ContentType      string `json:"content_type" gorm:"type:varchar(128)"`
	Size             int64  `json:"size" gorm:"bigint"`
	SHA256           string `json:"sha256" gorm:"type:char(64)"`
	StorageProfileID int    `json:"storage_profile_id"`
	ObjectID         string `json:"-" gorm:"type:varchar(64)"`
	ObjectKey        string `json:"-" gorm:"type:varchar(1024)"`
	UploadedBy       int    `json:"uploaded_by"`
	UploadedAt       int64  `json:"uploaded_at" gorm:"bigint"`
	CreatedAt        int64  `json:"-" gorm:"bigint"`
	Status           string `json:"status" gorm:"type:varchar(16);index"`
}

func (*UpstreamBillingAttachment) BeforeDelete(*gorm.DB) error { return ErrBillingEvidenceImmutable }

func (*UpstreamBillingStatementEvent) BeforeDelete(*gorm.DB) error {
	return ErrBillingEvidenceImmutable
}
func (*UpstreamBillingStatementEvent) BeforeUpdate(*gorm.DB) error {
	return ErrBillingEvidenceImmutable
}

func (*UpstreamBillingStatement) BeforeDelete(*gorm.DB) error { return ErrBillingEvidenceImmutable }

func (statement *UpstreamBillingStatement) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("ID", "GroupID", "Month", "Revision", "StartAt", "EndAt", "GroupVersion", "Snapshot", "SnapshotSHA256", "SourceSHA256", "RecordCount", "StorageProfileID", "CreatedBy", "CreatedAt") {
		return ErrBillingEvidenceImmutable
	}
	if !tx.Statement.Changed("ReviewEvidenceSHA256", "ReviewEvidenceSnapshot", "ConfirmedBy", "ConfirmedAt", "ArchivedAt", "ManifestSHA256", "PDFSHA256", "ExcelSHA256") {
		return nil
	}
	if statement.ID == "" {
		return ErrBillingEvidenceImmutable
	}
	var current UpstreamBillingStatement
	if err := tx.Session(&gorm.Session{NewDB: true}).Select("id", "status", "confirmed_at").Where("id = ?", statement.ID).First(&current).Error; err != nil {
		return err
	}
	if current.ConfirmedAt > 0 {
		return ErrBillingEvidenceImmutable
	}
	if current.Status != StatementPreparing && tx.Statement.Changed("ManifestSHA256", "PDFSHA256", "ExcelSHA256") {
		return ErrBillingEvidenceImmutable
	}
	return nil
}

func (*UpstreamBillingAttachment) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("ID", "StatementID", "Ordinal", "OriginalFilename", "ContentType", "Size", "SHA256", "StorageProfileID", "ObjectID", "ObjectKey", "UploadedBy", "CreatedAt") {
		return ErrBillingEvidenceImmutable
	}
	return nil
}

type UpstreamBillingReviewEvidence struct {
	SchemaVersion  int                         `json:"schema_version"`
	StatementID    string                      `json:"statement_id"`
	ManifestSHA256 string                      `json:"manifest_sha256"`
	Attachments    []UpstreamBillingAttachment `json:"attachments"`
}

type UpstreamBillingReview struct {
	Attachments    []UpstreamBillingAttachment `json:"attachments"`
	EvidenceSHA256 string                      `json:"evidence_sha256"`
	PendingUploads int64                       `json:"pending_uploads"`
}

// Upload reservation and confirmation lock the same statement. A confirmation
// cannot pass a pending upload, and a reservation cannot follow confirmation.
func ReserveUpstreamBillingAttachment(attachment *UpstreamBillingAttachment) error {
	if attachment == nil || attachment.StatementID == "" || attachment.UploadedBy <= 0 || attachment.Size <= 0 || attachment.Size > MaxUpstreamBillingAttachmentBytes || attachment.OriginalFilename == "" || len(attachment.OriginalFilename) > 255 || !utf8.ValidString(attachment.OriginalFilename) || strings.ContainsAny(attachment.OriginalFilename, "/\\") {
		return ErrUpstreamBillingInvalidAttachment
	}
	for _, character := range attachment.OriginalFilename {
		if unicode.IsControl(character) {
			return ErrUpstreamBillingInvalidAttachment
		}
	}
	digest, err := hex.DecodeString(attachment.SHA256)
	if err != nil || len(digest) != sha256.Size || strings.ToLower(attachment.SHA256) != attachment.SHA256 {
		return ErrUpstreamBillingInvalidAttachment
	}
	switch attachment.ContentType {
	case "application/pdf", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "text/csv; charset=utf-8", "image/png", "image/jpeg":
	default:
		return ErrUpstreamBillingInvalidAttachment
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var statement UpstreamBillingStatement
		if err := lockForUpdate(tx).Where("id = ?", attachment.StatementID).First(&statement).Error; err != nil {
			return err
		}
		if statement.Status != StatementDraft {
			return ErrBillingConflict
		}
		if err := statement.VerifySnapshot(); err != nil {
			return err
		}
		if err := expireUpstreamBillingUploads(tx, &statement); err != nil {
			return err
		}
		var active int64
		if err := tx.Model(&UpstreamBillingAttachment{}).Where("statement_id = ? AND status IN ?", statement.ID, []string{UpstreamBillingAttachmentUploading, UpstreamBillingAttachmentReady}).Count(&active).Error; err != nil {
			return err
		}
		if active >= MaxUpstreamBillingAttachments {
			return ErrUpstreamBillingAttachmentLimit
		}
		var ordinal int
		if err := tx.Model(&UpstreamBillingAttachment{}).Where("statement_id = ?", statement.ID).Select("COALESCE(MAX(ordinal), 0)").Scan(&ordinal).Error; err != nil {
			return err
		}
		attachment.ID, attachment.Ordinal = common.GetUUID(), ordinal+1
		attachment.StorageProfileID = statement.StorageProfileID
		attachment.CreatedAt, attachment.UploadedAt = common.GetTimestamp(), 0
		attachment.Status = UpstreamBillingAttachmentUploading
		attachment.ObjectKey = fmt.Sprintf("billing/upstream-%d/%s/%s/attachments/%s/%s", statement.GroupID, statement.Month, statement.ID, attachment.ID, attachment.SHA256)
		objectHash := sha256.Sum256([]byte(fmt.Sprintf("%d:%s", attachment.StorageProfileID, attachment.ObjectKey)))
		attachment.ObjectID = hex.EncodeToString(objectHash[:])
		if err := tx.Create(attachment).Error; err != nil {
			return err
		}
		return tx.Create(&UpstreamBillingStatementEvent{StatementID: statement.ID, ActorID: attachment.UploadedBy, Action: "attachment_upload", Note: attachment.ID + ": " + attachment.OriginalFilename, CreatedAt: attachment.CreatedAt}).Error
	})
}

func CompleteUpstreamBillingAttachment(id string, actorID int, succeeded bool) error {
	if id == "" || actorID <= 0 {
		return ErrUpstreamBillingInvalidAttachment
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var attachment UpstreamBillingAttachment
		if err := tx.Where("id = ?", id).First(&attachment).Error; err != nil {
			return err
		}
		var statement UpstreamBillingStatement
		if err := lockForUpdate(tx).Where("id = ?", attachment.StatementID).First(&statement).Error; err != nil {
			return err
		}
		if statement.Status != StatementDraft || attachment.UploadedBy != actorID {
			return ErrBillingConflict
		}
		if err := tx.Where("id = ?", id).First(&attachment).Error; err != nil {
			return err
		}
		if attachment.Status == UpstreamBillingAttachmentReady && succeeded {
			return nil
		}
		if attachment.Status != UpstreamBillingAttachmentUploading {
			return ErrBillingConflict
		}
		status, action := UpstreamBillingAttachmentFailed, "attachment_failed"
		uploadedAt := int64(0)
		if succeeded {
			var object StorageObject
			if err := tx.Where("object_id = ?", attachment.ObjectID).First(&object).Error; err != nil {
				return err
			}
			if object.OwnerUserID != attachment.UploadedBy || object.StorageProfileID != attachment.StorageProfileID || object.ObjectKey != attachment.ObjectKey || object.SHA256 != attachment.SHA256 || object.Size != attachment.Size || object.Purpose != StorageObjectPurposeBillingArchive || object.Status != StorageObjectStatusUploaded {
				return ErrBillingEvidenceIntegrity
			}
			status, action, uploadedAt = UpstreamBillingAttachmentReady, "attachment_ready", common.GetTimestamp()
		}
		result := tx.Model(&attachment).Where("status = ?", UpstreamBillingAttachmentUploading).Updates(map[string]any{"status": status, "uploaded_at": uploadedAt})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrBillingConflict
		}
		return tx.Create(&UpstreamBillingStatementEvent{StatementID: statement.ID, ActorID: actorID, Action: action, Note: attachment.ID + ": " + attachment.OriginalFilename, CreatedAt: common.GetTimestamp()}).Error
	})
}

func GetUpstreamBillingAttachment(statementID, id string) (*UpstreamBillingAttachment, error) {
	var attachment UpstreamBillingAttachment
	err := DB.Where("statement_id = ? AND id = ? AND status = ?", statementID, id, UpstreamBillingAttachmentReady).First(&attachment).Error
	return &attachment, err
}

func GetUpstreamBillingReview(statementID string) (*UpstreamBillingReview, error) {
	var review *UpstreamBillingReview
	err := DB.Transaction(func(tx *gorm.DB) error {
		var statement UpstreamBillingStatement
		if err := lockForUpdate(tx).Where("id = ?", statementID).First(&statement).Error; err != nil {
			return err
		}
		var err error
		review, _, err = upstreamBillingReviewEvidence(tx, &statement)
		return err
	})
	return review, err
}

func upstreamBillingReviewEvidence(tx *gorm.DB, statement *UpstreamBillingStatement) (*UpstreamBillingReview, []byte, error) {
	if err := expireUpstreamBillingUploads(tx, statement); err != nil {
		return nil, nil, err
	}
	attachments := make([]UpstreamBillingAttachment, 0)
	if err := tx.Where("statement_id = ? AND status = ?", statement.ID, UpstreamBillingAttachmentReady).Order("ordinal asc").Find(&attachments).Error; err != nil {
		return nil, nil, err
	}
	var pending int64
	if err := tx.Model(&UpstreamBillingAttachment{}).Where("statement_id = ? AND status = ?", statement.ID, UpstreamBillingAttachmentUploading).Count(&pending).Error; err != nil {
		return nil, nil, err
	}
	evidence := UpstreamBillingReviewEvidence{SchemaVersion: 1, StatementID: statement.ID, ManifestSHA256: statement.ManifestSHA256, Attachments: attachments}
	body, err := common.Marshal(evidence)
	if err != nil {
		return nil, nil, err
	}
	hash := sha256.Sum256(body)
	digest := hex.EncodeToString(hash[:])
	if statement.ReviewEvidenceSHA256 != "" && (statement.ReviewEvidenceSHA256 != digest || string(statement.ReviewEvidenceSnapshot) != string(body)) {
		return nil, nil, ErrBillingEvidenceIntegrity
	}
	return &UpstreamBillingReview{Attachments: attachments, EvidenceSHA256: digest, PendingUploads: pending}, body, nil
}

// A process can exit after reserving an upload. Expiry releases the pending
// review gate while keeping the reservation, original uploader and audit trail.
func expireUpstreamBillingUploads(tx *gorm.DB, statement *UpstreamBillingStatement) error {
	if statement.Status != StatementDraft {
		return nil
	}
	var expired []UpstreamBillingAttachment
	if err := tx.Where("statement_id = ? AND status = ? AND created_at < ?", statement.ID, UpstreamBillingAttachmentUploading, common.GetTimestamp()-15*60).Order("ordinal asc").Find(&expired).Error; err != nil {
		return err
	}
	for _, attachment := range expired {
		result := tx.Model(&attachment).Where("status = ?", UpstreamBillingAttachmentUploading).Update("status", UpstreamBillingAttachmentFailed)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrBillingConflict
		}
		if err := tx.Create(&UpstreamBillingStatementEvent{StatementID: statement.ID, ActorID: 0, Action: "attachment_expired", Note: fmt.Sprintf("%s: %s; uploader #%d", attachment.ID, attachment.OriginalFilename, attachment.UploadedBy), CreatedAt: common.GetTimestamp()}).Error; err != nil {
			return err
		}
	}
	return nil
}
