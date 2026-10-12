package kanidm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

func readinessKubeconfig(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	config := clientapi.Config{
		CurrentContext: "test",
		Clusters:       map[string]*clientapi.Cluster{"test": {Server: server.URL}},
		AuthInfos:      map[string]*clientapi.AuthInfo{"test": {}},
		Contexts:       map[string]*clientapi.Context{"test": {Cluster: "test", AuthInfo: "test"}},
	}
	path := filepath.Join(t.TempDir(), "kubeconfig")
	if err := clientcmd.WriteToFile(config, path); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestWaitForReadyWaitsForPodCreationAndReadiness(t *testing.T) {
	responses := []string{
		`{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"NotFound","code":404}`,
		`{"kind":"Pod","apiVersion":"v1","status":{}}`,
		`{"kind":"Pod","apiVersion":"v1","status":{"conditions":[{"type":"Ready","status":"Unknown"}]}}`,
		`{"kind":"Pod","apiVersion":"v1","status":{"conditions":[{"type":"ContainersReady","status":"True"},{"type":"Ready","status":"False"}]}}`,
		`{"kind":"Pod","apiVersion":"v1","metadata":{"deletionTimestamp":"2026-01-01T00:00:00Z"},"status":{"conditions":[{"type":"Ready","status":"True"}]}}`,
		`{"kind":"Pod","apiVersion":"v1","status":{"conditions":[{"type":"Ready","status":"True"}]}}`,
	}
	var requests atomic.Int32
	path := readinessKubeconfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/namespaces/kanidm/pods/kanidm-0" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		index := int(requests.Add(1)) - 1
		w.Header().Set("Content-Type", "application/json")
		if index == 0 {
			w.WriteHeader(http.StatusNotFound)
		}
		if index >= len(responses) {
			t.Error("did not stop polling a ready pod")
			index = len(responses) - 1
		}
		_, _ = w.Write([]byte(responses[index]))
	})
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	if err := waitForReady(ctx, path); err != nil {
		t.Fatal(err)
	}
	if got := int(requests.Load()); got != len(responses) {
		t.Fatalf("returned before the pod was ready: got %d requests, want %d", got, len(responses))
	}
}

func TestWaitForReadyReturnsAPIErrors(t *testing.T) {
	var requests atomic.Int32
	path := readinessKubeconfig(t, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"Forbidden","code":403}`))
	})
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := waitForReady(ctx, path); !apierrors.IsForbidden(err) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("retried a forbidden request %d times", got)
	}
}

func TestWaitForReadyHonorsDeadlineWhileMissing(t *testing.T) {
	path := readinessKubeconfig(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"NotFound","code":404}`))
	})
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	if err := waitForReady(ctx, path); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error while the pod was missing, got %v", err)
	}
}

func TestWaitForReadyCancelsInFlightRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	requestCanceled := make(chan struct{})
	path := readinessKubeconfig(t, func(_ http.ResponseWriter, r *http.Request) {
		cancel()
		<-r.Context().Done()
		close(requestCanceled)
	})
	if err := waitForReady(ctx, path); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("Kubernetes request was not canceled")
	}
}

func TestWaitForReadyRequiresExplicitKubeconfig(t *testing.T) {
	if err := waitForReady(t.Context(), ""); err == nil {
		t.Fatal("accepted an ambient kubeconfig")
	}
}
