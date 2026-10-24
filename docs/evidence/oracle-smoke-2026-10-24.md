# WordPress oracle smoke evidence — 2026-10-24 JST

## Scope and targets

- Branch/head used to build and run: `codex/phase-0-1` at
  `e22b21a90f7c9da08c6b3f40f4ad73c592ce2fd4`.
- Compose configuration: `a0d3bf085332d914d4d886a1880468f62b675add`.
- Project isolation: Compose project `gossip-phase01`, host ports `18080` and
  `18088`, disposable named volumes only.
- Inputs: checked-in Gossip source and owned WordPress fixture only. No
  third-party plugin/theme ZIP, Composer script, production secret, or
  production database was used.

## Image evidence

| Input | Resolved repository digest / local image ID |
| --- | --- |
| `wordpress:6.6.2-php8.3-apache` | `wordpress@sha256:04be81de853887fca707742785b32bf93fb4a692246dd1bc90b3790105071a9a` |
| `mysql:8.4.3` | `mysql@sha256:106d5197fd8e4892980469ad42eb20f7a336bd81509aae4ee175d852f5cc4565` |
| `wordpress:cli-2.10.0-php8.3` | `wordpress@sha256:fd01627919b2bbcff80924b1527813c193fde04658348c26874d1b890bfb8e61` |
| `curlimages/curl:8.11.0` | `curlimages/curl@sha256:83a505ba2ba62f208ed6e410c268b7b9aa48f0f7b403c8108b9773b44199dbba` |
| locally built `gossip-phase01-gossip` | `sha256:1d0cd3dce3a29dcc203fb8c65d28eba2c1581c9b77c1956df095b3b4e71552d5` |

The Go builder base resolved to
`golang:1.22.1-bookworm@sha256:d996c645c9934e770e64f05fc2bc103755197b43fd999b3aa5419142e1ee6d78`.

## Commands and outcomes

All commands used `GOSSIP_PORT=18080`, `WORDPRESS_ORACLE_PORT=18088`,
`--project-name gossip-phase01`, and `--env-file .env.example`.

| Action | Result |
| --- | --- |
| First combined `up -d --build` | CANCELED after a Go builder layer stopped making network progress for more than 100 seconds; no container or volume had been created |
| `docker compose ... build gossip` retry | PASS; the previously interrupted layer downloaded and the Linux Gossip binary image built |
| `docker compose ... up -d wordpress-mysql wordpress-oracle gossip` | PASS; MySQL and WordPress became healthy; Gossip served on port 18080 |
| `docker compose ... --profile oracle run --rm wordpress-bootstrap` | PASS; installed WordPress, activated owned theme/plugin, created category and Gutenberg post |
| `docker compose ... --profile compat run --rm compat-runner` | PASS: `WordPress oracle PASS; Gossip comparison UNVERIFIED` |
| Gossip `GET /healthz`, `/version`, `/api/v1/plugins`, `/posts/welcome` | PASS; health `ok`, build `dev`, sample manifest present, sample-plugin footer rendered |
| `docker compose ... down --volumes --remove-orphans` | PASS; the three containers, project network, and two disposable volumes were removed |

## Proven and not proven

The named WordPress oracle demonstrates the owned fixture's `send_headers` and
`the_content` hooks, Gutenberg post storage/rendering, block-theme/`theme.json`
runtime marker, and category/post REST observations on this pinned environment.
It also demonstrates that the Gossip development container starts and links the
native sample plugin.

It does **not** compare Gossip output with WordPress, prove general hook ordering,
support dynamic/classic themes, import a WordPress database into Gossip, execute
GoHP output, validate arbitrary third-party plugins, or establish production or
CGI readiness. The Gossip compatibility HTTP/trace contract remains undefined,
so the candidate runner intentionally reported Gossip comparison as UNVERIFIED.
