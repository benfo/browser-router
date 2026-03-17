//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const desktopTmpl = `[Desktop Entry]
Name=Browser Router
Exec={{.Exe}} open %u
Type=Application
Terminal=false
MimeType=x-scheme-handler/http;x-scheme-handler/https;
`

func register() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	desktopDir := filepath.Join(homeDir, ".local", "share", "applications")
	if err := os.MkdirAll(desktopDir, 0755); err != nil {
		return err
	}

	desktopPath := filepath.Join(desktopDir, "browser-router.desktop")
	f, err := os.Create(desktopPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl := template.Must(template.New("desktop").Parse(desktopTmpl))
	if err := tmpl.Execute(f, struct{ Exe string }{Exe: exe}); err != nil {
		return err
	}

	exec.Command("xdg-mime", "default", "browser-router.desktop", "x-scheme-handler/http").Run()
	exec.Command("xdg-mime", "default", "browser-router.desktop", "x-scheme-handler/https").Run()
	exec.Command("update-desktop-database", desktopDir).Run()

	fmt.Printf("Registered. Desktop file: %s\n", desktopPath)
	return nil
}

func unregister() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	desktopPath := filepath.Join(homeDir, ".local", "share", "applications", "browser-router.desktop")
	if err := os.Remove(desktopPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	fmt.Println("Unregistered.")
	return nil
}
