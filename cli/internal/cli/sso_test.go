package cli

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestSSORequiredArguments(t *testing.T) {
	for _, args := range [][]string{
		{"sso", "group", "ensure"},
		{"sso", "group", "ensure", "users"},
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

func TestSSOGroupValidationBeforeClusterAccess(t *testing.T) {
	_, _, err := operationExecute("sso", "--kubeconfig", "absent", "--url", "https://auth.example.test", "group", "ensure", "../invalid")
	if err == nil || !strings.Contains(err.Error(), "access group") {
		t.Fatalf("did not validate group before cluster access: %v", err)
	}
}

func TestRepeatedSSORedirectFlags(t *testing.T) {
	_, _, err := operationExecute("sso", "--kubeconfig", "absent", "--url", "https://auth.example.test", "client", "ensure", "immich", "--redirect-uri", "http://photos.example.test/auth/login", "--redirect-uri", "https://photos.example.test/user-settings")
	if err == nil || !strings.Contains(err.Error(), "redirect URI") {
		t.Fatalf("invalid first callback was lost when parsing repeated flags: %v", err)
	}
}

func TestSSOClaimValidationBeforeClusterAccess(t *testing.T) {
	for _, mapping := range []string{"immich_role", "immich_role:users", "immich_role:../admins=admin", "immich_role:admins=", "invalid/name:admins=admin"} {
		_, _, err := operationExecute("sso", "--kubeconfig", "absent", "--url", "https://auth.example.test", "client", "ensure", "immich", "--redirect-uri", "https://photos.example.test/auth/login", "--claim", "immich_role:users=user", "--claim", mapping)
		if err == nil || !strings.Contains(err.Error(), "claim mapping") {
			t.Fatalf("invalid claim %q reached cluster access: %v", mapping, err)
		}
	}
}

func TestSSODeadlineAllowsReadinessAndOperation(t *testing.T) {
	dir := t.TempDir()
	config := clientapi.Config{CurrentContext: "test", Clusters: map[string]*clientapi.Cluster{"test": {Server: "http://kubernetes.example.test"}}, AuthInfos: map[string]*clientapi.AuthInfo{"test": {}}, Contexts: map[string]*clientapi.Context{"test": {Cluster: "test", AuthInfo: "test"}}}
	kubeconfig := filepath.Join(dir, "kubeconfig")
	if err := clientcmd.WriteToFile(config, kubeconfig); err != nil {
		t.Fatal(err)
	}
	operationFake(t, "kubectl", "exit 0\n")
	original := http.DefaultTransport
	http.DefaultTransport = ssoDeadlineTransport{t: t}
	t.Cleanup(func() { http.DefaultTransport = original })
	for _, args := range [][]string{
		{"client", "ensure", "grafana", "--redirect-uri", "https://grafana.example.test/login/generic_oauth"},
		{"group", "ensure", "grafana_editors"},
		{"secrets", "sync"},
	} {
		_, _, err := operationExecute(append([]string{"sso", "--kubeconfig", kubeconfig, "--url", "https://auth.example.test"}, args...)...)
		if err == nil || !strings.Contains(err.Error(), "bootstrap credentials") {
			t.Fatalf("expected the readiness probe to stop at unavailable OpenBao credentials: %v", err)
		}
	}
}

type ssoDeadlineTransport struct{ t *testing.T }

func (transport ssoDeadlineTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	deadline, ok := request.Context().Deadline()
	if !ok || time.Until(deadline) <= 15*time.Minute {
		transport.t.Errorf("SSO operation cannot outlast its 15-minute Kanidm readiness wait: deadline in %s", time.Until(deadline))
	}
	return &http.Response{StatusCode: http.StatusForbidden, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"Forbidden","code":403}`)), Request: request}, nil
}
