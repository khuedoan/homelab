package forgejo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/khuedoan/homelab/toolbox/internal/secrets"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Push publishes the current commit to the GitOps repository.
func Push(ctx context.Context, kubeconfig, domain, source string) error {
	if len(validation.IsDNS1123Subdomain(domain)) != 0 {
		return fmt.Errorf("--domain must be a DNS domain name")
	}
	credentials, err := secrets.Read(ctx, kubeconfig, "forgejo.admin")
	if err != nil {
		return err
	}
	if credentials["password"] == "" {
		return fmt.Errorf("forgejo administrator credential is absent")
	}
	return push(ctx, source, "https://git."+domain+"/ops/homelab.git", credentials["password"])
}

func push(ctx context.Context, source, remote, password string) error {
	temp, err := os.MkdirTemp("", "homelab-forgejo-push-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(temp) }()
	askpass := filepath.Join(temp, "askpass")
	if err := os.WriteFile(askpass, []byte("#!/bin/sh\ncase \"$1\" in\n *Username*) printf '%s\\n' forgejo_admin ;;\n *) printf '%s\\n' \"$HOMELAB_FORGEJO_PASSWORD\" ;;\nesac\n"), 0700); err != nil {
		return err
	}
	command := exec.CommandContext(ctx, "git", "-c", "credential.helper=", "push", remote, "HEAD:refs/heads/main")
	command.Dir = source
	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS="+askpass, "HOMELAB_FORGEJO_PASSWORD="+password)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("push GitOps repository: %w", err)
	}
	return nil
}
