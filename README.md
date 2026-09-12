# gotth-webhooks

Consumer-neutral mechanics for bounded, signed outbound webhooks.

> **Distribution:** Forgejo remains canonical development. GitHub is the public
> clone and future Go/release endpoint after release admission. The base
> standalone implementation and receipt-time compatibility repair are
> technically admitted; no tag or compatibility promise exists yet. Release and
> compatibility remain blocked until one real product consumer supplies a
> contract, validates behavior against an exact dependency pin, and completes
> the release gates. See
> [the distribution contract](docs/distribution.md).

> Report public bugs through GitHub Issues and security vulnerabilities through
> GitHub private reporting.

## Boundary

The package sends opaque bytes only after the caller has authorized and
minimized them. Consumers own event definitions, subscriptions, recipients,
consent/privacy policy, durable scheduling, receipt storage, secret storage,
and receiver authorization. The package does not know what a GOTTH Board event
is and does not invent one.

V1 provides:

- HMAC-SHA-256 signatures binding target, delivery ID, attempt, timestamp,
  event type, content type, key ID, and payload digest;
- a stable delivery identity plus a sensitive derived semantics fingerprint
  for sender-side conflict detection;
- HTTPS-only HTTP/1 endpoint validation, public-address DNS checks immediately
  before literal-IP dialing, no ambient proxy, no implicit protocol replay,
  and no redirects;
- explicit bounded retry classification, `Retry-After`, cancellation,
  per-attempt timeout, response-header limit, and response-body limit;
- mandatory consumer-owned receipt recording before retry or return;
- explicit signing key IDs for bounded receiver-side secret rotation; and
- explicit concurrent-safe dispatcher retirement that rejects new delivery
  admission and releases owned idle transport connections.

It does not provide exactly-once delivery. A receiver can commit work before a
sender observes a timeout. The receiver must durably deduplicate the stable
delivery ID before effects, and the sender must reconcile unknown outcomes.
A process crash after a request but before receipt recording can leave an
unrecorded unknown attempt; recovery must not blindly resend it.

## Installation and compatibility

There is no release to install yet. Standalone technical admission does not
create a compatibility promise. Before release, one real consumer must supply
a contract and complete behavioral verification against an exact candidate
dependency pin as required by [the release policy](docs/RELEASING.md). The
local external-module fixture proves compilation only; it is not a real
consumer or compatibility oracle.

The current candidate module is `github.com/gotthboard/gotth-webhooks`, requires
Go 1.26.6, uses only the Go standard library, and supports HTTPS port 443. The
first compatibility contract remains unstable until the real-consumer gate and
release tag are complete.

## API

```go
type receipts struct{}

func (receipts) Record(ctx context.Context, receipt webhooks.Receipt) error {
	// Upsert by (DeliveryID, Attempt). Exact replay is idempotent; reject the
	// same DeliveryID with a different DeliveryFingerprint.
	return durableStoreReceipt(ctx, receipt)
}

dispatcher, err := webhooks.New(webhooks.Config{
	Secret: webhooks.Secret{
		KeyID: "2026-09-current",
		Value: currentSecretFromSecretManager,
	},
	Recorder: receipts{},
})
if err != nil {
	return err
}
defer dispatcher.Close()

result, err := dispatcher.Deliver(ctx, webhooks.Message{
	Endpoint:    authorizedEndpoint,
	DeliveryID:  durableDeliveryID,
	FirstAttempt: durableNextAttempt, // zero is shorthand for the first call
	EventType:   authorizedOpaqueType,
	ContentType: "application/json",
	Body:        minimizedOpaquePayload,
})
```

The all-zero retry/timeouts select three attempts, 250 ms initial delay, 10 s
maximum delay, 15 s attempt timeout, and 5 s receipt timeout. Nonzero policies
must satisfy the documented bounds in [the implementation spec](docs/implementation-spec.md).

`Deliver` is concurrency-safe for unrelated deliveries, and the dispatcher can
call one `Recorder` concurrently. Recorder implementations must therefore be
concurrency-safe. Calls sharing one
delivery ID must be serialized or leased durably by the consumer, which also
allocates a monotonically increasing `FirstAttempt` across invocations. On an
`ErrReceipt`, `Result.LastReceipt` contains the exact record for store
reconciliation without resending. Its stable fingerprint is sensitive derived
data that can support guessing and correlation; restrict access to stored
receipts and never log the fingerprint or whole result. Library-produced
`StartedAt` and `FinishedAt` values are explicitly UTC and truncated to exact
microsecond precision before both recorder and result exposure. An in-memory
map would not protect multiple processes or survive a crash, so the library
does not fake that guarantee.

`Close` is safe to call repeatedly and concurrently. Once its closed transition
occurs, new `Deliver` admissions return `ErrClosed`. It waits for calls already
admitted, including their retries and receipt recording, then releases the
dispatcher's owned idle HTTP connections. It does not cancel delivery, so a
dependency that violates its context contract can also prevent `Close` from
returning. A transport, wait, or recorder callback must not synchronously call
`Close` on its own dispatcher because it is part of the delivery being drained.
Consumers rotating signing generations should first stop new work
for the old generation, then close and discard that dispatcher after `Close`
returns.

## Receiver contract

Every request is POST and carries:

- `X-Gotth-Webhook-ID`
- `X-Gotth-Webhook-Event`
- `X-Gotth-Webhook-Attempt`
- `X-Gotth-Webhook-Timestamp`
- `X-Gotth-Webhook-Key-ID`
- `X-Gotth-Webhook-Signature: v1=<hex HMAC-SHA-256>`

The exact canonical input is specified in
[the implementation spec](docs/implementation-spec.md). Receivers must select a
bounded active/retired secret by key ID, recompute the MAC, compare with
`hmac.Equal`, enforce a timestamp window, bind the delivery ID to immutable
semantics, and durably deduplicate before side effects.

Rotation means constructing a dispatcher with the new current key/key ID while
receivers temporarily retain the old verification key for their declared
replay window. Key IDs are signed public metadata. Secrets are never stored in
receipts or errors.

## SSRF boundary and its limits

The production constructor does not accept custom transports, proxies, TLS
bypasses, private addresses, alternate ports, or redirect policy. Each new TCP
connection resolves and validates every answer before dialing a numeric public
address; mixed public/private answers fail closed. Existing validated
connections may be reused without repeated DNS lookup.

The process check rejects every allocation in the pinned IANA IPv4/IPv6
Special-Purpose Address Registry snapshot, even entries marked globally
reachable, and admits IPv6 only from allocated global-unicast `2000::/3`.
This is defense in depth, not a firewall. A compromised resolver, dialer,
kernel, route, NAT/service-mesh remap, public destination, or new IANA
allocation after the pinned snapshot can defeat assumptions below this
process. Operators must enforce independent egress policy. See
[architecture](docs/architecture.md), [address policy](docs/address-policy.md),
and [security policy](SECURITY.md).

## Scope exclusions

- Product events, subscriptions, payload schemas, privacy policy, or UI.
- Inbound webhooks, receiver routing, queues, schedulers, or workflows.
- Private endpoints, HTTP, custom ports, proxies, redirects, or mTLS in V1.
- Automatic secret management, receipt database schema, or deployment.

Release policy, contribution path, and current verification evidence are in
[releasing](docs/RELEASING.md), [contributing](CONTRIBUTING.md), and
[verification](docs/verification.md).
