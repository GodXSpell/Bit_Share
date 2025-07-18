package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// WiFiDirectManager handles WiFi Direct connections
type WiFiDirectManager struct {
	isRunning      bool
	listener       net.Listener
	connectedPeers map[string]*WiFiDirectPeer
	mutex          sync.RWMutex
	config         WiFiDirectConfig
}

// WiFiDirectConfig contains WiFi Direct configuration
type WiFiDirectConfig struct {
	GroupOwner     bool
	ServiceName    string
	ListenPort     int
	MaxConnections int
}

// WiFiDirectPeer represents a peer connected via WiFi Direct
type WiFiDirectPeer struct {
	ID        string
	Name      string
	Address   string
	Conn      net.Conn
	LastSeen  time.Time
	SignalDBM int // Signal strength in dBm
}

var (
	wifiDirectManager *WiFiDirectManager
	wdOnce            sync.Once
)

// GetWiFiDirectManager returns the singleton instance of WiFiDirectManager
func GetWiFiDirectManager() *WiFiDirectManager {
	wdOnce.Do(func() {
		wifiDirectManager = &WiFiDirectManager{
			isRunning:      false,
			connectedPeers: make(map[string]*WiFiDirectPeer),
			config: WiFiDirectConfig{
				GroupOwner:     true,
				ServiceName:    "BitShare",
				ListenPort:     9001,
				MaxConnections: 10,
			},
		}
	})
	return wifiDirectManager
}

// Start initializes and starts the WiFi Direct service
func (wdm *WiFiDirectManager) Start() error {
	wdm.mutex.Lock()
	defer wdm.mutex.Unlock()

	if wdm.isRunning {
		return errors.New("WiFi Direct service is already running")
	}

	// Check if WiFi Direct is supported on this device
	supported, err := isWiFiDirectSupported()
	if err != nil {
		return fmt.Errorf("failed to check WiFi Direct support: %w", err)
	}
	if !supported {
		return errors.New("WiFi Direct is not supported on this device")
	}

	// Start the WiFi Direct service
	if wdm.config.GroupOwner {
		// Start as group owner (acts like an access point)
		err = wdm.startAsGroupOwner()
	} else {
		// Start as client
		err = wdm.startAsClient()
	}

	if err != nil {
		return fmt.Errorf("failed to start WiFi Direct: %w", err)
	}

	wdm.isRunning = true
	go wdm.acceptConnections()

	return nil
}

// Stop stops the WiFi Direct service
func (wdm *WiFiDirectManager) Stop() error {
	wdm.mutex.Lock()
	defer wdm.mutex.Unlock()

	if !wdm.isRunning {
		return nil
	}

	// Close listener
	if wdm.listener != nil {
		wdm.listener.Close()
	}

	// Close all connections
	for _, peer := range wdm.connectedPeers {
		peer.Conn.Close()
	}

	wdm.isRunning = false
	return nil
}

// Discover scans for nearby WiFi Direct devices
func (wdm *WiFiDirectManager) Discover(timeout time.Duration) ([]PeerInfo, error) {
	// Implementation would use platform-specific APIs to discover WiFi Direct devices
	// This is a placeholder implementation

	// On real implementation, this would use:
	// - Windows: WiFi Direct API via Windows.Devices.WiFiDirect namespace
	// - Linux: wpa_supplicant and nl80211 interfaces
	// - Android: WifiP2pManager API
	// - iOS: MultipeerConnectivity framework

	peers := []PeerInfo{
		{
			ID:             "wd-device1",
			Name:           "Laptop-ABC",
			Address:        "192.168.49.10",
			Protocol:       "wifi-direct",
			SignalStrength: 85,
			LastSeen:       time.Now(),
			Capabilities:   []string{"transfer", "mesh"},
		},
		{
			ID:             "wd-device2",
			Name:           "Tablet-XYZ",
			Address:        "192.168.49.11",
			Protocol:       "wifi-direct",
			SignalStrength: 70,
			LastSeen:       time.Now(),
			Capabilities:   []string{"transfer"},
		},
	}

	return peers, nil
}

// Connect establishes a connection to a WiFi Direct peer
func (wdm *WiFiDirectManager) Connect(peerID string) error {
	// Implementation would connect to a specific peer
	// This is a placeholder implementation
	return nil
}

// SendData sends data to a connected peer
func (wdm *WiFiDirectManager) SendData(peerID string, data []byte) error {
	wdm.mutex.RLock()
	peer, exists := wdm.connectedPeers[peerID]
	wdm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("peer not connected: %s", peerID)
	}

	_, err := peer.Conn.Write(data)
	return err
}

// Helper methods
func (wdm *WiFiDirectManager) startAsGroupOwner() error {
	// Listen on the specified port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", wdm.config.ListenPort))
	if err != nil {
		return err
	}
	wdm.listener = listener

	// In a real implementation, this would advertise the service using platform-specific APIs
	fmt.Printf("Started WiFi Direct group with service name: %s\n", wdm.config.ServiceName)
	return nil
}

func (wdm *WiFiDirectManager) startAsClient() error {
	// In a real implementation, this would join an existing group
	return errors.New("client mode not implemented")
}

func (wdm *WiFiDirectManager) acceptConnections() {
	for wdm.isRunning {
		conn, err := wdm.listener.Accept()
		if err != nil {
			// Check if we're shutting down
			wdm.mutex.RLock()
			running := wdm.isRunning
			wdm.mutex.RUnlock()
			if !running {
				return
			}

			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		// Handle the connection in a new goroutine
		go wdm.handleConnection(conn)
	}
}

func (wdm *WiFiDirectManager) handleConnection(conn net.Conn) {
	// In a real implementation, this would handle the connection protocol
	// For now, just log the connection
	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("New connection from: %s\n", remoteAddr)

	// Create a new peer
	peer := &WiFiDirectPeer{
		ID:        fmt.Sprintf("wd-%x", time.Now().UnixNano()),
		Name:      fmt.Sprintf("Peer-%s", remoteAddr),
		Address:   remoteAddr,
		Conn:      conn,
		LastSeen:  time.Now(),
		SignalDBM: -50, // Placeholder value
	}

	// Add to connected peers
	wdm.mutex.Lock()
	wdm.connectedPeers[peer.ID] = peer
	wdm.mutex.Unlock()

	// Handle communication with this peer
	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Error reading from connection: %v\n", err)
			break
		}

		// Process received data
		fmt.Printf("Received %d bytes from %s\n", n, peer.ID)

		// Update last seen time
		wdm.mutex.Lock()
		if p, exists := wdm.connectedPeers[peer.ID]; exists {
			p.LastSeen = time.Now()
		}
		wdm.mutex.Unlock()
	}

	// Remove peer when connection closes
	wdm.mutex.Lock()
	delete(wdm.connectedPeers, peer.ID)
	wdm.mutex.Unlock()
	conn.Close()
}

// Helper functions for platform detection and support
func isWiFiDirectSupported() (bool, error) {
	// In a real implementation, this would check if WiFi Direct is supported on this device
	// For now, just return true
	return true, nil
}
