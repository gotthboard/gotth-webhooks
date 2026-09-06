# Dispatcher lifecycle repair

Status: `in_progress`; independent review 1 rejected the first candidate and
the drain-before-cleanup repair is active.

Add the smallest public lifecycle boundary needed to retire one dispatcher:
reject new delivery admission, wait for admitted calls to finish, and only then
close the owned transport's idle connections exactly once. This feature does
not own consumer generation policy, per-request identity/cancellation, or
remote publication.

Evidence belongs under `evidence/`; independent reviews belong under `review/`.

Source: `20a122362ede5c6c934d39e809ff52747887bd1f`

Verification: [evidence/verification.md](evidence/verification.md)
