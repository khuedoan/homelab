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
		Short:         "CLI tools for managing the homelab",
		Example:       "  toolbox status\n  toolbox users create johndoe \"John Doe\" johndoe@example.com\n  toolbox backup setup --namespace jellyfin --pvc jellyfin-data",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().String("kubeconfig", os.Getenv("KUBECONFIG"), "Kubeconfig path (defaults to infra/kubeconfig.yaml)")
	root.AddCommand(newStatusCmd(), newBackupCmd())
	for _, group := range []struct {
		name, description string
		commands          []*cobra.Command
	}{
		{"apps", "Manage application charts", []*cobra.Command{newAppCreateCmd()}},
		{"argocd", "Manage Argo CD", []*cobra.Command{newArgoCDPasswordCmd()}},
		{"cluster", "Manage k3s cluster access and enrollment", []*cobra.Command{newClusterEnrollCmd(), newClusterKubeconfigCmd()}},
		{"dns", "Inspect DNS records", []*cobra.Command{newDNSConfigCmd()}},
		{"helm", "Inspect Helm charts", []*cobra.Command{newHelmDiffCmd()}},
		{"infra", "Manage infrastructure prerequisites", []*cobra.Command{newInfraStateCmd()}},
		{"integrations", "Configure Gitea and Kanidm integrations", []*cobra.Command{newIntegrationsSetupCmd()}},
		{"screenshots", "Capture application screenshots", []*cobra.Command{newScreenshotsCmd()}},
		{"users", "Manage Kanidm accounts", []*cobra.Command{newOnboardUserCmd(), newKanidmPasswordCmd()}},
		{"wireguard", "Inspect WireGuard peers", []*cobra.Command{newWireguardConfigCmd()}},
	} {
		command := &cobra.Command{Use: group.name, Short: group.description}
		command.AddCommand(group.commands...)
		root.AddCommand(command)
	}
	return root
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
