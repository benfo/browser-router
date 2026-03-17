package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Default  BrowserTarget         `toml:"default"`
	Rules    []Rule                `toml:"rules"`
	Browsers map[string]BrowserDef `toml:"browsers"`
}

type Rule struct {
	Description string `toml:"description,omitempty"`
	Match       string `toml:"match"`
	Browser     string `toml:"browser"`
	Profile     string `toml:"profile,omitempty"`
}

type BrowserTarget struct {
	Browser string `toml:"browser"`
	Profile string `toml:"profile,omitempty"`
}

type BrowserDef struct {
	Windows string   `toml:"windows"`
	Darwin  string   `toml:"darwin"`
	Linux   string   `toml:"linux"`
	Args    []string `toml:"args,omitempty"`
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
	return filepath.Join(dir, "browser-router", "config.toml"), nil
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
	if err := toml.Unmarshal(data, &cfg); err != nil {
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

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
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
