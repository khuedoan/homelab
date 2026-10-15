package kanidm

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRecoverAdmin(t *testing.T) {
	dir := t.TempDir()
	command := `#!/bin/sh
[ "$*" = '--kubeconfig staging-config exec -n kanidm kanidm-0 -- /sbin/kanidmd scripting recover-account -c /data/server.toml idm_admin' ] || exit 1
printf '%s\n' "$KANIDM_RECOVERY_RESPONSE"
`
	if err := os.WriteFile(filepath.Join(dir, "kubectl"), []byte(command), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, tc := range []struct {
		name, response, password string
	}{
		{"recovered", `{"status":"ok","output":"recovered-canary"}`, "recovered-canary"},
		{"failed", `{"status":"error","output":"failure-canary"}`, ""},
		{"empty", `{"status":"ok","output":""}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("KANIDM_RECOVERY_RESPONSE", tc.response)
			password, err := recoverAdmin(context.Background(), "staging-config")
			if tc.password == "" {
				if err == nil || password != "" {
					t.Fatal("invalid recovery response accepted")
				}
				return
			}
			if err != nil || password != tc.password {
				t.Fatal("successful recovery did not return the password")
			}
		})
	}
}
