package e2e

import (
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayclient "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

var _ = Describe("LoadBalancer", Label("loadbalancer"), func() {
	It("serves HTTP through every advertised gateway address", func(ctx SpecContext) {
		gateway, err := gatewayclient.NewForConfig(lab.RESTConfig())
		Expect(err).NotTo(HaveOccurred())
		route, err := gateway.GatewayV1().HTTPRoutes(testNamespace).Get(ctx, "test-echo", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(route.Spec.Hostnames).To(HaveLen(1))
		hostname := string(route.Spec.Hostnames[0])
		transport := &http.Transport{Proxy: nil}
		DeferCleanup(transport.CloseIdleConnections)
		client := &http.Client{Transport: transport, Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		Eventually(func(g Gomega) {
			service, err := k8s.GetServiceContextE(GinkgoTB(), ctx, lab.Kubectl("istio-system"), "gateway-istio")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
			g.Expect(service.Status.LoadBalancer.Ingress).NotTo(BeEmpty())
			for _, ingress := range service.Status.LoadBalancer.Ingress {
				host := ingress.IP
				if host == "" {
					host = ingress.Hostname
				}
				request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+net.JoinHostPort(host, "80"), nil)
				g.Expect(err).NotTo(HaveOccurred())
				request.Host = hostname
				response, err := client.Do(request)
				g.Expect(err).NotTo(HaveOccurred())
				body, err := io.ReadAll(io.LimitReader(response.Body, 1024))
				closeErr := response.Body.Close()
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(closeErr).NotTo(HaveOccurred())
				g.Expect(response.StatusCode).To(Equal(http.StatusOK))
				g.Expect(string(body)).To(Equal("homelab-e2e"))
			}
		}).WithContext(ctx).WithTimeout(3 * time.Minute).WithPolling(3 * time.Second).Should(Succeed())
	}, SpecTimeout(5*time.Minute))
})
