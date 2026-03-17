package main

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type BrowserProfile struct {
	DirName     string // passed to --profile-directory or -P
	DisplayName string // human-readable label
}

type knownBrowser struct {
	ConfigKey   string // matches key in config Browsers map
	Label       string
	ProfileType string // "chromium" or "firefox"
	ExeWin      string
	ExeMac      string
	ExeLin      string
	DataWin     string // may use %ENV% vars
	DataMac     string // may use ~/ prefix
	DataLin     string
}

// knownBrowsers is the catalog of browsers we can auto-detect.
var knownBrowsers = []knownBrowser{
	{
		ConfigKey:   "chrome",
		Label:       "Google Chrome",
		ProfileType: "chromium",
		ExeWin:      `C:\Program Files\Google\Chrome\Application\chrome.exe`,
		ExeMac:      `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`,
		ExeLin:      `google-chrome`,
		DataWin:     `$LOCALAPPDATA\Google\Chrome\User Data`,
		DataMac:     `~/Library/Application Support/Google/Chrome`,
		DataLin:     `~/.config/google-chrome`,
	},
	{
		ConfigKey:   "edge",
		Label:       "Microsoft Edge",
		ProfileType: "chromium",
		ExeWin:      `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		ExeMac:      `/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge`,
		ExeLin:      `microsoft-edge`,
		DataWin:     `$LOCALAPPDATA\Microsoft\Edge\User Data`,
		DataMac:     `~/Library/Application Support/Microsoft Edge`,
		DataLin:     `~/.config/microsoft-edge`,
	},
	{
		ConfigKey:   "brave",
		Label:       "Brave",
		ProfileType: "chromium",
		ExeWin:      `C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`,
		ExeMac:      `/Applications/Brave Browser.app/Contents/MacOS/Brave Browser`,
		ExeLin:      `brave-browser`,
		DataWin:     `$LOCALAPPDATA\BraveSoftware\Brave-Browser\User Data`,
		DataMac:     `~/Library/Application Support/BraveSoftware/Brave-Browser`,
		DataLin:     `~/.config/BraveSoftware/Brave-Browser`,
	},
	{
		ConfigKey:   "vivaldi",
		Label:       "Vivaldi",
		ProfileType: "chromium",
		ExeWin:      `C:\Program Files\Vivaldi\Application\vivaldi.exe`,
		ExeMac:      `/Applications/Vivaldi.app/Contents/MacOS/Vivaldi`,
		ExeLin:      `vivaldi`,
		DataWin:     `$LOCALAPPDATA\Vivaldi\User Data`,
		DataMac:     `~/Library/Application Support/Vivaldi`,
		DataLin:     `~/.config/vivaldi`,
	},
	{
		ConfigKey:   "firefox",
		Label:       "Firefox",
		ProfileType: "firefox",
		ExeWin:      `C:\Program Files\Mozilla Firefox\firefox.exe`,
		ExeMac:      `/Applications/Firefox.app/Contents/MacOS/firefox`,
		ExeLin:      `firefox`,
		DataWin:     `$APPDATA\Mozilla\Firefox`,
		DataMac:     `~/Library/Application Support/Firefox`,
		DataLin:     `~/.mozilla/firefox`,
	},
}

func (kb knownBrowser) exe() string {
	switch runtime.GOOS {
	case "windows":
		return kb.ExeWin
	case "darwin":
		return kb.ExeMac
	default:
		return kb.ExeLin
	}
}

func (kb knownBrowser) dataDir() string {
	var raw string
	switch runtime.GOOS {
	case "windows":
		raw = kb.DataWin
	case "darwin":
		raw = kb.DataMac
	default:
		raw = kb.DataLin
	}
	raw = os.ExpandEnv(raw)
	if strings.HasPrefix(raw, "~/") {
		home, _ := os.UserHomeDir()
		raw = filepath.Join(home, raw[2:])
	}
	return raw
}

func (kb knownBrowser) isInstalled() bool {
	if _, err := os.Stat(kb.exe()); err == nil {
		return true
	}
	// On Linux the exe may just be a name on PATH
	if runtime.GOOS == "linux" {
		_, err := exec.LookPath(kb.exe())
		return err == nil
	}
	return false
}

func (kb knownBrowser) profiles() ([]BrowserProfile, error) {
	if kb.ProfileType == "firefox" {
		return firefoxProfiles(kb.dataDir())
	}
	return chromiumProfiles(kb.dataDir())
}

// findKnownBrowser looks up by ConfigKey (case-insensitive).
func findKnownBrowser(name string) (knownBrowser, bool) {
	name = strings.ToLower(name)
	for _, kb := range knownBrowsers {
		if strings.ToLower(kb.ConfigKey) == name {
			return kb, true
		}
	}
	return knownBrowser{}, false
}

// ── Chromium profile reader ───────────────────────────────────────────────────

func chromiumProfiles(dataDir string) ([]BrowserProfile, error) {
	data, err := os.ReadFile(filepath.Join(dataDir, "Local State"))
	if err != nil {
		return nil, err
	}

	var state struct {
		Profile struct {
			InfoCache map[string]struct {
				Name string `json:"name"`
			} `json:"info_cache"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	var profiles []BrowserProfile
	for dirName, info := range state.Profile.InfoCache {
		profiles = append(profiles, BrowserProfile{
			DirName:     dirName,
			DisplayName: info.Name,
		})
	}

	// Default first, then alphabetical by directory name
	sort.Slice(profiles, func(i, j int) bool {
		if profiles[i].DirName == "Default" {
			return true
		}
		if profiles[j].DirName == "Default" {
			return false
		}
		return profiles[i].DirName < profiles[j].DirName
	})

	return profiles, nil
}

// ── Firefox profile reader ────────────────────────────────────────────────────

func firefoxProfiles(dataDir string) ([]BrowserProfile, error) {
	f, err := os.Open(filepath.Join(dataDir, "profiles.ini"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var profiles []BrowserProfile
	var name, path string

	flush := func() {
		if path != "" {
			profiles = append(profiles, BrowserProfile{
				DirName:     filepath.Base(path),
				DisplayName: name,
			})
		}
		name, path = "", ""
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[Profile") {
			flush()
		} else if v, ok := strings.CutPrefix(line, "Name="); ok {
			name = v
		} else if v, ok := strings.CutPrefix(line, "Path="); ok {
			path = v
		}
	}
	flush()

	return profiles, scanner.Err()
}
