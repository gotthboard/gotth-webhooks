# Local verification evidence

Pre-commit candidate commands:

- `make verify` — PASS under Go 1.26.6-X:nodwarf5.
- `go test -race -coverprofile=/tmp/gotth-webhooks-coverage-final.out ./...`
  — PASS, 96.9% statements.
- Endpoint fuzz — PASS, 95,312 executions in 5 seconds.
- Signing fuzz — PASS, 85,673 executions in 5 seconds.
- External consumer `go test -mod=readonly ./...` in
  `/tmp/gotth-webhooks-consumer` — PASS.
- Benchmark command and raw hash are in `docs/performance.md`.

Repeated race, clean clone, final artifact hashes, and cold review are pending
until a coherent local commit exists. This evidence does not mark the feature
done.
