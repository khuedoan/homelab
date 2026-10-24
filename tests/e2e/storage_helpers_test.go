package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s/v2"
	"github.com/khuedoan/homelab/tests/internal/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

func restoreReader(t testing.TB, name, node, claim string) *corev1.Pod {
	t.Helper()

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: cluster.PodSpec(t, "testdata/restore-reader.yaml", map[string]string{
			"Node":  node,
			"Claim": claim,
		}),
	}
}

func deleteRestoreClaim(ctx context.Context, claim *corev1.PersistentVolumeClaim) error {
	current, err := lab.Client.CoreV1().PersistentVolumeClaims(claim.Namespace).Get(
		ctx,
		claim.Name,
		metav1.GetOptions{},
	)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read restored claim for cleanup: %w", err)
	}
	uid := claim.UID
	err = lab.Client.CoreV1().PersistentVolumeClaims(claim.Namespace).Delete(
		ctx,
		claim.Name,
		metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{UID: &uid},
		},
	)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete restored claim: %w", err)
	}
	err = wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		_, err := lab.Client.CoreV1().PersistentVolumeClaims(claim.Namespace).Get(
			ctx,
			claim.Name,
			metav1.GetOptions{},
		)
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
	if err != nil {
		return fmt.Errorf("wait for restored volume reclamation: %w", err)
	}
	return nil
}

func volumeHandle(ctx context.Context, t testing.TB, lab cluster.Cluster, namespace, name string) string {
	t.Helper()

	claim := k8s.GetPersistentVolumeClaimContext(t, ctx, lab.Kubectl(namespace), name)
	volume := k8s.GetPersistentVolumeContext(t, ctx, lab.Kubectl(namespace), claim.Spec.VolumeName)
	require.NotNil(t, volume.Spec.CSI)

	return volume.Spec.CSI.VolumeHandle
}

func reclaimableSource(ctx context.Context, name string) *corev1.PersistentVolumeClaim {
	GinkgoHelper()
	source, err := lab.Client.CoreV1().PersistentVolumeClaims(testNamespace).Get(ctx, name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	class, err := lab.Client.StorageV1().StorageClasses().Get(ctx, ptr.Deref(source.Spec.StorageClassName, ""), metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(class.ReclaimPolicy).To(HaveValue(Equal(corev1.PersistentVolumeReclaimDelete)))
	return source
}

func writeVolume(ctx context.Context, writer *corev1.Pod, volume sourceVolume, value string) {
	GinkgoHelper()
	output := execInPod(ctx, writer, "sh", "-ec",
		`printf '%s' "$1" > "$2/value"; sync; printf 'write-ok\n'`, "--", value, volume.Path)
	Expect(output).To(Equal("write-ok\n"))
}
