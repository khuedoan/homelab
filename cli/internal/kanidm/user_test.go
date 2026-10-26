package kanidm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestCreateUserPreservesMembershipsAndRejectsRerun(t *testing.T) {
	var person *entry
	members := map[string][]string{"users": {"existing-user"}, "developers": {"existing-developer"}, "admins": {"existing-admin"}}
	writes, intents := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer admin-token" {
			t.Error("missing admin authorization")
		}
		if r.Method != "GET" {
			writes++
		}
		switch {
		case r.Method == "GET" && r.URL.Path == "/v1/person/jane":
			_ = json.NewEncoder(w).Encode(person)
		case r.Method == "GET" && r.URL.Path == "/v1/person/jane/_credential/_update_intent":
			intents++
			_, _ = w.Write([]byte(`{"token":"setup+token&canary"}`))
		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/v1/group/"):
			name := strings.TrimPrefix(r.URL.Path, "/v1/group/")
			if _, ok := members[name]; !ok {
				t.Errorf("unknown group %s", name)
			}
			_ = json.NewEncoder(w).Encode(entry{Attrs: map[string][]string{"name": {name}}})
		case r.Method == "POST" && r.URL.Path == "/v1/person":
			if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
				t.Error(err)
			}
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/_attr/member"):
			name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/group/"), "/_attr/member")
			var added []string
			if err := json.NewDecoder(r.Body).Decode(&added); err != nil {
				t.Error(err)
			}
			members[name] = append(members[name], added...)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
		}
	}))
	defer server.Close()
	c := &client{url: server.URL, http: server.Client(), token: "admin-token"}
	config := UserConfig{Username: "jane", DisplayName: "Jane Doe", Email: "jane@example.test", Groups: []string{"developers", "admins", "users", "developers"}}
	link, err := c.createUser(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	if link != server.URL+"/ui/reset?token=setup%2Btoken%26canary" {
		t.Fatalf("wrong setup link %q", link)
	}
	wantPerson := &entry{Attrs: map[string][]string{"name": {"jane"}, "displayname": {"Jane Doe"}, "mail": {"jane@example.test"}}}
	if !reflect.DeepEqual(person, wantPerson) {
		t.Fatalf("person = %v, want %v", person, wantPerson)
	}
	wantMembers := map[string][]string{"users": {"existing-user", "jane"}, "developers": {"existing-developer", "jane"}, "admins": {"existing-admin", "jane"}}
	if !reflect.DeepEqual(members, wantMembers) || writes != 4 || intents != 1 {
		t.Fatalf("members=%v writes=%d intents=%d", members, writes, intents)
	}
	config.Groups = nil
	if _, err := c.createUser(t.Context(), config); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("rerun did not reject existing user: %v", err)
	}
	if writes != 4 || intents != 1 || !reflect.DeepEqual(members, wantMembers) {
		t.Fatal("rerun changed memberships or issued a credential reset")
	}
}

func TestCreateUserMissingGroupDoesNotCreatePerson(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("mutated before preflight completed: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Path == "/v1/group/users" {
			_, _ = w.Write([]byte(`{"attrs":{"name":["users"]}}`))
			return
		}
		_, _ = w.Write([]byte(`null`))
	}))
	defer server.Close()
	c := &client{url: server.URL, http: server.Client()}
	_, err := c.createUser(t.Context(), UserConfig{Username: "jane", Groups: []string{"developers"}})
	if err == nil || !strings.Contains(err.Error(), "access group developers does not exist") {
		t.Fatalf("missing group was accepted: %v", err)
	}
}

func TestCreateUserDefaultsToUsersAndReportsPartialCreation(t *testing.T) {
	for _, fail := range []string{"membership", "intent", "empty token"} {
		t.Run(fail, func(t *testing.T) {
			var requests []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.Path)
				switch r.URL.Path {
				case "/v1/person/jane":
					_, _ = w.Write([]byte(`null`))
				case "/v1/group/users":
					_, _ = w.Write([]byte(`{"attrs":{"name":["users"]}}`))
				case "/v1/person":
				case "/v1/group/users/_attr/member":
					if fail == "membership" {
						w.WriteHeader(403)
					}
				case "/v1/person/jane/_credential/_update_intent":
					if fail == "intent" {
						w.WriteHeader(403)
					} else {
						_, _ = w.Write([]byte(`{"token":""}`))
					}
				default:
					t.Errorf("unexpected group or request: %s", r.URL.Path)
				}
			}))
			defer server.Close()
			c := &client{url: server.URL, http: server.Client()}
			link, err := c.createUser(t.Context(), UserConfig{Username: "jane"})
			if err == nil || !strings.Contains(err.Error(), "user jane was created") || link != "" {
				t.Fatalf("partial creation not reported: link=%q err=%v requests=%v", link, err, requests)
			}
		})
	}
}

func TestCreateUserValidationBeforeClusterAccess(t *testing.T) {
	valid := UserConfig{Username: "jane", DisplayName: "Jane Doe", Email: "jane@example.test"}
	for _, test := range []struct {
		name string
		edit func(*UserConfig)
	}{
		{"username", func(c *UserConfig) { c.Username = "../admin" }},
		{"display name", func(c *UserConfig) { c.DisplayName = " " }},
		{"email", func(c *UserConfig) { c.Email = "Jane <jane@example.test>" }},
		{"email", func(c *UserConfig) { c.Email = "invalid" }},
		{"access group", func(c *UserConfig) { c.Groups = []string{"../admins"} }},
	} {
		config := valid
		test.edit(&config)
		_, err := CreateUser(t.Context(), Options{}, config)
		if err == nil || !strings.Contains(err.Error(), test.name) {
			t.Fatalf("invalid %s reached cluster access: %v", test.name, err)
		}
	}
}
