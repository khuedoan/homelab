package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/khuedoan/homelab/toolbox/internal/secrets"
	"github.com/spf13/cobra"
)

func newInfraSecretsCmd() *cobra.Command {
	var environment, kubeconfig string
	group := &cobra.Command{Use: "secrets", Short: "Publish infrastructure outputs to OpenBao"}
	sync := &cobra.Command{
		Use: "sync", Short: "Sync every Terragrunt unit's outputs in an environment",
		Long:    "Run from the repository root. Read existing Terraform outputs from infra/<environment> and write each to secret/infra/<unit>/<output> under the value field. Strings are preserved; other types are JSON encoded. Unchanged records and omitted paths are preserved. Requires terragrunt and an explicit cluster-admin kubeconfig. Secret values are never printed.",
		Example: "  toolbox infra secrets sync --environment staging --kubeconfig infra/staging/kubeconfig.yaml",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
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
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Synchronized %d infrastructure outputs to OpenBao.\n", len(records))
			return err
		},
	}
	sync.Flags().StringVar(&environment, "environment", "", "Environment directory under infra (required)")
	sync.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Target cluster kubeconfig (required)")
	requireFlags(sync, "environment", "kubeconfig")
	group.AddCommand(sync)
	return group
}
