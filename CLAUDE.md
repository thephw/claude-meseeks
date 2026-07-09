# CLAUDE.md

Guidance for Claude Code when working in this repository.

## What this is

`claude-meseeks` is a Claude Code **plugin** that plays a Mr. Meeseeks voice line when
Claude is genuinely waiting on the user, and when the user sends Claude a prompt. Playback
is handled by a self-contained Go binary (`meeseeks`) with the audio clips embedded in it.
Each category can be toggled on/off via plugin `userConfig` options. The repo root doubles
as the plugin, its own single-plugin marketplace, and the Go module.

## Layout

- `.claude-plugin/plugin.json` — plugin manifest. Does **not** declare a `hooks` key:
  Claude Code auto-loads `hooks/hooks.json`, so referencing it here causes a
  "Duplicate hooks file" load error. Declares `userConfig` boolean toggles
  (`enableDone`/`enableAsking`/`enableFeedback`, default `true`) — Claude Code exports these
  to hooks as `CLAUDE_PLUGIN_OPTION_<key>` env vars.
- `.claude-plugin/marketplace.json` — self-referential marketplace (`source: "./"`).
- `hooks/hooks.json` — `Notification` and `UserPromptSubmit` hooks, both → `scripts/play.sh notify`.
- `main.go` / `player.go` / `detach_{unix,windows}.go` — the `meeseeks` CLI.
- `scripts/play.sh` — hook launcher: exec the prebuilt `bin/` binary for this platform,
  else `go build` fallback, else silent. Forwards all args + stdin to the binary.
- `scripts/build.sh` — regenerates the `bin/` prebuilts for all platforms.
- `bin/meeseeks-<os>-<arch>` — committed prebuilt binaries (no Go needed to *use* the plugin).
- `audio/{done,asking,feedback}/` — clip pools, **embedded** into the binary via `//go:embed`.
  `done` holds only `ALL DONE`; `feedback` is the eager-acknowledgement pool played on prompt
  submit; `asking` is everything else (former permission clips + former `extra` pool).

## Trigger model (important — this is the fix for spurious audio)

`meeseeks notify` reads the hook event JSON on stdin and maps it to a clip category via the
pure function `categoryForEvent(hook_event_name, notification_type)`:

- `UserPromptSubmit` event → a `feedback` clip (you just sent Claude a prompt)
- `Notification` + `idle_prompt` → a `done` clip (Claude finished, it's your turn)
- `Notification` + `permission_prompt` → an `asking` clip (Claude needs approval)
- everything else (`agent_completed`, `agent_needs_input`, `auth_success`, `elicitation_*`)
  → **silence**

Both hooks (`Notification`, `UserPromptSubmit`) invoke the same `play.sh notify`; the binary
distinguishes them by `hook_event_name`. Playback is then gated per category by
`categoryEnabled`, which reads the `CLAUDE_PLUGIN_OPTION_<key>` toggle (defaults to enabled;
only an explicit `false`/`0`/`off`/`no` silences a category). Manual `meeseeks play <cat>` is
**not** gated — it always plays, for testing.

**Do not add a `Stop` hook.** `Stop` fires on every turn end including auto-continuations,
which plays sounds when the user isn't actually being waited on (the original bug). The
`hook_event_name`/`notification_type` filter is the reliable signal. `categoryForEvent` and
`categoryEnabled` are pure functions covered by `main_test.go`.

## Key invariants

- **Hooks must never block or fail.** `meeseeks` always exits 0 on the hook paths, plays
  detached (own process group), and is silent when a category is empty, no player is found,
  or the payload is missing/garbage.
- **Audio is embedded**, so changing clips requires a rebuild: edit `audio/`, run
  `./scripts/build.sh`, commit `bin/`.
- **Filenames must not contain apostrophes** (`'`) — `go:embed` rejects them — and must end
  in `.mp3`. Categorizing is just moving files between `audio/{done,asking,feedback}/`.
- Playback shells out to the first available system player: `afplay` → `ffplay` → `mpg123`
  → `paplay` → `aplay` → Windows PowerShell.
- Cross-compilation is pure Go (no cgo, since we shell out), so `scripts/build.sh` works
  from any platform.

## Testing changes

```
export PATH="$HOME/.asdf/shims:$PATH"          # Go is pinned via .tool-versions (asdf)
go test ./...                                   # unit tests (event mapping + toggles)
go vet ./...
go build -o /tmp/meeseeks .

/tmp/meeseeks list all                          # inventory
/tmp/meeseeks play feedback --wait              # hear a clip (blocking)
printf '{"hook_event_name":"UserPromptSubmit"}'                             | /tmp/meeseeks notify   # → feedback
printf '{"hook_event_name":"Notification","notification_type":"idle_prompt"}' | /tmp/meeseeks notify   # → done
printf '{"hook_event_name":"Notification","notification_type":"idle_prompt"}' | CLAUDE_PLUGIN_OPTION_enableDone=false /tmp/meeseeks notify   # → silent (toggled off)
printf '{"hook_event_name":"Notification","notification_type":"agent_completed"}' | /tmp/meeseeks notify   # → silent
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
