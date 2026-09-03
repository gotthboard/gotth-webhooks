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

Pending fresh review after repair and verification.
