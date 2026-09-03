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
- return stable sentinels and non-sensitive result/receipt metadata;
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
  The exact non-sensitive receipt remains available for reconciliation.
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
