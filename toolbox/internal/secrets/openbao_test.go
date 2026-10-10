package secrets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestReadRejectsAmbientKubeconfig(t *testing.T) {
	if _, err := Read(context.Background(), "", "infra/test/value"); err == nil {
		t.Fatal("accepted an ambient kubeconfig")
	}
}

func TestSyncWritesKVAndSkipsUnchangedRecords(t *testing.T) {
	existing := false
	writes := 0
	mountReady := false
	bootstrap := bootstrapHandler(&mountReady)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if bootstrap(w, r) {
			return
		}
		if r.URL.Path != "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/secret/data/infra/cloudflare/cert_manager_token" || r.Header.Get("X-Vault-Token") != "root-canary" {
			t.Errorf("unexpected request %s", r.URL.Path)
			w.WriteHeader(403)
			return
		}
		if !mountReady {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"errors":["no handler for route"]}`))
			return
		}
		switch r.Method {
		case "GET":
			if !existing {
				w.WriteHeader(404)
				_, _ = w.Write([]byte(`{"errors":["missing"]}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":{"data":{"value":"token-canary"},"metadata":{"version":1}}}`))
			return
		case "PUT":
		default:
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for i := 0; i < 2; i++ {
		if err := Sync(ctx, path, records); err != nil {
			t.Fatal(err)
		}
	}
	if writes != 1 {
		t.Fatalf("expected one KV version across two syncs, got %d writes", writes)
	}
}

func bootstrapHandler(mountReady *bool) func(http.ResponseWriter, *http.Request) bool {
	credentialReads, healthReads, mountReads := 0, 0, 0
	return func(w http.ResponseWriter, r *http.Request) bool {
		switch r.URL.Path {
		case "/api/v1/namespaces/openbao/secrets/openbao-unseal":
			credentialReads++
			switch credentialReads {
			case 1:
				w.WriteHeader(404)
				_, _ = w.Write([]byte(`{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"NotFound","code":404}`))
			case 2:
				_, _ = w.Write([]byte(`{"apiVersion":"v1","kind":"Secret","data":{}}`))
			default:
				_, _ = w.Write([]byte(`{"apiVersion":"v1","kind":"Secret","data":{"vault-root":"cm9vdC1jYW5hcnk="}}`))
			}
		case "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/sys/health":
			healthReads++
			_ = json.NewEncoder(w).Encode(map[string]bool{"initialized": true, "sealed": healthReads == 1})
		case "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/sys/mounts":
			mountReads++
			if mountReads == 1 {
				_, _ = w.Write([]byte(`{"data":{}}`))
			} else {
				*mountReady = true
				_, _ = w.Write([]byte(`{"data":{"secret/":{"type":"kv","options":{"version":"2"}}}}`))
			}
		default:
			return false
		}
		return true
	}
}
