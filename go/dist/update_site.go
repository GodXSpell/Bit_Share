package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	version     = "1.0.0"
	releaseDir  = "releases"
	downloadDir = "download"
)

func main() {
	fmt.Println("BitShare Release Tool")
	fmt.Println("====================")

	// Create release directory if it doesn't exist
	if err := os.MkdirAll(releaseDir, 0755); err != nil {
		fmt.Printf("Error creating release directory: %v\n", err)
		os.Exit(1)
	}

	// Create download directory if it doesn't exist
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		fmt.Printf("Error creating download directory: %v\n", err)
		os.Exit(1)
	}

	// Build for all platforms
	buildAllPlatforms()

	// Create installation scripts
	createInstallationScripts()

	// Create packages
	createPackages()

	fmt.Println("\nAll done! Release packages created.")
	fmt.Println("You can now upload the releases to your distribution server.")
}

func buildAllPlatforms() {
	platforms := []struct {
		os   string
		arch string
	}{
		{"windows", "amd64"},
		{"darwin", "amd64"}, // macOS Intel
		{"darwin", "arm64"}, // macOS Apple Silicon
		{"linux", "amd64"},
		{"linux", "arm64"},
	}

	fmt.Println("\nBuilding BitShare for all platforms...")

	for _, platform := range platforms {
		binaryName := "bitshare"
		if platform.os == "windows" {
			binaryName += ".exe"
		}

		outputPath := filepath.Join(releaseDir, fmt.Sprintf("bitshare-%s-%s", platform.os, platform.arch), binaryName)

		// Create directory
		err := os.MkdirAll(filepath.Dir(outputPath), 0755)
		if err != nil {
			fmt.Printf("Error creating directory for %s-%s: %v\n", platform.os, platform.arch, err)
			continue
		}

		fmt.Printf("Building for %s/%s...\n", platform.os, platform.arch)
		cmd := exec.Command("go", "build", "-o", outputPath)
		cmd.Env = append(os.Environ(),
			"GOOS="+platform.os,
			"GOARCH="+platform.arch,
		)

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error building for %s/%s: %v\n", platform.os, platform.arch, err)
			fmt.Println(stderr.String())
			continue
		}

		fmt.Printf("Successfully built for %s/%s\n", platform.os, platform.arch)
	}
}

func createInstallationScripts() {
	fmt.Println("\nCreating installation scripts...")

	// Windows PowerShell script
	windowsScript := `
# BitShare Installation Script
$ErrorActionPreference = "Stop"
$InstallDir = "$env:LOCALAPPDATA\BitShare"
$BinaryPath = "$InstallDir\bitshare.exe"

# Create installation directory
Write-Host "Creating installation directory..."
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Copy files
Write-Host "Copying files..."
Copy-Item -Path ".\bitshare.exe" -Destination $BinaryPath -Force

# Add to PATH
$userPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)
if ($userPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$userPath;$InstallDir", [EnvironmentVariableTarget]::User)
    Write-Host "Added BitShare to user PATH."
}

Write-Host "BitShare has been installed successfully!" -ForegroundColor Green
Write-Host "You can now run 'bitshare' from any terminal."
`

	// Unix shell script
	unixScript := `#!/bin/bash
# BitShare Installation Script
set -e

# Determine installation directory
if [[ "$(uname)" == "Darwin" ]]; then
    # macOS
    INSTALL_DIR="$HOME/Applications/BitShare"
    BIN_DIR="/usr/local/bin"
else
    # Linux
    INSTALL_DIR="$HOME/.local/bin/bitshare"
    BIN_DIR="$HOME/.local/bin"
fi

BINARY_PATH="$INSTALL_DIR/bitshare"

echo "BitShare Installer"
echo "-----------------"

# Create installation directory
echo "Creating installation directory..."
mkdir -p "$INSTALL_DIR"

# Copy files
echo "Copying files..."
cp ./bitshare "$BINARY_PATH"
chmod +x "$BINARY_PATH"

# Create symlink
mkdir -p "$BIN_DIR"
if [[ -f "$BIN_DIR/bitshare" ]]; then
    echo "Removing old symlink..."
    rm "$BIN_DIR/bitshare"
fi

echo "Creating symlink..."
ln -s "$BINARY_PATH" "$BIN_DIR/bitshare"

# Check if BIN_DIR is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo ""
    echo "WARNING: $BIN_DIR is not in your PATH."
    echo "To run bitshare from any location, add this line to your ~/.bashrc, ~/.zshrc or similar:"
    echo "    export PATH=\"\$PATH:$BIN_DIR\""
fi

echo ""
echo "BitShare has been installed successfully!"
echo "You can now run 'bitshare' from the terminal."
`

	// Write Windows script
	windowsScriptPath := filepath.Join(releaseDir, "bitshare-windows-amd64", "install.ps1")
	if err := os.WriteFile(windowsScriptPath, []byte(windowsScript), 0644); err != nil {
		fmt.Printf("Error writing Windows installation script: %v\n", err)
	} else {
		fmt.Println("Created Windows installation script")
	}

	// Write Unix script
	unixScriptPath := filepath.Join(releaseDir, "bitshare-darwin-amd64", "install.sh")
	if err := os.WriteFile(unixScriptPath, []byte(unixScript), 0644); err != nil {
		fmt.Printf("Error writing macOS (Intel) installation script: %v\n", err)
	} else {
		fmt.Println("Created macOS (Intel) installation script")
	}

	unixScriptPath = filepath.Join(releaseDir, "bitshare-darwin-arm64", "install.sh")
	if err := os.WriteFile(unixScriptPath, []byte(unixScript), 0644); err != nil {
		fmt.Printf("Error writing macOS (Apple Silicon) installation script: %v\n", err)
	} else {
		fmt.Println("Created macOS (Apple Silicon) installation script")
	}

	unixScriptPath = filepath.Join(releaseDir, "bitshare-linux-amd64", "install.sh")
	if err := os.WriteFile(unixScriptPath, []byte(unixScript), 0644); err != nil {
		fmt.Printf("Error writing Linux installation script: %v\n", err)
	} else {
		fmt.Println("Created Linux installation script")
	}
}

func createPackages() {
	fmt.Println("\nCreating release packages...")

	// Create ZIP for Windows
	createZip("bitshare-windows-amd64")

	// Create ZIP for macOS
	createZip("bitshare-darwin-amd64")
	createZip("bitshare-darwin-arm64")

	// Create tar.gz for Linux
	createTarGz("bitshare-linux-amd64")
	createTarGz("bitshare-linux-arm64")
}

func createZip(directory string) {
	outputFile := filepath.Join(releaseDir, directory+".zip")
	fmt.Printf("Creating ZIP package for %s...\n", directory)

	// Create a buffer to write our ZIP to.
	buf := new(bytes.Buffer)

	// Create a new ZIP archive.
	w := zip.NewWriter(buf)

	// Walk the directory and add all files to the ZIP.
	srcDir := filepath.Join(releaseDir, directory)
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		f, err := w.Create(relPath)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = f.Write(data)
		return err
	})

	if err != nil {
		fmt.Printf("Error creating ZIP for %s: %v\n", directory, err)
		return
	}

	// Close the ZIP writer.
	if err := w.Close(); err != nil {
		fmt.Printf("Error closing ZIP writer for %s: %v\n", directory, err)
		return
	}

	// Write the ZIP file.
	if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		fmt.Printf("Error writing ZIP file for %s: %v\n", directory, err)
		return
	}

	// Copy the ZIP to the download directory
	downloadPath := filepath.Join(downloadDir, directory+".zip")
	if err := os.WriteFile(downloadPath, buf.Bytes(), 0644); err != nil {
		fmt.Printf("Error copying ZIP to download directory for %s: %v\n", directory, err)
		return
	}

	fmt.Printf("Created ZIP package for %s\n", directory)
}

func createTarGz(directory string) {
	outputFile := filepath.Join(releaseDir, directory+".tar.gz")
	fmt.Printf("Creating tar.gz package for %s...\n", directory)

	// For simplicity, we'll use the tar command
	if runtime.GOOS == "windows" {
		fmt.Printf("tar.gz creation not supported on Windows in this script\n")
		return
	}

	srcDir := filepath.Join(releaseDir, directory)
	cmd := exec.Command("tar", "-czf", outputFile, "-C", filepath.Dir(srcDir), filepath.Base(srcDir))

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error creating tar.gz for %s: %v\n", directory, err)
		fmt.Println(stderr.String())
		return
	}

	// Copy the tar.gz to the download directory
	downloadPath := filepath.Join(downloadDir, directory+".tar.gz")
	data, err := os.ReadFile(outputFile)
	if err != nil {
		fmt.Printf("Error reading tar.gz for %s: %v\n", directory, err)
		return
	}

	if err := os.WriteFile(downloadPath, data, 0644); err != nil {
		fmt.Printf("Error copying tar.gz to download directory for %s: %v\n", directory, err)
		return
	}

	fmt.Printf("Created tar.gz package for %s\n", directory)
}
