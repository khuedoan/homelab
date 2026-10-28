package e2e

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/cluster"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v8/clientset/versioned"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

var _ = Describe("Snapshots", Label("snapshots"), func() {
	It("restores original bytes while preserving newer source data", func(ctx SpecContext) {
		t := GinkgoTB()
		client, err := snapshotclient.NewForConfig(lab.RESTConfig())
		Expect(err).NotTo(HaveOccurred())
		for _, volume := range sourceVolumes {
			By(volume.Class)
			source, err := lab.Client.CoreV1().PersistentVolumeClaims(testNamespace).Get(ctx, volume.Claim, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			class, err := lab.Client.StorageV1().StorageClasses().Get(ctx, ptr.Deref(source.Spec.StorageClassName, ""), metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(class.ReclaimPolicy).To(HaveValue(Equal(corev1.PersistentVolumeReclaimDelete)))
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
			snapshotClassName := matches[0]
			writer := fixturePod(ctx, "writer")

			By("writing the original data and taking a CSI snapshot")
			Expect(execInPod(ctx, writer, "sh", "-ec", fmt.Sprintf(
				"printf snapshot-original-data > %s/value; sync; printf 'write-ok\\n'", volume.Path,
			))).To(Equal("write-ok\n"))
			snapshots := client.SnapshotV1().VolumeSnapshots(testNamespace)
			snapshot, err := snapshots.Create(ctx, &snapshotv1.VolumeSnapshot{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "homelab-test-"},
				Spec: snapshotv1.VolumeSnapshotSpec{
					VolumeSnapshotClassName: &snapshotClassName,
					Source:                  snapshotv1.VolumeSnapshotSource{PersistentVolumeClaimName: &source.Name},
				},
			}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { deleteSnapshot(t, client, snapshot) })
			Eventually(func(g Gomega) {
				current, err := snapshots.Get(ctx, snapshot.Name, metav1.GetOptions{})
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(current.Status).NotTo(BeNil())
				if current.Status.Error != nil {
					StopTrying("snapshot failed: " + ptr.Deref(current.Status.Error.Message, "unknown error")).Now()
				}
				g.Expect(current.Status.ReadyToUse).To(HaveValue(BeTrue()))
			}).WithContext(ctx).WithTimeout(5 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())
			AddReportEntry("snapshot/"+volume.Class, snapshot.Name)

			By("overwriting the source after the snapshot")
			Expect(execInPod(ctx, writer, "sh", "-ec", fmt.Sprintf(
				"printf source-after-snapshot > %s/value; sync; printf 'change-ok\\n'", volume.Path,
			))).To(Equal("change-ok\n"))

			By("restoring the snapshot and checking both volumes")
			snapshotGroup := snapshotv1.SchemeGroupVersion.Group
			restored, err := lab.Client.CoreV1().PersistentVolumeClaims(testNamespace).Create(ctx, &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "homelab-test-restore-"},
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: source.Spec.StorageClassName,
					AccessModes:      source.Spec.AccessModes,
					Resources:        *source.Spec.Resources.DeepCopy(),
					DataSource: &corev1.TypedLocalObjectReference{
						APIGroup: &snapshotGroup, Kind: "VolumeSnapshot", Name: snapshot.Name,
					},
				},
			}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { deleteRestoreClaim(t, restored) })
			reader := restoreReader(t, "read-"+restored.Name, writer.Spec.NodeName, restored.Name)
			Expect(cluster.RunPod(ctx, t, lab, testNamespace, reader)).To(Equal("snapshot-original-data"), "restored PVC must contain the snapshot value")
			Expect(execInPod(ctx, fixturePod(ctx, "writer"), "cat", volume.Path+"/value")).To(Equal("source-after-snapshot"), "source PVC must retain the post-snapshot value")
			Expect(volumeHandle(ctx, t, lab, testNamespace, restored.Name)).NotTo(Equal(volumeHandle(ctx, t, lab, testNamespace, source.Name)))
		}
	}, SpecTimeout(15*time.Minute))
})

func deleteSnapshot(t testing.TB, client snapshotclient.Interface, snapshot *snapshotv1.VolumeSnapshot) {
	t.Helper()
	snapshots := client.SnapshotV1().VolumeSnapshots(testNamespace)
	uid := snapshot.UID
	cleanup, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	current, err := snapshots.Get(cleanup, snapshot.Name, metav1.GetOptions{})
	if !apierrors.IsNotFound(err) {
		require.NoError(t, err, "read snapshot for cleanup")
	}
	contentName := ""
	if current != nil && current.Status != nil {
		contentName = ptr.Deref(current.Status.BoundVolumeSnapshotContentName, "")
	}
	err = snapshots.Delete(cleanup, snapshot.Name, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid}})
	if !apierrors.IsNotFound(err) {
		require.NoError(t, err, "delete snapshot")
	}
	err = wait.PollUntilContextCancel(cleanup, 2*time.Second, true, func(ctx context.Context) (bool, error) {
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
	require.NoError(t, err, "wait for snapshot reclamation")
}
