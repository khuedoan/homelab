package fixture

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDiscoverStorage(t *testing.T) {
	block := &storagev1.StorageClass{
		ObjectMeta:  metav1.ObjectMeta{Name: "renamed-block"},
		Provisioner: "rook-ceph.rbd.csi.ceph.com",
	}
	shared := &storagev1.StorageClass{
		ObjectMeta:  metav1.ObjectMeta{Name: "renamed-shared"},
		Provisioner: "rook-ceph.cephfs.csi.ceph.com",
	}
	other := &storagev1.StorageClass{
		ObjectMeta:  metav1.ObjectMeta{Name: "unrelated"},
		Provisioner: "other.csi.example.com",
	}
	for _, tc := range []struct {
		name    string
		classes []runtime.Object
		want    []Storage
	}{
		{"both provisioners", []runtime.Object{block, shared, other}, []Storage{
			{Class: "renamed-block", Mode: corev1.ReadWriteOnce},
			{Class: "renamed-shared", Mode: corev1.ReadWriteMany},
		}},
		{"missing CephFS", []runtime.Object{block}, nil},
		{"missing RBD", []runtime.Object{shared}, nil},
		{"unrelated provisioner", []runtime.Object{other}, nil},
		{"no classes", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			storage, err := DiscoverStorage(t.Context(), fake.NewSimpleClientset(tc.classes...))
			if tc.want == nil {
				require.ErrorContains(t, err, "Ceph RBD and CephFS storage classes are required")
				require.Nil(t, storage)
				return
			}
			require.NoError(t, err)
			require.ElementsMatch(t, tc.want, storage)
		})
	}
}
