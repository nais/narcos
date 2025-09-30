#!/usr/bin/env bash
#MISE description="Generate release information using git-cliff"
#MISE depends=["build", "completions"]
set -euo pipefail

binary="narc"
if [[ "$GOOS" == "windows" ]]; then
  binary="narc.exe"
  mv "bin/narc" "bin/$binary"

  if [[ -n "$SIGN_CERT" && -n "$SIGN_KEY" ]]; then
    sudo apt-get update
    sudo apt-get install --yes osslsigncode

    echo "$SIGN_CERT" > nais.crt
    echo "$SIGN_KEY" > nais.key

    osslsigncode sign -certs nais.crt -key nais.key -n "narc-cli" -i "https://github.com/nais/narcos" -verbose -in "bin/$binary" -out "bin/narc-signed"
    mv "bin/narc-signed" "bin/$binary"
  fi
fi

tar -zcf "narc-cli_${GOOS}_${GOARCH}.tgz" ./completions -C bin/ "$binary"
