package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"fileshare/internal/firewall"
	"fileshare/internal/transfer"
	"fileshare/internal/utils"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "receive":
		if len(os.Args) < 3 || len(os.Args) > 4 {
			fmt.Println("Usage: receive <port_no> [destination_directory]")
			os.Exit(1)
		}
		port, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("Invalid port number: %v\n", err)
			os.Exit(1)
		}
		if port < 1 || port > 65535 {
			fmt.Println("Port number must be between 1 and 65535")
			os.Exit(1)
		}
		destDir := "." // Default to current directory
		if len(os.Args) == 4 {
			destDir = os.Args[3]
		}
		startReceiver(port, destDir)

	case "send":
		if len(os.Args) != 5 {
			fmt.Println("Usage: send <ip> <port_no> <file_path>")
			os.Exit(1)
		}
		ip := os.Args[2]

		// Validate IP address
		if net.ParseIP(ip) == nil {
			fmt.Printf("Invalid IP address: %s\n", ip)
			os.Exit(1)
		}

		port, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("Invalid port number: %v\n", err)
			os.Exit(1)
		}
		if port < 1 || port > 65535 {
			fmt.Println("Port number must be between 1 and 65535")
			os.Exit(1)
		}
		filePath := os.Args[4]
		startSender(ip, port, filePath)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

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
		fmt.Println("\n🛑 Shutting down...")
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
		os.Exit(1)
	}
}

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
			os.Exit(1)
		}
		fmt.Printf("File found: %s\n", foundPath)
		filePath = foundPath
	}

	// Check if file is readable
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Cannot read file: %v\n", err)
		os.Exit(1)
	}
	file.Close()

	err = transfer.SendFile(filePath, ip, port)
	if err != nil {
		fmt.Printf("Error sending file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("File sent successfully!")
}

func printUsage() {
	fmt.Println("Welcome to BitShare")
	fmt.Println("Hope you find it useful!")
	fmt.Println("-----------------")
	fmt.Println("Usage:")
	fmt.Println("  Send a file:")
	fmt.Println("    gomain send <ip> <port_no> \"<file_path_or_name>\"")
	fmt.Println("\n  Receive a file:")
	fmt.Println("    gomain receive <port_no> [destination_directory]")
	fmt.Println("\nExamples:")
	fmt.Println("  - Send a file using its full path (use quotes if the path has spaces):")
	fmt.Println("    gomain send 192.168.1.5 9000 \"C:\\Users\\YourUser\\My Documents\\report.docx\"")
	fmt.Println("\n  - Send a file by name (searches Downloads, Documents, Desktop):")
	fmt.Println("    gomain send 192.168.1.5 9000 \"MyPicture.jpg\"")
	fmt.Println("\n  - Receive a file into the current folder:")
	fmt.Println("    gomain receive 9000")
	fmt.Println("\n  - Receive a file into a specific folder:")
	fmt.Println("    gomain receive 9000 C:\\Users\\YourUser\\Downloads")
	fmt.Println("\n--- How It Works: A Scenario ---")
	fmt.Println("1. On the RECEIVING computer (let's say its IP is 192.168.1.10):")
	fmt.Println("   - The user runs this to listen on port 9000 and save files to their Downloads folder:")
	fmt.Println("     gomain receive 9000 \"C:\\Users\\ReceiverUser\\Downloads\"")
	fmt.Println("\n2. On the SENDING computer:")
	fmt.Println("   - You want to send 'MyReport.pdf' which is on your Desktop.")
	fmt.Println("   - You run this command, pointing to the receiver's IP and port:")
	fmt.Println("     gomain send 192.168.1.10 9000 \"MyReport.pdf\"")
	fmt.Println("\n3. Result:")
	fmt.Println("   - The tool finds 'MyReport.pdf' on your Desktop and sends it.")
	fmt.Println("   - The receiving computer saves it to 'C:\\Users\\ReceiverUser\\Downloads\\MyReport.pdf'.")
}

// needsFirewallRule is no longer needed as we always attempt to add a rule
/*
func needsFirewallRule(port int) bool {
	// Simple port availability check
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true // Port might be blocked
	}
	listener.Close()
	return false
}
*/
