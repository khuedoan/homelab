package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/khuedoan/homelab/toolbox/internal/process"
)

type giteaClient struct {
	base, user, password string
	client               *http.Client
}

func (g giteaClient) request(ctx context.Context, method, path string, body, result any) error {
	var input io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		input = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, g.base+path, input)
	if err != nil {
		return fmt.Errorf("build Gitea request: %w", err)
	}
	req.SetBasicAuth(g.user, g.password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("Gitea %s request failed", method)
	}
	defer resp.Body.Close()
	want := http.StatusOK
	if method == http.MethodPost {
		want = http.StatusCreated
	}
	if resp.StatusCode != want {
		return fmt.Errorf("Gitea %s returned HTTP %d", method, resp.StatusCode)
	}
	if json.NewDecoder(resp.Body).Decode(result) != nil {
		return fmt.Errorf("Gitea returned invalid JSON")
	}
	return nil
}

func (g giteaClient) integrations(ctx context.Context, run process.Runner, woodpecker string) error {
	path := "/api/v1/users/" + url.PathEscape(g.user) + "/tokens"
	var tokens []struct {
		Name string `json:"name"`
	}
	if err := g.request(ctx, "GET", path, nil, &tokens); err != nil {
		return err
	}
	found := false
	for _, token := range tokens {
		if token.Name == "renovate" {
			found = true
		}
	}
	if !found {
		var token struct {
			SHA1 string `json:"sha1"`
		}
		body := map[string]any{"name": "renovate", "scopes": []string{"write:repository", "read:user", "write:issue", "read:organization", "read:misc"}}
		if err := g.request(ctx, "POST", path, body, &token); err != nil {
			return err
		}
		if token.SHA1 == "" {
			return fmt.Errorf("Gitea returned an empty access token")
		}
		if err := applySecret(ctx, run, "gitea.renovate", map[string]string{"token": token.SHA1}); err != nil {
			return err
		}
	}
	path = "/api/v1/user/applications/oauth2"
	var apps []struct {
		Name string `json:"name"`
	}
	if err := g.request(ctx, "GET", path, nil, &apps); err != nil {
		return err
	}
	for _, app := range apps {
		if app.Name == "woodpecker" {
			return nil
		}
	}
	var app struct {
		ID     string `json:"client_id"`
		Secret string `json:"client_secret"`
	}
	body := map[string]any{"name": "woodpecker", "redirect_uris": []string{"https://" + woodpecker + "/authorize"}, "confidential_client": true}
	if err := g.request(ctx, "POST", path, body, &app); err != nil {
		return err
	}
	if app.ID == "" || app.Secret == "" {
		return fmt.Errorf("Gitea returned empty OAuth credentials")
	}
	return applySecret(ctx, run, "gitea.woodpecker", map[string]string{"client_id": app.ID, "client_secret": app.Secret})
}
