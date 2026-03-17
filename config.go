package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	Default  BrowserTarget         `json:"default"`
	Rules    []Rule                `json:"rules"`
	Browsers map[string]BrowserDef `json:"browsers"`
}

type Rule struct {
	Description string `json:"description,omitempty"`
	Match       string `json:"match"`
	Browser     string `json:"browser"`
	Profile     string `json:"profile,omitempty"`
}

type BrowserTarget struct {
	Browser string `json:"browser"`
	Profile string `json:"profile,omitempty"`
}

type BrowserDef struct {
	Windows string   `json:"windows"`
	Darwin  string   `json:"darwin"`
	Linux   string   `json:"linux"`
	Args    []string `json:"args,omitempty"`
}

func (b BrowserDef) Executable() string {
	switch runtime.GOOS {
	case "windows":
		return b.Windows
	case "darwin":
		return b.Darwin
	default:
		return b.Linux
	}
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "browser-router", "config.json"), nil
}

func loadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaultConfig(), nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func defaultConfig() *Config {
	return &Config{
		Default: BrowserTarget{Browser: "chrome"},
		Rules:   []Rule{},
		Browsers: map[string]BrowserDef{
			"chrome": {
				Windows: `C:\Program Files\Google\Chrome\Application\chrome.exe`,
				Darwin:  `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`,
				Linux:   `google-chrome`,
			},
			"edge": {
				Windows: `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
				Darwin:  `/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge`,
				Linux:   `microsoft-edge`,
			},
			"firefox": {
				Windows: `C:\Program Files\Mozilla Firefox\firefox.exe`,
				Darwin:  `/Applications/Firefox.app/Contents/MacOS/firefox`,
				Linux:   `firefox`,
			},
		},
	}
}
