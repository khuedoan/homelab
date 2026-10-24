package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/khuedoan/homelab/tests/internal/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	velerov2 "github.com/vmware-tanzu/velero/pkg/apis/velero/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Backups", Label("backups"), func() {
	It("restores uploaded data into an independent volume", func(ctx SpecContext) {
		t := GinkgoTB()
		scheme := runtime.NewScheme()
		Expect(velerov1.AddToScheme(scheme)).To(Succeed())
		Expect(velerov2.AddToScheme(scheme)).To(Succeed())
		client, err := crclient.New(lab.RESTConfig(), crclient.Options{Scheme: scheme, Log: GinkgoLogr})
		Expect(err).NotTo(HaveOccurred())
		restoreNamespace := cluster.Namespace(ctx, t, lab.Client)

		By("writing the original bytes to both persistent source volumes")
		for _, volume := range sourceVolumes {
			source, err := lab.Client.CoreV1().PersistentVolumeClaims(testNamespace).Get(ctx, volume.Claim, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			class, err := lab.Client.StorageV1().StorageClasses().Get(ctx, ptr.Deref(source.Spec.StorageClassName, ""), metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			Expect(class.ReclaimPolicy).To(HaveValue(Equal(corev1.PersistentVolumeReclaimDelete)))
			Expect(execInPod(ctx, fixturePod(ctx, "writer"), "sh", "-ec", fmt.Sprintf(
				"printf backup-original-data > %s/value; sync; printf 'write-ok\\n'", volume.Path,
			))).To(Equal("write-ok\n"))
		}

		By("uploading both source volumes to object storage")
		backup := &velerov1.Backup{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "homelab-test-", Namespace: "velero"},
			Spec: velerov1.BackupSpec{
				IncludedNamespaces: []string{testNamespace},
				IncludedResources:  []string{"persistentvolumeclaims", "persistentvolumes"},
				LabelSelector:      &metav1.LabelSelector{MatchLabels: map[string]string{"homelab.khuedoan.com/test-source": "true"}},
				SnapshotMoveData:   ptr.To(true),
				TTL:                metav1.Duration{Duration: time.Hour},
			},
		}
		Expect(client.Create(ctx, backup)).To(Succeed())
		defer deleteBackup(t, client, backup)
		AddReportEntry("backup", backup.Namespace+"/"+backup.Name)
		waitForBackup(ctx, client, backup.Name)
		var uploads velerov2.DataUploadList
		err = client.List(ctx, &uploads, crclient.InNamespace("velero"), crclient.MatchingLabels{"velero.io/backup-name": backup.Name})
		Expect(err).NotTo(HaveOccurred())
		Expect(uploads.Items).To(HaveLen(2), "both source volumes must move through object storage")
		var uploadedClaims []string
		for _, upload := range uploads.Items {
			Expect(upload.Status.Phase).To(Equal(velerov2.DataUploadPhaseCompleted))
			Expect(upload.Spec.SourceNamespace).To(Equal(testNamespace))
			uploadedClaims = append(uploadedClaims, upload.Spec.SourcePVC)
			AddReportEntry("upload/"+upload.Spec.SourcePVC, upload.Name)
		}
		Expect(uploadedClaims).To(ConsistOf("test-rwo", "test-rwx"))

		By("overwriting both sources after the backup")
		for _, volume := range sourceVolumes {
			Expect(execInPod(ctx, fixturePod(ctx, "writer"), "sh", "-ec", fmt.Sprintf(
				"printf source-after-backup > %s/value; sync; printf 'overwrite-ok\\n'", volume.Path,
			))).To(Equal("overwrite-ok\n"))
		}

		By("restoring into a separate namespace")
		restore := &velerov1.Restore{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "homelab-test-", Namespace: "velero"},
			Spec:       velerov1.RestoreSpec{BackupName: backup.Name, NamespaceMapping: map[string]string{testNamespace: restoreNamespace}},
		}
		Expect(client.Create(ctx, restore)).To(Succeed())
		defer func() {
			cleanup, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			err := client.Delete(cleanup, restore, crclient.Preconditions{UID: &restore.UID})
			if !apierrors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
			}
		}()
		AddReportEntry("restore", restore.Namespace+"/"+restore.Name)
		waitForRestore(ctx, client, restore.Name)
		var downloads velerov2.DataDownloadList
		err = client.List(ctx, &downloads, crclient.InNamespace("velero"), crclient.MatchingLabels{"velero.io/restore-name": restore.Name})
		Expect(err).NotTo(HaveOccurred())
		Expect(downloads.Items).To(HaveLen(2), "both volumes must return from object storage")
		var downloadedClaims []string
		for _, download := range downloads.Items {
			Expect(download.Status.Phase).To(Equal(velerov2.DataDownloadPhaseCompleted))
			Expect(download.Spec.TargetVolume.Namespace).To(Equal(restoreNamespace))
			downloadedClaims = append(downloadedClaims, download.Spec.TargetVolume.PVC)
			AddReportEntry("download/"+download.Spec.TargetVolume.PVC, download.Name)
		}
		Expect(downloadedClaims).To(ConsistOf("test-rwo", "test-rwx"))

		By("verifying independent CSI volumes and the restored bytes")
		for _, volume := range sourceVolumes {
			By(volume.Class)
			sourceHandle := volumeHandle(ctx, t, lab, testNamespace, volume.Claim)
			restoredHandle := volumeHandle(ctx, t, lab, restoreNamespace, volume.Claim)
			Expect(restoredHandle).NotTo(Equal(sourceHandle))
			writer := fixturePod(ctx, "writer")
			reader := restoreReader(t, "read-"+volume.Claim, writer.Spec.NodeName, volume.Claim)
			Expect(cluster.RunPod(ctx, t, lab, restoreNamespace, reader)).To(Equal("backup-original-data"))
			Expect(execInPod(ctx, writer, "cat", volume.Path+"/value")).To(Equal("source-after-backup"))
		}
	}, SpecTimeout(30*time.Minute))
})
