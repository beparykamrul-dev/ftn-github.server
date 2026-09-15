#!/usr/bin/env bash
set -euo pipefail

ROOT="${FTN_ROOT:-/opt/FTN_ser_AI}"
BRANCH="${FTN_BRANCH:-main}"
SERVICE="${FTN_SERVICE:-ftn-ser-ai}"

cd "$ROOT"
git fetch origin "$BRANCH"
git checkout "$BRANCH"
git pull --ff-only origin "$BRANCH"

if [ -f package.json ]; then
  npm install
  if npm run | grep -q ' build'; then npm run build; fi
fi

if command -v systemctl >/dev/null 2>&1 && systemctl list-unit-files | grep -q "^${SERVICE}\.service"; then
  systemctl restart "$SERVICE"
  systemctl --no-pager --full status "$SERVICE"
fi
