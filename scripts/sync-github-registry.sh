#!/usr/bin/env bash
set -euo pipefail

ROOT="${FTN_GITHUB_ROOT:-/opt/ftn-github.server}"
OUT="$ROOT/config/github-repositories.generated.yaml"
API="https://api.github.com/user/repos?affiliation=owner&per_page=100&page="

command -v curl >/dev/null || { echo "curl is required"; exit 1; }
command -v python3 >/dev/null || { echo "python3 is required"; exit 1; }
: "${GITHUB_TOKEN:?GITHUB_TOKEN must be supplied at runtime}"

mkdir -p "$(dirname "$OUT")"
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

page=1
printf 'version: 1\nsource: github\nowner: beparykamrul-dev\nrepositories:\n' > "$OUT"
while :; do
  body="$(curl -fsSL \
    -H 'Accept: application/vnd.github+json' \
    -H "Authorization: Bearer $GITHUB_TOKEN" \
    -H 'X-GitHub-Api-Version: 2022-11-28' \
    "${API}${page}")"
  printf '%s' "$body" > "$TMP"
  count="$(python3 - "$TMP" <<'PY'
import json, sys
p=json.load(open(sys.argv[1]))
for r in p:
    if r.get('owner', {}).get('login') != 'beparykamrul-dev':
        continue
    print(f"  - name: {r['name']!r}")
    print(f"    full_name: {r['full_name']!r}")
    print(f"    default_branch: {r.get('default_branch','main')!r}")
    print(f"    private: {str(bool(r.get('private'))).lower()}")
    print(f"    archived: {str(bool(r.get('archived'))).lower()}")
    print(f"    clone_url: {r['clone_url']!r}")
print(f"__COUNT__={len(p)}")
PY
)"
  n="$(printf '%s\n' "$count" | sed -n 's/^__COUNT__=//p')"
  printf '%s\n' "$count" | sed '/^__COUNT__=/d' >> "$OUT"
  [ "$n" -lt 100 ] && break
  page=$((page + 1))
done

printf '\npolicy:\n  secrets_persisted: false\n  tokens_persisted: false\n  auto_deploy: false\n' >> "$OUT"
chmod 0644 "$OUT"
printf 'GitHub registry synced: %s\n' "$OUT"
