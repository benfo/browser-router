//go:build handler

package main

import (
	"os"
	"time"
)

// main is the entry point for browser-router-open.exe — the windowsgui binary
// registered with Windows as the default browser handler.
// The registry command is: "browser-router-open.exe" "%1"
// It receives the URL as the sole argument, routes it, and exits silently.
func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	rawURL := os.Args[1]

	cfg, err := loadConfig()
	if err != nil {
		os.Exit(1)
	}

	target, rule := route(rawURL, cfg)
	broadcastToMonitor(monitorEvent{
		Time:    time.Now(),
		URL:     rawURL,
		Browser: target.Browser,
		Profile: target.Profile,
		Rule:    rule,
	})
	launch(target, rawURL, cfg) //nolint
}
