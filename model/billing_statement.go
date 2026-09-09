package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	StatementPreparing = "preparing"
	StatementDraft     = "draft"
	StatementIssued    = "issued"
	StatementDisputed  = "disputed"
	StatementConfirmed = "confirmed"
	StatementVoid      = "void"
	StatementFailed    = "failed"
)

type BillingStatement struct {
	ID               string  `json:"id" gorm:"primaryKey;type:varchar(64)"`
	UserID           int     `json:"user_id" gorm:"uniqueIndex:idx_billing_statement_revision,priority:1;index:idx_billing_statement_list,priority:1"`
	Month            string  `json:"month" gorm:"type:varchar(7);uniqueIndex:idx_billing_statement_revision,priority:2"`
	Revision         int     `json:"revision" gorm:"uniqueIndex:idx_billing_statement_revision,priority:3"`
	ActiveKey        *string `json:"-" gorm:"type:varchar(64);uniqueIndex"`
	Status           string  `json:"status" gorm:"type:varchar(16);index:idx_billing_statement_jobs,priority:1"`
	StartAt          int64   `json:"start_at" gorm:"bigint"`
	EndAt            int64   `json:"end_at" gorm:"bigint"`
	FromSequence     int64   `json:"-" gorm:"bigint"`
	ToSequence       int64   `json:"-" gorm:"bigint"`
	ProfileVersion   int64   `json:"profile_version" gorm:"bigint"`
	Snapshot         string  `json:"snapshot" gorm:"type:text"`
	SnapshotSHA256   string  `json:"snapshot_sha256" gorm:"type:char(64)"`
	ManifestSHA256   string  `json:"manifest_sha256" gorm:"type:char(64)"`
	PDFSHA256        string  `json:"pdf_sha256" gorm:"type:char(64)"`
	StorageProfileID int     `json:"storage_profile_id"`
	CreatedBy        int     `json:"created_by"`
	CreatedAt        int64   `json:"created_at" gorm:"bigint;index:idx_billing_statement_list,priority:2"`
	IssuedBy         int     `json:"issued_by"`
	IssuedAt         int64   `json:"issued_at" gorm:"bigint"`
	DueAt            int64   `json:"due_at" gorm:"bigint"`
	ConfirmedAt      int64   `json:"confirmed_at" gorm:"bigint"`
	LeaseOwner       string  `json:"-" gorm:"type:varchar(64)"`
	LeaseUntil       int64   `json:"-" gorm:"bigint;index:idx_billing_statement_jobs,priority:2"`
	LastError        string  `json:"last_error,omitempty" gorm:"type:text"`
}

type BillingStatementEvent struct {
	ID             int64  `json:"id" gorm:"primaryKey"`
	StatementID    string `json:"statement_id" gorm:"type:varchar(64);index:idx_billing_statement_events,priority:1"`
	ActorID        int    `json:"actor_id"`
	Action         string `json:"action" gorm:"type:varchar(32)"`
	Note           string `json:"note" gorm:"type:text"`
	ManifestSHA256 string `json:"manifest_sha256" gorm:"type:char(64)"`
	PDFSHA256      string `json:"pdf_sha256" gorm:"type:char(64)"`
	SessionID      string `json:"-" gorm:"type:varchar(128)"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint;index:idx_billing_statement_events,priority:2"`
}

// Artifact metadata is permanent, and also registered with StorageObject so
// profile identity cannot be changed underneath archived statements.
type BillingArtifact struct {
	ID          int64  `json:"id" gorm:"primaryKey"`
	StatementID string `json:"statement_id" gorm:"type:varchar(64);uniqueIndex:idx_billing_artifact,priority:1"`
	Kind        string `json:"kind" gorm:"type:varchar(16);uniqueIndex:idx_billing_artifact,priority:2"`
	Ordinal     int    `json:"ordinal" gorm:"uniqueIndex:idx_billing_artifact,priority:3"`
	ObjectID    string `json:"-" gorm:"type:varchar(64)"`
	ObjectKey   string `json:"-" gorm:"type:varchar(1024)"`
	SHA256      string `json:"sha256" gorm:"type:char(64)"`
	Size        int64  `json:"size" gorm:"bigint"`
	Rows        int64  `json:"rows" gorm:"bigint"`
}

func GetBillingStatement(id string) (*BillingStatement, error) {
	var statement BillingStatement
	err := DB.First(&statement, "id = ?", id).Error
	return &statement, err
}

// CompleteBillingStatementArchive publishes only the current lease owner's
// result. Metadata is updated atomically with the state and lease release.
func CompleteBillingStatementArchive(statement *BillingStatement, owner string, succeeded bool) error {
	updates := map[string]interface{}{"lease_until": 0, "lease_owner": ""}
	if succeeded {
		if statement.ManifestSHA256 == "" || statement.PDFSHA256 == "" {
			return ErrBillingEvidenceIntegrity
		}
		updates["status"], updates["last_error"] = StatementDraft, ""
		// Use Go field names so GORM resolves acronym-heavy fields against the
		// same schema used by AutoMigrate, without renaming existing columns.
		updates["ManifestSHA256"], updates["PDFSHA256"] = statement.ManifestSHA256, statement.PDFSHA256
	} else {
		updates["status"], updates["last_error"] = StatementFailed, "Archive preparation failed. Check server logs and retry."
	}
	result := DB.Model(&BillingStatement{}).Where("id = ? AND status = ? AND lease_owner = ?", statement.ID, StatementPreparing, owner).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrBillingConflict
	}
	return nil
}

func ListBillingStatements(userID int, before int64, beforeID, status string, customer bool) ([]BillingStatement, error) {
	statements := make([]BillingStatement, 0)
	query := DB.Where("user_id = ?", userID)
	if customer {
		query = query.Where("issued_at > 0")
	}
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	if before > 0 {
		query = query.Where("created_at < ? OR (created_at = ? AND id < ?)", before, before, beforeID)
	}
	err := query.Order("created_at desc").Order("id desc").Limit(100).Find(&statements).Error
	return statements, err
}

// Caller supplies a snapshot built under the account lock. It contains no
// sensitive request bodies or credentials.
func CreateBillingStatement(statement *BillingStatement, buildSnapshot func(*BillingAccount, []BillingHour) (string, string, error)) error {
	start, end, err := BillingMonthBounds(statement.Month)
	if err != nil {
		return err
	}
	if common.GetTimestamp() < end+86400 {
		return errors.New("statements can be prepared 24 hours after the month closes")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		account, err := lockBillingAccount(tx, statement.UserID)
		if err != nil {
			return err
		}
		if account.AccountingStartAt == 0 {
			return ErrBillingNotConfigured
		}
		if account.AccountingStartAt >= end {
			return errors.New("month precedes the accounting start")
		}
		if account.CompanyTitle == "" || account.TaxID == "" {
			return errors.New("company title and tax ID are required")
		}
		var pending int64
		if err := tx.Model(&BillingOperation{}).Where("user_id = ? AND state = ? AND created_at < ?", statement.UserID, "reserved", end).Count(&pending).Error; err != nil {
			return err
		}
		if pending > 0 {
			return errors.New("unfinished billing operations must be reconciled before preparing a statement")
		}
		if account.AccountingStartAt > start {
			start = account.AccountingStartAt
		}
		statement.StartAt, statement.EndAt = start, end
		statement.FromSequence, statement.ToSequence = account.Sequence+1, account.Sequence
		statement.ProfileVersion = account.ProfileVersion
		hours := make([]BillingHour, 0)
		if err := tx.Where("user_id = ? AND hour >= ? AND hour < ?", statement.UserID, start/3600*3600, end).Order("hour asc").Find(&hours).Error; err != nil {
			return err
		}
		// Summaries cannot define the evidence boundary: a missing first or last
		// hour would otherwise silently remove real entries from the archive.
		// This one-time, indexed customer/month query freezes canonical bounds;
		// the worker independently reconciles every included entry with the hours.
		var bounds struct {
			FirstSequence int64
			LastSequence  int64
		}
		if err := tx.Model(&BillingEntry{}).
			Where("user_id = ? AND posted_at >= ? AND posted_at < ? AND sequence >= ? AND sequence <= ? AND kind <> ?", statement.UserID, start, end, account.StartSequence, account.Sequence, "funding").
			Select("COALESCE(MIN(sequence), 0) AS first_sequence, COALESCE(MAX(sequence), 0) AS last_sequence").Scan(&bounds).Error; err != nil {
			return err
		}
		if bounds.FirstSequence > 0 {
			statement.FromSequence, statement.ToSequence = bounds.FirstSequence, bounds.LastSequence
		}
		for _, hour := range hours {
			if hour.FirstSequence < account.StartSequence || hour.LastSequence < hour.FirstSequence || hour.LastSequence > account.Sequence {
				return errors.New("invalid hourly billing sequence bounds")
			}
		}
		statement.Snapshot, statement.SnapshotSHA256, err = buildSnapshot(account, hours)
		if err != nil {
			return err
		}
		var latest BillingStatement
		err = tx.Where("user_id = ? AND month = ?", statement.UserID, statement.Month).Order("revision desc").First(&latest).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if latest.ID != "" && latest.Status != StatementVoid {
			return errors.New("an active statement already exists for this month")
		}
		statement.Revision = latest.Revision + 1
		key := fmt.Sprintf("%d:%s", statement.UserID, statement.Month)
		statement.ActiveKey = &key
		statement.Status, statement.CreatedAt = StatementPreparing, common.GetTimestamp()
		if err := tx.Create(statement).Error; err != nil {
			return err
		}
		return tx.Create(&BillingStatementEvent{StatementID: statement.ID, ActorID: statement.CreatedBy, Action: "prepare", CreatedAt: statement.CreatedAt}).Error
	})
}

func ChangeBillingStatement(id, action, digest, note, sessionID string, actorID int, admin bool) (*BillingStatement, error) {
	var statement BillingStatement
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := lockForUpdate(tx).First(&statement, "id = ?", id).Error; err != nil {
			return err
		}
		if !admin && statement.UserID != actorID {
			return gorm.ErrRecordNotFound
		}
		if action == "issue" || action == "confirm" {
			if err := statement.VerifySnapshot(); err != nil {
				return err
			}
		}
		now := common.GetTimestamp()
		updates := map[string]interface{}{}
		switch action {
		case "issue":
			if !admin || statement.Status != StatementDraft || statement.ManifestSHA256 == "" || statement.PDFSHA256 == "" {
				return ErrBillingConflict
			}
			updates["status"], updates["issued_at"], updates["issued_by"], updates["due_at"] = StatementIssued, now, actorID, now+7*86400
		case "confirm":
			// The controller additionally requires a live first-party session.
			if statement.UserID != actorID || sessionID == "" || digest == "" || digest != statement.ManifestSHA256 {
				return ErrBillingConflict
			}
			if statement.Status == StatementConfirmed {
				return nil
			}
			if statement.Status != StatementIssued && statement.Status != StatementDisputed {
				return ErrBillingConflict
			}
			updates["status"], updates["confirmed_at"] = StatementConfirmed, now
		case "dispute":
			if statement.UserID != actorID || sessionID == "" || statement.Status != StatementIssued || note == "" {
				return ErrBillingConflict
			}
			updates["status"] = StatementDisputed
		case "reply":
			if !admin || statement.Status != StatementDisputed || note == "" {
				return ErrBillingConflict
			}
		case "void":
			if !admin || statement.Status == StatementConfirmed || statement.Status == StatementVoid || note == "" {
				return ErrBillingConflict
			}
			updates["status"], updates["active_key"] = StatementVoid, nil
		case "retry":
			if !admin || statement.Status != StatementFailed {
				return ErrBillingConflict
			}
			updates["status"], updates["last_error"], updates["lease_until"] = StatementPreparing, "", 0
		default:
			return errors.New("invalid statement action")
		}
		if len([]rune(note)) > 2000 {
			return errors.New("statement note exceeds 2000 characters")
		}
		if err := tx.Model(&statement).Updates(updates).Error; err != nil {
			return err
		}
		event := BillingStatementEvent{StatementID: id, ActorID: actorID, Action: action, Note: note, SessionID: sessionID, CreatedAt: now, ManifestSHA256: statement.ManifestSHA256, PDFSHA256: statement.PDFSHA256}
		return tx.Create(&event).Error
	})
	if err != nil {
		return nil, err
	}
	return GetBillingStatement(id)
}

func GetBillingStatementEntries(statement *BillingStatement, after int64, limit int) ([]BillingEntry, error) {
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}
	entries := make([]BillingEntry, 0)
	err := DB.Where("user_id = ? AND sequence >= ? AND sequence > ? AND sequence <= ? AND posted_at >= ? AND posted_at < ? AND kind <> ?", statement.UserID, statement.FromSequence, after, statement.ToSequence, statement.StartAt, statement.EndAt, "funding").Order("sequence asc").Limit(limit).Find(&entries).Error
	return entries, err
}
