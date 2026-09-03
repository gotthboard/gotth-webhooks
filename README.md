# gotth-webhooks

Reserved for reusable outbound webhook delivery mechanics shared by GOTTH
applications.

## Intended boundary

This project may eventually own canonical request signing, timestamp and replay
protection, endpoint validation, delivery idempotency, bounded retry policy,
receipt recording, secret rotation, and conformance fixtures. Consumers own
event definitions, subscriptions, authorization, payload minimization, and the
decision that an event may cross the application boundary.

No webhook work begins until a consumer PRD names its events, recipients,
privacy constraints, delivery guarantees, and operator controls.

## Non-goals

- A generic event bus, plugin runtime, workflow engine, or inbound API.
- Automatically exporting application records or confidential fields.
- Unbounded retries or caller-selected internal-network destinations.

## Status

Placeholder only. There is no implementation, API, release, tag, compatibility
promise, or dependency to pin.
