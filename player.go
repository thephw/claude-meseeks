package main

import (
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
)

// categoryClips returns the embedded *.mp3 paths for a category (e.g. "done").
func categoryClips(category string) ([]string, error) {
	dir := path.Join("audio", category)
	entries, err := fs.ReadDir(audioFS, dir)
	if err != nil {
		return nil, err
	}
	var clips []string
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".mp3" {
			continue
		}
		clips = append(clips, path.Join(dir, e.Name()))
	}
	return clips, nil
}

// extractClip writes an embedded clip to a stable per-user cache path (once) and
// returns the on-disk path. Extraction is needed because system players read
// files, not embedded bytes; caching avoids rewriting on every invocation.
func extractClip(embeddedPath string) (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		root = os.TempDir()
	}
	dest := filepath.Join(root, "claude-meseeks", filepath.FromSlash(embeddedPath))

	if fi, err := os.Stat(dest); err == nil && fi.Size() > 0 {
		return dest, nil // already extracted
	}
	data, err := audioFS.ReadFile(embeddedPath)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return "", err
	}
	return dest, nil
}

// player is a system audio program and the args needed to play a file with it.
type player struct {
	bin  string
	args func(file string) []string
}

// detectPlayer finds the first available audio player. mp3-capable players come
// first; paplay/aplay are WAV-oriented fallbacks. Mirrors the old play.sh order.
func detectPlayer() (player, bool) {
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("powershell.exe"); err == nil {
			return player{p, func(f string) []string {
				return []string{"-NoProfile", "-Command",
					"(New-Object Media.SoundPlayer '" + f + "').PlaySync()"}
			}}, true
		}
	}
	candidates := []player{
		{"afplay", func(f string) []string { return []string{f} }},
		{"ffplay", func(f string) []string { return []string{"-nodisp", "-autoexit", "-loglevel", "quiet", f} }},
		{"mpg123", func(f string) []string { return []string{"-q", f} }},
		{"paplay", func(f string) []string { return []string{f} }},
		{"aplay", func(f string) []string { return []string{"-q", f} }},
	}
	for _, c := range candidates {
		if p, err := exec.LookPath(c.bin); err == nil {
			c.bin = p
			return c, true
		}
	}
	return player{}, false
}

// playFile plays a file. When wait is false it detaches the player into its own
// process group and returns immediately, so a long clip never blocks the hook
// and survives this process exiting. Returns nil when no player is available.
func playFile(file string, wait bool) error {
	p, ok := detectPlayer()
	if !ok {
		return nil
	}
	cmd := exec.Command(p.bin, p.args(file)...)
	if wait {
		return cmd.Run()
	}
	cmd.SysProcAttr = detachAttr()
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
