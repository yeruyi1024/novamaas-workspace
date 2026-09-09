package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memoryBillingArchive struct {
	files  map[string][]byte
	writes int
}

func (store *memoryBillingArchive) Put(_ context.Context, statementID, kind string, ordinal, userID int, contentType string, body []byte, rows int64) (*model.BillingArtifact, error) {
	hash := sha256.Sum256(body)
	store.files[kind] = append([]byte(nil), body...)
	store.writes++
	return &model.BillingArtifact{SHA256: hex.EncodeToString(hash[:]), Size: int64(len(body)), Rows: rows}, nil
}

func TestBillingPDFIsDeterministicAndEmbedsCorporateIdentity(t *testing.T) {
	account := &model.BillingAccount{UserID: 11, CompanyTitle: "上海示例科技有限公司", TaxID: "91310000EXAMPLE", AccountingStartAt: 1769875200}
	snapshot, err := BuildBillingSnapshot(account, "2026-02", []model.BillingHour{{Hour: 1769875200, Charge: 123450000, Refund: 4500000, Count: 18}, {Hour: 1769961600, Charge: 56750000, Count: 12}})
	require.NoError(t, err)
	statement := &model.BillingStatement{ID: "sample-statement-202602", Month: "2026-02", Revision: 1, UserID: 11, CreatedAt: 1772380800, StartAt: 1769875200, EndAt: 1772294400}
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement.Snapshot, statement.SnapshotSHA256 = string(body), hex.EncodeToString(hash[:])
	first, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	second, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.Equal(t, first, second, "retries must reuse the exact original PDF bytes")
	assert.Contains(t, string(first), "/FontFile2", "Chinese font must be embedded, not rely on host fonts")
	pdfHash := sha256.Sum256(first)
	statement.PDFSHA256, statement.ManifestSHA256 = hex.EncodeToString(pdfHash[:]), statement.SnapshotSHA256
	statement.ConfirmedAt = statement.CreatedAt + 86400
	receipt, err := RenderBillingStatementPDF(statement, snapshot, true)
	require.NoError(t, err)
	unchanged, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.Equal(t, first, unchanged, "confirmation must not change the original PDF")
	if output := os.Getenv("BILLING_PDF_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.WriteFile(output, receipt, 0600))
	}
}

func TestBillingPDFV2FreezesBrandingIdentityAndOriginalAfterConfirmation(t *testing.T) {
	start, end, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	account := &model.BillingAccount{UserID: 4, CompanyTitle: "上海示例科技有限公司（演示数据）", TaxID: "DEMO-NOT-A-REAL-TAX-ID", AccountingStartAt: start}
	snapshot, err := BuildBillingSnapshot(account, "2026-08", []model.BillingHour{{Hour: start + 9*3600, Charge: 500000, Count: 1}, {Hour: start + 14*86400 + 9*3600, Refund: 125000, Count: 1}, {Hour: start + 19*86400 + 15*3600, Charge: 750000, Count: 1}})
	require.NoError(t, err)
	snapshot.PDFTemplateVersion, snapshot.Issuer = 2, "聚合算力网关"
	snapshot.Username, snapshot.DisplayName = "demo_customer", "示例客户（仅展示版式）"
	snapshot.Currency = BillingCurrency{Code: "CNY", Symbol: "¥", Rate: "7", QuotaPerUnit: "500000"}
	for i := range snapshot.Days {
		row := snapshot.Days[i]
		snapshot.Days[i], err = billingRow(row.Label, row.ChargeQuota, row.RefundQuota, row.Count, snapshot.Currency)
		require.NoError(t, err)
	}
	snapshot.Total, err = billingRow(snapshot.Month, snapshot.ChargeQuota, snapshot.RefundQuota, snapshot.Total.Count, snapshot.Currency)
	require.NoError(t, err)
	statement := &model.BillingStatement{ID: "DEMO-202608-DOCUMENT-PREVIEW", UserID: 4, Month: snapshot.Month, Revision: 2, StartAt: start, EndAt: end, CreatedAt: end + 2*86400}
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement.Snapshot, statement.SnapshotSHA256 = string(body), hex.EncodeToString(hash[:])
	original, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.Contains(t, string(original), "/Subtype /Image", "the platform mark must be embedded")
	savedName := common.SystemName
	common.SystemName = "A later platform name"
	t.Cleanup(func() { common.SystemName = savedName })
	pdfHash := sha256.Sum256(original)
	statement.PDFSHA256, statement.ManifestSHA256 = hex.EncodeToString(pdfHash[:]), statement.SnapshotSHA256
	statement.IssuedAt, statement.ConfirmedAt = statement.CreatedAt+3600, statement.CreatedAt+7200
	unchanged, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.Equal(t, original, unchanged, "brand settings and live status must not rewrite an original")
	receipt, err := RenderBillingStatementPDF(statement, snapshot, true)
	require.NoError(t, err)
	retry, err := RenderBillingStatementPDF(statement, snapshot, true)
	require.NoError(t, err)
	assert.Equal(t, receipt, retry, "confirmation receipt retries are deterministic")
	if output := os.Getenv("BILLING_PDF_V2_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.WriteFile(output, receipt, 0600))
	}
	snapshot.PDFTemplateVersion = 99
	_, err = RenderBillingStatementPDF(statement, snapshot, false)
	assert.ErrorContains(t, err, "template version", "unknown templates cannot silently alter an archived file")
}

func TestBillingPDFV2SupportsTheAcceptedCorporateIdentityLength(t *testing.T) {
	start, end, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	snapshot, err := BuildBillingSnapshot(&model.BillingAccount{UserID: 4, CompanyTitle: strings.Repeat("企", 200), TaxID: strings.Repeat("X", 64), AccountingStartAt: start}, "2026-08", nil)
	require.NoError(t, err)
	snapshot.PDFTemplateVersion = 2
	snapshot.Username, snapshot.DisplayName = "example_customer", "合作伙伴_企业_示例客户"
	snapshot.HistoryImportID = "history-00000000000000000000000000000004"
	statement := &model.BillingStatement{ID: "long-corporate-identity", UserID: 4, Month: "2026-08", Revision: 1, StartAt: start, EndAt: end, CreatedAt: end + 86400, SnapshotSHA256: strings.Repeat("1", 64), PDFSHA256: strings.Repeat("2", 64), ManifestSHA256: strings.Repeat("3", 64), ConfirmedAt: end + 86400}
	_, err = RenderBillingStatementPDF(statement, snapshot, true)
	require.NoError(t, err, "valid maximum-length company and tax ID must fit without hiding or clipping content")
}

func TestBillingArchiveRejectsMissingDetailsBeforePublishingPDF(t *testing.T) {
	truncate(t)
	snapshot, err := BuildBillingSnapshot(&model.BillingAccount{UserID: 41}, "2026-02", []model.BillingHour{{Hour: 1769875200, Charge: 100, Count: 1}})
	require.NoError(t, err)
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "missing-detail", UserID: 41, Month: "2026-02", Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]), StartAt: 1769875200, EndAt: 1772294400, FromSequence: 1, ToSequence: 1}
	store := &memoryBillingArchive{files: map[string][]byte{}}
	err = BuildBillingArchive(context.Background(), statement, store)
	require.ErrorContains(t, err, "do not reconcile")
	assert.Zero(t, store.writes, "no PDF or manifest may be published when financial details are missing")
}

func TestBillingArchiveFreezesDetailsAndVerifiesTotals(t *testing.T) {
	truncate(t)
	posted := int64(1769875200)
	entry := &model.BillingEntry{EventKey: "source", UserID: 42, Sequence: 1, PostedAt: posted, Kind: "usage", Quota: 100, RequestID: "request", ModelName: "model"}
	require.NoError(t, model.DB.Create(entry).Error)
	snapshot, err := BuildBillingSnapshot(&model.BillingAccount{UserID: 42, CompanyTitle: "测试企业", TaxID: "TEST"}, "2026-02", []model.BillingHour{{Hour: posted, Charge: 100, Count: 1}})
	require.NoError(t, err)
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "frozen-detail", UserID: 42, Month: "2026-02", Revision: 1, Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]), StartAt: posted, EndAt: 1772294400, CreatedAt: 1772380800, FromSequence: 1, ToSequence: 1}
	store := &memoryBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))
	require.Len(t, store.files, 4)
	var manifest BillingManifest
	require.NoError(t, common.Unmarshal(store.files["manifest"], &manifest))
	assert.Equal(t, int64(1), manifest.Rows)
	assert.Equal(t, int64(100), manifest.ChargeQuota)
	assert.Equal(t, statement.PDFSHA256, manifest.PDFSHA256)
	assert.NotEmpty(t, statement.ManifestSHA256)
	statement.Snapshot += " "
	err = BuildBillingArchive(context.Background(), statement, store)
	assert.ErrorIs(t, err, model.ErrBillingEvidenceIntegrity)
}

func TestBillingArchiveRejectsWrongDayEvenWhenMonthlyTotalMatches(t *testing.T) {
	truncate(t)
	start := int64(1769875200)
	require.NoError(t, model.DB.Create(&model.BillingEntry{EventKey: "wrong-day", UserID: 45, Sequence: 1, PostedAt: start + 86400, Kind: "usage", Quota: 100}).Error)
	snapshot, err := BuildBillingSnapshot(&model.BillingAccount{UserID: 45}, "2026-02", []model.BillingHour{{Hour: start, Charge: 100, Count: 1}})
	require.NoError(t, err)
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "wrong-day", UserID: 45, Month: "2026-02", Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]), StartAt: start, EndAt: 1772294400, FromSequence: 1, ToSequence: 1}
	store := &memoryBillingArchive{files: map[string][]byte{}}
	err = BuildBillingArchive(context.Background(), statement, store)
	require.ErrorContains(t, err, "daily details")
	assert.NotContains(t, store.files, "pdf")
	assert.NotContains(t, store.files, "manifest")
}

func TestBillingRoundingDifferencePreservesExactTotal(t *testing.T) {
	currency := BillingCurrency{Code: "USD", Symbol: "$", Rate: "1", QuotaPerUnit: "3"}
	row, err := billingRow("day", 1, 0, 1, currency)
	require.NoError(t, err)
	total, err := billingRow("month", 2, 0, 2, currency)
	require.NoError(t, err)
	difference, err := billingRoundingDifference([]BillingRow{row, row}, total)
	require.NoError(t, err)
	assert.Equal(t, "0.333333", row.Amount)
	assert.Equal(t, "0.666667", total.Amount)
	assert.Equal(t, "0.000001", difference)
}

func TestBillingArchiveRejectsMissingFirstHourInsteadOfOmittingItsEntries(t *testing.T) {
	truncate(t)
	seedUser(t, 91, 1000000)
	start, _, err := model.BillingMonthBounds("2020-02")
	require.NoError(t, err)
	require.NoError(t, model.DB.Create(&model.BillingAccount{UserID: 91, AccountingStartAt: start, StartSequence: 1, Sequence: 2, CompanyTitle: "Test", TaxID: "TEST"}).Error)
	entries := []model.BillingEntry{
		{EventKey: "missing-first-hour", UserID: 91, Sequence: 1, PostedAt: start + 3600, Kind: "usage", Quota: 200},
		{EventKey: "surviving-hour", UserID: 91, Sequence: 2, PostedAt: start + 7200, Kind: "usage", Quota: 100},
	}
	require.NoError(t, model.DB.Create(&entries).Error)
	require.NoError(t, model.DB.Create(&model.BillingHour{UserID: 91, Hour: start + 7200, Charge: 100, Count: 1, FirstSequence: 2, LastSequence: 2}).Error)
	statement, err := PrepareBillingStatement(91, 1, 1, "2020-02")
	require.NoError(t, err)
	store := &memoryBillingArchive{files: map[string][]byte{}}
	err = BuildBillingArchive(context.Background(), statement, store)
	require.Error(t, err, "a missing boundary summary must not silently remove official entries from evidence")
	assert.NotContains(t, store.files, "pdf")
	assert.NotContains(t, store.files, "manifest")
}

func TestDurableStreamingReservationsSettleOnlyOnce(t *testing.T) {
	truncate(t)
	seedUser(t, 46, 100000)
	seedToken(t, 47, 46, "billing-realtime", 100000)
	start := int64(0)
	_, err := model.SaveBillingAccount(46, 1, 0, "Realtime customer", "TAX", &start)
	require.NoError(t, err)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	info := &relaycommon.RelayInfo{UserId: 46, TokenId: 47, TokenKey: "billing-realtime", OriginModelName: "gpt-4o-mini", UsingGroup: "default", UserGroup: "default", ForcePreConsume: true, RequestId: "realtime-request"}
	info.UserSetting.BillingPreference = "wallet_only"
	session, apiErr := NewBillingSession(ctx, info, 100)
	require.Nil(t, apiErr)
	info.Billing = session
	require.NoError(t, info.Billing.Reserve(250))
	require.NoError(t, info.Billing.Reserve(400))
	reserved := session.GetPreConsumedQuota()
	require.Equal(t, 400, reserved)
	var entries int64
	require.NoError(t, model.DB.Model(&model.BillingEntry{}).Where("user_id = ?", 46).Count(&entries).Error)
	assert.Zero(t, entries, "streaming reservations are not final charges")
	require.NoError(t, session.Settle(reserved))
	require.NoError(t, session.Settle(reserved))
	quota, err := model.GetUserQuota(46, true)
	require.NoError(t, err)
	assert.Equal(t, 100000-reserved, quota)
	var ledger []model.BillingEntry
	require.NoError(t, model.DB.Where("user_id = ?", 46).Find(&ledger).Error)
	require.Len(t, ledger, 1)
	assert.Equal(t, int64(reserved), ledger[0].Quota)
}
