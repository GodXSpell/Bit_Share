package p2p

import (
	"fmt"
	"time"
)

// PeerInfo contains information about a discovered peer
type PeerInfo struct {
	ID             string
	Name           string
	Address        string
	Protocol       string // "wifi-direct", "bluetooth", "tcp"
	SignalStrength int    // 0-100%
	LastSeen       time.Time
	Capabilities   []string
}

// ScanOptions configures the peer scan behavior
type ScanOptions struct {
	Timeout      time.Duration
	WifiDirect   bool
	Bluetooth    bool
	TCP          bool
	MaxDistance  int  // For protocols that support distance estimation
	IncludeCache bool // Include previously seen but currently unreachable peers
}

// DefaultScanOptions returns the default scan configuration
func DefaultScanOptions() ScanOptions {
	return ScanOptions{
		Timeout:      30 * time.Second,
		WifiDirect:   true,
		Bluetooth:    true,
		TCP:          true,
		MaxDistance:  100,
		IncludeCache: true,
	}
}

// ScanForPeers searches for peers using available protocols
func ScanForPeers() ([]PeerInfo, error) {
	return ScanForPeersWithOptions(DefaultScanOptions())
}

// ScanForPeersWithOptions searches for peers with custom options
func ScanForPeersWithOptions(options ScanOptions) ([]PeerInfo, error) {
	results := make([]PeerInfo, 0)
	errorsCh := make(chan error, 3)
	resultsCh := make(chan []PeerInfo, 3)

	// Create a timeout context
	doneCh := make(chan struct{})
	defer close(doneCh)

	// Launch scans for each protocol in parallel
	activeScanners := 0

	if options.WifiDirect {
		activeScanners++
		go func() {
			peers, err := scanWifiDirect(doneCh)
			if err != nil {
				errorsCh <- fmt.Errorf("WiFi Direct scan error: %w", err)
			} else {
				resultsCh <- peers
			}
		}()
	}

	if options.Bluetooth {
		activeScanners++
		go func() {
			peers, err := scanBluetooth(doneCh)
			if err != nil {
				errorsCh <- fmt.Errorf("Bluetooth scan error: %w", err)
			} else {
				resultsCh <- peers
			}
		}()
	}

	if options.TCP {
		activeScanners++
		go func() {
			peers, err := scanTCP(doneCh)
			if err != nil {
				errorsCh <- fmt.Errorf("TCP scan error: %w", err)
			} else {
				resultsCh <- peers
			}
		}()
	}

	// Wait for results with timeout
	timer := time.NewTimer(options.Timeout)
	defer timer.Stop()

	var errors []error
	for activeScanners > 0 {
		select {
		case err := <-errorsCh:
			errors = append(errors, err)
			activeScanners--
		case peers := <-resultsCh:
			results = append(results, peers...)
			activeScanners--
		case <-timer.C:
			return results, fmt.Errorf("scan timeout, partial results returned (errors: %v)", errors)
		}
	}

	// Include cached peers if requested
	if options.IncludeCache {
		cachedPeers := getCachedPeers()
		// Filter out duplicates
		for _, peer := range cachedPeers {
			isDuplicate := false
			for _, existing := range results {
				if peer.ID == existing.ID {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				peer.SignalStrength = 0 // Indicate this is a cached peer
				results = append(results, peer)
			}
		}
	}

	return results, nil
}

// Protocol-specific scan implementations
func scanWifiDirect(done <-chan struct{}) ([]PeerInfo, error) {
	// Implementation for WiFi Direct discovery
	// This is a placeholder for the actual implementation
	return []PeerInfo{
		{
			ID:             "wfd-1234",
			Name:           "Laptop-WFD",
			Address:        "192.168.49.1",
			Protocol:       "wifi-direct",
			SignalStrength: 80,
			LastSeen:       time.Now(),
			Capabilities:   []string{"transfer", "mesh"},
		},
	}, nil
}

func scanBluetooth(done <-chan struct{}) ([]PeerInfo, error) {
	// Implementation for Bluetooth discovery
	// This is a placeholder for the actual implementation
	return []PeerInfo{
		{
			ID:             "bt-5678",
			Name:           "Phone-BT",
			Address:        "00:11:22:33:44:55",
			Protocol:       "bluetooth",
			SignalStrength: 60,
			LastSeen:       time.Now(),
			Capabilities:   []string{"transfer"},
		},
	}, nil
}

func scanTCP(done <-chan struct{}) ([]PeerInfo, error) {
	// Implementation for TCP/IP discovery
	// This is a placeholder for the actual implementation
	return []PeerInfo{
		{
			ID:             "tcp-9012",
			Name:           "Desktop-TCP",
			Address:        "192.168.1.100",
			Protocol:       "tcp",
			SignalStrength: 100,
			LastSeen:       time.Now(),
			Capabilities:   []string{"transfer", "mesh", "relay"},
		},
	}, nil
}

func getCachedPeers() []PeerInfo {
	// Return previously discovered but currently offline peers
	// This is a placeholder for the actual implementation
	return []PeerInfo{}
}
