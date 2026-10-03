package cli

import (
	"github.com/khuedoan/homelab/toolbox/internal/identity"
	"github.com/spf13/cobra"
)

func newIntegrationsSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use: "setup", Short: "Configure Gitea and Kanidm integrations", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return identity.SetupIntegrations(cmd.Context(), commandRunner(cmd))
		},
	}
}
