package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func launch(target BrowserTarget, rawURL string, cfg *Config) error {
	def, ok := cfg.Browsers[target.Browser]
	if !ok {
		return fmt.Errorf("unknown browser %q — add it with: browser-router browsers add", target.Browser)
	}

	exe := def.Executable()
	if exe == "" {
		return fmt.Errorf("no executable configured for browser %q on %s", target.Browser, runtime.GOOS)
	}

	args := buildArgs(target, rawURL, def)
	cmd := exec.Command(exe, args...)
	return cmd.Start()
}

func buildArgs(target BrowserTarget, rawURL string, def BrowserDef) []string {
	args := append([]string{}, def.Args...)

	if target.Profile != "" {
		switch target.Browser {
		case "chrome", "edge", "brave", "vivaldi":
			args = append(args, "--profile-directory="+target.Profile)
		case "firefox":
			args = append(args, "-P", target.Profile)
		}
	}

	args = append(args, rawURL)
	return args
}
