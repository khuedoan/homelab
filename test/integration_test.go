package test

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
)

func TestArgoCDCheck(t *testing.T) {
	t.Parallel()

	// Setup the kubectl config and context. Here we choose to use the defaults, which is:
	// - $KUBECONFIG for the kubectl config file
	// - Current context of the kubectl config file
	options := k8s.NewKubectlOptions("", "", "argocd")

	// This will wait up to 10 seconds for the service to become available, to ensure that we can access it
	k8s.WaitUntilIngressAvailable(t, options, "argocd-server", 10, 1*time.Second)

	// Now we verify that the service will successfully boot and start serving requests
	ingress := k8s.GetIngress(t, options, "argocd-server")

	// Setup a TLS configuration to submit with the helper, a blank struct is acceptable
	tlsConfig := tls.Config{}

	// Test the endpoint for up to 5 minutes. This will only fail if we timeout waiting for the service to return a 200 response
	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("https://%s", ingress.Spec.Rules[0].Host),
		&tlsConfig,
		30,
		30*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)
}
