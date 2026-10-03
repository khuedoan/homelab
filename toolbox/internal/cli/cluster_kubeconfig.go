package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/khuedoan/homelab/toolbox/internal/cluster"
	"github.com/spf13/cobra"
)

func newClusterKubeconfigCmd() *cobra.Command {
	var environment, output string
	var timeout time.Duration
	command := &cobra.Command{
		Use:     "kubeconfig",
		Short:   "Export authenticated kubeconfig for the control-plane VIP",
		Long:    "Fetch kubeconfig from the configured initializer over verified root SSH. Verify authenticated VIP readiness before writing credentials. Files are published atomically with mode 0600. Use --output - to send credentials to stdout.",
		Example: "  toolbox cluster kubeconfig --environment staging\n  toolbox cluster kubeconfig --environment production --output ~/.kube/homelab.yaml",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if timeout <= 0 {
				return fmt.Errorf("--timeout must be positive")
			}
			if output == "" {
				return fmt.Errorf("--output must be a path or -")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			fmt.Fprintf(cmd.ErrOrStderr(), "Fetching kubeconfig for %s...\n", environment)
			data, err := cluster.Kubeconfig(ctx, environment)
			if err != nil {
				return err
			}
			if err := writeKubeconfig(output, data, cmd.OutOrStdout()); err != nil {
				return err
			}
			if output != "-" {
				fmt.Fprintf(cmd.OutOrStdout(), "Kubeconfig written to %s (0600).\n", output)
			}
			return nil
		},
	}
	command.Flags().StringVar(&environment, "environment", "", "Environment to export: staging or production (required)")
	command.Flags().StringVarP(&output, "output", "o", "infra/kubeconfig.yaml", "Destination path, or - for stdout")
	command.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "Maximum time for export and readiness checks")
	_ = command.MarkFlagRequired("environment")
	return command
}

func writeKubeconfig(destination string, data []byte, stdout io.Writer) error {
	if destination == "-" {
		_, err := stdout.Write(data)
		return err
	}
	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".kubeconfig-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), destination)
}
