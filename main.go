package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
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
			target := route(args[0], cfg)
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

// ── browsers ──────────────────────────────────────────────────────────────────

func browsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "browsers",
		Short: "Manage browser definitions",
	}
	cmd.AddCommand(browsersListCmd(), browsersAddCmd(), browsersSetDefaultCmd())
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

func browsersAddCmd() *cobra.Command {
	var name, windows, darwin, linux string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add or update a browser definition",
		Example: `  browser-router browsers add --name brave --windows "C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			cfg.Browsers[name] = BrowserDef{
				Windows: windows,
				Darwin:  darwin,
				Linux:   linux,
			}

			if err := saveConfig(cfg); err != nil {
				return err
			}

			fmt.Printf("Browser %q saved.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Browser identifier (required)")
	cmd.Flags().StringVar(&windows, "windows", "", "Executable path on Windows")
	cmd.Flags().StringVar(&darwin, "mac", "", "Executable path on macOS")
	cmd.Flags().StringVar(&linux, "linux", "", "Executable path on Linux")
	return cmd
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
				data, _ := json.MarshalIndent(cfg, "", "  ")
				fmt.Println(string(data))
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
