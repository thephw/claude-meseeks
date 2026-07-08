# CLAUDE.md

Guidance for Claude Code when working in this repository.

## What this is

`claude-meseeks` is a Claude Code **plugin** that plays a Mr. Meeseeks voice line
whenever Claude stops and is waiting for the user. The repo root doubles as both the
plugin and its own single-plugin marketplace.

## Layout

- `.claude-plugin/plugin.json` — plugin manifest. Does **not** declare a `hooks` key:
  Claude Code auto-loads `hooks/hooks.json`, so referencing it here causes a
  "Duplicate hooks file" load error.
- `.claude-plugin/marketplace.json` — self-referential marketplace (`source: "./"`).
- `hooks/hooks.json` — maps `Stop` → `play.sh done` and `Notification` → `play.sh asking`.
- `scripts/play.sh` — picks a random `.mp3` from `audio/<category>/` and plays it detached.
- `audio/{done,asking,extra}/` — clip pools. `done` = turn-end, `asking` = needs-input,
  `extra` = kept but unused by default.

## Key invariants

- **The hook must never block or fail.** `play.sh` always `exit 0`, plays in a detached
  background subshell, and exits quietly if the category dir is missing/empty.
- **`${CLAUDE_PLUGIN_ROOT}`** resolves to the repo root at runtime; `play.sh` falls back
  to its own location when the var is unset (so it's testable standalone).
- **Categorizing clips is just file placement** — move `.mp3`s between `audio/done`,
  `audio/asking`, `audio/extra`. No code changes needed; the script globs the folder.
- Player auto-detect order (mp3-capable first): `afplay` → `ffplay` → `mpg123` →
  `paplay` → `aplay` → Windows PowerShell.

## Testing changes

Run the script directly without installing the plugin:

```
CLAUDE_PLUGIN_ROOT="$(pwd)" bash scripts/play.sh done      # hear a turn-end clip
CLAUDE_PLUGIN_ROOT="$(pwd)" bash scripts/play.sh asking    # hear a needs-input clip
CLAUDE_PLUGIN_ROOT="$(pwd)" bash scripts/play.sh missing   # silent, exit 0
```

Validate config after edits: `python3 -m json.tool` on the three JSON files, and
`bash -n scripts/play.sh`.

## Conventions

- Clip filenames keep their original ` - AUDIO FROM JAYUZUMI.COM.mp3` suffix; the script
  matches `*.mp3` so exact names don't matter.
- Audio is Rick and Morty IP included for personal use — see the note in `README.md`
  before redistributing publicly.
