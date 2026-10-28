package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s/v2"
	"github.com/khuedoan/homelab/tests/internal/cluster"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func restoreReader(t testing.TB, name, node, claim string) *corev1.Pod {
	t.Helper()
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: cluster.PodSpec(t, "testdata/restore-reader.yaml", map[string]string{
			"Node": node, "Claim": claim,
		}),
	}
}

func deleteRestoreClaim(t testing.TB, claim *corev1.PersistentVolumeClaim) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	current, err := lab.Client.CoreV1().PersistentVolumeClaims(claim.Namespace).Get(ctx, claim.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return
	}
	require.NoError(t, err, "read restored claim for cleanup")
	uid := claim.UID
	err = lab.Client.CoreV1().PersistentVolumeClaims(claim.Namespace).Delete(ctx, claim.Name, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid}})
	if !apierrors.IsNotFound(err) {
		require.NoError(t, err, "delete restored claim")
	}
	err = wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		_, err := lab.Client.CoreV1().PersistentVolumeClaims(claim.Namespace).Get(ctx, claim.Name, metav1.GetOptions{})
		if !apierrors.IsNotFound(err) {
			return false, err
		}
		if current.Spec.VolumeName == "" {
			return true, nil
		}
		_, err = lab.Client.CoreV1().PersistentVolumes().Get(ctx, current.Spec.VolumeName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	})
	require.NoError(t, err, "wait for restored volume reclamation")
}

func volumeHandle(ctx context.Context, t testing.TB, lab cluster.Cluster, namespace, name string) string {
	t.Helper()
	claim := k8s.GetPersistentVolumeClaimContext(t, ctx, lab.Kubectl(namespace), name)
	volume := k8s.GetPersistentVolumeContext(t, ctx, lab.Kubectl(namespace), claim.Spec.VolumeName)
	require.NotNil(t, volume.Spec.CSI)
	return volume.Spec.CSI.VolumeHandle
}
