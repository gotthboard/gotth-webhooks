package webhooks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type dependencies struct {
	transport http.RoundTripper
	now       func() time.Time
	wait      func(context.Context, time.Duration) error
}

// Dispatcher is an immutable outbound sender that is safe for concurrent use
// when its Recorder fulfills the interface's concurrency contract. Calls
// sharing a delivery ID still require consumer-owned durable coordination.
type Dispatcher struct {
	config    validatedConfig
	transport http.RoundTripper
	now       func() time.Time
	wait      func(context.Context, time.Duration) error
}

// New validates configuration, copies the secret, and constructs an owned
// hardened HTTP transport. Complexity: time O(k), Omega(k), tight Theta(k);
// auxiliary space O(k), Omega(k), tight Theta(k); k is secret bytes.
func New(cfg Config) (*Dispatcher, error) {
	validated, err := validateConfig(cfg)
	if err != nil {
		return nil, err
	}
	return newDispatcher(validated, dependencies{
		transport: newHTTPTransport(validated.attemptTimeout),
		now:       time.Now,
		wait:      waitContext,
	}), nil
}

// newDispatcher binds already validated configuration to production or
// package-internal test dependencies. Complexity: time and auxiliary space
// O(1), Omega(1), tight Theta(1).
func newDispatcher(config validatedConfig, deps dependencies) *Dispatcher {
	return &Dispatcher{config: config, transport: deps.transport, now: deps.now, wait: deps.wait}
}

// Deliver performs bounded signed attempts and records each actual attempt
// before retry or return. Complexity: CPU time O(A*(b+r)), Omega(b+r), no
// input-independent tight Theta bound; auxiliary space O(b+r), Omega(b), no
// single tight bound; A is configured attempts, b is body bytes, and r is at
// most MaxResponseBytes+1; network, recorder, timer, DNS, and TLS costs are
// delegated and bounded by configured contexts and transport limits.
func (d *Dispatcher) Deliver(ctx context.Context, msg Message) (Result, error) {
	if ctx == nil {
		return Result{}, fmt.Errorf("%w: nil context", ErrInvalid)
	}
	validated, err := validateMessage(msg)
	if err != nil {
		return Result{}, err
	}
	result := Result{DeliveryID: validated.deliveryID}
	if validated.firstAttempt > 1_000_000_000-d.config.retry.MaxAttempts+1 {
		return Result{}, fmt.Errorf("%w: attempt sequence exceeds 1000000000", ErrInvalid)
	}
	for offset := 0; offset < d.config.retry.MaxAttempts; offset++ {
		attempt := validated.firstAttempt + offset
		if err := ctx.Err(); err != nil {
			result.Outcome = OutcomeCanceled
			return result, err
		}
		receipt, retryAfter, cause := d.attempt(ctx, validated, attempt)
		result = Result{
			DeliveryID:  validated.deliveryID,
			Attempts:    offset + 1,
			LastAttempt: attempt,
			Delivered:   receipt.Outcome == OutcomeDelivered,
			StatusCode:  receipt.StatusCode,
			Outcome:     receipt.Outcome,
			LastReceipt: receipt,
		}
		if err := d.record(ctx, receipt); err != nil {
			return result, err
		}
		switch receipt.Outcome {
		case OutcomeDelivered:
			return result, nil
		case OutcomePermanent:
			if errors.Is(cause, ErrResponseTooLarge) {
				return result, fmt.Errorf("%w: %w", ErrPermanent, ErrResponseTooLarge)
			}
			if errors.Is(cause, ErrDestination) {
				return result, fmt.Errorf("%w: %w", ErrPermanent, ErrDestination)
			}
			return result, fmt.Errorf("%w: status %d", ErrPermanent, receipt.StatusCode)
		case OutcomeCanceled:
			if err := ctx.Err(); err != nil {
				return result, err
			}
			return result, context.Canceled
		case OutcomeRetryable:
			if offset+1 == d.config.retry.MaxAttempts {
				return result, ErrExhausted
			}
		}
		delay := retryDelay(d.config.retry, offset+1, retryAfter, d.now())
		if err := d.wait(ctx, delay); err != nil {
			return result, err
		}
	}
	panic("unreachable bounded attempt loop")
}

// attempt executes one request and returns only bounded classification data.
// Complexity: CPU time O(b+r), Omega(b), no input-independent tight Theta
// bound. Auxiliary space is O(m+r+T(b)), Omega(1), with no single tight bound:
// m is request metadata, r is the constant-bounded response copy buffer, and
// T(b) is any request buffering delegated to RoundTripper.
// The request body is already owned and is hashed/read without another body
// copy. Transport latency is delegated and bounded by attemptTimeout.
func (d *Dispatcher) attempt(parent context.Context, msg validatedMessage, attempt int) (Receipt, string, error) {
	started := d.now().UTC()
	receipt := Receipt{DeliveryID: msg.deliveryID, DeliveryFingerprint: msg.fingerprint, Attempt: attempt, RequestTimestamp: started.Unix(), StartedAt: started}
	attemptCtx, cancel := context.WithTimeout(parent, d.config.attemptTimeout)
	defer cancel()
	req, _, err := buildRequest(attemptCtx, msg, d.config.secret, attempt, receipt.RequestTimestamp)
	if err != nil {
		receipt.FinishedAt = d.now().UTC()
		receipt.Outcome = OutcomePermanent
		receipt.ErrorCode = ErrorTransport
		return receipt, "", err
	}
	resp, err := d.transport.RoundTrip(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		receipt.FinishedAt = d.now().UTC()
		receipt.Outcome, receipt.ErrorCode = classifyAttemptFailure(parent, attemptCtx, err)
		return receipt, "", err
	}
	receipt.StatusCode = resp.StatusCode
	retryAfter := resp.Header.Get("Retry-After")
	receipt.ResponseBytes, err = consumeResponse(resp.Body)
	receipt.FinishedAt = d.now().UTC()
	if errors.Is(err, ErrResponseTooLarge) {
		receipt.Outcome = OutcomePermanent
		receipt.ErrorCode = ErrorResponseLimit
		return receipt, retryAfter, err
	}
	if err != nil {
		receipt.Outcome, receipt.ErrorCode = classifyAttemptFailure(parent, attemptCtx, err)
		return receipt, retryAfter, err
	}
	receipt.Outcome = classifyStatus(resp.StatusCode)
	if receipt.Outcome != OutcomeDelivered {
		receipt.ErrorCode = ErrorHTTPStatus
	}
	return receipt, retryAfter, nil
}

// classifyAttemptFailure applies cancellation, deterministic-failure, and
// transient allowlist precedence. Complexity: time O(w), Omega(1), no tight
// bound; auxiliary space O(1), Omega(1), tight Theta(1); w is wrapped-error
// depth delegated to errors.Is/As.
func classifyAttemptFailure(parent, attemptCtx context.Context, err error) (Outcome, ErrorCode) {
	if parent.Err() != nil {
		return OutcomeCanceled, ErrorCanceled
	}
	if errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
		return OutcomeRetryable, ErrorTimeout
	}
	if errors.Is(err, ErrDestination) {
		return OutcomePermanent, ErrorDestination
	}
	if isDeterministicTransportFailure(err) {
		return OutcomePermanent, ErrorTransport
	}
	if isRetryableTransportFailure(err) {
		return OutcomeRetryable, ErrorTransport
	}
	return OutcomePermanent, ErrorTransport
}

// record attempts durable receipt persistence with caller values preserved but
// cancellation detached. Complexity: CPU time and auxiliary space O(1),
// Omega(1), tight Theta(1); recorder I/O is delegated and bounded by
// receiptTimeout.
func (d *Dispatcher) record(parent context.Context, receipt Receipt) error {
	recordCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), d.config.receiptTimeout)
	defer cancel()
	if err := d.config.recorder.Record(recordCtx, receipt); err != nil {
		return ErrReceipt
	}
	return nil
}
