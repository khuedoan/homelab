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
		Long:    "Ensure an R2 bucket exists without changing existing buckets. Authentication uses CLOUDFLARE_TFSTATE_API_TOKEN. No Kubernetes cluster or Terraform state is required.",
		Example: "  toolbox infra state ensure --bucket tfstate-production",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			token := os.Getenv("CLOUDFLARE_TFSTATE_API_TOKEN")
			if token == "" || account == "" || bucket == "" {
				return fmt.Errorf("CLOUDFLARE_TFSTATE_API_TOKEN, --account-id (or CLOUDFLARE_ACCOUNT_ID), and --bucket are required")
			}
			api, err := cloudflare.NewWithAPIToken(token)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Minute)
			defer cancel()
			if err := state.EnsureR2Bucket(ctx, api, account, bucket); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "State bucket %s is ready.\n", bucket)
			return nil
		},
	}
	ensure.Flags().StringVar(&account, "account-id", os.Getenv("CLOUDFLARE_ACCOUNT_ID"), "Cloudflare account ID (defaults to CLOUDFLARE_ACCOUNT_ID)")
	ensure.Flags().StringVar(&bucket, "bucket", "", "R2 bucket name (required)")
	_ = ensure.MarkFlagRequired("bucket")
	group.AddCommand(ensure)
	return group
}
