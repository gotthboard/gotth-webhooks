package webhooks

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

const maxResolvedAddresses = 16

type resolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

type dialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

type safeDialer struct {
	resolver resolver
	dialer   dialer
}

type retryableTransportError struct{ cause error }

// Error returns a fixed redacted class. Time and auxiliary space are O(1),
// Omega(1), tight Theta(1).
func (e *retryableTransportError) Error() string { return "retryable transport failure" }

// Unwrap returns the stored cause without copying it. Time and auxiliary space
// are O(1), Omega(1), tight Theta(1).
func (e *retryableTransportError) Unwrap() error { return e.cause }

var errDialFailure = errors.New("dial failure")

// DialContext resolves a hostname, rejects an entire unsafe or oversized
// answer, and passes only validated numeric addresses to the underlying
// dialer. Complexity: CPU time O(n+a*(p+w)), Omega(1), no input-independent
// tight Theta bound; auxiliary space O(n+a+d), Omega(1), no single tight
// bound. n is dial-address bytes scanned or materialized by address parsing
// and numeric dial-target construction; a is the resolved address count; p is
// the fixed denied-prefix count; w is the maximum number of wrapped/joined
// error nodes visited by classification; and d is the maximum joined-error
// traversal depth. Resolver and dial I/O latency and resolver-result allocation
// are delegated and externally bounded by context and maxResolvedAddresses.
func (d safeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("%w: unsupported network", ErrDestination)
	}
	host, port, err := canonicalPort(address)
	if err != nil {
		return nil, err
	}

	var addresses []netip.Addr
	if literal, parseErr := netip.ParseAddr(host); parseErr == nil {
		addresses = []netip.Addr{literal}
	} else {
		addresses, err = d.resolver.LookupNetIP(ctx, "ip", host)
		if err != nil {
			if isAdmittedTransientTransportFailure(err) {
				return nil, &retryableTransportError{cause: err}
			}
			return nil, fmt.Errorf("DNS lookup failed: %w", err)
		}
	}
	if len(addresses) == 0 || len(addresses) > maxResolvedAddresses {
		return nil, fmt.Errorf("%w: DNS answer count outside 1..%d", ErrDestination, maxResolvedAddresses)
	}
	for _, candidate := range addresses {
		if !isPublicAddress(candidate) {
			return nil, fmt.Errorf("%w: DNS answer includes non-public address", ErrDestination)
		}
	}

	var lastErr error
	allTransient := true
	for _, candidate := range addresses {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		conn, dialErr := d.dialer.DialContext(ctx, network, net.JoinHostPort(candidate.Unmap().String(), port))
		if dialErr == nil {
			return conn, nil
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !isAdmittedTransientTransportFailure(dialErr) {
			allTransient = false
		}
		lastErr = dialErr
	}
	// Retry only when every attempted candidate failed with an admitted
	// transient class. Any deterministic or unknown member of an exhausted
	// mixed-address set makes the aggregate permanent and redacted.
	if allTransient {
		return nil, &retryableTransportError{cause: lastErr}
	}
	return nil, errDialFailure
}

// newHTTPTransport constructs the production-owned no-proxy HTTPS transport.
// Dispatcher invokes RoundTrip exactly once per attempt, so redirect handling
// is never entered. Complexity: time and auxiliary space O(1), Omega(1), tight Theta(1);
// network costs occur only during later requests.
func newHTTPTransport(attemptTimeout time.Duration) *http.Transport {
	baseDialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	checked := safeDialer{resolver: net.DefaultResolver, dialer: baseDialer}
	transport := &http.Transport{
		Proxy:                  nil,
		DialContext:            checked.DialContext,
		ForceAttemptHTTP2:      true,
		MaxIdleConns:           32,
		MaxIdleConnsPerHost:    8,
		MaxConnsPerHost:        16,
		IdleConnTimeout:        30 * time.Second,
		TLSHandshakeTimeout:    10 * time.Second,
		ResponseHeaderTimeout:  attemptTimeout,
		ExpectContinueTimeout:  time.Second,
		MaxResponseHeaderBytes: maxResponseHeaderBytes,
		TLSClientConfig:        &tls.Config{MinVersion: tls.VersionTLS12},
	}
	return transport
}

// isRetryableTransportFailure is an allowlist, not a catch-all. Production
// safe-dial failures, timeouts, connection loss, and truncated connections can
// be transient. Unknown errors and deterministic TLS/HTTP protocol failures
// are permanent. Complexity: time O(w), Omega(1), no input-independent tight
// bound; auxiliary space O(d), Omega(1), no input-independent tight bound; w
// is wrapped/joined error nodes visited and d is maximum join-tree depth in
// delegated errors.Is/As traversal.
func isRetryableTransportFailure(err error) bool {
	var marked *retryableTransportError
	if errors.As(err, &marked) {
		return true
	}
	return isAdmittedTransientTransportFailure(err)
}

// isAdmittedTransientTransportFailure is the shared typed allowlist for raw
// resolver, dial, and transport errors. Resource/configuration errors take
// precedence over net.Error.Temporary because the latter is deprecated and,
// for example, reports EMFILE as temporary. Complexity: time O(w), Omega(1),
// no input-independent tight bound; auxiliary space O(d), Omega(1), no
// input-independent tight bound; w is wrapped/joined error nodes visited and d
// is maximum join-tree depth in delegated errors.Is/As traversal.
func isAdmittedTransientTransportFailure(err error) bool {
	for _, permanent := range []error{
		syscall.EACCES,
		syscall.EPERM,
		syscall.EMFILE,
		syscall.ENFILE,
		syscall.EINVAL,
	} {
		if errors.Is(err, permanent) {
			return false
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, net.ErrClosed) {
		return true
	}
	for _, transient := range []error{
		syscall.ECONNABORTED,
		syscall.ECONNREFUSED,
		syscall.ECONNRESET,
		syscall.EHOSTDOWN,
		syscall.EHOSTUNREACH,
		syscall.ENETDOWN,
		syscall.ENETUNREACH,
		syscall.EPIPE,
		syscall.ETIMEDOUT,
	} {
		if errors.Is(err, transient) {
			return true
		}
	}
	return false
}

// isDeterministicTransportFailure identifies exported TLS and HTTP protocol
// failures that cannot improve merely by repeating the same request. Unknown
// errors are also permanent, but keeping these cases explicit prevents a
// future retry allowlist from swallowing them. Complexity: time O(w), Omega(1),
// no input-independent tight bound; auxiliary space O(d), Omega(1), no
// input-independent tight bound; w is wrapped/joined error nodes visited and d
// is maximum join-tree depth in delegated errors.As/Is traversal.
func isDeterministicTransportFailure(err error) bool {
	var certificateError *tls.CertificateVerificationError
	var recordError tls.RecordHeaderError
	var alertError tls.AlertError
	var protocolError *http.ProtocolError
	return errors.As(err, &certificateError) ||
		errors.As(err, &recordError) ||
		errors.As(err, &alertError) ||
		errors.As(err, &protocolError) ||
		errors.Is(err, http.ErrLineTooLong)
}
