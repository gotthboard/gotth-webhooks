# Dispatcher lifecycle repair

Status: `in_progress`; source and development evidence complete, independent
review and admission remain orchestrator-owned.

Add the smallest public lifecycle boundary needed to retire one dispatcher:
reject new delivery admission, let admitted calls finish, and close the owned
transport's idle connections exactly once. This feature does not own consumer
generation policy, request tracking, cancellation, or remote publication.

Evidence belongs under `evidence/`; independent reviews belong under `review/`.

Source: `20a122362ede5c6c934d39e809ff52747887bd1f`

Verification: [evidence/verification.md](evidence/verification.md)
