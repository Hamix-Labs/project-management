# Build hamix-desktop with required Wails tags (SSOT for go build flags).
# See https://wails.io/docs/guides/manual-builds/ and ADR-0095.
#
# Usage (repo root): .\scripts\build-desktop.ps1 [-Out path] [-Help]
param(
    [string]$Out = "",
    [switch]$Help
)

$ErrorActionPreference = "Stop"
$repo = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

function Show-Help {
    @"
Build cmd/hamix-desktop with Wails production tags.

Usage: .\scripts\build-desktop.ps1 [-Out path] [-Help]

  -Out path   Output binary (default: repo-root hamix-desktop-dev[.exe])
  -Help       Show this help

Always uses: go build -tags desktop,production
Do not use plain ``go build ./cmd/hamix-desktop`` — untagged builds are a Wails stub.
Dev console binary intentionally omits -H windowsgui (keeps Ctrl+C / logs).
"@
}

if ($Help) {
    Show-Help
    exit 0
}

$exeSuffix = if ($IsWindows -or $env:OS -match 'Windows') { '.exe' } else { '' }
if (-not $Out) {
    $Out = Join-Path $repo ("hamix-desktop-dev" + $exeSuffix)
}

Push-Location $repo
try {
    & go build -tags "desktop,production" -o $Out "./cmd/hamix-desktop"
    if ($LASTEXITCODE -ne 0) {
        throw "go build hamix-desktop failed (exit $LASTEXITCODE)"
    }
    Write-Host "built $Out (-tags desktop,production)" -ForegroundColor DarkGray
} finally {
    Pop-Location
}
