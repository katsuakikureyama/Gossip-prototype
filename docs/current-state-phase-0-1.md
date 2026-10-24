# Current state — Phase 0–1

Last updated: 2026-10-24 (JST)

## Present source state

- Native string-filter dispatch now has a real WordPress differential PASS:
  five owned cases, two calls each, against WordPress 6.6.2 / PHP 8.3.13.
  Target `df11d5e680a849a526e14f35383447c5fc13d8bf`; command, image/fixture hashes
  and exclusions are in [the hook report](evidence/hook-differential-2026-10-24.md).

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
`codex/phase-0-1`. The full disposable runtime smoke used head
`e22b21a90f7c9da08c6b3f40f4ad73c592ce2fd4`; its exact image digests and
cleanup are in [the oracle evidence record](evidence/oracle-smoke-2026-10-24.md).

| Command (2026-10-24 JST, one run for this target/config) | Result |
| --- | --- |
| `go test ./...` | PASS: Plugin SDK, HTML escaping/sample render hook, keyed development editor save, release failure preservation, B activation and tamper rejection |
| `sh scripts/demo-ab-build.sh` | PASS: real sample-plugin binary in A, failed B build leaves state byte-for-byte unchanged, verified B build activates |
| `sh scripts/check-oracle-fixture.sh` (configuration target `a0d3bf0`) | PASS: required fixture layout and non-floating version tags present |
| `docker compose --env-file .env.example config -q` (configuration target `a0d3bf0`) | PASS: Compose resolves and validates without pulling or running images |
| isolated Compose build/up + `wordpress-bootstrap` + `compat-runner` (runtime head `e22b21a`) | PASS after one network-stalled builder pull was canceled and retried; pinned WordPress oracle hooks, Gutenberg, block theme/`theme.json`, post and category markers passed; Gossip comparison explicitly UNVERIFIED |
| Gossip container `healthz` / version / plugin inventory / welcome render | PASS: `ok`, development build identity, sample manifest and plugin footer observed |
| `docker compose ... down --volumes --remove-orphans` | PASS: validation containers, network and two disposable volumes removed |

Only the recorded pinned WordPress/PHP/MySQL/CLI/curl images and owned fixture
were executed. No third-party plugin/theme ZIP, Composer script, production
secret, production database, or production data was used.

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
- Production WordPress adapter and HTTP compatibility contract. The local
  `gossip.hook-fixture/v1` harness compares only owned string-filter dispatch;
  the separate HTTP runner still provides oracle-only evidence.
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
