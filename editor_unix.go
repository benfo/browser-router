//go:build darwin || linux

package main

import (
	"os"
	"os/exec"
)

func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		for _, candidate := range []string{"nano", "vim", "vi"} {
			if p, err := exec.LookPath(candidate); err == nil {
				editor = p
				break
			}
		}
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
