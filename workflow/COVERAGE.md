# Coverage map

| Requirement | Design/spec | Implementation | Tests | Status |
| --- | --- | --- | --- | --- |
| WHK-001 | PRD/boundary | opaque message validation | public API, validation, fuzz | candidate-covered |
| WHK-002/003 | architecture/network | URL/query parser, pinned IANA policy, safe dialer, transport | URL/query and registry boundaries, mixed answer, transient DNS lookup, rebinding, TLS integration | candidate-covered |
| WHK-004/010 | spec/wire | signed request, canonical content type, and key ID | OpenSSL vector, every-field mutation, control rejection, spaces/UTF-8 canonical signature, rotation-independent fingerprint | candidate-covered |
| WHK-005/006 | spec/public API | random ID, limits, content-type controls, semantics fingerprint | uniqueness, input and canonical-output length boundaries, all C0/DEL positions, no-side-effect Deliver rejection, fuzz | candidate-covered |
| WHK-007/008 | failure model | typed retry allowlist, status, timeout, bounded response | malformed redirect, coincident deadline/permanent precedence, caller cancellation precedence, TLS/protocol/header, capped delta/date, error/cancel/overflow tests | candidate-covered |
| WHK-009/012 | receipt model | detached record and sensitive fingerprint | ordering, failure stop, unknown outcomes, fingerprint stability | candidate-covered |
| WHK-011 | concurrency | immutable dispatcher, concurrent Recorder obligation | overlapping recorder calls, concurrent deliveries, repeated race | candidate-covered |
| WHK-013 | distribution | `LICENSE` and policy docs | license inventory | candidate-covered |

These statuses cover only the local implementation candidate. Admission is
blocked because no real consumer requirement or pin validates the public API;
the synthetic compile fixture is not a compatibility oracle. Revised coverage
is 97.3% at exact source object
`f1fd980e8e6bd184815b871fb5b7513a999725d1`; each production function touched
by the latest repair is 100% covered. Exact-revision commands and
artifact hashes are recorded in the feature evidence.
