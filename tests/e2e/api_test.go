package e2e

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func checkAPI(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	ctx, client := t.Context(), cluster.Client
	err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 3*time.Minute, true, func(ctx context.Context) (bool, error) {
		body, err := client.CoreV1().RESTClient().Get().AbsPath("/readyz").DoRaw(ctx)
		return err == nil && strings.TrimSpace(string(body)) == "ok", nil
	})
	require.NoError(t, err, "VIP API is not ready")
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	require.NoError(t, err)
	names := make([]string, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		names = append(names, node.Name)
	}
	slices.Sort(names)
	require.Equal(t, cluster.Target.Names, names, "unexpected or missing cluster members")
	for _, node := range nodes.Items {
		_, controlPlane := node.Labels["node-role.kubernetes.io/control-plane"]
		require.True(t, controlPlane, "%s is not a control-plane node", node.Name)
		ready, internalIP := false, false
		for _, condition := range node.Status.Conditions {
			ready = ready || condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue
		}
		for _, address := range node.Status.Addresses {
			internalIP = internalIP || address.Type == corev1.NodeInternalIP && address.Address == cluster.Target.Hosts[node.Name].IP
		}
		require.True(t, ready, "%s is not Ready", node.Name)
		require.True(t, internalIP, "%s has the wrong InternalIP", node.Name)
		require.NotEmpty(t, node.UID)
		localUID := fixture.SSH(t, cluster.Target, node.Name, "k3s kubectl get node "+node.Name+" -o jsonpath='{.metadata.uid}'")
		require.Equal(t, string(node.UID), localUID, "%s belongs to a different cluster", node.Name)
	}
	identity, err := client.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, identity.UID)
	for _, name := range cluster.Target.Names {
		localUID := fixture.SSH(t, cluster.Target, name, "k3s kubectl get namespace kube-system -o jsonpath='{.metadata.uid}'")
		require.Equal(t, string(identity.UID), localUID, "%s cluster identity differs from the VIP", name)
	}
}
