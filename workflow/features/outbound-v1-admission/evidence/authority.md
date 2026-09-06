# Authority evidence

Design authorities read before implementation:

- Go 1.26.6 local docs for `http.Client`, `http.Transport.DialContext`,
  `net.Resolver.LookupNetIP`, `net.Dialer.DialContext`, `url.URL.RequestURI`,
  `crypto/hmac`, `crypto/rand.Reader`, and `context.WithoutCancel`.
- Go source at `/usr/lib/go/src/net/http/transport.go` and `h2_bundle.go` for
  direct connection pool keys and HTTP/2 address-key behavior.
- RFC 2104 SHA-256: `64d5245a9101929025336e470e3737f118704001249d503c85a86e19fe9fbb01`.
- RFC 3986 SHA-256: `3102dae4b68cebe40337730312fcb612297b8928547267e8b3d1ee6002b2d683`.
- RFC 9110 SHA-256: `21c1cdce6ab0e5509b04d84a28000836c7a087cf786efe6f04877ebfff47232a`.
- RFC 6585 SHA-256: `f6d55d1b491cd515c35827cf9181753b23b2a68c4df14e56d83dc445b3876e58`.
- IANA IPv4 Special-Purpose Address Registry, updated 2025-10-09 and fetched
  read-only 2026-09-03: SHA-256
  `cf24e11f41b7d42c68debe2d18b97cac815084ec413ebb3b244f704028a16f20`.
- IANA IPv6 Special-Purpose Address Registry, updated 2025-10-09 and fetched
  read-only 2026-09-03: SHA-256
  `c17f4380ba84fb2160dae82ebfd8bd155a5853cfab624ed3a9fd251638a8be02`.

Raw RFC snapshots remain task scratch under `/tmp/gotth-webhooks-authority`.
The exact IANA registry snapshots are retained as repository test fixtures at
`pkg/webhooks/testdata/iana-ipv4-special-registry.xml` and
`pkg/webhooks/testdata/iana-ipv6-special-registry.xml`. Their test-enforced
hashes match the values above; test parsing also checks registry IDs, the
2025-10-09 update date, all 26 IPv4 prefixes, all 25 IPv6 prefixes, and
coverage of each allocation by the compact production deny table.

The internal Judge used context-broker 0.1.0 (broker SHA-256
`8826786c571a5b906b23bda74e91dcbcf055a592b315e3512564b7c6da9924bb`)
against clean source `338b885077a5870b5645b2e55a257ac487eab638`, base
`0bba05927d7922e693a6211ffc41ee3ab91ba451`, mode `judge`, budgets 12
files/100 lines/30,000 bytes. It was a cache miss at
`/home/linus/.cache/openclaw-code-context/dfbe6fd02ca720e0/a1e27a0c24371b33/ed711b9bd96717fb8573186b3a74baabe34bd6a4ce7e4d2f082dad99116c0639.json`.
The packet byte budget was not exhausted, but the repository scan hit its
100-line emission bound and was marked truncated. Every consequential finding
was independently verified in complete production source, Go contracts, and
focused tests.

The pass-5 repair used context-broker 0.1.0 (same broker SHA-256) against clean
head `c30635434b88718f99ee5125052fdc309e737db9`, base
`0bba05927d7922e693a6211ffc41ee3ab91ba451`, mode `handoff`, budgets 16
files/100 lines/30,000 bytes. It was a cache miss at
`/home/linus/.cache/openclaw-code-context/dfbe6fd02ca720e0/8c113b1e0c50836a/56db1714b87bf8e9c97406e34801ca425995a0df449890f81d83b2cfbe0c36e7.json`.
The packet emitted 10 files/100 lines/12,809 bytes and marked its broader scan
truncated. It was navigation evidence only; the cited source, tests, Git
objects, Go contracts, and retained artifacts were read or verified directly.

The pass-6 repair directly inspected Go 1.26.6 `errors/wrap.go` recursion and
`net/http/request.go`, `response.go`, `transfer.go`, and `transport.go` before
changing classification. In particular, the owned transport's private
malformed-response and response-header-limit errors were captured through a
real `http.Transport` test fixture instead of inferred from deprecated or
unrelated exported sentinels. No new broker packet or external fetch was needed.

The pass-10 repair read the adjacent `validateMessage`, `buildRequest`,
`attempt`, `Deliver`, and `consumeResponse` contracts as one call chain. It
propagates the callee's already-declared HMAC costs and treats caller-supplied
`io.ReadCloser.Read`/`Close` CPU, allocation, I/O, and latency as delegated.
No runtime source, new authority fetch, broker packet, or graph was required.

The pass-11 repair directly inspected Go 1.26.6 `io.copyBuffer` dispatch and
`io.discard.ReadFrom` at `/usr/lib/go/src/io/io.go:407-415,662-675`. A Reader
may repeatedly return `(0, nil)`, so response byte count alone does not bound
local copy-loop iterations. The contract now carries completed Read callback
count through `consumeResponse`, attempt, and Deliver. No runtime source, new
fetch, broker packet, or graph was required.

The fresh post-pass-11 audit directly inspected Go 1.26.6
`net/http/h2_bundle.go` request retry logic, `net/http.Transport` protocol and
HTTP/1 retry selection, `io.ReadAtLeast`/`io.ReadFull`, and `time.NewTimer`.
HTTP/2 may replay a reconstructible request after `REFUSED_STREAM`, selected
peer protocol errors, or graceful GOAWAY inside one `RoundTrip`; HTTP/1 does
not replay this unmarked POST after request bytes are written. `io.ReadFull`
can repeat zero-progress callbacks, and a timer becomes eligible after at least
its delay rather than guaranteeing a return-time upper bound. These source
contracts drive the current narrow repair and do not constitute admission.

The registry-completeness repair used the exact already-hashed snapshots above;
it did not fetch new authority data. Independent report
`/tmp/gotth-webhooks-independent-judge-1.md`, SHA-256
`0fed933a862d1310ce09612f2cd7919bbc9848a0bf9f86855e8c950f2148c93c`,
identified the missing independent oracle and the admission-gate conflation at
reviewed clean head `381b680e4b6e0945746d5d2db87b3a60bc797924`.

The receipt-time repair directly inspected Go 1.26.6 documentation and source
for `time.Time.UTC` and `time.Time.Truncate`. `UTC` sets the location to UTC;
`Truncate` rounds down to a duration multiple since the zero time using the
absolute instant and strips the monotonic reading. The implementation therefore
uses `value.UTC().Truncate(time.Microsecond)` explicitly, never a local zone or
round-to-nearest conversion. No external fetch or new authority snapshot was
needed. The triggering independent report is
`/tmp/gotth-bb-v4-independent-judge-10.md`, SHA-256
`e377a40d6149f4bc2b10c51af10ca7955ac3ccccdf6ed9c1e339b2fb0f872fa0`.
