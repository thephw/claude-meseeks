package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// categories are the clip categories that can be individually muted.
var categories = []string{"done", "asking", "feedback"}

// statePath returns the path to the persisted per-category enable state.
// Preference order: $XDG_CONFIG_HOME, then ~/.config, then os.UserConfigDir()
// (which covers Windows). This parallels the ~/.cache/claude-meseeks build
// cache used by scripts/play.sh.
func statePath() (string, error) {
	var dir string
	switch {
	case os.Getenv("XDG_CONFIG_HOME") != "":
		dir = os.Getenv("XDG_CONFIG_HOME")
	default:
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			dir = filepath.Join(home, ".config")
		} else if ucd, err := os.UserConfigDir(); err == nil {
			dir = ucd
		} else {
			return "", err
		}
	}
	return filepath.Join(dir, "claude-meseeks", "state.json"), nil
}

// readState loads the per-category override map. Best-effort: a missing file or
// unparseable contents yields an empty map, and it never errors — the notify
// hot path must not fail. A category absent from the map means "use the
// default" (enabled), so a first run with no file behaves exactly as before.
func readState() map[string]bool {
	p, err := statePath()
	if err != nil {
		return map[string]bool{}
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return map[string]bool{}
	}
	var m map[string]bool
	if err := json.Unmarshal(data, &m); err != nil || m == nil {
		return map[string]bool{}
	}
	return m
}

// writeState persists the override map atomically (temp file + rename). Only
// user-invoked verbs (enable/disable/toggle) call this — never notify.
func writeState(m map[string]bool) error {
	p, err := statePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp", p)
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// resolveCategories expands a category argument to the concrete categories it
// affects. Only the explicit "all" means every category. An empty or
// unrecognized arg returns nil so callers report an error — deliberately: the
// slash commands pass the category through a shell placeholder, and if that ever
// fails to substitute we want a visible "specify a category" error, not a silent
// "mute everything".
func resolveCategories(arg string) []string {
	if arg == "all" {
		return append([]string(nil), categories...)
	}
	for _, c := range categories {
		if c == arg {
			return []string{arg}
		}
	}
	return nil
}
