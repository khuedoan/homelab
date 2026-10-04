package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "toolbox",
		Short:         "Bootstrap infrastructure and manage k3s cluster access",
		Example:       "  toolbox cluster enroll --environment staging\n  toolbox cluster kubeconfig --environment staging --output infra/staging/kubeconfig.yaml",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cluster := &cobra.Command{Use: "cluster", Short: "Manage k3s cluster access and enrollment"}
	cluster.AddCommand(newClusterEnrollCmd(), newClusterKubeconfigCmd())
	infra := &cobra.Command{Use: "infra", Short: "Manage infrastructure prerequisites"}
	infra.AddCommand(newInfraStateCmd())
	root.AddCommand(cluster, infra)
	return root
}

func requireFlags(cmd *cobra.Command, flags ...string) {
	for _, flag := range flags {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
}

// Execute runs the toolbox CLI and exits on failure.
func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		// A diagnostic-write failure cannot change the already-failed exit status.
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
