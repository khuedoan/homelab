package cli

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKanidmIntegrations(t *testing.T) {
	t.Setenv("KUBECONFIG", "ambient.yaml")
	t.Setenv("PRESERVED_ENV", "unchanged")
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
	operationFake(t, "kubectl", `
[ "$KUBECONFIG" = selected.yaml ] && [ "$PRESERVED_ENV" = unchanged ] || exit 17
case "$1" in
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
apply) cat > "$SECRET_FILE"; printf private-password; printf private-password >&2;;
*) exit 9;; esac`)
	operationFake(t, "kanidm", `
[ "$KUBECONFIG" = selected.yaml ] && [ "$PRESERVED_ENV" = unchanged ] || exit 17
case "$1 $2 $3" in
"login --url "*) printf 'Password:'; read -r password; [ "$password" = private-password ] || exit 8;;
"group create --url") [ "$7" = editor ] || exit 8;;
"system oauth2 create") [ "$8" = dex ] && [ "${10}" = https://dex.test/callback ] || exit 8;;
"system oauth2 warning-insecure-client-disable-pkce") :;;
"system oauth2 create-scope-map") [ "$*" = 'system oauth2 create-scope-map --url https://kanidm.test --name idm_admin dex editor openid profile email groups' ] || exit 8;;
"system oauth2 show-basic-secret") printf '{"secret":"oauth-secret"}';;
*) exit 9;; esac`)
	cmd := newRootCmd()
	var logs bytes.Buffer
	cmd.SetOut(&logs)
	cmd.SetErr(&logs)
	cmd.SetArgs([]string{"integrations", "setup", "--kubeconfig", "selected.yaml"})
	if err := cmd.ExecuteContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"client_secret":"oauth-secret"`) || !strings.Contains(string(data), `"name":"kanidm.dex"`) {
		t.Fatal("missing OAuth Secret")
	}
	if logs.Len() != 0 || os.Getenv("KUBECONFIG") != "ambient.yaml" {
		t.Fatalf("credential output leaked: %s", &logs)
	}
}
