# Outbound V1 admission

This feature replaces the placeholder with the first consumer-neutral outbound
webhook implementation. `workflow.toml` remains the state authority. The
standalone technical implementation is admitted after implementation,
evidence, and independent orchestrator review completed. The first
orchestrator review rejected commit `a5a3b10`; later reviews produced bounded
source and evidence repairs. Two fresh independent reviews of exact evidence
head `92c3de2` returned CLEAN, so the feature is `done`.

Technical admission does not release the module or promise compatibility. Per
`docs/RELEASING.md`, a real-consumer contract, behavioral validation, and
exact dependency pin remain hard release/compatibility gates. That
owner/product work is open, but it is not a blocker to standalone technical
implementation admission.

Post-admission source `c92f9aab1537bd49d035b7019ef7e00af44d5679`
repairs the receipt persistence contract by converting each produced start and
finish timestamp to UTC and truncating it to exact microsecond precision. This
repair is an unreleased review candidate pending two fresh independent reviews;
the existing `done` state describes the admitted base and is not an admission
claim for this candidate. Real-consumer behavioral validation and the exact pin
remain separate release/compatibility gates.
