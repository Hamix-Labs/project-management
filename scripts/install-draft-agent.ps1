# Builds sidecars/hamix-draft-agent, writes the repo-root launcher, and sets
# HAMIX_DRAFT_AGENT_BIN so taskapi/desktop can spawn it without PATH luck.
# Required: missing sidecar dir or a failed build throws (fail-boot).
$ErrorActionPreference = "Stop"

$repo = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$sidecar = Join-Path $repo "sidecars\hamix-draft-agent"
if (-not (Test-Path -LiteralPath $sidecar)) {
    throw "sidecars/hamix-draft-agent is missing; draft-assist cannot start without it."
}

Push-Location $sidecar
try {
    if (Get-Command pnpm -ErrorAction SilentlyContinue) {
        & pnpm install --silent
        if ($LASTEXITCODE -ne 0) { throw "pnpm install failed for hamix-draft-agent ($LASTEXITCODE)" }
        & pnpm run build
        if ($LASTEXITCODE -ne 0) { throw "pnpm build failed for hamix-draft-agent ($LASTEXITCODE)" }
    } else {
        & npm install --silent
        if ($LASTEXITCODE -ne 0) { throw "npm install failed for hamix-draft-agent ($LASTEXITCODE)" }
        & npm run build
        if ($LASTEXITCODE -ne 0) { throw "npm build failed for hamix-draft-agent ($LASTEXITCODE)" }
    }
} finally {
    Pop-Location
}

$bundle = Join-Path $sidecar "dist\hamix-draft-agent.js"
if (-not (Test-Path -LiteralPath $bundle)) {
    throw "sidecar build did not produce $bundle"
}

Copy-Item -LiteralPath $bundle -Destination (Join-Path $repo "hamix-draft-agent.js") -Force
$cmdShim = Join-Path $repo "hamix-draft-agent.cmd"
@"
@echo off
node "%~dp0hamix-draft-agent.js" %*
"@ | Set-Content -LiteralPath $cmdShim -Encoding ASCII

$env:HAMIX_DRAFT_AGENT_BIN = $cmdShim
