# taskapi + Vite from repo root: .\scripts\dev.ps1  (needs .env / DATABASE_URL)
# Schema migrate is a separate step: .\scripts\migrate.ps1
param(
    [switch]$Migrate,
    [switch]$Help
)

$ErrorActionPreference = "Stop"
$repo = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

function Show-DevHelp {
    @"
taskapi + Vite dev servers (does not migrate by default).

Usage: .\scripts\dev.ps1 [-Migrate] [-Help]

  -Migrate   Run .\scripts\migrate.ps1 first (convenience sugar)
  -Help      Show this help

Two-step workflow:
  1. .\scripts\migrate.ps1     # after git pull with schema changes
  2. .\scripts\dev.ps1          # daily — starts API + Vite only
"@
}

if ($Help) {
    Show-DevHelp
    exit 0
}

function Stop-ListenerOnPort([int]$Port) {
    Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue |
        ForEach-Object { if ($_.OwningProcess) { Stop-Process -Id $_.OwningProcess -Force -ErrorAction SilentlyContinue } }
    Start-Sleep -Milliseconds 300
}

function Get-DevReadinessTimeoutSec {
    $sec = [int](& go run ./cmd/devconfig -readiness-timeout-sec 2>$null)
    if ($sec -le 0) { return 150 }
    return $sec
}

function Wait-TaskAPIPort([int]$Port, [int]$TimeoutSec, [System.Diagnostics.Process]$Proc) {
    $until = (Get-Date).AddSeconds($TimeoutSec)
    while ((Get-Date) -lt $until) {
        if ($Proc.HasExited) { return "exited" }
        try {
            $tcp = New-Object System.Net.Sockets.TcpClient
            $tcp.Connect("127.0.0.1", $Port)
            $tcp.Close()
            return "ready"
        } catch {
            Start-Sleep -Milliseconds 150
        }
    }
    if (-not $Proc.HasExited) { return "timeout" }
    return "exited"
}

$port = if ($env:DEV_TASKAPI_PORT) { [int]$env:DEV_TASKAPI_PORT } else { 8080 }
$exeSuffix = if ($IsWindows) { '.exe' } else { '' }
$exe = Join-Path $repo ("taskapi-dev" + $exeSuffix)
$mcpExe = Join-Path $repo ("hamix-agent-mcp" + $exeSuffix)
$draftMcpExe = Join-Path $repo ("hamix-draft-mcp" + $exeSuffix)
$readinessSec = if ($Migrate) { Get-DevReadinessTimeoutSec } else { 30 }

$envFile = Join-Path $repo ".env"
$envExample = Join-Path $repo ".env.example"
if (-not (Test-Path -LiteralPath $envFile)) {
    throw @"
.env not found at:
  $envFile

Copy .env.example to .env and set DATABASE_URL:
  Copy-Item '$envExample' '$envFile'

See CONTRIBUTING.md for setup.
"@
}

Push-Location $repo
$api = $null
try {
    if ($Migrate) {
        & (Join-Path $PSScriptRoot "migrate.ps1")
    }

    & go mod download
    Set-Location (Join-Path $repo "web")
    & npm install
    Set-Location $repo
    & go build -o $exe "./cmd/taskapi"
    # Agent MCP host for Cursor execute/verify (LookPath from the taskapi process).
    & go build -o $mcpExe "./cmd/hamix-agent-mcp"
    # Draft-assist MCP host (Plan 4) for compose-page draft AI (bound to taskapi via --bind).
    & go build -o $draftMcpExe "./cmd/hamix-draft-mcp"

    # Draft-assist sidecar (Node). Build the ESM bundle and drop a launcher on
    # PATH so taskapi can spawn `hamix-draft-agent` the same way it spawns
    # `hamix-agent-mcp`.
    $sidecar = Join-Path $repo "sidecars\hamix-draft-agent"
    if (Test-Path -LiteralPath $sidecar) {
        Push-Location $sidecar
        try {
            if (Get-Command pnpm -ErrorAction SilentlyContinue) {
                & pnpm install --silent
                & pnpm run build
            } else {
                & npm install --silent
                & npm run build
            }
        } finally {
            Pop-Location
        }
        $bundle = Join-Path $sidecar "dist\hamix-draft-agent.js"
        $rootJs = Join-Path $repo "hamix-draft-agent.js"
        Copy-Item -LiteralPath $bundle -Destination $rootJs -Force
        $cmdShim = Join-Path $repo "hamix-draft-agent.cmd"
        @"
@echo off
node "%~dp0hamix-draft-agent.js" %*
"@ | Set-Content -LiteralPath $cmdShim -Encoding ASCII
    }

    $env:PATH = "$repo" + [IO.Path]::PathSeparator + $env:PATH

    Stop-ListenerOnPort $port
    $api = Start-Process -FilePath $exe -ArgumentList "-port", "$port" -WorkingDirectory $repo -PassThru -NoNewWindow
    if ($null -eq $api) { throw "failed to start taskapi" }

    $result = Wait-TaskAPIPort $port $readinessSec $api
    if ($result -eq "ready" -and -not $api.HasExited) {
        Set-Location (Join-Path $repo "web")
        npm run dev
        return
    }

    if ($null -ne $api -and -not $api.HasExited) {
        Stop-Process -Id $api.Id -Force -ErrorAction SilentlyContinue
    }

    if ($result -eq "timeout") {
        throw "taskapi did not listen on :$port within ${readinessSec}s (still starting?). Try .\scripts\migrate.ps1 if schema changed, or check logs/taskapi-*.jsonl"
    }

    $code = if ($null -ne $api) { $api.ExitCode } else { -1 }
    throw "taskapi exited on :$port (exit $code) before listening. See stderr above, logs/taskapi-*.jsonl, or run .\scripts\migrate.ps1 if schema changed."
} finally {
    Pop-Location
    if ($null -ne $api -and -not $api.HasExited) {
        Stop-Process -Id $api.Id -Force -ErrorAction SilentlyContinue
    }
}
