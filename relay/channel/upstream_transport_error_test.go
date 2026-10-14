package channel

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/url"
	"syscall"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamTransportDiagnosisPreservesCauseWithoutRequestSecrets(t *testing.T) {
	for _, test := range []struct {
		name     string
		err      error
		category string
		timeout  bool
	}{
		{"canceled", context.Canceled, "request_canceled", false},
		{"deadline", context.DeadlineExceeded, "timeout", true},
		{"dns", &net.DNSError{Err: "lookup failed", Name: "private.example.test"}, "dns_error", false},
		{"dns timeout", &net.DNSError{Err: "lookup timed out", Name: "private.example.test", IsTimeout: true}, "dns_error", true},
		{"certificate", x509.UnknownAuthorityError{}, "tls_error", false},
		{"connection refused", &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}, "connection_refused", false},
		{"connection reset", &net.OpError{Op: "read", Err: syscall.ECONNRESET}, "connection_reset", false},
		{"closed connection", io.ErrUnexpectedEOF, "connection_closed", false},
		{"unclassified", errors.New("secret-response-detail"), "transport_error", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := &url.Error{Op: "Post", URL: "https://private-user:private-password@private.example.test/v1?api_key=private-key", Err: test.err}
			diagnosis := classifyUpstreamTransportError(err)
			assert.Equal(t, test.category, diagnosis.Category)
			assert.Equal(t, test.timeout, diagnosis.Timeout)
			encoded, marshalError := common.Marshal(diagnosis)
			require.NoError(t, marshalError)
			assert.NotContains(t, string(encoded), "private-")
			assert.NotContains(t, string(encoded), "secret-response-detail")
		})
	}
}
