//go:build windows

package main

import (
	"os"
	"syscall"
)

func init() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	attachConsole := kernel32.NewProc("AttachConsole")

	// ATTACH_PARENT_PROCESS = 0xFFFFFFFF
	// If launched by the OS (no parent console): fails silently — no console window, no flash
	// If launched from cmd/PowerShell: succeeds — output flows to the existing console
	ret, _, _ := attachConsole.Call(0xFFFFFFFF)
	if ret != 0 {
		// Re-bind Go's stdout/stderr to the now-attached console handles
		stdout, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
		stderr, _ := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)
		os.Stdout = os.NewFile(uintptr(stdout), "stdout")
		os.Stderr = os.NewFile(uintptr(stderr), "stderr")
	}
}
