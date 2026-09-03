# Local repair verification evidence

Independent orchestrator review rejected exact commit
`a5a3b1060989f220bf97c4733a3248ac4c7e9130`; its evidence does not support
admission. Findings 2 through 9 are repaired at exact source tip
`f0d4008102380ad135e4a2d32190b8470af0fef8`.

Exact source-tip commands and results:

- `make verify` — PASS under Go 1.26.6-X:nodwarf5.
- `go test -mod=readonly -count=1 -race -coverprofile=... ./...` — PASS,
  95.8%. Artifact hashes:
  - coverage: `ced87114ceb4a71b8c0920c8453cb104e743fe538f0d532b2c383f2d38950cd0`;
  - log: `306a4e4e576feee7253c40d2346bce7faa613b33eded66054ee6a3a90cd221f3`;
  - function summary:
    `b1dfa157cae2728bd4699ad3971f09833603faff43e1ce173d37890c9bdf635a`.
- Fifty uncached full race runs — PASS.
- `FuzzParseEndpoint` — PASS, 81,989 executions/5 seconds.
- `FuzzSignedRequestDeterministic` — PASS, 61,895 executions/5 seconds.
- Detached clone `/tmp/gotth-webhooks-f0d4008-clean.Bgr3NZ/repo`, detached at
  exact object `f0d4008102380ad135e4a2d32190b8470af0fef8`, with fresh empty
  `GOCACHE`: `make verify` PASS at 95.8%; tracked status clean.
- Synthetic module `/tmp/gotth-webhooks-consumer`:
  `go test -mod=readonly -count=1 ./...` — PASS. This is syntax evidence only,
  not a real consumer or compatibility pin.
- Exact benchmark raw SHA-256:
  `227104feb824d0481a73510edaa7604d974565f9ab62e7af47297ab868b13ed4`.
- Internal cold passes found ASCII-OWS/concurrency wording and false permanent
  status reporting; `b34880e` and `f0d4008` repaired them. Fresh pass 4 is
  source-CLEAN.

The exact clean-clone commit and source object are identical. The following
older evidence is retained only as rejected-history provenance.

Candidate and exact-source commands:

- `make verify` — PASS under Go 1.26.6-X:nodwarf5.
- `go test -race -coverprofile=/tmp/gotth-webhooks-coverage-final.out ./...`
  — PASS; superseded by the exact repair-source 97.0% run below.
- Exact repair-source coverage at `1defba3673c8e8948501c9f65997b58573d6a844`:
  PASS, 97.0%. Artifacts:
  - `/tmp/gotth-webhooks-1defba3.coverage.out`, SHA-256
    `9743e14290bc3a3e747c02335edcdc420002cf1aeef7883d323584cf7c2a3424`.
  - `/tmp/gotth-webhooks-1defba3.coverage.log`, SHA-256
    `685a33dc5670137ca6728f377eb7ec5c2d4040b61e1e09ddec17c3e58d70b354`.
  - `/tmp/gotth-webhooks-1defba3.coverage.func`, SHA-256
    `0702f83d41a77ed8a1d38b5b219f796432f5f731aa930d0a0330221bc12fcf77`.
- Fifty uncached `go test -count=1 -race ./...` runs after repairs — PASS.
- Exact repair-source endpoint fuzz — PASS, 79,297 executions/5 seconds.
- Exact repair-source signing fuzz — PASS, 71,456 executions/5 seconds.
- External consumer `go test -mod=readonly ./...` in
  `/tmp/gotth-webhooks-consumer` — PASS.
- Benchmark command and raw hash are in `docs/performance.md`.

- Detached clean clone of pre-repair commit `338b885` passed `make verify`;
  the rejected candidate did not have committed exact-final evidence.
- Performance exact-source artifact and uncontrolled-host variance are in
  `docs/performance.md`.

The earlier internal cold Judge missed the independent findings and is
superseded for admission purposes. This evidence does not mark the feature
done.
