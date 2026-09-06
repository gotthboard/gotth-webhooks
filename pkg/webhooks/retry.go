package webhooks

import (
	"context"
	"io"
	"net/http"
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
// plus one overflow byte. Local CPU time is O(c+min(n,L+1)) and local auxiliary
// space is O(1). Whole-function CPU time is
// O(c+min(n,L+1)+Bt(c,min(n,L+1))), Omega(1), with no input-independent tight
// bound; whole-function auxiliary space is O(1+Bs(c,min(n,L+1))), Omega(1),
// with no input-independent tight bound. n is response bytes, L is
// MaxResponseBytes, c is the number of body Read callbacks/local copy-loop
// iterations before return, and Bt/Bs are delegated response-body
// CPU/allocation across those Read callbacks and one Close when body is
// non-nil. Response-body I/O and latency are also delegated. A body can return
// (0, nil) repeatedly, so there is no finite local CPU bound in n alone without
// body cooperation.
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
// Complexity: worst-case time O(a+P(r)), Omega(1), with no input-independent
// tight bound because delay saturation can stop the loop immediately; auxiliary
// space O(S(r)), Omega(1), with no tight bound established by the delegated
// parser contracts. P(r) and S(r) are parseRetryAfter time and space, including
// delegated http.ParseTime costs, for r Retry-After bytes; a is the prior
// attempt number.
func retryDelay(policy RetryPolicy, attempt int, retryAfter string, now time.Time) time.Duration {
	delay := policy.InitialDelay
	for step := 1; step < attempt && delay < policy.MaxDelay; step++ {
		if delay > policy.MaxDelay/2 {
			delay = policy.MaxDelay
			break
		}
		delay *= 2
	}
	if parsed, ok := parseRetryAfter(retryAfter, now, policy.MaxDelay); ok && parsed > delay {
		delay = parsed
	}
	return delay
}

// parseRetryAfter parses RFC 9110 delay-seconds (ASCII 1*DIGIT) or HTTP-date,
// rejects values beyond the response-header bound, and saturates valid delays
// at maxDelay without integer overflow. Complexity: time O(n), Omega(1), no
// input-independent tight Theta bound; HTTP-date allocation is delegated;
// n is header bytes and auxiliary space is otherwise O(1).
func parseRetryAfter(value string, now time.Time, maxDelay time.Duration) (time.Duration, bool) {
	if len(value) > maxResponseHeaderBytes || maxDelay <= 0 {
		return 0, false
	}
	value = strings.Trim(value, " \t")
	if value == "" {
		return 0, false
	}
	digits := true
	var seconds uint64
	capSeconds := uint64(maxDelay / time.Second)
	saturated := false
	for i := range len(value) {
		c := value[i]
		if c < '0' || c > '9' {
			digits = false
			break
		}
		if saturated {
			continue
		}
		digit := uint64(c - '0')
		if seconds > capSeconds/10 || (seconds == capSeconds/10 && digit > capSeconds%10) {
			saturated = true
			continue
		}
		seconds = seconds*10 + digit
	}
	if digits {
		if saturated {
			return maxDelay, true
		}
		return time.Duration(seconds) * time.Second, true
	}
	when, err := http.ParseTime(value)
	if err != nil || !when.After(now) {
		return 0, false
	}
	delay := when.Sub(now)
	if delay > maxDelay {
		return maxDelay, true
	}
	return delay, true
}

// waitContext sleeps without losing cancellation responsiveness. Complexity:
// CPU time and auxiliary space O(1), Omega(1), tight Theta(1). The timer becomes
// eligible after at least delay d; cancellation may win earlier, while actual
// return latency has no finite upper bound in d alone because runtime scheduling
// is delegated.
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
