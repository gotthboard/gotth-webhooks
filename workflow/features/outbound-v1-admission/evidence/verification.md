# Local repair verification evidence

Independent orchestrator review rejected exact commit
`a5a3b1060989f220bf97c4733a3248ac4c7e9130`; its evidence does not support
admission. Findings 2 through 9 are repaired at exact source tip
`7a0a940b10774b66eab3a5badd832181e560226a`. Independent post-repair reviews
rejected evidence heads `8696c6ebeebb0111f070d10d9b98768df066c932`
and `7c3fc0b18018b41a50b94b0d13952f7dba6b5aa0` for exhausted-dial,
wire-specification, empty-port, complexity, and changelog defects. The new
source tip repairs them without changing the consumer blocker.

Exact source-tip commands and results:

- `make verify` — PASS under Go 1.26.6-X:nodwarf5.
- `go test -mod=readonly -count=1 -race -coverprofile=... ./...` — PASS,
  96.5%. Artifact hashes:
  - coverage: `72f5d9ba7613587f4f4cc318ea8d0c4fe4e3924ae774ef2fa744fbfa2bf2ebe8`;
  - log: `80cd7074ef6734e52f7093509085ef64f64c8bbf16b141492fc7d46a199ff131`;
  - function summary:
    `0c2e57d2444c7451a424e2d3baa35968d72d44bb187bff1d9c26dbe37171479f`.
- Fifty uncached full race runs — PASS; log SHA-256
  `3c318c052b4ecbd04a9e80c614d49f8988e54d51027485bce4f245caa17c96c8`.
- `FuzzParseEndpoint` — PASS, 64,799 executions/5 seconds.
- `FuzzSignedRequestDeterministic` — PASS, 76,791 executions/5 seconds.
- Detached clone `/tmp/gotth-webhooks-7a0a940-clean.fnWUVi/repo`, detached at
  exact object `7a0a940b10774b66eab3a5badd832181e560226a`, with fresh empty
  `GOCACHE`: `make verify` PASS at 96.5%; tracked status clean.
- Synthetic module `/tmp/gotth-webhooks-consumer`:
  `go test -mod=readonly -count=1 ./...` — PASS. This is syntax evidence only,
  not a real consumer or compatibility pin.
- Exact benchmark raw SHA-256:
  `90fca519b715bc1e0571460acf4b8e1c1f504cf3393b1554e7b7acb095976e97`.
- Internal cold passes found ASCII-OWS/concurrency wording, false permanent
  status reporting, exhausted-dial classification, exact signed-target and
  empty-port defects, and cost-contract defects. Fresh pass 10 is source-CLEAN.

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
