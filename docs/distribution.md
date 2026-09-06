# Distribution Contract

## Endpoints

- Canonical development source:
  <https://git.dannyhunn.com/gotthboard/gotth-webhooks>
- Public clone and, only after implementation admission, future releases:
  <https://github.com/gotthboard/gotth-webhooks>
- Public bug tracker:
  <https://github.com/gotthboard/gotth-webhooks/issues>
- Private vulnerability reports:
  <https://github.com/gotthboard/gotth-webhooks/security/advisories/new>

Forgejo pushes one way to GitHub. GitHub does not feed commits or tags back to
Forgejo. A ref is distributed only when the exact object ID is visible at both
endpoints.

## Maturity and compatibility

Current status: implemented standalone technical admission candidate; no final
orchestrator admission, release, tag, public compatibility promise, or
real-consumer pin. Local green evidence does not itself admit the
implementation.

## Installation

There is no release to install. The candidate module path is
`github.com/gotthboard/gotth-webhooks`. A designated real-consumer adoption
candidate must use an exact source dependency pin for the required pre-release
behavioral verification; that pin is evidence, not a compatibility promise.
Other consumers must not treat the module as installable until all release
gates pass and an exact release tag exists.

The repository pins Go 1.26.6 where a Go module exists. Supported protocol,
runtime, database, and tool versions remain the ones stated in the README and
project verification documents; this distribution change does not widen those
contracts.

## Licensing gate

The maintainer selected MIT and the repository now carries the exact license
text. Release publication remains blocked on implementation admission,
consumer verification, clean exact-ref verification, and explicit release
authorization.

## Migration traceability

| Requirement | Repository implementation | Verification |
| --- | --- | --- |
| DIST-001 | Existing history, tags, worktrees, and mirror direction remain unchanged | pinned ref and worktree inventory |
| WHK-001..013 | Local outbound implementation remains unreleased and scoped to consumer-neutral mechanics | trace and verification audit |
| DIST-003/004 | README, contribution, security, changelog, and release contracts describe public use and support | documentation audit |
| WHK-013 | MIT decision is represented by the exact license file | license inventory |
| DIST-008 | Forgejo remains source and GitHub remains the one-way mirror target | push-mirror configuration and exact ref comparison |
