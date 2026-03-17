//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sys/windows/registry"
)

func register() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	type kv struct{ key, value, data string }

	entries := []struct {
		path string
		kvs  []kv
	}{
		{
			`Software\Classes\BrowserRouter`,
			[]kv{
				{"", "", "Browser Router"},
				{"", "URL Protocol", ""},
			},
		},
		{
			`Software\Classes\BrowserRouter\DefaultIcon`,
			[]kv{{"", "", exe + ",0"}},
		},
		{
			`Software\Classes\BrowserRouter\shell\open\command`,
			[]kv{{"", "", fmt.Sprintf(`"%s" open "%%1"`, exe)}},
		},
		{
			`Software\Clients\StartMenuInternet\BrowserRouter\Capabilities`,
			[]kv{
				{"", "ApplicationName", "Browser Router"},
				{"", "ApplicationDescription", "Routes URLs to different browsers based on rules"},
			},
		},
		{
			`Software\Clients\StartMenuInternet\BrowserRouter\Capabilities\URLAssociations`,
			[]kv{
				{"", "http", "BrowserRouter"},
				{"", "https", "BrowserRouter"},
			},
		},
		{
			`Software\RegisteredApplications`,
			[]kv{{"", "BrowserRouter", `Software\Clients\StartMenuInternet\BrowserRouter\Capabilities`}},
		},
	}

	for _, e := range entries {
		k, _, err := registry.CreateKey(registry.CURRENT_USER, e.path, registry.ALL_ACCESS)
		if err != nil {
			return fmt.Errorf("create key %s: %w", e.path, err)
		}
		for _, kv := range e.kvs {
			k.SetStringValue(kv.key, kv.data)
		}
		k.Close()
	}

	fmt.Println("Registered successfully.")
	fmt.Println("Opening Windows Settings > Default Apps — set 'Browser Router' as your default browser there.")
	exec.Command("cmd", "/c", "start", "ms-settings:defaultapps").Start()
	return nil
}

func unregister() error {
	// Delete leaves first, then parents
	leafFirst := []string{
		`Software\Classes\BrowserRouter\shell\open\command`,
		`Software\Classes\BrowserRouter\shell\open`,
		`Software\Classes\BrowserRouter\shell`,
		`Software\Classes\BrowserRouter\DefaultIcon`,
		`Software\Classes\BrowserRouter`,
		`Software\Clients\StartMenuInternet\BrowserRouter\Capabilities\URLAssociations`,
		`Software\Clients\StartMenuInternet\BrowserRouter\Capabilities`,
		`Software\Clients\StartMenuInternet\BrowserRouter`,
	}

	for _, path := range leafFirst {
		registry.DeleteKey(registry.CURRENT_USER, path)
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\RegisteredApplications`, registry.ALL_ACCESS)
	if err == nil {
		k.DeleteValue("BrowserRouter")
		k.Close()
	}

	fmt.Println("Unregistered successfully.")
	return nil
}
