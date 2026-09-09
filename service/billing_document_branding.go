package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"golang.org/x/net/html"
)

type billingDocumentBranding struct {
	Issuer  string
	LogoPNG []byte
	Footer  string
}

var ErrBillingDocumentBranding = errors.New("billing document branding is invalid")

// Branding is fetched once, before taking an accounting lock. Freeze the actual
// pixels and visible footer, not a mutable (potentially signed) URL, so retries
// and a later confirmation receipt never need to fetch current system settings.
func captureBillingDocumentBranding(ctx context.Context) (*billingDocumentBranding, error) {
	common.OptionMapRWMutex.RLock()
	issuer, logoURL, footer := common.SystemName, strings.TrimSpace(common.Logo), common.Footer
	serverAddress := system_setting.ServerAddress
	common.OptionMapRWMutex.RUnlock()
	plainFooter, err := billingFooterText(footer)
	if err != nil {
		return nil, err
	}
	logo := billingPlatformMarkV2
	if logoURL != "" {
		parsed, err := url.Parse(logoURL)
		if err != nil {
			return nil, errors.New("invalid billing Logo URL")
		}
		if !parsed.IsAbs() && strings.HasPrefix(logoURL, "/") {
			base, err := url.Parse(serverAddress)
			if err != nil {
				return nil, errors.New("invalid ServerAddress for billing Logo URL")
			}
			parsed = base.ResolveReference(parsed)
		}
		if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil {
			return nil, errors.New("billing Logo URL must be an HTTP(S) image without credentials")
		}
		if err := ValidateSSRFProtectedFetchURL(parsed.String()); err != nil {
			return nil, errors.New("billing Logo URL is blocked by the configured fetch security policy")
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
		if err != nil {
			return nil, errors.New("invalid billing Logo request")
		}
		request.Header.Set("Accept", "image/png,image/jpeg,image/webp,image/gif")
		client := GetSSRFProtectedHTTPClient()
		if client == nil {
			return nil, errors.New("billing Logo HTTP client is not initialized")
		}
		response, err := client.Do(request)
		if err != nil {
			return nil, errors.New("billing Logo download failed; check the system Logo URL and network access")
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("billing Logo download returned HTTP %d", response.StatusCode)
		}
		const maxDownload = 2 * 1024 * 1024
		if response.ContentLength > maxDownload {
			return nil, errors.New("billing Logo exceeds 2 MiB")
		}
		logo, err = io.ReadAll(io.LimitReader(response.Body, maxDownload+1))
		if err != nil || len(logo) > maxDownload {
			return nil, errors.New("billing Logo could not be read within the 2 MiB limit")
		}
	}
	logo, err = normalizeBillingLogo(logo)
	if err != nil {
		return nil, err
	}
	return &billingDocumentBranding{Issuer: issuer, LogoPNG: logo, Footer: plainFooter}, nil
}

// Rasterize to a bounded PNG without distorting the aspect ratio or fetching
// external image resources. The limit also keeps the full JSON snapshot within
// MySQL's existing TEXT column, including base64 overhead and all 31 days.
func normalizeBillingLogo(body []byte) ([]byte, error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		return nil, errors.New("billing Logo must be a valid PNG, JPEG, GIF or WebP image")
	}
	if config.Width > 4096 || config.Height > 4096 || int64(config.Width)*int64(config.Height) > 16*1024*1024 {
		return nil, errors.New("billing Logo dimensions exceed 4096 pixels or 16 megapixels")
	}
	source, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("billing Logo image is damaged")
	}
	// At 38 PDF points even 128 pixels exceeds 240 dpi. Prefer 256 pixels,
	// reducing only unusually complex images to keep the archive compact.
	for _, maxSide := range []int{256, 128, 96} {
		width, height := config.Width, config.Height
		if longest := max(width, height); longest > maxSide {
			width = max(1, width*maxSide/longest)
			height = max(1, height*maxSide/longest)
		}
		normalized := image.NewNRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(normalized, normalized.Bounds(), source, source.Bounds(), draw.Src, nil)
		var encoded bytes.Buffer
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(&encoded, normalized); err != nil {
			return nil, err
		}
		if encoded.Len() <= 24*1024 {
			return encoded.Bytes(), nil
		}
	}
	return nil, errors.New("billing Logo cannot fit the document image limit")
}

// Site footers can contain HTML. Archive only visible text, never execute HTML
// or fetch a footer's scripts/images while preparing a financial document.
func billingFooterText(footer string) (string, error) {
	if len(footer) > 64*1024 {
		return "", errors.New("billing Footer HTML exceeds 64 KiB")
	}
	tokenizer := html.NewTokenizer(strings.NewReader(footer))
	var text strings.Builder
	hidden := ""
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			if err := tokenizer.Err(); err != io.EOF {
				return "", err
			}
			break
		}
		token := tokenizer.Token()
		if hidden != "" {
			if tokenType == html.EndTagToken && token.Data == hidden {
				hidden = ""
			}
			continue
		}
		if tokenType == html.StartTagToken {
			switch token.Data {
			case "script", "style", "iframe", "noscript", "template":
				hidden = token.Data
				continue
			}
		}
		if tokenType == html.TextToken {
			text.WriteString(token.Data)
		} else if tokenType == html.StartTagToken || tokenType == html.EndTagToken || tokenType == html.SelfClosingTagToken {
			switch token.Data {
			case "br", "p", "div", "li", "hr", "section", "footer":
				text.WriteByte(' ')
			}
		}
	}
	plain := strings.Join(strings.FieldsFunc(text.String(), func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }), " ")
	if len([]rune(plain)) > 300 {
		return "", errors.New("billing Footer exceeds 300 visible characters; shorten the system Footer")
	}
	return plain, nil
}

// Customer download names are presentation only. Internal UUIDs, object keys
// and evidence digests remain unchanged; downloading again keeps the same name.
func BillingArtifactFilename(statement *model.BillingStatement, kind string, ordinal int) string {
	if kind == "pdf" {
		return "月度对账单_" + time.Unix(statement.CreatedAt, 0).In(billingLocation).Format("20060102150405") + ".pdf"
	}
	if kind == "receipt" {
		return "对账确认回执_" + time.Unix(statement.ConfirmedAt, 0).In(billingLocation).Format("20060102150405") + ".pdf"
	}
	extension := "json"
	if kind == "details" {
		extension = "jsonl.gz"
	}
	return fmt.Sprintf("statement-%s-r%d-%s-%d.%s", statement.Month, statement.Revision, kind, ordinal, extension)
}
