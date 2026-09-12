#!/usr/bin/env bash
set -euo pipefail

scratch=$1
candidate=$2
bundle=$scratch/candidate.bundle
repo=$scratch/repo
evidence=$scratch/evidence
export PATH=/tmp/gotth-webhooks-c92f9aa-receipt-time.cLMg65/go/bin:$PATH

mkdir -p "$evidence"
git clone "$bundle" "$repo" >"$evidence/clone.log" 2>&1
git -C "$repo" checkout --detach "$candidate" >>"$evidence/clone.log" 2>&1

run_gate() {
  local name=$1
  shift
  printf 'START %s\n' "$name"
  (
    set -euo pipefail
    cd "$repo"
    printf 'gate=%s\n' "$name"
    printf 'started_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    printf 'head_before=%s\n' "$(git rev-parse HEAD)"
    printf 'status_before_begin\n'
    git status --porcelain=v1 --untracked-files=all
    printf 'status_before_end\n'
    "$@"
    printf 'head_after=%s\n' "$(git rev-parse HEAD)"
    printf 'status_after_begin\n'
    git status --porcelain=v1 --untracked-files=all
    printf 'status_after_end\n'
    printf 'finished_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  ) >"$evidence/$name.log" 2>&1
  printf 'PASS %s\n' "$name"
}

(
  printf 'candidate=%s\n' "$(git -C "$repo" rev-parse HEAD)"
  printf 'tree=%s\n' "$(git -C "$repo" rev-parse HEAD^{tree})"
  printf 'bundle_sha256='; sha256sum "$bundle" | cut -d' ' -f1
  printf 'go_version=%s\n' "$(go version)"
  printf 'kernel=%s\n' "$(uname -srmo)"
) >"$evidence/metadata.txt"

run_gate diff-check git diff --check 60dcd948c81f2ece3aa5c6b13876021a56341329..HEAD
run_gate full env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly -count=1 ./...
run_gate verify env GOTOOLCHAIN=local GOFLAGS= make verify
run_gate race-coverage env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly -count=1 -race -coverprofile="$evidence/coverage.out" ./...
(cd "$repo" && go tool cover -func="$evidence/coverage.out") >"$evidence/coverage-functions.txt"
run_gate full-race50 env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly -race -count=50 ./...
run_gate focused-race50 env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly -race -count=50 ./pkg/webhooks -run '^(TestDispatcherClose.*|TestCloseWaitsForAdmittedDelivery.*|TestDispatcherCopiesShareLifecycle)$'
run_gate allocation-proof env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly -count=50 ./pkg/webhooks -run '^TestDispatcherClosedPathAllocations$'
run_gate fuzz-endpoint env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly ./pkg/webhooks -run '^$' -fuzz '^FuzzParseEndpoint$' -fuzztime=5s
run_gate fuzz-signing env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly ./pkg/webhooks -run '^$' -fuzz '^FuzzSignedRequestDeterministic$' -fuzztime=5s
run_gate fuzz-content-type env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly ./pkg/webhooks -run '^$' -fuzz '^FuzzCanonicalContentType$' -fuzztime=5s
run_gate benchmark env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly ./pkg/webhooks -run '^$' -bench '^(BenchmarkDeliver|BenchmarkClosedDeliver)$' -benchmem -benchtime=100ms -count=10

git -C "$repo" status --porcelain=v1 --untracked-files=all >"$evidence/final-status.txt"
sha256sum "$evidence"/* >"$evidence/SHA256SUMS"
printf 'EVIDENCE %s\n' "$evidence"
