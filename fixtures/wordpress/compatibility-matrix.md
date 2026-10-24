# WordPress oracle compatibility matrix (Phase 0)

Reference: the image tags in `.env.example`, fixed fixture content in this
directory, and a disposable Compose volume. This is an oracle/smoke fixture,
not a claim of WordPress compatibility or a permission to run arbitrary PHP.

| Surface | Oracle evidence | Phase 0 status | Candidate comparison |
| --- | --- | --- | --- |
| Core hook invocation | Checked-in plugin adds `X-Gossip-Oracle-Hook` from `send_headers` and appends a content marker via `the_content` | testable against WordPress | unsupported until Gossip exposes a versioned hook trace contract; this fixture does not prove full hook ordering |
| Gutenberg stored blocks | `seed/gutenberg-post.html` is stored in `wp_posts.post_content` and read through REST/rendered page | testable against WordPress storage/render output | unsupported: raw block comments alone do not constitute a block runtime |
| `theme.json` / block theme | checked-in v3 `theme.json`, template and template part; HTML exposes the color preset variable and theme title | testable against WordPress | unsupported: no Gossip theme.json renderer exists yet |
| `wp_posts` and terms | seeded `post` plus `gossip-oracle` category are queried through REST | testable against WordPress | unsupported: Gossip DB/import mapping has not been specified |
| Checked-in plugin activation | fixture plugin is activated by `wp-cli`; marker/header prove it executes | testable against WordPress | unsupported: plugin SDK adapter and isolated builder are not implemented |
| Arbitrary third-party plugin ZIP / Composer | deliberately absent | unsupported by design | requires inventory, license/dependency review, isolated build, and smoke tests |
| `wp_template` database overrides / dynamic blocks | deliberately absent | unsupported | requires a dedicated fixture and renderer contract |

`tests/compat/wordpress-oracle-smoke.sh` checks the rows marked testable. Set
`CANDIDATE_BASE_URL` and `REQUIRE_CANDIDATE=1` only after a versioned Gossip
compatibility HTTP contract exists; an HTTP lookalike is not a parity claim.
