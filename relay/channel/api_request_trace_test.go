package channel

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	rootconstant "github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type upstreamTraceRoundTripper func(*http.Request) (*http.Response, error)

func (transport upstreamTraceRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return transport(req)
}

type upstreamTraceChatAdaptor struct {
	Adaptor
	url       string
	urlErr    error
	headerErr error
}

func (adaptor *upstreamTraceChatAdaptor) GetRequestURL(*relaycommon.RelayInfo) (string, error) {
	return adaptor.url, adaptor.urlErr
}

func (adaptor *upstreamTraceChatAdaptor) SetupRequestHeader(*gin.Context, *http.Header, *relaycommon.RelayInfo) error {
	return adaptor.headerErr
}

type upstreamTraceTaskAdaptor struct {
	TaskAdaptor
	url       string
	urlErr    error
	headerErr error
}

func (adaptor *upstreamTraceTaskAdaptor) BuildRequestURL(*relaycommon.RelayInfo) (string, error) {
	return adaptor.url, adaptor.urlErr
}

func (adaptor *upstreamTraceTaskAdaptor) BuildRequestHeader(*gin.Context, *http.Request, *relaycommon.RelayInfo) error {
	return adaptor.headerErr
}

func invokeUpstreamTraceRequest(kind string, ctx *gin.Context, info *relaycommon.RelayInfo, requestURL string, urlErr, headerErr error) (*http.Response, error) {
	body := strings.NewReader(`{"input":"test"}`)
	switch kind {
	case "task":
		return DoTaskApiRequest(&upstreamTraceTaskAdaptor{url: requestURL, urlErr: urlErr, headerErr: headerErr}, ctx, info, body)
	case "chat":
		return DoApiRequest(&upstreamTraceChatAdaptor{url: requestURL, urlErr: urlErr, headerErr: headerErr}, ctx, info, body)
	case "form":
		return DoFormRequest(&upstreamTraceChatAdaptor{url: requestURL, urlErr: urlErr, headerErr: headerErr}, ctx, info, body)
	default:
		req, err := http.NewRequest(http.MethodPost, requestURL, body)
		if err != nil {
			return nil, err
		}
		return DoRequest(ctx, req, info)
	}
}

func TestUpstreamTraceBelongsOnlyToCurrentHTTPAttempt(t *testing.T) {
	service.InitHttpClient()
	gin.SetMode(gin.TestMode)
	client := service.GetHttpClient()
	require.NotNil(t, client)
	originalTransport := client.Transport
	t.Cleanup(func() { client.Transport = originalTransport })
	transportErr := errors.New("upstream connection failed before response")

	for _, kind := range []string{"chat", "task", "form", "direct"} {
		t.Run(kind, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/relay", strings.NewReader("client body"))
			ctx.Set(common.RequestIdKey, "local-request")
			ctx.Header(common.RequestIdKey, "local-request")
			info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{}}
			for _, attempt := range []struct {
				name, upstreamID string
				err              error
			}{
				{"initial response", "upstream-first", nil},
				{"response without request id", "", nil},
				{"new response id", "upstream-new", nil},
				{"transport failure", "", transportErr},
				{"response without id after transport failure", "", nil},
				{"recovered response id", "upstream-recovered", nil},
			} {
				client.Transport = upstreamTraceRoundTripper(func(req *http.Request) (*http.Response, error) {
					assert.Empty(t, ctx.GetString(common.UpstreamRequestIdKey), "an HTTP retry must clear its predecessor before transport starts")
					diagnostic, _ := common.GetContextKey(ctx, rootconstant.ContextKeyUpstreamTransportError)
					assert.Nil(t, diagnostic, "a retry must clear the previous transport diagnostic")
					if attempt.err != nil {
						return nil, attempt.err
					}
					header := make(http.Header)
					if attempt.upstreamID != "" {
						header.Set(common.RequestIdKey, attempt.upstreamID)
					}
					return &http.Response{StatusCode: http.StatusBadGateway, Header: header, Body: io.NopCloser(strings.NewReader("upstream response")), Request: req}, nil
				})
				resp, err := invokeUpstreamTraceRequest(kind, ctx, info, "https://upstream.test/relay", nil, nil)
				if attempt.err != nil {
					require.Error(t, err, attempt.name)
					var apiError *types.NewAPIError
					require.ErrorAs(t, err, &apiError)
					assert.Equal(t, types.ErrorCodeDoRequestFailed, apiError.GetErrorCode())
					assert.Equal(t, http.StatusInternalServerError, apiError.StatusCode)
					assert.Equal(t, "upstream error: do request failed", apiError.Error())
					assert.Nil(t, resp)
				} else {
					require.NoError(t, err, attempt.name)
					require.NotNil(t, resp)
					require.NoError(t, resp.Body.Close())
					assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
				}
				diagnostic, _ := common.GetContextKey(ctx, rootconstant.ContextKeyUpstreamTransportError)
				if attempt.err != nil {
					assert.NotNil(t, diagnostic, "the current transport failure must remain diagnosable after its public message is hidden")
				} else {
					assert.Nil(t, diagnostic)
				}
				assert.Equal(t, attempt.upstreamID, ctx.GetString(common.UpstreamRequestIdKey), attempt.name)
				assert.Equal(t, "local-request", ctx.GetString(common.RequestIdKey))
				assert.Equal(t, "local-request", recorder.Header().Get(common.RequestIdKey), "the local response request id must remain unchanged")
			}
		})
	}
}

func TestUpstreamTraceIsClearedBeforeRequestConstructionFails(t *testing.T) {
	service.InitHttpClient()
	gin.SetMode(gin.TestMode)
	client := service.GetHttpClient()
	require.NotNil(t, client)
	originalTransport := client.Transport
	t.Cleanup(func() { client.Transport = originalTransport })
	transportCalled := false
	client.Transport = upstreamTraceRoundTripper(func(*http.Request) (*http.Response, error) {
		transportCalled = true
		return nil, errors.New("unexpected transport call")
	})

	for _, kind := range []string{"chat", "task", "form"} {
		for _, failure := range []struct {
			name, requestURL  string
			urlErr, headerErr error
		}{
			{"url builder", "", errors.New("invalid provider route"), nil},
			{"invalid url", "\n", nil, nil},
			{"header builder", "https://upstream.test/relay", nil, errors.New("invalid provider header")},
		} {
			t.Run(kind+"/"+failure.name, func(t *testing.T) {
				ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
				ctx.Request = httptest.NewRequest(http.MethodPost, "/relay", nil)
				ctx.Set(common.UpstreamRequestIdKey, "previous-attempt")
				common.SetContextKey(ctx, rootconstant.ContextKeyUpstreamTransportError, "previous transport diagnostic")
				info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{}}
				resp, err := invokeUpstreamTraceRequest(kind, ctx, info, failure.requestURL, failure.urlErr, failure.headerErr)
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Empty(t, ctx.GetString(common.UpstreamRequestIdKey), "a local construction failure has no upstream trace")
				diagnostic, _ := common.GetContextKey(ctx, rootconstant.ContextKeyUpstreamTransportError)
				assert.Nil(t, diagnostic, "a construction failure must not inherit an unrelated transport diagnostic")
			})
		}
	}
	assert.False(t, transportCalled, "construction errors must not issue HTTP requests")
}

func TestUpstreamTraceIsClearedBeforeProxyConfigurationFails(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/relay", nil)
	ctx.Set(common.UpstreamRequestIdKey, "previous-attempt")
	common.SetContextKey(ctx, rootconstant.ContextKeyUpstreamTransportError, "previous transport diagnostic")
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{}}
	info.ChannelSetting.Proxy = "%"
	resp, err := invokeUpstreamTraceRequest("direct", ctx, info, "https://upstream.test/relay", nil, nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, ctx.GetString(common.UpstreamRequestIdKey))
	diagnostic, _ := common.GetContextKey(ctx, rootconstant.ContextKeyUpstreamTransportError)
	assert.Nil(t, diagnostic)
}
