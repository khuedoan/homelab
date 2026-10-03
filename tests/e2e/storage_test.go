package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func storagePod(name, node, claim, script string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			NodeSelector:  map[string]string{"kubernetes.io/hostname": node},
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{{
				Name: "storage", Image: "docker.io/library/busybox:1.37.0",
				Command:      []string{"sh", "-c", script},
				VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data"}},
			}},
			Volumes: []corev1.Volume{{Name: "data", VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: claim},
			}}},
		},
	}
}

func checkStorage(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	if len(cluster.Target.Config.Storage) == 0 {
		t.Skip("no storage classes configured")
	}
	for _, storage := range cluster.Target.Config.Storage {
		t.Run(storage.Class, func(t *testing.T) {
			namespace := fixture.Namespace(t, cluster.Client)
			claim := fixture.Volume(t, cluster.Client, namespace, storage, "1Gi")
			node := cluster.Target.Names[0]
			logs := runPod(t, cluster, namespace, storagePod("writer", node, claim.Name, "set -eu; printf 'homelab-persistent-data' >/data/value; sync; printf 'write-ok\\n'"))
			require.Equal(t, "write-ok\n", logs)
			writer, err := cluster.Client.CoreV1().Pods(namespace).Get(t.Context(), "writer", metav1.GetOptions{})
			require.NoError(t, err)
			uid := writer.UID
			err = cluster.Client.CoreV1().Pods(namespace).Delete(t.Context(), writer.Name, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid}})
			require.NoError(t, err)
			err = wait.PollUntilContextTimeout(t.Context(), 2*time.Second, time.Minute, true, func(ctx context.Context) (bool, error) {
				_, err := cluster.Client.CoreV1().Pods(namespace).Get(ctx, writer.Name, metav1.GetOptions{})
				if apierrors.IsNotFound(err) {
					return true, nil
				}
				return false, err
			})
			require.NoError(t, err)
			other := cluster.Target.Names[len(cluster.Target.Names)-1]
			logs = runPod(t, cluster, namespace, storagePod("reader", other, claim.Name, "set -eu; cat /data/value"))
			require.Equal(t, "homelab-persistent-data", logs, "data must survive deletion of the writer pod")
			if storage.Mode == corev1.ReadWriteMany {
				_, err := cluster.Client.CoreV1().Pods(namespace).Create(t.Context(), storagePod("holder", node, claim.Name, "sleep 600"), metav1.CreateOptions{})
				require.NoError(t, err)
				readyPod(t, cluster, namespace, "holder")
				if other == node {
					t.Log("RWX uses simultaneous pods on one node; cross-node sharing requires a larger inventory")
				}
				logs = runPod(t, cluster, namespace, storagePod("shared-reader", other, claim.Name, "set -eu; cat /data/value"))
				require.Equal(t, "homelab-persistent-data", logs, "simultaneous RWX mounts must expose the same data")
			}
		})
	}
}
