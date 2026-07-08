#!/usr/bin/env bash
set -euo pipefail

ROOT="${1:-$HOME/exchange-lab/sources}"

if [ ! -d "$ROOT" ]; then
  echo "source root not found: $ROOT" >&2
  exit 1
fi

for repo in "$ROOT"/*; do
  [ -d "$repo/.git" ] || continue
  name="$(basename "$repo")"
  echo "## $name"
  echo "path: $repo"
  echo "branch: $(git -C "$repo" rev-parse --abbrev-ref HEAD 2>/dev/null || true)"
  echo "latest_commit: $(git -C "$repo" rev-parse HEAD 2>/dev/null || true)"
  echo "license_file: $(find "$repo" -maxdepth 2 -iname 'LICENSE*' -type f | head -n 1)"
  echo "readme_file: $(find "$repo" -maxdepth 2 -iname 'README*' -type f | head -n 1)"
  echo "dependency_files:"
  find "$repo" -maxdepth 3 -type f \( -name 'go.mod' -o -name 'pom.xml' -o -name 'package.json' -o -name 'build.gradle' -o -name 'requirements.txt' -o -name 'Cargo.toml' \) | sed 's/^/  - /'
  echo "language_summary:"
  find "$repo" -type f | sed 's/.*\.//' | sort | uniq -c | sort -nr | head -n 12 | sed 's/^/  /'
  if command -v gitleaks >/dev/null 2>&1; then
    echo "secret_scan: running gitleaks"
    gitleaks detect --source "$repo" --no-banner --redact || true
  else
    echo "secret_scan: skipped, gitleaks not installed"
  fi
  echo
done

