package secrets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestSyncWritesKVAndSkipsUnchangedRecords(t *testing.T) {
	if _, err := Read(context.Background(), "", "infra/test/value"); err == nil {
		t.Fatal("accepted an ambient kubeconfig")
	}
	existing := false
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/namespaces/openbao/secrets/openbao-unseal" {
			_, _ = w.Write([]byte(`{"apiVersion":"v1","kind":"Secret","data":{"vault-root":"cm9vdC1jYW5hcnk="}}`))
			return
		}
		if r.URL.Path != "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/secret/data/infra/cloudflare/cert_manager_token" || r.Header.Get("X-Vault-Token") != "root-canary" {
			t.Errorf("unexpected request %s", r.URL.Path)
			w.WriteHeader(403)
			return
		}
		if r.Method == "GET" {
			if !existing {
				w.WriteHeader(404)
				_, _ = w.Write([]byte(`{"errors":["missing"]}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":{"data":{"value":"token-canary"},"metadata":{"version":1}}}`))
			return
		}
		if r.Method != "PUT" {
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(405)
			return
		}
		var body struct {
			Data map[string]string `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Data["value"] != "token-canary" || len(body.Data) != 1 {
			t.Errorf("incorrect KV write body")
		}
		existing = true
		writes++
		_, _ = w.Write([]byte(`{"data":{"version":1}}`))
	}))
	defer server.Close()
	config := clientapi.Config{CurrentContext: "test", Clusters: map[string]*clientapi.Cluster{"test": {Server: server.URL}}, AuthInfos: map[string]*clientapi.AuthInfo{"test": {}}, Contexts: map[string]*clientapi.Context{"test": {Cluster: "test", AuthInfo: "test"}}}
	path := filepath.Join(t.TempDir(), "kubeconfig")
	if err := clientcmd.WriteToFile(config, path); err != nil {
		t.Fatal(err)
	}
	records := map[string]map[string]string{"infra/cloudflare/cert_manager_token": {"value": "token-canary"}}
	for i := 0; i < 2; i++ {
		if err := Sync(context.Background(), path, records); err != nil {
			t.Fatal(err)
		}
	}
	if writes != 1 {
		t.Fatalf("expected one KV version across two syncs, got %d writes", writes)
	}
}
