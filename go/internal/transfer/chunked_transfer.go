package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ChunkInfo represents information about a file chunk
type ChunkInfo struct {
	Index     int
	Size      int64
	Offset    int64
	Checksum  string
	Completed bool
}

// FileTransferInfo contains information about a file transfer
type FileTransferInfo struct {
	FileID       string
	FileName     string
	FilePath     string
	FileSize     int64
	ChunkSize    int64
	Chunks       []ChunkInfo
	TotalChunks  int
	Completed    int
	StartTime    time.Time
	TransferRate int64 // bytes per second
	Status       string
	Error        error
	Mutex        sync.Mutex
}

// TransferOptions configures the behavior of file transfers
type TransferOptions struct {
	ChunkSize        int64         // Size of each chunk in bytes (default: 1MB)
	Parallelism      int           // Number of parallel transfers (default: 5)
	RetryCount       int           // Number of retries per chunk (default: 3)
	RetryDelay       time.Duration // Delay between retries (default: 1s)
	CompressData     bool          // Whether to compress data (default: true)
	VerifyChecksums  bool          // Whether to verify checksums (default: true)
	ProgressCallback func(*FileTransferInfo)
}

// DefaultTransferOptions returns the default transfer configuration
func DefaultTransferOptions() TransferOptions {
	return TransferOptions{
		ChunkSize:       1 * 1024 * 1024, // 1MB
		Parallelism:     5,
		RetryCount:      3,
		RetryDelay:      time.Second,
		CompressData:    true,
		VerifyChecksums: true,
		ProgressCallback: func(info *FileTransferInfo) {
			// Default progress reporting
			progress := float64(info.Completed) / float64(info.TotalChunks) * 100
			fmt.Printf("\rTransfer progress: %.1f%% (%d/%d chunks) - %.2f MB/s",
				progress, info.Completed, info.TotalChunks, float64(info.TransferRate)/(1024*1024))
		},
	}
}

// SendFileChunked sends a file using the chunked transfer protocol
func SendFileChunked(filePath, peerID string, options TransferOptions) error {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Generate file ID
	fileID := generateFileID(filePath)

	// Calculate chunks
	chunkSize := options.ChunkSize
	fileSize := fileInfo.Size()
	totalChunks := int((fileSize + chunkSize - 1) / chunkSize) // Ceiling division

	// Create transfer info
	transferInfo := &FileTransferInfo{
		FileID:      fileID,
		FileName:    filepath.Base(filePath),
		FilePath:    filePath,
		FileSize:    fileSize,
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		Chunks:      make([]ChunkInfo, totalChunks),
		StartTime:   time.Now(),
		Status:      "preparing",
	}

	// Prepare chunk info
	for i := 0; i < totalChunks; i++ {
		offset := int64(i) * chunkSize
		size := chunkSize
		if offset+size > fileSize {
			size = fileSize - offset
		}

		// Calculate checksum for this chunk
		checksum, err := calculateChunkChecksum(file, offset, size)
		if err != nil {
			return fmt.Errorf("failed to calculate checksum: %w", err)
		}

		transferInfo.Chunks[i] = ChunkInfo{
			Index:     i,
			Size:      size,
			Offset:    offset,
			Checksum:  checksum,
			Completed: false,
		}
	}

	// Send file metadata to peer
	err = sendFileMetadata(transferInfo, peerID)
	if err != nil {
		return fmt.Errorf("failed to send file metadata: %w", err)
	}

	// Start the transfer
	transferInfo.Status = "transferring"
	err = sendFileChunks(file, transferInfo, peerID, options)
	if err != nil {
		transferInfo.Status = "failed"
		transferInfo.Error = err
		return fmt.Errorf("failed to send file chunks: %w", err)
	}

	transferInfo.Status = "completed"
	return nil
}

// ReceiveFileChunked receives a file using the chunked transfer protocol
func ReceiveFileChunked(peerID, destDir string, options TransferOptions) error {
	// Receive file metadata
	transferInfo, err := receiveFileMetadata(peerID)
	if err != nil {
		return fmt.Errorf("failed to receive file metadata: %w", err)
	}

	// Create destination file
	destPath := filepath.Join(destDir, transferInfo.FileName)
	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Pre-allocate file size if possible
	err = file.Truncate(transferInfo.FileSize)
	if err != nil {
		return fmt.Errorf("failed to pre-allocate file: %w", err)
	}

	// Start receiving chunks
	transferInfo.Status = "receiving"
	err = receiveFileChunks(file, transferInfo, peerID, options)
	if err != nil {
		transferInfo.Status = "failed"
		transferInfo.Error = err
		return fmt.Errorf("failed to receive file chunks: %w", err)
	}

	transferInfo.Status = "completed"
	return nil
}

// Helper functions
func generateFileID(filePath string) string {
	// Generate a unique ID for this file
	hash := sha256.Sum256([]byte(filePath + time.Now().String()))
	return hex.EncodeToString(hash[:])[:16]
}

func calculateChunkChecksum(file *os.File, offset, size int64) (string, error) {
	// Calculate SHA-256 checksum of the chunk
	hasher := sha256.New()
	buffer := make([]byte, 64*1024) // Use a smaller buffer for reading

	// Move to the correct offset
	_, err := file.Seek(offset, io.SeekStart)
	if err != nil {
		return "", err
	}

	// Read the chunk and update the hash
	var totalRead int64
	for totalRead < size {
		readSize := size - totalRead
		if readSize > int64(len(buffer)) {
			readSize = int64(len(buffer))
		}

		n, err := file.Read(buffer[:readSize])
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			break
		}

		hasher.Write(buffer[:n])
		totalRead += int64(n)

		if err == io.EOF {
			break
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func sendFileMetadata(info *FileTransferInfo, peerID string) error {
	// Send file metadata to the peer
	// This is a placeholder for the actual implementation
	return nil
}

func receiveFileMetadata(peerID string) (*FileTransferInfo, error) {
	// Receive file metadata from the peer
	// This is a placeholder for the actual implementation
	return &FileTransferInfo{}, nil
}

func sendFileChunks(file *os.File, info *FileTransferInfo, peerID string, options TransferOptions) error {
	// Send file chunks to the peer
	// This is a placeholder for the actual implementation
	return nil
}

func receiveFileChunks(file *os.File, info *FileTransferInfo, peerID string, options TransferOptions) error {
	// Receive file chunks from the peer
	// This is a placeholder for the actual implementation
	return nil
}
