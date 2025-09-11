#!/usr/bin/env bash
#MISE description="Build narc"
go build \
  -ldflags "-s -w -X github.com/nais/narcos/internal/version.Version=${VERSION:-local}" \
  -o bin/narc ./