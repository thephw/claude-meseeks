# claude-meseeks 🔵

*"I'm Mr. Meeseeks! Look at me!"*

A [Claude Code](https://code.claude.com) plugin that plays a Mr. Meeseeks voice line
whenever Claude stops and is waiting for you.

- **When Claude finishes a turn** (`Stop` hook) → a satisfied/finished clip from `audio/done/`
  (*"All done!"*, *"Ooh yeah!"*, *"Yes siree!"* …).
- **When Claude pauses for permission or input** (`Notification` hook) → an asking/coaching
  clip from `audio/asking/` (*"Can you help me?"*, *"You mind if we get back to the task?"* …).

Clips are chosen at random within the appropriate category. Playback is detached and
non-blocking, so a long line never freezes your prompt.

## Install

This repository is both the plugin and its own marketplace.

```
/plugin marketplace add thephw/claude-meseeks
/plugin install claude-meseeks@claude-meseeks
```

Or, from a local clone:

```
/plugin marketplace add /path/to/claude-meseeks
/plugin install claude-meseeks@claude-meseeks
```

Restart or reload Claude Code and finish a turn — you should hear Meeseeks.

## Requirements

An audio player on your `PATH`. The script auto-detects, in order:
`afplay` (macOS, built in) → `ffplay` → `mpg123` → `paplay` → `aplay` → Windows PowerShell
`Media.SoundPlayer`. On macOS nothing extra is needed. On Linux, install `ffmpeg`
(for `ffplay`) or `mpg123`.

## Customizing clips

Audio lives under `audio/`, sorted into three folders:

- `audio/done/` — played on turn-end.
- `audio/asking/` — played on permission/input prompts.
- `audio/extra/` — kept but **unused by default** (longer/narrative/darker lines).

To change what plays, just move `.mp3` files between `done/` and `asking/`, or drop your
own `.mp3` files in. No code changes needed — the script picks a random file from whichever
folder matches the event.

## How it works

`hooks/hooks.json` maps the `Stop` and `Notification` events to
`scripts/play.sh <category>`, which picks a random `.mp3` from `audio/<category>/` and
plays it in a detached background process. The hook always exits 0, so it never blocks or
errors your session.

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
