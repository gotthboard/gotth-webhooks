# Dispatcher lifecycle drain repair verification

## Identity

- Rejected evidence HEAD: `065ba1d9f47219aab7e3664c275c8f9ba9b60838`
- Independent review SHA-256:
  `e0ca239ee5756c5c913d3323f321e61dafa8efd98959b2a4afcc054c77233c91`
- Repair source: `120db639e874419a71af4e10447b8e9ea1f73913`
- Repair source tree: `29884ceafedc180fc86606bcd84bb530e216cc68`
- Source bundle SHA-256:
  `fc3b7d5c634e71f18e056a35a7f6c9fcec1543dfbfc4bcacb37a8966cd72a06a`
- Development runner SHA-256:
  `495963e63ccbc6608947f29d7c16faf4311ad8cdbbe4033712200c98adc02acd`
- Expected-red runner SHA-256:
  `5ea918088fe2ff103dc0eee73ffd05c9d4fb587fa923fe8782f9da348bfa0bd8`
- Raw manifest SHA-256:
  `a0d54f5a7ee906157f28046e1e1b804fb06d6e21ec102421df27f7de60d639d2`
- Host/toolchain: `development`, Linux amd64, Go
  `go1.26.6-X:nodwarf5`.

The source bundle was cloned to a clean detached checkout. Each retained gate
records exact before/after HEAD and empty before/after porcelain status. A
second clone from the same bundle passed an uncached full test and remained
clean. Retained text output has only trailing horizontal whitespace normalized
before hashing.

## Rejection and repair

Independent review 1 found that the first candidate could clean the idle pool
before a delivery admitted by the atomic closed check reached `RoundTrip`.
Go's transport can clear its close-idle state when that later request enters,
allowing the resulting connection to return to the pool after the one-shot
cleanup.

The expected-red tests paused admitted deliveries before their first
`RoundTrip` and between retries. All three lifecycle regressions failed against
the rejected source; the retained expected-red log has SHA-256
`e1fb6c703d5ff4d40f64ca356f8d6df2315a35abc540bb8d61f8a00d5a3cce83`.

The repair shares a lifecycle object across dispatcher value copies. Delivery
admission and the closed transition use one mutex; successful admissions add
to a drain count before releasing that mutex. The first `Close` rejects later
admissions, waits for the count to reach zero, and then performs the sole idle
cleanup. This is aggregate accounting, not a per-request registry.

## Gates

- uncached full repository test: PASS;
- `make verify` format/version/vet/race/coverage gate: PASS;
- fresh uncached race coverage: PASS, 97.5% repository statements;
- full repository under race, 50 repetitions: PASS;
- lifecycle regressions under race, 50 repetitions: PASS;
- closed-path allocation assertions, 50 repetitions: PASS;
- `FuzzParseEndpoint`: PASS, 1,060,613 executions;
- `FuzzSignedRequestDeterministic`: PASS, 994,095 executions;
- `FuzzCanonicalContentType`: PASS, 34,468 executions;
- `BenchmarkDeliver` and `BenchmarkClosedDeliver`, ten 100 ms samples with
  allocation reporting: PASS;
- disposable external-consumer compile using public `Close` and `ErrClosed`:
  PASS; and
- second clean-clone uncached full test: PASS.

`Close` and `newDispatcher` are 100.0% statement covered. `Deliver` is 95.3%
covered, and all new admission, rejection, and drain statements are executed.
The remaining repository 2.5% and `Deliver` 4.7% gaps are pre-existing failure
branches outside this repair. `BenchmarkClosedDeliver` observed
110.3..123.7 ns/op, 0 B/op, and 0 allocations/op. Direct allocation assertions
also prove zero allocations for closed `Deliver` and repeated `Close`.

## Boundaries

`Close` may wait for admitted delivery work, including retry waits and receipt
recording. It cannot rescue a dependency that violates its context contract,
and a dependency callback must not synchronously close the dispatcher whose
delivery it is executing. The public configuration still cannot inject a
transport. No Board or notify tree, live destination, remote repository, PR,
merge, tag, release, or deployment was touched. Independent rereview and final
admission remain orchestrator-owned.
