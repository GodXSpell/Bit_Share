package main

import (
	"fmt"
	"os"
	"strconv"

	"fileshare/internal/transfer"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "send":
		if len(os.Args) != 5 {
			fmt.Println("Usage: go run main.go send <file_path> <receiver_ip> <port>")
			os.Exit(1)
		}
		filePath := os.Args[2]
		receiverIP := os.Args[3]
		port, err := strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Printf("Invalid port number: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Sending file '%s' to %s:%d\n", filePath, receiverIP, port)
		err = transfer.SendFile(filePath, receiverIP, port)
		if err != nil {
			fmt.Printf("Error sending file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("File sent successfully!")

	case "receive":
		if len(os.Args) != 4 {
			fmt.Println("Usage: go run main.go receive <save_path> <listen_port>")
			os.Exit(1)
		}
		savePath := os.Args[2]
		port, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Printf("Invalid port number: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Waiting to receive file on port %d...\n", port)
		err = transfer.ReceiveFile(savePath, port)
		if err != nil {
			fmt.Printf("Error receiving file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("File received and saved as '%s'\n", savePath)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("File Share Tool")
	fmt.Println("Usage:")
	fmt.Println("  Send:    go run main.go send <file_path> <receiver_ip> <port>")
	fmt.Println("  Receive: go run main.go receive <save_path> <listen_port>")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run main.go send test.txt 192.168.1.5 9000")
	fmt.Println("  go run main.go receive received_test.txt 9000")
}
