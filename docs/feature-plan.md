# Feature plan

| Slice | Requirements | Production units | Verification |
| --- | --- | --- | --- |
| Envelope | WHK-001/002/005/006 | values, URL normalization, ID generation | limits, malformed input, fuzz |
| Signature | WHK-004/010 | body digest, canonical bytes, HMAC headers | conformance vector and mutation tests |
| Network | WHK-002/003/008 | public-IP policy, resolving dialer, owned transport | prefix boundaries, rebinding sequence, TLS loopback harness |
| Delivery | WHK-007/008/009/012 | classification, bounded body, receipt, retry loop | all statuses, errors, cancellation, timeout, overflow |
| Lifecycle | WHK-011/014 | close admission flag, owned transport cleanup | closed precedence, admitted-call completion, concurrent/repeated close, race |
| Admission | WHK-011/013 | race, external consumer, docs, evidence | race repetition, coverage, fuzz, performance, review |

Only one slice may be unfinished. `workflow.toml` remains `in_progress` until
the implementation and local evidence pass and the orchestrator completes an
independent review. No local green run alone admits or releases the library.
