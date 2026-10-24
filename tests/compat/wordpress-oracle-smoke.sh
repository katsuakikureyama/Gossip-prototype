#!/bin/sh
set -eu

: "${ORACLE_BASE_URL:?ORACLE_BASE_URL is required}"

fail() {
  printf '%s\n' "compatibility smoke: FAIL: $*" >&2
  exit 1
}

contains() {
  needle=$1
  haystack=$2
  printf '%s' "$haystack" | grep -F -- "$needle" >/dev/null || fail "missing $needle"
}

headers=$(curl --fail --silent --show-error --location --head "$ORACLE_BASE_URL/?p=1")
contains 'X-Gossip-Oracle-Hook: send-headers-v1' "$headers"

post=$(curl --fail --silent --show-error --location "$ORACLE_BASE_URL/wp-json/wp/v2/posts?slug=gossip-oracle-gutenberg")
contains 'Gossip Oracle Gutenberg Fixture' "$post"
contains 'Gutenberg block markup remains source data.' "$post"

category=$(curl --fail --silent --show-error --location "$ORACLE_BASE_URL/wp-json/wp/v2/categories?slug=gossip-oracle")
contains 'Gossip Oracle' "$category"

page=$(curl --fail --silent --show-error --location "$ORACLE_BASE_URL/gossip-oracle-gutenberg/")
contains 'Gossip Oracle Theme' "$page"
contains 'plugin-content-filter-v1' "$page"
contains '--wp--preset--color--oracle-accent' "$page"

if [ -n "${CANDIDATE_BASE_URL:-}" ]; then
  candidate=$(curl --fail --silent --show-error --location "$CANDIDATE_BASE_URL/gossip-oracle-gutenberg/") || fail 'candidate request failed'
  contains 'Gossip Oracle Gutenberg Fixture' "$candidate"
  contains 'plugin-content-filter-v1' "$candidate"
  printf '%s\n' 'compatibility smoke: oracle + candidate response markers PASS'
elif [ "${REQUIRE_CANDIDATE:-0}" = '1' ]; then
  fail 'CANDIDATE_BASE_URL is required when REQUIRE_CANDIDATE=1'
else
  printf '%s\n' 'compatibility smoke: WordPress oracle PASS; Gossip comparison UNVERIFIED (no compatibility HTTP contract)'
fi
