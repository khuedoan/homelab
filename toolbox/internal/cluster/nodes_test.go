package cluster

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func testNodeAPI(t *testing.T, clusterID string, missing bool) (host, sshServers) {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/readyz":
			testWriteResponse(t, w, "ready")
		case "/api/v1/namespaces/kube-system":
			testWriteResponse(t, w, `{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"kube-system","uid":"cluster-uid"}}`)
		case "/api/v1/nodes/test":
			if clusterID != "cluster-uid" {
				t.Error("queried node in a foreign cluster")
			}
			if missing {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			testWriteResponse(t, w, `{"apiVersion":"v1","kind":"Node","metadata":{"name":"test","uid":"node-uid"}}`)
		default:
			t.Errorf("unexpected request: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	data, _ := testClientKubeconfig(t, server)
	node, remote, _ := testSSHServer(t, testRemoteFiles{"/etc/rancher/k3s/k3s.yaml": data}, map[string]string{
		"127.0.0.1:6443": server.Listener.Addr().String(),
	})
	return node, remote
}

func TestNodeUIDRequiresIntendedCluster(t *testing.T) {
	for _, tc := range []struct {
		name, clusterID string
		missing         bool
		wantUID         string
	}{
		{"member", "cluster-uid", false, "node-uid"},
		{"foreign cluster", "other-cluster", false, ""},
		{"missing node", "cluster-uid", true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node, remote := testNodeAPI(t, tc.clusterID, tc.missing)
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			uid, err := remote.nodeUID(ctx, node, tc.clusterID)
			if uid != tc.wantUID || (err != nil) != (tc.wantUID == "") {
				t.Fatalf("node UID = %q, %v; want %q", uid, err, tc.wantUID)
			}
			if tc.clusterID != "cluster-uid" && !strings.Contains(err.Error(), "another cluster") {
				t.Fatalf("unexpected cluster rejection: %v", err)
			}
		})
	}
}

func TestNativeAPIIdentityRequiresReadiness(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		uid    string
		valid  bool
	}{
		{"ready cluster", http.StatusOK, "cluster-uid", true},
		{"unready API", http.StatusServiceUnavailable, "cluster-uid", false},
		{"missing identity", http.StatusOK, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/readyz":
					w.WriteHeader(tc.status)
					testWriteResponse(t, w, "ready")
				case "/api/v1/namespaces/kube-system":
					w.Header().Set("Content-Type", "application/json")
					testWriteResponse(t, w, `{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"kube-system","uid":"`+tc.uid+`"}}`)
				default:
					t.Errorf("unexpected API request: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
			api, err := coreclient.NewForConfig(&rest.Config{Host: server.URL, TLSClientConfig: rest.TLSClientConfig{CAData: ca}})
			if err != nil {
				t.Fatal(err)
			}
			uid, err := apiIdentity(context.Background(), api)
			if (err == nil) != tc.valid || tc.valid && uid != "cluster-uid" {
				t.Fatalf("identity %q, error %v", uid, err)
			}
		})
	}
}

func TestMembershipRequiresMatchingUIDRoleAndReadiness(t *testing.T) {
	node := corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "joiner", UID: "node-uid", Labels: map[string]string{"node-role.kubernetes.io/control-plane": "true"}},
		Status:     corev1.NodeStatus{Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}}},
	}
	for _, tc := range []struct {
		name   string
		change func(*corev1.Node)
		valid  bool
	}{
		{"expected member", func(*corev1.Node) {}, true},
		{"foreign node", func(n *corev1.Node) { n.UID = "other-uid" }, false},
		{"missing node", func(n *corev1.Node) { n.Name = "other" }, false},
		{"unready node", func(n *corev1.Node) { n.Status.Conditions[0].Status = corev1.ConditionFalse }, false},
		{"worker", func(n *corev1.Node) { n.Labels = nil }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			observed := node.DeepCopy()
			tc.change(observed)
			err := verifyNodes([]corev1.Node{*observed}, map[string]string{"joiner": "node-uid"})
			if (err == nil) != tc.valid {
				t.Fatalf("accepted invalid membership: %v", err)
			}
		})
	}
	if err := verifyNodes(nil, map[string]string{"joiner": "node-uid"}); err == nil {
		t.Fatal("accepted missing member")
	}
}
