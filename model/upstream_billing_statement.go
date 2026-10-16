package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

func (statement *UpstreamBillingStatement) VerifySnapshot() error {
	if statement == nil || statement.Snapshot == "" {
		return ErrBillingEvidenceIntegrity
	}
	digest := sha256.Sum256([]byte(statement.Snapshot))
	if hex.EncodeToString(digest[:]) != statement.SnapshotSHA256 {
		return ErrBillingEvidenceIntegrity
	}
	return nil
}

func GetUpstreamBillingStatement(id string) (*UpstreamBillingStatement, error) {
	var statement UpstreamBillingStatement
	err := DB.Where("id = ?", id).First(&statement).Error
	if err == nil {
		err = statement.VerifySnapshot()
	}
	return &statement, err
}

func ListUpstreamBillingStatements(groupID int, before int64, beforeID, status string) ([]UpstreamBillingStatement, error) {
	query := DB.Model(&UpstreamBillingStatement{})
	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if before > 0 {
		query = query.Where("created_at < ? OR (created_at = ? AND id < ?)", before, before, beforeID)
	}
	statements := make([]UpstreamBillingStatement, 0)
	err := query.Order("created_at desc").Order("id desc").Limit(51).Find(&statements).Error
	return statements, err
}

// Freeze summary and detail rows together. Membership has an optimistic version
// and a group lock serializes both revisions and active-month uniqueness.
func CreateUpstreamBillingStatement(ctx context.Context, statement *UpstreamBillingStatement, records []UpstreamBillingStatementRecord) error {
	if statement == nil || statement.ID == "" || statement.GroupID <= 0 || statement.CreatedBy <= 0 || statement.StorageProfileID <= 0 {
		return errors.New("invalid upstream statement")
	}
	if err := statement.VerifySnapshot(); err != nil {
		return err
	}
	start, end, err := BillingMonthBounds(statement.Month)
	if err != nil || statement.StartAt != start || statement.EndAt != end {
		return errors.New("invalid upstream statement month")
	}
	return DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var group UpstreamBillingGroup
		if err := lockForUpdate(tx).First(&group, statement.GroupID).Error; err != nil {
			return err
		}
		if group.Version != statement.GroupVersion {
			return ErrBillingConflict
		}
		var frozen struct {
			Group UpstreamBillingGroupView `json:"group"`
		}
		if err := common.UnmarshalJsonStr(string(statement.Snapshot), &frozen); err != nil {
			return ErrBillingEvidenceIntegrity
		}
		for _, channel := range frozen.Group.Channels {
			var current Channel
			err := lockForUpdate(tx).Select("id", "cost_discount").First(&current, channel.ID).Error
			if errors.Is(err, gorm.ErrRecordNotFound) && channel.Missing {
				continue
			}
			if err != nil || channel.Missing {
				return ErrBillingConflict
			}
			discount, err := NormalizeCostDiscount(current.CostDiscount)
			if err != nil || discount != channel.CostDiscount {
				return ErrBillingConflict
			}
		}
		key := fmt.Sprintf("%d:%s", group.ID, statement.Month)
		var active int64
		if err := tx.Model(&UpstreamBillingStatement{}).Where("active_key = ?", key).Count(&active).Error; err != nil {
			return err
		}
		if active != 0 {
			return ErrBillingConflict
		}
		var revision int
		if err := tx.Model(&UpstreamBillingStatement{}).Where("group_id = ? AND month = ?", group.ID, statement.Month).Select("COALESCE(MAX(revision), 0)").Scan(&revision).Error; err != nil {
			return err
		}
		statement.Revision, statement.ActiveKey = revision+1, &key
		statement.Status, statement.CreatedAt = StatementPreparing, common.GetTimestamp()
		if err := tx.Create(statement).Error; err != nil {
			return err
		}
		for i := range records {
			if records[i].Sequence != int64(i+1) || records[i].OccurredAt < start || records[i].OccurredAt >= end {
				return ErrBillingEvidenceIntegrity
			}
			records[i].ID, records[i].StatementID = 0, statement.ID
			if records[i].CreatedAt == 0 {
				records[i].CreatedAt = statement.CreatedAt
			}
		}
		if len(records) > 0 {
			if statement.RecordCount != int64(len(records)) {
				return ErrBillingEvidenceIntegrity
			}
			if err := tx.CreateInBatches(records, 500).Error; err != nil {
				return err
			}
		}
		var count, invalid int64
		if err := tx.Model(&UpstreamBillingStatementRecord{}).Where("statement_id = ?", statement.ID).Count(&count).Error; err != nil {
			return err
		}
		if count != statement.RecordCount {
			return ErrBillingEvidenceIntegrity
		}
		if err := tx.Model(&UpstreamBillingStatementRecord{}).Where("statement_id = ? AND (occurred_at < ? OR occurred_at >= ?)", statement.ID, start, end).Count(&invalid).Error; err != nil {
			return err
		}
		if invalid != 0 {
			return ErrBillingEvidenceIntegrity
		}
		return tx.Create(&UpstreamBillingStatementEvent{StatementID: statement.ID, ActorID: statement.CreatedBy, Action: "create", CreatedAt: statement.CreatedAt}).Error
	})
}

func GetUpstreamBillingStatementRecords(statement *UpstreamBillingStatement, after int64, limit int) ([]UpstreamBillingStatementRecord, error) {
	if statement == nil || statement.ID == "" || after < 0 || limit <= 0 || limit > 1000 {
		return nil, errors.New("invalid upstream statement record page")
	}
	records := make([]UpstreamBillingStatementRecord, 0)
	err := DB.Where("statement_id = ? AND sequence > ?", statement.ID, after).Order("sequence asc").Limit(limit).Find(&records).Error
	return records, err
}

func CompleteUpstreamBillingStatementArchive(statement *UpstreamBillingStatement, owner string, succeeded bool) error {
	updates := map[string]any{"lease_until": 0, "lease_owner": ""}
	if succeeded {
		if err := statement.VerifySnapshot(); err != nil {
			return err
		}
		if len(statement.ManifestSHA256) != 64 || len(statement.PDFSHA256) != 64 || len(statement.ExcelSHA256) != 64 {
			return ErrBillingEvidenceIntegrity
		}
		updates["status"], updates["last_error"] = StatementDraft, ""
		updates["ManifestSHA256"], updates["PDFSHA256"], updates["ExcelSHA256"] = statement.ManifestSHA256, statement.PDFSHA256, statement.ExcelSHA256
	} else {
		updates["status"], updates["last_error"] = StatementFailed, "Archive preparation failed. Check server logs and retry."
	}
	result := DB.Model(statement).Where("id = ? AND status = ? AND lease_owner = ?", statement.ID, StatementPreparing, owner).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrBillingConflict
	}
	return nil
}

func ActUpstreamBillingStatement(id, action, digest, note, sessionID string, actorID int) (*UpstreamBillingStatement, error) {
	return ActUpstreamBillingStatementWithReview(id, action, digest, "", note, sessionID, actorID)
}

func ActUpstreamBillingStatementWithReview(id, action, digest, evidenceDigest, note, sessionID string, actorID int) (*UpstreamBillingStatement, error) {
	note = strings.TrimSpace(note)
	if actorID <= 0 || len(note) > 2000 {
		return nil, errors.New("invalid upstream statement action")
	}
	var statement UpstreamBillingStatement
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := lockForUpdate(tx).Where("id = ?", id).First(&statement).Error; err != nil {
			return err
		}
		if err := statement.VerifySnapshot(); err != nil {
			return err
		}
		now := common.GetTimestamp()
		updates := make(map[string]any)
		switch action {
		case "confirm":
			if sessionID == "" || len(sessionID) > 128 || digest == "" || digest != statement.ManifestSHA256 || len(evidenceDigest) != 64 {
				return ErrBillingConflict
			}
			if statement.Status != StatementDraft && statement.Status != StatementConfirmed {
				return ErrBillingConflict
			}
			review, evidence, err := upstreamBillingReviewEvidence(tx, &statement)
			if err != nil {
				return err
			}
			if review.PendingUploads > 0 {
				return ErrUpstreamBillingUploadPending
			}
			if review.EvidenceSHA256 != evidenceDigest {
				return ErrBillingConflict
			}
			if statement.Status == StatementConfirmed {
				return nil
			}
			for _, kind := range []string{"snapshot", "pdf", "xlsx", "manifest"} {
				var artifact BillingArtifact
				if err := tx.Where("statement_id = ? AND kind = ? AND ordinal = 0", id, kind).First(&artifact).Error; err != nil {
					return err
				}
				want := statement.SnapshotSHA256
				switch kind {
				case "pdf":
					want = statement.PDFSHA256
				case "xlsx":
					want = statement.ExcelSHA256
				case "manifest":
					want = statement.ManifestSHA256
				}
				if artifact.SHA256 != want || artifact.Size <= 0 {
					return ErrBillingEvidenceIntegrity
				}
			}
			updates["status"], updates["confirmed_by"], updates["confirmed_at"], updates["archived_at"] = StatementConfirmed, actorID, now, now
			updates["ReviewEvidenceSHA256"], updates["ReviewEvidenceSnapshot"] = evidenceDigest, UpstreamBillingSnapshotData(evidence)
		case "void":
			if note == "" {
				return errors.New("a void reason is required")
			}
			if statement.Status != StatementDraft && statement.Status != StatementFailed {
				return ErrBillingConflict
			}
			review, _, err := upstreamBillingReviewEvidence(tx, &statement)
			if err != nil {
				return err
			}
			if review.PendingUploads > 0 {
				return ErrUpstreamBillingUploadPending
			}
			updates["status"], updates["active_key"] = StatementVoid, nil
		case "retry":
			if statement.Status != StatementFailed {
				return ErrBillingConflict
			}
			updates["status"], updates["last_error"], updates["lease_until"] = StatementPreparing, "", 0
		default:
			return errors.New("invalid upstream statement action")
		}
		if err := tx.Model(&statement).Updates(updates).Error; err != nil {
			return err
		}
		return tx.Create(&UpstreamBillingStatementEvent{StatementID: id, ActorID: actorID, Action: action, Note: note, ManifestSHA256: digest, EvidenceSHA256: evidenceDigest, SessionID: sessionID, CreatedAt: now}).Error
	})
	if err != nil {
		return nil, err
	}
	return GetUpstreamBillingStatement(id)
}
