package identity

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/khuedoan/homelab/toolbox/internal/process"
)

func SetupIntegrations(ctx context.Context, run process.Runner) error {
	gitea, err := ingressHost(ctx, run, "gitea", "gitea")
	if err != nil {
		return err
	}
	kanidm, err := ingressHost(ctx, run, "kanidm", "kanidm")
	if err != nil {
		return err
	}
	dex, err := ingressHost(ctx, run, "dex", "dex")
	if err != nil {
		return err
	}
	woodpecker, err := ingressHost(ctx, run, "woodpecker", "woodpecker-server")
	if err != nil {
		return err
	}
	credentials, err := readSecret(ctx, run, "gitea", "gitea-admin-secret")
	if err != nil {
		return err
	}
	if credentials["username"] == "" || credentials["password"] == "" {
		return fmt.Errorf("Gitea admin secret is missing credentials")
	}
	g := giteaClient{
		base: "http://" + gitea, user: credentials["username"], password: credentials["password"],
		client: &http.Client{
			Timeout:       30 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
	if err := g.integrations(ctx, run, woodpecker); err != nil {
		return err
	}
	secret, err := readSecret(ctx, run, "global-secrets", "dex.gitea")
	if err != nil {
		return err
	}
	if secret["client_secret"] == "" {
		return fmt.Errorf("Dex Gitea secret has no client_secret")
	}
	pods, err := run.Output(ctx, "kubectl", "get", "pods", "--namespace", "gitea", "--selector=app=gitea", "--output=jsonpath={.items[0].metadata.name}")
	if err != nil {
		return err
	}
	pod := strings.TrimSpace(string(pods))
	if pod == "" {
		return fmt.Errorf("no Gitea pod found")
	}
	_, err = run.PrivateOutput(ctx, "kubectl", "exec", "--namespace", "gitea", pod, "--", "gitea", "admin", "auth", "add-oauth", "--name", "Dex", "--provider", "openidConnect", "--key", "gitea", "--secret", secret["client_secret"], "--auto-discover-url", "https://"+dex+"/.well-known/openid-configuration")
	if err != nil {
		return fmt.Errorf("configure Gitea Dex authentication: %w", err)
	}
	return configureKanidm(ctx, run, kanidm, dex)
}
