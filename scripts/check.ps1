# Hamix full verification — runs check-go.ps1 then check-web.ps1.
#
# Usage (repo root): .\scripts\check.ps1 [flags]
#
# Flags:
#   -Verbose          Stream full tool output
#   -GoOnly           Run check-go.ps1 only
#   -WebOnly          Run check-web.ps1 only
#   -Install          Pass -Install to check-web.ps1 (npm ci)
#   -SkipFunclog      Pass -SkipFunclog to check-go.ps1
#   -Help             Show options
#
# Default (quiet): one line per step on success; full output only on failure.
# CI runs check-go.sh and check-web.sh directly — see .github/workflows/ci.yml

param(
    [switch]$Help,
    [switch]$Verbose,
    [switch]$GoOnly,
    [switch]$WebOnly,
    [switch]$Install,
    [switch]$SkipFunclog
)

if ($Help -or $args -contains '--help' -or $args -contains '-h') {
    Get-Content $PSCommandPath | Select-Object -Skip 1 -First 15 | ForEach-Object { $_ -replace '^# ?', '' }
    exit 0
}

if ($GoOnly -and $WebOnly) {
    Write-Error "cannot use -GoOnly and -WebOnly together"
    exit 2
}

$scriptDir = $PSScriptRoot
# Hashtable splats so switches bind by name (string-array splats bind to $Group).
$goArgs = @{}
$webArgs = @{}
if ($Verbose) { $goArgs['Verbose'] = $true; $webArgs['Verbose'] = $true }
if ($SkipFunclog) { $goArgs['SkipFunclog'] = $true }
if ($Install) { $webArgs['Install'] = $true }

if ($WebOnly) {
    & (Join-Path $scriptDir "check-web.ps1") @webArgs
    exit $LASTEXITCODE
}

if ($GoOnly) {
    & (Join-Path $scriptDir "check-go.ps1") @goArgs
    exit $LASTEXITCODE
}

& (Join-Path $scriptDir "check-go.ps1") @goArgs
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

& (Join-Path $scriptDir "check-web.ps1") @webArgs
exit $LASTEXITCODE
