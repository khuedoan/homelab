package kanidm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/khuedoan/homelab/toolbox/internal/secrets"
	"k8s.io/apimachinery/pkg/util/validation"
)

type client struct {
	url     string
	http    *http.Client
	token   string
	session string
}

type entry struct {
	Attrs map[string][]string `json:"attrs"`
}

// Options identifies the target cluster and its SSO provider.
type Options struct {
	Kubeconfig string
	URL        string
	Bootstrap  bool
}

// ClientConfig describes an application registered with the SSO provider.
type ClientConfig struct {
	Name         string
	DisplayName  string
	RedirectURIs []string
	Public       bool
	RS256        bool
}

// EnsureClient provisions a client without rotating its secret or changing group membership.
func EnsureClient(ctx context.Context, options Options, config ClientConfig) error {
	name := config.Name
	if len(validation.IsDNS1123Label(name)) != 0 {
		return fmt.Errorf("client name must be a lowercase DNS label")
	}
	if len(config.RedirectURIs) == 0 {
		return fmt.Errorf("at least one redirect URI is required")
	}
	for _, uri := range config.RedirectURIs {
		if err := validateURL(uri); err != nil {
			return fmt.Errorf("redirect URI: %w", err)
		}
	}
	c, err := connect(ctx, options)
	if err != nil {
		return err
	}
	return c.ensureClient(ctx, config)
}

// SyncSecrets exports confidential client credentials without modifying Kanidm.
func SyncSecrets(ctx context.Context, options Options) error {
	c, err := connect(ctx, options)
	if err != nil {
		return err
	}
	records, err := c.clientRecords(ctx)
	if err != nil {
		return err
	}
	return secrets.Sync(ctx, options.Kubeconfig, records)
}

func validateURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
		return fmt.Errorf("a valid HTTPS URL without credentials or a fragment is required")
	}
	return nil
}

func connect(ctx context.Context, options Options) (*client, error) {
	if err := validateURL(options.URL); err != nil {
		return nil, fmt.Errorf("kanidm URL: %w", err)
	}
	credentials, err := secrets.Read(ctx, options.Kubeconfig, "kanidm/idm-admin")
	if err != nil {
		return nil, err
	}
	ready := exec.CommandContext(ctx, "kubectl", "--kubeconfig", options.Kubeconfig, "wait", "--namespace", "kanidm", "--for=condition=Ready", "pod/kanidm-0", "--timeout=15m")
	if err := ready.Run(); err != nil {
		return nil, fmt.Errorf("wait for Kanidm readiness failed")
	}
	password := credentials["password"]
	if password == "" {
		if !options.Bootstrap {
			return nil, fmt.Errorf("kanidm administrator credential is absent; use --bootstrap for initial account recovery")
		}
		password, err = recoverAdmin(ctx, options.Kubeconfig)
		if err != nil {
			return nil, err
		}
		if err := secrets.Sync(ctx, options.Kubeconfig, map[string]map[string]string{"kanidm/idm-admin": {"password": password}}); err != nil {
			return nil, err
		}
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c := &client{url: strings.TrimRight(options.URL, "/"), http: &http.Client{Jar: jar, Timeout: 30 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}}
	if err := c.login(ctx, password); err != nil {
		return nil, err
	}
	return c, nil
}

func recoverAdmin(ctx context.Context, kubeconfig string) (string, error) {
	command := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "exec", "-n", "kanidm", "kanidm-0", "--", "/sbin/kanidmd", "scripting", "recover-account", "-c", "/data/server.toml", "idm_admin")
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("recover Kanidm idm_admin failed")
	}
	var record struct {
		Status string `json:"status"`
		Output string `json:"output"`
	}
	if json.Unmarshal(output, &record) != nil || record.Status != "ok" || record.Output == "" {
		return "", fmt.Errorf("kanidm recovery returned no password")
	}
	return record.Output, nil
}

func (c *client) request(ctx context.Context, method, path string, input, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.url+path, body)
	if err != nil {
		return fmt.Errorf("invalid Kanidm request")
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.session != "" {
		req.Header.Set("X-KANIDM-AUTH-SESSION-ID", c.session)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("kanidm %s %s failed", method, path)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("kanidm %s %s returned HTTP %d", method, path, res.StatusCode)
	}
	if session := res.Header.Get("X-KANIDM-AUTH-SESSION-ID"); session != "" {
		c.session = session
	}
	if output == nil {
		return nil
	}
	if json.NewDecoder(res.Body).Decode(output) != nil {
		return fmt.Errorf("invalid Kanidm response for %s", path)
	}
	return nil
}

func (c *client) login(ctx context.Context, password string) error {
	steps := []any{
		map[string]any{"init2": map[string]any{"username": "idm_admin", "issue": "token", "privileged": true}},
		map[string]string{"begin": "password"},
		map[string]any{"cred": map[string]string{"password": password}},
	}
	for _, step := range steps {
		var response struct {
			State struct {
				Denied  *string `json:"denied"`
				Success string  `json:"success"`
			} `json:"state"`
		}
		if err := c.request(ctx, "POST", "/v1/auth", map[string]any{"step": step}, &response); err != nil {
			return err
		}
		if response.State.Denied != nil {
			return fmt.Errorf("kanidm administrator authentication denied")
		}
		c.token = response.State.Success
	}
	if c.token == "" {
		return fmt.Errorf("kanidm administrator authentication did not complete")
	}
	return nil
}

func (c *client) ensureClient(ctx context.Context, config ClientConfig) error {
	name, displayName := config.Name, config.DisplayName
	groupName := name + "_users"
	var group *entry
	if err := c.request(ctx, "GET", "/v1/group/"+groupName, nil, &group); err != nil {
		return err
	}
	if group == nil {
		if err := c.request(ctx, "POST", "/v1/group", entry{Attrs: map[string][]string{"name": {groupName}}}, nil); err != nil {
			return err
		}
	}
	if displayName == "" {
		displayName = name
	}
	var existing *entry
	if err := c.request(ctx, "GET", "/v1/oauth2/"+name, nil, &existing); err != nil {
		return err
	}
	clientType := "basic"
	if config.Public {
		clientType = "public"
	}
	if existing != nil && !slices.Contains(existing.Attrs["class"], "oauth2_resource_server_"+clientType) {
		return fmt.Errorf("SSO client %s has a different client type", name)
	}
	attrs := entry{Attrs: map[string][]string{
		"displayname":                     {displayName},
		"oauth2_rs_origin_landing":        {config.RedirectURIs[0]},
		"oauth2_rs_origin":                config.RedirectURIs,
		"oauth2_strict_redirect_uri":      {"true"},
		"oauth2_prefer_short_username":    {"true"},
		"oauth2_jwt_legacy_crypto_enable": {strconv.FormatBool(config.RS256)},
	}}
	if !config.Public {
		attrs.Attrs["oauth2_allow_insecure_client_disable_pkce"] = []string{"false"}
	}
	method, path := "PATCH", "/v1/oauth2/"+name
	if existing == nil {
		attrs.Attrs["name"] = []string{name}
		method, path = "POST", "/v1/oauth2/_"+clientType
	}
	if err := c.request(ctx, method, path, attrs, nil); err != nil {
		return err
	}
	return c.request(ctx, "POST", "/v1/oauth2/"+name+"/_scopemap/"+groupName, []string{"openid", "profile", "email", "groups"}, nil)
}

func (c *client) clientRecords(ctx context.Context) (map[string]map[string]string, error) {
	var clients []entry
	if err := c.request(ctx, "GET", "/v1/oauth2", nil, &clients); err != nil {
		return nil, err
	}
	records := make(map[string]map[string]string)
	for _, client := range clients {
		if !slices.Contains(client.Attrs["class"], "oauth2_resource_server_basic") {
			continue
		}
		names := client.Attrs["name"]
		if len(names) != 1 || len(validation.IsDNS1123Label(names[0])) != 0 {
			return nil, fmt.Errorf("kanidm returned an invalid client name")
		}
		name := names[0]
		var secret string
		if err := c.request(ctx, "GET", "/v1/oauth2/"+name+"/_basic_secret", nil, &secret); err != nil {
			return nil, err
		}
		if strings.TrimSpace(secret) == "" {
			return nil, fmt.Errorf("SSO client %s has no basic secret", name)
		}
		records[name+"/sso"] = map[string]string{"client_id": name, "client_secret": secret}
	}
	return records, nil
}
