# Installs the latest browser-router release on Windows.
#   irm https://raw.githubusercontent.com/benfo/browser-router/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "benfo/browser-router"
$installDir = Join-Path $env:LOCALAPPDATA "Programs\browser-router"
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }

$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$asset = $release.assets | Where-Object { $_.name -eq "browser-router_windows_$arch.zip" }
if (-not $asset) { throw "No windows/$arch build found in release $($release.tag_name)" }

$zip = Join-Path $env:TEMP $asset.name
Invoke-WebRequest $asset.browser_download_url -OutFile $zip
New-Item -ItemType Directory -Force $installDir | Out-Null
Expand-Archive $zip -DestinationPath $installDir -Force
Remove-Item $zip

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (($userPath -split ";") -notcontains $installDir) {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added $installDir to your user PATH (restart your terminal)"
}

Write-Host "Installed browser-router $($release.tag_name) to $installDir"
Write-Host "Run 'browser-router register' to set it as your default browser"
