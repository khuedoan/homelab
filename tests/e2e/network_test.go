package e2e

import (
	"fmt"
	"net"
	"testing"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func echoServer(t *testing.T, cluster fixture.Cluster, namespace string) *corev1.Pod {
	t.Helper()
	_, err := cluster.Client.CoreV1().Pods(namespace).Create(t.Context(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "echo", Labels: map[string]string{"app": "echo"}},
		Spec: corev1.PodSpec{
			NodeSelector:  map[string]string{"kubernetes.io/hostname": cluster.Target.Names[0]},
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{{
				Name: "echo", Image: "docker.io/library/busybox:1.37.0",
				Command: []string{"sh", "-c", "mkdir -p /www; printf homelab-e2e >/www/index.html; exec httpd -f -p 8080 -h /www"},
				ReadinessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.FromInt32(8080)},
				}},
			}},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	return readyPod(t, cluster, namespace, "echo")
}

func checkNetworking(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	namespace := fixture.Namespace(t, cluster.Client)
	server := echoServer(t, cluster, namespace)
	_, err := cluster.Client.CoreV1().Services(namespace).Create(t.Context(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "echo"},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "echo"},
			Ports:    []corev1.ServicePort{{Port: 8080, TargetPort: intstr.FromInt32(8080)}},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, server.Status.PodIP)
	probe := fmt.Sprintf(`set -eu
nslookup kubernetes.default.svc.cluster.local >/dev/null
test "$(wget -T 10 -qO- http://%s)" = homelab-e2e
test "$(wget -T 10 -qO- http://echo:8080)" = homelab-e2e
printf 'network-ok\n'`, net.JoinHostPort(server.Status.PodIP, "8080"))
	logs := runPod(t, cluster, namespace, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "probe"},
		Spec: corev1.PodSpec{
			NodeSelector: map[string]string{"kubernetes.io/hostname": cluster.Target.Names[len(cluster.Target.Names)-1]},
			Containers:   []corev1.Container{{Name: "probe", Image: "docker.io/library/busybox:1.37.0", Command: []string{"sh", "-c", probe}}},
		},
	})
	require.Equal(t, "network-ok\n", logs)
}
