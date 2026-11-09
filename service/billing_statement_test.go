package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	"github.com/xuri/excelize/v2"
)

type memoryBillingArchive struct {
	files  map[string][]byte
	writes int
}

func (store *memoryBillingArchive) Existing(_ context.Context, _ string, kind string, _ int) (*model.BillingArtifact, error) {
	body, exists := store.files[kind]
	if !exists {
		return nil, nil
	}
	hash := sha256.Sum256(body)
	return &model.BillingArtifact{SHA256: hex.EncodeToString(hash[:]), Size: int64(len(body))}, nil
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

func TestBillingPDFV8AddsExactCustomerModelSummaryBeforeConfirmationReceipt(t *testing.T) {
	start, end, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	snapshot, err := BuildBillingSnapshot(
		&model.BillingAccount{UserID: 4, CompanyTitle: "上海示例科技有限公司（演示数据）", TaxID: "DEMO-NOT-A-REAL-TAX-ID", AccountingStartAt: start},
		"2026-08",
		[]model.BillingHour{{Hour: start + 9*3600, Charge: 1500000, Refund: 125000, Count: 5}},
	)
	require.NoError(t, err)
	snapshot.PDFTemplateVersion = 8
	snapshot.PDFLogoPNG = billingPlatformMarkV2
	snapshot.Issuer = "聚合算力网关"
	snapshot.Username = "demo_customer"
	require.NoError(t, attachBillingModelRows(snapshot, []model.BillingModelTotal{
		{ModelName: "gpt-5", Charge: 800000, Count: 1, ChargeCount: 1, ActiveDays: 1, FirstPosted: start + 9*3600, LastPosted: start + 9*3600},
		{ModelName: "claude-sonnet", Charge: 400000, Refund: 125000, Count: 2, ChargeCount: 1, RefundCount: 1, ActiveDays: 2, FirstPosted: start + 2*86400, LastPosted: start + 10*86400},
		{ModelName: "企业级超长上下文推理与多模态分析模型专业版本", Charge: 200000, Count: 1, ChargeCount: 1, ActiveDays: 1, FirstPosted: start + 15*86400, LastPosted: start + 15*86400},
		{Charge: 100000, Count: 1, ChargeCount: 1, ActiveDays: 1, FirstPosted: start + 20*86400, LastPosted: start + 20*86400},
	}))
	require.Len(t, snapshot.Models, 4)
	assert.Equal(t, "gpt-5", snapshot.Models[0].ModelName)
	assert.Equal(t, "claude-sonnet", snapshot.Models[1].ModelName)
	assert.Empty(t, snapshot.Models[3].ModelName, "entries without a model remain a distinct OTHER group")
	assert.Equal(t, "53.33%", snapshot.Models[0].ChargeShare)
	assert.Equal(t, int64(1), snapshot.Models[0].ActiveDays)
	assert.Equal(t, int64(1), snapshot.Models[1].ChargeCount)
	assert.Equal(t, int64(1), snapshot.Models[1].RefundCount)
	assert.Equal(t, int64(2), snapshot.Models[1].ActiveDays)
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{
		ID: "DEMO-202608-MODEL-SUMMARY", UserID: 4, Month: snapshot.Month, Revision: 1,
		StartAt: start, EndAt: end, CreatedAt: end + 2*86400,
		Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]),
	}
	original, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	retry, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.Equal(t, original, retry)
	pdfHash := sha256.Sum256(original)
	statement.PDFSHA256 = hex.EncodeToString(pdfHash[:])
	statement.ManifestSHA256 = strings.Repeat("a", 64)
	statement.ConfirmedAt = statement.CreatedAt + 86400
	receipt, err := RenderBillingStatementPDF(statement, snapshot, true)
	require.NoError(t, err)
	assert.NotEqual(t, original, receipt)
	if output := os.Getenv("BILLING_PDF_V8_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.WriteFile(output, receipt, 0600))
	}
}

func TestBillingPDFV8PaginatesLargeModelSummary(t *testing.T) {
	start, end, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	snapshot, err := BuildBillingSnapshot(
		&model.BillingAccount{UserID: 4, CompanyTitle: "分页测试客户", TaxID: "TEST", AccountingStartAt: start},
		"2026-08",
		[]model.BillingHour{{Hour: start, Charge: 60, Count: 60}},
	)
	require.NoError(t, err)
	snapshot.PDFTemplateVersion = 8
	snapshot.PDFLogoPNG = billingPlatformMarkV2
	snapshot.Issuer = "聚合算力网关"
	snapshot.Username = "pagination_customer"
	totals := make([]model.BillingModelTotal, 0, 60)
	for index := 0; index < 60; index++ {
		totals = append(totals, model.BillingModelTotal{ModelName: fmt.Sprintf("model-%02d", index), Charge: 1, Count: 1, ChargeCount: 1, ActiveDays: 1, FirstPosted: start, LastPosted: start})
	}
	require.NoError(t, attachBillingModelRows(snapshot, totals))
	statement := &model.BillingStatement{
		ID: "model-pagination", UserID: 4, Month: "2026-08", Revision: 1,
		StartAt: start, EndAt: end, CreatedAt: end + 86400,
		SnapshotSHA256: strings.Repeat("1", 64),
	}
	pdf, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, strings.Count(string(pdf), "\n  /Type /Page\n"), 5, "all model rows must continue onto additional PDF pages")
	if output := os.Getenv("BILLING_PDF_V8_PAGINATION_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.WriteFile(output, pdf, 0600))
	}
	snapshot.PDFTemplateVersion = 9
	statement.ExcelSHA256 = strings.Repeat("2", 64)
	snapshot.RateReferenceAt = statement.CreatedAt
	for i := range snapshot.Models {
		snapshot.Models[i].ModelName = fmt.Sprintf("企业超长上下文分析模型专业版本-model-%02d", i)
		snapshot.Models[i].CurrentRates = []BillingGroupRate{{Group: "vip", Ratio: "0.65"}}
	}
	expanded, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err, "the expanded rate/price table must paginate without clipping rows or notes")
	assert.GreaterOrEqual(t, strings.Count(string(expanded), "\n  /Type /Page\n"), strings.Count(string(pdf), "\n  /Type /Page\n"), "wrapped model names must retain every row across pages")
	if output := os.Getenv("BILLING_PDF_V9_PAGINATION_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.WriteFile(output, expanded, 0600))
	}
	snapshot.PDFTemplateVersion = 10
	for i := range snapshot.Models {
		snapshot.Models[i].BillingGroup = strings.Repeat("历史计费组", 8)
		snapshot.Models[i].BillingRate = "0.86"
	}
	historical, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err, "historical group descriptions must paginate without clipping totals")
	if output := os.Getenv("BILLING_PDF_V10_PAGINATION_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.WriteFile(output, historical, 0600))
	}
	snapshot.PDFTemplateVersion = 11
	for index := range snapshot.Models {
		snapshot.Models[index].ModelName = fmt.Sprintf("model-%02d-very-long-model-name-to-check-wrapping-and-pagination", index)
	}
	wrapped, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err, "long model names must wrap without clipping totals or losing rows")
	if output := os.Getenv("BILLING_PDF_V11_PAGINATION_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.WriteFile(output, wrapped, 0600))
	}
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
	snapshot.PDFTemplateVersion = 8
	snapshot.PDFLogoPNG = billingPlatformMarkV2
	require.NoError(t, attachBillingModelRows(snapshot, []model.BillingModelTotal{{ModelName: "model", Charge: 100, Count: 1, ChargeCount: 1, ActiveDays: 1, FirstPosted: posted, LastPosted: posted}}))
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

func TestBillingArchiveReconcilesCaseDistinctModelNames(t *testing.T) {
	truncate(t)
	posted := int64(1769875200)
	for _, entry := range []model.BillingEntry{
		{EventKey: "upper-charge", UserID: 46, Sequence: 1, PostedAt: posted, Kind: "usage", Quota: 20, ModelName: "Foo"},
		{EventKey: "lower-charge", UserID: 46, Sequence: 2, PostedAt: posted + 1, Kind: "usage", Quota: 30, ModelName: "foo"},
		{EventKey: "upper-refund", UserID: 46, Sequence: 3, PostedAt: posted + 86400, Kind: "refund", Quota: -5, ModelName: "Foo"},
	} {
		require.NoError(t, model.DB.Create(&entry).Error)
	}
	snapshot, err := BuildBillingSnapshot(&model.BillingAccount{UserID: 46, CompanyTitle: "测试企业", TaxID: "TEST"}, "2026-02", []model.BillingHour{
		{Hour: posted, Charge: 50, Count: 2},
		{Hour: posted + 86400, Refund: 5, Count: 1},
	})
	require.NoError(t, err)
	snapshot.PDFTemplateVersion = 8
	snapshot.PDFLogoPNG = billingPlatformMarkV2
	require.NoError(t, attachBillingModelRows(snapshot, []model.BillingModelTotal{
		{ModelName: "Foo", Charge: 20, Refund: 5, Count: 2, ChargeCount: 1, RefundCount: 1, ActiveDays: 2, FirstPosted: posted, LastPosted: posted + 86400},
		{ModelName: "foo", Charge: 30, Count: 1, ChargeCount: 1, ActiveDays: 1, FirstPosted: posted + 1, LastPosted: posted + 1},
	}))
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "case-distinct-models", UserID: 46, Month: "2026-02", Revision: 1, Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]), StartAt: posted, EndAt: 1772294400, CreatedAt: 1772380800, FromSequence: 1, ToSequence: 3}
	store := &memoryBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))
	assert.NotEmpty(t, store.files["pdf"])
	assert.NotEmpty(t, store.files["manifest"])
}

func TestBillingArchiveRejectsModelSummaryThatDoesNotMatchFrozenDetails(t *testing.T) {
	truncate(t)
	posted := int64(1769875200)
	require.NoError(t, model.DB.Create(&model.BillingEntry{EventKey: "model-source", UserID: 43, Sequence: 1, PostedAt: posted, Kind: "usage", Quota: 100, ModelName: "model-a"}).Error)
	snapshot, err := BuildBillingSnapshot(&model.BillingAccount{UserID: 43, CompanyTitle: "测试企业", TaxID: "TEST"}, "2026-02", []model.BillingHour{{Hour: posted, Charge: 100, Count: 1}})
	require.NoError(t, err)
	snapshot.PDFTemplateVersion = 8
	snapshot.PDFLogoPNG = billingPlatformMarkV2
	require.NoError(t, attachBillingModelRows(snapshot, []model.BillingModelTotal{{ModelName: "model-b", Charge: 100, Count: 1, ChargeCount: 1, ActiveDays: 1, FirstPosted: posted, LastPosted: posted}}))
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "wrong-model-summary", UserID: 43, Month: "2026-02", Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]), StartAt: posted, EndAt: 1772294400, FromSequence: 1, ToSequence: 1}
	store := &memoryBillingArchive{files: map[string][]byte{}}
	err = BuildBillingArchive(context.Background(), statement, store)
	require.ErrorContains(t, err, "model details do not reconcile")
	assert.NotContains(t, store.files, "pdf")
	assert.NotContains(t, store.files, "manifest")
}

func TestBillingArchiveRejectsModelMetadataThatDoesNotMatchFrozenDetails(t *testing.T) {
	truncate(t)
	posted := int64(1769875200)
	for index, at := range []int64{posted, posted + 1} {
		require.NoError(t, model.DB.Create(&model.BillingEntry{EventKey: fmt.Sprintf("model-metadata-%d", index), UserID: 44, Sequence: int64(index + 1), PostedAt: at, Kind: "usage", Quota: 50, ModelName: "model-a"}).Error)
	}
	snapshot, err := BuildBillingSnapshot(&model.BillingAccount{UserID: 44, CompanyTitle: "测试企业", TaxID: "TEST"}, "2026-02", []model.BillingHour{{Hour: posted, Charge: 100, Count: 2}})
	require.NoError(t, err)
	snapshot.PDFTemplateVersion = 8
	snapshot.PDFLogoPNG = billingPlatformMarkV2
	require.NoError(t, attachBillingModelRows(snapshot, []model.BillingModelTotal{{ModelName: "model-a", Charge: 100, Count: 2, ChargeCount: 2, ActiveDays: 2, FirstPosted: posted, LastPosted: posted + 1}}))
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "wrong-model-metadata", UserID: 44, Month: "2026-02", Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]), StartAt: posted, EndAt: 1772294400, FromSequence: 1, ToSequence: 2}
	store := &memoryBillingArchive{files: map[string][]byte{}}
	err = BuildBillingArchive(context.Background(), statement, store)
	require.ErrorContains(t, err, "model details do not reconcile")
	assert.NotContains(t, store.files, "pdf")
	assert.NotContains(t, store.files, "manifest")
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

func TestAttachBillingAccountingDerivesProfitFromNetRevenue(t *testing.T) {
	row := BillingRow{ChargeQuota: 1000, RefundQuota: 200}
	currency := BillingCurrency{Code: "USD", Symbol: "$", Rate: "1", QuotaPerUnit: "100"}

	require.NoError(t, attachBillingAccounting(&row, 600, currency))
	assert.Equal(t, int64(600), row.CostQuota)
	assert.Equal(t, int64(200), row.ProfitQuota)
	assert.Equal(t, "6.000000", row.Cost)
	assert.Equal(t, "2.000000", row.Profit)
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
	_, err = PrepareBillingStatement(91, 1, 1, "2020-02")
	require.ErrorContains(t, err, "do not reconcile", "a missing boundary summary must not silently remove official entries from evidence")
	var statements int64
	require.NoError(t, model.DB.Model(&model.BillingStatement{}).Where("user_id = ?", 91).Count(&statements).Error)
	assert.Zero(t, statements, "an inconsistent statement must not be persisted for archival")
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

func TestBillingArchiveExcelListsFrozenRecordsWithTokenAndGroup(t *testing.T) {
	truncate(t)
	posted := int64(1769875200)
	require.NoError(t, model.DB.Create(&model.Token{Id: 7, UserId: 47, Key: "archive-detail-key", Name: "prod-key"}).Error)
	require.NoError(t, model.DB.Create(&model.BillingOperation{ID: "op-detail", UserID: 47, State: "settled", Reserved: 40, Actual: 100, RequestID: "req-detail", ModelName: "model", TokenID: 7}).Error)
	require.NoError(t, model.DB.Create(&model.BillingEntry{EventKey: "settle:op-detail", UserID: 47, Sequence: 1, PostedAt: posted, Kind: "usage", Quota: 100, RequestID: "req-detail", ModelName: "model", TokenID: 7, BillingGroup: "standard", BillingRate: "0.8"}).Error)
	require.NoError(t, model.DB.Create(&model.BillingEntry{EventKey: "detail-sheet-fallback", UserID: 47, Sequence: 2, PostedAt: posted + 60, Kind: "usage", Quota: 50, RequestID: "req-fallback", ModelName: "model", TokenID: 7}).Error)
	require.NoError(t, model.DB.Create(&model.Log{UserId: 47, RequestId: "req-fallback", Group: "selected-group", Type: model.LogTypeConsume, IsStream: true, PromptTokens: 77, CompletionTokens: 5, UseTime: 3, Ip: "10.1.1.1", CreatedAt: posted + 60}).Error)
	snapshot, err := BuildBillingSnapshot(&model.BillingAccount{UserID: 47, CompanyTitle: "测试企业", TaxID: "TEST"}, "2026-02", []model.BillingHour{{Hour: posted, Charge: 150, Count: 2}})
	require.NoError(t, err)
	snapshot.PDFTemplateVersion = 9
	snapshot.PDFLogoPNG = billingPlatformMarkV2
	snapshot.Username = "detail_customer"
	snapshot.DetailSheet = true
	require.NoError(t, attachBillingModelRows(snapshot, []model.BillingModelTotal{{ModelName: "model", Charge: 150, Count: 2, ChargeCount: 2, ActiveDays: 1, FirstPosted: posted, LastPosted: posted + 60}}))
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "detail-sheet", UserID: 47, Month: "2026-02", Revision: 1, Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]), StartAt: posted, EndAt: 1772294400, CreatedAt: 1772380800, FromSequence: 1, ToSequence: 2}
	store := &memoryBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))

	book, err := excelize.OpenReader(bytes.NewReader(store.files["xlsx"]))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("详细日志")
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, []string{"detail_customer", "prod-key", "standard", "model"}, rows[1][1:5])
	assert.Equal(t, "req-detail", rows[1][16])
	assert.Equal(t, []string{"0.000080", "0.000200", "0.000000", "0.000000", "0.000200"}, rows[1][9:14], "the reservation, shown apart, and the one charge the ledger holds")
	assert.Equal(t, []string{"", "0.000100", "0.000000", "0.000000", "0.000100"}, rows[2][9:14], "an entry without a reservation lists none")
	assert.Equal(t, "selected-group", rows[2][3], "entries without a recorded group use the usage log's group")
	assert.Equal(t, []string{"是", "77", "5"}, rows[2][5:8], "the usage log's stream flag and tokens")
	assert.Equal(t, []string{"3", "10.1.1.1"}, rows[2][14:16], "the usage log's duration and IP")
}

func TestBillingArchiveExcelMergesTaskHoldWithItsRefund(t *testing.T) {
	truncate(t)
	posted := int64(1769875200)
	// A task refund is booked under the task ID, not under the submit request.
	require.NoError(t, model.DB.Create(&model.BillingEntry{EventKey: "task-hold", UserID: 48, Sequence: 1, PostedAt: posted, Kind: "usage", Quota: 100, RequestID: "req-submit", ModelName: "video-model"}).Error)
	require.NoError(t, model.DB.Create(&model.BillingEntry{EventKey: "task-refund", UserID: 48, Sequence: 2, PostedAt: posted + 600, Kind: "task_adjustment", Quota: -30, RequestID: "task_abc", ModelName: "video-model"}).Error)
	require.NoError(t, model.DB.Create(&model.Log{UserId: 48, RequestId: "req-submit", Group: "video", Type: model.LogTypeConsume, Other: `{"task_id":"task_abc","group_ratio":0.86}`, CreatedAt: posted}).Error)
	require.NoError(t, model.DB.AutoMigrate(&model.Task{}))
	require.NoError(t, model.DB.Create(&model.Task{TaskID: "task_abc", UserId: 48, SubmitTime: posted, FinishTime: posted + 300, Data: []byte(`{"usage":{"total_tokens":1234}}`)}).Error)
	snapshot, err := BuildBillingSnapshot(&model.BillingAccount{UserID: 48, CompanyTitle: "测试企业", TaxID: "TEST"}, "2026-02", []model.BillingHour{{Hour: posted, Charge: 100, Count: 1}, {Hour: posted + 600 - (posted+600)%3600, Refund: 30, Count: 1}})
	require.NoError(t, err)
	snapshot.PDFTemplateVersion = 9
	snapshot.PDFLogoPNG = billingPlatformMarkV2
	snapshot.DetailSheet = true
	require.NoError(t, attachBillingModelRows(snapshot, []model.BillingModelTotal{{ModelName: "video-model", Charge: 100, Refund: 30, Count: 2, ChargeCount: 1, RefundCount: 1, ActiveDays: 1, FirstPosted: posted, LastPosted: posted + 600}}))
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "task-merge", UserID: 48, Month: "2026-02", Revision: 1, Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]), StartAt: posted, EndAt: 1772294400, CreatedAt: 1772380800, FromSequence: 1, ToSequence: 2}
	store := &memoryBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))

	book, err := excelize.OpenReader(bytes.NewReader(store.files["xlsx"]))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("详细日志")
	require.NoError(t, err)

	require.Len(t, rows, 2, "the hold and its refund are one row")
	assert.Equal(t, []string{"时间", "用户", "令牌", "所选分组", "模型", "输出 Token", "计费费率", "首次扣费", "补扣金额", "退款金额", "实际扣费金额", "耗时(秒)", "请求ID", "扣费说明"}, rows[0])
	assert.Equal(t, "req-submit", rows[1][12])
	assert.Equal(t, "video", rows[1][3], "the refund entry lends no group; the usage log does")
	assert.Equal(t, []string{"1,234", "86.00%"}, rows[1][5:7], "the task record's tokens and the rate from the usage log")
	assert.Equal(t, []string{"0.000200", "0.000000", "0.000060", "0.000140"}, rows[1][7:11], "the hold, no extra charge, the refund, and what the request cost")
	assert.Equal(t, "300", rows[1][11], "the task's run time")
}

func TestBillingArchiveExcelMergesAnOlderTaskHoldThatRecordedNoTaskID(t *testing.T) {
	truncate(t)
	posted := int64(1769875200)
	// The older video logs recorded no task ID; only the refund's log does.
	require.NoError(t, model.DB.Create(&model.BillingEntry{EventKey: "task-hold", UserID: 49, Sequence: 1, PostedAt: posted, Kind: "usage", Quota: 100, RequestID: "req-submit", ModelName: "video-model", SourceLogID: 501}).Error)
	require.NoError(t, model.DB.Create(&model.BillingEntry{EventKey: "task-refund", UserID: 49, Sequence: 2, PostedAt: posted + 3700, Kind: "history_refund", Quota: -30, RequestID: "req-refund", ModelName: "video-model", SourceLogID: 502}).Error)
	require.NoError(t, model.DB.Create(&model.Log{Id: 501, UserId: 49, RequestId: "req-submit", Type: model.LogTypeConsume, Other: `{"is_task":true}`, CreatedAt: posted}).Error)
	require.NoError(t, model.DB.Create(&model.Log{Id: 502, UserId: 49, RequestId: "req-refund", Type: model.LogTypeRefund, Other: `{"task_id":"task_old"}`, CreatedAt: posted + 3700}).Error)
	require.NoError(t, model.DB.AutoMigrate(&model.Task{}))
	require.NoError(t, model.DB.Create(&model.Task{TaskID: "task_old", UserId: 49, SubmitTime: posted + 1, FinishTime: posted + 3700}).Error)
	snapshot, err := BuildBillingSnapshot(&model.BillingAccount{UserID: 49, CompanyTitle: "测试企业", TaxID: "TEST"}, "2026-02", []model.BillingHour{{Hour: posted, Charge: 100, Count: 1}, {Hour: posted + 3600, Refund: 30, Count: 1}})
	require.NoError(t, err)
	snapshot.PDFTemplateVersion = 9
	snapshot.PDFLogoPNG = billingPlatformMarkV2
	snapshot.DetailSheet = true
	require.NoError(t, attachBillingModelRows(snapshot, []model.BillingModelTotal{{ModelName: "video-model", Charge: 100, Refund: 30, Count: 2, ChargeCount: 1, RefundCount: 1, ActiveDays: 1, FirstPosted: posted, LastPosted: posted + 3700}}))
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	hash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "task-merge-old", UserID: 49, Month: "2026-02", Revision: 1, Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(hash[:]), StartAt: posted, EndAt: 1772294400, CreatedAt: 1772380800, FromSequence: 1, ToSequence: 2}
	store := &memoryBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))

	book, err := excelize.OpenReader(bytes.NewReader(store.files["xlsx"]))
	require.NoError(t, err)
	defer book.Close()
	rows, err := book.GetRows("详细日志")
	require.NoError(t, err)

	require.Len(t, rows, 2, "the hold and its refund are one row although the hold's log has no task ID")
	assert.Equal(t, "req-submit", rows[1][len(rows[1])-2])
}
