# Current state — Phase 0–1

Last updated: 2026-10-24 (JST)

## Present source state

- A Go module, native Go Plugin SDK and sample plugin, build-information package,
  Go HTTP app with fixture/edit-view paths, release-slot manager and unit-test
  sources, Compose development kit, and WordPress oracle fixtures are present.
- This document does **not** record a completed command result for this exact
  source/configuration. Until parent integration records a command and outcome,
  they are **unverified implementation**, not tested compatibility.

## Documentation delivered in this change set

- Authoritative English project overview with a Japanese summary.
- Proposal-level architecture, security, licensing, Docker, CGI/Sakura, and
  roadmap documents.
- An evidence-led compatibility matrix that distinguishes untested work from a
  passing result.

Documentation is not evidence that a feature is safe for production.

## Unimplemented or unverified

- Functional verification of the Go HTTP app, content fixture, edit/view path,
  Plugin SDK, and release-slot manager for this exact revision/configuration.
- Execution of the Compose kit against its reviewed pinned WordPress/MySQL/CLI/
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
