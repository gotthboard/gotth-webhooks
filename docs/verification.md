# Verification status

Independent cold Judge pass 10 rejected exact evidence head
`f4ffb41c3a23978fcb990be05e60e919e53c8843` because caller cost contracts
omitted message identifiers and delegated HMAC work, while `consumeResponse`
claimed constant whole-function space despite an arbitrary caller body. Source
commit `bad5c8171e28666a32f1603205ee0218f8af2e67` repairs only comments. A fresh
independent review remains required; this record does not claim admission.

Exact source-tip evidence under Go 1.26.6-X:nodwarf5, Linux amd64:

- `make verify` — PASS.
- Uncached `go test -mod=readonly -count=1 -race -coverprofile=... ./...` —
  PASS, 97.3% statements; production logic is unchanged by this comments-only
  repair.
- Non-comment production source before and after the repair has identical
  SHA-256, and every changed production line is a `//` comment — PASS.
- Focused real-transport cancellation/deadline/permanent-error, transient-DNS,
  MIME-expansion, capped-HTTP-date, and concurrency race suite repeated fifty
  times — PASS.
- Endpoint fuzz — PASS, 34,690 executions/5 seconds.
- Signing fuzz — PASS, 30,646 executions/5 seconds.
- Content-type fuzz — PASS, 32,182 executions/5 seconds.
- Independent OpenSSL HMAC-SHA-256 computation — PASS, exact vector
  `5afc952ae607736b86e3570dddc25991115f80afdb46c49832a30a5671028f7f`.
- Detached clone at exact source `bad5c8171e28666a32f1603205ee0218f8af2e67`
  with a fresh empty `GOCACHE` ran
  `make verify` — PASS, clean detached worktree, 97.3% statements.
- Synthetic external module `go test -mod=readonly -count=1 ./...` — PASS.
  This proves public syntax only, not consumer behavior or compatibility.
- The 24 historical changelog records at the source object match the declared
  two-lineage order, exact Git timestamps, and exact affected-file sets; one
  current-commit placeholder is present as policy requires.
- All six retained RFC/IANA authority snapshots revalidated against their
  recorded SHA-256 values.
- The exact performance matrix and limitations are in `performance.md`; no
  speedup or production-latency claim is made.

Exact source artifacts and SHA-256 values follow. The final benchmark artifact
is retained from behavior-identical source `53872b9`; the non-comment identity
proof above makes that reuse explicit rather than pretending it was rerun:

- `/tmp/gotth-webhooks-bad5c81.verify.log` —
  `a68ff9bbb96d8682260a855a93f2b3e7188ffea9702a11582cda97d67f338445`.
- `/tmp/gotth-webhooks-bad5c81.coverage.out` —
  `9ef1df060a1e131bf6cd126c1317e899ad4f16ef649d150575ff600522eb1cbf`.
- `/tmp/gotth-webhooks-bad5c81.coverage.log` —
  `300a91638ab4284a8cdf753607fd577294cf123fc10e25dd8e0dfc7c5ea014ac`.
- `/tmp/gotth-webhooks-bad5c81.coverage.func` —
  `786edfd36b203526b1ae09871fb7e914539cd1f44383a4265aca7dc1fc77f46e`.
- `/tmp/gotth-webhooks-bad5c81.focused-race50.log` —
  `d8829092b8c6fa7e8d4521b07c4a8b0ab1095355849098f914f1a6b8d8c2ea6e`.
- `/tmp/gotth-webhooks-bad5c81.noncomment-identity.log` —
  `39b9095fb428ac296054bcf1ca77b283d3f11e02e06414ba18a2906b18b279a7`.
- `/tmp/gotth-webhooks-bad5c81.fuzz-endpoint.log` —
  `4b68a8109f69fb5bec63f3c5d12641897cfacd5b87bb108eb1bbd252329da767`.
- `/tmp/gotth-webhooks-bad5c81.fuzz-signing.log` —
  `bff60a2e3bc365721b443c7ffe3f99c8288a471483626277012885de882b2f5c`.
- `/tmp/gotth-webhooks-bad5c81.fuzz-content-type.log` —
  `1fe08b5b51e99a0d1ecd828335047d5a2f289efe0ae9ad28ada917222c729e27`.
- `/tmp/gotth-webhooks-bad5c81.hmac-openssl.log` —
  `c23d27896db18640b7451655ae1ff339012f4cc79c8c708eb886a161f3160944`.
- `/tmp/gotth-webhooks-bad5c81.clean-clone.log` —
  `be092ba97965a958354014deaae4db92ec16cd884c13b0a8f67483c569899eef`.
- `/tmp/gotth-webhooks-bad5c81.synthetic-consumer.log` —
  `11a15052b718207603dc4fcd769e70f74742b0b52418b037ec44d7c91b9131ce`.
- `/tmp/gotth-webhooks-bad5c81.provenance.log` —
  `f6aa139fce9cded15befd2fe4c9aacc63878a4e4c87e763b11d0f6068eaf5070`.
- `/tmp/gotth-webhooks-bad5c81.authority-hashes.log` —
  `6895ab208fddb8977864210f09c009bae8fffe4c655689babf8f188713c3787a`.
- `/tmp/gotth-webhooks-53872b9.benchmark.txt` —
  `207bb0404bf222a23455e21788644d0cd44fc2cb7161d0eaa22e2805ecb61fa8`.

The feature remains `in_progress` and unreleased. The newly authorized product
contract work is outside this source repair, and there is still no real
consumer adapter, behavioral proof, dependency pin, tag, or compatibility
promise. No push, PR, tag, release, deployment, or live request occurred.

PostgreSQL/database integration is N/A because this package has no adapter.
Graphify is N/A because this repair changes only adjacent comments in one
direct call chain and has no dependency ambiguity worth graph construction.
