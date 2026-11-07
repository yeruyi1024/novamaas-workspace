package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BillingArchiveWriter interface {
	Put(context.Context, string, string, int, int, string, []byte, int64) (*model.BillingArtifact, error)
}

type BillingArchiveStore interface {
	BillingArchiveWriter
	Existing(context.Context, string, string, int) (*model.BillingArtifact, error)
}
type BillingManifest struct {
	AccountingBasis   string                  `json:"accounting_basis,omitempty"`
	ExcelSHA256       string                  `json:"excel_sha256,omitempty"`
	SourceStatementID string                  `json:"source_statement_id,omitempty"`
	Currency          BillingCurrency         `json:"currency"`
	SchemaVersion     int                     `json:"schema_version"`
	StatementID       string                  `json:"statement_id"`
	UserID            int                     `json:"user_id"`
	Month             string                  `json:"month"`
	Revision          int                     `json:"revision"`
	Timezone          string                  `json:"timezone"`
	StartAt           int64                   `json:"start_at"`
	EndAt             int64                   `json:"end_at"`
	FromSequence      int64                   `json:"from_sequence"`
	ToSequence        int64                   `json:"to_sequence"`
	SnapshotSHA256    string                  `json:"snapshot_sha256"`
	PDFSHA256         string                  `json:"pdf_sha256"`
	Rows              int64                   `json:"rows"`
	ChargeQuota       int64                   `json:"charge_quota"`
	RefundQuota       int64                   `json:"refund_quota"`
	Chunks            []model.BillingArtifact `json:"chunks"`
}

var billingWorkerOnce sync.Once

func StartBillingStatementWorker() {
	billingWorkerOnce.Do(func() {
		gopool.Go(func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				if err := model.FlushBillingCorrectionAudits(); err != nil {
					common.SysError("billing correction audit outbox: " + err.Error())
				}
				if err := runBillingArchivePass(); err != nil {
					common.SysError("billing archive worker: " + err.Error())
				}
				<-ticker.C
			}
		})
	})
}
func runBillingArchivePass() error {
	now, owner := common.GetTimestamp(), common.GetUUID()
	var statement model.BillingStatement
	err := model.DB.Where("status = ? AND lease_until < ?", model.StatementPreparing, now).Order("created_at asc").First(&statement).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	claim := model.DB.Model(&statement).Where("status = ? AND lease_until < ?", model.StatementPreparing, now).Updates(map[string]interface{}{"lease_owner": owner, "lease_until": now + 300})
	if claim.Error != nil {
		return claim.Error
	}
	if claim.RowsAffected != 1 {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	defer close(done)
	// The lease is renewed independently of slow cloud I/O. Losing it cancels
	// the upload, and the final CAS prevents a stale worker publishing a draft.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				result := model.DB.Model(&model.BillingStatement{}).Where("id = ? AND status = ? AND lease_owner = ?", statement.ID, model.StatementPreparing, owner).Update("lease_until", common.GetTimestamp()+300)
				if result.Error != nil || result.RowsAffected != 1 {
					cancel()
					return
				}
			}
		}
	}()
	store, err := storageService.OpenBillingArchiveStore(statement.StorageProfileID, true)
	if err == nil {
		err = BuildBillingArchive(ctx, &statement, store)
	}
	if err != nil {
		common.SysError(fmt.Sprintf("billing statement %s archive failed: %v", statement.ID, err))
	}
	return model.CompleteBillingStatementArchive(&statement, owner, err == nil)
}

// BuildBillingArchive reads bounded, immutable per-account sequence pages.
// It verifies the detail totals against the independently maintained snapshot.
func BuildBillingArchive(ctx context.Context, statement *model.BillingStatement, store BillingArchiveStore) error {
	var snapshot BillingSnapshot
	if err := common.UnmarshalJsonStr(statement.Snapshot, &snapshot); err != nil {
		return err
	}
	if err := statement.VerifySnapshot(); err != nil {
		return err
	}
	manifest := BillingManifest{SchemaVersion: 1, StatementID: statement.ID, UserID: statement.UserID, Month: statement.Month, Revision: statement.Revision, SnapshotSHA256: statement.SnapshotSHA256, Chunks: make([]model.BillingArtifact, 0)}
	manifest.Currency = snapshot.Currency
	manifest.AccountingBasis = snapshot.AccountingBasis
	manifest.Timezone, manifest.StartAt, manifest.EndAt = snapshot.Timezone, statement.StartAt, statement.EndAt
	manifest.FromSequence, manifest.ToSequence = statement.FromSequence, statement.ToSequence
	dayTotals := make(map[string]BillingRow, len(snapshot.Days))
	modelTotals := make(map[string]model.BillingModelTotal, len(snapshot.Models))
	modelDays := make(map[string]map[string]struct{}, len(snapshot.Models))
	rate, err := decimal.NewFromString(snapshot.Currency.Rate)
	if err != nil || !rate.IsPositive() {
		return errors.New("invalid snapshot currency rate")
	}
	unit, err := decimal.NewFromString(snapshot.Currency.QuotaPerUnit)
	if err != nil || !unit.IsPositive() {
		return errors.New("invalid snapshot quota unit")
	}
	var excelDetails []BillingExcelDetail
	cursor := int64(0)
	for ordinal := 0; ; ordinal++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		entries, err := model.GetBillingStatementEntries(statement, cursor, 1000)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			break
		}
		var chunk bytes.Buffer
		compressed := gzip.NewWriter(&chunk)
		for _, entry := range entries {
			quota := entry.StatementAmountQuota()
			if quota > int64(common.MaxWalletQuota) || quota < -int64(common.MaxWalletQuota) {
				return errors.New("invalid archived quota")
			}
			if entry.Quota >= 0 {
				if quota > int64(common.MaxWalletQuota)-manifest.ChargeQuota {
					return errors.New("archive total overflow")
				}
				manifest.ChargeQuota += quota
			} else {
				if -quota > int64(common.MaxWalletQuota)-manifest.RefundQuota {
					return errors.New("archive refund overflow")
				}
				manifest.RefundQuota -= quota
			}
			day := time.Unix(entry.PostedAt, 0).In(billingLocation).Format("2006-01-02")
			total := dayTotals[day]
			if entry.Quota >= 0 {
				total.ChargeQuota += quota
			} else {
				total.RefundQuota -= quota
			}
			total.Count++
			dayTotals[day] = total
			modelTotal := modelTotals[entry.ModelName]
			modelTotal.ModelName = entry.ModelName
			if modelTotal.FirstPosted == 0 || entry.PostedAt < modelTotal.FirstPosted {
				modelTotal.FirstPosted = entry.PostedAt
			}
			if entry.PostedAt > modelTotal.LastPosted {
				modelTotal.LastPosted = entry.PostedAt
			}
			days := modelDays[entry.ModelName]
			if days == nil {
				days = make(map[string]struct{}, 1)
				modelDays[entry.ModelName] = days
			}
			days[day] = struct{}{}
			modelTotal.ActiveDays = int64(len(days))
			if entry.Quota >= 0 {
				modelTotal.Charge += quota
				modelTotal.ChargeCount++
			} else {
				modelTotal.Refund -= quota
				modelTotal.RefundCount++
			}
			modelTotal.Count++
			modelTotals[entry.ModelName] = modelTotal
			// Customer-visible immutable details intentionally exclude balances,
			// operator IDs, request bodies, and internal funding event keys.
			record := struct {
				Sequence      int64                              `json:"sequence"`
				PostedAt      int64                              `json:"posted_at"`
				Kind          string                             `json:"kind"`
				Quota         int64                              `json:"quota"`
				Amount        string                             `json:"amount"`
				RequestID     string                             `json:"request_id"`
				ModelName     string                             `json:"model_name"`
				ImportID      string                             `json:"import_id,omitempty"`
				SourceLogID   int                                `json:"source_log_id,omitempty"`
				SourceSHA256  string                             `json:"source_sha256,omitempty"`
				OriginalQuota *int64                             `json:"original_quota,omitempty"`
				Corrections   []model.BillingStatementCorrection `json:"corrections,omitempty"`
			}{entry.Sequence, entry.PostedAt, entry.Kind, quota, decimal.NewFromInt(quota).Mul(rate).Div(unit).StringFixed(6), entry.RequestID, entry.ModelName, entry.ImportID, entry.SourceLogID, entry.SourceSHA256, nil, entry.StatementCorrections}
			if entry.StatementQuota != nil {
				originalQuota := entry.Quota
				record.OriginalQuota = &originalQuota
			}
			if snapshot.DetailSheet {
				excelDetails = append(excelDetails, BillingExcelDetail{PostedAt: entry.PostedAt, Kind: entry.Kind, SourceLogID: entry.SourceLogID, EventKey: entry.EventKey, Amount: decimal.NewFromInt(quota).Mul(rate).Div(unit).String(), RequestID: entry.RequestID, ModelName: entry.ModelName, TokenID: entry.TokenID, BillingGroup: entry.BillingGroup, BillingRate: entry.BillingRate, Refund: entry.Quota < 0})
			}
			body, err := common.Marshal(record)
			if err != nil {
				_ = compressed.Close()
				return err
			}
			if _, err := compressed.Write(append(body, '\n')); err != nil {
				_ = compressed.Close()
				return err
			}
			manifest.Rows++
		}
		if err := compressed.Close(); err != nil {
			return err
		}
		artifact, err := store.Put(ctx, statement.ID, "details", ordinal, statement.UserID, "application/gzip", chunk.Bytes(), int64(len(entries)))
		if err != nil {
			return err
		}
		// Object metadata IDs are local bookkeeping, not part of canonical content.
		manifest.Chunks = append(manifest.Chunks, model.BillingArtifact{Kind: "details", Ordinal: ordinal, SHA256: artifact.SHA256, Size: artifact.Size, Rows: artifact.Rows})
		cursor = entries[len(entries)-1].Sequence
	}
	if manifest.Rows != snapshot.Total.Count || manifest.ChargeQuota != snapshot.ChargeQuota || manifest.RefundQuota != snapshot.RefundQuota {
		return errors.New("statement details do not reconcile with hourly snapshot")
	}
	for _, day := range snapshot.Days {
		total := dayTotals[day.Label]
		if total.ChargeQuota != day.ChargeQuota || total.RefundQuota != day.RefundQuota || total.Count != day.Count {
			return errors.New("statement daily details do not reconcile with snapshot")
		}
		delete(dayTotals, day.Label)
	}
	if len(dayTotals) > 0 {
		return errors.New("statement contains details outside the frozen calendar")
	}
	rowsToVerify := snapshot.Models
	if snapshot.PDFTemplateVersion >= 10 {
		// Split rates are frozen in the snapshot; reconcile their combined model
		// amounts against the ledger without consulting mutable logs on retries.
		byModel := make(map[string]BillingModelRow)
		for _, row := range snapshot.Models {
			total, ok := modelTotals[row.ModelName]
			if !ok || row.ActiveDays <= 0 || row.ActiveDays > total.ActiveDays || row.FirstPosted < total.FirstPosted || row.LastPosted > total.LastPosted || row.LastPosted < row.FirstPosted || row.ChargeQuota < 0 || row.RefundQuota < 0 || row.Count < 0 || row.ChargeCount < 0 || row.RefundCount < 0 || row.ChargeCount > row.Count || row.RefundCount != row.Count-row.ChargeCount {
				return errors.New("statement model rate details do not reconcile with snapshot")
			}
			sum := byModel[row.ModelName]
			sum.ModelName = row.ModelName
			if row.ChargeQuota > total.Charge-sum.ChargeQuota || row.RefundQuota > total.Refund-sum.RefundQuota || row.Count > total.Count-sum.Count {
				return errors.New("statement model rate totals exceed ledger amounts")
			}
			sum.ChargeQuota += row.ChargeQuota
			sum.RefundQuota += row.RefundQuota
			sum.Count += row.Count
			sum.ChargeCount += row.ChargeCount
			sum.RefundCount += row.RefundCount
			if sum.FirstPosted == 0 || row.FirstPosted < sum.FirstPosted {
				sum.FirstPosted = row.FirstPosted
			}
			if row.LastPosted > sum.LastPosted {
				sum.LastPosted = row.LastPosted
			}
			sum.ActiveDays = total.ActiveDays
			byModel[row.ModelName] = sum
		}
		rowsToVerify = make([]BillingModelRow, 0, len(byModel))
		for _, row := range byModel {
			rowsToVerify = append(rowsToVerify, row)
		}
	}
	if snapshot.PDFTemplateVersion >= 7 {
		for _, row := range rowsToVerify {
			total, ok := modelTotals[row.ModelName]
			if !ok || total.Charge != row.ChargeQuota || total.Refund != row.RefundQuota || total.Count != row.Count ||
				(snapshot.PDFTemplateVersion >= 8 && (total.ChargeCount != row.ChargeCount || total.RefundCount != row.RefundCount ||
					total.ActiveDays != row.ActiveDays || total.FirstPosted != row.FirstPosted || total.LastPosted != row.LastPosted)) {
				return errors.New("statement model details do not reconcile with snapshot")
			}
			delete(modelTotals, row.ModelName)
		}
		if len(modelTotals) > 0 {
			return errors.New("statement contains models outside the frozen summary")
		}
	}
	if _, err := store.Put(ctx, statement.ID, "snapshot", 0, statement.UserID, "application/json", []byte(statement.Snapshot), 0); err != nil {
		return err
	}
	if snapshot.PDFTemplateVersion >= 9 {
		artifact, err := store.Existing(ctx, statement.ID, "xlsx", 0)
		if err != nil {
			return err
		}
		if artifact == nil {
			if snapshot.DetailSheet {
				if err := enrichBillingExcelDetails(statement.UserID, excelDetails, rate, unit); err != nil {
					return err
				}
			}
			workbook, err := RenderBillingStatementExcelWithDetails(statement, &snapshot, excelDetails)
			if err != nil {
				return err
			}
			artifact, err = store.Put(ctx, statement.ID, "xlsx", 0, statement.UserID, billingExcelContentType, workbook, 0)
			if err != nil {
				return err
			}
		}
		manifest.SchemaVersion = 2
		if snapshot.AccountingBasis != "" {
			manifest.SchemaVersion = 3
		}
		manifest.SourceStatementID = statement.SourceStatementID
		manifest.ExcelSHA256, statement.ExcelSHA256 = artifact.SHA256, artifact.SHA256
	}
	pdf, err := RenderBillingStatementPDF(statement, &snapshot, false)
	if err != nil {
		return err
	}
	artifact, err := store.Put(ctx, statement.ID, "pdf", 0, statement.UserID, "application/pdf", pdf, 0)
	if err != nil {
		return err
	}
	manifest.PDFSHA256, statement.PDFSHA256 = artifact.SHA256, artifact.SHA256
	body, err := common.Marshal(manifest)
	if err != nil {
		return err
	}
	artifact, err = store.Put(ctx, statement.ID, "manifest", 0, statement.UserID, "application/json", body, manifest.Rows)
	if err != nil {
		return err
	}
	statement.ManifestSHA256 = artifact.SHA256
	return nil
}
