package identity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/khuedoan/homelab/toolbox/internal/process"
)

func integrationFake(t *testing.T, name, script string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return dir
}

func TestGiteaIntegrations(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "secrets")
			t.Setenv("SECRET_FILE", file)
			integrationFake(t, "kubectl", `cat >> "$SECRET_FILE"; printf '\n' >> "$SECRET_FILE"`)
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
			g := giteaClient{server.URL, "admin", "password", server.Client()}
			if err := g.integrations(t.Context(), process.New("", nil, nil, nil), "woodpecker.test"); err != nil {
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

func TestGiteaErrorsHideCredentials(t *testing.T) {
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
			g := giteaClient{server.URL, "admin", "private-password", server.Client()}
			err := g.integrations(t.Context(), process.New("", nil, nil, nil), "woodpecker.test")
			if err == nil || strings.Contains(err.Error(), "private-password") {
				t.Fatalf("unsafe error: %v", err)
			}
		})
	}
}

func TestCredentialOutputFailure(t *testing.T) {
	integrationFake(t, "kubectl", `echo private-password >&2; exit 17`)
	var logs bytes.Buffer
	run := process.New("", nil, &logs, &logs)

	_, err := readSecret(t.Context(), run, "gitea", "gitea-admin-secret")
	if err == nil || strings.Contains(err.Error(), "private-password") || logs.Len() != 0 {
		t.Fatalf("unsafe failure: %v, %s", err, &logs)
	}
}
