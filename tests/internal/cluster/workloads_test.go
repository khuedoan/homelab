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

func TestRunPodReturnsExactLogsAndDeletesItsUID(t *testing.T) {
	for _, test := range []struct {
		name          string
		phase         corev1.PodPhase
		cleanupStatus int
		wantErrors    []string
	}{
		{name: "succeeded", phase: corev1.PodSucceeded, cleanupStatus: http.StatusOK},
		{name: "failed", phase: corev1.PodFailed, cleanupStatus: http.StatusOK, wantErrors: []string{"workload result"}},
		{name: "succeeded with cleanup failure", phase: corev1.PodSucceeded, cleanupStatus: http.StatusForbidden, wantErrors: []string{"delete transient pod test/reader"}},
		{name: "failed with cleanup failure", phase: corev1.PodFailed, cleanupStatus: http.StatusForbidden, wantErrors: []string{"workload result", "delete transient pod test/reader"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var deleted atomic.Bool
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "reader", UID: "owned-reader"},
				Status:     corev1.PodStatus{Phase: test.phase, Message: "workload result"},
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodDelete:
					assert.Equal(t, "/api/v1/namespaces/test/pods/reader", r.URL.Path)
					if !expectPodDeletion(t, w, r, "owned-reader") {
						return
					}
					deleted.Store(true)
					w.WriteHeader(test.cleanupStatus)
					assert.NoError(t, json.NewEncoder(w).Encode(metav1.Status{Code: int32(test.cleanupStatus)}))
				case r.URL.Path == "/api/v1/namespaces/test/pods/reader/log":
					_, err := w.Write([]byte(" reader output\n\n"))
					assert.NoError(t, err)
				default:
					assert.NoError(t, json.NewEncoder(w).Encode(pod))
				}
			}))
			t.Cleanup(server.Close)
			client, err := kubernetes.NewForConfig(&rest.Config{
				Host:          server.URL,
				ContentConfig: rest.ContentConfig{ContentType: "application/json"},
			})
			require.NoError(t, err)
			cluster := Cluster{Client: client}

			logs, err := cluster.RunPod(t.Context(), "test", &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "reader"},
			})

			require.Equal(t, " reader output\n\n", logs)
			if len(test.wantErrors) == 0 {
				require.NoError(t, err)
			}
			for _, message := range test.wantErrors {
				require.ErrorContains(t, err, message)
			}
			require.True(t, deleted.Load(), "transient readers must be deleted after success or failure")
		})
	}
}

func TestRunPodDeletesItsUIDAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	var deleted atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			assert.NoError(t, json.NewEncoder(w).Encode(corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "reader", UID: "owned-reader"},
			}))
		case http.MethodGet:
			cancel()
			<-r.Context().Done()
		case http.MethodDelete:
			assert.Equal(t, "/api/v1/namespaces/test/pods/reader", r.URL.Path)
			if !expectPodDeletion(t, w, r, "owned-reader") {
				return
			}
			assert.ErrorIs(t, ctx.Err(), context.Canceled)
			deleted.Store(true)
			assert.NoError(t, json.NewEncoder(w).Encode(metav1.Status{Status: "Success"}))
		}
	}))
	t.Cleanup(server.Close)
	client, err := kubernetes.NewForConfig(&rest.Config{
		Host:          server.URL,
		ContentConfig: rest.ContentConfig{ContentType: "application/json"},
	})
	require.NoError(t, err)
	cluster := Cluster{Client: client}

	logs, err := cluster.RunPod(ctx, "test", &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "reader"}})

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, logs)
	require.True(t, deleted.Load(), "cancellation must not prevent deletion of the owned pod")
}

func expectPodDeletion(t *testing.T, w http.ResponseWriter, r *http.Request, uid string) bool {
	t.Helper()
	var options metav1.DeleteOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil {
		t.Error(err)
		http.Error(w, "invalid deletion options", http.StatusBadRequest)
		return false
	}
	if options.Preconditions == nil || options.Preconditions.UID == nil {
		t.Error("cleanup must select the created pod UID")
		http.Error(w, "missing UID precondition", http.StatusBadRequest)
		return false
	}
	return assert.Equal(t, uid, string(*options.Preconditions.UID))
}
