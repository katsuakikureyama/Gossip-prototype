# Docker development kit — PROPOSAL / UNVERIFIED

The Compose kit in this repository is an isolated local comparison environment.
It is not a production deployment recipe and must never mount production secrets
or a production WordPress database.

## Proposed services

| Service | Purpose | Required pinning/isolation |
| --- | --- | --- |
| `gossip` | Go HTTP CMS candidate | Go/toolchain and source revision pinned |
| `wordpress-oracle` | Reference implementation for controlled fixtures | Exact WordPress + PHP + plugin/theme versions pinned |
| `mysql` | Disposable fixture database | Image/version pinned; throwaway volume only |
| `gohp-builder` | Optional translator adapter | Exact GoHP revision/runtime; isolated fresh workspace |
| `runner` | Smoke/differential driver and report collector | No production credentials; explicit fixture input/output |

## Prerequisites and intended workflow

Docker Engine with Compose v2, enough local resources, and reviewed fixture
archives/environment files are required. Do not put credentials in a committed
`.env` file. Once implementation lands, the intended disposable workflow is:

```sh
docker compose build
docker compose up -d mysql wordpress-oracle gossip
docker compose run --rm runner <named-fixture>
docker compose down --volumes
```

Record Compose config/image digests, source revisions, fixture hashes, command,
test result, and report location before declaring PASS. Cleanup applies only to
the disposable development volumes created by this stack.

Required test families include native manifest/build and failed-candidate Slot A
preservation; content edit/view; DB-copy posts/terms/options; hook-order and
side-effect/exception comparisons; Gutenberg blocks, block themes, `theme.json`,
templates, and dynamic blocks; and GoHP compiler/report failure paths.

`compose.yaml` and `.env.example` define the kit and image pins. No successful
Compose command/result is recorded by this document yet.
