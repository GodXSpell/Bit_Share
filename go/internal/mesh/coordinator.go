package mesh

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// Config stores mesh network configuration
type Config struct {
	NodeName         string
	NodeID           string // Will be generated if empty
	ListenPort       int
	EnableWiFiDirect bool
	EnableBluetooth  bool
	EnableTCP        bool
	EnableRelay      bool     // Whether to use relay servers when direct connection fails
	RelayServers     []string // List of relay servers to use
	DataDir          string   // Directory to store mesh data
}

// NetworkMode indicates how peers can connect in the current network
type NetworkMode int

const (
	DirectMode NetworkMode = iota // Direct connections allowed
	RelayMode                     // Must use relay server
	MixedMode                     // Some peers direct, some relayed
)

// ConnectionInfo tracks network connectivity status
type ConnectionInfo struct {
	Mode                  NetworkMode
	ClientIsolation       bool
	NATType               string
	PublicIP              string
	RelayAvailable        bool
	LastConnectivityCheck time.Time
}

// Peer represents a node in the mesh network
type Peer struct {
	ID                string
	Name              string
	Address           string
	Protocol          string
	IsOnline          bool
	LastSeen          time.Time
	SignalStrength    int // 0-100%
	ConnectionQuality string
	Routes            []Route
}

// Route represents a path to a peer
type Route struct {
	DestinationID string
	NextHop       string
	HopCount      int
	Quality       int // 0-100%
}

var (
	meshConfig     Config
	isRunning      bool
	nodeID         string
	knownPeers     = make(map[string]*Peer)
	peersMutex     sync.RWMutex
	connectionInfo ConnectionInfo
)

// StartMeshNode initializes and starts the mesh network node
func StartMeshNode(config Config) error {
	if isRunning {
		return errors.New("mesh node is already running")
	}

	// Initialize node ID if not provided
	if config.NodeID == "" {
		config.NodeID = generateNodeID()
	}

	// Set default relay settings if not provided
	if config.EnableRelay && len(config.RelayServers) == 0 {
		config.RelayServers = []string{"relay1.bitshare.net:9100", "relay2.bitshare.net:9100"}
	}

	meshConfig = config
	nodeID = config.NodeID

	// Detect network conditions before starting protocol handlers
	detectNetworkConditions()

	// Start protocol handlers based on configuration
	if config.EnableWiFiDirect {
		go startWiFiDirectHandler(config.ListenPort)
	}

	if config.EnableBluetooth {
		go startBluetoothHandler()
	}

	if config.EnableTCP {
		go startTCPHandler(config.ListenPort)
	}

	// If client isolation detected, notify user
	if connectionInfo.ClientIsolation {
		fmt.Println("⚠️ Client isolation detected in your network")
		fmt.Println("→ Direct peer connections may be restricted")

		if config.EnableWiFiDirect {
			fmt.Println("→ Will attempt WiFi Direct for direct connections")
		}

		if config.EnableRelay {
			fmt.Println("→ Using relay servers for restricted connections")
		} else {
			fmt.Println("⚠️ Relay mode is disabled. Some peers may be unreachable")
			fmt.Println("→ Enable relay with --enable-relay flag to improve connectivity")
		}
	}

	// Start relay connection handler if enabled
	if config.EnableRelay {
		go startRelayHandler(config.RelayServers)
	}

	// Start the discovery service
	go startDiscoveryService()

	// Start the routing table maintenance
	go maintainRoutingTable()

	// Periodically check network conditions
	go monitorNetworkConditions()

	isRunning = true
	return nil
}

// StopMeshNode gracefully shuts down the mesh node
func StopMeshNode() {
	if !isRunning {
		return
	}

	// Send departure notice to known peers
	broadcastDeparture()

	// Stop all protocol handlers
	stopWiFiDirectHandler()
	stopBluetoothHandler()
	stopTCPHandler()

	isRunning = false
}

// GetKnownPeers returns the list of known peers in the network
func GetKnownPeers() ([]Peer, error) {
	if !isRunning {
		return nil, errors.New("mesh node is not running")
	}

	peersMutex.RLock()
	defer peersMutex.RUnlock()

	peers := make([]Peer, 0, len(knownPeers))
	for _, peer := range knownPeers {
		peers = append(peers, *peer)
	}

	return peers, nil
}

// FindPeerByIdOrName locates a peer by either ID or name
func FindPeerByIdOrName(idOrName string) (*Peer, error) {
	if !isRunning {
		return nil, errors.New("mesh node is not running")
	}

	peersMutex.RLock()
	defer peersMutex.RUnlock()

	// First, try exact match on ID
	if peer, exists := knownPeers[idOrName]; exists {
		return peer, nil
	}

	// Next, try case-insensitive match on ID
	for id, peer := range knownPeers {
		if strings.EqualFold(id, idOrName) {
			return peer, nil
		}
	}

	// Finally, try name match (which might not be unique)
	var matchedPeer *Peer
	matchCount := 0

	for _, peer := range knownPeers {
		// Exact name match
		if peer.Name == idOrName {
			return peer, nil
		}

		// Case-insensitive name match
		if strings.EqualFold(peer.Name, idOrName) {
			matchedPeer = peer
			matchCount++
		}
	}

	if matchCount == 1 {
		return matchedPeer, nil
	} else if matchCount > 1 {
		return nil, fmt.Errorf("multiple peers found with name '%s'. Please use a specific ID", idOrName)
	}

	return nil, fmt.Errorf("no peer found with ID or name '%s'", idOrName)
}

// Helper functions
func generateNodeID() string {
	// Generate a unique node ID based on hardware and timestamp
	return fmt.Sprintf("node-%x", time.Now().UnixNano())
}

func startWiFiDirectHandler(port int) {
	// Initialize WiFi Direct service
	// This is a placeholder for the actual implementation
	fmt.Println("Starting WiFi Direct handler on port", port)
}

func startBluetoothHandler() {
	// Initialize Bluetooth service
	// This is a placeholder for the actual implementation
	fmt.Println("Starting Bluetooth handler")
}

func startTCPHandler(port int) {
	// Initialize TCP service for fallback
	// This is a placeholder for the actual implementation
	fmt.Println("Starting TCP handler on port", port)
}

func stopWiFiDirectHandler() {
	// Clean up WiFi Direct resources
}

func stopBluetoothHandler() {
	// Clean up Bluetooth resources
}

func stopTCPHandler() {
	// Clean up TCP resources
}

func startDiscoveryService() {
	// Periodically discover new peers
	for isRunning {
		// Discover peers using available protocols
		discoverPeers()
		time.Sleep(60 * time.Second)
	}
}

func discoverPeers() {
	// Implementation for peer discovery
}

func maintainRoutingTable() {
	// Periodically update routing information
	for isRunning {
		// Update routes
		updateRoutes()
		time.Sleep(30 * time.Second)
	}
}

func updateRoutes() {
	// Implementation for route maintenance
}

func broadcastDeparture() {
	// Let peers know we're leaving the network
}

func IsClientIsolated() bool {
	return connectionInfo.ClientIsolation
}

// GetConnectionInfo returns current network connection information
func GetConnectionInfo() ConnectionInfo {
	return connectionInfo
}

// GetNetworkMode returns the current networking mode
func GetNetworkMode() NetworkMode {
	return connectionInfo.Mode
}

// ConnectToPeer attempts to establish the best possible connection to a peer
func ConnectToPeer(peerID string) error {
	peer, err := FindPeerByIdOrName(peerID)
	if err != nil {
		return err
	}

	// Try direct connection first
	directErr := connectDirectly(peer)
	if directErr == nil {
		fmt.Printf("Direct connection established to %s (%s)\n", peer.Name, peer.ID)
		return nil
	}

	// If direct fails and client isolation is detected, try WiFi Direct
	if connectionInfo.ClientIsolation && meshConfig.EnableWiFiDirect {
		wifiErr := connectViaWiFiDirect(peer)
		if wifiErr == nil {
			fmt.Printf("WiFi Direct connection established to %s (%s)\n", peer.Name, peer.ID)
			return nil
		}
	}

	// If all direct methods fail, try relay if enabled
	if meshConfig.EnableRelay {
		relayErr := connectViaRelay(peer)
		if relayErr == nil {
			fmt.Printf("Relay connection established to %s (%s)\n", peer.Name, peer.ID)
			return nil
		}
		return fmt.Errorf("failed to connect via relay: %v", relayErr)
	}

	// If we get here, all connection attempts failed
	return fmt.Errorf("failed to connect: direct connection error: %v", directErr)
}

// Helper functions for client isolation handling

func detectNetworkConditions() {
	connectionInfo.LastConnectivityCheck = time.Now()

	// Get public IP if possible
	publicIP, _ := getPublicIP()
	if publicIP != "" {
		connectionInfo.PublicIP = publicIP
	}

	// Check for client isolation
	isolated := detectClientIsolation()
	connectionInfo.ClientIsolation = isolated

	// Determine NAT type
	natType, _ := detectNATType()
	connectionInfo.NATType = natType

	// Set network mode based on conditions
	if isolated {
		if meshConfig.EnableRelay {
			connectionInfo.Mode = RelayMode
		} else if meshConfig.EnableWiFiDirect {
			connectionInfo.Mode = MixedMode
		} else {
			connectionInfo.Mode = RelayMode // Fallback to relay even if disabled
		}
	} else {
		connectionInfo.Mode = DirectMode
	}

	// Check relay connectivity
	if meshConfig.EnableRelay && len(meshConfig.RelayServers) > 0 {
		connectionInfo.RelayAvailable = checkRelayConnectivity(meshConfig.RelayServers[0])
	}
}

func detectClientIsolation() bool {
	// Try to detect client isolation by:
	// 1. Find other devices on local network via multicast
	// 2. Try to establish direct TCP connections to them
	// 3. If all connection attempts fail, likely client isolation

	// Simplified implementation for now:
	// Try to make a connection to a common local IP
	conn, err := net.DialTimeout("tcp", "239.255.255.250:1900", 500*time.Millisecond)
	if err == nil {
		conn.Close()
		return false
	}

	// Try to connect to common gateway IP addresses
	for _, ip := range []string{"192.168.1.1", "192.168.0.1", "10.0.0.1"} {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", ip), 500*time.Millisecond)
		if err == nil {
			conn.Close()
			// We can reach gateway but not peers - likely isolation
			return true
		}
	}

	// If unsure, assume no isolation
	return false
}

func getPublicIP() (string, error) {
	// Connect to external service to get public IP
	// This is a simplified implementation
	conn, err := net.DialTimeout("tcp", "api.ipify.org:80", 5*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Use local IP as fallback
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	return localAddr.IP.String(), nil
}

func detectNATType() (string, error) {
	// In a real implementation, this would use STUN protocol
	// to determine NAT type (Full Cone, Restricted, etc.)
	return "Unknown", nil
}

func checkRelayConnectivity(relayServer string) bool {
	// Check if we can connect to the relay server
	conn, err := net.DialTimeout("tcp", relayServer, 5*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func monitorNetworkConditions() {
	for isRunning {
		time.Sleep(5 * time.Minute)
		detectNetworkConditions()
	}
}

func startRelayHandler(relayServers []string) {
	// Connect to relay servers and listen for relay requests
	fmt.Println("Starting relay connection handler")

	for _, server := range relayServers {
		go connectToRelayServer(server)
	}
}

func connectToRelayServer(server string) {
	// Connect to relay server and maintain connection
	fmt.Printf("Connecting to relay server: %s\n", server)

	// In a real implementation, this would:
	// 1. Establish and maintain a WebSocket or TCP connection to the relay server
	// 2. Register this node with its nodeID
	// 3. Listen for incoming connection requests via the relay
	// 4. Handle relay protocol for NAT traversal
}

func connectDirectly(peer *Peer) error {
	// Try to establish a direct TCP connection
	return errors.New("not implemented")
}

func connectViaWiFiDirect(peer *Peer) error {
	// Try to establish a WiFi Direct connection
	return errors.New("not implemented")
}

func connectViaRelay(peer *Peer) error {
	// Try to establish a connection via relay server
	return errors.New("not implemented")
}

// IsNodeRunning checks if the mesh node is currently running
func IsNodeRunning() bool {
	return isRunning
}

// GetNodeName returns the name of the current node
func GetNodeName() string {
	return meshConfig.NodeName
}

// GetNodeID returns the ID of the current node
func GetNodeID() string {
	return nodeID
}
