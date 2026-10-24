# Owned WordPress post import — local fixture only

`gossip-import` previews a narrow WP-CLI JSON export and can create a **new**
Gossip snapshot after explicit confirmation. It never merges into or replaces
an existing snapshot. This is not a database migration, WXR importer, WordPress
renderer or production import tool. Only disposable owned fixtures are in scope.

## Local workflow

On the disposable WordPress oracle, use `wp post list` with the exact selected
fields below. Its JSON/field selection interface is documented by
[WordPress](https://developer.wordpress.org/cli/commands/post/list/).

```sh
wp post list --post_type=post --post_status=publish \
  --name=gossip-oracle-gutenberg \
  --fields=ID,post_type,post_status,post_password,post_name,post_title,post_content,post_modified_gmt \
  --format=json
```

Save the JSON output as `owned-posts.json` in a private local directory. From
the Gossip repository root, inspect it without writing a snapshot:

```sh
go run ./cmd/gossip-import -input /path/to/owned-posts.json
```

Review the listed source IDs, slugs, titles and warnings, as well as the input
content. Copy `source_sha256` from this preview when deliberately creating a
new snapshot:

```sh
go run ./cmd/gossip-import -input /path/to/owned-posts.json \
  -commit -expect-sha256 REVIEWED_SOURCE_SHA256 \
  -out .gossip/wordpress-fixture.json
go run ./cmd/gossip -data-file .gossip/wordpress-fixture.json
```

Visit `http://127.0.0.1:8080/posts/gossip-oracle-gutenberg`. The page shows
**escaped raw block/HTML source**, not WordPress-rendered blocks. The CLI does
not start a server or make imported data public by itself. Do not use private
or production content in this development workflow.

## Contract and exclusions

- Input is a non-empty UTF-8 JSON array, at most 16 MiB, with the eight selected
  fields above. Missing/null required fields, unknown fields, trailing data,
  duplicate IDs/slugs and invalid dates fail the whole batch, not just one row.
- Only `post_type=post`, `post_status=publish`, and an empty `post_password`
  are accepted. Draft/private/password-protected posts cannot silently become
  public by losing their visibility metadata.
- Slugs must match the current Gossip ASCII lowercase/hyphen contract. Encoded
  or Unicode slugs, pages, custom post types and empty titles are rejected;
  there is no lossy slug/title conversion.
- `post_name`, raw `post_title`, raw `post_content` and `post_modified_gmt`
  map to slug, title, content and UTC update time. IDs are preview provenance
  only. Authors, terms, attachments/media, metadata, options, permalinks and
  original WordPress rendering behavior are not imported.
- Preview returns `prepared` without creating directories, locks or snapshots.
  Commit requires `-commit`, `-out` and the exact input-byte SHA-256 from the
  preview. SHA-256 confirms reviewed bytes, not source authenticity.
- Save returns `saved`, not published or activated. The whole snapshot is
  serialized, synced and atomically linked into a new destination path on the
  same filesystem. Existing files, corrupt files, directories and symlinks
  cannot be replaced. Local hard-link support is required; no unsafe fallback
  exists. The regular editor continues to use atomic replacement for saves.
- Failed commits can leave an empty lock file or newly created parent directory,
  but do not partially publish a snapshot. A repeated commit fails if its
  destination exists. Lost stdout/process interruption after creation is an
  unknown result: inspect that snapshot before retrying; never delete it merely
  to make a retry succeed. No automated rollback/deletion is provided.
- The [local store limitations](local-content-store.md) still apply, including
  private local directories, cooperating single-writer locks, no network
  filesystem or power-loss durability guarantee, and no Windows support.

### Capability / authentication classification

Stable operation IDs `gossip.local.wordpress-posts.prepare` and
`gossip.local.wordpress-posts.save` are explicitly **excluded** from WebMCP.
They are offline developer CLI operations, authorized by the local OS user,
not a browser route, MCP protocol, internal service or production migration API.
There is no WebMCP registration, automatic submit, UI callsite, production
permission/audit ledger or supported agent API. No cookie/token forwarding is
introduced. Existing post display remains anonymous; edit remains dev-keyed.
The explicit exclusion is intentional until a production owner-service contract
and its authorization/audit tests exist.

## Evidence — 2026-10-24 JST

Implementation and tests: `6c78de7058f2762ce5a3e975769881440fa50c42`.
The exact source files tested were then committed without source changes.
Host: macOS arm64 / Go 1.22.1; Docker Engine 25.0.3. No new browser or platform
runtime claim is made for this import change.

The dedicated Compose project was `gossip-postimport`, with the WordPress host
port bound only to `127.0.0.1:18089`. The commands below used these immutable
image overrides (same images as the earlier oracle evidence):

```sh
export WORDPRESS_ORACLE_IMAGE=wordpress@sha256:04be81de853887fca707742785b32bf93fb4a692246dd1bc90b3790105071a9a
export WORDPRESS_CLI_IMAGE=wordpress@sha256:fd01627919b2bbcff80924b1527813c193fde04658348c26874d1b890bfb8e61
export WORDPRESS_MYSQL_IMAGE=mysql@sha256:106d5197fd8e4892980469ad42eb20f7a336bd81509aae4ee175d852f5cc4565
export WORDPRESS_ORACLE_PORT=127.0.0.1:18089
docker compose --project-name gossip-postimport --env-file .env.example \
  up -d --wait wordpress-mysql wordpress-oracle
docker compose --project-name gossip-postimport --env-file .env.example \
  --profile oracle run --rm --no-deps wordpress-bootstrap
```

After creating a private temporary directory with
`mktemp -d /private/tmp/gossip-postimport.XXXXXX`, export was performed with:

```sh
docker compose --project-name gossip-postimport --env-file .env.example \
  --profile oracle run --rm --no-deps --entrypoint wp wordpress-bootstrap \
  post list --post_type=post --post_status=publish --name=gossip-oracle-gutenberg \
  --fields=ID,post_type,post_status,post_password,post_name,post_title,post_content,post_modified_gmt \
  --format=json --allow-root > /private/tmp/gossip-postimport.4L9Exn/posts.json
GOSSIP_WP_POST_EXPORT=/private/tmp/gossip-postimport.4L9Exn/posts.json \
  go test -race -v ./internal/cms ./internal/wpimport ./cmd/gossip-import
```

Use your own newly created temporary path when reproducing. Export timestamp
and source hash will differ on a fresh oracle. The recorded export contained
owned post ID 5, slug `gossip-oracle-gutenberg`, UTC modification time
`2026-10-24 14:52:14`, and SHA-256
`5c5a5ffbb5910cacdbc20993ea3e1c801f6fe0f76dc6b080a9852fc53c434375`.
Its raw content matched `fixtures/wordpress/seed/gutenberg-post.html` exactly
(SHA-256 `f8ca67ec0073b92c557e62c0819e4633c3932e41a7d8be3c8da173a0aeb2d408`).

Result: **PASS** for all three packages, including the opt-in real-export
round trip, exact raw field preservation after reopen, anonymous HTTP handler
escaped display, existing CMS tests, unsafe-visibility rejection, stale-preview
rejection and no-overwrite checks. The HTTP check uses `httptest`, not a browser.
Without `GOSSIP_WP_POST_EXPORT`, that real-export test explicitly skips.

Cleanup command for this newly created project:

```sh
docker compose --project-name gossip-postimport --env-file .env.example \
  down --volumes --remove-orphans
```

The disposable containers, network and both project-owned WordPress/MySQL
volumes were removed. They can be recreated from the checked-in fixture; no
production data or existing project was modified. Temporary snapshots created
by Go tests were cleaned automatically. The small owned JSON export was retained
in its temporary directory as local evidence. Production DB migration,
WordPress HTTP/rendering parity, categories/media/options and production
authorization remain unimplemented/unverified.
