# Verification status

Independent cold Judge pass 11 rejected exact evidence head
`a23389dea35f10f979432b61e276c6e7e382863d` because response-consumption cost
contracts omitted local loop iterations when a body repeatedly returns
`(0, nil)`. Source commit `a3fb596b018069e65562c6a5f434236b9e7b37b4`
repairs only comments. A fresh independent review remains required; this record
does not claim admission.

Exact source-tip evidence under Go 1.26.6-X:nodwarf5, Linux amd64:

- `make verify` — PASS.
- Uncached `go test -mod=readonly -count=1 -race -coverprofile=... ./...` —
  PASS, 97.3% statements; production logic is unchanged by this comments-only
  repair.
- Non-comment production source at `f4ffb41`, `a23389d`, and `a3fb596` has one
  identical SHA-256, and every changed production line is a `//` comment — PASS.
- Focused real-transport cancellation/deadline/permanent-error, transient-DNS,
  MIME-expansion, capped-HTTP-date, and concurrency race suite repeated fifty
  times — PASS.
- Endpoint fuzz — PASS, 42,515 executions/5 seconds.
- Signing fuzz — PASS, 35,046 executions/5 seconds.
- Content-type fuzz — PASS, 28,877 executions/5 seconds.
- Independent OpenSSL HMAC-SHA-256 computation — PASS, exact vector
  `5afc952ae607736b86e3570dddc25991115f80afdb46c49832a30a5671028f7f`.
- Detached clone at exact source `a3fb596b018069e65562c6a5f434236b9e7b37b4`
  with a fresh empty `GOCACHE` ran
  `make verify` — PASS, clean detached worktree, 97.3% statements.
- Synthetic external module `go test -mod=readonly -count=1 ./...` — PASS.
  This proves public syntax only, not consumer behavior or compatibility.
- The 26 historical changelog records at the source object match the declared
  two-lineage order, exact Git timestamps, and exact affected-file sets; one
  current-commit placeholder is present as policy requires.
- All six retained RFC/IANA authority snapshots revalidated against their
  recorded SHA-256 values.
- The exact performance matrix and limitations are in `performance.md`; no
  speedup or production-latency claim is made.

Exact source artifacts and SHA-256 values follow. The final benchmark artifact
is retained from behavior-identical source `53872b9`; the three-object
non-comment identity proof above makes that reuse explicit rather than
pretending it was rerun:

- `/tmp/gotth-webhooks-a3fb596.verify.log` —
  `a68ff9bbb96d8682260a855a93f2b3e7188ffea9702a11582cda97d67f338445`.
- `/tmp/gotth-webhooks-a3fb596.coverage.out` —
  `6344dbdd4cee8afdd2c498822704d1f5968df7112933d17695e3ab24be05c18f`.
- `/tmp/gotth-webhooks-a3fb596.coverage.log` —
  `78e1367c10c435a4ed1fe9755d3c2087eef7aa37ecbf3a06cba240369c99f1c8`.
- `/tmp/gotth-webhooks-a3fb596.coverage.func` —
  `9a4dcda7aae1dbf3a87622de082c2c168ea31cebab42d9b294a236c801774f82`.
- `/tmp/gotth-webhooks-a3fb596.focused-race50.log` —
  `8bf2a9ea5958e280fa61413a129aa92e40ddde67e111b4cdda8096110dc93349`.
- `/tmp/gotth-webhooks-a3fb596.noncomment-identity.log` —
  `a52776d505ea6eb61ebdb9e55029a90370e5e289cd902ddb561062faa669581f`.
- `/tmp/gotth-webhooks-a3fb596.fuzz-endpoint.log` —
  `84d9ce066b125d2f0357d979030b7febe79b2116047a1267229342997f058366`.
- `/tmp/gotth-webhooks-a3fb596.fuzz-signing.log` —
  `f302c690bfa725a9f1cbcef21385ae6420dcb31bb234c33aa2dc553ac9ac01bc`.
- `/tmp/gotth-webhooks-a3fb596.fuzz-content-type.log` —
  `6ecaae70c3a58f79d0c8dc7e004dea9beb3f1edd0ccb0b38d49d473818dc2f70`.
- `/tmp/gotth-webhooks-a3fb596.hmac-openssl.log` —
  `c23d27896db18640b7451655ae1ff339012f4cc79c8c708eb886a161f3160944`.
- `/tmp/gotth-webhooks-a3fb596.clean-clone.log` —
  `7316b608b25ab894647abc90b478cd74c78ee8227157ea719683f8b3a13e12f5`.
- `/tmp/gotth-webhooks-a3fb596.synthetic-consumer.log` —
  `aa5274081bd4632f6ae349d2dc642cf39d4a5190675621279196803ae597e220`.
- `/tmp/gotth-webhooks-a3fb596.provenance.log` —
  `a99ecd69683370909516453b31a86d0aa9d0fb27a20791fcfe7b4a8ccd0dbd66`.
- `/tmp/gotth-webhooks-a3fb596.authority-hashes.log` —
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
