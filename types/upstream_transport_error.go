package types

// UpstreamTransportError retains a bounded, administrator-only diagnosis after
// the client-facing error has hidden the URL and other transport details.
type UpstreamTransportError struct {
	Category string `json:"category"`
	Timeout  bool   `json:"timeout"`
}
