# claude-meseeks 🔵

*"I'm Mr. Meeseeks! Look at me!"*

A [Claude Code](https://code.claude.com) plugin that plays a Mr. Meeseeks voice line
whenever Claude is genuinely waiting on *you*.

- **When Claude finishes and is waiting for your next prompt** → a satisfied/finished clip
  from `audio/done/` (*"All done!"*, *"Ooh yeah!"*, *"Yes siree!"* …).
- **When Claude needs your approval** → an asking/coaching clip from `audio/asking/`
  (*"Can you help me?"*, *"You mind if we get back to the task?"* …).

Both are driven by the `Notification` event, filtered by `notification_type` so it fires
**only** when you're actually needed. Autonomous work — auto-accept/bypass-permissions runs,
background-agent and subagent activity, auth refreshes — stays **silent**. Clips are random
within the category, and playback is detached and non-blocking, so a long line never freezes
your prompt.

## Install

This repository is both the plugin and its own marketplace.

```
/plugin marketplace add thephw/claude-meseeks
/plugin install mr-meeseeks@claude-meseeks
```

Or, from a local clone:

```
/plugin marketplace add /path/to/claude-meseeks
/plugin install mr-meeseeks@claude-meseeks
```

Restart or reload Claude Code and finish a turn — you should hear Meeseeks.

## Requirements

An audio player on your `PATH`. The tool auto-detects, in order:
`afplay` (macOS, built in) → `ffplay` → `mpg123` → `paplay` → `aplay` → Windows PowerShell
`Media.SoundPlayer`. On macOS nothing extra is needed. On Linux, install `ffmpeg`
(for `ffplay`) or `mpg123`.

No Go toolchain is required to *use* the plugin — prebuilt binaries ship in `bin/`. Go is
only needed to rebuild them (see below).

## The `meeseeks` CLI

Playback is handled by a small Go program, `meeseeks`, with the clips embedded directly in
the binary. You can drive it by hand too:

```
meeseeks play                      # random "done" clip, detached
meeseeks play asking               # random "asking" clip
meeseeks play extra --wait         # a wildcard clip, blocking until it finishes
meeseeks play --clip "ALL DONE"    # a specific clip by name
meeseeks list all                  # list every embedded clip
```

## How it works

`hooks/hooks.json` registers a single `Notification` hook that runs `scripts/play.sh notify`.
That launcher execs the prebuilt `bin/meeseeks-<os>-<arch>` for your platform (falling back
to `go build` from source if there's no matching binary, or staying silent if neither is
available), passing the event's JSON through on stdin.

`meeseeks notify` reads that JSON and looks at `notification_type`:

| `notification_type`                                      | Result           |
| -------------------------------------------------------- | ---------------- |
| `idle_prompt` (Claude done, awaiting your prompt)        | random `done`    |
| `permission_prompt` (Claude needs approval)              | random `asking`  |
| anything else (`agent_completed`, `auth_success`, …)     | silence          |

The chosen clip is extracted from the embedded audio to a cache dir and handed to a system
player in a detached process. Every path exits 0, so the hook never blocks or errors your
session.

> **Why not the `Stop` hook?** `Stop` fires at the end of *every* turn — including
> auto-continuations — so it plays sounds when you aren't actually being waited on. The
> `Notification` type filter is the reliable signal for "it's your turn."

## Customizing clips

Clips live under `audio/`, sorted into three folders that map to behavior:

- `audio/done/` — played on turn-end.
- `audio/asking/` — played on permission/input prompts.
- `audio/extra/` — embedded but **unused by the hooks** (longer/narrative/darker lines);
  still reachable via `meeseeks play extra`.

To change what plays, move `.mp3` files between the folders or drop your own in, then
**rebuild the binaries** so the new clips are re-embedded:

```
./scripts/build.sh    # regenerates bin/ for all platforms
```

Two constraints: filenames must end in `.mp3`, and — because of a `go:embed` restriction —
must not contain apostrophes (`'`).

## Why Meeseeks? On single-purpose sessions

The theme isn't just a joke — it's a working philosophy.

A Mr. Meeseeks is summoned to accomplish **one task**. It exists only until that task is
done, and then it poofs out of existence, satisfied. Give a Meeseeks a single, concrete goal
("help me finish this putt") and it's cheerful and effective. Give it a vague or unbounded
one, or keep it alive long past its purpose, and things degrade fast — *"existence is
pain, Jerry!"* — until you get a room full of increasingly unhinged Meeseeks.

A Claude Code session works best the same way:

- **Summon it for one goal.** A session scoped to a single, well-defined objective —
  "add this endpoint", "fix this failing test", "write this plugin" — is focused and sharp,
  the same way a fresh Meeseeks is.
- **Let it finish, then let it go.** When the goal is met, end the session. Start a new one
  for the next task. A fresh session with a clean context beats a stale one every time.
- **Beware the long-lived session.** Dragging one conversation across many unrelated goals
  is how you get the Meeseeks box problem: context piles up, focus drifts, earlier tangents
  pollute later work, and quality slides. Long ≠ productive.

So: treat each session like a Meeseeks. One purpose. Accomplish it. Poof. 🔵

## Credits

Inspired by and audio sourced from the
[Mr. Meeseeks Soundboard](https://jayuzumi.com/mr-meeseeks-soundboard) at jayuzumi.com.
Thanks for the clips! 🔵

## Note on the audio

The voice clips are from *Rick and Morty* (via the
[jayuzumi.com Mr. Meeseeks Soundboard](https://jayuzumi.com/mr-meeseeks-soundboard)) and are
included here for personal, non-commercial fun. They are the property of their respective
rights holders. Please consider those rights before redistributing this plugin publicly or
swap in your own audio.
