package transfer

import (
	"fmt"
	"io"
	"net"
	"os"
)

func SendFile(filePath string, ip string, port string) error {
	addr := ip + ":" + port
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("error connecting: %v", err)
	}
	defer conn.Close()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(conn, file)
	if err != nil {
		return fmt.Errorf("error sending file: %v", err)
	}

	fmt.Println("File sent successfully.")
	return nil
}

func ReceiveFile(destPath, listenAddr string) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Waiting for file...")

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Error accepting connection:", err)
		return
	}
	defer conn.Close()

	outFile, err := os.Create(destPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, conn)
	if err != nil {
		fmt.Println("Error receiving file:", err)
		return
	}
	fmt.Println("File received and saved to", destPath)
}
