# Verification status

A fresh severe audit rejected prior evidence head
`ef713a777a4949d1f6726f72cabe7a009d667883`. Go 1.26.6 HTTP/2 can replay a
reconstructible POST inside one `RoundTrip`, producing a wire send without a
distinct library receipt. The same audit found zero-progress callbacks missing
from `io.ReadFull` cost accounting and a false timer wall-time upper bound.

Source commit `bf64d724bd68bcb95f7180e1b79c16351e1d881c` pins the owned
transport to HTTP/1, proves negotiation against an HTTP/2-capable TLS server,
and corrects both comments. This is a repair and verification claim only. The
feature remains `in_progress`; orchestrator final admission and a real consumer
contract/pin remain open.

Exact source gates used Go 1.26.6-X:nodwarf5 on development Linux amd64 in an
isolated clean detached clone:

- `make verify` - PASS; format, vet, race, 97.3% statements.
- Uncached race coverage - PASS, 97.3%. `newHTTPTransport` and every other
  touched production function are 100%; no changed-surface statement gap.
- Full repository race x50 and consequential focused race x50 - PASS.
- Endpoint/signing/content-type fuzzing for five seconds each - PASS,
  1,120,515 / 946,687 / 1,168,851 executions.
- Independent OpenSSL HMAC vector - PASS, exact expected digest
  `5afc952ae607736b86e3570dddc25991115f80afdb46c49832a30a5671028f7f`.
- Synthetic external compile, benchmark observation, six authority hashes, and
  post-gate clean tracked status - PASS. The compile is syntax evidence only;
  the benchmark makes no speedup or production-latency claim.

Retained artifacts and SHA-256 values:

- verify `fbeaaaff513f31d76a852723fdeb791d1cae4b6eee7a4f008f1d94ca9b0262f2`;
- coverage/profile/function `698169dd8f21f1a9aa9b8f9151ae79622b71b2209c5f888ac979c120be0a6087` /
  `cfb599f2e94dc506076cbca45243ace5b476d4fbf393833e365346bf4b937dcc` /
  `7d66e803dbdbe2f1748fd7951a64c268fff3109d8cf8e125dedfcace21957830`;
- full/focused race50 `b040439bc51fa63318081001766cc6ae9883e47ae2706e65b217372ce6213cbd` /
  `871f9e664d1fb90d18eb92fdc61b97fd7310825f436e1055921e381a832d5a60`;
- fuzz endpoint/signing/content type `38223a9490e99b8c4a2d355fc076c972389e366056dfeb0033741e2531f19d2b` /
  `a1532fe2488478c0997318ced54d8f6fd0ed7b12a984b69c425745517ddc633e` /
  `3ded5fe730aa5f315966ab30db9084a78536e38776b0057194d796a5ed7bba14`;
- HMAC `952a876c8b7a817fa39f94cf08a16aecdc67173cb7c2960d9c5a69e489c58f1d`;
- synthetic compile `e327314e3c80bd9af6dc21df651c8c7fbc7cb901975d051a1603e039dc91bb73`;
- benchmark `8d67754d4ccd68ab227d770656a793d5799eef3dafce2200f5f1a00d0632eaeb`;
- Go contracts `ff1e886333ff13e9c5fd491732782e8c2346d455a6fbd3414a01121cba0146e3`;
- authority/provenance `4c4d82371b2c027cfc90c47a287106fe3251fbdb4a042927154b77010ed76c8a` /
  `2e3caf69710a01a744163aea1b9c987286ca665073497bbc622e6f8ebb0d2ff6`.

Artifacts use `/tmp/gotth-webhooks-bf64d72.<name>`. No push, PR, tag, release,
deployment, live request, remote mutation, or GOTTH Board mutation occurred.
Final admission is deliberately not claimed.
