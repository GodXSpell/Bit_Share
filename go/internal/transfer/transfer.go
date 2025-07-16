package transfer

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

// SendFile connects to a receiver and sends a file
func SendFile(filePath, receiverIP string, port int) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// Connect to receiver
	address := net.JoinHostPort(receiverIP, fmt.Sprintf("%d", port))
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to receiver: %v", err)
	}
	defer conn.Close()

	// Send filename first
	filename := filepath.Base(filePath)
	_, err = fmt.Fprintf(conn, "%s\n%d\n", filename, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to send file metadata: %v", err)
	}

	// Send file content
	_, err = io.Copy(conn, file)
	if err != nil {
		return fmt.Errorf("failed to send file content: %v", err)
	}

	return nil
}

// ReceiveFile starts a TCP listener and receives a file
func ReceiveFile(savePath string, port int) error {
	// Start TCP listener
	listener, err := net.Listen("tcp", net.JoinHostPort("", fmt.Sprintf("%d", port)))
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Listening on port %d...\n", port)

	// Accept connection
	conn, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("failed to accept connection: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connection established with %s\n", conn.RemoteAddr())

	// Read filename and size
	var filename string
	var fileSize int64
	_, err = fmt.Fscanf(conn, "%s\n%d\n", &filename, &fileSize)
	if err != nil {
		return fmt.Errorf("failed to read file metadata: %v", err)
	}

	fmt.Printf("Receiving file: %s (size: %d bytes)\n", filename, fileSize)

	// Create output file
	outputFile, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Receive file content
	bytesReceived, err := io.CopyN(outputFile, conn, fileSize)
	if err != nil {
		return fmt.Errorf("failed to receive file content: %v", err)
	}

	fmt.Printf("Successfully received %d bytes\n", bytesReceived)
	return nil
}
