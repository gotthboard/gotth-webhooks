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
- Raw exact-source output: `/tmp/gotth-webhooks-1defba3.benchmark.txt`,
  SHA-256 `ef9f5ac50e8d4b38b085323769d8c3b820e45a1d459a45d9d8d5f7c8ddbaec87`.

## Results

| Workload | p50 | p95 | p99 | p50 throughput | Bytes/op | Allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| empty payload | 108.519 us | 202.378 us | 202.378 us | 9,214.95 ops/s | 5,005 | 79 |
| small 1 KiB payload | 283.566 us | 340.730 us | 340.730 us | 3,526.52 ops/s | 6,032 | 80 |
| typical 64 KiB payload | 3.657090 ms | 5.568866 ms | 5.568866 ms | 273.44 ops/s | 71,091 | 80 |
| boundary 1 MiB payload | 71.510554 ms | 93.286854 ms | 93.286854 ms | 13.98 ops/s | 1,054,084 | 82 |
| pathological 64 KiB+1 response | 140.196 us | 171.925 us | 171.925 us | 7,132.87 ops/s | 5,197 | 84 |

The pathological fixture uses an in-memory reader, so fast overflow rejection
proves bounded work and failure, not network latency. Exact-source payload
throughput ranged from 3.01 to 18.93 MB/s across samples.

An earlier run on the same uncontrolled shared host produced p50 values 8.4 to
17.8 times faster while allocations remained effectively unchanged. No source
optimization explains that difference; concurrent host load does. The raw
earlier artifact remains `/tmp/gotth-webhooks-benchmark.txt` with SHA-256
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
