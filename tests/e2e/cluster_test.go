package e2e

import (
	"testing"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/khuedoan/homelab/tests/internal/testenv"
)

func TestCluster(t *testing.T) {
	cluster := fixture.Connect(t, testenv.ForTest(t))
	if len(cluster.Target.Names) == 1 {
		t.Log("Single-node inventory cannot exercise cross-node traffic or control-plane failover.")
	}
	if !t.Run("API", func(t *testing.T) { checkAPI(t, cluster) }) {
		return
	}
	t.Run("Networking", func(t *testing.T) { checkNetworking(t, cluster) })
	t.Run("LoadBalancer", func(t *testing.T) { checkLoadBalancer(t, cluster) })
	t.Run("Storage", func(t *testing.T) { checkStorage(t, cluster) })
	t.Run("Registry", func(t *testing.T) { checkRegistry(t, cluster) })
	t.Run("Apps", func(t *testing.T) { checkApps(t, cluster) })
}
