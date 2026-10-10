package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateEnsureArguments(t *testing.T) {
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", "")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "cf"), []byte("#!/bin/sh\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
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
