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
	"sync"
	"testing"
	"time"
)

func TestTLSLiteralDialAndRedirectDenialIntegration(t *testing.T) {
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
		http.Redirect(w, req, "/must-not-run", http.StatusFound)
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
	client := &http.Client{Transport: transport, Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer client.CloseIdleConnections()
	config, err := validateConfig(Config{Secret: Secret{KeyID: "key", Value: make([]byte, minSecretBytes)}, Recorder: &memoryRecorder{}, Retry: RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, AttemptTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	d := newDispatcher(config, dependencies{client: client, now: time.Now, wait: noWait})
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
