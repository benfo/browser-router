//go:build !handler

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func main() {
	root := &cobra.Command{
		Use:   "browser-router",
		Short: "Route URLs to different browsers based on configurable rules",
	}

	root.AddCommand(
		openCmd(),
		registerCmd(),
		unregisterCmd(),
		rulesCmd(),
		browsersCmd(),
		configCmd(),
		monitorCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// openCmd is called by the OS when a URL is opened.
// The registry entry sets the command to: browser-router.exe open "%1"
func openCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <url>",
		Short: "Open a URL using routing rules (called by the OS as default browser)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			target, rule := route(args[0], cfg)
			broadcastToMonitor(monitorEvent{
				Time:    time.Now(),
				URL:     args[0],
				Browser: target.Browser,
				Profile: target.Profile,
				Rule:    rule,
			})
			return launch(target, args[0], cfg)
		},
	}
}

func registerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "register",
		Short: "Register as the system default browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			return register()
		},
	}
}

func unregisterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unregister",
		Short: "Remove browser-router as the system default browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			return unregister()
		},
	}
}

// ── rules ─────────────────────────────────────────────────────────────────────

func rulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage routing rules",
	}
	cmd.AddCommand(rulesListCmd(), rulesAddCmd(), rulesRemoveCmd(), rulesTestCmd())
	return cmd
}

func rulesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all routing rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			if len(cfg.Rules) == 0 {
				fmt.Println("No rules configured. All URLs go to the default browser.")
				return nil
			}
			fmt.Printf("%-4s  %-35s  %-12s  %s\n", "#", "MATCH", "BROWSER", "PROFILE")
			fmt.Printf("%-4s  %-35s  %-12s  %s\n", "---", "-----", "-------", "-------")
			for i, r := range cfg.Rules {
				fmt.Printf("%-4d  %-35s  %-12s  %s\n", i+1, r.Match, r.Browser, r.Profile)
			}
			fmt.Printf("\nDefault: %s", cfg.Default.Browser)
			if cfg.Default.Profile != "" {
				fmt.Printf(" (profile: %s)", cfg.Default.Profile)
			}
			fmt.Println()
			return nil
		},
	}
}

func rulesAddCmd() *cobra.Command {
	var match, browser, profile, description string
	var position int

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a routing rule",
		Example: `  browser-router rules add --match "*.google.com" --browser chrome
  browser-router rules add --match "outlook.com" --browser chrome --profile "Work"
  browser-router rules add --match "localhost:3000" --browser edge
  browser-router rules add --match "regex:github\.com/myorg" --browser firefox`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if match == "" || browser == "" {
				return fmt.Errorf("--match and --browser are required")
			}

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			for _, r := range cfg.Rules {
				if r.Match == match && r.Browser == browser && r.Profile == profile {
					fmt.Println("Rule already exists.")
					return nil
				}
			}

			rule := Rule{
				Description: description,
				Match:       match,
				Browser:     browser,
				Profile:     profile,
			}

			if position > 0 && position <= len(cfg.Rules)+1 {
				idx := position - 1
				cfg.Rules = append(cfg.Rules[:idx], append([]Rule{rule}, cfg.Rules[idx:]...)...)
			} else {
				cfg.Rules = append(cfg.Rules, rule)
			}

			if err := saveConfig(cfg); err != nil {
				return err
			}

			fmt.Printf("Added: %s → %s", match, browser)
			if profile != "" {
				fmt.Printf(" (profile: %s)", profile)
			}
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVar(&match, "match", "", "URL pattern to match (required)")
	cmd.Flags().StringVar(&browser, "browser", "", "Browser name to use (required)")
	cmd.Flags().StringVar(&profile, "profile", "", "Browser profile name")
	cmd.Flags().StringVar(&description, "desc", "", "Human-readable description")
	cmd.Flags().IntVar(&position, "position", 0, "Insert at position (1-based, default: append)")
	return cmd
}

func rulesRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <index>",
		Short: "Remove a rule by its index number (see: rules list)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idx, err := strconv.Atoi(args[0])
			if err != nil || idx < 1 {
				return fmt.Errorf("index must be a positive integer")
			}

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			if idx > len(cfg.Rules) {
				return fmt.Errorf("index %d out of range (have %d rules)", idx, len(cfg.Rules))
			}

			removed := cfg.Rules[idx-1]
			cfg.Rules = append(cfg.Rules[:idx-1], cfg.Rules[idx:]...)

			if err := saveConfig(cfg); err != nil {
				return err
			}

			fmt.Printf("Removed rule #%d: %s → %s\n", idx, removed.Match, removed.Browser)
			return nil
		},
	}
}

func rulesTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "test <url>",
		Short:   "Show which browser a URL would open in without actually opening it",
		Args:    cobra.ExactArgs(1),
		Example: `  browser-router rules test https://mail.google.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			testURL(args[0], cfg)
			return nil
		},
	}
}

// ── monitor ───────────────────────────────────────────────────────────────────

func monitorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "monitor",
		Short: "Live view of URLs being opened, with interactive rule creation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMonitor()
		},
	}
}

// ── browsers ──────────────────────────────────────────────────────────────────

func browsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "browsers",
		Short: "Manage browser definitions",
	}
	cmd.AddCommand(browsersListCmd(), browsersSetDefaultCmd(), browsersProfilesCmd(), browsersDetectCmd())
	return cmd
}

func browsersListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured browsers",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			for name, def := range cfg.Browsers {
				marker := ""
				if name == cfg.Default.Browser {
					marker = "  [default]"
				}
				fmt.Printf("%s%s\n", name, marker)
				fmt.Printf("  %s\n", def.Executable())
			}
			return nil
		},
	}
}


func browsersSetDefaultCmd() *cobra.Command {
	var profile string

	cmd := &cobra.Command{
		Use:   "set-default <browser>",
		Short: "Set the fallback browser used when no rule matches",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			cfg.Default = BrowserTarget{Browser: args[0], Profile: profile}
			if err := saveConfig(cfg); err != nil {
				return err
			}
			fmt.Printf("Default set to %q\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&profile, "profile", "", "Default profile")
	return cmd
}

func browsersProfilesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "profiles <browser>",
		Short:   "List detected profiles for a browser",
		Example: `  browser-router browsers profiles chrome`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kb, ok := findKnownBrowser(args[0])
			if !ok {
				return fmt.Errorf("unknown browser %q — supported: chrome, edge, brave, vivaldi, firefox", args[0])
			}

			profiles, err := kb.profiles()
			if err != nil {
				return fmt.Errorf("could not read profiles: %w", err)
			}

			fmt.Printf("%s profiles:\n", kb.Label)
			for _, p := range profiles {
				fmt.Printf("  %-20s → %s\n", p.DirName, p.DisplayName)
			}
			return nil
		},
	}
}

func browsersDetectCmd() *cobra.Command {
	var addNames []string

	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Scan for installed browsers and their profiles",
		Example: `  browser-router browsers detect
  browser-router browsers detect --add chrome
  browser-router browsers detect --add chrome,edge
  browser-router browsers detect --add all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			addSet := map[string]bool{}
			addAll := false
			for _, n := range addNames {
				if n == "all" {
					addAll = true
				} else {
					addSet[strings.ToLower(n)] = true
				}
			}

			anyFound := false
			changed := false

			for _, kb := range knownBrowsers {
				if !kb.isInstalled() {
					continue
				}
				anyFound = true

				_, inConfig := cfg.Browsers[kb.ConfigKey]
				status := ""
				if !inConfig {
					status = "  [not in config]"
				}
				fmt.Printf("%s%s\n", kb.Label, status)
				fmt.Printf("  exe: %s\n", kb.exe())

				profiles, err := kb.profiles()
				if err != nil {
					fmt.Printf("  profiles: (could not read — %v)\n", err)
				} else {
					fmt.Printf("  profiles:\n")
					for _, p := range profiles {
						fmt.Printf("    %-20s → %s\n", p.DirName, p.DisplayName)
						fmt.Printf("      add rule: browser-router rules add --match \"<pattern>\" --browser %s --profile \"%s\"\n",
							kb.ConfigKey, p.DirName)
					}
				}

				shouldAdd := !inConfig && (addAll || addSet[kb.ConfigKey])
				if shouldAdd {
					cfg.Browsers[kb.ConfigKey] = BrowserDef{
						Windows: kb.ExeWin,
						Darwin:  kb.ExeMac,
						Linux:   kb.ExeLin,
					}
					fmt.Printf("  → added to config\n")
					changed = true
				}
				fmt.Println()
			}

			if !anyFound {
				fmt.Println("No known browsers detected.")
				return nil
			}

			if changed {
				if err := saveConfig(cfg); err != nil {
					return err
				}
				fmt.Println("Config saved.")
			} else if len(addNames) == 0 {
				fmt.Println("Tip: use --add <name> to add a browser, or --add all for all missing ones.")
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&addNames, "add", nil, "Browser(s) to add to config: a name, comma-separated names, or 'all'")
	return cmd
}

// ── config ────────────────────────────────────────────────────────────────────

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Config file management",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "show",
			Short: "Print the current config as JSON",
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, err := loadConfig()
				if err != nil {
					return err
				}
				data, _ := yaml.Marshal(cfg)
				fmt.Print(string(data))
				return nil
			},
		},
		&cobra.Command{
			Use:   "path",
			Short: "Print the config file location",
			RunE: func(cmd *cobra.Command, args []string) error {
				path, err := configPath()
				if err != nil {
					return err
				}
				fmt.Println(path)
				return nil
			},
		},
		&cobra.Command{
			Use:   "edit",
			Short: "Open the config file in your default editor",
			RunE: func(cmd *cobra.Command, args []string) error {
				path, err := configPath()
				if err != nil {
					return err
				}
				// Ensure the file exists before opening
				if _, err := os.Stat(path); os.IsNotExist(err) {
					cfg := defaultConfig()
					if err := saveConfig(cfg); err != nil {
						return err
					}
				}
				return openInEditor(path)
			},
		},
	)

	return cmd
}
