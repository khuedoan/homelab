package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/khuedoan/homelab/toolbox/internal/kanidm"
	"github.com/spf13/cobra"
)

func newSSOCmd() *cobra.Command {
	var options kanidm.Options
	group := &cobra.Command{Use: "sso", Short: "Provision SSO clients and sync their credentials to OpenBao"}
	group.PersistentFlags().StringVar(&options.Kubeconfig, "kubeconfig", "", "Target cluster kubeconfig (required)")
	group.PersistentFlags().StringVar(&options.URL, "url", "", "Kanidm HTTPS URL (required)")
	group.PersistentFlags().BoolVar(&options.Bootstrap, "bootstrap", false, "Recover idm_admin only when its OpenBao credential is absent")
	if err := group.MarkPersistentFlagRequired("kubeconfig"); err != nil {
		panic(err)
	}
	if err := group.MarkPersistentFlagRequired("url"); err != nil {
		panic(err)
	}
	var redirectURI, displayName string
	clients := &cobra.Command{Use: "client", Short: "Manage SSO clients"}
	ensure := &cobra.Command{
		Use: "ensure NAME", Short: "Create or update a client and its NAME_users access group",
		Long: "Ensure the client settings and access group exist without rotating its secret or changing group membership. Requires kubectl on PATH. Credentials are published separately with sso secrets sync.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()
			if err := kanidm.EnsureClient(ctx, options, args[0], displayName, redirectURI); err != nil {
				return err
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "SSO client %s is ready.\n", args[0])
			return err
		},
	}
	ensure.Flags().StringVar(&redirectURI, "redirect-uri", "", "Application HTTPS callback URI (required)")
	ensure.Flags().StringVar(&displayName, "display-name", "", "Client display name (defaults to NAME)")
	requireFlags(ensure, "redirect-uri")
	clients.AddCommand(ensure)
	secretGroup := &cobra.Command{Use: "secrets", Short: "Publish SSO client credentials to OpenBao"}
	sync := &cobra.Command{
		Use: "sync", Short: "Sync all confidential clients to sso/NAME in OpenBao",
		Long: "Read client IDs and existing secrets from Kanidm and publish client_id and client_secret fields to sso/NAME. Does not create or change clients. Unchanged OpenBao records retain their versions. Requires kubectl on PATH. Secret values are never printed.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()
			if err := kanidm.SyncSecrets(ctx, options); err != nil {
				return err
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "SSO client credentials synchronized to OpenBao.")
			return err
		},
	}
	secretGroup.AddCommand(sync)
	group.AddCommand(clients, secretGroup)
	return group
}
