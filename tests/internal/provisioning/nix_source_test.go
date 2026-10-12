package provisioning

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNixieStagesOnlyPublicFlakeInputs(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq is required by the provisioning wrapper")
	}
	temp := t.TempDir()
	flake := filepath.Join(temp, "terraform")
	if err := os.MkdirAll(filepath.Join(flake, "profiles"), 0700); err != nil {
		t.Fatal(err)
	}
	write := func(name, content string, mode os.FileMode) {
		t.Helper()
		if err := os.WriteFile(name, []byte(content), mode); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"flake.nix", "flake.lock", "default.nix", "configuration.nix", "disks.nix", "profiles/installer.nix", "profiles/k3s-server.nix", "profiles/kube-vip.nix", "profiles/rook-ceph.nix"} {
		write(filepath.Join(flake, name), "{}", 0600)
	}
	write(filepath.Join(flake, "hosts.json"), `{"node":{"mac_address":"00:11:22:33:44:55","ip":"192.0.2.1","password":"secret-canary","machine_id_hash":"private-canary"}}`, 0600)
	write(filepath.Join(flake, "cluster.json"), `{"init_host":"node","vip":"192.0.2.2","token":"secret-canary"}`, 0600)
	write(filepath.Join(flake, "backend.tf"), "secret-canary", 0600)
	write(filepath.Join(flake, "terraform.tfstate"), "secret-canary", 0600)
	if err := os.Mkdir(filepath.Join(flake, ".terraform"), 0700); err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(flake, ".terraform", "credentials"), "secret-canary", 0600)
	bin := filepath.Join(temp, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	// The stub never runs sudo, Nix or provisioning. Inspect the exact cwd and
	// arguments that would be passed to Nixie, before the wrapper removes it.
	write(filepath.Join(bin, "sudo"), `#!/bin/sh
set -eu
pwd > "$CAPTURE"
test "$PWD" != "$ORIGINAL_FLAKE"
test "$*" = "env PATH=$PATH nixie --address 192.0.2.10 --installer path:.#nixosConfigurations.installer --flake path:. --hosts $ORIGINAL_HOSTS --ssh-agent-socket /test-agent"
test "$(find . -type f | wc -l)" -eq 11
test ! -e backend.tf
test ! -e terraform.tfstate
test ! -e .terraform
! grep -R -E 'secret-canary|private-canary' .
jq -e '.node.ip == "192.0.2.1" and .node.mac_address == "00:11:22:33:44:55"' hosts.json >/dev/null
jq -e '.init_host == "node" and .vip == "192.0.2.2"' cluster.json >/dev/null
`, 0700)
	wrapper, err := filepath.Abs("../../../infra/_modules/nixos/nixie")
	if err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(temp, "capture")
	cmd := exec.Command("sh", wrapper, flake, filepath.Join(flake, "hosts.json"))
	cmd.Env = append(os.Environ(), "PATH="+bin+":"+os.Getenv("PATH"), "PXE_INTERFACE=test", "PXE_ADDRESS=192.0.2.10", "SSH_AUTH_SOCK=/test-agent", "CAPTURE="+capture, "ORIGINAL_FLAKE="+flake, "ORIGINAL_HOSTS="+filepath.Join(flake, "hosts.json"))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("wrapper: %v\n%s", err, out)
	}
	staged, err := os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(strings.TrimSpace(string(staged))); !os.IsNotExist(err) {
		t.Fatal("temporary provisioning source was not cleaned up")
	}
}
