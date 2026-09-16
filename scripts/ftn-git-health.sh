#!/usr/bin/env bash
set -euo pipefail

ROOT="${FTN_GITHUB_ROOT:-/opt/ftn-github.server}"
REGISTRY="$ROOT/config/git-service-registry.yaml"

fail=0
check(){
  local name="$1" url="$2"
  if curl -fsS --max-time 5 "$url" >/dev/null 2>&1; then
    printf 'OK   %s %s\n' "$name" "$url"
  else
    printf 'FAIL %s %s\n' "$name" "$url"
    fail=1
  fi
}

[ -f "$REGISTRY" ] || { echo "FAIL missing $REGISTRY"; exit 1; }

if command -v curl >/dev/null 2>&1; then
  check control-plane "http://127.0.0.1:8080/healthz"
else
  echo "WARN curl not installed; registry syntax check only"
fi

if command -v git >/dev/null 2>&1; then
  git -C "$ROOT" rev-parse --is-inside-work-tree >/dev/null
  printf 'OK   git %s\n' "$(git -C "$ROOT" rev-parse --short HEAD)"
fi

exit "$fail"
