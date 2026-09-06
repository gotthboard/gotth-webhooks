# Security Policy

## Supported versions

No release exists, so no version currently carries a long-term security-support
promise. The implementation branch is a review candidate, not a published
security contract. Supported versions will be listed here and in the changelog
after an admitted release.

## Security boundary

Consumers must authorize recipients and minimize opaque payloads before calling
the library. They must keep signing secrets in an appropriate secret manager,
use unique key IDs, rotate keys with a bounded receiver verification window,
coordinate same-ID sends durably, and make concurrency-safe receipt upserts
idempotent while rejecting a delivery ID reused with another fingerprint.

Receivers must verify the complete V1 canonical signature with `hmac.Equal`,
enforce a timestamp window, bind each delivery ID to immutable semantics, and
deduplicate durably before effects. HTTP delivery is not exactly once. Never
treat a timeout as proof that the receiver did nothing.

Endpoint checks allow only canonical HTTPS port 443, validate RFC 3986 raw
queries, reject every address in the pinned IANA special-purpose snapshot and
IPv6 outside allocated `2000::/3`, bypass ambient proxies, and classify every
3xx without redirect processing. The maintained prefix list is not a
replacement for host/network egress controls. Compromised DNS, routing, NAT,
service mesh, kernel, or a public endpoint remains outside the process trust
boundary.

A `Recorder` receives a deadline detached from caller cancellation so a canceled
request does not silently erase evidence. Go contexts are cooperative: a broken
recorder that ignores its context can still block. Recorder implementations are
trusted consumer infrastructure and must honor cancellation/deadlines and be
safe for concurrent calls.

Returned transport and recorder failures are reduced to stable sentinels plus
bounded `Result`/`LastReceipt`; raw errors are not propagated because Go's HTTP
errors can embed the complete endpoint query. The stable delivery fingerprint
is sensitive derived data: it enables correlation and low-entropy guessing, so
receipts and results containing it require restricted access and must not be
logged wholesale. Operators should record diagnostics inside trusted transport/
store boundaries without copying secrets into application logs. A crash
between send and record remains an unknown outcome and requires receiver
deduplication plus consumer reconciliation.

## Reporting a vulnerability

Do not disclose exploit details in a public GitHub issue, pull request,
discussion, or Forgejo ticket. Report vulnerabilities privately through
GitHub's private vulnerability-reporting form:

<https://github.com/gotthboard/gotth-webhooks/security/advisories/new>

Include the affected version or commit, impact, reproduction steps, and any
suggested remediation when available. The maintainers will acknowledge the
report and coordinate validation, remediation, and disclosure through the
private advisory.

Use GitHub Issues only for non-sensitive bugs. If the private form cannot be
used, open a GitHub issue containing no vulnerability details and ask the
maintainers to restore private reporting access.
