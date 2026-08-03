# Hamix Go verification — source of truth for the CI backend job.
#
# Steps: gofmt, go vet, scheduling boundary, go tests (per group), funclogmeasure
#
# Usage (repo root): .\scripts\check-go.ps1 [flags]
#
# Flags:
#   -Verbose            Stream full tool output (CI uses this)
#   -SkipFunclog        Skip funclogmeasure -enforce
#   -LintOnly           Lint steps only (includes test-group coverage guard)
#   -TestsOnly          go test only (use with -Group for CI matrix cells)
#   -Group <name>       Restrict go test to core|tasks|agents|harness
#   -Help               Show options
#
# CI: ./scripts/check-go.sh --lint-only --verbose
#     ./scripts/check-go.sh --tests-only --group=core --verbose

param(
    [switch]$Help,
    [switch]$Verbose,
    [switch]$SkipFunclog,
    [switch]$LintOnly,
    [switch]$TestsOnly,
    [string]$Group = ""
)

if ($Help -or $args -contains '--help' -or $args -contains '-h') {
    Get-Content $PSCommandPath | Select-Object -Skip 1 -First 17 | ForEach-Object { $_ -replace '^# ?', '' }
    exit 0
}

if ($LintOnly -and $TestsOnly) {
    Write-Error "cannot use -LintOnly and -TestsOnly together"
    exit 2
}

$ErrorActionPreference = "Stop"
$repo = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Set-Location $repo

. (Join-Path $PSScriptRoot "test-groups.ps1")

if ($Group) {
    $null = Get-GroupPackages $Group
}

$CheckStart = Get-Date
$script:Step = 0
$script:Passed = 0
$script:StepStats = ""

if ($TestsOnly) {
    $script:Total = if ($Group) { 2 } else { 1 }
} elseif ($LintOnly) {
    $script:Total = if ($SkipFunclog) { 7 } else { 8 }
} else {
    $script:Total = if ($SkipFunclog) { 10 } else { 11 }
}

function Format-Duration {
    param([TimeSpan]$Span)
    $secs = [int][Math]::Round($Span.TotalSeconds)
    if ($secs -lt 60) { return "${secs}s" }
    return "{0}m{1:D2}s" -f [Math]::Floor($secs / 60), ($secs % 60)
}

function Write-StepPrefix {
    $script:Step++
    Write-Host -NoNewline "[$($script:Step)/$($script:Total)] "
}

function Fail-Step {
    param(
        [string]$Name,
        [int]$Code = 1,
        [string]$Fix = ""
    )
    Write-Host ""
    Write-Host "check FAILED: $Name ($($script:Step)/$($script:Total))" -ForegroundColor Red
    if ($Fix) { Write-Host "  fix: $Fix" -ForegroundColor Red }
    exit $Code
}

$script:CheckStepBudgetSecs = 120

function Enforce-StepBudget {
    param(
        [string]$Label,
        [TimeSpan]$Elapsed
    )
    if ($Label -eq "npm ci") { return }
    $secs = [int][Math]::Round($Elapsed.TotalSeconds)
    if ($secs -le $script:CheckStepBudgetSecs) { return }

    $shown = Format-Duration $Elapsed
    if ($env:GITHUB_ACTIONS) {
        Write-Host "::error title=check step budget::$Label exceeded step budget ($shown > $($script:CheckStepBudgetSecs)s). Split the suite, slim the harness, or speed up the tests."
    }
    Write-Host ""
    Write-Host "check FAILED: $Label exceeded step budget ($shown > $($script:CheckStepBudgetSecs)s)" -ForegroundColor Red
    Write-Host "  This step must finish within $($script:CheckStepBudgetSecs)s (keep check steps under two minutes)."
    Write-Host "  Split the suite, slim the harness, or speed up the tests — do not raise the budget casually."
    exit 1
}

function Complete-Ok {
    param([string]$Detail = "")
    $elapsed = (Get-Date) - $CheckStart
    Write-Host ""
    Write-Host "check OK  $($script:Passed)/$($script:Total) passed  $(Format-Duration $elapsed)" -ForegroundColor Green
    if ($Detail) { Write-Host "  ($Detail)" }
    exit 0
}

function Write-OkLine {
    param(
        [string]$Label,
        [TimeSpan]$Elapsed,
        [string]$Stats = ""
    )
    $pad = [Math]::Max(1, 22 - $Label.Length)
    $line = (" " * $pad) + "ok $(Format-Duration $Elapsed)"
    if ($Stats) { $line += "  ($Stats)" }
    Write-Host $line -ForegroundColor Green
    Enforce-StepBudget $Label $Elapsed
}

function Invoke-CapturedStep {
    param(
        [string]$Label,
        [scriptblock]$Command,
        [scriptblock]$StatsParser = $null
    )
    Write-StepPrefix
    Write-Host -NoNewline "$Label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $log = [System.IO.Path]::GetTempFileName()
    $code = 0
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'

    try {
        if ($Verbose) {
            Write-Host "..." -ForegroundColor Cyan
            & $Command
            $code = $LASTEXITCODE
        } else {
            & $Command 2>&1 | Out-File -FilePath $log -Encoding utf8
            $code = $LASTEXITCODE
        }
        if ($null -eq $code) { $code = 0 }
    } catch {
        $code = 1
        if (-not $Verbose) { $_ | Out-File -FilePath $log -Encoding utf8 -Append }
    } finally {
        $ErrorActionPreference = $prevEap
        $sw.Stop()
    }

    if ($code -eq 0) {
        $stats = ""
        if ($StatsParser -and -not $Verbose -and (Test-Path $log)) {
            $stats = & $StatsParser $log
        }
        $script:Passed++
        Write-OkLine $Label $sw.Elapsed $stats
        Remove-Item $log -Force -ErrorAction SilentlyContinue
        return
    }

    Write-Host "FAILED" -ForegroundColor Red
    if (-not $Verbose -and (Test-Path $log)) { Get-Content $log }
    Remove-Item $log -Force -ErrorAction SilentlyContinue
    Fail-Step $Label $code
}

function Get-GoTestStats {
    param([string]$LogPath)
    $content = Get-Content $LogPath -Raw
    $count = ([regex]::Matches($content, '(?m)^(ok|FAIL|\?)')).Count
    if ($count -gt 0) { return "$count packages" }
    return ""
}

function Step-Gofmt {
    $label = "gofmt"
    Write-StepPrefix
    Write-Host -NoNewline "$label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $unformatted = [System.Collections.Generic.List[string]]::new()
    Get-ChildItem -Recurse -Filter '*.go' -File |
        Where-Object { $_.FullName -notmatch '\\vendor\\' } |
        ForEach-Object {
            $line = & gofmt -l $_.FullName
            if ($line) {
                foreach ($path in ($line -split "`n")) {
                    if ($path) { [void]$unformatted.Add($path) }
                }
            }
        }
    $sw.Stop()

    if ($unformatted.Count -gt 0) {
        Write-Host "FAILED" -ForegroundColor Red
        $unformatted
        Fail-Step $label 1 "gofmt -w on the files above"
    }

    $script:Passed++
    Write-OkLine $label $sw.Elapsed
}

function Step-SchedulingBoundary {
    $label = "scheduling boundary"
    Write-StepPrefix
    Write-Host -NoNewline "$label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $boundaryHits = & rg -n "gorm|store/|handler/|agents/" pkgs/tasks/scheduling/ -g "*.go" -g "!*_test.go" 2>$null
    $sw.Stop()

    if ($boundaryHits) {
        Write-Host "FAILED" -ForegroundColor Red
        Write-Host "scheduling must not import persistence or transport:" -ForegroundColor Red
        $boundaryHits
        Fail-Step $label 1
    }

    $script:Passed++
    Write-OkLine $label $sw.Elapsed
}

function Step-SSEPublishBoundary {
    $label = "sse publish boundary"
    Write-StepPrefix
    Write-Host -NoNewline "$label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $allHits = & rg -n "h\.hub\.Publish" pkgs/tasks/handler -g "*.go" -g "!*_test.go" 2>$null
    $badHits = $allHits | Where-Object { $_ -notmatch "sse_notify\.go" }
    $sw.Stop()

    if ($badHits) {
        Write-Host "FAILED" -ForegroundColor Red
        Write-Host "h.hub.Publish must only appear in pkgs/tasks/handler/sse_notify.go:" -ForegroundColor Red
        $badHits
        Fail-Step $label 1
    }

    $script:Passed++
    Write-OkLine $label $sw.Elapsed
}

function Step-ProjectsBoundary {
    $label = "projects boundary"
    Write-StepPrefix
    Write-Host -NoNewline "$label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $hits = @()
    if (& rg -q "github.com/.*/pkgs/tasks/handler" pkgs/projects/ -g "*.go" 2>$null) {
        $hits += & rg -n "github.com/.*/pkgs/tasks/handler" pkgs/projects/ -g "*.go" 2>$null
    }
    if (& rg -q "github.com/.*/pkgs/tasks/store/internal" pkgs/projects/ -g "*.go" 2>$null) {
        $hits += & rg -n "github.com/.*/pkgs/tasks/store/internal" pkgs/projects/ -g "*.go" 2>$null
    }
    $sw.Stop()

    if ($hits) {
        Write-Host "FAILED" -ForegroundColor Red
        Write-Host "pkgs/projects must not import pkgs/tasks/handler or pkgs/tasks/store/internal:" -ForegroundColor Red
        $hits
        Fail-Step $label 1
    }

    $script:Passed++
    Write-OkLine $label $sw.Elapsed
}

function Step-GitinventoryBoundary {
    $label = "gitinventory boundary"
    Write-StepPrefix
    Write-Host -NoNewline "$label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $hits = @()
    if (& rg -q "github.com/.*/pkgs/tasks/handler" pkgs/gitinventory/ -g "*.go" 2>$null) {
        $hits += & rg -n "github.com/.*/pkgs/tasks/handler" pkgs/gitinventory/ -g "*.go" 2>$null
    }
    if (& rg -q "github.com/.*/pkgs/tasks/store/internal" pkgs/gitinventory/ -g "*.go" 2>$null) {
        $hits += & rg -n "github.com/.*/pkgs/tasks/store/internal" pkgs/gitinventory/ -g "*.go" 2>$null
    }
    $sw.Stop()

    if ($hits) {
        Write-Host "FAILED" -ForegroundColor Red
        Write-Host "pkgs/gitinventory must not import pkgs/tasks/handler or pkgs/tasks/store/internal:" -ForegroundColor Red
        $hits
        Fail-Step $label 1
    }

    $script:Passed++
    Write-OkLine $label $sw.Elapsed
}

function Step-SettingsBoundary {
    $label = "settings boundary"
    Write-StepPrefix
    Write-Host -NoNewline "$label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $hits = @()
    if (& rg -q "github.com/.*/pkgs/tasks/handler" pkgs/settings/ -g "*.go" 2>$null) {
        $hits += & rg -n "github.com/.*/pkgs/tasks/handler" pkgs/settings/ -g "*.go" 2>$null
    }
    if (& rg -q "github.com/.*/pkgs/tasks/store/internal" pkgs/settings/ -g "*.go" 2>$null) {
        $hits += & rg -n "github.com/.*/pkgs/tasks/store/internal" pkgs/settings/ -g "*.go" 2>$null
    }
    $sw.Stop()

    if ($hits) {
        Write-Host "FAILED" -ForegroundColor Red
        Write-Host "pkgs/settings must not import pkgs/tasks/handler or pkgs/tasks/store/internal:" -ForegroundColor Red
        $hits
        Fail-Step $label 1
    }

    $script:Passed++
    Write-OkLine $label $sw.Elapsed
}

function Step-TestGroupCoverage {
    $label = "test group coverage"
    Write-StepPrefix
    Write-Host -NoNewline "$label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        Assert-GroupsCoverAll
    } catch {
        $sw.Stop()
        Write-Host "FAILED" -ForegroundColor Red
        Fail-Step $label 1
    }
    $sw.Stop()

    $script:Passed++
    Write-OkLine $label $sw.Elapsed
}

function Step-DesktopBuildTagsContract {
    $label = "desktop build tags"
    Write-StepPrefix
    Write-Host -NoNewline "$label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $errors = @()
    $tagNeedle = '-tags desktop,production'
    $tagNeedleAlt = '-tags "desktop,production"'

    foreach ($rel in @(
            'scripts/build-desktop.ps1',
            'scripts/build-desktop.sh'
        )) {
        $path = Join-Path $repo $rel
        if (-not (Test-Path $path)) {
            $errors += "missing $rel"
            continue
        }
        $text = Get-Content -Raw $path
        if ($text -notmatch [regex]::Escape($tagNeedle) -and $text -notmatch [regex]::Escape($tagNeedleAlt)) {
            $errors += "$rel must contain go build -tags desktop,production"
        }
    }

    foreach ($pair in @(
            @{ Rel = 'scripts/dev-desktop.ps1'; Needle = 'build-desktop.ps1' },
            @{ Rel = 'scripts/dev-desktop.sh'; Needle = 'build-desktop.sh' }
        )) {
        $path = Join-Path $repo $pair.Rel
        if (-not (Test-Path $path)) {
            $errors += "missing $($pair.Rel)"
            continue
        }
        $text = Get-Content -Raw $path
        if ($text -notmatch [regex]::Escape($pair.Needle)) {
            $errors += "$($pair.Rel) must invoke $($pair.Needle)"
        }
        # Forbid ad-hoc host builds; MCP/other go build lines are fine.
        if ($text -match 'go\s+build[^\r\n]*cmd/hamix-desktop') {
            $errors += "$($pair.Rel) must not go build ./cmd/hamix-desktop (use build-desktop.*)"
        }
    }

    $sw.Stop()

    if ($errors.Count -gt 0) {
        Write-Host "FAILED" -ForegroundColor Red
        Write-Host "Wails desktop build-tag contract (ADR-0095):" -ForegroundColor Red
        $errors | ForEach-Object { Write-Host "  $_" -ForegroundColor Red }
        Fail-Step $label 1 "use scripts/build-desktop.* with -tags desktop,production"
    }

    $script:Passed++
    Write-OkLine $label $sw.Elapsed
}

function Invoke-CoverageGate {
    $label = "coverage gate ($Group)"
    Write-StepPrefix
    Write-Host -NoNewline "$label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'

    $gateArgs = @{ Group = $Group }
    if ($script:CoverProfile) {
        $gateArgs['Profile'] = $script:CoverProfile
    }
    & "$PSScriptRoot\coverage-gate.ps1" @gateArgs
    $code = $LASTEXITCODE
    $ErrorActionPreference = $prevEap
    $sw.Stop()

    if ($code -eq 0) {
        $script:Passed++
        Write-OkLine $label $sw.Elapsed
        return
    }

    Write-Host "FAILED" -ForegroundColor Red
    Fail-Step $label $code
}

function Invoke-GoTest {
    param(
        [Parameter(Mandatory)][string]$Label,
        [Parameter(Mandatory)][string]$Targets,
        [bool]$WantCover = $false
    )
    $script:CoverProfile = ""

    if ($WantCover) {
        $script:CoverProfile = [System.IO.Path]::GetTempFileName()
        $coverArgs = @('-coverprofile=' + $script:CoverProfile)
    } else {
        $coverArgs = @()
    }

    Write-StepPrefix
    Write-Host -NoNewline "$Label "

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $log = [System.IO.Path]::GetTempFileName()
    $code = 0
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'

    try {
        if ($Verbose) {
            Write-Host "..." -ForegroundColor Cyan
            go test $Targets.Split(' ') -count=1 @coverArgs
            $code = $LASTEXITCODE
        } else {
            go test $Targets.Split(' ') -count=1 @coverArgs 2>&1 | Out-File -FilePath $log -Encoding utf8
            $code = $LASTEXITCODE
        }
        if ($null -eq $code) { $code = 0 }
    } catch {
        $code = 1
        if (-not $Verbose) { $_ | Out-File -FilePath $log -Encoding utf8 -Append }
    } finally {
        $ErrorActionPreference = $prevEap
        $sw.Stop()
    }

    if ($code -eq 0) {
        $stats = ""
        if (-not $Verbose -and (Test-Path $log)) {
            $stats = Get-GoTestStats $log
        }
        $script:Passed++
        Write-OkLine $Label $sw.Elapsed $stats
        Remove-Item $log -Force -ErrorAction SilentlyContinue
        return
    }

    Write-Host "FAILED" -ForegroundColor Red
    if (-not $Verbose -and (Test-Path $log)) { Get-Content $log }
    Remove-Item $log -Force -ErrorAction SilentlyContinue
    if ($script:CoverProfile) {
        Remove-Item $script:CoverProfile -Force -ErrorAction SilentlyContinue
        $script:CoverProfile = ""
    }
    Fail-Step $Label $code
}

Write-Host "Hamix check (Go)"
Write-Host ""

if ($TestsOnly) {
    if ($Group) {
        Invoke-GoTest "go-tests ($Group)" ((Get-GroupPackages $Group) -join ' ') $true
        Invoke-CoverageGate
        if ($script:CoverProfile) {
            Remove-Item $script:CoverProfile -Force -ErrorAction SilentlyContinue
            $script:CoverProfile = ""
        }
    } else {
        Invoke-GoTest "go test" './...' $false
    }
    Complete-Ok
}

Invoke-CapturedStep "check-brand" { & "$PSScriptRoot\check-brand.ps1" }
Step-Gofmt
Invoke-CapturedStep "schema revision" { & "$PSScriptRoot\check-schema-revision.ps1" }
Invoke-CapturedStep "go vet" { go vet ./... }
Step-SchedulingBoundary
Step-SSEPublishBoundary
Step-ProjectsBoundary
Step-GitinventoryBoundary
Step-SettingsBoundary
Step-DesktopBuildTagsContract

if ($LintOnly) {
    Step-TestGroupCoverage
} else {
    foreach ($g in Get-GroupNames) {
        Invoke-GoTest "go-tests ($g)" ((Get-GroupPackages $g) -join ' ') $false
    }
}

if (-not $SkipFunclog) {
    Invoke-CapturedStep "funclogmeasure" { go run ./cmd/funclogmeasure -enforce }
}

Complete-Ok
