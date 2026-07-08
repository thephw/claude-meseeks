# CLAUDE.md

Guidance for Claude Code when working in this repository.

## What this is

`claude-meseeks` is a Claude Code **plugin** that plays a Mr. Meeseeks voice line when
Claude is genuinely waiting on the user. Playback is handled by a self-contained Go binary
(`meeseeks`) with the audio clips embedded in it. The repo root doubles as the plugin, its
own single-plugin marketplace, and the Go module.

## Layout

- `.claude-plugin/plugin.json` — plugin manifest. Does **not** declare a `hooks` key:
  Claude Code auto-loads `hooks/hooks.json`, so referencing it here causes a
  "Duplicate hooks file" load error.
- `.claude-plugin/marketplace.json` — self-referential marketplace (`source: "./"`).
- `hooks/hooks.json` — a single `Notification` hook → `scripts/play.sh notify`.
- `main.go` / `player.go` / `detach_{unix,windows}.go` — the `meeseeks` CLI.
- `scripts/play.sh` — hook launcher: exec the prebuilt `bin/` binary for this platform,
  else `go build` fallback, else silent. Forwards all args + stdin to the binary.
- `scripts/build.sh` — regenerates the `bin/` prebuilts for all platforms.
- `bin/meeseeks-<os>-<arch>` — committed prebuilt binaries (no Go needed to *use* the plugin).
- `audio/{done,asking,extra}/` — clip pools, **embedded** into the binary via `//go:embed`.

## Trigger model (important — this is the fix for spurious audio)

Only the `Notification` event drives playback. `meeseeks notify` parses the event JSON on
stdin and switches on `notification_type`:

- `idle_prompt` → a `done` clip (Claude finished, it's your turn)
- `permission_prompt` → an `asking` clip (Claude needs approval)
- everything else (`agent_completed`, `agent_needs_input`, `auth_success`, `elicitation_*`)
  → **silence**

**Do not add a `Stop` hook.** `Stop` fires on every turn end including auto-continuations,
which plays sounds when the user isn't actually being waited on (the original bug). The
`notification_type` filter is the reliable "it's your turn" signal. This logic lives in the
pure function `categoryForNotification` and is covered by `main_test.go`.

## Key invariants

- **Hooks must never block or fail.** `meeseeks` always exits 0 on the hook paths, plays
  detached (own process group), and is silent when a category is empty, no player is found,
  or the payload is missing/garbage.
- **Audio is embedded**, so changing clips requires a rebuild: edit `audio/`, run
  `./scripts/build.sh`, commit `bin/`.
- **Filenames must not contain apostrophes** (`'`) — `go:embed` rejects them — and must end
  in `.mp3`. Categorizing is just moving files between `audio/{done,asking,extra}/`.
- Playback shells out to the first available system player: `afplay` → `ffplay` → `mpg123`
  → `paplay` → `aplay` → Windows PowerShell.
- Cross-compilation is pure Go (no cgo, since we shell out), so `scripts/build.sh` works
  from any platform.

## Testing changes

```
export PATH="$HOME/.asdf/shims:$PATH"          # Go is pinned via .tool-versions (asdf)
go test ./...                                   # unit tests (notification mapping)
go vet ./...
go build -o /tmp/meeseeks .

/tmp/meeseeks list all                          # inventory
/tmp/meeseeks play done --wait                  # hear a clip (blocking)
printf '{"notification_type":"idle_prompt"}'      | /tmp/meeseeks notify   # → done
printf '{"notification_type":"agent_completed"}'  | /tmp/meeseeks notify   # → silent
```

Validate manifests after edits: `python3 -m json.tool` on the three JSON files.

After changing behavior, rebuild prebuilts (`./scripts/build.sh`), reinstall
(`claude plugin uninstall`/`install`, or bump the version), and `/reload-plugins`.

## Git workflow

- **Never commit to `main` directly.** Always branch, push, and open a PR — even for
  small changes. Let the user merge.
- Create a topic branch (e.g. `feat/…`, `fix/…`, `docs/…`), commit there, push, and open
  the PR with `gh pr create`.

## Conventions

- Clip filenames keep their ` - AUDIO FROM JAYUZUMI.COM.mp3` suffix; `clipName` strips it
  for display. The binary matches `*.mp3`, so exact names don't matter to playback.
- Audio is Rick and Morty IP included for personal use — see `README.md` before
  redistributing publicly.
