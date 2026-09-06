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

Two fresh independent reviews of exact evidence head `92c3de2` returned CLEAN.
The standalone technical implementation is admitted and the feature is `done`.
Technical admission does not release the module or promise compatibility. A
real-consumer contract, behavioral validation, and exact dependency pin remain
hard release/compatibility gates under `docs/RELEASING.md`. No forbidden action
occurred.

## Post-admission receipt-time repair

Independent review report `/tmp/gotth-bb-v4-independent-judge-10.md` identified
raw nanosecond receipt timestamps as incompatible with the exact PostgreSQL
microsecond contract. Source `c92f9aab1537bd49d035b7019ef7e00af44d5679`
uses the existing private clock injection and one private helper to apply UTC
conversion and `time.Truncate(time.Microsecond)` to all start/finish readings.

The retained expected-red log proves that delivered, permanent transport
failure, and recorder-failure paths exposed the sub-microsecond input before
the fix. The repaired regression proves exact UTC/microsecond callback and
result values while preserving clock-call count, ordering, and duration. On
development, exact Go 1.26.6-X:nodwarf5 full, verify, fresh race coverage, full
race50, focused race50, three fuzz, OpenSSL HMAC, external syntax compile,
performance, and second clean-clone gates all passed. Coverage is 97.3%; the
new helper is 100.0% covered. Exact commands, paths, hashes, fuzz counts,
performance results, and limitations are in `docs/verification.md` and
`docs/performance.md`.

Two fresh independent reviews of exact evidence head
`7d8a1b7011056894ee4f2cd91a7d2f74c0f5ccf7` returned CLEAN. The base feature
and repair are technically admitted and remain unreleased. Real-consumer
behavioral validation and the exact dependency pin remain separate hard release
and compatibility gates. No forbidden action occurred.
