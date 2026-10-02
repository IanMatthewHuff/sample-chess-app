<#
.SYNOPSIS
    Builds and launches the chess engine (Go), backend (Python) and frontend (Vite),
    waits for each to become healthy, then opens the game in the browser.

.EXAMPLE
    .\start.ps1
    .\start.ps1 -NoBrowser
    .\start.ps1 -FrontendPort 19471 -BackendPort 19472 -EnginePort 19473
#>
param(
    [int]$FrontendPort = 18471,
    [int]$BackendPort = 18472,
    [int]$EnginePort = 18473,
    [switch]$NoBrowser
)

$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot
$runDir = Join-Path $root '.run'
New-Item -ItemType Directory -Force -Path $runDir | Out-Null
$pidFile = Join-Path $runDir 'pids.json'

if (Test-Path $pidFile) {
    Write-Host 'Stopping previous instance...'
    & (Join-Path $root 'stop.ps1')
}

foreach ($port in $FrontendPort, $BackendPort, $EnginePort) {
    if (Get-NetTCPConnection -State Listen -LocalPort $port -ErrorAction SilentlyContinue) {
        throw "Port $port is already in use. Pass different ports, e.g. .\start.ps1 -FrontendPort 19471 -BackendPort 19472 -EnginePort 19473"
    }
}

function Wait-Healthy([string]$name, [string]$url, [System.Diagnostics.Process]$proc, [string]$log) {
    for ($i = 0; $i -lt 60; $i++) {
        if ($proc.HasExited) {
            Get-Content "$log.err" -ErrorAction SilentlyContinue | Select-Object -Last 20 | Write-Host
            throw "$name exited during startup (see $log.err)"
        }
        try {
            Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 2 | Out-Null
            Write-Host "  $name is up: $url" -ForegroundColor Green
            return
        } catch { Start-Sleep -Milliseconds 500 }
    }
    throw "$name did not become healthy at $url (see $log.out / $log.err)"
}

function Start-ChessService([string]$name, [string]$file, [string[]]$arguments, [string]$dir) {
    $log = Join-Path $runDir $name
    Start-Process -FilePath $file -ArgumentList $arguments -WorkingDirectory $dir -PassThru -WindowStyle Hidden `
        -RedirectStandardOutput "$log.out" -RedirectStandardError "$log.err"
}

$started = @{}
try {
    Write-Host 'Building engine...'
    $engineExe = Join-Path $runDir 'chess-engine.exe'
    Push-Location (Join-Path $root 'engine')
    try { go build -o $engineExe . ; if ($LASTEXITCODE) { throw 'go build failed' } } finally { Pop-Location }

    $frontendDir = Join-Path $root 'frontend'
    if (-not (Test-Path (Join-Path $frontendDir 'node_modules'))) {
        Write-Host 'Installing frontend dependencies...'
        Push-Location $frontendDir
        try { npm install --no-audit --no-fund; if ($LASTEXITCODE) { throw 'npm install failed' } } finally { Pop-Location }
    }

    $env:ENGINE_PORT = $EnginePort
    $env:ENGINE_URL = "http://127.0.0.1:$EnginePort"
    $env:BACKEND_PORT = $BackendPort
    $env:BACKEND_URL = "http://127.0.0.1:$BackendPort"
    $env:FRONTEND_PORT = $FrontendPort

    Write-Host 'Starting services...'
    $engine = Start-ChessService 'engine' $engineExe @() (Join-Path $root 'engine')
    $started.engine = $engine.Id
    Wait-Healthy 'Engine' "http://127.0.0.1:$EnginePort/health" $engine (Join-Path $runDir 'engine')

    $backend = Start-ChessService 'backend' 'python' @('app.py') (Join-Path $root 'backend')
    $started.backend = $backend.Id
    Wait-Healthy 'Backend' "http://127.0.0.1:$BackendPort/api/health" $backend (Join-Path $runDir 'backend')

    $vite = Join-Path $frontendDir 'node_modules\vite\bin\vite.js'
    $frontend = Start-ChessService 'frontend' 'node' @("`"$vite`"") $frontendDir
    $started.frontend = $frontend.Id
    Wait-Healthy 'Frontend' "http://127.0.0.1:$FrontendPort/" $frontend (Join-Path $runDir 'frontend')
}
catch {
    $started | ConvertTo-Json | Set-Content $pidFile
    & (Join-Path $root 'stop.ps1')
    throw
}

$started | ConvertTo-Json | Set-Content $pidFile
$url = "http://127.0.0.1:$FrontendPort/"
Write-Host ''
Write-Host "Chess is running at $url" -ForegroundColor Cyan
Write-Host "Logs: $runDir   Stop with: .\stop.ps1"
if (-not $NoBrowser) { Start-Process $url }
