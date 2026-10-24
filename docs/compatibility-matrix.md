# Compatibility matrix

Compatibility is per pinned source version, platform, configuration, Gossip
revision, GoHP revision (if used), source/build hash, and test command. `PASS`
requires stored evidence; `generated` and `compiled` are not `PASS`.

| Surface | Reference | Current status | Evidence / limitation |
| --- | --- | --- | --- |
| Local native editor browser flow | Gossip runtime `19aa421`, harness `6d09fe5` | SMOKE PASS | Chrome form/save/303/view, reload, escaping and rejected writes; [evidence](evidence/browser-editor-2026-10-24.md); no restart persistence |
| Native Go Plugin SDK | Gossip-owned `v1alpha1` package | SMOKE PASS | `go test ./...` at `4f2ae36`; validates manifests, atomic registration and priority/registration ordering only |
| Fixture edit/view HTTP path | Gossip-owned in-memory fixture | SMOKE PASS | `go test ./...` at `4f2ae36`; escaping, sample render hook, disabled/keyed editor and save path only; not production auth/persistence |
| Native plugin build failure preserves Slot A | Release manager | SMOKE PASS | unit test plus `sh scripts/demo-ab-build.sh` at `4f2ae36`; failed real Go build leaves release state byte-for-byte unchanged |
| Native candidate Slot B selection | Release manager | SMOKE PASS | real sample-plugin binaries built with stamped IDs; verification, SHA-256, activation and tamper rejection covered at `4f2ae36` |
| Compose/oracle development kit | `.env.example`, Compose and owned fixtures | SMOKE PASS | static checks at `a0d3bf0`; disposable build/up/bootstrap/runner/cleanup at `e22b21a`; exact digests in the [evidence record](evidence/oracle-smoke-2026-10-24.md) |
| Pinned WordPress oracle | `wordpress:6.6.2-php8.3-apache` owned fixture | SMOKE PASS (oracle only) | WordPress installed and named markers passed; no Gossip differential/general core claim |
| Native string-filter dispatch | WordPress 6.6.2 real hook engine | DIFFERENTIAL PASS (five owned cases) | Priority, ties, empty hook, 0/999 boundaries and Unicode; two invocations each at `df11d5e`; [contract/evidence/exclusions](evidence/hook-differential-2026-10-24.md) |
| General WordPress hook behavior | WordPress 6.6.2 | UNTESTED | Dynamic registration/removal, recursion, callback identity, references, multiple arguments and core filter stack are not covered |
| Gutenberg block markup | WordPress 6.6.2 owned fixture | SMOKE PASS (oracle only) | Owned post stored, returned by REST and rendered; no Gossip block runtime exists |
| `theme.json` and block themes | WordPress 6.6.2 owned fixture | SMOKE PASS (oracle only) | Owned block theme activated and its named style/title markers rendered; no Gossip renderer exists |
| Templates, parts, `wp_template` overrides | Pinned release TBD | UNTESTED | No fixture/DB comparison |
| Dynamic blocks | Pinned release TBD | UNTESTED | No callback/runtime comparison |
| Classic themes / `functions.php` | Pinned theme/WP TBD | UNTESTED | No PHP claim |
| `wp_posts` and terms oracle observations | Owned disposable fixture | SMOKE PASS (oracle only) | Owned post/category created and queried through WordPress; no DB-copy/Gossip import test |
| WordPress DB/options import | Snapshot fixture TBD | UNTESTED | No DB-copy migration or Gossip mapping test |
| Owned WordPress fixture plugin | `gossip-oracle-fixture` 0.1.0 | SMOKE PASS (oracle only) | Activation, hook header and content marker passed; this is not third-party plugin compatibility |
| WordPress plugins/themes | Exact source/version | UNSUPPORTED / UNTESTED | No third-party report |
| GoHP translated PHP | draft PR #2 `feat/phase0-transpiler@56234c7` | UNTESTED | Owned-fixture CI only; `main` empty; `generated` is not compatibility evidence |
| GoHP builder adapter | `gossip.builder/v0.1` proposal | UNTESTED | No Gossip adapter implementation/evidence; tag/license/stable API undecided upstream |
| Composer/extensions/autoload | Per package/runtime | UNSUPPORTED | No adapter/isolation validation; GoHP Composer support is unproven |
| WordPress through GoHP | Exact WP source/version | UNSUPPORTED | No GoHP WordPress or database support demonstrated |
| CGI / Sakura FreeBSD | Actual plan TBD | UNTESTED | No host validation/resource measurement |

## Status vocabulary

| Status | Meaning |
| --- | --- |
| UNVERIFIED | Source/design exists but no recorded command/result for this revision/configuration. |
| UNTESTED | No relevant implementation or evidence exists. |
| GENERATED | Translator emitted output; no compile or behavior conclusion. |
| COMPILED | Candidate compiled; no behavior conclusion. |
| SMOKE PASS | Named fixture passed a narrow documented path. |
| DIFFERENTIAL PASS | Pinned-oracle comparison passed under recorded configuration. |
| UNSUPPORTED | Explicitly outside current implementation/tested scope. |
| BLOCKED | Evidence awaits a stated dependency or decision. |

Advancement requires a report with provenance/hash, tool/platform versions,
fixture, command/result, and known deviations. Failed/unknown outcomes remain
visible.

## Proposed `gossip.builder/v0.1` candidate result vocabulary

| Result | Meaning | May activate a slot? |
| --- | --- | --- |
| `rejected` | Intake/policy/schema/provenance validation declined the candidate | No |
| `generated` | Translator generated output with recorded hashes | No |
| `build_failed` | Candidate build failed | No |
| `differential_failed` | Oracle/smoke comparison failed | No |
| `infrastructure_failed` | Isolated builder, oracle, or runner did not produce reliable evidence | No |
| future passing result | Explicit, stored evidence meets the matrix rule | Only after Gossip policy/migration/deploy gates |

Every result records source hash, input/output artifact hashes, toolchain/version,
platform, report schema, and test context. Slot activation, database migration,
and deployment remain Gossip-owned actions.
