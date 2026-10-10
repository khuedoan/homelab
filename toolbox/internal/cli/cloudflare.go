package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func cloudflareLogin(ctx context.Context, account string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "env", "-u", "CLOUDFLARE_API_TOKEN", "-u", "CLOUDFLARE_API_KEY", "-u", "CLOUDFLARE_EMAIL", "CF_SEND_TELEMETRY=false", "cf", "--quiet", "auth", "whoami")
	raw, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("read cf OAuth login failed; run cf auth login")
	}
	var identity struct {
		AuthSource string `json:"authSource"`
		Accounts   []struct {
			ID string `json:"id"`
		} `json:"accounts"`
	}
	if err := json.Unmarshal(raw, &identity); err != nil {
		return "", "", fmt.Errorf("decode cf OAuth identity failed")
	}
	path, ok := strings.CutPrefix(identity.AuthSource, "OAuth token from ")
	if !ok || path == "" {
		return "", "", fmt.Errorf("cf OAuth login required; run cf auth login")
	}
	if account == "" {
		if len(identity.Accounts) != 1 || identity.Accounts[0].ID == "" {
			return "", "", fmt.Errorf("specify --account-id when cf has multiple or no accounts")
		}
		account = identity.Accounts[0].ID
	}
	raw, err = os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("read cf OAuth credential file failed")
	}
	var credentials struct {
		Token string `json:"oauth_token"`
	}
	if err := json.Unmarshal(raw, &credentials); err != nil || credentials.Token == "" {
		return "", "", fmt.Errorf("cf OAuth credential file has no valid token")
	}
	return credentials.Token, account, nil
}
