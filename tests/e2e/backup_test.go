package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	backups              = schema.GroupVersionResource{Group: "velero.io", Version: "v1", Resource: "backups"}
	restores             = schema.GroupVersionResource{Group: "velero.io", Version: "v1", Resource: "restores"}
	deleteBackupRequests = schema.GroupVersionResource{Group: "velero.io", Version: "v1", Resource: "deletebackuprequests"}
	dataUploads          = schema.GroupVersionResource{Group: "velero.io", Version: "v2alpha1", Resource: "datauploads"}
	dataDownloads        = schema.GroupVersionResource{Group: "velero.io", Version: "v2alpha1", Resource: "datadownloads"}
)

func waitForVelero(t *testing.T, cluster fixture.Cluster, resource schema.GroupVersionResource, name string) {
	t.Helper()
	err := wait.PollUntilContextTimeout(t.Context(), 3*time.Second, 10*time.Minute, true, func(ctx context.Context) (bool, error) {
		current, err := cluster.Dynamic.Resource(resource).Namespace("velero").Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		phase, _, err := unstructured.NestedString(current.Object, "status", "phase")
		if err != nil {
			return false, err
		}
		switch phase {
		case "Failed", "PartiallyFailed", "FailedValidation":
			return false, fmt.Errorf("%s %s ended with phase %s: %v", resource.Resource, name, phase, current.Object["status"])
		case "Completed":
			return true, nil
		default:
			return false, nil
		}
	})
	require.NoError(t, err, "%s %s did not complete", resource.Resource, name)
}

func checkBackups(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	classes, err := fixture.DiscoverStorage(t.Context(), cluster.Client)
	require.NoError(t, err)
	for _, storage := range classes {
		t.Run(storage.Class, func(t *testing.T) {
			checkVolumeBackup(t, cluster, storage)
		})
	}
}

func checkVolumeBackup(t *testing.T, cluster fixture.Cluster, storage fixture.Storage) {
	t.Helper()
	sourceNamespace := fixture.Namespace(t, cluster.Client)
	restoreNamespace := fixture.Namespace(t, cluster.Client)
	source := fixture.Volume(t, cluster.Client, sourceNamespace, storage, "1Gi")
	node := cluster.Target.Names[0]
	require.Equal(t, "write-ok\n", runStoragePod(t, cluster, sourceNamespace, storagePod("write-original", node, source.Name, "printf 'backup-original-data' >/data/value; sync; printf 'write-ok\\n'")))

	backup, err := cluster.Dynamic.Resource(backups).Namespace("velero").Create(t.Context(), &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "velero.io/v1", "kind": "Backup", "metadata": map[string]any{"name": sourceNamespace},
		"spec": map[string]any{
			"includedNamespaces": []any{sourceNamespace},
			"includedResources":  []any{"persistentvolumeclaims", "persistentvolumes"},
			"snapshotMoveData":   true,
			"ttl":                "1h",
		},
	}}, metav1.CreateOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		_, err := cluster.Dynamic.Resource(deleteBackupRequests).Namespace("velero").Create(ctx, &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "velero.io/v1", "kind": "DeleteBackupRequest", "metadata": map[string]any{"name": backup.GetName()},
			"spec": map[string]any{"backupName": backup.GetName()},
		}}, metav1.CreateOptions{})
		if err != nil {
			t.Errorf("request test backup deletion: %v", err)
			return
		}
		err = wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
			_, err := cluster.Dynamic.Resource(backups).Namespace("velero").Get(ctx, backup.GetName(), metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		})
		if err != nil {
			t.Errorf("wait for test backup deletion: %v", err)
		}
	})
	waitForVelero(t, cluster, backups, backup.GetName())
	uploads, err := cluster.Dynamic.Resource(dataUploads).Namespace("velero").List(t.Context(), metav1.ListOptions{
		LabelSelector: "velero.io/backup-name=" + backup.GetName(),
	})
	require.NoError(t, err)
	require.Len(t, uploads.Items, 1, "one volume must be uploaded to object storage")
	phase, _, err := unstructured.NestedString(uploads.Items[0].Object, "status", "phase")
	require.NoError(t, err)
	require.Equal(t, "Completed", phase)
	t.Logf("Backup %s completed with DataUpload %s", backup.GetName(), uploads.Items[0].GetName())

	require.Equal(t, "overwrite-ok\n", runStoragePod(t, cluster, sourceNamespace, storagePod("overwrite-source", node, source.Name, "printf 'source-after-backup' >/data/value; sync; printf 'overwrite-ok\\n'")))
	_, err = cluster.Dynamic.Resource(restores).Namespace("velero").Create(t.Context(), &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "velero.io/v1", "kind": "Restore", "metadata": map[string]any{"name": sourceNamespace},
		"spec": map[string]any{
			"backupName":       backup.GetName(),
			"namespaceMapping": map[string]any{sourceNamespace: restoreNamespace},
		},
	}}, metav1.CreateOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if err := cluster.Dynamic.Resource(restores).Namespace("velero").Delete(ctx, sourceNamespace, metav1.DeleteOptions{}); err != nil {
			t.Errorf("delete test restore: %v", err)
		}
	})
	waitForVelero(t, cluster, restores, sourceNamespace)

	downloads, err := cluster.Dynamic.Resource(dataDownloads).Namespace("velero").List(t.Context(), metav1.ListOptions{
		LabelSelector: "velero.io/restore-name=" + sourceNamespace,
	})
	require.NoError(t, err)
	require.Len(t, downloads.Items, 1, "one volume must be downloaded from object storage")
	phase, _, err = unstructured.NestedString(downloads.Items[0].Object, "status", "phase")
	require.NoError(t, err)
	require.Equal(t, "Completed", phase)

	sourceClaim, err := cluster.Client.CoreV1().PersistentVolumeClaims(sourceNamespace).Get(t.Context(), source.Name, metav1.GetOptions{})
	require.NoError(t, err)
	restoredClaim, err := cluster.Client.CoreV1().PersistentVolumeClaims(restoreNamespace).Get(t.Context(), source.Name, metav1.GetOptions{})
	require.NoError(t, err)
	sourceVolume, err := cluster.Client.CoreV1().PersistentVolumes().Get(t.Context(), sourceClaim.Spec.VolumeName, metav1.GetOptions{})
	require.NoError(t, err)
	restoredVolume, err := cluster.Client.CoreV1().PersistentVolumes().Get(t.Context(), restoredClaim.Spec.VolumeName, metav1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, sourceVolume.Spec.CSI)
	require.NotNil(t, restoredVolume.Spec.CSI)
	require.NotEqual(t, sourceVolume.Spec.CSI.VolumeHandle, restoredVolume.Spec.CSI.VolumeHandle, "restore must provision an independent CSI volume")
	t.Logf("Restore %s completed with DataDownload %s; CSI volume changed from %s to %s", sourceNamespace, downloads.Items[0].GetName(), sourceVolume.Spec.CSI.VolumeHandle, restoredVolume.Spec.CSI.VolumeHandle)
	require.Equal(t, "backup-original-data", runStoragePod(t, cluster, restoreNamespace, storagePod("read-restore", node, source.Name, "cat /data/value")))
	require.Equal(t, "source-after-backup", runStoragePod(t, cluster, sourceNamespace, storagePod("read-source", node, source.Name, "cat /data/value")))
	t.Log("Restored original bytes match; source retains the post-backup overwrite")
}
