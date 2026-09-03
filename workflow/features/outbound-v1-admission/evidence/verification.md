# Local repair verification evidence

Independent reviews rejected the prior candidates through exact head
`ff4fee3f5aa94cb5b44694050188793e5b181a8d`. Source commit
`be6f7150f3419ff3bcb971476119e7cc3f2cb10b` repairs the remaining
content-type control-byte contract. Exact candidate
`94f2b7d080d9959af3c1d23a5a4349b555ee240b` also corrects historical
changelog times and file attribution against Git. The consumer blocker is
unchanged.

Exact source-tip commands and results:

- `make verify` — PASS under Go 1.26.6-X:nodwarf5.
- `go test -mod=readonly -count=1 -race -coverprofile=... ./...` — PASS,
  96.5%. Artifact hashes:
  - coverage: `d726bb571e9c7aaa908017281cf2e939ceb441d9dee0458292585ac537d206f4`;
  - log: `0ea7a506966f3668504a627d4eee3385245895feae54a3f80008aec91ef1da57`;
  - function summary:
    `bd04b3a37fc42b96b3a1f891327d9606b51ecdb03031ae1b3edc44ca6cee0ef8`.
- Fifty uncached full race runs — PASS; log SHA-256
  `06a0544953f7e992d1d81f4fa435b4ace668f4e9e22a96a97497a9c13f266b2c`.
- `FuzzCanonicalContentType` — PASS, 115,255 executions/5 seconds; log
  SHA-256 `8b950e8ba634b6a64e87d9ec35e48671196975534ed789d368e73d5870674fa7`.
- `FuzzSignedRequestDeterministic` — PASS, 82,089 executions/5 seconds; log
  SHA-256 `e24e14b88b3e57395d5f48742d131b0bc0731b7081566383f9f5d089619290ce`.
- Detached clone `/tmp/gotth-webhooks-94f2b7d-clean.FPVPPK/repo`, detached at
  exact object `94f2b7d080d9959af3c1d23a5a4349b555ee240b`, with fresh empty
  `GOCACHE`: `make verify` PASS at 96.5%; tracked status clean.
- Synthetic module `/tmp/gotth-webhooks-consumer`:
  `go test -mod=readonly -count=1 ./...` — PASS. This is syntax evidence only,
  not a real consumer or compatibility pin.
- Exact benchmark raw SHA-256:
  `bcb578a6535e9af7ba178440e5fd1950c4b4316c3be9501534c9036fe2491bc7`.
- Exact file-list and second-resolution CDT heading-time comparison against
  Git passed for all 16 named changelog commits.
- Internal cold passes found ASCII-OWS/concurrency wording, false permanent
  status reporting, exhausted-dial classification, exact signed-target and
  empty-port defects, cost-contract defects, MIME controls, and changelog
  provenance. Fresh pass 13 is source-CLEAN.

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
