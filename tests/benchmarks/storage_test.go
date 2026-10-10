package benchmarks

import (
	"encoding/json"
	"testing"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func benchmarkStorage(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	classes, err := fixture.DiscoverStorage(t.Context(), cluster.Client)
	require.NoError(t, err)
	for _, storage := range classes {
		t.Run(storage.Class, func(t *testing.T) {
			namespace := fixture.Namespace(t, cluster.Client)
			claim := fixture.Volume(t, cluster.Client, namespace, storage, "1Gi")
			spec := corev1.PodSpec{
				NodeSelector: map[string]string{"kubernetes.io/hostname": cluster.Target.Names[0]},
				Containers: []corev1.Container{{
					Name: "storage", Image: "docker.io/zayashv/dbench@sha256:5186d2e2acca33d98b0f50a2fba3463bf0ebe30864c8dbd529f234651c462db4",
					Command: []string{
						"fio", "--name=homelab-storage", "--filename=/data/benchmark", "--size=256M",
						"--rw=randrw", "--rwmixread=70", "--bs=4k", "--ioengine=libaio", "--iodepth=16",
						"--direct=1", "--runtime=30", "--time_based", "--output-format=json",
					},
					VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data"}},
				}},
				Volumes: []corev1.Volume{{Name: "data", VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: claim.Name},
				}}},
			}
			logs := runJob(t, cluster, namespace, "fio", spec)
			var report struct {
				Jobs []struct {
					Error       int `json:"error"`
					Read, Write struct {
						Bytes     int64   `json:"io_bytes"`
						IOPS      float64 `json:"iops"`
						Bandwidth int64   `json:"bw"`
					}
				} `json:"jobs"`
			}
			require.NoError(t, json.Unmarshal([]byte(logs), &report), "%s", logs)
			require.Len(t, report.Jobs, 1)
			result := report.Jobs[0]
			require.Zero(t, result.Error, "fio reported an I/O error")
			require.Positive(t, result.Read.Bytes, "fio must perform reads")
			require.Positive(t, result.Write.Bytes, "fio must perform writes")
			require.Positive(t, result.Read.Bandwidth)
			require.Positive(t, result.Write.Bandwidth)
			t.Logf("%s read %.1f IOPS, %d KiB/s; write %.1f IOPS, %d KiB/s", storage.Class, result.Read.IOPS, result.Read.Bandwidth, result.Write.IOPS, result.Write.Bandwidth)
		})
	}
}
