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
  `53872b9c99d2a7a6d90035ed9564d760c464c298`:
  `/tmp/gotth-webhooks-53872b9.benchmark.txt`, SHA-256
  `207bb0404bf222a23455e21788644d0cd44fc2cb7161d0eaa22e2805ecb61fa8`.
- Later source `bad5c8171e28666a32f1603205ee0218f8af2e67` changes only full-line
  production comments. Its non-comment production-source hash equals the
  benchmark source's final evidence head, so no runtime benchmark rerun or new
  performance claim is manufactured. Identity evidence is recorded in
  `verification.md`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op range | Allocs/op range |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 35.369 us | 53.979 us | 53.979 us | 28,273.35 ops/s | 4,300..4,313 | 73 |
| small 1 KiB payload | 27.768 us | 38.873 us | 38.873 us | 36,012.68 ops/s | 5,325..5,341 | 74 |
| typical 64 KiB payload | 0.881020 ms | 1.339400 ms | 1.339400 ms | 1,135.05 ops/s | 70,107..70,390 | 74 |
| boundary 1 MiB payload | 11.349994 ms | 13.915099 ms | 13.915099 ms | 88.11 ops/s | 1,055,513..1,060,640 | 77..82 |
| pathological 64 KiB+1 response | 32.905 us | 44.497 us | 44.497 us | 30,390.52 ops/s | 4,487..4,495 | 78 |

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Exact-source payload
throughput ranged from 48.93 to 100.78 MB/s across typical and boundary
payload samples.

Earlier runs on the same uncontrolled shared host produced drastically
different wall times while allocations remained comparatively stable. The
causal-deadline repair only reorders bounded error-classification branches and
adds test-only real-transport fixtures; it is not treated as an optimization.
An earlier correctness repair also
removed the `http.Client`
redirect layer, so lower allocation counts across older runs include a
mechanism change rather than a controlled optimization comparison. The
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
