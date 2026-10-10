package forgejo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	sdk "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
)

func TestOAuthSyncPreservesCredentialsAndRepairsMissingSecret(t *testing.T) {
	var apps []*sdk.Oauth2
	writes := 0
	server := oauthTestServer(t, &apps, &writes)
	defer server.Close()
	c := testClient(t, server)
	first, err := syncOAuth(c, "https://ci.example.test/authorize", nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"client_id": "client-test", "client_secret": "secret-test-1", "redirect_uri": "https://ci.example.test/authorize"}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("credentials = %v", first)
	}
	second, err := syncOAuth(c, "https://ci.example.test/authorize", first)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(second, want) || writes != 1 {
		t.Fatalf("unchanged sync rotated credentials, writes=%d", writes)
	}
	if _, err := syncOAuth(c, "https://ci.example.test/authorize", nil); err != nil {
		t.Fatal(err)
	}
	if writes != 2 {
		t.Fatal("missing secret was not repaired")
	}
	if _, err := syncOAuth(c, "https://ci.changed.test/authorize", second); err != nil {
		t.Fatal(err)
	}
	if writes != 3 || apps[0].RedirectURIs[0] != "https://ci.changed.test/authorize" {
		t.Fatal("callback change was not reconciled")
	}
	recovered, err := syncOAuth(c, "https://ci.changed.test/authorize", second)
	if err != nil {
		t.Fatal(err)
	}
	if writes != 4 || recovered["client_secret"] != "secret-test-4" {
		t.Fatal("did not recover credentials after callback update was interrupted")
	}
}

func TestGitOpsCreatesOnlyMissingResources(t *testing.T) {
	org, repo := false, false
	creates := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/orgs/ops":
			if !org {
				w.WriteHeader(http.StatusNotFound)
			}
			_, _ = w.Write([]byte(`{}`))
		case "POST /api/v1/orgs":
			var input map[string]any
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				t.Fatal(err)
			}
			if input["username"] != "ops" || input["description"] != "Operations" {
				t.Errorf("unexpected organization: %v", input)
			}
			org = true
			creates++
			_, _ = w.Write([]byte(`{}`))
		case "GET /api/v1/repos/ops/homelab":
			if !repo {
				w.WriteHeader(http.StatusNotFound)
			}
			_, _ = w.Write([]byte(`{}`))
		case "POST /api/v1/org/ops/repos":
			var input map[string]any
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				t.Fatal(err)
			}
			if input["name"] != "homelab" || input["default_branch"] != "main" || input["private"] != false {
				t.Errorf("unexpected repository: %v", input)
			}
			repo = true
			creates++
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
		}
	}))
	defer server.Close()
	c := testClient(t, server)
	for range 2 {
		if err := syncGitOps(c); err != nil {
			t.Fatal(err)
		}
	}
	if creates != 2 {
		t.Fatalf("created %d resources; want 2", creates)
	}
}

func TestGitOpsDoesNotCreateOnUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("created resource after failed lookup")
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	c := testClient(t, server)
	if err := syncGitOps(c); err == nil {
		t.Fatal("ignored unauthorized lookup")
	}
}

func oauthTestServer(t *testing.T, apps *[]*sdk.Oauth2, writes *int) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "forgejo_admin" || password != "admin-test" {
			t.Error("missing administrator authentication")
		}
		switch r.Method {
		case "GET":
			if err := json.NewEncoder(w).Encode(*apps); err != nil {
				t.Error(err)
			}
		case "POST", "PATCH":
			var app sdk.Oauth2
			if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
				t.Error(err)
			}
			if app.Name != "woodpecker" || !app.ConfidentialClient || len(app.RedirectURIs) != 1 {
				t.Errorf("invalid OAuth client: %+v", app)
			}
			(*writes)++
			app.ID, app.ClientID, app.ClientSecret = 1, "client-test", fmt.Sprintf("secret-test-%d", *writes)
			listed := app
			listed.ClientSecret = ""
			*apps = []*sdk.Oauth2{&listed}
			if err := json.NewEncoder(w).Encode(app); err != nil {
				t.Error(err)
			}
		default:
			t.Errorf("unexpected request %s", r.Method)
		}
	}))
	return server
}

func testClient(t *testing.T, server *httptest.Server) *sdk.Client {
	t.Helper()
	c, err := sdk.NewClient(server.URL, sdk.SetHTTPClient(server.Client()), sdk.SetBasicAuth("forgejo_admin", "admin-test"), sdk.SetForgejoVersion(""))
	if err != nil {
		t.Fatal(err)
	}
	return c
}
