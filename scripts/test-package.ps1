#Requires -Version 7.0
[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [string]$PackageDirectory,
    [Parameter(Mandatory)]
    [ValidatePattern('^[0-9a-f]{40}$')]
    [string]$ExpectedCommit
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$binaryName = "blind-dev-setup-windows-x64.exe"
$archiveName = "blind-dev-setup-windows-x64.zip"
$checksums = @{}
foreach ($line in Get-Content -LiteralPath (Join-Path $PackageDirectory "SHA256SUMS.txt")) {
    if ($line -notmatch '^([0-9a-f]{64})  (blind-dev-setup-windows-x64\.(exe|zip))$') {
        throw "Unexpected checksum entry: $line"
    }
    $fileName = $Matches[2]
    if ($checksums.ContainsKey($fileName)) {
        throw "Duplicate checksum for $fileName."
    }
    $checksums[$fileName] = $Matches[1]
}
foreach ($fileName in @($binaryName, $archiveName)) {
    $actual = (Get-FileHash -LiteralPath (Join-Path $PackageDirectory $fileName) -Algorithm SHA256).Hash
    if (-not $checksums.ContainsKey($fileName) -or $actual -ne $checksums[$fileName]) {
        throw "Checksum mismatch or missing checksum for $fileName."
    }
}

$unpacked = Join-Path ([System.IO.Path]::GetTempPath()) ("blind-dev-package-" + [guid]::NewGuid().ToString("N"))
try {
    Expand-Archive -LiteralPath (Join-Path $PackageDirectory $archiveName) -DestinationPath $unpacked
    foreach ($fileName in @($binaryName, "BUILD-INFO.txt", "TESTE-WINDOWS.md", "RESULTADOS-VALIDACAO.md", "INICIAR-AQUI.cmd", "LEIA-ME.txt", "LICENSE")) {
        $path = Join-Path $unpacked $fileName
        if (-not (Test-Path -LiteralPath $path -PathType Leaf) -or (Get-Item -LiteralPath $path).Length -eq 0) {
            throw "Package is missing a nonempty $fileName."
        }
    }
    $packagedHash = (Get-FileHash -LiteralPath (Join-Path $unpacked $binaryName) -Algorithm SHA256).Hash
    if ($packagedHash -ne $checksums[$binaryName]) {
        throw "ZIP executable differs from the executable tested by CI."
    }
    $buildInfo = Get-Content -LiteralPath (Join-Path $unpacked "BUILD-INFO.txt")
    $manifest = Get-Content -LiteralPath (Join-Path (Split-Path -Parent $PSScriptRoot) "internal/bundle/windows-x64.json") -Raw | ConvertFrom-Json
    foreach ($expected in @("Commit: $ExpectedCommit", "Source state: clean", "Platform: windows-x64", "blind-dev-setup $($manifest.release)")) {
        if ($buildInfo -cnotcontains $expected) {
            throw "BUILD-INFO.txt does not contain the expected value: $expected"
        }
    }
    Write-Host "ZIP, executable and checksums match clean source commit $ExpectedCommit."
}
finally {
    if (Test-Path -LiteralPath $unpacked) {
        Remove-Item -LiteralPath $unpacked -Recurse -Force
    }
}
