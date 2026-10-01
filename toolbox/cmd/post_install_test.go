package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func postInstallFake(t *testing.T, name, script string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return dir
}

func postInstallTestCmd() *cobra.Command {
	cmd := newPostInstallCmd()
	cmd.SetContext(context.Background())
	return cmd
}

func TestPostInstallGitea(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "secrets")
			t.Setenv("SECRET_FILE", file)
			postInstallFake(t, "kubectl", `cat >> "$SECRET_FILE"; printf '\n' >> "$SECRET_FILE"`)
			posts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				u, p, ok := r.BasicAuth()
				if !ok || u != "admin" || p != "password" {
					t.Error("missing basic auth")
				}
				if r.Method == "GET" {
					if existing {
						fmt.Fprint(w, `[{"name":"renovate"},{"name":"woodpecker"}]`)
					} else {
						fmt.Fprint(w, `[]`)
					}
					return
				}
				posts++
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				w.WriteHeader(201)
				if strings.HasSuffix(r.URL.Path, "/tokens") {
					want := `["write:repository","read:user","write:issue","read:organization","read:misc"]`
					got, _ := json.Marshal(body["scopes"])
					if string(got) != want || body["name"] != "renovate" {
						t.Errorf("wrong token request: %s", got)
					}
					fmt.Fprint(w, `{"sha1":"token-value"}`)
				} else {
					got, _ := json.Marshal(body["redirect_uris"])
					if string(got) != `["https://woodpecker.test/authorize"]` || body["confidential_client"] != true {
						t.Error("wrong OAuth request")
					}
					fmt.Fprint(w, `{"client_id":"id-value","client_secret":"secret-value"}`)
				}
			}))
			defer server.Close()
			g := postInstallGitea{server.URL, "admin", "password", server.Client()}
			if err := g.integrations(postInstallTestCmd(), "woodpecker.test"); err != nil {
				t.Fatal(err)
			}
			if existing {
				if posts != 0 {
					t.Fatal("created existing integrations")
				}
				return
			}
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			if posts != 2 || len(lines) != 2 {
				t.Fatalf("posts=%d, secrets=%d", posts, len(lines))
			}
			for i, line := range lines {
				var secret struct {
					Kind       string
					Metadata   map[string]string
					StringData map[string]string
				}
				if err := json.Unmarshal([]byte(line), &secret); err != nil {
					t.Fatal(err)
				}
				if secret.Kind != "Secret" || secret.Metadata["namespace"] != "global-secrets" {
					t.Fatal("invalid Secret")
				}
				if i == 0 && (secret.Metadata["name"] != "gitea.renovate" || secret.StringData["token"] != "token-value") {
					t.Fatal("wrong token Secret")
				}
				if i == 1 && (secret.Metadata["name"] != "gitea.woodpecker" || secret.StringData["client_id"] != "id-value" || secret.StringData["client_secret"] != "secret-value") {
					t.Fatal("wrong OAuth Secret")
				}
			}
		})
	}
}

func TestPostInstallHTTPErrorsHideSecrets(t *testing.T) {
	for _, response := range []string{"error", "invalid", "empty"} {
		t.Run(response, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if response == "error" {
					w.WriteHeader(500)
					fmt.Fprint(w, "private-password")
					return
				}
				if response == "invalid" {
					fmt.Fprint(w, "private-password")
					return
				}
				if r.Method == "GET" {
					fmt.Fprint(w, "[]")
					return
				}
				w.WriteHeader(201)
				fmt.Fprint(w, "{}")
			}))
			defer server.Close()
			g := postInstallGitea{server.URL, "admin", "private-password", server.Client()}
			err := g.integrations(postInstallTestCmd(), "woodpecker.test")
			if err == nil || strings.Contains(err.Error(), "private-password") {
				t.Fatalf("unsafe error: %v", err)
			}
		})
	}
}

func TestPostInstallKanidm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Error("unexpected creation of existing Gitea integration")
		}
		fmt.Fprint(w, `[{"name":"renovate"},{"name":"woodpecker"}]`)
	}))
	defer server.Close()
	t.Setenv("GITEA_HOST", strings.TrimPrefix(server.URL, "http://"))
	file := filepath.Join(t.TempDir(), "secret")
	t.Setenv("SECRET_FILE", file)
	postInstallFake(t, "kubectl", `case "$1" in
get) case "$2 $5" in
"ingress gitea") printf '%s' "$GITEA_HOST";;
"ingress kanidm") printf kanidm.test;;
"ingress dex") printf dex.test;;
"ingress woodpecker-server") printf woodpecker.test;;
"secret gitea-admin-secret") printf '{"data":{"username":"YWRtaW4=","password":"cGFzc3dvcmQ="}}';;
"secret dex.gitea") printf '{"data":{"client_secret":"ZGV4LXNlY3JldA=="}}';;
"pods --selector=app=gitea") printf gitea-0;;
*) exit 9;; esac;;
exec) if [ "$3" = gitea ]; then
  [ "$*" = 'exec --namespace gitea gitea-0 -- gitea admin auth add-oauth --name Dex --provider openidConnect --key gitea --secret dex-secret --auto-discover-url https://dex.test/.well-known/openid-configuration' ] || exit 8
else printf 'recovery output\n{"password":"private-password"}\n'; fi;;
apply) cat > "$SECRET_FILE";;
*) exit 9;; esac`)
	postInstallFake(t, "kanidm", `case "$1 $2 $3" in
"login --url "*) printf 'Password:'; read -r password; [ "$password" = private-password ] || exit 8;;
"group create --url") [ "$7" = editor ] || exit 8;;
"system oauth2 create") [ "$8" = dex ] && [ "${10}" = https://dex.test/callback ] || exit 8;;
"system oauth2 warning-insecure-client-disable-pkce") :;;
"system oauth2 create-scope-map") [ "$*" = 'system oauth2 create-scope-map --url https://kanidm.test --name idm_admin dex editor openid profile email groups' ] || exit 8;;
"system oauth2 show-basic-secret") printf '{"secret":"oauth-secret"}';;
*) exit 9;; esac`)
	cmd := postInstallTestCmd()
	var logs bytes.Buffer
	cmd.SetOut(&logs)
	cmd.SetErr(&logs)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"client_secret":"oauth-secret"`) || !strings.Contains(string(data), `"name":"kanidm.dex"`) {
		t.Fatal("missing OAuth Secret")
	}
	if logs.Len() != 0 {
		t.Fatalf("credential output leaked: %s", &logs)
	}
}

func TestPostInstallSubprocessFailure(t *testing.T) {
	postInstallFake(t, "kubectl", `echo private-password >&2; exit 17`)
	cmd := postInstallTestCmd()
	var logs bytes.Buffer
	cmd.SetErr(&logs)
	_, err := postInstallSecret(cmd, "gitea", "gitea-admin-secret")
	if err == nil || strings.Contains(err.Error(), "private-password") || logs.Len() != 0 {
		t.Fatalf("unsafe failure: %v, %s", err, &logs)
	}
}
