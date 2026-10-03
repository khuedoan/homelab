package cli

import (
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{Use: "status", Short: "Show Argo CD applications and cluster ingresses", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			for _, args := range [][]string{{"get", "applicationsets", "--namespace", "argocd"}, {"get", "applications", "--namespace", "argocd"}, {"get", "ingress", "--all-namespaces"}} {
				if err := commandRunner(cmd).Run(cmd.Context(), nil, "kubectl", args...); err != nil {
					return err
				}
			}
			return nil
		}}
}
