package transfer

import (
	"fileshare/internal/utils"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	MaxFileSize = 10 * 1024 * 1024 * 1024 // 10GB limit
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

	// Check file size limit
	if fileInfo.Size() > MaxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d bytes)", fileInfo.Size(), MaxFileSize)
	}

	// Connect to receiver
	address := net.JoinHostPort(receiverIP, fmt.Sprintf("%d", port))
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to receiver: %v", err)
	}
	defer conn.Close()

	// Set connection timeout
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Send filename first
	filename := filepath.Base(filePath)
	fmt.Printf("Sending file: %s (%s)\n", filename, utils.FormatBytes(fileInfo.Size()))

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
func ReceiveFile(port int, destDir string) error {
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

	return receiveFileFromConnection(conn, destDir)
}

// ReceiveFileWithTimeout receives a file with connection timeout
func ReceiveFileWithTimeout(port int, timeout time.Duration, destDir string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Listening on port %d...\n", port)

	// Set accept timeout
	if tcpListener, ok := listener.(*net.TCPListener); ok {
		tcpListener.SetDeadline(time.Now().Add(timeout))
	}

	conn, err := listener.Accept()
	if err != nil {
		return fmt.Errorf("failed to accept connection: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Connection established with %s\n", conn.RemoteAddr())

	// Set read/write timeouts for security
	conn.SetReadDeadline(time.Now().Add(timeout))
	conn.SetWriteDeadline(time.Now().Add(timeout))

	return receiveFileFromConnection(conn, destDir)
}

// receiveFileFromConnection handles the file reception from an established connection
func receiveFileFromConnection(conn net.Conn, destDir string) error {
	// Read filename and size
	var filename string
	var fileSize int64
	_, err := fmt.Fscanf(conn, "%s\n%d\n", &filename, &fileSize)
	if err != nil {
		return fmt.Errorf("failed to read file metadata: %v", err)
	}

	// Security checks
	if fileSize <= 0 || fileSize > MaxFileSize {
		return fmt.Errorf("invalid file size: %d bytes", fileSize)
	}

	// Sanitize filename to prevent path traversal
	filename = filepath.Base(filename)
	if filename == "" || filename == "." || filename == ".." {
		return fmt.Errorf("invalid filename: %s", filename)
	}

	// Ensure destination directory exists
	if destDir != "" {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %v", err)
		}
	}

	// Create output file with original filename in the destination directory
	outputPath := filepath.Join(destDir, filename)

	// Get absolute path for user-friendly output
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		// If getting absolute path fails, use the original path for logging.
		// This is not a fatal error for the transfer itself.
		absPath = outputPath
	}
	fmt.Printf("Receiving file: %s (%s) -> %s\n", filename, utils.FormatBytes(fileSize), absPath)

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Receive file content
	bytesReceived, err := io.CopyN(outputFile, conn, fileSize)
	if err != nil {
		return fmt.Errorf("failed to receive file content: %v", err)
	}

	if bytesReceived != fileSize {
		return fmt.Errorf("incomplete transfer: received %d bytes, expected %d bytes", bytesReceived, fileSize)
	}

	fmt.Printf("Successfully received %s (%s) at %s\n", filename, utils.FormatBytes(bytesReceived), absPath)
	return nil
}
