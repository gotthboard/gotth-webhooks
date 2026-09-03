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
- Raw exact-candidate output at
  `94f2b7d080d9959af3c1d23a5a4349b555ee240b`:
  `/tmp/gotth-webhooks-94f2b7d.benchmark.txt`, SHA-256
  `bcb578a6535e9af7ba178440e5fd1950c4b4316c3be9501534c9036fe2491bc7`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op range | Allocs/op range |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 8.658 us | 12.560 us | 12.560 us | 115,500.12 ops/s | 4,296..4,300 | 73 |
| small 1 KiB payload | 18.202 us | 25.147 us | 25.147 us | 54,939.02 ops/s | 5,322..5,329 | 74 |
| typical 64 KiB payload | 0.424294 ms | 0.590677 ms | 0.590677 ms | 2,356.86 ops/s | 69,918..70,084 | 74 |
| boundary 1 MiB payload | 6.648275 ms | 8.922416 ms | 8.922416 ms | 150.41 ops/s | 1,054,117..1,056,735 | 76..78 |
| pathological 64 KiB+1 response | 10.516 us | 10.940 us | 10.940 us | 95,093.19 ops/s | 4,481..4,486 | 78 |

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Exact-candidate payload
throughput ranged from 110.95 to 196.40 MB/s across typical and boundary
payload samples.

Earlier runs on the same uncontrolled shared host produced drastically
different wall times while allocations remained comparatively stable. The
content-type repair adds a scan of at most 256 bytes and is not treated as an
optimization. An earlier correctness repair also removed the `http.Client`
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
