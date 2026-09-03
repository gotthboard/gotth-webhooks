# Coverage map

| Requirement | Design/spec | Implementation | Tests | Status |
| --- | --- | --- | --- | --- |
| WHK-001 | PRD/boundary | opaque message validation | public API, validation, fuzz | candidate-covered |
| WHK-002/003 | architecture/network | URL/query parser, pinned IANA policy, safe dialer, transport | URL/query and registry boundaries, mixed answer, rebinding, TLS integration | candidate-covered |
| WHK-004/010 | spec/wire | signed request and key ID | OpenSSL vector, every-field mutation, rotation-independent fingerprint | candidate-covered |
| WHK-005/006 | spec/public API | random ID, limits, semantics fingerprint | uniqueness, limit boundaries, fuzz | candidate-covered |
| WHK-007/008 | failure model | typed retry allowlist, status, timeout, bounded response | malformed redirect, TLS/protocol/header, delay/date/error/cancel/overflow tests | candidate-covered |
| WHK-009/012 | receipt model | detached record and sensitive fingerprint | ordering, failure stop, unknown outcomes, fingerprint stability | candidate-covered |
| WHK-011 | concurrency | immutable dispatcher, concurrent Recorder obligation | overlapping recorder calls, concurrent deliveries, repeated race | candidate-covered |
| WHK-013 | distribution | `LICENSE` and policy docs | license inventory | candidate-covered |

These statuses cover only the local implementation candidate. Admission is
blocked because no real consumer requirement or pin validates the public API;
the synthetic compile fixture is not a compatibility oracle. Revised coverage
and exact-revision evidence are recorded only after the repair source passes.
