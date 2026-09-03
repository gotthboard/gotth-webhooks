# Runtime boundary contract

## Supported targets

- Go 1.26.6 on Linux amd64 is the primary compiler/runtime contract.
- HTTPS over Go's standard `net/http.Transport`, `crypto/tls`, `net.Resolver`,
  `net.Dialer`, `crypto/hmac`, `crypto/sha256`, and `crypto/rand`.
- HTTP semantics follow RFC 9110; URL syntax follows RFC 3986; HMAC follows
  RFC 2104/FIPS 198; rate-limit response semantics follow RFC 6585.
- No database, proxy, custom transport, alternate TLS stack, or live service is
  supported by the library.

## Authorities read before design

- Go 1.26.6 `http.Client`: redirects are followed by default unless
  `CheckRedirect` intervenes; `Timeout` includes connection, redirects, and
  response-body reading; clients and transports are concurrency-safe and
  transports cache connections.
- Go 1.26.6 `http.Transport.DialContext`: dials may race with connection reuse,
  so address checks belong in the dial function and established validated
  connections may be reused.
- Go 1.26.6 `net/http` source: direct HTTPS connection-pool keys include target
  scheme and authority; only proxied HTTP uses an any-target pool key.
- Go 1.26.6 `net.Resolver.LookupNetIP`: returns IP addresses for a host;
  `net.Dialer.DialContext`: context bounds connection establishment but does
  not close an established connection when that context later expires.
- Go 1.26.6 `url.URL.RequestURI`: produces the encoded path/query actually used
  for an HTTP request.
- Go 1.26.6 `crypto/hmac`: HMAC is keyed authentication and receivers should
  use `hmac.Equal`; `crypto/rand.Reader` is a concurrency-safe CSPRNG backed by
  `getrandom(2)` on current Linux.
- RFC 9110 sections 4.2.4, 7.2, 9.2.2, 10.2.3, and 15: userinfo is unsafe,
  fragments are not request-target data, POST is not inherently idempotent,
  `Retry-After` permits date or delay forms, and status semantics are explicit.

## Correctness-relevant limits and oracles

All application limits are fixed in the implementation spec. Tests cover
limit-1, limit, limit+1, and materially beyond for body, response body, secret,
token, content type, attempts, timeouts, delays, address answer size, and
signature mutations where representable.

Completeness oracles are exact request-body SHA-256, captured request-target
and headers, exact attempt and recorder counts, response byte count including
the overflow probe, known address-answer cardinality, and listener-observed
numeric dial targets. A 2xx status without a completely bounded body read is
not success.

Warnings, redirects, partial response reads, mixed public/private DNS answers,
missing receipt records, and ambiguous URL forms fail closed. The library
changes no process-global, resolver-global, session, or environment state, so
restoration and pooled-session leakage tests are N/A. Connection reuse and the
DNS/rebinding limitations are explicit in the architecture.

Context deadlines are cooperative. The standard transport honors request
context, but a consumer `Recorder` that ignores its supplied context can block
past the deadline. Such a recorder violates the trusted interface contract;
Go cannot forcibly terminate it safely.
