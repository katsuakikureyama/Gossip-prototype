# WordPress oracle compatibility matrix (Phase 0)

Reference: the image tags in `.env.example`, fixed fixture content in this
directory, and a disposable Compose volume. This is an oracle/smoke fixture,
not a claim of WordPress compatibility or a permission to run arbitrary PHP.

| Surface | Oracle evidence | Phase 0 status | Candidate comparison |
| --- | --- | --- | --- |
| Core hook invocation | Checked-in plugin adds `X-Gossip-Oracle-Hook` from `send_headers` and appends a content marker via `the_content` | SMOKE PASS on pinned oracle; full ordering untested | unsupported until Gossip exposes a versioned hook trace contract; this fixture does not prove full hook ordering |
| Gutenberg stored blocks | `seed/gutenberg-post.html` is stored in `wp_posts.post_content` and read through REST/rendered page | SMOKE PASS on pinned oracle | unsupported: raw block comments alone do not constitute a Gossip block runtime |
| `theme.json` / block theme | checked-in v3 `theme.json`, template and template part; HTML exposes the color preset variable and theme title | SMOKE PASS on pinned oracle | unsupported: no Gossip theme.json renderer exists yet |
| `wp_posts` and terms | seeded `post` plus `gossip-oracle` category are queried through REST | SMOKE PASS on pinned oracle | unsupported: Gossip DB/import mapping has not been specified |
| Checked-in plugin activation | fixture plugin is activated by `wp-cli`; marker/header prove it executes | SMOKE PASS on pinned oracle | unsupported: translated-plugin adapter and isolated builder are not implemented |
| Arbitrary third-party plugin ZIP / Composer | deliberately absent | unsupported by design | requires inventory, license/dependency review, isolated build, and smoke tests |
| `wp_template` database overrides / dynamic blocks | deliberately absent | unsupported | requires a dedicated fixture and renderer contract |

`tests/compat/wordpress-oracle-smoke.sh` passed the named oracle rows on
2026-10-24 JST at head `e22b21a`; exact image digests and cleanup are in
`docs/evidence/oracle-smoke-2026-10-24.md`. Set
`CANDIDATE_BASE_URL` and `REQUIRE_CANDIDATE=1` only after a versioned Gossip
compatibility HTTP contract exists; an HTTP lookalike is not a parity claim.
