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
  `bf64d724bd68bcb95f7180e1b79c16351e1d881c`:
  `/tmp/gotth-webhooks-bf64d72.benchmark.txt`, SHA-256
  `8d67754d4ccd68ab227d770656a793d5799eef3dafce2200f5f1a00d0632eaeb`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op range | Allocs/op range |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 27.698 us | 29.526 us | 29.526 us | 36,103.69 ops/s | 4,339..4,347 | 73 |
| small 1 KiB payload | 34.271 us | 35.364 us | 35.364 us | 29,179.19 ops/s | 5,371..5,394 | 74 |
| typical 64 KiB payload | 0.306776 ms | 0.312970 ms | 0.312970 ms | 3,259.71 ops/s | 70,576..70,722 | 74 |
| boundary 1 MiB payload | 4.579924 ms | 4.722347 ms | 4.722347 ms | 218.34 ops/s | 1,059,659..1,060,640 | 76 |
| pathological 64 KiB+1 response | 37.076 us | 37.585 us | 37.585 us | 26,971.63 ops/s | 4,524..4,534 | 78 |

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Exact-source payload
throughput ranged from 213.63 to 228.95 MB/s across typical and boundary
payload samples at the p50 values.

Earlier runs on uncontrolled shared hosts produced drastically different wall
times while allocations remained comparatively stable. The
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
