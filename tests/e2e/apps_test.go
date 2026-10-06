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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

func checkRoute(t *testing.T, cluster fixture.Cluster, app testenv.App, path string) []http.Header {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	route, err := cluster.Dynamic.Resource(schema.GroupVersionResource{
		Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes",
	}).Namespace(app.Namespace).Get(ctx, app.Route, metav1.GetOptions{})
	require.NoError(t, err)
	hostnames, _, err := unstructured.NestedStringSlice(route.Object, "spec", "hostnames")
	require.NoError(t, err)
	require.NotEmpty(t, hostnames, "HTTPRoute %s/%s has no hostname", app.Namespace, app.Route)

	transport := &http.Transport{Proxy: nil}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	var headers []http.Header
	for _, hostname := range hostnames {
		require.NotEmpty(t, hostname, "HTTPRoute %s/%s has an empty hostname", app.Namespace, app.Route)
		var lastError error
		err := wait.PollUntilContextCancel(ctx, 5*time.Second, true, func(ctx context.Context) (bool, error) {
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+hostname+path, nil)
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
				lastError = fmt.Errorf("HTTPS %s%s returned HTTP %d", hostname, path, response.StatusCode)
				return false, nil
			}
			headers = append(headers, response.Header.Clone())
			return true, nil
		})
		require.NoError(t, err, "HTTPS %s/%s hostname %s%s: %v", app.Namespace, app.Route, hostname, path, lastError)
	}
	return headers
}

func checkApps(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	if len(cluster.Target.Config.Apps) == 0 {
		t.Skip("no application routes configured")
	}
	for _, app := range cluster.Target.Config.Apps {
		t.Run(app.Namespace+"/"+app.Route, func(t *testing.T) {
			checkRoute(t, cluster, app, "/")
		})
	}
}

func checkRegistry(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	if cluster.Target.Config.Registry == nil {
		t.Skip("no registry route configured")
	}
	for _, headers := range checkRoute(t, cluster, *cluster.Target.Config.Registry, "/v2/") {
		require.Equal(t, "registry/2.0", headers.Get("Docker-Distribution-Api-Version"), "endpoint must implement the Distribution v2 API")
	}
}
