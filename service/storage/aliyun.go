package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/runtimewatch"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

var (
	aliyunRegionPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$`)
	ossBucketPattern    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)
)

type aliyunOSSDriver struct {
	client          *oss.Client
	bucket          string
	idleConnections interface{ CloseIdleConnections() }
}

func newAliyunOSSDriver(profile *model.StorageProfile, credential *model.StorageCredential) (*aliyunOSSDriver, error) {
	if err := validateAliyunProfile(profile); err != nil {
		return nil, err
	}
	var provider credentials.CredentialsProvider
	switch credential.AuthType {
	case model.StorageCredentialAuthStatic:
		secret, err := decryptCredential(credential.EncryptedPayload)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(secret.AccessKeyID) == "" || strings.TrimSpace(secret.AccessKeySecret) == "" {
			return nil, fmt.Errorf("Aliyun OSS access key is incomplete")
		}
		provider = credentials.NewStaticCredentialsProvider(secret.AccessKeyID, secret.AccessKeySecret, secret.SecurityToken)
	case model.StorageCredentialAuthEnvironment:
		provider = credentials.NewEnvironmentVariableCredentialsProvider()
	default:
		return nil, fmt.Errorf("unsupported storage credential type: %s", credential.AuthType)
	}

	config := oss.LoadDefaultConfig().
		WithCredentialsProvider(provider).
		WithRegion(profile.Region).
		WithEndpoint(profile.Endpoint).
		WithConnectTimeout(10 * time.Second).
		WithReadWriteTimeout(60 * time.Second).
		WithRetryMaxAttempts(3)
	driver := &aliyunOSSDriver{bucket: profile.Bucket}
	driver.client = oss.NewClient(config, func(options *oss.Options) {
		if client, ok := options.HttpClient.(*http.Client); ok && runtimewatch.Enabled() {
			observed := *client
			transport := client.Transport
			if original, ok := runtimewatch.UnwrapTransport(transport).(*http.Transport); ok {
				clone := original.Clone()
				runtimewatch.ObserveHTTPConnections(clone)
				transport = clone
			}
			observed.Transport = runtimewatch.WrapTransport(transport, "storage")
			options.HttpClient = &observed
		}
		driver.idleConnections, _ = options.HttpClient.(interface{ CloseIdleConnections() })
	})
	return driver, nil
}

func (driver *aliyunOSSDriver) CloseIdleConnections() {
	if driver.idleConnections != nil {
		driver.idleConnections.CloseIdleConnections()
	}
}

func validateAliyunProfile(profile *model.StorageProfile) error {
	if profile == nil {
		return fmt.Errorf("storage profile is required")
	}
	if !aliyunRegionPattern.MatchString(strings.TrimSpace(profile.Region)) {
		return fmt.Errorf("invalid Aliyun OSS region")
	}
	if !ossBucketPattern.MatchString(strings.TrimSpace(profile.Bucket)) {
		return fmt.Errorf("invalid Aliyun OSS bucket")
	}
	endpoint, err := url.Parse(strings.TrimSpace(profile.Endpoint))
	if err != nil || endpoint.Scheme != "https" || endpoint.Hostname() == "" || endpoint.Port() != "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return fmt.Errorf("Aliyun OSS endpoint must be an HTTPS URL")
	}
	if endpoint.Path != "" && endpoint.Path != "/" {
		return fmt.Errorf("Aliyun OSS endpoint must not contain a path")
	}
	host := strings.ToLower(strings.TrimSuffix(endpoint.Hostname(), "."))
	if host != "aliyuncs.com" && !strings.HasSuffix(host, ".aliyuncs.com") {
		return fmt.Errorf("Aliyun OSS endpoint must use an official aliyuncs.com domain")
	}
	return nil
}

func (driver *aliyunOSSDriver) PutObject(ctx context.Context, objectKey string, contentType string, size int64, body io.Reader) (string, error) {
	cacheControl := "private, no-store"
	if strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "video/") || strings.HasPrefix(contentType, "audio/") {
		// Media objects use unique random keys with ForbidOverwrite=true and are
		// immutable once uploaded; private max-age enables browser disk caching
		// even when Aliyun OSS x-oss-process inherits object Cache-Control.
		cacheControl = "private, max-age=1800, immutable"
	}
	forbidOverwrite := "true"
	request := &oss.PutObjectRequest{
		Bucket:          oss.Ptr(driver.bucket),
		Key:             oss.Ptr(objectKey),
		Body:            body,
		ContentLength:   oss.Ptr(size),
		ContentType:     oss.Ptr(contentType),
		CacheControl:    oss.Ptr(cacheControl),
		ForbidOverwrite: oss.Ptr(forbidOverwrite),
	}
	// Snapshot JSON contains administrator-only request data. Never inherit a
	// public bucket ACL for these objects; media uploads retain their policy.
	if contentType == "application/json" {
		request.Acl = oss.ObjectACLPrivate
	}
	result, err := driver.client.PutObject(ctx, request)
	if err != nil {
		return "", err
	}
	if result.ETag == nil {
		return "", nil
	}
	return strings.Trim(*result.ETag, `"`), nil
}

func (driver *aliyunOSSDriver) PresignGet(ctx context.Context, objectKey string, ttl time.Duration, options ...ObjectGetOptions) (string, error) {
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(driver.bucket),
		Key:    oss.Ptr(objectKey),
	}
	if len(options) > 0 {
		if options[0].Process != "" {
			request.Process = oss.Ptr(options[0].Process)
		}
		if options[0].ResponseContentDisposition != "" {
			request.ResponseContentDisposition = oss.Ptr(options[0].ResponseContentDisposition)
		}
		if options[0].ResponseCacheControl != "" {
			request.ResponseCacheControl = oss.Ptr(options[0].ResponseCacheControl)
		}
	}
	result, err := driver.client.Presign(ctx, request, oss.PresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

func (driver *aliyunOSSDriver) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	result, err := driver.client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(driver.bucket),
		Key:    oss.Ptr(objectKey),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (driver *aliyunOSSDriver) DeleteObject(ctx context.Context, objectKey string) error {
	_, err := driver.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(driver.bucket),
		Key:    oss.Ptr(objectKey),
	})
	return err
}

func testObjectDriver(ctx context.Context, driver ObjectDriver) error {
	suffix, err := randomObjectSuffix()
	if err != nil {
		return err
	}
	objectKey := fmt.Sprintf("temporary/relay-media/healthcheck/%d-%s.txt", time.Now().Unix(), suffix)
	payload := []byte("storage healthcheck")
	if _, err = driver.PutObject(ctx, objectKey, "text/plain", int64(len(payload)), bytes.NewReader(payload)); err != nil {
		return fmt.Errorf("upload test object: %w", err)
	}
	needsCleanup := true
	defer func() {
		if !needsCleanup {
			return
		}
		deleteCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = driver.DeleteObject(deleteCtx, objectKey)
	}()

	signedURL, err := driver.PresignGet(ctx, objectKey, 15*time.Minute)
	if err != nil {
		return fmt.Errorf("sign test object: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, signedURL, nil)
	if err != nil {
		return fmt.Errorf("build signed URL request: %w", err)
	}
	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(request)
	if err != nil {
		return fmt.Errorf("download test object: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download test object returned HTTP %d", response.StatusCode)
	}
	downloaded, err := io.ReadAll(io.LimitReader(response.Body, int64(len(payload)+1)))
	if err != nil {
		return fmt.Errorf("read test object: %w", err)
	}
	if !bytes.Equal(downloaded, payload) {
		return fmt.Errorf("test object content mismatch")
	}
	if err = driver.DeleteObject(ctx, objectKey); err != nil {
		return fmt.Errorf("delete test object: %w", err)
	}
	needsCleanup = false
	return nil
}
