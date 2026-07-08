//go:build !windows

package main

import "syscall"

// detachAttr puts the player in its own process group so it is not killed when
// this process exits.
func detachAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
