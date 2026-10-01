package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/spf13/cobra"
)

func newPostInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use: "setup", Short: "Configure Gitea and Kanidm integrations", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return postInstall(cmd) },
	}
}

// Commands handling credentials must not forward output, including error output.
func postInstallOutput(cmd *cobra.Command, name string, args ...string) ([]byte, error) {
	p := exec.CommandContext(cmd.Context(), name, args...)
	out, err := p.Output()
	if err != nil {
		return nil, fmt.Errorf("%s failed: %w", name, err)
	}
	return out, nil
}

func postInstallSecret(cmd *cobra.Command, namespace, name string) (map[string]string, error) {
	out, err := postInstallOutput(cmd, "kubectl", "get", "secret", "--namespace", namespace, name, "--output=json")
	if err != nil {
		return nil, err
	}
	var secret struct {
		Data map[string]string `json:"data"`
	}
	if json.Unmarshal(out, &secret) != nil {
		return nil, fmt.Errorf("secret %s/%s has invalid JSON", namespace, name)
	}
	values := make(map[string]string)
	for key, value := range secret.Data {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("secret %s/%s has invalid %s", namespace, name, key)
		}
		values[key] = string(decoded)
	}
	return values, nil
}

func postInstallApplySecret(cmd *cobra.Command, name string, values map[string]string) error {
	// kubectl can echo input on failure, so suppress its output for Secret writes.
	object := map[string]any{"apiVersion": "v1", "kind": "Secret", "metadata": map[string]string{"name": name, "namespace": "global-secrets"}, "type": "Opaque", "stringData": values}
	data, err := json.Marshal(object)
	if err != nil {
		return err
	}
	p := exec.CommandContext(cmd.Context(), "kubectl", "apply", "--filename=-")
	p.Stdin = bytes.NewReader(data)
	if err := p.Run(); err != nil {
		return fmt.Errorf("apply secret %s: %w", name, err)
	}
	return nil
}

type postInstallGitea struct {
	base, user, password string
	client               *http.Client
}

func (g postInstallGitea) request(cmd *cobra.Command, method, path string, body, result any) error {
	var input io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		input = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(cmd.Context(), method, g.base+path, input)
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

func (g postInstallGitea) integrations(cmd *cobra.Command, woodpecker string) error {
	path := "/api/v1/users/" + url.PathEscape(g.user) + "/tokens"
	var tokens []struct {
		Name string `json:"name"`
	}
	if err := g.request(cmd, "GET", path, nil, &tokens); err != nil {
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
		if err := g.request(cmd, "POST", path, body, &token); err != nil {
			return err
		}
		if token.SHA1 == "" {
			return fmt.Errorf("Gitea returned an empty access token")
		}
		if err := postInstallApplySecret(cmd, "gitea.renovate", map[string]string{"token": token.SHA1}); err != nil {
			return err
		}
	}
	path = "/api/v1/user/applications/oauth2"
	var apps []struct {
		Name string `json:"name"`
	}
	if err := g.request(cmd, "GET", path, nil, &apps); err != nil {
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
	if err := g.request(cmd, "POST", path, body, &app); err != nil {
		return err
	}
	if app.ID == "" || app.Secret == "" {
		return fmt.Errorf("Gitea returned empty OAuth credentials")
	}
	return postInstallApplySecret(cmd, "gitea.woodpecker", map[string]string{"client_id": app.ID, "client_secret": app.Secret})
}

func postInstallLogin(cmd *cobra.Command, host, account, password string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Minute)
	defer cancel()
	p := exec.CommandContext(ctx, "kanidm", "login", "--url", "https://"+host, "--name", account)
	terminal, err := pty.Start(p)
	if err != nil {
		return fmt.Errorf("start Kanidm login: %w", err)
	}
	defer terminal.Close()
	stop := context.AfterFunc(ctx, func() { terminal.Close() })
	defer stop()
	var prompt strings.Builder
	buffer := make([]byte, 256)
	sent := false
	for {
		n, readErr := terminal.Read(buffer)
		if !sent {
			prompt.Write(buffer[:n])
			if strings.Contains(strings.ToLower(prompt.String()), "password:") {
				if _, err := io.WriteString(terminal, password+"\n"); err != nil {
					cancel()
					_ = p.Wait()
					return fmt.Errorf("send Kanidm password failed")
				}
				sent = true
			}
			if prompt.Len() > 65536 {
				cancel()
				break
			}
		}
		if readErr != nil {
			break
		}
	}
	if err := p.Wait(); err != nil {
		return fmt.Errorf("Kanidm login for %s failed: %w", account, err)
	}
	if !sent {
		return fmt.Errorf("Kanidm login for %s did not request a password", account)
	}
	return nil
}

func postInstallKanidm(cmd *cobra.Command, host, dex string) error {
	for _, account := range []string{"admin", "idm_admin"} {
		out, err := postInstallOutput(cmd, "kubectl", "exec", "--namespace", "kanidm", "kanidm-0", "--", "kanidmd", "recover-account", "--output", "json", account)
		if err != nil {
			return err
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		var recovery struct {
			Password string `json:"password"`
		}
		if json.Unmarshal([]byte(lines[len(lines)-1]), &recovery) != nil || recovery.Password == "" {
			return fmt.Errorf("Kanidm recovery for %s returned no password", account)
		}
		if err := postInstallLogin(cmd, host, account, recovery.Password); err != nil {
			return err
		}
	}
	options := []string{"--url", "https://" + host, "--name", "idm_admin"}
	// A failed create is an error, not proof that the object already exists.
	commands := [][]string{
		{"group", "create"}, {"system", "oauth2", "create"},
		{"system", "oauth2", "warning-insecure-client-disable-pkce"},
		{"system", "oauth2", "create-scope-map"}, {"system", "oauth2", "show-basic-secret"},
	}
	arguments := [][]string{{"editor"}, {"dex", "dex", "https://" + dex + "/callback"}, {"dex"}, {"dex", "editor", "openid", "profile", "email", "groups"}, {"--output", "json", "dex"}}
	for i, command := range commands {
		args := append(append(command, options...), arguments[i]...)
		out, err := postInstallOutput(cmd, "kanidm", args...)
		if err != nil {
			return fmt.Errorf("Kanidm %s: %w", strings.Join(command, " "), err)
		}
		if i == len(commands)-1 {
			var secret struct {
				Secret string `json:"secret"`
			}
			if json.Unmarshal(out, &secret) != nil || secret.Secret == "" {
				return fmt.Errorf("Kanidm returned no OAuth secret")
			}
			return postInstallApplySecret(cmd, "kanidm.dex", map[string]string{"client_id": "dex", "client_secret": secret.Secret})
		}
	}
	return nil
}

func postInstall(cmd *cobra.Command) error {
	gitea, err := ingressHost(cmd, "gitea", "gitea")
	if err != nil {
		return err
	}
	kanidm, err := ingressHost(cmd, "kanidm", "kanidm")
	if err != nil {
		return err
	}
	dex, err := ingressHost(cmd, "dex", "dex")
	if err != nil {
		return err
	}
	woodpecker, err := ingressHost(cmd, "woodpecker", "woodpecker-server")
	if err != nil {
		return err
	}
	credentials, err := postInstallSecret(cmd, "gitea", "gitea-admin-secret")
	if err != nil {
		return err
	}
	if credentials["username"] == "" || credentials["password"] == "" {
		return fmt.Errorf("Gitea admin secret is missing credentials")
	}
	g := postInstallGitea{base: "http://" + gitea, user: credentials["username"], password: credentials["password"], client: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	if err := g.integrations(cmd, woodpecker); err != nil {
		return err
	}
	secret, err := postInstallSecret(cmd, "global-secrets", "dex.gitea")
	if err != nil {
		return err
	}
	if secret["client_secret"] == "" {
		return fmt.Errorf("Dex Gitea secret has no client_secret")
	}
	pods, err := commandOutput(cmd, "kubectl", "get", "pods", "--namespace", "gitea", "--selector=app=gitea", "--output=jsonpath={.items[0].metadata.name}")
	if err != nil {
		return err
	}
	pod := strings.TrimSpace(string(pods))
	if pod == "" {
		return fmt.Errorf("no Gitea pod found")
	}
	_, err = postInstallOutput(cmd, "kubectl", "exec", "--namespace", "gitea", pod, "--", "gitea", "admin", "auth", "add-oauth", "--name", "Dex", "--provider", "openidConnect", "--key", "gitea", "--secret", secret["client_secret"], "--auto-discover-url", "https://"+dex+"/.well-known/openid-configuration")
	if err != nil {
		return fmt.Errorf("configure Gitea Dex authentication: %w", err)
	}
	return postInstallKanidm(cmd, kanidm, dex)
}
