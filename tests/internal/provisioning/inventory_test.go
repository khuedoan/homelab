package provisioning

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestActivePXEInventoriesHaveMACAddresses(t *testing.T) {
	for _, environment := range []string{"production", "staging"} {
		t.Run(environment, func(t *testing.T) {
			path := filepath.Join("../../../infra", environment, "metal/hosts.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var hosts map[string]struct {
				MAC string `json:"mac_address"`
			}
			if err := json.Unmarshal(data, &hosts); err != nil {
				t.Fatal(err)
			}
			if len(hosts) == 0 {
				t.Fatal("empty PXE inventory")
			}
			for name, host := range hosts {
				if _, err := net.ParseMAC(host.MAC); err != nil {
					t.Errorf("%s has no valid PXE MAC: %v", name, err)
				}
			}
			if _, active := hosts["hetzner-1"]; active {
				t.Fatal("remote Hetzner host must use a validated non-PXE provisioning path")
			}
		})
	}
}
