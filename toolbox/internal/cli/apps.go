package cli

import (
	"github.com/khuedoan/homelab/toolbox/internal/charts"
	"github.com/spf13/cobra"
)

func newAppCreateCmd() *cobra.Command {
	return &cobra.Command{Use: "create NAME", Short: "Scaffold an application Helm chart", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return charts.Create(args[0])
		},
	}
}
