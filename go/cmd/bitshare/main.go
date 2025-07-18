package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"fileshare/internal/mesh"
	"fileshare/internal/p2p"
	"fileshare/internal/transfer"
	"fileshare/internal/ui"
	"fileshare/internal/utils"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "interactive":
		// Start interactive terminal UI
		startInteractiveMode()

	case "start":
		// Start mesh node with optional parameters
		startMeshNode(os.Args[2:])

	case "scan":
		// Scan for nearby peers
		scanForPeers(os.Args[2:])

	case "list":
		// List known peers
		listPeers()

	case "connect":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bitshare connect <peer_id_or_address>")
			os.Exit(1)
		}
		connectToPeer(os.Args[2])

	case "send":
		if len(os.Args) < 4 {
			fmt.Println("Usage: bitshare send <peer_id> <file_path>")
			os.Exit(1)
		}
		sendFile(os.Args[2], os.Args[3])

	case "receive":
		if len(os.Args) < 3 {
			fmt.Println("Usage: bitshare receive <port_no> [destination_directory]")
			os.Exit(1)
		}
		port, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid port number: %v\n", err)
			os.Exit(1)
		}
		destDir := "."
		if len(os.Args) >= 4 {
			destDir = os.Args[3]
		}
		startReceiver(port, destDir)

	case "help":
		printDetailedHelp()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func startInteractiveMode() {
	// Initialize terminal UI
	termUI := ui.GetTerminalUI()
	err := termUI.Start()
	if err != nil {
		fmt.Printf("Failed to start interactive mode: %v\n", err)
		os.Exit(1)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start mesh node
	config := mesh.Config{
		NodeName:         utils.GenerateNodeName(),
		ListenPort:       9000,
		EnableWiFiDirect: true,
		EnableBluetooth:  true,
		EnableTCP:        true,
	}

	fmt.Println("Starting BitShare mesh node in interactive mode...")
	err = mesh.StartMeshNode(config)
	if err != nil {
		fmt.Printf("Failed to start mesh node: %v\n", err)
		termUI.Stop()
		os.Exit(1)
	}

	// Show dashboard
	termUI.ShowDashboard()

	// Process user input in a separate goroutine
	go processUserInput(termUI)

	// Wait for signal
	<-sigChan
	fmt.Println("\nShutting down...")
	mesh.StopMeshNode()
	termUI.Stop()
}

func processUserInput(termUI *ui.TerminalUI) {
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		command := strings.TrimSpace(input)
		if len(command) == 0 {
			continue
		}

		// Process single-character commands for UI navigation
		switch strings.ToLower(command) {
		case "d", "dashboard":
			termUI.ShowDashboard()
		case "p", "peers":
			termUI.ShowPeerList()
		case "t", "transfers":
			termUI.ShowTransferStatus()
		case "n", "network":
			termUI.ShowNetworkMap()
		case "s", "scan":
			go func() {
				fmt.Println("Scanning for peers...")
				peers, err := p2p.ScanForPeers()
				if err != nil {
					fmt.Printf("Scan error: %v\n", err)
				} else {
					fmt.Printf("Found %d peers\n", len(peers))
				}
			}()
		case "q", "quit", "exit":
			fmt.Println("Exiting...")
			// Send interrupt signal to trigger shutdown
			p, err := os.FindProcess(os.Getpid())
			if err == nil {
				p.Signal(syscall.SIGINT)
			}
			return
		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}

func startMeshNode(args []string) {
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Parse options from args if any
	config := mesh.Config{
		NodeName:         utils.GenerateNodeName(),
		ListenPort:       9000, // Default port
		EnableWiFiDirect: true,
		EnableBluetooth:  true,
		EnableTCP:        true,
	}

	// TODO: Parse additional options from args
	// This would handle flags like --port, --name, etc.

	fmt.Println("🌐 Starting BitShare mesh node...")
	err := mesh.StartMeshNode(config)
	if err != nil {
		fmt.Printf("❌ Failed to start mesh node: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Node started successfully as '%s'\n", config.NodeName)
	fmt.Println("📡 Listening for connections...")
	fmt.Println("Press Ctrl+C to stop")

	// Handle graceful shutdown
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down mesh node...")
		mesh.StopMeshNode()
		os.Exit(0)
	}()

	// Block until signal received
	select {}
}

func scanForPeers(args []string) {
	fmt.Println("🔍 Scanning for nearby peers...")

	peers, err := p2p.ScanForPeers()
	if err != nil {
		fmt.Printf("❌ Scan error: %v\n", err)
		os.Exit(1)
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

func listPeers() {
	peers, err := mesh.GetKnownPeers()
	if err != nil {
		fmt.Printf("❌ Error retrieving peers: %v\n", err)
		os.Exit(1)
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

func connectToPeer(peerID string) {
	fmt.Printf("Connecting to peer: %s\n", peerID)

	// This would use the mesh coordinator to establish a connection
	// For now, just print a message
	fmt.Println("Connection functionality not fully implemented")
}

func sendFile(peerID, filePath string) {
	// Validate file exists
	if !utils.FileExists(filePath) {
		fmt.Printf("File not found: %s\n", filePath)
		os.Exit(1)
	}

	fmt.Printf("Sending file %s to peer %s\n", filePath, peerID)

	// Create transfer options
	options := transfer.DefaultTransferOptions()

	// Start chunked transfer
	err := transfer.SendFileChunked(filePath, peerID, options)
	if err != nil {
		fmt.Printf("Error sending file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("File sent successfully!")
}

func startReceiver(port int, destDir string) {
	fmt.Printf("Starting receiver on port %d, saving to %s\n", port, destDir)

	// Ensure destination directory exists
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		fmt.Printf("Creating destination directory: %s\n", destDir)
		os.MkdirAll(destDir, 0755)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nStopping receiver...")
		os.Exit(0)
	}()

	fmt.Printf("Waiting for incoming files on port %d...\n", port)
	fmt.Println("Press Ctrl+C to stop")

	// Start receiver in blocking mode
	err := transfer.ReceiveFileWithTimeout(port, 0, destDir) // 0 timeout means no timeout
	if err != nil {
		fmt.Printf("Error receiving file: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("BitShare P2P Mesh Network")
	fmt.Println("========================")
	fmt.Println("Usage: bitshare <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  interactive    Start BitShare in interactive terminal UI mode")
	fmt.Println("  start          Start a mesh network node")
	fmt.Println("  scan           Scan for nearby peers")
	fmt.Println("  list           List known peers in the network")
	fmt.Println("  connect        Connect to a specific peer")
	fmt.Println("  send           Send a file to a peer")
	fmt.Println("  receive        Receive files on a specified port")
	fmt.Println("  help           Display detailed help")
	fmt.Println("\nRun 'bitshare help' for more detailed information.")
}

func printDetailedHelp() {
	fmt.Println("BitShare P2P Mesh Network - Detailed Help")
	fmt.Println("========================================")
	fmt.Println("\nBitShare is a high-performance P2P mesh network for file sharing")
	fmt.Println("that works without routers using direct connections between devices.")
	fmt.Println("\nCommands:")

	fmt.Println("\n  interactive")
	fmt.Println("    Start BitShare in interactive terminal UI mode")
	fmt.Println("    Usage: bitshare interactive")

	fmt.Println("\n  start [options]")
	fmt.Println("    Start a mesh network node")
	fmt.Println("    Options:")
	fmt.Println("      --port <number>    Port to listen on (default: 9000)")
	fmt.Println("      --name <name>      Node name (default: hostname)")
	fmt.Println("    Usage: bitshare start --port 9000 --name MyLaptop")

	fmt.Println("\n  scan")
	fmt.Println("    Scan for nearby peers using all available protocols")
	fmt.Println("    Usage: bitshare scan")

	fmt.Println("\n  list")
	fmt.Println("    List all known peers in the mesh network")
	fmt.Println("    Usage: bitshare list")

	fmt.Println("\n  connect <peer_id>")
	fmt.Println("    Connect to a specific peer")
	fmt.Println("    Usage: bitshare connect bob-laptop")

	fmt.Println("\n  send <peer_id> <file_path>")
	fmt.Println("    Send a file to a peer")
	fmt.Println("    Usage: bitshare send bob-laptop \"C:\\Users\\Alice\\Documents\\report.pdf\"")

	fmt.Println("\n  receive <port_no> [destination_directory]")
	fmt.Println("    Receive files on the specified port")
	fmt.Println("    Usage: bitshare receive 9000 \"C:\\Downloads\"")

	fmt.Println("\nFeatures:")
	fmt.Println("  - Direct P2P connections using WiFi Direct, TCP/IP, and Bluetooth")
	fmt.Println("  - Mesh network for extended range through intermediate nodes")
	fmt.Println("  - High-performance chunked file transfers with parallel streams")
	fmt.Println("  - Resumable downloads for reliability")
}
