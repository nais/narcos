#!/usr/bin/env bash
#MISE description="Package the Narc CLI binary as a Debian package"
#MISE depends=["build"]
export ARCH=$GOARCH
GOARCH="" go tool github.com/goreleaser/nfpm/v2/cmd/nfpm package \
  --packager deb \
  --config .nfpm.yaml \
  --target "narc-cli_${VERSION:-local}_${ARCH:-amd64}.deb"