#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
demo_root=$(mktemp -d "${TMPDIR:-/tmp}/gossip-ab.XXXXXX")
trap 'rm -rf "$demo_root"' EXIT INT TERM
cd "$repo_root"

go build -o "$demo_root/gossipctl" ./cmd/gossipctl

"$demo_root/gossipctl" build \
  --root "$demo_root/releases" \
  --source "$repo_root" \
  --slot A \
  --build-id demo-a \
  --package ./cmd/gossip >/dev/null

state_before=$("$demo_root/gossipctl" status --root "$demo_root/releases")
if "$demo_root/gossipctl" build \
  --root "$demo_root/releases" \
  --source "$repo_root" \
  --slot B \
  --build-id expected-failure \
  --package ./cmd/not-a-package >/dev/null 2>&1; then
  echo "expected candidate build to fail" >&2
  exit 1
fi
state_after=$("$demo_root/gossipctl" status --root "$demo_root/releases")
if [ "$state_before" != "$state_after" ]; then
  echo "failed build changed release state" >&2
  exit 1
fi

"$demo_root/gossipctl" build \
  --root "$demo_root/releases" \
  --source "$repo_root" \
  --slot B \
  --build-id demo-b \
  --package ./cmd/gossip >/dev/null
"$demo_root/gossipctl" activate --root "$demo_root/releases" --slot B

status=$("$demo_root/gossipctl" status --root "$demo_root/releases")
printf '%s\n' "$status" | grep '"active_slot": "B"' >/dev/null
printf '%s\n' "$status" | grep '"build_id": "demo-a"' >/dev/null
printf '%s\n' "$status" | grep '"build_id": "demo-b"' >/dev/null

echo "A/B demo PASS: failed build preserved A; verified B became active"
