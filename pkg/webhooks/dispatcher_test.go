package webhooks

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestDeliverRetriesRecordsAndSucceeds(t *testing.T) {
	t.Parallel()

	roundTrip := &scriptedTransport{steps: []transportStep{
		{status: http.StatusInternalServerError, body: "first"},
		{status: http.StatusTooManyRequests, retryAfter: "2", body: "second"},
		{status: http.StatusNoContent},
	}}
	recorder := &memoryRecorder{}
	var waits []time.Duration
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 3, InitialDelay: time.Second, MaxDelay: 1500 * time.Millisecond}, func(_ context.Context, delay time.Duration) error {
		waits = append(waits, delay)
		return nil
	})
	result, err := d.Deliver(context.Background(), validMessage())
	if err != nil {
		t.Fatal(err)
	}
	if !result.Delivered || result.Attempts != 3 || result.LastAttempt != 3 || result.LastReceipt.Attempt != 3 || result.StatusCode != http.StatusNoContent || result.Outcome != OutcomeDelivered {
		t.Fatalf("result = %+v", result)
	}
	if got, want := waits, []time.Duration{time.Second, 1500 * time.Millisecond}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("waits = %v, want %v", got, want)
	}
	receipts := recorder.snapshot()
	if len(receipts) != 3 || receipts[0].Outcome != OutcomeRetryable || receipts[1].Outcome != OutcomeRetryable || receipts[2].Outcome != OutcomeDelivered {
		t.Fatalf("receipts = %+v", receipts)
	}
	for i, receipt := range receipts {
		if receipt.Attempt != i+1 || receipt.DeliveryID != "delivery-1" || receipt.DeliveryFingerprint == ([32]byte{}) || receipt.DeliveryFingerprint != receipts[0].DeliveryFingerprint || receipt.ResponseBytes < 0 || receipt.StartedAt.IsZero() || receipt.FinishedAt.Before(receipt.StartedAt) {
			t.Errorf("receipt %d = %+v", i, receipt)
		}
	}
}

func TestDispatcherCloseRejectsNewCallsAndLetsAdmittedCallFinish(t *testing.T) {
	t.Parallel()

	transport := newLifecycleTransport()
	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, transport, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)

	deliveryDone := make(chan struct{})
	var deliveryResult Result
	var deliveryErr error
	go func() {
		deliveryResult, deliveryErr = d.Deliver(context.Background(), validMessage())
		close(deliveryDone)
	}()
	receiveBefore(t, transport.roundTripEntered, "delivery admission")

	firstCloseDone := make(chan struct{})
	go func() {
		d.Close()
		close(firstCloseDone)
	}()
	receiveBefore(t, transport.closeEntered, "transport cleanup")

	result, err := d.Deliver(nil, Message{})
	if !errors.Is(err, ErrClosed) || result != (Result{}) {
		t.Fatalf("closed Deliver result=%+v error=%v, want zero result and ErrClosed", result, err)
	}
	if got := transport.roundTripCount(); got != 1 {
		t.Fatalf("round trips after closed Deliver=%d, want 1", got)
	}
	if got := len(recorder.snapshot()); got != 0 {
		t.Fatalf("receipts before admitted delivery release=%d, want 0", got)
	}

	secondCloseDone := make(chan struct{})
	go func() {
		d.Close()
		close(secondCloseDone)
	}()
	select {
	case <-secondCloseDone:
		t.Fatal("concurrent Close returned before owned transport cleanup completed")
	case <-time.After(20 * time.Millisecond):
	}

	close(transport.releaseClose)
	receiveBefore(t, firstCloseDone, "first Close completion")
	receiveBefore(t, secondCloseDone, "concurrent Close completion")
	if got := transport.closeCount(); got != 1 {
		t.Fatalf("transport cleanup calls=%d, want 1", got)
	}

	select {
	case <-deliveryDone:
		t.Fatal("Close interrupted an admitted delivery")
	default:
	}
	close(transport.releaseRoundTrip)
	receiveBefore(t, deliveryDone, "admitted delivery completion")
	if deliveryErr != nil || !deliveryResult.Delivered || deliveryResult.Attempts != 1 {
		t.Fatalf("admitted delivery result=%+v error=%v", deliveryResult, deliveryErr)
	}
	if got := len(recorder.snapshot()); got != 1 {
		t.Fatalf("admitted delivery receipts=%d, want 1", got)
	}

	d.Close()
	if got := transport.closeCount(); got != 1 {
		t.Fatalf("transport cleanup calls after repeated Close=%d, want 1", got)
	}
}

func TestDispatcherCloseWithoutClosableTestTransport(t *testing.T) {
	t.Parallel()

	transport := &scriptedTransport{steps: []transportStep{{status: http.StatusNoContent}}}
	d := testDispatcher(t, &memoryRecorder{}, transport, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	d.Close()

	if result, err := d.Deliver(context.Background(), validMessage()); !errors.Is(err, ErrClosed) || result != (Result{}) {
		t.Fatalf("closed Deliver result=%+v error=%v, want zero result and ErrClosed", result, err)
	}
	if got := transport.callCount(); got != 0 {
		t.Fatalf("round trips=%d, want 0", got)
	}
}

func TestDeliverUsesConsumerDurableFirstAttempt(t *testing.T) {
	t.Parallel()

	recorder := &memoryRecorder{}
	transport := &scriptedTransport{steps: []transportStep{{status: http.StatusServiceUnavailable}, {status: http.StatusNoContent}}}
	d := testDispatcher(t, recorder, transport, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	msg := validMessage()
	msg.FirstAttempt = 41
	result, err := d.Deliver(context.Background(), msg)
	if err != nil || result.Attempts != 2 || result.LastAttempt != 42 {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	receipts := recorder.snapshot()
	if len(receipts) != 2 || receipts[0].Attempt != 41 || receipts[1].Attempt != 42 {
		t.Fatalf("receipts=%+v", receipts)
	}
}

func TestClassifyEveryStatusBoundary(t *testing.T) {
	t.Parallel()

	tests := map[int]Outcome{
		199: OutcomePermanent,
		200: OutcomeDelivered,
		299: OutcomeDelivered,
		300: OutcomePermanent,
		408: OutcomeRetryable,
		425: OutcomeRetryable,
		429: OutcomeRetryable,
		499: OutcomePermanent,
		500: OutcomeRetryable,
		599: OutcomeRetryable,
		600: OutcomePermanent,
		700: OutcomePermanent,
	}
	for status, want := range tests {
		if got := classifyStatus(status); got != want {
			t.Errorf("status %d = %s, want %s", status, got, want)
		}
	}
}

func TestDeliverPermanentAndRedirectResponsesDoNotRetry(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusBadRequest, http.StatusFound} {
		recorder := &memoryRecorder{}
		roundTrip := &scriptedTransport{steps: []transportStep{{status: status}}}
		d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
		result, err := d.Deliver(context.Background(), validMessage())
		if !errors.Is(err, ErrPermanent) {
			t.Fatalf("status %d error = %v", status, err)
		}
		if result.Attempts != 1 || result.Outcome != OutcomePermanent || roundTrip.callCount() != 1 || len(recorder.snapshot()) != 1 {
			t.Fatalf("status %d result=%+v calls=%d receipts=%d", status, result, roundTrip.callCount(), len(recorder.snapshot()))
		}
	}
}

func TestDeliverExhaustsAllowlistedTransportFailures(t *testing.T) {
	t.Parallel()

	roundTrip := &scriptedTransport{steps: []transportStep{{err: &retryableTransportError{cause: errors.New("broken")}}, {err: &retryableTransportError{cause: errors.New("still broken")}}}}
	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrExhausted) || result.Attempts != 2 || result.Outcome != OutcomeRetryable {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	for _, receipt := range recorder.snapshot() {
		if receipt.ErrorCode != ErrorTransport || receipt.StatusCode != 0 {
			t.Fatalf("receipt = %+v", receipt)
		}
	}
}

func TestDeliverDeterministicTransportFailuresArePermanent(t *testing.T) {
	t.Parallel()

	headerLimitError := captureHTTPTransportError(t, "HTTP/1.1 200 OK\r\nX-Large: "+strings.Repeat("x", 512)+"\r\n\r\n", 64)
	if !strings.Contains(headerLimitError.Error(), "server response headers exceeded") {
		t.Fatalf("header-limit transport error=%v", headerLimitError)
	}
	malformedResponseError := captureHTTPTransportError(t, "NOT-HTTP\r\n\r\n", maxResponseHeaderBytes)
	if !strings.Contains(malformedResponseError.Error(), "malformed HTTP response") {
		t.Fatalf("malformed-response transport error=%v", malformedResponseError)
	}
	tests := []struct {
		name string
		err  error
	}{
		{"unknown", errors.New("unknown transport failure")},
		{"certificate", &tls.CertificateVerificationError{Err: x509.UnknownAuthorityError{}}},
		{"TLS record", tls.RecordHeaderError{Msg: "not TLS"}},
		{"real malformed response", malformedResponseError},
		{"real response header limit", headerLimitError},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			transport := &scriptedTransport{steps: []transportStep{{err: tc.err}, {status: http.StatusNoContent}}}
			recorder := &memoryRecorder{}
			d := testDispatcher(t, recorder, transport, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
			result, err := d.Deliver(context.Background(), validMessage())
			if !errors.Is(err, ErrPermanent) || strings.Contains(err.Error(), "status 0") || result.Attempts != 1 || result.Outcome != OutcomePermanent || transport.callCount() != 1 {
				t.Fatalf("result=%+v error=%v calls=%d", result, err, transport.callCount())
			}
			if receipts := recorder.snapshot(); len(receipts) != 1 || receipts[0].ErrorCode != ErrorTransport {
				t.Fatalf("receipts=%+v", receipts)
			}
		})
	}
}

func TestClassifyAttemptFailurePrecedenceAtAttemptDeadline(t *testing.T) {
	t.Parallel()

	attemptCtx, cancelAttempt := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancelAttempt()
	headerLimitError := captureHTTPTransportError(t, "HTTP/1.1 200 OK\r\nX-Large: "+strings.Repeat("x", 512)+"\r\n\r\n", 64)
	malformedResponseError := captureHTTPTransportError(t, "NOT-HTTP\r\n\r\n", maxResponseHeaderBytes)
	tests := []struct {
		name    string
		err     error
		outcome Outcome
		code    ErrorCode
	}{
		{name: "destination", err: fmt.Errorf("wrapped: %w", ErrDestination), outcome: OutcomePermanent, code: ErrorDestination},
		{name: "certificate", err: &tls.CertificateVerificationError{Err: x509.UnknownAuthorityError{}}, outcome: OutcomePermanent, code: ErrorTransport},
		{name: "TLS record", err: tls.RecordHeaderError{Msg: "not TLS"}, outcome: OutcomePermanent, code: ErrorTransport},
		{name: "TLS alert", err: tls.AlertError(40), outcome: OutcomePermanent, code: ErrorTransport},
		{name: "real malformed response", err: malformedResponseError, outcome: OutcomePermanent, code: ErrorTransport},
		{name: "real response header limit", err: headerLimitError, outcome: OutcomePermanent, code: ErrorTransport},
		{name: "actual attempt deadline", err: fmt.Errorf("transport: %w", context.DeadlineExceeded), outcome: OutcomeRetryable, code: ErrorTimeout},
		{name: "transient without causal deadline", err: syscall.ETIMEDOUT, outcome: OutcomeRetryable, code: ErrorTransport},
		{name: "unknown without causal deadline", err: errors.New("unknown transport failure"), outcome: OutcomePermanent, code: ErrorTransport},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			outcome, code := classifyAttemptFailure(context.Background(), attemptCtx, tc.err)
			if outcome != tc.outcome || code != tc.code {
				t.Fatalf("classifyAttemptFailure() = (%s, %s), want (%s, %s)", outcome, code, tc.outcome, tc.code)
			}
		})
	}
}

func TestClassifyAttemptFailureCallerCancellationDominatesDeadlineAndPermanentError(t *testing.T) {
	t.Parallel()

	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()
	attemptCtx, cancelAttempt := context.WithDeadline(parent, time.Unix(0, 0))
	defer cancelAttempt()
	malformedResponseError := captureHTTPTransportError(t, "NOT-HTTP\r\n\r\n", maxResponseHeaderBytes)
	for _, err := range []error{
		fmt.Errorf("wrapped: %w", ErrDestination),
		&tls.CertificateVerificationError{Err: x509.UnknownAuthorityError{}},
		malformedResponseError,
		fmt.Errorf("transport: %w", context.DeadlineExceeded),
		syscall.ETIMEDOUT,
	} {
		outcome, code := classifyAttemptFailure(parent, attemptCtx, err)
		if outcome != OutcomeCanceled || code != ErrorCanceled {
			t.Fatalf("error %T classified as (%s, %s), want (%s, %s)", err, outcome, code, OutcomeCanceled, ErrorCanceled)
		}
	}
}

func TestDeliverExhaustsRetryableHTTPWithoutTransportCause(t *testing.T) {
	t.Parallel()

	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, &scriptedTransport{steps: []transportStep{{status: http.StatusServiceUnavailable}}}, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrExhausted) || result.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

func TestDeliverTruncatedResponseIsRetryable(t *testing.T) {
	t.Parallel()

	want := io.ErrUnexpectedEOF
	roundTrip := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: &trackedBody{Reader: errorReader{err: want}}}, nil
	})
	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrExhausted) || strings.Contains(err.Error(), want.Error()) || result.Outcome != OutcomeRetryable {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

func TestDeliverResponseProtocolFailureIsPermanent(t *testing.T) {
	t.Parallel()

	want := &http.ProtocolError{ErrorString: "invalid chunk framing"}
	roundTrip := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: &trackedBody{Reader: errorReader{err: want}}}, nil
	})
	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrPermanent) || strings.Contains(err.Error(), "status 200") || result.Attempts != 1 || result.Outcome != OutcomePermanent {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

func TestDeliverDestinationRejectionIsPermanentAndSanitized(t *testing.T) {
	t.Parallel()

	roundTrip := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("wrapped: %w", ErrDestination)
	})
	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	msg := validMessage()
	msg.Endpoint = "https://example.com/hook?private-token=do-not-leak"
	result, err := d.Deliver(context.Background(), msg)
	if !errors.Is(err, ErrPermanent) || !errors.Is(err, ErrDestination) || result.Attempts != 1 || strings.Contains(err.Error(), "do-not-leak") {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	receipt := recorder.snapshot()[0]
	if receipt.ErrorCode != ErrorDestination {
		t.Fatalf("receipt=%+v", receipt)
	}
}

func TestDeliverCancellationWhileReadingResponse(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	roundTrip := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: &trackedBody{Reader: cancelingReader{cancel: cancel, ctx: req.Context()}}}, nil
	})
	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(ctx, validMessage())
	if !errors.Is(err, context.Canceled) || result.Outcome != OutcomeCanceled {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	receipt := recorder.snapshot()[0]
	if receipt.ErrorCode != ErrorCanceled {
		t.Fatalf("receipt=%+v", receipt)
	}
}

func TestDeliverTimeoutWhileReadingResponse(t *testing.T) {
	t.Parallel()

	roundTrip := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: &trackedBody{Reader: contextReader{ctx: req.Context()}}}, nil
	})
	recorder := &memoryRecorder{}
	d := testDispatcherWithTimeout(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait, 100*time.Millisecond)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrExhausted) || result.Outcome != OutcomeRetryable {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	receipt := recorder.snapshot()[0]
	if receipt.ErrorCode != ErrorTimeout {
		t.Fatalf("receipt=%+v", receipt)
	}
}

func TestDeliverAttemptTimeoutIsRecordedAndBounded(t *testing.T) {
	t.Parallel()

	roundTrip := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})
	recorder := &memoryRecorder{}
	d := testDispatcherWithTimeout(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait, 100*time.Millisecond)
	started := time.Now()
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrExhausted) || time.Since(started) > time.Second {
		t.Fatalf("result=%+v error=%v elapsed=%s", result, err, time.Since(started))
	}
	receipts := recorder.snapshot()
	if len(receipts) != 1 || receipts[0].ErrorCode != ErrorTimeout || receipts[0].Outcome != OutcomeRetryable {
		t.Fatalf("receipts = %+v", receipts)
	}
}

func TestDeliverCallerCancellationRecordsWithDetachedContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	roundTrip := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		cancel()
		<-req.Context().Done()
		return nil, req.Context().Err()
	})
	recorder := &memoryRecorder{requireLiveContext: true}
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(ctx, validMessage())
	if !errors.Is(err, context.Canceled) || result.Attempts != 1 || result.Outcome != OutcomeCanceled {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	if receipts := recorder.snapshot(); len(receipts) != 1 || receipts[0].ErrorCode != ErrorCanceled {
		t.Fatalf("receipts = %+v", receipts)
	}
}

func TestDeliverResponseLimitIsPermanent(t *testing.T) {
	t.Parallel()

	roundTrip := &scriptedTransport{steps: []transportStep{{status: http.StatusOK, body: strings.Repeat("x", MaxResponseBytes+1)}}}
	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrPermanent) || !errors.Is(err, ErrResponseTooLarge) || result.Attempts != 1 {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	receipt := recorder.snapshot()[0]
	if receipt.ErrorCode != ErrorResponseLimit || receipt.ResponseBytes != MaxResponseBytes+1 || receipt.Outcome != OutcomePermanent {
		t.Fatalf("receipt = %+v", receipt)
	}
}

func TestDeliverReceiptFailureStopsRetries(t *testing.T) {
	t.Parallel()

	recorder := &memoryRecorder{err: errors.New("store unavailable secret-marker")}
	roundTrip := &scriptedTransport{steps: []transportStep{{status: http.StatusInternalServerError}, {status: http.StatusNoContent}}}
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrReceipt) || strings.Contains(err.Error(), "secret-marker") || result.Attempts != 1 || result.LastReceipt.DeliveryID != "delivery-1" || result.LastReceipt.Attempt != 1 || roundTrip.callCount() != 1 {
		t.Fatalf("result=%+v error=%v calls=%d", result, err, roundTrip.callCount())
	}
}

func TestReceiptTimesAreExactUTCMicroseconds(t *testing.T) {
	t.Parallel()

	local := time.FixedZone("test-non-UTC", 5*60*60+30*60)
	startedInput := time.Date(2026, 9, 6, 9, 10, 11, 456789123, local)
	finishedInput := startedInput.Add(2 * time.Millisecond)
	wantStarted := startedInput.UTC().Truncate(time.Microsecond)
	wantFinished := finishedInput.UTC().Truncate(time.Microsecond)

	tests := []struct {
		name          string
		transport     func() http.RoundTripper
		recorderError error
		wantError     error
		wantOutcome   Outcome
	}{
		{
			name: "delivered",
			transport: func() http.RoundTripper {
				return &scriptedTransport{steps: []transportStep{{status: http.StatusNoContent}}}
			},
			wantOutcome: OutcomeDelivered,
		},
		{
			name: "permanent transport failure",
			transport: func() http.RoundTripper {
				return &scriptedTransport{steps: []transportStep{{err: errors.New("permanent transport failure")}}}
			},
			wantError:   ErrPermanent,
			wantOutcome: OutcomePermanent,
		},
		{
			name: "receipt failure",
			transport: func() http.RoundTripper {
				return &scriptedTransport{steps: []transportStep{{status: http.StatusNoContent}}}
			},
			recorderError: errors.New("receipt unavailable"),
			wantError:     ErrReceipt,
			wantOutcome:   OutcomeDelivered,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			recorder := &memoryRecorder{err: tc.recorderError}
			validated, err := validateConfig(Config{
				Secret:   Secret{KeyID: "key-1", Value: []byte(strings.Repeat("s", minSecretBytes))},
				Recorder: recorder,
				Retry:    RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond},
			})
			if err != nil {
				t.Fatal(err)
			}
			clockValues := []time.Time{startedInput, finishedInput}
			clockCalls := 0
			clock := func() time.Time {
				if clockCalls >= len(clockValues) {
					t.Fatalf("clock called more than %d times", len(clockValues))
				}
				value := clockValues[clockCalls]
				clockCalls++
				return value
			}
			d := newDispatcher(validated, dependencies{transport: tc.transport(), now: clock, wait: noWait})

			result, err := d.Deliver(context.Background(), validMessage())
			if !errors.Is(err, tc.wantError) {
				t.Fatalf("error=%v want=%v", err, tc.wantError)
			}
			if result.Outcome != tc.wantOutcome || clockCalls != len(clockValues) {
				t.Fatalf("result=%+v clock calls=%d", result, clockCalls)
			}
			recorded := recorder.snapshot()
			if len(recorded) != 1 {
				t.Fatalf("recorded receipts=%d", len(recorded))
			}
			receipt := recorded[0]
			if receipt != result.LastReceipt {
				t.Fatalf("recorded receipt=%+v LastReceipt=%+v", receipt, result.LastReceipt)
			}
			if receipt.StartedAt != wantStarted || receipt.FinishedAt != wantFinished {
				t.Fatalf("receipt times=(%v, %v), want=(%v, %v)", receipt.StartedAt, receipt.FinishedAt, wantStarted, wantFinished)
			}
			if receipt.StartedAt.Location() != time.UTC || receipt.FinishedAt.Location() != time.UTC ||
				receipt.StartedAt.Nanosecond()%int(time.Microsecond) != 0 ||
				receipt.FinishedAt.Nanosecond()%int(time.Microsecond) != 0 {
				t.Fatalf("receipt times are not exact UTC microseconds: %+v", receipt)
			}
			if got, want := receipt.FinishedAt.Sub(receipt.StartedAt), finishedInput.Sub(startedInput); got != want || got < 0 {
				t.Fatalf("receipt duration=%s want=%s", got, want)
			}
			if receipt.RequestTimestamp != wantStarted.Unix() {
				t.Fatalf("request timestamp=%d want=%d", receipt.RequestTimestamp, wantStarted.Unix())
			}
		})
	}
}

func TestDeliverCopiesPayloadAndSigningSecret(t *testing.T) {
	t.Parallel()

	body := []byte("original")
	secretBytes := []byte(strings.Repeat("s", minSecretBytes))
	roundTrip := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		body[0] = 'X'
		secretBytes[0] = 'X'
		got, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		if string(got) != "original" {
			return nil, errors.New("payload was not copied")
		}
		return response(http.StatusNoContent, "", ""), nil
	})
	recorder := &memoryRecorder{}
	cfg := Config{Secret: Secret{KeyID: "key-1", Value: secretBytes}, Recorder: recorder, Retry: RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}}
	validated, err := validateConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	d := newDispatcher(validated, dependencies{transport: roundTrip, now: fixedClock, wait: noWait})
	msg := validMessage()
	msg.Body = body
	if _, err := d.Deliver(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
}

func TestDispatcherConcurrentUse(t *testing.T) {
	t.Parallel()

	recorder := &memoryRecorder{}
	roundTrip := roundTripperFunc(func(_ *http.Request) (*http.Response, error) { return response(http.StatusNoContent, "", ""), nil })
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := validMessage()
			msg.DeliveryID = "delivery-" + strconv.Itoa(i)
			if _, err := d.Deliver(context.Background(), msg); err != nil {
				t.Errorf("deliver: %v", err)
			}
		}()
	}
	wg.Wait()
	if got := len(recorder.snapshot()); got != 50 {
		t.Fatalf("receipts = %d", got)
	}
}

func TestDispatcherCallsRecorderConcurrently(t *testing.T) {
	t.Parallel()

	recorder := newOverlapRecorder(2)
	transport := roundTripperFunc(func(_ *http.Request) (*http.Response, error) { return response(http.StatusNoContent, "", ""), nil })
	d := testDispatcher(t, recorder, transport, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	var wg sync.WaitGroup
	for i := range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			msg := validMessage()
			msg.DeliveryID = "overlap-" + strconv.Itoa(i)
			if _, err := d.Deliver(context.Background(), msg); err != nil {
				t.Errorf("deliver: %v", err)
			}
		}()
	}
	wg.Wait()
	if recorder.maximum() < 2 {
		t.Fatalf("maximum concurrent Record calls=%d", recorder.maximum())
	}
}

func validMessage() Message {
	return Message{Endpoint: "https://example.com/hook", DeliveryID: "delivery-1", EventType: "event.one", ContentType: "application/json", Body: []byte("body")}
}

func fixedClock() time.Time { return time.Unix(1_700_000_000, 0).UTC() }

func noWait(context.Context, time.Duration) error { return nil }

func testDispatcher(t *testing.T, recorder Recorder, transport http.RoundTripper, retry RetryPolicy, wait func(context.Context, time.Duration) error) *Dispatcher {
	t.Helper()
	return testDispatcherWithTimeout(t, recorder, transport, retry, wait, time.Second)
}

func testDispatcherWithTimeout(t *testing.T, recorder Recorder, transport http.RoundTripper, retry RetryPolicy, wait func(context.Context, time.Duration) error, timeout time.Duration) *Dispatcher {
	t.Helper()
	validated, err := validateConfig(Config{Secret: Secret{KeyID: "key-1", Value: []byte(strings.Repeat("s", minSecretBytes))}, Recorder: recorder, Retry: retry, AttemptTimeout: timeout})
	if err != nil {
		t.Fatal(err)
	}
	return newDispatcher(validated, dependencies{transport: transport, now: fixedClock, wait: wait})
}

type memoryRecorder struct {
	mu                 sync.Mutex
	receipts           []Receipt
	err                error
	requireLiveContext bool
}

type overlapRecorder struct {
	mu      sync.Mutex
	active  int
	max     int
	want    int
	reached chan struct{}
	once    sync.Once
}

func newOverlapRecorder(want int) *overlapRecorder {
	return &overlapRecorder{want: want, reached: make(chan struct{})}
}

func (r *overlapRecorder) Record(ctx context.Context, _ Receipt) error {
	r.mu.Lock()
	r.active++
	if r.active > r.max {
		r.max = r.active
	}
	if r.active == r.want {
		r.once.Do(func() { close(r.reached) })
	}
	r.mu.Unlock()
	select {
	case <-r.reached:
	case <-ctx.Done():
		return ctx.Err()
	}
	r.mu.Lock()
	r.active--
	r.mu.Unlock()
	return nil
}

func (r *overlapRecorder) maximum() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.max
}

func (r *memoryRecorder) Record(ctx context.Context, receipt Receipt) error {
	if r.requireLiveContext && ctx.Err() != nil {
		return errors.New("record context was canceled")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.receipts = append(r.receipts, receipt)
	return r.err
}

func (r *memoryRecorder) snapshot() []Receipt {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Receipt(nil), r.receipts...)
}

type transportStep struct {
	status     int
	body       string
	retryAfter string
	err        error
}

type scriptedTransport struct {
	mu    sync.Mutex
	steps []transportStep
	calls int
}

type lifecycleTransport struct {
	mu               sync.Mutex
	roundTrips       int
	closeCalls       int
	roundTripEntered chan struct{}
	releaseRoundTrip chan struct{}
	closeEntered     chan struct{}
	releaseClose     chan struct{}
}

func newLifecycleTransport() *lifecycleTransport {
	return &lifecycleTransport{
		roundTripEntered: make(chan struct{}),
		releaseRoundTrip: make(chan struct{}),
		closeEntered:     make(chan struct{}),
		releaseClose:     make(chan struct{}),
	}
}

func (t *lifecycleTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	t.mu.Lock()
	t.roundTrips++
	t.mu.Unlock()
	close(t.roundTripEntered)
	<-t.releaseRoundTrip
	return response(http.StatusNoContent, "", ""), nil
}

func (t *lifecycleTransport) CloseIdleConnections() {
	t.mu.Lock()
	t.closeCalls++
	t.mu.Unlock()
	close(t.closeEntered)
	<-t.releaseClose
}

func (t *lifecycleTransport) roundTripCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.roundTrips
}

func (t *lifecycleTransport) closeCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.closeCalls
}

func receiveBefore(t *testing.T, ch <-chan struct{}, operation string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", operation)
	}
}

func (s *scriptedTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	step := s.steps[s.calls]
	s.calls++
	if step.err != nil {
		return nil, step.err
	}
	return response(step.status, step.body, step.retryAfter), nil
}

func (s *scriptedTransport) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func response(status int, body, retryAfter string) *http.Response {
	header := make(http.Header)
	if retryAfter != "" {
		header.Set("Retry-After", retryAfter)
	}
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body))}
}

func captureHTTPTransportError(t *testing.T, rawResponse string, maxHeaderBytes int64) error {
	t.Helper()

	clientConn, serverConn := net.Pipe()
	transport := &http.Transport{
		DialContext: func(context.Context, string, string) (net.Conn, error) {
			return clientConn, nil
		},
		MaxResponseHeaderBytes: maxHeaderBytes,
	}
	type serverResult struct {
		readErr  error
		writeErr error
	}
	serverDone := make(chan serverResult, 1)
	go func() {
		defer serverConn.Close()
		request, err := http.ReadRequest(bufio.NewReader(serverConn))
		if err != nil {
			serverDone <- serverResult{readErr: err}
			return
		}
		_ = request.Body.Close()
		_, err = io.WriteString(serverConn, rawResponse)
		serverDone <- serverResult{writeErr: err}
	}()
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.test/", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, transportErr := transport.RoundTrip(request)
	if response != nil && response.Body != nil {
		_ = response.Body.Close()
	}
	transport.CloseIdleConnections()
	server := <-serverDone
	if server.readErr != nil {
		t.Fatalf("server failed to read transport request: %v", server.readErr)
	}
	if transportErr == nil {
		t.Fatal("transport unexpectedly accepted malformed response")
	}
	return transportErr
}

type contextReader struct{ ctx context.Context }

func (r contextReader) Read([]byte) (int, error) {
	<-r.ctx.Done()
	return 0, r.ctx.Err()
}

type cancelingReader struct {
	cancel context.CancelFunc
	ctx    context.Context
}

func (r cancelingReader) Read([]byte) (int, error) {
	r.cancel()
	<-r.ctx.Done()
	return 0, r.ctx.Err()
}
