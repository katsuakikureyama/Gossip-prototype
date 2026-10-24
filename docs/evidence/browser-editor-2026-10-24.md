# Local editor browser smoke — 2026-10-24 JST

Result: PASS. Runtime source: `19aa4217bbab97878e275628919091505b386c56`.
Harness: `6d09fe51e7e4728b877b605b19d4593be2c16998` (identical files tested
immediately before committing). Node 24.4.1, Playwright 1.47.0, installed Chrome
156.0.8078.47, Go 1.22.1 on macOS arm64. One execution; no application changes.

Run from the repository root, using existing local Playwright and Chrome:

```sh
PLAYWRIGHT_MODULE=/absolute/path/to/node_modules/playwright \
GOSSIP_CHROME_PATH='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' \
  node scripts/smoke-editor.cjs
```

This run reused the installed Playwright package from the neighboring
`mawarimichi/ui-tests/node_modules/playwright` directory without modifying that
repository or adding a Gossip runtime dependency. No packages were downloaded.
With Playwright available through normal Node resolution, omit PLAYWRIGHT_MODULE.
With Playwright-managed Chromium already installed, omit GOSSIP_CHROME_PATH.

The harness builds a temporary Gossip binary, starts it on `127.0.0.1:18123`
with the development editor and a random per-run key, and uses the actual HTML
form and Save button. It verifies:

- native form POST returns 303 and redirects to the post;
- Japanese title/body match after saving and reloading;
- HTML/script input is rendered as text, not executed/inserted as elements;
- native sample plugin footer appears;
- missing-key GET/POST requests are rejected with 401;
- empty title is rejected by browser required-field validation and server 422;
- rejected writes leave the previously saved content unchanged.

The browser, child server and temporary binary directory were closed/removed
by the harness. No production or other user's content was used. The public
endpoint remained a local fixture. This is a test harness only and does not
add a UI, AI Assist or WebMCP capability.

The store is in memory: refresh works within the running process, but durable
storage/restart persistence is not implemented. Production login, authorization,
concurrent-edit handling and WordPress rendering/import remain outside this test.
