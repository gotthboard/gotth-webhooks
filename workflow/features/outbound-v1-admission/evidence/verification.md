# Local repair verification evidence

Fresh severe audit rejected `ef713a777a4949d1f6726f72cabe7a009d667883`
for unreceipted Go HTTP/2 internal POST replay plus false zero-progress entropy
and timer cost contracts. Source repair
`bf64d724bd68bcb95f7180e1b79c16351e1d881c` pins HTTP/1 and corrects the
contracts without changing workflow admission state.

On development under the copied exact Go 1.26.6-X:nodwarf5 toolchain, an
isolated clean detached clone passed `make verify`, uncached race coverage
(97.3%), full race x50, consequential focused race x50, three five-second fuzz
targets, the OpenSSL vector, external syntax compile, benchmark observation,
authority hashes, and clean-status checks. The changed executable function
`newHTTPTransport` is 100% covered; all comments-only touched functions are
also 100%. Exact artifact hashes and limitations are in `docs/verification.md`.

The bounded repair is source-clean in this audit but does not admit the
feature. A real consumer contract, behavioral validation, exact dependency pin,
and orchestrator final admission remain required. No forbidden action occurred.
