package p2p

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// TCPManager handles TCP/IP connections
type TCPManager struct {
	isRunning      bool
	listener       net.Listener
	connectedPeers map[string]*TCPPeer
	discoveryAddr  string
	listenPort     int
	mutex          sync.RWMutex
}

// TCPPeer represents a peer connected via TCP/IP
type TCPPeer struct {
	ID       string
	Name     string
	Address  string
	Conn     net.Conn
	LastSeen time.Time
}

// TCPDiscoveryMessage is used for peer discovery
type TCPDiscoveryMessage struct {
	MessageType  string   `json:"type"`
	NodeID       string   `json:"node_id"`
	NodeName     string   `json:"node_name"`
	Port         int      `json:"port"`
	Capabilities []string `json:"capabilities"`
}

var (
	tcpManager *TCPManager
	tcpOnce    sync.Once
)

// GetTCPManager returns the singleton instance of TCPManager
func GetTCPManager() *TCPManager {
	tcpOnce.Do(func() {
		tcpManager = &TCPManager{
			isRunning:      false,
			connectedPeers: make(map[string]*TCPPeer),
			discoveryAddr:  "255.255.255.255:9876", // Broadcast address for discovery
			listenPort:     9002,                   // Default port for TCP connections
		}
	})
	return tcpManager
}

// Start initializes and starts the TCP service
func (tm *TCPManager) Start(port int) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.isRunning {
		return errors.New("TCP service is already running")
	}

	if port > 0 {
		tm.listenPort = port
	}

	// Start listening on the specified port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tm.listenPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", tm.listenPort, err)
	}
	tm.listener = listener

	tm.isRunning = true

	// Start accepting connections
	go tm.acceptConnections()

	// Start discovery service
	go tm.startDiscoveryService()

	return nil
}

// Stop stops the TCP service
func (tm *TCPManager) Stop() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if !tm.isRunning {
		return nil
	}

	// Close listener
	if tm.listener != nil {
		tm.listener.Close()
	}

	// Close all connections
	for _, peer := range tm.connectedPeers {
		peer.Conn.Close()
	}

	tm.isRunning = false
	return nil
}

// Discover scans the local network for BitShare TCP peers
func (tm *TCPManager) Discover(timeout time.Duration) ([]PeerInfo, error) {
	results := make([]PeerInfo, 0)
	resultsChan := make(chan PeerInfo)
	errorChan := make(chan error)
	doneChan := make(chan bool)

	// Send broadcast discovery message
	go func() {
		conn, err := net.Dial("udp", tm.discoveryAddr)
		if err != nil {
			errorChan <- fmt.Errorf("failed to create UDP connection: %w", err)
			return
		}
		defer conn.Close()

		// Create discovery message
		msg := TCPDiscoveryMessage{
			MessageType:  "DISCOVER",
			NodeID:       "local-node", // Should be replaced with actual node ID
			NodeName:     "BitShare Node",
			Port:         tm.listenPort,
			Capabilities: []string{"transfer", "mesh"},
		}

		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			errorChan <- fmt.Errorf("failed to marshal discovery message: %w", err)
			return
		}

		_, err = conn.Write(jsonMsg)
		if err != nil {
			errorChan <- fmt.Errorf("failed to send discovery message: %w", err)
		}
		doneChan <- true
	}()

	// Listen for responses
	go func() {
		// Create UDP listener for discovery responses
		udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", tm.listenPort+1))
		if err != nil {
			errorChan <- fmt.Errorf("failed to resolve UDP address: %w", err)
			return
		}

		conn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			errorChan <- fmt.Errorf("failed to create UDP listener: %w", err)
			return
		}
		defer conn.Close()

		conn.SetReadDeadline(time.Now().Add(timeout))

		buffer := make([]byte, 1024)
		for {
			n, addr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					break
				}
				continue
			}

			var msg TCPDiscoveryMessage
			err = json.Unmarshal(buffer[:n], &msg)
			if err != nil {
				continue
			}

			if msg.MessageType == "DISCOVER_RESPONSE" {
				resultsChan <- PeerInfo{
					ID:             msg.NodeID,
					Name:           msg.NodeName,
					Address:        addr.IP.String(),
					Protocol:       "tcp",
					SignalStrength: 100, // Not applicable for TCP, use maximum
					LastSeen:       time.Now(),
					Capabilities:   msg.Capabilities,
				}
			}
		}
		doneChan <- true
	}()

	// Collect results with timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	// Wait for discovery message to be sent
	<-doneChan

	// Collect responses until timeout
	collectingResponses := true
	for collectingResponses {
		select {
		case peer := <-resultsChan:
			results = append(results, peer)
		case err := <-errorChan:
			return nil, err
		case <-timer.C:
			collectingResponses = false
		}
	}

	return results, nil
}

// Connect establishes a connection to a TCP peer
func (tm *TCPManager) Connect(peerAddress string, port int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", peerAddress, port))
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	// Create a new peer
	peer := &TCPPeer{
		ID:       fmt.Sprintf("tcp-%s-%d", peerAddress, port),
		Name:     fmt.Sprintf("Peer-%s", peerAddress),
		Address:  peerAddress,
		Conn:     conn,
		LastSeen: time.Now(),
	}

	// Add to connected peers
	tm.mutex.Lock()
	tm.connectedPeers[peer.ID] = peer
	tm.mutex.Unlock()

	// Handle communication with this peer in a separate goroutine
	go tm.handlePeer(peer)

	return nil
}

// SendData sends data to a connected peer
func (tm *TCPManager) SendData(peerID string, data []byte) error {
	tm.mutex.RLock()
	peer, exists := tm.connectedPeers[peerID]
	tm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("peer not connected: %s", peerID)
	}

	_, err := peer.Conn.Write(data)
	return err
}

// Helper methods
func (tm *TCPManager) acceptConnections() {
	for tm.isRunning {
		conn, err := tm.listener.Accept()
		if err != nil {
			// Check if we're shutting down
			tm.mutex.RLock()
			running := tm.isRunning
			tm.mutex.RUnlock()
			if !running {
				return
			}

			fmt.Printf("Error accepting TCP connection: %v\n", err)
			continue
		}

		// Handle the connection in a new goroutine
		go tm.handleConnection(conn)
	}
}

func (tm *TCPManager) handleConnection(conn net.Conn) {
	// In a real implementation, this would handle the connection protocol
	// For now, just log the connection
	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("New TCP connection from: %s\n", remoteAddr)

	// Create a new peer
	peer := &TCPPeer{
		ID:       fmt.Sprintf("tcp-%x", time.Now().UnixNano()),
		Name:     fmt.Sprintf("Peer-%s", remoteAddr),
		Address:  remoteAddr,
		Conn:     conn,
		LastSeen: time.Now(),
	}

	// Add to connected peers
	tm.mutex.Lock()
	tm.connectedPeers[peer.ID] = peer
	tm.mutex.Unlock()

	// Handle communication with this peer
	tm.handlePeer(peer)
}

func (tm *TCPManager) handlePeer(peer *TCPPeer) {
	reader := bufio.NewReader(peer.Conn)

	const maxMessageSize = 100 * 1024 * 1024 // 100MB maximum message size

	// Set read timeout to prevent hanging connections
	peer.Conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	// Use a single error logger function to reduce duplication
	logError := func(format string, args ...interface{}) {
		fmt.Printf("[TCP:%s] %s\n", peer.ID, fmt.Sprintf(format, args...))
	}

	for {
		// Read message length
		lengthBytes := make([]byte, 4)
		if _, err := io.ReadFull(reader, lengthBytes); err != nil {
			if err != io.EOF {
				logError("Read error: %v", err)
			}
			break
		}

		// Reset read deadline after successful read
		peer.Conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		// Parse and validate message length
		length := int(binary.BigEndian.Uint32(lengthBytes))
		if length <= 0 || length > maxMessageSize {
			logError("Invalid message length: %d", length)
			if length > maxMessageSize {
				break // Potential attack, close connection
			}
			continue
		}

		// Read message content
		message := make([]byte, length)
		if _, err := io.ReadFull(reader, message); err != nil {
			logError("Message read error: %v", err)
			break
		}

		// Update last seen time (only necessary field update)
		peer.LastSeen = time.Now()

		// Process message (only log errors, not every message)
		if err := tm.processMessage(peer, message); err != nil {
			logError("Processing error: %v", err)
			// Only break on fatal errors
			if isFatalError(err) {
				break
			}
		}
	}

	// Clean up peer connection
	tm.mutex.Lock()
	delete(tm.connectedPeers, peer.ID)
	tm.mutex.Unlock()
	peer.Conn.Close()
}

// isFatalError determines if an error should cause connection termination
func isFatalError(err error) bool {
	// Check for specific error types that warrant disconnection
	// For example: protocol violations, authentication failures
	return false // Default to non-fatal
}

// processMessage handles different message types more efficiently
func (tm *TCPManager) processMessage(peer *TCPPeer, message []byte) error {
	// Only log occasional messages or specific events, not every message
	if len(message) > 0 && message[0] == '{' {
		// Try parsing as JSON
		var msgHeader struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal(message, &msgHeader); err == nil {
			// Process based on message type with a more efficient switch
			switch msgHeader.Type {
			case "PING":
				return tm.sendPong(peer)
			case "DATA_TRANSFER", "MESH_ROUTE":
				return tm.routeMessage(peer, msgHeader.Type, message)
			}
			return nil
		}
	}

	// Handle binary data
	return tm.processBinaryMessage(peer, message)
}

// Simplified message handlers
func (tm *TCPManager) sendPong(peer *TCPPeer) error {
	// Send a simple pong response
	response := []byte(`{"type":"PONG","time":` + fmt.Sprint(time.Now().Unix()) + `}`)
	_, err := peer.Conn.Write(packMessage(response))
	return err
}

func (tm *TCPManager) routeMessage(peer *TCPPeer, msgType string, data []byte) error {
	// Currently a stub - mark unused parameters to avoid linter warnings
	// but keep the parameters for future implementation
	_ = msgType
	_ = data

	// Log that we received a message of this type
	fmt.Printf("[TCP:%s] Received %s message (%d bytes)\n", peer.ID, msgType, len(data))
	return nil
}

func (tm *TCPManager) processBinaryMessage(peer *TCPPeer, data []byte) error {
	// Currently a stub - mark unused parameters to avoid linter warnings
	// but keep the parameters for future implementation
	_ = data

	// Just log the binary message for now
	fmt.Printf("[TCP:%s] Received binary message (%d bytes)\n", peer.ID, len(data))
	return nil
}

// packMessage prepares a message for sending by adding length prefix
func packMessage(data []byte) []byte {
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(data)))
	return append(length, data...)
}

func (tm *TCPManager) startDiscoveryService() {
	// Listen for discovery messages
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", tm.listenPort+1))
	if err != nil {
		fmt.Printf("Failed to resolve UDP address for discovery: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Printf("Failed to create UDP listener for discovery: %v\n", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for tm.isRunning {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		var msg TCPDiscoveryMessage
		err = json.Unmarshal(buffer[:n], &msg)
		if err != nil {
			continue
		}

		if msg.MessageType == "DISCOVER" {
			// Send response
			response := TCPDiscoveryMessage{
				MessageType:  "DISCOVER_RESPONSE",
				NodeID:       "local-node", // Should be replaced with actual node ID
				NodeName:     "BitShare Node",
				Port:         tm.listenPort,
				Capabilities: []string{"transfer", "mesh"},
			}

			jsonResponse, err := json.Marshal(response)
			if err != nil {
				continue
			}

			conn.WriteToUDP(jsonResponse, addr)
		}
	}
}
