package e2e

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Storage", Label("storage"), func() {
	It("preserves bytes after the writer pod is deleted", func(ctx SpecContext) {
		writer := fixturePod(ctx, "writer")
		for _, volume := range sourceVolumes {
			By("writing " + volume.Class)
			Expect(execInPod(ctx, writer, "sh", "-ec", fmt.Sprintf(
				"printf homelab-persistent-data > %s/value; sync; printf 'write-ok\\n'", volume.Path,
			))).To(Equal("write-ok\n"))
		}
		By("deleting the observed writer pod")
		Expect(lab.Client.CoreV1().Pods(testNamespace).Delete(ctx, writer.Name, metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{UID: &writer.UID},
		})).To(Succeed())
		replacement := fixturePod(ctx, "writer")
		Expect(replacement.UID).NotTo(Equal(writer.UID))
		for _, volume := range sourceVolumes {
			By("reading " + volume.Class + " from the replacement pod")
			Expect(execInPod(ctx, replacement, "cat", volume.Path+"/value")).To(Equal("homelab-persistent-data"))
		}
	}, SpecTimeout(10*time.Minute))

	Context("with RWX storage", func() {
		It("shares bytes across simultaneous mounts", func(ctx SpecContext) {
			writer := fixturePod(ctx, "writer")
			Expect(execInPod(ctx, writer, "sh", "-ec",
				"printf homelab-persistent-data > /rwx/value; sync; printf 'write-ok\\n'",
			)).To(Equal("write-ok\n"))
			probes := probePods(ctx)
			for _, probe := range probes {
				By("reading the shared mount on " + probe.Spec.NodeName)
				Expect(execInPod(ctx, &probe, "cat", "/rwx/value")).To(Equal("homelab-persistent-data"))
			}
			Expect(fixturePod(ctx, "writer").UID).To(Equal(writer.UID), "the writer must remain mounted throughout the simultaneous-mount check")
		}, SpecTimeout(10*time.Minute))
	})
})
