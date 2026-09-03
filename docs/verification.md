# Verification status

Current local candidate evidence:

- Exact primary toolchain: Go 1.26.6-X:nodwarf5, Linux amd64.
- `make verify` passes format, `go vet -mod=readonly`, race, tests, and coverage.
- Statement coverage is 96.9%. Residual statements are defensive impossible
  branches after validated URL/request construction, MIME formatter rejection
  after successful parse, and bounded-loop exhaustiveness. Every claimed
  outcome, status class, response failure, cancellation point, retry path,
  receipt failure, signature field, URL/IP rejection, and DNS transition has a
  direct test.
- Real loopback TLS integration proves a public resolver result becomes a
  numeric dial target, TLS authenticates `example.com` with a test CA, the
  signed request arrives intact, and a 302 redirect is not followed.
- Race covers 50 concurrent deliveries through one dispatcher. Repeated race
  and final clean-ref verification remain pending the first local commit.
- Fuzz admission: endpoint parser 95,312 executions/5 seconds and signing
  determinism 85,673 executions/5 seconds, both without product failure.
- An external module at `/tmp/gotth-webhooks-consumer` imports and exercises the
  public API with `go test -mod=readonly ./...`.
- PostgreSQL/database integration is N/A: this repository adds no adapter.
- Remote Go 1.26.5 is not the required 1.26.6 contract and was not used.
- Graphify is N/A: one small package has direct source/test correspondence, so
  a graph would add index cost without resolving an architectural ambiguity.

The feature remains `in_progress`. Local green tests do not admit, tag, release,
push, deploy, or establish compatibility.
