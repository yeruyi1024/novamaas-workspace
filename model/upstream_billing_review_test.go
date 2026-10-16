package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func upstreamReviewStatementFixture(t *testing.T) *UpstreamBillingStatement {
	t.Helper()
	upstreamBillingModelFixture(t)
	hash := sha256.Sum256([]byte(`{"review":"fixture"}`))
	digest := hex.EncodeToString(hash[:])
	statement := &UpstreamBillingStatement{ID: common.GetUUID(), GroupID: 7, Month: "2026-09", Status: StatementDraft, StorageProfileID: 1, CreatedBy: 1, Snapshot: `{"review":"fixture"}`, SnapshotSHA256: digest, ManifestSHA256: digest, PDFSHA256: digest, ExcelSHA256: digest}
	require.NoError(t, DB.Create(statement).Error)
	for _, kind := range []string{"snapshot", "pdf", "xlsx", "manifest"} {
		require.NoError(t, DB.Create(&BillingArtifact{StatementID: statement.ID, Kind: kind, SHA256: digest, Size: 10}).Error)
	}
	return statement
}

func TestUpstreamReviewRejectsPendingAndStaleEvidenceThenFreezesAttachments(t *testing.T) {
	statement := upstreamReviewStatementFixture(t)
	initial, err := GetUpstreamBillingReview(statement.ID)
	require.NoError(t, err)
	assert.Len(t, initial.EvidenceSHA256, 64)
	hash := sha256.Sum256([]byte("model,cost\nA,7\n"))
	attachment := &UpstreamBillingAttachment{StatementID: statement.ID, OriginalFilename: "供应商账单.csv", ContentType: "text/csv; charset=utf-8", Size: 15, SHA256: hex.EncodeToString(hash[:]), UploadedBy: 4}
	require.NoError(t, ReserveUpstreamBillingAttachment(attachment))
	_, err = ActUpstreamBillingStatementWithReview(statement.ID, "confirm", statement.ManifestSHA256, initial.EvidenceSHA256, "", "session", 8)
	assert.ErrorIs(t, err, ErrUpstreamBillingUploadPending)
	_, err = ActUpstreamBillingStatement(statement.ID, "void", "", "Upload still running", "session", 8)
	assert.ErrorIs(t, err, ErrUpstreamBillingUploadPending, "void must serialize with the same pending reservation as confirmation")
	current, err := GetUpstreamBillingStatement(statement.ID)
	require.NoError(t, err)
	assert.Equal(t, StatementDraft, current.Status)
	assert.Error(t, CompleteUpstreamBillingAttachment(attachment.ID, 4, true), "unverified storage must not become a ready attachment")
	object := StorageObject{ObjectID: attachment.ObjectID, ObjectKey: attachment.ObjectKey, StorageProfileID: attachment.StorageProfileID, OwnerUserID: 4, SHA256: attachment.SHA256, Size: attachment.Size, Purpose: StorageObjectPurposeBillingArchive, Status: StorageObjectStatusUploaded}
	require.NoError(t, DB.Create(&object).Error)
	require.NoError(t, CompleteUpstreamBillingAttachment(attachment.ID, 4, true))
	review, err := GetUpstreamBillingReview(statement.ID)
	require.NoError(t, err)
	require.Len(t, review.Attachments, 1)
	assert.NotEqual(t, initial.EvidenceSHA256, review.EvidenceSHA256)
	assert.Equal(t, attachment.OriginalFilename, review.Attachments[0].OriginalFilename)
	_, err = ActUpstreamBillingStatementWithReview(statement.ID, "confirm", statement.ManifestSHA256, initial.EvidenceSHA256, "", "session", 8)
	assert.ErrorIs(t, err, ErrBillingConflict)
	_, err = ActUpstreamBillingStatement(statement.ID, "confirm", statement.ManifestSHA256, "", "session", 8)
	assert.ErrorIs(t, err, ErrBillingConflict, "the legacy action cannot bypass evidence confirmation")
	confirmed, err := ActUpstreamBillingStatementWithReview(statement.ID, "confirm", statement.ManifestSHA256, review.EvidenceSHA256, "Checked supplier statement", "session", 8)
	require.NoError(t, err)
	assert.Equal(t, review.EvidenceSHA256, confirmed.ReviewEvidenceSHA256)
	assert.Contains(t, string(confirmed.ReviewEvidenceSnapshot), attachment.OriginalFilename)
	assert.Contains(t, string(confirmed.ReviewEvidenceSnapshot), statement.ManifestSHA256)
	assert.Equal(t, 8, confirmed.ConfirmedBy)
	assert.ErrorIs(t, DB.Model(attachment).Update("original_filename", "replaced.csv").Error, ErrBillingEvidenceImmutable)
	assert.ErrorIs(t, DB.Delete(attachment).Error, ErrBillingEvidenceImmutable)
	assert.ErrorIs(t, DB.Model(confirmed).Update("review_evidence_sha256", strings.Repeat("0", 64)).Error, ErrBillingEvidenceImmutable)
	other := &UpstreamBillingAttachment{StatementID: statement.ID, OriginalFilename: "late.csv", ContentType: attachment.ContentType, Size: attachment.Size, SHA256: attachment.SHA256, UploadedBy: 4}
	assert.ErrorIs(t, ReserveUpstreamBillingAttachment(other), ErrBillingConflict)
	frozen, err := GetUpstreamBillingReview(statement.ID)
	require.NoError(t, err)
	assert.Equal(t, review.EvidenceSHA256, frozen.EvidenceSHA256)
	var event UpstreamBillingStatementEvent
	require.NoError(t, DB.Where("statement_id = ? AND action = ?", statement.ID, "confirm").First(&event).Error)
	assert.Equal(t, review.EvidenceSHA256, event.EvidenceSHA256)
	assert.Equal(t, "session", event.SessionID)
	// Simulate persistence corruption independently of the protected write API.
	require.NoError(t, DB.Exec("UPDATE upstream_billing_attachments SET sha256 = ? WHERE id = ?", strings.Repeat("0", 64), attachment.ID).Error)
	_, err = GetUpstreamBillingReview(statement.ID)
	assert.ErrorIs(t, err, ErrBillingEvidenceIntegrity)
}

func TestUpstreamReviewAllowsAnExplicitlyConfirmedEmptyAttachmentList(t *testing.T) {
	statement := upstreamReviewStatementFixture(t)
	review, err := GetUpstreamBillingReview(statement.ID)
	require.NoError(t, err)
	assert.Empty(t, review.Attachments)
	confirmed, err := ActUpstreamBillingStatementWithReview(statement.ID, "confirm", statement.ManifestSHA256, review.EvidenceSHA256, "", "session", 3)
	require.NoError(t, err)
	assert.Equal(t, StatementConfirmed, confirmed.Status)
	assert.Equal(t, review.EvidenceSHA256, confirmed.ReviewEvidenceSHA256)
	assert.Contains(t, string(confirmed.ReviewEvidenceSnapshot), `"attachments":[]`)
}

func TestUpstreamReviewExpiresCrashedUploadsAndLimitsOnlyActiveAttachments(t *testing.T) {
	statement := upstreamReviewStatementFixture(t)
	digest := strings.Repeat("a", 64)
	expired := &UpstreamBillingAttachment{ID: "expired-upload", StatementID: statement.ID, Ordinal: 1, OriginalFilename: "expired.csv", ContentType: "text/csv; charset=utf-8", Size: 1, SHA256: digest, Status: UpstreamBillingAttachmentUploading, UploadedBy: 5, CreatedAt: common.GetTimestamp() - 901}
	require.NoError(t, DB.Create(expired).Error)
	review, err := GetUpstreamBillingReview(statement.ID)
	require.NoError(t, err)
	assert.Zero(t, review.PendingUploads)
	var event UpstreamBillingStatementEvent
	require.NoError(t, DB.Where("statement_id = ? AND action = ?", statement.ID, "attachment_expired").First(&event).Error)
	assert.Zero(t, event.ActorID)
	assert.Contains(t, event.Note, "expired-upload")
	assert.Contains(t, event.Note, "uploader #5")
	var active []*UpstreamBillingAttachment
	for index := 0; index < MaxUpstreamBillingAttachments; index++ {
		attachment := &UpstreamBillingAttachment{StatementID: statement.ID, OriginalFilename: fmt.Sprintf("supplier-%d.csv", index), ContentType: "text/csv; charset=utf-8", Size: 1, SHA256: digest, UploadedBy: 5}
		require.NoError(t, ReserveUpstreamBillingAttachment(attachment))
		active = append(active, attachment)
	}
	extra := &UpstreamBillingAttachment{StatementID: statement.ID, OriginalFilename: "extra.csv", ContentType: "text/csv; charset=utf-8", Size: 1, SHA256: digest, UploadedBy: 5}
	assert.ErrorIs(t, ReserveUpstreamBillingAttachment(extra), ErrUpstreamBillingAttachmentLimit)
	require.NoError(t, CompleteUpstreamBillingAttachment(active[0].ID, 5, false))
	require.NoError(t, ReserveUpstreamBillingAttachment(extra), "failed attempts must not consume the twenty-file allowance")
}
