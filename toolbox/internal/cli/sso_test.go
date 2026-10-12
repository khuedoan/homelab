package cli

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestSSORequiredArguments(t *testing.T) {
	for _, args := range [][]string{
		{"sso", "client", "ensure"},
		{"sso", "client", "ensure", "grafana"},
		{"sso", "client", "ensure", "grafana", "--kubeconfig", "absent", "--url", "https://auth.example.test"},
		{"sso", "secrets", "sync"},
		{"sso", "secrets", "sync", "--kubeconfig", "absent"},
		{"sso", "openbao", "configure"},
		{"sso", "openbao", "configure", "--kubeconfig", "absent", "--url", "https://auth.example.test"},
	} {
		if _, _, err := operationExecute(args...); err == nil {
			t.Fatalf("accepted incomplete arguments: %v", args)
		}
	}
	_, _, err := operationExecute("sso", "--kubeconfig", "absent", "--url", "https://auth.example.test", "client", "ensure", "../invalid", "--redirect-uri", "https://app.example.test/callback")
	if err == nil || !strings.Contains(err.Error(), "client name") {
		t.Fatalf("did not validate the client before cluster access: %v", err)
	}
}

func TestSSOTimeoutDefault(t *testing.T) {
	timeout, err := newSSOCmd().PersistentFlags().GetDuration("timeout")
	if err != nil || timeout != 15*time.Minute {
		t.Fatalf("SSO operations need a 15-minute default budget, got %v (%v)", timeout, err)
	}
}

func ssoTestCommands() [][]string {
	return [][]string{
		{"client", "ensure", "grafana", "--redirect-uri", "https://app.example.test/callback"},
		{"secrets", "sync"},
		{"openbao", "configure", "--redirect-uri", "https://openbao.example.test/callback"},
	}
}

func TestSSORejectsNonpositiveTimeout(t *testing.T) {
	for _, command := range ssoTestCommands() {
		for _, timeout := range []string{"0", "-1s"} {
			args := []string{"sso", "--kubeconfig", "absent", "--url", "https://auth.example.test", "--timeout", timeout}
			_, _, err := operationExecute(append(args, command...)...)
			if err == nil || !strings.Contains(err.Error(), "--timeout must be positive") {
				t.Fatalf("accepted nonpositive timeout %s for %v: %v", timeout, command, err)
			}
		}
	}
}

func TestSSOCommandsHonorConfiguredAndParentDeadlines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/namespaces/openbao/secrets/openbao-unseal" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"NotFound","code":404}`))
	}))
	defer server.Close()
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
	for _, command := range ssoTestCommands() {
		for _, tc := range []struct {
			name, timeout string
			parentTimeout time.Duration
		}{
			{name: "configured", timeout: "100ms", parentTimeout: 5 * time.Second},
			{name: "parent", timeout: "1h", parentTimeout: 100 * time.Millisecond},
		} {
			t.Run(strings.Join(command[:2], "-")+"/"+tc.name, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), tc.parentTimeout)
				defer cancel()
				cmd := newRootCmd()
				args := []string{"sso", "--kubeconfig", path, "--url", "https://auth.example.test", "--timeout", tc.timeout}
				cmd.SetArgs(append(args, command...))
				started := time.Now()
				if err := cmd.ExecuteContext(ctx); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("expected deadline error, got %v", err)
				}
				if elapsed := time.Since(started); elapsed > 2*time.Second {
					t.Fatalf("shorter deadline was not honored: %v", elapsed)
				}
			})
		}
	}
}
