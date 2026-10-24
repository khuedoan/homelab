package e2e

import (
	"time"

	snapshotclient "github.com/kubernetes-csi/external-snapshotter/client/v8/clientset/versioned"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Snapshots", Label("snapshots"), func() {
	It("restores original bytes while preserving newer source data", func(ctx SpecContext) {
		t := GinkgoTB()
		client, err := snapshotclient.NewForConfig(lab.RESTConfig())
		Expect(err).NotTo(HaveOccurred())

		for _, volume := range sourceVolumes {
			By(volume.Class)
			source := reclaimableSource(ctx, volume.Claim)
			writer := fixturePod(ctx, "writer")

			By("writing the original data and taking a CSI snapshot")
			writeVolume(ctx, writer, volume, "snapshot-original-data")
			snapshot := takeSnapshot(ctx, client, source)
			AddReportEntry("snapshot/"+volume.Class, snapshot.Name)

			By("overwriting the source after the snapshot")
			writeVolume(ctx, writer, volume, "source-after-snapshot")

			By("restoring the snapshot and checking both volumes")
			restored := restoreSnapshot(ctx, snapshot, source)
			reader := restoreReader(t, "read-"+restored.Name, writer.Spec.NodeName, restored.Name)
			restoredData, err := lab.RunPod(ctx, testNamespace, reader)
			Expect(err).NotTo(HaveOccurred(), "reader logs: %s", restoredData)
			Expect(restoredData).To(Equal("snapshot-original-data"), "restored PVC must contain the snapshot value")

			sourceData := execInPod(ctx, fixturePod(ctx, "writer"), "cat", volume.Path+"/value")
			Expect(sourceData).To(Equal("source-after-snapshot"), "source PVC must retain the post-snapshot value")
			restoredHandle := volumeHandle(ctx, t, lab, testNamespace, restored.Name)
			sourceHandle := volumeHandle(ctx, t, lab, testNamespace, source.Name)
			Expect(restoredHandle).NotTo(Equal(sourceHandle))
		}
	}, SpecTimeout(15*time.Minute))
})
