package e2e

import (
	"time"

	"github.com/khuedoan/homelab/tests/internal/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	velerov2 "github.com/vmware-tanzu/velero/pkg/apis/velero/v2alpha1"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Backups", Label("backups"), func() {
	It("restores uploaded data into an independent volume", func(ctx SpecContext) {
		t := GinkgoTB()
		client := backupClient()
		restoreNamespace := cluster.Namespace(ctx, t, lab.Client)

		By("writing the original bytes to both persistent source volumes")
		for _, volume := range sourceVolumes {
			reclaimableSource(ctx, volume.Claim)
			writer := fixturePod(ctx, "writer")
			writeVolume(ctx, writer, volume, "backup-original-data")
		}

		By("uploading both source volumes to object storage")
		backup := createBackup(ctx, client)
		AddReportEntry("backup", backup.Namespace+"/"+backup.Name)
		waitForBackup(ctx, client, backup.Name)

		var uploads velerov2.DataUploadList
		err := client.List(
			ctx,
			&uploads,
			crclient.InNamespace("velero"),
			crclient.MatchingLabels{"velero.io/backup-name": backup.Name},
		)
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
			writer := fixturePod(ctx, "writer")
			writeVolume(ctx, writer, volume, "source-after-backup")
		}

		By("restoring into a separate namespace")
		restore := createRestore(ctx, client, backup, restoreNamespace)
		AddReportEntry("restore", restore.Namespace+"/"+restore.Name)
		waitForRestore(ctx, client, restore.Name)

		var downloads velerov2.DataDownloadList
		Eventually(func(g Gomega) {
			err := client.List(
				ctx,
				&downloads,
				crclient.InNamespace("velero"),
				crclient.MatchingLabels{"velero.io/restore-name": restore.Name},
			)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(downloads.Items).To(HaveLen(2), "both volumes must return from object storage")
			for _, download := range downloads.Items {
				g.Expect(download.Status.Phase).To(Equal(velerov2.DataDownloadPhaseCompleted))
			}
		}).
			WithContext(ctx).
			WithTimeout(5 * time.Minute).
			WithPolling(2 * time.Second).
			Should(Succeed())

		var downloadedClaims []string
		for _, download := range downloads.Items {
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
			restoredData, err := lab.RunPod(ctx, restoreNamespace, reader)
			Expect(err).NotTo(HaveOccurred(), "reader logs: %s", restoredData)
			Expect(restoredData).To(Equal("backup-original-data"))
			Expect(execInPod(ctx, writer, "cat", volume.Path+"/value")).To(Equal("source-after-backup"))
		}
	}, SpecTimeout(30*time.Minute))
})
