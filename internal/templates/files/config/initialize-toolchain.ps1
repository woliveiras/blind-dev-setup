# Generated hashes detect changes to the managed configuration before mise trusts it.
# They are not a signature: use verify from a trusted generator after suspected tampering.
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

try {
    $root = Split-Path -Parent $PSScriptRoot
    foreach ($entry in @(
        @{ Name = 'mise.toml'; Hash = '__MISE_CONFIG_SHA256__' },
        @{ Name = 'mise.lock'; Hash = '__MISE_LOCK_SHA256__' }
    )) {
        $path = Join-Path $PSScriptRoot $entry.Name
        if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
            throw "Configuracao obrigatoria ausente: $($entry.Name)."
        }
        if ((Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash -ne $entry.Hash) {
            throw "Configuracao gerenciada alterada: $($entry.Name). Execute verify com o gerador desta versao."
        }
    }

    $mise = Join-Path $root 'tools/mise/mise.exe'
    # Run outside user projects: their configuration must not be trusted implicitly.
    Push-Location -LiteralPath $root
    try {
        & $mise trust --yes (Join-Path $PSScriptRoot 'mise.toml')
        if ($LASTEXITCODE -ne 0) {
            throw 'O mise nao conseguiu autorizar a configuracao gerenciada.'
        }
        # Rebuild even existing shims; the bundle may have moved or changed drive letter.
        & $mise reshim --force
        if ($LASTEXITCODE -ne 0) {
            throw 'O mise nao conseguiu recriar os comandos do ambiente.'
        }
    }
    finally {
        Pop-Location
    }
}
catch {
    Write-Host "Erro: $($_.Exception.Message)"
    exit 1
}
exit 0
