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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

type Config struct {
	Environment         string        `json:"environment"`
	ExcludeEnvironments []string      `json:"exclude_environments"`
	Apps                []App         `json:"apps"`
	Storage             []Storage     `json:"storage"`
	Registry            *App          `json:"registry,omitempty"`
	LoadBalancer        *LoadBalancer `json:"load_balancer,omitempty"`
}

type App struct {
	Namespace string `json:"namespace"`
	Ingress   string `json:"ingress"`
}

type LoadBalancer struct {
	Namespace    string `json:"namespace"`
	Service      string `json:"service"`
	IngressClass string `json:"ingress_class"`
}

type Storage struct {
	Class string                            `json:"class"`
	Mode  corev1.PersistentVolumeAccessMode `json:"access_mode"`
}

type Cluster struct {
	Initializer string `json:"init_host"`
	VIP         string `json:"vip"`
}

type Host struct {
	MAC     string `json:"mac_address"`
	IP      string `json:"ip"`
	IPv6    string `json:"ipv6_address"`
	Machine string `json:"machine_id_hash"`
}

type Target struct {
	Root    string
	Config  Config
	Cluster Cluster
	Hosts   map[string]Host
	Names   []string
}

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
		node := inv.hosts[name]
		if len(validation.IsDNS1123Subdomain(name)) != 0 {
			return inv, fmt.Errorf("invalid %s hostname %q", environment, name)
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
				return inv, fmt.Errorf("%s/%s needs a unique management IP distinct from the VIP", environment, name)
			}
			*field = address.String()
			local[address] = true
		}
		for address := range local {
			inv.addresses[address] = true
		}
		mac, err := net.ParseMAC(node.MAC)
		if err != nil || len(mac) != 6 || mac[0]&1 != 0 || inv.macs[mac.String()] {
			return inv, fmt.Errorf("%s/%s needs a unique unicast MAC address", environment, name)
		}
		node.MAC = mac.String()
		inv.macs[node.MAC] = true
		inv.hosts[name] = node
	}
	inv.addresses[vip] = true
	return inv, nil
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

func Load(root, configFile string) (Target, error) {
	target := Target{Root: root}
	if configFile == "" {
		return target, fmt.Errorf("config file must be nonempty")
	}
	if !filepath.IsAbs(configFile) {
		configFile = filepath.Join(root, "tests", configFile)
	}
	if err := readJSON(configFile, &target.Config); err != nil {
		return target, err
	}
	environments := append([]string{target.Config.Environment}, target.Config.ExcludeEnvironments...)
	seen := map[string]bool{}
	for _, environment := range environments {
		if len(validation.IsDNS1123Label(environment)) != 0 {
			return target, fmt.Errorf("invalid environment name %q", environment)
		}
		if seen[environment] {
			return target, fmt.Errorf("duplicate environment %q", environment)
		}
		seen[environment] = true
	}
	apps := map[App]bool{}
	ingresses := append([]App(nil), target.Config.Apps...)
	if target.Config.Registry != nil {
		ingresses = append(ingresses, *target.Config.Registry)
	}
	for _, app := range ingresses {
		if len(validation.IsDNS1123Label(app.Namespace)) != 0 {
			return target, fmt.Errorf("invalid app namespace %q", app.Namespace)
		}
		if len(validation.IsDNS1123Subdomain(app.Ingress)) != 0 {
			return target, fmt.Errorf("invalid app ingress %q", app.Ingress)
		}
		if apps[app] {
			return target, fmt.Errorf("duplicate app target %s/%s", app.Namespace, app.Ingress)
		}
		apps[app] = true
	}
	classes := map[string]bool{}
	for _, storage := range target.Config.Storage {
		if len(validation.IsDNS1123Subdomain(storage.Class)) != 0 || classes[storage.Class] {
			return target, fmt.Errorf("storage class must be valid and unique: %q", storage.Class)
		}
		if storage.Mode != corev1.ReadWriteOnce && storage.Mode != corev1.ReadWriteMany {
			return target, fmt.Errorf("storage class %s requires ReadWriteOnce or ReadWriteMany", storage.Class)
		}
		classes[storage.Class] = true
	}
	if lb := target.Config.LoadBalancer; lb != nil {
		if len(validation.IsDNS1123Label(lb.Namespace)) != 0 || len(validation.IsDNS1035Label(lb.Service)) != 0 || len(validation.IsDNS1123Subdomain(lb.IngressClass)) != 0 {
			return target, fmt.Errorf("load_balancer requires a valid namespace, service, and ingress_class")
		}
	}
	selected, err := loadInventory(root, target.Config.Environment)
	if err != nil {
		return target, err
	}
	for _, environment := range target.Config.ExcludeEnvironments {
		excluded, err := loadExcluded(root, environment)
		if err != nil {
			return target, err
		}
		for address := range selected.addresses {
			if excluded.addresses[address] {
				return target, fmt.Errorf("%s overlaps %s inventory at address %s", target.Config.Environment, environment, address)
			}
		}
		for mac := range selected.macs {
			if excluded.macs[mac] {
				return target, fmt.Errorf("%s overlaps %s inventory at MAC %s", target.Config.Environment, environment, mac)
			}
		}
	}
	target.Cluster, target.Hosts, target.Names = selected.cluster, selected.hosts, selected.names
	return target, nil
}

func ForTest(t *testing.T) Target {
	t.Helper()
	if os.Getenv("E2E") != "1" || testing.Short() {
		t.Skip("live tests require E2E=1 without -short")
	}
	config := os.Getenv("TEST_CONFIG")
	if config == "" {
		t.Fatal("TEST_CONFIG must be nonempty")
	}
	root, err := Root()
	if err != nil {
		t.Fatal(err)
	}
	target, err := Load(root, config)
	if err != nil {
		t.Fatal(err)
	}
	return target
}
