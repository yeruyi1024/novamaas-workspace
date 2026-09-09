package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type billingLogoTransport struct {
	body          []byte
	status        int
	contentLength int64
	requests      []*http.Request
}

func (transport *billingLogoTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	transport.requests = append(transport.requests, request)
	return &http.Response{StatusCode: transport.status, ContentLength: transport.contentLength, Body: io.NopCloser(bytes.NewReader(transport.body)), Header: make(http.Header)}, nil
}

func configureBillingBranding(t *testing.T) *billingLogoTransport {
	t.Helper()
	savedName, savedLogo, savedFooter, savedAddress, savedClient := common.SystemName, common.Logo, common.Footer, system_setting.ServerAddress, httpClient
	fetchSetting := system_setting.GetFetchSetting()
	savedFetchSetting := *fetchSetting
	common.SystemName, common.Logo, common.Footer = "示例服务平台", "", ""
	system_setting.ServerAddress = "https://platform.example"
	fetchSetting.EnableSSRFProtection = false
	transport := &billingLogoTransport{status: http.StatusOK, body: billingPlatformMarkV2}
	httpClient = &http.Client{Transport: transport}
	t.Cleanup(func() {
		common.SystemName, common.Logo, common.Footer, system_setting.ServerAddress, httpClient = savedName, savedLogo, savedFooter, savedAddress, savedClient
		*fetchSetting = savedFetchSetting
	})
	return transport
}

func TestBillingBrandingCapturesConfiguredLogoAndFooterWithoutURLSecrets(t *testing.T) {
	transport := configureBillingBranding(t)
	common.Logo = "/company-logo.png?signature=do-not-archive"
	common.Footer = `<p>示例科技有限公司 &amp; 服务平台</p><script>do-not-render</script><div>ICP备案 123</div>`
	logo := image.NewNRGBA(image.Rect(0, 0, 300, 150))
	draw.Draw(logo, logo.Bounds(), image.NewUniform(color.NRGBA{R: 10, G: 40, B: 90, A: 255}), image.Point{}, draw.Src)
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, logo))
	transport.body = encoded.Bytes()

	branding, err := captureBillingDocumentBranding(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "示例服务平台", branding.Issuer)
	assert.Equal(t, "示例科技有限公司 & 服务平台 ICP备案 123", branding.Footer)
	decoded, err := png.Decode(bytes.NewReader(branding.LogoPNG))
	require.NoError(t, err)
	assert.Equal(t, image.Rect(0, 0, 256, 128), decoded.Bounds(), "wide logos retain their aspect ratio")
	assert.Equal(t, color.NRGBA{R: 10, G: 40, B: 90, A: 255}, color.NRGBAModel.Convert(decoded.At(120, 60)))
	require.Len(t, transport.requests, 1)
	assert.Equal(t, "https://platform.example/company-logo.png?signature=do-not-archive", transport.requests[0].URL.String())
	assert.Empty(t, transport.requests[0].Header.Get("Authorization"))
	body, err := common.Marshal(branding)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "signature")
	assert.NotContains(t, string(body), "do-not-render")
}

func TestBillingBrandingUsesDefaultOnlyWhenLogoIsUnconfigured(t *testing.T) {
	transport := configureBillingBranding(t)
	branding, err := captureBillingDocumentBranding(context.Background())
	require.NoError(t, err)
	assert.Empty(t, branding.Footer)
	assert.Empty(t, transport.requests, "the default logo needs no network request")
	_, err = png.Decode(bytes.NewReader(branding.LogoPNG))
	require.NoError(t, err)

	common.Logo = "https://platform.example/broken.png"
	transport.status = http.StatusNotFound
	_, err = captureBillingDocumentBranding(context.Background())
	assert.ErrorContains(t, err, "404", "a configured but broken logo must not silently issue the wrong brand")
}

func TestBillingBrandingRejectsUnsafeAndOversizedImages(t *testing.T) {
	for _, test := range []struct {
		name, logo, wantError string
		body                  []byte
		length                int64
		protection            bool
	}{
		{name: "local files", logo: "file:///etc/passwd", wantError: "HTTP(S)"},
		{name: "URL credentials", logo: "https://user:password@platform.example/logo.png", wantError: "credentials"},
		{name: "private metadata", logo: "http://169.254.169.254/latest/meta-data", protection: true, wantError: "security policy"},
		{name: "oversized response", logo: "https://platform.example/logo.png", length: 2097153, wantError: "2 MiB"},
		{name: "HTML instead of image", logo: "https://platform.example/logo.png", body: []byte("<html>login required</html>"), wantError: "valid PNG"},
	} {
		t.Run(test.name, func(t *testing.T) {
			transport := configureBillingBranding(t)
			common.Logo = test.logo
			transport.body, transport.contentLength = test.body, test.length
			if test.protection {
				configureSSRFTestFetchSetting(t)
			}
			_, err := captureBillingDocumentBranding(context.Background())
			assert.ErrorContains(t, err, test.wantError)
			if test.protection {
				assert.Empty(t, transport.requests)
			}
		})
	}
	oversized := image.NewNRGBA(image.Rect(0, 0, 4097, 1))
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, oversized))
	_, err := normalizeBillingLogo(encoded.Bytes())
	assert.ErrorContains(t, err, "dimensions")
}

func TestBillingFooterPreservesTextAndRejectsExcessiveContent(t *testing.T) {
	for _, test := range []struct{ source, expected string }{
		{"", ""},
		{"公司名称\t\n备案号", "公司名称 备案号"},
		{"<div>公司 <a href=\"https://example.com\">备案号</a><br>经营许可</div><style>hidden</style>", "公司 备案号 经营许可"},
		{"<script>alert(1)</script><img src=\"https://example.com/track\">", ""},
	} {
		text, err := billingFooterText(test.source)
		require.NoError(t, err)
		assert.Equal(t, test.expected, text)
	}
	_, err := billingFooterText(strings.Repeat("企", 301))
	assert.ErrorContains(t, err, "300")
}

func TestBillingBrandingFailureDoesNotCreateADraft(t *testing.T) {
	truncate(t)
	seedUser(t, 94, 1000000)
	start, _, err := model.BillingMonthBounds("2020-02")
	require.NoError(t, err)
	require.NoError(t, model.DB.Create(&model.BillingAccount{UserID: 94, AccountingStartAt: start, StartSequence: 1, CompanyTitle: "Test", TaxID: "TEST"}).Error)
	transport := configureBillingBranding(t)
	common.Logo = "https://platform.example/missing.png"
	transport.status = http.StatusNotFound
	_, err = PrepareBillingStatement(94, 1, 1, "2020-02")
	assert.ErrorIs(t, err, ErrBillingDocumentBranding)
	var count int64
	require.NoError(t, model.DB.Model(&model.BillingStatement{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestBillingDraftArchivesFrozenBrandingAfterSettingsChange(t *testing.T) {
	truncate(t)
	seedUser(t, 95, 1000000)
	start, _, err := model.BillingMonthBounds("2020-02")
	require.NoError(t, err)
	require.NoError(t, model.DB.Create(&model.BillingAccount{UserID: 95, AccountingStartAt: start, StartSequence: 1, CompanyTitle: "Test", TaxID: "TEST"}).Error)
	transport := configureBillingBranding(t)
	common.Logo, common.Footer = "https://platform.example/logo.png?signature=private", "<p>客户服务公司</p><p>备案号 123</p>"
	statement, err := PrepareBillingStatement(95, 1, 1, "2020-02")
	require.NoError(t, err)
	require.Len(t, transport.requests, 1)
	common.SystemName, common.Logo, common.Footer = "Later brand", "https://platform.example/missing.png", "Later footer"
	transport.status = http.StatusNotFound
	store := &memoryBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildBillingArchive(context.Background(), statement, store))
	var archived BillingSnapshot
	require.NoError(t, common.Unmarshal(store.files["snapshot"], &archived))
	assert.Equal(t, "示例服务平台", archived.Issuer)
	assert.Equal(t, "客户服务公司 备案号 123", archived.PDFFooter)
	assert.NotEmpty(t, archived.PDFLogoPNG)
	assert.NotContains(t, string(store.files["snapshot"]), "signature")
	assert.NotEmpty(t, store.files["pdf"])
	assert.NotEmpty(t, statement.PDFSHA256)
	assert.Len(t, transport.requests, 1, "the archive job must not re-fetch a changed logo")
}

func TestBillingPDFV3FreezesBrandingForBothOriginalAndReceipt(t *testing.T) {
	configureBillingBranding(t)
	common.Footer = "示例科技有限公司 | 备案信息与服务许可（演示数据）"
	branding, err := captureBillingDocumentBranding(context.Background())
	require.NoError(t, err)
	// Optional local fixture/output hooks allow visual QA without importing any
	// customer records or writing a real statement to the database or OSS.
	if input := os.Getenv("BILLING_PDF_V3_LOGO_INPUT"); input != "" {
		body, err := os.ReadFile(input)
		require.NoError(t, err)
		branding.LogoPNG, err = normalizeBillingLogo(body)
		require.NoError(t, err)
	}
	if issuer := os.Getenv("BILLING_PDF_V3_ISSUER"); issuer != "" {
		branding.Issuer = issuer
	}
	if footer := os.Getenv("BILLING_PDF_V3_FOOTER"); footer != "" {
		branding.Footer, err = billingFooterText(footer)
		require.NoError(t, err)
	}
	start, end, err := model.BillingMonthBounds("2026-08")
	require.NoError(t, err)
	currency := BillingCurrency{Code: "CNY", Symbol: "¥", Rate: "7", QuotaPerUnit: "500000"}
	snapshot, err := buildBillingSnapshot(&model.BillingAccount{UserID: 4, CompanyTitle: "上海示例科技有限公司（演示数据）", TaxID: "DEMO-NOT-A-REAL-TAX-ID", AccountingStartAt: start}, "2026-08", []model.BillingHour{{Hour: start + 9*3600, Charge: 500000, Count: 1}, {Hour: start + 14*86400, Refund: 125000, Count: 1}, {Hour: start + 19*86400, Charge: 750000, Count: 1}}, currency)
	require.NoError(t, err)
	snapshot.PDFTemplateVersion, snapshot.Issuer, snapshot.PDFLogoPNG, snapshot.PDFFooter = 3, branding.Issuer, branding.LogoPNG, branding.Footer
	snapshot.Username, snapshot.DisplayName = "demo_customer", "示例客户（仅展示版式）"
	body, err := common.Marshal(snapshot)
	require.NoError(t, err)
	require.Less(t, len(body), 60*1024)
	snapshotHash := sha256.Sum256(body)
	statement := &model.BillingStatement{ID: "DEMO-202608-DOCUMENT-PREVIEW", UserID: 4, Month: snapshot.Month, Revision: 1, StartAt: start, EndAt: end, CreatedAt: end + 2*86400, Snapshot: string(body), SnapshotSHA256: hex.EncodeToString(snapshotHash[:])}
	original, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.Contains(t, string(original), "/Subtype /Image")
	assert.Contains(t, string(original), "/FontFile2")
	common.SystemName, common.Logo, common.Footer = "Changed later", "https://unreachable.example/logo", "Changed footer"
	pdfHash := sha256.Sum256(original)
	statement.PDFSHA256, statement.ManifestSHA256 = hex.EncodeToString(pdfHash[:]), statement.SnapshotSHA256
	statement.ConfirmedAt = statement.CreatedAt + 7200
	unchanged, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.Equal(t, original, unchanged, "later configuration and confirmation never rewrite the original")
	receipt, err := RenderBillingStatementPDF(statement, snapshot, true)
	require.NoError(t, err)
	retry, err := RenderBillingStatementPDF(statement, snapshot, true)
	require.NoError(t, err)
	assert.Equal(t, receipt, retry, "the receipt uses the same frozen branding without network access")
	assert.NotEqual(t, original, receipt)
	if output := os.Getenv("BILLING_PDF_V3_OUTPUT_DIR"); output != "" {
		require.NoError(t, os.WriteFile(filepath.Join(output, BillingArtifactFilename(statement, "pdf", 0)), original, 0600))
		require.NoError(t, os.WriteFile(filepath.Join(output, BillingArtifactFilename(statement, "receipt", 0)), receipt, 0600))
	}
	snapshot.PDFFooter = ""
	noFooter, err := RenderBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	assert.NotEqual(t, original, noFooter, "configured footer is part of the actual PDF bytes")
	if output := os.Getenv("BILLING_PDF_V3_NO_FOOTER_OUTPUT"); output != "" {
		require.NoError(t, os.WriteFile(output, noFooter, 0600))
	}
	snapshot.PDFFooter = strings.Repeat("过长页脚", 1000)
	_, err = RenderBillingStatementPDF(statement, snapshot, false)
	assert.ErrorContains(t, err, "Footer exceeds", "never silently clip a legal footer")
}

func TestBillingArtifactFilenameUsesFrozenShanghaiTimestamp(t *testing.T) {
	statement := &model.BillingStatement{ID: "internal-uuid", Month: "2026-08", Revision: 2, CreatedAt: time.Date(2026, 9, 3, 0, 0, 0, 0, billingLocation).Unix(), ConfirmedAt: time.Date(2026, 9, 4, 8, 9, 10, 0, billingLocation).Unix()}
	for _, test := range []struct{ kind, expected string }{
		{"pdf", "月度对账单_20260903000000.pdf"},
		{"receipt", "对账确认回执_20260904080910.pdf"},
		{"details", "statement-2026-08-r2-details-0.jsonl.gz"},
		{"manifest", "statement-2026-08-r2-manifest-0.json"},
	} {
		name := BillingArtifactFilename(statement, test.kind, 0)
		assert.Equal(t, test.expected, name)
		header := mime.FormatMediaType("attachment", map[string]string{"filename": name})
		mediaType, params, err := mime.ParseMediaType(header)
		require.NoError(t, err)
		assert.Equal(t, "attachment", mediaType)
		assert.Equal(t, test.expected, params["filename"], "Chinese filename survives standard HTTP header encoding")
		assert.NotContains(t, header, statement.ID)
	}
}
