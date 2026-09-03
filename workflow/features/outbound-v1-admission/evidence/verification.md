# Local repair verification evidence

Independent cold Judge pass 10 rejected exact evidence head
`f4ffb41c3a23978fcb990be05e60e919e53c8843`. Source commit
`bad5c8171e28666a32f1603205ee0218f8af2e67` repairs only cost-contract comments:
it propagates identifier and HMAC inputs and separates local response-consume
space from delegated body costs. The workflow remains `in_progress`; fresh
independent review and a real consumer pin remain open.

Exact source-tip commands and results under Go 1.26.6-X:nodwarf5:

- `make verify` — PASS.
- `go test -mod=readonly -count=1 -race -coverprofile=... ./...` — PASS,
  97.3%.
- Non-comment production source at `f4ffb41` and `bad5c81` has identical
  SHA-256; every changed production line is a `//` comment — PASS.
- Fifty focused race repetitions covering concurrent dispatch/recording plus
  real malformed/header-limit cancellation and deadline precedence, transient
  DNS lookup, content-type expansion, and capped HTTP-date Retry-After — PASS.
- `FuzzParseEndpoint` — PASS, 34,690 executions/5 seconds.
- `FuzzSignedRequestDeterministic` — PASS, 30,646 executions/5 seconds.
- `FuzzCanonicalContentType` — PASS, 32,182 executions/5 seconds.
- Independent OpenSSL HMAC-SHA-256 vector — PASS, exact expected digest.
- Detached clean clone at exact source `bad5c8171e28666a32f1603205ee0218f8af2e67`
  with fresh empty `GOCACHE`: `make verify` PASS at 97.3%; tracked state clean.
- Synthetic external module: `go test -mod=readonly -count=1 ./...` — PASS.
  This is syntax evidence only, not a consumer contract or compatibility pin.
- Changelog audit — PASS for 24 named historical records, their declared
  two-lineage order, exact Git timestamps, exact file sets, and one permitted
  current-commit placeholder.
- Six retained RFC/IANA authority snapshot hashes — PASS.
- Exact five-regime benchmark observation — PASS; no speedup claim.
  The artifact is retained from behavior-identical source `53872b9`; the
  non-comment source identity proof documents why no rerun is represented.

Artifact SHA-256 values:

- verify: `a68ff9bbb96d8682260a855a93f2b3e7188ffea9702a11582cda97d67f338445`;
- coverage: `9ef1df060a1e131bf6cd126c1317e899ad4f16ef649d150575ff600522eb1cbf`;
- coverage log: `300a91638ab4284a8cdf753607fd577294cf123fc10e25dd8e0dfc7c5ea014ac`;
- coverage function summary: `786edfd36b203526b1ae09871fb7e914539cd1f44383a4265aca7dc1fc77f46e`;
- focused race50: `d8829092b8c6fa7e8d4521b07c4a8b0ab1095355849098f914f1a6b8d8c2ea6e`;
- non-comment identity: `39b9095fb428ac296054bcf1ca77b283d3f11e02e06414ba18a2906b18b279a7`;
- endpoint fuzz: `4b68a8109f69fb5bec63f3c5d12641897cfacd5b87bb108eb1bbd252329da767`;
- signing fuzz: `bff60a2e3bc365721b443c7ffe3f99c8288a471483626277012885de882b2f5c`;
- content-type fuzz: `1fe08b5b51e99a0d1ecd828335047d5a2f289efe0ae9ad28ada917222c729e27`;
- independent HMAC: `c23d27896db18640b7451655ae1ff339012f4cc79c8c708eb886a161f3160944`;
- clean clone: `be092ba97965a958354014deaae4db92ec16cd884c13b0a8f67483c569899eef`;
- synthetic consumer: `11a15052b718207603dc4fcd769e70f74742b0b52418b037ec44d7c91b9131ce`;
- changelog provenance: `f6aa139fce9cded15befd2fe4c9aacc63878a4e4c87e763b11d0f6068eaf5070`;
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
