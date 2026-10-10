package testenv

import (
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation"
)

// Cluster identifies the inventory member that initializes the API and its IPv4 VIP.
type Cluster struct {
	Initializer string `json:"init_host"`
	VIP         string `json:"vip"`
}

// Host records management addresses and identities checked over trusted SSH.
// Machine is an optional enrollment machine-ID digest, not the raw machine ID.
type Host struct {
	MAC     string `json:"mac_address"`
	IP      string `json:"ip"`
	IPv6    string `json:"ipv6_address"`
	Machine string `json:"machine_id_hash"`
}

// Target is a validated, normalized inventory with host names in sorted order.
type Target struct {
	Root        string
	Environment string
	Cluster     Cluster
	Hosts       map[string]Host
	Names       []string
}

// Root finds the repository containing tests/go.mod and infra above the working directory.
func Root() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		mod, modErr := os.Stat(filepath.Join(dir, "tests", "go.mod"))
		infra, infraErr := os.Stat(filepath.Join(dir, "infra"))
		if modErr == nil && !mod.IsDir() && infraErr == nil && infra.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repository root not found: need tests/go.mod and infra directory")
		}
		dir = parent
	}
}

func readJSON(path string, destination any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, destination); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

type inventory struct {
	cluster   Cluster
	hosts     map[string]Host
	names     []string
	addresses map[netip.Addr]bool
	macs      map[string]bool
}

func managementAddress(raw string) (netip.Addr, error) {
	address, err := netip.ParseAddr(raw)
	if err != nil {
		return netip.Addr{}, err
	}
	address = address.Unmap()
	if address.Zone() != "" || address.IsUnspecified() || address.IsLoopback() || address.IsMulticast() || address == netip.MustParseAddr("255.255.255.255") {
		return netip.Addr{}, fmt.Errorf("not a unicast management address")
	}
	return address, nil
}

func loadInventory(root, environment string) (inventory, error) {
	inv := inventory{addresses: map[netip.Addr]bool{}, macs: map[string]bool{}}
	base := filepath.Join(root, "infra", environment)
	if err := readJSON(filepath.Join(base, "cluster", "config.json"), &inv.cluster); err != nil {
		return inv, err
	}
	if err := readJSON(filepath.Join(base, "metal", "hosts.json"), &inv.hosts); err != nil {
		return inv, err
	}
	if _, exists := inv.hosts[inv.cluster.Initializer]; !exists {
		return inv, fmt.Errorf("%s initializer must exist in inventory", environment)
	}
	vip, err := managementAddress(inv.cluster.VIP)
	if err != nil || !vip.Is4() {
		return inv, fmt.Errorf("%s requires a unicast IPv4 VIP", environment)
	}
	inv.cluster.VIP = vip.String()
	for name := range inv.hosts {
		inv.names = append(inv.names, name)
	}
	sort.Strings(inv.names)
	for _, name := range inv.names {
		node, err := inv.normalizeHost(environment, name, vip)
		if err != nil {
			return inv, err
		}
		inv.hosts[name] = node
	}
	inv.addresses[vip] = true
	return inv, nil
}

func (inv inventory) normalizeHost(environment, name string, vip netip.Addr) (Host, error) {
	node := inv.hosts[name]
	if len(validation.IsDNS1123Subdomain(name)) != 0 {
		return node, fmt.Errorf("invalid %s hostname %q", environment, name)
	}
	if node.IP == "" {
		node.IP = node.IPv6
	}
	local := map[netip.Addr]bool{}
	for i, field := range []*string{&node.IP, &node.IPv6} {
		if i == 1 && *field == "" {
			continue
		}
		address, err := managementAddress(*field)
		if err != nil || address == vip || inv.addresses[address] {
			return node, fmt.Errorf("%s/%s needs a unique management IP distinct from the VIP", environment, name)
		}
		*field = address.String()
		local[address] = true
	}
	for address := range local {
		inv.addresses[address] = true
	}
	mac, err := net.ParseMAC(node.MAC)
	if err != nil || len(mac) != 6 || mac[0]&1 != 0 || inv.macs[mac.String()] {
		return node, fmt.Errorf("%s/%s needs a unique unicast MAC address", environment, name)
	}
	node.MAC = mac.String()
	inv.macs[node.MAC] = true
	return node, nil
}

func loadExcluded(root, environment string) (inventory, error) {
	inv := inventory{addresses: map[netip.Addr]bool{}, macs: map[string]bool{}}
	base := filepath.Join(root, "infra", environment)
	if err := readJSON(filepath.Join(base, "cluster", "config.json"), &inv.cluster); err != nil {
		return inv, err
	}
	if err := readJSON(filepath.Join(base, "metal", "hosts.json"), &inv.hosts); err != nil {
		return inv, err
	}
	addresses := []string{inv.cluster.VIP}
	for _, host := range inv.hosts {
		addresses = append(addresses, host.IP, host.IPv6)
		if host.MAC != "" {
			mac, err := net.ParseMAC(host.MAC)
			if err != nil {
				return inv, fmt.Errorf("%s has an invalid excluded MAC", environment)
			}
			inv.macs[mac.String()] = true
		}
	}
	for _, raw := range addresses {
		if raw == "" {
			continue
		}
		address, err := managementAddress(raw)
		if err != nil {
			return inv, fmt.Errorf("%s has an invalid excluded address", environment)
		}
		inv.addresses[address] = true
	}
	return inv, nil
}

// Load validates an environment inventory and rejects overlap with other environments.
func Load(root, environment string) (Target, error) {
	target := Target{Root: root, Environment: environment}
	if len(validation.IsDNS1123Label(environment)) != 0 {
		return target, fmt.Errorf("invalid environment name %q", environment)
	}
	selected, err := loadInventory(root, environment)
	if err != nil {
		return target, err
	}
	environments, err := filepath.Glob(filepath.Join(root, "infra", "*", "root.hcl"))
	if err != nil {
		return target, err
	}
	for _, path := range environments {
		other := filepath.Base(filepath.Dir(path))
		if other == environment {
			continue
		}
		excluded, err := loadExcluded(root, other)
		if err != nil {
			return target, err
		}
		for address := range selected.addresses {
			if excluded.addresses[address] {
				return target, fmt.Errorf("%s overlaps %s inventory at address %s", environment, other, address)
			}
		}
		for mac := range selected.macs {
			if excluded.macs[mac] {
				return target, fmt.Errorf("%s overlaps %s inventory at MAC %s", environment, other, mac)
			}
		}
	}
	target.Cluster, target.Hosts, target.Names = selected.cluster, selected.hosts, selected.names
	return target, nil
}

// ForTest requires an explicit environment for live tests and skips offline runs.
func ForTest(t *testing.T) Target {
	t.Helper()
	if os.Getenv("E2E") != "1" || testing.Short() {
		t.Skip("live tests require E2E=1 without -short")
	}
	environment := os.Getenv("TEST_ENV")
	if environment == "" {
		t.Fatal("TEST_ENV must be nonempty")
	}
	root, err := Root()
	if err != nil {
		t.Fatal(err)
	}
	target, err := Load(root, environment)
	if err != nil {
		t.Fatal(err)
	}
	return target
}
