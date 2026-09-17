#!/usr/bin/env bash
set -euo pipefail

ROOT="/opt/ftn-github.server"
APP="$ROOT/backend/git-center"
BIN="$APP/ftn-git-center"
UNIT="$ROOT/systemd/ftn-git-center.service"

if [[ ! -d "$APP" ]]; then
  echo "missing Git Center source: $APP" >&2
  exit 1
fi

command -v go >/dev/null 2>&1 || { echo "Go is required" >&2; exit 1; }
command -v systemctl >/dev/null 2>&1 || { echo "systemd is required" >&2; exit 1; }

mkdir -p "$APP"
cd "$APP"
go build -o "$BIN" .
chmod 0755 "$BIN"

install -m 0644 "$UNIT" /etc/systemd/system/ftn-git-center.service
systemctl daemon-reload
systemctl enable ftn-git-center.service
systemctl restart ftn-git-center.service

for _ in {1..20}; do
  if curl -fsS http://127.0.0.1:8096/healthz >/dev/null 2>&1; then
    echo "FTN Git Center: healthy on 127.0.0.1:8096"
    exit 0
  fi
  sleep 1
done

systemctl --no-pager --full status ftn-git-center.service || true
exit 1
