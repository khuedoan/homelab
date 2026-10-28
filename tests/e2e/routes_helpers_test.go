package e2e

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func expectHTTPSOK(ctx context.Context, route gatewayv1.HTTPRoute, path string) []http.Header {
	GinkgoHelper()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	hostnames := route.Spec.Hostnames
	Expect(hostnames).NotTo(BeEmpty(), "HTTPRoute %s/%s has no hostname", route.GetNamespace(), route.GetName())

	transport := &http.Transport{Proxy: nil}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second, CheckRedirect: sameHTTPSOrigin}
	var headers []http.Header
	for _, name := range hostnames {
		hostname := string(name)
		Expect(hostname).NotTo(BeEmpty(), "HTTPRoute %s/%s has an empty hostname", route.GetNamespace(), route.GetName())
		Eventually(func(g Gomega) {
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+hostname+path, nil)
			g.Expect(err).NotTo(HaveOccurred())
			response, err := client.Do(request)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(response.Body.Close()).To(Succeed())
			g.Expect(response.StatusCode).To(Equal(http.StatusOK), "HTTPS %s/%s hostname %s%s", route.GetNamespace(), route.GetName(), hostname, path)
			headers = append(headers, response.Header.Clone())
		}).WithContext(ctx).WithTimeout(3 * time.Minute).WithPolling(5 * time.Second).Should(Succeed())
	}
	return headers
}

func sameHTTPSOrigin(request *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return fmt.Errorf("too many redirects")
	}
	if request.URL.Scheme != "https" || request.URL.Host != via[0].URL.Host {
		return http.ErrUseLastResponse
	}
	return nil
}
