# Gossip

Gossip is an experimental Go-powered, WordPress-compatible CMS from Tulipa
Design. It aims to demonstrate compatibility with **pinned** WordPress releases
and selected plugins/themes while supporting native Go plugins from day one.
WordPress is the reference implementation; this project does not claim complete
WordPress parity.

> 日本語概要: Gossip は WordPress との互換性を検証可能な範囲で実証する Go
> ベースの CMS です。PHP/WordPress プラグインを無条件に実行するものではなく、
> 固定版・隔離ビルド・比較テストを通過したものだけを候補にします。現時点では
> 実験段階であり、本番利用や完全互換を約束しません。

## Status

Phase 0–1 is a proposal and starter-scaffold phase. The authoritative account of
tested and untested scope is the [compatibility matrix](docs/compatibility-matrix.md).
A successful compilation is a build result, not a functional-compatibility result.

## Project boundaries

- **Gossip** owns CMS semantics, the native Go Plugin SDK, compatibility reports,
  build release slots, activation, and rollback.
- **GoHP** is a separate general PHP-to-Go transpiler. Its compiler core must not
  gain WordPress-specific behavior. Gossip may consume its future pinned builder
  interface through an adapter.
- Reuse of WordPress HTML/CSS/JS/media, Gutenberg formats, or `theme.json`
  requires provenance and license review. Raw files do not prove a theme renders
  correctly without a block/theme runtime.
- This project has no marketplace, billing service, public SaaS deployment, or
  production migration in scope.

Architecture, security, and licensing positions are deliberately marked
**PROPOSAL** until reviewed; see [the documentation index](docs/INDEX.md).

### GoHP integration status

GoHP's `main` branch is currently empty. Its only relevant work is draft PR #2,
`feat/phase0-transpiler` at `56234c7`, with CI limited to owned fixtures. Its
tag, license, and stable API remain undecided. Gossip therefore proposes an
isolated `gossip.builder/v0.1` candidate adapter, not a GoHP compatibility
guarantee. GoHP has no demonstrated WordPress, Composer, or database support;
the proposed adapter records source/artifact/toolchain hashes and reports a
result such as `rejected`, `generated`, `build_failed`, `differential_failed`,
or `infrastructure_failed`. Gossip—not GoHP—owns slot activation, database
migration, and deployment.

## Development

The starter Go packages can be exercised locally with:

```sh
go test ./...
```

The repository includes a local Docker Compose WordPress-oracle kit. Its owned
fixture passed a pinned WordPress 6.6.2 oracle smoke on 2026-10-24; Gossip-to-
WordPress differential compatibility remains unimplemented and unverified. The
service boundaries and evidence requirements are in the
[Docker development kit](docs/docker-development-kit.md).

Do not use production secrets or a production WordPress database. Treat plugin
ZIPs, Composer metadata/scripts, PHP source, and unfamiliar build scripts as
untrusted. Metadata inspection must not execute PHP or Composer scripts.

Report vulnerabilities privately as described in [SECURITY.md](SECURITY.md).
Licensing remains provisional pending dependency/provenance audit; see
[third-party notices](LICENSES/THIRD_PARTY_NOTICES.md).

## Roadmap

1. Establish a minimal Go HTTP fixture with edit/view smoke paths and a native
   sample plugin.
2. Prove native plugin candidate builds, Slot B selection, and failure rollback
   while Slot A stays active.
3. Add a pinned WordPress oracle and publish differential results for hooks,
   blocks, `theme.json`, database fixtures, and selected plugins.
4. Define the isolated GoHP adapter protocol and first owned fixture only after
   the upstream translator interface is pinned.
5. Validate CGI deployment against Sakura's actual host constraints before
   claiming shared-host support.

See [the roadmap proposal](docs/roadmap-proposal.md) for acceptance evidence and
explicit non-goals.
