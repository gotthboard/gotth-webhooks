# Dispatcher lifecycle repair

Status: `done`. Independent review 1 rejected the first candidate. The
drain-before-cleanup repair closed that race; independent reviews 2 and 3 then
returned CLEAN on exact final candidate `48f13fc`, so the repair is technically
admitted.

Add the smallest public lifecycle boundary needed to retire one dispatcher:
reject new delivery admission, wait for admitted calls to finish, and only then
close the owned transport's idle connections exactly once. This feature does
not own consumer generation policy, per-request identity/cancellation, or
remote publication.

Evidence belongs under `evidence/`; independent reviews belong under `review/`.

Source: `120db639e874419a71af4e10447b8e9ea1f73913`

Initial verification: [evidence/verification.md](evidence/verification.md)

Review repair verification:
[evidence/repair-1/verification.md](evidence/repair-1/verification.md)

Final admission verification:
[evidence/final-48f13fc/verification.md](evidence/final-48f13fc/verification.md)

Technical admission does not release the module or promise compatibility. A
real-consumer contract, behavioral validation, exact dependency pin, and the
release gates remain required.
