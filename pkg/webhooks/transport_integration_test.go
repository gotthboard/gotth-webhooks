package webhooks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTLSLiteralDialAndMalformedRedirectDenialIntegration(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var paths []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()
		paths = append(paths, req.URL.Path)
		mu.Unlock()
		if req.TLS == nil || req.Host != "example.com:443" || req.Header.Get(HeaderSignature) == "" {
			t.Errorf("request lost TLS, canonical host, or signature")
		}
		body, err := io.ReadAll(req.Body)
		if err != nil || string(body) != "body" {
			t.Errorf("body=%q error=%v", body, err)
		}
		w.Header().Set("Location", ":bad")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	mapped := &mappingDialer{target: server.Listener.Addr().String()}
	checked := safeDialer{
		resolver: &sequenceResolver{answers: [][]netip.Addr{{netip.MustParseAddr("8.8.8.8")}}},
		dialer:   mapped,
	}
	transport := &http.Transport{
		Proxy:                  nil,
		DialContext:            checked.DialContext,
		TLSClientConfig:        &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots},
		MaxResponseHeaderBytes: 64 << 10,
	}
	defer transport.CloseIdleConnections()
	config, err := validateConfig(Config{Secret: Secret{KeyID: "key", Value: make([]byte, minSecretBytes)}, Recorder: &memoryRecorder{}, Retry: RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, AttemptTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	d := newDispatcher(config, dependencies{transport: transport, now: time.Now, wait: noWait})
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrPermanent) || result.StatusCode != http.StatusFound {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(paths) != 1 || paths[0] != "/hook" {
		t.Fatalf("redirect followed: paths=%v", paths)
	}
	if got := mapped.snapshot(); len(got) != 1 || got[0] != "8.8.8.8:443" {
		t.Fatalf("dial targets=%v", got)
	}
}

func TestProductionTransportPinsHTTP1AgainstHTTP2Server(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	requests := 0
	protocolMajor := 0
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()
		requests++
		protocolMajor = req.ProtoMajor
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	transport := newHTTPTransport(time.Second)
	defer transport.CloseIdleConnections()
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	transport.TLSClientConfig.RootCAs = roots
	checked := safeDialer{
		resolver: &sequenceResolver{answers: [][]netip.Addr{{netip.MustParseAddr("8.8.8.8")}}},
		dialer:   &mappingDialer{target: server.Listener.Addr().String()},
	}
	transport.DialContext = checked.DialContext

	d, _ := integrationDispatcher(t, transport, 1)
	result, err := d.Deliver(context.Background(), validMessage())
	if err != nil || !result.Delivered {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	mu.Lock()
	defer mu.Unlock()
	if requests != 1 || protocolMajor != 1 {
		t.Fatalf("requests=%d HTTP major=%d, want one HTTP/1 request", requests, protocolMajor)
	}
}

func TestResponseHeaderLimitIsPermanentIntegration(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Oversized", strings.Repeat("x", maxResponseHeaderBytes+1))
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	transport := mappedTLSTransport(server, true)
	defer transport.CloseIdleConnections()
	d, recorder := integrationDispatcher(t, transport, 2)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrPermanent) || result.Attempts != 1 || result.StatusCode != 0 || result.Outcome != OutcomePermanent {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	if receipts := recorder.snapshot(); len(receipts) != 1 || receipts[0].ErrorCode != ErrorTransport {
		t.Fatalf("receipts=%+v", receipts)
	}
}

func TestCertificateFailureIsPermanentIntegration(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	transport := mappedTLSTransport(server, false)
	defer transport.CloseIdleConnections()
	d, recorder := integrationDispatcher(t, transport, 2)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrPermanent) || result.Attempts != 1 || result.Outcome != OutcomePermanent {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	if receipts := recorder.snapshot(); len(receipts) != 1 || receipts[0].ErrorCode != ErrorTransport {
		t.Fatalf("receipts=%+v", receipts)
	}
}

func mappedTLSTransport(server *httptest.Server, trust bool) *http.Transport {
	roots := x509.NewCertPool()
	if trust {
		roots.AddCert(server.Certificate())
	}
	checked := safeDialer{
		resolver: &sequenceResolver{answers: [][]netip.Addr{{netip.MustParseAddr("8.8.8.8")}}},
		dialer:   &mappingDialer{target: server.Listener.Addr().String()},
	}
	return &http.Transport{
		Proxy:                  nil,
		DialContext:            checked.DialContext,
		TLSClientConfig:        &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots},
		MaxResponseHeaderBytes: maxResponseHeaderBytes,
	}
}

func integrationDispatcher(t *testing.T, transport http.RoundTripper, attempts int) (*Dispatcher, *memoryRecorder) {
	t.Helper()
	recorder := &memoryRecorder{}
	config, err := validateConfig(Config{
		Secret:         Secret{KeyID: "key", Value: make([]byte, minSecretBytes)},
		Recorder:       recorder,
		Retry:          RetryPolicy{MaxAttempts: attempts, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond},
		AttemptTimeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	return newDispatcher(config, dependencies{transport: transport, now: time.Now, wait: noWait}), recorder
}

type mappingDialer struct {
	mu        sync.Mutex
	target    string
	addresses []string
}

func (d *mappingDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d.mu.Lock()
	d.addresses = append(d.addresses, address)
	d.mu.Unlock()
	return (&net.Dialer{}).DialContext(ctx, network, d.target)
}

func (d *mappingDialer) snapshot() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.addresses...)
}
