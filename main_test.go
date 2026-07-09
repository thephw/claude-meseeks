package main

import "testing"

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

// TestCategoryEnabled verifies the per-category toggle: a category is enabled
// by default (unset/empty/unrecognized), and only an explicit falsey value
// silences it. Both the exact-case and upper-case env var names are honored.
func TestCategoryEnabled(t *testing.T) {
	cases := []struct {
		name     string
		category string
		env      map[string]string
		want     bool
	}{
		{"default enabled when unset", "done", nil, true},
		{"empty stays enabled", "asking", map[string]string{"CLAUDE_PLUGIN_OPTION_enableAsking": ""}, true},
		{"true stays enabled", "feedback", map[string]string{"CLAUDE_PLUGIN_OPTION_enableFeedback": "true"}, true},
		{"false disables", "done", map[string]string{"CLAUDE_PLUGIN_OPTION_enableDone": "false"}, false},
		{"zero disables", "asking", map[string]string{"CLAUDE_PLUGIN_OPTION_enableAsking": "0"}, false},
		{"off disables", "feedback", map[string]string{"CLAUDE_PLUGIN_OPTION_enableFeedback": "off"}, false},
		{"NO uppercase value disables", "done", map[string]string{"CLAUDE_PLUGIN_OPTION_enableDone": "NO"}, false},
		{"uppercase key honored", "done", map[string]string{"CLAUDE_PLUGIN_OPTION_ENABLEDONE": "false"}, false},
		{"garbage value stays enabled", "asking", map[string]string{"CLAUDE_PLUGIN_OPTION_enableAsking": "maybe"}, true},
		{"unknown category always enabled", "extra", map[string]string{"CLAUDE_PLUGIN_OPTION_enableExtra": "false"}, true},
	}
	for _, c := range cases {
		getenv := func(k string) string { return c.env[k] }
		if got := categoryEnabled(c.category, getenv); got != c.want {
			t.Errorf("%s: categoryEnabled(%q) = %v; want %v", c.name, c.category, got, c.want)
		}
	}
}
