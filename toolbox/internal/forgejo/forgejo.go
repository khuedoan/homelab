package forgejo

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	sdk "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	"github.com/khuedoan/homelab/toolbox/internal/secrets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/wait"
)

// Sync provisions the GitOps repository and Woodpecker OAuth credentials.
func Sync(ctx context.Context, kubeconfig, domain string) error {
	if len(validation.IsDNS1123Subdomain(domain)) != 0 {
		return fmt.Errorf("--domain must be a DNS domain name")
	}
	admin, err := secrets.Read(ctx, kubeconfig, "forgejo.admin")
	if err != nil {
		return err
	}
	if admin["password"] == "" {
		return fmt.Errorf("forgejo administrator credential is absent")
	}
	c, err := sdk.NewClient("https://git."+domain,
		sdk.SetBasicAuth("forgejo_admin", admin["password"]),
		sdk.SetContext(ctx),
		sdk.SetForgejoVersion(""),
		sdk.SetHTTPClient(&http.Client{Timeout: 30 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}),
	)
	if err != nil {
		return fmt.Errorf("create Forgejo client: %w", err)
	}
	err = wait.PollUntilContextCancel(ctx, time.Second, true, func(context.Context) (bool, error) {
		_, response, err := c.ServerVersion()
		if response == nil || response.StatusCode == http.StatusBadGateway || response.StatusCode == http.StatusServiceUnavailable || response.StatusCode == http.StatusGatewayTimeout {
			return false, nil
		}
		return err == nil, err
	})
	if err != nil {
		return fmt.Errorf("wait for Forgejo readiness: %w", err)
	}
	if err := syncGitOps(c); err != nil {
		return err
	}
	existing, err := secrets.Read(ctx, kubeconfig, "forgejo.woodpecker")
	if err != nil {
		return err
	}
	credentials, err := syncOAuth(c, "https://ci."+domain+"/authorize", existing)
	if err != nil {
		return err
	}
	return secrets.Sync(ctx, kubeconfig, map[string]map[string]string{"forgejo.woodpecker": credentials})
}

func syncGitOps(c *sdk.Client) error {
	if _, response, err := c.GetOrg("ops"); err != nil {
		if response == nil || response.StatusCode != http.StatusNotFound {
			return err
		}
		if _, _, err := c.CreateOrg(sdk.CreateOrgOption{Name: "ops", Description: "Operations"}); err != nil {
			return err
		}
	}
	if _, response, err := c.GetRepo("ops", "homelab"); err != nil {
		if response == nil || response.StatusCode != http.StatusNotFound {
			return err
		}
		if _, _, err := c.CreateOrgRepo("ops", sdk.CreateRepoOption{Name: "homelab", Private: false, DefaultBranch: "master"}); err != nil {
			return err
		}
	}
	return nil
}

func syncOAuth(c *sdk.Client, callback string, existing map[string]string) (map[string]string, error) {
	var found *sdk.Oauth2
	for page := 1; ; page++ {
		apps, _, err := c.ListOauth2(sdk.ListOauth2Option{ListOptions: sdk.ListOptions{Page: page, PageSize: 50}})
		if err != nil {
			return nil, err
		}
		for _, app := range apps {
			if app.Name != "woodpecker" {
				continue
			}
			if found != nil {
				return nil, fmt.Errorf("multiple Forgejo OAuth clients named woodpecker")
			}
			found = app
		}
		if len(apps) < 50 {
			break
		}
	}
	desired := sdk.CreateOauth2Option{Name: "woodpecker", ConfidentialClient: true, RedirectURIs: []string{callback}}
	if found != nil && found.ConfidentialClient && reflect.DeepEqual(found.RedirectURIs, desired.RedirectURIs) && existing["client_id"] == found.ClientID && existing["client_secret"] != "" && existing["redirect_uri"] == callback {
		return existing, nil
	}
	var app *sdk.Oauth2
	var err error
	if found == nil {
		app, _, err = c.CreateOauth2(desired)
	} else {
		app, _, err = c.UpdateOauth2(found.ID, desired)
	}
	if err != nil {
		return nil, err
	}
	if app.ClientID == "" || app.ClientSecret == "" {
		return nil, fmt.Errorf("forgejo returned incomplete OAuth credentials")
	}
	return map[string]string{"client_id": app.ClientID, "client_secret": app.ClientSecret, "redirect_uri": callback}, nil
}
