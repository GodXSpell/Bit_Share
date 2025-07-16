package main

import (
	"fileshare/internal/transfer"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go [send|receive]")
		return
	}

	mode := os.Args[1]
	switch mode {
	case "send":
		var ip string
		fmt.Print("Enter receiver IP (e.g. 192.168.1.23): ")
		fmt.Scanln(&ip)
		err := transfer.SendFile("test.txt", ip, "8080")
		if err != nil {
			log.Fatalf("Send error: %v", err)
		}

	case "receive":
		transfer.ReceiveFile("received_test.txt", ":9000")
	default:
		fmt.Println("Unknown mode:", mode)
	}
}
