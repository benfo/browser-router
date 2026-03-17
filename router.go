package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// route returns the target browser and the rule pattern that matched.
// Rule is empty string when the default browser is used.
func route(rawURL string, cfg *Config) (BrowserTarget, string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return cfg.Default, ""
	}

	for _, rule := range cfg.Rules {
		if matchRule(rule.Match, u) {
			return BrowserTarget{Browser: rule.Browser, Profile: rule.Profile}, rule.Match
		}
	}

	return cfg.Default, ""
}

// matchRule supports four pattern types:
//
//	regex:<pattern>  — match full URL with a regular expression
//	host:port        — exact host+port match  (e.g. "localhost:3000")
//	*.example.com    — wildcard subdomain match
//	example.com      — exact hostname match
func matchRule(pattern string, u *url.URL) bool {
	if strings.HasPrefix(pattern, "regex:") {
		re, err := regexp.Compile(strings.TrimPrefix(pattern, "regex:"))
		if err != nil {
			return false
		}
		return re.MatchString(u.String())
	}

	if strings.Contains(pattern, ":") {
		return u.Host == pattern
	}

	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // ".example.com"
		host := u.Hostname()
		return host == pattern[2:] || strings.HasSuffix(host, suffix)
	}

	return u.Hostname() == pattern
}

func testURL(rawURL string, cfg *Config) {
	target, rule := route(rawURL, cfg)
	def, ok := cfg.Browsers[target.Browser]

	fmt.Printf("URL:     %s\n", rawURL)
	fmt.Printf("Browser: %s\n", target.Browser)
	if target.Profile != "" {
		fmt.Printf("Profile: %s\n", target.Profile)
	}
	if rule != "" {
		fmt.Printf("Rule:    %s\n", rule)
	} else {
		fmt.Printf("Rule:    (none — using default)\n")
	}
	if ok {
		fmt.Printf("Exec:    %s\n", def.Executable())
	} else {
		fmt.Printf("Exec:    (browser %q not found in config)\n", target.Browser)
	}
}
