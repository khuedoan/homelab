package identity

import (
	"context"

	"github.com/khuedoan/homelab/toolbox/internal/process"
)

type User struct {
	Username string
	FullName string
	Email    string
}

func CreateUser(ctx context.Context, run process.Runner, user User) error {
	host, err := ingressHost(ctx, run, "kanidm", "kanidm")
	if err != nil {
		return err
	}
	options := []string{"--url", "https://" + host, "--name", "idm_admin"}
	steps := [][]string{
		append([]string{"person", "create", user.Username, user.FullName}, options...),
		append(append([]string{"person", "update", user.Username}, options...), "--mail", user.Email),
		append([]string{"group", "add-members", "editor", user.Username}, options...),
		append([]string{"person", "credential", "create-reset-token", user.Username}, options...),
	}
	for _, command := range steps {
		if err := run.Run(ctx, nil, "kanidm", command...); err != nil {
			return err
		}
	}
	return nil
}

func RecoverPassword(ctx context.Context, run process.Runner, account string) error {
	return run.Run(ctx, nil, "kubectl", "exec", "-it", "-n", "kanidm", "statefulset/kanidm", "--", "kanidmd", "recover-account", account)
}
