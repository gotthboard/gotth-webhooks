# Implementation specification

## Public contract

- `New(Config)`: validate policy, copy the signing secret, and construct the
  owned hardened transport.
- `NewDeliveryID()`: return a 192-bit cryptographically random base64url ID.
- `Dispatcher.Deliver(ctx, Message)`: validate and copy the message, perform at
  most `RetryPolicy.MaxAttempts`, durably record every actual attempt, and
  return the final known result.
- `Recorder.Record(ctx, Receipt)`: consumer-owned persistence boundary. Calls
  may be repeated after an unknown store outcome, so implementations must
  be concurrency-safe, upsert by `(delivery_id, attempt)`, and reject
  conflicting records.

`Config` contains one `Secret`, a bounded `RetryPolicy`, attempt and receipt
timeouts, and a required `Recorder`. Production dependencies are fixed. Tests
exercise internal dependency seams that are not available to consumers.

`Message` contains endpoint, stable delivery ID, consumer-durable first attempt
number, opaque event type, content type, and body. Zero first attempt defaults
to one and is suitable only for the first invocation. A consumer retry after a
process boundary must allocate the next monotonically increasing number from
durable state. `Deliver` copies the body before network work. The body limit is
1 MiB. Identifiers and event types are restricted printable tokens. Content
type must be 1 through 256 bytes, must parse as a media type, and is rejected
before parsing if any input byte is an ASCII C0 control (`0x00` through
`0x1f`) or DEL (`0x7f`). This includes CR, LF, and HTAB even where a MIME
parser might trim or preserve them. Non-ASCII input policy is otherwise
unchanged: MIME parameter values accepted by Go's parser remain supported and
deterministic formatting may serialize them with RFC 2231 UTF-8 percent
encoding.

## Wire contract

Requests are POST with these headers:

- `Content-Type`
- `X-Gotth-Webhook-ID`
- `X-Gotth-Webhook-Event`
- `X-Gotth-Webhook-Attempt`
- `X-Gotth-Webhook-Timestamp`
- `X-Gotth-Webhook-Key-ID`
- `X-Gotth-Webhook-Signature: v1=<lowercase hex HMAC-SHA-256>`

The signature input is UTF-8 bytes:

```text
gotth-webhook-signature-v1\n
POST\n
https://<lowercase-host-or-bracketed-IP>:443<request-target>\n
<delivery-id>\n
<attempt decimal>\n
<timestamp decimal>\n
<event-type>\n
<content-type>\n
<key-id>\n
<lowercase hex SHA-256 body digest>\n
```

The signed target is an absolute URI beginning with the literal bytes
`https://`. Its URI authority always includes port `443`; DNS names are
lowercase ASCII and IPv6 literals are bracketed. An empty path becomes `/`; the
parsed escaped path and validated raw query form the request target. Raw query
bytes must match RFC 3986 `pchar / "/" / "?"`; percent escapes require exactly
two hex digits. Valid query order and escape spelling are preserved. Ambiguous
opaque URLs, userinfo, fragments, encoded-host tricks, invalid raw query bytes,
explicit empty ports, and non-443 ports fail validation.

### Conformance vector

For endpoint
`https://EXAMPLE.com/hook?b=2&a=%2F%3f&flag&x=one+two/three?four`, delivery ID
`delivery-1`, attempt `2`, timestamp `1700000000`, event `thing.changed`, content
type `application/json`, key ID `key-1`, body `{}`, and the 32 ASCII-byte secret
`0123456789abcdef0123456789abcdef`, the exact signed bytes are:

```text
gotth-webhook-signature-v1
POST
https://example.com:443/hook?b=2&a=%2F%3f&flag&x=one+two/three?four
delivery-1
2
1700000000
thing.changed
application/json
key-1
44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a
```

The digest line has a final newline. The expected header is
`X-Gotth-Webhook-Signature: v1=5afc952ae607736b86e3570dddc25991115f80afdb46c49832a30a5671028f7f`.

Receivers must compare MACs in constant time, enforce their own timestamp
window, bind the delivery ID to immutable event/body semantics, and durably
deduplicate before producing side effects.

## Limits and retry state machine

- payload: 1 MiB;
- response body: 64 KiB plus one-byte overflow probe;
- response headers: 64 KiB;
- attempts: 1 through 10;
- attempt timeout: 100 ms through 2 minutes;
- receipt timeout: 100 ms through 30 seconds;
- initial/max retry delay: 1 ms through 1 minute, with max not below initial;
- secret: 32 through 1024 bytes;
- delivery/key/event token: 1 through 128 ASCII token characters;
- attempt number: 1 through 1,000,000,000, including all attempts in a call;
- content type: 1 through 256 bytes.

After a retryable outcome, the delay is exponential and saturating. A valid
`Retry-After` delta is ASCII `1*DIGIT`; parsing saturates without integer
overflow. A valid delta or HTTP date may increase delay but never exceed the
configured maximum, and values beyond the response-header bound are rejected.
Cancellation during an attempt or wait returns promptly.
No jitter is applied in V1 because deterministic policy is more useful to a
consumer-owned durable scheduler; callers should distribute scheduling above
this library when operating large fleets.

## Receipt contract

A receipt contains delivery ID, a stable SHA-256 delivery fingerprint,
one-based attempt, request timestamp, start and finish times, outcome, status
code when present, bounded response-byte count, and a stable error class. The
fingerprint binds normalized target, event type, content type, and body while
excluding attempt, timestamp, and signing key so retries and key rotation keep
one identity. It contains no raw endpoint, event, body, key ID, signature,
response body, or raw error string. The unkeyed fingerprint is still sensitive
derived data: it enables correlation and guesses of low-entropy endpoint/query/
body semantics. Receipt stores must restrict access and must not log or expose
the fingerprint. Its stable form lets the durable store reject one delivery ID
reused with different semantics across signing-key rotation.

`Record` runs after the response body closes or the transport returns. It uses
`context.WithoutCancel` plus the configured receipt timeout. A store failure is
wrapped with `ErrReceipt` and prevents further sends. A store implementation
must make exact replay idempotent and conflicting replay an error.

Go context deadlines are cooperative. The library supplies a bounded record
context but cannot force a broken `Recorder` implementation to return. The
recorder is trusted consumer infrastructure and must honor the context and be
safe for concurrent calls from one or more dispatchers.
When recording returns an unknown failure, `Result.LastReceipt` contains the
exact record so the consumer can query or replay the idempotent store operation
without issuing another HTTP request.

Raw transport and recorder errors are not propagated beyond the dispatcher;
standard HTTP errors can contain the full target URL, including a sensitive
query. Callers receive stable sentinels and inspect bounded result/receipt
classes, while treating any delivery fingerprint as sensitive. Destination
policy, deterministic TLS/protocol, and unknown transport failures are
permanent; only the documented typed transient allowlist is retryable.
For an exhausted multi-address dial, every observed error must be in that
allowlist before the aggregate is retryable. A local resource/configuration or
unknown error makes a mixed aggregate permanent and its details are redacted.

## Production-unit order

1. values, endpoint canonicalization, public-address policy, and random ID;
2. canonical signing and request construction;
3. retry classification, delay parsing, and bounded response consumption;
4. receipt recording and the delivery loop;
5. hardened transport and integration boundaries;
6. external consumer, performance, and review admission.

Each production function receives an adjacent complexity contract naming byte
inputs, address counts, attempts, I/O, allocations, and delegated costs.
