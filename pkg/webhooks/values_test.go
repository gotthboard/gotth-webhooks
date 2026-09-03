package webhooks

import (
	"context"
	"net/netip"
	"strings"
	"testing"
	"time"
)

func TestParseEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"dns", "https://EXAMPLE.com/hooks/a?b=1", "https://example.com:443/hooks/a?b=1"},
		{"default port", "https://example.com:443", "https://example.com:443/"},
		{"ipv6", "https://[2606:4700:4700::1111]/hook", "https://[2606:4700:4700::1111]:443/hook"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseEndpoint(tc.raw)
			if err != nil {
				t.Fatal(err)
			}
			if got.canonical != tc.want {
				t.Fatalf("canonical = %q, want %q", got.canonical, tc.want)
			}
		})
	}
}

func TestParseEndpointRejectsUnsafeForms(t *testing.T) {
	t.Parallel()

	bad := []string{
		"http://example.com/hook",
		"https://user@example.com/hook",
		"https://example.com/hook#fragment",
		"https://example.com:8443/hook",
		"https://localhost/hook",
		"https://127.0.0.1/hook",
		"https://[::1]/hook",
		"https://exa_mple.com/hook",
		"https://example.com/%zz",
		"https://example.com/\nnext",
		"https://example.com/" + strings.Repeat("a", maxEndpointBytes),
	}
	for _, raw := range bad {
		if _, err := parseEndpoint(raw); err == nil {
			t.Errorf("parseEndpoint(%q) succeeded", raw)
		}
	}
}

func TestPublicAddressPolicyBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ip   string
		want bool
	}{
		{"8.8.8.8", true},
		{"9.255.255.255", true},
		{"10.0.0.0", false},
		{"10.255.255.255", false},
		{"11.0.0.0", true},
		{"100.63.255.255", true},
		{"100.64.0.0", false},
		{"100.127.255.255", false},
		{"100.128.0.0", true},
		{"169.254.0.0", false},
		{"172.15.255.255", true},
		{"172.16.0.0", false},
		{"172.31.255.255", false},
		{"172.32.0.0", true},
		{"192.0.2.1", false},
		{"192.168.1.1", false},
		{"198.51.100.1", false},
		{"203.0.113.1", false},
		{"224.0.0.0", false},
		{"2606:4700:4700::1111", true},
		{"::1", false},
		{"::ffff:127.0.0.1", false},
		{"2001:db8::1", false},
		{"64:ff9b::7f00:1", false},
		{"fc00::1", false},
		{"fe80::1", false},
		{"fec0::1", false},
		{"ff00::1", false},
	}
	for _, tc := range tests {
		ip := netip.MustParseAddr(tc.ip)
		if got := isPublicAddress(ip); got != tc.want {
			t.Errorf("isPublicAddress(%s) = %v, want %v", ip, got, tc.want)
		}
	}
}

func TestDeliveryFingerprintBindsSemanticsButNotSigningKey(t *testing.T) {
	t.Parallel()

	base, err := validateMessage(Message{Endpoint: "https://example.com/hook", DeliveryID: "id", EventType: "event", ContentType: "application/json", Body: []byte("body")})
	if err != nil {
		t.Fatal(err)
	}
	variants := []Message{
		{Endpoint: "https://example.org/hook", DeliveryID: "id", EventType: "event", ContentType: "application/json", Body: []byte("body")},
		{Endpoint: "https://example.com/other", DeliveryID: "id", EventType: "event", ContentType: "application/json", Body: []byte("body")},
		{Endpoint: "https://example.com/hook", DeliveryID: "id", EventType: "other", ContentType: "application/json", Body: []byte("body")},
		{Endpoint: "https://example.com/hook", DeliveryID: "id", EventType: "event", ContentType: "text/plain", Body: []byte("body")},
		{Endpoint: "https://example.com/hook", DeliveryID: "id", EventType: "event", ContentType: "application/json", Body: []byte("changed")},
	}
	for i, variant := range variants {
		got, err := validateMessage(variant)
		if err != nil {
			t.Fatal(err)
		}
		if got.fingerprint == base.fingerprint {
			t.Errorf("variant %d did not change fingerprint", i)
		}
	}
}

func TestValidateTokenBoundaries(t *testing.T) {
	t.Parallel()

	if err := validateToken("x", "delivery ID"); err != nil {
		t.Fatal(err)
	}
	if err := validateToken(strings.Repeat("a", maxTokenBytes), "delivery ID"); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"", strings.Repeat("a", maxTokenBytes+1), "a b", "a\nb", "é"} {
		if err := validateToken(value, "delivery ID"); err == nil {
			t.Errorf("validateToken(%q) succeeded", value)
		}
	}
}

func TestValidateAndDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg, err := validateConfig(Config{
		Secret:   Secret{KeyID: "key-1", Value: make([]byte, minSecretBytes)},
		Recorder: discardRecorder{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.retry.MaxAttempts != defaultMaxAttempts || cfg.attemptTimeout != defaultAttemptTimeout || cfg.receiptTimeout != defaultReceiptTimeout {
		t.Fatalf("defaults not applied: %+v", cfg)
	}

	bad := []Config{
		{},
		{Secret: Secret{KeyID: "key", Value: make([]byte, minSecretBytes)}, Recorder: nil},
		{Secret: Secret{KeyID: "key", Value: make([]byte, minSecretBytes-1)}, Recorder: discardRecorder{}},
		{Secret: Secret{KeyID: "key", Value: make([]byte, maxSecretBytes+1)}, Recorder: discardRecorder{}},
		{Secret: Secret{KeyID: "bad key", Value: make([]byte, minSecretBytes)}, Recorder: discardRecorder{}},
		{Secret: Secret{KeyID: "key", Value: make([]byte, minSecretBytes)}, Recorder: discardRecorder{}, Retry: RetryPolicy{MaxAttempts: 11, InitialDelay: time.Millisecond, MaxDelay: time.Second}},
		{Secret: Secret{KeyID: "key", Value: make([]byte, minSecretBytes)}, Recorder: discardRecorder{}, AttemptTimeout: time.Millisecond},
	}
	for i, candidate := range bad {
		if _, err := validateConfig(candidate); err == nil {
			t.Errorf("bad config %d succeeded", i)
		}
	}
}

type discardRecorder struct{}

func (discardRecorder) Record(_ context.Context, _ Receipt) error { return nil }
