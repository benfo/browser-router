# Usage:
#   .\build.ps1           — build into current directory
#   .\build.ps1 install   — build and copy to $env:GOPATH\bin (replaces go install)

param([string]$target = "")

$outDir = if ($target -eq "install") {
    if ($env:GOBIN) { $env:GOBIN }
    elseif ($env:GOPATH) { Join-Path $env:GOPATH "bin" }
    else { Join-Path $env:USERPROFILE "go\bin" }
} else {
    "."
}

# windowsgui handler — registered as the OS default browser, no console flash
go build -tags handler -ldflags="-H windowsgui" -o "$outDir\browser-router-open.exe" .

# CLI tool — console app, all interactive commands
go build -o "$outDir\browser-router.exe" .

Write-Host "Built: $outDir\browser-router-open.exe + $outDir\browser-router.exe"
