package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
)

func routes(t *testing.T, cluster fixture.Cluster, namespace string) []unstructured.Unstructured {
	t.Helper()
	list, err := cluster.Dynamic.Resource(schema.GroupVersionResource{
		Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes",
	}).Namespace(namespace).List(t.Context(), metav1.ListOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, list.Items, "no HTTPRoutes found in namespace %q", namespace)
	return list.Items
}

func checkRoute(t *testing.T, route unstructured.Unstructured, path string) []http.Header {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	hostnames, _, err := unstructured.NestedStringSlice(route.Object, "spec", "hostnames")
	require.NoError(t, err)
	require.NotEmpty(t, hostnames, "HTTPRoute %s/%s has no hostname", route.GetNamespace(), route.GetName())

	transport := &http.Transport{Proxy: nil}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	var headers []http.Header
	for _, hostname := range hostnames {
		require.NotEmpty(t, hostname, "HTTPRoute %s/%s has an empty hostname", route.GetNamespace(), route.GetName())
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
		require.NoError(t, err, "HTTPS %s/%s hostname %s%s: %v", route.GetNamespace(), route.GetName(), hostname, path, lastError)
	}
	return headers
}

func checkApps(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	for _, route := range routes(t, cluster, metav1.NamespaceAll) {
		t.Run(route.GetNamespace()+"/"+route.GetName(), func(t *testing.T) {
			checkRoute(t, route, "/")
		})
	}
}

func checkRegistry(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	for _, route := range routes(t, cluster, "zot") {
		t.Run(route.GetNamespace()+"/"+route.GetName(), func(t *testing.T) {
			for _, headers := range checkRoute(t, route, "/v2/") {
				require.Equal(t, "registry/2.0", headers.Get("Docker-Distribution-Api-Version"), "endpoint must implement the Distribution v2 API")
			}
		})
	}
}
