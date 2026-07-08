#!/usr/bin/env bash
# Play a random Mr. Meeseeks voice line for the given category.
# Usage: play.sh [done|asking]
# Invoked by Claude Code Stop / Notification hooks. Must never block or fail the hook.

category="${1:-done}"
root="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
dir="${root}/audio/${category}"

# Nothing to play? Exit quietly — a hook must never error out.
[ -d "$dir" ] || exit 0

# Pick one random *.mp3 (handles spaces/apostrophes; NUL-delimited).
clip="$(find "$dir" -maxdepth 1 -type f -name '*.mp3' -print0 \
        | tr '\0' '\n' | sort -R | head -n 1)"
[ -n "$clip" ] || exit 0

# Choose the first available audio player for this platform.
play() {
  local f="$1"
  # afplay (macOS) and ffplay/mpg123 decode mp3; paplay/aplay are WAV-only fallbacks.
  if   command -v afplay        >/dev/null 2>&1; then afplay "$f"
  elif command -v ffplay        >/dev/null 2>&1; then ffplay -nodisp -autoexit -loglevel quiet "$f"
  elif command -v mpg123        >/dev/null 2>&1; then mpg123 -q "$f"
  elif command -v paplay        >/dev/null 2>&1; then paplay "$f"
  elif command -v aplay         >/dev/null 2>&1; then aplay -q "$f"
  elif command -v powershell.exe >/dev/null 2>&1; then
    powershell.exe -NoProfile -c "(New-Object Media.SoundPlayer '$f').PlaySync()"
  fi
}

# Detach so a long clip never freezes the prompt and survives the hook exiting.
( play "$clip" >/dev/null 2>&1 & ) &

exit 0
