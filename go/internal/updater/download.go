package updater

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// DownloadInfo contains information about how to get BitShare
type DownloadInfo struct {
	WebsiteURL     string
	RepositoryURL  string
	DirectDownload string
}

// GetDownloadInfo returns information about how to download BitShare
func GetDownloadInfo() DownloadInfo {
	// This would be updated to your actual distribution URLs
	info := DownloadInfo{
		WebsiteURL:    "https://bitshare.yourdomain.com",
		RepositoryURL: "https://github.com/yourusername/bitshare",
	}

	// Set platform-specific direct download link
	switch runtime.GOOS {
	case "windows":
		info.DirectDownload = "https://github.com/yourusername/bitshare/releases/latest/download/bitshare-windows-amd64.zip"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			info.DirectDownload = "https://github.com/yourusername/bitshare/releases/latest/download/bitshare-darwin-arm64.zip"
		} else {
			info.DirectDownload = "https://github.com/yourusername/bitshare/releases/latest/download/bitshare-darwin-amd64.zip"
		}
	default:
		info.DirectDownload = "https://github.com/yourusername/bitshare/releases/latest/download/bitshare-linux-amd64.tar.gz"
	}

	return info
}

// OpenDownloadPage opens the BitShare download page in the default browser
func OpenDownloadPage() error {
	info := GetDownloadInfo()
	url := info.WebsiteURL + "/download"

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

// GetDownloadCommand returns a command that users can use to download BitShare
func GetDownloadCommand() string {
	info := GetDownloadInfo()

	switch {
	case isCommandAvailable("wget"):
		return fmt.Sprintf("wget %s -O bitshare.zip", info.DirectDownload)
	case isCommandAvailable("curl"):
		return fmt.Sprintf("curl -L %s -o bitshare.zip", info.DirectDownload)
	default:
		return fmt.Sprintf("Please download from: %s", info.WebsiteURL)
	}
}

// Helper function to check if a command is available
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// ShowDownloadInstructions prints download instructions to stdout
func ShowDownloadInstructions() {
	info := GetDownloadInfo()
	fmt.Println("\n" + colorGreen + "BitShare Download Instructions" + colorReset)
	fmt.Println("================================")

	fmt.Println("\nDownload Links:")
	fmt.Printf("- Website: %s\n", info.WebsiteURL)
	fmt.Printf("- Repository: %s\n", info.RepositoryURL)

	fmt.Println("\nDownload Methods:")

	fmt.Println("\n1. Using Web Browser:")
	fmt.Printf("   Visit %s/download and choose your platform\n", info.WebsiteURL)

	fmt.Println("\n2. Using Command Line:")
	fmt.Printf("   %s\n", GetDownloadCommand())

	fmt.Println("\n3. Using Package Manager:")

	if runtime.GOOS == "darwin" {
		fmt.Println("   Using Homebrew:")
		fmt.Println("   brew tap yourusername/bitshare")
		fmt.Println("   brew install bitshare")
	} else if runtime.GOOS == "windows" {
		fmt.Println("   Using Scoop:")
		fmt.Println("   scoop bucket add bitshare https://github.com/yourusername/scoop-bitshare")
		fmt.Println("   scoop install bitshare")
	} else {
		fmt.Println("   Using your Linux package manager (if available):")
		fmt.Println("   Check your distribution's documentation")
	}

	fmt.Println("\nAfter downloading, follow the installation instructions included in the package.")
}

// colorGreen and colorReset variables should be defined elsewhere in the package
var (
	colorGreen = "\033[1;32m"
	colorReset = "\033[0m"
)

func init() {
	// Check if terminal supports colors and disable if not
	if !supportsColors() {
		colorGreen = ""
		colorReset = ""
	}
}

// supportsColors checks if the terminal supports ANSI color codes
func supportsColors() bool {
	// On Windows, check if running in modern terminals that support colors
	if os.Getenv("TERM") == "" && runtime.GOOS == "windows" {
		// Check for newer Windows terminals (Windows Terminal, VSCode, etc.)
		if os.Getenv("WT_SESSION") != "" || os.Getenv("TERM_PROGRAM") != "" {
			return true
		}
		return false
	}
	return true
}
