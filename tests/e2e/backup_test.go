package e2e

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	replicationSources      = schema.GroupVersionResource{Group: "volsync.backube", Version: "v1alpha1", Resource: "replicationsources"}
	replicationDestinations = schema.GroupVersionResource{Group: "volsync.backube", Version: "v1alpha1", Resource: "replicationdestinations"}
	cephClusters            = schema.GroupVersionResource{Group: "ceph.rook.io", Version: "v1", Resource: "cephclusters"}
)

func backupHealth(ctx context.Context, cluster fixture.Cluster) error {
	nodes, err := cluster.Client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		ready, memoryOK := false, false
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				ready = condition.Status == corev1.ConditionTrue
			}
			if condition.Type == corev1.NodeMemoryPressure {
				memoryOK = condition.Status == corev1.ConditionFalse
			}
		}
		if !ready || !memoryOK {
			return fmt.Errorf("node %s health changed: Ready=%t MemoryPressureClear=%t", node.Name, ready, memoryOK)
		}
	}
	ceph, err := cluster.Dynamic.Resource(cephClusters).Namespace("rook-ceph").Get(ctx, "rook-ceph", metav1.GetOptions{})
	if err != nil {
		return err
	}
	health, _, err := unstructured.NestedString(ceph.Object, "status", "ceph", "health")
	if err != nil {
		return err
	}
	state, _, err := unstructured.NestedString(ceph.Object, "status", "state")
	if err != nil {
		return err
	}
	if health != "HEALTH_OK" || state != "Created" {
		return fmt.Errorf("Ceph health changed: health=%s state=%s", health, state)
	}
	return nil
}

func waitForManualSync(t *testing.T, cluster fixture.Cluster, resource schema.GroupVersionResource, namespace, name, token string) {
	t.Helper()
	err := wait.PollUntilContextTimeout(t.Context(), 3*time.Second, 10*time.Minute, true, func(ctx context.Context) (bool, error) {
		if err := backupHealth(ctx, cluster); err != nil {
			return false, err
		}
		current, err := cluster.Dynamic.Resource(resource).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		result, _, err := unstructured.NestedString(current.Object, "status", "latestMoverStatus", "result")
		if err != nil {
			return false, err
		}
		if strings.Contains(strings.ToLower(result), "fail") || strings.Contains(strings.ToLower(result), "error") {
			return false, fmt.Errorf("%s %s failed: %s", resource.Resource, name, result)
		}
		last, _, err := unstructured.NestedString(current.Object, "status", "lastManualSync")
		return last == token, err
	})
	require.NoError(t, err, "%s %s did not complete manual sync", resource.Resource, name)
}

func checkBackups(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	classes, err := fixture.DiscoverStorage(t.Context(), cluster.Client)
	require.NoError(t, err)
	var storage *fixture.Storage
	for i := range classes {
		if classes[i].Mode == corev1.ReadWriteOnce {
			storage = &classes[i]
			break
		}
	}
	require.NotNil(t, storage, "Ceph RWO storage is required")
	require.NoError(t, backupHealth(t.Context(), cluster))
	namespace := fixture.Namespace(t, cluster.Client)
	source := fixture.Volume(t, cluster.Client, namespace, *storage, "1Gi")
	for _, name := range []string{"repository", "restore"} {
		_, err := cluster.Client.CoreV1().PersistentVolumeClaims(namespace).Create(t.Context(), &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: name}, Spec: *source.Spec.DeepCopy(),
		}, metav1.CreateOptions{})
		require.NoError(t, err)
	}

	password := make([]byte, 32)
	_, err = rand.Read(password)
	require.NoError(t, err)
	_, err = cluster.Client.CoreV1().Secrets(namespace).Create(t.Context(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "restic-config"},
		StringData: map[string]string{
			"RESTIC_REPOSITORY": "rest:http://rest-server:8000/repo",
			"RESTIC_PASSWORD":   hex.EncodeToString(password),
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	labels := map[string]string{"app": "rest-server"}
	nonRoot, user, group := true, int64(1000), int64(1000)
	server, err := cluster.Client.CoreV1().Pods(namespace).Create(t.Context(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "rest-server", Labels: labels},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{FSGroup: &group}, RestartPolicy: corev1.RestartPolicyAlways,
			Containers: []corev1.Container{{
				Name: "rest-server", Image: "docker.io/restic/rest-server:0.14.0",
				Env:            []corev1.EnvVar{{Name: "DISABLE_AUTHENTICATION", Value: "1"}},
				Ports:          []corev1.ContainerPort{{Name: "http", ContainerPort: 8000}},
				ReadinessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromString("http")}}},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("10m"), corev1.ResourceMemory: resource.MustParse("16Mi")},
					Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("128Mi")},
				},
				SecurityContext: &corev1.SecurityContext{RunAsNonRoot: &nonRoot, RunAsUser: &user, RunAsGroup: &group},
				VolumeMounts:    []corev1.VolumeMount{{Name: "repository", MountPath: "/data"}},
			}},
			Volumes: []corev1.Volume{{Name: "repository", VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "repository"},
			}}},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	_, err = cluster.Client.CoreV1().Services(namespace).Create(t.Context(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "rest-server"},
		Spec:       corev1.ServiceSpec{Selector: labels, Ports: []corev1.ServicePort{{Name: "http", Port: 8000, TargetPort: intstr.FromString("http")}}},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	err = wait.PollUntilContextTimeout(t.Context(), 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		if err := backupHealth(ctx, cluster); err != nil {
			return false, err
		}
		pod, err := cluster.Client.CoreV1().Pods(namespace).Get(ctx, server.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady {
				return condition.Status == corev1.ConditionTrue, nil
			}
		}
		return false, nil
	})
	require.NoError(t, err, "rest-server did not become ready")

	class, err := cluster.Client.StorageV1().StorageClasses().Get(t.Context(), storage.Class, metav1.GetOptions{})
	require.NoError(t, err)
	snapshot := snapshotClass(t, cluster, class.Provisioner)
	node := cluster.Target.Names[0]
	require.Equal(t, "write-ok\n", runStoragePod(t, cluster, namespace, storagePod("write-original", node, source.Name, "printf 'backup-original-data' >/data/value; sync; printf 'write-ok\\n'")))

	backupToken := "backup-" + time.Now().UTC().Format("20060102150405.000000000")
	_, err = cluster.Dynamic.Resource(replicationSources).Namespace(namespace).Create(t.Context(), &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "volsync.backube/v1alpha1", "kind": "ReplicationSource", "metadata": map[string]any{"name": "source"},
		"spec": map[string]any{"sourcePVC": source.Name, "trigger": map[string]any{"manual": backupToken}, "restic": map[string]any{
			"repository": "restic-config", "copyMethod": "Snapshot", "volumeSnapshotClassName": snapshot, "storageClassName": storage.Class,
			"cacheCapacity": "64Mi", "cacheStorageClassName": storage.Class, "cacheAccessModes": []any{"ReadWriteOnce"},
			"moverResources": map[string]any{"requests": map[string]any{"cpu": "10m", "memory": "64Mi"}, "limits": map[string]any{"memory": "256Mi"}},
			"retain":         map[string]any{"last": "1"},
		}},
	}}, metav1.CreateOptions{})
	require.NoError(t, err)
	waitForManualSync(t, cluster, replicationSources, namespace, "source", backupToken)

	require.Equal(t, "overwrite-ok\n", runStoragePod(t, cluster, namespace, storagePod("overwrite-source", node, source.Name, "printf 'source-after-backup' >/data/value; sync; printf 'overwrite-ok\\n'")))
	restoreToken := "restore-" + time.Now().UTC().Format("20060102150405.000000000")
	_, err = cluster.Dynamic.Resource(replicationDestinations).Namespace(namespace).Create(t.Context(), &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "volsync.backube/v1alpha1", "kind": "ReplicationDestination", "metadata": map[string]any{"name": "restore"},
		"spec": map[string]any{"trigger": map[string]any{"manual": restoreToken}, "restic": map[string]any{
			"repository": "restic-config", "destinationPVC": "restore", "copyMethod": "Direct",
			"cacheCapacity": "64Mi", "cacheStorageClassName": storage.Class, "cacheAccessModes": []any{"ReadWriteOnce"},
			"moverResources": map[string]any{"requests": map[string]any{"cpu": "10m", "memory": "64Mi"}, "limits": map[string]any{"memory": "256Mi"}},
		}},
	}}, metav1.CreateOptions{})
	require.NoError(t, err)
	waitForManualSync(t, cluster, replicationDestinations, namespace, "restore", restoreToken)

	require.Equal(t, "backup-original-data", runStoragePod(t, cluster, namespace, storagePod("read-restore", node, "restore", "cat /data/value")))
	require.Equal(t, "source-after-backup", runStoragePod(t, cluster, namespace, storagePod("read-source", node, source.Name, "cat /data/value")))
}
