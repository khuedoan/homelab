package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestUserCreateRepeatedGroups(t *testing.T) {
	cmd := newUserCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"create", "jane", "--kubeconfig", "absent", "--url", "https://auth.example.test", "--display-name", "Jane Doe", "--email", "jane@example.test", "--group", "../invalid", "--group", "developers"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "invalid access group") {
		t.Fatalf("first repeated group was lost or not validated: %v", err)
	}
}

func TestUserCreateUsesStoredAdminAndPrintsSetupLink(t *testing.T) {
	operationFake(t, "kubectl", "exit 0\n")
	var person map[string][]string
	var groups []string
	authSteps := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/namespaces/openbao/secrets/openbao-unseal":
			_, _ = w.Write([]byte(`{"apiVersion":"v1","kind":"Secret","data":{"vault-root":"cm9vdC1jYW5hcnk="}}`))
		case "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/sys/health":
			_, _ = w.Write([]byte(`{"initialized":true,"sealed":false}`))
		case "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/sys/mounts":
			_, _ = w.Write([]byte(`{"data":{"secret/":{"type":"kv","options":{"version":"2"}}}}`))
		case "/api/v1/namespaces/openbao/services/http:openbao:8200/proxy/v1/secret/data/kanidm/idm-admin":
			if r.Method != "GET" {
				t.Error("changed stored administrator credential")
			}
			_, _ = w.Write([]byte(`{"data":{"data":{"password":"stored-admin-canary"},"metadata":{"version":1}}}`))
		case "/v1/auth":
			authSteps++
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if authSteps == 3 && !strings.Contains(string(body["step"]), "stored-admin-canary") {
				t.Error("did not use stored administrator password")
			}
			_, _ = w.Write([]byte(`{"state":{"success":"bearer-canary"}}`))
		case "/v1/person/jane":
			_, _ = w.Write([]byte(`null`))
		case "/v1/group/users", "/v1/group/developers":
			_, _ = w.Write([]byte(`{"attrs":{"name":["existing"]}}`))
		case "/v1/person":
			var body struct {
				Attrs map[string][]string `json:"attrs"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			person = body.Attrs
		case "/v1/group/users/_attr/member", "/v1/group/developers/_attr/member":
			if r.Method != "POST" {
				t.Errorf("membership must append, got %s", r.Method)
			}
			groups = append(groups, r.URL.Path)
		case "/v1/person/jane/_credential/_update_intent":
			_, _ = w.Write([]byte(`{"token":"setup-canary","expiry_time":1792076400}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
		}
	}))
	defer server.Close()
	original, originalClient := http.DefaultTransport, http.DefaultClient.Transport
	http.DefaultTransport, http.DefaultClient.Transport = server.Client().Transport, server.Client().Transport
	t.Cleanup(func() { http.DefaultTransport, http.DefaultClient.Transport = original, originalClient })
	config := clientapi.Config{CurrentContext: "test", Clusters: map[string]*clientapi.Cluster{"test": {Server: server.URL}}, AuthInfos: map[string]*clientapi.AuthInfo{"test": {}}, Contexts: map[string]*clientapi.Context{"test": {Cluster: "test", AuthInfo: "test"}}}
	kubeconfig := filepath.Join(t.TempDir(), "kubeconfig")
	if err := clientcmd.WriteToFile(config, kubeconfig); err != nil {
		t.Fatal(err)
	}
	cmd := newUserCmd()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"create", "jane", "--kubeconfig", kubeconfig, "--url", server.URL, "--display-name", "Jane Doe", "--email", "jane@example.test", "--group", "developers"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), server.URL+"/ui/reset?token=setup-canary") {
		t.Fatalf("missing setup link: %s", output.String())
	}
	want := map[string][]string{"name": {"jane"}, "displayname": {"Jane Doe"}, "mail": {"jane@example.test"}}
	if !reflect.DeepEqual(person, want) || !reflect.DeepEqual(groups, []string{"/v1/group/users/_attr/member", "/v1/group/developers/_attr/member"}) || authSteps != 3 {
		t.Fatalf("incorrect onboarding: person=%v groups=%v authSteps=%d", person, groups, authSteps)
	}
}
