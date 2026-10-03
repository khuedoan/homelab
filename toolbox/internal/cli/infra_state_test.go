package cli

import (
	"testing"
)

func TestStateEnsureArguments(t *testing.T) {
	t.Setenv("CLOUDFLARE_TFSTATE_API_TOKEN", "")
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", "")
	for _, args := range [][]string{
		{"infra", "state", "ensure"},
		{"infra", "state", "ensure", "--bucket", "tfstate-test"},
		{"infra", "state", "ensure", "--bucket", "tfstate-test", "--account-id", "test-account"},
		{"infra", "state", "ensure", "--bucket", "tfstate-test", "extra"},
	} {
		if out, _, err := operationExecute(args...); err == nil || out != "" {
			t.Fatalf("accepted %v: %q, %v", args, out, err)
		}
	}
}
