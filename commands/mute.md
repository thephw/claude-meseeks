---
description: Mute a Mr. Meeseeks sound category (done, asking, feedback, or all)
argument-hint: [done|asking|feedback|all]
arguments: [category]
allowed-tools: Bash
disable-model-invocation: true
---
!`"${CLAUDE_PLUGIN_ROOT}"/scripts/play.sh disable "$category"`

The command above tried to mute the given category. If the output shows an error
(e.g. "specify a category"), relay it and remind the user to pass one of
done/asking/feedback/all. Otherwise, tell the user which categories are now on/off
using the status output above, in one short line.
