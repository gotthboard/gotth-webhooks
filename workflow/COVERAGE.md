# Coverage map

| Requirement | Design/spec | Implementation | Tests | Status |
| --- | --- | --- | --- | --- |
| WHK-001 | PRD/boundary | opaque message validation | public API, validation, fuzz | admitted-covered |
| WHK-002/003 | architecture/network | URL/query parser, pinned IANA policy, safe dialer, HTTP/1-only transport | URL/query boundaries, exact hashed IANA XML fixtures parsed allocation-by-allocation against the compact deny table, mixed answer, transient DNS lookup, rebinding, TLS integration, HTTP/2-capable server protocol pin | admitted-covered |
| WHK-004/010 | spec/wire | signed request, canonical content type, and key ID | OpenSSL vector, every-field mutation, control rejection, spaces/UTF-8 canonical signature, rotation-independent fingerprint | admitted-covered |
| WHK-005/006 | spec/public API | random ID, limits, content-type controls, semantics fingerprint | uniqueness, input and canonical-output length boundaries, all C0/DEL positions, no-side-effect Deliver rejection, fuzz | admitted-covered |
| WHK-007/008 | failure model | typed retry allowlist, causal timeout, status, bounded response | actual transport malformed/header-limit failures at expired deadline, causal deadline, caller cancellation precedence, TLS, capped delta/date, error/cancel/overflow tests | admitted-covered |
| WHK-009/012 | receipt model | detached record, exact UTC microsecond timestamps, and sensitive fingerprint | ordering, duration, UTC/microsecond alignment across success, transport failure, and recorder failure, failure stop, unknown outcomes, fingerprint stability | base admitted; repair review pending |
| WHK-011 | concurrency | immutable delivery configuration, concurrent Recorder obligation | overlapping recorder calls, concurrent deliveries, repeated race | admitted-covered |
| WHK-013 | distribution | `LICENSE` and policy docs | license inventory | admitted-covered |
| WHK-014 | lifecycle | atomic close admission and one owned-transport cleanup | closed precedence/no-side-effects, admitted-call completion, concurrent/repeated close, closable/non-closable transport branches, focused race x50 | candidate-covered; review pending |

These statuses cover the admitted standalone technical implementation except
for the explicitly marked post-admission receipt-time repair candidate. Two
fresh independent reviews of exact evidence head `92c3de2` returned CLEAN. A
real-consumer contract, behavioral validation, and exact dependency pin are
separate hard release/compatibility gates; the synthetic compile fixture is not
a compatibility oracle. Coverage is 97.3% at exact source object
`7de792774d2a9229f577d8d46dfb803b1b1c9bfe`. Production code is unchanged by
the fixture repair, and every address-policy production function remains at
100% statement coverage. Exact-revision commands and artifact hashes are
recorded in the feature evidence. The receipt-time repair has fresh 97.3%
statement coverage at exact source `c92f9aab1537bd49d035b7019ef7e00af44d5679`;
its canonicalization helper is 100.0% covered and two independent reviews
admitted it at `2b10dc0`. The dispatcher lifecycle candidate has fresh 97.4%
repository coverage at exact source
`20a122362ede5c6c934d39e809ff52747887bd1f`; `Close` and every changed
lifecycle decision are 100.0% covered. Its independent review remains
orchestrator-owned.
