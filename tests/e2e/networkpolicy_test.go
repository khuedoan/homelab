package e2e

import (
	"fmt"
	"net"
	"testing"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func checkNetworkPolicy(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	namespace := fixture.Namespace(t, cluster.Client)
	server := echoServer(t, cluster, namespace)
	address := "http://" + net.JoinHostPort(server.Status.PodIP, "8080")
	probe := func(name, access, command string) string {
		return runPod(t, cluster, namespace, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{"access": access}},
			Spec: corev1.PodSpec{
				AutomountServiceAccountToken: ptr.To(false),
				NodeSelector:                 map[string]string{"kubernetes.io/hostname": cluster.Target.Names[len(cluster.Target.Names)-1]},
				Containers: []corev1.Container{{
					Name: "probe", Image: "docker.io/library/busybox:1.37.0",
					Command: []string{"sh", "-ec", command},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1m"), corev1.ResourceMemory: resource.MustParse("8Mi")},
						Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("32Mi")},
					},
				}},
			},
		})
	}
	allowed := fmt.Sprintf(`test "$(wget -T 5 -qO- %s)" = homelab-e2e
printf 'allowed\n'`, address)
	require.Equal(t, "allowed\n", probe("before-policy", "blocked", allowed))
	_, err := cluster.Client.NetworkingV1().NetworkPolicies(namespace).Create(t.Context(), &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "deny-egress"},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"access": "blocked"}},
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	denied := fmt.Sprintf(`for attempt in 1 2 3 4; do
  if wget -T 3 -qO- %s >/dev/null 2>&1; then
    sleep 1
  else
    printf 'blocked\n'
    exit 0
  fi
done
printf 'Egress remained allowed\n' >&2
exit 1`, address)
	require.Equal(t, "blocked\n", probe("after-policy", "blocked", denied))
	require.Equal(t, "allowed\n", probe("unselected-control", "allowed", allowed), "the server must remain reachable from an unselected pod")
	t.Log("Verified egress before policy, denial after policy, and an unaffected control pod")
}
