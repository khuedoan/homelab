package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/khuedoan/homelab/toolbox/internal/forgejo"
	"github.com/spf13/cobra"
)

func newInfraForgejoCmd() *cobra.Command {
	var kubeconfig, domain string
	group := &cobra.Command{Use: "forgejo", Short: "Provision the GitOps repository and Woodpecker OAuth in Forgejo"}
	sync := &cobra.Command{
		Use: "sync", Short: "Provision Forgejo and synchronize Woodpecker OAuth credentials to OpenBao",
		Long: "Create the ops organization and its homelab GitOps repository if absent. Reuse unchanged Woodpecker OAuth credentials; regenerate only when configuration changes or credentials are absent. Requires an explicit cluster-admin kubeconfig and trusted HTTPS Forgejo endpoint. Secret values are never printed.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()
			if err := forgejo.Sync(ctx, kubeconfig, domain); err != nil {
				return err
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "GitOps repository and Woodpecker OAuth credentials synchronized.")
			return err
		},
	}
	sync.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Target cluster kubeconfig (required)")
	sync.Flags().StringVar(&domain, "domain", "", "Application base domain (required)")
	requireFlags(sync, "kubeconfig", "domain")
	group.AddCommand(sync)
	return group
}
