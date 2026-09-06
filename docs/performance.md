# Performance admission

## Decision

No optimization was introduced and no speedup is claimed. V1 validates and
copies once, hashes the body for a stable fingerprint, hashes it again for the
attempt signature, builds one request, consumes at most 64 KiB plus one
overflow byte, and records one receipt. At the 1 MiB request boundary the copy
and two SHA-256 passes dominate the in-process test.

Because there is no baseline/candidate optimization comparison, hotspot share
`P`, hotspot speedup `S_hotspot`, and Amdahl prediction
`1 / ((1-P) + P/S_hotspot)` are N/A. Numeric substitutions would be fabricated.
A future optimization must preserve byte ownership and signature/fingerprint
semantics and provide matched evidence before claiming an overall speedup.

## Environment and method

- Host: development, Linux amd64; AMD EPYC 7551P 32-Core Processor.
- Compiler/runtime: Go 1.26.6-X:nodwarf5.
- Command: `go test ./pkg/webhooks -run '^$' -bench '^BenchmarkDeliver$' -benchmem -benchtime=100ms -count=10`.
- Fixture: fixed clock, in-memory receipt sink, deterministic local
  `RoundTripper`; no DNS, TCP, TLS, remote server, durable database, or wait.
- Correctness oracle: normal iterations deliver; overflow iterations fail
  after exactly the bounded read.
- Ten sample means per workload; nearest-rank percentiles. With ten samples,
  p95 and p99 are both the maximum and are coarse.
- Raw exact-source output at
  `c92f9aab1537bd49d035b7019ef7e00af44d5679`:
  `/tmp/gotth-webhooks-c92f9aa-receipt-time.cLMg65/c92f9aa.performance.log`
  on development, SHA-256
  `c0ecd853e2b05cc0e016d1acfbdf30d7454ddabe6b565443d050709e0bc41987`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op range | Allocs/op range |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 28.331 us | 29.618 us | 29.618 us | 35,297.02 ops/s | 4,334..4,349 | 73 |
| small 1 KiB payload | 33.787 us | 34.925 us | 34.925 us | 29,597.18 ops/s | 5,373..5,393 | 74 |
| typical 64 KiB payload | 0.308038 ms | 0.315773 ms | 0.315773 ms | 3,246.35 ops/s | 70,536..70,711 | 74 |
| boundary 1 MiB payload | 4.202430 ms | 4.535586 ms | 4.535586 ms | 237.96 ops/s | 1,060,141..1,061,102 | 76..77 |
| pathological 64 KiB+1 response | 36.313 us | 37.059 us | 37.059 us | 27,538.35 ops/s | 4,523..4,542 | 78 |

The dispatcher lifecycle candidate adds a separate closed-admission benchmark
at exact source `20a122362ede5c6c934d39e809ff52747887bd1f`. Ten 100 ms samples
observed 74.66..87.41 ns/op, 0 B/op, and 0 allocations/op. Its raw log is
retained at
`workflow/features/dispatcher-lifecycle-repair/evidence/raw/benchmark.log`
with SHA-256
`4b63f563cc21c170d4506ea8f83544bc5cffe883b8f461e91f43bb5836a6a7d7`.
This proves the closed path does not validate or copy a message; it is not a
production scheduler or transport-latency claim.

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Exact-source payload
throughput ranged from 212.75 to 249.52 MB/s across typical and boundary
payload samples at the p50 values.

Earlier runs on uncontrolled shared hosts produced drastically different wall
times while allocations remained comparatively stable. The current receipt-
time repair adds one constant-time UTC conversion and microsecond truncation
per start and finish clock reading; the benchmark mechanism is unchanged and
the result is not treated as an optimization. An
earlier correctness repair removed the `http.Client` redirect layer, so lower
allocation counts across older runs include a mechanism change rather than a
controlled optimization comparison. The original raw artifact remains
`/tmp/gotth-webhooks-benchmark.txt` with SHA-256
`e4d7d750ece5515bb3ac9f9f8db41185c60b00320ace8af45e3e168d7e4de712`.
Therefore these wall-clock percentiles are environment observations, not a
performance promise or a reliable regression comparison. Stable allocation
counts and bounded scaling are the useful evidence.

## Limitations and next measurement

This evidence admits only the visible in-process cost shape. It makes no claim
about DNS, TCP, TLS, remote receiver latency, response streaming, durable
recorder latency, contention, retries, or fleet scheduling. The real-consumer
adoption/release gate must measure its actual recorder, egress path, payloads,
concurrency, and service objectives against the exact candidate dependency
pin. Re-profile before adding pooling, buffer reuse, streaming, or parallelism;
those complicate ownership or signing without current evidence.
