package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UpstreamBillingArchiveWriter interface {
	PutUpstream(context.Context, string, string, int, int, string, []byte, int64) (*model.BillingArtifact, error)
}

type UpstreamBillingManifest struct {
	SchemaVersion   int                      `json:"schema_version"`
	AccountingBasis string                   `json:"accounting_basis"`
	StatementID     string                   `json:"statement_id"`
	GroupID         int                      `json:"group_id"`
	GroupVersion    int64                    `json:"group_version"`
	Month           string                   `json:"month"`
	Revision        int                      `json:"revision"`
	Timezone        string                   `json:"timezone"`
	StartAt         int64                    `json:"start_at"`
	EndAt           int64                    `json:"end_at"`
	Currency        BillingCurrency          `json:"currency"`
	Rows            int64                    `json:"rows"`
	ChargeQuota     int64                    `json:"charge_quota"`
	RefundQuota     int64                    `json:"refund_quota"`
	SourceSHA256    string                   `json:"source_sha256"`
	SnapshotSHA256  string                   `json:"snapshot_sha256"`
	PDFSHA256       string                   `json:"pdf_sha256"`
	ExcelSHA256     string                   `json:"excel_sha256"`
	Chunks          []model.BillingArtifact  `json:"chunks"`
	Coverage        *UpstreamBillingCoverage `json:"coverage,omitempty"`
}

var upstreamBillingWorkerOnce sync.Once

func StartUpstreamBillingStatementWorker() {
	upstreamBillingWorkerOnce.Do(func() {
		gopool.Go(func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			var lastStagingCleanup time.Time
			for {
				if time.Since(lastStagingCleanup) >= time.Hour {
					lastStagingCleanup = time.Now()
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					_, err := model.PurgeUnpublishedUpstreamBillingRecords(ctx, common.GetTimestamp()-86400, 1000)
					cancel()
					if err != nil {
						common.SysError("upstream billing staging cleanup: " + err.Error())
					}
				}
				if err := runUpstreamBillingArchivePass(); err != nil {
					common.SysError("upstream billing archive worker: " + err.Error())
				}
				<-ticker.C
			}
		})
	})
}

func runUpstreamBillingArchivePass() error {
	now, owner := common.GetTimestamp(), common.GetUUID()
	var statement model.UpstreamBillingStatement
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
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				result := model.DB.Model(&model.UpstreamBillingStatement{}).Where("id = ? AND status = ? AND lease_owner = ?", statement.ID, model.StatementPreparing, owner).Update("lease_until", common.GetTimestamp()+300)
				if result.Error != nil || result.RowsAffected != 1 {
					cancel()
					return
				}
			}
		}
	}()
	store, err := storageService.OpenBillingArchiveStore(statement.StorageProfileID, true)
	if err == nil {
		err = BuildUpstreamBillingArchive(ctx, &statement, store)
	}
	if err != nil {
		common.SysError(fmt.Sprintf("upstream billing statement %s archive failed: %v", statement.ID, err))
	}
	return model.CompleteUpstreamBillingStatementArchive(&statement, owner, err == nil)
}

type upstreamArchiveModelKey struct{ model, upstream, group string }

// Details are verified against the source digest and independent calendar and
// model totals. Archive retries never consult current channels, prices or logs.
func BuildUpstreamBillingArchive(ctx context.Context, statement *model.UpstreamBillingStatement, store UpstreamBillingArchiveWriter) error {
	if err := statement.VerifySnapshot(); err != nil {
		return err
	}
	var snapshot UpstreamBillingSnapshot
	if err := common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot); err != nil {
		return err
	}
	if (snapshot.SchemaVersion != 1 && snapshot.SchemaVersion != 2) || snapshot.Group.ID != statement.GroupID || snapshot.Group.Version != statement.GroupVersion || snapshot.Month != statement.Month || snapshot.StartAt != statement.StartAt || snapshot.EndAt != statement.EndAt || snapshot.SourceSHA256 != statement.SourceSHA256 {
		return model.ErrBillingEvidenceIntegrity
	}
	rate, rateErr := decimal.NewFromString(snapshot.Currency.Rate)
	unit, unitErr := decimal.NewFromString(snapshot.Currency.QuotaPerUnit)
	if rateErr != nil || unitErr != nil || !rate.IsPositive() || !unit.IsPositive() {
		return model.ErrBillingEvidenceIntegrity
	}
	manifest := UpstreamBillingManifest{SchemaVersion: snapshot.SchemaVersion, AccountingBasis: snapshot.AccountingBasis, StatementID: statement.ID, GroupID: statement.GroupID, GroupVersion: statement.GroupVersion, Month: statement.Month, Revision: statement.Revision, Timezone: snapshot.Timezone, StartAt: statement.StartAt, EndAt: statement.EndAt, Currency: snapshot.Currency, SourceSHA256: statement.SourceSHA256, SnapshotSHA256: statement.SnapshotSHA256, Chunks: make([]model.BillingArtifact, 0)}
	if !snapshot.Coverage.Complete {
		manifest.Coverage = &snapshot.Coverage
	}
	total := UpstreamBillingRow{}
	days := map[string]UpstreamBillingRow{}
	models := map[upstreamArchiveModelKey]UpstreamBillingRow{}
	modelDates := map[upstreamArchiveModelKey]map[string]bool{}
	modelFirst := map[upstreamArchiveModelKey]int64{}
	modelLast := map[upstreamArchiveModelKey]int64{}
	dayModels := map[string]map[upstreamArchiveModelKey]UpstreamBillingRow{}
	dayModelTimes := map[string]map[upstreamArchiveModelKey][2]int64{}
	channelUsers := make(map[upstreamBillingChannelUserKey]UpstreamBillingRow)
	channelTotals := make(map[int]UpstreamBillingRow)
	channels := map[int]bool{}
	var missingChannels int64
	excludedChannels := make([]model.UpstreamBillingChannelView, 0)
	for _, channel := range snapshot.Group.Channels {
		if channel.Missing {
			missingChannels++
		}
		if channel.Eligible {
			channels[channel.ID] = true
		} else {
			excludedChannels = append(excludedChannels, channel)
		}
	}
	sourceHash := sha256.New()
	cursor := int64(0)
	for ordinal := 0; ; ordinal++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		records, err := model.GetUpstreamBillingStatementRecords(statement, cursor, 1000)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			break
		}
		var chunk bytes.Buffer
		compressed := gzip.NewWriter(&chunk)
		for _, record := range records {
			if record.Sequence != cursor+1 || record.OccurredAt < statement.StartAt || record.OccurredAt >= statement.EndAt || !channels[record.ChannelID] {
				_ = compressed.Close()
				return model.ErrBillingEvidenceIntegrity
			}
			if err := total.addArchiveRecord(record); err != nil {
				_ = compressed.Close()
				return err
			}
			if snapshot.SchemaVersion >= 2 {
				key := upstreamBillingChannelUserKey{record.ChannelID, record.UserID}
				user, channel := channelUsers[key], channelTotals[record.ChannelID]
				if err := user.addArchiveRecord(record); err != nil {
					_ = compressed.Close()
					return err
				}
				if err := channel.addArchiveRecord(record); err != nil {
					_ = compressed.Close()
					return err
				}
				channelUsers[key], channelTotals[record.ChannelID] = user, channel
			}
			day := time.Unix(record.OccurredAt, 0).In(billingLocation).Format("2006-01-02")
			dayTotal := days[day]
			if err := dayTotal.addArchiveRecord(record); err != nil {
				_ = compressed.Close()
				return err
			}
			days[day] = dayTotal
			key := upstreamArchiveModelKey{record.ModelName, record.UpstreamModelName, record.BillingGroup}
			modelTotal := models[key]
			if err := modelTotal.addArchiveRecord(record); err != nil {
				_ = compressed.Close()
				return err
			}
			models[key] = modelTotal
			if modelDates[key] == nil {
				modelDates[key] = map[string]bool{}
			}
			modelDates[key][day] = true
			if modelFirst[key] == 0 || record.OccurredAt < modelFirst[key] {
				modelFirst[key] = record.OccurredAt
			}
			if record.OccurredAt > modelLast[key] {
				modelLast[key] = record.OccurredAt
			}
			if dayModels[day] == nil {
				dayModels[day] = map[upstreamArchiveModelKey]UpstreamBillingRow{}
				dayModelTimes[day] = map[upstreamArchiveModelKey][2]int64{}
			}
			dayModelTotal := dayModels[day][key]
			if err := dayModelTotal.addArchiveRecord(record); err != nil {
				_ = compressed.Close()
				return err
			}
			dayModels[day][key] = dayModelTotal
			times := dayModelTimes[day][key]
			if times[0] == 0 || record.OccurredAt < times[0] {
				times[0] = record.OccurredAt
			}
			if record.OccurredAt > times[1] {
				times[1] = record.OccurredAt
			}
			dayModelTimes[day][key] = times
			body, err := common.Marshal(record)
			if err != nil {
				_ = compressed.Close()
				return err
			}
			body = append(body, '\n')
			_, _ = sourceHash.Write(body)
			if _, err := compressed.Write(body); err != nil {
				_ = compressed.Close()
				return err
			}
			cursor = record.Sequence
		}
		if err := compressed.Close(); err != nil {
			return err
		}
		artifact, err := store.PutUpstream(ctx, statement.ID, "details", ordinal, statement.CreatedBy, "application/gzip", chunk.Bytes(), int64(len(records)))
		if err != nil {
			return err
		}
		manifest.Chunks = append(manifest.Chunks, model.BillingArtifact{Kind: "details", Ordinal: ordinal, SHA256: artifact.SHA256, Size: artifact.Size, Rows: artifact.Rows})
	}
	if hex.EncodeToString(sourceHash.Sum(nil)) != statement.SourceSHA256 || total.Count != statement.RecordCount || !total.matchesArchiveRow(snapshot.Total, snapshot.Currency) {
		return model.ErrBillingEvidenceIntegrity
	}
	for _, row := range snapshot.Days {
		if !days[row.Label].matchesArchiveRow(row, snapshot.Currency) {
			return model.ErrBillingEvidenceIntegrity
		}
		delete(days, row.Label)
	}
	if len(days) != 0 {
		return model.ErrBillingEvidenceIntegrity
	}
	for _, row := range snapshot.Models {
		key := upstreamArchiveModelKey{row.ModelName, row.UpstreamModelName, row.BillingGroup}
		if !models[key].matchesArchiveRow(row.UpstreamBillingRow, snapshot.Currency) || row.ActiveDays != int64(len(modelDates[key])) || row.FirstPosted != modelFirst[key] || row.LastPosted != modelLast[key] {
			return model.ErrBillingEvidenceIntegrity
		}
		delete(models, key)
	}
	if len(models) != 0 {
		return model.ErrBillingEvidenceIntegrity
	}
	for _, row := range snapshot.DayModels {
		key := upstreamArchiveModelKey{row.ModelName, row.UpstreamModelName, row.BillingGroup}
		times := dayModelTimes[row.Date][key]
		if !dayModels[row.Date][key].matchesArchiveRow(row.UpstreamBillingRow, snapshot.Currency) || row.ActiveDays != 1 || row.FirstPosted != times[0] || row.LastPosted != times[1] {
			return model.ErrBillingEvidenceIntegrity
		}
		delete(dayModels[row.Date], key)
	}
	for _, rows := range dayModels {
		if len(rows) != 0 {
			return model.ErrBillingEvidenceIntegrity
		}
	}
	if snapshot.Coverage.MissingCostRecords != total.MissingCostRecords || snapshot.Coverage.MissingUsageRecords != total.MissingUsageRecords || snapshot.Coverage.EstimatedUsageRecords != total.EstimatedUsageRecords || snapshot.Coverage.PendingTasks != total.PendingTasks {
		return model.ErrBillingEvidenceIntegrity
	}
	complete := len(channels) > 0 && missingChannels == 0 && total.MissingCostRecords == 0 && total.MissingUsageRecords == 0 && total.EstimatedUsageRecords == 0 && total.PendingTasks == 0
	if snapshot.Coverage.Complete != complete || snapshot.Coverage.MissingChannels != missingChannels || len(snapshot.Coverage.ExcludedChannels) != len(excludedChannels) {
		return model.ErrBillingEvidenceIntegrity
	}
	for i, channel := range excludedChannels {
		if snapshot.Coverage.ExcludedChannels[i] != channel {
			return model.ErrBillingEvidenceIntegrity
		}
	}
	if err := verifyUpstreamBillingAllocation(&snapshot, channelUsers, channelTotals); err != nil {
		return err
	}
	manifest.Rows, manifest.ChargeQuota, manifest.RefundQuota = total.Count, total.ChargeQuota, total.RefundQuota
	if _, err := store.PutUpstream(ctx, statement.ID, "snapshot", 0, statement.CreatedBy, "application/json", []byte(statement.Snapshot), 0); err != nil {
		return err
	}
	workbook, err := RenderUpstreamBillingStatementExcel(statement, &snapshot)
	if err != nil {
		return err
	}
	artifact, err := store.PutUpstream(ctx, statement.ID, "xlsx", 0, statement.CreatedBy, billingExcelContentType, workbook, 0)
	if err != nil {
		return err
	}
	manifest.ExcelSHA256, statement.ExcelSHA256 = artifact.SHA256, artifact.SHA256
	pdf, err := RenderUpstreamBillingStatementPDF(statement, &snapshot, false)
	if err != nil {
		return err
	}
	artifact, err = store.PutUpstream(ctx, statement.ID, "pdf", 0, statement.CreatedBy, "application/pdf", pdf, 0)
	if err != nil {
		return err
	}
	manifest.PDFSHA256, statement.PDFSHA256 = artifact.SHA256, artifact.SHA256
	body, err := common.Marshal(manifest)
	if err != nil {
		return err
	}
	artifact, err = store.PutUpstream(ctx, statement.ID, "manifest", 0, statement.CreatedBy, "application/json", body, manifest.Rows)
	if err != nil {
		return err
	}
	statement.ManifestSHA256 = artifact.SHA256
	return nil
}

func (row *UpstreamBillingRow) addArchiveRecord(record model.UpstreamBillingStatementRecord) error {
	if record.CostQuota > int64(common.MaxQuota) || record.CostQuota < -int64(common.MaxQuota) || row.Count >= int64(common.MaxWalletQuota) {
		return model.ErrBillingEvidenceIntegrity
	}
	switch record.LogType {
	case model.LogTypeConsume:
		if record.CostQuota < 0 || (!record.CostMissing && record.CostQuota > int64(common.MaxWalletQuota)-row.ChargeQuota) {
			return model.ErrBillingEvidenceIntegrity
		}
		if !record.CostMissing {
			row.ChargeQuota += record.CostQuota
		}
		row.ChargeCount++
	case model.LogTypeRefund:
		if record.CostQuota > 0 || (!record.CostMissing && -record.CostQuota > int64(common.MaxWalletQuota)-row.RefundQuota) {
			return model.ErrBillingEvidenceIntegrity
		}
		if !record.CostMissing {
			row.RefundQuota -= record.CostQuota
		}
		row.RefundCount++
	default:
		return model.ErrBillingEvidenceIntegrity
	}
	row.Count++
	for _, pair := range []struct {
		target *int64
		value  int64
	}{
		{&row.Requests, record.Requests}, {&row.InputTokens, record.InputTokens}, {&row.OutputTokens, record.OutputTokens},
		{&row.CacheReadTokens, record.CacheReadTokens}, {&row.CacheWriteTokens, record.CacheWriteTokens}, {&row.CacheWrite5mTokens, record.CacheWrite5mTokens}, {&row.CacheWrite1hTokens, record.CacheWrite1hTokens}, {&row.VideoTokens, record.VideoTokens}, {&row.VideoSeconds, record.VideoSeconds},
	} {
		if pair.value < 0 || pair.value > int64(common.MaxWalletQuota)-*pair.target {
			return model.ErrBillingEvidenceIntegrity
		}
		*pair.target += pair.value
	}
	if record.CostMissing {
		row.MissingCostRecords++
	}
	if record.UsageMissing {
		row.MissingUsageRecords++
	}
	if record.UsageEstimated {
		row.EstimatedUsageRecords++
	}
	if record.TaskPending {
		row.PendingTasks++
	}
	return nil
}

func (row UpstreamBillingRow) matchesArchiveRow(expected UpstreamBillingRow, currency BillingCurrency) bool {
	amounts, err := billingRow("", row.ChargeQuota, row.RefundQuota, row.Count, currency)
	return err == nil && row.ChargeQuota == expected.ChargeQuota && row.RefundQuota == expected.RefundQuota && row.Count == expected.Count && amounts.Charge == expected.Charge && amounts.Refund == expected.Refund && amounts.Amount == expected.Amount && row.UpstreamBillingUsage == expected.UpstreamBillingUsage && row.ChargeCount == expected.ChargeCount && row.RefundCount == expected.RefundCount && row.MissingCostRecords == expected.MissingCostRecords && row.MissingUsageRecords == expected.MissingUsageRecords && row.EstimatedUsageRecords == expected.EstimatedUsageRecords && row.PendingTasks == expected.PendingTasks
}

// Confirmation is an additional artifact. Original PDF, Excel and evidence
// bytes keep the hashes that the financial administrator explicitly confirmed.
func BuildUpstreamBillingReceipt(ctx context.Context, statement *model.UpstreamBillingStatement, store UpstreamBillingArchiveWriter) (*model.BillingArtifact, error) {
	var snapshot UpstreamBillingSnapshot
	if err := statement.VerifySnapshot(); err != nil {
		return nil, err
	}
	if err := common.UnmarshalJsonStr(string(statement.Snapshot), &snapshot); err != nil {
		return nil, err
	}
	body, err := RenderUpstreamBillingStatementPDF(statement, &snapshot, true)
	if err != nil {
		return nil, err
	}
	return store.PutUpstream(ctx, statement.ID, "receipt", 0, statement.CreatedBy, "application/pdf", body, 0)
}

func UpstreamBillingArtifactFilename(statement *model.UpstreamBillingStatement, snapshot *UpstreamBillingSnapshot, kind string, ordinal int) string {
	name := strings.TrimSpace(snapshot.Group.Name)
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || strings.ContainsRune(`/\\:*?"<>|`, r) {
			return '_'
		}
		return r
	}, name)
	name = strings.Trim(name, ". ")
	if name == "" {
		name = fmt.Sprintf("upstream-%d", statement.GroupID)
	}
	if runes := []rune(name); len(runes) > 64 {
		name = string(runes[:64])
	}
	if kind == "pdf" || kind == "xlsx" {
		return fmt.Sprintf("%s_%s_上游对账单_V%02d.%s", name, statement.Month, statement.Revision, kind)
	}
	if kind == "receipt" {
		return fmt.Sprintf("%s_%s_上游对账确认回执_V%02d.pdf", name, statement.Month, statement.Revision)
	}
	extension := "json"
	if kind == "details" {
		extension = "jsonl.gz"
	}
	return fmt.Sprintf("upstream-%s-r%d-%s-%d.%s", statement.Month, statement.Revision, kind, ordinal, extension)
}
