package webhooks

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
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

func (d *recordingDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	d.addresses = append(d.addresses, address)
	if !d.succeed {
		return nil, errors.New("dial failed")
	}
	a, b := net.Pipe()
	_ = b.Close()
	return a, nil
}
