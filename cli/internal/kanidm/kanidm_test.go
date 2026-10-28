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
	scopeGroups        map[string]string
	clients            map[string]entry
	callbacks          map[string]string
	public             map[string]bool
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
		f.groups[name] = true
		f.creates++

	default:
		f.t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(500)
	}
}

func (f *fakeProvider) clientRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && (r.URL.Path == "/v1/oauth2/_basic" || r.URL.Path == "/v1/oauth2/_public") {
		var client entry
		_ = json.NewDecoder(r.Body).Decode(&client)
		name := client.Attrs["name"][0]
		if _, exists := f.clients[name]; exists {
			f.t.Error("recreated existing client")
		}
		clientType := "basic"
		if f.public[name] {
			clientType = "public"
			if _, exists := client.Attrs["oauth2_allow_insecure_client_disable_pkce"]; exists {
				f.t.Error("public clients cannot override PKCE")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		if r.URL.Path != "/v1/oauth2/_"+clientType {
			f.t.Error("incorrect client type")
		}
		client.Attrs["class"] = []string{"oauth2_resource_server_" + clientType}
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
		checkOAuthPatch(f.t, r, name, f.callbacks[name], f.public[name])
	case r.Method == "POST" && len(parts) == 3 && parts[1] == "_scopemap":
		f.scopeGroups[name] = parts[2]
		client := f.clients[name]
		mapping := parts[2] + `: {"openid", "profile", "email", "groups"}`
		var mappings []string
		for _, old := range client.Attrs["oauth2_rs_scope_map"] {
			if !strings.HasPrefix(old, parts[2]+": ") {
				mappings = append(mappings, old)
			}
		}
		client.Attrs["oauth2_rs_scope_map"] = append(mappings, mapping)
		f.clients[name] = client
		checkScopes(f.t, r)
	case r.Method == "DELETE" && len(parts) == 3 && parts[1] == "_scopemap":
		client := f.clients[name]
		var remaining []string
		for _, mapping := range client.Attrs["oauth2_rs_scope_map"] {
			if !strings.HasPrefix(mapping, parts[2]+": ") {
				remaining = append(remaining, mapping)
			}
		}
		client.Attrs["oauth2_rs_scope_map"] = remaining
		f.clients[name] = client
	case r.Method == "GET" && len(parts) == 2 && parts[1] == "_basic_secret":
		if !f.readingSecrets || f.public[name] {
			f.t.Error("client ensure read a secret")
		}
		_ = json.NewEncoder(w).Encode(name + "-canary")
	default:
		f.t.Errorf("unexpected mutation or request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(500)
	}
}

func TestClientProvisioningAndSecretSync(t *testing.T) {
	f := &fakeProvider{t: t, groups: make(map[string]bool), scopeGroups: make(map[string]string), clients: make(map[string]entry), callbacks: map[string]string{
		"grafana":      "https://grafana.example.test/login/generic_oauth",
		"forgejo":      "https://git.example.test/user/oauth2/SSO/callback",
		"argocd":       "https://argocd.example.test/auth/callback",
		"actualbudget": "https://budget.example.test/openid/callback",
		"zot":          "https://registry.example.test/zot/auth/callback/oidc",
		"openbao":      "https://openbao.example.test/ui/vault/auth/oidc/oidc/callback",
	}, public: map[string]bool{"zot": true}}
	server := httptest.NewServer(f)
	defer server.Close()
	jar, _ := cookiejar.New(nil)
	c := &client{url: server.URL, http: &http.Client{Jar: jar}}
	ctx := context.Background()
	if err := c.login(ctx, "admin-canary"); err != nil {
		t.Fatal(err)
	}
	if err := c.ensureClient(ctx, ClientConfig{Name: "grafana", RedirectURIs: []string{f.callbacks["grafana"]}}); err == nil || !strings.Contains(err.Error(), "sso group ensure users") {
		t.Fatalf("expected missing group guidance, got %v", err)
	}
	if f.mutations != 0 {
		t.Fatal("client provisioning mutated state before its group existed")
	}
	for range 2 {
		for _, name := range []string{"users", "developers", "admins"} {
			if err := c.ensureGroup(ctx, name); err != nil {
				t.Fatal(err)
			}
		}
	}
	if f.creates != 3 || f.mutations != 3 {
		t.Fatal("group ensure recreated groups or changed existing memberships")
	}
	for range 2 {
		for name, callback := range f.callbacks {
			if err := c.ensureClient(ctx, ClientConfig{Name: name, DisplayName: name, RedirectURIs: []string{callback}, Group: map[string]string{"openbao": "admins", "argocd": "developers"}[name], Public: f.public[name], RS256: name == "actualbudget"}); err != nil {
				t.Fatal(err)
			}
		}
	}
	if f.creates != 9 {
		t.Fatalf("expected three shared groups and six clients, got %d creations", f.creates)
	}
	if !reflect.DeepEqual(f.scopeGroups, map[string]string{"grafana": "users", "forgejo": "users", "argocd": "developers", "actualbudget": "users", "zot": "users", "openbao": "admins"}) {
		t.Fatalf("unexpected client access groups: %v", f.scopeGroups)
	}
	for range 2 {
		if err := c.ensureGroup(ctx, "grafana_editors"); err != nil {
			t.Fatal(err)
		}
	}
	f.clients["grafana"].Attrs["oauth2_rs_scope_map"] = append(f.clients["grafana"].Attrs["oauth2_rs_scope_map"], `former_users@auth.example.test: {"openid"}`)
	for range 2 {
		if err := c.ensureClient(ctx, ClientConfig{Name: "grafana", RedirectURIs: []string{f.callbacks["grafana"]}, Group: "grafana_editors"}); err != nil {
			t.Fatal(err)
		}
	}
	if got := f.clients["grafana"].Attrs["oauth2_rs_scope_map"]; !reflect.DeepEqual(got, []string{`grafana_editors: {"openid", "profile", "email", "groups"}`}) {
		t.Fatalf("client retained previous login access: %v", got)
	}
	f.clients["grafana"].Attrs["oauth2_rs_scope_map"] = []string{"malformed"}
	beforeInvalid := f.mutations
	if err := c.ensureClient(ctx, ClientConfig{Name: "grafana", RedirectURIs: []string{f.callbacks["grafana"]}, Group: "grafana_editors"}); err == nil {
		t.Fatal("accepted malformed scope map")
	}
	if f.mutations != beforeInvalid {
		t.Fatal("mutated client before validating existing scope maps")
	}
	f.readingSecrets = true
	before := f.mutations
	want := map[string]map[string]string{
		"grafana/sso":      {"client_id": "grafana", "client_secret": "grafana-canary"},
		"forgejo/sso":      {"client_id": "forgejo", "client_secret": "forgejo-canary"},
		"argocd/sso":       {"client_id": "argocd", "client_secret": "argocd-canary"},
		"actualbudget/sso": {"client_id": "actualbudget", "client_secret": "actualbudget-canary"},
		"openbao/sso":      {"client_id": "openbao", "client_secret": "openbao-canary"},
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

func checkOAuthPatch(t *testing.T, r *http.Request, name, callback string, public bool) {
	t.Helper()
	var body entry
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Error(err)
	}
	expected := map[string][]string{
		"displayname": {name}, "oauth2_rs_origin_landing": {callback},
		"oauth2_rs_origin":             {callback},
		"oauth2_strict_redirect_uri":   {"true"},
		"oauth2_prefer_short_username": {"true"},
	}
	expected["oauth2_jwt_legacy_crypto_enable"] = []string{"false"}
	if name == "actualbudget" {
		expected["oauth2_jwt_legacy_crypto_enable"] = []string{"true"}
	}
	if public {
		expected["oauth2_allow_localhost_redirect"] = []string{"false"}
	} else {
		expected["oauth2_allow_insecure_client_disable_pkce"] = []string{"false"}
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
		err := EnsureClient(context.Background(), options, ClientConfig{Name: test.name, RedirectURIs: []string{test.redirect}})
		if err == nil || strings.Contains(err.Error(), "kubeconfig") {
			t.Fatalf("did not reject invalid client configuration first: %v", err)
		}
	}
}

func TestNativeRedirectValidation(t *testing.T) {
	const landing = "https://files.example.test/"
	for _, test := range []struct {
		name      string
		redirects []string
		wantError string
	}{
		{"web", []string{landing, "https://files.example.test/callback"}, "kanidm URL"},
		{"android", []string{landing, "oc://android.opencloud.eu"}, "kanidm URL"},
		{"ios", []string{landing, "oc://ios.opencloud.eu"}, "kanidm URL"},
		{"private scheme", []string{landing, "com.example.app:/callback"}, "kanidm URL"},
		{"opaque URI", []string{landing, "com.example.app:callback"}, "kanidm URL"},
		{"native landing", []string{"oc://ios.opencloud.eu"}, "redirect URI"},
		{"insecure landing", []string{"http://files.example.test/"}, "redirect URI"},
		{"insecure callback", []string{landing, "http://files.example.test/callback"}, "redirect URI"},
		{"relative callback", []string{landing, "/callback"}, "redirect URI"},
		{"empty native callback", []string{landing, "oc:"}, "redirect URI"},
		{"empty HTTPS host", []string{landing, "https:///callback"}, "redirect URI"},
		{"native credentials", []string{landing, "oc://user:password@ios.opencloud.eu"}, "redirect URI"},
		{"native fragment", []string{landing, "oc://ios.opencloud.eu#fragment"}, "redirect URI"},
		{"web credentials", []string{landing, "https://user:password@files.example.test/callback"}, "redirect URI"},
		{"web fragment", []string{landing, "https://files.example.test/callback#fragment"}, "redirect URI"},
		{"malformed URI", []string{landing, "oc://%invalid"}, "redirect URI"},
	} {
		t.Run(test.name, func(t *testing.T) {
			// An invalid server URL stops accepted callbacks before any cluster access.
			err := EnsureClient(t.Context(), Options{URL: "http://auth.example.test"}, ClientConfig{Name: "files", Public: true, RedirectURIs: test.redirects})
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("wanted %q, got %v", test.wantError, err)
			}
		})
	}
}

func TestMultipleRedirectURIs(t *testing.T) {
	var got entry
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v1/group/"):
			_, _ = w.Write([]byte(`{"attrs":{"name":["users"]}}`))
		case r.Method == "GET":
			_, _ = w.Write([]byte(`{"attrs":{"class":["oauth2_resource_server_basic"]}}`))
		case r.Method == "PATCH":
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Error(err)
			}
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/_scopemap/"):
			checkScopes(t, r)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	c := &client{url: server.URL, http: server.Client()}
	callbacks := []string{"https://photos.example.test/auth/login", "https://photos.example.test/user-settings"}
	for _, redirects := range [][]string{callbacks, callbacks, callbacks[:1]} {
		if err := c.ensureClient(context.Background(), ClientConfig{Name: "immich", RedirectURIs: redirects}); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got.Attrs["oauth2_rs_origin"], redirects) {
			t.Fatalf("callbacks: %v", got.Attrs["oauth2_rs_origin"])
		}
		if !reflect.DeepEqual(got.Attrs["oauth2_rs_origin_landing"], []string{"https://photos.example.test/auth/login"}) {
			t.Fatalf("landing: %v", got.Attrs["oauth2_rs_origin_landing"])
		}
	}
	err := EnsureClient(context.Background(), Options{URL: "https://auth.example.test", Kubeconfig: "absent"}, ClientConfig{Name: "immich", RedirectURIs: []string{callbacks[0], "http://photos.example.test/user-settings"}})
	if err == nil || !strings.Contains(err.Error(), "redirect URI") {
		t.Fatalf("invalid second callback reached cluster: %v", err)
	}
}

func TestNativePublicClientReconciliation(t *testing.T) {
	var got entry
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/v1/group/users":
			_, _ = w.Write([]byte(`{"attrs":{"name":["users"]}}`))
		case r.Method == "GET" && r.URL.Path == "/v1/oauth2/files":
			_, _ = w.Write([]byte(`{"attrs":{"class":["oauth2_resource_server_public"]}}`))
		case r.Method == "PATCH" && r.URL.Path == "/v1/oauth2/files":
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Error(err)
			}
		case r.Method == "POST" && r.URL.Path == "/v1/oauth2/files/_scopemap/users":
			checkScopes(t, r)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	c := &client{url: server.URL, http: server.Client()}
	for _, test := range []struct {
		allow bool
		want  string
	}{{true, "true"}, {true, "true"}, {false, "false"}} {
		config := ClientConfig{Name: "files", Public: true, AllowLocalhostRedirects: test.allow,
			RedirectURIs: []string{"https://files.example.test/", "oc://android.opencloud.eu", "oc://ios.opencloud.eu"}}
		if err := c.ensureClient(t.Context(), config); err != nil {
			t.Fatal(err)
		}
		want := map[string][]string{
			"displayname": {"files"}, "oauth2_rs_origin_landing": {"https://files.example.test/"},
			"oauth2_rs_origin":                {"https://files.example.test/", "oc://android.opencloud.eu", "oc://ios.opencloud.eu"},
			"oauth2_strict_redirect_uri":      {"true"},
			"oauth2_prefer_short_username":    {"true"},
			"oauth2_jwt_legacy_crypto_enable": {"false"},
			"oauth2_allow_localhost_redirect": {test.want},
		}
		if !reflect.DeepEqual(got.Attrs, want) {
			t.Fatalf("native public client settings: got %v, want %v", got.Attrs, want)
		}
	}
}

func TestClientRejectsInvalidAccessGroup(t *testing.T) {
	err := EnsureClient(t.Context(), Options{URL: "https://auth.example.test", Kubeconfig: "absent"}, ClientConfig{Name: "grafana", Group: "../invalid", RedirectURIs: []string{"https://grafana.example.test/login/generic_oauth"}})
	if err == nil || !strings.Contains(err.Error(), "access group") {
		t.Fatalf("invalid group reached cluster access: %v", err)
	}
}

func TestClientClaimMappings(t *testing.T) {
	mappings := map[string][]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/v1/group/users":
			_, _ = w.Write([]byte(`{"attrs":{"name":["users"]}}`))
		case r.Method == "GET" && r.URL.Path == "/v1/oauth2/immich":
			_, _ = w.Write([]byte(`{"attrs":{"class":["oauth2_resource_server_basic"]}}`))
		case r.Method == "PATCH" && r.URL.Path == "/v1/oauth2/immich":
		case r.Method == "POST" && r.URL.Path == "/v1/oauth2/immich/_scopemap/users":
			checkScopes(t, r)
		case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/v1/oauth2/immich/_claimmap/immich_role/"):
			var values []string
			if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
				t.Error(err)
			}
			mappings[r.URL.Path] = values
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	c := &client{url: server.URL, http: server.Client()}
	config := ClientConfig{Name: "immich", RedirectURIs: []string{"https://photos.example.test/auth/login"}, Claims: []ClaimMapping{
		{Name: "immich_role", Group: "users", Value: "user"},
		{Name: "immich_role", Group: "admins", Value: "admin"},
	}}
	for range 2 {
		if err := c.ensureClient(t.Context(), config); err != nil {
			t.Fatal(err)
		}
	}
	want := map[string][]string{
		"/v1/oauth2/immich/_claimmap/immich_role/users":  {"user"},
		"/v1/oauth2/immich/_claimmap/immich_role/admins": {"admin"},
	}
	if !reflect.DeepEqual(mappings, want) {
		t.Fatalf("claim mappings: got %v, want %v", mappings, want)
	}
}
