package webhooks

import (
	"context"
	"time"
)

const (
	// MaxPayloadBytes is the largest accepted request body.
	MaxPayloadBytes = 1 << 20
	// MaxResponseBytes is the largest response body consumed.
	MaxResponseBytes = 64 << 10

	maxEndpointBytes       = 2048
	maxTokenBytes          = 128
	maxContentType         = 256
	maxResponseHeaderBytes = 64 << 10
	minSecretBytes         = 32
	maxSecretBytes         = 1024

	defaultMaxAttempts    = 3
	defaultAttemptTimeout = 15 * time.Second
	defaultReceiptTimeout = 5 * time.Second
	defaultInitialDelay   = 250 * time.Millisecond
	defaultMaxDelay       = 10 * time.Second
)

const (
	HeaderDeliveryID = "X-Gotth-Webhook-ID"
	HeaderEvent      = "X-Gotth-Webhook-Event"
	HeaderAttempt    = "X-Gotth-Webhook-Attempt"
	HeaderTimestamp  = "X-Gotth-Webhook-Timestamp"
	HeaderKeyID      = "X-Gotth-Webhook-Key-ID"
	HeaderSignature  = "X-Gotth-Webhook-Signature"
)

// Secret is one current signing key. KeyID is public routing metadata; Value
// is secret material and is copied by New.
type Secret struct {
	KeyID string
	Value []byte
}

// RetryPolicy bounds attempts and delays. The all-zero value selects the
// documented defaults; partially specified policies are invalid.
type RetryPolicy struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

// Config defines immutable dispatcher policy.
type Config struct {
	Secret         Secret
	Retry          RetryPolicy
	AttemptTimeout time.Duration
	ReceiptTimeout time.Duration
	Recorder       Recorder
}

// Message is one caller-authorized opaque delivery. DeliveryID must remain
// stable across retries of exactly the same logical event.
type Message struct {
	Endpoint   string
	DeliveryID string
	// FirstAttempt is a consumer-durable, globally monotonic attempt number for
	// this delivery. Zero means one and is suitable only for a first invocation.
	FirstAttempt int
	EventType    string
	ContentType  string
	Body         []byte
}

// Outcome is the stable result class persisted in a Receipt.
type Outcome string

const (
	OutcomeDelivered Outcome = "delivered"
	OutcomeRetryable Outcome = "retryable"
	OutcomePermanent Outcome = "permanent"
	OutcomeCanceled  Outcome = "canceled"
)

// ErrorCode is a stable, non-sensitive attempt failure class.
type ErrorCode string

const (
	ErrorNone          ErrorCode = ""
	ErrorTransport     ErrorCode = "transport"
	ErrorTimeout       ErrorCode = "timeout"
	ErrorCanceled      ErrorCode = "canceled"
	ErrorDestination   ErrorCode = "destination"
	ErrorHTTPStatus    ErrorCode = "http_status"
	ErrorResponseLimit ErrorCode = "response_limit"
)

// Receipt is the minimal durable audit record for one actual HTTP attempt. It
// intentionally excludes endpoint, event, payload, key, signature, response
// body, and raw error text. DeliveryFingerprint is sensitive derived data:
// callers must restrict its storage and must not log it. Library-produced
// StartedAt and FinishedAt values are in UTC and truncated to exact
// microsecond precision before Record and Result exposure.
type Receipt struct {
	DeliveryID          string
	DeliveryFingerprint [32]byte
	Attempt             int
	RequestTimestamp    int64
	StartedAt           time.Time
	FinishedAt          time.Time
	Outcome             Outcome
	StatusCode          int
	ResponseBytes       int64
	ErrorCode           ErrorCode
}

// Recorder persists attempt receipts. Implementations must be safe for
// concurrent Record calls, make an exact repeated (delivery ID, attempt)
// record idempotent, and reject a conflict.
type Recorder interface {
	Record(context.Context, Receipt) error
}

// Result reports the final known network outcome even when receipt recording
// subsequently fails.
type Result struct {
	DeliveryID  string
	Attempts    int
	LastAttempt int
	Delivered   bool
	StatusCode  int
	Outcome     Outcome
	// LastReceipt lets a caller reconcile or replay an exact Record operation
	// after ErrReceipt without sending the HTTP request again. Its
	// DeliveryFingerprint is sensitive derived data and must not be logged.
	LastReceipt Receipt
}
