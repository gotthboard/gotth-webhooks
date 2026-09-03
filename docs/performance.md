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

- Host: agenthost, Linux amd64; Intel Core i7-7660U at 2.50 GHz.
- Compiler/runtime: Go 1.26.6-X:nodwarf5.
- Command: `go test ./pkg/webhooks -run '^$' -bench '^BenchmarkDeliver$' -benchmem -benchtime=100ms -count=10`.
- Fixture: fixed clock, in-memory receipt sink, deterministic local
  `RoundTripper`; no DNS, TCP, TLS, remote server, durable database, or wait.
- Correctness oracle: normal iterations deliver; overflow iterations fail
  after exactly the bounded read.
- Ten sample means per workload; nearest-rank percentiles. With ten samples,
  p95 and p99 are both the maximum and are coarse.
- Raw exact-source output at
  `f0d4008102380ad135e4a2d32190b8470af0fef8`:
  `/tmp/gotth-webhooks-f0d4008.benchmark.txt`, SHA-256
  `227104feb824d0481a73510edaa7604d974565f9ab62e7af47297ab868b13ed4`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op range | Allocs/op range |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 8.521 us | 8.713 us | 8.713 us | 117,357.12 ops/s | 4,295..4,299 | 73 |
| small 1 KiB payload | 13.623 us | 14.180 us | 14.180 us | 73,405.27 ops/s | 5,320..5,330 | 74 |
| typical 64 KiB payload | 333.318 us | 402.268 us | 402.268 us | 3,000.14 ops/s | 69,923..70,053 | 74 |
| boundary 1 MiB payload | 5.193189 ms | 5.331998 ms | 5.331998 ms | 192.56 ops/s | 1,054,122..1,056,238 | 76..77 |
| pathological 64 KiB+1 response | 10.576 us | 10.860 us | 10.860 us | 94,553.71 ops/s | 4,480..4,487 | 78 |

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Exact-source payload
throughput ranged from 162.92 to 203.50 MB/s across typical and boundary payload
samples.

Earlier runs on the same uncontrolled shared host produced materially different
wall times while allocations remained comparatively stable. The correctness
repair also removed the `http.Client` redirect layer, so its lower allocation
counts are a mechanism change, not a controlled optimization comparison. The
original raw artifact remains `/tmp/gotth-webhooks-benchmark.txt` with SHA-256
`e4d7d750ece5515bb3ac9f9f8db41185c60b00320ace8af45e3e168d7e4de712`.
Therefore these wall-clock percentiles are environment observations, not a
performance promise or a reliable regression comparison. Stable allocation
counts and bounded scaling are the useful evidence.

## Limitations and next measurement

This evidence admits only the visible in-process cost shape. It makes no claim
about DNS, TCP, TLS, remote receiver latency, response streaming, durable
recorder latency, contention, retries, or fleet scheduling. The first consumer
pin must measure its actual recorder, egress path, payloads, concurrency, and
service objectives. Re-profile before adding pooling, buffer reuse, streaming,
or parallelism; those complicate ownership or signing without current evidence.
