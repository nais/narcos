#!/usr/bin/env bash
#MISE description="Check that go fix and gofumpt leave the code unchanged"
set -euo pipefail

mise run gofix
mise run fmt

if ! git diff --exit-code --name-only; then
  echo "Run 'mise run gofix' and 'mise run fmt' locally, then review and commit the changes listed above."
  exit 1
fi
