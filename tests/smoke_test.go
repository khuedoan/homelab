package test

import (
	"crypto/tls"
	"fmt"
	"os"
	"testing"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
)

func TestSmoke(t *testing.T) {
	t.Parallel()

	var mainApps = []struct {
		name      string
		namespace string
	}{
		{"argocd-server", "argocd"},
		{"gitea", "gitea"},
		{"grafana", "grafana"},
		{"homepage", "homepage"},
		{"kanidm", "kanidm"},
		{"zot", "zot"},
	}

	for _, app := range mainApps {
		app := app // https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(app.name, func(t *testing.T) {
			t.Parallel()

			options := k8s.NewKubectlOptions("", "", app.namespace)

			// Wait the service to become available to ensure that we can access it
			k8s.WaitUntilIngressAvailable(t, options, app.name, 30, 60*time.Second)

			// Now we verify that the service will successfully boot and start serving requests
			ingress := k8s.GetIngress(t, options, app.name)

			// Setup a TLS configuration, ignore the certificate because we may not use cert-manager (like the sandbox environment)
			tlsConfig := tls.Config{
				InsecureSkipVerify: os.Getenv("INSECURE_SKIP_VERIFY") != "",
			}

			// Test the endpoint, this will only fail if we timeout waiting for the service to return a 200 response
			http_helper.HttpGetWithRetryWithCustomValidation(
				t,
				fmt.Sprintf("https://%s", ingress.Spec.Rules[0].Host),
				&tlsConfig,
				30,
				60*time.Second,
				func(statusCode int, body string) bool {
					return statusCode == 200
				},
			)
		})
	}
}
