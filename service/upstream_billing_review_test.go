package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type upstreamSupplierFixtureStore struct{ fail error }

func (store upstreamSupplierFixtureStore) PutUpstreamAttachment(_ context.Context, attachment *model.UpstreamBillingAttachment, body []byte) error {
	if store.fail != nil {
		return store.fail
	}
	hash := sha256.Sum256(body)
	if hex.EncodeToString(hash[:]) != attachment.SHA256 {
		return model.ErrBillingEvidenceIntegrity
	}
	return model.DB.Create(&model.StorageObject{ObjectID: attachment.ObjectID, ObjectKey: attachment.ObjectKey, OwnerUserID: attachment.UploadedBy, StorageProfileID: attachment.StorageProfileID, Purpose: model.StorageObjectPurposeBillingArchive, SHA256: attachment.SHA256, Size: int64(len(body)), Status: model.StorageObjectStatusUploaded}).Error
}

func TestUpstreamSupplierAttachmentValidatesBytesAndDisplayNames(t *testing.T) {
	statement, snapshot, _ := frozenUpstreamBillingFixture(t)
	pdf, err := RenderUpstreamBillingStatementPDF(statement, snapshot, false)
	require.NoError(t, err)
	workbook, err := RenderUpstreamBillingStatementExcel(statement, snapshot)
	require.NoError(t, err)
	var pngImage, jpegImage bytes.Buffer
	require.NoError(t, png.Encode(&pngImage, image.NewRGBA(image.Rect(0, 0, 2, 2))))
	require.NoError(t, jpeg.Encode(&jpegImage, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil))
	for _, test := range []struct {
		name, filename, expectedName, contentType string
		body                                      []byte
	}{
		{"pdf", "supplier.PDF", "supplier.PDF", "application/pdf", pdf},
		{"xlsx", "supplier.xlsx", "supplier.xlsx", billingExcelContentType, workbook},
		{"csv basename", `C:\fakepath\供应商.csv`, "供应商.csv", "text/csv; charset=utf-8", []byte("model,cost\nA,7\n")},
		{"png", "scan.png", "scan.png", "image/png", pngImage.Bytes()},
		{"jpeg", "scan.jpeg", "scan.jpeg", "image/jpeg", jpegImage.Bytes()},
	} {
		t.Run(test.name, func(t *testing.T) {
			name, contentType, err := ValidateUpstreamBillingAttachment(test.filename, test.body)
			require.NoError(t, err)
			assert.Equal(t, test.expectedName, name)
			assert.Equal(t, test.contentType, contentType)
		})
	}
	oversizedPNG := append([]byte(nil), pngImage.Bytes()...)
	binary.BigEndian.PutUint32(oversizedPNG[16:20], 40001)
	binary.BigEndian.PutUint32(oversizedPNG[29:33], crc32.ChecksumIEEE(oversizedPNG[12:29]))
	for _, test := range []struct {
		name, filename string
		body           []byte
	}{
		{"extension disguise", "invoice.pdf", pngImage.Bytes()},
		{"image extension disguise", "scan.jpeg", pngImage.Bytes()},
		{"macro extension", "invoice.xlsm", workbook},
		{"legacy macro container", "invoice.xls", workbook},
		{"fake pdf magic", "fake.pdf", []byte("%PDF-1.7\nnot a document\n%%EOF")},
		{"truncated pdf", "partial.pdf", pdf[:len(pdf)-20]},
		{"csv nul", "supplier.csv", []byte("a,b\n1,\x00\n")},
		{"csv invalid utf8", "supplier.csv", []byte{0xff, 0xfe}},
		{"malformed csv", "supplier.csv", []byte("a,b\n\"unterminated,1")},
		{"control filename", "invoice\n.pdf", pdf},
		{"long filename", strings.Repeat("x", 252) + ".csv", []byte("a,1")},
		{"image pixel bound", "large.png", oversizedPNG},
		{"empty", "empty.csv", nil},
		{"file size", "huge.csv", bytes.Repeat([]byte{'x'}, int(model.MaxUpstreamBillingAttachmentBytes)+1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := ValidateUpstreamBillingAttachment(test.filename, test.body)
			assert.ErrorIs(t, err, model.ErrUpstreamBillingInvalidAttachment)
		})
	}
	archive, err := zip.NewReader(bytes.NewReader(workbook), int64(len(workbook)))
	require.NoError(t, err)
	var macro bytes.Buffer
	writer := zip.NewWriter(&macro)
	for _, file := range archive.File {
		reader, err := file.Open()
		require.NoError(t, err)
		entry, err := writer.Create(file.Name)
		require.NoError(t, err)
		_, err = io.Copy(entry, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
	}
	entry, err := writer.Create("xl/vbaProject.bin")
	require.NoError(t, err)
	_, err = entry.Write([]byte("macro"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	_, _, err = ValidateUpstreamBillingAttachment("hidden-macro.xlsx", macro.Bytes())
	assert.ErrorIs(t, err, model.ErrUpstreamBillingInvalidAttachment)
	var bomb bytes.Buffer
	compressed := zip.NewWriter(&bomb)
	entry, err = compressed.CreateRaw(&zip.FileHeader{Name: "xl/workbook.xml", Method: zip.Deflate, CompressedSize64: 2, UncompressedSize64: 101 * 1024 * 1024})
	require.NoError(t, err)
	_, err = entry.Write([]byte{3, 0})
	require.NoError(t, err)
	require.NoError(t, compressed.Close())
	_, _, err = ValidateUpstreamBillingAttachment("inflated.xlsx", bomb.Bytes())
	assert.ErrorIs(t, err, model.ErrUpstreamBillingInvalidAttachment)
}

func TestUpstreamSupplierAttachmentReviewReceiptPreservesOriginalDocuments(t *testing.T) {
	upstreamBillingServiceFixture(t)
	require.NoError(t, model.DB.AutoMigrate(&model.UpstreamBillingAttachment{}, &model.StorageObject{}))
	statement, snapshot, records := frozenUpstreamBillingFixture(t)
	statement.ID = "REVIEW-RECEIPT-UNIT-SAMPLE"
	for index := range records {
		records[index].StatementID = statement.ID
	}
	require.NoError(t, model.DB.Create(&records).Error)
	original := &memoryUpstreamBillingArchive{files: map[string][]byte{}}
	require.NoError(t, BuildUpstreamBillingArchive(context.Background(), statement, original))
	statement.Status = model.StatementDraft
	require.NoError(t, model.DB.Create(statement).Error)
	require.NoError(t, model.DB.Create(&original.artifacts).Error)
	pdfBefore := append([]byte(nil), original.files["pdf-0"]...)
	excelBefore := append([]byte(nil), original.files["xlsx-0"]...)
	attachment, err := UploadUpstreamBillingAttachmentWithStore(context.Background(), statement.ID, 4, "供应商账单📄.csv", []byte("model,cost\nA,26.6\n"), upstreamSupplierFixtureStore{})
	require.NoError(t, err)
	assert.Equal(t, model.UpstreamBillingAttachmentReady, attachment.Status)
	for index := 2; index <= model.MaxUpstreamBillingAttachments; index++ {
		_, err := UploadUpstreamBillingAttachmentWithStore(context.Background(), statement.ID, 4, fmt.Sprintf("%02d-%s.csv", index, strings.Repeat("供应商成本对账附件", 6)), []byte("model,cost\nA,26.6\n"), upstreamSupplierFixtureStore{})
		require.NoError(t, err)
	}
	review, err := model.GetUpstreamBillingReview(statement.ID)
	require.NoError(t, err)
	confirmed, err := model.ActUpstreamBillingStatementWithReview(statement.ID, "confirm", statement.ManifestSHA256, review.EvidenceSHA256, "reviewed", "real-test-session", 8)
	require.NoError(t, err)
	receipt, err := RenderUpstreamBillingStatementPDF(confirmed, snapshot, true)
	require.NoError(t, err)
	retry, err := RenderUpstreamBillingStatementPDF(confirmed, snapshot, true)
	require.NoError(t, err)
	assert.Equal(t, receipt, retry)
	text := billingPDFPageText(t, receipt, 0)
	assert.Contains(t, text, review.EvidenceSHA256)
	assert.Contains(t, text, attachment.SHA256)
	assert.Contains(t, text, "供应商账单[U+1F4C4].csv")
	assert.Contains(t, text, "#8")
	assert.NotEmpty(t, billingPDFPageText(t, receipt, 1), "the maximum supplier attachment list must continue on additional pages")
	pageCount := len(regexp.MustCompile(`(?m)^  /Type /Page$`).FindAll(receipt, -1))
	for page := 0; page < pageCount; page++ {
		pageText := billingPDFPageText(t, receipt, page)
		names := len(regexp.MustCompile(`附件 [0-9]{2} `).FindAllString(pageText, -1))
		assert.Equal(t, names, strings.Count(pageText, "上传人 #4"), "each filename, digest and uploader block must stay on one page")
	}
	pdfAfter, err := RenderUpstreamBillingStatementPDF(confirmed, snapshot, false)
	require.NoError(t, err)
	excelAfter, err := RenderUpstreamBillingStatementExcel(confirmed, snapshot)
	require.NoError(t, err)
	assert.Equal(t, pdfBefore, pdfAfter)
	assert.Equal(t, excelBefore, excelAfter)
	_, err = UploadUpstreamBillingAttachmentWithStore(context.Background(), statement.ID, 4, "late.csv", []byte("a,1"), upstreamSupplierFixtureStore{})
	assert.ErrorIs(t, err, model.ErrBillingConflict)
	if output := os.Getenv("UPSTREAM_REVIEW_TEST_OUTPUT"); output != "" {
		require.NoError(t, os.MkdirAll(output, 0700))
		require.NoError(t, os.WriteFile(filepath.Join(output, "review-receipt-unit-sample.pdf"), receipt, 0600))
		require.NoError(t, os.WriteFile(filepath.Join(output, "review-evidence-unit-sample.json"), []byte(confirmed.ReviewEvidenceSnapshot), 0600))
	}
}

func TestUpstreamSupplierUploadFailureLeavesAuditedFailedEvidence(t *testing.T) {
	upstreamBillingServiceFixture(t)
	require.NoError(t, model.DB.AutoMigrate(&model.UpstreamBillingAttachment{}, &model.StorageObject{}))
	statement, _, _ := frozenUpstreamBillingFixture(t)
	statement.Status = model.StatementDraft
	require.NoError(t, model.DB.Create(statement).Error)
	networkFailure := errors.New("storage unavailable")
	_, err := UploadUpstreamBillingAttachmentWithStore(context.Background(), statement.ID, 4, "supplier.csv", []byte("a,1"), upstreamSupplierFixtureStore{fail: networkFailure})
	assert.ErrorIs(t, err, networkFailure)
	review, err := model.GetUpstreamBillingReview(statement.ID)
	require.NoError(t, err)
	assert.Empty(t, review.Attachments)
	assert.Zero(t, review.PendingUploads)
	var failed model.UpstreamBillingAttachment
	require.NoError(t, model.DB.Where("statement_id = ?", statement.ID).First(&failed).Error)
	assert.Equal(t, model.UpstreamBillingAttachmentFailed, failed.Status)
	var audit model.UpstreamBillingStatementEvent
	require.NoError(t, model.DB.Where("statement_id = ? AND action = ?", statement.ID, "attachment_failed").First(&audit).Error)
	assert.Equal(t, 4, audit.ActorID)
	assert.Contains(t, audit.Note, failed.ID)
	assert.Contains(t, audit.Note, failed.OriginalFilename)
}
