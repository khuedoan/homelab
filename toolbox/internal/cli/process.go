package cli

import (
	"github.com/khuedoan/homelab/toolbox/internal/process"
	"github.com/spf13/cobra"
)

func commandRunner(cmd *cobra.Command) process.Runner {
	var kubeconfig string
	if flag := cmd.Flag("kubeconfig"); flag != nil {
		kubeconfig = flag.Value.String()
		if kubeconfig == "" {
			kubeconfig = "infra/kubeconfig.yaml"
		}
	}
	return process.New(kubeconfig, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
}
