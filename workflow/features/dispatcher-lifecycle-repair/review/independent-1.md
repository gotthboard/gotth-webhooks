# Independent Final Maintainer Review 1 of 2

Verdict: REJECT

- Candidate HEAD: `065ba1d9f47219aab7e3664c275c8f9ba9b60838`
- Candidate tree: `b21465ce457bc3d2697f8b8ac108ad920f278057`
- Base: `2b10dc0ae67feefe826eebc64315a5ce0c14b4f9`
- Base tree: `d700475b6abcb089ebc1fc95bfc6a56d6811359f`
- Initial worktree status: CLEAN
- Final worktree status: CLEAN
- Reviewed toolchain: `go version go1.26.6-X:nodwarf5 linux/amd64`

## Findings

### HIGH: The one-shot transport cleanup can run before an admitted delivery starts `RoundTrip`, leaving a post-close idle connection behind

References: `pkg/webhooks/dispatcher.go:62`, `pkg/webhooks/dispatcher.go:95`, `pkg/webhooks/dispatcher.go:115`, `pkg/webhooks/dispatcher_test.go:70`, `docs/implementation-spec.md:26`, `docs/implementation-spec.md:32`, `docs/architecture.md:130`, `docs/runtime-boundary.md:27`, `/usr/lib/go/src/net/http/transport.go:897`, `/usr/lib/go/src/net/http/transport.go:1182`, `/usr/lib/go/src/net/http/transport.go:1547`.

`Deliver` is admitted by the atomic load at `dispatcher.go:95`, but `Close` does not coordinate with that admitted call before it performs its sole `CloseIdleConnections` invocation at `dispatcher.go:62-68`. An admitted call can still be validating/copying its message, reading the clock, waiting between retries, or otherwise have not entered its next `RoundTrip` when `Close` completes cleanup. When that call subsequently starts HTTP/1 work, Go 1.26.6 `Transport.getConn` calls `queueForIdleConn`; that function explicitly clears `t.closeIdle` to undo `CloseIdleConnections` (`transport.go:1182-1184`). The new connection can then be returned to the idle pool and remain there until the 30-second idle timeout. A retry begun after `Close` has the same defect. `sync.Once` prevents the dispatcher from ever cleaning it up again.

This violates WHK-014's promise to let pre-close admitted deliveries finish while releasing the owned transport's idle resources. It also makes the architecture statement that active requests make request accounting unnecessary incomplete: that source behavior only protects a connection already in transport use, not an admitted delivery that enters transport later.

The lifecycle test does not cover the contract's linearization point. It waits for `lifecycleTransport.RoundTrip` and calls that event "delivery admission" at `dispatcher_test.go:70`; actual admission happened earlier at `dispatcher.go:95`. The fake transport also cannot model Go's later request clearing `closeIdle`, so 100% statement coverage and race repetition do not establish cleanup correctness.

Smallest acceptable fix: synchronize the closed transition with shared admission accounting. A successful `Deliver` admission must register before releasing the admission lock and unregister on every return; the first `Close` must reject later admissions, wait for all registered deliveries (including all retries and recording) to exit, and only then invoke `CloseIdleConnections` exactly once. Concurrent `Close` callers must continue waiting for that cleanup. Update the public contract to state that `Close` may wait for admitted deliveries, and add deterministic regressions that pause an admitted delivery before its first `RoundTrip` and between retry attempts, proving cleanup occurs only after both can no longer create or reuse transport connections. Use shared lifecycle state if pre-first-use `Dispatcher` copies remain supported; otherwise tighten the copy prohibition and record the behavioral compatibility break rather than calling the API change purely additive.

## Verification Evidence

- Exact candidate HEAD and tree matched the requested values before and after review.
- `git status --porcelain=v1 --untracked-files=all` was empty before and after review.
- `git diff --check 2b10dc0ae67feefe826eebc64315a5ce0c14b4f9..065ba1d9f47219aab7e3664c275c8f9ba9b60838` passed.
- All 16 retained raw evidence files passed `sha256sum -c`.
- Focused local `go test -mod=readonly -p=1 -race -count=1 ./pkg/webhooks -run '^TestDispatcherClose'` passed with `GOMAXPROCS=2`.
- Focused `go vet -mod=readonly ./pkg/webhooks` and `gofmt -d` checks passed.
- Retained `BenchmarkClosedDeliver` evidence reports 0 B/op and 0 allocs/op; the finding is transport lifetime correctness, not an allocation regression.
