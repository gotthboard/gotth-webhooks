# Coverage map

| Requirement | Design/spec | Implementation | Tests | Status |
| --- | --- | --- | --- | --- |
| WHK-001 | PRD/boundary | opaque message validation | public API, validation, fuzz | covered |
| WHK-002/003 | architecture/network | URL parser, address policy, safe dialer, transport | URL/IP boundaries, mixed answer, rebinding, TLS integration | covered |
| WHK-004/010 | spec/wire | signed request and key ID | OpenSSL vector, every-field mutation, rotation-independent fingerprint | covered |
| WHK-005/006 | spec/public API | random ID, limits, semantics fingerprint | uniqueness, limit boundaries, fuzz | covered |
| WHK-007/008 | failure model | retry, status, timeout, bounded response | status/delay/date/error/cancel/overflow tests | covered |
| WHK-009/012 | receipt model | detached record and minimal receipt | ordering, failure stop, unknown outcomes, fingerprint stability | covered |
| WHK-011 | concurrency | immutable dispatcher and standard client | race plus 50 concurrent calls | covered; repeated race pending exact commit |
| WHK-013 | distribution | `LICENSE` and policy docs | license inventory | covered |

Statement coverage is 96.9%. Residual defensive branches are listed in
`docs/verification.md`; percentage is iteration evidence, not the behavioral
oracle. There is no database, deployment, inbound handler, or product event
subsystem hidden outside this map.
