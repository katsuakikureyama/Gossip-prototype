#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

required='compose.yaml
.env.example
fixtures/wordpress/bootstrap.sh
fixtures/wordpress/compatibility-matrix.md
fixtures/wordpress/plugin/gossip-oracle-fixture.php
fixtures/wordpress/seed/gutenberg-post.html
fixtures/wordpress/theme/gossip-oracle/theme.json
fixtures/wordpress/theme/gossip-oracle/templates/index.html
tests/compat/wordpress-oracle-smoke.sh'

for path in $required; do
  [ -f "$path" ] || { printf '%s\n' "missing required oracle fixture: $path" >&2; exit 1; }
done

for image_var in WORDPRESS_ORACLE_IMAGE WORDPRESS_CLI_IMAGE WORDPRESS_MYSQL_IMAGE COMPAT_RUNNER_IMAGE; do
  value=$(grep "^$image_var=" .env.example | cut -d= -f2-)
  [ -n "$value" ] || { printf '%s\n' "missing image pin: $image_var" >&2; exit 1; }
  case "$value" in
    *:latest|latest) printf '%s\n' "floating image tag forbidden: $image_var=$value" >&2; exit 1 ;;
  esac
done

grep -F 'wordpress:6.6.2-php8.3-apache' .env.example >/dev/null
grep -F 'mysql:8.4.3' .env.example >/dev/null
grep -F 'X-Gossip-Oracle-Hook: send-headers-v1' fixtures/wordpress/plugin/gossip-oracle-fixture.php >/dev/null
grep -F '"version": 3' fixtures/wordpress/theme/gossip-oracle/theme.json >/dev/null
grep -F 'unsupported by design' fixtures/wordpress/compatibility-matrix.md >/dev/null
grep -F 'Gossip comparison UNVERIFIED' tests/compat/wordpress-oracle-smoke.sh >/dev/null

printf '%s\n' 'oracle fixture layout: PASS'
