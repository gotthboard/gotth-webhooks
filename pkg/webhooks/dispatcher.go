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
// hardened HTTP transport. All-input time is O(1+q+k) and auxiliary space
// O(1+k), both Omega(1), with no single tight bound because invalid
// configuration can return before secret work. The admitted path is Theta(q+k)
// time and Theta(k) auxiliary space; q and k are key-ID and secret bytes.
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
// before retry or return. All-input CPU time is
// O(Vt(e,c,b)+sum(i=1..A)(b+r_i+m+w_i+Pt(i,q_i)+Tt_i+Bt_i+Rt_i+Wt_i)), Omega(1), with no
// input-independent tight bound; auxiliary space is
// O(Vs(e,c,b)+b+m+max(i=1..A)(Ts_i+Bs_i+Rs_i+Ps(q_i)+Ws_i+d_i)), Omega(1), also with no
// input-independent tight bound. A is actual attempts (at most configured
// MaxAttempts); e, c, and b are endpoint, content-type, and body bytes; r_i is
// bounded response bytes; m is bounded request metadata; q_i is Retry-After
// bytes; w_i is total error-tree nodes visited by all errors.Is/As calls; and
// d_i is maximum joined-error depth. Vt/Vs are validation, Pt/Ps are
// retry-delay/header parsing, Tt/Ts are RoundTripper CPU/allocation, Bt/Bs are
// response-body Read/Close CPU/allocation, Rt/Rs are Recorder CPU/allocation,
// and Wt/Ws are retry-wait CPU/allocation. Sum and maximum attempt terms are
// zero when A is zero. Network, response-body, recorder, timer, DNS, TLS, and
// wait I/O/latency are delegated; contexts bound cooperative implementations,
// not a Recorder or body implementation that violates its contract.
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
			if cause != nil {
				return result, ErrPermanent
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
// All-input CPU time is O(b+r+m+w+Tt(b,r,m)+Bt(r)), Omega(1), with no
// input-independent tight bound; auxiliary space is O(m+Ts(b,r,m)+Bs(r)+d),
// Omega(1), with no input-independent tight bound. b is body bytes, r is
// bounded response bytes, m is bounded request metadata, w is total error-tree
// nodes visited by errors.Is/As calls, and d is maximum joined-error depth.
// Tt/Ts are delegated RoundTripper CPU/allocation and buffering costs; Bt/Bs
// are delegated response-body Read/Close CPU/allocation costs.
// The request body is already owned and is hashed/read without another body
// copy. Transport and response-body I/O latency are delegated and cooperatively
// bounded by attemptTimeout.
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
// transient allowlist precedence. Complexity: time O(w), Omega(1), no
// input-independent tight bound; auxiliary space O(d), Omega(1), no
// input-independent tight bound; w is total wrapped/joined error nodes visited
// by all delegated errors.Is/As traversals and d is maximum join-tree depth.
func classifyAttemptFailure(parent, attemptCtx context.Context, err error) (Outcome, ErrorCode) {
	if parent.Err() != nil {
		return OutcomeCanceled, ErrorCanceled
	}
	if errors.Is(err, ErrDestination) {
		return OutcomePermanent, ErrorDestination
	}
	if isDeterministicTransportFailure(err) {
		return OutcomePermanent, ErrorTransport
	}
	if errors.Is(err, context.DeadlineExceeded) ||
		(errors.Is(err, context.Canceled) && errors.Is(attemptCtx.Err(), context.DeadlineExceeded)) {
		return OutcomeRetryable, ErrorTimeout
	}
	if isRetryableTransportFailure(err) {
		return OutcomeRetryable, ErrorTransport
	}
	return OutcomePermanent, ErrorTransport
}

// record attempts durable receipt persistence with caller values preserved but
// cancellation detached. Complexity: CPU time O(1+Rt), Omega(1), with no
// input-independent tight bound; auxiliary space O(1+Rs), Omega(1), with no
// input-independent tight bound. Rt/Rs and recorder I/O/latency are delegated
// to Recorder; receiptTimeout cooperatively bounds conforming implementations.
func (d *Dispatcher) record(parent context.Context, receipt Receipt) error {
	recordCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), d.config.receiptTimeout)
	defer cancel()
	if err := d.config.recorder.Record(recordCtx, receipt); err != nil {
		return ErrReceipt
	}
	return nil
}
