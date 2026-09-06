#!/usr/bin/env bash
set -euo pipefail

repo=/tmp/gotth-webhooks-lifecycle-red.WUEJ9f
log=/tmp/gotth-webhooks-lifecycle-expected-red.log
go=/tmp/gotth-webhooks-c92f9aa-receipt-time.cLMg65/go/bin/go

set +e
(
  cd "$repo"
  printf 'production_head=065ba1d9f47219aab7e3664c275c8f9ba9b60838\n'
  printf 'review_head=65f8efa\n'
  printf 'go_version='; "$go" version
  "$go" test -mod=readonly -count=1 ./pkg/webhooks -run '^(TestCloseWaitsForAdmittedDeliveryBeforeFirstRoundTrip|TestCloseWaitsForAdmittedDeliveryBetweenRetries|TestDispatcherCopiesShareLifecycle)$'
) >"$log" 2>&1
rc=$?
set -e
cat "$log"
if [ "$rc" -eq 0 ]; then
  printf 'expected regressions to fail against rejected production\n' >&2
  exit 1
fi
printf 'expected_red_exit=%d\n' "$rc"
