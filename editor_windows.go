//go:build windows

package main

import (
	"os"
	"os/exec"
)

func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// notepad is always available; use code/notepad++ if present
		for _, candidate := range []string{"code", "notepad++", "notepad"} {
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
