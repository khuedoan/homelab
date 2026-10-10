package fixture

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Storage identifies a discovered Ceph class and its access mode.
type Storage struct {
	Class string
	Mode  corev1.PersistentVolumeAccessMode
}

// DiscoverStorage requires both RBD and CephFS and selects classes by CSI provisioner.
func DiscoverStorage(ctx context.Context, client kubernetes.Interface) ([]Storage, error) {
	classes, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list storage classes: %w", err)
	}
	var storage []Storage
	var rbd, cephfs bool
	for _, class := range classes.Items {
		switch {
		case strings.HasSuffix(class.Provisioner, ".rbd.csi.ceph.com"):
			storage = append(storage, Storage{Class: class.Name, Mode: corev1.ReadWriteOnce})
			rbd = true
		case strings.HasSuffix(class.Provisioner, ".cephfs.csi.ceph.com"):
			storage = append(storage, Storage{Class: class.Name, Mode: corev1.ReadWriteMany})
			cephfs = true
		}
	}
	if !rbd || !cephfs {
		return nil, fmt.Errorf("both Ceph RBD and CephFS storage classes are required; discovered RBD=%t, CephFS=%t", rbd, cephfs)
	}
	return storage, nil
}

// Volume creates a PVC only when its storage class uses the Delete reclaim policy.
// The caller must use Namespace so cleanup also waits for the backing PV's reclamation.
func Volume(t *testing.T, client *kubernetes.Clientset, namespace string, storage Storage, size string) *corev1.PersistentVolumeClaim {
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
