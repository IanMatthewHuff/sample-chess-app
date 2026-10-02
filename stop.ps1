# Stops the services started by start.ps1 (only the PIDs it recorded).
$pidFile = Join-Path $PSScriptRoot '.run\pids.json'
if (-not (Test-Path $pidFile)) {
    Write-Host 'Nothing to stop.'
    return
}
$pids = Get-Content $pidFile -Raw | ConvertFrom-Json
foreach ($entry in $pids.PSObject.Properties) {
    $proc = Get-Process -Id $entry.Value -ErrorAction SilentlyContinue
    if ($proc) {
        Stop-Process -Id $entry.Value -Force
        Write-Host "Stopped $($entry.Name) (PID $($entry.Value))"
    }
}
Remove-Item $pidFile
