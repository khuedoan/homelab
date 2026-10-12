package secrets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

type oidcServer struct {
	t                       *testing.T
	mounted                 bool
	enables, configs, roles int
}

func (f *oidcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Path == "/api/v1/namespaces/openbao/secrets/openbao-unseal" {
		_, _ = w.Write([]byte(`{"apiVersion":"v1","kind":"Secret","data":{"vault-root":"cm9vdC1jYW5hcnk="}}`))
		return
	}
	if r.Header.Get("X-Vault-Token") != "root-canary" {
		f.t.Error("missing OpenBao authentication")
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/")
	switch path {
	case "sys/health":
		_, _ = w.Write([]byte(`{"initialized":true,"sealed":false}`))
	case "sys/mounts":
		_, _ = w.Write([]byte(`{"data":{"secret/":{"type":"kv","options":{"version":"2"}}}}`))
	case "secret/data/sso/openbao":
		if r.Method != http.MethodGet {
			f.t.Error("changed SSO credentials")
		}
		_, _ = w.Write([]byte(`{"data":{"data":{"client_id":"openbao","client_secret":"secret-canary"},"metadata":{"version":1}}}`))
	case "sys/auth":
		mounts := map[string]any{}
		if f.mounted {
			mounts["oidc/"] = map[string]string{"type": "oidc"}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": mounts})
	case "sys/auth/oidc":
		if r.Method != http.MethodPost || f.mounted {
			f.t.Error("recreated or replaced the OIDC mount")
		}
		f.mounted = true
		f.enables++
	case "auth/oidc/config", "auth/oidc/role/sso":
		f.checkConfiguration(r, path)
	default:
		f.t.Errorf("unexpected request %s %s", r.Method, path)
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *oidcServer) checkConfiguration(r *http.Request, path string) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		f.t.Error(err)
		return
	}
	var want map[string]any
	if path == "auth/oidc/config" {
		f.configs++
		want = map[string]any{
			"oidc_discovery_url": "https://auth.example.test/oauth2/openid/openbao",
			"oidc_client_id":     "openbao", "oidc_client_secret": "secret-canary", "default_role": "sso",
			"jwt_supported_algs": []any{"ES256"},
		}
	} else {
		f.roles++
		want = map[string]any{
			"role_type": "oidc", "user_claim": "sub", "bound_audiences": []any{"openbao"},
			"allowed_redirect_uris": []any{"https://openbao.example.test/ui/vault/auth/oidc/oidc/callback"},
			"oidc_scopes":           []any{"profile", "email", "groups"}, "token_policies": []any{"default"}, "token_ttl": "1h",
		}
	}
	if !reflect.DeepEqual(body, want) {
		f.t.Errorf("incorrect configuration at %s", path)
	}
}

func TestConfigureOIDCPreservesMountAndLimitsTokenPrivileges(t *testing.T) {
	f := &oidcServer{t: t}
	server := httptest.NewServer(f)
	defer server.Close()
	config := clientapi.Config{CurrentContext: "test", Clusters: map[string]*clientapi.Cluster{"test": {Server: server.URL}}, AuthInfos: map[string]*clientapi.AuthInfo{"test": {}}, Contexts: map[string]*clientapi.Context{"test": {Cluster: "test", AuthInfo: "test"}}}
	path := filepath.Join(t.TempDir(), "kubeconfig")
	if err := clientcmd.WriteToFile(config, path); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := ConfigureOIDC(t.Context(), path, "https://auth.example.test/oauth2/openid/openbao", "https://openbao.example.test/ui/vault/auth/oidc/oidc/callback"); err != nil {
			t.Fatal(err)
		}
	}
	if f.enables != 1 || f.configs != 2 || f.roles != 2 {
		t.Fatalf("incorrect configuration counts: %d mounts, %d configs, %d roles", f.enables, f.configs, f.roles)
	}
}

func TestConfigureOIDCRejectsInvalidURLsBeforeClusterAccess(t *testing.T) {
	if err := ConfigureOIDC(context.Background(), "absent", "http://auth.example.test", "https://openbao.example.test/callback"); err == nil || !strings.Contains(err.Error(), "HTTPS") {
		t.Fatalf("expected URL validation, got %v", err)
	}
}
