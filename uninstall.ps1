# bgit Windows uninstall script
# Usage: irm https://raw.githubusercontent.com/byterings/bgit/main/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

$InstallDir = "$env:LOCALAPPDATA\bgit"

Write-Host ""
Write-Host "bgit Windows cleanup" -ForegroundColor Cyan
Write-Host "====================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Removing bgit from user PATH..."
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath) {
    $NewEntries = @()
    foreach ($Entry in ($UserPath -split ';')) {
        if (-not $Entry) {
            continue
        }

        $Keep = $true
        try {
            $Keep = ([IO.Path]::GetFullPath($Entry.Trim('"')) -ine [IO.Path]::GetFullPath($InstallDir))
        } catch {
            $Keep = $true
        }

        if ($Keep) {
            $NewEntries += $Entry
        }
    }
    [Environment]::SetEnvironmentVariable("Path", ($NewEntries -join ';'), "User")
    Write-Host "PATH cleanup complete" -ForegroundColor Green
}

if (Test-Path $InstallDir) {
    Write-Host "Removing install directory: $InstallDir"
    Remove-Item $InstallDir -Recurse -Force
    Write-Host "Install directory removed" -ForegroundColor Green
} else {
    Write-Host "Install directory not found" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "To remove bgit Git/SSH integration, run before deleting config:" -ForegroundColor Yellow
Write-Host "  bgit uninstall --remove-config"
Write-Host ""
