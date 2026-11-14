$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
$outputDirectory = Join-Path $projectRoot "dist"
$binaryName = "blind-dev-setup-windows-x64.exe"
$archiveName = "blind-dev-setup-windows-x64.zip"
$checksumsName = "SHA256SUMS.txt"
$outputFile = Join-Path $outputDirectory $binaryName
$archiveFile = Join-Path $outputDirectory $archiveName
$checksumsFile = Join-Path $outputDirectory $checksumsName
$packageDirectory = Join-Path $outputDirectory "package"
$packagingDirectory = Join-Path $projectRoot "packaging\windows"

Set-Location $projectRoot
go test ./...
if ($LASTEXITCODE -ne 0) {
    throw "Testes falharam. O pacote nao foi criado."
}
go vet ./...
if ($LASTEXITCODE -ne 0) {
    throw "Analise estatica falhou. O pacote nao foi criado."
}

New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null
foreach ($path in @($outputFile, $archiveFile, $checksumsFile)) {
    if (Test-Path $path) {
        Remove-Item -Force $path
    }
}
if (Test-Path $packageDirectory) {
    Remove-Item -Recurse -Force $packageDirectory
}

$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -trimpath -ldflags "-s -w" -o $outputFile ./cmd/blind-dev-setup
if ($LASTEXITCODE -ne 0) {
    throw "Compilacao falhou. O pacote nao foi criado."
}

New-Item -ItemType Directory -Force -Path $packageDirectory | Out-Null
Copy-Item $outputFile (Join-Path $packageDirectory $binaryName)
Copy-Item (Join-Path $packagingDirectory "INICIAR-AQUI.cmd") $packageDirectory
Copy-Item (Join-Path $packagingDirectory "LEIA-ME.txt") $packageDirectory
Copy-Item (Join-Path $projectRoot "LICENSE") $packageDirectory
# Keep the manual test guide's relative links usable before preparing a drive.
Copy-Item (Join-Path $projectRoot "internal\templates\files\docs\*.md") $packageDirectory
$sourceCommit = (& git rev-parse HEAD)
if ($LASTEXITCODE -ne 0) { throw "Nao foi possivel identificar o commit do pacote." }
$sourceStatus = (& git status --porcelain)
if ($LASTEXITCODE -ne 0) { throw "Nao foi possivel verificar o estado do codigo." }
$sourceState = if ($sourceStatus) { "modified" } else { "clean" }
$bundle = Get-Content -LiteralPath (Join-Path $projectRoot "internal\bundle\windows-x64.json") -Raw | ConvertFrom-Json
@(
    "blind-dev-setup $($bundle.release)",
    "Commit: $($sourceCommit.Trim())",
    "Source state: $sourceState",
    "Platform: windows-x64",
    "Compare SHA256SUMS.txt before testing. Record this commit in RESULTADOS-VALIDACAO.md."
) | Set-Content -LiteralPath (Join-Path $packageDirectory "BUILD-INFO.txt") -Encoding ascii

Compress-Archive -Path (Join-Path $packageDirectory "*") -DestinationPath $archiveFile -CompressionLevel Optimal

$checksumLines = foreach ($artifact in @($outputFile, $archiveFile)) {
    $hash = Get-FileHash -Algorithm SHA256 $artifact
    "$($hash.Hash.ToLowerInvariant())  $(Split-Path -Leaf $artifact)"
}
$checksumLines | Set-Content -Encoding ascii $checksumsFile

Remove-Item -Recurse -Force $packageDirectory
Get-FileHash -Algorithm SHA256 $outputFile, $archiveFile
