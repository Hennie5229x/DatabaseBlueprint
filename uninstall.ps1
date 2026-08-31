$ErrorActionPreference = "Stop"

$appName = "blue"
$installDirectory = Join-Path $env:LOCALAPPDATA "DatabaseBlueprint"
$targetBinary = Join-Path $installDirectory "$appName.exe"

if (Test-Path $targetBinary) {
    Remove-Item $targetBinary -Force
    Write-Host "Removed $appName from $targetBinary"
}
else {
    Write-Host "$appName is not installed at $targetBinary"
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathEntries = @($userPath -split ";" | Where-Object { $_ -and $_ -ne $installDirectory })
[Environment]::SetEnvironmentVariable("Path", ($pathEntries -join ";"), "User")

Write-Host "Restart PowerShell or Command Prompt for PATH changes to take effect."
