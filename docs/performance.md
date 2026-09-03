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
  `7a0a940b10774b66eab3a5badd832181e560226a`:
  `/tmp/gotth-webhooks-7a0a940.benchmark.txt`, SHA-256
  `90fca519b715bc1e0571460acf4b8e1c1f504cf3393b1554e7b7acb095976e97`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op range | Allocs/op range |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 92.875 us | 173.101 us | 173.101 us | 10,767.16 ops/s | 4,295..4,314 | 73 |
| small 1 KiB payload | 144.222 us | 176.175 us | 176.175 us | 6,933.75 ops/s | 5,313..5,343 | 74 |
| typical 64 KiB payload | 3.448488 ms | 3.697285 ms | 3.697285 ms | 289.98 ops/s | 69,853..70,125 | 74 |
| boundary 1 MiB payload | 54.951676 ms | 58.882721 ms | 58.882721 ms | 18.20 ops/s | 1,053,388..1,062,300 | 76..83 |
| pathological 64 KiB+1 response | 120.415 us | 365.591 us | 365.591 us | 8,304.61 ops/s | 4,473..4,537 | 78 |

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Exact-source payload
throughput ranged from 17.73 to 19.73 MB/s across typical and boundary payload
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
