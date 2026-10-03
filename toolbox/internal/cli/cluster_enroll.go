package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/khuedoan/homelab/toolbox/internal/cluster"
	"github.com/spf13/cobra"
)

func newClusterEnrollCmd() *cobra.Command {
	var environment string
	var timeout time.Duration
	command := &cobra.Command{
		Use:     "enroll",
		Short:   "Join installed servers to the configured k3s cluster",
		Long:    "Enroll installed servers using verified root SSH connections. Host keys must already be trusted in known_hosts. Existing conflicting credentials or cluster state are never overwritten. This does not install nodes or recover a lost cluster.",
		Example: "  toolbox cluster enroll --environment staging\n  toolbox cluster enroll --environment production --timeout 15m",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if timeout <= 0 {
				return fmt.Errorf("--timeout must be positive")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			if err := cluster.Enroll(ctx, environment, cmd.ErrOrStderr()); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "All configured servers are Ready in the intended cluster. Export kubeconfig through the cluster stage next.")
			return nil
		},
	}
	command.Flags().StringVar(&environment, "environment", "", "Environment to enroll: staging or production (required)")
	command.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "Maximum time for enrollment and readiness checks")
	_ = command.MarkFlagRequired("environment")
	return command
}
