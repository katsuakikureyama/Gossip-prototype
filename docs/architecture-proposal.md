# Architecture — PROPOSAL

## Intent and invariants

Gossip should reuse WordPress semantics and legally reusable assets where
demonstrated, rather than make a visually similar clean-room clone. Compatibility
is per pinned source/version/environment and must be backed by a report.

- Native Go plugins and translated PHP plugins meet the same owner-controlled
  Plugin SDK/Hook API through adapters.
- An artifact marked `generated` or `compiled` is not functionally compatible
  without differential or smoke evidence.
- A candidate is immutable and built away from the active release. The active
  slot remains untouched when candidate inspection, build, test, migration
  planning, or staging fails.
- Plugin inspection never runs arbitrary PHP, Composer scripts, or installer
  hooks. Executable steps occur only in an isolated build/test environment.
- WordPress data is initially imported and tested from a copy/snapshot, never a
  production database.

## Proposed component boundaries

| Component | Responsibility | Trust boundary |
| --- | --- | --- |
| HTTP/CGI runtime | Serve CMS requests using selected release slot | Does not build untrusted code |
| Core services | Content, permission, hooks, assets, sessions, audit | Server-side authorization/validation is authoritative |
| Native Plugin SDK | Stable Go manifest, hook, capability, lifecycle contract | Native artifact still needs inventory/approval |
| WordPress/GoHP adapter | Invoke pinned translator and map supported hooks | No direct PHP execution in production runtime |
| Builder/test runner | Inspect, translate, compile, compare candidate artifacts | Isolated filesystem/network/DB/secrets |
| Compatibility registry | Provenance, versions, test reports, status | Must not expose proprietary source/telemetry |
| Release-slot manager | Immutable Slot A/B candidates and rollback | Activation only after policy gates |
| WordPress oracle | Pinned comparison container/database fixture | Test-only, never production authority |

## Plugin intake and release lifecycle

```text
WP.org search / ZIP upload / native manifest
  -> non-executing inventory (hash, version, license, dependencies)
  -> compatibility-registry lookup
  -> isolated translation/build and differential or smoke test
  -> immutable candidate release (inactive slot)
  -> explicit approval + activation policy
  -> selected active slot; retained prior slot for rollback
```

Database migrations are separate: snapshot, dry run, compatibility/rollback
assessment, then owner-approved activation. Binary rollback does not itself
reverse arbitrary data migration.

## WordPress and GoHP boundary

Gossip may preserve WordPress database shape where useful for import, but schema
similarity is not a behavior promise. Test `wp_posts`, terms, options, block
markup, `theme.json`, templates/parts, `wp_template` overrides, dynamic blocks,
classic themes, `functions.php`, and hook ordering separately.

**PROPOSAL:** Gossip defines `gossip.builder/v0.1` as an isolated-candidate
adapter boundary. It records the GoHP source ref, source hash, generated/source
artifact hashes, toolchain versions, target platform, and test report. It uses a
fresh candidate directory and may invoke a pinned GoHP revision with the
proposed CLI:

```text
gohp -in source.php -out candidate.go -php-version 8.3 -report candidate.report.json
```

`generated` is a provenance result, never a compatibility PASS. The Gossip
adapter must distinguish `rejected`, `generated`, `build_failed`,
`differential_failed`, and `infrastructure_failed`; it may add a passing result
only with the compatibility-matrix evidence requirements. Gossip owns candidate
slot activation, database migration decisions, and deployment, rather than
delegating them to GoHP.

As of 2026-10-24, GoHP `main` is empty; its relevant draft work is PR #2 branch
`feat/phase0-transpiler` at `56234c7`, with CI limited to owned fixtures. GoHP
tags, license, and a stable API are undecided. It does not establish WordPress
core/plugin, Composer/autoload, or database support, nor PHP extensions,
references, closures, or PHP warning/exception equivalence.

## Execution modes and authentication

Target semantics are the same for a Go HTTP server (Docker/VPS/future managed
service) and Go native CGI (shared FreeBSD hosting). CGI candidates are built
off-host and uploaded as binaries; no on-host arbitrary compilation is assumed.

- Browser-facing Gossip routes use Gossip's own host-only relying-party session.
- Service-to-service calls use dedicated internal credentials, never browser
  cookies.
- MCP uses MCP bearer credentials; provider callbacks validate their signature
  and state; public routes remain anonymous.

No full WordPress-clone claim, automatic arbitrary PHP ZIP execution, production
deployment, paid marketplace, billing, public SaaS onboarding, or irreversible
database update is in scope for Phase 0–1.
