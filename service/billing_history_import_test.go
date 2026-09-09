package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type historyImportArchive struct {
	body []byte
	err  error
}

func (store *historyImportArchive) Put(_ context.Context, _ string, _ string, _ int, _ int, _ string, body []byte, _ int64) (*model.BillingArtifact, error) {
	if store.err != nil {
		return nil, store.err
	}
	store.body = body
	return &model.BillingArtifact{ID: 1}, nil
}

func seedBillingHistory(t *testing.T) int64 {
	t.Helper()
	truncate(t)
	seedUser(t, 96, 8765432)
	start, _, err := model.BillingMonthBounds("2020-04")
	require.NoError(t, err)
	require.NoError(t, model.DB.Create(&model.BillingAccount{UserID: 96, AccountingStartAt: start, StartSequence: 1, CompanyTitle: "History Customer", TaxID: "HISTORY", ProfileVersion: 1}).Error)
	require.NoError(t, model.LOG_DB.Create(&[]model.Log{
		{Id: 961, UserId: 96, CreatedAt: start + 3600, Type: model.LogTypeConsume, Quota: 500000, RequestId: "old-charge", Content: "never-archive-private-body", Other: "never-archive-admin-data"},
		{Id: 962, UserId: 96, CreatedAt: start + 2*86400, Type: model.LogTypeRefund, Quota: 125000, RequestId: "old-refund"},
		{Id: 963, UserId: 96, CreatedAt: start + 3*86400, Type: model.LogTypeConsume, Quota: 750000, RequestId: "old-second-charge"},
		{Id: 964, UserId: 96, CreatedAt: start + 3600, Type: model.LogTypeTopup, Quota: 9000000},
		{Id: 965, UserId: 97, CreatedAt: start + 3600, Type: model.LogTypeConsume, Quota: 9000000},
	}).Error)
	return start
}

func TestBillingHistoryImportRequiresReviewedEvidenceAndNeverChargesWallet(t *testing.T) {
	seedBillingHistory(t)
	ctx := context.Background()
	review, err := ReviewBillingHistory(ctx, 96, "2020-04")
	require.NoError(t, err)
	require.True(t, review.Ready)
	assert.Equal(t, 3, review.SourceCount)
	assert.Equal(t, int64(1250000), review.Snapshot.ChargeQuota)
	assert.Equal(t, int64(125000), review.Snapshot.RefundQuota)
	var count int64
	require.NoError(t, model.DB.Model(&model.BillingEntry{}).Count(&count).Error)
	assert.Zero(t, count, "review is not consent to import")
	store := &historyImportArchive{}
	batch, err := ConfirmBillingHistoryImport(ctx, 96, 1, 1, "2020-04", "00000000000000000000000000000096", review.SourceSHA256, "Checked against historical consumption and administrator funding records", "live-admin-session", store)
	require.NoError(t, err)
	assert.Equal(t, int64(3), batch.Records)
	assert.Equal(t, int64(1), batch.FromSequence)
	assert.Equal(t, int64(3), batch.ToSequence)
	var user model.User
	require.NoError(t, model.DB.First(&user, 96).Error)
	assert.Equal(t, 8765432, user.Quota)
	assert.Zero(t, user.UsedQuota)
	assert.Zero(t, user.RequestCount)
	var entries []model.BillingEntry
	require.NoError(t, model.DB.Order("sequence asc").Find(&entries).Error)
	require.Len(t, entries, 3)
	assert.Equal(t, int64(-125000), entries[1].Quota)
	for _, entry := range entries {
		assert.Zero(t, entry.WalletDelta)
		assert.Equal(t, batch.ID, entry.ImportID)
		assert.Equal(t, "unchanged_at_import", entry.BalanceBasis)
		assert.NotEmpty(t, entry.SourceSHA256)
	}
	reader, err := gzip.NewReader(bytes.NewReader(store.body))
	require.NoError(t, err)
	source, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	assert.Equal(t, review.Source, source)
	assert.NotContains(t, string(source), "never-archive")
	retry, err := ConfirmBillingHistoryImport(ctx, 96, 1, 1, "2020-04", "00000000000000000000000000000096", review.SourceSHA256, "retry", "live-admin-session", store)
	require.NoError(t, err)
	assert.Equal(t, batch.ID, retry.ID)
	require.NoError(t, model.DB.Model(&model.BillingEntry{}).Count(&count).Error)
	assert.Equal(t, int64(3), count, "retry must not import or charge twice")
	statement, err := PrepareBillingStatement(96, 1, 1, "2020-04")
	require.NoError(t, err)
	var snapshot BillingSnapshot
	require.NoError(t, common.UnmarshalJsonStr(statement.Snapshot, &snapshot))
	assert.Equal(t, review.Snapshot.Total, snapshot.Total)
	assert.Equal(t, batch.ID, snapshot.HistoryImportID)
	assert.Equal(t, batch.SourceSHA256, snapshot.HistorySourceSHA256)
	assert.Equal(t, "test_user", snapshot.Username)
	assert.Equal(t, 3, snapshot.PDFTemplateVersion)
	assert.NotEmpty(t, snapshot.PDFLogoPNG, "the archive freezes the platform image rather than a mutable URL")
	archive := &memoryBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildBillingArchive(ctx, statement, archive), "imported rows must reconcile with the independent hourly summary")
	reader, err = gzip.NewReader(bytes.NewReader(archive.files["details"]))
	require.NoError(t, err)
	archivedDetails, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	assert.Contains(t, string(archivedDetails), batch.ID)
	assert.Contains(t, string(archivedDetails), entries[0].SourceSHA256)
	assert.NotContains(t, string(archivedDetails), "balance")
}

func TestBillingHistoryRejectsChangedSourceBeforeArchivingOrImporting(t *testing.T) {
	start := seedBillingHistory(t)
	review, err := ReviewBillingHistory(context.Background(), 96, "2020-04")
	require.NoError(t, err)
	require.NoError(t, model.LOG_DB.Create(&model.Log{UserId: 96, CreatedAt: start + 7200, Type: model.LogTypeConsume, Quota: 1}).Error)
	store := &historyImportArchive{}
	_, err = ConfirmBillingHistoryImport(context.Background(), 96, 1, 1, "2020-04", "00000000000000000000000000000096", review.SourceSHA256, "checked", "session", store)
	assert.ErrorIs(t, err, model.ErrBillingConflict)
	assert.Empty(t, store.body)
	var entries int64
	require.NoError(t, model.DB.Model(&model.BillingEntry{}).Count(&entries).Error)
	assert.Zero(t, entries)
}

func TestBillingHistoryCloudFailureAndConcurrentDraftLeaveLedgerUnchanged(t *testing.T) {
	seedBillingHistory(t)
	review, err := ReviewBillingHistory(context.Background(), 96, "2020-04")
	require.NoError(t, err)
	_, err = ConfirmBillingHistoryImport(context.Background(), 96, 1, 1, "2020-04", "00000000000000000000000000000096", review.SourceSHA256, "checked", "session", &historyImportArchive{err: errors.New("store offline")})
	assert.ErrorContains(t, err, "store offline")
	require.NoError(t, model.DB.Create(&model.BillingStatement{ID: "concurrent-draft", UserID: 96, Month: "2020-04", Status: model.StatementDraft}).Error)
	batch := &model.BillingHistoryImport{ID: "concurrent-import", UserID: 96, Month: "2020-04", SourceSHA256: review.SourceSHA256, ArtifactID: 1, SessionID: "session", Note: "checked"}
	err = model.ApplyBillingHistoryImport(context.Background(), batch, review.Account, review.Records)
	assert.ErrorIs(t, err, model.ErrBillingHistoryBlocked)
	var entries, imports int64
	require.NoError(t, model.DB.Model(&model.BillingEntry{}).Count(&entries).Error)
	require.NoError(t, model.DB.Model(&model.BillingHistoryImport{}).Count(&imports).Error)
	assert.Zero(t, entries)
	assert.Zero(t, imports)
	blocked, err := ReviewBillingHistory(context.Background(), 96, "2020-04")
	require.NoError(t, err)
	assert.False(t, blocked.Ready)
	assert.Equal(t, "concurrent-draft", blocked.ExistingStatement)
}

func TestBillingDraftRejectsLegacyHistoryButAllowsAnActuallyEmptyMonth(t *testing.T) {
	seedBillingHistory(t)
	_, err := PrepareBillingStatement(96, 1, 1, "2020-04")
	assert.ErrorIs(t, err, ErrBillingHistoricalDataUnreconciled)
	statement, err := PrepareBillingStatement(96, 1, 1, "2020-05")
	require.NoError(t, err)
	assert.Equal(t, model.StatementPreparing, statement.Status)
	assert.NoError(t, ValidateBillingStatementSource(context.Background(), statement))
}
