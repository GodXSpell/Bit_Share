package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fileshare/internal/mesh"
)

// TerminalUI manages terminal-based user interface
type TerminalUI struct {
	isRunning    bool
	refreshRate  time.Duration
	width        int
	height       int
	activeScreen string
	mutex        sync.RWMutex
}

// TransferProgress tracks file transfer progress
type TransferProgress struct {
	FileName      string
	FileSize      int64
	BytesComplete int64
	StartTime     time.Time
	Status        string // "pending", "transferring", "complete", "failed"
	SpeedBps      int64  // Bytes per second
	Error         error
}

var (
	termUI *TerminalUI
	uiOnce sync.Once
)

// GetTerminalUI returns the singleton instance of TerminalUI
func GetTerminalUI() *TerminalUI {
	uiOnce.Do(func() {
		termUI = &TerminalUI{
			isRunning:    false,
			refreshRate:  time.Second,
			width:        80,
			height:       24,
			activeScreen: "dashboard",
		}
	})
	return termUI
}

// Start initializes and starts the terminal UI
func (ui *TerminalUI) Start() error {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	if ui.isRunning {
		return fmt.Errorf("terminal UI is already running")
	}

	// Initialize terminal
	width, height, err := getTerminalSize()
	if err == nil {
		ui.width = width
		ui.height = height
	}

	ui.isRunning = true

	// Start UI refresh goroutine
	go ui.refreshLoop()

	return nil
}

// Stop stops the terminal UI
func (ui *TerminalUI) Stop() {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.isRunning = false
}

// ShowDashboard displays the main dashboard
func (ui *TerminalUI) ShowDashboard() {
	ui.mutex.Lock()
	ui.activeScreen = "dashboard"
	ui.mutex.Unlock()
}

// ShowPeerList displays the list of peers
func (ui *TerminalUI) ShowPeerList() {
	ui.mutex.Lock()
	ui.activeScreen = "peers"
	ui.mutex.Unlock()
}

// ShowTransferStatus displays file transfer status
func (ui *TerminalUI) ShowTransferStatus() {
	ui.mutex.Lock()
	ui.activeScreen = "transfers"
	ui.mutex.Unlock()
}

// ShowNetworkMap displays the mesh network topology
func (ui *TerminalUI) ShowNetworkMap() {
	ui.mutex.Lock()
	ui.activeScreen = "network"
	ui.mutex.Unlock()
}

// UpdateTransferProgress updates the progress of a file transfer
func (ui *TerminalUI) UpdateTransferProgress(progress TransferProgress) {
	// In a full implementation, this would update the UI with the progress
	// For this placeholder, just print progress to console
	percentComplete := float64(progress.BytesComplete) / float64(progress.FileSize) * 100

	if progress.Status == "transferring" {
		elapsed := time.Since(progress.StartTime).Seconds()
		if elapsed > 0 {
			progress.SpeedBps = int64(float64(progress.BytesComplete) / elapsed)
		}
	}

	speedMBps := float64(progress.SpeedBps) / (1024 * 1024)

	fmt.Printf("\rTransfer: %s - %.1f%% complete (%.2f MB/s)",
		progress.FileName, percentComplete, speedMBps)
}

// Helper methods
func (ui *TerminalUI) refreshLoop() {
	ticker := time.NewTicker(ui.refreshRate)
	defer ticker.Stop()

	for range ticker.C {
		ui.mutex.RLock()
		if !ui.isRunning {
			ui.mutex.RUnlock()
			return
		}

		screen := ui.activeScreen
		width := ui.width
		height := ui.height
		ui.mutex.RUnlock()

		// Clear screen and render active view
		clearScreen()

		switch screen {
		case "dashboard":
			ui.renderDashboard(width, height)
		case "peers":
			ui.renderPeerList(width, height)
		case "transfers":
			ui.renderTransferStatus(width, height)
		case "network":
			ui.renderNetworkMap(width, height)
		}
	}
}

func (ui *TerminalUI) renderDashboard(width, height int) {
	// Render dashboard layout
	title := "BitShare P2P Mesh Network"
	header := centerText(title, width)
	divider := strings.Repeat("=", width)

	fmt.Println(header)
	fmt.Println(divider)

	// Node status
	nodeInfo := "Node: Unknown"
	fmt.Println(nodeInfo)
	fmt.Println()

	// Network summary
	fmt.Println("Network Status:")
	fmt.Println("  Peers: 0 online, 0 total")
	fmt.Println("  Protocols: WiFi Direct, TCP/IP, Bluetooth")
	fmt.Println()

	// Transfer summary
	fmt.Println("Transfers:")
	fmt.Println("  Active: 0, Queued: 0, Completed: 0")
	fmt.Println()

	// Commands
	fmt.Println(divider)
	fmt.Println("Commands: [P]eers | [T]ransfers | [N]etwork | [Q]uit")
}

func (ui *TerminalUI) renderPeerList(width, height int) {
	// Render peer list
	title := "Connected Peers"
	header := centerText(title, width)
	divider := strings.Repeat("=", width)

	fmt.Println(header)
	fmt.Println(divider)

	// Get peer list
	peers, err := mesh.GetKnownPeers()
	if err != nil {
		fmt.Println("Error retrieving peers:", err)
		return
	}

	if len(peers) == 0 {
		fmt.Println("No peers connected.")
		fmt.Println("Run 'scan' to discover nearby peers.")
	} else {
		// Table header
		fmt.Printf("%-20s %-15s %-10s %-8s\n", "NAME", "ID", "PROTOCOL", "STATUS")
		fmt.Println(strings.Repeat("-", width))

		// Peer rows
		for _, peer := range peers {
			status := "Offline"
			if peer.IsOnline {
				status = "Online"
			}
			fmt.Printf("%-20s %-15s %-10s %-8s\n",
				truncateString(peer.Name, 20),
				truncateString(peer.ID, 15),
				"", // Protocol not in mesh.Peer
				status)
		}
	}

	fmt.Println()
	fmt.Println(divider)
	fmt.Println("Commands: [D]ashboard | [T]ransfers | [N]etwork | [Q]uit")
}

func (ui *TerminalUI) renderTransferStatus(width, height int) {
	// Render transfer status
	title := "File Transfers"
	header := centerText(title, width)
	divider := strings.Repeat("=", width)

	fmt.Println(header)
	fmt.Println(divider)

	// This would show active and recent transfers
	fmt.Println("No active transfers.")

	fmt.Println()
	fmt.Println(divider)
	fmt.Println("Commands: [D]ashboard | [P]eers | [N]etwork | [Q]uit")
}

func (ui *TerminalUI) renderNetworkMap(width, height int) {
	// Render network map/topology
	title := "Mesh Network Map"
	header := centerText(title, width)
	divider := strings.Repeat("=", width)

	fmt.Println(header)
	fmt.Println(divider)

	// This would show a text-based visualization of the network topology
	fmt.Println("Network visualization not available in terminal mode.")
	fmt.Println("Use 'list' command to see connected peers.")

	fmt.Println()
	fmt.Println(divider)
	fmt.Println("Commands: [D]ashboard | [P]eers | [T]ransfers | [Q]uit")
}

// Helper functions
func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}

	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func clearScreen() {
	// ANSI escape sequence to clear screen and move cursor to top-left
	fmt.Print("\033[2J\033[H")
}

func getTerminalSize() (int, int, error) {
	// In a real implementation, this would use platform-specific APIs
	// to determine terminal size

	// For this placeholder, return default values
	return 80, 24, nil
}
