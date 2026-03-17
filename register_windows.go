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

	openCmd := fmt.Sprintf(`"%s" open "%%1"`, exe)

	// String values: path → map of name→data
	stringKeys := map[string]map[string]string{
		// ProgID — what the URL associations point to
		`Software\Classes\BrowserRouter`: {
			"":            "Browser Router",
			"URL Protocol": "",
		},
		`Software\Classes\BrowserRouter\DefaultIcon`: {
			"": exe + ",0",
		},
		`Software\Classes\BrowserRouter\shell\open\command`: {
			"": openCmd,
		},

		// StartMenuInternet registration — required for Default Apps UI
		`Software\Clients\StartMenuInternet\BrowserRouter`: {
			"": "Browser Router",
		},
		`Software\Clients\StartMenuInternet\BrowserRouter\DefaultIcon`: {
			"": exe + ",0",
		},
		`Software\Clients\StartMenuInternet\BrowserRouter\shell\open\command`: {
			"": openCmd,
		},
		`Software\Clients\StartMenuInternet\BrowserRouter\Capabilities`: {
			"ApplicationName":        "Browser Router",
			"ApplicationDescription": "Routes URLs to different browsers based on rules",
			"ApplicationIcon":        exe + ",0",
		},
		`Software\Clients\StartMenuInternet\BrowserRouter\Capabilities\URLAssociations`: {
			"http":  "BrowserRouter",
			"https": "BrowserRouter",
		},
		`Software\Clients\StartMenuInternet\BrowserRouter\Capabilities\FileAssociations`: {
			".htm":  "BrowserRouter",
			".html": "BrowserRouter",
		},

		// RegisteredApplications — links name to Capabilities
		`Software\RegisteredApplications`: {
			"BrowserRouter": `Software\Clients\StartMenuInternet\BrowserRouter\Capabilities`,
		},
	}

	for path, values := range stringKeys {
		k, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.ALL_ACCESS)
		if err != nil {
			return fmt.Errorf("create key %s: %w", path, err)
		}
		for name, data := range values {
			k.SetStringValue(name, data)
		}
		k.Close()
	}

	// InstallInfo requires a DWORD, so handle separately
	installInfo, _, err := registry.CreateKey(registry.CURRENT_USER,
		`Software\Clients\StartMenuInternet\BrowserRouter\InstallInfo`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("create InstallInfo key: %w", err)
	}
	installInfo.SetDWordValue("IconsVisible", 1)
	installInfo.Close()

	fmt.Println("Registered successfully.")
	fmt.Printf("Executable: %s\n\n", exe)
	fmt.Println("Opening Windows Settings > Default Apps.")
	fmt.Println("Search for 'Browser Router' and set it as your default browser.")
	exec.Command("cmd", "/c", "start", "ms-settings:defaultapps").Start()
	return nil
}

func unregister() error {
	leafFirst := []string{
		`Software\Classes\BrowserRouter\shell\open\command`,
		`Software\Classes\BrowserRouter\shell\open`,
		`Software\Classes\BrowserRouter\shell`,
		`Software\Classes\BrowserRouter\DefaultIcon`,
		`Software\Classes\BrowserRouter`,
		`Software\Clients\StartMenuInternet\BrowserRouter\Capabilities\URLAssociations`,
		`Software\Clients\StartMenuInternet\BrowserRouter\Capabilities\FileAssociations`,
		`Software\Clients\StartMenuInternet\BrowserRouter\Capabilities`,
		`Software\Clients\StartMenuInternet\BrowserRouter\InstallInfo`,
		`Software\Clients\StartMenuInternet\BrowserRouter\DefaultIcon`,
		`Software\Clients\StartMenuInternet\BrowserRouter\shell\open\command`,
		`Software\Clients\StartMenuInternet\BrowserRouter\shell\open`,
		`Software\Clients\StartMenuInternet\BrowserRouter\shell`,
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
