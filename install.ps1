$ErrorActionPreference = "Stop"

$appName = "blue"
$repository = "Hennie5229x/DatabaseBlueprint"
$installDirectory = Join-Path $env:LOCALAPPDATA "DatabaseBlueprint"
$targetBinary = Join-Path $installDirectory "$appName.exe"
$downloadUrl = "https://github.com/$repository/releases/latest/download/blue-windows-amd64.exe"
$temporaryFile = Join-Path ([System.IO.Path]::GetTempPath()) "blue-$([guid]::NewGuid()).exe"

try {
    Write-Host "Downloading $appName for windows/amd64..."
    Invoke-WebRequest -Uri $downloadUrl -OutFile $temporaryFile

    New-Item -ItemType Directory -Path $installDirectory -Force | Out-Null
    Move-Item -Path $temporaryFile -Destination $targetBinary -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $pathEntries = @($userPath -split ";" | Where-Object { $_ })
    if ($pathEntries -notcontains $installDirectory) {
        [Environment]::SetEnvironmentVariable("Path", (($pathEntries + $installDirectory) -join ";"), "User")
    }

    Write-Host "Installed $appName to $targetBinary"
    Write-Host "Restart PowerShell or Command Prompt before running '$appName'."
}
finally {
    if (Test-Path $temporaryFile) {
        Remove-Item $temporaryFile -Force
    }
}
