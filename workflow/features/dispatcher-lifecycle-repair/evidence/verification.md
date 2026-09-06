# Dispatcher lifecycle verification

## Identity

- Base: `2b10dc0ae67feefe826eebc64315a5ce0c14b4f9`
- Planning commit: `0a100b0a484706174b7f675f016cc78206e652b2`
- Exact source: `20a122362ede5c6c934d39e809ff52747887bd1f`
- Source tree: `6157878724ed9b20b40053625dd33ae0b09f6a4b`
- Source bundle SHA-256:
  `1cb0b180d7231e96b16d79637506c05faf3b9367f468e362e9343b3c27272c81`
- Runner SHA-256:
  `22f0d99585034f3e74f97b35f28d0f947a734f3884ed840233417869f86c1f8e`
- Host/toolchain: `development`, Linux amd64, Go
  `go1.26.6-X:nodwarf5`.

The source bundle was cloned to a fresh detached checkout. Every retained gate
log records exact before/after HEAD and empty before/after porcelain status.
The runner, raw logs, coverage profile, function report, metadata, and hashes
are retained beside this file. A second fresh clone from the same bundle also
passed an uncached full test with exact HEAD and clean status before and after.
Trailing horizontal whitespace emitted by tools is normalized before retained
text logs are hashed so the evidence commit remains `git diff --check` clean;
no substantive output is filtered.

## Gates

- uncached full repository test: PASS;
- `make verify` format/version/vet/race/coverage gate: PASS;
- fresh uncached race coverage: PASS, 97.4% repository statements;
- full repository under race, 50 repetitions: PASS;
- both lifecycle regressions under race, 50 repetitions: PASS;
- `FuzzParseEndpoint`: PASS, 1,047,369 executions;
- `FuzzSignedRequestDeterministic`: PASS, 897,677 executions;
- `FuzzCanonicalContentType`: PASS, 1,017,642 executions;
- `BenchmarkDeliver` and `BenchmarkClosedDeliver`, ten 100 ms samples with
  allocation reporting: PASS;
- disposable external-consumer compile using public `Close` and `ErrClosed`:
  PASS; and
- second clean-clone uncached full test: PASS.

The `Close` function is 100.0% statement covered. The changed lifecycle issue
surface covers the closable and non-closable internal transport branches,
closed-before-validation precedence, zero-result `ErrClosed`, no transport or
recorder side effects after close admission, one cleanup invocation across
concurrent/repeated closes, concurrent callers waiting for cleanup, and an
already-admitted delivery completing after `Close`. The repository's remaining
2.6% uncovered statements predate this repair; no changed production statement
is uncovered.

`BenchmarkClosedDeliver` observed 74.66..87.41 ns/op, 0 B/op, and 0
allocations/op. This is an in-process mechanism observation, not a latency or
throughput promise.

## Boundaries

The public configuration still cannot inject a transport or cleanup hook.
Tests use the existing package-internal transport seam. `Close` does not cancel
admitted calls or track signing generations; consumers own producer shutdown
and generation quiescence. No network destination, live consumer, Board tree,
remote repository, PR, merge, tag, release, or deployment was touched.

Independent review and final admission remain open.
