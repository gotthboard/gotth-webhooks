# Product requirements

## Problem

GOTTH applications need to deliver authorized outbound webhook payloads
without each consumer improvising signatures, retry rules, or network egress
defenses. HTTP delivery can be duplicated and can have an unknown outcome. A
reusable library must expose that fact instead of selling fake exactly-once
semantics.

The current GOTTH Board product PRD places product webhooks in v4 and does not
define events, recipients, subscriptions, payload schemas, or privacy rules.
Those decisions therefore remain consumer-owned. This candidate explores
consumer-neutral outbound mechanics over opaque caller-provided bytes, but its
public API is not admissible until one real consumer requirement validates the
surface. A synthetic compile fixture is not that requirement.

## Requirements

- `WHK-001`: Accept an opaque event type and payload only after the consumer
  has authorized and minimized them; never define or infer product events.
- `WHK-002`: Accept only canonical HTTPS endpoints without userinfo,
  fragments, non-default ports, or ambiguous URL forms.
- `WHK-003`: Resolve a destination immediately before dialing, reject every
  address in the pinned IANA IPv4/IPv6 Special-Purpose Address Registries
  regardless of their Global flag, admit IPv6 only from allocated global
  unicast `2000::/3`, reject any other non-global address, dial only a validated
  literal address, ignore ambient proxy settings, and never process redirects.
- `WHK-004`: Sign each POST with HMAC-SHA-256 over a versioned canonical input
  binding the target, delivery ID, attempt, timestamp, event type, content
  type, key ID, and payload digest.
- `WHK-005`: Put a stable caller-supplied delivery ID and attempt number on
  every request so a receiver can reject replays and deduplicate retries.
- `WHK-006`: Provide cryptographically random delivery IDs and reject malformed
  identifiers, key IDs, event types, secrets, content types, and oversized
  bodies.
- `WHK-007`: Retry only an explicit typed allowlist of transient transport failures and
  HTTP statuses, cap `Retry-After`, and stop on cancellation, permanent
  failure, exhausted attempts, or receipt-recording failure.
- `WHK-008`: Bound connection, TLS, whole-attempt, response-header, and response
  body work. Drain no unbounded response and return cancellation promptly.
- `WHK-009`: While the process survives, record every attempted request result
  through a required consumer-owned durable interface before returning or
  starting another attempt. Records contain metadata, never payloads, secrets,
  signatures, raw endpoints, response bodies, or raw errors. The stable
  semantics fingerprint is sensitive derived data, not log-safe metadata. A
  crash between HTTP completion and recording remains an explicit unknown outcome.
- `WHK-010`: Support secret rotation explicitly through a signed key ID. A
  dispatcher signs with exactly one configured current key; receivers own the
  bounded active/retired verification-key set and retirement window.
- `WHK-011`: Remain safe for concurrent `Deliver` calls, require `Recorder`
  implementations to accept concurrent calls, and require the consumer to
  serialize or durably coordinate calls sharing a delivery ID and
  allocate monotonically increasing attempts across invocations.
- `WHK-012`: State that receiver processing and HTTP delivery are at-least-once
  possibilities, not exactly once; a timeout or transport error can follow a
  completed receiver action.
- `WHK-013`: Publish the source under the maintainer-selected MIT license.

## Non-goals

- Product events, subscriptions, recipients, authorization, consent, privacy
  policy, operator UI, or payload minimization.
- Inbound webhook routing or automatic receiver-side processing.
- A queue, scheduler, event bus, workflow engine, database adapter, or secret
  manager.
- Exactly-once delivery, exactly-once receiver effects, or automatic recovery
  from consumer process crashes.
- Private-network destinations, custom ports, HTTP, ambient proxies, redirects,
  mTLS, or arbitrary caller-provided transports in V1.

## Acceptance

Local implementation requirements trace to design, source, tests, and evidence.
Admission additionally requires one real consumer contract and pin; that
product input is currently blocked. Verification
includes format, vet, race, repeated race, statement coverage, fuzzing,
boundary and negative tests, a real loopback TLS transport test using only
test-internal dependencies, an external-consumer compile, performance
admission, and cold review. No push, tag, release, deployment, or live service
mutation is part of this feature.
