package e2e

import (
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayclient "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

var _ = Describe("Apps", Label("apps"), func() {
	It("serves the test application over HTTPS", func(ctx SpecContext) {
		client, err := gatewayclient.NewForConfig(lab.RESTConfig())
		Expect(err).NotTo(HaveOccurred())
		route, err := client.GatewayV1().HTTPRoutes(testNamespace).Get(ctx, "test-echo", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(route.Spec.Hostnames).To(HaveLen(1))

		transport := &http.Transport{Proxy: nil}
		DeferCleanup(transport.CloseIdleConnections)
		https := &http.Client{Transport: transport, Timeout: 10 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}
		Eventually(func(g Gomega) {
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+string(route.Spec.Hostnames[0]), nil)
			g.Expect(err).NotTo(HaveOccurred())
			response, err := https.Do(request)
			g.Expect(err).NotTo(HaveOccurred())
			body, err := io.ReadAll(io.LimitReader(response.Body, 1024))
			closeErr := response.Body.Close()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(closeErr).NotTo(HaveOccurred())
			g.Expect(response.StatusCode).To(Equal(http.StatusOK))
			g.Expect(string(body)).To(Equal("homelab-e2e"))
		}).WithContext(ctx).WithTimeout(3 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
	}, SpecTimeout(5*time.Minute))
})

var _ = Describe("Registry", Label("registry"), func() {
	It("implements the Distribution v2 API over HTTPS", func(ctx SpecContext) {
		client, err := gatewayclient.NewForConfig(lab.RESTConfig())
		Expect(err).NotTo(HaveOccurred())

		routes, err := client.GatewayV1().HTTPRoutes("zot").List(ctx, metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(routes.Items).NotTo(BeEmpty())
		for _, route := range routes.Items {
			By(route.GetNamespace() + "/" + route.GetName())
			for _, headers := range expectHTTPSOK(ctx, route, "/v2/") {
				Expect(headers.Get("Docker-Distribution-Api-Version")).To(Equal("registry/2.0"), "endpoint must implement the Distribution v2 API")
			}
		}
	}, SpecTimeout(5*time.Minute))
})
