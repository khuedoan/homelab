package e2e

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

func externalHTTP(ctx context.Context, client *http.Client, address, hostname string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address, nil)
	if err != nil {
		return err
	}
	request.Host = hostname
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 1024))
	err = errors.Join(readErr, response.Body.Close())
	if err != nil {
		return fmt.Errorf("read or close external HTTP response: %w", err)
	}
	if response.StatusCode != http.StatusOK || string(body) != "homelab-e2e" {
		return fmt.Errorf("external HTTP returned %d with body %q", response.StatusCode, body)
	}
	return nil
}

func checkLoadBalancer(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	namespace := fixture.Namespace(t, cluster.Client)
	echoServer(t, cluster, namespace)
	_, err := cluster.Client.CoreV1().Services(namespace).Create(t.Context(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "echo"},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"app": "echo"},
			Ports:    []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt32(8080)}},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	hostname := namespace + ".invalid"
	service, err := cluster.Client.CoreV1().Services("istio-system").Get(t.Context(), "gateway-istio", metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, corev1.ServiceTypeLoadBalancer, service.Spec.Type)
	_, err = cluster.Dynamic.Resource(schema.GroupVersionResource{
		Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes",
	}).Namespace(namespace).Create(t.Context(), &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]any{
			"name": "echo",
			"annotations": map[string]any{
				"external-dns.alpha.kubernetes.io/exclude": "true",
			},
		},
		"spec": map[string]any{
			"parentRefs": []any{map[string]any{
				"group": "gateway.networking.k8s.io", "kind": "Gateway",
				"name": "gateway", "namespace": "istio-system", "sectionName": "http",
			}},
			"hostnames": []any{hostname},
			"rules": []any{map[string]any{
				"backendRefs": []any{map[string]any{"name": "echo", "port": int64(80)}},
			}},
		},
	}}, metav1.CreateOptions{})
	require.NoError(t, err)
	transport := &http.Transport{Proxy: nil}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	var lastError error
	err = wait.PollUntilContextTimeout(t.Context(), 3*time.Second, 3*time.Minute, true, func(ctx context.Context) (bool, error) {
		current, err := cluster.Client.CoreV1().Services(service.Namespace).Get(ctx, service.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if len(current.Status.LoadBalancer.Ingress) == 0 {
			lastError = fmt.Errorf("Service has no external address")
			return false, nil
		}
		for _, ingress := range current.Status.LoadBalancer.Ingress {
			host := ingress.IP
			if host == "" {
				host = ingress.Hostname
			}
			if err := externalHTTP(ctx, client, net.JoinHostPort(host, "80"), hostname); err != nil {
				lastError = err
				return false, nil
			}
		}
		return true, nil
	})
	require.NoError(t, err, "LoadBalancer is not reachable from the test runner: %v", lastError)
}

func TestExternalHTTP(t *testing.T) {
	for _, tc := range []struct {
		name, body, wantError string
		status                int
	}{
		{"echo", "homelab-e2e", "", http.StatusOK},
		{"wrong backend", "another-backend", `external HTTP returned 200 with body "another-backend"`, http.StatusOK},
		{"unavailable", "homelab-e2e", `external HTTP returned 503 with body "homelab-e2e"`, http.StatusServiceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Host != "probe.invalid" || r.Method != http.MethodGet {
					t.Errorf("request = %s %s; want GET probe.invalid", r.Method, r.Host)
				}
				w.WriteHeader(tc.status)
				if _, err := io.WriteString(w, tc.body); err != nil {
					t.Errorf("write response: %v", err)
				}
			}))
			t.Cleanup(server.Close)
			err := externalHTTP(t.Context(), server.Client(), strings.TrimPrefix(server.URL, "http://"), "probe.invalid")
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
				return
			}
			require.NoError(t, err)
		})
	}
}
