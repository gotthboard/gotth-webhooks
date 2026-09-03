# Verification status

Independent cold Judge pass 5 rejected exact head
`c30635434b88718f99ee5125052fdc309e737db9` for deadline/permanent error
precedence, incomplete network cost contracts, stale changelog accounting, and
four coverage gaps. Source commit
`f1fd980e8e6bd184815b871fb5b7513a999725d1` addresses those findings. A fresh
independent review remains required; this record does not claim admission.

Exact source-tip evidence under Go 1.26.6-X:nodwarf5, Linux amd64:

- `make verify` — PASS.
- Uncached `go test -mod=readonly -count=1 -race -coverprofile=... ./...` —
  PASS, 97.3% statements; all touched production functions are 100% covered.
- Focused cancellation/deadline/permanent-error, transient-DNS,
  MIME-expansion, capped-HTTP-date, and concurrency race suite repeated fifty
  times — PASS.
- Endpoint fuzz — PASS, 99,546 executions/5 seconds.
- Signing fuzz — PASS, 77,965 executions/5 seconds.
- Content-type fuzz — PASS, 91,802 executions/5 seconds.
- Independent OpenSSL HMAC-SHA-256 computation — PASS, exact vector
  `5afc952ae607736b86e3570dddc25991115f80afdb46c49832a30a5671028f7f`.
- Detached clone at the exact source object with a fresh empty `GOCACHE` ran
  `make verify` — PASS, clean detached worktree, 97.3% statements.
- Synthetic external module `go test -mod=readonly -count=1 ./...` — PASS.
  This proves public syntax only, not consumer behavior or compatibility.
- The 20 historical changelog records at the source object match the declared
  two-lineage order, exact Git timestamps, and exact affected-file sets; one
  current-commit placeholder is present as policy requires.
- All six retained RFC/IANA authority snapshots revalidated against their
  recorded SHA-256 values.
- The exact performance matrix and limitations are in `performance.md`; no
  speedup or production-latency claim is made.

Exact source artifacts and SHA-256 values:

- `/tmp/gotth-webhooks-f1fd980.verify.log` —
  `a68ff9bbb96d8682260a855a93f2b3e7188ffea9702a11582cda97d67f338445`.
- `/tmp/gotth-webhooks-f1fd980.coverage.out` —
  `7bf789422f4a247dc6cf3c57ca5a1d34add685aa235ba5f4da80736c5e37feef`.
- `/tmp/gotth-webhooks-f1fd980.coverage.log` —
  `9be78b6a981773b5a46ebccdc10b30bd37892aa7657386386958564b70160929`.
- `/tmp/gotth-webhooks-f1fd980.coverage.func` —
  `15f005a235c5683685ac58824aa61abf44c453d7b6ad18a64d005ca8697794ea`.
- `/tmp/gotth-webhooks-f1fd980.focused-race50.log` —
  `c470b715acea25d41b9eea1020ce8e1d239ad419712777894062ea5701c39062`.
- `/tmp/gotth-webhooks-f1fd980.fuzz-endpoint.log` —
  `f7434c2c758065553595926ad6a18159bfcc0b25a53d96e58f5d72a29e9fd30a`.
- `/tmp/gotth-webhooks-f1fd980.fuzz-signing.log` —
  `0c9f3d8d74f65663e423fc3170528c96184cf9a794824f59e5d715189222d900`.
- `/tmp/gotth-webhooks-f1fd980.fuzz-content-type.log` —
  `36f4bfb75c7a0d5e2e8d2a3cf892a6eeadc5acddecc3cfb4bc89eb36df0f3bf8`.
- `/tmp/gotth-webhooks-f1fd980.hmac-openssl.log` —
  `c23d27896db18640b7451655ae1ff339012f4cc79c8c708eb886a161f3160944`.
- `/tmp/gotth-webhooks-f1fd980.clean-clone.log` —
  `0b38c6a29eff11b4513f3555acee8b361ae8e95bb07b9ff41bf8e1dc6c9512ab`.
- `/tmp/gotth-webhooks-f1fd980.synthetic-consumer.log` —
  `c7a8c308ddd88137cf0bc359f1fd749625636c35e91a8216ee54b2a139080263`.
- `/tmp/gotth-webhooks-f1fd980.provenance.log` —
  `93638eff1049dbcc7ef5a4db8dd7726944168d48d6309af41739057549ca0023`.
- `/tmp/gotth-webhooks-f1fd980.authority-hashes.log` —
  `6895ab208fddb8977864210f09c009bae8fffe4c655689babf8f188713c3787a`.
- `/tmp/gotth-webhooks-f1fd980.benchmark.txt` —
  `d45c180a8261f2e7024b7eb7d351e71fe046e76cd5a5311e30c7293fa508d44f`.

The feature remains `in_progress` and unreleased. The newly authorized product
contract work is outside this source repair, and there is still no real
consumer adapter, behavioral proof, dependency pin, tag, or compatibility
promise. No push, PR, tag, release, deployment, or live request occurred.

PostgreSQL/database integration is N/A because this package has no adapter.
Graphify is N/A because the repair affects one direct classifier plus bounded
tests/comments and has no dependency ambiguity worth graph construction.
