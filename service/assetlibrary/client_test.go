package assetlibrary

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestProviderClientBearerPreservesCustomPathAndOfficialEnvelope(t *testing.T) {
	config := &model.AssetChannelConfig{
		ChannelID: 12, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://assets.example.com/volcengine/assets/",
		Region: DefaultRegion, Service: DefaultService, APIVersion: DefaultAPIVersion, ProjectName: "project-a", QPM: 1000,
	}
	client := &volcActionClient{config: config, secret: "test-api-key", now: time.Now}
	client.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "/volcengine/assets/", request.URL.Path)
		assert.Equal(t, "CreateAsset", request.URL.Query().Get("Action"))
		assert.Equal(t, DefaultAPIVersion, request.URL.Query().Get("Version"))
		assert.Equal(t, "Bearer test-api-key", request.Header.Get("Authorization"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"GroupId":"group-upstream"`)
		assert.Contains(t, string(body), `"ProjectName":"project-a"`)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"ResponseMetadata":{"RequestId":"req-1"},"Result":{"Id":"asset-created"}}`)),
		}, nil
	})}

	id, err := client.createAsset(context.Background(), "group-upstream", "portrait", model.AssetTypeImage, "https://storage.example.com/source.png")
	require.NoError(t, err)
	assert.Equal(t, "asset-created", id)
}

func TestProviderClientAcceptsCompatibleDataEnvelope(t *testing.T) {
	config := &model.AssetChannelConfig{ChannelID: 13, AuthType: model.AssetChannelAuthBearer, BaseURL: "https://assets.example.com", Region: DefaultRegion, Service: DefaultService, APIVersion: DefaultAPIVersion, QPM: 1000}
	client := &volcActionClient{config: config, secret: "key", now: time.Now}
	client.httpClient = &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"code":0,"data":{"Id":"group-compatible"}}`))}, nil
	})}

	id, err := client.createGroup(context.Background(), "group", "description")
	require.NoError(t, err)
	assert.Equal(t, "group-compatible", id)
}

func TestProviderClientAKSKSignatureIncludesPayloadAndCustomPath(t *testing.T) {
	fixed := time.Date(2026, time.September, 22, 8, 30, 45, 0, time.UTC)
	config := &model.AssetChannelConfig{
		ChannelID: 14, AuthType: model.AssetChannelAuthAKSK, AccessKeyID: "AKIDEXAMPLE", BaseURL: DefaultBaseURL,
		Region: DefaultRegion, Service: DefaultService, APIVersion: DefaultAPIVersion, QPM: 60,
	}
	client := &volcActionClient{config: config, secret: "secret-example", now: func() time.Time { return fixed }}
	body := []byte(`{"Id":"asset-1"}`)
	request, err := http.NewRequest(http.MethodPost, "https://ark.cn-beijing.volcengineapi.com/custom/path?Version=2024-01-01&Action=GetAsset", bytes.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	client.sign(request, body)

	digest := sha256.Sum256(body)
	assert.Equal(t, hex.EncodeToString(digest[:]), request.Header.Get("X-Content-Sha256"))
	assert.Equal(t, "20260922T083045Z", request.Header.Get("X-Date"))
	authorization := request.Header.Get("Authorization")
	assert.Contains(t, authorization, "Credential=AKIDEXAMPLE/20260922/cn-beijing/ark/request")
	assert.Contains(t, authorization, "SignedHeaders=content-type;host;x-content-sha256;x-date")
	assert.Contains(t, authorization, "Signature=")
	assert.Equal(t, "Action=GetAsset&Version=2024-01-01", canonicalQuery(request.URL.Query()))
	assert.Equal(t, "/custom/path", canonicalURI(request.URL))
}

func TestVolcActionClientSupportsDocumentedAKSKProviderPaths(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		wantPath string
	}{
		{
			name:     "XingShu Wuji asset endpoint",
			baseURL:  "https://mintel.591ll.com/render/api/assets/",
			wantPath: "/render/api/assets/",
		},
		{
			name:     "YuMi asset endpoint",
			baseURL:  "https://video.flaios.cn/api/v3/",
			wantPath: "/api/v3/",
		},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &model.AssetChannelConfig{
				ChannelID: index + 100, Protocol: model.AssetChannelProtocolVolcAction,
				AuthType: model.AssetChannelAuthAKSK, AccessKeyID: "provider-ak", BaseURL: test.baseURL,
				Region: DefaultRegion, Service: DefaultService, APIVersion: DefaultAPIVersion, QPM: 1000,
			}
			client := &volcActionClient{config: config, secret: "provider-sk", now: time.Now}
			client.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				assert.Equal(t, test.wantPath, request.URL.Path)
				assert.Equal(t, "ListAssetGroups", request.URL.Query().Get("Action"))
				assert.Equal(t, DefaultAPIVersion, request.URL.Query().Get("Version"))
				assert.Contains(t, request.Header.Get("Authorization"), "Credential=provider-ak/")
				assert.Contains(t, request.Header.Get("Authorization"), "/cn-beijing/ark/request")
				body, err := io.ReadAll(request.Body)
				require.NoError(t, err)
				var payload struct {
					Filter struct {
						GroupType string `json:"GroupType"`
					} `json:"Filter"`
					PageNumber int `json:"PageNumber"`
					PageSize   int `json:"PageSize"`
				}
				require.NoError(t, common.Unmarshal(body, &payload))
				assert.Equal(t, "AIGC", payload.Filter.GroupType)
				assert.Equal(t, 1, payload.PageNumber)
				assert.Equal(t, 1, payload.PageSize)
				digest := sha256.Sum256(body)
				assert.Equal(t, hex.EncodeToString(digest[:]), request.Header.Get("X-Content-Sha256"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"ResponseMetadata":{"RequestId":"req-compatible"},"Result":{"Items":[],"TotalCount":0}}`)),
				}, nil
			})}

			require.NoError(t, client.test(context.Background()))
		})
	}
}

func TestValidateChannelConfigAllowsHTTPSPathsAndOnlyLoopbackHTTP(t *testing.T) {
	base := model.AssetChannelConfig{Protocol: model.AssetChannelProtocolVolcAction, AuthType: model.AssetChannelAuthBearer, EncryptedCredential: "encrypted", Region: DefaultRegion, Service: DefaultService, APIVersion: DefaultAPIVersion, QPM: 60}
	base.BaseURL = "https://provider.example.com/custom/assets/path"
	require.NoError(t, validateChannelConfig(&base))

	base.BaseURL = "http://127.0.0.1:8080/assets"
	require.NoError(t, validateChannelConfig(&base))

	base.BaseURL = "http://provider.example.com/assets"
	assert.ErrorContains(t, validateChannelConfig(&base), "HTTPS")

	base.BaseURL = "https://provider.example.com/assets?fixed=query"
	assert.ErrorContains(t, validateChannelConfig(&base), "invalid")

	base.Protocol = model.AssetChannelProtocolYoufangREST
	base.AuthType = model.AssetChannelAuthBearer
	base.BaseURL = DefaultYoufangBaseURL
	base.Region = ""
	base.Service = ""
	base.APIVersion = ""
	require.NoError(t, validateChannelConfig(&base))

	base.AuthType = model.AssetChannelAuthAKSK
	assert.ErrorContains(t, validateChannelConfig(&base), "requires bearer")
}

func TestProtocolDefaultsKeepLegacyConfigsOnVolcAndInitializeYoufang(t *testing.T) {
	legacy := model.AssetChannelConfig{AuthType: model.AssetChannelAuthBearer, BaseURL: "https://provider.example.com", Region: DefaultRegion, Service: DefaultService, APIVersion: DefaultAPIVersion}
	applyProtocolDefaults(&legacy)
	assert.Equal(t, model.AssetChannelProtocolVolcAction, legacy.Protocol)
	assert.Equal(t, "https://provider.example.com", legacy.BaseURL)

	youfang := model.AssetChannelConfig{Protocol: model.AssetChannelProtocolYoufangREST}
	applyProtocolDefaults(&youfang)
	assert.Equal(t, model.AssetChannelAuthBearer, youfang.AuthType)
	assert.Equal(t, DefaultYoufangBaseURL, youfang.BaseURL)
}

func TestYoufangRESTClientCreatesAssetWithSKBearerAndDefaultGroup(t *testing.T) {
	config := &model.AssetChannelConfig{
		ChannelID: 31, Protocol: model.AssetChannelProtocolYoufangREST, AuthType: model.AssetChannelAuthBearer,
		BaseURL: "https://asset-inference-doubao.yoofang.com/gateway", QPM: 1000,
	}
	client := &youfangRESTClient{config: config, secret: "sk-test-key"}
	client.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/gateway/api/v1/assets", request.URL.Path)
		assert.Empty(t, request.URL.RawQuery)
		assert.Equal(t, "Bearer sk-test-key", request.Header.Get("Authorization"))
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Equal(t, "https://storage.example.com/source.png", gjson.GetBytes(body, "url").String())
		assert.Equal(t, "portrait", gjson.GetBytes(body, "name").String())
		assert.Equal(t, "Image", gjson.GetBytes(body, "asset_type").String())
		assert.False(t, gjson.GetBytes(body, "group_id").Exists())
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"id":"asset-youfang-1","ref":"asset://asset-youfang-1","status":"Processing"}`)),
		}, nil
	})}

	assert.False(t, client.requiresGroup())
	id, err := client.createAsset(context.Background(), "ignored-local-group", "portrait", model.AssetTypeImage, "https://storage.example.com/source.png")
	require.NoError(t, err)
	assert.Equal(t, "asset-youfang-1", id)
}

func TestYoufangRESTClientUsesResourcePathsForStatusDeleteAndTest(t *testing.T) {
	tests := []struct {
		name       string
		channelID  int
		invoke     func(*youfangRESTClient) (providerResult, error)
		method     string
		path       string
		query      string
		response   string
		wantStatus string
	}{
		{
			name: "get asset", channelID: 32, method: http.MethodGet, path: "/api/v1/assets/asset-yf-2",
			response: `{"id":"asset-yf-2","status":"Active"}`, wantStatus: "Active",
			invoke: func(client *youfangRESTClient) (providerResult, error) {
				return client.getAsset(context.Background(), "asset-yf-2")
			},
		},
		{
			name: "delete asset", channelID: 33, method: http.MethodDelete, path: "/api/v1/assets/asset-yf-2",
			response: `{}`,
			invoke: func(client *youfangRESTClient) (providerResult, error) {
				return providerResult{}, client.deleteAsset(context.Background(), "asset-yf-2")
			},
		},
		{
			name: "test connection", channelID: 34, method: http.MethodGet, path: "/api/v1/assets", query: "page=1&size=1",
			response: `{"items":[],"page":1,"size":1}`,
			invoke: func(client *youfangRESTClient) (providerResult, error) {
				return providerResult{}, client.test(context.Background())
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &model.AssetChannelConfig{ChannelID: test.channelID, Protocol: model.AssetChannelProtocolYoufangREST, AuthType: model.AssetChannelAuthBearer, BaseURL: DefaultYoufangBaseURL, QPM: 1000}
			client := &youfangRESTClient{config: config, secret: "sk-test-key"}
			client.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				assert.Equal(t, test.method, request.Method)
				assert.Equal(t, test.path, request.URL.Path)
				assert.Equal(t, test.query, request.URL.RawQuery)
				assert.Equal(t, "Bearer sk-test-key", request.Header.Get("Authorization"))
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(test.response))}, nil
			})}

			result, err := test.invoke(client)
			require.NoError(t, err)
			assert.Equal(t, test.wantStatus, result.Status)
		})
	}
}
