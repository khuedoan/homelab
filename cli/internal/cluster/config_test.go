package cluster

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnrollmentInventory(t *testing.T) {
	for _, tc := range []struct {
		name, hosts, cluster string
		valid                bool
	}{
		{"explicit seed and IPv6", `{"z-seed":{"ip":"192.0.2.9"},"a-joiner":{"ipv6_address":"2001:db8::1"}}`, `{"init_host":"z-seed","vip":"192.0.2.100"}`, true},
		{"no discovered address", `{"seed":{}}`, `{"init_host":"seed","vip":"192.0.2.100"}`, false},
		{"absent seed", `{"seed":{"ip":"192.0.2.9"}}`, `{"init_host":"missing","vip":"192.0.2.100"}`, false},
		{"duplicate IP", `{"seed":{"ip":"192.0.2.9"},"other":{"ip":"192.0.2.9"}}`, `{"init_host":"seed","vip":"192.0.2.100"}`, false},
		{"management VIP", `{"seed":{"ip":"192.0.2.100"}}`, `{"init_host":"seed","vip":"192.0.2.100"}`, false},
		{"bad name", `{"bad/name":{"ip":"192.0.2.9"}}`, `{"init_host":"bad/name","vip":"192.0.2.100"}`, false},
		{"bad VIP", `{"seed":{"ip":"192.0.2.9"}}`, `{"init_host":"seed","vip":"0.0.0.0"}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			for path, data := range map[string]string{
				"infra/staging/metal/hosts.json":    tc.hosts,
				"infra/staging/cluster/config.json": tc.cluster,
			} {
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := loadConfig("staging")
			if (err == nil) != tc.valid {
				t.Fatalf("config %#v, error %v", cfg, err)
			}
			if tc.valid && (cfg.seed.name != "z-seed" || cfg.seed.address.String() != "192.0.2.9:22" || cfg.vip.String() != "192.0.2.100" || len(cfg.joiners) != 1 || cfg.joiners[0].address.String() != "[2001:db8::1]:22") {
				t.Fatalf("wrong initializer or management address: %#v", cfg)
			}
		})
	}
}
