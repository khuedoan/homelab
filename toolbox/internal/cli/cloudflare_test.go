package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCloudflareLoginUsesOAuth(t *testing.T) {
	dir := t.TempDir()
	credentials := filepath.Join(dir, "credentials.json")
	if err := os.WriteFile(credentials, []byte(`{"oauth_token":"oauth-canary"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	identity, err := json.Marshal(map[string]any{
		"authSource": "OAuth token from " + credentials,
		"accounts":   []map[string]string{{"id": "account-canary"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\n[ -z \"${CLOUDFLARE_API_TOKEN:-}${CLOUDFLARE_API_KEY:-}${CLOUDFLARE_EMAIL:-}\" ] || exit 1\n[ \"$*\" = '--quiet auth whoami' ] || exit 1\nprintf '%s' '" + string(identity) + "'\n"
	if err := os.WriteFile(filepath.Join(dir, "cf"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("CLOUDFLARE_API_TOKEN", "ambient-canary")
	t.Setenv("CLOUDFLARE_API_KEY", "ambient-canary")
	t.Setenv("CLOUDFLARE_EMAIL", "ambient-canary")
	token, account, err := cloudflareLogin(context.Background(), "")
	if err != nil || token != "oauth-canary" || account != "account-canary" {
		t.Fatal("did not use the cf OAuth credentials and account")
	}
}
