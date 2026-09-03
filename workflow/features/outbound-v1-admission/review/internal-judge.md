# Internal cold Judge loop

## Pass 1 — FAIL

Reviewed commit: `338b885077a5870b5645b2e55a257ac487eab638`

Exact blockers:

1. `ErrDestination` policy failures were classified as generic retryable
   transport errors. That wastes attempts on invariant denial and hides the
   SSRF rejection behind `ErrExhausted`.
2. Exhausted transport and recorder errors propagated raw lower-layer text.
   Go `url.Error` includes the complete target URL, so a secret-bearing query
   could leak into logs despite receipt minimization.
3. Receipt prose implied stronger durability than the mechanism supplies. A
   process can crash after HTTP send and before `Record`.
4. `buildRequest` claimed auxiliary `Omega(b+m)` although SHA-256 streams over
   the already-owned body without allocating proportional to body bytes.

Smallest acceptable repairs:

- classify policy denial as permanent and preserve `ErrDestination`; keep true
  DNS/network failure retryable;
- return stable sentinels and bounded result/receipt metadata;
- document crash recovery honestly; and
- correct the cost contract without changing the mechanism.

Userspace/trust: no released userspace exists. Repairs narrow exposure and
retry behavior to the stated contract. No permission, confirmation, channel,
or external-state boundary changes.

Authority: Go `http.Client.Do`/`url.Error` behavior and the committed mechanism
were inspected; no folklore substituted for the contract.

## Pass 2

### Verdict

Accept with constraints (`PASS` for worker handoff; independent admission still
belongs to the orchestrator).

### Exact flaw

The four pass-1 flaws are repaired without expanding the product boundary.
No new correctness, trust, userspace, or cost blocker remains in the reviewed
source and documentation.

### Why it is admissible

- Address-policy failures stop after one attempt and preserve both
  `ErrPermanent` and `ErrDestination`; transient DNS/network failures retry.
- Returned errors cannot carry the endpoint query or raw recorder diagnostic.
  The exact receipt remains available for reconciliation. Later independent
  review correctly identified its fingerprint as sensitive derived data; the
  earlier safety characterization is withdrawn.
- Crash-between-send-and-record is explicit in PRD, architecture, README, and
  security policy; exactly-once is rejected.
- The corrected signature allocation contract matches the streaming hash and
  canonical-buffer mechanism.
- Format, vet, full race, 50 uncached race runs, 97.0% coverage, TLS integration,
  two fuzz targets, external consumer compile, and performance evidence pass.

### Boundary notes

Userspace is new and unreleased. No workflow, confirmation, permission, channel,
or trust-semantics regression exists. HTTPS:443/no-proxy/no-redirect remains a
deliberately narrow V1. Consumers still own authorization, minimization,
same-ID coordination, durable attempt allocation, recorder correctness, and
receiver deduplication.

Documented behavior was checked in Go docs/source and RFC authorities. The
context-broker Judge packet was navigation-only, was truncated at its scan
line bound, and no conclusion relies on it alone.

## Independent orchestrator review — REJECTED, repair pending

Reviewed commit: `a5a3b1060989f220bf97c4733a3248ac4c7e9130`

The orchestrator rejected admission. The decisive product blocker remains: no
real consumer contract validates the public API, and the synthetic external
module proves syntax only. Source findings 2 through 9 were independently
verified and are being repaired, but those repairs cannot turn the candidate
into an admitted or releasable contract. A fresh internal review and exact
repair-source evidence are required before handoff.

## Repair pass 1 — NARROW AND RETRY

Reviewed commit: `8516e191623380c642a26a65fa47d7c812b69c51`

The technical orchestrator findings were repaired, but the cold pass found two
narrow defects: `strings.TrimSpace` accepted non-ASCII whitespace despite the
ASCII Retry-After grammar, and the public Dispatcher concurrency comment did
not state its dependence on a concurrency-safe Recorder. No broader rewrite
was justified.

## Repair pass 2 — TECHNICALLY CLEAN / ADMISSION BLOCKED

Reviewed commit: `b34880e2b1c94eb20528aab9f69b3a669c5f3362`

Both pass-1 defects are fixed and directly tested. Findings 2 through 9 are
CLEAN under exact revision-matched gates. Finding 1 remains an owner/product
blocker: no real consumer contract or pin exists. The synthetic module is only
a syntax fixture. The workflow therefore remains `in_progress`, and no release
or compatibility claim is admissible.

## Repair pass 3 — NARROW AND RETRY

Reviewed commit: `079dddaa9b1c63aa9601bba2b7482e4012a6ecea`

The permanent-failure sentinel still described only HTTP responses, and
transport/response-protocol failures could return misleading `status 0` or
`status 200` text. The smallest repair was to return numeric status only when
status classification caused the failure, without exposing raw transport data.

## Repair pass 4 — SOURCE CLEAN / ADMISSION BLOCKED

Reviewed commit: `f0d4008102380ad135e4a2d32190b8470af0fef8`

The stable permanent sentinel now covers every permanent delivery failure, and
fake transport/protocol status text is gone. Findings 2 through 9 are source-
CLEAN. Exact gates pass at this object. Finding 1 remains the unchanged owner/
product blocker.

## Independent post-repair review — NARROW AND RETRY

Reviewed commit: `8696c6ebeebb0111f070d10d9b98768df066c932`

The external maintainer Judge found one remaining retry-classification escape
hatch and two false complexity contracts. Exhausted address dials were all
marked transient, including deterministic/local and unknown failures;
`canonicalPort` falsely claimed body-sized auxiliary space; and `retryDelay`
falsely claimed work lower-bounded by attempt count. The consumer blocker
remained independently decisive.

## Repair pass 7 — SOURCE CLEAN / ADMISSION BLOCKED

Reviewed commit: `62894967d7f2c8760d816d3f6d0c57928c8cab67`

Exhausted dial errors now use the shared typed allowlist and a documented
permanent-dominates mixed policy, with cancellation preserved and permanent
details redacted. Required failure classes and one-attempt delivery behavior
are directly tested. Both cost contracts now delegate unknown library costs
symbolically and avoid false lower/tight bounds. Exact source gates pass.
Finding 1 remains the unchanged consumer-contract blocker.

## Independent cold pass 3 — NARROW AND RETRY

Reviewed commit: `7c3fc0b18018b41a50b94b0d13952f7dba6b5aa0`

The published wire grammar called the scheme-inclusive signed target an
authority and omitted the literal `https://` bytes. Two retry-wrapper methods
lacked cost contracts, several early-rejection paths retained false
proportional lower/tight bounds, and the changelog attributed evidence files to
the source commit rather than the evidence commit. An orchestrator checkpoint
also proved explicit empty ports were accepted as if omitted.

## Repair pass 10 — SOURCE CLEAN / ADMISSION BLOCKED

Reviewed commit: `7a0a940b10774b66eab3a5badd832181e560226a`

Explicit empty DNS/IPv6 ports now fail before side effects while omitted and
`:443` forms sign identically. Wire grammar and the documented/source golden
vector agree on the absolute HTTPS target bytes. Missing and false cost
contracts were repaired across the early-rejection audit. Exact source gates
pass. Changelog attribution is handled separately in evidence history so it
can name existing commits without self-reference. Finding 1 remains the
unchanged consumer-contract blocker.

## Independent cold pass 4 — NARROW AND RETRY

Reviewed commit: `ff4fee3f5aa94cb5b44694050188793e5b181a8d`

Go's MIME parser trimmed some leading/trailing controls and preserved HTAB in
quoted parameters, violating the stated no-controls contract. Historical
changelog records also assigned the distribution work to the formatting-only
`0bba059` commit and used several non-Git heading times. The consumer blocker
remained independently decisive.

## Repair pass 13 — SOURCE CLEAN / ADMISSION BLOCKED

Reviewed commit: `94f2b7d080d9959af3c1d23a5a4349b555ee240b`

A direct pre-parse scan now rejects all ASCII C0 bytes and DEL. Tests cover
every rejected byte in leading, trailing, and quoted-parameter positions at
the canonicalization, signing-validation, and public Deliver boundaries; no
transport or receipt side effect occurs. Valid spaces and quoted UTF-8 MIME
parameter values remain canonical and signature-stable. Historical changelog
entries are split by actual commit, with every named object's exact file list
and second-resolution CDT author/committer time mechanically matched to Git.
Exact gates, fifty race runs, affected fuzzing, benchmark observation, clean
clone, external syntax fixture, and pass-13 cold review pass. Admission remains
blocked solely on the absent real-consumer contract and pin.

## Independent cold pass 5 — NARROW AND RETRY

Reviewed commit: `c30635434b88718f99ee5125052fdc309e737db9`

The independent Judge found that an expired attempt context overrode observed
destination/TLS/protocol failures, three network cost contracts excluded
variable parsing/error traversal, changelog evidence omitted meaningful commits
and had no declared ordering, and four contract branches lacked direct tests.
The consumer-pin blocker remained separately open.

## Repair pass 16 — FINDINGS ADDRESSED / INDEPENDENT RECHECK PENDING

Reviewed source: `f1fd980e8e6bd184815b871fb5b7513a999725d1`

Caller cancellation remains first; observed destination, certificate,
TLS-record/alert, and HTTP-protocol failures now remain permanent across a
coincident attempt deadline. Network comments include address-byte parsing and
wrapped/joined-error traversal. Tests directly cover every cited gap, and all
touched production functions are at 100% statement coverage. Changelog records
now declare their lineage order, include `8696c6e` and `c306354`, and use the
permitted current-commit placeholder instead of endless attribution commits.
Exact source gates pass at 97.3%. This is not a CLEAN claim: the orchestrator
must run a fresh independent pass and, if clean, an independent double-check.
Admission also still requires a real consumer adapter, behavioral proof, and
dependency pin.
