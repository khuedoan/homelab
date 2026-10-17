package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/khuedoan/homelab/cli/internal/secrets"
	"github.com/spf13/cobra"
)

func newInfraSecretsCmd() *cobra.Command {
	var environment, kubeconfig string
	group := &cobra.Command{Use: "secrets", Short: "Publish infrastructure secrets to OpenBao"}
	sync := &cobra.Command{
		Use: "sync", Short: "Sync every Terragrunt unit's secrets output in an environment",
		Long:    "Run from the repository root. Read each module's secrets output, a map of namespace/secret paths to string fields, and publish it to OpenBao. Other outputs are ignored. Unchanged records and omitted paths are preserved. Requires terragrunt and an explicit cluster-admin kubeconfig. Secret values are never printed.",
		Example: "  lab infra secrets sync --environment staging --kubeconfig infra/staging/kubeconfig.yaml",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Minute)
			defer cancel()
			records, err := secrets.EnvironmentRecords(ctx, environment)
			if err != nil {
				return err
			}
			if len(records) != 0 {
				if err := secrets.Sync(ctx, kubeconfig, records); err != nil {
					return err
				}
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Synchronized %d infrastructure secrets to OpenBao.\n", len(records))
			return err
		},
	}
	sync.Flags().StringVar(&environment, "environment", "", "Environment directory under infra (required)")
	sync.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Target cluster kubeconfig (required)")
	requireFlags(sync, "environment", "kubeconfig")
	group.AddCommand(sync)
	return group
}
