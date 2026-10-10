package cli

import (
	"strings"
	"testing"
)

func TestSSORequiredArguments(t *testing.T) {
	for _, args := range [][]string{
		{"sso", "client", "ensure"},
		{"sso", "client", "ensure", "grafana"},
		{"sso", "client", "ensure", "grafana", "--kubeconfig", "absent", "--url", "https://auth.example.test"},
		{"sso", "secrets", "sync"},
		{"sso", "secrets", "sync", "--kubeconfig", "absent"},
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
