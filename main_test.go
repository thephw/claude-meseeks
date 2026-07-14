package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

// TestCategoryForEvent pins the core of the auto-mode fix: only genuine
// "waiting for the user" events (and the UserPromptSubmit feedback event)
// produce a sound; everything else — especially background-agent and auth
// events that fire while Claude is still working — must stay silent.
func TestCategoryForEvent(t *testing.T) {
	cases := []struct {
		hookEventName    string
		notificationType string
		wantCategory     string
		wantPlay         bool
	}{
		// UserPromptSubmit maps to feedback regardless of notification_type.
		{"UserPromptSubmit", "", "feedback", true},
		{"UserPromptSubmit", "idle_prompt", "feedback", true},
		// Notification events switch on notification_type.
		{"Notification", "idle_prompt", "done", true},
		{"Notification", "permission_prompt", "asking", true},
		{"Notification", "agent_completed", "", false},
		{"Notification", "agent_needs_input", "", false},
		{"Notification", "auth_success", "", false},
		{"Notification", "elicitation_dialog", "", false},
		{"Notification", "elicitation_complete", "", false},
		{"Notification", "", "", false},
		{"Notification", "something_new", "", false},
		// No event name at all: fall back to notification_type mapping.
		{"", "idle_prompt", "done", true},
		{"", "", "", false},
	}
	for _, c := range cases {
		gotCategory, gotPlay := categoryForEvent(c.hookEventName, c.notificationType)
		if gotCategory != c.wantCategory || gotPlay != c.wantPlay {
			t.Errorf("categoryForEvent(%q, %q) = (%q, %v); want (%q, %v)",
				c.hookEventName, c.notificationType, gotCategory, gotPlay, c.wantCategory, c.wantPlay)
		}
	}
}

// TestCategoryEnabled verifies the toggle precedence: an explicit state-file
// entry wins over the env var, which wins over the default (enabled). A category
// is enabled by default (unset/empty/unrecognized), and only an explicit falsey
// value silences it. Both the exact-case and upper-case env var names are honored.
func TestCategoryEnabled(t *testing.T) {
	cases := []struct {
		name     string
		category string
		env      map[string]string
		state    map[string]bool
		want     bool
	}{
		{"default enabled when unset", "done", nil, nil, true},
		{"empty stays enabled", "asking", map[string]string{"CLAUDE_PLUGIN_OPTION_enableAsking": ""}, nil, true},
		{"true stays enabled", "feedback", map[string]string{"CLAUDE_PLUGIN_OPTION_enableFeedback": "true"}, nil, true},
		{"false disables", "done", map[string]string{"CLAUDE_PLUGIN_OPTION_enableDone": "false"}, nil, false},
		{"zero disables", "asking", map[string]string{"CLAUDE_PLUGIN_OPTION_enableAsking": "0"}, nil, false},
		{"off disables", "feedback", map[string]string{"CLAUDE_PLUGIN_OPTION_enableFeedback": "off"}, nil, false},
		{"NO uppercase value disables", "done", map[string]string{"CLAUDE_PLUGIN_OPTION_enableDone": "NO"}, nil, false},
		{"uppercase key honored", "done", map[string]string{"CLAUDE_PLUGIN_OPTION_ENABLEDONE": "false"}, nil, false},
		{"garbage value stays enabled", "asking", map[string]string{"CLAUDE_PLUGIN_OPTION_enableAsking": "maybe"}, nil, true},
		{"unknown category always enabled", "extra", map[string]string{"CLAUDE_PLUGIN_OPTION_enableExtra": "false"}, nil, true},
		// State file takes precedence over env and default.
		{"state false beats env true", "done", map[string]string{"CLAUDE_PLUGIN_OPTION_enableDone": "true"}, map[string]bool{"done": false}, false},
		{"state true beats env false", "feedback", map[string]string{"CLAUDE_PLUGIN_OPTION_enableFeedback": "false"}, map[string]bool{"feedback": true}, true},
		{"state false beats default", "asking", nil, map[string]bool{"asking": false}, false},
		{"state for other category ignored", "done", nil, map[string]bool{"asking": false}, true},
	}
	for _, c := range cases {
		getenv := func(k string) string { return c.env[k] }
		if got := categoryEnabled(c.category, getenv, c.state); got != c.want {
			t.Errorf("%s: categoryEnabled(%q) = %v; want %v", c.name, c.category, got, c.want)
		}
	}
}

// TestResolveCategories checks argument expansion: "all"/empty means every
// category, a known name maps to itself, and an unknown name yields nil.
func TestResolveCategories(t *testing.T) {
	cases := []struct {
		arg  string
		want []string
	}{
		{"all", []string{"done", "asking", "feedback"}},
		{"", nil}, // empty is an error, NOT "all" — guards against a failed slash-command substitution
		{"done", []string{"done"}},
		{"feedback", []string{"feedback"}},
		{"nope", nil},
	}
	for _, c := range cases {
		if got := resolveCategories(c.arg); !reflect.DeepEqual(got, c.want) {
			t.Errorf("resolveCategories(%q) = %v; want %v", c.arg, got, c.want)
		}
	}
}

// TestStateRoundTrip verifies persistence: writeState then readState returns the
// same map, and readState on a missing file yields an empty (never nil) map.
func TestStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	if got := readState(); len(got) != 0 {
		t.Errorf("readState() on missing file = %v; want empty", got)
	}

	want := map[string]bool{"feedback": false, "done": true}
	if err := writeState(want); err != nil {
		t.Fatalf("writeState: %v", err)
	}
	if p, _ := statePath(); filepath.Dir(filepath.Dir(p)) != dir {
		t.Errorf("statePath() = %q; not under XDG_CONFIG_HOME %q", p, dir)
	}
	if got := readState(); !reflect.DeepEqual(got, want) {
		t.Errorf("readState() = %v; want %v", got, want)
	}
}
