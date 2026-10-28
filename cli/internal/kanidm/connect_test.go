package kanidm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestConnectReadsCredentialAfterReadiness(t *testing.T) {
	dir := t.TempDir()
	readyPath := filepath.Join(dir, "ready")
	createdPath := filepath.Join(dir, "created")
	recoveryPath := filepath.Join(dir, "recovered")
	command := `#!/bin/sh
case "$*" in
  *' wait '*)
    case "$*" in
      *'--for=create --for=condition=Ready pod/kanidm-0 --timeout=15m') touch "$SSO_CREATED_PATH"; touch "$SSO_READY_PATH" ;;
      *) exit 1 ;;
    esac
    ;;
  *' recover-account '*) touch "$SSO_RECOVERY_PATH"; printf '%s\n' '{"status":"ok","output":"rotated-canary"}' ;;
  *) exit 1 ;;
esac
`
	if err := os.WriteFile(filepath.Join(dir, "kubectl"), []byte(command), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("SSO_READY_PATH", readyPath)
	t.Setenv("SSO_CREATED_PATH", createdPath)
	t.Setenv("SSO_RECOVERY_PATH", recoveryPath)
	writes := 0
	storedPassword := ""
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/namespaces/openbao/secrets/openbao-unseal":
			_, _ = w.Write([]byte(`{"apiVersion":"v1","kind":"Secret","data":{"vault-root":"cm9vdC1jYW5hcnk="}}`))
		case "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/sys/health":
			_, _ = w.Write([]byte(`{"initialized":true,"sealed":false}`))
		case "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/sys/mounts":
			_, _ = w.Write([]byte(`{"data":{"secret/":{"type":"kv","options":{"version":"2"}}}}`))
		case "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/secret/data/kanidm/idm-admin":
			if r.Method == "PUT" {
				var record struct {
					Data map[string]string `json:"data"`
				}
				if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
					t.Error(err)
				}
				storedPassword = record.Data["password"]
				writes++
				_, _ = w.Write([]byte(`{"data":{"version":2}}`))
				return
			}
			if _, err := os.Stat(readyPath); err != nil {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errors":["missing"]}`))
				return
			}
			if storedPassword == "" {
				storedPassword = "admin-canary"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"data": map[string]string{"password": storedPassword}, "metadata": map[string]int{"version": writes + 1}}})
		case "/v1/auth":
			var body struct {
				Step struct {
					Cred *struct {
						Password string `json:"password"`
					} `json:"cred"`
				} `json:"step"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			wantPassword := "admin-canary"
			if _, err := os.Stat(recoveryPath); err == nil {
				wantPassword = "rotated-canary"
			}
			if body.Step.Cred != nil && body.Step.Cred.Password != wantPassword {
				_, _ = w.Write([]byte(`{"state":{"denied":"stale recovery"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"state":{"success":"bearer-canary"}}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	original := http.DefaultTransport
	originalClientTransport := http.DefaultClient.Transport
	http.DefaultTransport = server.Client().Transport
	http.DefaultClient.Transport = server.Client().Transport
	t.Cleanup(func() {
		http.DefaultTransport = original
		http.DefaultClient.Transport = originalClientTransport
	})
	config := clientapi.Config{CurrentContext: "test", Clusters: map[string]*clientapi.Cluster{"test": {Server: server.URL}}, AuthInfos: map[string]*clientapi.AuthInfo{"test": {}}, Contexts: map[string]*clientapi.Context{"test": {Cluster: "test", AuthInfo: "test"}}}
	kubeconfig := filepath.Join(dir, "kubeconfig")
	if err := clientcmd.WriteToFile(config, kubeconfig); err != nil {
		t.Fatal(err)
	}
	c, err := connect(t.Context(), Options{Kubeconfig: kubeconfig, URL: server.URL, Bootstrap: true})
	if err != nil {
		t.Fatalf("credential published during readiness was ignored: %v", err)
	}
	if c.token != "bearer-canary" || writes != 0 || storedPassword != "admin-canary" {
		t.Fatalf("unexpected recovered credential or login token: writes=%d token=%q", writes, c.token)
	}
	if _, err := os.Stat(createdPath); err != nil {
		t.Fatalf("did not wait for the missing Pod to be created: %v", err)
	}
	if _, err := os.Stat(recoveryPath); !os.IsNotExist(err) {
		t.Fatalf("recovered an account whose credential already exists: %v", err)
	}
}

func TestConnectRejectsAmbientKubeconfig(t *testing.T) {
	_, err := connect(t.Context(), Options{URL: "https://auth.example.test"})
	if err == nil || err.Error() != "an explicit kubeconfig path is required" {
		t.Fatalf("did not reject an ambient kubeconfig before cluster access: %v", err)
	}
}
