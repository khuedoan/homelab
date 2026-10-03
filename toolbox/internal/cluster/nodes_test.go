package cluster

import (
	"context"
	"crypto/x509"
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
	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

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
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/readyz":
					w.Write([]byte("ready"))
				case "/api/v1/namespaces/kube-system":
					w.Write([]byte(`{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"kube-system","uid":"cluster-uid"}}`))
				case "/api/v1/nodes/test":
					if tc.clusterID != "cluster-uid" {
						t.Error("queried node in a foreign cluster")
					}
					if tc.missing {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					w.Write([]byte(`{"apiVersion":"v1","kind":"Node","metadata":{"name":"test","uid":"node-uid"}}`))
				default:
					t.Errorf("unexpected request: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			certificate := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
			key, err := x509.MarshalPKCS8PrivateKey(server.TLS.Certificates[0].PrivateKey)
			if err != nil {
				t.Fatal(err)
			}
			data, err := clientcmd.Write(clientapi.Config{
				CurrentContext: "default",
				Contexts:       map[string]*clientapi.Context{"default": {Cluster: "default", AuthInfo: "default"}},
				Clusters:       map[string]*clientapi.Cluster{"default": {Server: localAPI, CertificateAuthorityData: certificate}},
				AuthInfos: map[string]*clientapi.AuthInfo{"default": {
					ClientCertificateData: certificate,
					ClientKeyData:         pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}),
				}},
			})
			if err != nil {
				t.Fatal(err)
			}
			node, remote, _ := testSSHServer(t, testRemoteFiles{"/etc/rancher/k3s/k3s.yaml": data}, map[string]string{
				"127.0.0.1:6443": server.Listener.Addr().String(),
			})
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
					w.Write([]byte("ready"))
				case "/api/v1/namespaces/kube-system":
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"kube-system","uid":"` + tc.uid + `"}}`))
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
