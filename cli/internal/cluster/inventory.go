package cluster

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type host struct {
	name    string
	address netip.AddrPort
}

type config struct {
	seed    host
	joiners []host
	vip     netip.Addr
}

func loadConfig(environment string) (config, error) {
	var cfg config
	if environment != "staging" && environment != "production" {
		return cfg, fmt.Errorf("--environment must be staging or production")
	}
	var cluster struct {
		Initializer string `json:"init_host"`
		VIP         string `json:"vip"`
	}
	if err := readJSON(filepath.Join("infra", environment, "cluster/config.json"), &cluster); err != nil {
		return cfg, err
	}
	var inventory map[string]struct {
		IP   string `json:"ip"`
		IPv6 string `json:"ipv6_address"`
	}
	if err := readJSON(filepath.Join("infra", environment, "metal/hosts.json"), &inventory); err != nil {
		return cfg, err
	}
	vip, err := netip.ParseAddr(cluster.VIP)
	if err != nil || !vip.Is4() || !vip.IsGlobalUnicast() {
		return cfg, fmt.Errorf("cluster VIP must be a unicast IPv4 address")
	}
	cfg.vip = vip
	validName := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)
	seen := make(map[netip.Addr]bool)
	for name, entry := range inventory {
		if len(name) > 63 || !validName.MatchString(name) {
			return cfg, fmt.Errorf("invalid node name %q", name)
		}
		address := entry.IP
		if address == "" {
			address = entry.IPv6
		}
		ip, err := netip.ParseAddr(address)
		if err != nil || !ip.IsGlobalUnicast() || ip == vip || seen[ip] {
			return cfg, fmt.Errorf("node %s needs a unique management IP from installation, distinct from the VIP", name)
		}
		seen[ip] = true
		node := host{name, netip.AddrPortFrom(ip, 22)}
		if name == cluster.Initializer {
			cfg.seed = node
		} else {
			cfg.joiners = append(cfg.joiners, node)
		}
	}
	if cfg.seed.name == "" {
		return cfg, fmt.Errorf("configured initializer is absent from inventory")
	}
	slices.SortFunc(cfg.joiners, func(a, b host) int { return strings.Compare(a.name, b.name) })
	return cfg, nil
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}
