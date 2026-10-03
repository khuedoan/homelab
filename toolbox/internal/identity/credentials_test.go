package identity

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/khuedoan/homelab/toolbox/internal/process"
)

func TestPrivateCredentials(t *testing.T) {
	t.Setenv("KUBECONFIG", "ambient.yaml")
	t.Setenv("PRESERVED_ENV", "unchanged")
	file := filepath.Join(t.TempDir(), "secret.json")
	t.Setenv("SECRET_FILE", file)
	integrationFake(t, "kubectl", `
[ "$KUBECONFIG" = selected.yaml ] && [ "$PRESERVED_ENV" = unchanged ] || exit 17
case "$1" in
get) printf '{"data":{"password":"cHJpdmF0ZS1wYXNzd29yZA=="}}';;
apply) cat > "$SECRET_FILE"; printf private-password; printf private-password >&2; [ "$FAIL_APPLY" != yes ] || exit 18;;
*) exit 19;; esac`)
	integrationFake(t, "kanidm", `
[ "$KUBECONFIG" = selected.yaml ] && [ "$PRESERVED_ENV" = unchanged ] || exit 17
printf 'Password:'
read -r password
[ "$password" = private-password ] || exit 18
printf private-password`)
	var logs bytes.Buffer
	run := process.New("selected.yaml", nil, &logs, &logs)
	secret, err := readSecret(t.Context(), run, "gitea", "gitea-admin-secret")
	if err != nil || secret["password"] != "private-password" {
		t.Fatalf("credential read failed: %v", err)
	}
	if err := applySecret(t.Context(), run, "integration-test", map[string]string{"password": secret["password"]}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var published struct {
		Kind       string
		Metadata   struct{ Name, Namespace string }
		StringData map[string]string
	}
	if err := json.Unmarshal(data, &published); err != nil {
		t.Fatal(err)
	}
	if published.Kind != "Secret" || published.Metadata.Name != "integration-test" || published.Metadata.Namespace != "global-secrets" || published.StringData["password"] != "private-password" {
		t.Fatal("incorrect published credential")
	}
	if err := loginKanidm(t.Context(), run, "kanidm.test", "admin", "private-password"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FAIL_APPLY", "yes")
	err = applySecret(t.Context(), run, "integration-test", map[string]string{"password": secret["password"]})
	if err == nil || !strings.HasPrefix(err.Error(), "apply secret integration-test:") || strings.Contains(err.Error(), "private-password") {
		t.Fatalf("unsafe Secret write failure: %v", err)
	}
	if logs.Len() != 0 || os.Getenv("KUBECONFIG") != "ambient.yaml" {
		t.Fatal("credential command leaked output or changed parent configuration")
	}
}
