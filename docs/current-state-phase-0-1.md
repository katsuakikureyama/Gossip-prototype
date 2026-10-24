# Current state — Phase 0–1

Last updated: 2026-10-24 (JST)

## Present source state

- A Go module, native Go Plugin SDK and sample plugin, build-information package,
  Go HTTP app with fixture/edit-view paths, release-slot manager and unit-test
  sources, Compose development kit, and WordPress oracle fixtures are present.
- Native hook ordering, the keyed development-only edit/view path, immutable A/B
  binary construction, failed-candidate preservation, and Slot B activation have
  passed the narrow tests recorded below. These results do not establish
  WordPress/PHP compatibility or production readiness.

## Verification record

Go implementation target: `4f2ae36757b6c0c34cefbf0ab0e684179c1cb246`.
Oracle/Compose configuration target after the internal-network URL correction:
`a0d3bf085332d914d4d886a1880468f62b675add`. Both are on branch
`codex/phase-0-1`; the later evidence-only commit does not alter either target.

| Command (2026-10-24 JST, one run for this target/config) | Result |
| --- | --- |
| `go test ./...` | PASS: Plugin SDK, HTML escaping/sample render hook, keyed development editor save, release failure preservation, B activation and tamper rejection |
| `sh scripts/demo-ab-build.sh` | PASS: real sample-plugin binary in A, failed B build leaves state byte-for-byte unchanged, verified B build activates |
| `sh scripts/check-oracle-fixture.sh` (configuration target `a0d3bf0`) | PASS: required fixture layout and non-floating version tags present |
| `docker compose --env-file .env.example config -q` (configuration target `a0d3bf0`) | PASS: Compose resolves and validates without pulling or running images |

No Docker image, WordPress/PHP process, MySQL database, third-party ZIP,
Composer script, production secret, or production data was executed or used.

## Documentation delivered in this change set

- Authoritative English project overview with a Japanese summary.
- Proposal-level architecture, security, licensing, Docker, CGI/Sakura, and
  roadmap documents.
- An evidence-led compatibility matrix that distinguishes untested work from a
  passing result.

Documentation is not evidence that a feature is safe for production.

## Unimplemented or unverified

- Production session/authorization, persistent content/database mapping, plugin
  intake, and isolated build execution. The current editor is a loopback-default,
  explicit development fixture protected only by a per-run key.
- Execution of the Compose kit against its version-pinned WordPress/MySQL/CLI/
  runner images.
- WordPress core, Gutenberg/block theme, `theme.json`, classic-theme, database,
  and third-party plugin compatibility.
- GoHP builder adapter and translated WordPress/PHP artifact execution.
- Any Sakura/CGI shared-host execution or resource measurement.
- Production migration, SaaS tenancy, payments, marketplace, public release,
  main-branch merge, and license finalization.

## Production migration requirements

No production database or secret may be used in Phase 0–1. A later migration
requires a source-version inventory, restorable snapshot, dry run against a
copy, validation against a pinned oracle, rollback evidence, and explicit owner
approval. This is a gate, not an implementation status.
