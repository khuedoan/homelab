package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/khuedoan/homelab/cli/internal/state"
	"github.com/spf13/cobra"
)

func newInfraStateCmd() *cobra.Command {
	group := &cobra.Command{Use: "state", Short: "Manage infrastructure state storage"}
	var account, bucket string
	group.PersistentFlags().StringVar(&account, "account-id", os.Getenv("CLOUDFLARE_ACCOUNT_ID"), "Cloudflare account ID (defaults to CLOUDFLARE_ACCOUNT_ID)")
	group.PersistentFlags().StringVar(&bucket, "bucket", "", "R2 bucket name (required)")
	if err := group.MarkPersistentFlagRequired("bucket"); err != nil {
		panic(err)
	}
	ensure := &cobra.Command{
		Use:     "ensure",
		Short:   "Create the Cloudflare R2 state bucket if absent",
		Long:    "Ensure an R2 bucket exists without changing existing buckets. Authentication uses the cf CLI OAuth login, matching the Cloudflare Terraform provider. Requires cf on PATH. No Kubernetes cluster or Terraform state is required.",
		Example: "  lab infra state ensure --bucket homelab-production-tfstate",
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
	credentials := &cobra.Command{
		Use:   "credentials",
		Short: "Issue expiring R2 credentials for the OpenTofu backend",
		Long:  "Emit AWS credential-process JSON using the cf CLI OAuth login. Grants object read/write access to one R2 bucket for one hour. Requires OAuth account token creation and R2 object write scopes. Credentials remain in process memory and expire automatically.",
		Args:  cobra.NoArgs,
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
			result, err := state.CreateR2Credentials(ctx, api, accountID, bucket)
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		},
	}
	group.AddCommand(ensure, credentials)
	return group
}
