package e2e

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	volumeSnapshots = schema.GroupVersionResource{
		Group: "snapshot.storage.k8s.io", Version: "v1", Resource: "volumesnapshots",
	}
	volumeSnapshotClasses = schema.GroupVersionResource{
		Group: "snapshot.storage.k8s.io", Version: "v1", Resource: "volumesnapshotclasses",
	}
)

func snapshotClass(t *testing.T, cluster fixture.Cluster, driver string) string {
	t.Helper()
	classes, err := cluster.Dynamic.Resource(volumeSnapshotClasses).List(t.Context(), metav1.ListOptions{})
	require.NoError(t, err)
	var matches []string
	for _, class := range classes.Items {
		classDriver, _, err := unstructured.NestedString(class.Object, "driver")
		require.NoError(t, err)
		deletionPolicy, _, err := unstructured.NestedString(class.Object, "deletionPolicy")
		require.NoError(t, err)
		if classDriver == driver && deletionPolicy == "Delete" {
			matches = append(matches, class.GetName())
		}
	}
	sort.Strings(matches)
	require.NotEmpty(t, matches, "no delete-policy VolumeSnapshotClass uses driver %s", driver)
	return matches[0]
}

func runStoragePod(t *testing.T, cluster fixture.Cluster, namespace string, pod *corev1.Pod) string {
	t.Helper()
	pod.Spec.Containers[0].Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1m"),
			corev1.ResourceMemory: resource.MustParse("8Mi"),
		},
		Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("32Mi")},
	}
	logs := runPod(t, cluster, namespace, pod)
	created, err := cluster.Client.CoreV1().Pods(namespace).Get(t.Context(), pod.Name, metav1.GetOptions{})
	require.NoError(t, err)
	uid := created.UID
	err = cluster.Client.CoreV1().Pods(namespace).Delete(t.Context(), created.Name, metav1.DeleteOptions{
		Preconditions: &metav1.Preconditions{UID: &uid},
	})
	require.NoError(t, err)
	err = wait.PollUntilContextTimeout(t.Context(), 2*time.Second, time.Minute, true, func(ctx context.Context) (bool, error) {
		_, err := cluster.Client.CoreV1().Pods(namespace).Get(ctx, created.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
	require.NoError(t, err)
	return logs
}

func checkSnapshots(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	classes, err := fixture.DiscoverStorage(t.Context(), cluster.Client)
	require.NoError(t, err)
	for _, storage := range classes {
		t.Run(storage.Class, func(t *testing.T) {
			namespace := fixture.Namespace(t, cluster.Client)
			source := fixture.Volume(t, cluster.Client, namespace, storage, "1Gi")
			class, err := cluster.Client.StorageV1().StorageClasses().Get(t.Context(), storage.Class, metav1.GetOptions{})
			require.NoError(t, err)
			snapshotClassName := snapshotClass(t, cluster, class.Provisioner)
			node := cluster.Target.Names[0]

			logs := runStoragePod(t, cluster, namespace, storagePod(
				"write-original", node, source.Name,
				"set -eu; printf 'snapshot-original-data' >/data/value; sync; printf 'write-ok\\n'",
			))
			require.Equal(t, "write-ok\n", logs)

			snapshot, err := cluster.Dynamic.Resource(volumeSnapshots).Namespace(namespace).Create(t.Context(),
				&unstructured.Unstructured{Object: map[string]any{
					"apiVersion": "snapshot.storage.k8s.io/v1",
					"kind":       "VolumeSnapshot",
					"metadata":   map[string]any{"name": "data"},
					"spec": map[string]any{
						"volumeSnapshotClassName": snapshotClassName,
						"source":                  map[string]any{"persistentVolumeClaimName": source.Name},
					},
				}}, metav1.CreateOptions{})
			require.NoError(t, err)
			err = wait.PollUntilContextTimeout(t.Context(), 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
				current, err := cluster.Dynamic.Resource(volumeSnapshots).Namespace(namespace).Get(ctx, snapshot.GetName(), metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				if message, found, err := unstructured.NestedString(current.Object, "status", "error", "message"); err != nil {
					return false, err
				} else if found {
					return false, fmt.Errorf("snapshot failed: %s", message)
				}
				ready, _, err := unstructured.NestedBool(current.Object, "status", "readyToUse")
				return ready, err
			})
			require.NoError(t, err, "snapshot %s/%s did not become ready", namespace, snapshot.GetName())

			logs = runStoragePod(t, cluster, namespace, storagePod(
				"change-source", node, source.Name,
				"set -eu; printf 'source-after-snapshot' >/data/value; sync; printf 'change-ok\\n'",
			))
			require.Equal(t, "change-ok\n", logs)

			snapshotGroup := volumeSnapshots.Group
			restore, err := cluster.Client.CoreV1().PersistentVolumeClaims(namespace).Create(t.Context(), &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "restore"},
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: &storage.Class,
					AccessModes:      []corev1.PersistentVolumeAccessMode{storage.Mode},
					Resources: corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("1Gi"),
					}},
					DataSource: &corev1.TypedLocalObjectReference{
						APIGroup: &snapshotGroup, Kind: "VolumeSnapshot", Name: snapshot.GetName(),
					},
				},
			}, metav1.CreateOptions{})
			require.NoError(t, err)

			logs = runStoragePod(t, cluster, namespace, storagePod("read-restore", node, restore.Name, "cat /data/value"))
			require.Equal(t, "snapshot-original-data", logs, "restored PVC must contain the snapshot value")
			logs = runStoragePod(t, cluster, namespace, storagePod("read-source", node, source.Name, "cat /data/value"))
			require.Equal(t, "source-after-snapshot", logs, "source PVC must retain the post-snapshot value")
		})
	}
}
