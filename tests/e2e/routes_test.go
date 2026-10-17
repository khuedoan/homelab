package e2e

import (
	"testing"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func TestRouteDiscovery(t *testing.T) {
	client := fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{
		{Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes"}: "HTTPRouteList",
	}, &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1", "kind": "HTTPRoute",
		"metadata": map[string]any{"namespace": "pairdrop", "name": "pairdrop"},
	}}, &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1", "kind": "HTTPRoute",
		"metadata": map[string]any{"namespace": "zot", "name": "zot"},
	}})
	cluster := fixture.Cluster{Dynamic: client}
	for _, tc := range []struct {
		namespace string
		want      []string
	}{
		{metav1.NamespaceAll, []string{"pairdrop/pairdrop", "zot/zot"}},
		{"zot", []string{"zot/zot"}},
	} {
		t.Run("namespace="+tc.namespace, func(t *testing.T) {
			var names []string
			for _, route := range routes(t, cluster, tc.namespace) {
				names = append(names, route.GetNamespace()+"/"+route.GetName())
			}
			require.ElementsMatch(t, tc.want, names)
		})
	}
}
