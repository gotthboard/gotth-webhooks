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
  `f1fd980e8e6bd184815b871fb5b7513a999725d1`:
  `/tmp/gotth-webhooks-f1fd980.benchmark.txt`, SHA-256
  `d45c180a8261f2e7024b7eb7d351e71fe046e76cd5a5311e30c7293fa508d44f`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op range | Allocs/op range |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 11.652 us | 13.075 us | 13.075 us | 85,822.18 ops/s | 4,297..4,301 | 73 |
| small 1 KiB payload | 17.190 us | 22.384 us | 22.384 us | 58,173.36 ops/s | 5,324..5,329 | 74 |
| typical 64 KiB payload | 0.449464 ms | 0.646433 ms | 0.646433 ms | 2,224.87 ops/s | 69,948..70,192 | 74 |
| boundary 1 MiB payload | 5.455674 ms | 8.243447 ms | 8.243447 ms | 183.30 ops/s | 1,054,828..1,056,458 | 76..78 |
| pathological 64 KiB+1 response | 13.277 us | 15.005 us | 15.005 us | 75,318.22 ops/s | 4,480..4,486 | 78 |

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Exact-source payload
throughput ranged from 101.38 to 196.39 MB/s across typical and boundary
payload samples.

Earlier runs on the same uncontrolled shared host produced drastically
different wall times while allocations remained comparatively stable. The
deadline-precedence repair only reorders bounded error-classification branches
and is not treated as an optimization. An earlier correctness repair also
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
