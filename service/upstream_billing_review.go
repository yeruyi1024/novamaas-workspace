package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	storageService "github.com/QuantumNous/new-api/service/storage"
)

type UpstreamBillingAttachmentWriter interface {
	PutUpstreamAttachment(context.Context, *model.UpstreamBillingAttachment, []byte) error
}

type upstreamBillingOSSAttachmentWriter struct{}

func (upstreamBillingOSSAttachmentWriter) PutUpstreamAttachment(ctx context.Context, attachment *model.UpstreamBillingAttachment, body []byte) error {
	store, err := storageService.OpenBillingArchiveStore(attachment.StorageProfileID, true)
	if err != nil {
		return err
	}
	return store.PutUpstreamAttachment(ctx, attachment, body)
}

func UploadUpstreamBillingAttachment(ctx context.Context, statementID string, actorID int, filename string, body []byte) (*model.UpstreamBillingAttachment, error) {
	return UploadUpstreamBillingAttachmentWithStore(ctx, statementID, actorID, filename, body, upstreamBillingOSSAttachmentWriter{})
}

func UploadUpstreamBillingAttachmentWithStore(ctx context.Context, statementID string, actorID int, filename string, body []byte, store UpstreamBillingAttachmentWriter) (*model.UpstreamBillingAttachment, error) {
	if store == nil {
		return nil, model.ErrUpstreamBillingInvalidAttachment
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	filename, contentType, err := ValidateUpstreamBillingAttachment(filename, body)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(body)
	attachment := &model.UpstreamBillingAttachment{StatementID: statementID, OriginalFilename: filename, ContentType: contentType, Size: int64(len(body)), SHA256: hex.EncodeToString(hash[:]), UploadedBy: actorID}
	if err := model.ReserveUpstreamBillingAttachment(attachment); err != nil {
		return nil, err
	}
	uploadErr := store.PutUpstreamAttachment(ctx, attachment, body)
	if err := model.CompleteUpstreamBillingAttachment(attachment.ID, actorID, uploadErr == nil); err != nil {
		if uploadErr == nil {
			return nil, err
		}
		common.SysError("complete failed upstream attachment " + attachment.ID + ": " + err.Error())
	}
	if uploadErr != nil {
		return nil, uploadErr
	}
	return model.GetUpstreamBillingAttachment(statementID, attachment.ID)
}

// Validate uploaded bytes rather than trusting a browser MIME type. The source
// stays byte-for-byte intact; safe display names never become object paths.
func ValidateUpstreamBillingAttachment(filename string, body []byte) (string, string, error) {
	if len(body) == 0 || int64(len(body)) > model.MaxUpstreamBillingAttachmentBytes || !utf8.ValidString(filename) {
		return "", "", model.ErrUpstreamBillingInvalidAttachment
	}
	for _, character := range filename {
		if unicode.IsControl(character) {
			return "", "", model.ErrUpstreamBillingInvalidAttachment
		}
	}
	filename = path.Base(strings.ReplaceAll(strings.TrimSpace(filename), "\\", "/"))
	if filename == "." || filename == ".." || filename == "/" || filename == "" || len(filename) > 255 {
		return "", "", model.ErrUpstreamBillingInvalidAttachment
	}
	switch strings.ToLower(path.Ext(filename)) {
	case ".pdf":
		trimmed := bytes.TrimSpace(body)
		if (!bytes.HasPrefix(body, []byte("%PDF-1.")) && !bytes.HasPrefix(body, []byte("%PDF-2.0"))) || !bytes.HasSuffix(trimmed, []byte("%%EOF")) {
			return "", "", model.ErrUpstreamBillingInvalidAttachment
		}
		index := bytes.LastIndex(trimmed, []byte("startxref"))
		if index < 0 {
			return "", "", model.ErrUpstreamBillingInvalidAttachment
		}
		offset, err := strconv.ParseInt(strings.TrimSpace(string(trimmed[index+9:len(trimmed)-5])), 10, 64)
		if err != nil || offset < 0 || offset >= int64(index) {
			return "", "", model.ErrUpstreamBillingInvalidAttachment
		}
		start := body[offset:min(int64(len(body)), offset+512)]
		if !bytes.HasPrefix(start, []byte("xref")) && !bytes.Contains(start, []byte("/Type /XRef")) && !bytes.Contains(start, []byte("/Type/XRef")) {
			return "", "", model.ErrUpstreamBillingInvalidAttachment
		}
		return filename, "application/pdf", nil
	case ".xlsx":
		if err := validateUpstreamBillingWorkbook(body); err != nil {
			return "", "", err
		}
		return filename, billingExcelContentType, nil
	case ".csv":
		if !utf8.Valid(body) || bytes.IndexByte(body, 0) >= 0 {
			return "", "", model.ErrUpstreamBillingInvalidAttachment
		}
		reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(body, []byte{0xef, 0xbb, 0xbf})))
		reader.FieldsPerRecord = -1
		reader.ReuseRecord = true
		for {
			_, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", "", model.ErrUpstreamBillingInvalidAttachment
			}
		}
		return filename, "text/csv; charset=utf-8", nil
	case ".png", ".jpg", ".jpeg":
		config, format, err := image.DecodeConfig(bytes.NewReader(body))
		if err != nil || config.Width <= 0 || config.Height <= 0 || config.Width > 20000 || config.Height > 20000 || int64(config.Width)*int64(config.Height) > 40_000_000 {
			return "", "", model.ErrUpstreamBillingInvalidAttachment
		}
		if strings.EqualFold(path.Ext(filename), ".png") {
			if format != "png" || !bytes.HasSuffix(body, []byte{0, 0, 0, 0, 'I', 'E', 'N', 'D', 0xae, 0x42, 0x60, 0x82}) {
				return "", "", model.ErrUpstreamBillingInvalidAttachment
			}
			return filename, "image/png", nil
		}
		if format != "jpeg" || !bytes.HasSuffix(bytes.TrimSpace(body), []byte{0xff, 0xd9}) {
			return "", "", model.ErrUpstreamBillingInvalidAttachment
		}
		return filename, "image/jpeg", nil
	default:
		return "", "", model.ErrUpstreamBillingInvalidAttachment
	}
}

func validateUpstreamBillingWorkbook(body []byte) error {
	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil || len(archive.File) > 4096 {
		return model.ErrUpstreamBillingInvalidAttachment
	}
	var expanded uint64
	required := map[string]bool{"[Content_Types].xml": false, "_rels/.rels": false, "xl/workbook.xml": false, "xl/_rels/workbook.xml.rels": false}
	worksheet := false
	for _, file := range archive.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if strings.Contains(file.Name, "\\") || strings.HasPrefix(file.Name, "/") || path.Clean(file.Name) != file.Name || strings.HasPrefix(file.Name, "../") {
			return model.ErrUpstreamBillingInvalidAttachment
		}
		lower := strings.ToLower(file.Name)
		if strings.Contains(lower, "vbaproject") || strings.HasPrefix(lower, "xl/macrosheets/") {
			return model.ErrUpstreamBillingInvalidAttachment
		}
		if file.UncompressedSize64 > 100*1024*1024-expanded {
			return model.ErrUpstreamBillingInvalidAttachment
		}
		expanded += file.UncompressedSize64
		reader, err := file.Open()
		if err != nil {
			return model.ErrUpstreamBillingInvalidAttachment
		}
		if file.Name == "[Content_Types].xml" || file.Name == "xl/workbook.xml" {
			data, readErr := io.ReadAll(io.LimitReader(reader, 2*1024*1024+1))
			reader.Close()
			if readErr != nil || len(data) > 2*1024*1024 {
				return model.ErrUpstreamBillingInvalidAttachment
			}
			var root struct{ XMLName xml.Name }
			if xml.Unmarshal(data, &root) != nil {
				return model.ErrUpstreamBillingInvalidAttachment
			}
			if file.Name == "[Content_Types].xml" && (root.XMLName.Local != "Types" || !bytes.Contains(data, []byte("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml")) || bytes.Contains(bytes.ToLower(data), []byte("macroenabled"))) {
				return model.ErrUpstreamBillingInvalidAttachment
			}
			if file.Name == "xl/workbook.xml" && root.XMLName.Local != "workbook" {
				return model.ErrUpstreamBillingInvalidAttachment
			}
		} else {
			copied, readErr := io.Copy(io.Discard, io.LimitReader(reader, int64(file.UncompressedSize64)+1))
			reader.Close()
			if readErr != nil || uint64(copied) != file.UncompressedSize64 {
				return model.ErrUpstreamBillingInvalidAttachment
			}
		}
		if _, ok := required[file.Name]; ok {
			if required[file.Name] {
				return model.ErrUpstreamBillingInvalidAttachment
			}
			required[file.Name] = true
		}
		if strings.HasPrefix(file.Name, "xl/worksheets/") && strings.HasSuffix(file.Name, ".xml") {
			worksheet = true
		}
	}
	for name, found := range required {
		if !found {
			return fmt.Errorf("%w: missing workbook member %s", model.ErrUpstreamBillingInvalidAttachment, name)
		}
	}
	if !worksheet {
		return model.ErrUpstreamBillingInvalidAttachment
	}
	return nil
}
