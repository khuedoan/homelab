package e2e

import (
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Networking", Label("networking"), func() {
	It("resolves cluster DNS from a pod", func(ctx SpecContext) {
		for _, probe := range probePods(ctx) {
			By(probe.Spec.NodeName)
			Expect(execInPod(ctx, &probe, "sh", "-ec", "nslookup kubernetes.default.svc.cluster.local >/dev/null; printf 'dns-ok\\n'")).To(Equal("dns-ok\n"))
		}
	}, SpecTimeout(5*time.Minute))

	It("carries traffic directly to a pod", func(ctx SpecContext) {
		server := fixturePod(ctx, "writer")
		address := "http://" + net.JoinHostPort(server.Status.PodIP, "8080")
		for _, probe := range probePods(ctx) {
			By(probe.Spec.NodeName)
			Expect(execInPod(ctx, &probe, "wget", "-T", "10", "-qO-", address)).To(Equal("homelab-e2e"))
		}
	}, SpecTimeout(5*time.Minute))

	It("carries traffic through a Service", func(ctx SpecContext) {
		for _, probe := range probePods(ctx) {
			By(probe.Spec.NodeName)
			Expect(execInPod(ctx, &probe, "wget", "-T", "10", "-qO-", "http://test-echo:8080")).To(Equal("homelab-e2e"))
		}
	}, SpecTimeout(5*time.Minute))
})
