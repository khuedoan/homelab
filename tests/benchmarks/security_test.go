package benchmarks

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func benchmarkSecurity(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	for _, node := range cluster.Target.Names {
		t.Run(node, func(t *testing.T) {
			namespace := fixture.Namespace(t, cluster.Client)
			spec := corev1.PodSpec{
				NodeSelector: map[string]string{"kubernetes.io/hostname": node},
				HostPID:      true,
				Tolerations:  []corev1.Toleration{{Operator: corev1.TolerationOpExists}},
				Containers: []corev1.Container{{
					Name: "kube-bench", Image: "docker.io/aquasec/kube-bench:v0.7.2",
					Command: []string{"kube-bench", "run", "--benchmark", "k3s-cis-1.7", "--targets", "master,node,etcd", "--json"},
				}},
			}
			for i, path := range []string{"/var/lib/rancher/k3s", "/var/lib/kubelet", "/etc/rancher/k3s", "/etc/systemd", "/nix/store"} {
				name := fmt.Sprintf("host-%d", i)
				spec.Volumes = append(spec.Volumes, corev1.Volume{Name: name, VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{Path: path, Type: ptr.To(corev1.HostPathDirectory)},
				}})
				spec.Containers[0].VolumeMounts = append(spec.Containers[0].VolumeMounts, corev1.VolumeMount{Name: name, MountPath: path, ReadOnly: true})
			}
			logs := runJob(t, cluster, namespace, "kube-bench", spec)
			var report struct {
				Controls []struct {
					Target string `json:"node_type"`
					Pass   int    `json:"total_pass"`
					Fail   int    `json:"total_fail"`
					Warn   int    `json:"total_warn"`
				}
			}
			require.NoError(t, json.Unmarshal([]byte(logs), &report), "%s", logs)
			targets := make([]string, 0, len(report.Controls))
			for _, result := range report.Controls {
				targets = append(targets, result.Target)
				t.Logf("%s %s CIS checks: %d pass, %d fail, %d manual warnings", node, result.Target, result.Pass, result.Fail, result.Warn)
				require.Positive(t, result.Pass+result.Fail, "CIS checks must execute rather than only warn")
				require.Zero(t, result.Fail, "CIS failures on %s/%s:\n%s", node, result.Target, logs)
			}
			require.ElementsMatch(t, []string{"master", "node", "etcd"}, targets)
		})
	}
}
