package fixture

import (
	"testing"

	"github.com/khuedoan/homelab/tests/internal/testenv"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Volume(t *testing.T, client *kubernetes.Clientset, namespace string, storage testenv.Storage, size string) *corev1.PersistentVolumeClaim {
	t.Helper()
	class, err := client.StorageV1().StorageClasses().Get(t.Context(), storage.Class, metav1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, class.ReclaimPolicy)
	require.Equal(t, corev1.PersistentVolumeReclaimDelete, *class.ReclaimPolicy, "test volumes must be reclaimable without manual deletion")
	claim, err := client.CoreV1().PersistentVolumeClaims(namespace).Create(t.Context(), &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "data"},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storage.Class,
			AccessModes:      []corev1.PersistentVolumeAccessMode{storage.Mode},
			Resources: corev1.VolumeResourceRequirements{Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(size),
			}},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	return claim
}
