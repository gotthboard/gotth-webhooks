# Verification status

Independent cold Judge pass 6 rejected exact evidence head
`3eafee7e28eb806fab9e990c39224ef27a7c1e0c` because attempt-context readiness
could make actual untyped `net/http` protocol failures retryable and because
caller cost contracts omitted error-tree and delegated work. Source commit
`53872b9c99d2a7a6d90035ed9564d760c464c298` addresses those findings. A fresh
independent review remains required; this record does not claim admission.

Exact source-tip evidence under Go 1.26.6-X:nodwarf5, Linux amd64:

- `make verify` — PASS.
- Uncached `go test -mod=readonly -count=1 -race -coverprofile=... ./...` —
  PASS, 97.3% statements; all touched production functions are 100% covered.
- Focused real-transport cancellation/deadline/permanent-error, transient-DNS,
  MIME-expansion, capped-HTTP-date, and concurrency race suite repeated fifty
  times — PASS.
- Endpoint fuzz — PASS, 29,216 executions/5 seconds.
- Signing fuzz — PASS, 26,549 executions/5 seconds.
- Content-type fuzz — PASS, 36,860 executions/5 seconds.
- Independent OpenSSL HMAC-SHA-256 computation — PASS, exact vector
  `5afc952ae607736b86e3570dddc25991115f80afdb46c49832a30a5671028f7f`.
- Detached clone at exact source `53872b9c99d2a7a6d90035ed9564d760c464c298`
  with a fresh empty `GOCACHE` ran
  `make verify` — PASS, clean detached worktree, 97.3% statements.
- Synthetic external module `go test -mod=readonly -count=1 ./...` — PASS.
  This proves public syntax only, not consumer behavior or compatibility.
- The 22 historical changelog records at the source object match the declared
  two-lineage order, exact Git timestamps, and exact affected-file sets; one
  current-commit placeholder is present as policy requires.
- All six retained RFC/IANA authority snapshots revalidated against their
  recorded SHA-256 values.
- The exact performance matrix and limitations are in `performance.md`; no
  speedup or production-latency claim is made.

Exact source artifacts and SHA-256 values:

- `/tmp/gotth-webhooks-53872b9.verify.log` —
  `dc6e46524f43c1cfa012ff0947680570138fb9c0a7a0bc050f6427f3868cbf51`.
- `/tmp/gotth-webhooks-53872b9.coverage.out` —
  `ddd101779a1079eed2102e65627b2248a28c3831bc36aae14f4ea39d2cf58d9d`.
- `/tmp/gotth-webhooks-53872b9.coverage.log` —
  `486156377102dd8591eb9c932e60a646023cb7a7b84b246ef549709a268d4085`.
- `/tmp/gotth-webhooks-53872b9.coverage.func` —
  `66fe5b801d794154c8e24dec5b7188e25ab0c70419c79a50b529ac7d8fb0bb09`.
- `/tmp/gotth-webhooks-53872b9.focused-race50.log` —
  `5722702e59f58187f533fedcc18494ff2fc1b1da7c4a85d6dec3bf26b0a5a225`.
- `/tmp/gotth-webhooks-53872b9.fuzz-endpoint.log` —
  `67937b280acab7a7906e065d751304f15e66cef71031cf5d464fc3680f9c3979`.
- `/tmp/gotth-webhooks-53872b9.fuzz-signing.log` —
  `7e3f43f47a92dd6c13e8b7a42d59dfd44e833fa8b9f2ae8447b3af0251cbecc0`.
- `/tmp/gotth-webhooks-53872b9.fuzz-content-type.log` —
  `75250d2b32c22530d12635d8dd6d4b9db4f14fe4c5fd7758f8bd90d635869b35`.
- `/tmp/gotth-webhooks-53872b9.hmac-openssl.log` —
  `c23d27896db18640b7451655ae1ff339012f4cc79c8c708eb886a161f3160944`.
- `/tmp/gotth-webhooks-53872b9.clean-clone.log` —
  `b44b0b6fd2bb39fb5c6febe54ff3a9fb2544e58e8df4837082447c4239e0cef7`.
- `/tmp/gotth-webhooks-53872b9.synthetic-consumer.log` —
  `ecaa98fffadfe75aae8b7ff093bd588751a1a2c05fc17d422491357f812430b6`.
- `/tmp/gotth-webhooks-53872b9.provenance.log` —
  `af201a96ad88e26b47a792f568014a0c9a998e2bf21adf536e490eb85808a1c3`.
- `/tmp/gotth-webhooks-53872b9.authority-hashes.log` —
  `6895ab208fddb8977864210f09c009bae8fffe4c655689babf8f188713c3787a`.
- `/tmp/gotth-webhooks-53872b9.benchmark.txt` —
  `207bb0404bf222a23455e21788644d0cd44fc2cb7161d0eaa22e2805ecb61fa8`.

The feature remains `in_progress` and unreleased. The newly authorized product
contract work is outside this source repair, and there is still no real
consumer adapter, behavioral proof, dependency pin, tag, or compatibility
promise. No push, PR, tag, release, deployment, or live request occurred.

PostgreSQL/database integration is N/A because this package has no adapter.
Graphify is N/A because the repair affects one direct classifier plus bounded
tests/comments and has no dependency ambiguity worth graph construction.
