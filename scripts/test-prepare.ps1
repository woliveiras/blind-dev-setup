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

function Invoke-Wizard {
    param(
        [string]$FilePath,
        [string[]]$ArgumentList = @(),
        [AllowEmptyCollection()]
        [string[]]$Answers = @()
    )

    # Give this invocation its own temp directory. The wizard keeps technical
    # output in blind-dev-setup-*.log; this scope avoids reading older logs or
    # depending on localized UI strings when an installation fails.
    $invocationTemp = Join-Path $WorkDirectory ("wizard-temp-" + [guid]::NewGuid().ToString("N"))
    New-Item -ItemType Directory -Path $invocationTemp | Out-Null
    $previousTemp = $env:TEMP
    $previousTmp = $env:TMP
    $lines = [System.Collections.Generic.List[string]]::new()
    $wizardExitCode = $null
    try {
        $env:TEMP = $invocationTemp
        $env:TMP = $invocationTemp
        # Preserve live progress while collecting the full UI transcript.
        $Answers | & $FilePath @ArgumentList 2>&1 | ForEach-Object {
            $line = [string]$_
            $lines.Add($line)
            Write-Host $line
        }
        $wizardExitCode = $LASTEXITCODE
    }
    finally {
        $env:TEMP = $previousTemp
        $env:TMP = $previousTmp
    }
    if ($wizardExitCode -ne 0) {
        try {
            foreach ($log in @(Get-ChildItem -LiteralPath $invocationTemp -Filter "blind-dev-setup-*.log" -File)) {
                Write-Host "Technical log tail: $($log.FullName)"
                Get-Content -LiteralPath $log.FullName -Tail 120 | Out-Host
            }
        }
        catch {
            Write-Warning "Could not read wizard diagnostics: $_"
        }
        throw "Wizard failed with exit code $wizardExitCode."
    }
    return ($lines -join [Environment]::NewLine)
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
    $bundleRoot = Join-Path $target "blind-dev-setup"
    $sentinel = Join-Path $target "preserve-existing-file.txt"
    $sentinelContent = "Wizard must preserve existing files."
    Set-Content -LiteralPath $sentinel -Value $sentinelContent -Encoding ascii

    # Redirected no-argument startup uses accessible mode automatically; it
    # must accept cancellation without writing. The VHD is not USB, so the
    # full test explicitly selects its path, retaining every consent prompt.
    $null = Invoke-Wizard -FilePath $executable -Answers @("0")
    $wizardArguments = @("wizard", "--accessible", "--target", $target, "--cache", $cache)
    $cancelScenarios = @(
        @{ Name = "exit at target selection"; Answers = @("0") },
        @{ Name = "exit before license acceptance"; Answers = @("1", "0") },
        @{ Name = "exit before destination confirmation"; Answers = @("1", "ACEITO", "0") },
        @{ Name = "EOF before license acceptance"; Answers = @("1") },
        @{ Name = "EOF before destination confirmation"; Answers = @("1", "ACEITO") }
    )
    foreach ($scenario in $cancelScenarios) {
        Write-Host "Testing accessible wizard cancellation: $($scenario.Name)."
        $null = Invoke-Wizard -FilePath $executable -ArgumentList $wizardArguments -Answers $scenario.Answers
        if ((Test-Path -LiteralPath $bundleRoot) -or (Test-Path -LiteralPath $cache) -or
            @(Get-ChildItem -LiteralPath $target -Force -Filter ".blind-dev-setup-staging-*").Count -ne 0) {
            throw "Cancelled wizard wrote an installation, staging directory or cache: $($scenario.Name)."
        }
        if ((Get-Content -LiteralPath $sentinel -Raw).Trim() -ne $sentinelContent) {
            throw "Cancelled wizard modified an existing file: $($scenario.Name)."
        }
    }

    # Exercise target choice, license acceptance, destination confirmation,
    # real preparation, automatic verification, and exit through the accessible
    # huh forms in one process/window. The visual UI remains a manual test.
    $transcript = Invoke-Wizard -FilePath $executable -ArgumentList $wizardArguments -Answers @(
        "1", "ACEITO", "SIM", "0"
    )
    if (-not $transcript.Contains("Preparação e verificação concluídas.")) {
        throw "Wizard did not report completion of preparation and verification."
    }
    if ((Get-Content -LiteralPath $sentinel -Raw).Trim() -ne $sentinelContent) {
        throw "Wizard modified an existing file while preparing the environment."
    }
    Invoke-Checked -FilePath $executable -ArgumentList @("verify", "--target", $target)

    if (-not (Test-Path -LiteralPath (Join-Path $bundleRoot "READY.txt") -PathType Leaf)) {
        throw "prepare did not publish READY.txt."
    }
    $staging = @(Get-ChildItem -LiteralPath $target -Force -Filter ".blind-dev-setup-staging-*")
    if ($staging.Count -ne 0) {
        throw "prepare left staging directories behind."
    }
    # Selecting an existing environment must offer verification, not overwrite it.
    $existingTranscript = Invoke-Wizard -FilePath $executable -ArgumentList $wizardArguments -Answers @(
        "1", "SIM", "0"
    )
    if (-not $existingTranscript.Contains("Verificação concluída.")) {
        throw "Wizard did not verify the existing environment."
    }

    # Test ordinary PATH resolution, without mise exec or the runner's tools.
    $smokePowerShell = Join-Path $WorkDirectory "check-toolchain.ps1"
    @'
param([switch]$RunExamples)

$ErrorActionPreference = "Stop"
$PSNativeCommandUseErrorActionPreference = $false
$root = [System.IO.Path]::GetFullPath($env:BLIND_DEV_ROOT).TrimEnd('\') + '\'
$manifest = Get-Content -LiteralPath (Join-Path $root "config\manifest.json") -Raw | ConvertFrom-Json
foreach ($tool in @("python", "node", "uv", "pnpm")) {
    $command = Get-Command $tool -ErrorAction Stop
    if (-not $command.Source.StartsWith($root, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "$tool resolved outside the portable environment: $($command.Source)"
    }
    Write-Host "$tool resolves to $($command.Source)"
    $version = (& $tool --version | Out-String).Trim()
    if ($LASTEXITCODE -ne 0) {
        throw "$tool --version failed with exit code $LASTEXITCODE."
    }
    $expected = [regex]::Escape($manifest.toolchain.$tool)
    if ($version -notmatch "(?<![0-9.])$expected(?![0-9.])") {
        throw "$tool returned unexpected version: $version"
    }
    Write-Host $version
}
# Portable Git must be usable after its silent extraction and post-install step.
$gitCommand = Get-Command git -ErrorAction Stop
$expectedGitPath = Join-Path $root "tools\git\cmd\git.exe"
if (-not [string]::Equals($gitCommand.Source, $expectedGitPath, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Git resolved outside the portable environment: $($gitCommand.Source)"
}
$gitVersion = (& git --version | Out-String).Trim()
if ($LASTEXITCODE -ne 0) { throw "Portable git --version failed." }
$gitComponent = @($manifest.components | Where-Object id -eq "git")[0]
$expectedGitVersion = $gitComponent.version -replace '^(\d+\.\d+\.\d+)\.(\d+)$', '$1.windows.$2'
if ($gitVersion -ne "git version $expectedGitVersion") {
    throw "Portable Git returned unexpected version: $gitVersion"
}
Write-Host "$gitVersion resolves to $($gitCommand.Source)"
$bashExecutable = Join-Path $root "tools\git\bin\bash.exe"
$bashOutput = (& $bashExecutable --noprofile --norc -c 'printf portable-git-ok' | Out-String).Trim()
if ($LASTEXITCODE -ne 0 -or $bashOutput -ne "portable-git-ok") {
    throw "Portable Bash could not execute the smoke command: $bashOutput"
}
Write-Host "Portable Bash execution passed: $bashExecutable"
if ($RunExamples) {
    $pythonExecutable = (& python -c 'import sys; print(sys.executable)' | Out-String).Trim()
    if ($LASTEXITCODE -ne 0 -or -not $pythonExecutable.StartsWith($root, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Python interpreter resolved outside the portable environment: $pythonExecutable"
    }
    Push-Location -LiteralPath (Join-Path $root "workspace\python-example")
    try {
        & uv sync --python $pythonExecutable --no-python-downloads
        if ($LASTEXITCODE -ne 0) { throw "uv sync failed." }
        & uv run --no-python-downloads python debug_example.py
        if ($LASTEXITCODE -ne 0) { throw "Python debug example failed." }
        $pythonBase = (& uv run --no-python-downloads python -c 'import sys; print(sys.base_prefix)' | Out-String).Trim()
        if ($LASTEXITCODE -ne 0 -or -not $pythonBase.StartsWith($root, [System.StringComparison]::OrdinalIgnoreCase)) {
            throw "uv used a Python installation outside the portable environment: $pythonBase"
        }
    }
    finally { Pop-Location }
    Push-Location -LiteralPath (Join-Path $root "workspace\node-workspace")
    try {
        & pnpm install
        if ($LASTEXITCODE -ne 0) { throw "pnpm install failed." }
        & pnpm start
        if ($LASTEXITCODE -ne 0) { throw "Node workspace example failed." }
    }
    finally { Pop-Location }
    Write-Host "Python and Node workspace examples passed."
}
'@ | Set-Content -LiteralPath $smokePowerShell -Encoding ascii

    foreach ($phase in @("published", "relocated")) {
        if ($phase -eq "relocated") {
            # Change only the partition belonging to our own mounted image.
            $usedLetters = @(Get-Volume | Where-Object DriveLetter | ForEach-Object { [string]$_.DriveLetter })
            $newLetter = 90..70 | ForEach-Object { [char]$_ } |
                Where-Object { [string]$_ -notin $usedLetters } | Select-Object -First 1
            if (-not $newLetter) {
                throw "No free drive letter for the portability test."
            }
            $partition | Set-Partition -NewDriveLetter $newLetter
            $target = "${newLetter}:\"
            $bundleRoot = Join-Path $target "blind-dev-setup"
            Invoke-Checked -FilePath $executable -ArgumentList @("verify", "--target", $target)
        }
        Write-Host "Testing $phase toolchain through PATH at $bundleRoot."
        $smokeScript = Join-Path $WorkDirectory "check-toolchain.cmd"
        $exampleArgument = if ($phase -eq "published") { " -RunExamples" } else { "" }
        @(
            "@echo off",
            "setlocal",
            'set "PATH=%SystemRoot%\System32;%SystemRoot%;%SystemRoot%\System32\WindowsPowerShell\v1.0"',
            "call `"$bundleRoot\config\initialize-toolchain.cmd`"",
            "if errorlevel 1 exit /b 1",
            'cd /d "%BLIND_DEV_ROOT%\workspace"',
            "if errorlevel 1 exit /b 1",
            "`"$PSHOME\pwsh.exe`" -NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -File `"$smokePowerShell`"$exampleArgument",
            "exit /b %errorlevel%"
        ) | Set-Content -LiteralPath $smokeScript -Encoding ascii
        Invoke-Checked -FilePath "cmd.exe" -ArgumentList @("/d", "/c", "`"$smokeScript`"")
    }
    # A valid but modified config must be rejected before automatic trust.
    $miseConfig = Join-Path $bundleRoot "config\mise.toml"
    $originalConfig = [System.IO.File]::ReadAllBytes($miseConfig)
    try {
        Add-Content -LiteralPath $miseConfig -Value "`n# CI: modified managed configuration" -Encoding utf8
        & cmd.exe /d /c "`"$smokeScript`"" | Out-Host
        if ($LASTEXITCODE -eq 0) {
            throw "Toolchain initialization accepted a modified mise.toml."
        }
        Write-Host "Toolchain initialization correctly rejected a modified mise.toml."
    }
    finally {
        [System.IO.File]::WriteAllBytes($miseConfig, $originalConfig)
    }
    Invoke-Checked -FilePath "cmd.exe" -ArgumentList @("/d", "/c", "`"$smokeScript`"")

    # Keep the install directory intact: verify must notice the missing binary.
    $nodeExecutable = Join-Path $bundleRoot "data\mise\installs\node\$($manifest.toolchain.node)\node.exe"
    $nodeBackup = "$nodeExecutable.e2e-backup"
    Move-Item -LiteralPath $nodeExecutable -Destination $nodeBackup
    try {
        & $executable verify --target $target | Out-Host
        if ($LASTEXITCODE -eq 0) {
            throw "verify accepted a toolchain installation without node.exe."
        }
        Write-Host "verify correctly rejected a missing toolchain executable."
    }
    finally {
        Move-Item -LiteralPath $nodeBackup -Destination $nodeExecutable
    }
    Invoke-Checked -FilePath $executable -ArgumentList @("verify", "--target", $target)
    Write-Host "Accessible wizard preparation, EOF/cancel, existing verification, PATH commands, drive-letter change and rejection checks passed."
}
finally {
    Remove-TestVolume
}
