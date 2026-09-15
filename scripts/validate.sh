#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

for f in \
  "$ROOT/config/nodes.example.yaml" \
  "$ROOT/config/sources.example.yaml" \
  "$ROOT/config/services.example.yaml"; do
  test -f "$f" || { echo "missing: $f"; exit 1; }
done

if command -v python3 >/dev/null 2>&1; then
  python3 - "$ROOT" <<'PY'
import pathlib, sys
root = pathlib.Path(sys.argv[1])
for p in root.glob("**/*"):
    if p.is_file() and p.name not in {".gitignore"}:
        data = p.read_bytes()
        if b"BEGIN OPENSSH PRIVATE KEY" in data or b"PRIVATE KEY-----" in data:
            raise SystemExit(f"secret-like private key found in {p}")
print("FTN GitHub Server repository validation: OK")
PY
else
  echo "Repository files exist. Install Python 3 for secret-pattern validation."
fi
