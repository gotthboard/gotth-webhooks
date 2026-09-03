# Verification status

Independent review rejected the prior candidates through exact head
`ff4fee3f5aa94cb5b44694050188793e5b181a8d`. The last two findings were MIME
control-byte acceptance and false historical changelog provenance. Source
commit `be6f7150f3419ff3bcb971476119e7cc3f2cb10b` rejects all ASCII C0 controls
and DEL before MIME parsing. Exact candidate
`94f2b7d080d9959af3c1d23a5a4349b555ee240b` additionally splits and corrects
the historical records against Git. It contains the complete local technical
repair; the separate consumer blocker remains open.

Exact source-tip evidence under Go 1.26.6-X:nodwarf5, Linux amd64:

- `make verify` — PASS.
- Uncached `go test -mod=readonly -count=1 -race -coverprofile=... ./...` —
  PASS, 96.5% statements.
- Fifty uncached full race runs — PASS.
- Content-type fuzz — PASS, 115,255 executions/5 seconds.
- Signing fuzz — PASS, 82,089 executions/5 seconds.
- Detached clone at the exact object with a new empty `GOCACHE` ran
  `make verify` — PASS, clean worktree, 96.5% statements.
- Synthetic external module `go test -mod=readonly -count=1 ./...` — PASS.
- Exact performance matrix and raw hash are in `performance.md`; it is an
  uncontrolled-host observation and makes no speed claim.
- Every changelog heading time and exact file list was mechanically compared
  with Git for all 16 named commit records — PASS.
- Internal cold passes found and repaired ASCII-OWS/concurrency wording, false
  permanent-status reporting, exhausted-dial classification, exact signed-target
  specification, explicit-empty-port validation, cost contracts, content-type
  control handling, and provenance. Fresh source pass 13 is CLEAN for the
  technical scope.

Coverage artifacts and hashes:

- `/tmp/gotth-webhooks-94f2b7d.coverage.out` —
  `d726bb571e9c7aaa908017281cf2e939ceb441d9dee0458292585ac537d206f4`.
- `/tmp/gotth-webhooks-94f2b7d.coverage.log` —
  `0ea7a506966f3668504a627d4eee3385245895feae54a3f80008aec91ef1da57`.
- `/tmp/gotth-webhooks-94f2b7d.coverage.func` —
  `bd04b3a37fc42b96b3a1f891327d9606b51ecdb03031ae1b3edc44ca6cee0ef8`.
- `/tmp/gotth-webhooks-94f2b7d.race50.log` —
  `06a0544953f7e992d1d81f4fa435b4ace668f4e9e22a96a97497a9c13f266b2c`.
- `/tmp/gotth-webhooks-94f2b7d.fuzz-content-type.log` —
  `8b950e8ba634b6a64e87d9ec35e48671196975534ed789d368e73d5870674fa7`.
- `/tmp/gotth-webhooks-94f2b7d.fuzz-signing.log` —
  `e24e14b88b3e57395d5f48742d131b0bc0731b7081566383f9f5d089619290ce`.

The external module at `/tmp/gotth-webhooks-consumer` is a synthetic compile
fixture. It proves public syntax only; it is not a real product consumer,
behavioral oracle, compatibility pin, or answer to the admission blocker.

The feature remains `in_progress` and unreleased. No local green result can
admit, tag, release, push, deploy, or establish compatibility without an actual
consumer requirement and pin.

PostgreSQL/database integration is N/A because there is no adapter. Graphify is
N/A because one compact package has direct source/test correspondence and no
dependency ambiguity that justifies index/cache cost.
