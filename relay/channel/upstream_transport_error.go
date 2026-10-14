package channel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"syscall"

	hosttypes "github.com/QuantumNous/new-api/types"
)

func classifyUpstreamTransportError(err error) hosttypes.UpstreamTransportError {
	diagnosis := hosttypes.UpstreamTransportError{Category: "transport_error"}
	var networkError net.Error
	if errors.As(err, &networkError) {
		diagnosis.Timeout = networkError.Timeout()
	}
	var dnsError *net.DNSError
	var certificateError *tls.CertificateVerificationError
	var unknownAuthority x509.UnknownAuthorityError
	var invalidCertificate x509.CertificateInvalidError
	var hostnameError x509.HostnameError
	switch {
	case errors.Is(err, context.Canceled):
		diagnosis.Category = "request_canceled"
	case errors.Is(err, context.DeadlineExceeded):
		diagnosis.Category, diagnosis.Timeout = "timeout", true
	case errors.As(err, &dnsError):
		diagnosis.Category = "dns_error"
	case errors.As(err, &certificateError), errors.As(err, &unknownAuthority), errors.As(err, &invalidCertificate), errors.As(err, &hostnameError):
		diagnosis.Category = "tls_error"
	case errors.Is(err, syscall.ECONNREFUSED):
		diagnosis.Category = "connection_refused"
	case errors.Is(err, syscall.ECONNRESET), errors.Is(err, syscall.EPIPE):
		diagnosis.Category = "connection_reset"
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		diagnosis.Category = "connection_closed"
	case diagnosis.Timeout:
		diagnosis.Category = "timeout"
	}
	return diagnosis
}
