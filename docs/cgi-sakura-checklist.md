# Go CGI / Sakura shared-host checklist — UNVERIFIED

This is a validation plan, not confirmation that Sakura supports a configuration.
Confirm the actual plan, region, OS/FreeBSD version, and host documentation
before implementing or advertising support.

## Host contract

- [ ] Confirm native CGI binary allowance, executable location, mode/ownership,
  architecture, dynamic-library restrictions, and size limits.
- [ ] Confirm CGI environment, body limits, stdin/stdout, header requirements,
  timeouts, process/concurrency limits, logging, errors, and signals.
- [ ] Confirm rewrite/front-controller support, `PATH_INFO`, query strings,
  HTTPS/proxy headers, and static-asset rules.
- [ ] Measure cold-start/per-request costs and resource limits with a harmless
  fixture; retain commands and observations.

## Runtime and plugin safety

- [ ] Build reproducible target binaries off-host; record Go version,
  GOOS/GOARCH, hashes, and provenance.
- [ ] Validate request/body/header/redirect/cookie/host/escaping/error behavior.
- [ ] Validate host-only sessions, CSRF, logout, and no cookie forwarding to
  internal services.
- [ ] Validate SQLite/MySQL using test credentials and a disposable DB.
- [ ] Validate assets, cache headers, uploads, writable paths, and permissions.
- [ ] Never build, translate, Composer-install, or execute arbitrary plugins on
  the shared host; upload only approved immutable artifacts.
- [ ] Test selected-slot serving and binary rollback; do not claim database
  rollback without a separately tested migration plan.
- [ ] Confirm logs/audit omit secrets and proprietary plugin source.

Do not label CGI/Sakura `PASS` until relevant boxes have evidence for the actual
plan. Missing capability, unsafe permissions, unmeasured limits, or untested
session/DB behavior remains visible in the compatibility matrix.
