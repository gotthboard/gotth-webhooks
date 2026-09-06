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
  `7de792774d2a9229f577d8d46dfb803b1b1c9bfe`:
  `/tmp/gotth-webhooks-7de7927-repair.0tljnO/7de7927.performance.log`
  on development, SHA-256
  `73f7d021f17edfcc39c2f6dbe22e42faed12c346fd929ae8f6fa2114ddbe450a`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op range | Allocs/op range |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 28.537 us | 30.048 us | 30.048 us | 35,042.23 ops/s | 4,337..4,350 | 73 |
| small 1 KiB payload | 34.263 us | 35.588 us | 35.588 us | 29,186.00 ops/s | 5,373..5,385 | 74 |
| typical 64 KiB payload | 0.308946 ms | 0.314287 ms | 0.314287 ms | 3,236.81 ops/s | 70,580..70,673 | 74 |
| boundary 1 MiB payload | 4.531674 ms | 4.921437 ms | 4.921437 ms | 220.67 ops/s | 1,059,238..1,061,233 | 75..77 |
| pathological 64 KiB+1 response | 36.108 us | 37.381 us | 37.381 us | 27,694.69 ops/s | 4,527..4,536 | 78 |

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Exact-source payload
throughput ranged from 212.13 to 231.39 MB/s across typical and boundary
payload samples at the p50 values.

Earlier runs on uncontrolled shared hosts produced drastically different wall
times while allocations remained comparatively stable. The current IANA
completeness repair adds only test fixtures and test code; the production
benchmark mechanism is unchanged and is not treated as an optimization. An
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
