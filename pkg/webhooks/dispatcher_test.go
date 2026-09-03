package webhooks

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

func TestDeliverExhaustsTransportFailures(t *testing.T) {
	t.Parallel()

	roundTrip := &scriptedTransport{steps: []transportStep{{err: errors.New("broken")}, {err: errors.New("still broken")}}}
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

func TestDeliverExhaustsRetryableHTTPWithoutTransportCause(t *testing.T) {
	t.Parallel()

	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, &scriptedTransport{steps: []transportStep{{status: http.StatusServiceUnavailable}}}, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrExhausted) || result.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

func TestDeliverResponseReadFailureIsRetryable(t *testing.T) {
	t.Parallel()

	want := errors.New("response stream failed")
	roundTrip := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: &trackedBody{Reader: errorReader{err: want}}}, nil
	})
	recorder := &memoryRecorder{}
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 1, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrExhausted) || !errors.Is(err, want) || result.Outcome != OutcomeRetryable {
		t.Fatalf("result=%+v error=%v", result, err)
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

	recorder := &memoryRecorder{err: errors.New("store unavailable")}
	roundTrip := &scriptedTransport{steps: []transportStep{{status: http.StatusInternalServerError}, {status: http.StatusNoContent}}}
	d := testDispatcher(t, recorder, roundTrip, RetryPolicy{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}, noWait)
	result, err := d.Deliver(context.Background(), validMessage())
	if !errors.Is(err, ErrReceipt) || result.Attempts != 1 || result.LastReceipt.DeliveryID != "delivery-1" || result.LastReceipt.Attempt != 1 || roundTrip.callCount() != 1 {
		t.Fatalf("result=%+v error=%v calls=%d", result, err, roundTrip.callCount())
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
	d := newDispatcher(validated, dependencies{client: &http.Client{Transport: roundTrip}, now: fixedClock, wait: noWait})
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
	return newDispatcher(validated, dependencies{client: &http.Client{Transport: transport}, now: fixedClock, wait: wait})
}

type memoryRecorder struct {
	mu                 sync.Mutex
	receipts           []Receipt
	err                error
	requireLiveContext bool
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
