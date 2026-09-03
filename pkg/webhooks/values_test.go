package webhooks

import (
	"context"
	"errors"
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
		{"query order and escapes", "https://example.com/hook?b=2&a=%2F%3f&flag&x=one+two/three?four", "https://example.com:443/hook?b=2&a=%2F%3f&flag&x=one+two/three?four"},
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
		"https://example.com:/hook",
		"https://[2001:4860:4860::8888]:/hook",
		"https://example.com:8443/hook",
		"https://localhost/hook",
		"https://127.0.0.1/hook",
		"https://[::1]/hook",
		"https://exa_mple.com/hook",
		"https://example.com/%zz",
		"https://example.com/hook?q=has space",
		"https://example.com/hook?q=%",
		"https://example.com/hook?q=%0",
		"https://example.com/hook?q=%zz",
		"https://example.com/hook?q=[bad]",
		"https://example.com/hook?q=bad\\query",
		"https://example.com/hook?q=bad\tquery",
		"https://example.com/hook?q=\x80",
		"https://example.com/\nnext",
		"https://example.com/" + strings.Repeat("a", maxEndpointBytes),
	}
	for _, raw := range bad {
		if _, err := parseEndpoint(raw); err == nil {
			t.Errorf("parseEndpoint(%q) succeeded", raw)
		}
	}
}

func TestIANASpecialPurposeRegistrySnapshot(t *testing.T) {
	t.Parallel()

	// Compact encompassing prefixes from the IPv4 and IPv6 registries last
	// updated 2025-10-09. Nested registry allocations are covered by the broad
	// registry entries 192.0.0.0/24 and 2001::/23.
	want := []string{
		"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8",
		"169.254.0.0/16", "172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24",
		"192.31.196.0/24", "192.52.193.0/24", "192.88.99.0/24", "192.168.0.0/16",
		"192.175.48.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24",
		"240.0.0.0/4", "::/128", "::1/128", "::ffff:0:0/96", "64:ff9b::/96",
		"64:ff9b:1::/48", "100::/64", "100:0:0:1::/64", "2001::/23",
		"2001:db8::/32", "2002::/16", "2620:4f:8000::/48", "3fff::/20",
		"5f00::/16", "fc00::/7", "fe80::/10",
	}
	if len(ianaSpecialPurposePrefixes) != len(want) {
		t.Fatalf("special prefix count=%d want=%d", len(ianaSpecialPurposePrefixes), len(want))
	}
	for i, text := range want {
		prefix := netip.MustParsePrefix(text).Masked()
		if got := ianaSpecialPurposePrefixes[i].Masked(); got != prefix {
			t.Fatalf("special prefix[%d]=%s want=%s", i, got, prefix)
		}
		for _, address := range []netip.Addr{prefix.Addr(), lastAddress(prefix)} {
			if isPublicAddress(address) {
				t.Errorf("registry boundary %s in %s was accepted", address, prefix)
			}
		}
	}
}

func TestAllocatedGlobalIPv6Boundary(t *testing.T) {
	t.Parallel()

	for _, address := range []string{"100:0:0:1::1", "1fff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", "3fff::1", "4000::", "5f00::1"} {
		if isPublicAddress(netip.MustParseAddr(address)) {
			t.Errorf("unallocated or special IPv6 address %s was accepted", address)
		}
	}
	if !isPublicAddress(netip.MustParseAddr("2606:4700:4700::1111")) {
		t.Fatal("ordinary allocated public IPv6 address was rejected")
	}
}

func lastAddress(prefix netip.Prefix) netip.Addr {
	prefix = prefix.Masked()
	if prefix.Addr().Is4() {
		bytes := prefix.Addr().As4()
		for bit := prefix.Bits(); bit < 32; bit++ {
			bytes[bit/8] |= 1 << (7 - bit%8)
		}
		return netip.AddrFrom4(bytes)
	}
	bytes := prefix.Addr().As16()
	for bit := prefix.Bits(); bit < 128; bit++ {
		bytes[bit/8] |= 1 << (7 - bit%8)
	}
	return netip.AddrFrom16(bytes)
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

func TestCanonicalContentTypeRejectsEveryASCIIControlPosition(t *testing.T) {
	t.Parallel()

	for _, value := range contentTypeControlValues() {
		if _, err := canonicalContentType(value); !errors.Is(err, ErrInvalid) {
			t.Errorf("canonicalContentType(%q) error = %v", value, err)
		}
	}
}

func TestCanonicalContentTypePreservesSupportedText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "optional spaces and quoted space",
			value: ` Application/JSON ; Charset="UTF-8"; note="hello world" `,
			want:  `application/json; charset=UTF-8; note="hello world"`,
		},
		{
			name:  "quoted UTF-8 parameter",
			value: `text/plain; title="café"`,
			want:  `text/plain; title*=utf-8''caf%C3%A9`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := canonicalContentType(tc.value)
			if err != nil || got != tc.want {
				t.Fatalf("canonicalContentType(%q) = %q, %v; want %q", tc.value, got, err, tc.want)
			}
		})
	}
}

func contentTypeControlValues() []string {
	values := make([]string, 0, 33*3)
	controls := make([]byte, 0, 33)
	for control := byte(0); control < 0x20; control++ {
		controls = append(controls, control)
	}
	controls = append(controls, 0x7f)
	for _, control := range controls {
		values = append(values,
			string([]byte{control})+"application/json",
			"application/json"+string([]byte{control}),
			`application/json; note="a`+string([]byte{control})+`b"`,
		)
	}
	return values
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
