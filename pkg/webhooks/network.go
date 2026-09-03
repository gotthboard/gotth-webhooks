package webhooks

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/netip"
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

// DialContext resolves a hostname, rejects an entire unsafe or oversized
// answer, and passes only validated numeric addresses to the underlying
// dialer. Complexity: CPU time O(a*p), Omega(1), no input-independent tight
// Theta bound; auxiliary space O(a), Omega(1), no single tight bound; a is
// resolved address count and p is the fixed denied-prefix count; DNS and dial
// latency are delegated and externally bounded by context.
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
			return nil, fmt.Errorf("%w: DNS lookup failed: %v", ErrDestination, err)
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
	for _, candidate := range addresses {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		conn, dialErr := d.dialer.DialContext(ctx, network, net.JoinHostPort(candidate.Unmap().String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	return nil, fmt.Errorf("dial validated destination: %w", lastErr)
}

// newHTTPClient constructs the production-owned no-proxy, no-redirect HTTPS
// client. Complexity: time and auxiliary space O(1), Omega(1), tight Theta(1);
// network costs occur only during later requests.
func newHTTPClient(attemptTimeout time.Duration) *http.Client {
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
		MaxResponseHeaderBytes: 64 << 10,
		TLSClientConfig:        &tls.Config{MinVersion: tls.VersionTLS12},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   attemptTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
