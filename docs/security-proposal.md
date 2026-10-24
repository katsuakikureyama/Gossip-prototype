# Security — PROPOSAL

This design proposal is not a statement that its controls are implemented.

## Threat model

Treat plugin/theme ZIPs, PHP source, Composer files/scripts, build scripts,
generated Go code, manifests, uploads, WordPress DB copies, and third-party
fixtures as untrusted. A successful compiler exit code does not make them safe.
Risks include arbitrary build-time execution, supply-chain substitution,
host/tenant escape, secret disclosure, unsafe activation/rollback, destructive
migrations, permission bypass, and proprietary-source leakage.

## Required controls

| Stage | Proposed control |
| --- | --- |
| Intake | Hash/version/provenance/license/dependency inventory without executing PHP or Composer scripts |
| Build | Fresh isolated workspace; pinned toolchain; no production secrets; restricted filesystem/network/database |
| Test | Disposable fixture DB and WordPress oracle; source/build hashes and versions in reports |
| Release | Immutable candidate; explicit status; failed candidate cannot alter active slot |
| Migration | Snapshot, dry run, compatible rollback plan, verified result, and owner approval before switch |
| Runtime | Server-side permission/resource-role checks, audit events, least-privilege credentials, future per-tenant isolation |
| Disclosure | Private reporting; no secrets or proprietary source in fixtures, logs, or telemetry |

## Authentication and unresolved controls

Never transfer browser sessions/cookies to the GoHP builder, WordPress oracle,
MCP client, or provider callback. Apply the route-specific classifications in
the [architecture proposal](architecture-proposal.md#execution-modes-and-authentication).

Still unverified: sandbox/process isolation and egress policy; signature/hash
provenance flow; runtime permissions/audit; slot concurrency/recovery; CGI
environment/header/path handling; and future SaaS tenant/database/filesystem/
network/secret isolation.
