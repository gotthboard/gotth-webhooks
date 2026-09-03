package webhooks

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"testing"
	"time"
)

func TestSafeDialerRejectsMixedAnswerBeforeDial(t *testing.T) {
	t.Parallel()

	dial := &recordingDialer{}
	safe := safeDialer{
		resolver: &sequenceResolver{answers: [][]netip.Addr{{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("127.0.0.1")}}},
		dialer:   dial,
	}
	if _, err := safe.DialContext(context.Background(), "tcp", "example.com:443"); !errors.Is(err, ErrDestination) {
		t.Fatalf("error = %v, want ErrDestination", err)
	}
	if len(dial.addresses) != 0 {
		t.Fatalf("dialed rejected answer: %v", dial.addresses)
	}
}

func TestSafeDialerRevalidatesEachNewConnection(t *testing.T) {
	t.Parallel()

	dial := &recordingDialer{succeed: true}
	safe := safeDialer{
		resolver: &sequenceResolver{answers: [][]netip.Addr{{netip.MustParseAddr("8.8.8.8")}, {netip.MustParseAddr("10.0.0.1")}}},
		dialer:   dial,
	}
	conn, err := safe.DialContext(context.Background(), "tcp", "example.com:443")
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()
	if _, err := safe.DialContext(context.Background(), "tcp", "example.com:443"); !errors.Is(err, ErrDestination) {
		t.Fatalf("second error = %v, want ErrDestination", err)
	}
	if got, want := len(dial.addresses), 1; got != want {
		t.Fatalf("dial count = %d, want %d", got, want)
	}
	if dial.addresses[0] != "8.8.8.8:443" {
		t.Fatalf("dial address = %q", dial.addresses[0])
	}
}

func TestSafeDialerAddressCountBoundary(t *testing.T) {
	t.Parallel()

	for _, count := range []int{maxResolvedAddresses - 1, maxResolvedAddresses, maxResolvedAddresses + 1, maxResolvedAddresses * 4} {
		answers := make([]netip.Addr, count)
		for i := range answers {
			answers[i] = netip.MustParseAddr("8.8.8.8")
		}
		dial := &recordingDialer{succeed: true}
		safe := safeDialer{resolver: &sequenceResolver{answers: [][]netip.Addr{answers}}, dialer: dial}
		conn, err := safe.DialContext(context.Background(), "tcp", "example.com:443")
		if count <= maxResolvedAddresses {
			if err != nil {
				t.Errorf("count %d: %v", count, err)
			} else {
				_ = conn.Close()
			}
		} else if !errors.Is(err, ErrDestination) {
			t.Errorf("count %d error = %v, want ErrDestination", count, err)
		}
	}
}

func TestSafeDialerRejectsUnsupportedNetworkAndPort(t *testing.T) {
	t.Parallel()

	safe := safeDialer{resolver: &sequenceResolver{}, dialer: &recordingDialer{}}
	for _, tc := range []struct{ network, address string }{{"udp", "example.com:443"}, {"tcp", "example.com:80"}, {"tcp", "bad"}} {
		if _, err := safe.DialContext(context.Background(), tc.network, tc.address); !errors.Is(err, ErrDestination) {
			t.Errorf("%s %s: %v", tc.network, tc.address, err)
		}
	}
}

func TestSafeDialerExhaustionUsesTypedTransientAllowlist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		dialError error
		retryable bool
	}{
		{name: "access denied", dialError: syscall.EACCES},
		{name: "file descriptor exhaustion", dialError: syscall.EMFILE},
		{name: "invalid argument", dialError: syscall.EINVAL},
		{name: "plain unknown", dialError: errors.New("unclassified dial failure")},
		{name: "timeout", dialError: classifiedNetError{timeout: true}, retryable: true},
		{name: "temporary", dialError: classifiedNetError{temporary: true}, retryable: true},
		{name: "connection refused", dialError: syscall.ECONNREFUSED, retryable: true},
		{name: "connection reset", dialError: syscall.ECONNRESET, retryable: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dial := &scriptedErrorDialer{errors: []error{tc.dialError}}
			safe := safeDialer{resolver: &sequenceResolver{}, dialer: dial}
			_, err := safe.DialContext(context.Background(), "tcp", "8.8.8.8:443")
			if err == nil {
				t.Fatal("dial unexpectedly succeeded")
			}
			if got := isRetryableTransportFailure(err); got != tc.retryable {
				t.Fatalf("retryable=%v want=%v error=%v", got, tc.retryable, err)
			}
			if !tc.retryable && err != errDialFailure {
				t.Fatalf("permanent error was not redacted: %v", err)
			}
		})
	}
}

func TestSafeDialerMixedExhaustionPermanentDominates(t *testing.T) {
	t.Parallel()

	answers := [][]netip.Addr{{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("1.1.1.1")}}
	for _, dialErrors := range [][]error{
		{syscall.ECONNREFUSED, syscall.EACCES},
		{errors.New("unknown"), syscall.ECONNRESET},
	} {
		dial := &scriptedErrorDialer{errors: dialErrors}
		safe := safeDialer{resolver: &sequenceResolver{answers: answers}, dialer: dial}
		_, err := safe.DialContext(context.Background(), "tcp", "example.com:443")
		if err != errDialFailure || isRetryableTransportFailure(err) {
			t.Fatalf("mixed errors=%v produced retryable/unredacted error %v", dialErrors, err)
		}
		if len(dial.addresses) != 2 {
			t.Fatalf("dial count=%d", len(dial.addresses))
		}
	}

	allTransient := &scriptedErrorDialer{errors: []error{syscall.ECONNREFUSED, classifiedNetError{temporary: true}}}
	safe := safeDialer{resolver: &sequenceResolver{answers: answers}, dialer: allTransient}
	_, err := safe.DialContext(context.Background(), "tcp", "example.com:443")
	if err == nil || !isRetryableTransportFailure(err) {
		t.Fatalf("all-transient aggregate=%v", err)
	}
}

func TestSafeDialerCancellationDominatesAggregate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	dial := &scriptedErrorDialer{errors: []error{syscall.ECONNREFUSED}, beforeReturn: cancel}
	safe := safeDialer{resolver: &sequenceResolver{}, dialer: dial}
	if _, err := safe.DialContext(ctx, "tcp", "8.8.8.8:443"); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}

func TestSafeDialerMarksTransientDNSLookupForRetry(t *testing.T) {
	t.Parallel()

	dial := &recordingDialer{}
	safe := safeDialer{
		resolver: &resolverFunc{fn: func(context.Context, string, string) ([]netip.Addr, error) {
			return nil, classifiedNetError{temporary: true}
		}},
		dialer: dial,
	}
	_, err := safe.DialContext(context.Background(), "tcp", "example.com:443")
	if err == nil || !isRetryableTransportFailure(err) {
		t.Fatalf("transient DNS lookup error=%v, want retryable classification", err)
	}
	if len(dial.addresses) != 0 {
		t.Fatalf("dialed after failed DNS lookup: %v", dial.addresses)
	}
}

func TestSafeDialerPermanentFailureStopsDeliveryAfterOneAttempt(t *testing.T) {
	t.Parallel()

	dial := &scriptedErrorDialer{errors: []error{syscall.EACCES}}
	safe := safeDialer{
		resolver: &sequenceResolver{answers: [][]netip.Addr{{netip.MustParseAddr("8.8.8.8")}}},
		dialer:   dial,
	}
	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		_, err := safe.DialContext(req.Context(), "tcp", req.URL.Host)
		return nil, err
	})
	recorder := &memoryRecorder{}
	dispatcher := testDispatcher(t, recorder, transport, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := dispatcher.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrPermanent) || result.Attempts != 1 || result.Outcome != OutcomePermanent || len(dial.addresses) != 1 {
		t.Fatalf("result=%+v error=%v dials=%v", result, err, dial.addresses)
	}
	if err.Error() != ErrPermanent.Error() {
		t.Fatalf("public error leaked dial detail: %v", err)
	}
	receipts := recorder.snapshot()
	if len(receipts) != 1 || receipts[0].ErrorCode != ErrorTransport {
		t.Fatalf("receipts=%+v", receipts)
	}
}

type sequenceResolver struct {
	answers [][]netip.Addr
	calls   int
}

func (r *sequenceResolver) LookupNetIP(_ context.Context, _, _ string) ([]netip.Addr, error) {
	if r.calls >= len(r.answers) {
		return nil, errors.New("no answer")
	}
	answer := r.answers[r.calls]
	r.calls++
	return answer, nil
}

type recordingDialer struct {
	addresses []string
	succeed   bool
}

type scriptedErrorDialer struct {
	errors       []error
	addresses    []string
	calls        int
	beforeReturn func()
}

func (d *scriptedErrorDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.addresses = append(d.addresses, address)
	err := d.errors[d.calls]
	d.calls++
	if d.beforeReturn != nil {
		d.beforeReturn()
	}
	return nil, err
}

type classifiedNetError struct {
	timeout   bool
	temporary bool
}

func (e classifiedNetError) Error() string   { return "classified network failure" }
func (e classifiedNetError) Timeout() bool   { return e.timeout }
func (e classifiedNetError) Temporary() bool { return e.temporary }

func (d *recordingDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.addresses = append(d.addresses, address)
	if !d.succeed {
		return nil, errors.New("dial failed")
	}
	a, b := net.Pipe()
	_ = b.Close()
	return a, nil
}
