package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/khuedoan/homelab/tests/internal/testenv"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func checkIngress(t *testing.T, cluster fixture.Cluster, app testenv.App, path string) http.Header {
	t.Helper()
	transport := &http.Transport{}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	var lastError error
	var headers http.Header
	err := wait.PollUntilContextTimeout(t.Context(), 5*time.Second, 3*time.Minute, true, func(ctx context.Context) (bool, error) {
		ingress, err := cluster.Client.NetworkingV1().Ingresses(app.Namespace).Get(ctx, app.Ingress, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if len(ingress.Spec.Rules) == 0 || ingress.Spec.Rules[0].Host == "" {
			return false, fmt.Errorf("ingress %s/%s has no hostname", app.Namespace, app.Ingress)
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+ingress.Spec.Rules[0].Host+path, nil)
		if err != nil {
			return false, err
		}
		response, err := client.Do(request)
		if err != nil {
			lastError = err
			return false, nil
		}
		if err := response.Body.Close(); err != nil {
			return false, fmt.Errorf("close HTTPS response: %w", err)
		}
		if response.StatusCode != http.StatusOK {
			lastError = fmt.Errorf("HTTP %d", response.StatusCode)
			return false, nil
		}
		headers = response.Header.Clone()
		return true, nil
	})
	require.NoError(t, err, "HTTPS %s/%s%s: %v", app.Namespace, app.Ingress, path, lastError)
	return headers
}

func checkApps(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	if len(cluster.Target.Config.Apps) == 0 {
		t.Skip("no application ingresses configured")
	}
	for _, app := range cluster.Target.Config.Apps {
		t.Run(app.Namespace+"/"+app.Ingress, func(t *testing.T) {
			checkIngress(t, cluster, app, "/")
		})
	}
}

func checkRegistry(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	if cluster.Target.Config.Registry == nil {
		t.Skip("no registry ingress configured")
	}
	headers := checkIngress(t, cluster, *cluster.Target.Config.Registry, "/v2/")
	require.Equal(t, "registry/2.0", headers.Get("Docker-Distribution-Api-Version"), "endpoint must implement the Distribution v2 API")
}
