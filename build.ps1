# Build browser-router without a console window (no flash when launched by OS)
go build -ldflags="-H windowsgui" -o browser-router.exe .
Write-Host "Built: browser-router.exe"
