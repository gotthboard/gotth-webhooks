package webhooks

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type validatedConfig struct {
	secret         Secret
	retry          RetryPolicy
	attemptTimeout time.Duration
	receiptTimeout time.Duration
	recorder       Recorder
}

type validatedMessage struct {
	endpoint     endpoint
	deliveryID   string
	firstAttempt int
	eventType    string
	contentType  string
	body         []byte
	fingerprint  [32]byte
}

type endpoint struct {
	url       *url.URL
	canonical string
}

// ianaSpecialPurposePrefixes is the compact, encompassing-prefix form of every
// allocation in the IANA IPv4 and IPv6 Special-Purpose Address Registries,
// both last updated 2025-10-09. The policy rejects every registry allocation,
// including entries whose registry Global flag is true. See the pinned source
// hashes in workflow/features/outbound-v1-admission/evidence/authority.md.
var ianaSpecialPurposePrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.31.196.0/24"),
	netip.MustParsePrefix("192.52.193.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("192.175.48.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("::ffff:0:0/96"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("100:0:0:1::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("2620:4f:8000::/48"),
	netip.MustParsePrefix("3fff::/20"),
	netip.MustParsePrefix("5f00::/16"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
}

var allocatedGlobalIPv6 = netip.MustParsePrefix("2000::/3")

// NewDeliveryID returns a 192-bit CSPRNG-backed, base64url delivery identity.
// All-input time and auxiliary space are O(n), Omega(1), with no single tight
// bound because entropy acquisition can fail early; the successful path is
// Theta(n). n is the fixed 24-byte entropy input; delegated costs are
// crypto/rand.Reader and base64 encoding.
func NewDeliveryID() (string, error) {
	return newDeliveryID(rand.Reader)
}

// newDeliveryID isolates the entropy read so its failure contract is directly
// testable. All-input time and auxiliary space are O(n), Omega(1), with no
// single tight bound because the delegated reader can fail before n bytes; the
// successful read/encode path is Theta(n). n is the fixed 24-byte entropy
// input.
func newDeliveryID(source io.Reader) (string, error) {
	var raw [24]byte
	if _, err := io.ReadFull(source, raw[:]); err != nil {
		return "", fmt.Errorf("delivery ID entropy: %w", err)
	}
	return "wh_" + base64.RawURLEncoding.EncodeToString(raw[:]), nil
}

// validateConfig applies all-zero defaults, validates fixed bounds, and copies
// secret material. All-input time is O(1+q+k) and auxiliary space O(1+k), both
// Omega(1), with no single tight bound because recorder, key, or length checks
// can reject before secret processing. The admitted path is Theta(q+k) time and
// Theta(k) auxiliary space due to token scanning and the secret copy; q and k
// are key-ID and secret bytes.
func validateConfig(cfg Config) (validatedConfig, error) {
	if cfg.Recorder == nil {
		return validatedConfig{}, fmt.Errorf("%w: recorder is required", ErrInvalid)
	}
	if err := validateToken(cfg.Secret.KeyID, "key ID"); err != nil {
		return validatedConfig{}, err
	}
	if len(cfg.Secret.Value) < minSecretBytes || len(cfg.Secret.Value) > maxSecretBytes {
		return validatedConfig{}, fmt.Errorf("%w: secret length must be %d..%d bytes", ErrInvalid, minSecretBytes, maxSecretBytes)
	}

	retry := cfg.Retry
	if retry == (RetryPolicy{}) {
		retry = RetryPolicy{MaxAttempts: defaultMaxAttempts, InitialDelay: defaultInitialDelay, MaxDelay: defaultMaxDelay}
	}
	if retry.MaxAttempts < 1 || retry.MaxAttempts > 10 || retry.InitialDelay < time.Millisecond || retry.MaxDelay < retry.InitialDelay || retry.MaxDelay > time.Minute {
		return validatedConfig{}, fmt.Errorf("%w: invalid retry policy", ErrInvalid)
	}
	attemptTimeout := cfg.AttemptTimeout
	if attemptTimeout == 0 {
		attemptTimeout = defaultAttemptTimeout
	}
	if attemptTimeout < 100*time.Millisecond || attemptTimeout > 2*time.Minute {
		return validatedConfig{}, fmt.Errorf("%w: invalid attempt timeout", ErrInvalid)
	}
	receiptTimeout := cfg.ReceiptTimeout
	if receiptTimeout == 0 {
		receiptTimeout = defaultReceiptTimeout
	}
	if receiptTimeout < 100*time.Millisecond || receiptTimeout > 30*time.Second {
		return validatedConfig{}, fmt.Errorf("%w: invalid receipt timeout", ErrInvalid)
	}
	secret := Secret{KeyID: cfg.Secret.KeyID, Value: append([]byte(nil), cfg.Secret.Value...)}
	return validatedConfig{secret: secret, retry: retry, attemptTimeout: attemptTimeout, receiptTimeout: receiptTimeout, recorder: cfg.Recorder}, nil
}

// validateMessage validates and copies the complete caller boundary.
// All-input time is O(V(e,c)+d+v+b) and auxiliary space O(W(e,c)+v+b), both
// Omega(1), with no single tight bound because ordered checks can reject before
// later fields. V/W are delegated endpoint and MIME validation costs; e, c, d,
// v, and b are endpoint, content-type, delivery-ID, event-type, and body bytes.
// The fully accepted path scans d/v and copies and hashes all b body bytes.
func validateMessage(msg Message) (validatedMessage, error) {
	ep, err := parseEndpoint(msg.Endpoint)
	if err != nil {
		return validatedMessage{}, err
	}
	if err := validateToken(msg.DeliveryID, "delivery ID"); err != nil {
		return validatedMessage{}, err
	}
	if err := validateToken(msg.EventType, "event type"); err != nil {
		return validatedMessage{}, err
	}
	firstAttempt := msg.FirstAttempt
	if firstAttempt == 0 {
		firstAttempt = 1
	}
	if firstAttempt < 1 || firstAttempt > 1_000_000_000 {
		return validatedMessage{}, fmt.Errorf("%w: first attempt must be 1..1000000000", ErrInvalid)
	}
	contentType, err := canonicalContentType(msg.ContentType)
	if err != nil {
		return validatedMessage{}, err
	}
	if len(msg.Body) > MaxPayloadBytes {
		return validatedMessage{}, fmt.Errorf("%w: body exceeds %d bytes", ErrInvalid, MaxPayloadBytes)
	}
	body := append([]byte(nil), msg.Body...)
	return validatedMessage{endpoint: ep, deliveryID: msg.DeliveryID, firstAttempt: firstAttempt, eventType: msg.EventType, contentType: contentType, body: body, fingerprint: deliveryFingerprint(ep.canonical, msg.EventType, contentType, body)}, nil
}

// deliveryFingerprint binds immutable logical-delivery semantics independently
// of attempt timestamp and signing-key rotation. Time is
// Theta(1+e+v+c+b); auxiliary space is Theta(1+e+v+c); e, v, c, and b are
// endpoint, event, content-type, and body bytes. SHA-256 state is constant and
// the body is streamed into it.
func deliveryFingerprint(endpoint, eventType, contentType string, body []byte) [32]byte {
	hash := sha256.New()
	_, _ = io.WriteString(hash, "gotth-webhook-delivery-v1\n")
	_, _ = io.WriteString(hash, endpoint+"\n")
	_, _ = io.WriteString(hash, eventType+"\n")
	_, _ = io.WriteString(hash, contentType+"\n")
	_, _ = hash.Write(body)
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

// validateToken accepts a deliberately small ASCII metadata alphabet.
// Complexity: time O(n), Omega(1), no input-independent tight Theta bound;
// auxiliary space O(1), Omega(1), tight Theta(1); n is token bytes.
func validateToken(value, name string) error {
	if len(value) == 0 || len(value) > maxTokenBytes {
		return fmt.Errorf("%w: %s length must be 1..%d bytes", ErrInvalid, name, maxTokenBytes)
	}
	for i := range len(value) {
		c := value[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || strings.ContainsRune("._:-", rune(c))) {
			return fmt.Errorf("%w: %s contains an invalid byte", ErrInvalid, name)
		}
	}
	return nil
}

// canonicalContentType rejects ASCII controls, then parses and deterministically
// formats one MIME media type. Non-ASCII bytes retain mime.ParseMediaType's
// policy and are serialized by mime.FormatMediaType. All-input time is
// O(n+Mtime(n)) and auxiliary space O(n+Mspace(n)), both Omega(1), with no
// single tight bound because the length guard or control scan can reject before
// parsing. Mtime/Mspace are delegated mime.ParseMediaType and
// mime.FormatMediaType costs for n admitted-length input bytes.
func canonicalContentType(value string) (string, error) {
	if len(value) == 0 || len(value) > maxContentType {
		return "", fmt.Errorf("%w: content type length must be 1..%d bytes", ErrInvalid, maxContentType)
	}
	for i := range len(value) {
		if value[i] < 0x20 || value[i] == 0x7f {
			return "", fmt.Errorf("%w: content type contains an ASCII control byte", ErrInvalid)
		}
	}
	mediaType, params, err := mime.ParseMediaType(value)
	if err != nil {
		return "", fmt.Errorf("%w: invalid content type", ErrInvalid)
	}
	canonical := mime.FormatMediaType(mediaType, params)
	if canonical == "" || len(canonical) > maxContentType {
		return "", fmt.Errorf("%w: invalid content type", ErrInvalid)
	}
	return canonical, nil
}

// parseEndpoint returns the exact normalized HTTPS target used for both the
// request and signature. All-input time is O(n+Utime(n)) and auxiliary space
// O(n+Uspace(n)), both Omega(1), with no single tight bound because the length
// and syntax checks can reject before normalization. Utime/Uspace are delegated
// URL parsing costs; an accepted n-byte endpoint creates proportional
// lowercase/canonical strings.
func parseEndpoint(raw string) (endpoint, error) {
	if len(raw) == 0 || len(raw) > maxEndpointBytes {
		return endpoint{}, fmt.Errorf("%w: endpoint length must be 1..%d bytes", ErrInvalid, maxEndpointBytes)
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Opaque != "" || u.User != nil || u.Fragment != "" || u.RawFragment != "" {
		return endpoint{}, fmt.Errorf("%w: endpoint must be an unambiguous HTTPS URL", ErrInvalid)
	}
	if strings.HasSuffix(u.Host, ":") {
		return endpoint{}, fmt.Errorf("%w: endpoint port must be omitted or 443", ErrInvalid)
	}
	host := strings.ToLower(u.Hostname())
	if host == "" || strings.Contains(host, "%") || !validASCIIHost(host) {
		return endpoint{}, fmt.Errorf("%w: invalid endpoint host", ErrInvalid)
	}
	port := u.Port()
	if port != "" && port != "443" {
		return endpoint{}, fmt.Errorf("%w: endpoint port must be 443", ErrInvalid)
	}
	if ip, err := netip.ParseAddr(host); err == nil && !isPublicAddress(ip) {
		return endpoint{}, fmt.Errorf("%w: endpoint IP is not public", ErrDestination)
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	authority := host
	if strings.Contains(host, ":") {
		authority = "[" + host + "]"
	}
	authority += ":443"
	if !validRawQuery(u.RawQuery) {
		return endpoint{}, fmt.Errorf("%w: invalid endpoint query", ErrInvalid)
	}
	canonical := "https://" + authority + path
	if u.ForceQuery || u.RawQuery != "" {
		canonical += "?" + u.RawQuery
	}
	parsed, err := url.Parse(canonical)
	if err != nil {
		return endpoint{}, fmt.Errorf("%w: canonical endpoint", ErrInvalid)
	}
	return endpoint{url: parsed, canonical: canonical}, nil
}

// validRawQuery accepts RFC 3986 query bytes exactly as supplied. It preserves
// order and escape spelling while rejecting malformed percent triplets and
// bytes outside pchar / "/" / "?". Complexity: time O(n), Omega(1), no
// input-independent tight Theta bound; auxiliary space O(1), Omega(1), tight
// Theta(1); n is raw-query bytes.
func validRawQuery(raw string) bool {
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c == '%' {
			if i+2 >= len(raw) || !isHex(raw[i+1]) || !isHex(raw[i+2]) {
				return false
			}
			i += 2
			continue
		}
		if !isQueryByte(c) {
			return false
		}
	}
	return true
}

// isHex classifies one byte in constant time and space.
func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// isQueryByte classifies one byte against the fixed RFC 3986 query alphabet in
// constant time and space.
func isQueryByte(c byte) bool {
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
		return true
	}
	return strings.ContainsRune("-._~!$&'()*+,;=:@/?", rune(c))
}

// validASCIIHost validates an IP literal or an RFC-compatible conservative
// DNS label subset. All-input time is O(n+Itime(n)) and auxiliary space
// O(n+Ispace(n)), both Omega(1), with no input-independent tight bound. Itime/
// Ispace are delegated netip.ParseAddr costs; an accepted DNS path scans n host
// bytes and allocates label views, while IP and early-rejection paths diverge.
func validASCIIHost(host string) bool {
	if ip, err := netip.ParseAddr(host); err == nil {
		return ip.Zone() == ""
	}
	if len(host) > 253 || strings.HasSuffix(host, ".") || !strings.Contains(host, ".") {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := range len(label) {
			c := label[i]
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
				return false
			}
		}
	}
	return true
}

// isPublicAddress admits only global unicast addresses outside every pinned
// IANA special-purpose allocation. IPv6 must also be in IANA's allocated
// 2000::/3 global-unicast block. Complexity: time O(p), Omega(1), no
// input-independent tight Theta bound; auxiliary space O(1), Omega(1), tight
// Theta(1); p is the fixed denied-prefix table length.
func isPublicAddress(ip netip.Addr) bool {
	if !ip.IsValid() {
		return false
	}
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() {
		return false
	}
	if ip.Is6() && !allocatedGlobalIPv6.Contains(ip) {
		return false
	}
	for _, prefix := range ianaSpecialPurposePrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

// canonicalPort validates a dial target and returns its host and fixed port.
// All-input time is S(n)+O(1), Omega(1), with no tighter bound asserted by the
// public contract; S(n) is delegated net.SplitHostPort work for n address bytes.
// Successful-path auxiliary space is O(1), Omega(1), tight Theta(1), because
// returned host/port strings are substring views. Rejection additionally
// delegates bounded error allocation to net.SplitHostPort and fmt.Errorf.
func canonicalPort(address string) (string, string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil || port != strconv.Itoa(443) {
		return "", "", fmt.Errorf("%w: unexpected dial address", ErrDestination)
	}
	return host, port, nil
}
