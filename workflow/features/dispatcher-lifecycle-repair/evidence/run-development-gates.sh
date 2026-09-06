#!/usr/bin/env bash
set -euo pipefail

repo=/tmp/gotth-webhooks-lifecycle-clean-20a1223
evidence=/tmp/gotth-webhooks-lifecycle-evidence-20a1223
source_commit=20a1223
bundle=/tmp/gotth-webhooks-lifecycle-20a1223.bundle
go=/tmp/gotth-webhooks-c92f9aa-receipt-time.cLMg65/go/bin/go
export PATH="$(dirname "$go"):$PATH"

rm -rf "$evidence"
mkdir -p "$evidence"

run_gate() {
  local name=$1
  shift
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
}

(
  printf 'source_commit=%s\n' "$(git -C "$repo" rev-parse HEAD)"
  printf 'source_tree=%s\n' "$(git -C "$repo" rev-parse HEAD^{tree})"
  printf 'go_version=%s\n' "$(go version)"
  printf 'kernel=%s\n' "$(uname -srmo)"
  printf 'bundle_sha256='; sha256sum "$bundle" | cut -d' ' -f1
) >"$evidence/metadata.txt"

run_gate full env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly -count=1 ./...
run_gate verify env GOTOOLCHAIN=local GOFLAGS= make verify
run_gate race-coverage env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly -count=1 -race -coverprofile="$evidence/coverage.out" ./...
(cd "$repo" && go tool cover -func="$evidence/coverage.out") >"$evidence/coverage-functions.txt"
run_gate full-race50 env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly -race -count=50 ./...
run_gate focused-race50 env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly -race -count=50 ./pkg/webhooks -run '^TestDispatcherClose'
run_gate fuzz-endpoint env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly ./pkg/webhooks -run '^$' -fuzz '^FuzzParseEndpoint$' -fuzztime=5s
run_gate fuzz-signing env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly ./pkg/webhooks -run '^$' -fuzz '^FuzzSignedRequestDeterministic$' -fuzztime=5s
run_gate fuzz-content-type env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly ./pkg/webhooks -run '^$' -fuzz '^FuzzCanonicalContentType$' -fuzztime=5s
run_gate benchmark env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly ./pkg/webhooks -run '^$' -bench '^(BenchmarkDeliver|BenchmarkClosedDeliver)$' -benchmem -benchtime=100ms -count=10

external=$(mktemp -d /tmp/gotth-webhooks-lifecycle-external.XXXXXX)
trap 'rm -rf "$external"' EXIT
cd "$external"
go mod init example.test/lifecycle-consumer >"$evidence/external-init.log" 2>&1
go mod edit -require=github.com/gotthboard/gotth-webhooks@v0.0.0
go mod edit -replace=github.com/gotthboard/gotth-webhooks="$repo"
cat >main.go <<'EOF'
package main

import (
  "context"
  "errors"

  webhooks "github.com/gotthboard/gotth-webhooks/pkg/webhooks"
)

type recorder struct{}

func (recorder) Record(context.Context, webhooks.Receipt) error { return nil }

func main() {
  dispatcher, err := webhooks.New(webhooks.Config{
    Secret: webhooks.Secret{KeyID: "key", Value: make([]byte, 32)},
    Recorder: recorder{},
  })
  if err != nil {
    panic(err)
  }
  dispatcher.Close()
  dispatcher.Close()
  _, err = dispatcher.Deliver(context.Background(), webhooks.Message{})
  if !errors.Is(err, webhooks.ErrClosed) {
    panic("closed classification missing")
  }
}
EOF
(
  printf 'gate=external-consumer\n'
  printf 'source_commit=%s\n' "$(git -C "$repo" rev-parse HEAD)"
  env GOTOOLCHAIN=local GOFLAGS= go test -mod=mod ./...
) >"$evidence/external-consumer.log" 2>&1

second=/tmp/gotth-webhooks-lifecycle-second-20a1223
rm -rf "$second"
git clone "$bundle" "$second" >"$evidence/clean-clone.log" 2>&1
git -C "$second" checkout --detach "$source_commit" >>"$evidence/clean-clone.log" 2>&1
(
  cd "$second"
  printf 'head_before=%s\n' "$(git rev-parse HEAD)"
  printf 'status_before_begin\n'
  git status --porcelain=v1 --untracked-files=all
  printf 'status_before_end\n'
  env GOTOOLCHAIN=local GOFLAGS= go test -mod=readonly -count=1 ./...
  printf 'head_after=%s\n' "$(git rev-parse HEAD)"
  printf 'status_after_begin\n'
  git status --porcelain=v1 --untracked-files=all
  printf 'status_after_end\n'
) >>"$evidence/clean-clone.log" 2>&1

git -C "$repo" status --porcelain=v1 --untracked-files=all >"$evidence/final-status.txt"
sed -i 's/[[:blank:]]*$//' "$evidence"/*.log "$evidence"/*.txt
sha256sum "$evidence"/* >"$evidence/SHA256SUMS"
printf 'evidence=%s\n' "$evidence"
