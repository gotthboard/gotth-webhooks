# Local repair verification evidence

Independent cold Judge pass 6 rejected exact evidence head
`3eafee7e28eb806fab9e990c39224ef27a7c1e0c`. Source commit
`53872b9c99d2a7a6d90035ed9564d760c464c298` requires causal timeout errors,
uses real `net/http` transport failures in regressions, and repairs the cost
contracts without changing the consumer-owned product boundary. The workflow
remains `in_progress`; fresh independent review and a real consumer pin remain
open.

Exact source-tip commands and results under Go 1.26.6-X:nodwarf5:

- `make verify` — PASS.
- `go test -mod=readonly -count=1 -race -coverprofile=... ./...` — PASS,
  97.3%. Touched production functions `classifyAttemptFailure`,
  `safeDialer.DialContext`, `canonicalContentType`, and `parseRetryAfter` are
  100% covered.
- Fifty focused race repetitions covering concurrent dispatch/recording plus
  real malformed/header-limit cancellation and deadline precedence, transient DNS lookup,
  content-type expansion, and capped HTTP-date Retry-After — PASS.
- `FuzzParseEndpoint` — PASS, 29,216 executions/5 seconds.
- `FuzzSignedRequestDeterministic` — PASS, 26,549 executions/5 seconds.
- `FuzzCanonicalContentType` — PASS, 36,860 executions/5 seconds.
- Independent OpenSSL HMAC-SHA-256 vector — PASS, exact expected digest.
- Detached clean clone at exact source `53872b9c99d2a7a6d90035ed9564d760c464c298`
  with fresh empty `GOCACHE`: `make verify` PASS at 97.3%; tracked state clean.
- Synthetic external module: `go test -mod=readonly -count=1 ./...` — PASS.
  This is syntax evidence only, not a consumer contract or compatibility pin.
- Changelog audit — PASS for 22 named historical records, their declared
  two-lineage order, exact Git timestamps, exact file sets, and one permitted
  current-commit placeholder.
- Six retained RFC/IANA authority snapshot hashes — PASS.
- Exact five-regime benchmark observation — PASS; no speedup claim.

Artifact SHA-256 values:

- verify: `dc6e46524f43c1cfa012ff0947680570138fb9c0a7a0bc050f6427f3868cbf51`;
- coverage: `ddd101779a1079eed2102e65627b2248a28c3831bc36aae14f4ea39d2cf58d9d`;
- coverage log: `486156377102dd8591eb9c932e60a646023cb7a7b84b246ef549709a268d4085`;
- coverage function summary: `66fe5b801d794154c8e24dec5b7188e25ab0c70419c79a50b529ac7d8fb0bb09`;
- focused race50: `5722702e59f58187f533fedcc18494ff2fc1b1da7c4a85d6dec3bf26b0a5a225`;
- endpoint fuzz: `67937b280acab7a7906e065d751304f15e66cef71031cf5d464fc3680f9c3979`;
- signing fuzz: `7e3f43f47a92dd6c13e8b7a42d59dfd44e833fa8b9f2ae8447b3af0251cbecc0`;
- content-type fuzz: `75250d2b32c22530d12635d8dd6d4b9db4f14fe4c5fd7758f8bd90d635869b35`;
- independent HMAC: `c23d27896db18640b7451655ae1ff339012f4cc79c8c708eb886a161f3160944`;
- clean clone: `b44b0b6fd2bb39fb5c6febe54ff3a9fb2544e58e8df4837082447c4239e0cef7`;
- synthetic consumer: `ecaa98fffadfe75aae8b7ff093bd588751a1a2c05fc17d422491357f812430b6`;
- changelog provenance: `af201a96ad88e26b47a792f568014a0c9a998e2bf21adf536e490eb85808a1c3`;
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
