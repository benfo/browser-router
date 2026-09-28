# browser-router

Set it as your default browser and it opens each URL in the browser/profile you choose by rule.

## Install

Windows:

```powershell
irm https://raw.githubusercontent.com/benfo/browser-router/main/install.ps1 | iex
```

macOS / Linux:

```sh
go install github.com/benfo/browser-router@latest
```

Or grab an archive from [Releases](https://github.com/benfo/browser-router/releases).

## Setup

```sh
browser-router browsers detect --add all   # add installed browsers
browser-router register                    # make it the default browser
```

On Windows, then pick "Browser Router" in Settings > Default Apps.

## Rules

```sh
browser-router rules add --match "*.google.com" --browser chrome
browser-router rules add --match "outlook.com" --browser chrome --profile "Work"
browser-router rules add --match "regex:github\.com/myorg" --browser firefox
browser-router rules list
browser-router rules test https://mail.google.com
```

Rules are checked top to bottom, and the first match wins. Anything that doesn't match opens in the default browser (`browsers set-default`).

`browser-router monitor` shows URLs as they're routed. `browser-router config edit` opens the config file.

## Release

```sh
git tag v0.1.0 && git push origin v0.1.0
```

## Roadmap

- [ ] Resolve rule profiles by display name (e.g. `profile: Work`) via Chrome's `Local State`, since folder names like `Profile 1` differ per machine. Exact folder-name match wins.
- [ ] Config path override (`BROWSER_ROUTER_CONFIG` or `config path --set`) so the config can live in a synced folder.
- [ ] UI for rule management (Wails, Fyne, Tauri or embedded webview), once the CLI is stable.
- [ ] System tray icon.
