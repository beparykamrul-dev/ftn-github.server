#!/usr/bin/env bash
set -euo pipefail

ROOT="${FTN_GITHUB_ROOT:-/opt/ftn-github.server}"
REGISTRY="$ROOT/config/github-repositories.generated.yaml"
REPOS_ROOT="${FTN_GIT_REPOS_ROOT:-/opt/ftn-git/repos}"
SYNC_SOURCE="${FTN_GIT_SOURCE_SYNC:-false}"

"$ROOT/scripts/sync-github-registry.sh"
mkdir -p "$REPOS_ROOT"

if [ "$SYNC_SOURCE" != "true" ]; then
  echo "Metadata sync complete; source sync disabled (FTN_GIT_SOURCE_SYNC=false)."
  exit 0
fi

command -v git >/dev/null || { echo "git is required"; exit 1; }
command -v python3 >/dev/null || { echo "python3 is required"; exit 1; }
: "${GITHUB_TOKEN:?GITHUB_TOKEN must be supplied at runtime}"

export GIT_CONFIG_COUNT=1
export GIT_CONFIG_KEY_0=http.extraheader
export GIT_CONFIG_VALUE_0="Authorization: Bearer $GITHUB_TOKEN"

python3 - "$REGISTRY" "$REPOS_ROOT" <<'PY'
import pathlib, subprocess, sys
p = pathlib.Path(sys.argv[1])
root = pathlib.Path(sys.argv[2])
text = p.read_text()
repos = []
for block in text.split('\n  - name: ')[1:]:
    lines = block.splitlines()
    name = lines[0].strip().strip("'")
    full = next((x.split(': ',1)[1].strip().strip("'") for x in lines if x.strip().startswith('full_name: ')), '')
    clone = next((x.split(': ',1)[1].strip().strip("'") for x in lines if x.strip().startswith('clone_url: ')), '')
    if name and full and clone:
        repos.append((name, full, clone))
for name, full, clone in repos:
    dest = root / name
    if (dest / '.git').exists():
        subprocess.run(['git','-C',str(dest),'fetch','--prune','origin'], check=True)
        subprocess.run(['git','-C',str(dest),'remote','set-url','origin',clone], check=True)
    else:
        subprocess.run(['git','clone','--filter=blob:none',clone,str(dest)], check=True)
    print('SYNC', full)
PY
