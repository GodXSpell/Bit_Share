package utils

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// GetAllLocalIPs returns a slice of all non-loopback local IP addresses.
func GetAllLocalIPs() ([]string, error) {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			ips = append(ips, ip.String())
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no network interfaces found")
	}
	return ips, nil
}

// FileExists checks if a file exists and is not a directory.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// FormatBytes converts a number of bytes into a human-readable string.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// FindFileInCommonDirs searches for a file in standard user directories (Desktop, Documents, Downloads).
func FindFileInCommonDirs(filename string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find user home directory: %v", err)
	}

	searchDirs := []string{
		"Downloads",
		"Documents",
		"Desktop",
	}

	for _, dir := range searchDirs {
		searchPath := filepath.Join(homeDir, dir, filename)
		if FileExists(searchPath) {
			return searchPath, nil
		}
	}

	return "", fmt.Errorf("file not found in Desktop, Documents, or Downloads")
}

// GenerateNodeName creates a friendly name for this node
func GenerateNodeName() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-device"
	}

	// Clean up hostname and add a suffix for uniqueness
	cleanName := strings.Replace(hostname, ".", "-", -1)
	cleanName = strings.Replace(cleanName, " ", "-", -1)

	return cleanName
}
