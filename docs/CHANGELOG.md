# Changelog

This repository records user-visible and compatibility-relevant changes here.
Released sections use Semantic Versioning; unreleased work remains under
`Unreleased` and does not imply a tag.

## Unreleased

Records are grouped by lineage. The active outbound-V1 feature lineage appears
first in ascending Git commit time; the pre-feature distribution lineage
follows, also in ascending Git commit time. A final `current commit` placeholder
belongs to the commit that contains it and avoids an impossible self-hash.

### 2026-09-03 11:15:14 CDT — Implement bounded outbound webhook mechanics

Commit: `338b885077a5870b5645b2e55a257ac487eab638`

Affected files:

- `.gitignore`
- `.go-version`
- `CONTRIBUTING.md`
- `LICENSE`
- `Makefile`
- `README.md`
- `SECURITY.md`
- `docs/CHANGELOG.md`
- `docs/RELEASING.md`
- `docs/architecture.md`
- `docs/distribution.md`
- `docs/feature-plan.md`
- `docs/implementation-spec.md`
- `docs/performance.md`
- `docs/prd.md`
- `docs/runtime-boundary.md`
- `docs/verification.md`
- `go.mod`
- `pkg/webhooks/boundary_test.go`
- `pkg/webhooks/dispatcher.go`
- `pkg/webhooks/dispatcher_test.go`
- `pkg/webhooks/doc.go`
- `pkg/webhooks/errors.go`
- `pkg/webhooks/fuzz_test.go`
- `pkg/webhooks/network.go`
- `pkg/webhooks/network_test.go`
- `pkg/webhooks/performance_test.go`
- `pkg/webhooks/retry.go`
- `pkg/webhooks/signing.go`
- `pkg/webhooks/signing_test.go`
- `pkg/webhooks/transport_integration_test.go`
- `pkg/webhooks/types.go`
- `pkg/webhooks/values.go`
- `pkg/webhooks/values_test.go`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/README.md`
- `workflow/features/outbound-v1-admission/README.md`
- `workflow/features/outbound-v1-admission/evidence/authority.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`

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

### 2026-09-03 11:23:30 CDT — Repair cold-review delivery boundaries

Commit: `1defba3673c8e8948501c9f65997b58573d6a844`

Affected files:

- `README.md`
- `SECURITY.md`
- `docs/CHANGELOG.md`
- `docs/architecture.md`
- `docs/implementation-spec.md`
- `docs/prd.md`
- `pkg/webhooks/boundary_test.go`
- `pkg/webhooks/dispatcher.go`
- `pkg/webhooks/dispatcher_test.go`
- `pkg/webhooks/network.go`
- `pkg/webhooks/signing.go`
- `pkg/webhooks/types.go`
- `workflow.toml`
- `workflow/features/outbound-v1-admission/review/internal-judge.md`

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

### 2026-09-03 11:29:10 CDT — Record local verification handoff

Commit: `a5a3b1060989f220bf97c4733a3248ac4c7e9130`

Affected files:

- `docs/CHANGELOG.md`
- `docs/performance.md`
- `docs/verification.md`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/features/outbound-v1-admission/evidence/authority.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`
- `workflow/features/outbound-v1-admission/review/internal-judge.md`

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

### 2026-09-03 12:01:50 CDT — Repair independent-review boundaries

Commit: `8516e191623380c642a26a65fa47d7c812b69c51`

Affected files:

- `README.md`
- `SECURITY.md`
- `docs/CHANGELOG.md`
- `docs/address-policy.md`
- `docs/architecture.md`
- `docs/implementation-spec.md`
- `docs/prd.md`
- `docs/runtime-boundary.md`
- `docs/verification.md`
- `pkg/webhooks/boundary_test.go`
- `pkg/webhooks/dispatcher.go`
- `pkg/webhooks/dispatcher_test.go`
- `pkg/webhooks/network.go`
- `pkg/webhooks/performance_test.go`
- `pkg/webhooks/retry.go`
- `pkg/webhooks/signing_test.go`
- `pkg/webhooks/transport_integration_test.go`
- `pkg/webhooks/types.go`
- `pkg/webhooks/values.go`
- `pkg/webhooks/values_test.go`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/features/outbound-v1-admission/README.md`
- `workflow/features/outbound-v1-admission/evidence/authority.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`
- `workflow/features/outbound-v1-admission/review/internal-judge.md`

Explanation:

Pin reject-all-special address policy to the 2025-10-09 IANA registries, call
the owned RoundTripper directly so malformed redirects remain classifiable,
use a typed transient retry allowlist with deterministic and unknown failures
permanent, validate raw-query grammar, classify the stable fingerprint as
sensitive, parse Retry-After without overflow, require concurrent Recorders,
and correct test/evidence/cost claims.

Verification:

- exact source-tip format, vet, race, race50, 95.8% coverage, fuzz, fresh-cache
  clean clone, synthetic external compile, benchmark, and internal cold review

Risks / non-goals:

- The public API still lacks a real consumer contract. The feature remains
  `in_progress`, unreleased, unadmitted, and blocked on owner/product input.

### 2026-09-03 12:10:21 CDT — Tighten ASCII and concurrency contracts

Commit: `b34880e2b1c94eb20528aab9f69b3a669c5f3362`

Affected files:

- `pkg/webhooks/boundary_test.go`
- `pkg/webhooks/dispatcher.go`
- `pkg/webhooks/retry.go`

Explanation:

Restrict Retry-After optional-whitespace trimming to ASCII SP/HTAB, add the
non-ASCII rejection case, and qualify Dispatcher concurrency on the Recorder
fulfilling its existing concurrency contract.

Verification:

- focused and full race tests

Risks / non-goals:

- This does not widen accepted header syntax or provide a Recorder adapter.

### 2026-09-03 12:16:46 CDT — Record independent-boundary repair evidence

Commit: `079dddaa9b1c63aa9601bba2b7482e4012a6ecea`

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

Record the exact repair verification, performance, workflow, and internal
review evidence without changing the public implementation.

Verification:

- exact commands, revisions, artifact paths, and hashes in feature evidence

Risks / non-goals:

- Evidence does not admit or release the API.

### 2026-09-03 12:19:39 CDT — Avoid false webhook status errors

Commit: `f0d4008102380ad135e4a2d32190b8470af0fef8`

Affected files:

- `docs/prd.md`
- `pkg/webhooks/dispatcher.go`
- `pkg/webhooks/dispatcher_test.go`
- `pkg/webhooks/errors.go`

Explanation:

Return the stable `ErrPermanent` sentinel for permanent transport and
response-protocol failures without fabricating HTTP status text.

Verification:

- focused and full exact-source tests

Risks / non-goals:

- Stable public classification deliberately omits raw lower-layer diagnostics.

### 2026-09-03 12:31:17 CDT — Record final webhook repair evidence

Commit: `8696c6ebeebb0111f070d10d9b98768df066c932`

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

Record the verified failure-boundary repair state, retained performance and
verification artifacts, workflow status, coverage posture, and internal-review
result while leaving admission blocked on a real consumer.

Verification:

- exact commands, source revision, artifact paths, hashes, and review result in
  the evidence commit

Risks / non-goals:

- Evidence did not admit or release the API and was later superseded by narrower
  source and evidence repairs.

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

### 2026-09-03 13:41:32 CDT — Record endpoint and cost repair evidence

Commit: `6155c63e370ff70a3fb11bf97b3bb58949ea2b0e`

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

Record exact-source verification for `7a0a940`, correct the historical scopes
of `6289496` and `7c3fc0b`, retain the rejected-review provenance, and keep the
consumer blocker explicit. This entry is added by a later changelog-only audit
commit so it can name the evidence object without a placeholder or false
self-reference.

Verification:

- exact source revision, commands, artifact paths, and hashes in the evidence
  commit

Risks / non-goals:

- Recording evidence does not admit, release, push, tag, deploy, or establish
  compatibility for the candidate.

### 2026-09-03 13:42:04 CDT — Add exact evidence-commit attribution

Commit: `ff4fee3f5aa94cb5b44694050188793e5b181a8d`

Affected files:

- `docs/CHANGELOG.md`

Explanation:

Add the exact `6155c63e370ff70a3fb11bf97b3bb58949ea2b0e` evidence-commit
record after that object existed, avoiding a placeholder or self-reference.

Verification:

- exact `git show --name-only --format=fuller` attribution audit

Risks / non-goals:

- Changelog attribution does not alter implementation or admission state.

### 2026-09-03 14:03:37 CDT — Reject controls in webhook content types

Commit: `be6f7150f3419ff3bcb971476119e7cc3f2cb10b`

Affected files:

- `docs/implementation-spec.md`
- `pkg/webhooks/boundary_test.go`
- `pkg/webhooks/fuzz_test.go`
- `pkg/webhooks/signing_test.go`
- `pkg/webhooks/values.go`
- `pkg/webhooks/values_test.go`

Explanation:

Reject every ASCII C0 control byte and DEL before MIME parsing. Exercise each
byte in leading, trailing, and quoted-parameter positions through direct
canonicalization, signing validation, and the public Config/Message/Deliver
path, proving validation precedes transport and receipt recording. Preserve
accepted spaces and MIME parser support for quoted UTF-8 parameter values.

Verification:

- focused and full race tests, coverage, content-type/signature fuzzing, and
  exact-source clean-clone verification

Risks / non-goals:

- Non-ASCII policy outside the rejected ASCII control set is unchanged.
- The real-consumer blocker remains open; this does not admit or release the
  API.

### 2026-09-03 14:07:09 CDT — Correct webhook changelog provenance

Commit: `94f2b7d080d9959af3c1d23a5a4349b555ee240b`

Affected files:

- `docs/CHANGELOG.md`

Explanation:

Replace broad and grouped historical records with one exact record per Git
object. Match every named commit's affected-file list and second-resolution
CDT author/committer time, including the substantive distribution commits and
their formatting-only follow-up.

Verification:

- exact file-list and heading-time comparison against Git for all 16 named
  commits

Risks / non-goals:

- This repairs evidence provenance only; it does not rewrite Git history or
  change implementation, admission, or release state.

### 2026-09-03 14:17:24 CDT — Record content-type repair evidence

Commit: `905b4ddd1f7e6bcba5a8e25e7313eff26e886dd0`

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

Record exact-candidate coverage, race, fuzz, performance, clean-clone,
provenance-audit, external syntax, and internal review evidence for
`94f2b7d080d9959af3c1d23a5a4349b555ee240b`. Preserve the real-consumer
blocker and unreleased workflow state.

Verification:

- exact commands, revisions, artifact paths, hashes, and review result in the
  evidence commit

Risks / non-goals:

- Evidence does not admit, release, push, tag, deploy, or establish
  compatibility for the candidate.

### 2026-09-03 14:17:51 CDT — Audit content-type evidence attribution

Commit: `c30635434b88718f99ee5125052fdc309e737db9`

Affected files:

- `docs/CHANGELOG.md`

Explanation:

Add the exact record for the preceding content-type evidence commit. This
historical self-attribution pattern is replaced by the policy-permitted current
commit placeholder below so future evidence does not require endless
changelog-only follow-ups.

Verification:

- exact `git show --name-only --format=fuller` attribution audit

Risks / non-goals:

- Changelog attribution does not alter implementation or admission state.

### 2026-09-03 15:01:34 CDT — Repair deadline precedence and coverage gaps

Commit: `f1fd980e8e6bd184815b871fb5b7513a999725d1`

Affected files:

- `docs/CHANGELOG.md`
- `docs/architecture.md`
- `docs/implementation-spec.md`
- `pkg/webhooks/boundary_test.go`
- `pkg/webhooks/dispatcher.go`
- `pkg/webhooks/dispatcher_test.go`
- `pkg/webhooks/network.go`
- `pkg/webhooks/network_test.go`

Explanation:

Preserve caller cancellation first, then keep observed destination,
certificate, TLS-record/alert, and HTTP-protocol failures permanent when the
attempt deadline expires at the same edge. State address parsing and wrapped
error traversal costs honestly. Add direct regressions for deadline/permanent
precedence, transient DNS lookup classification, canonical MIME expansion past
the output bound, and capped HTTP-date Retry-After handling. Document the
changelog lineage order and replace self-referential attribution churn with the
permitted current-commit placeholder.

Verification:

- focused race tests for the repaired and newly covered boundaries
- full repository verification and retained evidence are recorded separately

Risks / non-goals:

- Unknown transport errors remain permanent unless the attempt deadline is the
  only observed typed classification; no retry allowlist is widened.
- The real-consumer contract and dependency pin remain separate admission
  blockers. No push, tag, release, deployment, or live request is performed.

### 2026-09-03 15:11:32 CDT — Record deadline-precedence repair evidence

Commit: `3eafee7e28eb806fab9e990c39224ef27a7c1e0c`

Affected files:

- `docs/CHANGELOG.md`
- `docs/performance.md`
- `docs/verification.md`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/features/outbound-v1-admission/evidence/authority.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`
- `workflow/features/outbound-v1-admission/review/internal-judge.md`

Explanation:

Record exact-source race, coverage, fuzz, independent HMAC, benchmark,
clean-clone, synthetic-consumer, authority-hash, and changelog-provenance
evidence for `f1fd980e8e6bd184815b871fb5b7513a999725d1`. Update the strict coverage map
with every formerly missing branch and retain `in_progress` state for fresh
independent review and the separate real-consumer pin.

Verification:

- exact commands, revision, results, paths, and SHA-256 values in feature
  evidence

Risks / non-goals:

- This evidence does not claim a clean independent review, admission, release,
  compatibility, or consumer behavior.
- No push, PR, tag, deployment, live request, or remote mutation is performed.

### 2026-09-03 15:45:00 CDT — Require causal deadline classification

Commit: `53872b9c99d2a7a6d90035ed9564d760c464c298`

Affected files:

- `docs/CHANGELOG.md`
- `docs/architecture.md`
- `docs/implementation-spec.md`
- `pkg/webhooks/dispatcher.go`
- `pkg/webhooks/dispatcher_test.go`
- `pkg/webhooks/network.go`

Explanation:

Classify an attempt timeout only when the returned transport error chain carries
deadline or cancellation evidence. Keep real malformed-response and
response-header-limit failures permanent even if the attempt context happens to
expire at the same edge. Replace synthetic transport sentinels in the deadline
precedence regressions with errors captured from the production `net/http`
transport. Account explicitly for joined-error traversal, response-body,
resolver, dialer, retry parsing, and Recorder costs.

Verification:

- focused real-transport deadline-precedence race tests
- full source and retained-evidence gates recorded separately

Risks / non-goals:

- Causal deadline errors remain retryable; typed transient dial and transport
  failures retain their existing allowlist behavior.
- The real-consumer contract and dependency pin remain separate admission
  blockers. No push, tag, release, deployment, or live request is performed.

### 2026-09-03 16:05:00 CDT — Record causal-deadline repair evidence

Commit: `f4ffb41c3a23978fcb990be05e60e919e53c8843`

Affected files:

- `docs/CHANGELOG.md`
- `docs/performance.md`
- `docs/verification.md`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/features/outbound-v1-admission/evidence/authority.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`
- `workflow/features/outbound-v1-admission/review/internal-judge.md`

Explanation:

Record exact-source race, coverage, real-transport focused tests, fuzz,
independent HMAC, benchmark, clean-clone, synthetic-consumer, authority-hash,
and changelog-provenance evidence for
`53872b9c99d2a7a6d90035ed9564d760c464c298`. Preserve `in_progress` state for
fresh independent review and the separate real-consumer dependency pin.

Verification:

- exact commands, revision, results, paths, and SHA-256 values in feature
  evidence

Risks / non-goals:

- This evidence does not claim a clean independent review, admission, release,
  compatibility, or consumer behavior.
- No push, PR, tag, deployment, live request, or remote mutation is performed.

### 2026-09-03 16:20:00 CDT — Complete webhook cost aggregation contracts

Commit: `bad5c8171e28666a32f1603205ee0218f8af2e67`

Affected files:

- `docs/CHANGELOG.md`
- `pkg/webhooks/dispatcher.go`
- `pkg/webhooks/retry.go`

Explanation:

Carry complete message-identifier validation and delegated HMAC key-processing
costs through the attempt and delivery aggregates. Separate
`consumeResponse`'s constant local working space from arbitrary delegated
response-body Read/Close CPU, allocation, I/O, and latency. This is a
comments-only production-source repair; executable behavior is unchanged.

Verification:

- non-comment production-source identity proof
- full source and retained-evidence gates recorded separately

Risks / non-goals:

- No API, runtime, retry, transport, receipt, or allocation behavior changes.
- The real-consumer contract and dependency pin remain separate admission
  blockers. No push, tag, release, deployment, or live request is performed.

### 2026-09-03 16:40:00 CDT — Record comments-only contract evidence

Commit: `a23389dea35f10f979432b61e276c6e7e382863d`

Affected files:

- `docs/CHANGELOG.md`
- `docs/performance.md`
- `docs/verification.md`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/features/outbound-v1-admission/evidence/authority.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`
- `workflow/features/outbound-v1-admission/review/internal-judge.md`

Explanation:

Record exact-source non-comment identity, race, coverage, focused, fuzz, HMAC,
clean-clone, synthetic-consumer, authority-hash, and provenance evidence for
the comments-only source `bad5c8171e28666a32f1603205ee0218f8af2e67`.
Preserve `in_progress` state for fresh independent review and the separate
real-consumer dependency pin.

Verification:

- exact commands, revision, results, paths, and SHA-256 values in feature
  evidence

Risks / non-goals:

- This evidence does not claim a clean independent review, admission, release,
  compatibility, or consumer behavior.
- No push, PR, tag, deployment, live request, or remote mutation is performed.

### 2026-09-03 16:50:00 CDT — Account for zero-progress body reads

Commit: `a3fb596b018069e65562c6a5f434236b9e7b37b4`

Affected files:

- `docs/CHANGELOG.md`
- `pkg/webhooks/dispatcher.go`
- `pkg/webhooks/retry.go`

Explanation:

Add response-body Read callback/local loop iteration counts to
`consumeResponse`, attempt, and Deliver cost contracts. This accounts for a
caller body that repeatedly returns `(0, nil)` without advancing the byte
limit, while retaining explicit delegated Read/Close CPU, allocation, I/O, and
latency costs. This is comments-only; executable behavior is unchanged.

Verification:

- non-comment production-source identity proof
- full source and retained-evidence gates recorded separately

Risks / non-goals:

- No new runtime no-progress policy is introduced and no API or behavior
  changes.
- The real-consumer contract and dependency pin remain separate admission
  blockers. No push, tag, release, deployment, or live request is performed.

### 2026-09-03 17:10:00 CDT — Record response-callback contract evidence

Commit: `ef713a777a4949d1f6726f72cabe7a009d667883`

Affected files:

- `docs/CHANGELOG.md`
- `docs/performance.md`
- `docs/verification.md`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/features/outbound-v1-admission/evidence/authority.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`
- `workflow/features/outbound-v1-admission/review/internal-judge.md`

Explanation:

Record exact-source three-revision non-comment identity, race, coverage,
focused, fuzz, HMAC, clean-clone, synthetic-consumer, authority-hash, and
provenance evidence for comments-only source
`a3fb596b018069e65562c6a5f434236b9e7b37b4`. Preserve `in_progress` state for
fresh independent review and the separate real-consumer dependency pin.

Verification:

- exact commands, revision, results, paths, and SHA-256 values in feature
  evidence

Risks / non-goals:

- This evidence does not claim a clean independent review, admission, release,
  compatibility, or consumer behavior.
- No push, PR, tag, deployment, live request, or remote mutation is performed.

### 2026-09-05 22:00:09 CDT — Prevent unrecorded transport replay

Commit: `bf64d724bd68bcb95f7180e1b79c16351e1d881c`

Affected files:

- `README.md`
- `docs/CHANGELOG.md`
- `docs/architecture.md`
- `docs/implementation-spec.md`
- `docs/runtime-boundary.md`
- `pkg/webhooks/boundary_test.go`
- `pkg/webhooks/network.go`
- `pkg/webhooks/retry.go`
- `pkg/webhooks/transport_integration_test.go`
- `pkg/webhooks/values.go`

Explanation:

Pin the owned transport to HTTP/1 so Go 1.26.6 cannot replay a webhook POST
inside one `RoundTrip` without a distinct durable receipt. Prove HTTP/1 is
negotiated against a TLS server that advertises HTTP/2. Correct the remaining
zero-progress entropy-reader and timer-latency cost contracts found by the
same fresh audit.

Verification:

- focused transport-policy and TLS protocol integration tests
- exact-commit repository, race, coverage, fuzz, and clean-clone gates are
  recorded separately

Risks / non-goals:

- V1 no longer negotiates HTTP/2; protocol breadth is deliberately traded for
  observable one-receipt-per-send behavior.
- This repair does not admit or release the API. The real-consumer contract,
  dependency pin, and orchestrator final admission remain open.
- No push, PR, tag, deployment, live request, or remote mutation is performed.

### 2026-09-05 22:20:00 CDT — Record transport replay repair evidence

Commit: current commit; hash assigned by Git after commit

Affected files:

- `docs/CHANGELOG.md`
- `docs/performance.md`
- `docs/verification.md`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/features/outbound-v1-admission/evidence/authority.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`
- `workflow/features/outbound-v1-admission/review/internal-judge.md`

Explanation:

Record the fresh severe audit findings and exact Go 1.26.6 development-host
verification for source `bf64d724bd68bcb95f7180e1b79c16351e1d881c`.
Preserve `in_progress` state and the separate consumer/admission blockers.

Verification:

- exact format, vet, race, coverage, race50, focused race50, fuzz, HMAC,
  external syntax, benchmark, authority, provenance, and clean-clone evidence

Risks / non-goals:

- Source-clean worker evidence is not orchestrator final admission.
- No push, PR, tag, release, deployment, live request, remote mutation, or
  GOTTH Board mutation is performed.

### 2026-09-05 22:38:50 CDT - Test IANA deny table against pinned registries

Commit: `7de792774d2a9229f577d8d46dfb803b1b1c9bfe`

Affected files:

- `docs/address-policy.md`
- `pkg/webhooks/testdata/iana-ipv4-special-registry.xml`
- `pkg/webhooks/testdata/iana-ipv6-special-registry.xml`
- `pkg/webhooks/values_test.go`

Explanation:

Retain the exact already-hashed 2025-10-09 IANA IPv4 and IPv6 registry XML
snapshots as test fixtures. Verify fixture hashes and registry metadata, parse
all registry prefixes, and require every allocation to be covered by the
unchanged compact production deny table. This supplies an independent
completeness oracle without adding a test-time or runtime fetch.

Verification:

- focused local fixture/hash/registry coverage tests with `GOMAXPROCS=2` and
  `-p=1`
- exact fixture byte comparison and SHA-256 checks
- exact-source development-host gates recorded in the following evidence entry

Risks / non-goals:

- Production code and the compact runtime prefix table are unchanged.
- The snapshots remain pinned and require deliberate maintenance before a
  release or security-policy refresh.

### 2026-09-05 22:50:00 CDT - Record registry and admission-gate repair

Commit: current commit; hash assigned by Git after commit

Affected files:

- `README.md`
- `docs/CHANGELOG.md`
- `docs/distribution.md`
- `docs/performance.md`
- `docs/prd.md`
- `docs/RELEASING.md`
- `docs/verification.md`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/features/outbound-v1-admission/README.md`
- `workflow/features/outbound-v1-admission/evidence/authority.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`
- `workflow/features/outbound-v1-admission/review/internal-judge.md`

Explanation:

Record the independent-review findings and exact Go 1.26.6 development-host
verification for source `7de792774d2a9229f577d8d46dfb803b1b1c9bfe`.
Separate standalone technical implementation admission from real-consumer
adoption and release. Preserve the real-consumer contract, behavioral
validation, and exact dependency pin as hard release/compatibility gates under
`docs/RELEASING.md`.

Verification:

- uncached full, format, vet, race, 97.3% coverage, full/focused race50, three
  fuzz targets, OpenSSL HMAC, external syntax, benchmark, and clean-clone gates
- every retained gate log records exact before/after HEAD and status

Risks / non-goals:

- Workflow state remains `in_progress`; final technical admission is
  orchestrator-owned.
- No release or compatibility promise is made. No push, merge, PR, tag,
  release, deployment, live request, or GOTTH Board mutation is performed.

### 2026-09-05 23:09:12 CDT - Admit standalone webhook implementation

Commit: current commit; hash assigned by Git after commit

Affected files:

- `README.md`
- `docs/CHANGELOG.md`
- `docs/verification.md`
- `workflow.events.jsonl`
- `workflow.toml`
- `workflow/COVERAGE.md`
- `workflow/features/outbound-v1-admission/README.md`
- `workflow/features/outbound-v1-admission/evidence/verification.md`
- `workflow/features/outbound-v1-admission/review/independent-12.md`
- `workflow/features/outbound-v1-admission/review/independent-13.md`

Explanation:

Admit the standalone technical implementation after two fresh independent
CLEAN reviews of exact evidence head `92c3de2`. Preserve real-consumer contract,
behavioral validation, exact dependency pin, release verification, and explicit
release authorization as separate hard release/compatibility gates.

Verification:

- exact-source Go 1.26.6 development-host matrix recorded in feature evidence
- independent CLEAN review and independent CLEAN double-check
- final documentation diff and repository cleanliness checks

Risks / non-goals:

- Technical admission does not create a tag, release, or compatibility promise.
- No push, merge, PR, tag, release, deployment, live request, or GOTTH Board
  mutation is performed by this admission commit.

### 2026-09-03 00:52:57 CDT — Establish GitHub public distribution

Commit: `9bb9a46e35e5b9de70ed17507f233f6bdd9fc0d4`

Affected files:

- `CONTRIBUTING.md`
- `README.md`
- `SECURITY.md`
- `docs/CHANGELOG.md`
- `docs/RELEASING.md`
- `docs/distribution.md`

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

### 2026-09-03 01:05:08 CDT — Tighten distribution and release claims

Commit: `3cb9334b09b020a98a2cae103b65b2a3fd79f95e`

Affected files:

- `README.md`
- `docs/RELEASING.md`
- `docs/distribution.md`

Explanation:

Clarify the placeholder's maturity and separate GitHub distribution from the
canonical Forgejo development repository without claiming a release.

Verification:

- documentation contract audit

Risks / non-goals:

- No release, mirror-direction, ownership, or account-type change is made.

### 2026-09-03 01:07:20 CDT — Normalize release-policy formatting

Commit: `0bba05927d7922e693a6211ffc41ee3ab91ba451`

Affected files:

- `docs/RELEASING.md`

Explanation:

Normalize policy formatting only. The substantive distribution contract is
owned by `9bb9a46e35e5b9de70ed17507f233f6bdd9fc0d4` with the narrower claim repair
in `3cb9334b09b020a98a2cae103b65b2a3fd79f95e`.

Verification:

- exact `git show --name-only --format=fuller` attribution audit

Risks / non-goals:

- Formatting does not establish or modify distribution policy.
