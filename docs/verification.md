# Verification status

Independent review of evidence head
`381b680e4b6e0945746d5d2db87b3a60bc797924` found that the copied compact
IANA prefix list was not an independent completeness oracle and that status
text incorrectly coupled standalone technical admission to consumer adoption.

Source commit `7de792774d2a9229f577d8d46dfb803b1b1c9bfe` retains the exact
already-hashed IANA IPv4 and IPv6 XML snapshots as test fixtures. The new test
verifies each fixture hash and registry metadata, parses all 26 IPv4 and 25
IPv6 allocations, and requires every allocation to be covered by the unchanged
compact production deny table. No test or runtime fetch occurs.

The feature remains `in_progress`; final standalone technical admission is
orchestrator-owned and is not claimed here. A real-consumer contract,
behavioral validation, and exact dependency pin remain separate hard
release/compatibility gates under `docs/RELEASING.md`.

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
