package cli

import (
	"fmt"

	"github.com/khuedoan/homelab/toolbox/internal/identity"
	"github.com/spf13/cobra"
)

func newKanidmPasswordCmd() *cobra.Command {
	return &cobra.Command{Use: "reset-password ACCOUNT", Short: "Recover a Kanidm account password", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: Kanidm admin can do anything in the cluster, only use it for just enough initial setup or in emergencies.")
			return identity.RecoverPassword(cmd.Context(), commandRunner(cmd), args[0])
		},
	}
}

func newOnboardUserCmd() *cobra.Command {
	return &cobra.Command{Use: "create USERNAME FULL_NAME EMAIL", Short: "Create a Kanidm user and issue a credential reset token", Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			user := identity.User{Username: args[0], FullName: args[1], Email: args[2]}
			return identity.CreateUser(cmd.Context(), commandRunner(cmd), user)
		},
	}
}
