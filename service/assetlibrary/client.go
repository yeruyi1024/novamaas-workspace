package assetlibrary

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const (
	DefaultBaseURL        = "https://ark.cn-beijing.volcengineapi.com"
	DefaultYoufangBaseURL = "https://asset-inference-doubao.yoofang.com"
	DefaultRegion         = "cn-beijing"
	DefaultService        = "ark"
	DefaultAPIVersion     = "2024-01-01"
)

type assetProvider interface {
	requiresGroup() bool
	createGroup(ctx context.Context, name string, description string) (string, error)
	createAsset(ctx context.Context, groupID string, name string, assetType string, sourceURL string) (string, error)
	getAsset(ctx context.Context, id string) (providerResult, error)
	deleteAsset(ctx context.Context, id string) error
	test(ctx context.Context) error
}

type volcActionClient struct {
	config     *model.AssetChannelConfig
	secret     string
	httpClient *http.Client
	now        func() time.Time
}

var channelRateState = struct {
	sync.Mutex
	next map[int]time.Time
}{next: make(map[int]time.Time)}

type providerResult struct {
	ID      string
	Status  string
	Code    string
	Message string
}

func newProviderClient(config *model.AssetChannelConfig) (assetProvider, error) {
	if config == nil {
		return nil, errors.New("asset channel configuration is required")
	}
	secret, err := decryptChannelCredential(config.EncryptedCredential)
	if err != nil {
		return nil, err
	}
	base := volcActionClient{
		config: config,
		secret: secret,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return errors.New("asset provider redirects are not allowed")
			},
		},
		now: time.Now,
	}
	switch normalizeProtocol(config.Protocol) {
	case model.AssetChannelProtocolVolcAction:
		return &base, nil
	case model.AssetChannelProtocolYoufangREST:
		return &youfangRESTClient{config: config, secret: secret, httpClient: base.httpClient}, nil
	default:
		return nil, errors.New("unsupported asset provider protocol")
	}
}

func (client *volcActionClient) requiresGroup() bool { return true }

func (client *volcActionClient) createGroup(ctx context.Context, name string, description string) (string, error) {
	payload := map[string]any{
		"Name":        name,
		"Description": description,
		"GroupType":   "AIGC",
	}
	if client.config.ProjectName != "" {
		payload["ProjectName"] = client.config.ProjectName
	}
	result, err := client.call(ctx, "CreateAssetGroup", payload)
	if err != nil {
		return "", err
	}
	if result.ID == "" {
		return "", errors.New("asset provider returned an empty group ID")
	}
	return result.ID, nil
}

func (client *volcActionClient) createAsset(ctx context.Context, groupID string, name string, assetType string, sourceURL string) (string, error) {
	payload := map[string]any{
		"GroupId":   groupID,
		"Name":      name,
		"AssetType": strings.ToUpper(assetType[:1]) + strings.ToLower(assetType[1:]),
		"URL":       sourceURL,
	}
	if client.config.ProjectName != "" {
		payload["ProjectName"] = client.config.ProjectName
	}
	result, err := client.call(ctx, "CreateAsset", payload)
	if err != nil {
		return "", err
	}
	if result.ID == "" {
		return "", errors.New("asset provider returned an empty asset ID")
	}
	return result.ID, nil
}

func (client *volcActionClient) getAsset(ctx context.Context, id string) (providerResult, error) {
	payload := map[string]any{"Id": id}
	if client.config.ProjectName != "" {
		payload["ProjectName"] = client.config.ProjectName
	}
	return client.call(ctx, "GetAsset", payload)
}

func (client *volcActionClient) deleteAsset(ctx context.Context, id string) error {
	payload := map[string]any{"Id": id}
	if client.config.ProjectName != "" {
		payload["ProjectName"] = client.config.ProjectName
	}
	_, err := client.call(ctx, "DeleteAsset", payload)
	return err
}

func (client *volcActionClient) test(ctx context.Context) error {
	payload := map[string]any{
		"Filter":     map[string]any{"GroupType": "AIGC"},
		"PageNumber": 1,
		"PageSize":   1,
	}
	if client.config.ProjectName != "" {
		payload["ProjectName"] = client.config.ProjectName
	}
	_, err := client.call(ctx, "ListAssetGroups", payload)
	return err
}

func (client *volcActionClient) call(ctx context.Context, action string, payload any) (result providerResult, err error) {
	started := time.Now()
	status, requestBytes, responseBytes := 0, int64(0), int64(0)
	errorKind := ""
	var requestURL string
	var requestBody, loggedResponseBody []byte
	defer func() {
		if err != nil && errorKind == "" {
			errorKind = "unknown"
		}
		recordRequestAttempt(ctx, client.config.ChannelID, normalizeProtocol(client.config.Protocol), action, http.MethodPost, "/", status, time.Since(started), requestBytes, responseBytes, errorKind, requestURL, requestBody, loggedResponseBody)
	}()
	if err := waitForChannelRate(ctx, client.config.ChannelID, client.config.QPM); err != nil {
		errorKind = "rate_limit"
		return providerResult{}, err
	}
	body, err := common.Marshal(payload)
	if err != nil {
		errorKind = "prepare"
		return providerResult{}, err
	}
	requestBytes = int64(len(body))
	requestBody = body
	endpoint, err := url.Parse(client.config.BaseURL)
	if err != nil {
		errorKind = "prepare"
		return providerResult{}, err
	}
	query := endpoint.Query()
	query.Set("Action", action)
	query.Set("Version", client.config.APIVersion)
	endpoint.RawQuery = query.Encode()
	requestURL = endpoint.String()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		errorKind = "prepare"
		return providerResult{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	switch client.config.AuthType {
	case model.AssetChannelAuthBearer:
		request.Header.Set("Authorization", "Bearer "+client.secret)
	case model.AssetChannelAuthAKSK:
		client.sign(request, body)
	default:
		errorKind = "prepare"
		return providerResult{}, errors.New("unsupported asset provider authentication type")
	}
	response, err := client.httpClient.Do(request)
	if err != nil {
		errorKind = "network"
		return providerResult{}, fmt.Errorf("call asset provider %s: %w", action, err)
	}
	defer response.Body.Close()
	status = response.StatusCode
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 4*1024*1024))
	if err != nil {
		errorKind = "response_read"
		return providerResult{}, fmt.Errorf("read asset provider response: %w", err)
	}
	responseBytes = int64(len(responseBody))
	loggedResponseBody = responseBody
	var envelope map[string]any
	if len(responseBody) > 0 {
		if err = common.Unmarshal(responseBody, &envelope); err != nil {
			errorKind = "response_decode"
			return providerResult{}, fmt.Errorf("decode asset provider response: %w", err)
		}
	}
	if providerErr := providerResponseError(response.StatusCode, envelope); providerErr != nil {
		errorKind = "upstream"
		return providerResult{}, fmt.Errorf("asset provider %s failed: %w", action, providerErr)
	}
	resultMap := mapValue(envelope, "Result")
	if len(resultMap) == 0 {
		resultMap = mapValue(envelope, "Data")
	}
	return providerResult{
		ID:      stringValue(resultMap, "Id", "AssetId", "GroupId"),
		Status:  stringValue(resultMap, "Status"),
		Code:    nestedErrorCode(resultMap),
		Message: nestedErrorMessage(resultMap),
	}, nil
}

func waitForChannelRate(ctx context.Context, channelID int, qpm int) error {
	if qpm <= 0 {
		qpm = 60
	}
	interval := time.Minute / time.Duration(qpm)
	now := time.Now()
	channelRateState.Lock()
	allowedAt := channelRateState.next[channelID]
	if allowedAt.Before(now) {
		allowedAt = now
	}
	channelRateState.next[channelID] = allowedAt.Add(interval)
	channelRateState.Unlock()
	if delay := time.Until(allowedAt); delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	return nil
}

func (client *volcActionClient) sign(request *http.Request, body []byte) {
	now := client.now().UTC()
	xDate := now.Format("20060102T150405Z")
	shortDate := now.Format("20060102")
	payloadHashBytes := sha256.Sum256(body)
	payloadHash := hex.EncodeToString(payloadHashBytes[:])
	request.Header.Set("X-Date", xDate)
	request.Header.Set("X-Content-Sha256", payloadHash)

	canonicalHeaders := "content-type:" + strings.TrimSpace(request.Header.Get("Content-Type")) + "\n" +
		"host:" + request.URL.Host + "\n" +
		"x-content-sha256:" + payloadHash + "\n" +
		"x-date:" + xDate + "\n"
	signedHeaders := "content-type;host;x-content-sha256;x-date"
	canonicalRequest := strings.Join([]string{
		request.Method,
		canonicalURI(request.URL),
		canonicalQuery(request.URL.Query()),
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
	canonicalHash := sha256.Sum256([]byte(canonicalRequest))
	scope := strings.Join([]string{shortDate, client.config.Region, client.config.Service, "request"}, "/")
	stringToSign := "HMAC-SHA256\n" + xDate + "\n" + scope + "\n" + hex.EncodeToString(canonicalHash[:])
	kDate := hmacSHA256([]byte(client.secret), []byte(shortDate))
	kRegion := hmacSHA256(kDate, []byte(client.config.Region))
	kService := hmacSHA256(kRegion, []byte(client.config.Service))
	kSigning := hmacSHA256(kService, []byte("request"))
	signature := hex.EncodeToString(hmacSHA256(kSigning, []byte(stringToSign)))
	request.Header.Set("Authorization", fmt.Sprintf(
		"HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		client.config.AccessKeyID, scope, signedHeaders, signature,
	))
}

func canonicalURI(value *url.URL) string {
	if value.EscapedPath() == "" {
		return "/"
	}
	return value.EscapedPath()
}

func canonicalQuery(values url.Values) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(values))
	for _, key := range keys {
		items := append([]string(nil), values[key]...)
		sort.Strings(items)
		for _, item := range items {
			parts = append(parts, uriEncode(key)+"="+uriEncode(item))
		}
	}
	return strings.Join(parts, "&")
}

func uriEncode(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}

func hmacSHA256(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	_, _ = hash.Write(data)
	return hash.Sum(nil)
}

func providerResponseError(statusCode int, envelope map[string]any) error {
	metadata := mapValue(envelope, "ResponseMetadata")
	errorMap := mapValue(metadata, "Error")
	if len(errorMap) == 0 {
		errorMap = mapValue(envelope, "Error")
	}
	code := stringValue(errorMap, "Code")
	message := stringValue(errorMap, "Message")
	if code != "" || message != "" {
		return &upstreamAssetError{statusCode: statusCode, code: code, message: message}
	}
	for candidate, value := range envelope {
		if !strings.EqualFold(candidate, "code") {
			continue
		}
		switch typed := value.(type) {
		case float64:
			if typed != 0 {
				return &upstreamAssetError{statusCode: statusCode, code: fmt.Sprintf("code %.0f", typed), message: stringValue(envelope, "message", "msg")}
			}
		case string:
			if typed != "" && typed != "0" && !strings.EqualFold(typed, "success") {
				return &upstreamAssetError{statusCode: statusCode, code: "code " + typed, message: stringValue(envelope, "message", "msg")}
			}
		}
	}
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return &upstreamAssetError{statusCode: statusCode, code: fmt.Sprintf("HTTP %d", statusCode), message: stringValue(envelope, "message", "msg", "detail")}
	}
	return nil
}

func mapValue(source map[string]any, key string) map[string]any {
	for candidate, value := range source {
		if strings.EqualFold(candidate, key) {
			if result, ok := value.(map[string]any); ok {
				return result
			}
		}
	}
	return map[string]any{}
}

func stringValue(source map[string]any, keys ...string) string {
	for _, key := range keys {
		for candidate, value := range source {
			if strings.EqualFold(candidate, key) {
				if text, ok := value.(string); ok {
					return text
				}
			}
		}
	}
	return ""
}

func nestedErrorMessage(result map[string]any) string {
	if message := stringValue(mapValue(result, "Error"), "Message", "Code"); message != "" {
		return message
	}
	return stringValue(result, "Message", "msg", "reason", "detail")
}

func nestedErrorCode(result map[string]any) string {
	if code := stringValue(mapValue(result, "Error"), "Code"); code != "" {
		return code
	}
	return stringValue(result, "Code")
}
