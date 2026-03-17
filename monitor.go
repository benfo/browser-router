package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const monitorPort = 7717

type monitorEvent struct {
	Time    time.Time `json:"time"`
	URL     string    `json:"url"`
	Browser string    `json:"browser"`
	Profile string    `json:"profile"`
	Rule    string    `json:"rule"` // empty = no rule matched, default browser used
}

// broadcastToMonitor sends a URL open event to any running monitor session.
// Fire-and-forget over UDP — if no monitor is listening this is a no-op.
func broadcastToMonitor(event monitorEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: monitorPort,
	})
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(50 * time.Millisecond))
	conn.Write(data)
}

// runMonitor starts the live monitor. It listens for UDP events from open
// and lets the user interactively create rules from captured URLs.
func runMonitor() error {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: monitorPort,
	})
	if err != nil {
		return fmt.Errorf("could not start monitor (is another instance running?): %w", err)
	}
	defer conn.Close()

	eventCh := make(chan monitorEvent, 20)
	inputCh := make(chan string, 5)

	// Receive UDP packets
	go func() {
		buf := make([]byte, 8192)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			var event monitorEvent
			if json.Unmarshal(buf[:n], &event) == nil {
				eventCh <- event
			}
		}
	}()

	// Read stdin lines
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputCh <- scanner.Text()
		}
	}()

	var events []monitorEvent

	fmt.Println("Browser Router — live monitor")
	fmt.Println("Waiting for URLs... (open a link anywhere to see it here)")
	fmt.Println("Type a URL number + Enter to create a rule from it. Ctrl+C to stop.")
	fmt.Println(strings.Repeat("─", 60))

	for {
		select {
		case event := <-eventCh:
			events = append(events, event)
			printEvent(len(events), event)

		case line := <-inputCh:
			line = strings.TrimSpace(strings.TrimLeft(line, "rR "))
			idx, err := strconv.Atoi(line)
			if err != nil || idx < 1 || idx > len(events) {
				if line != "" {
					fmt.Printf("Enter a number between 1 and %d\n", len(events))
				}
				continue
			}
			createRuleInteractive(events[idx-1], inputCh)
		}
	}
}

func printEvent(idx int, e monitorEvent) {
	timeStr := e.Time.Local().Format("15:04:05")

	target := e.Browser
	if e.Profile != "" {
		target = e.Browser + " / " + e.Profile
	}

	ruleStr := "(no rule — default)"
	if e.Rule != "" {
		ruleStr = "matched: " + e.Rule
	}

	fmt.Printf("\n#%-3d %s  %s\n", idx, timeStr, e.URL)
	fmt.Printf("          → %-25s %s\n", target, ruleStr)
}

func createRuleInteractive(e monitorEvent, inputCh <-chan string) {
	u, err := url.Parse(e.URL)
	if err != nil {
		fmt.Println("Could not parse URL.")
		return
	}

	suggestions := suggestPatterns(u)

	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Create rule for: %s\n\n", e.URL)
	fmt.Println("Suggested patterns:")
	for i, s := range suggestions {
		fmt.Printf("  %d)  %s\n", i+1, s)
	}
	fmt.Println()

	prompt := func(msg string) string {
		fmt.Print(msg)
		return strings.TrimSpace(<-inputCh)
	}

	// Pattern
	patternInput := prompt(fmt.Sprintf("Choose [1-%d] or enter custom pattern (blank to cancel): ", len(suggestions)))
	if patternInput == "" {
		fmt.Println("Cancelled.")
		fmt.Println(strings.Repeat("─", 60))
		return
	}
	pattern := patternInput
	if idx, err := strconv.Atoi(patternInput); err == nil && idx >= 1 && idx <= len(suggestions) {
		pattern = suggestions[idx-1]
	}

	// Browser
	browserInput := prompt(fmt.Sprintf("Browser [%s]: ", e.Browser))
	if browserInput == "" {
		browserInput = e.Browser
	}

	// Profile
	profileInput := prompt("Profile (blank for none): ")

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	for _, r := range cfg.Rules {
		if r.Match == pattern && r.Browser == browserInput && r.Profile == profileInput {
			fmt.Println("Rule already exists.")
			fmt.Println(strings.Repeat("─", 60))
			return
		}
	}

	// Prepend so the new rule takes priority over existing ones
	rule := Rule{Match: pattern, Browser: browserInput, Profile: profileInput}
	cfg.Rules = append([]Rule{rule}, cfg.Rules...)

	if err := saveConfig(cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}

	fmt.Printf("\nRule added: %s → %s", pattern, browserInput)
	if profileInput != "" {
		fmt.Printf(" / %s", profileInput)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
}

// suggestPatterns generates candidate rule patterns from a parsed URL.
func suggestPatterns(u *url.URL) []string {
	var out []string
	host := u.Hostname()
	parts := strings.Split(host, ".")

	// host:port — most specific, suggest first for localhost-style URLs
	if u.Port() != "" {
		out = append(out, u.Host)
	}

	// exact host
	out = append(out, host)

	// wildcard domain — only when there's a meaningful subdomain
	// (skip for bare domains like "localhost" or IPs)
	if len(parts) > 2 && !isIP(host) {
		out = append(out, "*."+strings.Join(parts[1:], "."))
	}

	// host + first path segment (when path is meaningful)
	path := strings.Trim(u.Path, "/")
	if path != "" {
		segments := strings.SplitN(path, "/", 2)
		if segments[0] != "" {
			out = append(out, host+"/"+segments[0])
		}
	}

	return out
}

func isIP(host string) bool {
	return net.ParseIP(host) != nil
}
