# hamix-desktop from repo root: .\scripts\dev-desktop.ps1  (needs .env / DATABASE_URL)
# Schema migrate is a separate step: .\scripts\migrate.ps1
# Browser/API path remains: .\scripts\dev.ps1
param(
    [switch]$Migrate,
    [switch]$Live,
    [switch]$Help
)

$ErrorActionPreference = "Stop"
$repo = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

function Show-DevDesktopHelp {
    @"
Hamix desktop (Wails) — embeds the SPA + shared API runtime in one window.

Usage: .\scripts\dev-desktop.ps1 [-Migrate] [-Live] [-Help]

  -Migrate   Run .\scripts\migrate.ps1 first (convenience sugar)
  -Live      Use ``wails dev`` (Vite HMR) instead of a production SPA embed
  -Help      Show this help

Two-step workflow (same as browser):
  1. .\scripts\migrate.ps1
  2. .\scripts\dev-desktop.ps1

Default (no -Live): npm build → copy into cmd/hamix-desktop/frontend/dist →
  .\scripts\build-desktop.ps1 (go build -tags desktop,production) → run binary.
DSN: DATABASE_URL in .env (or first-run / Settings UI). See docs/adr/ADR-0095-desktop-wails-host.md.
"@
}

if ($Help) {
    Show-DevDesktopHelp
    exit 0
}

$envFile = Join-Path $repo ".env"
$envExample = Join-Path $repo ".env.example"
if (-not (Test-Path -LiteralPath $envFile)) {
    throw @"
.env not found at:
  $envFile

Copy .env.example to .env and set DATABASE_URL:
  Copy-Item '$envExample' '$envFile'

See CONTRIBUTING.md for setup. (Desktop can also use the in-app setup screen after launch,
but local agent/MCP wiring expects the same .env as .\scripts\dev.ps1.)
"@
}

$exeSuffix = if ($IsWindows -or $env:OS -match 'Windows') { '.exe' } else { '' }
$desktopExe = Join-Path $repo ("hamix-desktop-dev" + $exeSuffix)
$mcpExe = Join-Path $repo ("hamix-agent-mcp" + $exeSuffix)
$draftMcpExe = Join-Path $repo ("hamix-draft-mcp" + $exeSuffix)
$desktopDir = Join-Path $repo "cmd\hamix-desktop"
$copyScript = Join-Path $desktopDir "scripts\copy-web-dist.mjs"

Push-Location $repo
$desktop = $null
try {
    if ($Migrate) {
        & (Join-Path $PSScriptRoot "migrate.ps1")
    }

    & go mod download
    Set-Location (Join-Path $repo "web")
    & npm install
    Set-Location $repo

    # Agent MCP host for Cursor execute/verify (LookPath from the desktop process).
    & go build -o $mcpExe "./cmd/hamix-agent-mcp"
    # Draft-assist MCP host (Plan 4) for compose-page draft AI.
    & go build -o $draftMcpExe "./cmd/hamix-draft-mcp"
    $env:PATH = "$repo" + [IO.Path]::PathSeparator + $env:PATH

    if ($Live) {
        $wails = Get-Command wails -ErrorAction SilentlyContinue
        if ($null -eq $wails) {
            throw @"
wails CLI not found (required for -Live).

Install once:
  go install github.com/wailsapp/wails/v2/cmd/wails@latest

Or omit -Live to build the SPA and run the embedded desktop binary.
"@
        }
        Set-Location $desktopDir
        & wails dev
        return
    }

    Set-Location (Join-Path $repo "web")
    & npm run build
    Set-Location $repo
    & node $copyScript
    # Wails tags live only in build-desktop.* (ADR-0095 / manual builds guide).
    & (Join-Path $PSScriptRoot "build-desktop.ps1") -Out $desktopExe

    Write-Host "starting $desktopExe (Ctrl+C quits the window process)" -ForegroundColor DarkGray
    $desktop = Start-Process -FilePath $desktopExe -WorkingDirectory $repo -PassThru -NoNewWindow
    if ($null -eq $desktop) { throw "failed to start hamix-desktop" }
    Wait-Process -Id $desktop.Id
} finally {
    Pop-Location
    if ($null -ne $desktop -and -not $desktop.HasExited) {
        Stop-Process -Id $desktop.Id -Force -ErrorAction SilentlyContinue
    }
}
