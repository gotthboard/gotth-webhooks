package webhooks

import "errors"

var (
	// ErrInvalid identifies rejected configuration or message input.
	ErrInvalid = errors.New("webhooks: invalid input")
	// ErrPermanent identifies an HTTP response that must not be retried.
	ErrPermanent = errors.New("webhooks: permanent delivery failure")
	// ErrExhausted identifies a retryable outcome after all attempts were used.
	ErrExhausted = errors.New("webhooks: attempts exhausted")
	// ErrReceipt identifies failure to durably record a known attempt outcome.
	ErrReceipt = errors.New("webhooks: receipt recording failed")
	// ErrResponseTooLarge identifies a response body beyond the fixed limit.
	ErrResponseTooLarge = errors.New("webhooks: response body too large")
	// ErrDestination identifies DNS or address-policy rejection.
	ErrDestination = errors.New("webhooks: destination rejected")
)
