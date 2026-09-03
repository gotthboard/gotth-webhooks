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
- Raw output: `/tmp/gotth-webhooks-benchmark.txt`, SHA-256
  `e4d7d750ece5515bb3ac9f9f8db41185c60b00320ace8af45e3e168d7e4de712`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op | Allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 12.942 us | 13.919 us | 13.919 us | 77,267.81 ops/s | 4,996 | 79 |
| small 1 KiB payload | 19.585 us | 22.237 us | 22.237 us | 51,059.48 ops/s | 6,023 | 80 |
| typical 64 KiB payload | 341.090 us | 427.475 us | 427.475 us | 2,931.78 ops/s | 70,669 | 80 |
| boundary 1 MiB payload | 5.318247 ms | 5.632210 ms | 5.632210 ms | 188.03 ops/s | 1,056,504 | 83 |
| pathological 64 KiB+1 response | 11.468 us | 14.985 us | 14.985 us | 87,199.16 ops/s | 5,181 | 84 |

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Benchmark-reported
payload throughput ranged from 46.05 to 197.89 MB/s across samples.

## Limitations and next measurement

This evidence admits only the visible in-process cost shape. It makes no claim
about DNS, TCP, TLS, remote receiver latency, response streaming, durable
recorder latency, contention, retries, or fleet scheduling. The first consumer
pin must measure its actual recorder, egress path, payloads, concurrency, and
service objectives. Re-profile before adding pooling, buffer reuse, streaming,
or parallelism; those complicate ownership or signing without current evidence.
