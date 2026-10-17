package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/khuedoan/homelab/cli/internal/kanidm"
	"github.com/spf13/cobra"
)

func newUserCmd() *cobra.Command {
	group := &cobra.Command{Use: "user", Short: "Manage homelab users"}
	var options kanidm.Options
	var config kanidm.UserConfig
	create := &cobra.Command{
		Use: "create USERNAME", Short: "Create a user and print their credential setup link",
		Long: "Create a Kanidm person in users and any additional existing groups. Share the one-use link privately so the user can set up passkeys or a password and TOTP. Existing users are rejected without changing credentials or memberships. Requires kubectl on PATH and the administrator credential in OpenBao.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Minute)
			defer cancel()
			config.Username = args[0]
			link, err := kanidm.CreateUser(ctx, options, config)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "User %s created. Share this one-use credential setup link privately:\n%s\n", config.Username, link)
			return err
		},
	}
	create.Flags().StringVar(&options.Kubeconfig, "kubeconfig", "", "Target cluster kubeconfig (required)")
	create.Flags().StringVar(&options.URL, "url", "", "Kanidm HTTPS URL (required)")
	create.Flags().StringVar(&config.DisplayName, "display-name", "", "User's full name (required)")
	create.Flags().StringVar(&config.Email, "email", "", "User's email address (required)")
	create.Flags().StringArrayVar(&config.Groups, "group", nil, "Additional existing access group, such as developers or admins (repeat; users is always included)")
	requireFlags(create, "kubeconfig", "url", "display-name", "email")
	group.AddCommand(create)
	return group
}
