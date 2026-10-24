# Local content store — contract and evidence

The development fixture can persist posts in a versioned JSON snapshot. This is
not a WordPress database/import format or a production content backend.

## Operation

`go run ./cmd/gossip -dev-editor` uses `.gossip/content.json` relative to the
working directory. Set `GOSSIP_DEV_EDITOR_KEY` first as shown in the README.
`-data-file` selects a different path; an empty argument selects memory only.
The Docker default remains memory-only because its non-root scratch filesystem
has no writable data directory. Persistent Docker operation requires a mounted
directory writable by UID 65532 and an explicit snapshot path; this mounted
configuration has not been runtime-tested.

- A missing snapshot is seeded once. Existing snapshots are loaded without
  overwriting their content with seed values.
- Schema version 1 contains a `posts` array. Invalid JSON, unknown fields,
  unsupported schema, duplicate slugs, invalid slugs and blank titles fail
  startup without replacing the snapshot. The snapshot limit is 16 MiB.
- Save serializes writes, writes a same-directory temporary file, syncs and
  closes that file, then atomically renames it over the snapshot. Only then does
  the in-memory view change and the existing HTTP save handler return success.
  Failures before rename retain the old snapshot and in-memory content.
- A nonblocking OS file lock rejects a second cooperating process using the
  same path. Process exit, including SIGKILL, releases the lock. The `.lock`
  file intentionally remains; never remove or replace it while a server runs.
- Newly created directories use mode 0700 and snapshot/lock files use 0600.
  Use a private local directory; do not allow untrusted users to replace paths
  or files. Network filesystems, path aliases and external snapshot writers are
  outside the contract.
- The parent directory is not fsynced. Sudden power loss/filesystem failure
  durability is not promised. A kill during save may leave a temporary file;
  no in-flight interrupted-write or recovery claim is made. Back up the
  snapshot separately. Windows is not currently a supported build target.

## Capability and authorization classification

`gossip.dev.post.save`, `gossip.dev.post.edit.inspect`, and
`gossip.public.post.read` are explicitly **excluded** from WebMCP publication
in this experimental fixture. There is no capability discovery/agent API,
automatic submit, production permission model or audit ledger. The storage
change reuses the existing owner-service HTTP save operation; it adds no
second write route. Native, compat and unsupported browser modes are unverified;
WebMCP is disabled by design here, not marked supported.

Public post reads remain anonymous. Development edit inspection/save remains
an explicit keyed local fixture, not a production RP session. No browser cookie
is forwarded to another service. A 303 means the local save completed; 401/422
reject authentication/validation, and storage failure returns 500. Publishing,
deleting, external sending, payments, production sessions, idempotent retries
and audit-backed status inspection are not implemented. A lost HTTP response
does not prove failure: inspect the post before deciding to repeat an edit.

## Verification — 2026-10-24 JST

Target: `1539a2191f376cd71bfb772ccce32c8f179a1426` (implementation and harness).
Tests ran on the exact files subsequently committed under this SHA; no source
changes occurred between verification and commit.

Environment: macOS arm64, Go 1.22.1, Node 24.4.1, Playwright 1.47.0 and
Chrome 156.0.8078.47. Commands ran from the repository root.

| Command | Result |
| --- | --- |
| `go test -race ./internal/cms` | PASS: store reopen, lock rejection, failed-write disk/memory preservation, corrupt snapshot rejection, concurrent saves and existing CMS tests |
| `PLAYWRIGHT_MODULE=/Users/kureyamayuri/Desktop/mawarimichi/ui-tests/node_modules/playwright GOSSIP_CHROME_PATH='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' node scripts/smoke-editor.cjs` | PASS: real form submit/303/view, Japanese content, escaped markup, plugin footer, reload, missing key and empty-title rejection, unchanged rejected writes; SIGKILL after successful save, restart, view and editor values preserved |
| `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o "$task_build_dir/gossip-linux" ./cmd/gossip` | COMPILED: static Linux arm64 binary; runtime not tested for this revision |
| `GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -o "$task_build_dir/gossip-freebsd" ./cmd/gossip` | COMPILED: static FreeBSD amd64 binary; no Sakura/runtime claim |

For the cross-build commands, `task_build_dir` was created with
`mktemp -d /private/tmp/gossip-storage-build.XXXXXX`. Browser smoke owns its own
temporary binary/snapshot, random development key and server on loopback port
18123; it closes the browser/server and removes its temporary directory.
Reproduction requires an installed Playwright module and browser; override the
two environment paths for your machine. No production data or secrets used.

The earlier [browser report](evidence/browser-editor-2026-10-24.md) records the
historical memory-only implementation. This report supersedes its persistence
limitation for local runs only. WordPress DB mapping/import, power-loss behavior,
multi-host operation and production authorization remain unverified/unimplemented.
