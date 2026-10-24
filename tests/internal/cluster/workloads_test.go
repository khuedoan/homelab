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

type failedPodTB struct{ *testing.T }

func (*failedPodTB) Errorf(string, ...any) { panic("expected workload failure") }

func TestRunPodDeletesItsUIDAfterFailureAndCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var deleted atomic.Bool
	failed := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "reader", UID: "owned-reader"}, Status: corev1.PodStatus{Phase: corev1.PodFailed}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodDelete:
			var options metav1.DeleteOptions
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&options))
			assert.NotNil(t, options.Preconditions)
			assert.Equal(t, "owned-reader", string(*options.Preconditions.UID))
			deleted.Store(true)
			assert.NoError(t, json.NewEncoder(w).Encode(metav1.Status{Status: "Success"}))
		case r.URL.Path == "/api/v1/namespaces/test/pods/reader/log":
			_, err := w.Write([]byte("reader failed"))
			assert.NoError(t, err)
			cancel()
		default:
			assert.NoError(t, json.NewEncoder(w).Encode(failed))
		}
	}))
	t.Cleanup(server.Close)
	client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL, ContentConfig: rest.ContentConfig{ContentType: "application/json"}})
	require.NoError(t, err)
	cluster := Cluster{Client: client}

	require.PanicsWithValue(t, "expected workload failure", func() {
		RunPod(ctx, &failedPodTB{t}, cluster, "test", &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "reader"}})
	})

	require.True(t, deleted.Load(), "failed readers must not remain in the permanent namespace")
}
