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
  `62894967d7f2c8760d816d3f6d0c57928c8cab67`:
  `/tmp/gotth-webhooks-6289496.benchmark.txt`, SHA-256
  `1b94650cbbdae3d5f0b2a8782aa66cb5b8edff1df320353ace8e68bc115ff13b`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op range | Allocs/op range |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 18.675 us | 24.197 us | 24.197 us | 53,547.52 ops/s | 4,300..4,306 | 73 |
| small 1 KiB payload | 21.885 us | 29.439 us | 29.439 us | 45,693.40 ops/s | 5,320..5,335 | 74 |
| typical 64 KiB payload | 559.188 us | 1.223087 ms | 1.223087 ms | 1,788.31 ops/s | 69,993..70,258 | 74 |
| boundary 1 MiB payload | 9.587198 ms | 10.445094 ms | 10.445094 ms | 104.31 ops/s | 1,055,457..1,057,763 | 77..79 |
| pathological 64 KiB+1 response | 13.822 us | 18.092 us | 18.092 us | 72,348.43 ops/s | 4,483..4,489 | 78 |

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Exact-source payload
throughput ranged from 53.58 to 159.76 MB/s across typical and boundary payload
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
