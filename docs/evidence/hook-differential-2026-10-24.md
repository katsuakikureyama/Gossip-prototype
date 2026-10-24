# Native hook differential — 2026-10-24 JST

Result: **DIFFERENTIAL PASS for the five owned string-filter cases only**.
Implementation target: `df11d5e680a849a526e14f35383447c5fc13d8bf`.
The run used the identical files immediately before that commit; subsequent
documentation changes do not change the tested source/configuration.

## Contract and reproduction

`tests/compat/hooks-v1.json` is the shared `gossip.hook-fixture/v1` input:
case ID, initial string, and ordered registrations with priority/prefix/suffix.
`hooks-oracle.php` loads real WordPress `wp-includes/plugin.php` and its hook
engine directly, registering owned callbacks through `add_filter` and invoking
`apply_filters`. The Go adapter registers equivalent transformations through
the native Plugin SDK. The test compares callback index traces and final values,
including a second invocation against the same registration state.

No HTTP route, database, WordPress installation or production credentials are
needed. This test protocol is excluded from public UI/AI/WebMCP capabilities:
it is only a local test harness, not an end-user operation. It does not add a
production WordPress adapter.

Prerequisite: Docker and the already pulled official WordPress digest below.
The command will fail rather than pull a missing image; without the environment
variable, ordinary `go test` explicitly SKIPs this integration test.

```sh
GOSSIP_WP_ORACLE_IMAGE=wordpress@sha256:04be81de853887fca707742785b32bf93fb4a692246dd1bc90b3790105071a9a \
  go test -count=1 -v ./tests/compat
```

The oracle runs as UID 65532, read-only, without network or Linux capabilities,
with a read-only fixture mount, 128 MiB memory/32 PID/1 CPU limits, no-new-privileges,
and automatic container removal. The test has a 45-second process timeout.

## Recorded evidence

- Command above: PASS, one execution for this configuration.
- WordPress: `6.6.2`; actual PHP: `8.3.13`; host Go: `1.22.1`.
- Fixture SHA-256: `b77b22bb03eb7639387e8ad0e70701db707affad19ba0ff43b647ecefa44ebed`.
- Oracle image digest: `wordpress@sha256:04be81de853887fca707742785b32bf93fb4a692246dd1bc90b3790105071a9a`.
- The container exited successfully and was removed by `docker run --rm`; no volumes were created.

| Case | Observed callback indexes | Identical result on both runtimes (both invocations) |
| --- | --- | --- |
| empty | `[]` | `seed` |
| priority | `[1,2,0]` | `C(B(A(seed)))` |
| equal-priority | `[0,1,2]` | `third(second(first(seed)))` |
| boundaries-and-ties | `[1,2,0]` | `last(zero-second(zero-first()))` |
| unicode | `[1,0]` | `「前こんにちは & <content>後」` |

Not covered: dynamic add/remove during dispatch, recursion/reentrancy,
WordPress callback identity/duplicate registrations, multiple arguments,
references, exceptions, negative priorities or priorities above 999 (currently
rejected by the SDK), `the_content` core filter stack, theme rendering, PHP
translation, database import, or third-party plugins. These remain unsupported
or unverified; this report is not general WordPress compatibility.
