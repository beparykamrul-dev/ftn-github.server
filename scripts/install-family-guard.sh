#!/usr/bin/env bash
set -euo pipefail

ROOT="${FTN_FAMILY_GUARD_ROOT:-/opt/ftn-family-guard}"
REPO="${FTN_FAMILY_GUARD_REPO:-https://github.com/beparykamrul-dev/ftn-github.server.git}"
BRANCH="${FTN_FAMILY_GUARD_BRANCH:-main}"
BIN="$ROOT/bin/ftn-family-guard"
ENV_DIR="/etc/ftn"
ENV_FILE="$ENV_DIR/family-guard.env"
SERVICE="ftn-family-guard"

command -v git >/dev/null || { echo "git is required"; exit 1; }
command -v go >/dev/null || { echo "Go 1.23+ is required"; exit 1; }

mkdir -p "$ENV_DIR"
if [ -d "$ROOT/.git" ]; then
  git -C "$ROOT" fetch origin "$BRANCH"
  git -C "$ROOT" checkout "$BRANCH"
  git -C "$ROOT" pull --ff-only origin "$BRANCH"
elif [ -e "$ROOT" ]; then
  echo "Refusing to overwrite existing non-Git path: $ROOT" >&2
  exit 1
else
  mkdir -p "$(dirname "$ROOT")"
  git clone --branch "$BRANCH" --depth 1 "$REPO" "$ROOT"
fi

mkdir -p "$ROOT/bin"
cd "$ROOT/backend/family-guard"
go mod download
go test ./...
go build -trimpath -ldflags='-s -w' -o "$BIN" .

if [ ! -f "$ENV_FILE" ]; then
  cat > "$ENV_FILE" <<'EOF'
FTN_FAMILY_GUARD_ADDR=:8095
FTN_FAMILY_GUARD_ALLOWED_ORIGIN=
FTN_FAMILY_GUARD_API_TOKEN=
FTN_FAMILY_GUARD_DATABASE_URL=
EOF
  chmod 600 "$ENV_FILE"
fi

if command -v systemctl >/dev/null 2>&1; then
  if [ -f "$ROOT/systemd/ftn-family-guard.service.example" ]; then
    install -m 0644 "$ROOT/systemd/ftn-family-guard.service.example" /etc/systemd/system/ftn-family-guard.service
    systemctl daemon-reload
    systemctl enable --now "$SERVICE"
    systemctl --no-pager --full status "$SERVICE"
  fi
  install -m 0644 "$ROOT/systemd/ftn-git-sync.service" /etc/systemd/system/ftn-git-sync.service
  install -m 0644 "$ROOT/systemd/ftn-git-sync.timer" /etc/systemd/system/ftn-git-sync.timer
  systemctl daemon-reload
  systemctl enable --now ftn-git-sync.timer
fi

echo "Family Guard binary: $BIN"
echo "Git registry: $ROOT/config/github-repositories.generated.yaml"
