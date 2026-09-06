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

## Independent cold pass 6 — NARROW AND RETRY

Reviewed evidence head: `3eafee7e28eb806fab9e990c39224ef27a7c1e0c`

The independent Judge proved that attempt-context readiness was not causal
evidence: actual private `net/http` malformed-response and response-header-limit
errors could become retryable if the attempt deadline happened to expire at the
same edge. The source tests relied on obsolete or unrelated exported sentinels.
The Judge also found stale constant-space error-tree accounting and missing
response-body, retry parsing, resolver/dialer, transport, and Recorder terms in
caller contracts. Report: `/tmp/webhooks-judge-6-independent.md`, SHA-256
`3cdfea4a90a5239a73ad8e71794d4fe0b64e626c86f4909f35da764c62d3b265`.

## Repair pass 17 — FINDINGS ADDRESSED / INDEPENDENT RECHECK PENDING

Reviewed source: `53872b9c99d2a7a6d90035ed9564d760c464c298`

Timeout classification now requires deadline or cancellation evidence in the
returned error chain; caller cancellation remains first, typed transient
classes remain retryable, and unknown errors remain permanent. Tests obtain the
real malformed and header-limit errors from `http.Transport` and prove their
permanence against an already-expired attempt context while separately proving
a causal deadline retry. Complexity contracts now aggregate visited error-tree
nodes, join depth, retry parsing/wait, body, transport, resolver/dialer, and
Recorder costs and state that the DNS answer-count guard is post-resolution.
Exact source gates pass at 97.3%, focused race repetitions and all fuzz targets
pass, and the clean-clone and retained-authority checks pass. This is not a
CLEAN claim; a fresh independent review is still required, and consumer pin
admission remains separately open.

## Independent cold pass 10 — NARROW AND RETRY

Reviewed evidence head: `f4ffb41c3a23978fcb990be05e60e919e53c8843`

The independent Judge found one comments-only contract gap. Deliver and attempt
did not propagate `buildRequest`'s HMAC key-processing terms, Deliver's
validation aggregate omitted delivery-ID and event bytes, and
`consumeResponse` called an arbitrary `io.ReadCloser` while claiming constant
whole-function auxiliary space. Report:
`/tmp/webhooks-judge-10-independent.md`, SHA-256
`e3caa86f71774ffbeade2689a7a6b74df42833583cbde750a8f09441b8b3dcfe`.

## Repair pass 18 — FINDING ADDRESSED / INDEPENDENT RECHECK PENDING

Reviewed source: `bad5c8171e28666a32f1603205ee0218f8af2e67`

Deliver's validation aggregate now covers endpoint, content type, delivery ID,
event type, and body bytes. Deliver and attempt propagate secret-key length and
delegated HMAC time/space. `consumeResponse` separates its bounded local loop
from delegated body Read/Close CPU, allocation, I/O, and latency. A mechanical
proof gives identical SHA-256 for all non-test production Go source after
removing full-line comments at the preceding and repaired objects, and rejects
any changed production line that is not a `//` comment. Exact verify/build,
uncached race coverage, focused race50, three fuzz targets, HMAC, synthetic
compile, authority hashes, provenance, and detached source clone pass. This is
not a CLEAN claim; fresh independent review and the consumer pin remain open.

## Independent cold pass 11 — NARROW AND RETRY

Reviewed evidence head: `a23389dea35f10f979432b61e276c6e7e382863d`

The independent Judge found that `consumeResponse` bounded local CPU only by
bytes even though an arbitrary body may return `(0, nil)` repeatedly. The same
local callback-loop work was absent from attempt and Deliver. Report:
`/tmp/webhooks-judge-11-independent.md`, SHA-256
`c54a8c8046fad7d664b859cc5cbeb5e9f2274a9232f746186c0393e69eca5c07`.

## Repair pass 19 — FINDING ADDRESSED / INDEPENDENT RECHECK PENDING

Reviewed source: `a3fb596b018069e65562c6a5f434236b9e7b37b4`

The response contracts now define `c`/`c_i` as body Read callback/local loop
iterations and add those terms to local CPU. Delegated body costs explicitly
cover those Reads and the one Close when a body exists. The prose states that
no finite byte-only CPU bound exists for a body repeatedly returning `(0,
nil)`. Three-revision normalized source hashing and comment-only diff assertions
prove executable production source unchanged from `f4ffb41` through `a23389d`
to this repair. Exact verify/build, uncached race coverage, focused race50,
three fuzz targets, HMAC, synthetic compile, authority hashes, provenance, and
detached source clone pass. This is not a CLEAN claim; fresh independent review
and the consumer pin remain open.

## Fresh post-pass-11 audit - NARROW AND RETRY

Reviewed evidence head: `ef713a777a4949d1f6726f72cabe7a009d667883`

The receipt invariant was still false under Go 1.26.6. The enabled HTTP/2
transport can replay a reconstructible POST inside one `RoundTrip` after
`REFUSED_STREAM`, selected peer protocol errors, or graceful GOAWAY. That
second wire send has no distinct library attempt or receipt. The same audit
found that `io.ReadFull` cost comments omitted repeated `(0, nil)` callbacks
and `waitContext` falsely treated a timer delay as a wall-time upper bound.

## Bounded replay repair - SOURCE CLEAN / ADMISSION OPEN

Reviewed source: `bf64d724bd68bcb95f7180e1b79c16351e1d881c`

The owned transport now explicitly permits HTTP/1 only. A real TLS integration
test uses an HTTP/2-capable server and proves exactly one HTTP/1 request. The
entropy and timer comments now match Go source. Exact Go 1.26.6 development-host
format, vet, race, coverage, race50, focused race50, fuzz, HMAC, external syntax,
benchmark, authority, and clean-clone gates pass. Changed executable coverage
is 100%. This audit does not admit the feature; the real consumer contract/pin
and orchestrator final admission remain open.

## Independent cold pass 14 - NARROW AND RETRY

Reviewed evidence head: `381b680e4b6e0945746d5d2db87b3a60bc797924`

The independent Judge found one high-severity SSRF evidence flaw: the test
copied the same compact IANA prefix table as production, so an allocation
omitted from both transcriptions remained untested. Registry hashes alone did
not connect XML rows to the deny table. The Judge also found that status text
incorrectly coupled standalone technical admission to a real-consumer release
pin. Report: `/tmp/gotth-webhooks-independent-judge-1.md`, SHA-256
`0fed933a862d1310ce09612f2cd7919bbc9848a0bf9f86855e8c950f2148c93c`.

## Repair pass 20 - FINDINGS ADDRESSED / FINAL TECHNICAL ADMISSION OPEN

Reviewed source: `7de792774d2a9229f577d8d46dfb803b1b1c9bfe`

The exact already-hashed IPv4/IPv6 XML snapshots are now repository test
fixtures. A test verifies hashes and registry metadata, parses every allocation
prefix, and proves each is covered by the unchanged compact production deny
table. PRD, README, workflow feature, coverage, and evidence text now separate
standalone technical admission from consumer adoption and release.
`docs/RELEASING.md` remains authoritative: a real-consumer contract,
behavioral validation, and exact dependency pin are hard release/compatibility
gates.

Exact Go 1.26.6 development-host full, verify, race coverage, race50, focused
race50, fuzz, OpenSSL HMAC, external compile, benchmark, and clean-clone gates
pass. This is a worker repair verdict, not final admission. The orchestrator
owns that decision, and workflow state remains `in_progress`.
