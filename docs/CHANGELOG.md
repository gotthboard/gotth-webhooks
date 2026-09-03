# Changelog

This repository records user-visible and compatibility-relevant changes here.
Released sections use Semantic Versioning; unreleased work remains under
`Unreleased` and does not imply a tag.

## Unreleased

### 2026-09-03 11:08 CDT — Implement bounded outbound webhook mechanics

Commit: `338b885077a5870b5645b2e55a257ac487eab638`

Affected files:

- `LICENSE`, `.go-version`, `go.mod`, `Makefile`, `.gitignore`
- `pkg/webhooks/**`
- `README.md`, `CONTRIBUTING.md`, `SECURITY.md`, `docs/**`
- `workflow.toml`, `workflow.events.jsonl`, `workflow/**`

Explanation:

Replace the placeholder with a local, unreleased Go 1.26.6 admission candidate
for opaque outbound HTTPS webhook mechanics. Add canonical HMAC-SHA-256
signatures, stable delivery identity and semantics fingerprints, public-network
DNS/literal-dial enforcement, no-proxy/no-redirect transport policy, bounded
retry/timeout/response handling, consumer-owned durable receipts, explicit key
IDs for rotation, MIT licensing, lifecycle documents, traceability, and tests.

Verification:

- `make verify`
- race and 97.0% statement coverage
- endpoint/signature fuzz campaigns
- loopback TLS/literal-dial/redirect-denial integration
- external consumer compile
- five-regime in-process performance matrix

Risks / non-goals:

- No exactly-once claim, product event, subscription, payload policy, inbound
  API, database adapter, private endpoint, proxy, redirect, alternate port, or
  custom production transport.
- Candidate remains `in_progress` pending exact-commit verification and
  independent orchestrator review. No push, PR, tag, release, or deployment.

### 2026-09-03 11:20 CDT — Repair cold-review delivery boundaries

Commit: `1defba3673c8e8948501c9f65997b58573d6a844`

Affected files:

- `pkg/webhooks/{dispatcher,network,signing,types}.go`
- focused dispatcher and network tests
- boundary, security, lifecycle, workflow, and evidence documents

Explanation:

Make destination-policy rejection permanent instead of wasting retries, leave
transient DNS lookup failure retryable, suppress raw transport/recorder errors
that can embed secret-bearing endpoint queries, correct the signing auxiliary
space contract, and state the process-crash gap in receipt recording directly.

Verification:

- focused race suite after repairs
- full exact-ref gates repeated after the repair commit

Risks / non-goals:

- Stable error identities and `Result.LastReceipt` preserve reconciliation;
  raw lower-layer diagnostics deliberately remain inside trusted boundaries.
- This does not add pre-send storage, a database adapter, or exactly-once claims.

### 2026-09-03 11:27 CDT — Record local verification handoff

Commit: `a5a3b1060989f220bf97c4733a3248ac4c7e9130`

Affected files:

- performance, verification, workflow evidence, and internal review records

Explanation:

Record exact repair-source coverage, fuzz, repeated race, performance, and
internal Judge results while leaving canonical state `in_progress` for the
orchestrator's independent review.

Verification:

- exact commands, revisions, artifact paths, and hashes in feature evidence

Risks / non-goals:

- Evidence does not create a release or compatibility promise. Independent
  review later rejected this candidate; the rejection is not superseded merely
  by local source repairs.

### 2026-09-03 12:05 CDT — Repair independent-review boundaries

Commit: `f0d4008102380ad135e4a2d32190b8470af0fef8` (source tip; includes repair
batches `8516e191623380c642a26a65fa47d7c812b69c51` and
`b34880e2b1c94eb20528aab9f69b3a669c5f3362`)

Affected files:

- address, query, transport, retry, receipt, and complexity source/tests
- public boundary, security, workflow, and evidence documents

Explanation:

Pin reject-all-special address policy to the 2025-10-09 IANA registries, call
the owned RoundTripper directly so malformed redirects remain classifiable,
use a typed transient retry allowlist with deterministic and unknown failures
permanent, validate raw-query grammar, classify the stable fingerprint as
sensitive, parse Retry-After without overflow, require concurrent Recorders,
and correct test/evidence/cost claims.
Permanent transport and response-protocol failures return the stable
`ErrPermanent` sentinel without fabricated HTTP status text.

Verification:

- exact source-tip format, vet, race, race50, 95.8% coverage, fuzz, fresh-cache
  clean clone, synthetic external compile, benchmark, and internal cold review

Risks / non-goals:

- The public API still lacks a real consumer contract. The feature remains
  `in_progress`, unreleased, unadmitted, and blocked on owner/product input.

### 2026-09-03 12:56:51 CDT — Close post-repair dial classification gaps

Commit: `62894967d7f2c8760d816d3f6d0c57928c8cab67`

Affected files:

- `docs/architecture.md`
- `docs/implementation-spec.md`
- `pkg/webhooks/network.go`
- `pkg/webhooks/network_test.go`
- `pkg/webhooks/retry.go`
- `pkg/webhooks/values.go`

Explanation:

Stop treating every exhausted address dial as retryable. Resolver, dialer, and
dispatcher failures now use one explicit typed transient allowlist; local
resource/configuration and unknown errors are permanent and redacted. A mixed
address set is retryable only when every observed failure is admitted
transient. Correct the successful `canonicalPort` auxiliary-space contract and
the saturation-sensitive `retryDelay` lower/tight bounds.

Verification:

- exact source-tip focused/full race tests, 96.5% coverage, fifty uncached race
  runs, fuzz, fresh-cache clean clone, synthetic external compile, benchmark,
  and cold source review

Risks / non-goals:

- The permanent-dominates mixed policy fails closed and can suppress retries
  when another address had a transient failure; it prevents deterministic
  local failure from being hidden by an alternate address.
- The real-consumer blocker remains open. This does not admit or release the
  API.

### 2026-09-03 13:06:15 CDT — Record post-repair dial evidence

Commit: `7c3fc0b18018b41a50b94b0d13952f7dba6b5aa0`

Affected files:

- `docs/CHANGELOG.md`
- `docs/performance.md`
- `docs/verification.md`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`
- `workflow/features/outbound-v1-admission/review/internal-judge.md`

Explanation:

Record exact-source verification for `6289496`, including coverage, repeated
race, fuzz, performance, clean-clone, external syntax fixture, and source-review
results. This historical entry is added after the evidence commit so its object
ID and actual file scope are recorded without a self-reference placeholder.

Verification:

- exact commands, source revision, artifact paths, and hashes in the evidence
  commit

Risks / non-goals:

- The evidence commit was subsequently rejected for narrower wire, complexity,
  and changelog defects. It did not admit or release the API.

### 2026-09-03 13:32:43 CDT — Tighten endpoint and cost contracts

Commit: `7a0a940b10774b66eab3a5badd832181e560226a`

Affected files:

- `docs/architecture.md`
- `docs/implementation-spec.md`
- `pkg/webhooks/boundary_test.go`
- `pkg/webhooks/dispatcher.go`
- `pkg/webhooks/network.go`
- `pkg/webhooks/signing.go`
- `pkg/webhooks/signing_test.go`
- `pkg/webhooks/values.go`
- `pkg/webhooks/values_test.go`

Explanation:

Reject explicit empty endpoint ports, publish the exact scheme-inclusive signed
target grammar and golden vector, add missing retry-wrapper cost contracts, and
correct all audited early-rejection lower/tight bounds. Omitted ports and
explicit `:443` remain equivalent.

Verification:

- exact source-tip focused/full race tests, 96.5% coverage, fifty-run race,
  fuzz, fresh-cache clean clone, synthetic external compile, benchmark, and
  cold source review

Risks / non-goals:

- This is still an unreleased candidate. The real-consumer contract and pin
  remain absent and independently block admission.

### 2026-09-03 00:42 CDT — Establish GitHub public distribution

Commit: `0bba05927d7922e693a6211ffc41ee3ab91ba451`

Affected files:

- `README.md`
- `CONTRIBUTING.md`
- `SECURITY.md`
- `docs/distribution.md`
- `docs/RELEASING.md`

Explanation:

Declare GitHub as the public distribution endpoint while retaining Forgejo as
canonical development, define maturity and support honestly, and document the
independent release process. The reserved namespace remains a documentation-only placeholder and makes no API or release claim.

Verification:

- exact old-import search
- documentation contract audit

Risks / non-goals:

- No license is selected.
- No existing tag is changed and no new release is created.
- Mirror direction, repository ownership, and account type are unchanged.
