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

Raw RFC and IANA registry snapshots are retained as task scratch under
`/tmp/gotth-webhooks-authority`; the runtime boundary records consequential
conclusions.

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
