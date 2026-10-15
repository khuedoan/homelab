package state

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

// R2Credentials implements the AWS SDK credential_process response protocol.
type R2Credentials struct {
	Version         int
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string
	Expiration      time.Time
}

// CreateR2Credentials grants one hour of object access to one state bucket.
// The account token expires directly, so no permanent parent token is needed.
func CreateR2Credentials(ctx context.Context, api *cloudflare.API, account, bucket string) (R2Credentials, error) {
	endpoint := "/accounts/" + account + "/tokens"
	response, err := api.Raw(ctx, http.MethodGet, endpoint+"/permission_groups", nil, nil)
	if err != nil {
		return R2Credentials{}, fmt.Errorf("list account token permissions: %w", err)
	}
	var groups []cloudflare.APITokenPermissionGroups
	if err := json.Unmarshal(response.Result, &groups); err != nil {
		return R2Credentials{}, fmt.Errorf("decode account token permissions: %w", err)
	}
	var permission cloudflare.APITokenPermissionGroups
	for _, group := range groups {
		if group.Name == "Workers R2 Storage Bucket Item Write" {
			permission = group
			break
		}
	}
	if permission.ID == "" {
		return R2Credentials{}, errors.New("cf OAuth login cannot grant R2 object access; run cf auth login with R2 object write and account token creation scopes")
	}
	name := "terraform-" + bucket
	resource := "com.cloudflare.edge.r2.bucket." + account + "_default_" + bucket
	expires := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	grant := cloudflare.APIToken{
		Name:      name,
		ExpiresOn: &expires,
		Policies: []cloudflare.APITokenPolicies{{
			Effect:           "allow",
			Resources:        map[string]interface{}{resource: "*"},
			PermissionGroups: []cloudflare.APITokenPermissionGroups{{ID: permission.ID}},
		}},
	}
	if err := pruneR2Credentials(ctx, api, endpoint, grant); err != nil {
		return R2Credentials{}, err
	}
	response, err = api.Raw(ctx, http.MethodPost, endpoint, grant, nil)
	if err != nil {
		return R2Credentials{}, fmt.Errorf("create expiring R2 state token: %w", err)
	}
	var token cloudflare.APIToken
	if err := json.Unmarshal(response.Result, &token); err != nil || token.ID == "" || token.Value == "" {
		return R2Credentials{}, errors.New("account token response has no valid R2 credentials")
	}
	secret := sha256.Sum256([]byte(token.Value))
	return R2Credentials{Version: 1, AccessKeyID: token.ID, SecretAccessKey: hex.EncodeToString(secret[:]), Expiration: expires}, nil
}

func pruneR2Credentials(ctx context.Context, api *cloudflare.API, endpoint string, grant cloudflare.APIToken) error {
	var expired []string
	for page := 1; ; page++ {
		response, err := api.Raw(ctx, http.MethodGet, fmt.Sprintf("%s?include_expired=true&per_page=50&page=%d", endpoint, page), nil, nil)
		if err != nil {
			return fmt.Errorf("list expired R2 state tokens: %w", err)
		}
		var tokens []cloudflare.APIToken
		if err := json.Unmarshal(response.Result, &tokens); err != nil {
			return fmt.Errorf("decode R2 state tokens: %w", err)
		}
		for _, token := range tokens {
			if expiredR2Credential(token, grant) {
				expired = append(expired, token.ID)
			}
		}
		if len(tokens) < 50 {
			break
		}
	}
	for _, id := range expired {
		if _, err := api.Raw(ctx, http.MethodDelete, endpoint+"/"+id, nil, nil); err != nil {
			var missing *cloudflare.NotFoundError
			if !errors.As(err, &missing) {
				return fmt.Errorf("delete expired R2 state token: %w", err)
			}
		}
	}

	return nil
}

func expiredR2Credential(token, grant cloudflare.APIToken) bool {
	if token.ID == "" || token.Name != grant.Name || token.ExpiresOn == nil || !token.ExpiresOn.Before(time.Now()) || len(token.Policies) != 1 {
		return false
	}
	policy, expected := token.Policies[0], grant.Policies[0]
	return policy.Effect == expected.Effect && maps.Equal(policy.Resources, expected.Resources) && len(policy.PermissionGroups) == 1 && policy.PermissionGroups[0].ID == expected.PermissionGroups[0].ID
}

// EnsureR2Bucket creates a missing bucket and rechecks after a failed create.
// Inspection errors other than not-found never trigger creation.
func EnsureR2Bucket(ctx context.Context, api *cloudflare.API, account, bucket string) error {
	rc := cloudflare.AccountIdentifier(account)
	_, err := api.GetR2Bucket(ctx, rc, bucket)
	if err == nil {
		return nil
	}
	var missing *cloudflare.NotFoundError
	if !errors.As(err, &missing) {
		return fmt.Errorf("inspect state bucket %q: %w", bucket, err)
	}
	_, err = api.CreateR2Bucket(ctx, rc, cloudflare.CreateR2BucketParameters{Name: bucket})
	if err != nil {
		// A competing provisioner or a lost response can leave the bucket ready.
		if _, checkErr := api.GetR2Bucket(ctx, rc, bucket); checkErr == nil {
			return nil
		}
		return fmt.Errorf("create state bucket %q: %w", bucket, err)
	}
	return nil
}
