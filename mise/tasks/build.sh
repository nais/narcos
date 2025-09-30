#!/usr/bin/env bash
#MISE description="Build narc"
set -euo pipefail

version="${VERSION:-local}"

go build \
  -ldflags "-s -w -X github.com/nais/narcos/internal/version.Version=$version" \
  -o bin/narc ./