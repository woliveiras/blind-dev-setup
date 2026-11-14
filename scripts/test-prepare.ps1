#Requires -Version 7.0
#Requires -RunAsAdministrator
[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [string]$WorkDirectory,
    [switch]$Cleanup
)

$ErrorActionPreference = "Stop"
$PSNativeCommandUseErrorActionPreference = $false
Set-StrictMode -Version Latest

$projectRoot = Split-Path -Parent $PSScriptRoot
$WorkDirectory = [System.IO.Path]::GetFullPath($WorkDirectory)
$imagePath = Join-Path $WorkDirectory "prepare.vhdx"
$ownerFile = Join-Path $WorkDirectory ".blind-dev-setup-e2e"

function Remove-TestVolume {
    # A second invocation from the workflow also handles interrupted test steps.
    if (-not (Test-Path -LiteralPath $ownerFile -PathType Leaf)) {
        return
    }
    if (Test-Path -LiteralPath $imagePath) {
        $image = Get-DiskImage -ImagePath $imagePath
        if ($image.Attached) {
            Dismount-DiskImage -ImagePath $imagePath
        }
        Remove-Item -LiteralPath $imagePath -Force
    }
    Remove-Item -LiteralPath $WorkDirectory -Recurse -Force
    Write-Host "Temporary prepare volume and downloads removed."
}

function Invoke-Checked {
    param([string]$FilePath, [string[]]$ArgumentList)

    & $FilePath @ArgumentList | Out-Host
    if ($LASTEXITCODE -ne 0) {
        throw "$FilePath failed with exit code $LASTEXITCODE."
    }
}

if ($Cleanup) {
    Remove-TestVolume
    return
}

$executable = Join-Path $projectRoot "dist\blind-dev-setup-windows-x64.exe"
if (-not (Test-Path -LiteralPath $executable -PathType Leaf)) {
    throw "Run scripts/build.ps1 before the prepare integration test."
}
# Never reuse an existing directory or disk image.
New-Item -ItemType Directory -Path $WorkDirectory | Out-Null
Set-Content -LiteralPath $ownerFile -Value "Temporary prepare integration test" -Encoding ascii

try {
    # DiskPart creates a dynamic image without requiring the Hyper-V module.
    # 32 GiB is its virtual capacity; only written blocks consume host space.
    $diskpartScript = Join-Path $WorkDirectory "create-volume.txt"
    @"
create vdisk file="$imagePath" maximum=32768 type=expandable
exit
"@ | Set-Content -LiteralPath $diskpartScript -Encoding ascii
    Invoke-Checked -FilePath "diskpart.exe" -ArgumentList @("/s", $diskpartScript)

    # Resolve the disk from our exact image, never from a guessed disk number.
    $image = Mount-DiskImage -ImagePath $imagePath -PassThru
    $disks = @($image | Get-Disk)
    if ($disks.Count -ne 1) {
        throw "Expected exactly one disk backed by the temporary image."
    }
    $disk = $disks[0]
    if ($disk.IsBoot -or $disk.IsSystem -or $disk.PartitionStyle -ne "RAW") {
        throw "Refusing to initialize a system disk or a disk that is not empty."
    }
    $partition = $disk | Initialize-Disk -PartitionStyle GPT -PassThru |
        New-Partition -AssignDriveLetter -UseMaximumSize
    $volume = $partition | Format-Volume -FileSystem NTFS -NewFileSystemLabel "PREPARE-E2E" -Confirm:$false
    $manifest = Get-Content -LiteralPath (Join-Path $projectRoot "internal\bundle\windows-x64.json") -Raw |
        ConvertFrom-Json
    if ($volume.FileSystem -ne "NTFS" -or $volume.SizeRemaining -lt $manifest.minimum_free_bytes) {
        throw "Temporary volume does not meet prepare's NTFS/free-space requirements."
    }
    if (-not $partition.DriveLetter) {
        throw "Temporary volume has no drive letter."
    }
    $target = "$($partition.DriveLetter):\"
    if ($target.TrimEnd('\') -eq $env:SystemDrive) {
        throw "Temporary target must not be the system volume."
    }
    Write-Host "Testing prepare on $target ($($volume.SizeRemaining) bytes free)."
    $cache = Join-Path $WorkDirectory "downloads"
    Invoke-Checked -FilePath $executable -ArgumentList @("plan", "--target", $target)
    Invoke-Checked -FilePath $executable -ArgumentList @(
        "prepare", "--target", $target, "--cache", $cache, "--accept-licenses"
    )
    Invoke-Checked -FilePath $executable -ArgumentList @("verify", "--target", $target)

    $bundleRoot = Join-Path $target "blind-dev-setup"
    if (-not (Test-Path -LiteralPath (Join-Path $bundleRoot "READY.txt") -PathType Leaf)) {
        throw "prepare did not publish READY.txt."
    }
    $staging = @(Get-ChildItem -LiteralPath $target -Force -Filter ".blind-dev-setup-staging-*")
    if ($staging.Count -ne 0) {
        throw "prepare left staging directories behind."
    }

    # Exercise the published location after the staging directory was renamed.
    # Match DEV-SHELL's trust/reshim setup without opening an interactive shell.
    $smokeScript = Join-Path $WorkDirectory "check-toolchain.cmd"
    $commands = @(
        "@echo off",
        "setlocal",
        "call `"$bundleRoot\config\environment.cmd`"",
        "if errorlevel 1 exit /b 1",
        'cd /d "%BLIND_DEV_ROOT%\workspace"',
        '"%BLIND_DEV_ROOT%\tools\mise\mise.exe" trust --yes "%MISE_CONFIG_FILE%"',
        "if errorlevel 1 exit /b 1",
        '"%BLIND_DEV_ROOT%\tools\mise\mise.exe" reshim',
        "if errorlevel 1 exit /b 1"
    )
    foreach ($tool in @("python", "node", "uv", "pnpm")) {
        $version = $manifest.toolchain.$tool
        $commands += "`"%BLIND_DEV_ROOT%\tools\mise\mise.exe`" exec $tool@$version -- $tool --version"
        $commands += "if errorlevel 1 exit /b 1"
    }
    $commands += "exit /b 0"
    $commands | Set-Content -LiteralPath $smokeScript -Encoding ascii
    Invoke-Checked -FilePath "cmd.exe" -ArgumentList @("/d", "/c", "`"$smokeScript`"")
    Write-Host "prepare, verify and the published toolchain passed."
}
finally {
    Remove-TestVolume
}
