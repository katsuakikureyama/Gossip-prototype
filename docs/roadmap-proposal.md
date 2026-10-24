# Phase 0–1 roadmap — PROPOSAL

1. **Starter runtime:** minimal Go HTTP app, fixture content, development-only
   keyed edit/view smoke path, versioned build ID, native sample plugin.
2. **Plugin/release contract:** native manifest/hook interface, non-executing
   inventory, immutable candidate build, Slot A/B selection, failed-build
   preservation and rollback demonstration.
3. **Oracle evidence:** pinned WordPress/MySQL Docker fixture, named comparisons,
   compatibility report registry/matrix.
4. **GoHP boundary:** implement the proposed `gossip.builder/v0.1` isolated
   candidate adapter only after pinning the draft source ref; emit hashes,
   toolchain data, and `rejected`/`generated`/failure result statuses. Make
   unsupported WordPress/Composer/database/extensions behavior explicit.
5. **CGI feasibility:** build off-host and run the Sakura checklist on an
   approved test host, never production.

| Gate | Required evidence | Does not prove |
| --- | --- | --- |
| Native build | **Phase 0 narrow PASS at `4f2ae36`:** sample pipeline, failed candidate leaves Slot A unchanged, Slot B selection | WordPress/PHP compatibility or production-safe plugin isolation |
| Oracle | Pinned reference plus named reports | Broad WordPress parity |
| Docker kit | Repeatable disposable-fixture run | Production readiness |
| GoHP adapter | Pinned builder report + isolated test | Arbitrary WP plugin/theme support |
| CGI | Actual-host checklist and measurements | All Sakura plan support |

Coordinate Gossip and GoHP with versioned interfaces, fixtures, hashes, and
explicit issues/PRs. Do not block native SDK work on GoHP or add WordPress
behavior to its compiler core. GoHP's tag, license, and stable API are pending;
the only current candidate is draft PR #2 `feat/phase0-transpiler@56234c7`.
Marketplace/billing, public SaaS, automatic
arbitrary activation, production migration, and a full-clone promise are
deferred pending owner approval and compatibility/security/licensing evidence.
