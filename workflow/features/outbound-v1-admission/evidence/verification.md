# Local repair verification evidence

Independent orchestrator review rejected exact commit
`a5a3b1060989f220bf97c4733a3248ac4c7e9130`; its evidence does not support
admission. Findings 2 through 9 are repaired at exact source tip
`62894967d7f2c8760d816d3f6d0c57928c8cab67`. Independent post-repair review
rejected evidence HEAD `8696c6ebeebb0111f070d10d9b98768df066c932`
for exhausted-dial classification and two cost-contract defects; the new source
tip repairs them without changing the consumer blocker.

Exact source-tip commands and results:

- `make verify` — PASS under Go 1.26.6-X:nodwarf5.
- `go test -mod=readonly -count=1 -race -coverprofile=... ./...` — PASS,
  96.5%. Artifact hashes:
  - coverage: `626b1fbe8b8bc4ee3c6f7bdf71f9026f51455a3e4671c73c653ba7f3cd7fda10`;
  - log: `9cba9d76f5574b683031e61395a8f6a7557ee0ab44f55d1cfdd6245ed10db7ca`;
  - function summary:
    `5597984a4f783ad1ecf6254eefd28d00380dfb62cd3d16854358a3ee387dc070`.
- Fifty uncached full race runs — PASS; log SHA-256
  `99f2c29650aee72ea154664861664b09fb229c5e36d0176feef2ee09c83e3781`.
- `FuzzParseEndpoint` — PASS, 69,826 executions/5 seconds.
- `FuzzSignedRequestDeterministic` — PASS, 56,792 executions/5 seconds.
- Detached clone `/tmp/gotth-webhooks-6289496-clean.GzxYIp/repo`, detached at
  exact object `62894967d7f2c8760d816d3f6d0c57928c8cab67`, with fresh empty
  `GOCACHE`: `make verify` PASS at 96.5%; tracked status clean.
- Synthetic module `/tmp/gotth-webhooks-consumer`:
  `go test -mod=readonly -count=1 ./...` — PASS. This is syntax evidence only,
  not a real consumer or compatibility pin.
- Exact benchmark raw SHA-256:
  `1b94650cbbdae3d5f0b2a8782aa66cb5b8edff1df320353ace8e68bc115ff13b`.
- Internal cold passes found ASCII-OWS/concurrency wording, false permanent
  status reporting, exhausted-dial classification, and two cost-contract
  defects. Fresh pass 7 is source-CLEAN.

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
