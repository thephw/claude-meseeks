//go:build windows

package main

import "syscall"

const createNewProcessGroup = 0x00000200 // CREATE_NEW_PROCESS_GROUP

// detachAttr starts the player in a new process group so it is not tied to this
// process's lifetime.
func detachAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: createNewProcessGroup}
}
