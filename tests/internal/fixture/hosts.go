package fixture

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"strings"
	"testing"

	"github.com/khuedoan/homelab/tests/internal/testenv"
	"github.com/stretchr/testify/require"
)

type networkInterface struct {
	Address   string `json:"address"`
	Addresses []struct {
		Local string `json:"local"`
	} `json:"addr_info"`
}

func interfaces(t *testing.T, target testenv.Target, name string) []networkInterface {
	t.Helper()
	var interfaces []networkInterface
	require.NoError(t, json.Unmarshal([]byte(SSH(t, target, name, "ip -j address")), &interfaces))
	return interfaces
}

func checkHosts(t *testing.T, target testenv.Target) {
	t.Helper()
	for _, name := range target.Names {
		node := target.Hosts[name]
		t.Logf("Verifying node %s at %s", name, node.IP)
		require.Equal(t, name, SSH(t, target, name, "hostname"))
		foundMAC, foundIP := false, false
		for _, nic := range interfaces(t, target, name) {
			mac, err := net.ParseMAC(nic.Address)
			foundMAC = foundMAC || err == nil && mac.String() == node.MAC
			for _, address := range nic.Addresses {
				foundIP = foundIP || address.Local == node.IP
			}
		}
		require.True(t, foundMAC, "%s does not have its configured MAC", name)
		require.True(t, foundIP, "%s does not own its management IP", name)
		if node.Machine == "" {
			continue
		}
		machineID := strings.ToLower(SSH(t, target, name, "cat /etc/machine-id"))
		decoded, err := hex.DecodeString(machineID)
		if err != nil || len(decoded) != 16 {
			t.Fatalf("%s returned an invalid machine identity", name)
		}
		digest := hmac.New(sha256.New, []byte("code.khuedoan.com/nixie/machine-id/v1"))
		digest.Write([]byte(machineID))
		require.Equal(t, node.Machine, hex.EncodeToString(digest.Sum(nil)), "%s machine identity changed", name)
	}
}
