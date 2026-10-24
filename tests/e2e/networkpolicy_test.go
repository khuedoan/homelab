package e2e

import (
	"context"
	"errors"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	utilexec "k8s.io/client-go/util/exec"
)

var _ = Describe("NetworkPolicy", Label("networkpolicy"), func() {
	It("denies egress from a selected pod", func(ctx SpecContext) {
		server := fixturePod(ctx, "writer")
		probes := probePods(ctx)
		template := &probes[0]
		for i := range probes {
			if probes[i].Spec.NodeName != server.Spec.NodeName {
				template = &probes[i]
				break
			}
		}
		if len(lab.Target.Names) > 1 {
			Expect(template.Spec.NodeName).NotTo(Equal(server.Spec.NodeName))
		}
		control := fixturePod(ctx, "control")
		Expect(control.UID).NotTo(Equal(server.UID))
		address := "http://" + net.JoinHostPort(server.Status.PodIP, "8080")
		By("checking the backend from an independent unselected workload")
		Expect(execInPod(ctx, control, "wget", "-T", "5", "-qO-", address)).To(Equal("homelab-e2e"))

		By("denying a new connection from a fresh selected probe")
		var selected *corev1.Pod
		// Reusing a source can retain an HBONE connection established before policy propagation.
		Eventually(func(g Gomega, attempt context.Context) {
			candidate := newPolicyProbe(attempt, template)
			output, err := lab.Exec(attempt, testNamespace, candidate.Name, "wget", "-T", "3", "-qO-", address)
			var exit utilexec.ExitError
			g.Expect(errors.As(err, &exit)).To(
				BeTrue(),
				"wget must fail inside the container, rather than the exec transport: %v; output: %q",
				err,
				output,
			)
			g.Expect(exit.ExitStatus()).To(Equal(1))
			selected = candidate
		}).
			WithContext(ctx).
			WithTimeout(2 * time.Minute).
			WithPolling(time.Second).
			Should(Succeed())

		By("checking that the same backend remains healthy from an unselected workload")
		Expect(execInPod(ctx, control, "wget", "-T", "5", "-qO-", address)).To(Equal("homelab-e2e"))
		allowProbe(ctx, selected)
		Eventually(func(g Gomega) {
			output, err := lab.Exec(ctx, testNamespace, selected.Name, "wget", "-T", "5", "-qO-", address)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(output).To(Equal("homelab-e2e"))
		}).
			WithContext(ctx).
			WithTimeout(time.Minute).
			WithPolling(time.Second).
			Should(Succeed())
	}, SpecTimeout(5*time.Minute))
})
