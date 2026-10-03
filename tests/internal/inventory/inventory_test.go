package inventory

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

type inventoryFixture struct {
	selected Cluster
	excluded Cluster
	hosts    map[string]Host
	others   map[string]Host
}

func newFixture() inventoryFixture {
	return inventoryFixture{
		selected: Cluster{
			Initializer: "node-a",
			VIP:         "192.0.2.100",
		},
		excluded: Cluster{
			Initializer: "node-b",
			VIP:         "192.0.2.200",
		},
		hosts: map[string]Host{
			"node-a": {
				IP:  "192.0.2.10",
				MAC: "02:00:00:00:00:10",
			},
		},
		others: map[string]Host{
			"node-b": {
				IP:  "192.0.2.20",
				MAC: "02:00:00:00:00:20",
			},
		},
	}
}

func (f inventoryFixture) write(t *testing.T, root string) {
	t.Helper()

	writeJSON(t, root, "infra/dev/cluster/config.json", f.selected)
	writeJSON(t, root, "infra/dev/metal/hosts.json", f.hosts)
	writeJSON(t, root, "infra/prod/cluster/config.json", f.excluded)
	writeJSON(t, root, "infra/prod/metal/hosts.json", f.others)
}

func TestLoadValidInventory(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.write(t, root)

	target, err := Load(root, "dev")

	require.NoError(t, err)
	require.Equal(t, Target{
		Root:        root,
		Environment: "dev",
		Cluster: Cluster{
			Initializer: "node-a",
			VIP:         "192.0.2.100",
		},
		Hosts: map[string]Host{
			"node-a": {
				IP:  "192.0.2.10",
				MAC: "02:00:00:00:00:10",
			},
		},
		Names: []string{"node-a"},
	}, target)
}

func TestLoadRejectsMissingInitializer(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.selected.Initializer = "missing"
	f.write(t, root)

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "dev initializer must exist in inventory")
}

func TestLoadRejectsInvalidVIP(t *testing.T) {
	for _, tc := range []struct {
		name string
		vip  string
	}{
		{name: "IPv6", vip: "2001:db8::100"},
		{name: "broadcast", vip: "255.255.255.255"},
		{name: "multicast", vip: "224.0.0.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, root := newFixture(), t.TempDir()
			f.selected.VIP = tc.vip
			f.write(t, root)

			_, err := Load(root, "dev")

			require.ErrorContains(t, err, "unicast IPv4 VIP")
		})
	}
}

func TestLoadRejectsInvalidManagementAddress(t *testing.T) {
	for _, tc := range []struct {
		name    string
		address string
	}{
		{name: "missing discovered IP", address: ""},
		{name: "VIP as management IP", address: "192.0.2.100"},
		{name: "loopback", address: "127.0.0.1"},
		{name: "unspecified", address: "::"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, root := newFixture(), t.TempDir()
			host := f.hosts["node-a"]
			host.IP = tc.address
			f.hosts["node-a"] = host
			f.write(t, root)

			_, err := Load(root, "dev")

			require.ErrorContains(t, err, "unique management IP distinct from the VIP")
		})
	}
}

func TestLoadRejectsInvalidMAC(t *testing.T) {
	for _, tc := range []struct {
		name string
		mac  string
	}{
		{name: "invalid", mac: "not-a-mac"},
		{name: "multicast", mac: "01:00:00:00:00:10"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, root := newFixture(), t.TempDir()
			host := f.hosts["node-a"]
			host.MAC = tc.mac
			f.hosts["node-a"] = host
			f.write(t, root)

			_, err := Load(root, "dev")

			require.ErrorContains(t, err, "unique unicast MAC address")
		})
	}
}

func TestLoadRejectsInvalidHostname(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.hosts["../bad"] = f.hosts["node-a"]
	f.write(t, root)

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "invalid dev hostname")
}

func TestLoadRejectsDuplicateHostIdentity(t *testing.T) {
	for _, tc := range []struct {
		name      string
		host      Host
		wantError string
	}{
		{
			name:      "management address",
			host:      Host{IP: "192.0.2.10", MAC: "02:00:00:00:00:30"},
			wantError: "unique management IP",
		},
		{
			name:      "normalized MAC",
			host:      Host{IP: "192.0.2.30", MAC: "02-00-00-00-00-10"},
			wantError: "unique unicast MAC",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, root := newFixture(), t.TempDir()
			f.hosts["node-c"] = tc.host
			f.write(t, root)

			_, err := Load(root, "dev")

			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestLoadRejectsSecondaryAddressMatchingAnotherHost(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.hosts["node-c"] = Host{
		IP:   "192.0.2.30",
		IPv6: "::ffff:192.0.2.10",
		MAC:  "02:00:00:00:00:30",
	}
	f.write(t, root)

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "dev/node-c needs a unique management IP distinct from the VIP")
}

func TestLoadRejectsSecondaryAddressMatchingVIP(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	host := f.hosts["node-a"]
	host.IPv6 = "::ffff:192.0.2.100"
	f.hosts["node-a"] = host
	f.write(t, root)

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "unique management IP distinct from the VIP")
}

func TestLoadRejectsVIPOverlapWithExcludedEnvironment(t *testing.T) {
	for _, tc := range []struct {
		name      string
		vip       string
		wantError string
	}{
		{
			name:      "excluded VIP",
			vip:       "192.0.2.200",
			wantError: "overlaps prod inventory at address 192.0.2.200",
		},
		{
			name:      "excluded host",
			vip:       "192.0.2.20",
			wantError: "overlaps prod inventory at address 192.0.2.20",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, root := newFixture(), t.TempDir()
			f.selected.VIP = tc.vip
			f.write(t, root)

			_, err := Load(root, "dev")

			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestLoadRejectsManagementAddressOverlapWithExcludedEnvironment(t *testing.T) {
	for _, tc := range []struct {
		name      string
		address   string
		wantError string
	}{
		{
			name:      "excluded host",
			address:   "192.0.2.20",
			wantError: "overlaps prod inventory at address 192.0.2.20",
		},
		{
			name:      "excluded VIP",
			address:   "192.0.2.200",
			wantError: "overlaps prod inventory at address 192.0.2.200",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, root := newFixture(), t.TempDir()
			host := f.hosts["node-a"]
			host.IP = tc.address
			f.hosts["node-a"] = host
			f.write(t, root)

			_, err := Load(root, "dev")

			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestLoadRejectsMACOverlapWithExcludedEnvironment(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	host := f.hosts["node-a"]
	host.MAC = "02:00:00:00:00:20"
	f.hosts["node-a"] = host
	f.write(t, root)

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "overlaps prod inventory at MAC 02:00:00:00:00:20")
}

func TestLoadRejectsNormalizedIPv6OverlapWithExcludedEnvironment(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	host := f.hosts["node-a"]
	host.IPv6 = "2001:db8::20"
	f.hosts["node-a"] = host
	host = f.others["node-b"]
	host.IPv6 = "2001:0db8:0:0::20"
	f.others["node-b"] = host
	f.write(t, root)

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "overlaps prod inventory at address 2001:db8::20")
}

func TestLoadRejectsUnsafeExcludedAddress(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.excluded.VIP = "127.0.0.1"
	f.write(t, root)

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "prod has an invalid excluded address")
}

func TestLoadRejectsInvalidExcludedMAC(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	host := f.others["node-b"]
	host.MAC = "invalid"
	f.others["node-b"] = host
	f.write(t, root)

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "prod has an invalid excluded MAC")
}

func TestLoadNormalizesMappedIPv4Inventory(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.selected.VIP = "::ffff:192.0.2.100"
	f.hosts["node-a"] = Host{
		IP:   "192.0.2.10",
		IPv6: "::ffff:192.0.2.10",
		MAC:  "02-AB-00-00-00-10",
	}
	f.write(t, root)
	target, err := Load(root, "dev")
	require.NoError(t, err)
	require.Equal(t, Cluster{
		Initializer: "node-a",
		VIP:         "192.0.2.100",
	}, target.Cluster)
	require.Equal(
		t,
		map[string]Host{
			"node-a": {
				IP:   "192.0.2.10",
				IPv6: "192.0.2.10",
				MAC:  "02:ab:00:00:00:10",
			},
		},
		target.Hosts,
	)
}

func TestLoadRejectsMappedIPv4OverlapWithExcludedEnvironment(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.excluded.VIP = "::ffff:192.0.2.10"
	f.write(t, root)

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "dev overlaps prod inventory at address 192.0.2.10")
}

func TestLoadAllowsExcludedInventoryWithoutDiscovery(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.others = map[string]Host{
		"metal": {MAC: "02:00:00:00:00:20"},
		"cloud": {IPv6: "2001:db8::20"},
	}
	f.write(t, root)

	target, err := Load(root, "dev")

	require.NoError(t, err)
	require.Equal(t, []string{"node-a"}, target.Names)
}

func TestLoadRejectsMACOverlapWithUndiscoveredExcludedHost(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.others = map[string]Host{
		"metal": {MAC: "02:00:00:00:00:20"},
		"cloud": {IPv6: "2001:db8::20"},
	}
	host := f.hosts["node-a"]
	host.MAC = "02:00:00:00:00:20"
	f.hosts["node-a"] = host
	f.write(t, root)

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "overlaps prod inventory at MAC 02:00:00:00:00:20")
}

func TestEnvironmentValidation(t *testing.T) {
	for _, environment := range []string{"", "../dev", "/dev", "bad.name"} {
		_, err := Load(t.TempDir(), environment)
		require.ErrorContains(t, err, "invalid environment name")
	}
}

func TestLoadNormalizesIPv6FallbackAndMAC(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.hosts["node-a"] = Host{
		IPv6: "2001:0db8:0:0::10",
		MAC:  "02-AB-00-00-00-10",
	}
	f.write(t, root)

	target, err := Load(root, "dev")

	require.NoError(t, err)
	require.Equal(t, Host{
		IP:   "2001:db8::10",
		IPv6: "2001:db8::10",
		MAC:  "02:ab:00:00:00:10",
	}, target.Hosts["node-a"])
}

func TestLoadSortsHostNames(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.hosts["aaa"] = Host{
		IP:  "192.0.2.30",
		MAC: "02:00:00:00:00:30",
	}
	f.write(t, root)

	target, err := Load(root, "dev")

	require.NoError(t, err)
	require.Equal(t, []string{"aaa", "node-a"}, target.Names)
}

func TestLoadReportsMissingSelectedInventory(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.write(t, root)

	_, err := Load(root, "missing")

	require.ErrorContains(t, err, "infra/missing/cluster/config.json")
}

func TestLoadReportsMalformedJSON(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.write(t, root)
	config := filepath.Join(root, "infra/dev/cluster/config.json")
	require.NoError(t, os.WriteFile(config, []byte("{"), 0600))

	_, err := Load(root, "dev")

	require.ErrorContains(t, err, "parse "+config)
}

func TestMissingExcludedInventory(t *testing.T) {
	f, root := newFixture(), t.TempDir()
	f.write(t, root)
	require.NoError(t, os.MkdirAll(filepath.Join(root, "infra/other/metal"), 0700))
	_, err := Load(root, "dev")
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
			t.Setenv("TEST_ENV", "does-not-exist")
			reached := false
			t.Run("gate", func(t *testing.T) {
				ForTest(t)
				reached = true
			})
			require.False(t, reached, "offline gate must skip before inventory loading")
		})
	}
}
