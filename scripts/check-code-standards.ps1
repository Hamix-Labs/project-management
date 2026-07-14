# Layout guardrails (see .cursor/rules/web-layout.mdc and backend/go/layout.mdc).
# Exit 0 when clean; exit 1 when a rule is violated.
# Run from repository root: pwsh -File scripts/check-code-standards.ps1

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
if (-not (Test-Path (Join-Path $root "go.mod"))) {
    Write-Error "Run from repository root (go.mod not found next to scripts/)."
    exit 1
}

$failed = $false

function Test-IsUnderWebSrcApi {
    param([string]$FullPath)
    # CI (Linux) uses `/`; Windows uses `\`. `-like '*\web\src\api\*'` only matched Windows.
    $n = $FullPath.Replace('\', '/')
    return $n.Contains('/web/src/api/')
}

# TypeScript: fetch() must only appear under web/src/api/ (exclude tests).
$srcRoot = Join-Path $root (Join-Path "web" "src")
if (Test-Path $srcRoot) {
    $tsFiles = Get-ChildItem -Path $srcRoot -Recurse -Include *.ts, *.tsx -File |
        Where-Object {
            (-not (Test-IsUnderWebSrcApi $_.FullName)) -and
            $_.Name -notmatch '\.test\.(ts|tsx)$' -and
            ($_.FullName.Replace('\', '/') -notmatch '/test/')
        }
    # Match global `fetch(` only: exclude `.refetch(` / `prefetch(` and JSDoc
    # lines like "stats fetch (" where `fetch` is not the global.
    $fetchPat = '(?:^|[^\w.])fetch\s*\('
    foreach ($f in $tsFiles) {
        $text = Get-Content -LiteralPath $f.FullName -Raw
        if ($null -eq $text) { continue }
        if ($text -match $fetchPat) {
            Write-Host "VIOLATION: fetch( outside web/src/api/: $($f.FullName)" -ForegroundColor Red
            $failed = $true
        }
    }
}

$stylesRoot = Join-Path $srcRoot (Join-Path "app" "styles")
if (Test-Path $stylesRoot) {
    $cssFiles = Get-ChildItem -Path $stylesRoot -Recurse -Filter *.css -File
    $componentCssFiles = $cssFiles | Where-Object {
        $_.FullName.Replace('\', '/') -notmatch '/web/src/app/styles/tokens/'
    }
    $rawColorPat = '#[0-9a-fA-F]{3,8}\b|rgba?\(|hsla?\('
    foreach ($f in $componentCssFiles) {
        $text = Get-Content -LiteralPath $f.FullName -Raw
        if ($null -eq $text) { continue }
        if ($text -match $rawColorPat) {
            Write-Host "VIOLATION: raw color outside web style tokens: $($f.FullName)" -ForegroundColor Red
            $failed = $true
        }
    }

    $tooSmallRemPat = 'font-size:\s*0\.[0-6][0-9]*rem'
    foreach ($f in $componentCssFiles) {
        $text = Get-Content -LiteralPath $f.FullName -Raw
        if ($null -eq $text) { continue }
        if ($text -match $tooSmallRemPat) {
            Write-Host "VIOLATION: font-size below --text-xs in component CSS: $($f.FullName)" -ForegroundColor Red
            $failed = $true
        }
    }
}

# Note: pkgs/tasks/domain embeds GORM struct tags and gorm.io/datatypes; a
# naive "no gorm in domain" check would false-positive. Tightening domain
# purity is a later CODE_STANDARDS stage (split models vs pure domain).

# Go: handler must not import database drivers directly.
$handlerRoot = Join-Path $root (Join-Path "pkgs" (Join-Path "tasks" "handler"))
if (Test-Path $handlerRoot) {
    $goFiles = Get-ChildItem -Path $handlerRoot -Recurse -Filter *.go -File |
        Where-Object { $_.Name -notmatch '_test\.go$' }
    foreach ($f in $goFiles) {
        $text = Get-Content -LiteralPath $f.FullName -Raw
        if ($text -match 'database/sql|jackc/pgx|gorm\.io/gorm') {
            Write-Host "VIOLATION: handler imports DB stack: $($f.FullName)" -ForegroundColor Red
            $failed = $true
        }
    }
}

# Go: readpolicy/writepolicy pure subpackages must not import HTTP or DB stack.
$policyDirs = @(
    (Join-Path $handlerRoot "readpolicy"),
    (Join-Path $handlerRoot "writepolicy")
)
foreach ($dir in $policyDirs) {
    if (-not (Test-Path $dir)) { continue }
    $policyFiles = Get-ChildItem -Path $dir -Filter *.go -File |
        Where-Object { $_.Name -notmatch '_test\.go$' }
    foreach ($f in $policyFiles) {
        $text = Get-Content -LiteralPath $f.FullName -Raw
        if ($text -match 'database/sql|jackc/pgx|gorm\.io/gorm|net/http') {
            Write-Host "VIOLATION: handler policy subpackage imports HTTP/DB: $($f.FullName)" -ForegroundColor Red
            $failed = $true
        }
    }
}

# TypeScript: mutations pure modules must not import React.
$mutationsRoot = Join-Path $srcRoot (Join-Path "tasks" "mutations")
if (Test-Path $mutationsRoot) {
    $mutationPureFiles = Get-ChildItem -Path $mutationsRoot -Filter *.ts -File |
        Where-Object { $_.Name -notmatch '\.test\.ts$' }
    foreach ($f in $mutationPureFiles) {
        $text = Get-Content -LiteralPath $f.FullName -Raw
        if ($null -eq $text) { continue }
        if ($text -match 'from\s+["'']react["'']|from\s+["'']react/') {
            Write-Host "VIOLATION: mutations pure module imports react: $($f.FullName)" -ForegroundColor Red
            $failed = $true
        }
    }
}

# TypeScript: create slice pure modules must not import React or modal components.
$createRoot = Join-Path $srcRoot (Join-Path "tasks" "create")
if (Test-Path $createRoot) {
    $createPureFiles = Get-ChildItem -Path $createRoot -Filter *.ts -File |
        Where-Object { $_.Name -notmatch '\.test\.ts$' }
    foreach ($f in $createPureFiles) {
        $text = Get-Content -LiteralPath $f.FullName -Raw
        if ($null -eq $text) { continue }
        if ($text -match 'from\s+["'']react["'']|from\s+["'']react/') {
            Write-Host "VIOLATION: create pure module imports react: $($f.FullName)" -ForegroundColor Red
            $failed = $true
        }
        if ($text -match 'task-create-modal') {
            Write-Host "VIOLATION: create pure module imports task-create-modal: $($f.FullName)" -ForegroundColor Red
            $failed = $true
        }
    }
}

# TypeScript: vertical modules must not cross-import (web-layout.mdc).
$featureDirs = @(
    @{ Name = "tasks"; Path = (Join-Path $srcRoot "tasks"); Forbidden = @("projects", "worktrees") },
    @{ Name = "projects"; Path = (Join-Path $srcRoot "projects"); Forbidden = @("tasks", "settings", "worktrees") },
    @{ Name = "settings"; Path = (Join-Path $srcRoot "settings"); Forbidden = @("tasks", "projects", "worktrees") },
    @{ Name = "worktrees"; Path = (Join-Path $srcRoot "worktrees"); Forbidden = @("tasks", "projects", "settings") }
)
$featureImportPat = 'from\s+["'']@/(tasks|projects|settings|worktrees)/'
foreach ($feat in $featureDirs) {
    if (-not (Test-Path $feat.Path)) { continue }
    $featFiles = Get-ChildItem -Path $feat.Path -Recurse -Include *.ts, *.tsx -File
    foreach ($f in $featFiles) {
        $text = Get-Content -LiteralPath $f.FullName -Raw
        if ($null -eq $text) { continue }
        $matches = [regex]::Matches($text, $featureImportPat)
        foreach ($m in $matches) {
            $imported = $m.Groups[1].Value
            if ($feat.Forbidden -contains $imported) {
                Write-Host "VIOLATION: $($feat.Name) feature imports @$imported/: $($f.FullName)" -ForegroundColor Red
                $failed = $true
            }
        }
    }
}

# TypeScript: vertical production code must not call invalidateQueries outside allowed subdirs.
$verticals = @(
    @{ Name = "projects"; Path = (Join-Path $srcRoot "projects"); AllowSubdirs = @("mutations") },
    @{ Name = "worktrees"; Path = (Join-Path $srcRoot "worktrees"); AllowSubdirs = @("mutations") },
    @{ Name = "tasks"; Path = (Join-Path $srcRoot "tasks"); AllowSubdirs = @("mutations", "sync") }
)
$invalidatePat = 'invalidateQueries'
foreach ($vertical in $verticals) {
    if (-not (Test-Path $vertical.Path)) { continue }
    $allowSubdirs = $vertical.AllowSubdirs
    $verticalFiles = Get-ChildItem -Path $vertical.Path -Recurse -Include *.ts, *.tsx -File |
        Where-Object { $_.Name -notmatch '\.test\.(ts|tsx)$' }
    foreach ($f in $verticalFiles) {
        $normalized = $f.FullName.Replace('\', '/')
        $allowed = $false
        foreach ($subdir in $allowSubdirs) {
            if ($normalized -match "/$subdir/") { $allowed = $true; break }
        }
        if ($allowed) { continue }
        $text = Get-Content -LiteralPath $f.FullName -Raw
        if ($null -eq $text) { continue }
        if ($text -match $invalidatePat) {
            Write-Host "VIOLATION: invalidateQueries outside allowed paths in $($vertical.Name)/: $($f.FullName)" -ForegroundColor Red
            $failed = $true
        }
    }
}

# Warn-only file-size scan (CODE_STANDARDS.mdc). Yellow/red prints; exit stays 0.
$sizeWarnCount = 0

function Get-CodeStandardsSizeZone {
    param([string]$RelPath, [string]$FileName)
    $n = $RelPath.Replace('\', '/').ToLowerInvariant()
    $fn = $FileName.ToLowerInvariant()

    if ($fn -match '_test\.go$') {
        return @{ Zone = 'Go *_test.go'; Green = 400; Red = 600 }
    }
    if ($n -match '/cmd/[^/]+/main\.go$') {
        return @{ Zone = 'Go cmd/main.go'; Green = 80; Red = 120 }
    }
    if ($n -match '/domain/') {
        return @{ Zone = 'Go domain'; Green = 200; Red = 350 }
    }
    if ($n -match '/store/internal/' -or $fn -match '^store_') {
        return @{ Zone = 'Go store'; Green = 300; Red = 500 }
    }
    if ($fn -match '^handler_.*\.go$') {
        return @{ Zone = 'Go handler_*.go'; Green = 300; Red = 500 }
    }
    if ($fn -match '_json\.go$') {
        return @{ Zone = 'Go *_json.go'; Green = 200; Red = 350 }
    }
    if ($n -match '/middleware/') {
        return @{ Zone = 'Go middleware'; Green = 150; Red = 250 }
    }
    if ($fn -match '\.go$') {
        return @{ Zone = 'Go general'; Green = 400; Red = 800 }
    }
    if ($fn -match 'page\.tsx$') {
        return @{ Zone = 'TS *Page.tsx'; Green = 80; Red = 150 }
    }
    if ($fn -match '^use.+\.ts$') {
        return @{ Zone = 'TS use*.ts hook'; Green = 80; Red = 150 }
    }
    if ($n -match '/web/src/api/') {
        return @{ Zone = 'TS api/*.ts'; Green = 200; Red = 350 }
    }
    if ($n -match '/utils/') {
        return @{ Zone = 'TS utils/*.ts'; Green = 150; Red = 250 }
    }
    if ($fn -match '\.test\.tsx$') {
        return @{ Zone = 'TS *.test.tsx'; Green = 300; Red = 500 }
    }
    if ($fn -match '\.css$' -and $n -notmatch '/web/src/app/styles/tokens/') {
        return @{ Zone = 'TS component CSS'; Green = 200; Red = 350 }
    }
    if ($fn -match '(section|panel|view|layout)\.tsx$') {
        return @{ Zone = 'TS container component'; Green = 120; Red = 200 }
    }
    if ($fn -match '\.tsx$') {
        return @{ Zone = 'TS presentational component'; Green = 150; Red = 250 }
    }
    if ($fn -match '\.ts$') {
        return @{ Zone = 'TS general'; Green = 200; Red = 400 }
    }
    return $null
}

function Test-IsGeneratedGo {
    param([string]$Text)
    return ($null -ne $Text) -and ($Text -match '(?m)^// Code generated\b')
}

$sizeScanRoots = @(
    (Join-Path $root "pkgs"),
    (Join-Path $root "cmd"),
    (Join-Path $root "internal")
)
$goSizeFiles = @()
foreach ($scanRoot in $sizeScanRoots) {
    if (-not (Test-Path $scanRoot)) { continue }
    $goSizeFiles += Get-ChildItem -Path $scanRoot -Recurse -Filter *.go -File
}
if (Test-Path $srcRoot) {
    $goSizeFiles += Get-ChildItem -Path $srcRoot -Recurse -Include *.ts, *.tsx, *.css -File |
        Where-Object { $_.FullName.Replace('\', '/') -notmatch '/node_modules/|/dist/' }
}

foreach ($f in $goSizeFiles) {
    $rel = $f.FullName.Substring($root.Length).TrimStart('\', '/')
    $zone = Get-CodeStandardsSizeZone -RelPath $rel -FileName $f.Name
    if ($null -eq $zone) { continue }

    $text = Get-Content -LiteralPath $f.FullName -Raw
    if ($f.Extension -eq '.go' -and (Test-IsGeneratedGo $text)) { continue }

    $lines = (Get-Content -LiteralPath $f.FullName | Measure-Object -Line).Lines
    if ($lines -le $zone.Green) { continue }

    $sizeWarnCount++
    if ($lines -gt $zone.Red) {
        Write-Host "SIZE (red): $lines lines [$($zone.Zone)] $rel" -ForegroundColor Red
    } else {
        Write-Host "SIZE (yellow): $lines lines [$($zone.Zone)] $rel" -ForegroundColor Yellow
    }
}

if ($failed) {
    exit 1
}
if ($sizeWarnCount -gt 0) {
    Write-Host "check-code-standards: OK ($sizeWarnCount file-size warning(s); warn-only)" -ForegroundColor Green
} else {
    Write-Host "check-code-standards: OK" -ForegroundColor Green
}
exit 0
