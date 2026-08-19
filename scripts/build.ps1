$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
$outputDirectory = Join-Path $projectRoot "dist"
$outputFile = Join-Path $outputDirectory "blind-dev-setup-windows-x64.exe"

Set-Location $projectRoot
go test ./...
go vet ./...
New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -trimpath -ldflags "-s -w" -o $outputFile ./cmd/blind-dev-setup
Get-FileHash -Algorithm SHA256 $outputFile
