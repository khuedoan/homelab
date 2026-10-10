package testenv

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeJSON(t *testing.T, root, path string, value any) {
	t.Helper()
	file := filepath.Join(root, path)
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0700))
	data, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(file, data, 0600))
}

type fixture struct {
	config   Config
	selected Cluster
	excluded Cluster
	hosts    map[string]Host
	others   map[string]Host
}

func newFixture() fixture {
	return fixture{
		config:   Config{Environment: "dev", ExcludeEnvironments: []string{"prod"}, DNSDomain: "unit.invalid", GitOpsNamespace: "cd"},
		selected: Cluster{Initializer: "node-a", VIP: "192.0.2.100"},
		excluded: Cluster{Initializer: "node-b", VIP: "192.0.2.200"},
		hosts:    map[string]Host{"node-a": {IP: "192.0.2.10", MAC: "02:00:00:00:00:10"}},
		others:   map[string]Host{"node-b": {IP: "192.0.2.20", MAC: "02:00:00:00:00:20"}},
	}
}

func (f fixture) write(t *testing.T, root string) {
	t.Helper()
	writeJSON(t, root, "tests/config/target.json", f.config)
	writeJSON(t, root, "infra/dev/cluster/config.json", f.selected)
	writeJSON(t, root, "infra/dev/metal/hosts.json", f.hosts)
	writeJSON(t, root, "infra/prod/cluster/config.json", f.excluded)
	writeJSON(t, root, "infra/prod/metal/hosts.json", f.others)
}

func TestSafetyCases(t *testing.T) {
	for _, tc := range []struct {
		name, initializer, vip, address, mac, wantError string
	}{
		{"valid", "node-a", "192.0.2.100", "192.0.2.10", "02:00:00:00:00:10", ""},
		{"missing initializer", "missing", "192.0.2.100", "192.0.2.10", "02:00:00:00:00:10", "dev initializer must exist in inventory"},
		{"excluded VIP", "node-a", "192.0.2.200", "192.0.2.10", "02:00:00:00:00:10", "overlaps prod inventory at address 192.0.2.200"},
		{"excluded IP as VIP", "node-a", "192.0.2.20", "192.0.2.10", "02:00:00:00:00:10", "overlaps prod inventory at address 192.0.2.20"},
		{"excluded IP", "node-a", "192.0.2.100", "192.0.2.20", "02:00:00:00:00:10", "overlaps prod inventory at address 192.0.2.20"},
		{"excluded MAC", "node-a", "192.0.2.100", "192.0.2.10", "02:00:00:00:00:20", "overlaps prod inventory at MAC 02:00:00:00:00:20"},
		{"missing discovered IP", "node-a", "192.0.2.100", "", "02:00:00:00:00:10", "unique management IP distinct from the VIP"},
		{"VIP as management IP", "node-a", "192.0.2.100", "192.0.2.100", "02:00:00:00:00:10", "unique management IP distinct from the VIP"},
		{"invalid MAC", "node-a", "192.0.2.100", "192.0.2.10", "not-a-mac", "unique unicast MAC address"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, root := newFixture(), t.TempDir()
			f.selected = Cluster{Initializer: tc.initializer, VIP: tc.vip}
			f.hosts["node-a"] = Host{IP: tc.address, MAC: tc.mac}
			f.write(t, root)
			target, err := Load(root, "config/target.json")
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, Target{Root: root, Config: f.config, Cluster: Cluster{Initializer: "node-a", VIP: "192.0.2.100"}, Hosts: map[string]Host{"node-a": {IP: "192.0.2.10", MAC: "02:00:00:00:00:10"}}, Names: []string{"node-a"}}, target)
		})
	}
}

func TestConfigValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		config    Config
		wantError string
	}{
		{"missing environment", Config{}, "invalid environment name"},
		{"traversal", Config{Environment: "../dev"}, "invalid environment name"},
		{"absolute environment", Config{Environment: "/dev"}, "invalid environment name"},
		{"excluded traversal", Config{Environment: "dev", ExcludeEnvironments: []string{"../prod"}}, "invalid environment name"},
		{"self exclusion", Config{Environment: "dev", ExcludeEnvironments: []string{"dev"}}, "duplicate environment"},
		{"duplicate exclusion", Config{Environment: "dev", ExcludeEnvironments: []string{"prod", "prod"}}, "duplicate environment"},
		{"invalid storage class", Config{Environment: "dev", Storage: []Storage{{Class: "../disk", Mode: "ReadWriteOnce"}}}, "storage class must be valid and unique"},
		{"empty storage mode", Config{Environment: "dev", Storage: []Storage{{Class: "disk"}}}, "requires ReadWriteOnce or ReadWriteMany"},
		{"unknown storage mode", Config{Environment: "dev", Storage: []Storage{{Class: "disk", Mode: "RWX"}}}, "requires ReadWriteOnce or ReadWriteMany"},
		{"duplicate storage class", Config{Environment: "dev", Storage: []Storage{{Class: "disk", Mode: "ReadWriteOnce"}, {Class: "disk", Mode: "ReadWriteMany"}}}, "storage class must be valid and unique"},
		{"empty load balancer", Config{Environment: "dev", LoadBalancer: &LoadBalancer{}}, "load_balancer requires a valid namespace, service, gateway, and listener"},
		{"invalid GitOps namespace", Config{Environment: "dev", GitOpsNamespace: "../cd"}, "invalid GitOps namespace"},
		{"missing DNS domain", Config{Environment: "dev"}, "invalid DNS domain"},
		{"invalid DNS domain", Config{Environment: "dev", DNSDomain: "../cluster.local"}, "invalid DNS domain"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			writeJSON(t, root, "tests/config/target.json", tc.config)
			_, err := Load(root, "config/target.json")
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestLoadBalancerValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		lb   LoadBalancer
	}{
		{"empty namespace", LoadBalancer{Service: "controller", Gateway: "edge", Listener: "http"}},
		{"invalid namespace", LoadBalancer{Namespace: "bad.name", Service: "controller", Gateway: "edge", Listener: "http"}},
		{"empty service", LoadBalancer{Namespace: "edge", Gateway: "edge", Listener: "http"}},
		{"invalid service", LoadBalancer{Namespace: "edge", Service: "../lb", Gateway: "edge", Listener: "http"}},
		{"empty gateway", LoadBalancer{Namespace: "edge", Service: "controller", Listener: "http"}},
		{"invalid gateway", LoadBalancer{Namespace: "edge", Service: "controller", Gateway: "../edge", Listener: "http"}},
		{"empty listener", LoadBalancer{Namespace: "edge", Service: "controller", Gateway: "edge"}},
		{"invalid listener", LoadBalancer{Namespace: "edge", Service: "controller", Gateway: "edge", Listener: "bad.name"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, root := newFixture(), t.TempDir()
			f.config.LoadBalancer = &tc.lb
			f.write(t, root)
			_, err := Load(root, "config/target.json")
			require.EqualError(t, err, "load_balancer requires a valid namespace, service, gateway, and listener")
		})
	}
}

func TestInventoryValidation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		change    func(*fixture)
		wantError string
	}{
		{"IPv6 VIP", func(f *fixture) { f.selected.VIP = "2001:db8::100" }, "unicast IPv4 VIP"},
		{"broadcast VIP", func(f *fixture) { f.selected.VIP = "255.255.255.255" }, "unicast IPv4 VIP"},
		{"multicast VIP", func(f *fixture) { f.selected.VIP = "224.0.0.1" }, "unicast IPv4 VIP"},
		{"loopback address", func(f *fixture) { h := f.hosts["node-a"]; h.IP = "127.0.0.1"; f.hosts["node-a"] = h }, "unique management IP"},
		{"unspecified address", func(f *fixture) { h := f.hosts["node-a"]; h.IP = "::"; f.hosts["node-a"] = h }, "unique management IP"},
		{"multicast MAC", func(f *fixture) { h := f.hosts["node-a"]; h.MAC = "01:00:00:00:00:10"; f.hosts["node-a"] = h }, "unique unicast MAC"},
		{"bad hostname", func(f *fixture) { f.hosts["../bad"] = f.hosts["node-a"] }, "invalid dev hostname"},
		{"duplicate address", func(f *fixture) { f.hosts["node-c"] = Host{IP: "192.0.2.10", MAC: "02:00:00:00:00:30"} }, "unique management IP"},
		{"duplicate normalized MAC", func(f *fixture) { f.hosts["node-c"] = Host{IP: "192.0.2.30", MAC: "02-00-00-00-00-10"} }, "unique unicast MAC"},
		{"excluded unsafe address", func(f *fixture) { f.excluded.VIP = "127.0.0.1" }, "prod has an invalid excluded address"},
		{"excluded unsafe MAC", func(f *fixture) { h := f.others["node-b"]; h.MAC = "invalid"; f.others["node-b"] = h }, "prod has an invalid excluded MAC"},
		{"selected IP equals excluded VIP", func(f *fixture) { h := f.hosts["node-a"]; h.IP = "192.0.2.200"; f.hosts["node-a"] = h }, "overlaps prod inventory at address 192.0.2.200"},
		{"primary overlaps another secondary", func(f *fixture) {
			f.hosts["node-c"] = Host{IP: "192.0.2.30", IPv6: "::ffff:192.0.2.10", MAC: "02:00:00:00:00:30"}
		}, "dev/node-c needs a unique management IP distinct from the VIP"},
		{"excluded secondary IPv6 overlap", func(f *fixture) {
			h := f.hosts["node-a"]
			h.IPv6 = "2001:db8::20"
			f.hosts["node-a"] = h
			h = f.others["node-b"]
			h.IPv6 = "2001:0db8:0:0::20"
			f.others["node-b"] = h
		}, "overlaps prod inventory at address 2001:db8::20"},
		{"secondary IP equals VIP", func(f *fixture) { h := f.hosts["node-a"]; h.IPv6 = "::ffff:192.0.2.100"; f.hosts["node-a"] = h }, "unique management IP distinct from the VIP"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, root := newFixture(), t.TempDir()
			tc.change(&f)
			f.write(t, root)
			_, err := Load(root, "config/target.json")
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestMappedIPv4Inventory(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.selected.VIP = "::ffff:192.0.2.100"
	f.hosts["node-a"] = Host{IP: "192.0.2.10", IPv6: "::ffff:192.0.2.10", MAC: "02-AB-00-00-00-10"}
	f.write(t, root)
	target, err := Load(root, "config/target.json")
	require.NoError(t, err)
	require.Equal(t, Cluster{Initializer: "node-a", VIP: "192.0.2.100"}, target.Cluster)
	require.Equal(t, map[string]Host{"node-a": {IP: "192.0.2.10", IPv6: "192.0.2.10", MAC: "02:ab:00:00:00:10"}}, target.Hosts)
	f.excluded.VIP = "::ffff:192.0.2.10"
	f.write(t, root)
	_, err = Load(root, "config/target.json")
	require.EqualError(t, err, "dev overlaps prod inventory at address 192.0.2.10")
}

func TestCapabilities(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.write(t, root)
	writeJSON(t, root, "tests/config/target.json", json.RawMessage(`{
		"environment": "dev",
		"exclude_environments": ["prod"],
		"dns_domain": "unit.invalid",
		"storage": [
			{"class": "block", "access_mode": "ReadWriteOnce"},
			{"class": "shared", "access_mode": "ReadWriteMany"}
		],
		"load_balancer": {"namespace": "edge", "service": "controller", "gateway": "edge", "listener": "http"},
		"gitops_namespace": "cd"
	}`))
	target, err := Load(root, "config/target.json")
	require.NoError(t, err)
	require.Equal(t, Config{
		Environment: "dev", ExcludeEnvironments: []string{"prod"},
		DNSDomain:       "unit.invalid",
		Storage:         []Storage{{Class: "block", Mode: "ReadWriteOnce"}, {Class: "shared", Mode: "ReadWriteMany"}},
		LoadBalancer:    &LoadBalancer{Namespace: "edge", Service: "controller", Gateway: "edge", Listener: "http"},
		GitOpsNamespace: "cd",
	}, target.Config)
}

func TestExcludedInventoryWithoutDiscovery(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.others = map[string]Host{
		"metal": {MAC: "02:00:00:00:00:20"},
		"cloud": {IPv6: "2001:db8::20"},
	}
	f.write(t, root)
	target, err := Load(root, "config/target.json")
	require.NoError(t, err)
	require.Equal(t, []string{"node-a"}, target.Names)
	host := f.hosts["node-a"]
	host.MAC = "02:00:00:00:00:20"
	f.hosts["node-a"] = host
	f.write(t, root)
	_, err = Load(root, "config/target.json")
	require.ErrorContains(t, err, "overlaps prod inventory at MAC 02:00:00:00:00:20")
}

func TestNormalizationAndPaths(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.hosts["node-a"] = Host{IPv6: "2001:0db8:0:0::10", MAC: "02-AB-00-00-00-10", Machine: "machine-hash"}
	f.hosts["aaa"] = Host{IP: "192.0.2.30", MAC: "02:00:00:00:00:30"}
	f.write(t, root)
	for _, path := range []string{"config/target.json", filepath.Join(root, "tests/config/target.json")} {
		target, err := Load(root, path)
		require.NoError(t, err)
		require.Equal(t, []string{"aaa", "node-a"}, target.Names)
		require.Equal(t, Host{IP: "2001:db8::10", IPv6: "2001:db8::10", MAC: "02:ab:00:00:00:10", Machine: "machine-hash"}, target.Hosts["node-a"])
	}
	_, err := Load(root, "")
	require.EqualError(t, err, "config file must be nonempty")
	_, err = Load(root, "missing.json")
	require.ErrorContains(t, err, "read "+filepath.Join(root, "tests/missing.json"))
	require.NoError(t, os.WriteFile(filepath.Join(root, "tests/config/target.json"), []byte("{"), 0600))
	_, err = Load(root, "config/target.json")
	require.ErrorContains(t, err, "parse "+filepath.Join(root, "tests/config/target.json"))
}

func TestMissingExcludedInventory(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.config.ExcludeEnvironments = append(f.config.ExcludeEnvironments, "other")
	f.write(t, root)
	_, err := Load(root, "config/target.json")
	require.ErrorContains(t, err, "infra/other/cluster/config.json")
}

func TestRoot(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "tests/internal/nested"), 0700))
	require.NoError(t, os.Mkdir(filepath.Join(root, "infra"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "tests/go.mod"), []byte("module example"), 0600))
	t.Chdir(filepath.Join(root, "tests/internal/nested"))
	found, err := Root()
	require.NoError(t, err)
	require.Equal(t, root, found)
	require.NoError(t, os.Remove(filepath.Join(root, "tests/go.mod")))
	_, err = Root()
	require.ErrorContains(t, err, "repository root not found")
}

func TestForTestSkipsOffline(t *testing.T) {
	for _, e2e := range []string{"", "0", "1"} {
		t.Run("E2E="+e2e, func(t *testing.T) {
			if e2e == "1" && !testing.Short() {
				t.Skip("this case requires -short")
			}
			t.Setenv("E2E", e2e)
			t.Setenv("TEST_CONFIG", "does-not-exist.json")
			reached := false
			t.Run("gate", func(t *testing.T) { ForTest(t); reached = true })
			require.False(t, reached, "offline gate must skip before config loading")
		})
	}
}
