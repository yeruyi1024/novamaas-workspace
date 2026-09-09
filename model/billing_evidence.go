package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"gorm.io/gorm"
)

var ErrBillingEvidenceImmutable = errors.New("billing evidence is append-only")
var ErrBillingEvidenceIntegrity = errors.New("billing snapshot integrity check failed")

func (statement *BillingStatement) VerifySnapshot() error {
	digest := sha256.Sum256([]byte(statement.Snapshot))
	if statement.Snapshot == "" || hex.EncodeToString(digest[:]) != statement.SnapshotSHA256 {
		return ErrBillingEvidenceIntegrity
	}
	return nil
}

func (*BillingEntry) BeforeUpdate(*gorm.DB) error          { return ErrBillingEvidenceImmutable }
func (*BillingEntry) BeforeDelete(*gorm.DB) error          { return ErrBillingEvidenceImmutable }
func (*BillingAccountEvent) BeforeUpdate(*gorm.DB) error   { return ErrBillingEvidenceImmutable }
func (*BillingAccountEvent) BeforeDelete(*gorm.DB) error   { return ErrBillingEvidenceImmutable }
func (*BillingStatementEvent) BeforeUpdate(*gorm.DB) error { return ErrBillingEvidenceImmutable }
func (*BillingStatementEvent) BeforeDelete(*gorm.DB) error { return ErrBillingEvidenceImmutable }
func (*BillingArtifact) BeforeUpdate(*gorm.DB) error       { return ErrBillingEvidenceImmutable }
func (*BillingArtifact) BeforeDelete(*gorm.DB) error       { return ErrBillingEvidenceImmutable }
func (*BillingStatement) BeforeDelete(*gorm.DB) error      { return ErrBillingEvidenceImmutable }
func (*BillingStatement) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("UserID", "Month", "Revision", "StartAt", "EndAt", "FromSequence", "ToSequence", "ProfileVersion", "Snapshot", "SnapshotSHA256", "StorageProfileID", "CreatedBy", "CreatedAt") {
		return ErrBillingEvidenceImmutable
	}
	return nil
}
