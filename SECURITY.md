# Security policy

## Supported versions

No released or supported Gossip version exists yet. Phase 0–1 artifacts are
experimental and must not be used with production secrets or production WordPress
databases.

## Reporting a vulnerability

Please do not file a public issue for a suspected vulnerability. Contact the
repository owner privately through GitHub's security-advisory/reporting channel
when available, or use the contact method listed on the owner's GitHub profile.
Include affected revision, a minimal reproduction, impact, and any suggested
mitigation. Do not include credentials, private keys, or proprietary plugin
source.

We will acknowledge reports when a private channel is available, assess scope,
and coordinate disclosure before publishing details. No response-time SLA is
promised during this pre-release phase.

## Build and plugin safety

Plugin/theme archives, PHP, Composer metadata/scripts, generated code, and
unfamiliar build scripts are untrusted. Proposed isolation and release controls
are documented in [Security — PROPOSAL](docs/security-proposal.md); they are not
yet verified implementation claims.
