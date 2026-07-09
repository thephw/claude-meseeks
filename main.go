// Command meeseeks plays a random Mr. Meeseeks voice line.
//
// It is the engine behind the claude-meseeks Claude Code plugin: the
// Notification and UserPromptSubmit hooks invoke `meeseeks notify`, which reads
// the hook event on stdin and plays the matching clip category (subject to the
// per-category enable toggles). The audio clips are embedded into the binary,
// so it is fully self-contained; playback is handed off to whatever system
// audio player is available.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"strings"
)

//go:embed audio
var audioFS embed.FS

const version = "0.3.0"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		return
	}
	switch args[0] {
	case "play":
		cmdPlay(args[1:]) // best-effort; never fails the hook
	case "notify":
		cmdNotify() // reads a Notification hook payload on stdin
	case "list":
		os.Exit(cmdList(args[1:]))
	case "version", "-v", "--version":
		fmt.Println("meeseeks " + version)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "meeseeks: unknown command %q\n\n", args[0])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Print(`meeseeks — play a Mr. Meeseeks voice line

Usage:
  meeseeks play [category] [--wait] [--clip <substring>]
  meeseeks notify           # decide from a hook event payload on stdin
  meeseeks list [category|all]
  meeseeks version

Categories: done (default), asking, feedback

play picks a random clip from the category and plays it detached (non-blocking).
It ignores the enable toggles, so it always plays — handy for testing a clip.
  --wait            play in the foreground, blocking until the clip finishes
  --clip <substr>   play the first clip whose filename contains <substr>

notify reads a Claude Code hook event (JSON) on stdin and plays a clip only when
the user is genuinely being waited on (and that category's toggle is enabled):
  UserPromptSubmit  -> a "feedback" clip (you just sent Claude a prompt)
  idle_prompt       -> a "done" clip (Claude finished, it's your turn)
  permission_prompt -> an "asking" clip (Claude needs your approval)
  anything else     -> silence (background-agent, auth, elicitation events…)

Toggles: set the plugin's enableDone / enableAsking / enableFeedback options
(exported as CLAUDE_PLUGIN_OPTION_* env vars) to "false" to silence a category.

Examples:
  meeseeks play                      # random "done" clip
  meeseeks play asking
  meeseeks play --clip "ALL DONE" --wait
  meeseeks list all
`)
}

// cmdNotify reads a hook event payload from stdin and plays a clip only for
// events that mean "Claude is waiting on the user" (or that the user just gave
// feedback). Every other Notification type (agent_completed, auth_success,
// elicitation_*, …) is silent — this is what prevents sounds during autonomous
// / auto-accept work. Playback is further gated by the per-category enable
// toggle, so a disabled category stays silent even on a matching event.
func cmdNotify() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(data) == 0 {
		return
	}
	var payload struct {
		HookEventName    string `json:"hook_event_name"`
		NotificationType string `json:"notification_type"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return
	}
	category, ok := categoryForEvent(payload.HookEventName, payload.NotificationType)
	if ok && categoryEnabled(category, os.Getenv) {
		playCategory(category, false)
	}
}

// categoryForEvent maps a hook event to a clip category. A UserPromptSubmit
// event means the user just sent Claude a prompt; otherwise we treat it as a
// Notification event and switch on the notification type. The bool is false for
// anything that should stay silent. Pure function — unit tested.
func categoryForEvent(hookEventName, notificationType string) (string, bool) {
	if hookEventName == "UserPromptSubmit" { // you just gave Claude feedback
		return "feedback", true
	}
	switch notificationType {
	case "idle_prompt": // Claude finished and is waiting for your next prompt
		return "done", true
	case "permission_prompt": // Claude needs your approval to proceed
		return "asking", true
	default: // agent_completed, agent_needs_input, auth_success, elicitation_* …
		return "", false
	}
}

// categoryEnabled reports whether automatic playback is enabled for a category.
// It reads the plugin's per-category toggle, exported by Claude Code as a
// CLAUDE_PLUGIN_OPTION_<key> environment variable. Both the exact-case and
// upper-case variants are checked, since the exported casing is unspecified. A
// category defaults to enabled: an unset or unrecognized value plays, and only
// an explicit falsey value ("false"/"0"/"off"/"no") silences it. Pure given
// getenv — unit tested. getenv is injected so tests need not touch os.Environ.
func categoryEnabled(category string, getenv func(string) string) bool {
	var key string
	switch category {
	case "done":
		key = "enableDone"
	case "asking":
		key = "enableAsking"
	case "feedback":
		key = "enableFeedback"
	default:
		return true
	}
	val := getenv("CLAUDE_PLUGIN_OPTION_" + key)
	if val == "" {
		val = getenv("CLAUDE_PLUGIN_OPTION_" + strings.ToUpper(key))
	}
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "false", "0", "off", "no":
		return false
	default: // "", "true", or anything else -> enabled
		return true
	}
}

// cmdPlay selects a clip and plays it. Every failure path is silent: a hook
// must never error, so "no clips" or "no player" simply produces no sound.
func cmdPlay(args []string) {
	category := "done"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		category = args[0]
		args = args[1:]
	}

	fs := flag.NewFlagSet("play", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	wait := fs.Bool("wait", false, "block until playback finishes")
	clip := fs.String("clip", "", "play the clip whose name contains this substring")
	if err := fs.Parse(args); err != nil {
		return
	}

	if *clip == "" {
		playCategory(category, *wait)
		return
	}

	// --clip: play the first clip whose name contains the substring.
	clips, err := categoryClips(category)
	if err != nil {
		return
	}
	needle := strings.ToLower(*clip)
	for _, c := range clips {
		if strings.Contains(strings.ToLower(c), needle) {
			if dest, err := extractClip(c); err == nil {
				_ = playFile(dest, *wait)
			}
			return
		}
	}
}

// playCategory picks a random clip from the category and plays it. Silent (no
// error) when the category is empty or missing — hooks must never fail.
func playCategory(category string, wait bool) {
	clips, err := categoryClips(category)
	if err != nil || len(clips) == 0 {
		return
	}
	dest, err := extractClip(clips[rand.Intn(len(clips))])
	if err != nil {
		return
	}
	_ = playFile(dest, wait)
}

// cmdList prints the clips in a category (or all categories).
func cmdList(args []string) int {
	cats := []string{"done", "asking", "feedback"}
	if len(args) > 0 && args[0] != "all" {
		cats = []string{args[0]}
	}
	for i, cat := range cats {
		clips, err := categoryClips(cat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "meeseeks: no such category %q\n", cat)
			return 1
		}
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s (%d)\n", cat, len(clips))
		for _, c := range clips {
			fmt.Println("  " + clipName(c))
		}
	}
	return 0
}

// clipName turns "audio/done/ALL DONE - AUDIO FROM JAYUZUMI.COM.mp3" into "ALL DONE".
func clipName(embeddedPath string) string {
	name := strings.TrimSuffix(path.Base(embeddedPath), ".mp3")
	return strings.TrimSuffix(name, " - AUDIO FROM JAYUZUMI.COM")
}
