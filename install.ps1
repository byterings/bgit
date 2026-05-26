# bgit Windows installation script
# Usage: irm https://raw.githubusercontent.com/byterings/bgit/main/install.ps1 | iex
#
# This script downloads bgit and installs it to your user directory.
# Review this script before running: https://github.com/byterings/bgit/blob/main/install.ps1

$ErrorActionPreference = "Stop"

$GithubRepo = "byterings/bgit"
$InstallDir = "$env:LOCALAPPDATA\bgit"

Write-Host ""
Write-Host "Fetching latest version..." -ForegroundColor Cyan

try {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$GithubRepo/releases/latest" -UseBasicParsing
    $Version = $Release.tag_name -replace '^v', ''
} catch {
    Write-Host "Error: Could not fetch latest version" -ForegroundColor Red
    exit 1
}

# Detect architecture based on published Windows release artifacts
$Arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()) {
    "X64" { "amd64" }
    "Arm64" { "arm64" }
    default { $null }
}

if (-not $Arch) {
    Write-Host "Error: Unsupported Windows architecture. Supported: amd64, arm64" -ForegroundColor Red
    exit 1
}

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
    $NewPath = if ([string]::IsNullOrWhiteSpace($UserPath)) { $InstallDir } else { "$UserPath;$InstallDir" }
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
Write-Host "  bgit setup         # One-time setup (hooks + defaults)"
Write-Host "  bgit add           # Add your first identity"
Write-Host "  bgit use <alias>   # Switch identity"
Write-Host ""
Write-Host "Note: You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
Write-Host ""
