# gotth-webhooks

> **Distribution:** GitHub is the public clone, and future release endpoint.
> Forgejo remains canonical development and the issue/contribution location.
> See [the distribution contract](docs/distribution.md).


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

## Installation, compatibility, and support

Planned placeholder only. There is no implementation, API, support promise, or release.

There is nothing to install or import. Do not add this repository as a
dependency.

The repository has no selected license and no long-term support promise.
Versioning, release admission, security reporting, and contribution details are
in [the release policy](docs/RELEASING.md), [security policy](SECURITY.md), and
[contribution guide](CONTRIBUTING.md).
