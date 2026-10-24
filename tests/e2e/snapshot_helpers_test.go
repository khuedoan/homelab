package e2e

import (
	"context"
	"fmt"
	"sort"
	"time"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v8/clientset/versioned"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

func snapshotClassFor(ctx context.Context, client snapshotclient.Interface, source *corev1.PersistentVolumeClaim) string {
	GinkgoHelper()
	class, err := lab.Client.StorageV1().StorageClasses().Get(ctx, ptr.Deref(source.Spec.StorageClassName, ""), metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	classes, err := client.SnapshotV1().VolumeSnapshotClasses().List(ctx, metav1.ListOptions{})
	Expect(err).NotTo(HaveOccurred())

	var matches []string
	for _, candidate := range classes.Items {
		if candidate.Driver == class.Provisioner && candidate.DeletionPolicy == snapshotv1.VolumeSnapshotContentDelete {
			matches = append(matches, candidate.Name)
		}
	}
	sort.Strings(matches)
	Expect(matches).NotTo(BeEmpty(), "no delete-policy snapshot class uses driver %s", class.Provisioner)
	return matches[0]
}

func takeSnapshot(ctx context.Context, client snapshotclient.Interface, source *corev1.PersistentVolumeClaim) *snapshotv1.VolumeSnapshot {
	GinkgoHelper()
	class := snapshotClassFor(ctx, client, source)
	snapshots := client.SnapshotV1().VolumeSnapshots(testNamespace)
	snapshot, err := snapshots.Create(ctx, &snapshotv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "homelab-test-"},
		Spec: snapshotv1.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &class,
			Source:                  snapshotv1.VolumeSnapshotSource{PersistentVolumeClaimName: &source.Name},
		},
	}, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(deleteSnapshot, client, snapshot, NodeTimeout(3*time.Minute))

	Eventually(func(g Gomega) {
		current, err := snapshots.Get(ctx, snapshot.Name, metav1.GetOptions{})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(current.Status).NotTo(BeNil())
		if current.Status.Error != nil {
			StopTrying("snapshot failed: " + ptr.Deref(current.Status.Error.Message, "unknown error")).Now()
		}
		g.Expect(current.Status.ReadyToUse).To(HaveValue(BeTrue()))
	}).WithContext(ctx).WithTimeout(5 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())
	return snapshot
}

func restoreSnapshot(ctx context.Context, snapshot *snapshotv1.VolumeSnapshot, source *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	GinkgoHelper()
	group := snapshotv1.SchemeGroupVersion.Group
	claim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "homelab-test-restore-"},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: source.Spec.StorageClassName,
			AccessModes:      source.Spec.AccessModes,
			Resources:        *source.Spec.Resources.DeepCopy(),
			DataSource: &corev1.TypedLocalObjectReference{
				APIGroup: &group,
				Kind:     "VolumeSnapshot",
				Name:     snapshot.Name,
			},
		},
	}
	restored, err := lab.Client.CoreV1().PersistentVolumeClaims(testNamespace).Create(ctx, claim, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(deleteRestoreClaim, restored, NodeTimeout(3*time.Minute))
	return restored
}

func deleteSnapshot(ctx context.Context, client snapshotclient.Interface, snapshot *snapshotv1.VolumeSnapshot) error {
	snapshots := client.SnapshotV1().VolumeSnapshots(testNamespace)
	uid := snapshot.UID
	current, err := snapshots.Get(ctx, snapshot.Name, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("read snapshot for cleanup: %w", err)
	}
	contentName := ""
	if current != nil && current.Status != nil {
		contentName = ptr.Deref(current.Status.BoundVolumeSnapshotContentName, "")
	}
	err = snapshots.Delete(
		ctx,
		snapshot.Name,
		metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{UID: &uid},
		},
	)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete snapshot: %w", err)
	}
	err = wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		_, err := snapshots.Get(ctx, snapshot.Name, metav1.GetOptions{})
		if !apierrors.IsNotFound(err) {
			return false, err
		}
		if contentName == "" {
			return true, nil
		}
		_, err = client.SnapshotV1().VolumeSnapshotContents().Get(ctx, contentName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}

		return false, err
	})
	if err != nil {
		return fmt.Errorf("wait for snapshot reclamation: %w", err)
	}
	return nil
}
