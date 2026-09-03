package webhooks

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"testing"
	"time"
)

func TestNewAndProductionTransportPolicy(t *testing.T) {
	t.Parallel()

	if _, err := New(Config{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid New error = %v", err)
	}
	d, err := New(Config{Secret: Secret{KeyID: "key-1", Value: []byte(strings.Repeat("s", minSecretBytes))}, Recorder: discardRecorder{}})
	if err != nil {
		t.Fatal(err)
	}
	transport, ok := d.transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", d.transport)
	}
	if transport.Proxy != nil || transport.DialContext == nil || transport.MaxResponseHeaderBytes != maxResponseHeaderBytes || transport.TLSClientConfig.MinVersion != 0x0303 {
		t.Fatalf("unsafe transport policy: %+v", transport)
	}
}

func TestNewDeliveryID(t *testing.T) {
	t.Parallel()

	seen := make(map[string]struct{}, 128)
	for range 128 {
		id, err := NewDeliveryID()
		if err != nil {
			t.Fatal(err)
		}
		if err := validateToken(id, "delivery ID"); err != nil {
			t.Fatal(err)
		}
		if _, exists := seen[id]; exists {
			t.Fatalf("duplicate delivery ID %q", id)
		}
		seen[id] = struct{}{}
	}
	if _, err := newDeliveryID(errorReader{err: errors.New("entropy unavailable")}); err == nil {
		t.Fatal("entropy failure was ignored")
	}
}

func TestValidateMessageBoundaries(t *testing.T) {
	t.Parallel()

	for _, size := range []int{MaxPayloadBytes - 1, MaxPayloadBytes} {
		msg := validMessage()
		msg.Body = make([]byte, size)
		got, err := validateMessage(msg)
		if err != nil || len(got.body) != size {
			t.Fatalf("size %d: body=%d error=%v", size, len(got.body), err)
		}
	}
	msg := validMessage()
	for _, size := range []int{MaxPayloadBytes + 1, MaxPayloadBytes * 4} {
		msg.Body = make([]byte, size)
		if _, err := validateMessage(msg); !errors.Is(err, ErrInvalid) {
			t.Fatalf("oversize %d error = %v", size, err)
		}
	}
	for _, mutate := range []func(*Message){
		func(m *Message) { m.DeliveryID = "bad id" },
		func(m *Message) { m.EventType = "bad\nevent" },
		func(m *Message) { m.ContentType = "not a media type" },
		func(m *Message) { m.ContentType = strings.Repeat("a", maxContentType+1) },
		func(m *Message) { m.FirstAttempt = -1 },
		func(m *Message) { m.FirstAttempt = 1_000_000_001 },
	} {
		candidate := validMessage()
		mutate(&candidate)
		if _, err := validateMessage(candidate); !errors.Is(err, ErrInvalid) {
			t.Errorf("candidate %+v error = %v", candidate, err)
		}
	}
	msg = validMessage()
	msg.ContentType = "Application/JSON; Charset=UTF-8"
	got, err := validateMessage(msg)
	if err != nil || got.contentType != "application/json; charset=UTF-8" {
		t.Fatalf("canonical content type=%q error=%v", got.contentType, err)
	}
}

func TestConfigurationLimitBoundaries(t *testing.T) {
	t.Parallel()

	base := func() Config {
		return Config{
			Secret:         Secret{KeyID: "key", Value: make([]byte, minSecretBytes)},
			Recorder:       discardRecorder{},
			Retry:          RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond},
			AttemptTimeout: 100 * time.Millisecond,
			ReceiptTimeout: 100 * time.Millisecond,
		}
	}
	for _, size := range []int{minSecretBytes, maxSecretBytes} {
		cfg := base()
		cfg.Secret.Value = make([]byte, size)
		if _, err := validateConfig(cfg); err != nil {
			t.Errorf("secret size %d: %v", size, err)
		}
	}
	for _, size := range []int{minSecretBytes - 1, maxSecretBytes + 1, maxSecretBytes * 4} {
		cfg := base()
		cfg.Secret.Value = make([]byte, size)
		if _, err := validateConfig(cfg); !errors.Is(err, ErrInvalid) {
			t.Errorf("invalid secret size %d: %v", size, err)
		}
	}
	for _, attempts := range []int{1, 10} {
		cfg := base()
		cfg.Retry.MaxAttempts = attempts
		if _, err := validateConfig(cfg); err != nil {
			t.Errorf("attempts %d: %v", attempts, err)
		}
	}
	for _, attempts := range []int{-1, 0, 11, 100} {
		cfg := base()
		cfg.Retry.MaxAttempts = attempts
		if _, err := validateConfig(cfg); !errors.Is(err, ErrInvalid) {
			t.Errorf("invalid attempts %d: %v", attempts, err)
		}
	}
	for _, timeout := range []time.Duration{100 * time.Millisecond, 2 * time.Minute} {
		cfg := base()
		cfg.AttemptTimeout = timeout
		if _, err := validateConfig(cfg); err != nil {
			t.Errorf("attempt timeout %s: %v", timeout, err)
		}
	}
	for _, timeout := range []time.Duration{100*time.Millisecond - 1, 2*time.Minute + 1, 10 * time.Minute} {
		cfg := base()
		cfg.AttemptTimeout = timeout
		if _, err := validateConfig(cfg); !errors.Is(err, ErrInvalid) {
			t.Errorf("invalid attempt timeout %s: %v", timeout, err)
		}
	}
	for _, timeout := range []time.Duration{100 * time.Millisecond, 30 * time.Second} {
		cfg := base()
		cfg.ReceiptTimeout = timeout
		if _, err := validateConfig(cfg); err != nil {
			t.Errorf("receipt timeout %s: %v", timeout, err)
		}
	}
	for _, timeout := range []time.Duration{100*time.Millisecond - 1, 30*time.Second + 1, 5 * time.Minute} {
		cfg := base()
		cfg.ReceiptTimeout = timeout
		if _, err := validateConfig(cfg); !errors.Is(err, ErrInvalid) {
			t.Errorf("invalid receipt timeout %s: %v", timeout, err)
		}
	}
	for _, retry := range []RetryPolicy{
		{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond},
		{MaxAttempts: 10, InitialDelay: time.Minute, MaxDelay: time.Minute},
	} {
		cfg := base()
		cfg.Retry = retry
		if _, err := validateConfig(cfg); err != nil {
			t.Errorf("retry %+v: %v", retry, err)
		}
	}
	for _, retry := range []RetryPolicy{
		{MaxAttempts: 1, InitialDelay: time.Millisecond - 1, MaxDelay: time.Millisecond},
		{MaxAttempts: 1, InitialDelay: time.Second, MaxDelay: time.Second - 1},
		{MaxAttempts: 1, InitialDelay: time.Second, MaxDelay: time.Minute + 1},
		{MaxAttempts: 1, InitialDelay: time.Second, MaxDelay: 10 * time.Minute},
	} {
		cfg := base()
		cfg.Retry = retry
		if _, err := validateConfig(cfg); !errors.Is(err, ErrInvalid) {
			t.Errorf("invalid retry %+v: %v", retry, err)
		}
	}
}

func TestContentTypeLengthBoundaries(t *testing.T) {
	t.Parallel()

	validOfSize := func(size int) string { return "application/" + strings.Repeat("a", size-len("application/")) }
	for _, size := range []int{maxContentType - 1, maxContentType} {
		if got, err := canonicalContentType(validOfSize(size)); err != nil || len(got) != size {
			t.Errorf("content type size %d got=%d error=%v", size, len(got), err)
		}
	}
	for _, size := range []int{maxContentType + 1, maxContentType * 4} {
		if _, err := canonicalContentType(validOfSize(size)); !errors.Is(err, ErrInvalid) {
			t.Errorf("invalid content type size %d error=%v", size, err)
		}
	}
}

func TestDeliverRejectsAttemptSequenceOverflow(t *testing.T) {
	t.Parallel()

	d := testDispatcher(t, &memoryRecorder{}, &scriptedTransport{}, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	msg := validMessage()
	msg.FirstAttempt = 1_000_000_000
	if _, err := d.Deliver(context.Background(), msg); !errors.Is(err, ErrInvalid) {
		t.Fatalf("overflow error=%v", err)
	}
}

func TestRetryDelayAndRetryAfterForms(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0).UTC()
	policy := RetryPolicy{InitialDelay: time.Second, MaxDelay: 5 * time.Second}
	tests := []struct {
		attempt int
		header  string
		want    time.Duration
	}{
		{1, "", time.Second},
		{2, "", 2 * time.Second},
		{4, "", 5 * time.Second},
		{1, "3", 3 * time.Second},
		{1, "3600", 5 * time.Second},
		{1, "+5", time.Second},
		{1, "9223372036854775807", 5 * time.Second},
		{1, strings.Repeat("9", 1000), 5 * time.Second},
		{1, "-1", time.Second},
		{1, "garbage", time.Second},
		{1, now.Add(4 * time.Second).Format(http.TimeFormat), 4 * time.Second},
		{1, now.Add(-time.Second).Format(http.TimeFormat), time.Second},
	}
	for _, tc := range tests {
		if got := retryDelay(policy, tc.attempt, tc.header, now); got != tc.want {
			t.Errorf("attempt=%d header=%q got=%s want=%s", tc.attempt, tc.header, got, tc.want)
		}
	}
}

func TestRetryAfterHeaderBoundAndSaturation(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0).UTC()
	for _, size := range []int{maxResponseHeaderBytes - 1, maxResponseHeaderBytes} {
		value := strings.Repeat("9", size)
		if got, ok := parseRetryAfter(value, now, 1500*time.Millisecond); !ok || got != 1500*time.Millisecond {
			t.Fatalf("size=%d delay=%s ok=%v", size, got, ok)
		}
	}
	if _, ok := parseRetryAfter(strings.Repeat("9", maxResponseHeaderBytes+1), now, time.Minute); ok {
		t.Fatal("oversized Retry-After accepted")
	}
	for _, value := range []string{"+5", "-1", "1x", "\x805"} {
		if _, ok := parseRetryAfter(value, now, time.Minute); ok {
			t.Errorf("invalid delay-seconds %q accepted", value)
		}
	}
}

func TestConsumeResponseBoundariesAndFailures(t *testing.T) {
	t.Parallel()

	for _, size := range []int{MaxResponseBytes - 1, MaxResponseBytes, MaxResponseBytes + 1, MaxResponseBytes * 4} {
		body := &trackedBody{Reader: strings.NewReader(strings.Repeat("x", size))}
		n, err := consumeResponse(body)
		if !body.closed {
			t.Errorf("size %d body not closed", size)
		}
		if size <= MaxResponseBytes && (err != nil || n != int64(size)) {
			t.Errorf("size %d n=%d err=%v", size, n, err)
		}
		if size > MaxResponseBytes && (!errors.Is(err, ErrResponseTooLarge) || n != MaxResponseBytes+1) {
			t.Errorf("size %d n=%d err=%v", size, n, err)
		}
	}
	if n, err := consumeResponse(nil); n != 0 || err != nil {
		t.Fatalf("nil body n=%d err=%v", n, err)
	}
	wantRead := errors.New("read failure")
	if _, err := consumeResponse(&trackedBody{Reader: errorReader{err: wantRead}}); !errors.Is(err, wantRead) {
		t.Fatalf("read error = %v", err)
	}
	wantClose := errors.New("close failure")
	if _, err := consumeResponse(&trackedBody{Reader: strings.NewReader(""), closeErr: wantClose}); !errors.Is(err, wantClose) {
		t.Fatalf("close error = %v", err)
	}
}

func TestWaitContextCompletesAndCancels(t *testing.T) {
	t.Parallel()

	if err := waitContext(context.Background(), time.Millisecond); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitContext(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error = %v", err)
	}
}

func TestDeliverPreflightAndWaitCancellation(t *testing.T) {
	t.Parallel()

	recorder := &memoryRecorder{}
	transport := &scriptedTransport{steps: []transportStep{{status: http.StatusInternalServerError}}}
	d := testDispatcher(t, recorder, transport, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, func(ctx context.Context, _ time.Duration) error {
		return context.Canceled
	})
	if _, err := d.Deliver(nil, validMessage()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("nil context error=%v", err)
	}
	invalid := validMessage()
	invalid.Endpoint = "http://example.com"
	if _, err := d.Deliver(context.Background(), invalid); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid message error=%v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := d.Deliver(ctx, validMessage())
	if !errors.Is(err, context.Canceled) || result.Attempts != 0 || transport.callCount() != 0 {
		t.Fatalf("pre-cancel result=%+v err=%v calls=%d", result, err, transport.callCount())
	}
	result, err = d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, context.Canceled) || result.Attempts != 1 || transport.callCount() != 1 {
		t.Fatalf("wait cancel result=%+v err=%v calls=%d", result, err, transport.callCount())
	}
}

func TestDeliverRejectsInvalidRawQueryBeforeTransport(t *testing.T) {
	t.Parallel()

	transport := &scriptedTransport{}
	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, transport, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	for _, endpoint := range []string{
		"https://example.com/hook?q=raw space",
		"https://example.com/hook?q=%zz",
		"https://example.com/hook?q=raw[bracket]",
	} {
		msg := validMessage()
		msg.Endpoint = endpoint
		if result, err := d.Deliver(context.Background(), msg); !errors.Is(err, ErrInvalid) || result.Attempts != 0 {
			t.Errorf("endpoint=%q result=%+v error=%v", endpoint, result, err)
		}
	}
	if transport.callCount() != 0 || len(recorder.snapshot()) != 0 {
		t.Fatalf("invalid queries reached side effects: calls=%d receipts=%d", transport.callCount(), len(recorder.snapshot()))
	}
}

func TestSafeDialerAdditionalFailurePaths(t *testing.T) {
	t.Parallel()

	lookupFailure := &resolverFunc{fn: func(context.Context, string, string) ([]netip.Addr, error) { return nil, errors.New("DNS down") }}
	if _, err := (safeDialer{resolver: lookupFailure, dialer: &recordingDialer{}}).DialContext(context.Background(), "tcp", "example.com:443"); err == nil || errors.Is(err, ErrDestination) {
		t.Fatalf("lookup error=%v", err)
	}
	if _, err := (safeDialer{resolver: &sequenceResolver{answers: [][]netip.Addr{{}}}, dialer: &recordingDialer{}}).DialContext(context.Background(), "tcp", "example.com:443"); !errors.Is(err, ErrDestination) {
		t.Fatalf("empty answer error=%v", err)
	}
	dial := &recordingDialer{}
	if _, err := (safeDialer{resolver: &sequenceResolver{}, dialer: dial}).DialContext(context.Background(), "tcp", "8.8.8.8:443"); err == nil {
		t.Fatal("literal dial unexpectedly succeeded")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (safeDialer{resolver: &sequenceResolver{}, dialer: &recordingDialer{succeed: true}}).DialContext(ctx, "tcp", "8.8.8.8:443"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled dial error=%v", err)
	}
}

func TestValidASCIIHostEdges(t *testing.T) {
	t.Parallel()

	for _, host := range []string{"example.com", "a-b.example"} {
		if !validASCIIHost(host) {
			t.Errorf("valid host rejected: %s", host)
		}
	}
	for _, host := range []string{"", "single", ".example", "example.", "-a.example", "a-.example", strings.Repeat("a", 64) + ".example", "EXAMPLE.com"} {
		if validASCIIHost(host) {
			t.Errorf("invalid host accepted: %q", host)
		}
	}
}

func TestInvalidAddressAndReceiptTimeoutConfig(t *testing.T) {
	t.Parallel()

	if isPublicAddress(netip.Addr{}) {
		t.Fatal("invalid address accepted")
	}
	cfg := Config{Secret: Secret{KeyID: "key", Value: make([]byte, minSecretBytes)}, Recorder: discardRecorder{}, ReceiptTimeout: 31 * time.Second}
	if _, err := validateConfig(cfg); !errors.Is(err, ErrInvalid) {
		t.Fatalf("receipt timeout error=%v", err)
	}
}

type trackedBody struct {
	io.Reader
	closed   bool
	closeErr error
}

func (b *trackedBody) Close() error { b.closed = true; return b.closeErr }

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

type resolverFunc struct {
	fn func(context.Context, string, string) ([]netip.Addr, error)
}

func (r *resolverFunc) LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error) {
	return r.fn(ctx, network, host)
}
