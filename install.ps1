# bgit Windows installation script
# Usage: irm https://raw.githubusercontent.com/byterings/bgit/main/install.ps1 | iex
#
# This script downloads bgit and installs it to your user directory.
# Review this script before running: https://github.com/byterings/bgit/blob/main/install.ps1

$ErrorActionPreference = "Stop"

$Version = "0.1.0"
$GithubRepo = "byterings/bgit"
$InstallDir = "$env:LOCALAPPDATA\bgit"

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

$Binary = "bgit-windows-$Arch.exe"
$DownloadUrl = "https://github.com/$GithubRepo/releases/download/v$Version/$Binary"

Write-Host ""
Write-Host "bgit installer v$Version" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan
Write-Host ""

# Create install directory
if (-not (Test-Path $InstallDir)) {
    Write-Host "Creating directory: $InstallDir"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Download binary
$TempFile = Join-Path $env:TEMP "bgit.exe"
Write-Host "Downloading from: $DownloadUrl"

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempFile -UseBasicParsing
} catch {
    Write-Host "Error: Failed to download bgit" -ForegroundColor Red
    Write-Host "Please check your internet connection or download manually from:" -ForegroundColor Yellow
    Write-Host "https://github.com/$GithubRepo/releases" -ForegroundColor Yellow
    exit 1
}

# Move to install directory
$DestPath = Join-Path $InstallDir "bgit.exe"
Move-Item -Path $TempFile -Destination $DestPath -Force

Write-Host "Installed to: $DestPath" -ForegroundColor Green

# Add to PATH if not already present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "Adding bgit to PATH..."
    $NewPath = "$UserPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    $env:Path = "$env:Path;$InstallDir"
    Write-Host "Added to PATH" -ForegroundColor Green
} else {
    Write-Host "Already in PATH" -ForegroundColor Green
}

# Verify installation
Write-Host ""
Write-Host "Verifying installation..."
try {
    $VersionOutput = & $DestPath --version 2>&1
    Write-Host "bgit installed successfully!" -ForegroundColor Green
    Write-Host $VersionOutput
} catch {
    Write-Host "Installation completed, but verification failed." -ForegroundColor Yellow
    Write-Host "Please restart your terminal and run: bgit --version"
}

Write-Host ""
Write-Host "Get started:" -ForegroundColor Cyan
Write-Host "  bgit add           # Add your first identity"
Write-Host "  bgit use <alias>   # Switch identity"
Write-Host "  bgit list          # List all identities"
Write-Host ""
Write-Host "Note: You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
Write-Host ""
