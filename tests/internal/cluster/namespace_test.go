package cluster

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestNamespaceCleanupSurvivesCancellationAndWaitsForReclamation(t *testing.T) {
	var deletes, volumeChecks atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/namespaces":
			assert.NoError(t, json.NewEncoder(w).Encode(corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-one", UID: "owned-namespace"}}))
		case r.URL.Path == "/api/v1/namespaces/test-one/serviceaccounts/default":
			assert.NoError(t, json.NewEncoder(w).Encode(corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "default"}}))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/namespaces/test-one":
			var options metav1.DeleteOptions
			if err := json.NewDecoder(r.Body).Decode(&options); err != nil {
				t.Error(err)
				w.WriteHeader(400)
				return
			}
			if options.Preconditions == nil || options.Preconditions.UID == nil || *options.Preconditions.UID != "owned-namespace" {
				t.Error("cleanup did not select its observed UID")
			}
			deletes.Add(1)
			assert.NoError(t, json.NewEncoder(w).Encode(metav1.Status{Status: "Success"}))
		case r.URL.Path == "/api/v1/namespaces/test-one":
			w.WriteHeader(http.StatusNotFound)
			assert.NoError(t, json.NewEncoder(w).Encode(metav1.Status{Reason: metav1.StatusReasonNotFound, Code: 404}))
		case r.URL.Path == "/api/v1/persistentvolumes":
			list := corev1.PersistentVolumeList{}
			if volumeChecks.Add(1) == 1 {
				list.Items = []corev1.PersistentVolume{{Spec: corev1.PersistentVolumeSpec{ClaimRef: &corev1.ObjectReference{Namespace: "test-one"}}}}
			}
			assert.NoError(t, json.NewEncoder(w).Encode(list))
		default:
			t.Errorf("unexpected Kubernetes request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL, ContentConfig: rest.ContentConfig{ContentType: "application/json"}})
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	t.Run("workload", func(t *testing.T) {
		namespace := Namespace(ctx, t, client)
		require.Equal(t, "test-one", namespace)
		cancel()
	})
	require.Equal(t, int32(1), deletes.Load(), "namespace cleanup must run once")
	require.Equal(t, int32(2), volumeChecks.Load(), "cleanup must observe PV reclamation")
}
