package firewall

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
)

// FirewallRule represents a temporary firewall rule.
type FirewallRule struct {
	Name string
	Port int
}

// AddTempRule adds a temporary firewall rule to allow incoming TCP traffic on a specific port.
// It returns a FirewallRule object that can be used to remove the rule later.
func AddTempRule(port int) (*FirewallRule, error) {
	ruleName := fmt.Sprintf("fileshare-port-%d", port)
	rule := &FirewallRule{
		Name: ruleName,
		Port: port,
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		portStr := strconv.Itoa(port)
		// Command to add a rule to Windows Defender Firewall
		cmd = exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
			"name="+ruleName, "dir=in", "action=allow", "protocol=TCP", "localport="+portStr)
	case "linux":
		// Placeholder for Linux (e.g., using iptables or ufw)
		// sudo ufw allow <port>/tcp
		return nil, fmt.Errorf("firewall management not implemented for Linux")
	case "darwin":
		// Placeholder for macOS (e.g., using pfctl)
		return nil, fmt.Errorf("firewall management not implemented for macOS")
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to add firewall rule: %v. Try running as administrator", err)
	}

	return rule, nil
}

// RemoveRule removes the firewall rule that was previously added.
func (r *FirewallRule) RemoveRule() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Command to remove the rule from Windows Defender Firewall
		cmd = exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name="+r.Name)
	case "linux":
		return fmt.Errorf("firewall management not implemented for Linux")
	case "darwin":
		return fmt.Errorf("firewall management not implemented for macOS")
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove firewall rule: %v. Try running as administrator", err)
	}

	return nil
}
