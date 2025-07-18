package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// Update check frequency
	UpdateCheckInterval = 24 * time.Hour

	// GitHub release API URL - replace with your actual repository
	ReleaseURL = "https://api.github.com/repos/yourusername/bitshare/releases/latest"

	// Current version
	Version = "1.0.0"
)

var (
	// Path to the update settings file
	settingsPath string
)

// UpdateSettings stores user preferences for updates
type UpdateSettings struct {
	LastCheck       time.Time `json:"last_check"`
	AutoUpdate      bool      `json:"auto_update"`
	UpdateAvailable bool      `json:"update_available"`
	NewVersion      string    `json:"new_version"`
	DownloadURL     string    `json:"download_url"`
}

// ReleaseInfo stores information about a GitHub release
type ReleaseInfo struct {
	TagName     string      `json:"tag_name"`
	Name        string      `json:"name"`
	Body        string      `json:"body"`
	PublishedAt time.Time   `json:"published_at"`
	Assets      []AssetInfo `json:"assets"`
}

// AssetInfo represents a release asset (downloadable file)
type AssetInfo struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int    `json:"size"`
}

func init() {
	// Set up settings path in the user's config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	settingsPath = filepath.Join(configDir, "BitShare", "update.json")

	// Create directory if it doesn't exist
	os.MkdirAll(filepath.Dir(settingsPath), 0755)

	// Load or create default settings
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		saveSettings(&UpdateSettings{
			LastCheck:  time.Time{},
			AutoUpdate: false,
		})
	}
}

// CheckForUpdates checks if an update is available
func CheckForUpdates(force bool) (*UpdateSettings, bool, error) {
	settings, err := loadSettings()
	if err != nil {
		return nil, false, err
	}

	// Check if we need to check for updates
	if !force && time.Since(settings.LastCheck) < UpdateCheckInterval {
		return settings, settings.UpdateAvailable, nil
	}

	// Update last check time
	settings.LastCheck = time.Now()

	// Check for updates
	release, err := getLatestRelease()
	if err != nil {
		return settings, false, err
	}

	// Check if version is newer
	newVersion := strings.TrimPrefix(release.TagName, "v")
	if isNewer(newVersion, Version) {
		settings.UpdateAvailable = true
		settings.NewVersion = newVersion

		// Find the appropriate download for this platform
		downloadURL := findDownloadURL(release)
		if downloadURL != "" {
			settings.DownloadURL = downloadURL
		}
	} else {
		settings.UpdateAvailable = false
	}

	// Save settings
	err = saveSettings(settings)
	return settings, settings.UpdateAvailable, err
}

// InstallUpdate downloads and installs the latest version
func InstallUpdate() error {
	settings, err := loadSettings()
	if err != nil {
		return err
	}

	if !settings.UpdateAvailable {
		return fmt.Errorf("no updates available")
	}

	if settings.DownloadURL == "" {
		return fmt.Errorf("no download URL available for this platform")
	}

	// Download the update
	fmt.Println("Downloading update...")
	downloadPath := filepath.Join(os.TempDir(), "bitshare-update.zip")
	err = downloadFile(settings.DownloadURL, downloadPath)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Extract and install the update
	fmt.Println("Installing update...")

	// Platform specific installation
	switch runtime.GOOS {
	case "windows":
		return installUpdateWindows(downloadPath)
	default:
		return installUpdatePosix(downloadPath)
	}
}

// EnableAutoUpdate enables or disables automatic updates
func EnableAutoUpdate(enable bool) error {
	settings, err := loadSettings()
	if err != nil {
		return err
	}

	settings.AutoUpdate = enable
	return saveSettings(settings)
}

// ShouldAutoUpdate returns true if automatic updates are enabled
func ShouldAutoUpdate() (bool, error) {
	settings, err := loadSettings()
	if err != nil {
		return false, err
	}

	return settings.AutoUpdate, nil
}

// Helper functions
func loadSettings() (*UpdateSettings, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return &UpdateSettings{}, err
	}

	var settings UpdateSettings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		return &UpdateSettings{}, err
	}

	return &settings, nil
}

func saveSettings(settings *UpdateSettings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsPath, data, 0644)
}

func getLatestRelease() (*ReleaseInfo, error) {
	resp, err := http.Get(ReleaseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release ReleaseInfo
	err = json.Unmarshal(body, &release)
	if err != nil {
		return nil, err
	}

	return &release, nil
}

func isNewer(version, currentVersion string) bool {
	// Simple version comparison - in a real implementation, use proper semver comparison
	return version > currentVersion
}

func findDownloadURL(release *ReleaseInfo) string {
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Look for matching asset
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, platform) && strings.Contains(name, arch) {
			return asset.DownloadURL
		}
	}

	// Fallback to platform-only match
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, platform) {
			return asset.DownloadURL
		}
	}

	return ""
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// Platform-specific update installation
func installUpdateWindows(downloadPath string) error {
	// For Windows, we'll use PowerShell commands to extract and install
	// This is a placeholder implementation
	fmt.Println("Windows update not fully implemented")
	return nil
}

func installUpdatePosix(downloadPath string) error {
	// For Linux/macOS, we'll use shell commands to extract and install
	// This is a placeholder implementation
	fmt.Println("Unix update not fully implemented")
	return nil
}
