# Local repair verification evidence

Independent cold Judge pass 11 rejected exact evidence head
`a23389dea35f10f979432b61e276c6e7e382863d`. Source commit
`a3fb596b018069e65562c6a5f434236b9e7b37b4` adds response Read callback/local
loop counts to the response-consume, attempt, and delivery cost contracts. It
changes only comments. The workflow remains `in_progress`; fresh independent
review and a real consumer pin remain open.

Exact source-tip commands and results under Go 1.26.6-X:nodwarf5:

- `make verify` — PASS.
- `go test -mod=readonly -count=1 -race -coverprofile=... ./...` — PASS,
  97.3%.
- Non-comment production source at `f4ffb41`, `a23389d`, and `a3fb596` has one
  identical SHA-256; every changed production line is a `//` comment — PASS.
- Fifty focused race repetitions covering concurrent dispatch/recording plus
  real malformed/header-limit cancellation and deadline precedence, transient
  DNS lookup, content-type expansion, and capped HTTP-date Retry-After — PASS.
- `FuzzParseEndpoint` — PASS, 42,515 executions/5 seconds.
- `FuzzSignedRequestDeterministic` — PASS, 35,046 executions/5 seconds.
- `FuzzCanonicalContentType` — PASS, 28,877 executions/5 seconds.
- Independent OpenSSL HMAC-SHA-256 vector — PASS, exact expected digest.
- Detached clean clone at exact source `a3fb596b018069e65562c6a5f434236b9e7b37b4`
  with fresh empty `GOCACHE`: `make verify` PASS at 97.3%; tracked state clean.
- Synthetic external module: `go test -mod=readonly -count=1 ./...` — PASS.
  This is syntax evidence only, not a consumer contract or compatibility pin.
- Changelog audit — PASS for 26 named historical records, their declared
  two-lineage order, exact Git timestamps, exact file sets, and one permitted
  current-commit placeholder.
- Six retained RFC/IANA authority snapshot hashes — PASS.
- Exact five-regime benchmark observation — PASS; no speedup claim.
  The artifact is retained from behavior-identical source `53872b9`; the
  non-comment source identity proof documents why no rerun is represented.

Artifact SHA-256 values:

- verify: `a68ff9bbb96d8682260a855a93f2b3e7188ffea9702a11582cda97d67f338445`;
- coverage: `6344dbdd4cee8afdd2c498822704d1f5968df7112933d17695e3ab24be05c18f`;
- coverage log: `78e1367c10c435a4ed1fe9755d3c2087eef7aa37ecbf3a06cba240369c99f1c8`;
- coverage function summary: `9a4dcda7aae1dbf3a87622de082c2c168ea31cebab42d9b294a236c801774f82`;
- focused race50: `8bf2a9ea5958e280fa61413a129aa92e40ddde67e111b4cdda8096110dc93349`;
- non-comment identity: `a52776d505ea6eb61ebdb9e55029a90370e5e289cd902ddb561062faa669581f`;
- endpoint fuzz: `84d9ce066b125d2f0357d979030b7febe79b2116047a1267229342997f058366`;
- signing fuzz: `f302c690bfa725a9f1cbcef21385ae6420dcb31bb234c33aa2dc553ac9ac01bc`;
- content-type fuzz: `6ecaae70c3a58f79d0c8dc7e004dea9beb3f1edd0ccb0b38d49d473818dc2f70`;
- independent HMAC: `c23d27896db18640b7451655ae1ff339012f4cc79c8c708eb886a161f3160944`;
- clean clone: `7316b608b25ab894647abc90b478cd74c78ee8227157ea719683f8b3a13e12f5`;
- synthetic consumer: `aa5274081bd4632f6ae349d2dc642cf39d4a5190675621279196803ae597e220`;
- changelog provenance: `a99ecd69683370909516453b31a86d0aa9d0fb27a20791fcfe7b4a8ccd0dbd66`;
- authority hashes: `6895ab208fddb8977864210f09c009bae8fffe4c655689babf8f188713c3787a`;
- benchmark: `207bb0404bf222a23455e21788644d0cd44fc2cb7161d0eaa22e2805ecb61fa8`.

No relevant changed-surface coverage gap remains. Repository-wide uncovered
branches are unchanged and lie outside this repair; the source-level total is
97.3%. Performance evidence remains an uncontrolled-host observation. No
database/recorder adapter is invented here.

The separate admission blocker remains: a concrete consumer must implement the
authorized product contract, verify behavior against its durable receipt and
scheduling boundaries, and pin an exact dependency. No local fixture can fake
that proof.
