package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"fileshare/internal/firewall"
	"fileshare/internal/mesh"
	"fileshare/internal/p2p"
	"fileshare/internal/transfer"
	"fileshare/internal/updater"
	"fileshare/internal/utils"
)

// Create a startup file and add a helper to check if the terminal supports colors
func init() {
	// Check if terminal supports colors and disable if not
	if !supportsColors() {
		// Replace color codes with empty strings
		colorReset = ""
		colorGreen = ""
		colorBlue = ""
		colorCyan = ""
	}

	// Check for updates if auto-update is enabled
	go func() {
		autoUpdate, _ := updater.ShouldAutoUpdate()
		if autoUpdate {
			settings, updateAvailable, _ := updater.CheckForUpdates(false)
			if updateAvailable {
				fmt.Printf("\nA new version of BitShare is available: %s\n", settings.NewVersion)
				fmt.Println("Run 'bitshare update install' to update")
			}
		}
	}()
}

// Constants for terminal colors
var (
	colorReset = "\033[0m"
	colorGreen = "\033[1;32m"
	colorBlue  = "\033[1;34m"
	colorCyan  = "\033[1;36m"
)

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

func main() {
	// If no arguments are provided, start interactive mode by default
	if len(os.Args) == 1 {
		startInteractiveMode()
		return
	}

	command := os.Args[1]

	// Handle special case for interactive mode
	if command == "interactive" || command == "shell" || command == "terminal" {
		startInteractiveMode()
		return
	}

	executeCommand(os.Args[1:])
}

// startInteractiveMode launches BitShare as an interactive terminal application
func startInteractiveMode() {
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Handle Ctrl+C gracefully
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Exiting BitShare terminal...")
		os.Exit(0)
	}()

	// Start mesh node in background
	config := mesh.Config{
		NodeName:         utils.GenerateNodeName(),
		ListenPort:       9000, // Default port
		EnableWiFiDirect: true,
		EnableBluetooth:  true,
		EnableTCP:        true,
		EnableRelay:      true, // Enable relay by default in interactive mode
	}

	fmt.Println("🌐 Starting BitShare in interactive mode...")
	err := mesh.StartMeshNode(config)
	if err != nil {
		fmt.Printf("❌ Warning: Failed to start mesh node: %v\n", err)
		fmt.Println("Some functionality may be limited.")
	} else {
		fmt.Printf("✅ Node started successfully as '%s'\n", config.NodeName)
	}

	// Display welcome message and instructions
	displayWelcomeMessage()

	// Start the command prompt loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\033[1;36mbitshare> \033[0m") // Cyan prompt
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command:", err)
			continue
		}

		// Process the command
		cmdString = strings.TrimSpace(cmdString)
		if cmdString == "" {
			continue
		}

		// Handle built-in commands
		if handleBuiltinCommand(cmdString) {
			continue
		}

		// Parse the command string into arguments
		args := parseCommand(cmdString)
		if len(args) == 0 {
			continue
		}

		// Execute the command
		executeCommand(args)
	}
}

// parseCommand splits a command string into arguments, respecting quoted strings
func parseCommand(cmd string) []string {
	var args []string
	var currentArg strings.Builder
	inQuotes := false
	escapeNext := false

	for _, char := range cmd {
		if escapeNext {
			currentArg.WriteRune(char)
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' {
			inQuotes = !inQuotes
			continue
		}

		if char == ' ' && !inQuotes {
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
			continue
		}

		currentArg.WriteRune(char)
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return args
}

// handleBuiltinCommand processes special built-in commands
func handleBuiltinCommand(cmd string) bool {
	cmd = strings.ToLower(cmd)

	switch cmd {
	case "exit", "quit", "bye":
		fmt.Println("Exiting BitShare terminal. Goodbye!")
		// Stop mesh node if running
		mesh.StopMeshNode()
		os.Exit(0)
		return true

	case "clear", "cls":
		// Clear the screen
		fmt.Print("\033[H\033[2J") // ANSI escape sequence to clear screen
		return true

	case "help":
		printInteractiveHelp()
		return true

	case "status":
		printNodeStatus()
		return true
	}

	return false
}

// executeCommand runs a BitShare command with the given arguments
func executeCommand(args []string) {
	if len(args) == 0 {
		return
	}

	command := args[0]

	switch command {
	case "start":
		startMeshNode()

	case "scan":
		scanNetwork()

	case "list":
		listPeers()

	case "install", "--install":
		showInstallationInfo()

	case "receive":
		if len(args) < 2 || len(args) > 3 {
			fmt.Println("Usage: receive <port_no> [destination_directory]")
			return
		}
		port, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("Invalid port number: %v\n", err)
			return
		}
		if port < 1 || port > 65535 {
			fmt.Println("Port number must be between 1 and 65535")
			return
		}
		destDir := "." // Default to current directory
		if len(args) == 3 {
			destDir = args[2]
		}

		// Start receiver in non-blocking mode
		go func() {
			startReceiver(port, destDir)
		}()
		fmt.Printf("Receiver started on port %d. Files will be saved to %s\n", port, destDir)
		fmt.Println("You can continue using other commands while receiving.")

	case "send":
		if len(args) != 4 {
			fmt.Println("Usage: send <peer_id_or_ip> <port_no> <file_path>")
			return
		}
		ip := args[1]

		port, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Printf("Invalid port number: %v\n", err)
			return
		}

		filePath := args[3]

		// Start sender in a goroutine so it doesn't block the terminal
		go func() {
			if net.ParseIP(ip) == nil {
				// This might be a peer ID or name, try to resolve it
				fmt.Printf("Looking up peer: %s\n", ip)
				peer, err := mesh.FindPeerByIdOrName(ip)
				if err != nil {
					fmt.Printf("Error finding peer: %v\n", err)
					return
				}

				if peer != nil {
					fmt.Printf("Found peer %s (%s)\n", peer.Name, peer.ID)
					// Use the peer's address
					if len(peer.Routes) > 0 {
						bestRoute := findBestRoute(peer.Routes)
						fmt.Printf("Using route via: %s\n", bestRoute.NextHop)
						ip = bestRoute.NextHop
					} else if peer.Address != "" {
						ip = peer.Address
						fmt.Printf("Using direct connection to: %s\n", ip)
					} else {
						fmt.Printf("Error: Peer found but no address information available\n")
						return
					}
				} else {
					fmt.Printf("Error: Could not find peer named '%s'\n", ip)
					fmt.Println("Run 'scan' to discover available peers.")
					return
				}
			}

			// Now we have a valid IP to connect to
			if !utils.FileExists(filePath) {
				fmt.Printf("File not found at '%s'. Searching in common directories...\n", filePath)
				foundPath, err := utils.FindFileInCommonDirs(filePath)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					absPath, _ := filepath.Abs(filePath)
					fmt.Printf("Looked for file at: %s\n", absPath)
					return
				}
				fmt.Printf("File found: %s\n", foundPath)
				filePath = foundPath
			}

			// Check if file is readable
			file, err := os.Open(filePath)
			if err != nil {
				fmt.Printf("Cannot read file: %v\n", err)
				return
			}
			file.Close()

			fmt.Printf("Sending %s to %s:%d...\n", filepath.Base(filePath), ip, port)
			err = transfer.SendFile(filePath, ip, port)
			if err != nil {
				fmt.Printf("Error sending file: %v\n", err)
				return
			}

			fmt.Println("File sent successfully!")
		}()
		fmt.Println("Transfer started in background. You can continue using other commands.")

	case "help":
		printInteractiveHelp()

	case "update":
		if len(args) < 2 {
			fmt.Println("Usage: update <check|install|auto>")
			fmt.Println("  - check: Check for available updates")
			fmt.Println("  - install: Install the latest update")
			fmt.Println("  - auto: Configure automatic updates")
			return
		}

		updateSubcommand := args[1]
		switch updateSubcommand {
		case "check":
			settings, updateAvailable, err := updater.CheckForUpdates(true)
			if err != nil {
				fmt.Printf("Error checking for updates: %v\n", err)
				return
			}

			if updateAvailable {
				fmt.Printf("A new version is available: %s\n", settings.NewVersion)
				fmt.Println("Run 'bitshare update install' to update")
			} else {
				fmt.Println("You are running the latest version!")
			}

		case "install":
			err := updater.InstallUpdate()
			if err != nil {
				fmt.Printf("Error installing update: %v\n", err)
			} else {
				fmt.Println("Update installed successfully! Please restart BitShare.")
			}

		case "auto":
			if len(args) < 3 || (args[2] != "--enable" && args[2] != "--disable") {
				fmt.Println("Usage: update auto --enable|--disable")
				return
			}

			enable := args[2] == "--enable"
			err := updater.EnableAutoUpdate(enable)
			if err != nil {
				fmt.Printf("Error configuring auto-updates: %v\n", err)
				return
			}

			if enable {
				fmt.Println("Automatic updates enabled")
			} else {
				fmt.Println("Automatic updates disabled")
			}

		default:
			fmt.Printf("Unknown update subcommand: %s\n", updateSubcommand)
			fmt.Println("Valid subcommands: check, install, auto")
		}

	case "download":
		updater.ShowDownloadInstructions()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Type 'help' for a list of commands.")
	}
}

// displayWelcomeMessage shows initial welcome information
func displayWelcomeMessage() {
	fmt.Println("\033[1;32m===============================================\033[0m")
	fmt.Println("\033[1;32m         Welcome to BitShare Terminal         \033[0m")
	fmt.Println("\033[1;32m===============================================\033[0m")
	fmt.Println("\033[0;36mBitShare P2P Mesh Network File Transfer\033[0m")
	fmt.Println("\nMesh node is running in the background. Type commands directly at the prompt.")
	fmt.Println("\nCommon commands:")
	fmt.Println("  \033[1mscan\033[0m           - Scan for nearby peers")
	fmt.Println("  \033[1mlist\033[0m           - List known peers")
	fmt.Println("  \033[1mreceive <port>\033[0m - Start receiving files")
	fmt.Println("  \033[1msend <peer> <port> <file>\033[0m - Send a file")
	fmt.Println("  \033[1mhelp\033[0m           - Show more commands")
	fmt.Println("  \033[1mquit\033[0m           - Exit BitShare")
	fmt.Println("\033[0;36mType commands directly at the prompt below:\033[0m")
}

// printInteractiveHelp shows the help information in interactive mode
func printInteractiveHelp() {
	fmt.Println("\n\033[1mBitShare Terminal Commands:\033[0m")
	fmt.Println("\n\033[1;34mCore Commands:\033[0m")
	fmt.Println("  \033[1mscan\033[0m                    - Scan for nearby peers")
	fmt.Println("  \033[1mlist\033[0m                    - List known peers in the network")
	fmt.Println("  \033[1mreceive <port> [dir]\033[0m    - Start receiving files on specified port")
	fmt.Println("  \033[1msend <peer> <port> <file>\033[0m - Send a file to a peer")

	fmt.Println("\n\033[1;34mNetwork Commands:\033[0m")
	fmt.Println("  \033[1mstart\033[0m                   - Restart the mesh network node")
	fmt.Println("  \033[1mstatus\033[0m                  - Show current node and network status")

	fmt.Println("\n\033[1;34mTerminal Commands:\033[0m")
	fmt.Println("  \033[1mhelp\033[0m                    - Show this help information")
	fmt.Println("  \033[1mclear\033[0m                   - Clear the terminal screen")
	fmt.Println("  \033[1mquit\033[0m, \033[1mexit\033[0m, \033[1mbye\033[0m        - Exit BitShare")

	fmt.Println("\n\033[1;34mInstallation and Updates:\033[0m")
	fmt.Println("  \033[1minstall\033[0m                  - Show installation instructions")
	fmt.Println("  \033[1mdownload\033[0m                 - Show download instructions")
	fmt.Println("  \033[1mupdate check\033[0m             - Check for updates")
	fmt.Println("  \033[1mupdate install\033[0m           - Install available updates")

	fmt.Println("\n\033[1mExamples:\033[0m")
	fmt.Println("  scan")
	fmt.Println("  receive 9000 C:\\Downloads")
	fmt.Println("  send bob-laptop 9000 report.pdf")
	fmt.Println("  send 192.168.1.10 9000 \"My Document.docx\"")
}

// printNodeStatus shows the current status of the mesh node
func printNodeStatus() {
	fmt.Println("\n\033[1mBitShare Node Status:\033[0m")

	// Check if mesh node is running
	isRunning := mesh.IsNodeRunning()
	if isRunning {
		fmt.Println("  Mesh Node: \033[1;32mRunning\033[0m")

		// Get node info
		nodeName := mesh.GetNodeName()
		nodeID := mesh.GetNodeID()
		fmt.Printf("  Node Name: %s\n", nodeName)
		fmt.Printf("  Node ID: %s\n", nodeID)

		// Get connection info
		connInfo := mesh.GetConnectionInfo()
		fmt.Printf("  Network Mode: %s\n", getNetworkModeString(connInfo.Mode))
		fmt.Printf("  Client Isolation: %v\n", connInfo.ClientIsolation)

		// Get peer count
		peers, _ := mesh.GetKnownPeers()
		onlinePeers := 0
		for _, peer := range peers {
			if peer.IsOnline {
				onlinePeers++
			}
		}
		fmt.Printf("  Peers: %d online, %d total\n", onlinePeers, len(peers))

	} else {
		fmt.Println("  Mesh Node: \033[1;31mNot Running\033[0m")
		fmt.Println("  Type 'start' to start the mesh node")
	}
}

// getNetworkModeString converts the network mode enum to a human-readable string
func getNetworkModeString(mode mesh.NetworkMode) string {
	switch mode {
	case mesh.DirectMode:
		return "Direct (P2P connections available)"
	case mesh.RelayMode:
		return "Relay (Using relay servers)"
	case mesh.MixedMode:
		return "Mixed (Some direct, some relayed)"
	default:
		return "Unknown"
	}
}

// startReceiver starts a file receiver on the given port and directory
func startReceiver(port int, destDir string) {
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// On Windows, firewall rules are often necessary. We will always try to add one.
	rule, err := firewall.AddTempRule(port)
	if err != nil {
		fmt.Printf("⚠️  Firewall rule not added: %v\n", err)
		fmt.Printf("💡 If connection fails, manually allow port %d or run as administrator\n", port)
	} else {
		fmt.Printf("✓ Temporary firewall rule added for port %d\n", port)
		// Ensure rule is removed on exit
		defer func() {
			if err := rule.RemoveRule(); err != nil {
				fmt.Printf("⚠️  Could not remove firewall rule: %v\n", err)
			} else {
				fmt.Printf("✓ Firewall rule removed\n")
			}
		}()
	}

	// Handle graceful shutdown
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down receiver...")
		// The deferred cleanup will run when os.Exit is called.
		os.Exit(0)
	}()

	// Get local IP for user information
	localIPs, err := utils.GetAllLocalIPs()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not determine local IP addresses: %v\n", err)
	}

	fmt.Printf("📡 Receiver: Listening on port %d\n", port)
	if len(localIPs) > 0 {
		fmt.Println("🌐 Your IP addresses are:")
		for _, ip := range localIPs {
			fmt.Printf("  - %s\n", ip)
			if strings.HasPrefix(ip, "169.254.") {
				fmt.Println("  ⚠️  Warning: This IP looks like an APIPA address. Your computer may not be connected to the network correctly. Please check your network connection.")
			}
		}
		fmt.Printf("🔗 Others can connect to: %s:%d\n", localIPs[0], port)
	}
	fmt.Printf("💾 Files will be saved to: %s\n", destDir)
	fmt.Printf("Press Ctrl+C to stop\n")

	// Set connection timeout for security (increased for larger files)
	err = transfer.ReceiveFileWithTimeout(port, 300*time.Second, destDir)
	if err != nil {
		fmt.Printf("Error receiving file: %v\n", err)
	}
}

// startSender initiates a file transfer to the given IP and port
func startSender(ip string, port int, filePath string) {
	// Remove quotes if present (useful for drag-and-drop)
	if strings.HasPrefix(filePath, "\"") && strings.HasSuffix(filePath, "\"") {
		filePath = filePath[1 : len(filePath)-1]
	}

	if !utils.FileExists(filePath) {
		fmt.Printf("File not found at '%s'. Searching in common directories...\n", filePath)
		foundPath, err := utils.FindFileInCommonDirs(filePath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			absPath, _ := filepath.Abs(filePath)
			fmt.Printf("Looked for file at: %s\n", absPath)
			fmt.Println("Hint: If your path contains spaces, make sure to wrap it in quotes.")
			return
		}
		fmt.Printf("File found: %s\n", foundPath)
		filePath = foundPath
	}

	// Check if file is readable
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Cannot read file: %v\n", err)
		return
	}
	file.Close()

	err = transfer.SendFile(filePath, ip, port)
	if err != nil {
		fmt.Printf("Error sending file: %v\n", err)
		return
	}

	fmt.Println("File sent successfully!")
}

// startMeshNode starts or restarts the mesh network node
func startMeshNode() {
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Handle graceful shutdown
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down mesh node...")
		mesh.StopMeshNode()
		os.Exit(0)
	}()

	// Initialize mesh networking
	config := mesh.Config{
		NodeName:         utils.GenerateNodeName(),
		ListenPort:       9000, // Default port
		EnableWiFiDirect: true,
		EnableBluetooth:  true,
		EnableTCP:        true,
		EnableRelay:      true, // Enable relay by default
	}

	fmt.Println("🌐 Starting BitShare mesh node...")
	err := mesh.StartMeshNode(config)
	if err != nil {
		fmt.Printf("❌ Failed to start mesh node: %v\n", err)
		return
	}

	fmt.Printf("✅ Node started successfully as '%s'\n", config.NodeName)
	fmt.Println("📡 Listening for connections...")
	fmt.Println("Press Ctrl+C to stop")

	// Block until signal received
	select {}
}

// scanNetwork scans the local network for peers
func scanNetwork() {
	fmt.Println("🔍 Scanning for nearby peers...")

	// Scan across all available protocols
	peers, err := p2p.ScanForPeers()
	if err != nil {
		fmt.Printf("❌ Scan error: %v\n", err)
		return
	}

	if len(peers) == 0 {
		fmt.Println("No peers found. Try again or check connection settings.")
		return
	}

	fmt.Printf("Found %d peers:\n", len(peers))
	for i, peer := range peers {
		fmt.Printf("%d. %s (%s) - Protocol: %s, Signal: %d%%\n",
			i+1, peer.Name, peer.ID, peer.Protocol, peer.SignalStrength)
	}
}

// listPeers lists all known peers in the mesh network
func listPeers() {
	peers, err := mesh.GetKnownPeers()
	if err != nil {
		fmt.Printf("❌ Error retrieving peers: %v\n", err)
		return
	}

	if len(peers) == 0 {
		fmt.Println("No known peers. Run 'scan' to discover peers.")
		return
	}

	fmt.Println("Known peers in the mesh network:")
	fmt.Println("--------------------------------")
	for i, peer := range peers {
		status := "⚫ Offline"
		if peer.IsOnline {
			status = "🟢 Online"
		}
		fmt.Printf("%d. %s (%s) - %s\n", i+1, peer.Name, peer.ID, status)
		fmt.Printf("   Routes: %d, Connection Quality: %s\n",
			len(peer.Routes), peer.ConnectionQuality)
	}
}

// Helper function to find the best route to a peer
func findBestRoute(routes []mesh.Route) mesh.Route {
	// Start with the first route
	bestRoute := routes[0]

	// Look for direct routes first (hopCount = 1)
	for _, route := range routes {
		if route.HopCount == 1 {
			return route
		}
	}

	// If no direct routes, find the highest quality route
	for _, route := range routes {
		if route.Quality > bestRoute.Quality {
			bestRoute = route
		}
	}

	return bestRoute
}

// Original printUsage function - kept for compatibility with command-line mode
func printUsage() {
	fmt.Println("Welcome to BitShare P2P Mesh Network")
	fmt.Println("Hope you find it useful!")
	fmt.Println("-----------------")
	fmt.Println("Usage:")
	fmt.Println("  Start mesh node:")
	fmt.Println("    bitshare start")
	fmt.Println("\n  Scan for peers:")
	fmt.Println("    bitshare scan")
	fmt.Println("\n  List known peers:")
	fmt.Println("    bitshare list")
	fmt.Println("\n  Send a file:")
	fmt.Println("    bitshare send <peer_id_or_name_or_ip> <port_no> \"<file_path_or_name>\"")
	fmt.Println("\n  Receive a file:")
	fmt.Println("    bitshare receive <port_no> [destination_directory]")
	fmt.Println("\n  Start interactive mode:")
	fmt.Println("    bitshare")
	fmt.Println("    or")
	fmt.Println("    bitshare interactive")

	fmt.Println("\nFor detailed help, start interactive mode and type 'help'")
}

// showInstallationInfo displays instructions for installing BitShare system-wide
func showInstallationInfo() {
	fmt.Println("\n" + colorGreen + "BitShare Installation Instructions" + colorReset)
	fmt.Println("================================")

	if runtime.GOOS == "windows" {
		fmt.Println("\nWindows Installation:")
		fmt.Println("1. Open PowerShell as Administrator")
		fmt.Println("2. Navigate to the BitShare directory")
		fmt.Println("   cd " + filepath.Dir(os.Args[0]))
		fmt.Println("3. Run the installation script:")
		fmt.Println("   .\\install.ps1")
	} else {
		fmt.Println("\nLinux/macOS Installation:")
		fmt.Println("1. Open Terminal")
		fmt.Println("2. Navigate to the BitShare directory")
		fmt.Println("   cd " + filepath.Dir(os.Args[0]))
		fmt.Println("3. Run the installation script:")
		fmt.Println("   ./install.sh")
		fmt.Println("\nAlternatively, use make:")
		fmt.Println("   make install")
	}

	fmt.Println("\nAfter installation, you can run BitShare from any terminal with:")
	fmt.Println("   bitshare")
}
