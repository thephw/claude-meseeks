#!/usr/bin/env bash
# Launcher for the `meeseeks` Go binary — the hook entry point.
# Usage: play.sh <args...>   (forwarded verbatim to the binary)
#   e.g. play.sh notify        (Notification hook; reads JSON on stdin)
#        play.sh play done      (manual)
#
# Runs the prebuilt binary for this platform; if none matches, builds from
# source (requires Go) into a cache. Must never block or fail the hook, so
# every path ends in exit 0. The binary itself detaches playback. stdin is
# inherited through exec so `notify` can read the hook payload.

root="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
[ "$#" -eq 0 ] && set -- play done

# Normalize OS/arch to the bin/ naming scheme.
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
esac

prebuilt="${root}/bin/meeseeks-${os}-${arch}"
if [ -x "$prebuilt" ]; then
  exec "$prebuilt" "$@"
fi

# Fallback: build from source if a Go toolchain is available.
if command -v go >/dev/null 2>&1; then
  cache="${HOME}/.cache/claude-meseeks"
  built="${cache}/meeseeks"
  mkdir -p "$cache" 2>/dev/null
  if [ ! -x "$built" ] || [ "${root}/main.go" -nt "$built" ]; then
    ( cd "$root" && go build -o "$built" . ) >/dev/null 2>&1 || exit 0
  fi
  [ -x "$built" ] && exec "$built" "$@"
fi

# No binary and no Go: stay silent rather than break the session.
exit 0
