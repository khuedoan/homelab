package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretsSyncArguments(t *testing.T) {
	for _, args := range [][]string{
		{"infra", "secrets", "sync"},
		{"infra", "secrets", "sync", "--environment", "staging"},
		{"infra", "secrets", "sync", "--kubeconfig", "test"},
		{"infra", "secrets", "sync", "extra"},
	} {
		if _, _, err := operationExecute(args...); err == nil {
			t.Fatalf("accepted incomplete arguments: %v", args)
		}
	}
}

func TestSecretsSyncReadsAllOutputsBeforeWriting(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, file := range []string{"infra/test/root.hcl", "infra/test/a/terragrunt.hcl", "infra/test/b/terragrunt.hcl"} {
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	operationFake(t, "terragrunt", `case "$PWD" in
  */a) printf '%s\n' '{"text":{"value":"secret-canary"}}' ;;
  */b) printf 'secret-canary' >&2; exit 1 ;;
esac
`)
	out, stderr, err := operationExecute("infra", "secrets", "sync", "--environment", "test", "--kubeconfig", "absent")
	if err == nil || !strings.Contains(err.Error(), "outputs for b failed") || strings.Contains(out+stderr+err.Error(), "secret-canary") {
		t.Fatalf("unexpected sync failure: stdout=%q stderr=%q error=%v", out, stderr, err)
	}
}
