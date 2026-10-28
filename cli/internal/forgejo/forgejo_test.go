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
	var team *sdk.Team
	assigned := false
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
			if input["name"] != "homelab" || input["default_branch"] != "master" || input["private"] != false {
				t.Errorf("unexpected repository: %v", input)
			}
			repo = true
			creates++
			_, _ = w.Write([]byte(`{}`))
		case "GET /api/v1/orgs/ops/teams":
			teams := []*sdk.Team{}
			if team != nil {
				teams = append(teams, team)
			}
			_ = json.NewEncoder(w).Encode(teams)
		case "POST /api/v1/orgs/ops/teams", "PATCH /api/v1/teams/7":
			var input sdk.CreateTeamOption
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				t.Fatal(err)
			}
			if input.Name != "developers" || input.Permission != sdk.AccessModeWrite || input.CanCreateOrgRepo || input.IncludesAllRepositories || !reflect.DeepEqual(input.UnitsMap, map[string]string{"repo.code": "write", "repo.pulls": "write"}) {
				t.Errorf("unexpected developer team permissions: %+v", input)
			}
			team = &sdk.Team{ID: 7, Name: input.Name, Permission: input.Permission, UnitsMap: input.UnitsMap}
			if r.Method == "POST" {
				creates++
			}
			_ = json.NewEncoder(w).Encode(team)
		case "PUT /api/v1/teams/7/repos/ops/homelab":
			assigned = true
			w.WriteHeader(http.StatusNoContent)
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
	if creates != 3 || !assigned {
		t.Fatalf("created %d resources; want 3 and GitOps repository assignment", creates)
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

func TestDeveloperTeamLookupFailureDoesNotGrantAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("mutated permissions after failed lookup: %s", r.Method)
		}
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()
	if err := syncDevelopers(testClient(t, server)); err == nil {
		t.Fatal("ignored denied team lookup")
	}
}

func TestDeveloperTeamFoundOnLaterPage(t *testing.T) {
	pages := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/orgs/ops/teams":
			page := r.URL.Query().Get("page")
			pages = append(pages, page)
			if page == "1" {
				teams := make([]sdk.Team, 50)
				for i := range teams {
					teams[i] = sdk.Team{ID: int64(i + 20), Name: fmt.Sprintf("other-%d", i)}
				}
				_ = json.NewEncoder(w).Encode(teams)
			} else {
				_, _ = w.Write([]byte(`[{"id":7,"name":"developers"}]`))
			}
		case "PATCH /api/v1/teams/7":
			_, _ = w.Write([]byte(`{"id":7,"name":"developers"}`))
		case "PUT /api/v1/teams/7/repos/ops/homelab":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()
	if err := syncDevelopers(testClient(t, server)); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(pages, []string{"1", "2"}) {
		t.Fatalf("pages: %v", pages)
	}
}
