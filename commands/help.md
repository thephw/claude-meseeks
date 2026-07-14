---
description: How to configure Mr. Meeseeks (which sounds play, and how to mute them)
---
Explain the following to the user clearly and concisely (this is a help screen —
do not run any commands, just relay this):

**Mr. Meeseeks plays a voice line in three situations, each its own category:**

- **done** — Claude finished and it's your turn (idle prompt).
- **asking** — Claude needs your approval (permission prompt).
- **feedback** — you just submitted a prompt to Claude.

**All three play by default. To change what plays, use these commands:**

- `/mr-meeseeks:mute <category>` — silence a category (e.g. `/mr-meeseeks:mute feedback`)
- `/mr-meeseeks:unmute <category>` — turn it back on
- `/mr-meeseeks:status` — show which categories are on or off
- `/mr-meeseeks:help` — this screen

`<category>` is required and is one of `done`, `asking`, `feedback`, or `all`.

**Where the setting lives:** choices are saved to
`~/.config/claude-meseeks/state.json` (or `$XDG_CONFIG_HOME/claude-meseeks/`).
They take effect immediately — no `/reload-plugins` or restart needed.

**Why not the `/plugin` "Configure options" screen?** Claude Code currently renders
every plugin option there as a free-text field with no working toggle
([claude-code#74289](https://github.com/anthropics/claude-code/issues/74289)), so
these commands are the reliable way to configure Mr. Meeseeks. Power users can also
set `CLAUDE_PLUGIN_OPTION_enableDone` / `enableAsking` / `enableFeedback` to `false`
in `settings.json`; the state file above takes precedence over those.

You can also drive it directly from the CLI: `meeseeks status`,
`meeseeks disable feedback`, `meeseeks enable all`, `meeseeks toggle done`.
