package secrets

import (
	"context"
	"fmt"
	"net/url"

	bao "github.com/openbao/openbao/api/v2"
)

// ConfigureOIDC enables UI SSO using the synchronized OpenBao client credentials.
func ConfigureOIDC(ctx context.Context, kubeconfig, issuer, redirectURI string) error {
	for _, value := range []string{issuer, redirectURI} {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
			return fmt.Errorf("OpenBao SSO requires valid HTTPS issuer and callback URLs")
		}
	}
	client, err := openBaoClient(ctx, kubeconfig)
	if err != nil {
		return err
	}
	credentials, err := readRecord(ctx, client.KVv2("secret"), "openbao/sso")
	if err != nil {
		return err
	}
	if credentials["client_id"] != "openbao" || credentials["client_secret"] == "" {
		return fmt.Errorf("OpenBao SSO credentials are absent; run sso secrets sync first")
	}
	mounts, err := client.Sys().ListAuthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("read OpenBao auth methods failed")
	}
	if mount := mounts["oidc/"]; mount != nil {
		if mount.Type != "oidc" {
			return fmt.Errorf("OpenBao auth/oidc is already used by a different auth method")
		}
	} else if err := client.Sys().EnableAuthWithOptionsWithContext(ctx, "oidc", &bao.EnableAuthOptions{Type: "oidc", Description: "SSO"}); err != nil {
		return fmt.Errorf("enable OpenBao OIDC failed")
	}
	if _, err := client.Logical().WriteWithContext(ctx, "auth/oidc/config", map[string]any{
		"oidc_discovery_url": issuer,
		"oidc_client_id":     credentials["client_id"],
		"oidc_client_secret": credentials["client_secret"],
		"default_role":       "sso",
		"jwt_supported_algs": []string{"ES256"},
	}); err != nil {
		return fmt.Errorf("configure OpenBao OIDC failed")
	}
	if _, err := client.Logical().WriteWithContext(ctx, "auth/oidc/role/sso", map[string]any{
		"role_type":             "oidc",
		"user_claim":            "sub",
		"bound_audiences":       []string{"openbao"},
		"allowed_redirect_uris": []string{redirectURI},
		"oidc_scopes":           []string{"profile", "email", "groups"},
		"token_policies":        []string{"default"},
		"token_ttl":             "1h",
	}); err != nil {
		return fmt.Errorf("configure OpenBao SSO role failed")
	}
	return nil
}
