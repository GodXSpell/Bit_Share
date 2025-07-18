<#
.SYNOPSIS
    Installs BitShare as a system-wide application
.DESCRIPTION
    This script builds BitShare, copies it to a permanent location,
    and adds it to the system PATH so you can run it from any terminal.
#>

$ErrorActionPreference = "Stop"
$InstallDir = "$env:LOCALAPPDATA\BitShare"
$BinaryPath = "$InstallDir\bitshare.exe"

# Check for administrator rights
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Warning "For the best experience, run this script as administrator to add BitShare to the system PATH."
    $AddToPath = $false
} else {
    $AddToPath = $true
}

# Ensure Go is installed
try {
    $goVersion = go version
    Write-Host "Found Go: $goVersion"
} catch {
    Write-Error "Go is not installed or not in PATH. Please install Go first: https://golang.org/dl/"
    exit 1
}

# Create installation directory
Write-Host "Creating installation directory..."
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
} else {
    Write-Host "Installation directory already exists."
}

# Build BitShare
Write-Host "Building BitShare..."
go build -o "$BinaryPath" .
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to build BitShare."
    exit 1
}

# Add to PATH
if ($AddToPath) {
    Write-Host "Adding BitShare to PATH..."
    $userPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)
    if ($userPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("PATH", "$userPath;$InstallDir", [EnvironmentVariableTarget]::User)
        Write-Host "Added BitShare to user PATH."
    } else {
        Write-Host "BitShare is already in PATH."
    }
}

# Create shortcut in Start Menu
$StartMenuPath = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\BitShare"
if (!(Test-Path $StartMenuPath)) {
    New-Item -ItemType Directory -Path $StartMenuPath -Force | Out-Null
}

$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("$StartMenuPath\BitShare.lnk")
$Shortcut.TargetPath = $BinaryPath
$Shortcut.Save()

Write-Host "`nBitShare has been installed successfully!" -ForegroundColor Green
Write-Host "You can now run 'bitshare' from any terminal."
if (!$AddToPath) {
    Write-Host "`nManual PATH Setup:" -ForegroundColor Yellow
    Write-Host "To run bitshare from any location, add this to your PATH:"
    Write-Host $InstallDir
}
Write-Host "`nPress any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
