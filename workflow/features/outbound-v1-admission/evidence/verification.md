# Local repair verification evidence

Independent review rejected evidence head
`381b680e4b6e0945746d5d2db87b3a60bc797924` because the IANA deny-table
test duplicated production's compact list instead of deriving completeness
from the registries. It also found that status text incorrectly made a
real-consumer pin a prerequisite for standalone technical admission.

Source repair `7de792774d2a9229f577d8d46dfb803b1b1c9bfe` commits the exact
already-hashed IPv4 and IPv6 registry XML snapshots under
`pkg/webhooks/testdata`. Tests verify fixture hashes and 2025-10-09 registry
metadata, parse all 26 IPv4 and 25 IPv6 prefixes, and prove each is covered by
the unchanged compact production deny table. No fetch occurs in tests or
runtime.

On development under exact Go 1.26.6-X:nodwarf5, an isolated clean detached
clone passed the uncached full suite, `make verify`, uncached race coverage
(97.3%), full race x50, consequential focused race x50, three five-second fuzz
targets, the OpenSSL vector, external syntax compile, and benchmark
observation. Every gate retained before/after HEAD and clean-status traces.
Exact artifact paths, hashes, coverage gaps, and limitations are in
`docs/verification.md`.

The source repair is technically clean in this worker audit, but final
standalone technical admission remains orchestrator-owned and the feature
stays `in_progress`. Technical admission does not release the module or
promise compatibility. A real-consumer contract, behavioral validation, and
exact dependency pin remain hard release/compatibility gates under
`docs/RELEASING.md`. No forbidden action occurred.
