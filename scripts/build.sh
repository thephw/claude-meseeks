#!/usr/bin/env bash
# Rebuild the committed prebuilt binaries in bin/ for all supported platforms.
# Run this whenever the Go source or the audio/ clips change, then commit bin/.
#
# Requires a Go toolchain (this repo pins one via asdf / .tool-versions).
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
mkdir -p bin

targets="darwin/arm64 darwin/amd64 linux/arm64 linux/amd64 windows/amd64"
for t in $targets; do
  os="${t%/*}"; arch="${t#*/}"
  out="bin/meeseeks-${os}-${arch}"
  [ "$os" = windows ] && out="${out}.exe"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -ldflags="-s -w" -o "$out" .
  chmod +x "$out"
  echo "built $out"
done
