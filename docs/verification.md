# Verification status

## Admitted dispatcher lifecycle repair

Independent review 1 rejected evidence HEAD `065ba1d9` because cleanup could
precede an admitted delivery's first or retried `RoundTrip`. Repair source
`120db639e874419a71af4e10447b8e9ea1f73913` now serializes admission with the
closed transition, waits for all admitted deliveries, and then performs the
sole owned-transport cleanup. Dispatcher value copies share this lifecycle.
The review and expected-red proofs are retained with the repair evidence.

Exact Go 1.26.6 development-host evidence passed full, verify, race coverage,
full/focused race x50, all three existing fuzz targets, allocation, benchmark,
external-consumer, and second-clean-clone gates. Repository coverage is 97.5%;
`Close` and every changed lifecycle decision are 100.0% covered. The closed
path observed 0 B/op and 0 allocations/op.

Raw repair logs, runner, coverage profile, identities, hashes, limitations, and
the exact test oracle are retained under
`workflow/features/dispatcher-lifecycle-repair/evidence/repair-1`. Exact final
candidate `48f13fc30dcc5096c88f39da3890103ef581f55b`, tree
`4bf21b3ffeee8f35d5fd731b78b26f7abf0eea60`, passed a second complete
development-host gate after incorporating current canonical `main`. Those logs
and hashes are retained under
`workflow/features/dispatcher-lifecycle-repair/evidence/final-48f13fc`.
Independent reviews 2 and 3 returned CLEAN on that exact candidate, so the
feature is technically admitted. No remote mutation or release action occurred.

## Post-admission receipt-time repair

Independent review report `/tmp/gotth-bb-v4-independent-judge-10.md` found
that Dispatcher receipts exposed unquantized clock nanoseconds despite the
downstream exact PostgreSQL microsecond contract. The report's SHA-256 is
`e377a40d6149f4bc2b10c51af10ca7955ac3ccccdf6ed9c1e339b2fb0f872fa0`.
Source commit
`c92f9aab1537bd49d035b7019ef7e00af44d5679` uses the existing private clock
hook and converts every receipt start and finish reading with
`value.UTC().Truncate(time.Microsecond)` before recorder or result exposure.

The expected-red regression exercised delivered, permanent transport failure,
and receipt-recording failure paths with non-UTC, sub-microsecond clock values.
Before the source repair, all three paths retained nanoseconds; its retained
log has SHA-256
`125cbc71301011b2bfecff555bfbee4348ce904a1dff974845eba6d8f30e0fd2`.
After the repair, the same test proves exact UTC values, microsecond alignment,
callback/result identity, two clock reads, nonnegative ordering, and the
expected two-millisecond interval. Focused local tests and vet passed with
`GOMAXPROCS=2` and `-p=1`.

Development-host verification used Go 1.26.6-X:nodwarf5 on Linux amd64 in
isolated detached clone
`/tmp/gotth-webhooks-c92f9aa-receipt-time.cLMg65/repo`. The source bundle is
SHA-256 `71b672d8247743aa70609ea20a9ff11482cb0cef234e1d5080a933252fda69c2`.
The exact runner is retained as `gotth-webhooks-c92f9aa-run-gates.sh`, SHA-256
`ebe4fb352a579c961614091dfaef0597f0726610d5cd606f7082283f8711fad2`.
Every gate log records before/after HEAD at exact source
`c92f9aab1537bd49d035b7019ef7e00af44d5679`, empty before/after status, and
exit 0. The exact gates were:

- uncached `go test -mod=readonly -count=1 ./...`;
- `make verify`;
- uncached `go test -mod=readonly -count=1 -race
  -coverprofile=/tmp/gotth-webhooks-c92f9aa-receipt-time.cLMg65/c92f9aa.coverage.out
  ./...`;
- full repository `go test -mod=readonly -race -count=50 ./...`;
- focused `go test -mod=readonly -race -count=50 ./pkg/webhooks -run
  '^(TestReceiptTimesAreExactUTCMicroseconds|TestDeliverRetriesRecordsAndSucceeds|TestDeliverPermanentAndRedirectResponsesDoNotRetry|TestDeliverExhaustsAllowlistedTransportFailures|TestDeliverCancellationWhileReadingResponse|TestDeliverReceiptFailureStopsRetries)$'`;
- `FuzzParseEndpoint`, `FuzzSignedRequestDeterministic`, and
  `FuzzCanonicalContentType` with `-run '^$' -fuzztime=5s`;
- independent OpenSSL HMAC vector;
- disposable external-consumer syntax compile;
- ten 100 ms samples of `BenchmarkDeliver` with allocation reporting; and
- a second detached clean clone from the retained source bundle followed by
  an uncached full test and final clean-status assertion.

All gates passed. Fresh race coverage is 97.3% statements and
`canonicalReceiptTime` is 100.0% covered. Fuzz execution counts were 1,059,828
endpoint, 925,711 signing, and 1,139,852 content-type inputs. The external
consumer remains syntax evidence only; no real-consumer behavior or
compatibility claim follows from it.

Retained local copies are under
`/tmp/gotth-webhooks-c92f9aa-receipt-time-evidence`. Development artifacts and
SHA-256 values are:

- full / verify / race coverage
  `fdd3034a4b53c77a0eeb998be281be5703acc1af58ae5c2a66745d5a04724d21` /
  `d6c2080179097333418da01612719071809391b81068bb368f396c77ad64b8b0` /
  `30daa5ddbea4e220039fd5cf8f340e605196932bf0b6f42ed3cb0b5db9893150`;
- coverage profile / function report
  `a10415f1428d14224866e8441eb838fa0e72cb03c865e9b314cf6213feee36d0` /
  `1e0d5111bf66ba246f0b4c1f090675e2ee5b7608fae15eff02b945466ab80425`;
- full / focused race50
  `5354783bba07f618e0df4060724cc36e6750e3f0455594569010a5576f5f02e4` /
  `3cfefe6049eb93894660502aee9a376aaf6be2338890aaf2e9bdb2649a3bcefa`;
- fuzz endpoint / signing / content type
  `233af201e7250c68b9a16f479627245ad7267d4d2057743a9be008e8ac4f9fee` /
  `c2ac4ef0c1f854238a65ce36ec664957a5377418677f7e2f65bb6d5dbea4a282` /
  `260d26306e1d193421ad252b31f039390656a43e435e6644c4fd9e228f89b04a`;
- HMAC / external compile / benchmark / clean clone
  `19270d5e0c972b50007e0c369e6eddf8dadb42b4cba7acb62f3025bc314bc46e` /
  `353f5e64970032f1bcf524b7a8540a84cc4eaf6f154205536ae5cb40957281f9` /
  `c0ecd853e2b05cc0e016d1acfbdf30d7454ddabe6b565443d050709e0bc41987` /
  `2cffa6aa86a31e21a1048d5993ff8a4990767ff0064809de1121dd52e4b39409`.

Two fresh independent reviews of exact evidence head
`7d8a1b7011056894ee4f2cd91a7d2f74c0f5ccf7` returned CLEAN. Their canonical
records are `independent-14.md` and `independent-15.md`. The standalone base and
this post-admission compatibility repair are technically admitted. The
real-consumer behavior and exact dependency pin remain separate hard
release/compatibility gates under `docs/RELEASING.md`. No push,
merge, PR, tag, release, deployment, live request, remote-repository mutation,
or GOTTH Board mutation occurred.

Independent review of evidence head
`381b680e4b6e0945746d5d2db87b3a60bc797924` found that the copied compact
IANA prefix list was not an independent completeness oracle and that status
text incorrectly coupled standalone technical admission to consumer adoption.

Source commit `7de792774d2a9229f577d8d46dfb803b1b1c9bfe` retains the exact
already-hashed IANA IPv4 and IPv6 XML snapshots as test fixtures. The new test
verifies each fixture hash and registry metadata, parses all 26 IPv4 and 25
IPv6 allocations, and requires every allocation to be covered by the unchanged
compact production deny table. No test or runtime fetch occurs.

Two fresh independent reviews of exact evidence head
`92c3de21d512341cf862068f36c94ca7a70ed1b3` returned CLEAN. The standalone
technical implementation is admitted. A real-consumer contract, behavioral
validation, and exact dependency pin remain separate hard release/compatibility
gates under `docs/RELEASING.md`.

Exact source gates used Go 1.26.6-X:nodwarf5 on development Linux amd64 in the
isolated clean detached clone
`/tmp/gotth-webhooks-7de7927-repair.0tljnO/repo`:

- Uncached full repository test and `make verify` - PASS; format, vet, race,
  and 97.3% statements.
- Fresh uncached race coverage - PASS, 97.3%. Production code is unchanged;
  `isPublicAddress` and every other address-policy production function remain
  100% covered.
- Full repository race x50 and consequential focused race x50 - PASS.
- Endpoint/signing/content-type fuzzing for five seconds each - PASS,
  1,022,075 / 964,273 / 1,160,609 executions.
- Independent OpenSSL HMAC vector - PASS, exact expected digest
  `5afc952ae607736b86e3570dddc25991115f80afdb46c49832a30a5671028f7f`.
- Synthetic external compile and benchmark observation - PASS. The compile is
  syntax evidence only; the benchmark makes no speedup or production-latency
  claim.
- Every gate log records Go version, before/after HEAD
  `7de792774d2a9229f577d8d46dfb803b1b1c9bfe`, empty before/after tracked
  status, and exit 0. Final clone HEAD/status matched.

Retained development-host artifacts and SHA-256 values:

- full / verify / race coverage
  `a63dcb7f681bca8e961f2348fa7afc7c65d6d7652840ad31152b2a0a788a4fa5` /
  `dbb26bb8d1f8f168b00f3c4d8f2d287895b89f7693fb5215ca5e3cb146d1c6e9` /
  `98652860153e3e4864cabee8cf0102693ce6b83da0189db0d61ee08ff4582e64`;
- coverage profile / function report
  `cfb599f2e94dc506076cbca45243ace5b476d4fbf393833e365346bf4b937dcc` /
  `7d66e803dbdbe2f1748fd7951a64c268fff3109d8cf8e125dedfcace21957830`;
- full / focused race50
  `b4f57dc60a7704380d7340df2407fd00c91684e17f188792b17de10aa1497d2b` /
  `beb75b06a0426854288e55a919a50cd7acf3352e0b7178b7f974b6101458fa07`;
- fuzz endpoint / signing / content type
  `5f34127c695103fbc63b20ed91a010f2ba8bf1b1f061da668239cf56051ad587` /
  `ff050498584086f10d0be59311b99caa51e42efd9071b99da0b0759516040dd0` /
  `553dc4aaa05e05ddbdf3d0ea9d51d0134f183efd897d2f73fed394317975d0ea`;
- HMAC / synthetic compile / benchmark
  `47edb1a11d52294cb50f275d3239328f1c523ab69aabd7ef05ace704b2c81176` /
  `96f60988ac01f2d2add5a0c8ef515f859f3a82aa0f9b80aeea6d013b5fd3b7bc` /
  `73f7d021f17edfcc39c2f6dbe22e42faed12c346fd929ae8f6fa2114ddbe450a`.

The source bundle is SHA-256
`1e092901aac0dde3353afb5039b0e90b66df151a7de73fee7b0d60659cbb08e7`;
the independent review report is
`0fed933a862d1310ce09612f2cd7919bbc9848a0bf9f86855e8c950f2148c93c`.
No push, PR, tag, release, deployment, live request, remote repository
mutation, or GOTTH Board mutation occurred.
