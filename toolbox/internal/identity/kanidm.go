package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/khuedoan/homelab/toolbox/internal/process"
)

func loginKanidm(ctx context.Context, run process.Runner, host, account, password string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	p := run.Command(ctx, "kanidm", "login", "--url", "https://"+host, "--name", account)
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

func configureKanidm(ctx context.Context, run process.Runner, host, dex string) error {
	for _, account := range []string{"admin", "idm_admin"} {
		out, err := run.PrivateOutput(ctx, "kubectl", "exec", "--namespace", "kanidm", "kanidm-0", "--", "kanidmd", "recover-account", "--output", "json", account)
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
		if err := loginKanidm(ctx, run, host, account, recovery.Password); err != nil {
			return err
		}
	}
	options := []string{"--url", "https://" + host, "--name", "idm_admin"}
	steps := []struct{ command, arguments []string }{
		{[]string{"group", "create"}, []string{"editor"}},
		{[]string{"system", "oauth2", "create"}, []string{"dex", "dex", "https://" + dex + "/callback"}},
		{[]string{"system", "oauth2", "warning-insecure-client-disable-pkce"}, []string{"dex"}},
		{[]string{"system", "oauth2", "create-scope-map"}, []string{"dex", "editor", "openid", "profile", "email", "groups"}},
	}
	for _, step := range steps {
		args := append(append(step.command, options...), step.arguments...)
		if _, err := run.PrivateOutput(ctx, "kanidm", args...); err != nil {
			return fmt.Errorf("Kanidm %s: %w", strings.Join(step.command, " "), err)
		}
	}
	args := append([]string{"system", "oauth2", "show-basic-secret"}, options...)
	args = append(args, "--output", "json", "dex")
	out, err := run.PrivateOutput(ctx, "kanidm", args...)
	if err != nil {
		return fmt.Errorf("Kanidm system oauth2 show-basic-secret: %w", err)
	}
	var secret struct {
		Secret string `json:"secret"`
	}
	if json.Unmarshal(out, &secret) != nil || secret.Secret == "" {
		return fmt.Errorf("Kanidm returned no OAuth secret")
	}
	return applySecret(ctx, run, "kanidm.dex", map[string]string{"client_id": "dex", "client_secret": secret.Secret})
}
