package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func detailArchiveFixture(t *testing.T, currency BillingCurrency, entries []model.BillingEntry) (*model.BillingStatement, *BillingSnapshot) {
	t.Helper()
	truncate(t)
	start, end, err := model.BillingMonthBounds("2026-02")
	require.NoError(t, err)
	var charge int64
	for i := range entries {
		entries[i].UserID, entries[i].Sequence = 47, int64(i+1)
		entries[i].PostedAt, entries[i].ModelName = start+int64(i), "model"
		entries[i].EventKey, entries[i].Kind = fmt.Sprintf("detail-%d", i), "usage"
		charge += entries[i].Quota
	}
	require.NoError(t, model.DB.Create(&entries).Error)
	snapshot, err := buildBillingSnapshot(&model.BillingAccount{UserID: 47, AccountingStartAt: start, CompanyTitle: "Customer", TaxID: "TEST"}, "2026-02", []model.BillingHour{{Hour: start, Charge: charge, Count: int64(len(entries))}}, currency)
	require.NoError(t, err)
	snapshot.PDFTemplateVersion, snapshot.PDFLogoPNG, snapshot.DetailSheet = 9, billingPlatformMarkV2, true
	require.NoError(t, attachBillingModelRows(snapshot, []model.BillingModelTotal{{ModelName: "model", Charge: charge, Count: int64(len(entries)), ChargeCount: int64(len(entries)), ActiveDays: 1, FirstPosted: start, LastPosted: start + int64(len(entries)-1)}}))
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	return &model.BillingStatement{ID: "detail-statement", UserID: 47, Month: "2026-02", Revision: 1, Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]), StartAt: start, EndAt: end, CreatedAt: end + 86400, FromSequence: 1, ToSequence: int64(len(entries))}, snapshot
}

func TestBillingDetailRoundingIsDisclosedAndReconcilesWithTheOverview(t *testing.T) {
	statement, snapshot := detailArchiveFixture(t, BillingCurrency{Code: "CNY", Symbol: "¥", Rate: "7.1", QuotaPerUnit: "500000"}, []model.BillingEntry{
		{Quota: 1, RequestID: "a"}, {Quota: 1, RequestID: "b"}, {Quota: 1, RequestID: "c"},
	})
	store := &memoryBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))
	book, err := excelize.OpenReader(bytes.NewReader(store.files["xlsx"]))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("详细日志", excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	column := -1
	for i, title := range rows[0] {
		if title == "实际扣费金额" {
			column = i
		}
	}
	require.NotEqual(t, -1, column)
	sum := decimal.Zero
	for _, row := range rows[1:] {
		amount, err := decimal.NewFromString(row[column])
		require.NoError(t, err)
		sum = sum.Add(amount)
	}
	assert.Equal(t, "0.000042", sum.StringFixed(6))
	assert.Equal(t, "0.000043", snapshot.Total.Amount)
	assert.Equal(t, "0.000000", snapshot.RoundingDifference, "day rounding does not describe request rounding")
	overview, err := book.GetRows("总览", excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	adjustments := make(map[string]string)
	for _, row := range overview {
		if len(row) > 1 {
			adjustments[row[0]] = row[1]
		}
	}
	assert.Equal(t, "0.000001", adjustments["详细日志消费舍入差额"])
	assert.Equal(t, "0.000001", adjustments["详细日志净额舍入差额"])
	assert.NotEmpty(t, adjustments["详细日志舍入说明"])
	difference, err := decimal.NewFromString(adjustments["详细日志净额舍入差额"])
	require.NoError(t, err)
	assert.Equal(t, snapshot.Total.Amount, sum.Add(difference).StringFixed(6))
}

type interruptedBillingArchive struct {
	memoryBillingArchive
	failPDF bool
}

func (store *interruptedBillingArchive) Put(ctx context.Context, statementID, kind string, ordinal, userID int, contentType string, body []byte, rows int64) (*model.BillingArtifact, error) {
	if kind == "pdf" && store.failPDF {
		return nil, errors.New("PDF upload unavailable")
	}
	if previous, exists := store.files[kind]; exists && !bytes.Equal(previous, body) {
		return nil, model.ErrBillingConflict
	}
	return store.memoryBillingArchive.Put(ctx, statementID, kind, ordinal, userID, contentType, body, rows)
}

func TestBillingArchiveRetryReusesExcelAfterTokenRename(t *testing.T) {
	statement, _ := detailArchiveFixture(t, BillingCurrency{Code: "CNY", Symbol: "¥", Rate: "7", QuotaPerUnit: "500000"}, []model.BillingEntry{{Quota: 500000, RequestID: "request", TokenID: 47}})
	require.NoError(t, model.DB.Create(&model.Token{Id: 47, UserId: 47, Key: "archive-test-key", Name: "before"}).Error)
	store := &interruptedBillingArchive{memoryBillingArchive: memoryBillingArchive{files: map[string][]byte{}}, failPDF: true}
	require.EqualError(t, BuildBillingArchive(context.Background(), statement, store), "PDF upload unavailable")
	original := append([]byte(nil), store.files["xlsx"]...)
	require.NotEmpty(t, original)
	require.NoError(t, model.DB.Model(&model.Token{}).Where("id = ?", 47).Update("name", "after").Error)
	store.failPDF = false

	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))
	assert.Equal(t, original, store.files["xlsx"], "a retry preserves the first archived workbook")
	var manifest BillingManifest
	require.NoError(t, common.Unmarshal(store.files["manifest"], &manifest))
	hash := sha256.Sum256(original)
	assert.Equal(t, hex.EncodeToString(hash[:]), manifest.ExcelSHA256)
	assert.NotEmpty(t, store.files["pdf"])
}
