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

Raw RFC text is retained as task scratch under
`/tmp/gotth-webhooks-authority`; the runtime boundary records consequential
conclusions.
