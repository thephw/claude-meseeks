package main

import "testing"

// TestCategoryForNotification pins the core of the auto-mode fix: only genuine
// "waiting for the user" notification types produce a sound; everything else —
// especially background-agent and auth events that fire while Claude is still
// working — must stay silent.
func TestCategoryForNotification(t *testing.T) {
	cases := []struct {
		notificationType string
		wantCategory     string
		wantPlay         bool
	}{
		{"idle_prompt", "done", true},
		{"permission_prompt", "asking", true},
		{"agent_completed", "", false},
		{"agent_needs_input", "", false},
		{"auth_success", "", false},
		{"elicitation_dialog", "", false},
		{"elicitation_complete", "", false},
		{"", "", false},
		{"something_new", "", false},
	}
	for _, c := range cases {
		gotCategory, gotPlay := categoryForNotification(c.notificationType)
		if gotCategory != c.wantCategory || gotPlay != c.wantPlay {
			t.Errorf("categoryForNotification(%q) = (%q, %v); want (%q, %v)",
				c.notificationType, gotCategory, gotPlay, c.wantCategory, c.wantPlay)
		}
	}
}
