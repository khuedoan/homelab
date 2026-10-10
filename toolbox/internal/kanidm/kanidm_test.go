package kanidm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type fakeProvider struct {
	t                  *testing.T
	groups             map[string]bool
	clients            map[string]entry
	callbacks          map[string]string
	creates, mutations int
	readingSecrets     bool
}

func (f *fakeProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Path == "/v1/auth" {
		testAuthentication(f.t, w, r)
		return
	}
	if r.Header.Get("Authorization") != "Bearer bearer-canary" {
		f.t.Error("missing authorization")
	}
	if r.Method != "GET" {
		f.mutations++
		if f.readingSecrets {
			f.t.Error("secret sync mutated Kanidm")
		}
	}
	switch {
	case r.Method == "GET" && r.URL.Path == "/v1/oauth2":
		entries := []entry{{Attrs: map[string][]string{"name": {"public-client"}, "class": {"oauth2_resource_server_public"}}}}
		for _, client := range f.clients {
			entries = append(entries, client)
		}
		_ = json.NewEncoder(w).Encode(entries)
	case strings.HasPrefix(r.URL.Path, "/v1/group"):
		f.groupRequest(w, r)
	case strings.HasPrefix(r.URL.Path, "/v1/oauth2/"):
		f.clientRequest(w, r)
	default:
		f.t.Errorf("unexpected mutation or request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(500)
	}
}

func (f *fakeProvider) groupRequest(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v1/group/"):
		name := strings.TrimPrefix(r.URL.Path, "/v1/group/")
		if f.groups[name] {
			_ = json.NewEncoder(w).Encode(entry{Attrs: map[string][]string{"name": {name}, "member": {"existing-user"}}})
		} else {
			_, _ = w.Write([]byte(`null`))
		}
	case r.Method == "POST" && r.URL.Path == "/v1/group":
		var group entry
		_ = json.NewDecoder(r.Body).Decode(&group)
		name := group.Attrs["name"][0]
		if _, ok := f.callbacks[strings.TrimSuffix(name, "_users")]; !ok {
			f.t.Errorf("unexpected group %s", name)
		}
		f.groups[name] = true
		f.creates++

	default:
		f.t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(500)
	}
}

func (f *fakeProvider) clientRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.URL.Path == "/v1/oauth2/_basic" {
		var client entry
		_ = json.NewDecoder(r.Body).Decode(&client)
		name := client.Attrs["name"][0]
		if _, exists := f.clients[name]; exists {
			f.t.Error("recreated existing client")
		}
		client.Attrs["class"] = []string{"oauth2_resource_server_basic"}
		f.clients[name] = client
		f.creates++
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/oauth2/"), "/")
	name := parts[0]
	switch {
	case r.Method == "GET" && len(parts) == 1:
		if client, exists := f.clients[name]; exists {
			_ = json.NewEncoder(w).Encode(client)
		} else {
			_, _ = w.Write([]byte(`null`))
		}
	case r.Method == "PATCH" && len(parts) == 1:
		checkOAuthPatch(f.t, r, name, f.callbacks[name])
	case r.Method == "POST" && len(parts) == 3 && parts[1] == "_scopemap" && parts[2] == name+"_users":
		checkScopes(f.t, r)
	case r.Method == "GET" && len(parts) == 2 && parts[1] == "_basic_secret":
		if !f.readingSecrets {
			f.t.Error("client ensure read a secret")
		}
		_ = json.NewEncoder(w).Encode(name + "-canary")
	default:
		f.t.Errorf("unexpected mutation or request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(500)
	}
}

func TestClientProvisioningAndSecretSync(t *testing.T) {
	f := &fakeProvider{t: t, groups: make(map[string]bool), clients: make(map[string]entry), callbacks: map[string]string{
		"grafana": "https://grafana.example.test/login/generic_oauth",
		"forgejo": "https://git.example.test/user/oauth2/sso/callback",
	}}
	server := httptest.NewServer(f)
	defer server.Close()
	jar, _ := cookiejar.New(nil)
	c := &client{url: server.URL, http: &http.Client{Jar: jar}}
	ctx := context.Background()
	if err := c.login(ctx, "admin-canary"); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		for name, callback := range f.callbacks {
			if err := c.ensureClient(ctx, name, name, callback); err != nil {
				t.Fatal(err)
			}
		}
	}
	if f.creates != 4 {
		t.Fatalf("expected two groups and two clients, got %d creations", f.creates)
	}
	f.readingSecrets = true
	before := f.mutations
	want := map[string]map[string]string{
		"sso/grafana": {"client_id": "grafana", "client_secret": "grafana-canary"},
		"sso/forgejo": {"client_id": "forgejo", "client_secret": "forgejo-canary"},
	}
	for range 2 {
		got, err := c.clientRecords(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected secret records: %v", got)
		}
	}
	if f.mutations != before {
		t.Fatal("secret sync changed clients or memberships")
	}
}

func testAuthentication(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	var req map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		t.Error(err)
	}
	step := string(req["step"])
	switch {
	case strings.Contains(step, "init2"):
		if !strings.Contains(step, `"privileged":true`) {
			t.Error("expected write-capable authentication")
		}
		http.SetCookie(w, &http.Cookie{Name: "auth-session-id", Value: "session", Path: "/"})
		_, _ = w.Write([]byte(`{"state":{"choose":["password"]}}`))
	case strings.Contains(step, "begin"):
		if _, err := r.Cookie("auth-session-id"); err != nil {
			t.Error("lost authentication session")
		}
		_, _ = w.Write([]byte(`{"state":{"continue":["password"]}}`))
	default:
		if !strings.Contains(step, "admin-canary") {
			t.Error("missing administrator credential")
		}
		_, _ = w.Write([]byte(`{"state":{"success":"bearer-canary"}}`))
	}
}

func TestRequestErrorsDoNotExposeCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte("secret-canary"))
	}))
	defer server.Close()
	c := &client{url: server.URL, http: server.Client()}
	err := c.login(context.Background(), "secret-canary")
	if err == nil || strings.Contains(err.Error(), "secret-canary") {
		t.Fatalf("unsafe error: %v", err)
	}
}

func checkOAuthPatch(t *testing.T, r *http.Request, name, callback string) {
	t.Helper()
	var body entry
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Error(err)
	}
	expected := map[string][]string{
		"displayname": {name}, "oauth2_rs_origin_landing": {callback},
		"oauth2_rs_origin":                          {callback},
		"oauth2_allow_insecure_client_disable_pkce": {"false"}, "oauth2_prefer_short_username": {"true"},
	}
	if !reflect.DeepEqual(body.Attrs, expected) {
		t.Errorf("unexpected OAuth patch: %v", body.Attrs)
	}
}

func checkScopes(t *testing.T, r *http.Request) {
	t.Helper()
	var scopes []string
	if err := json.NewDecoder(r.Body).Decode(&scopes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(scopes, []string{"openid", "profile", "email", "groups"}) {
		t.Errorf("unexpected scopes %v", scopes)
	}
}

func TestClientValidationBeforeClusterAccess(t *testing.T) {
	options := Options{URL: "https://auth.example.test", Kubeconfig: "absent"}
	for _, test := range []struct{ name, redirect string }{
		{"../invalid", "https://app.example.test/callback"},
		{"app", "http://app.example.test/callback"},
		{"app", "https://user:password@app.example.test/callback"},
	} {
		err := EnsureClient(context.Background(), options, test.name, "", test.redirect)
		if err == nil || strings.Contains(err.Error(), "kubeconfig") {
			t.Fatalf("did not reject invalid client configuration first: %v", err)
		}
	}
}
