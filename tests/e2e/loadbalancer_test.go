package e2e

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

func checkLoadBalancer(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	namespace := fixture.Namespace(t, cluster.Client)
	echoServer(t, cluster, namespace)
	serviceType := corev1.ServiceTypeLoadBalancer
	if cluster.Target.Config.LoadBalancer != nil {
		serviceType = corev1.ServiceTypeClusterIP
	}
	service, err := cluster.Client.CoreV1().Services(namespace).Create(t.Context(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "echo"},
		Spec: corev1.ServiceSpec{
			Type:     serviceType,
			Selector: map[string]string{"app": "echo"},
			Ports:    []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt32(8080)}},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	hostname := namespace + ".invalid"
	if lb := cluster.Target.Config.LoadBalancer; lb != nil {
		service, err = cluster.Client.CoreV1().Services(lb.Namespace).Get(t.Context(), lb.Service, metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, corev1.ServiceTypeLoadBalancer, service.Spec.Type)
		_, err = cluster.Client.NetworkingV1().Ingresses(namespace).Create(t.Context(), &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "echo", Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/ssl-redirect":       "false",
				"nginx.ingress.kubernetes.io/force-ssl-redirect": "false",
			}},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &lb.IngressClass,
				Rules: []networkingv1.IngressRule{{Host: hostname, IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{Paths: []networkingv1.HTTPIngressPath{{
						Path: "/", PathType: ptr.To(networkingv1.PathTypePrefix),
						Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{
							Name: "echo", Port: networkingv1.ServiceBackendPort{Number: 80},
						}},
					}}},
				}}},
			},
		}, metav1.CreateOptions{})
		require.NoError(t, err)
	}
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
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+net.JoinHostPort(host, "80"), nil)
			if err != nil {
				return false, err
			}
			request.Host = hostname
			response, err := client.Do(request)
			if err != nil {
				lastError = err
				return false, nil
			}
			body, err := io.ReadAll(io.LimitReader(response.Body, 1024))
			response.Body.Close()
			if err != nil || response.StatusCode != http.StatusOK || string(body) != "homelab-e2e" {
				lastError = fmt.Errorf("external HTTP returned %d with body %q: %v", response.StatusCode, body, err)
				return false, nil
			}
		}
		return true, nil
	})
	require.NoError(t, err, "LoadBalancer is not reachable from the test runner: %v", lastError)
}
