// Package webhooks sends bounded, HMAC-authenticated outbound webhook
// requests. Consumers retain authority over event meaning, authorization,
// payload minimization, persistence, and receiver deduplication.
//
// Delivery is not exactly once. A receiver can finish work before the sender
// observes a timeout. Receivers must durably deduplicate the stable delivery ID
// before producing effects.
package webhooks
