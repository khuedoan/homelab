package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/khuedoan/homelab/toolbox/internal/state"
	"github.com/spf13/cobra"
)

func newInfraStateCmd() *cobra.Command {
	var account, bucket string
	group := &cobra.Command{Use: "state", Short: "Manage infrastructure state storage"}
	ensure := &cobra.Command{
		Use:     "ensure",
		Short:   "Create the Cloudflare R2 state bucket if absent",
		Long:    "Ensure an R2 bucket exists without changing existing buckets. Authentication uses the cf CLI OAuth login, matching the Cloudflare Terraform provider. Requires cf on PATH. No Kubernetes cluster or Terraform state is required.",
		Example: "  toolbox infra state ensure --bucket homelab-production-tfstate",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Minute)
			defer cancel()
			token, accountID, err := cloudflareLogin(ctx, account)
			if err != nil {
				return err
			}
			api, err := cloudflare.NewWithAPIToken(token)
			if err != nil {
				return err
			}
			if err := state.EnsureR2Bucket(ctx, api, accountID, bucket); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "State bucket %s is ready.\n", bucket)
			return err
		},
	}
	ensure.Flags().StringVar(&account, "account-id", os.Getenv("CLOUDFLARE_ACCOUNT_ID"), "Cloudflare account ID (defaults to CLOUDFLARE_ACCOUNT_ID)")
	ensure.Flags().StringVar(&bucket, "bucket", "", "R2 bucket name (required)")
	requireFlags(ensure, "bucket")
	group.AddCommand(ensure)
	return group
}
