# Local repair verification evidence

Independent cold Judge pass 5 rejected exact head
`c30635434b88718f99ee5125052fdc309e737db9`. Source commit
`f1fd980e8e6bd184815b871fb5b7513a999725d1` repairs the four technical
findings without changing the consumer-owned product boundary. The workflow
remains `in_progress`; fresh independent review and a real consumer pin remain
open.

Exact source-tip commands and results under Go 1.26.6-X:nodwarf5:

- `make verify` — PASS.
- `go test -mod=readonly -count=1 -race -coverprofile=... ./...` — PASS,
  97.3%. Touched production functions `classifyAttemptFailure`,
  `safeDialer.DialContext`, `canonicalContentType`, and `parseRetryAfter` are
  100% covered.
- Fifty focused race repetitions covering concurrent dispatch/recording plus
  cancellation, deadline/permanent precedence, transient DNS lookup,
  content-type expansion, and capped HTTP-date Retry-After — PASS.
- `FuzzParseEndpoint` — PASS, 99,546 executions/5 seconds.
- `FuzzSignedRequestDeterministic` — PASS, 77,965 executions/5 seconds.
- `FuzzCanonicalContentType` — PASS, 91,802 executions/5 seconds.
- Independent OpenSSL HMAC-SHA-256 vector — PASS, exact expected digest.
- Detached clean clone at exact source `f1fd980e8e6bd184815b871fb5b7513a999725d1`
  with fresh empty `GOCACHE`: `make verify` PASS at 97.3%; tracked state clean.
- Synthetic external module: `go test -mod=readonly -count=1 ./...` — PASS.
  This is syntax evidence only, not a consumer contract or compatibility pin.
- Changelog audit — PASS for 20 named historical records, their declared
  two-lineage order, exact Git timestamps, exact file sets, and one permitted
  current-commit placeholder.
- Six retained RFC/IANA authority snapshot hashes — PASS.
- Exact five-regime benchmark observation — PASS; no speedup claim.

Artifact SHA-256 values:

- verify: `a68ff9bbb96d8682260a855a93f2b3e7188ffea9702a11582cda97d67f338445`;
- coverage: `7bf789422f4a247dc6cf3c57ca5a1d34add685aa235ba5f4da80736c5e37feef`;
- coverage log: `9be78b6a981773b5a46ebccdc10b30bd37892aa7657386386958564b70160929`;
- coverage function summary: `15f005a235c5683685ac58824aa61abf44c453d7b6ad18a64d005ca8697794ea`;
- focused race50: `c470b715acea25d41b9eea1020ce8e1d239ad419712777894062ea5701c39062`;
- endpoint fuzz: `f7434c2c758065553595926ad6a18159bfcc0b25a53d96e58f5d72a29e9fd30a`;
- signing fuzz: `0c9f3d8d74f65663e423fc3170528c96184cf9a794824f59e5d715189222d900`;
- content-type fuzz: `36f4bfb75c7a0d5e2e8d2a3cf892a6eeadc5acddecc3cfb4bc89eb36df0f3bf8`;
- independent HMAC: `c23d27896db18640b7451655ae1ff339012f4cc79c8c708eb886a161f3160944`;
- clean clone: `0b38c6a29eff11b4513f3555acee8b361ae8e95bb07b9ff41bf8e1dc6c9512ab`;
- synthetic consumer: `c7a8c308ddd88137cf0bc359f1fd749625636c35e91a8216ee54b2a139080263`;
- changelog provenance: `93638eff1049dbcc7ef5a4db8dd7726944168d48d6309af41739057549ca0023`;
- authority hashes: `6895ab208fddb8977864210f09c009bae8fffe4c655689babf8f188713c3787a`;
- benchmark: `d45c180a8261f2e7024b7eb7d351e71fe046e76cd5a5311e30c7293fa508d44f`.

No relevant changed-surface coverage gap remains. Repository-wide uncovered
branches are unchanged and lie outside this repair; the source-level total is
97.3%. Performance evidence remains an uncontrolled-host observation. No
database/recorder adapter is invented here.

The separate admission blocker remains: a concrete consumer must implement the
authorized product contract, verify behavior against its durable receipt and
scheduling boundaries, and pin an exact dependency. No local fixture can fake
that proof.
