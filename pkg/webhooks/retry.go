package webhooks

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// classifyStatus maps the complete HTTP status space to the V1 retry contract.
// Complexity: time and auxiliary space O(1), Omega(1), tight Theta(1).
func classifyStatus(status int) Outcome {
	if status >= 200 && status <= 299 {
		return OutcomeDelivered
	}
	if status == http.StatusRequestTimeout || status == http.StatusTooEarly || status == http.StatusTooManyRequests || (status >= 500 && status <= 599) {
		return OutcomeRetryable
	}
	return OutcomePermanent
}

// consumeResponse closes a response after reading at most the configured limit
// plus one overflow byte. Complexity: time O(min(n,L+1)), Omega(1), no
// input-independent tight Theta bound; auxiliary space O(1), Omega(1), tight
// Theta(1); n is response bytes and L is MaxResponseBytes; read/close I/O costs
// are delegated to the response body.
func consumeResponse(body io.ReadCloser) (int64, error) {
	if body == nil {
		return 0, nil
	}
	n, readErr := io.Copy(io.Discard, io.LimitReader(body, MaxResponseBytes+1))
	closeErr := body.Close()
	if n > MaxResponseBytes {
		return n, ErrResponseTooLarge
	}
	if readErr != nil {
		return n, readErr
	}
	return n, closeErr
}

// retryDelay calculates saturating exponential delay and honors a valid
// Retry-After only when it increases the delay, always capped by MaxDelay.
// Complexity: time O(a+r), Omega(a), tight Theta(a+r) for valid date parsing;
// auxiliary space inherits http.ParseTime and is O(r), Omega(1), with no tight
// bound established by its public contract; a is prior attempt number and r is
// Retry-After bytes.
func retryDelay(policy RetryPolicy, attempt int, retryAfter string, now time.Time) time.Duration {
	delay := policy.InitialDelay
	for step := 1; step < attempt && delay < policy.MaxDelay; step++ {
		if delay > policy.MaxDelay/2 {
			delay = policy.MaxDelay
			break
		}
		delay *= 2
	}
	if parsed, ok := parseRetryAfter(retryAfter, now); ok && parsed > delay {
		delay = parsed
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
	}
	return delay
}

// parseRetryAfter parses RFC 9110 delay-seconds or HTTP-date and rejects
// negative/past values. Complexity: time and auxiliary space inherit integer
// or HTTP-date parsing over n bytes; O(n), Omega(1), tight Theta not established
// across both forms; n is header bytes.
func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		if seconds < 0 {
			return 0, false
		}
		if seconds > int64(time.Minute/time.Second) {
			return time.Minute, true
		}
		return time.Duration(seconds) * time.Second, true
	}
	when, err := http.ParseTime(value)
	if err != nil || !when.After(now) {
		return 0, false
	}
	return when.Sub(now), true
}

// waitContext sleeps without losing cancellation responsiveness. Complexity:
// CPU time and auxiliary space O(1), Omega(1), tight Theta(1); wall time is at
// most delay d and is delegated to the runtime timer/context scheduler.
func waitContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
