# Architecture

## Boundary and authority split

The sole public package is `pkg/webhooks`. The consumer decides whether an
event may leave the application, chooses the endpoint, defines the event type,
minimizes the opaque body, allocates durable delivery identity, coordinates
concurrent retries, stores receipts, and manages verification-key retirement.

The library validates the mechanical envelope, signs it, resolves and dials a
public destination through an owned HTTP transport, classifies the result,
records an attempt, and performs only bounded in-process retries. It never
stores a payload or defines its meaning.

```text
consumer authorization/minimization
              |
              v
       Dispatcher.Deliver
       | validate/copy
       | sign attempt
       | resolve -> reject special IPs -> dial literal
       | POST once, no proxy, no redirect
       | read bounded response
       v
       Recorder.Record ----> consumer-owned durable store
              |
              +---- retry only after the record succeeds
```

## Signing and identity

Every attempt carries a stable delivery ID plus a one-based attempt number and
Unix-second timestamp. HMAC-SHA-256 signs a domain-separated canonical input
whose fields are newline-delimited after validation forbids newlines. The body
is represented by its SHA-256 digest. The normalized absolute HTTPS target
includes the literal `https://` scheme, lowercase DNS name or bracketed IP,
explicit `:443`, and exact request target, so a signature cannot be moved to
another scheme, host, path, or query.

The key ID is public routing metadata, not a secret. It is covered by the MAC.
A dispatcher has one current signing key. Rotation creates a new dispatcher
with a new current key; receivers retain old keys only for their declared
maximum replay/retention window. Secrets are copied at construction and never
appear in receipts or returned errors.

## Network boundary

The production constructor owns `http.Transport`; callers cannot inject a
RoundTripper, dialer, proxy, or TLS bypass. Only HTTPS port 443 is accepted.
The dispatcher calls the owned transport's `RoundTrip` exactly once per
attempt. It never invokes `http.Client` redirect processing, so a 3xx response
is returned for status classification even when `Location` is malformed. The
transport disables proxies. For each new TCP connection its
dial hook resolves the hostname, normalizes mapped addresses, rejects the
entire answer if any address is loopback, private, link-local, multicast,
unspecified, or in the library's explicit special-purpose prefix table, then
dials only a validated numeric address. TLS still authenticates the original
hostname through Go's standard transport.

Connection reuse does not re-resolve DNS for every request; it reuses a TCP/TLS
connection already established to a validated address. Every new connection
repeats resolution and validation. This closes the ordinary DNS-rebinding
time-of-check/time-of-use gap because unvalidated hostnames are never passed to
the dialer. It cannot defend against a compromised resolver, a malicious
dialer/kernel, route changes after address validation, NAT or service-mesh
remapping, a compromised public destination, or an already-established public
server pivoting at the application layer. Operators must enforce egress policy
below this library as a second boundary.

The address policy rejects every allocation in the IANA IPv4 and IPv6
Special-Purpose Address Registries snapshot last updated 2025-10-09, including
allocations whose registry Global flag is true. IPv6 outside IANA's allocated
global-unicast `2000::/3` is also rejected. The exact source hashes and compact
prefix derivation are pinned in [address policy](address-policy.md). A new IANA
reservation can predate a library update, so production network policy must
remain fail-closed independently of this user-space check.

## Delivery and failure model

An HTTP attempt has four terminal classifications: delivered, retryable,
permanent, or canceled. A 2xx response is delivered. Status 408, 425, 429, and
5xx are retryable. Other HTTP responses, including every 3xx, are permanent.
Only explicitly typed transient failures are retryable: deadline timeouts,
safe-dialer-marked transient lookup/dial failures, selected connection errno
values, and EOF/truncation. Destination rejection, certificate validation, TLS
alerts/record failures, and HTTP protocol/header-limit failures are permanent.
An unknown transport error is permanent when no attempt deadline expired.
Caller cancellation remains canceled and has
first precedence. After that check, an observed destination, certificate,
TLS-record/alert, or HTTP-protocol failure remains permanent even when the
per-attempt deadline expires at the same edge; the attempt-deadline fallback
applies only after those observed permanent classes are excluded.
When every validated address fails to dial, the aggregate is retryable only if
every observed dial failure belongs to the transient allowlist; any local
resource/configuration or unknown failure in a mixed set makes the aggregate
permanent. This fail-closed rule avoids hiding a deterministic failure behind a
different address's transient failure.

While the process survives, the library records an attempt result before
another attempt or return. Recording uses a separate bounded context so caller
cancellation does not silently erase the receipt. If recording fails, delivery
stops with an explicit receipt error; `Result.LastReceipt` exposes the exact
record for reconciliation. Its delivery fingerprint is sensitive derived data
and must not be logged or exposed broadly. A process crash after sending but
before recording can leave an unrecorded unknown attempt. No library can
atomically commit a receiver effect and sender receipt across HTTP. Consumers
and receivers must use stable identity for durable deduplication and reconcile
unknown/in-flight work before another send.

## Concurrency

Dispatchers contain immutable configuration plus Go's concurrency-safe HTTP
transport and may serve concurrent unrelated deliveries. The same `Recorder`
can be called concurrently and must implement synchronization appropriate to
its store. Two calls with the same
delivery ID can both send; preventing that requires a consumer-owned durable
lease or uniqueness mechanism. The consumer also persists and allocates the
next attempt number across invocations. Hiding a partial in-memory map here
would be garbage because it would fail across processes and restarts.
